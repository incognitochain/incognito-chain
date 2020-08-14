package committeestate

import (
	"math/rand"
	"sync"

	"github.com/incognitochain/incognito-chain/privacy"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
)

type BeaconCommitteeStateV2 struct {
	blockHeight uint64
	blockHash   common.Hash

	beaconCommitteePool []incognitokey.CommitteePublicKey
	beaconPool          []incognitokey.CommitteePublicKey
	beaconCommonPool    []incognitokey.CommitteePublicKey
	assignBCCheckpoint  int

	shardCommitteePool map[byte][]incognitokey.CommitteePublicKey
	shardPool          map[byte][]incognitokey.CommitteePublicKey
	shardCommonPool    []incognitokey.CommitteePublicKey
	assignShCheckpoint int

	autoStake      map[string]bool                   // committee public key => reward receiver payment address
	rewardReceiver map[string]privacy.PaymentAddress // incognito public key => reward receiver payment address
	stakingTx      map[string]common.Hash            // committee public key => reward receiver payment address

	mu *sync.RWMutex
}

type ChainInfoGetter interface {
	CommitteeGetter(blkHash common.Hash, committeeID int) ([]incognitokey.CommitteePublicKey, error)
	SubstituteGetter(blkHash common.Hash, committeeID int) ([]incognitokey.CommitteePublicKey, error)
	CandidateGetter(blkHash common.Hash, getBeacon bool) ([]incognitokey.CommitteePublicKey, error)
	GetActiveShards(blkHash common.Hash) int
}

type BeststateGetter interface {
	CommitteeGetter(committeeID int) ([]incognitokey.CommitteePublicKey, error)
	SubstituteGetter(committeeID int) ([]incognitokey.CommitteePublicKey, error)
	CandidateGetter(getBeacon bool) ([]incognitokey.CommitteePublicKey, error)
	GetActiveShards() int
}

type BeaconCommitteeEngineV2 struct {
	beaconHeight      uint64
	beaconHash        common.Hash
	finalState        *BeaconCommitteeStateV2
	uncommitteedState *BeaconCommitteeStateV2
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

// func (b *BeaconCommitteeEngineV2) GetBeaconCommitteeAtBlock(blkHash *common.Hash) ([]incognitokey.CommitteePublicKey, error) {
// 	if blkHash.IsEqual(&b.beaconHash) {
// 		//remove hardcode
// 		return b.curCommitteeInfo.CommitteeGetter(-1)
// 	}
// 	return b.oldCommitteeInfo.CommitteeGetter(*blkHash, -1)
// }

// func (b *BeaconCommitteeEngineV2) GetBeaconPoolAtBlock(blkHash *common.Hash) ([]incognitokey.CommitteePublicKey, error) {
// 	if blkHash.IsEqual(&b.beaconHash) {
// 		//remove hardcode
// 		return b.curCommitteeInfo.SubstituteGetter(-1)
// 	}
// 	return b.oldCommitteeInfo.SubstituteGetter(*blkHash, -1)
// }

// func (b *BeaconCommitteeEngineV2) GetBeaconCommonPoolAtBlock(blkHash *common.Hash) ([]incognitokey.CommitteePublicKey, error) {
// 	if blkHash.IsEqual(&b.beaconHash) {
// 		//remove hardcode
// 		return b.curCommitteeInfo.CandidateGetter(true)
// 	}
// 	return b.oldCommitteeInfo.CandidateGetter(*blkHash, true)
// }

// func (b *BeaconCommitteeEngineV2) GetShardCommitteeAtBlock(blkHash *common.Hash, shardID byte) ([]incognitokey.CommitteePublicKey, error) {
// 	if blkHash.IsEqual(&b.beaconHash) {
// 		//remove hardcode
// 		return b.curCommitteeInfo.CommitteeGetter(int(shardID))
// 	}
// 	return b.oldCommitteeInfo.CommitteeGetter(*blkHash, int(shardID))
// }

// func (b *BeaconCommitteeEngineV2) GetShardPoolAtBlock(blkHash *common.Hash, shardID byte) ([]incognitokey.CommitteePublicKey, error) {
// 	if blkHash.IsEqual(&b.beaconHash) {
// 		//remove hardcode
// 		return b.curCommitteeInfo.SubstituteGetter(int(shardID))
// 	}
// 	return b.oldCommitteeInfo.SubstituteGetter(*blkHash, int(shardID))
// }

// func (b *BeaconCommitteeEngineV2) GetShardCommonPoolAtBlock(blkHash *common.Hash) ([]incognitokey.CommitteePublicKey, error) {
// 	if blkHash.IsEqual(&b.beaconHash) {
// 		//remove hardcode
// 		return b.curCommitteeInfo.CandidateGetter(false)
// 	}
// 	return b.oldCommitteeInfo.CandidateGetter(*blkHash, false)
// }

func (b *BeaconCommitteeEngineV2) GetCurrentActiveShards() int {
	// return b.finalState.GetActiveShards()
	return 5
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

func (b *BeaconCommitteeEngineV2) AssignSubstitutePoolUsingRandomInstruction(
	blkHash common.Hash,
	seed int64,
) ([]string, map[byte][]string) {
	//Still update
	subsSizeMap := map[byte]int{}
	numShards := b.GetCurrentActiveShards()
	for i := 0; i < numShards; i++ {
		shardSubtitutes, ok := b.finalState.shardPool[byte(i)]
		if !ok {
			//TODO handle error?
		}
		subsSizeMap[byte(i)] = len(shardSubtitutes)
	}
	shardCandidates := b.finalState.shardCommonPool[:b.finalState.assignShCheckpoint]
	beaconCandidates := b.finalState.beaconCommonPool[:b.finalState.assignBCCheckpoint]
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

func (b *BeaconCommitteeEngineV2) getListShardSwapOut(shardID byte) []string {
	return []string{}
}

func (b *BeaconCommitteeEngineV2) getListShardSwapIn(shardID byte) []string {
	return []string{}
}

func (b *BeaconCommitteeEngineV2) getListBeaconSwapOut() []string {
	return []string{}
}

func (b *BeaconCommitteeEngineV2) getListBeaconSwapIn() []string {
	return []string{}
}

func (b *BeaconCommitteeEngineV2) SwapValidator() (
	in map[int][]string,
	out map[int][]string,
) {
	return in, out
}
