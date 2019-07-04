package mubft

import (
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/metrics"
	"time"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/wire"
)

type BFTProtocol struct {
	cBFTMsg   chan wire.Message
	EngineCfg *EngineConfig

	cQuit    chan interface{}
	cTimeout chan interface{}

	phase string

	pendingBlock    interface{}
	blockCreateTime time.Duration
	RoundData       struct {
		MinBeaconHeight uint64
		BestStateHash   common.Hash
		IsProposer      bool
		Layer           string
		ShardID         byte
		Committee       []string
		// ClosestPoolState map[byte]uint64
		Round int
	}
	multiSigScheme *multiSigScheme

	proposeCh  chan wire.Message
	earlyMsgCh chan wire.Message

	startTime time.Time
}

func (protocol *BFTProtocol) Start() (interface{}, error) {
	protocol.cQuit = make(chan interface{})
	protocol.proposeCh = make(chan wire.Message)
	protocol.earlyMsgCh = make(chan wire.Message)
	protocol.phase = BFT_LISTEN
	defer close(protocol.cQuit)
	if protocol.RoundData.IsProposer {
		protocol.phase = BFT_PROPOSE
	}

	Logger.log.Info("Starting PBFT protocol for " + protocol.RoundData.Layer)
	protocol.multiSigScheme = new(multiSigScheme)
	protocol.multiSigScheme.Init(protocol.EngineCfg.UserKeySet, protocol.RoundData.Committee)
	err := protocol.multiSigScheme.Prepare()
	if err != nil {
		return nil, err
	}
	go protocol.earlyMsgHandler()
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
		protocol.cTimeout = make(chan interface{})
		switch protocol.phase {
		case BFT_PROPOSE:
			if err := protocol.phasePropose(); err != nil {
				return nil, err
			}
		case BFT_LISTEN:
			if err := protocol.phaseListen(); err != nil {
				return nil, err
			}
		case BFT_AGREE:
			if err := protocol.phaseAgree(); err != nil {
				return nil, err
			}
		case BFT_COMMIT:
			if err := protocol.phaseCommit(); err != nil {
				return nil, err
			}
			return protocol.pendingBlock, nil
		}
	}
}

func (protocol *BFTProtocol) CreateBlockMsg() {
	start := time.Now()
	var elasped time.Duration
	var msg wire.Message
	//fmt.Println("[db] CreateBlockMsg")
	if protocol.RoundData.Layer == common.BEACON_ROLE {

		newBlock, err := protocol.EngineCfg.BlockGen.NewBlockBeacon(&protocol.EngineCfg.UserKeySet.PaymentAddress, protocol.RoundData.Round, protocol.EngineCfg.BlockChain.Synker.GetClosestShardToBeaconPoolState())
		go metrics.AnalyzeTimeSeriesMetricData(map[string]interface{}{
			metrics.Measurement: metrics.BeaconBlock,
			metrics.MeasurementValue: float64(time.Since(start).Seconds()),
			metrics.Tag: metrics.NodeIDTag,
			metrics.TagValue: protocol.EngineCfg.UserKeySet.PaymentAddress.String(),
		})
		elasped = time.Since(start)
		if err != nil {
			Logger.log.Error(err)
			protocol.closeProposeCh()
		} else {
			timeSinceLastBlk := time.Since(time.Unix(protocol.EngineCfg.BlockChain.BestState.Beacon.BestBlock.Header.Timestamp, 0))
			if timeSinceLastBlk < common.MinBeaconBlkInterval {
				fmt.Println("BFT: Wait for ", (common.MinBeaconBlkInterval - timeSinceLastBlk).Seconds())
				time.Sleep(common.MinBeaconBlkInterval - timeSinceLastBlk)
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

		newBlock, err := protocol.EngineCfg.BlockGen.NewBlockShard(protocol.EngineCfg.UserKeySet, protocol.RoundData.ShardID, protocol.RoundData.Round, protocol.EngineCfg.BlockChain.Synker.GetClosestCrossShardPoolState(), protocol.RoundData.MinBeaconHeight, start)
		go metrics.AnalyzeTimeSeriesMetricData(map[string]interface{}{
			metrics.Measurement: metrics.ShardBlock,
			metrics.MeasurementValue: float64(time.Since(start).Seconds()),
			metrics.Tag: metrics.NodeIDTag,
			metrics.TagValue: protocol.EngineCfg.UserKeySet.PaymentAddress.String(),
		})
		elasped = time.Since(start)
		if err != nil {
			Logger.log.Error(err)
			protocol.closeProposeCh()
		} else {
			timeSinceLastBlk := time.Since(time.Unix(protocol.EngineCfg.BlockChain.BestState.Shard[protocol.RoundData.ShardID].BestBlock.Header.Timestamp, 0))
			if timeSinceLastBlk < common.MinShardBlkInterval {
				fmt.Println("BFT: Wait for ", (common.MinShardBlkInterval - timeSinceLastBlk).Seconds())
				time.Sleep(common.MinShardBlkInterval - timeSinceLastBlk)
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
	Logger.log.Critical("BFT: Block create time is", elasped)
	select {
	case <-protocol.proposeCh:
		Logger.log.Critical("☠︎ Oops block create time longer than timeout")
	default:
		protocol.blockCreateTime = elasped
		protocol.proposeCh <- msg
	}
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
	case d := <-protocol.cTimeout:
		if d != nil {
			close(protocol.cTimeout)
		}
		return
	default:
		close(protocol.cTimeout)
	}
}

func (protocol *BFTProtocol) closeProposeCh() {
	select {
	case d := <-protocol.proposeCh:
		if d != nil {
			close(protocol.proposeCh)
		}
		return
	default:
		close(protocol.proposeCh)
	}
}

func (protocol *BFTProtocol) earlyMsgHandler() {
	var agreeMsgs []wire.Message
	var commitMsgs []wire.Message
	go func() {
		for {
			select {
			case <-protocol.cQuit:
				return
			default:
				if protocol.phase == BFT_AGREE {
					for _, msg := range agreeMsgs {
						protocol.cBFTMsg <- msg
					}
					agreeMsgs = []wire.Message{}
				}
				if protocol.phase == BFT_COMMIT {
					for _, msg := range commitMsgs {
						protocol.cBFTMsg <- msg
					}
					commitMsgs = []wire.Message{}
				}
				time.Sleep(10 * time.Millisecond)
			}
		}
	}()

	for {
		select {
		case <-protocol.cQuit:
			return
		case earlyMsg := <-protocol.earlyMsgCh:
			fmt.Println("Early msg received")
			switch earlyMsg.MessageType() {
			case wire.CmdBFTAgree:
				if protocol.phase == BFT_LISTEN {
					if common.IndexOfStr(earlyMsg.(*wire.MessageBFTAgree).Pubkey, protocol.RoundData.Committee) >= 0 {
						agreeMsgs = append(agreeMsgs, earlyMsg)
					}
				}
			case wire.CmdBFTCommit:
				if protocol.phase == BFT_AGREE {
					newSig := bftCommittedSig{
						ValidatorsIdxR: earlyMsg.(*wire.MessageBFTCommit).ValidatorsIdx,
						Sig:            earlyMsg.(*wire.MessageBFTCommit).CommitSig,
					}
					R := earlyMsg.(*wire.MessageBFTCommit).R
					err := protocol.multiSigScheme.VerifyCommitSig(earlyMsg.(*wire.MessageBFTCommit).Pubkey, newSig.Sig, R, newSig.ValidatorsIdxR)
					if err == nil {
						commitMsgs = append(commitMsgs, earlyMsg)
					}
				}
			}
		}
	}
}

func getTimeout(phase string, committeeSize int) time.Duration {
	assumedDelay := time.Duration(committeeSize) * MaxNetworkDelayTime
	switch phase {
	case BFT_PROPOSE:
		return assumedDelay + ListenTimeout
	case BFT_LISTEN:
		return assumedDelay + ListenTimeout
	case BFT_AGREE:
		return assumedDelay + AgreeTimeout
	case BFT_COMMIT:
		return assumedDelay + CommitTimeout
	}
	return 0
}
