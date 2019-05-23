package database

import (
	"math/big"

	"github.com/constant-money/constant-chain/common"
)

// DatabaseInterface provides the interface that is used to store blocks.
type DatabaseInterface interface {
	Put(key, value []byte) error
	Get(key []byte) ([]byte, error)
	Delete(key []byte) error
	HasValue(key []byte) (bool, error)

	// Block
	StoreShardBlock(interface{}, byte) error
	StoreShardBlockHeader(interface{}, *common.Hash, byte) error
	FetchBlock(*common.Hash) ([]byte, error)
	HasBlock(*common.Hash) (bool, error)
	DeleteBlock(*common.Hash, uint64, byte) error

	StoreIncomingCrossShard(shardID byte, crossShardID byte, blkHeight uint64, crossBlkHash *common.Hash) error
	HasIncomingCrossShard(shardID byte, crossShardID byte, crossBlkHash *common.Hash) error
	GetIncomingCrossShard(shardID byte, crossShardID byte, crossBlkHash *common.Hash) (uint64, error)
	// StoreOutgoingCrossShard(shardID byte, crossShardID byte, blkHeight uint64, crossBlk interface{}) error
	// HasOutgoingCrossShard(shardID byte, crossShardID byte, blkHeight uint64) error
	// GetOutgoingCrossShard(shardID byte, crossShardID byte, blkHeight uint64) ([]byte, error)

	StoreAcceptedShardToBeacon(shardID byte, blkHeight uint64, shardBlkHash *common.Hash) error
	HasAcceptedShardToBeacon(shardID byte, shardBlkHash *common.Hash) error
	GetAcceptedShardToBeacon(shardID byte, shardBlkHash *common.Hash) (uint64, error)

	// Beacon
	StoreBeaconBlock(interface{}) error
	StoreBeaconBlockHeader(interface{}, *common.Hash) error
	FetchBeaconBlock(*common.Hash) ([]byte, error)
	HasBeaconBlock(*common.Hash) (bool, error)
	FetchBeaconBlockChain() ([]*common.Hash, error)
	DeleteBeaconBlock(*common.Hash, uint64) error

	//Crossshard
	StoreCrossShardNextHeight(byte, byte, uint64, uint64) error
	FetchCrossShardNextHeight(byte, byte, uint64) (uint64, error)

	// Block index
	StoreShardBlockIndex(*common.Hash, uint64, byte) error
	GetIndexOfBlock(*common.Hash) (uint64, byte, error)
	GetBlockByIndex(uint64, byte) (*common.Hash, error)

	// Block index
	StoreBeaconBlockIndex(*common.Hash, uint64) error
	GetIndexOfBeaconBlock(*common.Hash) (uint64, error)
	GetBeaconBlockHashByIndex(uint64) (*common.Hash, error)

	// Transaction index
	StoreTransactionIndex(txId *common.Hash, blockHash *common.Hash, indexInBlock int) error
	GetTransactionIndexById(txId *common.Hash) (*common.Hash, int, *DatabaseError)

	// Best state of chain
	StorePrevBestState(interface{}, bool, byte) error
	FetchPrevBestState(bool, byte) ([]byte, error)

	// Best state of chain
	StoreShardBestState(interface{}, byte) error
	FetchShardBestState(byte) ([]byte, error)
	CleanShardBestState() error

	// Best state of chain
	StoreBeaconBestState(interface{}) error
	StoreCommitteeByHeight(uint64, interface{}) error
	StoreCommitteeByEpoch(uint64, interface{}) error
	StoreStabilityInfoByHeight(uint64, interface{}) error

	FetchCommitteeByHeight(uint64) ([]byte, error)
	FetchStabilityInfoByHeight(uint64) ([]byte, error)
	FetchCommitteeByEpoch(uint64) ([]byte, error)
	HasCommitteeByEpoch(uint64) (bool, error)
	FetchBeaconBestState() ([]byte, error)
	CleanBeaconBestState() error

	// SerialNumber
	StoreSerialNumbers(tokenID *common.Hash, data []byte, shardID byte) error
	FetchSerialNumbers(tokenID *common.Hash, shardID byte) ([][]byte, error)
	HasSerialNumber(tokenID *common.Hash, data []byte, shardID byte) (bool, error)
	CleanSerialNumbers() error

	// PedersenCommitment
	StoreCommitments(tokenID *common.Hash, pubkey []byte, commitment []byte, shardID byte) error
	StoreOutputCoins(tokenID *common.Hash, pubkey []byte, outputcoin []byte, shardID byte) error
	FetchCommitments(tokenID *common.Hash, shardID byte) ([][]byte, error)
	HasCommitment(tokenID *common.Hash, commitment []byte, shardID byte) (bool, error)
	HasCommitmentIndex(tokenID *common.Hash, commitmentIndex uint64, shardID byte) (bool, error)
	GetCommitmentByIndex(tokenID *common.Hash, commitmentIndex uint64, shardID byte) ([]byte, error)
	GetCommitmentIndex(tokenID *common.Hash, commitment []byte, shardID byte) (*big.Int, error)
	GetCommitmentLength(tokenID *common.Hash, shardID byte) (*big.Int, error)
	GetCommitmentIndexsByPubkey(tokenID *common.Hash, pubkey []byte, shardID byte) ([][]byte, error)
	GetOutcoinsByPubkey(tokenID *common.Hash, pubkey []byte, shardID byte) ([][]byte, error)
	CleanCommitments() error

	// SNDerivator
	StoreSNDerivators(tokenID *common.Hash, data big.Int, shardID byte) error
	FetchSNDerivator(tokenID *common.Hash, shardID byte) ([]big.Int, error)
	HasSNDerivator(tokenID *common.Hash, data big.Int, shardID byte) (bool, error)
	CleanSNDerivator() error

	// Fee estimator
	StoreFeeEstimator([]byte, byte) error
	GetFeeEstimator(byte) ([]byte, error)
	CleanFeeEstimator() error

	// Custom token
	StoreCustomToken(tokenID *common.Hash, data []byte) error                                                    // store custom token. Param: tokenID, txInitToken-id, data tx
	StoreCustomTokenTx(tokenID *common.Hash, shardID byte, blockHeight uint64, txIndex int32, data []byte) error // store custom token tx. Param: tokenID, shardID, block height, tx-id, data tx
	ListCustomToken() ([][]byte, error)                                                                          // get list all custom token which issued in network, return init tx hash
	CustomTokenIDExisted(tokenID *common.Hash) bool                                                              // check tokenID existed in network, return init tx hash
	PrivacyCustomTokenIDExisted(tokenID *common.Hash) bool                                                       // check privacy tokenID existed in network
	CustomTokenTxs(tokenID *common.Hash) ([]*common.Hash, error)                                                 // from token id get all custom txs
	GetCustomTokenPaymentAddressUTXO(tokenID *common.Hash, paymentAddress []byte) (map[string]string, error)     // get list of utxo of an paymentaddress.pubkey of a token
	GetCustomTokenPaymentAddressesBalance(tokenID *common.Hash) (map[string]uint64, error)                       // get balance of all paymentaddress of a token (only return payment address with balance > 0)
	UpdateRewardAccountUTXO(tokenID *common.Hash, paymentAddress []byte, txHash *common.Hash, voutIndex int) error
	GetCustomTokenListPaymentAddress(*common.Hash) ([][]byte, error) // get all paymentaddress owner that have balance > 0 of a custom token
	GetCustomTokenPaymentAddressesBalanceUnreward(tokenID *common.Hash) (map[string]uint64, error)

	// privacy Custom token
	StorePrivacyCustomToken(tokenID *common.Hash, data []byte) error // store custom token. Param: tokenID, txInitToken-id, data tx
	StorePrivacyCustomTokenTx(tokenID *common.Hash, shardID byte, blockHeight uint64, txIndex int32, txHash []byte) error
	ListPrivacyCustomToken() ([][]byte, error)                          // get list all custom token which issued in network
	PrivacyCustomTokenTxs(tokenID *common.Hash) ([]*common.Hash, error) // from token id get all custom txs

	StorePrivacyCustomTokenCrossShard(tokenID *common.Hash, tokenValue []byte) error // store custom token cross shard privacy
	ListPrivacyCustomTokenCrossShard() ([][]byte, error)
	PrivacyCustomTokenIDCrossShardExisted(tokenID *common.Hash) bool

	// Reserve
	StoreIssuingInfo(reqTxID common.Hash, amount uint64, instType string) error
	GetIssuingInfo(reqTxID common.Hash) (uint64, string, error)
	StoreContractingInfo(reqTxID common.Hash, amount uint64, redeem uint64, instType string) error
	GetContractingInfo(reqTxID common.Hash) (uint64, uint64, string, error)

	// Centralized bridge
	CountUpDepositedAmtByTokenID(*common.Hash, uint64) error
	DeductAmtByTokenID(*common.Hash, uint64) error
	GetBridgeTokensAmounts() ([][]byte, error)
	IsBridgeTokenExisted(*common.Hash) (bool, error)

	Close() error
}
