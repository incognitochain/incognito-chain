package frombeaconins

import (
	"github.com/constant-money/constant-chain/database"
	"github.com/constant-money/constant-chain/metadata"
	"github.com/constant-money/constant-chain/privacy"
	"github.com/constant-money/constant-chain/transaction"
)

type RewardProposalSubmitterIns struct {
	receiverAddress *privacy.PaymentAddress
	amount          uint64
}

func NewRewardProposalSubmitterIns(receiverAddress *privacy.PaymentAddress, amount uint64) *RewardProposalSubmitterIns {
	return &RewardProposalSubmitterIns{receiverAddress: receiverAddress, amount: amount}
}

func (rewardProposalSubmitterIns RewardProposalSubmitterIns) GetStringFormat() ([]string, error) {
	panic("implement me")
}

func (rewardProposalSubmitterIns RewardProposalSubmitterIns) BuildTransaction(
	minerPrivateKey *privacy.SpendingKey,
	db database.DatabaseInterface,
) (metadata.Transaction, error) {
	meta := metadata.NewRewardGOVProposalSubmitterMetadata()
	tx := transaction.Tx{}
	receiverAddress := rewardProposalSubmitterIns.receiverAddress
	amount := rewardProposalSubmitterIns.amount

	err := tx.InitTxSalary(amount, receiverAddress, minerPrivateKey, db, meta)
	return &tx, err
}
