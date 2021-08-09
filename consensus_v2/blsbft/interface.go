package blsbft

import (
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/multiview"
	"github.com/incognitochain/incognito-chain/portal/portalv4"
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
	InsertAndBroadcastBlock(block types.BlockInterface) error
	InsertAndBroadcastBlockWithPrevValidationData(types.BlockInterface, string) error
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
	ReplacePreviousValidationData(previousBlockHash common.Hash, newValidationData string) error
	// GetSigningCommitteesFromBestView must be retrieve from a shard view, because it's based on the committee state version
	GetSigningCommittees(
		proposerIndex int,
		committees []incognitokey.CommitteePublicKey,
		blockVersion int,
	) []incognitokey.CommitteePublicKey
	GetPortalParamsV4(beaconHeight uint64) portalv4.PortalParams
}

type CommitteeChainHandler interface {
	CommitteesFromViewHashForShard(committeeHash common.Hash, shardID byte) ([]incognitokey.CommitteePublicKey, error)
	FinalView() multiview.View
}