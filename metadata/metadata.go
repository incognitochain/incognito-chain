package metadata

import (
	"encoding/json"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/database"
	"github.com/incognitochain/incognito-chain/ethrelaying/les"
	zkp "github.com/incognitochain/incognito-chain/privacy/zeroknowledge"
)

type MetadataBase struct {
	Type int
}

func NewMetadataBase(thisType int) *MetadataBase {
	return &MetadataBase{Type: thisType}
}

func calculateSize(meta Metadata) uint64 {
	metaBytes, err := json.Marshal(meta)
	if err != nil {
		return 0
	}
	return uint64(len(metaBytes))
}

func (mb *MetadataBase) IsMinerCreatedMetaType() bool {
	metaType := mb.GetType()
	for _, mType := range minerCreatedMetaTypes {
		if metaType == mType {
			return true
		}
	}
	return false
}

func (mb *MetadataBase) CalculateSize() uint64 {
	return 0
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
	hash := common.HashH([]byte(record))
	return &hash
}

func (mb *MetadataBase) CheckTransactionFee(tx Transaction, minFeePerKbTx uint64) bool {
	txFee := tx.GetTxFee()
	fullFee := minFeePerKbTx * tx.GetTxActualSize()
	return !(txFee < fullFee)
}

func (mb *MetadataBase) VerifyMultiSigs(
	tx Transaction,
	db database.DatabaseInterface,
) (bool, error) {
	return true, nil
}

func (mb *MetadataBase) BuildReqActions(tx Transaction, bcr BlockchainRetriever, shardID byte) ([][]string, error) {
	return [][]string{}, nil
}

func (mb *MetadataBase) ProcessWhenInsertBlockShard(tx Transaction, retriever BlockchainRetriever) error {
	return nil
}

func (mb *MetadataBase) VerifyMinerCreatedTxBeforeGettingInBlock(
	txsInBlock []Transaction,
	txsUsed []int,
	insts [][]string,
	instsUsed []int,
	shardID byte,
	txr Transaction,
	bcr BlockchainRetriever,
	accumulatedValues *AccumulatedValues,
) (bool, error) {
	return true, nil
}

// This is tx struct which is really saved in tx mempool
type TxDesc struct {
	// Tx is the transaction associated with the entry.
	Tx Transaction

	// Height is the best block's height when the entry was added to the the source pool.
	Height uint64

	// Fee is the total fee the transaction associated with the entry pays.
	Fee uint64

	// FeeToken is the total token fee the transaction associated with the entry pays.
	// FeeToken is zero if tx is PRV transaction
	FeeToken uint64

	// FeePerKB is the fee the transaction pays in coin per 1000 bytes.
	FeePerKB int32
}

// Interface for mempool which is used in metadata
type MempoolRetriever interface {
	GetSerialNumbersHashH() map[common.Hash][]common.Hash
	GetTxsInMem() map[common.Hash]TxDesc
}

// Interface for blockchain which is used in metadata
type BlockchainRetriever interface {
	GetTxChainHeight(tx Transaction) (uint64, error)
	GetChainHeight(byte) uint64
	GetBeaconHeight() uint64
	GetCustomTokenTxs(*common.Hash) (map[common.Hash]Transaction, error)
	GetTransactionByHash(common.Hash) (byte, common.Hash, int, Transaction, error)
	GetCurrentBeaconBlockHeight(byte) uint64
	GetAllCommitteeValidatorCandidate() (map[byte][]string, map[byte][]string, []string, []string, []string, []string, []string, []string)
	GetDatabase() database.DatabaseInterface
	GetTxValue(txid string) (uint64, error)
	GetShardIDFromTx(txid string) (byte, error)
	GetLightEthereum() *les.LightEthereum
}

// Interface for all types of metadata in tx
type Metadata interface {
	GetType() int
	Hash() *common.Hash
	CheckTransactionFee(Transaction, uint64) bool
	ValidateTxWithBlockChain(tx Transaction, bcr BlockchainRetriever, b byte, db database.DatabaseInterface) (bool, error)
	// isContinue, ok, err
	ValidateSanityData(bcr BlockchainRetriever, tx Transaction) (bool, bool, error)
	ValidateMetadataByItself() bool // TODO: need to define the method for metadata
	VerifyMultiSigs(Transaction, database.DatabaseInterface) (bool, error)
	BuildReqActions(tx Transaction, bcr BlockchainRetriever, shardID byte) ([][]string, error)
	ProcessWhenInsertBlockShard(tx Transaction, bcr BlockchainRetriever) error
	CalculateSize() uint64
	VerifyMinerCreatedTxBeforeGettingInBlock([]Transaction, []int, [][]string, []int, byte, Transaction, BlockchainRetriever, *AccumulatedValues) (bool, error)
	IsMinerCreatedMetaType() bool
}

// Interface for all type of transaction
type Transaction interface {
	Hash() *common.Hash
	ValidateTransaction(bool, database.DatabaseInterface, byte, *common.Hash) (bool, error)
	GetMetadataType() int
	GetType() string
	GetLockTime() int64
	GetTxActualSize() uint64
	GetSenderAddrLastByte() byte
	GetTxFee() uint64
	GetTxFeeToken() uint64
	ListSerialNumbersHashH() []common.Hash
	CheckTxVersion(int8) bool
	CheckTransactionFee(minFeePerKbTx uint64) bool
	IsSalaryTx() bool
	ValidateTxWithCurrentMempool(MempoolRetriever) error
	ValidateTxWithBlockChain(BlockchainRetriever, byte, database.DatabaseInterface) error
	ValidateSanityData(BlockchainRetriever) (bool, error)
	ValidateTxByItself(bool, database.DatabaseInterface, BlockchainRetriever, byte) (bool, error)
	ValidateType() bool
	GetMetadata() Metadata
	SetMetadata(Metadata)
	GetInfo() []byte
	ValidateConstDoubleSpendWithBlockchain(BlockchainRetriever, byte, database.DatabaseInterface) error

	GetSender() []byte
	GetSigPubKey() []byte
	IsPrivacy() bool
	IsCoinsBurning() bool
	CalculateTxValue() uint64
	GetProof() *zkp.PaymentProof

	// Get receivers' data for tx
	GetReceivers() ([][]byte, []uint64)
	GetUniqueReceiver() (bool, []byte, uint64)
	GetTransferData() (bool, []byte, uint64, *common.Hash)

	// Get receivers' data for custom token tx (nil for normal tx)
	GetTokenReceivers() ([][]byte, []uint64)
	GetTokenUniqueReceiver() (bool, []byte, uint64)

	GetMetadataFromVinsTx(BlockchainRetriever) (Metadata, error)
	GetTokenID() *common.Hash
	VerifyMinerCreatedTxBeforeGettingInBlock([]Transaction, []int, [][]string, []int, byte, BlockchainRetriever, *AccumulatedValues) (bool, error)
}
