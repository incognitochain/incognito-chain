package frombeaconins

import (
	"encoding/json"
	"github.com/constant-money/constant-chain/blockchain/component"
	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/database"
	"github.com/constant-money/constant-chain/metadata"
	"github.com/constant-money/constant-chain/privacy"
	"github.com/constant-money/constant-chain/transaction"
	"strconv"
)

type RewardProposalSubmitterIns struct {
	ReceiverAddress *privacy.PaymentAddress
	Amount          uint64
	BoardType       common.BoardType
}

func NewRewardProposalSubmitterIns(receiverAddress *privacy.PaymentAddress, amount uint64, boardType common.BoardType) *RewardProposalSubmitterIns {
	return &RewardProposalSubmitterIns{ReceiverAddress: receiverAddress, Amount: amount, BoardType: boardType}
}

func (rewardProposalSubmitterIns RewardProposalSubmitterIns) GetStringFormat() ([]string, error) {
	content, err := json.Marshal(rewardProposalSubmitterIns)
	if err != nil {
		return nil, err
	}
	shardID := GetShardIDFromPaymentAddressBytes(*rewardProposalSubmitterIns.ReceiverAddress)
	var metadataType int
	if rewardProposalSubmitterIns.BoardType == common.DCBBoard {
		metadataType = component.RewardDCBProposalSubmitterIns
	} else {
		metadataType = component.RewardGOVProposalSubmitterIns
	}
	return []string{
		strconv.Itoa(metadataType),
		strconv.Itoa(int(shardID)),
		string(content),
	}, nil
}

func (rewardProposalSubmitterIns RewardProposalSubmitterIns) BuildTransaction(
	minerPrivateKey *privacy.SpendingKey,
	db database.DatabaseInterface,
	boardType common.BoardType,
) (metadata.Transaction, error) {
	var meta metadata.Metadata
	if boardType == common.DCBBoard {
		meta = metadata.NewRewardDCBProposalSubmitterMetadata()
	} else {
		meta = metadata.NewRewardGOVProposalSubmitterMetadata()
	}
	tx := transaction.Tx{}
	receiverAddress := rewardProposalSubmitterIns.ReceiverAddress
	amount := rewardProposalSubmitterIns.Amount

	err := tx.InitTxSalary(amount, receiverAddress, minerPrivateKey, db, meta)
	return &tx, err
}
