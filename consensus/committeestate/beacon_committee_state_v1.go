package committeestate

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"strconv"
	"strings"
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

func (b *BeaconCommitteeStateV1) UpdateCommitteeState(newBeaconHeight uint64, newBeaconHash common.Hash, beaconInstruction []string) (*CommitteeChange, error) {
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

// validate swap instruction sanity
// new reward receiver only present in replace committee
// ["swap" "inPubkey1,inPubkey2,..." "outPupkey1, outPubkey2,..." "shard" "shardID" "punishedPubkey1,..." "newRewardReceiver1,..."]
// ["swap" "inPubkey1,inPubkey2,..." "outPupkey1, outPubkey2,..." "beacon" "punishedPubkey1,..." "newRewardReceiver1,..."]
func validateSwapInstructionSanity(instruction []string, shardID byte) error {
	if len(instruction) != 5 || len(instruction) != 6 {
		return NewCommitteeStateError(ErrSwapInstructionSanity, fmt.Errorf("invalid instruction length, %+v, %+v", len(instruction), instruction))
	}
	if instruction[0] != swapAction {
		return NewCommitteeStateError(ErrSwapInstructionSanity, fmt.Errorf("invalid swap action, %+v", instruction))
	}
	// beacon instruction
	if len(instruction) == 5 && instruction[3] != beaconInst {
		return NewCommitteeStateError(ErrSwapInstructionSanity, fmt.Errorf("invalid swap beacon instruction, %+v", instruction))
	}
	// shard instruction
	if len(instruction) == 6 && (instruction[3] != shardInst || instruction[4] != strconv.Itoa(int(shardID))) {
		return NewCommitteeStateError(ErrSwapInstructionSanity, fmt.Errorf("invalid swap shard instruction, %+v", instruction))
	}
	return nil
}

// validate stake instruction sanity
// ["stake", "pubkey1,pubkey2,..." "shard" "txStake1,txStake2,..." "rewardReceiver1,rewardReceiver2,..." "flag1,flag2..."]
// ["stake", "pubkey1,pubkey2,..." "beacon" "txStake1,txStake2,..." "rewardReceiver1,rewardReceiver2,..." "flag1,flag2..."]
func validateStakeInstructionSanity(instruction []string) error {
	if len(instruction) != 6 {
		return NewCommitteeStateError(ErrStakeInstructionSanity, fmt.Errorf("invalid length, %+v", instruction))
	}
	if instruction[0] != stakeAction {
		return NewCommitteeStateError(ErrStakeInstructionSanity, fmt.Errorf("invalid swap action, %+v", instruction))
	}
	if instruction[2] != shardInst && instruction[2] != beaconInst {
		return NewCommitteeStateError(ErrStakeInstructionSanity, fmt.Errorf("invalid swap action, %+v", instruction))
	}
	publicKeys := strings.Split(instruction[1], splitter)
	txStakes := strings.Split(instruction[3], splitter)
	rewardReceivers := strings.Split(instruction[4], splitter)
	autoStakings := strings.Split(instruction[5], splitter)
	if len(publicKeys) != len(txStakes) {
		return NewCommitteeStateError(ErrStakeInstructionSanity, fmt.Errorf("invalid public key & tx stake length, %+v", instruction))
	}
	if len(rewardReceivers) != len(txStakes) {
		return NewCommitteeStateError(ErrStakeInstructionSanity, fmt.Errorf("invalid reward receivers & tx stake length, %+v", instruction))
	}
	if len(rewardReceivers) != len(autoStakings) {
		return NewCommitteeStateError(ErrStakeInstructionSanity, fmt.Errorf("invalid reward receivers & tx auto staking length, %+v", instruction))
	}
	return nil
}

func validateStopAutoStakeInstructionSanity(instruction []string) error {
	if len(instruction) != 2 {
		return NewCommitteeStateError(ErrStopAutoStakeInstructionSanity, fmt.Errorf("invalid length, %+v", instruction))
	}
	if instruction[0] != stopAutoStake {
		return NewCommitteeStateError(ErrStopAutoStakeInstructionSanity, fmt.Errorf("invalid stop auto stake action, %+v", instruction))
	}
}
