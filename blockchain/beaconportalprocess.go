package blockchain

import (
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
)

//todo:
func (blockchain *BlockChain) processPortalInstructions(portalStateDB *statedb.StateDB, block *BeaconBlock) error {
	//// Note: should comment this code if you need to create local chain.
	//if blockchain.config.ChainParams.Net == Testnet && block.Header.Height < 1580600 {
	//	return nil
	//}
	//beaconHeight := block.Header.Height - 1
	//currentPortalState, err := InitCurrentPortalStateFromDB(portalStateDB)
	//if err != nil {
	//	Logger.log.Error(err)
	//	return nil
	//}
	//
	//portalParams := blockchain.GetPortalParams(block.GetHeight())
	//
	//// re-use update info of bridge
	//updatingInfoByTokenID := map[common.Hash]UpdatingInfo{}
	//
	//for _, inst := range block.Body.Instructions {
	//	if len(inst) < 4 {
	//		continue // Not error, just not Portal instruction
	//	}
	//
	//	var err error
	//	switch inst[0] {
	//	// ============ Exchange rate ============
	//	case strconv.Itoa(metadata.PortalExchangeRatesMeta):
	//		err = blockchain.processPortalExchangeRates(portalStateDB, beaconHeight, inst, currentPortalState, portalParams)
	//
	//	// ============ Custodian ============
	//	// custodian deposit collateral
	//	case strconv.Itoa(metadata.PortalCustodianDepositMeta):
	//		err = blockchain.processPortalCustodianDeposit(portalStateDB, beaconHeight, inst, currentPortalState, portalParams)
	//	// custodian withdraw collateral
	//	case strconv.Itoa(metadata.PortalCustodianWithdrawRequestMeta):
	//		err = blockchain.processPortalCustodianWithdrawRequest(portalStateDB, beaconHeight, inst, currentPortalState, portalParams)
	//	// custodian deposit collateral v3
	//	case strconv.Itoa(metadata.PortalCustodianDepositMetaV3):
	//		err = blockchain.processPortalCustodianDepositV3(portalStateDB, beaconHeight, inst, currentPortalState, portalParams)
	//	// custodian request withdraw collateral v3
	//	case strconv.Itoa(metadata.PortalCustodianWithdrawRequestMetaV3):
	//		err = blockchain.processPortalCustodianWithdrawV3(portalStateDB, beaconHeight, inst, currentPortalState, portalParams)
	//
	//	// ============ Porting flow ============
	//	// porting request
	//	case strconv.Itoa(metadata.PortalRequestPortingMeta), strconv.Itoa(metadata.PortalRequestPortingMetaV3):
	//		err = blockchain.processPortalUserRegister(portalStateDB, beaconHeight, inst, currentPortalState, portalParams)
	//	// request ptoken
	//	case strconv.Itoa(metadata.PortalUserRequestPTokenMeta):
	//		err = blockchain.processPortalUserReqPToken(portalStateDB, beaconHeight, inst, currentPortalState, portalParams, updatingInfoByTokenID)
	//
	//	// ============ Redeem flow ============
	//	// redeem request
	//	case strconv.Itoa(metadata.PortalRedeemRequestMeta), strconv.Itoa(metadata.PortalRedeemRequestMetaV3):
	//		err = blockchain.processPortalRedeemRequest(portalStateDB, beaconHeight, inst, currentPortalState, portalParams, updatingInfoByTokenID)
	//	// custodian request matching waiting redeem requests
	//	case strconv.Itoa(metadata.PortalReqMatchingRedeemMeta):
	//		err = blockchain.processPortalReqMatchingRedeem(portalStateDB, beaconHeight, inst, currentPortalState, portalParams)
	//	case strconv.Itoa(metadata.PortalPickMoreCustodianForRedeemMeta):
	//		err = blockchain.processPortalPickMoreCustodiansForTimeOutWaitingRedeemReq(portalStateDB, beaconHeight, inst, currentPortalState, portalParams)
	//	// request unlock collateral
	//	case strconv.Itoa(metadata.PortalRequestUnlockCollateralMeta), strconv.Itoa(metadata.PortalRequestUnlockCollateralMetaV3):
	//		err = blockchain.processPortalUnlockCollateral(portalStateDB, beaconHeight, inst, currentPortalState, portalParams)
	//
	//	// ============ Liquidation ============
	//	// liquidation custodian run away
	//	case strconv.Itoa(metadata.PortalLiquidateCustodianMeta), strconv.Itoa(metadata.PortalLiquidateCustodianMetaV3):
	//		err = blockchain.processPortalLiquidateCustodian(portalStateDB, beaconHeight, inst, currentPortalState, portalParams)
	//	//liquidation exchange rates
	//	case strconv.Itoa(metadata.PortalLiquidateTPExchangeRatesMeta):
	//		err = blockchain.processLiquidationTopPercentileExchangeRates(portalStateDB, beaconHeight, inst, currentPortalState, portalParams)
	//	// custodian topup
	//	case strconv.Itoa(metadata.PortalCustodianTopupMetaV2):
	//		err = blockchain.processPortalCustodianTopup(portalStateDB, beaconHeight, inst, currentPortalState, portalParams)
	//	// top up for waiting porting
	//	case strconv.Itoa(metadata.PortalTopUpWaitingPortingRequestMeta):
	//		err = blockchain.processPortalTopUpWaitingPorting(portalStateDB, beaconHeight, inst, currentPortalState, portalParams)
	//	// redeem from liquidation pool
	//	case strconv.Itoa(metadata.PortalRedeemFromLiquidationPoolMeta):
	//		err = blockchain.processPortalRedeemLiquidateExchangeRates(portalStateDB, beaconHeight, inst, currentPortalState, portalParams, updatingInfoByTokenID)
	//	// expired waiting porting request
	//	case strconv.Itoa(metadata.PortalExpiredWaitingPortingReqMeta):
	//		err = blockchain.processPortalExpiredPortingRequest(portalStateDB, beaconHeight, inst, currentPortalState, portalParams)
	//
	//	// liquidation by exchange rate v3
	//	case strconv.Itoa(metadata.PortalLiquidateByRatesMetaV3):
	//		err = blockchain.processLiquidationByExchangeRatesV3(portalStateDB, beaconHeight, inst, currentPortalState, portalParams)
	//	// redeem from liquidation pool v3
	//	case strconv.Itoa(metadata.PortalRedeemFromLiquidationPoolMetaV3):
	//		err = blockchain.processPortalRedeemFromLiquidationPoolV3(portalStateDB, beaconHeight, inst, currentPortalState, portalParams, updatingInfoByTokenID)
	//	// custodian topup v3
	//	case strconv.Itoa(metadata.PortalCustodianTopupMetaV3):
	//		err = blockchain.processPortalCustodianTopupV3(portalStateDB, beaconHeight, inst, currentPortalState, portalParams)
	//	// top up for waiting porting v3
	//	case strconv.Itoa(metadata.PortalTopUpWaitingPortingRequestMetaV3):
	//		err = blockchain.processPortalTopUpWaitingPortingV3(portalStateDB, beaconHeight, inst, currentPortalState, portalParams)
	//
	//	// ============ Reward ============
	//	// portal reward
	//	case strconv.Itoa(metadata.PortalRewardMeta), strconv.Itoa(metadata.PortalRewardMetaV3):
	//		err = blockchain.processPortalReward(portalStateDB, beaconHeight, inst, currentPortalState, portalParams)
	//	// request withdraw reward
	//	case strconv.Itoa(metadata.PortalRequestWithdrawRewardMeta):
	//		err = blockchain.processPortalWithdrawReward(portalStateDB, beaconHeight, inst, currentPortalState, portalParams)
	//	// total custodian reward instruction
	//	case strconv.Itoa(metadata.PortalTotalRewardCustodianMeta):
	//		err = blockchain.processPortalTotalCustodianReward(portalStateDB, beaconHeight, inst, currentPortalState, portalParams)
	//
	//	// ============ Portal smart contract ============
	//	// todo: add more metadata need to unlock token from sc
	//	case strconv.Itoa(metadata.PortalCustodianWithdrawConfirmMetaV3),
	//		strconv.Itoa(metadata.PortalRedeemFromLiquidationPoolConfirmMetaV3),
	//		strconv.Itoa(metadata.PortalLiquidateRunAwayCustodianConfirmMetaV3):
	//		err = blockchain.processPortalConfirmWithdrawInstV3(portalStateDB, beaconHeight, inst, currentPortalState, portalParams)
	//	}
	//
	//	if err != nil {
	//		Logger.log.Error(err)
	//		return nil
	//	}
	//}
	//
	////save final exchangeRates
	//blockchain.pickExchangesRatesFinal(currentPortalState)
	//
	//// update info of bridge portal token
	//for _, updatingInfo := range updatingInfoByTokenID {
	//	var updatingAmt uint64
	//	var updatingType string
	//	if updatingInfo.countUpAmt > updatingInfo.deductAmt {
	//		updatingAmt = updatingInfo.countUpAmt - updatingInfo.deductAmt
	//		updatingType = "+"
	//	}
	//	if updatingInfo.countUpAmt < updatingInfo.deductAmt {
	//		updatingAmt = updatingInfo.deductAmt - updatingInfo.countUpAmt
	//		updatingType = "-"
	//	}
	//	err := statedb.UpdateBridgeTokenInfo(
	//		portalStateDB,
	//		updatingInfo.tokenID,
	//		updatingInfo.externalTokenID,
	//		updatingInfo.isCentralized,
	//		updatingAmt,
	//		updatingType,
	//	)
	//	if err != nil {
	//		return err
	//	}
	//}
	//
	//// store updated currentPortalState to leveldb with new beacon height
	//err = storePortalStateToDB(portalStateDB, currentPortalState)
	//if err != nil {
	//	Logger.log.Error(err)
	//}

	return nil
}


func (blockchain *BlockChain) processRelayingInstructions(block *BeaconBlock) error {
	//relayingState, err := blockchain.InitRelayingHeaderChainStateFromDB()
	//if err != nil {
	//	Logger.log.Error(err)
	//	return nil
	//}
	//
	//// because relaying instructions in received beacon block were sorted already as desired so dont need to do sorting again over here
	//for _, inst := range block.Body.Instructions {
	//	if len(inst) < 4 {
	//		continue // Not error, just not relaying instruction
	//	}
	//	var err error
	//	switch inst[0] {
	//	//case strconv.Itoa(metadata.RelayingBNBHeaderMeta):
	//	//	err = blockchain.processRelayingBNBHeaderInst(inst, relayingState)
	//	case strconv.Itoa(metadata.RelayingBTCHeaderMeta):
	//		err = blockchain.processRelayingBTCHeaderInst(inst, relayingState)
	//	}
	//	if err != nil {
	//		Logger.log.Error(err)
	//	}
	//}

	// store updated relayingState to leveldb with new beacon height
	//err = relayingState.BNBHeaderChain.StoreBNBChainState()
	//if err != nil {
	//	Logger.log.Error(err)
	//}
	return nil
}