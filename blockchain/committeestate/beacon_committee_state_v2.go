package committeestate

import (
	"bytes"
	"crypto/rand"
	"math/big"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
)

// type BeaconCommitteeEngine struct {
// 	beaconHeight                      uint64
// 	beaconHash                        common.Hash
// 	beaconCommitteeStateV1            *BeaconCommitteeStateV1
// 	uncommittedBeaconCommitteeStateV1 *BeaconCommitteeStateV1
// }

func (b *BeaconCommitteeEngine) AssignSubstitutePoolUsingRandomInstruction(
	seed int64,
) ([]string, map[byte][]string) {

	return []string{}, map[byte][]string{}
}

func (b *BeaconCommitteeEngine) AssignShardsPoolUsingRandomInstruction(
	seed int64,
	numShards int,
	candidatesList []string,
	subsSizeMap map[byte]int,
) map[byte][]string {
	// numShards := len(b.beaconCommitteeStateV1.shardSubstitute)
	sumArr := make([]uint64, numShards)
	totalCandidates := 0
	for i := 0; i < numShards; i++ {
		totalCandidates += subsSizeMap[byte(i)]
	}
	sumArr[0] = uint64(totalCandidates / subsSizeMap[byte(0)])
	for i := 1; i < numShards; i++ {
		sumArr[i] = sumArr[i-1] + uint64(totalCandidates/subsSizeMap[byte(i)])
	}
	reader := bytes.NewReader(common.Uint64ToBytes(uint64(seed)))
	res := map[byte][]string{}
	for _, c := range candidatesList {
		pos, _ := rand.Int(reader, big.NewInt(int64(sumArr[numShards-1])))
		sID := 0
		for sID = 0; uint64(pos.Int64()) > sumArr[sID]; sID++ {
		}
		res[byte(sID)] = append(res[byte(sID)], c)
		for j := sID; j < numShards; j++ {
			sumArr[j]++
		}
	}
	return res
}

func (b *BeaconCommitteeEngine) AssignBeaconUsingRandomInstruction(
	seed int64,
) []string {
	res, _ := incognitokey.CommitteeKeyListToString(b.beaconCommitteeStateV1.nextEpochBeaconCandidate)
	return res
}
