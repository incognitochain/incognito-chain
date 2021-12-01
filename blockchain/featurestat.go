package blockchain

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/consensus"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/wire"
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
	}
	fmt.Println("debugfeature InitFeatureStat")
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
			DefaultFeatureStat.Report()
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
		for _, feature := range features.Features {
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

func (featureStat *FeatureStat) addNode(timestamp int, key string, features []string) {
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
