package database

import (
	"math/big"

	"github.com/incognitochain/incognito-chain/common"
	rCommon "github.com/incognitochain/incognito-chain/ethrelaying/common"
)

// DatabaseInterface provides the interface that is used to store blocks.
type DatabaseInterface interface {
	Put(key, value []byte) error
	Get(key []byte) ([]byte, error)
	Delete(key []byte) error
	HasValue(key []byte) (bool, error)

	// Block
	StoreShardBlock(interface{}, common.Hash, byte) error
	FetchBlock(common.Hash) ([]byte, error)
	HasBlock(common.Hash) (bool, error)
	DeleteBlock(common.Hash, uint64, byte) error

	StoreIncomingCrossShard(shardID byte, crossShardID byte, blkHeight uint64, crossBlkHash common.Hash) error
	HasIncomingCrossShard(shardID byte, crossShardID byte, crossBlkHash common.Hash) error
	GetIncomingCrossShard(shardID byte, crossShardID byte, crossBlkHash common.Hash) (uint64, error)
	DeleteIncomingCrossShard(shardID byte, crossShardID byte, crossBlkHash common.Hash) error

	StoreAcceptedShardToBeacon(shardID byte, blkHeight uint64, shardBlkHash common.Hash) error
	HasAcceptedShardToBeacon(shardID byte, shardBlkHash common.Hash) error
	GetAcceptedShardToBeacon(shardID byte, shardBlkHash common.Hash) (uint64, error)
	DeleteAcceptedShardToBeacon(shardID byte, shardBlkHash common.Hash) error

	// Beacon
	StoreBeaconBlock(interface{}, common.Hash) error
	FetchBeaconBlock(common.Hash) ([]byte, error)
	HasBeaconBlock(common.Hash) (bool, error)
	DeleteBeaconBlock(common.Hash, uint64) error

	//Crossshard
	StoreCrossShardNextHeight(byte, byte, uint64, uint64) error
	FetchCrossShardNextHeight(byte, byte, uint64) (uint64, error)
	RestoreCrossShardNextHeights(byte, byte, uint64) error

	// Block index
	StoreShardBlockIndex(common.Hash, uint64, byte) error
	GetIndexOfBlock(common.Hash) (uint64, byte, error)
	GetBlockByIndex(uint64, byte) (common.Hash, error)

	// Block index
	StoreBeaconBlockIndex(common.Hash, uint64) error
	GetIndexOfBeaconBlock(common.Hash) (uint64, error)
	GetBeaconBlockHashByIndex(uint64) (common.Hash, error)

	// Transaction index
	StoreTransactionIndex(txId common.Hash, blockHash common.Hash, indexInBlock int) error
	GetTransactionIndexById(txId common.Hash) (common.Hash, int, *DatabaseError)
	DeleteTransactionIndex(txId common.Hash) error

	// Best state of shard chain
	StorePrevBestState([]byte, bool, byte) error
	FetchPrevBestState(bool, byte) ([]byte, error)
	CleanBackup(bool, byte) error
	StoreShardBestState(interface{}, byte) error
	FetchShardBestState(byte) ([]byte, error)
	CleanShardBestState() error

	// Best state of beacon chain
	StoreBeaconBestState(interface{}) error
	StoreCommitteeByHeight(uint64, interface{}) error
	StoreCommitteeByEpoch(uint64, interface{}) error
	DeleteCommitteeByEpoch(uint64) error

	FetchCommitteeByEpoch(uint64) ([]byte, error)
	HasCommitteeByEpoch(uint64) (bool, error)
	FetchBeaconBestState() ([]byte, error)
	CleanBeaconBestState() error

	// SerialNumber
	StoreSerialNumbers(tokenID common.Hash, serialNumber [][]byte, shardID byte) error
	HasSerialNumber(tokenID common.Hash, data []byte, shardID byte) (bool, error)
	BackupSerialNumbersLen(tokenID common.Hash, shardID byte) error
	RestoreSerialNumber(tokenID common.Hash, shardID byte, serialNumbers [][]byte) error
	// DeleteSerialNumber(tokenID common.Hash, data []byte, shardID byte) error
	CleanSerialNumbers() error

	// PedersenCommitment
	StoreCommitments(tokenID common.Hash, pubkey []byte, commitment [][]byte, shardID byte) error
	StoreOutputCoins(tokenID common.Hash, publicKey []byte, outputCoinArr [][]byte, shardID byte) error
	HasCommitment(tokenID common.Hash, commitment []byte, shardID byte) (bool, error)
	HasCommitmentIndex(tokenID common.Hash, commitmentIndex uint64, shardID byte) (bool, error)
	GetCommitmentByIndex(tokenID common.Hash, commitmentIndex uint64, shardID byte) ([]byte, error)
	GetCommitmentIndex(tokenID common.Hash, commitment []byte, shardID byte) (*big.Int, error)
	GetCommitmentLength(tokenID common.Hash, shardID byte) (*big.Int, error)
	GetOutcoinsByPubkey(tokenID common.Hash, pubkey []byte, shardID byte) ([][]byte, error)
	BackupCommitmentsOfPubkey(tokenID common.Hash, shardID byte, pubkey []byte) error
	RestoreCommitmentsOfPubkey(tokenID common.Hash, shardID byte, pubkey []byte, commitments [][]byte) error
	BackupOutputCoin(tokenID common.Hash, pubkey []byte, shardID byte) error
	RestoreOutputCoin(tokenID common.Hash, pubkey []byte, shardID byte) error
	CleanCommitments() error

	// SNDerivator
	StoreSNDerivators(tokenID common.Hash, sndArray [][]byte, shardID byte) error
	HasSNDerivator(tokenID common.Hash, data []byte, shardID byte) (bool, error)
	CleanSNDerivator() error

	// Fee estimator
	StoreFeeEstimator([]byte, byte) error
	GetFeeEstimator(byte) ([]byte, error)
	CleanFeeEstimator() error

	// Custom token
	StoreCustomToken(tokenID common.Hash, data []byte) error // store custom token. Param: tokenID, txInitToken-id, data tx
	DeleteCustomToken(tokenID common.Hash) error
	StoreCustomTokenTx(tokenID common.Hash, shardID byte, blockHeight uint64, txIndex int32, data []byte) error // store custom token tx. Param: tokenID, shardID, block height, tx-id, data tx
	DeleteCustomTokenTx(tokenID common.Hash, txIndex int32, shardID byte, blockHeight uint64) error
	ListCustomToken() ([][]byte, error)                                                                     // get list all custom token which issued in network, return init tx hash
	CustomTokenIDExisted(tokenID common.Hash) bool                                                          // check tokenID existed in network, return init tx hash
	PrivacyCustomTokenIDExisted(tokenID common.Hash) bool                                                   // check privacy tokenID existed in network
	CustomTokenTxs(tokenID common.Hash) ([]common.Hash, error)                                              // from token id get all custom txs
	GetCustomTokenPaymentAddressUTXO(tokenID common.Hash, paymentAddress []byte) (map[string]string, error) // get list of utxo of an paymentaddress.pubkey of a token
	GetCustomTokenPaymentAddressesBalance(tokenID common.Hash) (map[string]uint64, error)                   // get balance of all paymentaddress of a token (only return payment address with balance > 0)

	// privacy Custom token
	StorePrivacyCustomToken(tokenID common.Hash, data []byte) error // store custom token. Param: tokenID, txInitToken-id, data tx
	DeletePrivacyCustomToken(tokenID common.Hash) error
	StorePrivacyCustomTokenTx(tokenID common.Hash, shardID byte, blockHeight uint64, txIndex int32, txHash []byte) error
	DeletePrivacyCustomTokenTx(tokenID common.Hash, txIndex int32, shardID byte, blockHeight uint64) error
	ListPrivacyCustomToken() ([][]byte, error)                        // get list all custom token which issued in network
	PrivacyCustomTokenTxs(tokenID common.Hash) ([]common.Hash, error) // from token id get all custom txs

	StorePrivacyCustomTokenCrossShard(tokenID common.Hash, tokenValue []byte) error // store custom token cross shard privacy
	ListPrivacyCustomTokenCrossShard() ([][]byte, error)
	PrivacyCustomTokenIDCrossShardExisted(tokenID common.Hash) bool
	DeletePrivacyCustomTokenCrossShard(tokenID common.Hash) error

	// Centralized bridge
	GetBridgeTokensAmounts() ([][]byte, error)
	IsBridgeTokenExisted(common.Hash) (bool, error)
	UpdateAmtByTokenID(common.Hash, uint64, string) error

	// Decentralized bridge
	InsertETHTxHashIssued(rCommon.Hash) error
	IsETHTxHashIssued(rCommon.Hash) (bool, error)

	// block reward
	AddShardRewardRequest(epoch uint64, shardID byte, rewardAmount uint64) error
	GetRewardOfShardByEpoch(epoch uint64, shardID byte) (uint64, error)
	AddCommitteeReward(committeeAddress []byte, amount uint64) error
	GetCommitteeReward(committeeAddress []byte) (uint64, error)
	RemoveCommitteeReward(committeeAddress []byte, amount uint64) error

	Close() error
}
