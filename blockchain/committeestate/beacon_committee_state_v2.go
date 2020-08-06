package committeestate

import (
	"math/rand"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
)

func (b *BeaconCommitteeEngine) AssignSubstitutePoolUsingRandomInstruction(
	blkHash common.Hash,
	seed int64,
) ([]string, map[byte][]string) {
	//Still update
	subsSizeMap := map[byte]int{}
	numShards := len(b.beaconCommitteeStateV1.shardCommittee)
	for i := 0; i < numShards; i++ {
		shardCandidates, err := b.SubstituteGetter(blkHash, i)
		if err != nil {
			//TODO handle error?
		}
		subsSizeMap[byte(i)] = len(shardCandidates)
	}
	shardCandidates, err := b.CandidateGetter(blkHash, false)
	if err != nil {
		//TODO handle error?
	}
	beaconCandidates, err := b.CandidateGetter(blkHash, true)
	if err != nil {
		//TODO handle error?
	}
	shCandListString, _ := incognitokey.CommitteeKeyListToString(shardCandidates)
	bcCandListString, _ := incognitokey.CommitteeKeyListToString(beaconCandidates)
	newShSubs := b.AssignShardsPoolUsingRandomInstruction(
		seed,
		numShards,
		shCandListString,
		subsSizeMap,
	)
	newBcSubs := b.AssignBeaconUsingRandomInstruction(
		seed,
		bcCandListString,
	)
	return newBcSubs, newShSubs
}

func (b *BeaconCommitteeEngine) AssignShardsPoolUsingRandomInstruction(
	seed int64,
	numShards int,
	candidatesList []string,
	subsSizeMap map[byte]int,
) map[byte][]string {
	//Still update
	sumArr := make([]uint64, numShards)
	totalCandidates := 0
	for i := 0; i < numShards; i++ {
		totalCandidates += subsSizeMap[byte(i)]
	}
	rand.Seed(seed)
	res := map[byte][]string{}
	for _, c := range candidatesList {
		sumArr[0] = uint64(totalCandidates / subsSizeMap[byte(0)])
		for i := 1; i < numShards; i++ {
			sumArr[i] = sumArr[i-1] + uint64(totalCandidates/subsSizeMap[byte(i)])
		}
		pos := 1 + rand.Intn(int(sumArr[numShards-1]))
		sID := 0
		for sID = 0; pos > int(sumArr[sID]); sID++ {
		}
		res[byte(sID)] = append(res[byte(sID)], c)
		subsSizeMap[byte(sID)] = subsSizeMap[byte(sID)] + 1
		totalCandidates++
	}
	return res
}

func (b *BeaconCommitteeEngine) AssignBeaconUsingRandomInstruction(
	seed int64,
	candidateList []string,
) []string {
	//TODO
	return candidateList
}
