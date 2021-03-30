package consensus_v2

import (
	"fmt"
	"strings"
	"time"

	"github.com/incognitochain/incognito-chain/metrics/monitor"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/consensus"
	"github.com/incognitochain/incognito-chain/consensus_v2/blsbft"
	blsbft2 "github.com/incognitochain/incognito-chain/consensus_v2/blsbftv2"
	blsbft3 "github.com/incognitochain/incognito-chain/consensus_v2/blsbftv3"
	blsbft4 "github.com/incognitochain/incognito-chain/consensus_v2/blsbftv4"
	signatureschemes2 "github.com/incognitochain/incognito-chain/consensus_v2/signatureschemes"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/wire"
)

type Engine struct {
	BFTProcess             map[int]blsbft.Actor          // chainID -> consensus
	validators             []*consensus.Validator        // list of validator
	syncingValidators      map[int][]consensus.Validator // syncing validators
	syncingValidatorsIndex map[string]int                // syncing validators index
	version                map[int]int                   // chainID -> version

	consensusName string
	config        *EngineConfig
	IsEnabled     int //0 > stop, 1: running

	//legacy code -> single process
	userMiningPublicKeys *incognitokey.CommitteePublicKey
	userKeyListString    string
	currentMiningProcess blsbft.Actor
}

//just get role of first validator
//this function support NODE monitor (getmininginfo) which assumed running only 1 validator
func (s *Engine) GetUserRole() (string, string, int) {
	for _, validator := range s.validators {
		return validator.State.Layer, validator.State.Role, validator.State.ChainID
	}
	return "", "", -2
}

func (s *Engine) GetCurrentValidators() []*consensus.Validator {
	return s.validators
}

func (s *Engine) SyncingValidatorsByShardID(shardID int) []string {
	res := []string{}
	for _, validator := range s.syncingValidators[shardID] {
		res = append(res, validator.MiningKey.GetPublicKey().GetMiningKeyBase58(common.BlsConsensus))
	}
	return res
}

func (s *Engine) GetOneValidator() *consensus.Validator {
	if len(s.validators) > 0 {
		return s.validators[0]
	}
	return nil
}

func (s *Engine) GetOneValidatorForEachConsensusProcess() map[int]*consensus.Validator {
	chainValidator := make(map[int]*consensus.Validator)
	role := ""
	layer := ""
	chainID := -2
	pubkey := ""
	if len(s.validators) > 0 {
		for _, validator := range s.validators {
			if validator.State.ChainID != -2 {
				_, ok := chainValidator[validator.State.ChainID]
				if ok {
					if validator.State.Role == common.CommitteeRole {
						chainValidator[validator.State.ChainID] = validator
						pubkey = validator.MiningKey.GetPublicKeyBase58()
						role = validator.State.Role
						chainID = validator.State.ChainID
						layer = validator.State.Layer
					}
				} else {
					chainValidator[validator.State.ChainID] = validator
					pubkey = validator.MiningKey.GetPublicKeyBase58()
					role = validator.State.Role
					chainID = validator.State.ChainID
					layer = validator.State.Layer
				}
			} else {
				if role == "" { //role not set, and userkey in waiting role
					role = validator.State.Role
					layer = validator.State.Layer
				}
			}
		}
	}
	monitor.SetGlobalParam("Role", role)
	monitor.SetGlobalParam("Layer", layer)
	monitor.SetGlobalParam("ShardID", chainID)
	monitor.SetGlobalParam("MINING_PUBKEY", pubkey)
	//fmt.Println("GetOneValidatorForEachConsensusProcess", chainValidator[1])
	return chainValidator
}

func (engine *Engine) WatchCommitteeChange() {

	defer func() {
		time.AfterFunc(time.Second*3, engine.WatchCommitteeChange)
	}()

	//check if enable
	if engine.IsEnabled == 0 || engine.config == nil {
		return
	}

	ValidatorGroup := make(map[int][]consensus.Validator)
	for _, validator := range engine.validators {
		engine.userMiningPublicKeys = validator.MiningKey.GetPublicKey()
		engine.userKeyListString = validator.PrivateSeed
		role, chainID := engine.config.Node.GetPubkeyMiningState(validator.MiningKey.GetPublicKey())
		//Logger.Log.Info("validator key", validator.MiningKey.GetPublicKeyBase58())
		if chainID == -1 {
			validator.State = consensus.MiningState{role, "beacon", -1}
		} else if chainID > -1 {
			validator.State = consensus.MiningState{role, "shard", chainID}
			if role == common.PendingRole {
				if len(engine.syncingValidators[chainID]) != 0 {
					engine.syncingValidators[chainID] = append(
						engine.syncingValidators[chainID][:engine.syncingValidatorsIndex[validator.MiningKey.GetPublicKeyBase58()]],
						engine.syncingValidators[chainID][engine.syncingValidatorsIndex[validator.MiningKey.GetPublicKeyBase58()]+1:]...)
				}
			}
			if role == common.SyncingRole {
				if _, ok := engine.syncingValidatorsIndex[validator.MiningKey.GetPublicKeyBase58()]; !ok {
					engine.syncingValidators[chainID] = append(engine.syncingValidators[chainID], *validator)
					engine.syncingValidatorsIndex[validator.MiningKey.GetPublicKeyBase58()] = len(engine.syncingValidators[chainID]) - 1
				}
			}
		} else {
			if role != "" {
				validator.State = consensus.MiningState{role, "shard", -2}
			} else {
				validator.State = consensus.MiningState{role, "", -2}
			}
		}

		//group all validator as committee by chainID
		if role == common.CommitteeRole {
			//fmt.Println("Consensus", chainID, validator.PrivateSeed, validator.State)
			ValidatorGroup[chainID] = append(ValidatorGroup[chainID], *validator)
		}
	}

	miningProc := blsbft.Processor(nil)
	for chainID, validators := range ValidatorGroup {
		chainName := "beacon"
		if chainID >= 0 {
			chainName = fmt.Sprintf("shard-%d", chainID)
		}
		engine.updateVersion(chainID)
		if _, ok := engine.BFTProcess[chainID]; !ok {
			engine.initProcess(chainID, chainName)
		} else { //if not run correct version => stop and init
			if engine.version[chainID] == 1 {
				if _, ok := engine.BFTProcess[chainID].(*blsbft.BLSBFT); !ok {
					engine.BFTProcess[chainID].Destroy()
					engine.initProcess(chainID, chainName)
				}
			}
			if engine.version[chainID] == 2 {
				if _, ok := engine.BFTProcess[chainID].(*blsbft2.BLSBFT_V2); !ok {
					engine.BFTProcess[chainID].Destroy()
					engine.initProcess(chainID, chainName)
				}
			}
			if engine.version[chainID] == 3 {
				if _, ok := engine.BFTProcess[chainID].(*blsbft3.BLSBFT_V3); !ok {
					engine.BFTProcess[chainID].Destroy()
					engine.initProcess(chainID, chainName)
				}
			}
			if engine.version[chainID] == 4 {
				if _, ok := engine.BFTProcess[chainID].(*blsbft4.BLSBFT_V4); !ok {
					engine.BFTProcess[chainID].Destroy()
					engine.initProcess(chainID, chainName)
				}
			}
		}
		validatorMiningKey := []signatureschemes2.MiningKey{}
		for _, validator := range validators {
			validatorMiningKey = append(validatorMiningKey, validator.MiningKey)
		}

		engine.BFTProcess[chainID].LoadUserKeys(validatorMiningKey)
		engine.BFTProcess[chainID].Start()
		miningProc = engine.BFTProcess[chainID]
	}

	for chainID, proc := range engine.BFTProcess {
		if _, ok := ValidatorGroup[chainID]; !ok {
			proc.Stop()
		}
	}

	engine.currentMiningProcess = miningProc
}

func NewConsensusEngine() *Engine {
	Logger.Log.Infof("CONSENSUS: NewConsensusEngine")
	engine := &Engine{
		BFTProcess:             make(map[int]blsbft.Actor),
		syncingValidators:      make(map[int][]consensus.Validator),
		syncingValidatorsIndex: make(map[string]int),
		consensusName:          common.BlsConsensus,
		version:                make(map[int]int),
	}
	return engine
}

func (engine *Engine) initProcess(chainID int, chainName string) {
	var bftActor blsbft.Actor
	if chainID == -1 {
		bftActor = blsbft.NewActorWithValue(
			engine.config.Blockchain.BeaconChain, engine.version[chainID],
			chainID, chainName, engine.config.Node, Logger.Log)
	} else {
		bftActor = blsbft.NewActorWithValue(
			engine.config.Blockchain.ShardChain[chainID], engine.version[chainID],
			chainID, chainName, engine.config.Node, Logger.Log)
	}
	engine.BFTProcess[chainID] = bftActor
}

func (engine *Engine) updateVersion(chainID int) {
	chainEpoch := uint64(1)
	chainHeight := uint64(1)
	if chainID == -1 {
		chainEpoch = engine.config.Blockchain.BeaconChain.GetEpoch()
		chainHeight = engine.config.Blockchain.BeaconChain.GetBestViewHeight()
	} else {
		chainEpoch = engine.config.Blockchain.ShardChain[chainID].GetEpoch()
		chainHeight = engine.config.Blockchain.ShardChain[chainID].GetBestView().GetBeaconHeight()
	}

	if chainEpoch >= engine.config.Blockchain.GetConfig().ChainParams.ConsensusV2Epoch {
		engine.version[chainID] = blsbft.MultiViewsVersion
	}

	if chainHeight >= engine.config.Blockchain.GetConfig().ChainParams.ConsensusV3Height {
		engine.version[chainID] = blsbft.SlashingVersion
	}

	if chainHeight >= engine.config.Blockchain.GetConfig().ChainParams.ConsensusV4Height {
		engine.version[chainID] = blsbft.MultiSubsetsVersion
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
		engine.validators = []*consensus.Validator{&consensus.Validator{PrivateSeed: privateSeed, MiningKey: *miningKey}}
	} else if engine.config.Node.GetMiningKeys() != "" {
		keys := strings.Split(engine.config.Node.GetMiningKeys(), ",")
		engine.validators = []*consensus.Validator{}
		for _, key := range keys {
			miningKey, err := GetMiningKeyFromPrivateSeed(key)
			if err != nil {
				panic(err)
			}
			engine.validators = append(engine.validators, &consensus.Validator{PrivateSeed: key, MiningKey: *miningKey})
		}
		engine.validators = engine.validators[:1] //allow only 1 key
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

func (engine *Engine) GetAllValidatorKeyState() map[string]consensus.MiningState {
	result := make(map[string]consensus.MiningState)
	for _, validator := range engine.validators {
		result[validator.PrivateSeed] = validator.State
	}
	return result
}

func (engine *Engine) IsCommitteeInShard(shardID byte) bool {
	if shard, ok := engine.BFTProcess[int(shardID)]; ok {
		return shard.IsStarted()
	}
	return false
}
