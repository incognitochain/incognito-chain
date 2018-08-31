package blockchain

import (
	"fmt"
	// "math"

	"github.com/ninjadotorg/cash-prototype/transaction"
	"github.com/ninjadotorg/cash-prototype/common"
	"math"
)

var (
	zeroHash common.Hash
)

func nonNilBytes(bz []byte) []byte {
	if bz == nil {
		return []byte{}
	}
	return bz
}

func CountSigOps(tx *transaction.Tx) float64 {

	totalSigOps := 0.0
	for _, txIn := range tx.TxIn {
		//@todo need implement function calc value of input
		fmt.Print(txIn.PreviousOutPoint)
	}

	for _, txOut := range tx.TxOut {

		totalSigOps -= txOut.Value
	}

	return totalSigOps
}

// IsCoinBaseTx determines whether or not a transaction is a coinbase.  A coinbase
// is a special transaction created by miners that has no inputs.  This is
// represented in the block chain by a transaction with a single input that has
// a previous output transaction index set to the maximum value along with a
// zero hash.
//
// This function only differs from IsCoinBase in that it works with a raw wire
// transaction as opposed to a higher level util transaction.
func IsCoinBaseTx(tx transaction.Tx) bool {
	// A coin base must only have one transaction input.
	if len(tx.TxIn) != 1 {
		return false
	}

	// The previous output of a coin base must have a max value index and
	// a zero hash.
	prevOut := &tx.TxIn[0].PreviousOutPoint
	if prevOut.Vout != math.MaxUint32 || prevOut.Hash != zeroHash {
		return false
	}

	return true
}
