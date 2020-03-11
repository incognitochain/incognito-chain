package blockchain

// beacon build instruction for portal liquidation when custodians run away - don't send public tokens back to users.
//todo:
func buildCustodianRunAwayLiquidationInst(
) []string {
	return []string{}
	//custodianDepositContent := metadata.PortalCustodianDepositContent{
	//	IncogAddressStr: custodianAddressStr,
	//	RemoteAddresses: remoteAddresses,
	//	DepositedAmount: depositedAmount,
	//	TxReqID:         txReqID,
	//	ShardID:         shardID,
	//}
	//custodianDepositContentBytes, _ := json.Marshal(custodianDepositContent)
	//return []string{
	//	strconv.Itoa(metaType),
	//	strconv.Itoa(int(shardID)),
	//	status,
	//	string(custodianDepositContentBytes),
	//}
}

func buildMinAspectRatioCollateralLiquidationInst(beaconHeight uint64, currentPortalState *CurrentPortalState)  ([]string, error) {
	if len(currentPortalState.CustodianPoolState) <= 0 {
		return []string{}, nil
	}

	/*keyExchangeRate := lvdb.NewFinalExchangeRatesKey(beaconHeight)
	exchangeRate, ok := currentPortalState.CustodianPoolState[keyExchangeRate]
	if !ok {
		return []string{}, nil
	}


	for i, v := range currentPortalState.CustodianPoolState {
		detectMinAspectRatio()
		v.HoldingPubTokens[]Op

		convertToPRV := exchangeRate.
	}*/

	return []string{}, nil
}
