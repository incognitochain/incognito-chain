package transaction

import (
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
	"github.com/ninjadotorg/constant/metadata"
	"github.com/ninjadotorg/constant/privacy"
)

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
		// TODO(@0xbunyip): check if txtype should be set to txnormal instead of txsalary
		tx := new(Tx)
		err := tx.InitTxSalary(amounts[i], paymentAddress, producerPrivateKey, db, meta)
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
