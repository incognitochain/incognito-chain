package constantbft

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/pkg/errors"

	"github.com/constant-money/constant-chain/common"

	"github.com/constant-money/constant-chain/wire"
)

type BFTProtocol struct {
	cBFTMsg   chan wire.Message
	EngineCfg *EngineConfig

	cQuit    chan struct{}
	cTimeout chan struct{}

	phase string

	pendingBlock interface{}

	RoundData struct {
		BestStateHash    common.Hash
		ProposerOffset   int
		IsProposer       bool
		Layer            string
		ShardID          byte
		Committee        []string
		ClosestPoolState map[byte]uint64
	}
	multiSigScheme *multiSigScheme

	proposeCh   chan wire.Message
	desyncMsgCh chan wire.Message

	startTime time.Time
}

func (protocol *BFTProtocol) Start() (interface{}, error) {
	protocol.proposeCh = make(chan wire.Message)
	protocol.desyncMsgCh = make(chan wire.Message)
	protocol.phase = PBFT_LISTEN
	if protocol.RoundData.IsProposer {
		protocol.phase = PBFT_PROPOSE
	}

	Logger.log.Info("Starting PBFT protocol for " + protocol.RoundData.Layer)
	protocol.multiSigScheme = new(multiSigScheme)
	protocol.multiSigScheme.Init(protocol.EngineCfg.UserKeySet, protocol.RoundData.Committee)
	err := protocol.multiSigScheme.Prepare()
	if err != nil {
		return nil, err
	}

	//    single-node start    //
	// go protocol.CreateBlockMsg()
	// <-protocol.proposeCh
	// if protocol.pendingBlock != nil {
	// 	return protocol.pendingBlock, nil
	// }
	// return nil, errors.New("can't produce block")
	//    single-node end    //

	for {
		protocol.startTime = time.Now()
		fmt.Println("BFT: New Phase", time.Since(protocol.startTime).Seconds())
		protocol.cTimeout = make(chan struct{})
		select {
		case <-protocol.cQuit:
			return nil, errors.New("Consensus quit")
		default:
			switch protocol.phase {
			case PBFT_PROPOSE:
				if err := protocol.phasePropose(); err != nil {
					return nil, err
				}
				go common.SendMetricDataToGrafana(protocol.EngineCfg.UserKeySet.PaymentAddress.String(), float64(time.Since(protocol.startTime).Seconds()), "Propose")
			case PBFT_LISTEN:
				if err := protocol.phaseListen(); err != nil {
					return nil, err
				}
				go common.SendMetricDataToGrafana(protocol.EngineCfg.UserKeySet.PaymentAddress.String(), float64(time.Since(protocol.startTime).Seconds()), "Listen")
			case PBFT_PREPARE:
				if err := protocol.phasePrepare(); err != nil {
					return nil, err
				}
				go common.SendMetricDataToGrafana(protocol.EngineCfg.UserKeySet.PaymentAddress.String(), float64(time.Since(protocol.startTime).Seconds()), "Prepare")
			case PBFT_COMMIT:
				if err := protocol.phaseCommit(); err != nil {
					return nil, err
				}
				go common.SendMetricDataToGrafana(protocol.EngineCfg.UserKeySet.PaymentAddress.String(), float64(time.Since(protocol.startTime).Seconds()), "Commit")
				return protocol.pendingBlock, nil
			}
		}
	}
}

func (protocol *BFTProtocol) CreateBlockMsg() {
	start := time.Now()
	var msg wire.Message
	if protocol.RoundData.Layer == common.BEACON_ROLE {

		newBlock, err := protocol.EngineCfg.BlockGen.NewBlockBeacon(&protocol.EngineCfg.UserKeySet.PaymentAddress, protocol.RoundData.ProposerOffset, protocol.RoundData.ClosestPoolState)
		go common.SendMetricDataToGrafana(protocol.EngineCfg.UserKeySet.PaymentAddress.String(), float64(time.Since(start).Seconds()), "BeaconBlock")
		if err != nil {
			Logger.log.Error(err)
			protocol.closeProposeCh()
		} else {
			timeSinceLastBlk := time.Since(time.Unix(protocol.EngineCfg.BlockChain.BestState.Beacon.BestBlock.Header.Timestamp, 0))
			if timeSinceLastBlk < common.MinBlkInterval {
				fmt.Println("BFT: Wait for ", (common.MinBlkInterval - timeSinceLastBlk).Seconds())
				time.Sleep(common.MinBlkInterval - timeSinceLastBlk)
			}

			err = protocol.EngineCfg.BlockGen.FinalizeBeaconBlock(newBlock, protocol.EngineCfg.UserKeySet)

			if err != nil {
				Logger.log.Error(err)
				protocol.closeProposeCh()
			} else {
				jsonBlock, _ := json.Marshal(newBlock)
				msg, err = MakeMsgBFTPropose(jsonBlock, protocol.RoundData.Layer, protocol.RoundData.ShardID, protocol.EngineCfg.UserKeySet)
				if err != nil {
					Logger.log.Error(err)
					protocol.closeProposeCh()
				} else {
					protocol.pendingBlock = newBlock
					protocol.multiSigScheme.dataToSig = newBlock.Header.Hash()
				}
			}
		}
	} else {

		newBlock, err := protocol.EngineCfg.BlockGen.NewBlockShard(protocol.EngineCfg.UserKeySet, protocol.RoundData.ShardID, protocol.RoundData.ProposerOffset, protocol.RoundData.ClosestPoolState)
		go common.SendMetricDataToGrafana(protocol.EngineCfg.UserKeySet.PaymentAddress.String(), float64(time.Since(start).Seconds()), "ShardBlock")
		if err != nil {
			Logger.log.Error(err)
			protocol.closeProposeCh()
		} else {
			timeSinceLastBlk := time.Since(time.Unix(protocol.EngineCfg.BlockChain.BestState.Shard[protocol.RoundData.ShardID].BestBlock.Header.Timestamp, 0))
			if timeSinceLastBlk < common.MinBlkInterval {
				fmt.Println("BFT: Wait for ", (common.MinBlkInterval - timeSinceLastBlk).Seconds())
				time.Sleep(common.MinBlkInterval - timeSinceLastBlk)
			}

			err = protocol.EngineCfg.BlockGen.FinalizeShardBlock(newBlock, protocol.EngineCfg.UserKeySet)

			if err != nil {
				Logger.log.Error(err)
				protocol.closeProposeCh()
			} else {
				jsonBlock, _ := json.Marshal(newBlock)
				msg, err = MakeMsgBFTPropose(jsonBlock, protocol.RoundData.Layer, protocol.RoundData.ShardID, protocol.EngineCfg.UserKeySet)
				if err != nil {
					Logger.log.Error(err)
					protocol.closeProposeCh()
				} else {
					protocol.pendingBlock = newBlock
					protocol.multiSigScheme.dataToSig = newBlock.Header.Hash()
				}
			}
		}
	}
	elasped := time.Since(start)
	Logger.log.Critical("BFT: Block create time is", elasped)
	select {
	case <-protocol.proposeCh:
		Logger.log.Critical("Oops block create time longer than timeout")
	default:
		protocol.proposeCh <- msg
	}
	return
}

func (protocol *BFTProtocol) forwardMsg(msg wire.Message) {
	if protocol.RoundData.Layer == common.BEACON_ROLE {
		go protocol.EngineCfg.Server.PushMessageToBeacon(msg)
	} else {
		go protocol.EngineCfg.Server.PushMessageToShard(msg, protocol.RoundData.ShardID)
	}
}

func (protocol *BFTProtocol) closeTimeoutCh() {
	select {
	case <-protocol.cTimeout:
		return
	default:
		close(protocol.cTimeout)
	}
}

func (protocol *BFTProtocol) closeProposeCh() {
	select {
	case <-protocol.proposeCh:
		return
	default:
		close(protocol.proposeCh)
	}
}

func (protocol *BFTProtocol) RoundDesyncDetector() {

}
