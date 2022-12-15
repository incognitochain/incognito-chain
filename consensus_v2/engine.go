package consensus_v2

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/incognitochain/incognito-chain/blockchain"

	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common/base58"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/consensus"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/consensus_v2/blsbft"
	signatureschemes2 "github.com/incognitochain/incognito-chain/consensus_v2/signatureschemes"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/metrics/monitor"
	"github.com/incognitochain/incognito-chain/pubsub"
	"github.com/incognitochain/incognito-chain/wire"
)

type Engine struct {
	bftProcess map[int]blsbft.Actor   // chainID -> consensus
	validators []*consensus.Validator // list of validator
	version    map[int]int            // chainID -> version

	consensusName string
	config        *EngineConfig
	IsEnabled     int //0 > stop, 1: running

	//legacy code -> single process
	userMiningPublicKeys *incognitokey.CommitteePublicKey
	userKeyListString    string
	currentMiningProcess blsbft.Actor
}

// just get role of first validator
// this function support NODE monitor (getmininginfo) which assumed running only 1 validator
func (s *Engine) GetUserRole() (string, string, int) {
	for _, validator := range s.validators {
		return validator.State.Layer, validator.State.Role, validator.State.ChainID
	}
	return "", "", -2
}

func (s *Engine) GetCurrentValidators() []*consensus.Validator {
	return s.validators
}

func (s *Engine) GetValidators() []*consensus.Validator {
	res := []*consensus.Validator{}
	for _, validator := range s.validators {
		res = append(res, validator)
	}
	return res
}

func (engine *Engine) GetOneValidator() *consensus.Validator {
	if len(engine.validators) > 0 {
		return engine.validators[0]
	}
	return nil
}

func (engine *Engine) GetOneValidatorForEachConsensusProcess() map[int]*consensus.Validator {
	chainValidator := make(map[int]*consensus.Validator)
	role := ""
	layer := ""
	chainID := -2
	if len(engine.validators) > 0 {
		for _, validator := range engine.validators {
			if validator.State.InBeaconWaiting {
				chainValidator[-1] = validator
			}
			if validator.State.ChainID != -2 {
				_, ok := chainValidator[validator.State.ChainID]
				if ok {
					if validator.State.Role == common.CommitteeRole {
						chainValidator[validator.State.ChainID] = validator
						role = validator.State.Role
						chainID = validator.State.ChainID
						layer = validator.State.Layer
					}
				} else {
					chainValidator[validator.State.ChainID] = validator
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
	//Logger.Log.Infof("Validator Role %+v, Layer %+v, ChainID %+v", role, layer, chainID)
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

	validatorGroup := make(map[int][]consensus.Validator)
	for _, validator := range engine.validators {
		engine.userMiningPublicKeys = validator.MiningKey.GetPublicKey()
		engine.userKeyListString = validator.PrivateSeed
		role, chainID := engine.config.Node.GetPubkeyMiningState(validator.MiningKey.GetPublicKey())
		inBeaconWaiting := engine.config.Node.IsInBeaconWaitingList(validator.MiningKey.GetPublicKey())

		keyBytes := validator.MiningKey.PubKey[common.BlsConsensus]
		logKey := base58.Base58Check{}.Encode(keyBytes, common.Base58Version)
		log.Printf("validator key %+v, shardID %+v, role %+v \n", logKey, chainID, role)

		if chainID == common.BeaconChainID {
			validator.State = consensus.MiningState{role, common.BeaconChainKey, common.BeaconChainID, inBeaconWaiting}
		} else if chainID > common.BeaconChainID {
			validator.State = consensus.MiningState{role, common.ShardChainKey, chainID, inBeaconWaiting}
		} else {
			if role != "" {
				validator.State = consensus.MiningState{role, common.ShardChainKey, -2, inBeaconWaiting}
			} else {
				validator.State = consensus.MiningState{role, "", -2, inBeaconWaiting}
			}
		}

		//group all validator as committee by chainID
		if role == common.CommitteeRole {
			//fmt.Println("Consensus", chainID, validator.PrivateSeed, validator.State)
			validatorGroup[chainID] = append(validatorGroup[chainID], *validator)
		}
	}

	miningProc := blsbft.Actor(nil)
	for chainID, validators := range validatorGroup {
		engine.NotifyNewRole(chainID, common.CommitteeRole)
		chainName := common.BeaconChainKey
		if chainID >= 0 {
			chainName = fmt.Sprintf("%s-%d", common.ShardChainKey, chainID)

			//check chain sync up
			shardFinalizeHeight := engine.config.Blockchain.BeaconChain.GetBestView().(*blockchain.BeaconBestState).BestShardHeight[byte(chainID)]
			shardHeight := engine.config.Blockchain.ShardChain[chainID].GetBestView().GetHeight()
			if shardHeight+10 < shardFinalizeHeight {
				continue
			}
		}

		oldVersion := engine.version[chainID]
		engine.updateVersion(chainID)

		if bftActor, ok := engine.bftProcess[chainID]; !ok {
			engine.initProcess(chainID, chainName)
		} else {
			if oldVersion < types.INSTANT_FINALITY_VERSION_V2 && engine.version[chainID] >= types.INSTANT_FINALITY_VERSION_V2 {
				bftActor.Destroy()
				engine.initProcess(chainID, chainName)
			}
			engine.bftProcess[chainID].SetBlockVersion(engine.version[chainID])
		}

		validatorMiningKey := []signatureschemes2.MiningKey{}
		for _, validator := range validators {
			validatorMiningKey = append(validatorMiningKey, validator.MiningKey)
		}
		engine.bftProcess[chainID].LoadUserKeys(validatorMiningKey)
		engine.bftProcess[chainID].Start()
		miningProc = engine.bftProcess[chainID]
	}

	for chainID, proc := range engine.bftProcess {
		if _, ok := validatorGroup[chainID]; !ok {
			if proc.IsStarted() {
				proc.Stop()
				engine.NotifyNewRole(chainID, common.WaitingRole)
			}
		}
	}

	engine.currentMiningProcess = miningProc
}

func NewConsensusEngine() *Engine {
	Logger.Log.Infof("CONSENSUS: NewConsensusEngine")
	engine := &Engine{
		bftProcess:    make(map[int]blsbft.Actor),
		consensusName: common.BlsConsensus,
		version:       make(map[int]int),
	}
	blsbft.ByzantineDetectorObject = blsbft.NewByzantineDetector(Logger.Log)
	go blsbft.ByzantineDetectorObject.Loop()
	return engine
}

func (engine *Engine) initProcess(chainID int, chainName string) {
	var bftActor blsbft.Actor
	blockVersion := engine.version[chainID]
	if chainID == -1 {
		bftActor = blsbft.NewActorWithValue(
			engine.config.Blockchain.BeaconChain,
			engine.config.Blockchain.BeaconChain,
			engine.version[chainID], blockVersion,
			chainID, chainName, engine.config.Node, Logger.Log)
	} else {
		bftActor = blsbft.NewActorWithValue(
			engine.config.Blockchain.ShardChain[chainID],
			engine.config.Blockchain.BeaconChain,
			engine.version[chainID], blockVersion,
			chainID, chainName, engine.config.Node, Logger.Log)
	}
	engine.bftProcess[chainID] = bftActor
	Logger.Log.Infof("CONSENSUS: init process, chain %+v, chain-name %+v, version %+v",
		chainID, chainName, blockVersion)
}

func (engine *Engine) updateVersion(chainID int) {
	newVersion := engine.getBlockVersion(chainID)
	oldVersion := 0
	if chainID == -1 {
		oldVersion = engine.config.Blockchain.BeaconChain.GetBestView().GetBlock().GetVersion()
	} else {
		oldVersion = engine.config.Blockchain.ShardChain[chainID].GetBestView().GetBlock().GetVersion()
	}

	if newVersion < oldVersion {
		panic("Init wrong verson ")
	}
	engine.version[chainID] = newVersion
}

func (engine *Engine) Init(config *EngineConfig) {
	engine.config = config
	go engine.WatchCommitteeChange()
}

func (engine *Engine) Start() error {
	defer Logger.Log.Infof("CONSENSUS: Start")

	if engine.config.Node.GetPrivateKey() != "" {
		engine.loadKeysFromPrivateKey()
	} else if engine.config.Node.GetMiningKeys() != "" {
		engine.loadKeysFromMiningKey()
	}
	engine.IsEnabled = 1
	return nil
}

func (engine *Engine) loadKeysFromPrivateKey() {
	privateSeed, err := GenMiningKeyFromPrivateKey(engine.config.Node.GetPrivateKey())
	if err != nil {
		panic(err)
	}
	miningKey, err := GetMiningKeyFromPrivateSeed(privateSeed)
	if err != nil {
		panic(err)
	}
	engine.validators = []*consensus.Validator{
		&consensus.Validator{PrivateSeed: privateSeed, MiningKey: *miningKey},
	}
}

func (engine *Engine) loadKeysFromMiningKey() {
	keys := strings.Split(engine.config.Node.GetMiningKeys(), ",")
	engine.validators = []*consensus.Validator{}
	for _, key := range keys {
		miningKey, err := GetMiningKeyFromPrivateSeed(key)
		if err != nil {
			panic(err)
		}
		engine.validators = append(engine.validators, &consensus.Validator{
			PrivateSeed: key, MiningKey: *miningKey,
		})
	}
	// @NOTICE: hack code, only allow one key
	//engine.validators = engine.validators[:1] //allow only 1 key

	//set monitor pubkey
	pubkeys := []string{}
	for _, val := range engine.validators {
		pubkeys = append(pubkeys, val.MiningKey.GetPublicKey().GetMiningKeyBase58("bls"))
	}
	monitor.SetGlobalParam("MINING_PUBKEY", strings.Join(pubkeys, ","))
}

func (engine *Engine) Stop() error {
	Logger.Log.Infof("CONSENSUS: Stop")
	for _, BFTProcess := range engine.bftProcess {
		BFTProcess.Stop()
	}
	engine.IsEnabled = 0
	return nil
}

func (engine *Engine) OnBFTMsg(msg *wire.MessageBFT) {
	for _, process := range engine.bftProcess {
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
	if shard, ok := engine.bftProcess[int(shardID)]; ok {
		return shard.IsStarted()
	}
	return false
}

func (engine *Engine) getBlockVersion(chainID int) int {
	chainEpoch := uint64(1)
	chainHeight := uint64(1)
	triggerFeature := make(map[string]uint64)
	if chainID == -1 {
		chainEpoch = engine.config.Blockchain.BeaconChain.GetEpoch()
		chainHeight = engine.config.Blockchain.BeaconChain.GetBestViewHeight()
		triggerFeature = engine.config.Blockchain.BeaconChain.GetFinalView().(*blockchain.BeaconBestState).TriggeredFeature
	} else {
		chainEpoch = engine.config.Blockchain.ShardChain[chainID].GetEpoch()
		chainHeight = engine.config.Blockchain.ShardChain[chainID].GetBestView().GetBeaconHeight()
		triggerFeature = engine.config.Blockchain.ShardChain[chainID].GetFinalView().(*blockchain.ShardBestState).TriggeredFeature
	}

	//get last trigger feature that change block version
	latestFeature := ""
	latestTriggerHeight := uint64(0)
	for f, h := range triggerFeature {
		if _, ok := config.Param().FeatureVersion[f]; ok {
			if latestTriggerHeight < h {
				latestTriggerHeight = h
				latestFeature = f
			}
		}
	}
	if version, ok := config.Param().FeatureVersion[latestFeature]; ok {
		return int(version)
	}

	//legacy flow
	if triggerFeature[blockchain.INSTANT_FINALITY_FEATURE] != 0 {
		return types.INSTANT_FINALITY_VERSION
	}

	if chainHeight >= config.Param().ConsensusParam.BlockProducingV3Height {
		return types.BLOCK_PRODUCINGV3_VERSION
	}

	if chainHeight >= config.Param().ConsensusParam.Lemma2Height {
		return types.LEMMA2_VERSION
	}

	if chainHeight >= config.Param().ConsensusParam.StakingFlowV3Height {
		return types.SHARD_SFV3_VERSION
	}

	if chainHeight >= config.Param().ConsensusParam.StakingFlowV2Height {
		return types.SHARD_SFV2_VERSION
	}

	if chainEpoch >= config.Param().ConsensusParam.ConsensusV2Epoch {
		return types.MULTI_VIEW_VERSION
	}

	return types.BFT_VERSION
}

// BFTProcess for testing only
func (engine *Engine) BFTProcess() map[int]blsbft.Actor {
	return engine.bftProcess
}

func (engine *Engine) NotifyNewRole(newCID int, newRole string) {
	engine.config.PubSubManager.PublishMessage(
		pubsub.NewMessage(pubsub.NodeRoleDetailTopic, &pubsub.NodeRole{
			CID:  newCID,
			Role: newRole,
		}),
	)
}
