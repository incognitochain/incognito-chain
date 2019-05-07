package frombeaconins

import (
	"encoding/json"
	"strconv"

	"github.com/constant-money/constant-chain/blockchain/component"
	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/database"
	"github.com/constant-money/constant-chain/metadata"
	"github.com/constant-money/constant-chain/privacy"
	"github.com/constant-money/constant-chain/transaction"
)

type RewardProposalVoterIns struct {
	ReceiverAddress *privacy.PaymentAddress
	Amount          uint64
	BoardType       common.BoardType
}

func NewRewardProposalVoterIns(receiverAddress *privacy.PaymentAddress, amount uint64, boardType common.BoardType) *RewardProposalVoterIns {
	return &RewardProposalVoterIns{ReceiverAddress: receiverAddress, Amount: amount, BoardType: boardType}
}

func (rewardProposalVoterIns RewardProposalVoterIns) GetStringFormat() ([]string, error) {
	content, err := json.Marshal(rewardProposalVoterIns)
	if err != nil {
		return nil, err
	}
	shardID := GetShardIDFromPaymentAddressBytes(*rewardProposalVoterIns.ReceiverAddress)
	var metadataType int
	if rewardProposalVoterIns.BoardType == common.DCBBoard {
		metadataType = component.RewardDCBProposalVoterIns
	} else {
		metadataType = component.RewardGOVProposalVoterIns
	}
	return []string{
		strconv.Itoa(metadataType),
		strconv.Itoa(int(shardID)),
		string(content),
	}, nil
}

func (rewardProposalVoterIns RewardProposalVoterIns) BuildTransaction(
	minerPrivateKey *privacy.PrivateKey,
	db database.DatabaseInterface,
	boardType common.BoardType,
) (metadata.Transaction, error) {
	var meta metadata.Metadata
	if boardType == common.DCBBoard {
		meta = metadata.NewRewardDCBProposalVoterMetadata()
	} else {
		meta = metadata.NewRewardGOVProposalVoterMetadata()
	}
	tx := transaction.Tx{}
	receiverAddress := rewardProposalVoterIns.ReceiverAddress
	amount := rewardProposalVoterIns.Amount

	err := tx.InitTxSalary(amount, receiverAddress, minerPrivateKey, db, meta)
	return &tx, err
}
