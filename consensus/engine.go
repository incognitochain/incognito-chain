package consensus

import (
	"fmt"
	"sync"
	"time"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/consensus/blsbft"
	blsbft2 "github.com/incognitochain/incognito-chain/consensus/blsbftv2"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/pubsub"
	"github.com/incognitochain/incognito-chain/wire"
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

	version int

	lock *sync.Mutex
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

func (s *Engine) WatchCommitteeChange() {

	defer func() {
		time.AfterFunc(time.Second*3, s.WatchCommitteeChange)
	}()

	//check if enable
	if s.IsEnabled == 0 || s.config == nil {
		fmt.Println("CONSENSUS: enable", s.IsEnabled, s.config == nil)
		return
	}

	//extract role, layer, chainID
	role, chainID := s.config.Node.GetUserMiningState()
	Logger.Log.Infof("Node state role: %v chainID: %v", role, chainID)
	s.curringMiningState.chainID = chainID
	s.curringMiningState.role = role

	if chainID == -2 {
		s.curringMiningState.role = ""
		s.curringMiningState.layer = ""
		s.NotifyBeaconRole(false)
		s.NotifyShardRole(-2)
	} else if chainID == -1 {
		s.curringMiningState.layer = "beacon"
		s.NotifyBeaconRole(true)
		s.NotifyShardRole(-1)
	} else if chainID >= 0 {
		s.curringMiningState.layer = "shard"
		s.NotifyBeaconRole(false)
		s.NotifyShardRole(chainID)
	} else {
		panic("User Mining State Error")
	}

	for _, BFTProcess := range s.BFTProcess {
		if role != "committee" || chainID != BFTProcess.GetChainID() {
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
			if len(s.config.Blockchain.ShardChain)-1 < chainID {
				panic("Chain " + chainName + " not available")
			}
			if s.version == 1 {
				if chainID == -1 {
					s.BFTProcess[chainID] = blsbft.NewInstance(s.config.Blockchain.BeaconChain, chainName, chainID, s.config.Node, Logger.Log)
				} else {
					s.BFTProcess[chainID] = blsbft.NewInstance(s.config.Blockchain.ShardChain[chainID], chainName, chainID, s.config.Node, Logger.Log)
				}
			} else {
				if chainID == -1 {
					s.BFTProcess[chainID] = blsbft2.NewInstance(s.config.Blockchain.BeaconChain, chainName, chainID, s.config.Node, Logger.Log)
				} else {
					s.BFTProcess[chainID] = blsbft2.NewInstance(s.config.Blockchain.ShardChain[chainID], chainName, chainID, s.config.Node, Logger.Log)
				}
			}

		}

		s.BFTProcess[chainID].Start()
		miningProcess = s.BFTProcess[chainID]
		s.currentMiningProcess = s.BFTProcess[chainID]
		if err := s.LoadMiningKeys(s.userKeyListString); err != nil {
			panic(err)
		}

	}
	s.currentMiningProcess = miningProcess
}

func NewConsensusEngine() *Engine {
	Logger.Log.Infof("CONSENSUS: NewConsensusEngine")
	engine := &Engine{
		BFTProcess:           make(map[int]ConsensusInterface),
		consensusName:        common.BlsConsensus,
		userMiningPublicKeys: make(map[string]*incognitokey.CommitteePublicKey),
		version:              2,
		lock:                 new(sync.Mutex),
	}
	return engine
}

func (engine *Engine) Init(config *EngineConfig) {
	engine.config = config
	go engine.WatchCommitteeChange()
}

func (engine *Engine) Start() error {
	defer Logger.Log.Infof("CONSENSUS: Start", engine.userKeyListString)
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
	Logger.Log.Infof("CONSENSUS: Stop")
	for _, BFTProcess := range engine.BFTProcess {
		BFTProcess.Stop()
		engine.currentMiningProcess = nil
	}
	engine.IsEnabled = 0
	return nil
}

func (engine *Engine) OnBFTMsg(msg *wire.MessageBFT) {
	if engine.currentMiningProcess == nil {
		Logger.Log.Warnf("Current mining process still nil")
		return
	}
	if engine.currentMiningProcess.GetChainKey() == msg.ChainKey {
		engine.currentMiningProcess.ProcessBFTMsg(msg)
	}
}

func (engine *Engine) NotifyBeaconRole(beaconRole bool) {
	engine.config.PubSubManager.PublishMessage(pubsub.NewMessage(pubsub.BeaconRoleTopic, beaconRole))
}
func (engine *Engine) NotifyShardRole(shardRole int) {
	engine.config.PubSubManager.PublishMessage(pubsub.NewMessage(pubsub.ShardRoleTopic, shardRole))
}
