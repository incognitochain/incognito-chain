package blockchain

import (
	"errors"
	"strings"

	"github.com/incognitochain/incognito-chain/blockchain/types"
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
	pendingValidator []string,
	beaconCommittees []string,
	minCommitteeSize int,
	epoch uint64,
) ([]string, []string, []string) {
	newPendingValidator := pendingValidator
	swapInstruction, newBeaconCommittees := GetBeaconSwapInstructionKeyListV2(genesisParam, epoch)
	remainBeaconCommittees := beaconCommittees[minCommitteeSize:]
	return swapInstruction, newPendingValidator, append(newBeaconCommittees, remainBeaconCommittees...)
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
	producersBlackList map[string]uint8,
	swapOffset int,
) ([]string, []string, []string, []string, error) {
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

	Logger.log.Info("[unstake] postProcessIncurredInstructions")

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
