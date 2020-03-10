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

func buildMinAspectRatioCollateralLiquidationInst()  []string {
	return []string{}
}
