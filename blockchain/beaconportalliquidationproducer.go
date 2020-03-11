package blockchain

import (
	"encoding/json"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/database/lvdb"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/wallet"
	"strconv"
)

// beacon build instruction for portal liquidation when custodians run away - don't send public tokens back to users.
func buildCustodianRunAwayLiquidationInst(
	redeemID string,
	tokenID string,
	redeemPubTokenAmount uint64,
	mintedCollateralAmount uint64,
	redeemerIncAddrStr string,
	custodianIncAddrStr string,
	metaType int,
	shardID byte,
	status string,
) []string {
	liqCustodianContent := metadata.PortalLiquidateCustodianContent{
		UniqueRedeemID:         redeemID,
		TokenID:                tokenID,
		RedeemPubTokenAmount:   redeemPubTokenAmount,
		MintedCollateralAmount: mintedCollateralAmount,
		RedeemerIncAddressStr:  redeemerIncAddrStr,
		CustodianIncAddressStr: custodianIncAddrStr,
		ShardID:                shardID,
	}
	liqCustodianContentBytes, _ := json.Marshal(liqCustodianContent)
	return []string{
		strconv.Itoa(metaType),
		strconv.Itoa(int(shardID)),
		status,
		string(liqCustodianContentBytes),
	}
}

func checkAndBuildInstForCustodianLiquidation(
	beaconHeight uint64,
	currentPortalState *CurrentPortalState,
) ([][]string, error) {

	insts := [][]string{}

	// get exchange rate
	exchangeRateKey := lvdb.NewFinalExchangeRatesKey(beaconHeight)
	exchangeRate := currentPortalState.FinalExchangeRates[exchangeRateKey]
	if exchangeRate == nil {
		Logger.log.Errorf("[checkAndBuildInstForCustodianLiquidation] Error when get exchange rate")
	}

	for redeemReqKey, redeemReq := range currentPortalState.WaitingRedeemRequests {
		if beaconHeight - redeemReq.BeaconHeight >= common.PortalTimeOutCustodianSendPubTokenBack {
			// get shardId of redeemer
			redeemerKey, err := wallet.Base58CheckDeserialize(redeemReq.RedeemerAddress)
			if err != nil {
				Logger.log.Errorf("[checkAndBuildInstForCustodianLiquidation] Error when deserializing redeemer address string in redeemID %v - %v\n: ",
					redeemReq.UniqueRedeemID, err)
				continue
			}
			shardID := common.GetShardIDFromLastByte(redeemerKey.KeySet.PaymentAddress.Pk[len(redeemerKey.KeySet.PaymentAddress.Pk)-1])

			// get tokenSymbol from redeemTokenID
			tokenSymbol := ""
			for tokenSym, incTokenID := range metadata.PortalSupportedTokenMap {
				if incTokenID == redeemReq.TokenID {
					tokenSymbol = tokenSym
					break
				}
			}

			for cusIncAddr, matchCusDetail := range redeemReq.Custodians {
				// calculate minted collateral amount
				mintedAmountInPToken := matchCusDetail.Amount * common.PercentReceivedCollateralAmount / 100
				mintedAmountInPRV, err := exchangeRate.ExchangePToken2PRVByTokenId(tokenSymbol, mintedAmountInPToken)
				if err != nil {
					Logger.log.Errorf("[checkAndBuildInstForCustodianLiquidation] Error when exchanging ptoken to prv amount %v\n: ", err)
					inst := buildCustodianRunAwayLiquidationInst(
						redeemReq.UniqueRedeemID,
						redeemReq.TokenID,
						matchCusDetail.Amount,
						0,
						redeemReq.RedeemerAddress,
						cusIncAddr,
						metadata.PortalLiquidateCustodianMeta,
						shardID,
						common.PortalLiquidateCustodianFailedChainStatus,
					)
					insts = append(insts, inst)
					continue
				}

				// update custodian state (total collateral, holding public tokens, locked amount, free collateral)
				cusStateKey := lvdb.NewCustodianStateKey(beaconHeight, cusIncAddr)
				custodianState := currentPortalState.CustodianPoolState[cusStateKey]

				if custodianState.TotalCollateral < mintedAmountInPRV ||
					custodianState.LockedAmountCollateral[tokenSymbol] < mintedAmountInPRV {
					Logger.log.Errorf("[checkAndBuildInstForCustodianLiquidation] Total collateral %v, locked amount %v " +
						"should be greater than minted amount %v\n: ",
						custodianState.TotalCollateral, custodianState.LockedAmountCollateral[tokenSymbol], mintedAmountInPRV)
					inst := buildCustodianRunAwayLiquidationInst(
						redeemReq.UniqueRedeemID,
						redeemReq.TokenID,
						matchCusDetail.Amount,
						mintedAmountInPRV,
						redeemReq.RedeemerAddress,
						cusIncAddr,
						metadata.PortalLiquidateCustodianMeta,
						shardID,
						common.PortalLiquidateCustodianFailedChainStatus,
					)
					insts = append(insts, inst)
					continue
				}

				err = updateCustodianStateAfterLiquidateCustodian(custodianState, mintedAmountInPRV, tokenSymbol)
				if err != nil {
					Logger.log.Errorf("[checkAndBuildInstForCustodianLiquidation] Error when updating %v\n: ", err)
					inst := buildCustodianRunAwayLiquidationInst(
						redeemReq.UniqueRedeemID,
						redeemReq.TokenID,
						matchCusDetail.Amount,
						mintedAmountInPRV,
						redeemReq.RedeemerAddress,
						cusIncAddr,
						metadata.PortalLiquidateCustodianMeta,
						shardID,
						common.PortalLiquidateCustodianFailedChainStatus,
					)
					insts = append(insts, inst)
					continue
				}

				// remove matching custodian from matching custodians list in waiting redeem request
				delete(currentPortalState.WaitingRedeemRequests[redeemReqKey].Custodians, cusIncAddr)

				// build instruction
				inst := buildCustodianRunAwayLiquidationInst(
					redeemReq.UniqueRedeemID,
					redeemReq.TokenID,
					matchCusDetail.Amount,
					mintedAmountInPRV,
					redeemReq.RedeemerAddress,
					cusIncAddr,
					metadata.PortalLiquidateCustodianMeta,
					shardID,
					common.PortalLiquidateCustodianSuccessChainStatus,
				)
				insts = append(insts, inst)
			}

			// remove redeem request from waiting redeem requests list
			if len(currentPortalState.WaitingRedeemRequests[redeemReqKey].Custodians) == 0 {
				delete(currentPortalState.WaitingRedeemRequests, redeemReqKey)
			}
		}
	}

	return insts, nil
}

func buildMinAspectRatioCollateralLiquidationInst() []string {
	return []string{}
}
