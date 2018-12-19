package blockchain

// type ConstitutionInfo struct {
// 	StartedBlockHeight uint32
// 	ExecuteDuration    uint32
// 	ProposalTXID       common.Hash
// }

// func (dcbConstitution DCBConstitution) GetEndedBlockHeight() uint32 {
// 	return dcbConstitution.StartedBlockHeight + dcbConstitution.ExecuteDuration
// }

// func (govConstitution GOVConstitution) GetEndedBlockHeight() uint32 {
// 	return govConstitution.StartedBlockHeight + govConstitution.ExecuteDuration
// }

// type DCBConstitution struct {
// 	ConstitutionInfo
// 	CurrentDCBNationalWelfare int32
// 	DCBParams                 params.DCBParams
// }

// type DCBConstitutionHelper struct{}
// type GOVConstitutionHelper struct{}

// func (DCBConstitutionHelper) GetEndedBlockHeight(blockgen *BlkTmplGenerator, shardID byte) uint32 {
// 	BestBlock := blockgen.chain.BestState[shardID].BestBlock
// 	lastDCBConstitution := BestBlock.Header.DCBConstitution
// 	return lastDCBConstitution.StartedBlockHeight + lastDCBConstitution.ExecuteDuration
// }

// func (GOVConstitutionHelper) GetEndedBlockHeight(blockgen *BlkTmplGenerator, shardID byte) uint32 {
// 	BestBlock := blockgen.chain.BestState[shardID].BestBlock
// 	lastGOVConstitution := BestBlock.Header.GOVConstitution
// 	return lastGOVConstitution.StartedBlockHeight + lastGOVConstitution.ExecuteDuration
// }

// func (DCBConstitutionHelper) GetStartedNormalVote(blockgen *BlkTmplGenerator, shardID byte) uint32 {
// 	BestBlock := blockgen.chain.BestState[shardID].BestBlock
// 	lastDCBConstitution := BestBlock.Header.DCBConstitution
// 	return lastDCBConstitution.StartedBlockHeight - common.EncryptionPhaseDuration
// }

// // func (DCBConstitutionHelper) CheckVotingProposalType(tx metadata.Transaction) bool {
// // 	return tx.GetMetadataType() == metadata.VoteDCBProposalMeta
// // }

// // func (DCBConstitutionHelper) GetAmountVoteToken(tx metadata.Transaction) uint64 {
// // 	return tx.(*transaction.TxCustomToken).GetAmountOfVote()
// // }

// // func (GOVConstitutionHelper) GetStartedNormalVote(blockgen *BlkTmplGenerator, shardID byte) int32 {
// // 	BestBlock := blockgen.chain.BestState[shardID].BestBlock
// // 	lastGOVConstitution := BestBlock.Header.GOVConstitution
// // 	return lastGOVConstitution.StartedBlockHeight - common.EncryptionPhaseDuration
// // }

// func (GOVConstitutionHelper) GetStartedNormalVote(blockgen *BlkTmplGenerator, shardID byte) uint32 {
// 	BestBlock := blockgen.chain.BestState[shardID].BestBlock
// 	lastGOVConstitution := BestBlock.Header.GOVConstitution
// 	return lastGOVConstitution.StartedBlockHeight - common.EncryptionPhaseDuration
// }

// // func (GOVConstitutionHelper) CheckVotingProposalType(tx metadata.Transaction) bool {
// // 	return tx.GetMetadataType() == metadata.VoteGOVProposalMeta
// // }

// // func (GOVConstitutionHelper) GetAmountVoteToken(tx metadata.Transaction) uint64 {
// // 	return tx.(*transaction.TxCustomToken).GetAmountOfVote()
// // }

// // func (DCBConstitutionHelper) TxAcceptProposal(originTx metadata.Transaction) metadata.Transaction {
// // 	acceptTx := transaction.Tx{
// // 		Metadata: &metadata.AcceptDCBProposalMetadata{
// // 			DCBProposalTXID: *originTx.Hash(),
// // 		},
// // 	}
// // 	return &acceptTx
// // }

// func (DCBConstitutionHelper) TxAcceptProposal(txId *common.Hash) metadata.Transaction {
// 	acceptTx := transaction.Tx{
// 		Metadata: &metadata.AcceptDCBProposalMetadata{
// 			DCBProposalTXID: *txId,
// 		},
// 	}
// 	return &acceptTx
// }

// func (GOVConstitutionHelper) TxAcceptProposal(txId *common.Hash) metadata.Transaction {
// 	acceptTx := transaction.Tx{
// 		Metadata: &metadata.AcceptGOVProposalMetadata{
// 			GOVProposalTXID: *txId,
// 		},
// 	}
// 	return &acceptTx
// }

// func (DCBConstitutionHelper) GetLowerCaseBoardType() string {
// 	return "dcb"
// }

// func (GOVConstitutionHelper) GetLowerCaseBoardType() string {
// 	return "gov"
// }
