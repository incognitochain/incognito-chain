package frombeaconins

import (
	"github.com/constant-money/constant-chain/database"
	"github.com/constant-money/constant-chain/metadata"
	"github.com/constant-money/constant-chain/privacy"
	"github.com/constant-money/constant-chain/transaction"
)

type RewardProposalWinnerIns struct {
	paymentAddress privacy.PaymentAddress
	prize          uint32
}

func NewRewardProposalWinnerIns(paymentAddress privacy.PaymentAddress, prize uint32) *RewardProposalWinnerIns {
	return &RewardProposalWinnerIns{paymentAddress: paymentAddress, prize: prize}
}

func (rewardProposalWinnerIns RewardProposalWinnerIns) GetStringFormat() ([]string, error) {
	panic("implement me")
}

func (rewardProposalWinnerIns RewardProposalWinnerIns) BuildTransaction(
	minerPrivateKey *privacy.SpendingKey,
	db database.DatabaseInterface,
) (metadata.Transaction, error) {
	meta := metadata.NewRewardProposalWinnerMetadata(rewardProposalWinnerIns.paymentAddress, rewardProposalWinnerIns.prize)
	tx := transaction.NewEmptyTx(minerPrivateKey, db, meta)
	return tx, nil
}
