package blockchain

/*
Use these function to validate common data in blockchain
 */

import (
	"github.com/ninjadotorg/cash/common"
	"github.com/ninjadotorg/cash/transaction"
)

/*
IsSalaryTx determines whether or not a transaction is a salary.
*/
func IsSalaryTx(tx transaction.Transaction) bool {
	// Check normal tx(not an action tx)
	if tx.GetType() != common.TxNormalType {
		return true
	}
	normalTx, ok := tx.(*transaction.Tx)
	if !ok {
		return false
	}
	// Check nullifiers in every Descs
	descs := normalTx.Descs
	if len(descs) != 1 {
		return false
	} else {
		if descs[0].Reward > 0 {
			return true
		}
	}
	return false
}
