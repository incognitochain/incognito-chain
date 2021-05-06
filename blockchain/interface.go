package blockchain

import (
	"context"
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/consensus"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/multiview"
	"github.com/incognitochain/incognito-chain/pubsub"
	"time"
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
	EstimateFee(numBlocks uint64, tokenId *common.Hash) (uint64, error)
	GetLimitFeeForNativeToken() uint64
}

type ConsensusEngine interface {
	ValidateProducerPosition(blk types.BlockInterface, lastProposerIdx int, committee []incognitokey.CommitteePublicKey, minCommitteeSize int) error
	ValidateProducerSig(block types.BlockInterface, consensusType string) error
	ValidateBlockCommitteSig(block types.BlockInterface, committee []incognitokey.CommitteePublicKey) error
	GetCurrentMiningPublicKey() (string, string)
	// GetCurrentValidators() []*consensus.Validator
	// GetOneValidatorForEachConsensusProcess() map[int]*consensus.Validator
	// GetMiningPublicKeyByConsensus(consensusName string) (string, error)
	GetAllMiningPublicKeys() []string
	ExtractBridgeValidationData(block types.BlockInterface) ([][]byte, []int, error)
	GetAllValidatorKeyState() map[string]consensus.MiningState
	GetUserRole() (string, string, int)
	IsCommitteeInChain(chainID int) bool
	// CommitteeChange(chainName string)
}

type Server interface {
	PushBlockToAll(block types.BlockInterface, previousValidationData string, isBeacon bool) error
}

type Highway interface {
	BroadcastCommittee(uint64, []incognitokey.CommitteePublicKey, map[byte][]incognitokey.CommitteePublicKey, map[byte][]incognitokey.CommitteePublicKey)
}

type Syncker interface {
	GetCrossShardBlocksForShardProducer(curView *ShardBestState, list map[byte][]uint64) map[byte][]interface{}
	GetCrossShardBlocksForShardValidator(curView *ShardBestState, list map[byte][]uint64) (map[byte][]interface{}, error)
	SyncMissingBeaconBlock(ctx context.Context, peerID string, fromHash common.Hash)
	SyncMissingShardBlock(ctx context.Context, peerID string, sid byte, fromHash common.Hash)
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

type ChainInterface interface {
	GetBestView() multiview.View
	GetFinalView() multiview.View
}
