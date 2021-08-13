package pdex

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strconv"

	v2 "github.com/incognitochain/incognito-chain/blockchain/pdex/v2utils"
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
	poolPairs map[string]*PoolPairState,
	waitingContributions map[string]rawdbv2.Pdexv3Contribution,
	deletedWaitingContributions map[string]rawdbv2.Pdexv3Contribution,
	nftIDs map[string]bool,
) (
	map[string]*PoolPairState,
	map[string]rawdbv2.Pdexv3Contribution, map[string]rawdbv2.Pdexv3Contribution,
	map[string]bool, error,
) {
	var err error
	switch inst[1] {
	case common.PDEContributionWaitingChainStatus:
		waitingContributions, _, err = sp.waitingContribution(stateDB, inst, waitingContributions, deletedWaitingContributions)
		if err != nil {
			return poolPairs, waitingContributions, deletedWaitingContributions, nftIDs, err
		}
	case common.PDEContributionRefundChainStatus:
		waitingContributions, deletedWaitingContributions, _, err = sp.refundContribution(stateDB, inst, waitingContributions, deletedWaitingContributions)
		if err != nil {
			return poolPairs, waitingContributions, deletedWaitingContributions, nftIDs, err
		}
	case common.PDEContributionMatchedChainStatus:
		waitingContributions, deletedWaitingContributions, poolPairs, nftIDs, _, err = sp.matchContribution(
			stateDB, inst, beaconHeight, waitingContributions, deletedWaitingContributions, poolPairs, nftIDs)
		if err != nil {
			return poolPairs, waitingContributions, deletedWaitingContributions, nftIDs, err
		}
	case common.PDEContributionMatchedNReturnedChainStatus:
		waitingContributions,
			deletedWaitingContributions, poolPairs, nftIDs, _, err = sp.matchAndReturnContribution(
			stateDB, inst, beaconHeight,
			waitingContributions, deletedWaitingContributions, poolPairs, nftIDs)
		if err != nil {
			return poolPairs, waitingContributions, deletedWaitingContributions, nftIDs, err
		}
	}
	return poolPairs, waitingContributions, deletedWaitingContributions, nftIDs, nil
}

func (sp *stateProcessorV2) waitingContribution(
	stateDB *statedb.StateDB,
	inst []string,
	waitingContributions map[string]rawdbv2.Pdexv3Contribution,
	deletedWaitingContributions map[string]rawdbv2.Pdexv3Contribution,
) (map[string]rawdbv2.Pdexv3Contribution, *metadata.PDEContributionStatus, error) {
	waitingAddLiquidityInst := instruction.WaitingAddLiquidity{}
	err := waitingAddLiquidityInst.FromStringSlice(inst)
	if err != nil {
		return waitingContributions, nil, err
	}
	err = sp.verifyWaitingContribution(waitingAddLiquidityInst.Contribution(), waitingContributions, deletedWaitingContributions)
	if err != nil {
		return waitingContributions, nil, err
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
		return waitingContributions, nil, err
	}

	return waitingContributions, &contribStatus, nil
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
	*metadata.PDEContributionStatus,
	error,
) {
	refundAddLiquidityInst := instruction.RefundAddLiquidity{}
	err := refundAddLiquidityInst.FromStringSlice(inst)
	if err != nil {
		return waitingContributions, deletedWaitingContributions, nil, err
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
		return waitingContributions, deletedWaitingContributions, nil, err
	}

	return waitingContributions, deletedWaitingContributions, &contribStatus, nil
}

func (sp *stateProcessorV2) matchContribution(
	stateDB *statedb.StateDB,
	inst []string,
	beaconHeight uint64,
	waitingContributions map[string]rawdbv2.Pdexv3Contribution,
	deletedWaitingContributions map[string]rawdbv2.Pdexv3Contribution,
	poolPairs map[string]*PoolPairState, nftIDs map[string]bool,
) (
	map[string]rawdbv2.Pdexv3Contribution, map[string]rawdbv2.Pdexv3Contribution,
	map[string]*PoolPairState, map[string]bool,
	*metadata.PDEContributionStatus, error,
) {
	matchAddLiquidityInst := instruction.MatchAddLiquidity{}
	err := matchAddLiquidityInst.FromStringSlice(inst)
	if err != nil {
		return waitingContributions, deletedWaitingContributions, poolPairs, nftIDs, nil, err
	}
	matchContribution := matchAddLiquidityInst.Contribution()
	existedWaitingContribution, found := waitingContributions[matchContribution.PairHash()]
	if !found {
		err := fmt.Errorf("ERROR: could not find out existing waiting contribution with unique pair id: %s", matchContribution.PairHash())
		return waitingContributions, deletedWaitingContributions, poolPairs, nftIDs, nil, err
	}
	_, found = deletedWaitingContributions[matchContribution.PairHash()]
	if found {
		err := fmt.Errorf("Pair Hash %v has been existed in list deletedWaitingContributions", matchContribution.PairHash())
		return waitingContributions, deletedWaitingContributions, poolPairs, nftIDs, nil, err
	}

	matchContributionValue := matchContribution.Value()
	poolPair := initPoolPairState(existedWaitingContribution, matchContribution.Value())
	poolPairID := generatePoolPairKey(
		existedWaitingContribution.TokenID().String(),
		matchContributionValue.TokenID().String(),
		existedWaitingContribution.TxReqID().String(),
	)
	tempAmt := big.NewInt(0).Mul(
		big.NewInt(0).SetUint64(existedWaitingContribution.Amount()),
		big.NewInt(0).SetUint64(matchContributionValue.Amount()),
	)
	shareAmount := big.NewInt(0).Sqrt(tempAmt).Uint64()
	_, nftIDs, err = poolPair.addShare(
		existedWaitingContribution.NftID(),
		nftIDs, shareAmount, beaconHeight,
		existedWaitingContribution.TxReqID().String(),
	)

	if err != nil {
		return waitingContributions, deletedWaitingContributions, poolPairs, nftIDs, nil, err
	}
	poolPairs[poolPairID] = poolPair

	deletedWaitingContributions[matchContribution.PairHash()] = existedWaitingContribution
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
		return waitingContributions, deletedWaitingContributions, poolPairs, nftIDs, nil, err
	}

	return waitingContributions, deletedWaitingContributions, poolPairs, nftIDs, &contribStatus, nil
}

func (sp *stateProcessorV2) matchAndReturnContribution(
	stateDB *statedb.StateDB,
	inst []string,
	beaconHeight uint64,
	waitingContributions map[string]rawdbv2.Pdexv3Contribution,
	deletedWaitingContributions map[string]rawdbv2.Pdexv3Contribution,
	poolPairs map[string]*PoolPairState,
	nftIDs map[string]bool,
) (
	map[string]rawdbv2.Pdexv3Contribution, map[string]rawdbv2.Pdexv3Contribution,
	map[string]*PoolPairState, map[string]bool,
	*metadata.PDEContributionStatus,
	error,
) {
	matchAndReturnAddLiquidity := instruction.MatchAndReturnAddLiquidity{}
	err := matchAndReturnAddLiquidity.FromStringSlice(inst)
	if err != nil {
		return waitingContributions, deletedWaitingContributions, poolPairs, nftIDs, nil, err
	}
	matchAndReturnContribution := matchAndReturnAddLiquidity.Contribution()
	matchAndReturnContributionValue := matchAndReturnContribution.Value()
	waitingContribution, found := waitingContributions[matchAndReturnContribution.PairHash()]
	var contribStatus metadata.PDEContributionStatus
	if found {
		if matchAndReturnContributionValue.PoolPairID() != waitingContribution.PoolPairID() {
			err := fmt.Errorf("Expect poolPairID %v but get %v", waitingContribution.PoolPairID(), matchAndReturnContributionValue.PoolPairID())
			return waitingContributions, deletedWaitingContributions, poolPairs, nftIDs, nil, err
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
		err = poolPair.addReserveData(amount0, amount1, matchAndReturnAddLiquidity.ShareAmount())
		if err != nil {
			return waitingContributions, deletedWaitingContributions, poolPairs, nftIDs, nil, err
		}
		_, nftIDs, err = poolPair.addShare(
			waitingContribution.NftID(),
			nftIDs,
			matchAndReturnAddLiquidity.ShareAmount(),
			beaconHeight,
			waitingContribution.TxReqID().String(),
		)
		if err != nil {
			return waitingContributions, deletedWaitingContributions, poolPairs, nftIDs, nil, err
		}
		deletedWaitingContributions[matchAndReturnContribution.PairHash()] = waitingContribution
		delete(waitingContributions, matchAndReturnContribution.PairHash())
	} else {
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
			return waitingContributions, deletedWaitingContributions, poolPairs, nftIDs, nil, err
		}

		err = statedb.TrackPDEContributionStatus(
			stateDB,
			rawdbv2.PDEContributionStatusPrefix,
			[]byte(matchAndReturnContribution.PairHash()),
			contribStatusBytes,
		)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while tracking pde contribution status: %+v", err)
			return waitingContributions, deletedWaitingContributions, poolPairs, nftIDs, nil, err
		}
	}

	return waitingContributions, deletedWaitingContributions, poolPairs, nftIDs, &contribStatus, nil
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

func (sp *stateProcessorV2) trade(
	stateDB *statedb.StateDB,
	inst []string,
	pairs map[string]*PoolPairState,
) (map[string]*PoolPairState, error) {
	var currentTrade *instruction.Action
	var trackedStatus metadataPdexv3.TradeStatus
	switch inst[1] {
	case strconv.Itoa(metadataPdexv3.TradeAcceptedStatus):
		currentTrade = &instruction.Action{Content: metadataPdexv3.AcceptedTrade{}}
		err := currentTrade.FromStringSlice(inst)
		if err != nil {
			return pairs, err
		}

		// skip error checking since concrete type is specified above
		md, _ := currentTrade.Content.(*metadataPdexv3.AcceptedTrade)
		for index, pairID := range md.TradePath {
			pair, exists := pairs[pairID]
			if !exists {
				return pairs, fmt.Errorf("Cannot find pair %s for trade", pairID)
			}
			reserveState := &v2.TradingPair{&pair.state}
			err := reserveState.ApplyReserveChanges(md.PairChanges[index][0], md.PairChanges[index][1])
			if err != nil {
				return pairs, err
			}

			orderbook := pair.orderbook
			ordersById := make(map[string]*Order)
			for _, ord := range orderbook.orders {
				ordersById[ord.Id()] = ord
			}
			for id, change := range md.OrderChanges[index] {
				currentOrder, exists := ordersById[id]
				if !exists {
					return pairs, fmt.Errorf("Cannot find order ID %s for trade", id)
				}
				err := (&v2.MatchingOrder{currentOrder}).ApplyBalanceChanges(change[0], change[1])
				if err != nil {
					return pairs, err
				}
			}
		}

		trackedStatus = metadataPdexv3.TradeStatus{
			BuyAmount:  md.Amount,
			TokenToBuy: md.TokenToBuy,
		}
	case strconv.Itoa(metadataPdexv3.TradeRefundedStatus):
		currentTrade = &instruction.Action{Content: metadataPdexv3.RefundedTrade{}}
		err := currentTrade.FromStringSlice(inst)
		if err != nil {
			return pairs, err
		}
	default:
		return pairs, fmt.Errorf("Invalid status %s from instruction", inst[1])
	}

	// store tracked trade status
	trackedStatus.Status = currentTrade.GetStatus()
	marshaledTrackedStatus, err := json.Marshal(trackedStatus)
	if err != nil {
		return pairs, err
	}
	err = statedb.TrackPdexv3Status(
		stateDB,
		statedb.Pdexv3TradeStatusPrefix(),
		currentTrade.RequestTxID[:],
		marshaledTrackedStatus,
	)
	return pairs, nil
}

func (sp *stateProcessorV2) withdrawLiquidity(
	stateDB *statedb.StateDB,
	inst []string,
	poolPairs map[string]*PoolPairState,
) (map[string]*PoolPairState, error) {
	var err error
	switch inst[1] {
	case common.PDEWithdrawalRejectedChainStatus:
		err = sp.rejectWithdrawLiquidity(stateDB, inst)
	case common.PDEWithdrawalAcceptedChainStatus:
		poolPairs, err = sp.acceptWithdrawLiquidity(stateDB, inst, poolPairs)
	}
	if err != nil {
		return poolPairs, err
	}

	return poolPairs, err
}

func (sp *stateProcessorV2) rejectWithdrawLiquidity(
	stateDB *statedb.StateDB, inst []string,
) error {
	rejectWithdrawLiquidity := instruction.NewRejectWithdrawLiquidity()
	err := rejectWithdrawLiquidity.FromStringSlice(inst)
	if err != nil {
		return err
	}
	err = statedb.TrackPdexv3Status(
		stateDB,
		statedb.Pdexv3WithdrawLiquidityStatusPrefix(),
		[]byte(rejectWithdrawLiquidity.TxReqID().String()),
		[]byte{byte(common.PDEWithdrawalRejectedStatus)},
	)
	return err
}

func (sp *stateProcessorV2) acceptWithdrawLiquidity(
	stateDB *statedb.StateDB, inst []string,
	poolPairs map[string]*PoolPairState,
) (map[string]*PoolPairState, error) {
	acceptWithdrawLiquidity := instruction.NewAcceptWithdrawLiquidity()
	err := acceptWithdrawLiquidity.FromStringSlice(inst)
	if err != nil {
		return poolPairs, err
	}
	poolPair, ok := poolPairs[acceptWithdrawLiquidity.PoolPairID()]
	if !ok || poolPair == nil {
		err := fmt.Errorf("Can't find poolPairID %s", acceptWithdrawLiquidity.PoolPairID())
		return poolPairs, err
	}
	share, ok := poolPair.shares[acceptWithdrawLiquidity.NftID().String()]
	if !ok || share == nil {
		err := fmt.Errorf("Can't find nftID %s", acceptWithdrawLiquidity.NftID().String())
		return poolPairs, err
	}
	_, _, _, err = poolPair.deductShare(
		acceptWithdrawLiquidity.NftID().String(),
		acceptWithdrawLiquidity.ShareAmount(),
	)
	err = statedb.TrackPdexv3Status(
		stateDB,
		statedb.Pdexv3WithdrawLiquidityStatusPrefix(),
		[]byte(acceptWithdrawLiquidity.TxReqID().String()),
		[]byte{byte(common.PDEWithdrawalAcceptedStatus)},
	)
	return poolPairs, err
}

func (sp *stateProcessorV2) addOrder(
	stateDB *statedb.StateDB,
	inst []string,
	pairs map[string]PoolPairState,
	orderbooks map[string]Orderbook,
) (map[string]PoolPairState, map[string]Orderbook, error) {
	switch inst[1] {
	case strconv.Itoa(metadataPdexv3.OrderAcceptedStatus):
		currentTrade := &instruction.Action{Content: metadataPdexv3.AcceptedAddOrder{}}
		err := currentTrade.FromStringSlice(inst)
		if err != nil {
			return pairs, orderbooks, err
		}
	case strconv.Itoa(metadataPdexv3.OrderRefundedStatus):
		currentTrade := &instruction.Action{Content: metadataPdexv3.RefundedAddOrder{}}
		err := currentTrade.FromStringSlice(inst)
		if err != nil {
			return pairs, orderbooks, err
		}
	default:
		return pairs, orderbooks, fmt.Errorf("Invalid status %s from instruction", inst[1])
	}
	// TODO : apply state changes
	return pairs, orderbooks, nil
}
