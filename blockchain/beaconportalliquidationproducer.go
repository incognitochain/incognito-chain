package blockchain

import (
	"encoding/json"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/database/lvdb"
	"github.com/incognitochain/incognito-chain/metadata"
	"strconv"
)

// beacon build instruction for portal liquidation when custodians run away - don't send public tokens back to users.
//todo:
func buildCustodianRunAwayLiquidationInst(
	redeemID string,
	tokenID string,
	redeemPubTokenAmount uint64,
	mintedCollateralAmount uint64,
	redeemerIncAddrStr string,
	custodianIncAddrStr string,
	metaType int,
	status string,
) []string {
	liqCustodianContent := metadata.PortalLiquidateCustodianContent{
		UniqueRedeemID:         redeemID,
		TokenID:                tokenID,
		RedeemPubTokenAmount:   redeemPubTokenAmount,
		MintedCollateralAmount: mintedCollateralAmount,
		RedeemerIncAddressStr:  redeemerIncAddrStr,
		CustodianIncAddressStr: custodianIncAddrStr,
	}
	liqCustodianContentBytes, _ := json.Marshal(liqCustodianContent)
	return []string{
		strconv.Itoa(metaType),
		strconv.Itoa(-1), // shardID
		status,
		string(liqCustodianContentBytes),
	}
}

func checkAndBuildInstForCustodianLiquidation(
	beaconHeight uint64,
	currentPortalState *CurrentPortalState,
) ([][]string, error) {

	insts := [][]string{}
	for _, redeemReq := range currentPortalState.WaitingRedeemRequests {
		if beaconHeight-redeemReq.BeaconHeight >= common.PortalTimeOutCustodianSendPubTokenBack {
			for cusIncAddr, matchCusDetail := range redeemReq.Custodians {
				// calculate minted collateral amount
				mintedAmount := matchCusDetail.Amount * common.PercentReceivedCollateralAmount / 100

				// update waiting redeem request (remove custodian from matching custodians list)
				//delete(redeemReq.Custodians, cusIncAddr)

				// update custodian state (total collateral, holding public tokens, locked amount, free collateral)
				// get tokenSymbol from redeemTokenID
				tokenSymbol := ""
				for tokenSym, incTokenID := range metadata.PortalSupportedTokenMap {
					if incTokenID == redeemReq.TokenID {
						tokenSymbol = tokenSym
						break
					}
				}
				cusStateKey := lvdb.NewCustodianStateKey(beaconHeight, cusIncAddr)
				currentPortalState.CustodianPoolState[cusStateKey].TotalCollateral -= mintedAmount
				currentPortalState.CustodianPoolState[cusStateKey].HoldingPubTokens[tokenSymbol] -= matchCusDetail.Amount

				if currentPortalState.CustodianPoolState[cusStateKey].HoldingPubTokens[tokenSymbol] > 0 {
					currentPortalState.CustodianPoolState[cusStateKey].LockedAmountCollateral[tokenSymbol] -= mintedAmount
				} else {
					unlockedCollateralAmount := currentPortalState.CustodianPoolState[cusStateKey].LockedAmountCollateral[tokenSymbol] - mintedAmount
					currentPortalState.CustodianPoolState[cusStateKey].FreeCollateral += unlockedCollateralAmount
					currentPortalState.CustodianPoolState[cusStateKey].LockedAmountCollateral[tokenSymbol] = 0
				}

				// build instruction
				inst := buildCustodianRunAwayLiquidationInst(
					redeemReq.UniqueRedeemID,
					redeemReq.TokenID,
					matchCusDetail.Amount,
					mintedAmount,
					redeemReq.RedeemerAddress,
					cusIncAddr,
					metadata.PortalLiquidateCustodianMeta,
					"",
				)
				insts = append(insts, inst)
			}
		}
	}

	return insts, nil
}

func buildMinAspectRatioCollateralLiquidationInst() []string {
	return []string{}
}
