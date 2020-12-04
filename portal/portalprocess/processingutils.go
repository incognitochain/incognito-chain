package portalprocess

import (
	"github.com/incognitochain/incognito-chain/basemeta"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/portal"
	"sort"
	"strconv"
)

func CollectPortalInsts (pm *PortalManager, metaType int, action []string, shardID byte) {
	switch metaType {
	case basemeta.PortalCustodianDepositMeta:
		pm.PortalInstructions[basemeta.PortalCustodianDepositMeta].putAction(action, shardID)

	// off this metatype in portal v3
	//case basemeta.PortalRequestPortingMeta:
	//	pm.PortalInstructions[basemeta.PortalRequestPortingMeta].putAction(action, shardID)
	case basemeta.PortalRequestPortingMetaV3:
		pm.PortalInstructions[basemeta.PortalRequestPortingMetaV3].putAction(action, shardID)
	case basemeta.PortalUserRequestPTokenMeta:
		pm.PortalInstructions[basemeta.PortalUserRequestPTokenMeta].putAction(action, shardID)
	case basemeta.PortalExchangeRatesMeta:
		pm.PortalInstructions[basemeta.PortalExchangeRatesMeta].putAction(action, shardID)
	case basemeta.PortalCustodianWithdrawRequestMeta:
		pm.PortalInstructions[basemeta.PortalCustodianWithdrawRequestMeta].putAction(action, shardID)
	// off this metatype in portal v3
	//case basemeta.PortalRedeemRequestMeta:
	//	pm.PortalInstructions[basemeta.PortalRedeemRequestMeta].putAction(action, shardID)
	case basemeta.PortalRedeemRequestMetaV3:
		pm.PortalInstructions[basemeta.PortalRedeemRequestMetaV3].putAction(action, shardID)
	// off this metatype in portal v3
	//case basemeta.PortalRequestUnlockCollateralMeta:
	//	pm.PortalInstructions[basemeta.PortalRequestUnlockCollateralMeta].putAction(action, shardID)
	case basemeta.PortalRequestUnlockCollateralMetaV3:
		pm.PortalInstructions[basemeta.PortalRequestUnlockCollateralMetaV3].putAction(action, shardID)
	case basemeta.PortalRequestWithdrawRewardMeta:
		pm.PortalInstructions[basemeta.PortalRequestWithdrawRewardMeta].putAction(action, shardID)
	case basemeta.PortalRedeemFromLiquidationPoolMetaV3:
		pm.PortalInstructions[basemeta.PortalRedeemFromLiquidationPoolMetaV3].putAction(action, shardID)
	case basemeta.PortalCustodianTopupMetaV2:
		pm.PortalInstructions[basemeta.PortalCustodianTopupMetaV2].putAction(action, shardID)
	case basemeta.PortalReqMatchingRedeemMeta:
		pm.PortalInstructions[basemeta.PortalReqMatchingRedeemMeta].putAction(action, shardID)
	case basemeta.PortalTopUpWaitingPortingRequestMeta:
		pm.PortalInstructions[basemeta.PortalTopUpWaitingPortingRequestMeta].putAction(action, shardID)
	case basemeta.PortalCustodianTopupMetaV3:
		pm.PortalInstructions[basemeta.PortalCustodianTopupMetaV3].putAction(action, shardID)
	case basemeta.PortalTopUpWaitingPortingRequestMetaV3:
		pm.PortalInstructions[basemeta.PortalTopUpWaitingPortingRequestMetaV3].putAction(action, shardID)

	case basemeta.PortalCustodianDepositMetaV3:
		pm.PortalInstructions[basemeta.PortalCustodianDepositMetaV3].putAction(action, shardID)
	case basemeta.PortalCustodianWithdrawRequestMetaV3:
		pm.PortalInstructions[basemeta.PortalCustodianWithdrawRequestMetaV3].putAction(action, shardID)

	case basemeta.RelayingBNBHeaderMeta:
		pm.RelayingChains[basemeta.RelayingBNBHeaderMeta].putAction(action)
	case basemeta.RelayingBTCHeaderMeta:
		pm.RelayingChains[basemeta.RelayingBTCHeaderMeta].putAction(action)
	default:
		return
	}
}

//TODO:
// check liquidation
func autoCheckAndCreatePortalLiquidationInsts(
	bc basemeta.ChainRetriever,
	beaconHeight uint64,
	shardHeights map[byte]uint64,
	currentPortalState *CurrentPortalState,
	portalParams portal.PortalParams,
	pm *PortalManager) ([][]string, error) {
	insts := [][]string{}

	// check there is any waiting porting request timeout
	expiredWPortingProcessor := pm.PortalInstructions[basemeta.PortalExpiredWaitingPortingReqMeta]
	expiredWaitingPortingInsts, err := expiredWPortingProcessor.BuildNewInsts(bc, "", 0, currentPortalState, beaconHeight, shardHeights, portalParams, nil)
	if err != nil {
		Logger.log.Errorf("Error when check and build custodian liquidation %v\n", err)
	}
	if len(expiredWaitingPortingInsts) > 0 {
		insts = append(insts, expiredWaitingPortingInsts...)
	}
	Logger.log.Infof("There are %v instruction for expired waiting porting in portal\n", len(expiredWaitingPortingInsts))

	// case 1: check there is any custodian doesn't send public tokens back to user after TimeOutCustodianReturnPubToken
	// get custodian's collateral to return user
	liquidateCustodianProcessor := pm.PortalInstructions[basemeta.PortalLiquidateCustodianMetaV3]
	custodianLiqInsts, err := liquidateCustodianProcessor.BuildNewInsts(bc, "", 0, currentPortalState, beaconHeight, shardHeights, portalParams, nil)
	if err != nil {
		Logger.log.Errorf("Error when check and build custodian liquidation %v\n", err)
	}
	if len(custodianLiqInsts) > 0 {
		insts = append(insts, custodianLiqInsts...)
	}
	Logger.log.Infof("There are %v instruction for custodian liquidation in portal\n", len(custodianLiqInsts))

	// case 2: check collateral's value (locked collateral amount) drops below MinRatio
	liquidationByRateProcessor := pm.PortalInstructions[basemeta.PortalLiquidateByRatesMetaV3]
	exchangeRatesLiqInsts, err := liquidationByRateProcessor.BuildNewInsts(bc, "", 0, currentPortalState, beaconHeight, shardHeights, portalParams, nil)
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
	p portalInstructionProcessor,
	bc basemeta.ChainRetriever,
	stateDB *statedb.StateDB,
	currentPortalState *CurrentPortalState,
	beaconHeight uint64,
	shardHeights map[byte]uint64,
	portalParams portal.PortalParams) ([][]string, error) {

	instructions := [][]string{}
	actions := p.getActions()
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
			optionalData, err := p.prepareDataForBlockProducer(stateDB, contentStr)
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
				portalParams,
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
func HandlePortalInsts(
	bc basemeta.ChainRetriever,
	stateDB *statedb.StateDB,
	beaconHeight uint64,
	shardHeights map[byte]uint64,
	currentPortalState *CurrentPortalState,
	rewardForCustodianByEpoch map[common.Hash]uint64,
	portalParams portal.PortalParams,
	pm *PortalManager,
) ([][]string, error) {
	instructions := [][]string{}

	oldMatchedRedeemRequests := CloneRedeemRequests(currentPortalState.MatchedRedeemRequests)

	// auto-liquidation portal instructions
	portalLiquidationInsts, err := autoCheckAndCreatePortalLiquidationInsts(
		bc,
		beaconHeight,
		shardHeights,
		currentPortalState,
		portalParams,
		pm,
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
	for metaType := range pm.PortalInstructions {
		metaTypes = append(metaTypes, metaType)
	}
	sort.Ints(metaTypes)
	for _, metaType := range metaTypes {
		actions := pm.PortalInstructions[metaType]
		newInst, err := buildNewPortalInstsFromActions(
			actions,
			bc,
			stateDB,
			currentPortalState,
			beaconHeight,
			shardHeights,
			portalParams)

		if err != nil {
			Logger.log.Error(err)
		}
		if len(newInst) > 0 {
			instructions = append(instructions, newInst...)
		}
	}

	// check and create instruction for picking more custodians for timeout waiting redeem requests
	var pickCustodiansForRedeemReqInsts [][]string

	pickCustodiansProcessor := pm.PortalInstructions[basemeta.PortalPickMoreCustodianForRedeemMeta]
	pickCustodiansForRedeemReqInsts, err = pickCustodiansProcessor.BuildNewInsts(bc, "", 0,  currentPortalState,beaconHeight, shardHeights,
		portalParams, nil)
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

func ProcessPortalInstructions(
	portalStateDB *statedb.StateDB,
	portalParams portal.PortalParams,
	beaconHeight uint64,
	instructions [][]string,
	pm *PortalManager,
	epoch uint64) error {
	currentPortalState, err := InitCurrentPortalStateFromDB(portalStateDB)
	if err != nil {
		Logger.log.Error(err)
		return nil
	}

	// re-use update info of bridge
	updatingInfoByTokenID := map[common.Hash]basemeta.UpdatingInfo{}

	for _, inst := range instructions {
		if len(inst) < 4 {
			continue // Not error, just not Portal instruction
		}

		var err error
		metaType, _ := strconv.Atoi(inst[0])
		processor := pm.GetPortalInstProcessorByMetaType(metaType)
		if processor != nil {
			err = processor.ProcessInsts(portalStateDB, beaconHeight, inst, currentPortalState, portalParams, updatingInfoByTokenID)
			if err != nil {
				Logger.log.Errorf("Process portal instruction err: %v, inst %+v", err, inst)
			}
			continue
		}

		switch inst[0] {
		// ============ Reward ============
		// portal reward
		case strconv.Itoa(basemeta.PortalRewardMeta), strconv.Itoa(basemeta.PortalRewardMetaV3):
			err = ProcessPortalReward(portalStateDB, beaconHeight, inst, currentPortalState, portalParams, epoch)
		// total custodian reward instruction
		case strconv.Itoa(basemeta.PortalTotalRewardCustodianMeta):
			err = ProcessPortalTotalCustodianReward(portalStateDB, beaconHeight, inst, currentPortalState, portalParams, epoch)

		// ============ Portal smart contract ============
		// todo: add more metadata need to unlock token from sc
		case strconv.Itoa(basemeta.PortalCustodianWithdrawConfirmMetaV3),
			strconv.Itoa(basemeta.PortalRedeemFromLiquidationPoolConfirmMetaV3),
			strconv.Itoa(basemeta.PortalLiquidateRunAwayCustodianConfirmMetaV3):
			err = ProcessPortalConfirmWithdrawInstV3(portalStateDB, beaconHeight, inst, currentPortalState, portalParams)
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
	portalParams portal.PortalParams) error {
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