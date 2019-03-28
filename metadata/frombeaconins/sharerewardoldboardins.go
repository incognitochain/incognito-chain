package frombeaconins

import (
	"encoding/json"
	"github.com/constant-money/constant-chain/blockchain/component"
	"strconv"

	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/database"
	"github.com/constant-money/constant-chain/metadata"
	"github.com/constant-money/constant-chain/privacy"
	"github.com/constant-money/constant-chain/transaction"
)

type ShareRewardOldBoardIns struct {
	ChairPaymentAddress privacy.PaymentAddress
	VoterPaymentAddress privacy.PaymentAddress
	BoardType           common.BoardType
	AmountOfCoin        uint64
}

func (shareRewardOldBoardIns *ShareRewardOldBoardIns) GetStringFormat() ([]string, error) {
	content, err := json.Marshal(shareRewardOldBoardIns)
	if err != nil {
		return nil, err
	}
	shardID := GetShardIDFromPaymentAddressBytes(shareRewardOldBoardIns.VoterPaymentAddress)
	var metadataType int
	if shareRewardOldBoardIns.BoardType == common.DCBBoard {
		metadataType = component.ShareRewardOldDCBBoardIns
	} else {
		metadataType = component.ShareRewardOldGOVBoardIns
	}
	return []string{
		strconv.Itoa(metadataType),
		strconv.Itoa(int(shardID)),
		string(content),
	}, nil
}

func NewShareRewardOldBoardMetadataIns(
	chairPaymentAddress privacy.PaymentAddress,
	voterPaymentAddress privacy.PaymentAddress,
	boardType common.BoardType,
	amountOfCoin uint64,
) *ShareRewardOldBoardIns {
	return &ShareRewardOldBoardIns{
		ChairPaymentAddress: chairPaymentAddress,
		VoterPaymentAddress: voterPaymentAddress,
		BoardType:           boardType,
		AmountOfCoin:        amountOfCoin,
	}
}

func (shareRewardOldBoardIns *ShareRewardOldBoardIns) BuildTransaction(
	minerPrivateKey *privacy.SpendingKey,
	db database.DatabaseInterface,
) (metadata.Transaction, error) {
	rewardShareOldBoardMeta := metadata.NewShareRewardOldBoardMetadata(
		shareRewardOldBoardIns.ChairPaymentAddress,
		shareRewardOldBoardIns.VoterPaymentAddress,
		shareRewardOldBoardIns.BoardType,
	)
	tx := transaction.Tx{}
	err := tx.InitTxSalary(
		shareRewardOldBoardIns.AmountOfCoin,
		&shareRewardOldBoardIns.VoterPaymentAddress,
		minerPrivateKey,
		db,
		rewardShareOldBoardMeta,
	)
	if err != nil {
		return nil, err
	}
	return &tx, nil
}
