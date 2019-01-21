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
// 	// Voting transaction
// 	// Check if it is the case we need to apply a new proposal
// 	// 1. newNW < lastNW * 0.9
// 	// 2. current block height == last Constitution start time + last Constitution execute duration
// 	if blockgen.chain.readyNewConstitution(dcbHelper) {
// 		blockgen.chain.config.DataBase.SetEncryptionLastBlockHeight(dcbHelper.GetBoardType(), uint32(prevBlock.Header.Height+1))
// 		blockgen.chain.config.DataBase.SetEncryptFlag(dcbHelper.GetBoardType(), uint32(common.Lv3EncryptionFlag))
// 		tx, err := blockgen.createAcceptConstitutionAndPunishTxAndRewardSubmitter(shardID, DCBConstitutionHelper{}, privatekey)
// 		coinbases = append(coinbases, tx...)
// 		if err != nil {
// 			Logger.log.Error(err)
// 			return nil, err
// 		}
// 		rewardTx, err := blockgen.createRewardProposalWinnerTx(shardID, DCBConstitutionHelper{})
// 		coinbases = append(coinbases, rewardTx)
// 	}
// 	if blockgen.chain.readyNewConstitution(govHelper) {
// 		blockgen.chain.config.DataBase.SetEncryptionLastBlockHeight(govHelper.GetBoardType(), uint32(prevBlock.Header.Height+1))
// 		blockgen.chain.config.DataBase.SetEncryptFlag(govHelper.GetBoardType(), uint32(common.Lv3EncryptionFlag))
// 		tx, err := blockgen.createAcceptConstitutionAndPunishTxAndRewardSubmitter(shardID, GOVConstitutionHelper{}, privatekey)
// 		coinbases = append(coinbases, tx...)
// 		if err != nil {
// 			Logger.log.Error(err)
// 			return nil, err
// 		}
// 		rewardTx, err := blockgen.createRewardProposalWinnerTx(shardID, GOVConstitutionHelper{})
// 		coinbases = append(coinbases, rewardTx)
// 	}

// 	if blockgen.neededNewDCBGovernor(shardID) {
// 		coinbases = append(coinbases, blockgen.UpdateNewGovernor(DCBConstitutionHelper{}, shardID, privatekey)...)
// 	}
// 	if blockgen.neededNewGOVGovernor(shardID) {
// 		coinbases = append(coinbases, blockgen.UpdateNewGovernor(GOVConstitutionHelper{}, shardID, privatekey)...)
// 	}

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
