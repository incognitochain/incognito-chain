package blockchain

import (
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/database/lvdb"
	"github.com/incognitochain/incognito-chain/metadata"
)

func (blockchain *BlockChain) processPortalLiquidateCustodian(
	beaconHeight uint64, instructions []string,
	currentPortalState *CurrentPortalState) error {

	db := blockchain.GetDatabase()

	// unmarshal instructions content
	var actionData metadata.PortalLiquidateCustodianContent
	err := json.Unmarshal([]byte(instructions[3]), &actionData)
	if err != nil {
		Logger.log.Errorf("Can not unmarshal instruction content %v\n", err)
		return nil
	}

	// get tokenSymbol from redeemTokenID
	tokenSymbol := ""
	for tokenSym, incTokenID := range metadata.PortalSupportedTokenMap {
		if incTokenID == actionData.TokenID {
			tokenSymbol = tokenSym
			break
		}
	}

	reqStatus := instructions[2]
	if reqStatus == common.PortalLiquidateCustodianSuccessChainStatus {
		// update custodian state (total collateral, holding public tokens, locked amount, free collateral)
		cusStateKey := lvdb.NewCustodianStateKey(beaconHeight, actionData.CustodianIncAddressStr)
		custodianState := currentPortalState.CustodianPoolState[cusStateKey]

		if custodianState.TotalCollateral < actionData.MintedCollateralAmount ||
			custodianState.LockedAmountCollateral[tokenSymbol] < actionData.MintedCollateralAmount {
			Logger.log.Errorf("[checkAndBuildInstForCustodianLiquidation] Total collateral %v, locked amount %v "+
				"should be greater than minted amount %v\n: ",
				custodianState.TotalCollateral, custodianState.LockedAmountCollateral[tokenSymbol], actionData.MintedCollateralAmount)
			return fmt.Errorf("[checkAndBuildInstForCustodianLiquidation] Total collateral %v, locked amount %v "+
				"should be greater than minted amount %v\n: ",
				custodianState.TotalCollateral, custodianState.LockedAmountCollateral[tokenSymbol], actionData.MintedCollateralAmount)
		}

		err = updateCustodianStateAfterLiquidateCustodian(custodianState, actionData.MintedCollateralAmount, tokenSymbol)
		if err != nil {
			Logger.log.Errorf("[checkAndBuildInstForCustodianLiquidation] Error when updating %v\n: ", err)
			return err
		}

		// remove matching custodian from matching custodians list in waiting redeem request
		waitingRedeemReqKey := lvdb.NewWaitingRedeemReqKey(beaconHeight, actionData.UniqueRedeemID)
		delete(currentPortalState.WaitingRedeemRequests[waitingRedeemReqKey].Custodians, actionData.CustodianIncAddressStr)

		// remove redeem request from waiting redeem requests list
		if len(currentPortalState.WaitingRedeemRequests[waitingRedeemReqKey].Custodians) == 0 {
			delete(currentPortalState.WaitingRedeemRequests, waitingRedeemReqKey)

			// update status of redeem request with redeemID to liquidated status
			err = updateRedeemRequestStatusByRedeemId(actionData.UniqueRedeemID, common.PortalRedeemReqLiquidatedStatus, db)
			if err != nil {
				Logger.log.Errorf("ERROR: an error occurred while updating redeem request status by redeemID: %+v", err)
				return nil
			}
		}

		// track liquidation custodian status by redeemID and custodian address into DB
		custodianLiquidationTrackKey := lvdb.NewPortalLiquidationCustodianKey(actionData.UniqueRedeemID, actionData.CustodianIncAddressStr)
		custodianLiquidationTrackData := metadata.PortalLiquidateCustodianStatus{
			Status:                 common.PortalLiquidateCustodianSuccessStatus,
			UniqueRedeemID:         actionData.UniqueRedeemID,
			TokenID:                actionData.TokenID,
			RedeemPubTokenAmount:   actionData.RedeemPubTokenAmount,
			MintedCollateralAmount: actionData.MintedCollateralAmount,
			RedeemerIncAddressStr:  actionData.RedeemerIncAddressStr,
			CustodianIncAddressStr: actionData.CustodianIncAddressStr,
			ShardID:                actionData.ShardID,
			LiquidatedBeaconHeight: beaconHeight + 1,
		}
		custodianLiquidationTrackDataBytes, _ := json.Marshal(custodianLiquidationTrackData)
		err = db.TrackLiquidateCustodian(
			[]byte(custodianLiquidationTrackKey),
			custodianLiquidationTrackDataBytes,
		)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while tracking liquidation custodian: %+v", err)
			return nil
		}

	} else if reqStatus == common.PortalLiquidateCustodianFailedChainStatus {
		// track liquidation custodian status by redeemID and custodian address into DB
		custodianLiquidationTrackKey := lvdb.NewPortalLiquidationCustodianKey(actionData.UniqueRedeemID, actionData.CustodianIncAddressStr)
		custodianLiquidationTrackData := metadata.PortalLiquidateCustodianStatus{
			Status:                 common.PortalLiquidateCustodianFailedStatus,
			UniqueRedeemID:         actionData.UniqueRedeemID,
			CustodianIncAddressStr: actionData.CustodianIncAddressStr,
			LiquidatedBeaconHeight: beaconHeight + 1,
		}
		custodianLiquidationTrackDataBytes, _ := json.Marshal(custodianLiquidationTrackData)
		err = db.TrackLiquidateCustodian(
			[]byte(custodianLiquidationTrackKey),
			custodianLiquidationTrackDataBytes,
		)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while tracking liquidation custodian: %+v", err)
			return nil
		}
	}

	return nil
}

func (blockchain *BlockChain) processLiquidationTopPercentileExchangeRates(beaconHeight uint64, instructions []string,
currentPortalState *CurrentPortalState) error {
	db := blockchain.GetDatabase()

	// unmarshal instructions content
	var actionData metadata.PortalLiquidateTopPercentileExchangeRatesContent
	err := json.Unmarshal([]byte(instructions[3]), &actionData)
	if err != nil {
		Logger.log.Errorf("Can not unmarshal instruction content %v\n", err)
		return nil
	}


	keyExchangeRate := lvdb.NewFinalExchangeRatesKey(beaconHeight)
	exchangeRate, ok := currentPortalState.FinalExchangeRates[keyExchangeRate]
	if !ok {
		Logger.log.Errorf("Exchange rate not found", err)
		return nil
	}

	cusStateKey := lvdb.NewCustodianStateKey(beaconHeight, actionData.CustodianAddress)
	custodianState, ok := currentPortalState.CustodianPoolState[cusStateKey]
	//todo: check custodian exist on db
	if !ok {
		Logger.log.Errorf("Custodian not found")
		return nil
	}

	reqStatus := instructions[2]
	if reqStatus == common.PortalLiquidateTPExchangeRatesSuccessChainStatus {
		//validation
		result, err := detectMinAspectRatio(custodianState.HoldingPubTokens, custodianState.LockedAmountCollateral, exchangeRate)
		if err != nil {
			Logger.log.Errorf("Detect min aspect ratio error %v", err)
			return nil
		}

		tpList := make(map[string]int)
		newCustodian := custodianState

		for ptoken, tpValue := range result {
			if isTp120, ok := isTP120(tpValue); ok {
				tpList[ptoken] = tpValue
				if isTp120 {
					//update custodian
					newCustodian.LockedAmountCollateral[ptoken] = 0
					newCustodian.HoldingPubTokens[ptoken] = 0
					newCustodian.TotalCollateral = newCustodian.TotalCollateral - custodianState.LockedAmountCollateral[ptoken]

					//todo: update final liquidate
					//load liquidate
					//save liquidate
				}
			}
		}

		if len(tpList) > 0 {
			newTPKey := lvdb.NewPortalLiquidateTPExchangeRatesKey(beaconHeight, custodianState.IncognitoAddress)
			newTPExchangeRates, _ := NewLiquidateTopPercentileExchangeRates(
				tpList,
				newCustodian.IncognitoAddress,
				newCustodian.LockedAmountCollateral,
				newCustodian.HoldingPubTokens,
				*exchangeRate,
				common.PortalLiquidateTPExchangeRatesSuccessChainStatus,
				)

			err := db.StoreLiquidateTopPercentileExchangeRates([]byte(newTPKey), newTPExchangeRates)
			if err != nil {
				Logger.log.Errorf("ERROR: an error occurred while store liquidation TP exchange rates %v", err)
				return nil
			}

			currentPortalState.CustodianPoolState[cusStateKey] = newCustodian
		}
	} else if reqStatus == common.PortalLiquidateTPExchangeRatesFailedChainStatus {
		newTPKey := lvdb.NewPortalLiquidateTPExchangeRatesKey(beaconHeight, custodianState.IncognitoAddress)
		newTPExchangeRates, _ := NewLiquidateTopPercentileExchangeRates(
			nil,
			custodianState.IncognitoAddress,
			custodianState.LockedAmountCollateral,
			custodianState.HoldingPubTokens,
			*exchangeRate,
			common.PortalLiquidateTPExchangeRatesFailedChainStatus,
		)

		err := db.StoreLiquidateTopPercentileExchangeRates([]byte(newTPKey), newTPExchangeRates)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occurred while store liquidation TP exchange rates %v", err)
			return nil
		}
	}

	return nil
}
