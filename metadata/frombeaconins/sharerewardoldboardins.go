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
	ChairPaymentAddress privacy.PaymentAddress
	VoterPaymentAddress privacy.PaymentAddress
	BoardType           common.BoardType
	AmountOfCoin        uint64
	AmountOfToken       uint64
}

func (txShareRewardOldBoardMetadataIns *TxShareRewardOldBoardMetadataIns) GetStringFormat() ([]string, error) {
	content, err := json.Marshal(txShareRewardOldBoardMetadataIns)
	if err != nil {
		return nil, err
	}
	shardID := GetShardIDFromPaymentAddressBytes(txShareRewardOldBoardMetadataIns.VoterPaymentAddress)
	var metadataType int
	if txShareRewardOldBoardMetadataIns.BoardType == common.DCBBoard {
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
		ChairPaymentAddress: chairPaymentAddress,
		VoterPaymentAddress: voterPaymentAddress,
		BoardType:           boardType,
		AmountOfCoin:        amountOfCoin,
		AmountOfToken:       amountOfToken,
	}
}

func (txShareRewardOldBoardMetadataIns *TxShareRewardOldBoardMetadataIns) BuildTransaction(
	minerPrivateKey *privacy.SpendingKey,
	db database.DatabaseInterface,
) (metadata.Transaction, error) {
	rewardShareOldBoardMeta := metadata.NewShareRewardOldBoardMetadata(
		txShareRewardOldBoardMetadataIns.ChairPaymentAddress,
		txShareRewardOldBoardMetadataIns.VoterPaymentAddress,
		txShareRewardOldBoardMetadataIns.BoardType,
	)
	var propertyID common.Hash
	if txShareRewardOldBoardMetadataIns.BoardType == common.DCBBoard {
		propertyID = common.DCBTokenID
	} else {
		propertyID = common.GOVTokenID
	}
	txTokenData := transaction.TxTokenData{
		Type:       transaction.CustomTokenInit,
		Amount:     txShareRewardOldBoardMetadataIns.AmountOfToken,
		PropertyID: propertyID,
		Vins:       []transaction.TxTokenVin{},
		Vouts:      []transaction.TxTokenVout{},
	}

	txCustomToken, err := NewVoteCustomTokenTx(
		txShareRewardOldBoardMetadataIns.AmountOfCoin,
		&txShareRewardOldBoardMetadataIns.VoterPaymentAddress,
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
