package blockchain

import (
	"fmt"
	"sync"
	"time"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/consensus"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/wire"
)

var DefaultFeatureStat *FeatureStat

type NodeFeatureInfo struct {
	Features  []string
	Timestamp int
}
type FeatureStat struct {
	blockchain *BlockChain
	nodes      map[string]NodeFeatureInfo // committeePK : feature lists
	lock       *sync.RWMutex
	msg        chan *wire.MessageFeature
}

type FeatureReportInfo struct {
	ValidatorStat map[string]map[int]uint64 // feature -> shardid -> stat
	CommitteeStat map[string]map[int]uint64 // feature -> shardid -> stat
	ValidatorSize map[int]int               // chainid -> all validator size
}

//filter validators that is in sync pool ( inSyncPool = true), or in pending&committee (isInPool = false), then create feature stat message
func CreateNewFeatureStatMessage(beaconView *BeaconBestState, inSyncPool bool, reportFeatures []string, validators []*consensus.Validator) (*wire.MessageFeature, error) {

	if len(reportFeatures) == 0 {
		return nil, nil
	}

	validatorFromUserKeys, validatorStr := beaconView.ExtractPendingAndCommittee(validators)
	if inSyncPool {
		validatorFromUserKeys, validatorStr = beaconView.ExtractAllFinishSyncingValidators(validators)
	}
	if len(validatorFromUserKeys) == 0 {
		return nil, nil
	}
	featureSyncValidators := []string{}
	featureSyncSignatures := [][]byte{}

	signBytes := []byte{}
	for _, v := range reportFeatures {
		signBytes = append([]byte(wire.CmdMsgFeatureStat), []byte(v)...)
	}
	timestamp := time.Now().Unix()
	timestampStr := fmt.Sprintf("%v", timestamp)
	signBytes = append(signBytes, []byte(timestampStr)...)

	for i, v := range validatorFromUserKeys {
		dataSign := signBytes[:]
		signature, err := v.MiningKey.BriSignData(append(dataSign, []byte(validatorStr[i])...))
		if err != err {
			continue
		}
		featureSyncSignatures = append(featureSyncSignatures, signature)
		featureSyncValidators = append(featureSyncValidators, validatorStr[i])
	}
	if len(featureSyncValidators) == 0 {
		return nil, nil
	}
	Logger.log.Infof("Send Feature Stat Message, key %+v \n signature %+v", featureSyncValidators, featureSyncSignatures)
	msg := wire.NewMessageFeature(int(timestamp), featureSyncValidators, featureSyncSignatures, reportFeatures)

	return msg, nil
}

func (bc *BlockChain) InitFeatureStat() {
	DefaultFeatureStat = &FeatureStat{
		blockchain: bc,
		nodes:      make(map[string]NodeFeatureInfo),
		lock:       new(sync.RWMutex),
		msg:        make(chan *wire.MessageFeature, 5000),
	}

	go func() {
		for {
			select {
			case msg := <-DefaultFeatureStat.msg:
				bc.ReceiveFeatureReport(msg.Timestamp, msg.CommitteePublicKey, msg.Signature, msg.Feature)
			}
		}
	}()

	//send message periodically
	go func() {
		for {
			bc.SendFeatureStat()
			time.Sleep(10 * time.Second)
		}
	}()
}

func (bc *BlockChain) SendFeatureStat() {
	if bc.BeaconChain.GetMultiView() == nil || bc.BeaconChain.GetBestView() == nil {
		return
	}
	beaconView := bc.BeaconChain.GetBestView().(*BeaconBestState)

	//if validator in sync pool, send feature stat for all untrigger and triggger feature, even before checkpoint
	unTriggerFeatures := beaconView.getUntriggerFeature(false)
	updatedFeature := unTriggerFeatures
	for feature, _ := range beaconView.TriggeredFeature {
		updatedFeature = append(updatedFeature, feature)
	}
	msg, err := CreateNewFeatureStatMessage(beaconView, true, updatedFeature, bc.config.ConsensusEngine.GetValidators())
	if err != nil {
		Logger.log.Error(err)
		return
	}
	if msg != nil {
		if err := bc.config.Server.PushMessageToBeacon(msg, nil); err != nil {
			Logger.log.Errorf("Send Feature Stat Message Public Message to beacon, error %+v", err)
		}
	}

	//only after checkpoint, send feature state for  validator not in sync pool
	unTriggerFeatures = beaconView.getUntriggerFeature(true)
	msg, err = CreateNewFeatureStatMessage(beaconView, false, unTriggerFeatures, bc.config.ConsensusEngine.GetValidators())
	if err != nil {
		Logger.log.Error(err)
		return
	}
	if msg != nil {
		if err := bc.config.Server.PushMessageToBeacon(msg, nil); err != nil {
			Logger.log.Errorf("Send Feature Stat Message Public Message to beacon, error %+v", err)
		}
	}

	//DefaultFeatureStat.Report()
	DefaultFeatureStat.lock.Lock()
	for id, node := range DefaultFeatureStat.nodes {
		if time.Now().Unix()-int64(node.Timestamp) > 30*60 {
			delete(DefaultFeatureStat.nodes, id)
		}
	}
	DefaultFeatureStat.lock.Unlock()
}

func (stat *FeatureStat) ReceiveMsg(msg *wire.MessageFeature) {
	if len(stat.msg) >= cap(stat.msg)*90/100 {
		Logger.log.Warn("Drop feature stat message", msg.CommitteePublicKey, len(stat.msg), cap(stat.msg))
		return
	}
	Logger.log.Info("Receive a MsgFeature", msg.CommitteePublicKey, msg.Feature)
	stat.msg <- msg
}

//contain all trigger & untrigger feature
func (stat *FeatureStat) IsContainLatestFeature(curView *BeaconBestState, cpk string) bool {
	stat.lock.RLock()
	defer stat.lock.RUnlock()

	nodeFeatures := stat.nodes[cpk].Features
	//get feature that beacon is checking for trigger
	unTriggerFeatures := curView.getUntriggerFeature(false)
	fmt.Println("IsContainLatestFeature", cpk, unTriggerFeatures, stat.nodes[cpk].Features)
	//check if node contain the untriggered feature
	for _, feature := range unTriggerFeatures {
		if common.IndexOfStr(feature, nodeFeatures) == -1 {
			//fmt.Println("node", cpk, "not content feature", feature, nodeFeatures, len(nodeFeatures))
			return false
		}
	}

	//check if node contain the triggered feature
	for feature, _ := range curView.TriggeredFeature {
		if common.IndexOfStr(feature, nodeFeatures) == -1 {
			//fmt.Println("node", cpk, "not content feature", feature, nodeFeatures, len(nodeFeatures))
			return false
		}
	}
	return true
}

func (stat *FeatureStat) Report(beaconView *BeaconBestState) FeatureReportInfo {
	validatorStat := make(map[string]map[int]uint64)
	committeeStat := make(map[string]map[int]uint64)
	validatorSize := make(map[int]int)

	beaconCommittee, err := incognitokey.CommitteeKeyListToString(stat.blockchain.BeaconChain.GetCommittee())
	if err != nil {
		Logger.log.Error(err)
	}
	validatorSize[-1] = len(beaconCommittee)
	shardCommmittee := map[int][]string{}
	pendingCommmittee := map[int][]string{}
	for i := 0; i < stat.blockchain.GetActiveShardNumber(); i++ {
		shardCommmittee[i], err = incognitokey.CommitteeKeyListToString(beaconView.GetAShardCommittee(byte(i)))
		pendingCommmittee[i], err = incognitokey.CommitteeKeyListToString(beaconView.GetAShardPendingValidator(byte(i)))
		validatorSize[i] = len(shardCommmittee[i]) + len(pendingCommmittee[i])
		if err != nil {
			Logger.log.Error(err)
		}
	}
	unTriggerFeatures := beaconView.getUntriggerFeature(true)
	stat.lock.Lock()
	defer stat.lock.Unlock()
	for key, features := range stat.nodes {

		//check valid trigger feature and remove duplicate
		featureList := map[string]bool{}
		for _, feature := range features.Features {
			if _, ok := config.Param().AutoEnableFeature[feature]; ok {
				if common.IndexOfStr(feature, unTriggerFeatures) > -1 {
					featureList[feature] = true
				}
			}
		}

		//count
		for feature, _ := range featureList {
			if validatorStat[feature] == nil {
				validatorStat[feature] = make(map[int]uint64)
			}
			if committeeStat[feature] == nil {
				committeeStat[feature] = make(map[int]uint64)
			}

			//check in beacon
			if common.IndexOfStr(key, beaconCommittee) > -1 {
				validatorStat[feature][-1]++
				committeeStat[feature][-1]++
			}

			//check in shard
			for i := 0; i < stat.blockchain.GetActiveShardNumber(); i++ {
				//if in pending -> increase validator set
				if common.IndexOfStr(key, pendingCommmittee[i]) > -1 {
					validatorStat[feature][i]++
				}
				//if in committee, increase validator set
				if common.IndexOfStr(key, shardCommmittee[i]) > -1 {
					validatorStat[feature][i]++
					committeeStat[feature][i]++
				}
			}
		}
	}

	//Logger.log.Infof("=========== \n%+v", validatorStat)
	return FeatureReportInfo{
		validatorStat,
		committeeStat,
		validatorSize,
	}

}

func (featureStat *FeatureStat) addNode(timestamp int, key string, features []string) {
	featureStat.lock.RLock()
	defer featureStat.lock.RUnlock()

	//not update from old message
	if _, ok := featureStat.nodes[key]; ok && featureStat.nodes[key].Timestamp > timestamp {
		return
	}

	featureStat.nodes[key] = NodeFeatureInfo{
		features, timestamp,
	}

}

func (featureStat *FeatureStat) containExpectedFeature(key string, expectedFeature []string) bool {
	for _, f := range expectedFeature {
		if common.IndexOfStr(f, featureStat.nodes[key].Features) == -1 {
			return false
		}
	}
	return true

}
