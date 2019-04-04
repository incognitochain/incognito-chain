package frombeaconins

import (
	"encoding/json"
	"strconv"

	"github.com/constant-money/constant-chain/transaction"

	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/database"
	"github.com/constant-money/constant-chain/metadata"
	"github.com/constant-money/constant-chain/privacy"
)

type TxSendBackTokenToOldSupporterIns struct {
	BoardType      common.BoardType
	PaymentAddress privacy.PaymentAddress
	Amount         uint64
	PropertyID     common.Hash
}

func NewTxSendBackTokenToOldSupporterIns(boardType common.BoardType, paymentAddress privacy.PaymentAddress, amount uint64, propertyID common.Hash) *TxSendBackTokenToOldSupporterIns {
	return &TxSendBackTokenToOldSupporterIns{BoardType: boardType, PaymentAddress: paymentAddress, Amount: amount, PropertyID: propertyID}
}

func (txSendBackTokenToOldSupporterIns *TxSendBackTokenToOldSupporterIns) GetStringFormat() ([]string, error) {
	content, err := json.Marshal(txSendBackTokenToOldSupporterIns)
	if err != nil {
		return nil, err
	}
	shardID := GetShardIDFromPaymentAddressBytes(txSendBackTokenToOldSupporterIns.PaymentAddress)
	return []string{
		strconv.Itoa(metadata.SendBackTokenToOldSupporterMeta),
		strconv.Itoa(int(shardID)),
		string(content),
	}, nil
}

func (txSendBackTokenToOldSupporterIns *TxSendBackTokenToOldSupporterIns) BuildTransaction(
	minerPrivateKey *privacy.PrivateKey,
	db database.DatabaseInterface,
	bcr metadata.BlockchainRetriever,
	shardID byte,
) (metadata.Transaction, error) {
	return NewSendBackTokenToOldSupporterTx(
		txSendBackTokenToOldSupporterIns.BoardType,
		minerPrivateKey,
		db,
		txSendBackTokenToOldSupporterIns.PaymentAddress,
		txSendBackTokenToOldSupporterIns.Amount,
		bcr,
		shardID,
	)
}

func NewSendBackTokenToOldSupporterIns(
	paymentAddress privacy.PaymentAddress,
	amount uint64,
	propertyID common.Hash,
) *TxSendBackTokenToOldSupporterIns {
	return &TxSendBackTokenToOldSupporterIns{
		PaymentAddress: paymentAddress,
		Amount:         amount,
		PropertyID:     propertyID,
	}
}

func NewSendBackTokenToOldSupporterTx(
	boardType common.BoardType,
	minerPrivateKey *privacy.PrivateKey,
	db database.DatabaseInterface,
	paymentAddress privacy.PaymentAddress,
	amount uint64,
	bcr metadata.BlockchainRetriever,
	shardID byte,
) (metadata.Transaction, error) {

	//create token params
	customTokenParamTx := mintDCBTokenParam
	customTokenParamTx.Receiver = []transaction.TxTokenVout{{
		Value:          amount,
		PaymentAddress: paymentAddress,
	}}
	if boardType == common.GOVBoard {
		customTokenParamTx = mintGOVTokenParam
	}
	customTokenParamTx.Amount = amount

	//CALL DB
	//listCustomTokens, err := GetListCustomTokens(db, bcr)
	//if err != nil {
	//	return nil, err
	//}
	txCustom := &transaction.TxCustomToken{}
	err1 := txCustom.Init(
		minerPrivateKey,
		[]*privacy.PaymentInfo{},
		nil,
		0,
		&customTokenParamTx,
		//listCustomTokens,
		db,
		metadata.NewSendBackTokenToOldSupporterMetadata(),
		false,
		shardID,
	)
	if err1 != nil {
		return nil, err1
	}
	txCustom.Type = common.TxCustomTokenType
	return txCustom, nil
}
