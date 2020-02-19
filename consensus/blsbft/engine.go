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
	userKeyListString    string
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
	return s.curringMiningState.layer, s.curringMiningState.role, s.curringMiningState.chainID
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
	if s.userMiningPublicKeys == nil || s.userMiningPublicKeys[s.consensusName] == nil {
		return nil
	}
	return s.userMiningPublicKeys[s.consensusName]
}

func (s *Engine) WatchCommitteeChange() {

	defer func() {
		time.AfterFunc(time.Second, s.WatchCommitteeChange)
	}()

	//check if enable
	if s.IsEnabled == 0 || s.config == nil {
		fmt.Println("CONSENSUS: enable", s.IsEnabled, s.config == nil)
		return
	}

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

		}
		s.BFTProcess[chainID].Start()
		miningProcess = s.BFTProcess[chainID]
		s.currentMiningProcess = s.BFTProcess[chainID]
		err := s.LoadMiningKeys(s.userKeyListString)
		if err != nil {
			panic(err)
		}
	}
	//fmt.Println("CONSENSUS:", role, chainID)
	s.currentMiningProcess = miningProcess
}

func NewConsensusEngine() *Engine {
	fmt.Println("CONSENSUS: NewConsensusEngine")
	engine := &Engine{
		BFTProcess:           make(map[int]ConsensusInterface),
		consensusName:        common.BlsConsensus,
		userMiningPublicKeys: make(map[string]*incognitokey.CommitteePublicKey),
	}
	return engine
}

func (engine *Engine) Init(config *EngineConfig) {
	engine.config = config
	go engine.WatchCommitteeChange()
}

func (engine *Engine) Start() error {
	if engine.config.Node.GetPrivateKey() != "" {
		keyList, err := engine.GenMiningKeyFromPrivateKey(engine.config.Node.GetPrivateKey())
		if err != nil {
			panic(err)
		}
		engine.userKeyListString = keyList
	} else if engine.config.Node.GetMiningKeys() != "" {
		engine.userKeyListString = engine.config.Node.GetMiningKeys()
	}
	err := engine.LoadMiningKeys(engine.userKeyListString)
	if err != nil {
		panic(err)
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
