package blockchain

import (
	"context"
	"github.com/incognitochain/incognito-chain/wire"
	libp2p "github.com/libp2p/go-libp2p-core/peer"
	"time"

	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/consensus"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/metadata"
	portalprocessv4 "github.com/incognitochain/incognito-chain/portal/portalv4/portalprocess"
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
	MaybeAcceptSalaryTransactionForBlockProducing(byte, metadata.Transaction, int64, *ShardBestState) (*metadata.TxDesc, error)
	//CheckTransactionFee
	// CheckTransactionFee(tx metadata.Transaction) (uint64, error)
	// Check tx validate by it self
	// ValidateTxByItSelf(tx metadata.Transaction) bool
}

type FeeEstimator interface {
	RegisterBlock(block *types.ShardBlock) error
	EstimateFee(numBlocks uint64, tokenId *common.Hash) (uint64, error)
	GetLimitFeeForNativeToken() uint64
}

type ConsensusEngine interface {
	ValidateProducerPosition(blk types.BlockInterface, lastProposerIdx int, committee []incognitokey.CommitteePublicKey, minCommitteeSize int) error
	ValidateProducerSig(block types.BlockInterface, consensusType string) error
	ValidateBlockCommitteSig(block types.BlockInterface, committee []incognitokey.CommitteePublicKey) error
	GetCurrentMiningPublicKey() (string, string)
	GetAllMiningPublicKeys() []string
	ExtractBridgeValidationData(block types.BlockInterface) ([][]byte, []int, error)
	ExtractPortalV4ValidationData(block types.BlockInterface) ([]*portalprocessv4.PortalSig, error)
	GetAllValidatorKeyState() map[string]consensus.MiningState
	GetUserRole() (string, string, int)
	GetValidators() []*consensus.Validator
}

type Server interface {
	PushBlockToAll(block types.BlockInterface, previousValidationData string, isBeacon bool) error
	PushMessageToBeacon(msg wire.Message, exclusivePeerIDs map[libp2p.ID]bool) error
}

type Highway interface {
	BroadcastCommittee(uint64, []incognitokey.CommitteePublicKey, map[byte][]incognitokey.CommitteePublicKey, map[byte][]incognitokey.CommitteePublicKey)
	GetConnectionStatus() interface{}
}

type Syncker interface {
	GetCrossShardBlocksForShardProducer(view *ShardBestState, list map[byte][]uint64) map[byte][]interface{}
	GetCrossShardBlocksForShardValidator(view *ShardBestState, list map[byte][]uint64) (map[byte][]interface{}, error)
	SyncMissingBeaconBlock(ctx context.Context, peerID string, fromHash common.Hash)
	SyncMissingShardBlock(ctx context.Context, peerID string, sid byte, fromHash common.Hash)
	ReceiveBlock(block interface{}, previousValidationData string, peerID string)
}

type TxsCrawler interface {
	// RemoveTx remove tx from tx resource
	RemoveTxs(txs []metadata.Transaction)
	GetTxsTranferForNewBlock(sView interface{}, bcView interface{}, maxSize uint64, maxTime time.Duration) []metadata.Transaction
	CheckValidatedTxs(txs []metadata.Transaction) (valid []metadata.Transaction, needValidate []metadata.Transaction)
}

type Pubsub interface {
	PublishMessage(message *pubsub.Message)
}
