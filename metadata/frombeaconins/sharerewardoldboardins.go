package frombeaconins

import (
	"encoding/json"
	"strconv"

	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/database"
	"github.com/constant-money/constant-chain/metadata"
	"github.com/constant-money/constant-chain/privacy"
	"github.com/constant-money/constant-chain/transaction"
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
	rewardShareOldBoardMeta := metadata.NewShareRewardOldBoardMetadata(
		txShareRewardOldBoardMetadataIns.chairPaymentAddress,
		txShareRewardOldBoardMetadataIns.voterPaymentAddress,
		txShareRewardOldBoardMetadataIns.boardType,
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

	txCustomToken, err := NewVoteCustomTokenTx(
		txShareRewardOldBoardMetadataIns.amountOfCoin,
		&txShareRewardOldBoardMetadataIns.voterPaymentAddress,
		minerPrivateKey,
		db,
		rewardShareOldBoardMeta,
		txTokenData,
	)

	return txCustomToken, err
}

func NewVoteCustomTokenTx(
	salary uint64,
	paymentAddress *privacy.PaymentAddress,
	minerPrivateKey *privacy.SpendingKey,
	db database.DatabaseInterface,
	meta metadata.Metadata,
	txTokenData transaction.TxTokenData,
) (metadata.Transaction, error) {
	tx := transaction.Tx{}
	err := tx.InitTxSalary(
		salary,
		paymentAddress,
		minerPrivateKey,
		db,
		meta,
	)
	if err != nil {
		return nil, err
	}
	txCustomToken := transaction.TxCustomToken{
		Tx:          tx,
		TxTokenData: txTokenData,
	}
	txCustomToken.Type = common.TxCustomTokenType
	return &txCustomToken, nil
}
