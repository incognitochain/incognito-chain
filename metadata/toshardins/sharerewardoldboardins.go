package toshardins

import (
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
	"github.com/ninjadotorg/constant/metadata"
	"github.com/ninjadotorg/constant/privacy"
	"github.com/ninjadotorg/constant/transaction"
)

type ShareRewardOldBoardMetadataIns struct {
	chairPaymentAddress privacy.PaymentAddress
	voterPaymentAddress privacy.PaymentAddress
	boardType           byte
	amountOfCoin        uint64
	amountOfToken       uint64
}

func (shareRewardOldBoardMetadataIns *ShareRewardOldBoardMetadataIns) GetStringFormat() []string {
	panic("implement me")
}

func NewShareRewardOldBoardMetadataIns(
	chairPaymentAddress privacy.PaymentAddress,
	voterPaymentAddress privacy.PaymentAddress,
	boardType byte,
	amountOfCoin uint64,
	amountOfToken uint64,
) *ShareRewardOldBoardMetadataIns {
	return &ShareRewardOldBoardMetadataIns{
		chairPaymentAddress: chairPaymentAddress,
		voterPaymentAddress: voterPaymentAddress,
		boardType:           boardType,
		amountOfCoin:        amountOfCoin,
		amountOfToken:       amountOfToken,
	}
}

func (shareRewardOldBoardMetadataIns *ShareRewardOldBoardMetadataIns) BuildTransaction(
	minerPrivateKey *privacy.SpendingKey,
	db database.DatabaseInterface,
) metadata.Transaction {
	tx := transaction.Tx{}
	rewardShareOldBoardMeta := metadata.NewShareRewardOldBoardMetadata(
		shareRewardOldBoardMetadataIns.chairPaymentAddress,
		shareRewardOldBoardMetadataIns.voterPaymentAddress,
		shareRewardOldBoardMetadataIns.boardType,
	)
	tx.InitTxSalary(
		shareRewardOldBoardMetadataIns.amountOfCoin,
		&shareRewardOldBoardMetadataIns.voterPaymentAddress,
		minerPrivateKey,
		db,
		rewardShareOldBoardMeta,
	)
	var propertyID common.Hash
	if shareRewardOldBoardMetadataIns.boardType == common.DCBBoard {
		propertyID = common.DCBTokenID
	} else {
		propertyID = common.GOVTokenID
	}
	txTokenData := transaction.TxTokenData{
		Type:       transaction.CustomTokenInit,
		Amount:     shareRewardOldBoardMetadataIns.amountOfToken,
		PropertyID: propertyID,
		Vins:       []transaction.TxTokenVin{},
		Vouts:      []transaction.TxTokenVout{},
	}

	txCustomToken := transaction.TxCustomToken{
		Tx:          tx,
		TxTokenData: txTokenData,
	}
	return &txCustomToken
}
