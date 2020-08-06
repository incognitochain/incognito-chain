package committeestate

import (
	"math/rand"

	"github.com/incognitochain/incognito-chain/incognitokey"
)

func (b *BeaconCommitteeEngine) AssignSubstitutePoolUsingRandomInstruction(
	seed int64,
) ([]string, map[byte][]string) {
	_ = b.GetCandidateBeaconWaitingForCurrentRandom()
	return []string{}, map[byte][]string{}
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
) []string {
	res, _ := incognitokey.CommitteeKeyListToString(b.beaconCommitteeStateV1.nextEpochBeaconCandidate)
	return res
}
