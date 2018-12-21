package blockchain

import (
	"github.com/ninjadotorg/constant/database"
	"github.com/ninjadotorg/constant/metadata"
	privacy "github.com/ninjadotorg/constant/privacy-protocol"
	"github.com/ninjadotorg/constant/transaction"
	"github.com/pkg/errors"
)

func buildRefundTx(
	receiver privacy.PaymentAddress,
	amount uint64,
	producerPrivateKey *privacy.SpendingKey,
	db database.DatabaseInterface,
) (*transaction.Tx, error) {
	pks := [][]byte{receiver.Pk[:]}
	tks := [][]byte{receiver.Tk[:]}
	amounts := []uint64{amount}
	txs, err := buildCoinbaseTxs(pks, tks, amounts, producerPrivateKey, db)
	if err != nil {
		return nil, err
	}
	return txs[0], nil // only one tx in slice
}

func (blockgen *BlkTmplGenerator) buildCMBRefund(sourceTxns []*metadata.TxDesc, chainID byte, producerPrivateKey *privacy.SpendingKey) ([]*transaction.Tx, error) {
	// Get old block
	refunds := []*transaction.Tx{}
	header := blockgen.chain.BestState[chainID].BestBlock.Header
	lookbackBlockHeight := header.Height - metadata.CMBInitRefundPeriod
	if lookbackBlockHeight < 0 {
		return refunds, nil
	}
	lookbackBlock, err := blockgen.chain.GetBlockByBlockHeight(lookbackBlockHeight, chainID)
	if err != nil {
		Logger.log.Error(err)
		return refunds, nil
	}

	// Build refund tx for each cmb init request that hasn't been approved
	for _, tx := range lookbackBlock.Transactions {
		if tx.GetMetadataType() != metadata.CMBInitRequestMeta {
			continue
		}

		// Check if CMB is still not approved in previous blocks
		meta := tx.GetMetadata().(*metadata.CMBInitRequest)
		_, capital, _, state, err := blockgen.chain.GetCMB(meta.MainAccount.ToBytes())
		if err != nil {
			// Unexpected error, cannot create a block if CMB init request is not refundable
			return nil, errors.Errorf("error retrieving cmb for building refund")
		}
		if state == metadata.CMBRequested {
			refund, err := buildRefundTx(meta.MainAccount, capital, producerPrivateKey, blockgen.chain.GetDatabase())
			if err != nil {
				return nil, errors.Errorf("error building refund tx for cmb init request")
			}
			refunds = append(refunds, refund)
		}
	}
	return refunds, nil
}
