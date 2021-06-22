package blockchain

import (
	"reflect"
	"strconv"

	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/metadata"
)

func (blockchain *BlockChain) processPDEInstructions(beaconView *BeaconBestState, beaconBlock *types.BeaconBlock) (*CurrentPDEState, error) {
	if !hasPDEInstruction(beaconBlock.Body.Instructions) {
		return beaconView.pdeState, nil
	}
	beaconHeight := beaconBlock.Header.Height - 1
	currentPDEState, err := InitCurrentPDEStateFromDB(
		beaconView.featureStateDB, beaconHeight)
	if err != nil {
		return nil, err
	}

	return currentPDEState, nil
}

func getDiffPDEState(previous *CurrentPDEState, current *CurrentPDEState) (diffState *CurrentPDEState) {
	if current == nil {
		return nil
	}
	if previous == nil {
		return current
	}

	diffState = new(CurrentPDEState)
	diffState.WaitingPDEContributions = make(map[string]*rawdbv2.PDEContribution)
	diffState.DeletedWaitingPDEContributions = make(map[string]*rawdbv2.PDEContribution)
	diffState.PDEPoolPairs = make(map[string]*rawdbv2.PDEPoolForPair)
	diffState.PDEShares = make(map[string]uint64)
	diffState.PDETradingFees = make(map[string]uint64)

	for k, v := range current.WaitingPDEContributions {
		if m, ok := previous.WaitingPDEContributions[k]; !ok || !reflect.DeepEqual(m, v) {
			diffState.WaitingPDEContributions[k] = v
		}
	}
	for k, v := range current.DeletedWaitingPDEContributions {
		if m, ok := previous.DeletedWaitingPDEContributions[k]; !ok || !reflect.DeepEqual(m, v) {
			diffState.DeletedWaitingPDEContributions[k] = v
		}
	}
	for k, v := range current.PDEPoolPairs {
		if m, ok := previous.PDEPoolPairs[k]; !ok || !reflect.DeepEqual(m, v) {
			diffState.PDEPoolPairs[k] = v
		}
	}
	for k, v := range current.PDEShares {
		if m, ok := previous.PDEShares[k]; !ok || !reflect.DeepEqual(m, v) {
			diffState.PDEShares[k] = v
		}
	}
	for k, v := range current.PDETradingFees {
		if m, ok := previous.PDETradingFees[k]; !ok || !reflect.DeepEqual(m, v) {
			diffState.PDETradingFees[k] = v
		}
	}
	return diffState
}

func hasPDEInstruction(instructions [][]string) bool {
	hasPDEXInstruction := false
	for _, inst := range instructions {
		if len(inst) < 2 {
			continue // Not error, just not PDE instruction
		}
		switch inst[0] {
		case strconv.Itoa(metadata.PDEContributionMeta):
			hasPDEXInstruction = true
			break
		case strconv.Itoa(metadata.PDETradeRequestMeta):
			hasPDEXInstruction = true
			break
		case strconv.Itoa(metadata.PDEWithdrawalRequestMeta):
			hasPDEXInstruction = true
			break
		case strconv.Itoa(metadata.PDEFeeWithdrawalRequestMeta):
			hasPDEXInstruction = true
			break
		case strconv.Itoa(metadata.PDEPRVRequiredContributionRequestMeta):
			hasPDEXInstruction = true
			break
		case strconv.Itoa(metadata.PDECrossPoolTradeRequestMeta):
			hasPDEXInstruction = true
			break
		case strconv.Itoa(metadata.PDETradingFeesDistributionMeta):
			hasPDEXInstruction = true
			break
		}
	}
	return hasPDEXInstruction
}
