package blockchain

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/consensus"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/wire"
	"sync"
	"time"
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
}

type FeatureReportInfo struct {
	ValidatorStat map[string]map[int]uint64 // feature -> shardid -> stat
	ProposeStat   map[string]map[int]uint64 // feature -> shardid -> stat
	ValidatorSize map[int]int               // chainid -> all validator size
}

func CreateNewFeatureStatMessage(beaconView *BeaconBestState, validators []*consensus.Validator, unTriggerFeatures []string) (*wire.MessageFeature, error) {

	if len(validators) == 0 {
		return nil, nil
	}

	if len(unTriggerFeatures) == 0 {
		return nil, nil
	}

	validatorFromUserKeys, syncValidator := beaconView.ExtractPendingAndCommittee(validators)
	featureSyncValidators := []string{}
	featureSyncSignatures := [][]byte{}

	signBytes := []byte{}
	for _, v := range unTriggerFeatures {
		signBytes = append([]byte(wire.CmdMsgFeatureStat), []byte(v)...)
	}
	timestamp := time.Now().Unix()
	timestampStr := fmt.Sprintf("%v", timestamp)
	signBytes = append(signBytes, []byte(timestampStr)...)

	for i, v := range validatorFromUserKeys {
		dataSign := signBytes[:]
		signature, err := v.MiningKey.BriSignData(append(dataSign, []byte(syncValidator[i])...))
		if err != err {
			continue
		}
		featureSyncSignatures = append(featureSyncSignatures, signature)
		featureSyncValidators = append(featureSyncValidators, syncValidator[i])
	}
	if len(featureSyncValidators) == 0 {
		return nil, nil
	}
	Logger.log.Infof("Send Feature Stat Message, key %+v \n signature %+v", featureSyncValidators, featureSyncSignatures)
	msg := wire.NewMessageFeature(int(timestamp), featureSyncValidators, featureSyncSignatures, unTriggerFeatures)

	return msg, nil
}

func (bc *BlockChain) InitFeatureStat() {
	DefaultFeatureStat = &FeatureStat{
		blockchain: bc,
		nodes:      make(map[string]NodeFeatureInfo),
		lock:       new(sync.RWMutex),
	}
	//send message periodically
	go func() {
		for {
			time.Sleep(5 * time.Second)

			//get untrigger feature
			beaconView := bc.BeaconChain.GetBestView().(*BeaconBestState)
			unTriggerFeatures := beaconView.getUntriggerFeature()
			msg, err := CreateNewFeatureStatMessage(beaconView, bc.config.ConsensusEngine.GetValidators(), unTriggerFeatures)

			if err != nil {
				Logger.log.Error(err)
				continue
			}

			if msg == nil {
				continue
			}

			if err := bc.config.Server.PushMessageToBeacon(msg, nil); err != nil {
				Logger.log.Errorf("Send Feature Stat Message Public Message to beacon, error %+v", err)
			}
			//DefaultFeatureStat.Report()
		}

	}()

}

func (stat *FeatureStat) IsContainLatestFeature(curView *BeaconBestState, cpk string) bool {
	nodeFeatures := stat.nodes[cpk].Features
	//get feature that beacon is preparing to trigger
	unTriggerFeatures := curView.getUntriggerFeature()

	//check if node contain the untriggered feature
	for _, feature := range unTriggerFeatures {
		if common.IndexOfStr(feature, nodeFeatures) == -1 {
			//fmt.Println("node", cpk, "not content feature", feature, nodeFeatures, len(nodeFeatures))
			return false
		}
	}
	return true
}

func (stat *FeatureStat) Report(beaconView *BeaconBestState) FeatureReportInfo {
	validatorStat := make(map[string]map[int]uint64)
	proposeStat := make(map[string]map[int]uint64)
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
	unTriggerFeatures := beaconView.getUntriggerFeature()
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
			if proposeStat[feature] == nil {
				proposeStat[feature] = make(map[int]uint64)
			}
			//check in beacon
			if common.IndexOfStr(key, beaconCommittee) > -1 {
				validatorStat[feature][-1]++
				proposeStat[feature][-1]++
			}

			//check in shard
			for i := 0; i < stat.blockchain.GetActiveShardNumber(); i++ {
				//if in pending -> increase validator set
				if common.IndexOfStr(key, pendingCommmittee[i]) > -1 {
					validatorStat[feature][-1]++
				}
				//if in committee, increase validator set
				if common.IndexOfStr(key, shardCommmittee[i]) > -1 {
					validatorStat[feature][i]++
					//if in proposer, increase proposer
					if common.IndexOfStr(key, shardCommmittee[i]) < config.Param().CommitteeSize.NumberOfFixedShardBlockValidator {
						proposeStat[feature][i]++
					}
				}
			}
		}
	}

	//Logger.log.Infof("=========== \n%+v", validatorStat)
	return FeatureReportInfo{
		validatorStat,
		proposeStat,
		validatorSize,
	}

}

func (featureStat *FeatureStat) addNode(timestamp int, key string, features []string) {
	featureStat.lock.RLock()
	defer featureStat.lock.RUnlock()

	//not update from old message
	if _, ok := featureStat.nodes[key]; ok && featureStat.nodes[key].Timestamp > timestamp {
		panic(1)
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
