package blockchain

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/incognitokey"
)

var DeafaultFeatureStat *FeatureStat

type FeatureStat struct {
	blockchain *BlockChain
	nodes      map[string][]string // committeePK : feature lists
}

type FeatureReportInfo struct {
	ValidatorStat map[string]map[int]uint64 // feature -> shardid -> stat
	ProposeStat   map[string]map[int]uint64 // feature -> shardid -> stat
}

func (bc *BlockChain) InitFeatureStat() {
	DeafaultFeatureStat = &FeatureStat{
		blockchain: bc,
		nodes:      make(map[string][]string),
	}
}

func (stat *FeatureStat) Report() FeatureReportInfo {
	validatorStat := make(map[string]map[int]uint64)
	proposeStat := make(map[string]map[int]uint64)

	beaconCommittee, err := incognitokey.ExtractPublickeysFromCommitteeKeyList(stat.blockchain.BeaconChain.GetCommittee(), common.BlsConsensus)
	if err != nil {
		Logger.log.Error(err)
	}

	shardCommmittee := map[int][]string{}
	for i := 0; i < stat.blockchain.GetActiveShardNumber(); i++ {
		shardCommmittee[i], err = incognitokey.ExtractPublickeysFromCommitteeKeyList(stat.blockchain.ShardChain[i].GetCommittee(), common.BlsConsensus)
		if err != nil {
			Logger.log.Error(err)
		}
	}

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
			validatorStat[feature][chainCommitteeID]++
			//if key is proposer
			if chainCommitteeID == -1 && chainCommiteeIndex[-1] < 7 {
				proposeStat[feature][-1]++
			}
			if chainCommitteeID > -1 && chainCommiteeIndex[chainCommitteeID] < config.Param().CommitteeSize.NumberOfFixedShardBlockValidator {
				proposeStat[feature][chainCommitteeID]++
			}

		}
	}

	return FeatureReportInfo{
		validatorStat,
		proposeStat,
	}

}

func (featureStat *FeatureStat) addNode(key string, features []string) {
	featureStat.nodes[key] = features
}
