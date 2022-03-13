package pdex

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strconv"

	"github.com/incognitochain/incognito-chain/blockchain/pdex/v2utils"
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
	pairHashCache   map[string]common.Hash
	withdrawTxCache map[string]uint64
	rewardCache     map[string]map[common.Hash]uint64
	receiverCache   map[string]map[common.Hash]metadataPdexv3.ReceiverInfo
	stateProcessorBase
}

func (sp *stateProcessorV2) clearCache() {
	sp.pairHashCache = make(map[string]common.Hash)
	sp.withdrawTxCache = make(map[string]uint64)
	sp.rewardCache = make(map[string]map[common.Hash]uint64)
	sp.receiverCache = make(map[string]map[common.Hash]metadataPdexv3.ReceiverInfo)
}

func (sp *stateProcessorV2) addLiquidity(
	stateDB *statedb.StateDB,
	inst []string,
	beaconHeight uint64,
	poolPairs map[string]*PoolPairState,
	waitingContributions map[string]rawdbv2.Pdexv3Contribution,
	deletedWaitingContributions map[string]rawdbv2.Pdexv3Contribution,
	params *Params,
) (
	map[string]*PoolPairState,
	map[string]rawdbv2.Pdexv3Contribution, map[string]rawdbv2.Pdexv3Contribution, error,
) {
	var err error
	switch inst[1] {
	case common.PDEContributionWaitingChainStatus:
		waitingContributions, _, err = sp.waitingContribution(stateDB, inst, waitingContributions)
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
			stateDB, inst, beaconHeight, waitingContributions, deletedWaitingContributions, poolPairs, params)
		if err != nil {
			return poolPairs, waitingContributions, deletedWaitingContributions, err
		}
	case common.PDEContributionMatchedNReturnedChainStatus:
		waitingContributions,
			deletedWaitingContributions, poolPairs, _, err = sp.matchAndReturnContribution(
			stateDB, inst, beaconHeight,
			waitingContributions, deletedWaitingContributions, poolPairs, params)
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
) (map[string]rawdbv2.Pdexv3Contribution, *v2.ContributionStatus, error) {
	waitingAddLiquidityInst := instruction.WaitingAddLiquidity{}
	err := waitingAddLiquidityInst.FromStringSlice(inst)
	if err != nil {
		return waitingContributions, nil, err
	}
	err = sp.verifyWaitingContribution(waitingAddLiquidityInst.Contribution(), waitingContributions)
	if err != nil {
		return waitingContributions, nil, err
	}
	contribution := waitingAddLiquidityInst.Contribution()
	contributionValue := contribution.Value()
	waitingContributions[contribution.PairHash()] = contributionValue

	contribStatus := v2.ContributionStatus{
		Token0ID:                contributionValue.TokenID().String(),
		Token0ContributedAmount: contributionValue.Amount(),
		Status:                  common.PDEContributionWaitingStatus,
		PoolPairID:              contributionValue.PoolPairID(),
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
) error {
	_, found := waitingContributions[contribution.PairHash()]
	if found {
		err := fmt.Errorf("Pair Hash %v has been existed in list waitingContributions", contribution.PairHash())
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
		Status:     common.PDEContributionRefundStatus,
		PoolPairID: refundContributionValue.PoolPairID(),
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
	params *Params,
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
	lmLockedBlocks := uint64(0)
	if _, exists := params.PDEXRewardPoolPairsShare[poolPairID]; exists {
		lmLockedBlocks = params.MiningRewardPendingBlocks
	}
	err = poolPair.addShare(
		existedWaitingContribution.NftID(),
		shareAmount,
		beaconHeight, lmLockedBlocks,
	)

	if err != nil {
		return waitingContributions, deletedWaitingContributions, poolPairs, nil, err
	}
	poolPairs[poolPairID] = poolPair

	deletedWaitingContributions[matchContribution.PairHash()] = existedWaitingContribution
	delete(waitingContributions, matchContribution.PairHash())

	contribStatus := v2.ContributionStatus{
		Status:     common.PDEContributionAcceptedStatus,
		PoolPairID: matchContributionValue.PoolPairID(),
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
	params *Params,
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
			amount1 = matchAndReturnContributionValue.Amount() - matchAndReturnAddLiquidity.ReturnAmount()
		} else {
			amount1 = matchAndReturnAddLiquidity.ExistedTokenActualAmount()
			amount0 = matchAndReturnContributionValue.Amount() - matchAndReturnAddLiquidity.ReturnAmount()
		}
		err = poolPair.updateReserveData(amount0, amount1, matchAndReturnAddLiquidity.ShareAmount(), addOperator)
		if err != nil {
			return waitingContributions, deletedWaitingContributions, poolPairs, nil, err
		}
		lmLockedBlocks := uint64(0)
		if _, exists := params.PDEXRewardPoolPairsShare[waitingContribution.PoolPairID()]; exists {
			lmLockedBlocks = params.MiningRewardPendingBlocks
		}
		err = poolPair.addShare(
			waitingContribution.NftID(),
			matchAndReturnAddLiquidity.ShareAmount(),
			beaconHeight,
			lmLockedBlocks,
		)
		if err != nil {
			return waitingContributions, deletedWaitingContributions, poolPairs, nil, err
		}
		sp.pairHashCache[matchAndReturnContribution.PairHash()] = matchAndReturnContributionValue.TxReqID()
		deletedWaitingContributions[matchAndReturnContribution.PairHash()] = waitingContribution
		delete(waitingContributions, matchAndReturnContribution.PairHash())
	} else {
		if matchAndReturnAddLiquidity.ExistedTokenID().String() < matchAndReturnContributionValue.TokenID().String() {
			contribStatus = v2.ContributionStatus{
				Status:                  common.PDEContributionMatchedNReturnedStatus,
				Token0ID:                matchAndReturnAddLiquidity.ExistedTokenID().String(),
				Token0ContributedAmount: matchAndReturnAddLiquidity.ExistedTokenActualAmount(),
				Token0ReturnedAmount:    matchAndReturnAddLiquidity.ExistedTokenReturnAmount(),
				Token1ID:                matchAndReturnContributionValue.TokenID().String(),
				Token1ContributedAmount: matchAndReturnContributionValue.Amount() - matchAndReturnAddLiquidity.ReturnAmount(),
				Token1ReturnedAmount:    matchAndReturnAddLiquidity.ReturnAmount(),
				PoolPairID:              matchAndReturnContributionValue.PoolPairID(),
			}
		} else {
			contribStatus = v2.ContributionStatus{
				Status:                  common.PDEContributionMatchedNReturnedStatus,
				Token1ID:                matchAndReturnAddLiquidity.ExistedTokenID().String(),
				Token1ContributedAmount: matchAndReturnAddLiquidity.ExistedTokenActualAmount(),
				Token1ReturnedAmount:    matchAndReturnAddLiquidity.ExistedTokenReturnAmount(),
				Token0ID:                matchAndReturnContributionValue.TokenID().String(),
				Token0ContributedAmount: matchAndReturnContributionValue.Amount() - matchAndReturnAddLiquidity.ReturnAmount(),
				Token0ReturnedAmount:    matchAndReturnAddLiquidity.ReturnAmount(),
				PoolPairID:              matchAndReturnContributionValue.PoolPairID(),
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
			sp.pairHashCache[matchAndReturnContribution.PairHash()].Bytes(),
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
	inst []string,
	params *Params,
	stakingPoolStates map[string]*StakingPoolState,
) (*Params, map[string]*StakingPoolState, error) {
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
		*params = Params(actionData.Content)
		reqTrackStatus = metadataPdexv3.ParamsModifyingSuccessStatus
		stakingPoolStates = addStakingPoolState(stakingPoolStates, params.StakingPoolsShare)
	} else {
		reqTrackStatus = metadataPdexv3.ParamsModifyingFailedStatus
	}

	modifyingReqStatus := metadataPdexv3.ParamsModifyingRequestStatus{
		Status:       reqTrackStatus,
		ErrorMsg:     actionData.ErrorMsg,
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
		Logger.log.Errorf("PDex Params Modifying: An error occurred while tracking request tx - Error: %v", err)
	}

	return params, stakingPoolStates, nil
}

func (sp *stateProcessorV2) trade(
	stateDB *statedb.StateDB,
	inst []string,
	pairs map[string]*PoolPairState,
	params *Params,
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
			reserveState := v2utils.NewTradingPairWithValue(
				&pair.state,
			)
			err := reserveState.ApplyReserveChanges(md.PairChanges[index][0], md.PairChanges[index][1])
			if err != nil {
				return pairs, err
			}

			for tokenID, amount := range md.RewardEarned[index] {
				// split reward between LPs and LOPs by weighted ratio
				ratio := params.DefaultOrderTradingRewardRatioBPS
				if params.OrderTradingRewardRatioBPS != nil {
					bps, ok := params.OrderTradingRewardRatioBPS[pairID]
					if ok {
						ratio = bps
					}
				}

				remain := new(big.Int).SetUint64(0)

				// add staking pools and protocol fees
				pair.protocolFees, pair.stakingPoolFees, remain = reserveState.AddStakingAndProtocolFee(
					tokenID, new(big.Int).SetUint64(amount), pair.protocolFees, pair.stakingPoolFees,
					params.TradingProtocolFeePercent, params.TradingStakingPoolRewardPercent, params.StakingRewardTokens,
				)

				ammMakingVolume, orderMakingVolumes, tradeDirection := v2utils.GetMakingVolumes(
					md.PairChanges[index], md.OrderChanges[index],
					pair.orderbook.NftIDs(),
				)

				ammReward, orderRewards := v2.SplitTradingReward(
					remain, ratio, BPS,
					ammMakingVolume, orderMakingVolumes,
				)

				makingToken := reserveState.Token0ID()
				if tradeDirection == v2.TradeDirectionSell0 {
					makingToken = reserveState.Token1ID()
				}

				// add volume to LOPs
				if _, ok := params.PDEXRewardPoolPairsShare[pairID]; ok && params.DAOContributingPercent > 0 {
					if _, ok := pair.makingVolume[makingToken]; !ok {
						pair.makingVolume[makingToken] = NewMakingVolume()
					}
					for nftID, amount := range orderMakingVolumes {
						pair.makingVolume[makingToken].AddVolume(nftID, amount)
					}
				}

				// add reward to LOPs
				for nftID, reward := range orderRewards {
					if _, ok := pair.orderRewards[nftID]; !ok {
						pair.orderRewards[nftID] = NewOrderReward()
					}
					pair.orderRewards[nftID].AddReward(tokenID, reward)
				}

				// add reward to LPs
				pair.lpFeesPerShare = reserveState.AddLPFee(
					tokenID, new(big.Int).SetUint64(ammReward), BaseLPFeesPerShare,
					pair.lpFeesPerShare,
				)
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
	txID := currentTrade.RequestTxID()
	err = statedb.TrackPdexv3Status(
		stateDB,
		statedb.Pdexv3TradeStatusPrefix(),
		txID[:],
		marshaledTrackedStatus,
	)
	return pairs, err
}

func (sp *stateProcessorV2) withdrawLiquidity(
	stateDB *statedb.StateDB,
	inst []string,
	poolPairs map[string]*PoolPairState,
	lmLockedBlocks uint64,
) (map[string]*PoolPairState, error) {
	var err error
	switch inst[1] {
	case common.PDEWithdrawalRejectedChainStatus:
		_, err = sp.rejectWithdrawLiquidity(stateDB, inst)
	case common.PDEWithdrawalAcceptedChainStatus:
		poolPairs, _, err = sp.acceptWithdrawLiquidity(stateDB, inst, poolPairs, lmLockedBlocks)
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
		Status: common.Pdexv3RejectStatus,
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
	stateDB *statedb.StateDB,
	inst []string,
	poolPairs map[string]*PoolPairState,
	lmLockedBlocks uint64,
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
			acceptWithdrawLiquidity.ShareAmount(),
			acceptWithdrawLiquidity.NftID().String(),
			subOperator,
			0,
			0)
		if err != nil {
			return poolPairs, nil, err
		}
		withdrawStatus = &v2.WithdrawStatus{
			Status:       common.Pdexv3AcceptStatus,
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
			md.Token0Balance, md.Token1Balance, md.TradeDirection, md.Receiver)
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
	txID := currentOrder.RequestTxID()
	err = statedb.TrackPdexv3Status(
		stateDB,
		statedb.Pdexv3AddOrderStatusPrefix(),
		txID[:],
		marshaledTrackedStatus,
	)
	return pairs, err
}

func (sp *stateProcessorV2) withdrawOrder(
	stateDB *statedb.StateDB,
	inst []string,
	pairs map[string]*PoolPairState,
) (map[string]*PoolPairState, error) {
	var currentOrder *instruction.Action
	var trackedStatus metadataPdexv3.WithdrawOrderStatus
	var txID common.Hash
	suffixWithToken := []byte{}
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
		txID = currentOrder.RequestTxID()
		suffixWithToken = append(txID[:], md.TokenID[:]...)

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
		txID = currentOrder.RequestTxID()
	default:
		return pairs, fmt.Errorf("Invalid status %s from instruction", inst[1])
	}

	// store tracked order status
	trackedStatus.Status = currentOrder.GetStatus()
	marshaledTrackedStatus, err := json.Marshal(trackedStatus)
	if err != nil {
		return pairs, err
	}

	// store accepted / rejected status
	err = statedb.TrackPdexv3Status(
		stateDB,
		statedb.Pdexv3WithdrawOrderStatusPrefix(),
		txID[:],
		marshaledTrackedStatus,
	)

	// store withdrawal info (tokenID & amount) specific to this instruction
	if len(suffixWithToken) > 0 {
		err := statedb.TrackPdexv3Status(
			stateDB,
			statedb.Pdexv3WithdrawOrderStatusPrefix(),
			suffixWithToken,
			marshaledTrackedStatus,
		)
		if err != nil {
			return pairs, err
		}
	}

	return pairs, err
}

func (sp *stateProcessorV2) withdrawLPFee(
	stateDB *statedb.StateDB,
	inst []string,
	pairs map[string]*PoolPairState,
) (map[string]*PoolPairState, error) {
	if len(inst) != 4 {
		msg := fmt.Sprintf("Length of instruction is not valid expect %v but get %v", 4, len(inst))
		Logger.log.Errorf(msg)
		return pairs, errors.New(msg)
	}

	// unmarshal instructions content
	var actionData metadataPdexv3.WithdrawalLPFeeContent
	err := json.Unmarshal([]byte(inst[3]), &actionData)
	if err != nil {
		msg := fmt.Sprintf("Could not unmarshal instruction content %v - Error: %v\n", inst[3], err)
		Logger.log.Errorf(msg)
		return pairs, err
	}

	withdrawalStatus := inst[2]
	var reqTrackStatus int
	if withdrawalStatus == metadataPdexv3.RequestAcceptedChainStatus {
		// check conditions
		poolPair, isExisted := pairs[actionData.PoolPairID]
		if !isExisted {
			msg := fmt.Sprintf("Could not find pair %s for withdrawal", actionData.PoolPairID)
			Logger.log.Errorf(msg)
			return pairs, errors.New(msg)
		}

		share, isExisted := poolPair.shares[actionData.NftID.String()]
		if isExisted {
			// update state after fee withdrawal
			share.tradingFees = resetKeyValueToZero(share.tradingFees)
			share.lastLPFeesPerShare = poolPair.LpFeesPerShare()
			share.lastLmRewardsPerShare = poolPair.LmRewardsPerShare()
		}

		_, isExisted = poolPair.orderRewards[actionData.NftID.String()]
		if isExisted {
			delete(poolPair.orderRewards, actionData.NftID.String())
		}

		reqTrackStatus = metadataPdexv3.WithdrawLPFeeSuccessStatus

		_, found := sp.receiverCache[actionData.TxReqID.String()]
		if !found {
			sp.receiverCache[actionData.TxReqID.String()] = map[common.Hash]metadataPdexv3.ReceiverInfo{}
		}
		sp.receiverCache[actionData.TxReqID.String()][actionData.TokenID] = actionData.Receiver
	} else {
		reqTrackStatus = metadataPdexv3.WithdrawLPFeeFailedStatus
	}

	if reqTrackStatus == metadataPdexv3.WithdrawProtocolFeeSuccessStatus && !actionData.IsLastInst {
		return pairs, nil
	}

	withdrawalReqStatus := metadataPdexv3.WithdrawalLPFeeStatus{
		Status:    reqTrackStatus,
		Receivers: sp.receiverCache[actionData.TxReqID.String()],
	}
	withdrawalReqStatusBytes, _ := json.Marshal(withdrawalReqStatus)

	err = statedb.TrackPdexv3Status(
		stateDB,
		statedb.Pdexv3WithdrawalLPFeeStatusPrefix(),
		[]byte(actionData.TxReqID.String()),
		withdrawalReqStatusBytes,
	)
	if err != nil {
		Logger.log.Errorf("PDex v3 Withdrawal LP Fee: An error occurred while tracking request tx - Error: %v", err)
	}
	return pairs, err
}

func (sp *stateProcessorV2) withdrawProtocolFee(
	stateDB *statedb.StateDB,
	inst []string,
	pairs map[string]*PoolPairState,
) (map[string]*PoolPairState, error) {
	if len(inst) != 4 {
		msg := fmt.Sprintf("Length of instruction is not valid expect %v but get %v", 4, len(inst))
		Logger.log.Errorf(msg)
		return pairs, errors.New(msg)
	}

	// unmarshal instructions content
	var actionData metadataPdexv3.WithdrawalProtocolFeeContent
	err := json.Unmarshal([]byte(inst[3]), &actionData)
	if err != nil {
		msg := fmt.Sprintf("Could not unmarshal instruction content %v - Error: %v\n", inst[3], err)
		Logger.log.Errorf(msg)
		return pairs, err
	}

	withdrawalStatus := inst[2]
	var reqTrackStatus int
	if withdrawalStatus == metadataPdexv3.RequestAcceptedChainStatus {
		// check conditions
		poolPair, isExisted := pairs[actionData.PoolPairID]
		if !isExisted {
			msg := fmt.Sprintf("Could not find pair %s for withdrawal", actionData.PoolPairID)
			Logger.log.Errorf(msg)
			return pairs, errors.New(msg)
		}

		// update state after fee withdrawal
		poolPair.protocolFees = resetKeyValueToZero(poolPair.protocolFees)
		reqTrackStatus = metadataPdexv3.WithdrawProtocolFeeSuccessStatus

		_, found := sp.rewardCache[actionData.TxReqID.String()]
		if !found {
			sp.rewardCache[actionData.TxReqID.String()] = map[common.Hash]uint64{}
		}
		sp.rewardCache[actionData.TxReqID.String()][actionData.TokenID] = actionData.Amount
	} else {
		reqTrackStatus = metadataPdexv3.WithdrawProtocolFeeFailedStatus
	}

	if reqTrackStatus == metadataPdexv3.WithdrawProtocolFeeSuccessStatus && !actionData.IsLastInst {
		return pairs, nil
	}

	withdrawalReqStatus := metadataPdexv3.WithdrawalProtocolFeeStatus{
		Status: reqTrackStatus,
		Amount: sp.rewardCache[actionData.TxReqID.String()],
	}
	withdrawalReqStatusBytes, _ := json.Marshal(withdrawalReqStatus)

	err = statedb.TrackPdexv3Status(
		stateDB,
		statedb.Pdexv3WithdrawalProtocolFeeStatusPrefix(),
		[]byte(actionData.TxReqID.String()),
		withdrawalReqStatusBytes,
	)
	if err != nil {
		Logger.log.Errorf("PDex v3 Withdrawal Protocol Fee: An error occurred while tracking request tx - Error: %v", err)
	}
	return pairs, err
}

func (sp *stateProcessorV2) mintBlockReward(
	stateDB *statedb.StateDB,
	inst []string,
	pairs map[string]*PoolPairState,
) (map[string]*PoolPairState, error) {
	if len(inst) != 4 {
		msg := fmt.Sprintf("Length of instruction is not valid expect %v but get %v", 4, len(inst))
		Logger.log.Errorf(msg)
		return pairs, errors.New(msg)
	}

	// unmarshal instructions content
	var actionData metadataPdexv3.MintBlockRewardContent
	err := json.Unmarshal([]byte(inst[3]), &actionData)
	if err != nil {
		msg := fmt.Sprintf("Could not unmarshal instruction content %v - Error: %v\n", inst[3], err)
		Logger.log.Errorf(msg)
		return pairs, err
	}

	pair, isExisted := pairs[actionData.PoolPairID]
	if !isExisted {
		msg := fmt.Sprintf("Could not find pair %s for minting", actionData.PoolPairID)
		Logger.log.Errorf(msg)
		return pairs, fmt.Errorf(msg)
	}

	pairReward := actionData.Amount

	pair.lmRewardsPerShare = v2utils.NewTradingPairWithValue(
		&pair.state,
	).AddLMRewards(
		actionData.TokenID, new(big.Int).SetUint64(pairReward), BaseLPFeesPerShare,
		pair.lmRewardsPerShare,
	)

	return pairs, err
}

func (sp *stateProcessorV2) userMintNft(
	stateDB *statedb.StateDB, inst []string, nftIDs map[string]uint64,
) (map[string]uint64, *v2.MintNftStatus, error) {
	if len(inst) != 3 {
		return nftIDs, nil, fmt.Errorf("Expect length of instruction is %v but get %v", 3, len(inst))
	}
	status := byte(0)
	nftID := utils.EmptyString
	txReqID := common.Hash{}
	var burntAmount uint64
	if inst[0] != strconv.Itoa(metadataCommon.Pdexv3UserMintNftRequestMeta) {
		return nftIDs, nil, fmt.Errorf("Expect metaType is %v but get %s", metadataCommon.Pdexv3UserMintNftRequestMeta, inst[1])
	}
	switch inst[1] {
	case common.Pdexv3RejectUserMintNftStatus:
		refundInst := instruction.NewRejectUserMintNft()
		err := refundInst.FromStringSlice(inst)
		if err != nil {
			return nftIDs, nil, err
		}
		burntAmount = refundInst.Amount()
		txReqID = refundInst.TxReqID()
		status = common.Pdexv3RejectStatus
	case common.Pdexv3AcceptUserMintNftStatus:
		acceptInst := instruction.NewAcceptUserMintNft()
		err := acceptInst.FromStringSlice(inst)
		if err != nil {
			return nftIDs, nil, err
		}
		nftID = acceptInst.NftID().String()
		burntAmount = acceptInst.BurntAmount()
		nftIDs[acceptInst.NftID().String()] = acceptInst.BurntAmount()
		txReqID = acceptInst.TxReqID()
		status = common.Pdexv3AcceptStatus
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
	var status byte
	var nftID, stakingPoolID string
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
		status = common.Pdexv3AcceptStatus
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
		status = common.Pdexv3RejectStatus
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
	var status byte
	var nftID, stakingPoolID string
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
		status = common.Pdexv3AcceptStatus
		stakingPoolID = acceptInst.StakingPoolID().String()
		liquidity = acceptInst.Amount()
		nftID = acceptInst.NftID().String()
		stakingPoolState := stakingPoolStates[stakingPoolID]
		err = stakingPoolState.updateLiquidity(nftID, liquidity, beaconHeight, subOperator)
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
		status = common.Pdexv3RejectStatus
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

func (sp *stateProcessorV2) distributeStakingReward(
	stateDB *statedb.StateDB,
	inst []string,
	stakingPools map[string]*StakingPoolState,
) (map[string]*StakingPoolState, error) {
	if len(inst) != 4 {
		msg := fmt.Sprintf("Length of instruction is not valid expect %v but get %v", 4, len(inst))
		Logger.log.Errorf(msg)
		return stakingPools, errors.New(msg)
	}

	// unmarshal instructions content
	var actionData metadataPdexv3.DistributeStakingRewardContent
	err := json.Unmarshal([]byte(inst[3]), &actionData)
	if err != nil {
		msg := fmt.Sprintf("Could not unmarshal instruction content %v - Error: %v\n", inst[3], err)
		Logger.log.Errorf(msg)
		return stakingPools, err
	}

	pool, isExisted := stakingPools[actionData.StakingPoolID]
	if !isExisted {
		msg := fmt.Sprintf("Could not find staking pool %v for distributing", actionData.StakingPoolID)
		Logger.log.Errorf(msg)
		return stakingPools, fmt.Errorf(msg)
	}

	for rewardToken, rewardAmount := range actionData.Rewards {
		pool.AddReward(rewardToken, rewardAmount)
	}

	return stakingPools, err
}

func (sp *stateProcessorV2) withdrawStakingReward(
	stateDB *statedb.StateDB,
	inst []string,
	pools map[string]*StakingPoolState,
) (map[string]*StakingPoolState, error) {
	if len(inst) != 4 {
		msg := fmt.Sprintf("Length of instruction is not valid expect %v but get %v", 4, len(inst))
		Logger.log.Errorf(msg)
		return pools, errors.New(msg)
	}

	// unmarshal instructions content
	var actionData metadataPdexv3.WithdrawalStakingRewardContent
	err := json.Unmarshal([]byte(inst[3]), &actionData)
	if err != nil {
		msg := fmt.Sprintf("Could not unmarshal instruction content %v - Error: %v\n", inst[3], err)
		Logger.log.Errorf(msg)
		return pools, err
	}

	withdrawalStatus := inst[2]
	var reqTrackStatus int
	if withdrawalStatus == metadataPdexv3.RequestAcceptedChainStatus {
		// check conditions
		pool, isExisted := pools[actionData.StakingPoolID]
		if !isExisted {
			msg := fmt.Sprintf("Could not find staking pool %s for withdrawal", actionData.StakingPoolID)
			Logger.log.Errorf(msg)
			return pools, errors.New(msg)
		}

		share, isExisted := pool.stakers[actionData.NftID.String()]
		if !isExisted {
			msg := fmt.Sprintf("Could not find staker %s for withdrawal", actionData.NftID.String())
			Logger.log.Errorf(msg)
			return pools, errors.New(msg)
		}

		// update state after reward withdrawal
		share.rewards = resetKeyValueToZero(share.rewards)
		share.lastRewardsPerShare = pool.RewardsPerShare()

		reqTrackStatus = metadataPdexv3.WithdrawStakingRewardSuccessStatus

		_, found := sp.receiverCache[actionData.TxReqID.String()]
		if !found {
			sp.receiverCache[actionData.TxReqID.String()] = map[common.Hash]metadataPdexv3.ReceiverInfo{}
		}
		sp.receiverCache[actionData.TxReqID.String()][actionData.TokenID] = actionData.Receiver
	} else {
		reqTrackStatus = metadataPdexv3.WithdrawStakingRewardFailedStatus
	}

	if reqTrackStatus == metadataPdexv3.WithdrawProtocolFeeSuccessStatus && !actionData.IsLastInst {
		return pools, nil
	}

	withdrawalReqStatus := metadataPdexv3.WithdrawalStakingRewardStatus{
		Status:    reqTrackStatus,
		Receivers: sp.receiverCache[actionData.TxReqID.String()],
	}
	withdrawalReqStatusBytes, _ := json.Marshal(withdrawalReqStatus)

	err = statedb.TrackPdexv3Status(
		stateDB,
		statedb.Pdexv3WithdrawalStakingRewardStatusPrefix(),
		[]byte(actionData.TxReqID.String()),
		withdrawalReqStatusBytes,
	)
	if err != nil {
		Logger.log.Errorf("PDex v3 Withdrawal Staking Reward Fee: An error occurred while tracking request tx - Error: %v", err)
	}
	return pools, err
}

func (sp *stateProcessorV2) distributeMiningOrderReward(stateDB *statedb.StateDB,
	inst []string,
	pairs map[string]*PoolPairState,
) (map[string]*PoolPairState, error) {
	if len(inst) != 4 {
		msg := fmt.Sprintf("Length of instruction is not valid expect %v but get %v", 4, len(inst))
		Logger.log.Errorf(msg)
		return pairs, errors.New(msg)
	}

	// unmarshal instructions content
	var actionData metadataPdexv3.DistributeMiningOrderRewardContent
	err := json.Unmarshal([]byte(inst[3]), &actionData)
	if err != nil {
		msg := fmt.Sprintf("Could not unmarshal instruction content %v - Error: %v\n", inst[3], err)
		Logger.log.Errorf(msg)
		return pairs, err
	}

	pair, isExisted := pairs[actionData.PoolPairID]
	if !isExisted {
		msg := fmt.Sprintf("Could not find pair %s for minting order reward", actionData.PoolPairID)
		Logger.log.Errorf(msg)
		return pairs, fmt.Errorf(msg)
	}

	orderRewards := v2.SplitOrderRewardLiquidityMining(
		pair.makingVolume[actionData.MakingTokenID].volume,
		new(big.Int).SetUint64(actionData.Amount),
		actionData.TokenID,
	)

	for nftID, reward := range orderRewards {
		if _, ok := pair.orderRewards[nftID]; !ok {
			pair.orderRewards[nftID] = NewOrderReward()
		}
		pair.orderRewards[nftID].AddReward(actionData.TokenID, reward)
	}

	delete(pair.makingVolume, actionData.MakingTokenID)

	return pairs, nil
}
