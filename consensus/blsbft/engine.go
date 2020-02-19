package blsbft

import (
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
	IsEnabled            int //0 > stop, 1: running

	curringMiningState struct {
		layer   string
		role    string
		chainID int
	}
}

func (engine *Engine) GetUserLayer() (string, int) {
	return engine.curringMiningState.layer, engine.curringMiningState.chainID
}

func (s *Engine) GetUserRole() (string, string, int) {
	return s.curringMiningState.role, s.curringMiningState.layer, s.curringMiningState.chainID
}

func (engine *Engine) IsOngoing(chainName string) bool {
	if engine.currentMiningProcess == nil {
		return false
	}
	return engine.currentMiningProcess.IsOngoing()
}

//TODO: remove all places use this function
func (engine *Engine) CommitteeChange(chainName string) {
	return
}

func (s *Engine) GetMiningPublicKeys() *incognitokey.CommitteePublicKey {
	if s.userMiningPublicKeys[s.consensusName] == nil {
		return nil
	}
	return s.userMiningPublicKeys[s.consensusName]
}

func (s *Engine) WatchCommitteeChange() {
	defer func() {
		time.AfterFunc(time.Second, s.WatchCommitteeChange)
	}()

	//extract role, layer, chainID
	role, chainID := s.config.Node.GetUserMiningState()
	s.curringMiningState.chainID = chainID
	s.curringMiningState.role = role

	if chainID == -2 {
		s.curringMiningState.role = ""
		s.curringMiningState.layer = ""
	} else if chainID == -1 {
		s.curringMiningState.layer = "beacon"
	} else if chainID >= 0 {
		s.curringMiningState.layer = "shard"
	} else {
		panic("User Mining State Error")
	}

	//check if enable
	if s.IsEnabled == 0 {
		return
	}

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
	engine.IsEnabled = 1
	return nil
}

func (engine *Engine) Stop() error {
	for _, BFTProcess := range engine.BFTProcess {
		BFTProcess.Stop()
		engine.currentMiningProcess = nil
	}
	engine.IsEnabled = 0
	return nil
}

func (engine *Engine) OnBFTMsg(msg *wire.MessageBFT) {
	if engine.currentMiningProcess.GetChainKey() == msg.ChainKey {
		engine.currentMiningProcess.ProcessBFTMsg(msg)
	}
}
