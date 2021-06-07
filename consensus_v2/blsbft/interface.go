package blsbft

import (
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
	signatureschemes2 "github.com/incognitochain/incognito-chain/consensus_v2/signatureschemes"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/multiview"
	"github.com/incognitochain/incognito-chain/wire"
	peer "github.com/libp2p/go-libp2p-peer"
	"time"
)

//Used interfaces

//NodeInterface
type NodeInterface interface {
	PushMessageToChain(msg wire.Message, chain common.ChainInterface) error
	PushBlockToAll(block types.BlockInterface, previousValidationData string, isBeacon bool) error
	IsEnableMining() bool
	GetMiningKeys() string
	GetPrivateKey() string
	GetUserMiningState() (role string, chainID int)
	RequestMissingViewViaStream(peerID string, hashes [][]byte, fromCID int, chainName string) (err error)
	GetSelfPeerID() peer.ID
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
	GetProposerByTimeSlotFromCommitteeList(
		ts int64,
		committees []incognitokey.CommitteePublicKey,
	) (incognitokey.CommitteePublicKey, int, error)
	CommitteesFromViewHashForShard(committeeHash common.Hash, shardID byte) ([]incognitokey.CommitteePublicKey, error)
	ReplacePreviousValidationData(previousBlockHash common.Hash, newValidationData string) error
	// GetSigningCommitteesFromBestView must be retrieve from a shard view, because it's based on the committee state version
	GetSigningCommittees(
		proposerIndex int,
		committees []incognitokey.CommitteePublicKey,
		blockVersion int,
	) []incognitokey.CommitteePublicKey
}

// TODO: @hung consider split interface
type CommitteeChainHandler interface {
	CommitteesFromViewHashForShard(committeeHash common.Hash, shardID byte) ([]incognitokey.CommitteePublicKey, error)
	FinalView() multiview.View
}

//Actor
type Actor interface {
	// GetConsensusName - retrieve consensus name
	GetConsensusName() string
	GetChainKey() string
	GetChainID() int
	// GetUserPublicKey - get user public key of loaded mining key
	GetUserPublicKey() *incognitokey.CommitteePublicKey
	// Start - start consensus
	Run() error
	// Stop - stop consensus
	Stop() error
	// IsOngoing - check whether consensus is currently voting on a block
	IsStarted() bool
	// ProcessBFTMsg - process incoming BFT message
	ProcessBFTMsg(msg *wire.MessageBFT)
	// LoadUserKey - load user mining key
	LoadUserKeys(miningKey []signatureschemes2.MiningKey)
	// ValidateData - validate data with this consensus signature scheme
	ValidateData(data []byte, sig string, publicKey string) error
	// SignData - sign data with this consensus signature scheme
	SignData(data []byte) (string, error)
	BlockVersion() int
}
