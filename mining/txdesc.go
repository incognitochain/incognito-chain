package mining

import (
	"time"
	"github.com/ninjadotorg/cash-prototype/transaction"
)

/**
 TxDesc is the object which is used to saved into mempool for mining processing
 */
type TxDesc struct {
	// Tx is the transaction associated with the entry.
	Tx transaction.Transaction

	// Added is the time when the entry was added to the source pool.
	Added time.Time

	// Height is the best block's height when the entry was added to the the source pool.
	Height int32

	// Fee is the total fee the transaction associated with the entry pays.
	Fee uint64

	//@todo add more properties to TxDesc if we need more laster
}
