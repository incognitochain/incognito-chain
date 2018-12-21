package blockchain

import (
	"github.com/ninjadotorg/constant/database"
	"github.com/ninjadotorg/constant/metadata"
	"github.com/ninjadotorg/constant/privacy-protocol"
	"github.com/ninjadotorg/constant/transaction"
)

func buildCoinbaseTxs(
	pks, tks [][]byte,
	amounts []uint64,
	producerPrivateKey *privacy.SpendingKey,
	db database.DatabaseInterface,
) ([]*transaction.Tx, error) {
	txs := []*transaction.Tx{}
	for i := 0; i < len(pks); i++ {
		paymentAddress := &privacy.PaymentAddress{
			Pk: pks[i],
			Tk: tks[i],
		}
		// TODO(@0xbunyip): check if txtype should be set to txnormal instead of txsalary
		tx, err := transaction.CreateTxSalary(amounts[i], paymentAddress, producerPrivateKey, db)
		if err != nil {
			return nil, err
		}
		txs = append(txs, tx)
	}
	return txs, nil
}

func buildDividendTxs(
	infos []metadata.DividendInfo,
	proposal *metadata.DividendProposal,
	producerPrivateKey *privacy.SpendingKey,
	db database.DatabaseInterface,
) ([]*transaction.Tx, error) {
	pks := [][]byte{}
	tks := [][]byte{}
	amounts := []uint64{}
	for _, info := range infos {
		pks = append(pks, info.TokenHolder.Pk)
		tks = append(tks, info.TokenHolder.Tk)
		amounts = append(amounts, info.Amount)
	}

	txs, err := buildCoinbaseTxs(pks, tks, amounts, producerPrivateKey, db)
	if err != nil {
		return nil, err
	}
	for index, tx := range txs {
		paymentAddress := privacy.PaymentAddress{
			Pk: pks[index][:],
			Tk: tks[index][:],
		}
		tx.Metadata = &metadata.Dividend{
			PayoutID:       proposal.PayoutID,
			TokenID:        proposal.TokenID,
			PaymentAddress: paymentAddress,
		}
	}
	return txs, nil
}
