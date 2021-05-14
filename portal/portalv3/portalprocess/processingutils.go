package portalprocess

import (
	"sort"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
)

func CollectPortalInstsV3(pv3 map[int]PortalInstructionProcessorV3, metaType int, action []string, shardID byte) {
	switch metaType {
	case metadata.PortalCustodianDepositMeta:
		pv3[metadata.PortalCustodianDepositMeta].PutAction(action, shardID)

	// off this metatype in portal v3
	//case metadata.PortalRequestPortingMeta:
	//	pm.PortalInstProcessorsV3[metadata.PortalRequestPortingMeta].PutAction(action, shardID)
	case metadata.PortalRequestPortingMetaV3:
		pv3[metadata.PortalRequestPortingMetaV3].PutAction(action, shardID)
	case metadata.PortalUserRequestPTokenMeta:
		pv3[metadata.PortalUserRequestPTokenMeta].PutAction(action, shardID)
	case metadata.PortalExchangeRatesMeta:
		pv3[metadata.PortalExchangeRatesMeta].PutAction(action, shardID)
	case metadata.PortalCustodianWithdrawRequestMeta:
		pv3[metadata.PortalCustodianWithdrawRequestMeta].PutAction(action, shardID)
	// off this metatype in portal v3
	//case metadata.PortalRedeemRequestMeta:
	//	pm.PortalInstProcessorsV3[metadata.PortalRedeemRequestMeta].PutAction(action, shardID)
	case metadata.PortalRedeemRequestMetaV3:
		pv3[metadata.PortalRedeemRequestMetaV3].PutAction(action, shardID)
	// off this metatype in portal v3
	//case metadata.PortalRequestUnlockCollateralMeta:
	//	pm.PortalInstProcessorsV3[metadata.PortalRequestUnlockCollateralMeta].PutAction(action, shardID)
	case metadata.PortalRequestUnlockCollateralMetaV3:
		pv3[metadata.PortalRequestUnlockCollateralMetaV3].PutAction(action, shardID)
	case metadata.PortalRequestWithdrawRewardMeta:
		pv3[metadata.PortalRequestWithdrawRewardMeta].PutAction(action, shardID)
	case metadata.PortalRedeemFromLiquidationPoolMetaV3:
		pv3[metadata.PortalRedeemFromLiquidationPoolMetaV3].PutAction(action, shardID)
	case metadata.PortalCustodianTopupMetaV2:
		pv3[metadata.PortalCustodianTopupMetaV2].PutAction(action, shardID)
	case metadata.PortalReqMatchingRedeemMeta:
		pv3[metadata.PortalReqMatchingRedeemMeta].PutAction(action, shardID)
	case metadata.PortalTopUpWaitingPortingRequestMeta:
		pv3[metadata.PortalTopUpWaitingPortingRequestMeta].PutAction(action, shardID)
	case metadata.PortalCustodianTopupMetaV3:
		pv3[metadata.PortalCustodianTopupMetaV3].PutAction(action, shardID)
	case metadata.PortalTopUpWaitingPortingRequestMetaV3:
		pv3[metadata.PortalTopUpWaitingPortingRequestMetaV3].PutAction(action, shardID)

	case metadata.PortalCustodianDepositMetaV3:
		pv3[metadata.PortalCustodianDepositMetaV3].PutAction(action, shardID)
	case metadata.PortalCustodianWithdrawRequestMetaV3:
		pv3[metadata.PortalCustodianWithdrawRequestMetaV3].PutAction(action, shardID)

		// TODO: move to relaying package
	//case metadata.RelayingBNBHeaderMeta:
	//	pv3.RelayingChains[metadata.RelayingBNBHeaderMeta].putAction(action)
	//case metadata.RelayingBTCHeaderMeta:
	//	pv3.RelayingChains[metadata.RelayingBTCHeaderMeta].putAction(action)
	default:
		return
	}
}

// auto check and liquidate
func autoCheckAndCreatePortalLiquidationInsts(
	bc metadata.ChainRetriever,
	beaconHeight uint64,
	shardHeights map[byte]uint64,
	currentPortalState *CurrentPortalState,
	portalParam config.PortalParam,
	pv3 map[int]PortalInstructionProcessorV3) ([][]string, error) {
	insts := [][]string{}

	// check there is any waiting porting request timeout
	expiredWPortingProcessor := pv3[metadata.PortalExpiredWaitingPortingReqMeta]
	expiredWaitingPortingInsts, err := expiredWPortingProcessor.BuildNewInsts(bc, "", 0, currentPortalState, beaconHeight, shardHeights, portalParam, nil)
	if err != nil {
		Logger.log.Errorf("Error when check and build custodian liquidation %v\n", err)
	}
	if len(expiredWaitingPortingInsts) > 0 {
		insts = append(insts, expiredWaitingPortingInsts...)
	}
	Logger.log.Infof("There are %v instruction for expired waiting porting in portal\n", len(expiredWaitingPortingInsts))

	// case 1: check there is any custodian doesn't send public tokens back to user after TimeOutCustodianReturnPubToken
	// get custodian's collateral to return user
	liquidateCustodianProcessor := pv3[metadata.PortalLiquidateCustodianMetaV3]
	custodianLiqInsts, err := liquidateCustodianProcessor.BuildNewInsts(bc, "", 0, currentPortalState, beaconHeight, shardHeights, portalParam, nil)
	if err != nil {
		Logger.log.Errorf("Error when check and build custodian liquidation %v\n", err)
	}
	if len(custodianLiqInsts) > 0 {
		insts = append(insts, custodianLiqInsts...)
	}
	Logger.log.Infof("There are %v instruction for custodian liquidation in portal\n", len(custodianLiqInsts))

	// case 2: check collateral's value (locked collateral amount) drops below MinRatio
	liquidationByRateProcessor := pv3[metadata.PortalLiquidateByRatesMetaV3]
	exchangeRatesLiqInsts, err := liquidationByRateProcessor.BuildNewInsts(bc, "", 0, currentPortalState, beaconHeight, shardHeights, portalParam, nil)
	if err != nil {
		Logger.log.Errorf("Error when check and build exchange rates liquidation %v\n", err)
	}
	if len(exchangeRatesLiqInsts) > 0 {
		insts = append(insts, exchangeRatesLiqInsts...)
	}

	Logger.log.Infof("There are %v instruction for exchange rates liquidation in portal\n", len(exchangeRatesLiqInsts))

	return insts, nil
}

func buildNewPortalInstsFromActions(
	p PortalInstructionProcessorV3,
	bc metadata.ChainRetriever,
	stateDB *statedb.StateDB,
	currentPortalState *CurrentPortalState,
	beaconHeight uint64,
	shardHeights map[byte]uint64,
	portalParam config.PortalParam) ([][]string, error) {

	instructions := [][]string{}
	actions := p.GetActions()
	var shardIDKeys []int
	for k := range actions {
		shardIDKeys = append(shardIDKeys, int(k))
	}

	sort.Ints(shardIDKeys)
	for _, value := range shardIDKeys {
		shardID := byte(value)
		actions := actions[shardID]
		for _, action := range actions {
			contentStr := action[1]
			optionalData, err := p.PrepareDataForBlockProducer(stateDB, contentStr)
			if err != nil {
				Logger.log.Errorf("Error when preparing data before processing instruction %+v", err)
				continue
			}
			newInst, err := p.BuildNewInsts(
				bc,
				contentStr,
				shardID,
				currentPortalState,
				beaconHeight,
				shardHeights,
				portalParam,
				optionalData,
			)
			if err != nil {
				Logger.log.Errorf("Error when building new instructions : %v", err)
				continue
			}
			if len(newInst) > 0 {
				instructions = append(instructions, newInst...)
			}
		}
	}

	return instructions, nil
}

// Build instructions portal reward for each beacon block
func handlePortalRewardInsts(
	beaconHeight uint64,
	currentPortalState *CurrentPortalState,
	rewardForCustodianByEpoch map[common.Hash]uint64,
	newMatchedRedeemReqIDs []string,
) ([][]string, error) {
	instructions := [][]string{}

	// Build instructions portal reward for each beacon block
	portalRewardInsts, err := buildPortalRewardsInsts(beaconHeight, currentPortalState, rewardForCustodianByEpoch, newMatchedRedeemReqIDs)
	if err != nil {
		Logger.log.Error(err)
	}
	if len(portalRewardInsts) > 0 {
		instructions = append(instructions, portalRewardInsts...)
	}

	return instructions, nil
}

// handle portal instructions for block producer
func HandlePortalInstsV3(
	bc metadata.ChainRetriever,
	stateDB *statedb.StateDB,
	beaconHeight uint64,
	shardHeights map[byte]uint64,
	currentPortalState *CurrentPortalState,
	rewardForCustodianByEpoch map[common.Hash]uint64,
	portalParam config.PortalParam,
	pv3 map[int]PortalInstructionProcessorV3,
) ([][]string, error) {
	instructions := [][]string{}

	oldMatchedRedeemRequests := CloneRedeemRequests(currentPortalState.MatchedRedeemRequests)

	// auto-liquidation portal instructions
	portalLiquidationInsts, err := autoCheckAndCreatePortalLiquidationInsts(
		bc,
		beaconHeight,
		shardHeights,
		currentPortalState,
		portalParam,
		pv3,
	)
	if err != nil {
		Logger.log.Error(err)
	}
	if len(portalLiquidationInsts) > 0 {
		instructions = append(instructions, portalLiquidationInsts...)
	}

	// producer portal instructions for actions from shards
	// sort metadata type map to make it consistent for every run
	var metaTypes []int
	for metaType := range pv3 {
		metaTypes = append(metaTypes, metaType)
	}
	sort.Ints(metaTypes)
	for _, metaType := range metaTypes {
		actions := pv3[metaType]
		newInst, err := buildNewPortalInstsFromActions(
			actions,
			bc,
			stateDB,
			currentPortalState,
			beaconHeight,
			shardHeights,
			portalParam)

		if err != nil {
			Logger.log.Error(err)
		}
		if len(newInst) > 0 {
			instructions = append(instructions, newInst...)
		}
	}

	// check and create instruction for picking more custodians for timeout waiting redeem requests
	var pickCustodiansForRedeemReqInsts [][]string

	pickCustodiansProcessor := pv3[metadata.PortalPickMoreCustodianForRedeemMeta]
	pickCustodiansForRedeemReqInsts, err = pickCustodiansProcessor.BuildNewInsts(bc, "", 0, currentPortalState, beaconHeight, shardHeights,
		portalParam, nil)
	if err != nil {
		Logger.log.Error(err)
	}
	if len(pickCustodiansForRedeemReqInsts) > 0 {
		instructions = append(instructions, pickCustodiansForRedeemReqInsts...)
	}

	// get new matched redeem request at beacon height
	newMatchedRedeemReqIDs := getNewMatchedRedeemReqIDs(oldMatchedRedeemRequests, currentPortalState.MatchedRedeemRequests)

	// calculate rewards (include porting fee and redeem fee) for custodians and build instructions at beaconHeight
	portalRewardsInsts, err := handlePortalRewardInsts(
		beaconHeight,
		currentPortalState,
		rewardForCustodianByEpoch,
		newMatchedRedeemReqIDs,
	)

	if err != nil {
		Logger.log.Error(err)
	}
	if len(portalRewardsInsts) > 0 {
		instructions = append(instructions, portalRewardsInsts...)
	}

	return instructions, nil
}

func ProcessPortalInstsV3(
	portalStateDB *statedb.StateDB,
	portalParam config.PortalParam,
	beaconHeight uint64,
	instructions [][]string,
	pv3 map[int]PortalInstructionProcessorV3,
	epoch uint64) error {
	currentPortalState, err := InitCurrentPortalStateFromDB(portalStateDB)
	if err != nil {
		Logger.log.Error(err)
		return nil
	}

	// re-use update info of bridge
	updatingInfoByTokenID := map[common.Hash]metadata.UpdatingInfo{}

	for _, inst := range instructions {
		if len(inst) < 4 {
			continue // Not error, just not Portal instruction
		}

		var err error
		metaType, _ := strconv.Atoi(inst[0])
		processor := GetPortalInstProcessorByMetaType(pv3, metaType)
		if processor != nil {
			err = processor.ProcessInsts(portalStateDB, beaconHeight, inst, currentPortalState, portalParam, updatingInfoByTokenID)
			if err != nil {
				Logger.log.Errorf("Process portal instruction err: %v, inst %+v", err, inst)
			}
			continue
		}

		switch inst[0] {
		// ============ Reward ============
		// portal reward
		case strconv.Itoa(metadata.PortalRewardMeta), strconv.Itoa(metadata.PortalRewardMetaV3):
			err = ProcessPortalReward(portalStateDB, beaconHeight, inst, currentPortalState, portalParam, epoch)
		// total custodian reward instruction
		case strconv.Itoa(metadata.PortalTotalRewardCustodianMeta):
			err = ProcessPortalTotalCustodianReward(portalStateDB, beaconHeight, inst, currentPortalState, portalParam, epoch)

		// ============ Portal smart contract ============
		case strconv.Itoa(metadata.PortalCustodianWithdrawConfirmMetaV3),
			strconv.Itoa(metadata.PortalRedeemFromLiquidationPoolConfirmMetaV3),
			strconv.Itoa(metadata.PortalLiquidateRunAwayCustodianConfirmMetaV3):
			err = ProcessPortalConfirmWithdrawInstV3(portalStateDB, beaconHeight, inst, currentPortalState)
		}

		if err != nil {
			Logger.log.Errorf("Process portal instruction err: %v, inst %+v", err, inst)
		}

	}

	// pick the final exchangeRates
	PickExchangesRatesFinal(currentPortalState)

	// update info of bridge portal token
	for _, updatingInfo := range updatingInfoByTokenID {
		var updatingAmt uint64
		var updatingType string
		if updatingInfo.CountUpAmt > updatingInfo.DeductAmt {
			updatingAmt = updatingInfo.CountUpAmt - updatingInfo.DeductAmt
			updatingType = "+"
		}
		if updatingInfo.CountUpAmt < updatingInfo.DeductAmt {
			updatingAmt = updatingInfo.DeductAmt - updatingInfo.CountUpAmt
			updatingType = "-"
		}
		err := statedb.UpdateBridgeTokenInfo(
			portalStateDB,
			updatingInfo.TokenID,
			updatingInfo.ExternalTokenID,
			updatingInfo.IsCentralized,
			updatingAmt,
			updatingType,
		)
		if err != nil {
			return err
		}
	}

	// store updated currentPortalState to leveldb with new beacon height
	err = StorePortalStateToDB(portalStateDB, currentPortalState)
	if err != nil {
		Logger.log.Error(err)
	}

	return nil
}

func calcMedian(ratesList []uint64) uint64 {
	mNumber := len(ratesList) / 2

	if len(ratesList)%2 == 0 {
		return (ratesList[mNumber-1] + ratesList[mNumber]) / 2
	}

	return ratesList[mNumber]
}

func ProcessPortalConfirmWithdrawInstV3(
	portalStateDB *statedb.StateDB,
	beaconHeight uint64,
	instructions []string,
	currentPortalState *CurrentPortalState,
) error {
	if currentPortalState == nil {
		Logger.log.Errorf("current portal state is nil")
		return nil
	}

	if len(instructions) != 8 {
		return nil // skip the instruction
	}

	txReqIDStr := instructions[6]
	txReqID, _ := common.Hash{}.NewHashFromStr(txReqIDStr)

	// store withdraw confirm proof
	err := statedb.StoreWithdrawCollateralConfirmProof(portalStateDB, *txReqID, beaconHeight+1)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occurred while store custodian withdraw confirm proof: %+v", err)
		return nil
	}
	return nil
}
