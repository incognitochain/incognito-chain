package frombeaconins

import (
	"encoding/json"
	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/database"
	"github.com/constant-money/constant-chain/metadata"
	"github.com/constant-money/constant-chain/privacy"
	"github.com/constant-money/constant-chain/transaction"
	"strconv"
)

type TxSendBackTokenVoteFailIns struct {
	PaymentAddress privacy.PaymentAddress
	Amount         uint64
	PropertyID     common.Hash
}

func (txSendBackTokenVoteFailIns *TxSendBackTokenVoteFailIns) GetStringFormat() ([]string, error) {
	content, err := json.Marshal(txSendBackTokenVoteFailIns)
	if err != nil {
		return nil, err
	}
	shardID := GetShardIDFromPaymentAddressBytes(txSendBackTokenVoteFailIns.PaymentAddress)
	return []string{
		strconv.Itoa(metadata.SendBackTokenVoteFailMeta),
		strconv.Itoa(int(shardID)),
		string(content),
	}, nil
}

func GetShardIDFromPaymentAddressBytes(paymentAddress privacy.PaymentAddress) byte {
	lastByte := paymentAddress.Pk[len(paymentAddress.Pk)-1]
	return common.GetShardIDFromLastByte(lastByte)
}

func (txSendBackTokenVoteFailIns *TxSendBackTokenVoteFailIns) BuildTransaction(
	minerPrivateKey *privacy.SpendingKey,
	db database.DatabaseInterface,
) (metadata.Transaction, error) {
	tx := NewSendBackTokenVoteFailTx(
		minerPrivateKey,
		db,
		txSendBackTokenVoteFailIns.PaymentAddress,
		txSendBackTokenVoteFailIns.Amount,
		txSendBackTokenVoteFailIns.PropertyID,
	)
	return tx, nil
}

func NewSendBackTokenVoteFailIns(
	paymentAddress privacy.PaymentAddress,
	amount uint64,
	propertyID common.Hash,
) *TxSendBackTokenVoteFailIns {
	return &TxSendBackTokenVoteFailIns{
		PaymentAddress: paymentAddress,
		Amount:         amount,
		PropertyID:     propertyID,
	}
}

func NewSendBackTokenVoteFailTx(
	minerPrivateKey *privacy.SpendingKey,
	db database.DatabaseInterface,
	paymentAddress privacy.PaymentAddress,
	amount uint64,
	propertyID common.Hash,
) metadata.Transaction {
	txTokenVout := transaction.TxTokenVout{
		Value:          amount,
		PaymentAddress: paymentAddress,
	}
	newTx := transaction.TxCustomToken{
		TxTokenData: transaction.TxTokenData{
			Type:       transaction.CustomTokenInit,
			Amount:     amount,
			PropertyID: propertyID,
			Vins:       []transaction.TxTokenVin{},
			Vouts:      []transaction.TxTokenVout{txTokenVout},
		},
	}
	newTx.Type = common.TxCustomTokenType
	newTx.SetMetadata(metadata.NewSendBackTokenVoteFailMetadata())
	return &newTx
}
