package blockchain

import (
	// "math"

	"github.com/ninjadotorg/cash-prototype/transaction"
	"github.com/ninjadotorg/cash-prototype/common"
)

var (
	zeroHash common.Hash
)

/*func CountSigOps(tx *transaction.Tx) float64 {
	totalSigOps := 0.0
	*//*for _, txIn := range tx.TxIn {
		//@todo need implement function calc value of input
		fmt.Print(txIn.PreviousOutPoint)
	}

	for _, txOut := range tx.TxOut {

		totalSigOps -= txOut.Value
	}*//*

	return totalSigOps
}*/

/**
IsCoinBaseTx determines whether or not a transaction is a coinbase.
*/
func IsCoinBaseTx(tx transaction.Transaction) bool {
	// Check normal tx(not an action tx)
	normalTx, ok := tx.(*transaction.Tx)
	if !ok {
		return false
	}
	// Check nullifiers in every Descs
	descs := normalTx.Desc
	for _, desc := range descs {
		if len(desc.Nullifiers) > 0 {
			return false
		}
	}
	return true
}
