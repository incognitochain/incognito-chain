package database

import (
	"math/big"

	"github.com/incognitochain/incognito-chain/common"
)

type BatchData struct {
	Key   []byte
	Value []byte
}

// DatabaseInterface provides the interface that is used to store blocks, txs, or any data of Incognito network.
type DatabaseInterface interface {
	// basic function
	Put(key, value []byte) error
	PutBatch(data []BatchData) error
	Get(key []byte) ([]byte, error)
	Delete(key []byte) error
	HasValue(key []byte) (bool, error)
	Close() error

	// Process on Block data
	StoreShardBlock(v interface{}, hash common.Hash, shardID byte, bd *[]BatchData) error
	FetchBlock(hash common.Hash) ([]byte, error)
	HasBlock(hash common.Hash) (bool, error)
	DeleteBlock(hash common.Hash, idx uint64, shardID byte) error

	// Process on Incomming Cross shard data
	StoreIncomingCrossShard(shardID byte, crossShardID byte, blkHeight uint64, crossBlkHash common.Hash, bd *[]BatchData) error
	HasIncomingCrossShard(shardID byte, crossShardID byte, crossBlkHash common.Hash) error
	GetIncomingCrossShard(shardID byte, crossShardID byte, crossBlkHash common.Hash) (uint64, error)
	DeleteIncomingCrossShard(shardID byte, crossShardID byte, crossBlkHash common.Hash) error

	// Process on Shard -> Beacon
	StoreAcceptedShardToBeacon(shardID byte, blkHeight uint64, shardBlkHash common.Hash) error
	HasAcceptedShardToBeacon(shardID byte, shardBlkHash common.Hash) error
	GetAcceptedShardToBeacon(shardID byte, shardBlkHash common.Hash) (uint64, error)
	DeleteAcceptedShardToBeacon(shardID byte, shardBlkHash common.Hash) error

	// Beacon
	StoreBeaconBlock(v interface{}, hash common.Hash, bd *[]BatchData) error
	FetchBeaconBlock(hash common.Hash) ([]byte, error)
	HasBeaconBlock(hash common.Hash) (bool, error)
	DeleteBeaconBlock(hash common.Hash, idx uint64) error

	//Crossshard
	StoreCrossShardNextHeight(fromShard byte, toShard byte, curHeight uint64, nextHeight uint64) error
	FetchCrossShardNextHeight(fromShard, toShard byte, curHeight uint64) (uint64, error)
	RestoreCrossShardNextHeights(fromShard byte, toShard byte, curHeight uint64) error

	// Block index
	StoreShardBlockIndex(hash common.Hash, idx uint64, shardID byte, bd *[]BatchData) error
	GetIndexOfBlock(hash common.Hash) (uint64, byte, error)
	GetBlockByIndex(idx uint64, shardID byte) (common.Hash, error)

	// Block index for beacon
	StoreBeaconBlockIndex(hash common.Hash, idx uint64) error
	GetIndexOfBeaconBlock(hash common.Hash) (uint64, error)
	GetBeaconBlockHashByIndex(idx uint64) (common.Hash, error)

	// Transaction index
	StoreTransactionIndex(txId common.Hash, blockHash common.Hash, indexInBlock int, bd *[]BatchData) error
	GetTransactionIndexById(txId common.Hash) (common.Hash, int, error)
	DeleteTransactionIndex(txId common.Hash) error

	// Best state of Prev
	StorePrevBestState(val []byte, isBeacon bool, shardID byte) error
	FetchPrevBestState(isBeacon bool, shardID byte) ([]byte, error)
	CleanBackup(isBeacon bool, shardID byte) error

	// Best state of shard chain
	StoreShardBestState(v interface{}, shardID byte, bd *[]BatchData) error
	FetchShardBestState(shardID byte) ([]byte, error)
	CleanShardBestState() error

	// Best state of beacon chain
	StoreBeaconBestState(v interface{}, bd *[]BatchData) error
	FetchBeaconBestState() ([]byte, error)
	CleanBeaconBestState() error

	// Commitee with epoch
	StoreShardCommitteeByHeight(height uint64, v interface{}) error
	StoreRewardReceiverByHeight(height uint64, v interface{}) error
	StoreBeaconCommitteeByHeight(height uint64, v interface{}) error
	StoreAutoStakingByHeight(height uint64, v interface{}) error
	DeleteCommitteeByHeight(blkEpoch uint64) error
	FetchShardCommitteeByHeight(height uint64) ([]byte, error)
	FetchRewardReceiverByHeight(height uint64) ([]byte, error)
	FetchBeaconCommitteeByHeight(height uint64) ([]byte, error)
	FetchAutoStakingByHeight(height uint64) ([]byte, error)
	HasShardCommitteeByHeight(height uint64) (bool, error)

	// SerialNumber
	StoreSerialNumbers(tokenID common.Hash, serialNumber [][]byte, shardID byte) error
	HasSerialNumber(tokenID common.Hash, data []byte, shardID byte) (bool, error)
	ListSerialNumber(tokenID common.Hash, shardID byte) (map[string]uint64, error)
	BackupSerialNumbersLen(tokenID common.Hash, shardID byte) error
	RestoreSerialNumber(tokenID common.Hash, shardID byte, serialNumbers [][]byte) error
	CleanSerialNumbers() error

	// PedersenCommitment
	StoreCommitments(tokenID common.Hash, pubkey []byte, commitment [][]byte, shardID byte) error
	StoreOutputCoins(tokenID common.Hash, publicKey []byte, outputCoinArr [][]byte, shardID byte) error
	HasCommitment(tokenID common.Hash, commitment []byte, shardID byte) (bool, error)
	ListCommitment(tokenID common.Hash, shardID byte) (map[string]uint64, error)
	ListCommitmentIndices(tokenID common.Hash, shardID byte) (map[uint64]string, error)
	HasCommitmentIndex(tokenID common.Hash, commitmentIndex uint64, shardID byte) (bool, error)
	GetCommitmentByIndex(tokenID common.Hash, commitmentIndex uint64, shardID byte) ([]byte, error)
	GetCommitmentIndex(tokenID common.Hash, commitment []byte, shardID byte) (*big.Int, error)
	GetCommitmentLength(tokenID common.Hash, shardID byte) (*big.Int, error)
	GetOutcoinsByPubkey(tokenID common.Hash, pubkey []byte, shardID byte) ([][]byte, error)
	BackupCommitmentsOfPubkey(tokenID common.Hash, shardID byte, pubkey []byte) error
	RestoreCommitmentsOfPubkey(tokenID common.Hash, shardID byte, pubkey []byte, commitments [][]byte) error
	DeleteOutputCoin(tokenID common.Hash, publicKey []byte, outputCoinArr [][]byte, shardID byte) error
	CleanCommitments() error

	// SNDerivator
	StoreSNDerivators(tokenID common.Hash, sndArray [][]byte) error
	HasSNDerivator(tokenID common.Hash, data []byte) (bool, error)
	CleanSNDerivator() error
	ListSNDerivator(tokenID common.Hash) ([][]byte, error)

	// Tx for Public key
	StoreTxByPublicKey(publicKey []byte, txID common.Hash, shardID byte) error
	GetTxByPublicKey(publicKey []byte) (map[byte][]common.Hash, error)

	// Fee estimator
	StoreFeeEstimator(val []byte, shardID byte) error
	GetFeeEstimator(shardID byte) ([]byte, error)
	CleanFeeEstimator() error

	// privacy token
	StorePrivacyToken(tokenID common.Hash, data []byte) error // store privacy token. Param: tokenID, txInitToken-id, data tx
	DeletePrivacyToken(tokenID common.Hash) error
	StorePrivacyTokenTx(tokenID common.Hash, shardID byte, blockHeight uint64, txIndex int32, txHash []byte) error
	DeletePrivacyTokenTx(tokenID common.Hash, txIndex int32, shardID byte, blockHeight uint64) error
	ListPrivacyToken() ([][]byte, error)                        // get list all privacy token which issued in network
	PrivacyTokenIDExisted(tokenID common.Hash) bool             // check privacy tokenID existed in network
	PrivacyTokenTxs(tokenID common.Hash) ([]common.Hash, error) // from token id get all privacy token txs

	// Privacy token for Cross Shard
	StorePrivacyTokenCrossShard(tokenID common.Hash, tokenValue []byte) error // store privacy token cross shard privacy
	ListPrivacyTokenCrossShard() ([][]byte, error)
	PrivacyTokenIDCrossShardExisted(tokenID common.Hash) bool
	DeletePrivacyTokenCrossShard(tokenID common.Hash) error

	// Centralized bridge
	BackupBridgedTokenByTokenID(tokenID common.Hash) error
	RestoreBridgedTokenByTokenID(tokenID common.Hash) error

	// Incognito -> Ethereum relay
	StoreBurningConfirm(txID common.Hash, height uint64, bd *[]BatchData) error
	GetBurningConfirm(txID common.Hash) (uint64, error)

	// Decentralized bridge
	IsBridgeTokenExistedByType(incTokenID common.Hash, isCentralized bool) (bool, error)
	InsertETHTxHashIssued(uniqETHTx []byte) error
	IsETHTxHashIssued(uniqETHTx []byte) (bool, error)
	CanProcessTokenPair(externalTokenID []byte, incTokenID common.Hash) (bool, error)
	CanProcessCIncToken(incTokenID common.Hash) (bool, error)
	UpdateBridgeTokenInfo(incTokenID common.Hash, externalTokenID []byte, isCentralized bool, updatingAmt uint64, updateType string, bd *[]BatchData) error
	GetAllBridgeTokens() ([]byte, error)
	TrackBridgeReqWithStatus(txReqID common.Hash, status byte, bd *[]BatchData) error
	GetBridgeReqWithStatus(txReqID common.Hash) (byte, error)

	// Block reward
	AddShardRewardRequest(epoch uint64, shardID byte, amount uint64, tokenID common.Hash, bd *[]BatchData) error
	GetRewardOfShardByEpoch(epoch uint64, shardID byte, tokenID common.Hash) (uint64, error)
	AddCommitteeReward(committeeAddress []byte, amount uint64, tokenID common.Hash) error
	GetCommitteeReward(committeeAddress []byte, tokenID common.Hash) (uint64, error)
	RemoveCommitteeReward(committeeAddress []byte, amount uint64, tokenID common.Hash, bd *[]BatchData) error
	ListCommitteeReward() map[string]map[common.Hash]uint64

	BackupShardRewardRequest(epoch uint64, shardID byte, tokenID common.Hash) error  //beacon
	BackupCommitteeReward(committeeAddress []byte, tokenID common.Hash) error        //shard
	RestoreShardRewardRequest(epoch uint64, shardID byte, tokenID common.Hash) error //beacon
	RestoreCommitteeReward(committeeAddress []byte, tokenID common.Hash) error       //shard

	// slash
	GetProducersBlackList(beaconHeight uint64) (map[string]uint8, error)
	StoreProducersBlackList(beaconHeight uint64, producersBlackList map[string]uint8) error

	// pde
	DeleteWaitingPDEContributionByPairID(beaconHeight uint64, pairID string) error
	ContributeToPDE(beaconHeight uint64, pairID string, contributorAddressStr string, tokenIDStr string, contributedAmount uint64) error
	GetPDEPoolForPair(beaconHeight uint64, tokenIDToBuyStr string, tokenIDToSellStr string) ([]byte, error)
	UpdatePDEPoolForPair(beaconHeight uint64, token1IDStr string, token2IDStr string, pdePoolForPairBytes []byte) error
	AddTradeFeeUp(beaconHeight uint64, token1IDStr string, token2IDStr string, tokenIDToBuyStr string, amt uint64) error
	DeductTradeFee(beaconHeight uint64, token1IDStr string, token2IDStr string, tokenIDToBuyStr string, amt uint64) error
	GetSharesOfContributorForTokenIDOnAPair(token1IDStr string, token2IDStr string, contributedTokenIDStr string, contributorAddrStr string) (uint64, error)
	GetTotalSharesForTokenIDOnAPair(token1IDStr string, token2IDStr string, contributedTokenIDStr string) (uint64, error)
	GetAllRecordsByPrefix(beaconHeight uint64, prefix []byte) ([][]byte, [][]byte, error)
	DeductSharesForWithdrawal(beaconHeight uint64, token1IDStr string, token2IDStr string, targetingTokenIDStr string, withdrawerAddressStr string, amt uint64) error
	GetLatestPDEPoolForPair(tokenIDToBuyStr string, tokenIDToSellStr string) ([]byte, error)
	TrackPDEStatus(prefix []byte, suffix []byte, status byte) error
	GetPDEStatus(prefix []byte, suffix []byte) (byte, error)
}
