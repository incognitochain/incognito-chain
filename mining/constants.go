package mining

import (
	"time"

	"github.com/ninjadotorg/cash-prototype/transaction"
)

const (
	DEFAULT_ADDRESS_FOR_BURNING      = "0x0000000000"
	NUMBER_OF_MAKING_DECISION_AGENTS = 3
	MAX_OF_MAKING_DECISION_AGENTS    = 21
	DEFAULT_COINS                    = 5
	DEFAULT_BONDS                    = 0
)

type TxDesc struct {
	// Tx is the transaction associated with the entry.
	Tx transaction.Transaction

	// Added is the time when the entry was added to the source pool.
	Added time.Time

	// Height is the block height when the entry was added to the the source pool.
	Height int32

	// Fee is the total fee the transaction associated with the entry pays.
	Fee int64

	//@todo add more properties to TxDesc if we need more laster
}
