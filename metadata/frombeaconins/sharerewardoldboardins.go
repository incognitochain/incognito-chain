package frombeaconins

import (
	"encoding/json"
	"strconv"

	"github.com/big0t/constant-chain/common"
	"github.com/big0t/constant-chain/database"
	"github.com/big0t/constant-chain/metadata"
	"github.com/big0t/constant-chain/privacy"
	"github.com/big0t/constant-chain/transaction"
)

type TxShareRewardOldBoardMetadataIns struct {
	chairPaymentAddress privacy.PaymentAddress
	voterPaymentAddress privacy.PaymentAddress
	boardType           common.BoardType
	amountOfCoin        uint64
	amountOfToken       uint64
}

func (txShareRewardOldBoardMetadataIns *TxShareRewardOldBoardMetadataIns) GetStringFormat() ([]string, error) {
	content, err := json.Marshal(txShareRewardOldBoardMetadataIns)
	if err != nil {
		return nil, err
	}
	shardID := GetShardIDFromPaymentAddressBytes(txShareRewardOldBoardMetadataIns.voterPaymentAddress)
	var metadataType int
	if txShareRewardOldBoardMetadataIns.boardType == common.DCBBoard {
		metadataType = metadata.ShareRewardOldDCBBoardMeta
	} else {
		metadataType = metadata.ShareRewardOldGOVBoardMeta
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
	amountOfToken uint64,
) *TxShareRewardOldBoardMetadataIns {
	return &TxShareRewardOldBoardMetadataIns{
		chairPaymentAddress: chairPaymentAddress,
		voterPaymentAddress: voterPaymentAddress,
		boardType:           boardType,
		amountOfCoin:        amountOfCoin,
		amountOfToken:       amountOfToken,
	}
}

func (txShareRewardOldBoardMetadataIns *TxShareRewardOldBoardMetadataIns) BuildTransaction(
	minerPrivateKey *privacy.SpendingKey,
	db database.DatabaseInterface,
) (metadata.Transaction, error) {
	tx := transaction.Tx{}
	rewardShareOldBoardMeta := metadata.NewShareRewardOldBoardMetadata(
		txShareRewardOldBoardMetadataIns.chairPaymentAddress,
		txShareRewardOldBoardMetadataIns.voterPaymentAddress,
		txShareRewardOldBoardMetadataIns.boardType,
	)
	tx.InitTxSalary(
		txShareRewardOldBoardMetadataIns.amountOfCoin,
		&txShareRewardOldBoardMetadataIns.voterPaymentAddress,
		minerPrivateKey,
		db,
		rewardShareOldBoardMeta,
	)
	var propertyID common.Hash
	if txShareRewardOldBoardMetadataIns.boardType == common.DCBBoard {
		propertyID = common.DCBTokenID
	} else {
		propertyID = common.GOVTokenID
	}
	txTokenData := transaction.TxTokenData{
		Type:       transaction.CustomTokenInit,
		Amount:     txShareRewardOldBoardMetadataIns.amountOfToken,
		PropertyID: propertyID,
		Vins:       []transaction.TxTokenVin{},
		Vouts:      []transaction.TxTokenVout{},
	}

	txCustomToken := transaction.TxCustomToken{
		Tx:          tx,
		TxTokenData: txTokenData,
	}
	return &txCustomToken, nil
}
