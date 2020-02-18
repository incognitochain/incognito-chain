package blsbft

import (
	"errors"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/wire"
	"time"
)

type Engine struct {
	BFTProcess           map[int]ConsensusInterface //chainID -> consensus
	userMiningPublicKeys map[string]*incognitokey.CommitteePublicKey
	consensusName        string
	currentMiningProcess ConsensusInterface
	config               *EngineConfig
	status               int //0 > stop, 1: running
}

func (engine *Engine) GetUserLayer() (string, int) {
	panic("implement me")
}

func (engine *Engine) IsOngoing(chainName string) bool {
	panic("implement me")
}

func (engine *Engine) CommitteeChange(chainName string) {
	panic("implement me")
}

func (s *Engine) GetMiningPublicKeys() incognitokey.CommitteePublicKey {
	return *s.userMiningPublicKeys[s.consensusName]
}

func (s *Engine) GetUserRole() (string, string, int) {
	return "", "", 0
}

func (s *Engine) WatchCommitteeChange() {
	defer func() {
		time.AfterFunc(time.Second, s.WatchCommitteeChange)
	}()

	if s.status == 0 {
		return
	}

	role, chainID := s.config.Node.GetUserMiningState()
	for _, BFTProcess := range s.BFTProcess {
		if role == "" || chainID != BFTProcess.GetChainID() {
			BFTProcess.Stop()
		}
	}

	var miningProcess ConsensusInterface = nil
	if role == "committee" {
		chainName := "beacon"
		if chainID >= 0 {
			chainName = fmt.Sprintf("shard-%d", chainID)
		}
		if _, ok := s.BFTProcess[chainID]; !ok {
			if s.config.Blockchain.Chains[chainName] == nil {
				panic("Chain " + chainName + " not available")
			}
			s.BFTProcess[chainID] = NewInstance(s.config.Blockchain.Chains[chainName], chainName, chainID, s.config.Node, Logger.log)
			if err := s.BFTProcess[chainID].LoadUserKey(s.config.Node.GetMiningKeys()); err != nil {
				Logger.log.Error("Cannot load mining keys")
			}
		}
		if err := s.BFTProcess[chainID].Start(); err != nil {
			Logger.log.Error("Cannot start BFT Process")
		}
		miningProcess = s.BFTProcess[chainID]
	}

	s.currentMiningProcess = miningProcess
}

func NewConsensusEngine(config *EngineConfig) *Engine {
	engine := &Engine{
		BFTProcess:    make(map[int]ConsensusInterface),
		consensusName: common.BlsConsensus,
		config:        config,
	}
	go engine.WatchCommitteeChange()
	return engine
}

func (engine *Engine) OnBFTMsg(msg *wire.MessageBFT) {
	if engine.currentMiningProcess.GetChainKey() == msg.ChainKey {
		engine.currentMiningProcess.ProcessBFTMsg(msg)
	}
}

func (engine *Engine) Start() error {
	if engine.config.Node.GetPrivateKey() != "" {
		keyList, err := engine.GenMiningKeyFromPrivateKey(engine.config.Node.GetPrivateKey())
		if err != nil {
			panic(err)
		}
		err = engine.LoadMiningKeys(keyList)
		if err != nil {
			panic(err)
		}
	} else if engine.config.Node.GetMiningKeys() != "" {
		err := engine.LoadMiningKeys(engine.config.Node.GetMiningKeys())
		if err != nil {
			panic(err)
		}
	}
	engine.status = 1
	return nil
}

func (engine *Engine) Stop() error {
	for _, BFTProcess := range engine.BFTProcess {
		BFTProcess.Stop()
		engine.currentMiningProcess = nil
	}
	engine.status = 0
	return nil
}

func (engine *Engine) ExtractBridgeValidationData(block common.BlockInterface) ([][]byte, []int, error) {
	if engine.currentMiningProcess != nil {
		return engine.currentMiningProcess.ExtractBridgeValidationData(block)
	}
	return nil, nil, NewConsensusError(ConsensusTypeNotExistError, errors.New(block.GetConsensusType()))
}
