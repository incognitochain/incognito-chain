package metadata

import (
	"time"

	"github.com/ninjadotorg/constant/common"
)

// This is tx struct which is really saved in tx mempool
type TxDesc struct {
	// Tx is the transaction associated with the entry.
	Tx Transaction

	// Added is the time when the entry was added to the source pool.
	Added time.Time

	// Height is the best block's height when the entry was added to the the source pool.
	Height int32

	// Fee is the total fee the transaction associated with the entry pays.
	Fee uint64

	// FeePerKB is the fee the transaction pays in coin per 1000 bytes.
	FeePerKB int32
}

type MempoolRetriever interface {
	GetPoolNullifiers() map[common.Hash][][]byte
	GetTxsInMem() map[common.Hash]TxDesc
}

type BlockchainRetriever interface {
	GetNulltifiersList(byte) ([][]byte, error)
	GetCustomTokenTxs(*common.Hash) (map[common.Hash]Transaction, error)
}

type TxRetriever interface {
	GetTxFee() uint64
}

type Metadata interface {
	Validate() error
	Process() error
	CheckTransactionFee(TxRetriever, uint64) bool
	ValidateTxWithBlockChain(BlockchainRetriever, byte) (bool, error)
}

// Interface for all type of transaction
type Transaction interface {
	Hash() *common.Hash
	ValidateTransaction() bool
	GetType() string
	GetTxVirtualSize() uint64
	GetSenderAddrLastByte() byte
	GetTxFee() uint64
	ListNullifiers() [][]byte
	CheckTxVersion(int8) bool
	CheckTransactionFee(uint64) bool
	IsSalaryTx() bool
	ValidateTxWithCurrentMempool(MempoolRetriever) error
	ValidateTxWithBlockChain(BlockchainRetriever, byte) error
}
