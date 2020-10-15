package blockchain

import (
	"errors"
	"strings"

	"github.com/incognitochain/incognito-chain/blockchain/committeestate"
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/instruction"
)

func GetStakingCandidate(beaconBlock types.BeaconBlock) ([]string, []string) {
	beacon := []string{}
	shard := []string{}
	beaconBlockBody := beaconBlock.Body
	for _, v := range beaconBlockBody.Instructions {
		if len(v) < 1 {
			continue
		}
		if v[0] == instruction.STAKE_ACTION && v[2] == "beacon" {
			beacon = strings.Split(v[1], ",")
		}
		if v[0] == instruction.STAKE_ACTION && v[2] == "shard" {
			shard = strings.Split(v[1], ",")
		}
	}

	return beacon, shard
}

func filterValidators(
	validators []string,
	producersBlackList map[string]uint8,
	isExistenceIncluded bool,
) []string {
	resultingValidators := []string{}
	for _, pv := range validators {
		_, found := producersBlackList[pv]
		if (found && isExistenceIncluded) || (!found && !isExistenceIncluded) {
			resultingValidators = append(resultingValidators, pv)
		}
	}
	return resultingValidators
}

func isBadProducer(badProducers []string, producer string) bool {
	for _, badProducer := range badProducers {
		if badProducer == producer {
			return true
		}
	}
	return false
}

func CreateBeaconSwapActionForKeyListV2(
	genesisParam *GenesisParams,
	beaconCommittees []string,
	minCommitteeSize int,
	epoch uint64,
) ([]string, []string) {
	swapInstruction, newBeaconCommittees := GetBeaconSwapInstructionKeyListV2(genesisParam, epoch)
	remainBeaconCommittees := beaconCommittees[minCommitteeSize:]
	return swapInstruction, append(newBeaconCommittees, remainBeaconCommittees...)
}

func swap(
	badPendingValidators []string,
	goodPendingValidators []string,
	currentGoodProducers []string,
	currentBadProducers []string,
	maxCommittee int,
	offset int,
) ([]string, []string, []string, []string, error) {
	// if swap offset = 0 then do nothing
	if offset == 0 {
		// return pendingValidators, currentGoodProducers, currentBadProducers, []string{}, errors.New("no pending validator for swapping")
		return append(goodPendingValidators, badPendingValidators...), currentGoodProducers, currentBadProducers, []string{}, nil
	}
	if offset > maxCommittee {
		return append(goodPendingValidators, badPendingValidators...), currentGoodProducers, currentBadProducers, []string{}, errors.New("try to swap too many validators")
	}
	tempValidators := []string{}
	swapValidator := currentBadProducers
	diff := maxCommittee - len(currentGoodProducers)
	if diff >= offset {
		tempValidators = append(tempValidators, goodPendingValidators[:offset]...)
		currentGoodProducers = append(currentGoodProducers, tempValidators...)
		goodPendingValidators = goodPendingValidators[offset:]
		return append(goodPendingValidators, badPendingValidators...), currentGoodProducers, swapValidator, tempValidators, nil
	}
	offset -= diff
	tempValidators = append(tempValidators, goodPendingValidators[:diff]...)
	goodPendingValidators = goodPendingValidators[diff:]
	currentGoodProducers = append(currentGoodProducers, tempValidators...)

	// out pubkey: swapped out validator
	swapValidator = append(swapValidator, currentGoodProducers[:offset]...)
	// unqueue validator with index from 0 to offset-1 from currentValidators list
	currentGoodProducers = currentGoodProducers[offset:]
	// in pubkey: unqueue validator with index from 0 to offset-1 from pendingValidators list
	tempValidators = append(tempValidators, goodPendingValidators[:offset]...)
	// enqueue new validator to the remaning of current validators list
	currentGoodProducers = append(currentGoodProducers, goodPendingValidators[:offset]...)
	// save new pending validators list
	goodPendingValidators = goodPendingValidators[offset:]
	return append(goodPendingValidators, badPendingValidators...), currentGoodProducers, swapValidator, tempValidators, nil
}

// SwapValidator consider these list as queue structure
// unqueue a number of validator out of currentValidators list
// enqueue a number of validator into currentValidators list <=> unqueue a number of validator out of pendingValidators list
// return value: #1 remaining pendingValidators, #2 new currentValidators #3 swapped out validator, #4 incoming validator #5 error
func SwapValidator(
	pendingValidators []string,
	currentValidators []string,
	maxCommittee int,
	minCommittee int,
	offset int,
	swapOffset int,
) ([]string, []string, []string, []string, error) {
	producersBlackList := make(map[string]uint8)
	goodPendingValidators := filterValidators(pendingValidators, producersBlackList, false)
	badPendingValidators := filterValidators(pendingValidators, producersBlackList, true)
	currentBadProducers := filterValidators(currentValidators, producersBlackList, true)
	currentGoodProducers := filterValidators(currentValidators, producersBlackList, false)
	goodPendingValidatorsLen := len(goodPendingValidators)
	currentGoodProducersLen := len(currentGoodProducers)
	// number of good producer more than minimum needed producer to continue
	if currentGoodProducersLen >= minCommittee {
		// current number of good producer reach maximum committee size => swap
		if currentGoodProducersLen == maxCommittee {
			offset = swapOffset
		}
		// if not then number of good producer are less than maximum committee size
		// push more pending validator into committee list

		// if number of current good pending validators are less than maximum push offset
		// then push all good pending validator into committee
		if offset > goodPendingValidatorsLen {
			offset = goodPendingValidatorsLen
		}
		return swap(badPendingValidators, goodPendingValidators, currentGoodProducers, currentBadProducers, maxCommittee, offset)
	}

	minProducersNeeded := minCommittee - currentGoodProducersLen
	if len(pendingValidators) >= minProducersNeeded {
		if offset < minProducersNeeded {
			offset = minProducersNeeded
		} else if offset > goodPendingValidatorsLen {
			offset = goodPendingValidatorsLen
		}
		return swap(badPendingValidators, goodPendingValidators, currentGoodProducers, currentBadProducers, maxCommittee, offset)
	}

	producersNumCouldBeSwapped := len(goodPendingValidators) + len(currentValidators) - minCommittee
	swappedProducers := []string{}
	remainingProducers := []string{}
	for _, producer := range currentValidators {
		if isBadProducer(currentBadProducers, producer) && len(swappedProducers) < producersNumCouldBeSwapped {
			swappedProducers = append(swappedProducers, producer)
			continue
		}
		remainingProducers = append(remainingProducers, producer)
	}
	newProducers := append(remainingProducers, goodPendingValidators...)
	return badPendingValidators, newProducers, swappedProducers, goodPendingValidators, nil
}

func (beaconBestState *BeaconBestState) postProcessIncurredInstructions(instructions [][]string) error {

	for _, inst := range instructions {
		switch inst[0] {
		case instruction.RETURN_ACTION:
			returnStakingIns, err := instruction.ValidateAndImportReturnStakingInstructionFromString(inst)
			if err != nil {
				return err
			}
			err = statedb.DeleteStakerInfo(beaconBestState.consensusStateDB, returnStakingIns.PublicKeysStruct)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (beaconBestState *BeaconBestState) preProcessInstructionsFromShardBlock(instructions [][]string, shardID byte) *shardInstruction {
	res := &shardInstruction{
		swapInstructions: make(map[byte][]*instruction.SwapInstruction),
	}
	// extract instructions
	for _, inst := range instructions {
		if len(inst) > 0 {
			if inst[0] == instruction.STAKE_ACTION {
				if err := instruction.ValidateStakeInstructionSanity(inst); err != nil {
					Logger.log.Errorf("SKIP Stake Instruction Error %+v", err)
					continue
				}
				tempStakeInstruction := instruction.ImportStakeInstructionFromString(inst)
				res.stakeInstructions = append(res.stakeInstructions, tempStakeInstruction)
			}
			if inst[0] == instruction.SWAP_ACTION {
				// validate swap instruction
				// only allow shard to swap committee for it self
				if err := instruction.ValidateSwapInstructionSanity(inst); err != nil {
					Logger.log.Errorf("SKIP Swap Instruction Error %+v", err)
					continue
				}
				tempSwapInstruction := instruction.ImportSwapInstructionFromString(inst)
				res.swapInstructions[shardID] = append(res.swapInstructions[shardID], tempSwapInstruction)
			}
			if inst[0] == instruction.STOP_AUTO_STAKE_ACTION {
				if err := instruction.ValidateStopAutoStakeInstructionSanity(inst); err != nil {
					Logger.log.Errorf("SKIP Stop Auto Stake Instruction Error %+v", err)
					continue
				}
				tempStopAutoStakeInstruction := instruction.ImportStopAutoStakeInstructionFromString(inst)
				res.stopAutoStakeInstructions = append(res.stopAutoStakeInstructions, tempStopAutoStakeInstruction)
			}
			if inst[0] == instruction.UNSTAKE_ACTION {
				if err := instruction.ValidateUnstakeInstructionSanity(inst); err != nil {
					Logger.log.Errorf("SKIP Stop Auto Stake Instruction Error %+v", err)
					continue
				}
				tempUnstakeInstruction := instruction.ImportUnstakeInstructionFromString(inst)
				res.unstakeInstructions = append(res.unstakeInstructions, tempUnstakeInstruction)
			}
		}
	}

	if len(res.stakeInstructions) != 0 {
		Logger.log.Info("Beacon Producer/ Process Stakers List ", res.stakeInstructions)
	}
	if len(res.swapInstructions[shardID]) != 0 {
		Logger.log.Info("Beacon Producer/ Process Stakers List ", res.swapInstructions[shardID])
	}

	return res
}

func (beaconBestState *BeaconBestState) processStakeInstructionFromShardBlock(
	shardInstructions *shardInstruction, validStakePublicKeys []string) (
	*shardInstruction, *duplicateKeyStakeInstruction) {

	duplicateKeyStakeInstruction := &duplicateKeyStakeInstruction{}
	newShardInstructions := shardInstructions
	stakeInstructions := []*instruction.StakeInstruction{}
	stakeShardPublicKeys := []string{}
	stakeShardTx := []string{}
	stakeShardRewardReceiver := []string{}
	stakeShardAutoStaking := []bool{}
	tempValidStakePublicKeys := []string{}

	// Process Stake Instruction form Shard Block
	// Validate stake instruction => extract only valid stake instruction
	for _, stakeInstruction := range shardInstructions.stakeInstructions {
		tempStakePublicKey := stakeInstruction.PublicKeys
		duplicateStakePublicKeys := []string{}
		// list of stake public keys and stake transaction and reward receiver must have equal length

		tempStakePublicKey = beaconBestState.GetValidStakers(tempStakePublicKey)
		tempStakePublicKey = common.GetValidStaker(stakeShardPublicKeys, tempStakePublicKey)
		tempStakePublicKey = common.GetValidStaker(validStakePublicKeys, tempStakePublicKey)

		if len(tempStakePublicKey) > 0 {
			stakeShardPublicKeys = append(stakeShardPublicKeys, tempStakePublicKey...)
			for i, v := range stakeInstruction.PublicKeys {
				if common.IndexOfStr(v, tempStakePublicKey) > -1 {
					stakeShardTx = append(stakeShardTx, stakeInstruction.TxStakes[i])
					stakeShardRewardReceiver = append(stakeShardRewardReceiver, stakeInstruction.RewardReceivers[i])
					stakeShardAutoStaking = append(stakeShardAutoStaking, stakeInstruction.AutoStakingFlag[i])
				}
			}
		}

		if beaconBestState.beaconCommitteeEngine.Version() == committeestate.SLASHING_VERSION &&
			(len(stakeInstruction.PublicKeys) != len(tempStakePublicKey)) {
			duplicateStakePublicKeys = common.DifferentElementStrings(stakeInstruction.PublicKeys, tempStakePublicKey)
			if len(duplicateStakePublicKeys) > 0 {
				stakingTxs := []string{}
				autoStaking := []bool{}
				rewardReceivers := []string{}
				for i, v := range stakeInstruction.PublicKeys {
					if common.IndexOfStr(v, duplicateStakePublicKeys) > -1 {
						stakingTxs = append(stakingTxs, stakeInstruction.TxStakes[i])
						rewardReceivers = append(rewardReceivers, stakeInstruction.RewardReceivers[i])
						autoStaking = append(autoStaking, stakeInstruction.AutoStakingFlag[i])
					}
				}
				duplicateStakeInstruction := instruction.NewStakeInstructionWithValue(
					duplicateStakePublicKeys,
					stakeInstruction.Chain,
					stakingTxs,
					rewardReceivers,
					autoStaking,
				)
				duplicateKeyStakeInstruction.instructions = append(duplicateKeyStakeInstruction.instructions, duplicateStakeInstruction)
			}
		}
	}

	if len(stakeShardPublicKeys) > 0 {
		tempValidStakePublicKeys = append(tempValidStakePublicKeys, stakeShardPublicKeys...)
		tempStakeShardInstruction := instruction.NewStakeInstructionWithValue(
			stakeShardPublicKeys,
			instruction.SHARD_INST,
			stakeShardTx, stakeShardRewardReceiver,
			stakeShardAutoStaking,
		)
		stakeInstructions = append(stakeInstructions, tempStakeShardInstruction)
		validStakePublicKeys = append(validStakePublicKeys, stakeShardPublicKeys...)
	}

	newShardInstructions.stakeInstructions = stakeInstructions
	return newShardInstructions, duplicateKeyStakeInstruction
}

func (beaconBestState *BeaconBestState) processStopAutoStakeInstructionFromShardBlock(
	shardInstructions *shardInstruction, allCommitteeValidatorCandidate []string) *shardInstruction {

	stopAutoStakingPublicKeys := []string{}
	stopAutoStakeInstructions := []*instruction.StopAutoStakeInstruction{}

	for _, stopAutoStakeInstruction := range shardInstructions.stopAutoStakeInstructions {
		for _, tempStopAutoStakingPublicKey := range stopAutoStakeInstruction.CommitteePublicKeys {
			if common.IndexOfStr(tempStopAutoStakingPublicKey, allCommitteeValidatorCandidate) > -1 {
				stopAutoStakingPublicKeys = append(stopAutoStakingPublicKeys, tempStopAutoStakingPublicKey)
			}
		}
	}

	if len(stopAutoStakingPublicKeys) > 0 {
		tempStopAutoStakeInstruction := instruction.NewStopAutoStakeInstructionWithValue(stopAutoStakingPublicKeys)
		stopAutoStakeInstructions = append(stopAutoStakeInstructions, tempStopAutoStakeInstruction)
	}

	shardInstructions.stopAutoStakeInstructions = stopAutoStakeInstructions
	return shardInstructions
}

func (beaconBestState *BeaconBestState) processUnstakeInstructionFromShardBlock(
	shardInstructions *shardInstruction,
	allCommitteeValidatorCandidate []string,
	shardID byte,
	validUnstakePublicKeys map[string]bool) *shardInstruction {
	unstakingPublicKeys := []string{}
	unstakeInstructions := []*instruction.UnstakeInstruction{}

	for _, unstakeInstruction := range shardInstructions.unstakeInstructions {
		for _, tempUnstakePublicKey := range unstakeInstruction.CommitteePublicKeys {
			if validUnstakePublicKeys[tempUnstakePublicKey] {
				Logger.log.Infof("SHARD %v | UNSTAKE duplicated unstake instruction | ", shardID)
				continue
			}
			if common.IndexOfStr(tempUnstakePublicKey, allCommitteeValidatorCandidate) > -1 {
				unstakingPublicKeys = append(unstakingPublicKeys, tempUnstakePublicKey)
			}
			validUnstakePublicKeys[tempUnstakePublicKey] = true
		}
	}
	if len(unstakingPublicKeys) > 0 {
		tempUnstakeInstruction := instruction.NewUnstakeInstructionWithValue(unstakingPublicKeys)
		tempUnstakeInstruction.SetCommitteePublicKeys(unstakingPublicKeys)
		unstakeInstructions = append(unstakeInstructions, tempUnstakeInstruction)
	}

	shardInstructions.unstakeInstructions = unstakeInstructions
	return shardInstructions

}

func (shardInstruction *shardInstruction) compose() {
	stakeInstruction := &instruction.StakeInstruction{}
	unstakeInstruction := &instruction.UnstakeInstruction{}
	stopAutoStakeInstruction := &instruction.StopAutoStakeInstruction{}

	for _, v := range shardInstruction.stakeInstructions {
		stakeInstruction.PublicKeys = append(stakeInstruction.PublicKeys, v.PublicKeys...)
		stakeInstruction.PublicKeyStructs = append(stakeInstruction.PublicKeyStructs, v.PublicKeyStructs...)
		stakeInstruction.TxStakeHashes = append(stakeInstruction.TxStakeHashes, v.TxStakeHashes...)
		stakeInstruction.TxStakes = append(stakeInstruction.TxStakes, v.TxStakes...)
		stakeInstruction.RewardReceivers = append(stakeInstruction.RewardReceivers, v.RewardReceivers...)
		stakeInstruction.RewardReceiverStructs = append(stakeInstruction.RewardReceiverStructs, v.RewardReceiverStructs...)
		stakeInstruction.Chain = v.Chain
		stakeInstruction.AutoStakingFlag = append(stakeInstruction.AutoStakingFlag, v.AutoStakingFlag...)
	}

	for _, v := range shardInstruction.unstakeInstructions {
		unstakeInstruction.CommitteePublicKeys = append(unstakeInstruction.CommitteePublicKeys, v.CommitteePublicKeys...)
		unstakeInstruction.CommitteePublicKeysStruct = append(unstakeInstruction.CommitteePublicKeysStruct, v.CommitteePublicKeysStruct...)
	}

	for _, v := range shardInstruction.stopAutoStakeInstructions {
		stopAutoStakeInstruction.CommitteePublicKeys = append(stopAutoStakeInstruction.CommitteePublicKeys, v.CommitteePublicKeys...)
	}

	shardInstruction.stakeInstructions = []*instruction.StakeInstruction{}
	shardInstruction.stakeInstructions = append(shardInstruction.stakeInstructions, stakeInstruction)
	shardInstruction.unstakeInstructions = []*instruction.UnstakeInstruction{}
	shardInstruction.unstakeInstructions = append(shardInstruction.unstakeInstructions, unstakeInstruction)
	shardInstruction.stopAutoStakeInstructions = []*instruction.StopAutoStakeInstruction{}
	shardInstruction.stopAutoStakeInstructions = append(shardInstruction.stopAutoStakeInstructions, stopAutoStakeInstruction)
}
