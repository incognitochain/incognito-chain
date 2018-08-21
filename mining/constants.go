package mining

import (
	"time"

	"github.com/ninjadotorg/money-prototype/transaction"
	"github.com/ninjadotorg/money-prototype/blockchain"
)

const (

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

	Fees []int64
}


type BlkTmplGenerator struct {
	txSource    []*TxDesc
	chain       *blockchain.BlockChain
}