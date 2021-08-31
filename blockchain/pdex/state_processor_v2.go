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
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	metadataPdexv3 "github.com/incognitochain/incognito-chain/metadata/pdexv3"
	"github.com/incognitochain/incognito-chain/utils"
)

type stateProcessorV2 struct {
	pairHashCache   map[string]string
	withdrawTxCache map[string]uint64
	stateProcessorBase
}

func (sp *stateProcessorV2) clearCache() {
	sp.pairHashCache = make(map[string]string)
	sp.withdrawTxCache = make(map[string]uint64)
}

func (sp *stateProcessorV2) addLiquidity(
	stateDB *statedb.StateDB,
	inst []string,
	beaconHeight uint64,
	poolPairs map[string]*PoolPairState,
	waitingContributions map[string]rawdbv2.Pdexv3Contribution,
	deletedWaitingContributions map[string]rawdbv2.Pdexv3Contribution,
) (
	map[string]*PoolPairState,
	map[string]rawdbv2.Pdexv3Contribution, map[string]rawdbv2.Pdexv3Contribution, error,
) {
	var err error
	switch inst[1] {
	case common.PDEContributionWaitingChainStatus:
		waitingContributions, _, err = sp.waitingContribution(stateDB, inst, waitingContributions, deletedWaitingContributions)
		if err != nil {
			return poolPairs, waitingContributions, deletedWaitingContributions, err
		}
	case common.PDEContributionRefundChainStatus:
		waitingContributions, deletedWaitingContributions, _, err = sp.refundContribution(stateDB, inst, waitingContributions, deletedWaitingContributions)
		if err != nil {
			return poolPairs, waitingContributions, deletedWaitingContributions, err
		}
	case common.PDEContributionMatchedChainStatus:
		waitingContributions, deletedWaitingContributions, poolPairs, _, err = sp.matchContribution(
			stateDB, inst, beaconHeight, waitingContributions, deletedWaitingContributions, poolPairs)
		if err != nil {
			return poolPairs, waitingContributions, deletedWaitingContributions, err
		}
	case common.PDEContributionMatchedNReturnedChainStatus:
		waitingContributions,
			deletedWaitingContributions, poolPairs, _, err = sp.matchAndReturnContribution(
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
) (map[string]rawdbv2.Pdexv3Contribution, *v2.ContributionStatus, error) {
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

	contribStatus := v2.ContributionStatus{
		Token0ID:                contributionValue.TokenID().String(),
		Token0ContributedAmount: contributionValue.Amount(),
		Status:                  common.PDEContributionWaitingChainStatus,
	}
	contribStatusBytes, _ := json.Marshal(contribStatus)
	err = statedb.TrackPdexv3Status(
		stateDB,
		statedb.Pdexv3ContributionStatusPrefix(),
		contributionValue.TxReqID().Bytes(),
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
	*v2.ContributionStatus,
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
	refundContributionValue := refundContribution.Value()

	contribStatus := v2.ContributionStatus{
		Status: common.PDEContributionRefundChainStatus,
	}
	contribStatusBytes, _ := json.Marshal(contribStatus)
	err = statedb.TrackPdexv3Status(
		stateDB,
		statedb.Pdexv3ContributionStatusPrefix(),
		refundContributionValue.TxReqID().Bytes(),
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
	poolPairs map[string]*PoolPairState,
) (
	map[string]rawdbv2.Pdexv3Contribution, map[string]rawdbv2.Pdexv3Contribution,
	map[string]*PoolPairState,
	*v2.ContributionStatus, error,
) {
	matchAddLiquidityInst := instruction.MatchAddLiquidity{}
	err := matchAddLiquidityInst.FromStringSlice(inst)
	if err != nil {
		return waitingContributions, deletedWaitingContributions, poolPairs, nil, err
	}
	matchContribution := matchAddLiquidityInst.Contribution()
	existedWaitingContribution, found := waitingContributions[matchContribution.PairHash()]
	if !found {
		err := fmt.Errorf("ERROR: could not find out existing waiting contribution with unique pair id: %s", matchContribution.PairHash())
		return waitingContributions, deletedWaitingContributions, poolPairs, nil, err
	}
	_, found = deletedWaitingContributions[matchContribution.PairHash()]
	if found {
		err := fmt.Errorf("Pair Hash %v has been existed in list deletedWaitingContributions", matchContribution.PairHash())
		return waitingContributions, deletedWaitingContributions, poolPairs, nil, err
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
	err = poolPair.addShare(
		existedWaitingContribution.NftID(),
		shareAmount, beaconHeight,
		existedWaitingContribution.TxReqID().String(),
	)

	if err != nil {
		return waitingContributions, deletedWaitingContributions, poolPairs, nil, err
	}
	poolPairs[poolPairID] = poolPair

	deletedWaitingContributions[matchContribution.PairHash()] = existedWaitingContribution
	delete(waitingContributions, matchContribution.PairHash())

	contribStatus := v2.ContributionStatus{
		Status: common.PDEContributionMatchedChainStatus,
	}
	contribStatusBytes, _ := json.Marshal(contribStatus)
	err = statedb.TrackPdexv3Status(
		stateDB,
		statedb.Pdexv3ContributionStatusPrefix(),
		matchContributionValue.TxReqID().Bytes(),
		contribStatusBytes,
	)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while tracking pde accepted contribution status: %+v", err)
		return waitingContributions, deletedWaitingContributions, poolPairs, nil, err
	}
	err = statedb.TrackPdexv3Status(
		stateDB,
		statedb.Pdexv3ContributionStatusPrefix(),
		existedWaitingContribution.TxReqID().Bytes(),
		contribStatusBytes,
	)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while tracking pde accepted contribution status: %+v", err)
		return waitingContributions, deletedWaitingContributions, poolPairs, nil, err
	}
	return waitingContributions, deletedWaitingContributions, poolPairs, &contribStatus, nil
}

func (sp *stateProcessorV2) matchAndReturnContribution(
	stateDB *statedb.StateDB,
	inst []string,
	beaconHeight uint64,
	waitingContributions map[string]rawdbv2.Pdexv3Contribution,
	deletedWaitingContributions map[string]rawdbv2.Pdexv3Contribution,
	poolPairs map[string]*PoolPairState,
) (
	map[string]rawdbv2.Pdexv3Contribution, map[string]rawdbv2.Pdexv3Contribution,
	map[string]*PoolPairState,
	*v2.ContributionStatus,
	error,
) {
	matchAndReturnAddLiquidity := instruction.MatchAndReturnAddLiquidity{}
	err := matchAndReturnAddLiquidity.FromStringSlice(inst)
	if err != nil {
		return waitingContributions, deletedWaitingContributions, poolPairs, nil, err
	}
	matchAndReturnContribution := matchAndReturnAddLiquidity.Contribution()
	matchAndReturnContributionValue := matchAndReturnContribution.Value()
	waitingContribution, found := waitingContributions[matchAndReturnContribution.PairHash()]
	var contribStatus v2.ContributionStatus
	if found {
		if matchAndReturnContributionValue.PoolPairID() != waitingContribution.PoolPairID() {
			err := fmt.Errorf("Expect poolPairID %v but get %v", waitingContribution.PoolPairID(), matchAndReturnContributionValue.PoolPairID())
			return waitingContributions, deletedWaitingContributions, poolPairs, nil, err
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
		err = poolPair.updateReserveData(amount0, amount1, matchAndReturnAddLiquidity.ShareAmount(), addOperator)
		if err != nil {
			return waitingContributions, deletedWaitingContributions, poolPairs, nil, err
		}
		err = poolPair.addShare(
			waitingContribution.NftID(),
			matchAndReturnAddLiquidity.ShareAmount(),
			beaconHeight,
			waitingContribution.TxReqID().String(),
		)
		if err != nil {
			return waitingContributions, deletedWaitingContributions, poolPairs, nil, err
		}
		sp.pairHashCache[matchAndReturnContribution.PairHash()] = matchAndReturnContributionValue.TxReqID().String()
		deletedWaitingContributions[matchAndReturnContribution.PairHash()] = waitingContribution
		delete(waitingContributions, matchAndReturnContribution.PairHash())
	} else {
		if matchAndReturnAddLiquidity.ExistedTokenID().String() < matchAndReturnContributionValue.TokenID().String() {
			contribStatus = v2.ContributionStatus{
				Status:                  common.PDEContributionMatchedNReturnedChainStatus,
				Token0ID:                matchAndReturnAddLiquidity.ExistedTokenID().String(),
				Token0ContributedAmount: matchAndReturnAddLiquidity.ExistedTokenActualAmount(),
				Token0ReturnedAmount:    matchAndReturnAddLiquidity.ExistedTokenReturnAmount(),
				Token1ID:                matchAndReturnContributionValue.TokenID().String(),
				Token1ContributedAmount: matchAndReturnContributionValue.Amount() - matchAndReturnAddLiquidity.ReturnAmount(),
				Token1ReturnedAmount:    matchAndReturnAddLiquidity.ReturnAmount(),
			}
		} else {
			contribStatus = v2.ContributionStatus{
				Status:                  common.PDEContributionMatchedNReturnedChainStatus,
				Token1ID:                matchAndReturnAddLiquidity.ExistedTokenID().String(),
				Token1ContributedAmount: matchAndReturnAddLiquidity.ExistedTokenActualAmount(),
				Token1ReturnedAmount:    matchAndReturnAddLiquidity.ExistedTokenReturnAmount(),
				Token0ID:                matchAndReturnContributionValue.TokenID().String(),
				Token0ContributedAmount: matchAndReturnContributionValue.Amount() - matchAndReturnAddLiquidity.ReturnAmount(),
				Token0ReturnedAmount:    matchAndReturnAddLiquidity.ReturnAmount(),
			}
		}

		contribStatusBytes, _ := json.Marshal(contribStatus)
		err = statedb.TrackPdexv3Status(
			stateDB,
			statedb.Pdexv3ContributionStatusPrefix(),
			matchAndReturnContributionValue.TxReqID().Bytes(),
			contribStatusBytes,
		)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while tracking pde contribution status: %+v", err)
			return waitingContributions, deletedWaitingContributions, poolPairs, nil, err
		}
		err = statedb.TrackPdexv3Status(
			stateDB,
			statedb.Pdexv3ContributionStatusPrefix(),
			[]byte(sp.pairHashCache[matchAndReturnContribution.PairHash()]),
			contribStatusBytes,
		)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while tracking pde contribution status: %+v", err)
			return waitingContributions, deletedWaitingContributions, poolPairs, nil, err
		}
	}

	return waitingContributions, deletedWaitingContributions, poolPairs, &contribStatus, nil
}

func (sp *stateProcessorV2) modifyParams(
	stateDB *statedb.StateDB,
	beaconHeight uint64,
	inst []string,
	params Params,
	stakingPoolStates map[string]*StakingPoolState,
) (Params, map[string]*StakingPoolState, error) {
	if len(inst) != 4 {
		msg := fmt.Sprintf("Length of instruction is not valid expect %v but get %v", 4, len(inst))
		Logger.log.Errorf(msg)
		return params, stakingPoolStates, errors.New(msg)
	}

	// unmarshal instructions content
	var actionData metadataPdexv3.ParamsModifyingContent
	err := json.Unmarshal([]byte(inst[3]), &actionData)
	if err != nil {
		msg := fmt.Sprintf("Could not unmarshal instruction content %v - Error: %v\n", inst[3], err)
		Logger.log.Errorf(msg)
		return params, stakingPoolStates, err
	}

	modifyingStatus := inst[2]
	var reqTrackStatus int
	if modifyingStatus == metadataPdexv3.RequestAcceptedChainStatus {
		params = Params(actionData.Content)
		reqTrackStatus = metadataPdexv3.ParamsModifyingSuccessStatus
		stakingPoolStates = addStakingPoolState(stakingPoolStates, params.StakingPoolsShare)
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

	return params, stakingPoolStates, nil
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
		currentTrade = &instruction.Action{Content: &metadataPdexv3.AcceptedTrade{}}
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
			// write changes to state
			pairs[pairID] = pair
		}

		trackedStatus = metadataPdexv3.TradeStatus{
			BuyAmount:  md.Amount,
			TokenToBuy: md.TokenToBuy,
		}
	case strconv.Itoa(metadataPdexv3.TradeRefundedStatus):
		currentTrade = &instruction.Action{Content: &metadataPdexv3.RefundedTrade{}}
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
	beaconHeight uint64,
) (map[string]*PoolPairState, error) {
	var err error
	switch inst[1] {
	case common.PDEWithdrawalRejectedChainStatus:
		_, err = sp.rejectWithdrawLiquidity(stateDB, inst)
	case common.PDEWithdrawalAcceptedChainStatus:
		poolPairs, _, err = sp.acceptWithdrawLiquidity(stateDB, inst, poolPairs, beaconHeight)
	}
	if err != nil {
		return poolPairs, err
	}
	return poolPairs, err
}

func (sp *stateProcessorV2) rejectWithdrawLiquidity(
	stateDB *statedb.StateDB, inst []string,
) (*v2.WithdrawStatus, error) {
	rejectWithdrawLiquidity := instruction.NewRejectWithdrawLiquidity()
	err := rejectWithdrawLiquidity.FromStringSlice(inst)
	if err != nil {
		return nil, err
	}
	withdrawStatus := v2.WithdrawStatus{
		Status: common.PDEWithdrawalRejectedChainStatus,
	}
	contentBytes, _ := json.Marshal(withdrawStatus)
	err = statedb.TrackPdexv3Status(
		stateDB,
		statedb.Pdexv3WithdrawLiquidityStatusPrefix(),
		rejectWithdrawLiquidity.TxReqID().Bytes(),
		contentBytes,
	)
	return &withdrawStatus, err
}

func (sp *stateProcessorV2) acceptWithdrawLiquidity(
	stateDB *statedb.StateDB, inst []string,
	poolPairs map[string]*PoolPairState,
	beaconHeight uint64,
) (map[string]*PoolPairState, *v2.WithdrawStatus, error) {
	acceptWithdrawLiquidity := instruction.NewAcceptWithdrawLiquidity()
	err := acceptWithdrawLiquidity.FromStringSlice(inst)
	if err != nil {
		return poolPairs, nil, err
	}
	poolPair, ok := poolPairs[acceptWithdrawLiquidity.PoolPairID()]
	if !ok || poolPair == nil {
		err := fmt.Errorf("Can't find poolPairID %s", acceptWithdrawLiquidity.PoolPairID())
		return poolPairs, nil, err
	}
	share, ok := poolPair.shares[acceptWithdrawLiquidity.NftID().String()]
	if !ok || share == nil {
		err := fmt.Errorf("Can't find nftID %s", acceptWithdrawLiquidity.NftID().String())
		return poolPairs, nil, err
	}
	poolPair.updateSingleTokenAmount(
		acceptWithdrawLiquidity.TokenID(),
		acceptWithdrawLiquidity.TokenAmount(), acceptWithdrawLiquidity.ShareAmount(), subOperator,
	)
	token0Amount, found := sp.withdrawTxCache[acceptWithdrawLiquidity.TxReqID().String()]
	if !found {
		sp.withdrawTxCache[acceptWithdrawLiquidity.TxReqID().String()] = acceptWithdrawLiquidity.TokenAmount()
	}
	var withdrawStatus *v2.WithdrawStatus
	if poolPair.state.Token1ID().String() == acceptWithdrawLiquidity.TokenID().String() {
		err = poolPair.updateShareValue(
			acceptWithdrawLiquidity.ShareAmount(), beaconHeight,
			acceptWithdrawLiquidity.NftID().String(), subOperator)
		if err != nil {
			return poolPairs, nil, err
		}
		withdrawStatus = &v2.WithdrawStatus{
			Status:       common.PDEWithdrawalAcceptedChainStatus,
			Token0ID:     poolPair.state.Token0ID().String(),
			Token0Amount: token0Amount,
			Token1ID:     poolPair.state.Token1ID().String(),
			Token1Amount: acceptWithdrawLiquidity.TokenAmount(),
		}
		contentBytes, _ := json.Marshal(withdrawStatus)
		err = statedb.TrackPdexv3Status(
			stateDB,
			statedb.Pdexv3WithdrawLiquidityStatusPrefix(),
			acceptWithdrawLiquidity.TxReqID().Bytes(),
			contentBytes,
		)
	}
	return poolPairs, withdrawStatus, err
}

func (sp *stateProcessorV2) addOrder(
	stateDB *statedb.StateDB,
	inst []string,
	pairs map[string]*PoolPairState,
) (map[string]*PoolPairState, error) {
	var currentOrder *instruction.Action
	var trackedStatus metadataPdexv3.AddOrderStatus
	switch inst[1] {
	case strconv.Itoa(metadataPdexv3.OrderAcceptedStatus):
		currentOrder = &instruction.Action{Content: &metadataPdexv3.AcceptedAddOrder{}}
		err := currentOrder.FromStringSlice(inst)
		if err != nil {
			return pairs, err
		}

		// skip error checking since concrete type is specified above
		md, _ := currentOrder.Content.(*metadataPdexv3.AcceptedAddOrder)
		trackedStatus.OrderID = md.OrderID

		pair, exists := pairs[md.PoolPairID]
		if !exists {
			return pairs, fmt.Errorf("Cannot find pair %s for new order", md.PoolPairID)
		}

		// fee for this request is deducted right away, while the fee stored in the order itself
		// starts from 0 and will accumulate over time
		newOrder := rawdbv2.NewPdexv3OrderWithValue(md.OrderID, md.NftID, md.Token0Rate, md.Token1Rate,
			md.Token0Balance, md.Token1Balance, md.TradeDirection, 0)
		pair.orderbook.InsertOrder(newOrder)
		// write changes to state
		pairs[md.PoolPairID] = pair
	case strconv.Itoa(metadataPdexv3.OrderRefundedStatus):
		currentOrder = &instruction.Action{Content: &metadataPdexv3.RefundedAddOrder{}}
		err := currentOrder.FromStringSlice(inst)
		if err != nil {
			return pairs, err
		}
	default:
		return pairs, fmt.Errorf("Invalid status %s from instruction", inst[1])
	}

	// store tracked order status
	trackedStatus.Status = currentOrder.GetStatus()
	marshaledTrackedStatus, err := json.Marshal(trackedStatus)
	if err != nil {
		return pairs, err
	}
	err = statedb.TrackPdexv3Status(
		stateDB,
		statedb.Pdexv3AddOrderStatusPrefix(),
		currentOrder.RequestTxID[:],
		marshaledTrackedStatus,
	)
	return pairs, nil
}

func (sp *stateProcessorV2) withdrawOrder(
	stateDB *statedb.StateDB,
	inst []string,
	pairs map[string]*PoolPairState,
) (map[string]*PoolPairState, error) {
	var currentOrder *instruction.Action
	var trackedStatus metadataPdexv3.WithdrawOrderStatus
	switch inst[1] {
	case strconv.Itoa(metadataPdexv3.WithdrawOrderAcceptedStatus):
		currentOrder = &instruction.Action{Content: &metadataPdexv3.AcceptedWithdrawOrder{}}
		err := currentOrder.FromStringSlice(inst)
		if err != nil {
			return pairs, err
		}

		// skip error checking since concrete type is specified above
		md, _ := currentOrder.Content.(*metadataPdexv3.AcceptedWithdrawOrder)
		trackedStatus.TokenID = md.TokenID
		trackedStatus.WithdrawAmount = md.Amount

		pair, exists := pairs[md.PoolPairID]
		if !exists {
			return pairs, fmt.Errorf("Cannot find pair %s for new order", md.PoolPairID)
		}

		for index, ord := range pair.orderbook.orders {
			if ord.Id() == md.OrderID {
				if md.TokenID == pair.state.Token0ID() {
					newBalance := ord.Token0Balance() - md.Amount
					if newBalance > ord.Token0Balance() {
						return pairs, fmt.Errorf("Cannot withdraw more than current token0 balance from order %s",
							md.OrderID)
					}
					ord.SetToken0Balance(newBalance)
					// remove order when both balances are cleared
					if newBalance == 0 && ord.Token1Balance() == 0 {
						pair.orderbook.RemoveOrder(index)
					}
				} else if md.TokenID == pair.state.Token1ID() {
					newBalance := ord.Token1Balance() - md.Amount
					if newBalance > ord.Token1Balance() {
						return pairs, fmt.Errorf("Cannot withdraw more than current token1 balance from order %s",
							md.OrderID)
					}
					ord.SetToken1Balance(newBalance)
					// remove order when both balances are cleared
					if newBalance == 0 && ord.Token0Balance() == 0 {
						pair.orderbook.RemoveOrder(index)
					}
				}
			}
		}

		// write changes to state
		pairs[md.PoolPairID] = pair
	case strconv.Itoa(metadataPdexv3.WithdrawOrderRejectedStatus):
		currentOrder = &instruction.Action{Content: &metadataPdexv3.RejectedWithdrawOrder{}}
		err := currentOrder.FromStringSlice(inst)
		if err != nil {
			return pairs, err
		}
	default:
		return pairs, fmt.Errorf("Invalid status %s from instruction", inst[1])
	}

	// store tracked order status
	trackedStatus.Status = currentOrder.GetStatus()
	marshaledTrackedStatus, err := json.Marshal(trackedStatus)
	if err != nil {
		return pairs, err
	}
	err = statedb.TrackPdexv3Status(
		stateDB,
		statedb.Pdexv3WithdrawOrderStatusPrefix(),
		currentOrder.RequestTxID[:],
		marshaledTrackedStatus,
	)
	return pairs, nil
}

func (sp *stateProcessorV2) userMintNft(
	stateDB *statedb.StateDB, inst []string, nftIDs map[string]uint64,
) (map[string]uint64, *v2.MintNftStatus, error) {
	if len(inst) != 3 {
		return nftIDs, nil, fmt.Errorf("Expect length of instruction is %v but get %v", 3, len(inst))
	}
	status := utils.EmptyString
	nftID := utils.EmptyString
	txReqID := common.Hash{}
	var burntAmount uint64
	if inst[0] != strconv.Itoa(metadataCommon.Pdexv3UserMintNftRequestMeta) {
		return nftIDs, nil, fmt.Errorf("Expect metaType is %v but get %s", metadataCommon.Pdexv3UserMintNftRequestMeta, inst[1])
	}
	switch inst[1] {
	case common.Pdexv3RejectUserMintNftStatus:
		status = common.Pdexv3RejectUserMintNftStatus
		refundInst := instruction.NewRejectUserMintNft()
		err := refundInst.FromStringSlice(inst)
		if err != nil {
			return nftIDs, nil, err
		}
		burntAmount = refundInst.Amount()
		txReqID = refundInst.TxReqID()
	case common.Pdexv3AcceptUserMintNftStatus:
		status = common.Pdexv3AcceptUserMintNftStatus
		acceptInst := instruction.NewAcceptUserMintNft()
		err := acceptInst.FromStringSlice(inst)
		if err != nil {
			return nftIDs, nil, err
		}
		nftID = acceptInst.NftID().String()
		burntAmount = acceptInst.BurntAmount()
		nftIDs[acceptInst.NftID().String()] = acceptInst.BurntAmount()
		txReqID = acceptInst.TxReqID()
	default:
		return nftIDs, nil, errors.New("Can not recognize status")
	}

	mintNftStatus := v2.MintNftStatus{
		NftID:       nftID,
		Status:      status,
		BurntAmount: burntAmount,
	}
	data, err := json.Marshal(mintNftStatus)
	if err != nil {
		return nftIDs, nil, err
	}

	err = statedb.TrackPdexv3Status(
		stateDB,
		statedb.Pdexv3UserMintNftStatusPrefix(),
		txReqID.Bytes(),
		data,
	)
	return nftIDs, &mintNftStatus, nil
}

func (sp *stateProcessorV2) staking(
	stateDB *statedb.StateDB,
	inst []string, nftIDs map[string]uint64, stakingPoolStates map[string]*StakingPoolState,
	beaconHeight uint64,
) (map[string]*StakingPoolState, *v2.StakingStatus, error) {
	if len(inst) < 2 {
		return stakingPoolStates, nil, fmt.Errorf("Length of inst is invalid %v", len(inst))
	}
	var status, nftID, stakingPoolID string
	var txReqID common.Hash
	var liquidity uint64
	switch inst[1] {
	case common.Pdexv3AcceptStakingStatus:
		acceptInst := instruction.NewAcceptStaking()
		err := acceptInst.FromStringSlice(inst)
		if err != nil {
			return stakingPoolStates, nil, err
		}
		txReqID = acceptInst.TxReqID()
		status = common.Pdexv3AcceptStakingStatus
		stakingPoolID = acceptInst.StakingPoolID().String()
		liquidity = acceptInst.Liquidity()
		nftID = acceptInst.NftID().String()
		stakingPoolState := stakingPoolStates[stakingPoolID]
		err = stakingPoolState.updateLiquidity(nftID, liquidity, beaconHeight, addOperator)
		if err != nil {
			return stakingPoolStates, nil, err
		}
	case common.Pdexv3RejectStakingStatus:
		rejectInst := instruction.NewRejectStaking()
		err := rejectInst.FromStringSlice(inst)
		if err != nil {
			return stakingPoolStates, nil, err
		}
		txReqID = rejectInst.TxReqID()
		status = common.Pdexv3RejectStakingStatus
		stakingPoolID = rejectInst.TokenID().String()
		liquidity = rejectInst.Amount()
	}
	stakingStatus := v2.StakingStatus{
		Status:        status,
		NftID:         nftID,
		StakingPoolID: stakingPoolID,
		Liquidity:     liquidity,
	}
	data, err := json.Marshal(stakingStatus)
	if err != nil {
		return stakingPoolStates, nil, err
	}
	err = statedb.TrackPdexv3Status(
		stateDB,
		statedb.Pdexv3StakingStatusPrefix(),
		txReqID.Bytes(),
		data,
	)
	return stakingPoolStates, &stakingStatus, nil
}

func (sp *stateProcessorV2) unstaking(
	stateDB *statedb.StateDB,
	inst []string, nftIDs map[string]uint64, stakingPoolStates map[string]*StakingPoolState,
	beaconHeight uint64,
) (map[string]*StakingPoolState, *v2.UnstakingStatus, error) {
	if len(inst) < 2 {
		return stakingPoolStates, nil, fmt.Errorf("Length of inst is invalid %v", len(inst))
	}
	var status, nftID, stakingPoolID string
	var txReqID common.Hash
	var liquidity uint64
	switch inst[1] {
	case common.Pdexv3AcceptUnstakingStatus:
		acceptInst := instruction.NewAcceptUnstaking()
		err := acceptInst.FromStringSlice(inst)
		if err != nil {
			return stakingPoolStates, nil, err
		}
		txReqID = acceptInst.TxReqID()
		status = common.Pdexv3AcceptStakingStatus
		stakingPoolID = acceptInst.StakingPoolID().String()
		liquidity = acceptInst.Amount()
		nftID = acceptInst.NftID().String()
		stakingPoolState := stakingPoolStates[stakingPoolID]
		err = stakingPoolState.updateLiquidity(nftID, liquidity, beaconHeight, addOperator)
		if err != nil {
			return stakingPoolStates, nil, err
		}
	case common.Pdexv3RejectUnstakingStatus:
		rejectInst := instruction.NewRejectUnstaking()
		err := rejectInst.FromStringSlice(inst)
		if err != nil {
			return stakingPoolStates, nil, err
		}
		txReqID = rejectInst.TxReqID()
		status = common.Pdexv3RejectStakingStatus
	}
	unstakingStatus := v2.UnstakingStatus{
		Status:        status,
		NftID:         nftID,
		StakingPoolID: stakingPoolID,
		Liquidity:     liquidity,
	}
	data, err := json.Marshal(unstakingStatus)
	if err != nil {
		return stakingPoolStates, nil, err
	}
	err = statedb.TrackPdexv3Status(
		stateDB,
		statedb.Pdexv3UnstakingStatusPrefix(),
		txReqID.Bytes(),
		data,
	)
	return stakingPoolStates, &unstakingStatus, nil
}
