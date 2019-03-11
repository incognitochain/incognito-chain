package transaction

import (
	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/database"
	"github.com/constant-money/constant-chain/metadata"
	"github.com/constant-money/constant-chain/privacy"
)

func BuildCoinbaseTx(
	paymentAddress *privacy.PaymentAddress,
	amount uint64,
	producerPrivateKey *privacy.SpendingKey,
	db database.DatabaseInterface,
	meta metadata.Metadata,
) (*Tx, error) {
	tx := &Tx{}
	// TODO(@0xbunyip): use another method that sets type to TxNormal (otherwise tx signature will be violated)
	err := tx.InitTxSalary(amount, paymentAddress, producerPrivateKey, db, meta)
	tx.Type = common.TxNormalType
	if err != nil {
		return nil, err
	}
	return tx, nil
}

func BuildCoinbaseTxs(
	paymentAddresses []*privacy.PaymentAddress,
	amounts []uint64,
	producerPrivateKey *privacy.SpendingKey,
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

func BuildDividendTxs(
	dividendID uint64,
	tokenID *common.Hash,
	receivers []*privacy.PaymentAddress,
	amounts []uint64,
	producerPrivateKey *privacy.SpendingKey,
	db database.DatabaseInterface,
) ([]*Tx, error) {
	metas := []metadata.Metadata{}
	for i := 0; i < len(receivers); i++ {
		dividendMeta := &metadata.DividendPayment{
			DividendID:   dividendID,
			TokenID:      tokenID,
			MetadataBase: metadata.MetadataBase{Type: metadata.DividendPaymentMeta},
		}
		metas = append(metas, dividendMeta)
	}
	return BuildCoinbaseTxs(receivers, amounts, producerPrivateKey, db, metas)
}

// BuildRefundTx - build a coinbase tx to refund constant with CMB policies
func BuildRefundTx(
	receiver privacy.PaymentAddress,
	amount uint64,
	producerPrivateKey *privacy.SpendingKey,
	db database.DatabaseInterface,
) (*Tx, error) {
	meta := &metadata.CMBInitRefund{
		MainAccount:  receiver,
		MetadataBase: metadata.MetadataBase{Type: metadata.CMBInitRefundMeta},
	}
	metaList := []metadata.Metadata{meta}
	amounts := []uint64{amount}
	txs, err := BuildCoinbaseTxs([]*privacy.PaymentAddress{&receiver}, amounts, producerPrivateKey, db, metaList)
	if err != nil {
		return nil, err
	}
	return txs[0], nil // only one tx in slice
}
