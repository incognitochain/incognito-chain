package metadata

import (
	"strconv"
	"time"

	"github.com/ninjadotorg/constant/blockchain/params"
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
	"github.com/ninjadotorg/constant/voting"
)

type MetadataBase struct {
	Type int
}

func NewMetadataBase(thisType int) *MetadataBase {
	return &MetadataBase{Type: thisType}
}

func (mb *MetadataBase) Validate() error {
	return nil
}

func (mb *MetadataBase) Process() error {
	return nil
}

func (mb *MetadataBase) GetType() int {
	return mb.Type
}

func (mb *MetadataBase) Hash() *common.Hash {
	record := strconv.Itoa(mb.Type)
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (mb *MetadataBase) CheckTransactionFee(tr Transaction, minFeePerKbTx uint64) bool {
	txFee := tr.GetTxFee()
	fullFee := minFeePerKbTx * tr.GetTxActualSize()
	if txFee < fullFee {
		return false
	}
	return true
}

// TODO(@0xankylosaurus): move TxDesc to mempool DTO
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
	GetSerialNumbers() map[common.Hash][][]byte
	GetTxsInMem() map[common.Hash]TxDesc
}

type BlockchainRetriever interface {
	GetHeight() int32
	GetCustomTokenTxs(*common.Hash) (map[common.Hash]Transaction, error)
	GetDCBParams() params.DCBParams
	GetDCBBoardPubKeys() [][]byte
	GetGOVParams() params.GOVParams
	GetGOVBoardPubKeys() [][]byte
	GetTransactionByHash(*common.Hash) (byte, *common.Hash, int, Transaction, error)
	GetOracleParams() *params.Oracle

	// For validating loan metadata
	GetLoanTxs([]byte) ([][]byte, error)
	GetNumberOfDCBGovernors() int
	GetNumberOfGOVGovernors() int
	GetLoanPayment([]byte) (uint64, uint64, uint32, error)
	GetLoanRequestMeta([]byte) (*LoanRequest, error)

	// For validating dividend
	GetAmountPerAccount(*DividendProposal) (uint64, []string, []uint64, error)

	// For validating crowdsale
	GetCrowdsaleData([]byte) (*voting.SaleData, error)
	GetCrowdsaleTxs([]byte) ([][]byte, error)
}

type Metadata interface {
	GetType() int
	Hash() *common.Hash
	CheckTransactionFee(Transaction, uint64) bool
	ValidateTxWithBlockChain(tx Transaction, bcr BlockchainRetriever, b byte, db database.DatabaseInterface) (bool, error)
	// isContinue, ok, err
	ValidateSanityData(bcr BlockchainRetriever, tx Transaction) (bool, bool, error)
	ValidateMetadataByItself() bool // TODO: need to define the method for metadata
}

// Interface for all type of transaction
type Transaction interface {
	Hash() *common.Hash
	ValidateTransaction(bool, database.DatabaseInterface, byte, *common.Hash) bool
	GetMetadataType() int
	GetType() string
	GetLockTime() int64
	GetTxActualSize() uint64
	GetSenderAddrLastByte() byte
	GetTxFee() uint64
	ListNullifiers() [][]byte
	CheckTxVersion(int8) bool
	CheckTransactionFee(minFeePerKbTx uint64) bool
	IsSalaryTx() bool
	ValidateTxWithCurrentMempool(MempoolRetriever) error
	ValidateTxWithBlockChain(BlockchainRetriever, byte, database.DatabaseInterface) error
	ValidateSanityData(BlockchainRetriever) (bool, error)
	ValidateTxByItself(bool, database.DatabaseInterface, BlockchainRetriever, byte) bool
	ValidateType() bool
	GetMetadata() Metadata
	SetMetadata(Metadata)
	ValidateConstDoubleSpendWithBlockchain(BlockchainRetriever, byte, database.DatabaseInterface) error

	GetJSPubKey() []byte
	GetReceivers() ([][]byte, []uint64)
	IsPrivacy() bool
	IsCoinsBurning() bool
	CloneTxThenUpdateMetadata(Metadata) []byte
}
