package consensus

import (
	"fmt"
	"time"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/consensus/blsbft"
	"github.com/incognitochain/incognito-chain/consensus/blsbftv2"
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
		time.AfterFunc(time.Second*3, s.WatchCommitteeChange)
	}()

	//check if enable
	if s.IsEnabled == 0 || s.config == nil {
		fmt.Println("CONSENSUS: enable", s.IsEnabled, s.config == nil)
		return
	}
	newConsensus := func(chain ChainInterface, chainKey string, chainID int, node NodeInterface, logger common.Logger, version int) ConsensusInterface {
		if version == 1 {
			return blsbft.NewInstance(chain, chainKey, chainID, node, logger)
		} else {
			return blsbftv2.NewInstance(chain, chainKey, chainID, node, logger)
		}
	}
	getChain := func(e *Engine, chainID int) (ChainInterface, string) {
		if chainID < -1 {
			return nil, ""
		}
		if chainID == -1 {
			return e.config.Blockchain.BeaconChain, common.BeaconRole
		}
		return s.config.Blockchain.ShardChain[chainID], common.ShardRole
	}

	//extract role, layer, chainID
	role, chainID := s.config.Node.GetUserMiningState()
	chain, layer := getChain(s, chainID)

	if chainID < -2 {
		panic("User Mining State Error")
	}
	for _, BFTProcess := range s.BFTProcess {
		if role == "" || chainID != BFTProcess.GetChainID() {
			BFTProcess.Stop()
		}
	}

	chainName := s.config.Node.GetNodeMode()
	if role == "committee" {
		chainName = common.BeaconRole
		if chainID >= 0 {
			chainName = fmt.Sprintf("%v-%d", common.ShardRole, chainID)
		}
		if _, ok := s.BFTProcess[chainID]; !ok {
			if len(s.config.Blockchain.ShardChain)-1 < chainID {
				panic("Chain " + chainName + " not available")
			}
		}
	}

	s.curringMiningState.chainID = chainID
	s.curringMiningState.role = role
	s.curringMiningState.layer = layer
	s.NotifyBeaconRole(chainID == -1)
	s.NotifyShardRole(chainID)
	s.BFTProcess[chainID] = newConsensus(chain, chainName, chainID, s.config.Node, Logger.Log, s.version)
	s.currentMiningProcess = s.BFTProcess[chainID]
	err := s.LoadMiningKeys(s.userKeyListString)
	if err != nil {
		panic(err)
	}
	if err := s.currentMiningProcess.Start(); err != nil {
		return
	}
}

func NewConsensusEngine() *Engine {
	fmt.Println("CONSENSUS: NewConsensusEngine")
	engine := &Engine{
		BFTProcess:           make(map[int]ConsensusInterface),
		consensusName:        common.BlsConsensus,
		userMiningPublicKeys: make(map[string]*incognitokey.CommitteePublicKey),
		version:              2,
	}
	return engine
}

func (engine *Engine) Init(config *EngineConfig) {

	engine.config = config
	if engine.config.Node.GetPrivateKey() != "" {
		keyList, err := engine.GenMiningKeyFromPrivateKey(engine.config.Node.GetPrivateKey())
		if err != nil {
			panic(err)
		}
		engine.userKeyListString = keyList
	} else if engine.config.Node.GetMiningKeys() != "" {
		engine.userKeyListString = engine.config.Node.GetMiningKeys()
	}
}

func (engine *Engine) Start() error {
	fmt.Println("CONSENSUS: Start")
	engine.IsEnabled = 1
	go engine.WatchCommitteeChange()
	return nil
}

func (engine *Engine) Stop() error {
	fmt.Println("CONSENSUS: Stop")
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

func (engine *Engine) NotifyBeaconRole(beaconRole bool) {
	engine.config.PubSubManager.PublishMessage(pubsub.NewMessage(pubsub.BeaconRoleTopic, beaconRole))
}
func (engine *Engine) NotifyShardRole(shardRole int) {
	engine.config.PubSubManager.PublishMessage(pubsub.NewMessage(pubsub.ShardRoleTopic, shardRole))
}
