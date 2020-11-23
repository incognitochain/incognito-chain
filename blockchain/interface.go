package blockchain

import (
	"context"
	"github.com/incognitochain/incognito-chain/multiview"

	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/pubsub"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/metadata"
)

type TxPool interface {
	// LastUpdated returns the last time a transaction was added to or
	// removed from the source pool.
	// LastUpdated() time.Time
	// MiningDescs returns a slice of mining descriptors for all the
	// transactions in the source pool.
	// MiningDescs() []*metadata.TxDesc
	// HaveTransaction returns whether or not the passed transaction hash
	// exists in the source pool.
	HaveTransaction(hash *common.Hash) bool
	// RemoveTx remove tx from tx resource
	RemoveTx(txs []metadata.Transaction, isInBlock bool)
	RemoveCandidateList([]string)
	EmptyPool() bool
	MaybeAcceptTransactionForBlockProducing(metadata.Transaction, int64, *ShardBestState) (*metadata.TxDesc, error)
	MaybeAcceptBatchTransactionForBlockProducing(byte, []metadata.Transaction, int64, *ShardBestState) ([]*metadata.TxDesc, error)
	//CheckTransactionFee
	// CheckTransactionFee(tx metadata.Transaction) (uint64, error)
	// Check tx validate by it self
	// ValidateTxByItSelf(tx metadata.Transaction) bool
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

type Server interface {
	// PublishNodeState() error

	// PushMessageGetBlockBeaconByHeight(from uint64, to uint64) error
	// PushMessageGetBlockBeaconByHash(blksHash []common.Hash, getFromPool bool, peerID libp2p.ID) error
	// PushMessageGetBlockBeaconBySpecificHeight(heights []uint64, getFromPool bool) error

	// PushMessageGetBlockShardByHeight(shardID byte, from uint64, to uint64) error
	// PushMessageGetBlockShardByHash(shardID byte, blksHash []common.Hash, getFromPool bool, peerID libp2p.ID) error
	// PushMessageGetBlockShardBySpecificHeight(shardID byte, heights []uint64, getFromPool bool) error

	PushBlockToAll(block types.BlockInterface, isBeacon bool) error
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

type ChainInterface interface {
	GetBestView() multiview.View
	GetFinalView() multiview.View
}
