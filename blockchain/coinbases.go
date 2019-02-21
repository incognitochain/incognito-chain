package blockchain

// func (blockgen *BlkTmplGenerator) buildCoinbases(
// 	shardID byte,
// 	privatekey *privacy.SpendingKey,
// 	txGroups *txGroups,
// 	salaryTx metadata.Transaction,
// ) ([]metadata.Transaction, error) {

// 	prevBlock := blockgen.chain.BestState[shardID].BestBlock
// 	dcbHelper := DCBConstitutionHelper{}
// 	govHelper := GOVConstitutionHelper{}
// 	coinbases := []metadata.Transaction{salaryTx}

// 	for _, tx := range txGroups.unlockTxs {
// 		coinbases = append(coinbases, tx)
// 	}
// 	for _, resTx := range txGroups.buySellResTxs {
// 		coinbases = append(coinbases, resTx)
// 	}
// 	for _, resTx := range txGroups.buyGOVTokensResTxs {
// 		coinbases = append(coinbases, resTx)
// 	}
// 	for _, resTx := range txGroups.buyBackResTxs {
// 		coinbases = append(coinbases, resTx)
// 	}
// 	for _, resTx := range txGroups.issuingResTxs {
// 		coinbases = append(coinbases, resTx)
// 	}
// 	for _, refundTx := range txGroups.refundTxs {
// 		coinbases = append(coinbases, refundTx)
// 	}
// 	for _, oracleRewardTx := range txGroups.oracleRewardTxs {
// 		coinbases = append(coinbases, oracleRewardTx)
// 	}
// 	return coinbases, nil
// }
