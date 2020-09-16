package consensus_multi

import (
	"fmt"
	signatureschemes2 "github.com/incognitochain/incognito-chain/consensus_multi/signatureschemes"
	"strings"
	"time"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/consensus_multi/blsbft"
	blsbft2 "github.com/incognitochain/incognito-chain/consensus_multi/blsbftv2"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/wire"
)

type MiningState struct {
	role    string
	layer   string
	chainID int
}

type Validator struct {
	miningKey   signatureschemes2.MiningKey
	privateSeed string
	state       MiningState
}

type Engine struct {
	BFTProcess map[int]ConsensusInterface //chainID -> consensus
	validators []*Validator               //list of validator
	version    map[int]int                //chainID -> version

	consensusName string
	config        *EngineConfig
	IsEnabled     int //0 > stop, 1: running

	//old code -> single process
	//TODO: remove old code
	userMiningPublicKeys *incognitokey.CommitteePublicKey
	userKeyListString    string
	currentMiningProcess ConsensusInterface
}

func (s *Engine) GetUserRole() (string, string, int) {
	for _, validator := range s.validators {
		return validator.state.layer, validator.state.role, validator.state.chainID
	}
	return "", "", -2
}

func (s *Engine) WatchCommitteeChange() {

	defer func() {
		time.AfterFunc(time.Second*3, s.WatchCommitteeChange)
	}()

	//check if enable
	if s.IsEnabled == 0 || s.config == nil {
		return
	}

	ValidatorGroup := make(map[int][]Validator)
	for _, validator := range s.validators {
		s.userMiningPublicKeys = validator.miningKey.GetPublicKey()
		s.userKeyListString = validator.privateSeed
		role, chainID := s.config.Node.GetPubkeyMiningState(validator.miningKey.GetPublicKey())
		if chainID == -1 {
			validator.state = MiningState{role, "beacon", -1}
		} else if chainID > -1 {
			validator.state = MiningState{role, "shard", chainID}
		} else {
			if role != "" {
				validator.state = MiningState{role, "shard", -2}
			} else {
				validator.state = MiningState{role, "", -2}
			}
		}

		//group all validator as committee by chainID
		if role == common.CommitteeRole {
			ValidatorGroup[chainID] = append(ValidatorGroup[chainID], *validator)
		}
	}

	miningProc := ConsensusInterface(nil)
	for chainID, validators := range ValidatorGroup {
		chainName := "beacon"
		if chainID >= 0 {
			chainName = fmt.Sprintf("shard-%d", chainID)
		}
		s.updateVersion(chainID)
		if _, ok := s.BFTProcess[chainID]; !ok {
			s.initProcess(chainID, chainName)
		} else { //if not run correct version => stop and init
			if s.version[chainID] == 1 {
				if _, ok := s.BFTProcess[chainID].(*blsbft.BLSBFT); !ok {
					s.BFTProcess[chainID].Stop()
					s.initProcess(chainID, chainName)
				}
			}
			if s.version[chainID] == 2 {
				if _, ok := s.BFTProcess[chainID].(*blsbft2.BLSBFT_V2); !ok {
					s.BFTProcess[chainID].Stop()
					s.initProcess(chainID, chainName)
				}
			}
		}
		validatorMiningKey := []signatureschemes2.MiningKey{}
		for _, validator := range validators {
			validatorMiningKey = append(validatorMiningKey, validator.miningKey)
		}
		s.BFTProcess[chainID].LoadUserKeys(validatorMiningKey)
		s.BFTProcess[chainID].Start()
		miningProc = s.BFTProcess[chainID]
	}
	s.currentMiningProcess = miningProc
}

func NewConsensusEngine() *Engine {
	Logger.Log.Infof("CONSENSUS: NewConsensusEngine")
	engine := &Engine{
		BFTProcess:    make(map[int]ConsensusInterface),
		consensusName: common.BlsConsensus,
		version:       make(map[int]int),
	}
	return engine
}

func (engine *Engine) initProcess(chainID int, chainName string) {
	if engine.version[chainID] == 1 {
		if chainID == -1 {
			engine.BFTProcess[chainID] = blsbft.NewInstance(engine.config.Blockchain.BeaconChain, chainName, chainID, engine.config.Node, Logger.Log)
		} else {
			engine.BFTProcess[chainID] = blsbft.NewInstance(engine.config.Blockchain.ShardChain[chainID], chainName, chainID, engine.config.Node, Logger.Log)
		}
	} else {
		if chainID == -1 {
			engine.BFTProcess[chainID] = blsbft2.NewInstance(engine.config.Blockchain.BeaconChain, chainName, chainID, engine.config.Node, Logger.Log)
		} else {
			engine.BFTProcess[chainID] = blsbft2.NewInstance(engine.config.Blockchain.ShardChain[chainID], chainName, chainID, engine.config.Node, Logger.Log)
		}
	}
}

func (engine *Engine) updateVersion(chainID int) {
	chainEpoch := uint64(1)
	if chainID == -1 {
		chainEpoch = engine.config.Blockchain.BeaconChain.GetEpoch()
	} else {
		chainEpoch = engine.config.Blockchain.ShardChain[chainID].GetEpoch()
	}

	if chainEpoch >= engine.config.Blockchain.GetConfig().ChainParams.ConsensusV2Epoch {
		engine.version[chainID] = 2
	} else {
		engine.version[chainID] = 1
	}
}

func (engine *Engine) Init(config *EngineConfig) {
	engine.config = config
	go engine.WatchCommitteeChange()
}

func (engine *Engine) Start() error {
	defer Logger.Log.Infof("CONSENSUS: Start")
	if engine.config.Node.GetPrivateKey() != "" {
		privateSeed, err := engine.GenMiningKeyFromPrivateKey(engine.config.Node.GetPrivateKey())
		if err != nil {
			panic(err)
		}
		miningKey, err := GetMiningKeyFromPrivateSeed(privateSeed)
		if err != nil {
			panic(err)
		}

		engine.validators = []*Validator{&Validator{privateSeed: privateSeed, miningKey: *miningKey}}
	} else if engine.config.Node.GetMiningKeys() != "" {
		//import validator keys : 'key1,key2'
		keys := strings.Split(engine.config.Node.GetMiningKeys(), ",")
		for _, key := range keys {
			miningKey, err := GetMiningKeyFromPrivateSeed(key)
			if err != nil {
				panic(err)
			}
			engine.validators = []*Validator{&Validator{privateSeed: key, miningKey: *miningKey}}
		}
	}
	engine.IsEnabled = 1
	return nil
}

func (engine *Engine) Stop() error {
	Logger.Log.Infof("CONSENSUS: Stop")
	for _, BFTProcess := range engine.BFTProcess {
		BFTProcess.Stop()
	}
	engine.IsEnabled = 0
	return nil
}

func (engine *Engine) OnBFTMsg(msg *wire.MessageBFT) {
	for _, process := range engine.BFTProcess {
		if process.IsStarted() && process.GetChainKey() == msg.ChainKey {
			process.ProcessBFTMsg(msg)
		}
	}
}
