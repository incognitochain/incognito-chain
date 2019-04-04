package frombeaconins

//type RewardProposalWinnerIns struct {
//	PaymentAddress privacy.PaymentAddress
//	Prize          uint32
//}
//
//func NewRewardProposalWinnerIns(paymentAddress privacy.PaymentAddress, prize uint32) *RewardProposalWinnerIns {
//	return &RewardProposalWinnerIns{PaymentAddress: paymentAddress, Prize: prize}
//}
//
//func (rewardProposalWinnerIns RewardProposalWinnerIns) GetStringFormat() ([]string, error) {
//	panic("implement me")
//}
//
//func (rewardProposalWinnerIns RewardProposalWinnerIns) BuildTransaction(
//	minerPrivateKey *privacy.PrivateKey,
//	db database.DatabaseInterface,
//) (metadata.Transaction, error) {
//	meta := metadata.NewRewardProposalWinnerMetadata(rewardProposalWinnerIns.PaymentAddress, rewardProposalWinnerIns.Prize)
//	tx := transaction.NewEmptyTx(minerPrivateKey, db, meta)
//	return tx, nil
//}
//r
