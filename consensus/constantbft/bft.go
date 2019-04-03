package constantbft

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/pkg/errors"

	"github.com/constant-money/constant-chain/common"

	"github.com/constant-money/constant-chain/cashec"

	"github.com/constant-money/constant-chain/blockchain"
	"github.com/constant-money/constant-chain/wire"
)

type BFTProtocol struct {
	cBFTMsg   chan wire.Message
	EngineCfg *EngineConfig

	ShardToBeaconPool blockchain.ShardToBeaconPool
	CrossShardPool    map[byte]blockchain.CrossShardPool
	BlockGen          *blockchain.BlkTmplGenerator
	BlockChain        *blockchain.BlockChain
	Server            serverInterface
	UserKeySet        *cashec.KeySet

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

	proposeCh chan wire.Message

	startTime time.Time
}

func (protocol *BFTProtocol) Start() (interface{}, error) {
	protocol.proposeCh = make(chan wire.Message)
	protocol.phase = PBFT_LISTEN
	if protocol.RoundData.IsProposer {
		protocol.phase = PBFT_PROPOSE
	}

	Logger.log.Info("Starting PBFT protocol for " + protocol.RoundData.Layer)
	protocol.multiSigScheme = new(multiSigScheme)
	protocol.multiSigScheme.Init(protocol.UserKeySet, protocol.RoundData.Committee)
	err := protocol.multiSigScheme.Prepare()
	if err != nil {
		return nil, err
	}

	//    single-node start    //
	go protocol.CreateBlockMsg()
	<-protocol.proposeCh
	if protocol.pendingBlock != nil {
		return protocol.pendingBlock, nil
	}
	return nil, errors.New("can't produce block")
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
			case PBFT_LISTEN:
				if err := protocol.phaseListen(); err != nil {
					return nil, err
				}
			case PBFT_PREPARE:
				if err := protocol.phasePrepare(); err != nil {
					return nil, err
				}
			case PBFT_COMMIT:
				if err := protocol.phaseCommit(); err != nil {
					return nil, err
				}
				return protocol.pendingBlock, nil
			}
		}
	}
}

func (protocol *BFTProtocol) CreateBlockMsg() {
	start := time.Now()
	var msg wire.Message
	if protocol.RoundData.Layer == common.BEACON_ROLE {
		newBlock, err := protocol.BlockGen.NewBlockBeacon(&protocol.UserKeySet.PaymentAddress, &protocol.UserKeySet.PrivateKey, protocol.RoundData.ProposerOffset, protocol.RoundData.ClosestPoolState)
		<-time.Tick(time.Second * 2) //single-node
		if err != nil {
			Logger.log.Error(err)
			protocol.closeProposeCh()
		} else {
			jsonBlock, _ := json.Marshal(newBlock)
			msg, err = MakeMsgBFTPropose(jsonBlock, protocol.RoundData.Layer, protocol.RoundData.ShardID, protocol.UserKeySet)
			if err != nil {
				Logger.log.Error(err)
				protocol.closeProposeCh()
			} else {
				protocol.pendingBlock = newBlock
				protocol.multiSigScheme.dataToSig = newBlock.Header.Hash()
			}
		}
	} else {
		newBlock, err := protocol.BlockGen.NewBlockShard(&protocol.UserKeySet.PaymentAddress, &protocol.UserKeySet.PrivateKey, protocol.RoundData.ShardID, protocol.RoundData.ProposerOffset, protocol.RoundData.ClosestPoolState)
		<-time.Tick(time.Second * 2) //single-node
		if err != nil {
			Logger.log.Error(err)
			protocol.closeProposeCh()
		} else {
			jsonBlock, _ := json.Marshal(newBlock)
			msg, err = MakeMsgBFTPropose(jsonBlock, protocol.RoundData.Layer, protocol.RoundData.ShardID, protocol.UserKeySet)
			if err != nil {
				Logger.log.Error(err)
				protocol.closeProposeCh()
			} else {
				protocol.pendingBlock = newBlock
				protocol.multiSigScheme.dataToSig = newBlock.Header.Hash()
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
		go protocol.Server.PushMessageToBeacon(msg)
	} else {
		go protocol.Server.PushMessageToShard(msg, protocol.RoundData.ShardID)
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
