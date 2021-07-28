package pdex

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	instruction "github.com/incognitochain/incognito-chain/instruction/pdexv3"
	"github.com/incognitochain/incognito-chain/metadata"
	metadataPdexv3 "github.com/incognitochain/incognito-chain/metadata/pdexv3"
)

type stateProcessorV2 struct {
	stateProcessorBase
}

func (sp *stateProcessorV2) addLiquidity(
	stateDB *statedb.StateDB,
	inst []string,
	beaconHeight uint64,
	poolPairs map[string]PoolPairState,
	waitingContributions map[string]rawdbv2.Pdexv3Contribution,
	deletedWaitingContributions map[string]rawdbv2.Pdexv3Contribution,
) (
	map[string]PoolPairState,
	map[string]rawdbv2.Pdexv3Contribution,
	map[string]rawdbv2.Pdexv3Contribution,
	error,
) {
	var err error
	switch inst[len(inst)-1] {
	case strconv.Itoa(common.PDEContributionWaitingStatus):
		waitingContributions, err = sp.waitingContribution(stateDB, inst, waitingContributions, deletedWaitingContributions)
		if err != nil {
			return poolPairs, waitingContributions, deletedWaitingContributions, err
		}
	case strconv.Itoa(common.PDEContributionRefundStatus):
		waitingContributions, deletedWaitingContributions, err = sp.refundContribution(stateDB, inst, waitingContributions, deletedWaitingContributions)
		if err != nil {
			return poolPairs, waitingContributions, deletedWaitingContributions, err
		}
	case strconv.Itoa(common.PDEContributionAcceptedStatus):
		waitingContributions, deletedWaitingContributions, poolPairs, err = sp.matchContribution(
			stateDB, inst, beaconHeight, waitingContributions, deletedWaitingContributions, poolPairs)
		if err != nil {
			return poolPairs, waitingContributions, deletedWaitingContributions, err
		}
	case strconv.Itoa(common.PDEContributionMatchedNReturnedStatus):
		waitingContributions, deletedWaitingContributions, poolPairs, err = sp.matchAndReturnContribution(
			stateDB, inst, beaconHeight,
			waitingContributions, deletedWaitingContributions, poolPairs)
		if err != nil {
			return poolPairs, waitingContributions, deletedWaitingContributions, err
		}
	}
	return poolPairs, waitingContributions, deletedWaitingContributions, nil
}

func (sp *stateProcessorV2) waitingContribution(
	stateDB *statedb.StateDB,
	inst []string,
	waitingContributions map[string]rawdbv2.Pdexv3Contribution,
	deletedWaitingContributions map[string]rawdbv2.Pdexv3Contribution,
) (map[string]rawdbv2.Pdexv3Contribution, error) {
	waitingAddLiquidityInst := instruction.WaitingAddLiquidity{}
	err := waitingAddLiquidityInst.FromStringSlice(inst)
	if err != nil {
		return waitingContributions, err
	}
	err = sp.verifyWaitingContribution(waitingAddLiquidityInst.Contribution(), waitingContributions, deletedWaitingContributions)
	if err != nil {
		return waitingContributions, err
	}
	contribution := waitingAddLiquidityInst.Contribution()
	contributionValue := contribution.Value()
	waitingContributions[contribution.PairHash()] = contributionValue

	contribStatus := metadata.PDEContributionStatus{
		Contributed1Amount: contributionValue.Amount(),
		TokenID1Str:        contributionValue.TokenID().String(),
		Status:             byte(common.PDEContributionWaitingStatus),
	}
	contribStatusBytes, _ := json.Marshal(contribStatus)
	err = statedb.TrackPDEContributionStatus(
		stateDB,
		rawdbv2.PDEContributionStatusPrefix,
		[]byte(contribution.PairHash()),
		contribStatusBytes,
	)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while tracking pde waiting contribution status: %+v", err)
		return waitingContributions, err
	}

	return waitingContributions, nil
}

func (sp *stateProcessorV2) verifyWaitingContribution(
	contribution statedb.Pdexv3ContributionState,
	waitingContributions map[string]rawdbv2.Pdexv3Contribution,
	deletedWaitingContributions map[string]rawdbv2.Pdexv3Contribution,
) error {
	_, found := waitingContributions[contribution.PairHash()]
	if found {
		err := fmt.Errorf("Pair Hash %v has been existed in list waitingContributions", contribution.PairHash())
		return err
	}
	_, found = deletedWaitingContributions[contribution.PairHash()]
	if found {
		err := fmt.Errorf("Pair Hash %v has been existed in list deletedWaitingContributions", contribution.PairHash())
		return err
	}
	return nil
}

func (sp *stateProcessorV2) refundContribution(
	stateDB *statedb.StateDB,
	inst []string,
	waitingContributions map[string]rawdbv2.Pdexv3Contribution,
	deletedWaitingContributions map[string]rawdbv2.Pdexv3Contribution,
) (
	map[string]rawdbv2.Pdexv3Contribution,
	map[string]rawdbv2.Pdexv3Contribution,
	error,
) {
	refundAddLiquidityInst := instruction.RefundAddLiquidity{}
	err := refundAddLiquidityInst.FromStringSlice(inst)
	if err != nil {
		return waitingContributions, deletedWaitingContributions, err
	}
	refundContribution := refundAddLiquidityInst.Contribution()
	existingWaitingContribution, found := waitingContributions[refundContribution.PairHash()]
	if found {
		deletedWaitingContributions[refundContribution.PairHash()] = existingWaitingContribution
		delete(waitingContributions, refundContribution.PairHash())
	}

	contribStatus := metadata.PDEContributionStatus{
		Status: byte(common.PDEContributionRefundStatus),
	}
	contribStatusBytes, _ := json.Marshal(contribStatus)
	err = statedb.TrackPDEContributionStatus(
		stateDB,
		rawdbv2.PDEContributionStatusPrefix,
		[]byte(refundContribution.PairHash()),
		contribStatusBytes,
	)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while tracking pde refund contribution status: %+v", err)
		return waitingContributions, deletedWaitingContributions, err
	}

	return waitingContributions, deletedWaitingContributions, nil
}

func (sp *stateProcessorV2) matchContribution(
	stateDB *statedb.StateDB,
	inst []string,
	beaconHeight uint64,
	waitingContributions map[string]rawdbv2.Pdexv3Contribution,
	deletedWaitingContributions map[string]rawdbv2.Pdexv3Contribution,
	poolPairs map[string]PoolPairState,
) (
	map[string]rawdbv2.Pdexv3Contribution,
	map[string]rawdbv2.Pdexv3Contribution,
	map[string]PoolPairState,
	error,
) {
	matchAddLiquidityInst := instruction.MatchAddLiquidity{}
	err := matchAddLiquidityInst.FromStringSlice(inst)
	if err != nil {
		return waitingContributions, deletedWaitingContributions, poolPairs, err
	}
	matchContribution := matchAddLiquidityInst.Contribution()
	existingWaitingContribution, found := waitingContributions[matchContribution.PairHash()]
	if !found {
		err := fmt.Errorf("ERROR: could not find out existing waiting contribution with unique pair id: %s", matchContribution.PairHash())
		return waitingContributions, deletedWaitingContributions, poolPairs, err
	}
	_, found = deletedWaitingContributions[matchContribution.PairHash()]
	if found {
		err := fmt.Errorf("Pair Hash %v has been existed in list deletedWaitingContributions", matchContribution.PairHash())
		return waitingContributions, deletedWaitingContributions, poolPairs, err
	}

	matchContributionValue := matchContribution.Value()
	poolPair := initPoolPairState(existingWaitingContribution, matchContribution.Value())
	poolPairID := generatePoolPairKey(
		existingWaitingContribution.TokenID().String(),
		matchContributionValue.TokenID().String(),
		existingWaitingContribution.TxReqID().String(),
	)
	poolPair.addShare(poolPairID, poolPair.state.ShareAmount(), beaconHeight)
	poolPairs[poolPairID] = *poolPair
	deletedWaitingContributions[matchContribution.PairHash()] = existingWaitingContribution
	delete(waitingContributions, matchContribution.PairHash())

	contribStatus := metadata.PDEContributionStatus{
		Status: byte(common.PDEContributionAcceptedStatus),
	}
	contribStatusBytes, _ := json.Marshal(contribStatus)
	err = statedb.TrackPDEContributionStatus(
		stateDB,
		rawdbv2.PDEContributionStatusPrefix,
		[]byte(matchContribution.PairHash()),
		contribStatusBytes,
	)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while tracking pde accepted contribution status: %+v", err)
		return waitingContributions, deletedWaitingContributions, poolPairs, err
	}

	return waitingContributions, deletedWaitingContributions, poolPairs, nil
}

func (sp *stateProcessorV2) matchAndReturnContribution(
	stateDB *statedb.StateDB,
	inst []string,
	beaconHeight uint64,
	waitingContributions map[string]rawdbv2.Pdexv3Contribution,
	deletedWaitingContributions map[string]rawdbv2.Pdexv3Contribution,
	poolPairs map[string]PoolPairState,
) (
	map[string]rawdbv2.Pdexv3Contribution,
	map[string]rawdbv2.Pdexv3Contribution,
	map[string]PoolPairState,
	error,
) {
	matchAndReturnAddLiquidity := instruction.MatchAndReturnAddLiquidity{}
	err := matchAndReturnAddLiquidity.FromStringSlice(inst)
	if err != nil {
		return waitingContributions, deletedWaitingContributions, poolPairs, err
	}
	matchAndReturnContribution := matchAndReturnAddLiquidity.Contribution()
	matchAndReturnContributionValue := matchAndReturnContribution.Value()
	waitingContribution, found := waitingContributions[matchAndReturnContribution.PairHash()]
	if found {
		if matchAndReturnContributionValue.PoolPairID() != waitingContribution.PoolPairID() {
			err := fmt.Errorf("Expect poolPairID %v but get %v", waitingContribution.PoolPairID(), matchAndReturnContributionValue.PoolPairID())
			return waitingContributions, deletedWaitingContributions, poolPairs, err
		}
		poolPair := poolPairs[waitingContribution.PoolPairID()]
		var amount0, amount1 uint64
		if matchAndReturnAddLiquidity.ExistedTokenID().String() < matchAndReturnContributionValue.TokenID().String() {
			amount0 = matchAndReturnAddLiquidity.ExistedTokenActualAmount()
			amount1 = matchAndReturnContributionValue.Amount()
		} else {
			amount1 = matchAndReturnAddLiquidity.ExistedTokenActualAmount()
			amount0 = matchAndReturnContributionValue.Amount()
		}
		poolPair.updateReserveData(amount0, amount1, matchAndReturnAddLiquidity.ShareAmount())
		poolPair.addShare(waitingContribution.PoolPairID(), matchAndReturnAddLiquidity.ShareAmount(), beaconHeight)
		deletedWaitingContributions[matchAndReturnContribution.PairHash()] = waitingContribution
		delete(waitingContributions, matchAndReturnContribution.PairHash())
	} else {
		var contribStatus metadata.PDEContributionStatus
		if matchAndReturnAddLiquidity.ExistedTokenID().String() < matchAndReturnContributionValue.TokenID().String() {
			contribStatus = metadata.PDEContributionStatus{
				Status:             common.PDEContributionMatchedNReturnedStatus,
				TokenID1Str:        matchAndReturnAddLiquidity.ExistedTokenID().String(),
				Contributed1Amount: matchAndReturnAddLiquidity.ExistedTokenActualAmount(),
				Returned1Amount:    matchAndReturnAddLiquidity.ExistedTokenReturnAmount(),
				TokenID2Str:        matchAndReturnContributionValue.TokenID().String(),
				Contributed2Amount: matchAndReturnContributionValue.Amount() - matchAndReturnAddLiquidity.ReturnAmount(),
				Returned2Amount:    matchAndReturnAddLiquidity.ReturnAmount(),
			}
		} else {
			contribStatus = metadata.PDEContributionStatus{
				Status:             common.PDEContributionMatchedNReturnedStatus,
				TokenID2Str:        matchAndReturnAddLiquidity.ExistedTokenID().String(),
				Contributed2Amount: matchAndReturnAddLiquidity.ExistedTokenActualAmount(),
				Returned2Amount:    matchAndReturnAddLiquidity.ExistedTokenReturnAmount(),
				TokenID1Str:        matchAndReturnContributionValue.TokenID().String(),
				Contributed1Amount: matchAndReturnContributionValue.Amount() - matchAndReturnAddLiquidity.ReturnAmount(),
				Returned1Amount:    matchAndReturnAddLiquidity.ReturnAmount(),
			}
		}

		contribStatusBytes, err := json.Marshal(contribStatus)
		if err != nil {
			return waitingContributions, deletedWaitingContributions, poolPairs, err
		}

		err = statedb.TrackPDEContributionStatus(
			stateDB,
			rawdbv2.PDEContributionStatusPrefix,
			[]byte(matchAndReturnContribution.PairHash()),
			contribStatusBytes,
		)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while tracking pde contribution status: %+v", err)
			return waitingContributions, deletedWaitingContributions, poolPairs, err
		}
	}

	return waitingContributions, deletedWaitingContributions, poolPairs, nil
}

func (sp *stateProcessorV2) modifyParams(
	stateDB *statedb.StateDB,
	beaconHeight uint64,
	inst []string,
	params Params,
) (Params, error) {
	if len(inst) != 4 {
		msg := fmt.Sprintf("Length of instruction is not valid expect %v but get %v", 4, len(inst))
		Logger.log.Errorf(msg)
		return params, errors.New(msg)
	}

	// unmarshal instructions content
	var actionData metadataPdexv3.ParamsModifyingContent
	err := json.Unmarshal([]byte(inst[3]), &actionData)
	if err != nil {
		msg := fmt.Sprintf("Could not unmarshal instruction content %v - Error: %v\n", inst[3], err)
		Logger.log.Errorf(msg)
		return params, err
	}

	modifyingStatus := inst[2]
	var reqTrackStatus int
	if modifyingStatus == metadataPdexv3.RequestAcceptedChainStatus {
		params = Params(actionData.Content)
		reqTrackStatus = metadataPdexv3.ParamsModifyingSuccessStatus
	} else {
		reqTrackStatus = metadataPdexv3.ParamsModifyingFailedStatus
	}

	modifyingReqStatus := metadataPdexv3.ParamsModifyingRequestStatus{
		Status:       reqTrackStatus,
		Pdexv3Params: metadataPdexv3.Pdexv3Params(actionData.Content),
	}
	modifyingReqStatusBytes, _ := json.Marshal(modifyingReqStatus)
	err = statedb.TrackPdexv3Status(
		stateDB,
		statedb.Pdexv3ParamsModifyingStatusPrefix(),
		[]byte(actionData.TxReqID.String()),
		modifyingReqStatusBytes,
	)
	if err != nil {
		Logger.log.Errorf("PDex Params Modifying: An error occurred while tracking shielding request tx - Error: %v", err)
	}

	return params, nil
}
