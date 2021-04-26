package blockchain

import (
	"context"
	"time"

	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/multiview"
	"github.com/incognitochain/incognito-chain/pubsub"
)

type TxPool interface {
	HaveTransaction(hash *common.Hash) bool
	// RemoveTx remove tx from tx resource
	RemoveTx(txs []metadata.Transaction, isInBlock bool)
	RemoveCandidateList([]string)
	EmptyPool() bool
	MaybeAcceptTransactionForBlockProducing(metadata.Transaction, int64, *ShardBestState) (*metadata.TxDesc, error)
	MaybeAcceptBatchTransactionForBlockProducing(byte, []metadata.Transaction, int64, *ShardBestState) ([]*metadata.TxDesc, error)
}

type FeeEstimator interface {
	RegisterBlock(block *types.ShardBlock) error
}

type ConsensusEngine interface {
	ValidateProducerPosition(blk types.BlockInterface, lastProposerIdx int, committee []incognitokey.CommitteePublicKey, minCommitteeSize int) error
	ValidateProducerSig(block types.BlockInterface, consensusType string) error
	ValidateBlockCommitteSig(block types.BlockInterface, committee []incognitokey.CommitteePublicKey) error
	// CommitteeChange(chainName string)
}

type Highway interface {
	BroadcastCommittee(uint64, []incognitokey.CommitteePublicKey, map[byte][]incognitokey.CommitteePublicKey, map[byte][]incognitokey.CommitteePublicKey)
}

type Syncker interface {
	GetCrossShardBlocksForShardProducer(toShard byte, list map[byte][]uint64) map[byte][]interface{}
	GetCrossShardBlocksForShardValidator(toShard byte, list map[byte][]uint64) (map[byte][]interface{}, error)
	SyncMissingBeaconBlock(ctx context.Context, peerID string, fromHash common.Hash)
	SyncMissingShardBlock(ctx context.Context, peerID string, sid byte, fromHash common.Hash)
}

type Pubsub interface {
	PublishMessage(message *pubsub.Message)
}

type Chain interface {
	BestViewCommitteeFromBlock() common.Hash
	GetFinalView() multiview.View
	GetBestView() multiview.View
	GetEpoch() uint64
	GetChainName() string
	GetConsensusType() string
	GetLastBlockTimeStamp() int64
	GetMinBlkInterval() time.Duration
	GetMaxBlkCreateTime() time.Duration
	IsReady() bool
	SetReady(bool)
	GetActiveShardNumber() int
	CurrentHeight() uint64
	GetCommitteeSize() int
	IsBeaconChain() bool
	GetCommittee() []incognitokey.CommitteePublicKey
	GetPendingCommittee() []incognitokey.CommitteePublicKey
	GetPubKeyCommitteeIndex(string) int
	GetLastProposerIndex() int
	UnmarshalBlock(blockString []byte) (types.BlockInterface, error)
	CreateNewBlock(
		version int,
		proposer string,
		round int,
		startTime int64,
		committees []incognitokey.CommitteePublicKey,
		hash common.Hash) (types.BlockInterface, error)
	CreateNewBlockFromOldBlock(
		oldBlock types.BlockInterface,
		proposer string,
		startTime int64,
		committees []incognitokey.CommitteePublicKey,
		hash common.Hash) (types.BlockInterface, error)
	InsertBlock(block types.BlockInterface, shouldValidate bool) error
	ValidateBlockSignatures(block types.BlockInterface, committees []incognitokey.CommitteePublicKey) error
	ValidatePreSignBlock(block types.BlockInterface, signingCommittees, committees []incognitokey.CommitteePublicKey) error
	GetShardID() int
	GetChainDatabase() incdb.Database

	//for new syncker
	GetBestViewHeight() uint64
	GetFinalViewHeight() uint64
	GetBestViewHash() string
	GetFinalViewHash() string
	GetViewByHash(hash common.Hash) multiview.View
	CommitteeEngineVersion() int
	GetProposerByTimeSlot(
		committeeViewHash common.Hash,
		shardID byte,
		ts int64,
		committees []incognitokey.CommitteePublicKey,
	) (incognitokey.CommitteePublicKey, int, error)
	CommitteesFromViewHashForShard(committeeHash common.Hash, shardID byte) ([]incognitokey.CommitteePublicKey, error)
	ReplacePreviousValidationData(previousBlockHash common.Hash, newValidationData string) error
	SigningCommittees(
		committeeViewHash common.Hash,
		proposerIndex int,
		committees []incognitokey.CommitteePublicKey,
		shardID byte,
	) []incognitokey.CommitteePublicKey
}
