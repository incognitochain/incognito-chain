package committeestate

import (
	"fmt"
	"strconv"
	"strings"
)

// compressStakeInstructions receives a list of instructions and compress into one instruction
func getVallidStakeInstructions(instruction [][]string) []string {

}

// ["stake", "pubkey1,pubkey2,..." "shard" "txStake1,txStake2,..." "rewardReceiver1,rewardReceiver2,..." "flag1,flag2..."]
// ["stake", "pubkey1,pubkey2,..." "beacon" "txStake1,txStake2,..." "rewardReceiver1,rewardReceiver2,..." "flag1,flag2..."]
func extractStakeInstruction(instruction []string) ([]string, []string, []string, []bool) {
	publicKeys := strings.Split(instruction[1], splitter)
	txStakes := strings.Split(instruction[3], splitter)
	rewardReceivers := strings.Split(instruction[4], splitter)
	tempAutoStakings := strings.Split(instruction[5], splitter)
	autoStakings := []bool{}
	for _, v := range tempAutoStakings {
		if v == "true" {
			autoStakings = append(autoStakings, true)
		} else {
			autoStakings = append(autoStakings, false)
		}
	}
	return publicKeys, txStakes, rewardReceivers, autoStakings
}

// validate stake instruction sanity
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

func validateStopAutoStakeInstructionSanity(instruction []string) error {
	if len(instruction) != 2 {
		return NewCommitteeStateError(ErrStopAutoStakeInstructionSanity, fmt.Errorf("invalid length, %+v", instruction))
	}
	if instruction[0] != stopAutoStake {
		return NewCommitteeStateError(ErrStopAutoStakeInstructionSanity, fmt.Errorf("invalid stop auto stake action, %+v", instruction))
	}
	return nil
}
