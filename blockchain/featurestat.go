package blockchain

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/wire"
	"time"
)

var DefaultFeatureStat *FeatureStat

type FeatureStat struct {
	blockchain *BlockChain
	nodes      map[string][]string // committeePK : feature lists
}

type FeatureReportInfo struct {
	ValidatorStat map[string]map[int]uint64 // feature -> shardid -> stat
	ProposeStat   map[string]map[int]uint64 // feature -> shardid -> stat
	ValidatorSize map[int]int               // chainid -> all validator size
}

func (bc *BlockChain) InitFeatureStat() {
	DefaultFeatureStat = &FeatureStat{
		blockchain: bc,
		nodes:      make(map[string][]string),
	}
	fmt.Println("debugfeature InitFeatureStat")
	//send message periodically
	go func() {
		for {
			time.Sleep(5 * time.Second)

			//get untrigger feature
			beaconView := bc.BeaconChain.GetBestView().(*BeaconBestState)
			unTriggerFeatures := beaconView.getUntriggerFeature()

			validatorFromUserKeys, syncValidator := beaconView.ExtractPendingAndCommittee(bc.config.ConsensusEngine.GetSyncingValidators())
			featureSyncValidators := []string{}
			featureSyncSignatures := [][]byte{}
			for i, v := range validatorFromUserKeys {
				signature, err := v.MiningKey.BriSignData([]byte(wire.CmdMsgFeatureStat))
				if err != nil {
					continue
				}
				featureSyncSignatures = append(featureSyncSignatures, signature)
				featureSyncValidators = append(featureSyncValidators, syncValidator[i])
			}
			if len(featureSyncValidators) == 0 {
				continue
			}
			Logger.log.Infof("Send Feature Stat Message, key %+v \n signature %+v", featureSyncValidators, featureSyncSignatures)
			msg := wire.NewMessageFeature(featureSyncValidators, featureSyncSignatures, unTriggerFeatures)
			if err := bc.config.Server.PushMessageToBeacon(msg, nil); err != nil {
				Logger.log.Errorf("Send Feature Stat Message Public Message to beacon, error %+v", err)
			}
			DefaultFeatureStat.Report()
		}

	}()

}

func (stat *FeatureStat) IsContainLatestFeature(curView *BeaconBestState, cpk string) bool {
	nodeFeatures := stat.nodes[cpk]
	//get feature that beacon is preparing to trigger
	unTriggerFeatures := curView.getUntriggerFeature()

	//check if node contain the untriggered feature
	for _, feature := range unTriggerFeatures {
		if common.IndexOfStr(feature, nodeFeatures) == -1 {
			return false
		}
	}
	return true
}

func (stat *FeatureStat) Report() FeatureReportInfo {
	validatorStat := make(map[string]map[int]uint64)
	proposeStat := make(map[string]map[int]uint64)
	validatorSize := make(map[int]int)

	beaconCommittee, err := incognitokey.CommitteeKeyListToString(stat.blockchain.BeaconChain.GetCommittee())
	if err != nil {
		Logger.log.Error(err)
	}
	validatorSize[-1] = len(beaconCommittee)
	shardCommmittee := map[int][]string{}
	for i := 0; i < stat.blockchain.GetActiveShardNumber(); i++ {
		shardCommmittee[i], err = incognitokey.CommitteeKeyListToString(stat.blockchain.ShardChain[i].GetCommittee())
		validatorSize[i] = len(shardCommmittee[i])
		if err != nil {
			Logger.log.Error(err)
		}
	}
	fmt.Println("debugfeature", len(stat.nodes))
	for key, features := range stat.nodes {
		//if key is in Committee list
		chainCommitteeID := -2
		chainCommiteeIndex := map[int]int{}
		if common.IndexOfStr(key, beaconCommittee) > -1 {
			chainCommitteeID = -1
			chainCommiteeIndex[-1] = common.IndexOfStr(key, beaconCommittee)
		}
		for i := 0; i < stat.blockchain.GetActiveShardNumber(); i++ {
			if common.IndexOfStr(key, shardCommmittee[i]) > -1 {
				chainCommitteeID = i
				chainCommiteeIndex[-1] = common.IndexOfStr(key, shardCommmittee[i])
			}
		}

		if chainCommitteeID == -2 {
			continue
		}

		//check valid trigger feature and remove duplicate
		featureList := map[string]bool{}
		for _, feature := range features {
			if _, ok := config.Param().AutoEnableFeature[feature]; ok {
				featureList[feature] = true
			}
		}

		//count
		for feature, _ := range featureList {
			if validatorStat[feature] == nil {
				validatorStat[feature] = make(map[int]uint64)
			}
			validatorStat[feature][chainCommitteeID]++
			//if key is proposer
			if chainCommitteeID == -1 && chainCommiteeIndex[-1] < 7 {
				if proposeStat[feature] == nil {
					proposeStat[feature] = make(map[int]uint64)
				}
				proposeStat[feature][-1]++
			}
			if chainCommitteeID > -1 && chainCommiteeIndex[chainCommitteeID] < config.Param().CommitteeSize.NumberOfFixedShardBlockValidator {
				if proposeStat[feature] == nil {
					proposeStat[feature] = make(map[int]uint64)
				}
				proposeStat[feature][chainCommitteeID]++
			}

		}
	}

	Logger.log.Infof("=========== \n%+v", validatorStat)
	return FeatureReportInfo{
		validatorStat,
		proposeStat,
		validatorSize,
	}

}

func (featureStat *FeatureStat) addNode(key string, features []string) {
	fmt.Println("debugfeature addnode", key)
	featureStat.nodes[key] = features
}

func (featureStat *FeatureStat) containExpectedFeature(key string, expectedFeature []string) bool {
	for _, f := range expectedFeature {
		if common.IndexOfStr(f, featureStat.nodes[key]) == -1 {
			return false
		}
	}
	return true

}
