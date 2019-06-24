package transaction

import (
	"errors"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/database"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/privacy"
)

func BuildCoinbaseTx(
	paymentAddress *privacy.PaymentAddress,
	amount uint64,
	producerPrivateKey *privacy.PrivateKey,
	db database.DatabaseInterface,
	meta metadata.Metadata,
) (*Tx, error) {
	tx := &Tx{}
	err := tx.InitTxSalary(amount, paymentAddress, producerPrivateKey, db, meta)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

func BuildCoinbaseTxByCoinID(
	payToAddress *privacy.PaymentAddress,
	amount uint64,
	payByPrivateKey *privacy.PrivateKey,
	db database.DatabaseInterface,
	meta metadata.Metadata,
	coinID common.Hash,
	txType int,
	coinName string,
	shardID byte,
) (metadata.Transaction, error) {
	switch txType {
	case NormalCoinType:
		tx := &Tx{}
		err := tx.InitTxSalary(amount, payToAddress, payByPrivateKey, db, meta)
		return tx, err
	case CustomTokenType:
		tx := &TxCustomToken{}
		receiver := &TxTokenVout{
			PaymentAddress: *payToAddress,
			Value:          amount,
		}
		tokenParams := &CustomTokenParamTx{
			PropertyID:     coinID.String(),
			PropertyName:   coinName,
			PropertySymbol: coinName,
			Amount:         amount,
			TokenTxType:    CustomTokenInit,
			Receiver:       []TxTokenVout{*receiver},
			Mintable:       true,
		}
		err := tx.Init(
			payByPrivateKey,
			nil,
			nil,
			0,
			tokenParams,
			//listCustomTokens,
			db,
			meta,
			false,
			shardID,
		)
		if err != nil {
			return nil, errors.New(err.Error())
		}
		return tx, nil
	case CustomTokenPrivacyType:
		var propertyID [common.HashSize]byte
		copy(propertyID[:], coinID[:])
		receiver := &privacy.PaymentInfo{
			Amount:         amount,
			PaymentAddress: *payToAddress,
		}
		propID := common.Hash(propertyID)
		tokenParams := &CustomTokenPrivacyParamTx{
			PropertyID:     propID.String(),
			PropertyName:   coinName,
			PropertySymbol: coinName,
			Amount:         amount,
			TokenTxType:    CustomTokenInit,
			Receiver:       []*privacy.PaymentInfo{receiver},
			TokenInput:     []*privacy.InputCoin{},
			Mintable:       true,
		}
		tx := &TxCustomTokenPrivacy{}
		err := tx.Init(
			payByPrivateKey,
			[]*privacy.PaymentInfo{},
			nil,
			0,
			tokenParams,
			db,
			meta,
			false,
			false,
			shardID,
		)
		if err != nil {
			return nil, errors.New(err.Error())
		}
		return tx, nil
	}
	return nil, nil
}

func BuildCoinbaseTxs(
	paymentAddresses []*privacy.PaymentAddress,
	amounts []uint64,
	producerPrivateKey *privacy.PrivateKey,
	db database.DatabaseInterface,
	metaList []metadata.Metadata,
) ([]*Tx, error) {
	txs := []*Tx{}
	for i, paymentAddress := range paymentAddresses {
		var meta metadata.Metadata
		if len(metaList) == 0 {
			meta = nil
		} else {
			meta = metaList[i]
		}
		tx, err := BuildCoinbaseTx(paymentAddress, amounts[i], producerPrivateKey, db, meta)
		if err != nil {
			return nil, err
		}
		txs = append(txs, tx)
	}
	return txs, nil
}
