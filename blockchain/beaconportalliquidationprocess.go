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

	// get pTokenID from actionData
	pTokenID := actionData.TokenID

	reqStatus := instructions[2]
	if reqStatus == common.PortalLiquidateCustodianSuccessChainStatus {
		// update custodian state (total collateral, holding public tokens, locked amount, free collateral)
		cusStateKey := lvdb.NewCustodianStateKey(beaconHeight, actionData.CustodianIncAddressStr)
		custodianState := currentPortalState.CustodianPoolState[cusStateKey]

		if custodianState.TotalCollateral < actionData.MintedCollateralAmount ||
			custodianState.LockedAmountCollateral[pTokenID] < actionData.MintedCollateralAmount {
			Logger.log.Errorf("[checkAndBuildInstForCustodianLiquidation] Total collateral %v, locked amount %v "+
				"should be greater than minted amount %v\n: ",
				custodianState.TotalCollateral, custodianState.LockedAmountCollateral[pTokenID], actionData.MintedCollateralAmount)
			return fmt.Errorf("[checkAndBuildInstForCustodianLiquidation] Total collateral %v, locked amount %v "+
				"should be greater than minted amount %v\n: ",
				custodianState.TotalCollateral, custodianState.LockedAmountCollateral[pTokenID], actionData.MintedCollateralAmount)
		}

		err = updateCustodianStateAfterLiquidateCustodian(custodianState, actionData.MintedCollateralAmount, pTokenID)
		if err != nil {
			Logger.log.Errorf("[checkAndBuildInstForCustodianLiquidation] Error when updating %v\n: ", err)
			return err
		}

		// remove matching custodian from matching custodians list in waiting redeem request
		waitingRedeemReqKey := lvdb.NewWaitingRedeemReqKey(beaconHeight, actionData.UniqueRedeemID)
		removeCustodianFromMatchingRedeemCustodians(currentPortalState.WaitingRedeemRequests[waitingRedeemReqKey].Custodians, actionData.CustodianIncAddressStr)

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
	err := json.Unmarshal([]byte(instructions[2]), &actionData)
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
		detectTPExchangeRates, err := detectTPRatio(custodianState.HoldingPubTokens, custodianState.LockedAmountCollateral, exchangeRate)
		if err != nil {
			Logger.log.Errorf("Detect tp ratio error %v", err)
			return nil
		}

		liquidateChange, err := GetLiquidateExchangeRatesChange(custodianState, detectTPExchangeRates)

		if err != nil {
			Logger.log.Errorf("ERROR: an error occurred while Get liquidate exchange rates change error %v", err)
			return nil
		}

		if len(liquidateChange) > 0 {
			for ptoken, liquidateTopPercentileExchangeRatesDetail := range liquidateChange {
				custodianState.LockedAmountCollateral[ptoken] = custodianState.LockedAmountCollateral[ptoken] - liquidateTopPercentileExchangeRatesDetail.HoldAmountFreeCollateral
				custodianState.HoldingPubTokens[ptoken] = custodianState.HoldingPubTokens[ptoken] - liquidateTopPercentileExchangeRatesDetail.HoldAmountPubToken
				custodianState.TotalCollateral = custodianState.TotalCollateral - liquidateTopPercentileExchangeRatesDetail.HoldAmountFreeCollateral
			}

			//update custodian
			currentPortalState.CustodianPoolState[cusStateKey] = custodianState

			//update
			liquidateExchangeRatesKey := lvdb.NewPortalLiquidateExchangeRatesKey(beaconHeight)
			liquidateExchangeRates, ok := currentPortalState.LiquidateExchangeRates[liquidateExchangeRatesKey]

			if !ok {
				item := make(map[string]lvdb.LiquidateExchangeRatesDetail)

				for ptoken, liquidateTopPercentileExchangeRatesDetail := range liquidateChange {
					item[ptoken] = lvdb.LiquidateExchangeRatesDetail{
						HoldAmountFreeCollateral: liquidateTopPercentileExchangeRatesDetail.HoldAmountFreeCollateral,
						HoldAmountPubToken: liquidateTopPercentileExchangeRatesDetail.HoldAmountPubToken,
					}
				}
				currentPortalState.LiquidateExchangeRates[liquidateExchangeRatesKey], _ = NewLiquidateExchangeRates(item)
			} else {
				for ptoken, liquidateTopPercentileExchangeRatesDetail := range liquidateChange {
					if _, ok := liquidateExchangeRates.Rates[ptoken]; !ok {
						liquidateExchangeRates.Rates[ptoken] = lvdb.LiquidateExchangeRatesDetail{
							HoldAmountFreeCollateral: liquidateTopPercentileExchangeRatesDetail.HoldAmountFreeCollateral,
							HoldAmountPubToken: liquidateTopPercentileExchangeRatesDetail.HoldAmountPubToken,
						}
					} else {
						liquidateExchangeRates.Rates[ptoken] = lvdb.LiquidateExchangeRatesDetail{
							HoldAmountFreeCollateral: liquidateExchangeRates.Rates[ptoken].HoldAmountFreeCollateral +  liquidateTopPercentileExchangeRatesDetail.HoldAmountFreeCollateral,
							HoldAmountPubToken: liquidateExchangeRates.Rates[ptoken].HoldAmountPubToken +  liquidateTopPercentileExchangeRatesDetail.HoldAmountPubToken,
						}
					}
				}

				currentPortalState.LiquidateExchangeRates[liquidateExchangeRatesKey] = liquidateExchangeRates
			}

			newTPKey := lvdb.NewPortalLiquidateTPExchangeRatesKey(beaconHeight, custodianState.IncognitoAddress)
			newTPExchangeRates, _ := NewLiquidateTopPercentileExchangeRates(
				custodianState.IncognitoAddress,
				liquidateChange,
				common.PortalLiquidateTPExchangeRatesSuccessChainStatus,
				)

			err := db.StoreLiquidateTopPercentileExchangeRates([]byte(newTPKey), newTPExchangeRates)
			if err != nil {
				Logger.log.Errorf("ERROR: an error occurred while store liquidation TP exchange rates %v", err)
				return nil
			}
		}
	} else if reqStatus == common.PortalLiquidateTPExchangeRatesFailedChainStatus {
		newTPKey := lvdb.NewPortalLiquidateTPExchangeRatesKey(beaconHeight, custodianState.IncognitoAddress)
		newTPExchangeRates, _ := NewLiquidateTopPercentileExchangeRates(
			custodianState.IncognitoAddress,
		nil,
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

func (blockchain *BlockChain) processPortalRedeemLiquidateExchangeRates(beaconHeight uint64, instructions []string, currentPortalState *CurrentPortalState, updatingInfoByTokenID map[common.Hash]UpdatingInfo) error {
	if currentPortalState == nil {
		Logger.log.Errorf("current portal state is nil")
		return nil
	}

	if len(instructions) != 4 {
		return nil // skip the instruction
	}

	// unmarshal instructions content
	var actionData metadata.PortalRedeemLiquidateExchangeRatesContent
	err := json.Unmarshal([]byte(instructions[3]), &actionData)
	if err != nil {
		Logger.log.Errorf("Can not unmarshal instruction content %v\n", err)
		return nil
	}


	db := blockchain.GetDatabase()

	reqStatus := instructions[2]
	if reqStatus == common.PortalRedeemLiquidateExchangeRatesSuccessChainStatus {
		keyExchangeRate := lvdb.NewFinalExchangeRatesKey(beaconHeight)
		_, ok := currentPortalState.FinalExchangeRates[keyExchangeRate]
		if !ok {
			Logger.log.Errorf("Exchange rate not found", err)
			return nil
		}

		liquidateExchangeRatesKey := lvdb.NewPortalLiquidateExchangeRatesKey(beaconHeight)
		liquidateExchangeRates, ok := currentPortalState.LiquidateExchangeRates[liquidateExchangeRatesKey]

		if !ok {
			Logger.log.Errorf("Liquidate exchange rates not found")
			return nil
		}

		liquidateByTokenID, ok := liquidateExchangeRates.Rates[actionData.TokenID]
		if !ok {
			Logger.log.Errorf("Liquidate exchange rates not found")
			return nil
		}

		totalPrv := actionData.TotalPTokenReceived

		liquidateExchangeRates.Rates[actionData.TokenID] = lvdb.LiquidateExchangeRatesDetail{
			HoldAmountFreeCollateral: liquidateByTokenID.HoldAmountFreeCollateral - totalPrv,
			HoldAmountPubToken: liquidateByTokenID.HoldAmountPubToken - actionData.RedeemAmount,
		}

		currentPortalState.LiquidateExchangeRates[liquidateExchangeRatesKey] = liquidateExchangeRates

		redeemKey := lvdb.NewRedeemLiquidateExchangeRatesKey(actionData.TxReqID.String())
		redeem, _ := NewRedeemLiquidateExchangeRates(
			actionData.TxReqID,
			actionData.TokenID,
			actionData.RedeemerIncAddressStr,
			actionData.RemoteAddress,
			actionData.RedeemAmount,
			actionData.RedeemFee,
			totalPrv,
			common.PortalRedeemLiquidateExchangeRatesSuccessStatus,
			)

		err = db.StoreRedeemLiquidationExchangeRates([]byte(redeemKey), redeem)
		if err != nil {
			Logger.log.Errorf("Store redeem liquidate exchange rates error %v\n", err)
			return nil
		}

		// update bridge/portal token info
		incTokenID, err := common.Hash{}.NewHashFromStr(actionData.TokenID)
		if err != nil {
			Logger.log.Errorf("ERROR: Can not new hash from porting incTokenID: %+v", err)
			return nil
		}
		updatingInfo, found := updatingInfoByTokenID[*incTokenID]
		if found {
			updatingInfo.deductAmt += actionData.RedeemAmount
		} else {
			updatingInfo = UpdatingInfo{
				countUpAmt:      0,
				deductAmt:       actionData.RedeemAmount,
				tokenID:         *incTokenID,
				externalTokenID: nil,
				isCentralized:   false,
			}
		}
		updatingInfoByTokenID[*incTokenID] = updatingInfo
	} else if reqStatus == common.PortalRedeemLiquidateExchangeRatesRejectedChainStatus {
		redeemKey := lvdb.NewRedeemLiquidateExchangeRatesKey(actionData.TxReqID.String())
		redeem, _ := NewRedeemLiquidateExchangeRates(
			actionData.TxReqID,
			actionData.TokenID,
			actionData.RedeemerIncAddressStr,
			actionData.RemoteAddress,
			actionData.RedeemAmount,
			actionData.RedeemFee,
			actionData.TotalPTokenReceived,
			common.PortalRedeemLiquidateExchangeRatesRejectedStatus,
		)

		err = db.StoreRedeemLiquidationExchangeRates([]byte(redeemKey), redeem)
		if err != nil {
			Logger.log.Errorf("Store redeem liquidate exchange rates error %v\n", err)
			return nil
		}
	}

	return nil
}
