package committeestate

import (
	"math/rand"
	"sync"

	"github.com/incognitochain/incognito-chain/privacy"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
)

type BeaconCommitteeStateV2 struct {
	beaconCommittee             []incognitokey.CommitteePublicKey
	beaconSubstitute            []incognitokey.CommitteePublicKey
	nextEpochShardCandidate     []incognitokey.CommitteePublicKey
	currentEpochShardCandidate  []incognitokey.CommitteePublicKey
	nextEpochBeaconCandidate    []incognitokey.CommitteePublicKey
	currentEpochBeaconCandidate []incognitokey.CommitteePublicKey
	shardCommittee              map[byte][]incognitokey.CommitteePublicKey
	shardSubstitute             map[byte][]incognitokey.CommitteePublicKey
	autoStake                   map[string]bool                   // committee public key => reward receiver payment address
	rewardReceiver              map[string]privacy.PaymentAddress // incognito public key => reward receiver payment address
	stakingTx                   map[string]common.Hash            // committee public key => reward receiver payment address
	unstake                     map[string]bool                   // committee public key => isExist ?

	mu *sync.RWMutex
}

type BeaconCommitteeEngineV2 struct {
	beaconHeight                      uint64
	beaconHash                        common.Hash
	beaconCommitteeStateV2            *BeaconCommitteeStateV2
	uncommittedBeaconCommitteeStateV2 *BeaconCommitteeStateV2
	CommitteeGetter                   func(blkHash common.Hash, committeeID int) ([]incognitokey.CommitteePublicKey, error)
	SubstituteGetter                  func(blkHash common.Hash, committeeID int) ([]incognitokey.CommitteePublicKey, error)
	CandidateGetter                   func(blkHash common.Hash, getBeacon bool) ([]incognitokey.CommitteePublicKey, error)
}

func (b *BeaconCommitteeEngineV2) GetBeaconHeight() uint64 {
	panic("implement me")
}

func (b *BeaconCommitteeEngineV2) GetBeaconHash() common.Hash {
	panic("implement me")
}

func (b *BeaconCommitteeEngineV2) GetBeaconCommittee() []incognitokey.CommitteePublicKey {
	panic("implement me")
}

func (b *BeaconCommitteeEngineV2) GetBeaconSubstitute() []incognitokey.CommitteePublicKey {
	panic("implement me")
}

func (b *BeaconCommitteeEngineV2) GetCandidateShardWaitingForCurrentRandom() []incognitokey.CommitteePublicKey {
	panic("implement me")
}

func (b *BeaconCommitteeEngineV2) GetCandidateBeaconWaitingForCurrentRandom() []incognitokey.CommitteePublicKey {
	panic("implement me")
}

func (b *BeaconCommitteeEngineV2) GetCandidateShardWaitingForNextRandom() []incognitokey.CommitteePublicKey {
	panic("implement me")
}

func (b *BeaconCommitteeEngineV2) GetCandidateBeaconWaitingForNextRandom() []incognitokey.CommitteePublicKey {
	panic("implement me")
}

func (b *BeaconCommitteeEngineV2) GetOneShardCommittee(shardID byte) []incognitokey.CommitteePublicKey {
	panic("implement me")
}

func (b *BeaconCommitteeEngineV2) GetShardCommittee() map[byte][]incognitokey.CommitteePublicKey {
	panic("implement me")
}

func (b *BeaconCommitteeEngineV2) GetOneShardSubstitute(shardID byte) []incognitokey.CommitteePublicKey {
	panic("implement me")
}

func (b *BeaconCommitteeEngineV2) GetShardSubstitute() map[byte][]incognitokey.CommitteePublicKey {
	panic("implement me")
}

func (b *BeaconCommitteeEngineV2) GetAutoStaking() map[string]bool {
	panic("implement me")
}

func (b *BeaconCommitteeEngineV2) GetStakingTx() map[string]common.Hash {
	panic("implement me")
}

func (b *BeaconCommitteeEngineV2) GetRewardReceiver() map[string]privacy.PaymentAddress {
	panic("implement me")
}

func (b *BeaconCommitteeEngineV2) GetAllCandidateSubstituteCommittee() []string {
	panic("implement me")
}

func (b *BeaconCommitteeEngineV2) Commit(hash *BeaconCommitteeStateHash) error {
	panic("implement me")
}

func (b *BeaconCommitteeEngineV2) AbortUncommittedBeaconState() {
	panic("implement me")
}

func (b *BeaconCommitteeEngineV2) UpdateCommitteeState(env *BeaconCommitteeStateEnvironment) (*BeaconCommitteeStateHash, *CommitteeChange, error) {
	panic("implement me")
}

func (b *BeaconCommitteeEngineV2) InitCommitteeState(env *BeaconCommitteeStateEnvironment) {
	panic("implement me")
}

func (b *BeaconCommitteeEngineV2) ValidateCommitteeRootHashes(rootHashes []common.Hash) (bool, error) {
	panic("implement me")
}

func (b *BeaconCommitteeEngineV2) GenerateAssignInstruction(candidates []string, numberOfPendingValidator map[byte]int, rand int64, assignOffset int, activeShards int) ([]string, map[byte][]string) {
	panic("implement me")
}

//Unstake : Get map unstake value
func (b *BeaconCommitteeEngineV2) Unstake() map[string]bool {
	b.beaconCommitteeStateV2.mu.RLock()
	defer b.beaconCommitteeStateV2.mu.RUnlock()
	unstake := make(map[string]bool)
	for k, v := range b.beaconCommitteeStateV2.unstake {
		unstake[k] = v
	}
	return unstake
}

func (b *BeaconCommitteeEngineV2) AssignSubstitutePoolUsingRandomInstruction(
	blkHash common.Hash,
	seed int64,
) ([]string, map[byte][]string) {
	//Still update
	subsSizeMap := map[byte]int{}
	numShards := len(b.beaconCommitteeStateV2.shardCommittee)
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

func (b *BeaconCommitteeEngineV2) AssignShardsPoolUsingRandomInstruction(
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

func (b *BeaconCommitteeEngineV2) AssignBeaconUsingRandomInstruction(
	seed int64,
	candidateList []string,
) []string {
	//TODO
	return candidateList
}
