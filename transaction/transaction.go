package transaction

import (
	"time"

	"github.com/ninjadotorg/cash-prototype/common"
)

type Transaction interface {
	Hash() *common.Hash
	ValidateTransaction() bool
	GetType() string
}

type TxDesc struct {
	// Tx is the transaction associated with the entry.
	Tx Transaction

	// Added is the time when the entry was added to the source pool.
	Added time.Time

	// Height is the block height when the entry was added to the the source pool.
	Height int32

	// Fee is the total fee the transaction associated with the entry pays.
	Fee float64

	//@todo add more properties to TxDesc if we need more laster
}
