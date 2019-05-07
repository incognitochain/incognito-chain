package transaction

import (
	"github.com/constant-money/constant-chain/database"
	"github.com/constant-money/constant-chain/metadata"
	"github.com/constant-money/constant-chain/privacy"
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
