package mining

import (
	"time"

	"github.com/ninjadotorg/cash-prototype/transaction"
	"github.com/ninjadotorg/cash-prototype/blockchain"
)

const (
	ACTION_PARAMS_TRANSACTION_TYPE = "ACTION_PARAMS"
	NUMBER_OF_LAST_BLOCKS = 10
	NUMBER_OF_MAKING_DECISION_AGENTS = 3
	DEFAULT_COINS = 5
	DEFAULT_BONDS = 5
)

type Policy struct {
	//@todo well will declare it later
}

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

type BlockTemplate struct {
	Block *blockchain.Block

	Fees []float64
}


type BlkTmplGenerator struct {
	txSource    []*TxDesc
	chain       *blockchain.BlockChain
}