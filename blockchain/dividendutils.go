package blockchain

// import (
// 	"github.com/ninjadotorg/constant/database"
// 	"github.com/ninjadotorg/constant/metadata"
// 	"github.com/ninjadotorg/constant/privacy"
// 	"github.com/ninjadotorg/constant/transaction"
// )

func buildCoinbaseTxs(
	pks, tks [][]byte,
	amounts []uint64,
	producerPrivateKey *privacy.SpendingKey,
	db database.DatabaseInterface,
	metaList []metadata.Metadata,
) ([]*transaction.Tx, error) {
	txs := []*transaction.Tx{}
	for i := 0; i < len(pks); i++ {
		var meta metadata.Metadata
		if metaList == nil || len(metaList) == 0 {
			meta = nil
		} else {
			meta = metaList[i]
		}
		paymentAddress := &privacy.PaymentAddress{
			Pk: pks[i],
			Tk: tks[i],
		}
		// TODO(@0xbunyip): check if txtype should be set to txnormal instead of txsalary
		tx := new(transaction.Tx)
		err := tx.InitTxSalary(amounts[i], paymentAddress, producerPrivateKey, db, meta)
		if err != nil {
			return nil, err
		}
		txs = append(txs, tx)
	}
	return txs, nil
}

// func buildDividendTxs(
// 	infos []metadata.DividendInfo,
// 	proposal *metadata.DividendProposal,
// 	producerPrivateKey *privacy.SpendingKey,
// 	db database.DatabaseInterface,
// ) ([]*transaction.Tx, error) {
// 	pks := [][]byte{}
// 	tks := [][]byte{}
// 	amounts := []uint64{}
// 	for _, info := range infos {
// 		pks = append(pks, info.TokenHolder.Pk)
// 		tks = append(tks, info.TokenHolder.Tk)
// 		amounts = append(amounts, info.Amount)
// 	}

	dividendMetaList := []metadata.Metadata{}
	for i := 0; i < len(pks); i++ {
		paymentAddress := privacy.PaymentAddress{
			Pk: pks[i],
			Tk: tks[i],
		}
		dividendMeta := &metadata.Dividend{
			PayoutID:       proposal.PayoutID,
			TokenID:        proposal.TokenID,
			PaymentAddress: paymentAddress,
			MetadataBase:   metadata.MetadataBase{Type: metadata.DividendMeta},
		}
		dividendMetaList = append(dividendMetaList, dividendMeta)
	}
	return buildCoinbaseTxs(pks, tks, amounts, producerPrivateKey, db, dividendMetaList)
}
