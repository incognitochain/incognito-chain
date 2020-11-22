package txpool

import (
	"time"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/metadata"
)

type BeaconViewRetriever interface {
	GetAllCommitteeValidatorCandidate() (map[byte][]incognitokey.CommitteePublicKey, map[byte][]incognitokey.CommitteePublicKey, []incognitokey.CommitteePublicKey, []incognitokey.CommitteePublicKey, []incognitokey.CommitteePublicKey, []incognitokey.CommitteePublicKey, []incognitokey.CommitteePublicKey, []incognitokey.CommitteePublicKey, error)
	GetAllCommitteeValidatorCandidateFlattenListFromDatabase() ([]string, error)
	GetAutoStakingList() map[string]bool
	GetBeaconFeatureStateDB() *statedb.StateDB
	GetBeaconRewardStateDB() *statedb.StateDB
	GetBeaconSlashStateDB() *statedb.StateDB
}

type ShardViewRetriever interface {
	GetEpoch() uint64
	GetBeaconHeight() uint64
	GetStakingTx() map[string]string
	ListShardPrivacyTokenAndPRV() []common.Hash
	GetShardRewardStateDB() *statedb.StateDB
	GetCopiedFeatureStateDB() *statedb.StateDB
}

type TxPoolManager interface {
	GetShardTxsPool(shardID byte) (TxPool, error)
	StartShardTxsPool(shardID byte) error
	StopShardTxsPool(shardID byte) error
}

type TxPool interface {
	// RemoveTx remove tx from tx resource
	Start()
	Stop()
	RemoveTxs(txHashes []string)
	GetTxsTranferForNewBlock(
		sView interface{},
		bcView interface{},
		maxSize uint64,
		maxTime time.Duration,
	) []metadata.Transaction
	CheckValidatedTxs(
		txs []metadata.Transaction,
	) (
		valid []metadata.Transaction,
		needValidate []metadata.Transaction,
	)
}

type BlockTxsVerifier interface {
	ValidateBlockTransactions(
		txP TxPool,
		sView interface{},
		bcView interface{},
		txs []metadata.Transaction,
	) bool
	ValidateBatchRangeProof([]metadata.Transaction) (bool, error)
}

type TxVerifier interface {
	ValidateAuthentications(metadata.Transaction) (bool, error)
	ValidateDataCorrectness(metadata.Transaction) (bool, error)
	ValidateTxZKProof(metadata.Transaction) (bool, error)

	ValidateWithBlockChain(
		tx metadata.Transaction,
		sView interface{},
		bcView interface{},
	) (bool, error)

	ValidateDoubleSpend(
		txs []metadata.Transaction,
		sView interface{},
		bcView interface{},
	) (bool, error)

	ValidateTxAndAddToListTxs(
		txNew metadata.Transaction,
		txs []metadata.Transaction,
		sView interface{},
		bcView interface{},
		better func(txA, txB metadata.Transaction) bool, // return true if we want txA when txA and txB is double spending
	) (bool, error)

	FilterDoubleSpend(
		txs []metadata.Transaction,
		sView interface{},
		bcView interface{},
		better func(txA, txB metadata.Transaction) bool, // return true if we want txA when txA and txB is double spending
	) ([]metadata.Transaction, error)
}
