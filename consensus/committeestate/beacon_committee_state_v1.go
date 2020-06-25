package committeestate

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
)

type BeaconCommitteeStateEnvironment struct {
	shardInstructions         [][]string
	newBeaconHeight           uint64
	epochLength               uint64
	epochBreakPointSwapNewKey []uint64
	randomNumber              uint64
}

func NewBeaconCommitteeStateEnvironment(shardInstructions [][]string, newBeaconHeight uint64, epochLength uint64, epochBreakPointSwapNewKey []uint64, randomNumber uint64) *BeaconCommitteeStateEnvironment {
	return &BeaconCommitteeStateEnvironment{shardInstructions: shardInstructions, newBeaconHeight: newBeaconHeight, epochLength: epochLength, epochBreakPointSwapNewKey: epochBreakPointSwapNewKey, randomNumber: randomNumber}
}

type BeaconCommitteeStateV1 struct {
	beaconHeight                           uint64
	beaconHash                             common.Hash
	beaconCommittee                        []incognitokey.CommitteePublicKey
	beaconSubstitute                       []incognitokey.CommitteePublicKey
	candidateShardWaitingForCurrentRandom  []incognitokey.CommitteePublicKey
	candidateBeaconWaitingForCurrentRandom []incognitokey.CommitteePublicKey
	candidateShardWaitingForNextRandom     []incognitokey.CommitteePublicKey
	candidateBeaconWaitingForNextRandom    []incognitokey.CommitteePublicKey
	shardCommittee                         map[byte][]incognitokey.CommitteePublicKey
	shardSubstitute                        map[byte][]incognitokey.CommitteePublicKey
	autoStaking                            map[string]bool
	rewardReceiver                         map[string]string
}

func NewBeaconCommitteeStateV1() *BeaconCommitteeStateV1 {
	return &BeaconCommitteeStateV1{}
}

func NewBeaconCommitteeStateV1WithValue(beaconCommittee []incognitokey.CommitteePublicKey, beaconPendingValidator []incognitokey.CommitteePublicKey, candidateShardWaitingForCurrentRandom []incognitokey.CommitteePublicKey, candidateBeaconWaitingForCurrentRandom []incognitokey.CommitteePublicKey, candidateShardWaitingForNextRandom []incognitokey.CommitteePublicKey, candidateBeaconWaitingForNextRandom []incognitokey.CommitteePublicKey, shardCommittee map[byte][]incognitokey.CommitteePublicKey, shardPendingValidator map[byte][]incognitokey.CommitteePublicKey, autoStaking map[string]bool, rewardReceiver map[string]string) *BeaconCommitteeStateV1 {
	return &BeaconCommitteeStateV1{beaconCommittee: beaconCommittee, beaconSubstitute: beaconPendingValidator, candidateShardWaitingForCurrentRandom: candidateShardWaitingForCurrentRandom, candidateBeaconWaitingForCurrentRandom: candidateBeaconWaitingForCurrentRandom, candidateShardWaitingForNextRandom: candidateShardWaitingForNextRandom, candidateBeaconWaitingForNextRandom: candidateBeaconWaitingForNextRandom, shardCommittee: shardCommittee, shardSubstitute: shardPendingValidator, autoStaking: autoStaking, rewardReceiver: rewardReceiver}
}

func (b BeaconCommitteeStateV1) GenerateBeaconCommitteeInstruction(env *BeaconCommitteeStateEnvironment) {
	panic("implement me")
}

func (b BeaconCommitteeStateV1) GenerateCommitteeRootHashes(beaconInstruction [][]string) ([]common.Hash, error) {
	panic("implement me")
}

func (b *BeaconCommitteeStateV1) UpdateCommitteeState(newBeaconHeight uint64, newBeaconHash common.Hash, beaconInstructions [][]string) (*CommitteeChange, error) {
	panic("implement me")
}

func (b BeaconCommitteeStateV1) ValidateCommitteeRootHashes(rootHashes []common.Hash) (bool, error) {
	panic("implement me")
}

func (b BeaconCommitteeStateV1) GetBeaconHeight() uint64 {
	return b.beaconHeight
}
func (b BeaconCommitteeStateV1) GetBeaconHash() common.Hash {
	return b.beaconHash
}

func (b BeaconCommitteeStateV1) GetBeaconCommittee() []incognitokey.CommitteePublicKey {
	return b.beaconCommittee
}

func (b BeaconCommitteeStateV1) GetBeaconSubstitute() []incognitokey.CommitteePublicKey {
	return b.beaconSubstitute
}

func (b BeaconCommitteeStateV1) GetCandidateShardWaitingForCurrentRandom() []incognitokey.CommitteePublicKey {
	return b.candidateShardWaitingForCurrentRandom
}

func (b BeaconCommitteeStateV1) GetCandidateBeaconWaitingForCurrentRandom() []incognitokey.CommitteePublicKey {
	return b.candidateBeaconWaitingForCurrentRandom
}

func (b BeaconCommitteeStateV1) GetCandidateShardWaitingForNextRandom() []incognitokey.CommitteePublicKey {
	return b.candidateShardWaitingForNextRandom
}

func (b BeaconCommitteeStateV1) GetCandidateBeaconWaitingForNextRandom() []incognitokey.CommitteePublicKey {
	return b.candidateBeaconWaitingForNextRandom
}

func (b BeaconCommitteeStateV1) GetShardCommittee(shardID byte) []incognitokey.CommitteePublicKey {
	return b.shardCommittee[shardID]
}

func (b BeaconCommitteeStateV1) GetShardSubstitute(shardID byte) []incognitokey.CommitteePublicKey {
	return b.shardSubstitute[shardID]
}

func (b BeaconCommitteeStateV1) GetAutoStaking() map[string]bool {
	return b.autoStaking
}

func (b BeaconCommitteeStateV1) GetRewardReceiver() map[string]string {
	return b.rewardReceiver
}

// validate a batch of stake, assign, swap instruction
func (b BeaconCommitteeStateV1) validateStakeInstructions(instructions [][]string, shardID byte) error {

	return nil
}
