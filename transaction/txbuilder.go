package transaction

import (
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
	infos []metadata.DividendPayment,
	proposal *metadata.DividendProposal,
	producerPrivateKey *privacy.SpendingKey,
	db database.DatabaseInterface,
) ([]*Tx, error) {
	amounts := []uint64{}
	dividendMetaList := []metadata.Metadata{}
	paymentAddresses := []*privacy.PaymentAddress{}
	for _, info := range infos {
		amounts = append(amounts, info.Amount)
		paymentAddress := info.TokenHolder
		paymentAddresses = append(paymentAddresses, &paymentAddress)
		dividendMeta := &metadata.Dividend{
			PayoutID:       proposal.PayoutID,
			TokenID:        proposal.TokenID,
			PaymentAddress: paymentAddress,
			MetadataBase:   metadata.MetadataBase{Type: metadata.DividendMeta},
		}
		dividendMetaList = append(dividendMetaList, dividendMeta)
	}
	return BuildCoinbaseTxs(paymentAddresses, amounts, producerPrivateKey, db, dividendMetaList)
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
