package blsbft

import (
	"time"

	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
	signatureschemes2 "github.com/incognitochain/incognito-chain/consensus_v2/signatureschemes"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/wire"
)

//Used interfaces

//NodeInterface
type NodeInterface interface {
	PushMessageToChain(msg wire.Message, chain common.ChainInterface) error
	GetMiningKeys() string
	GetPrivateKey() string
	GetUserMiningState() (role string, chainID int)
}

///

type ChainInterface interface {
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
	GetCommittee() []incognitokey.CommitteePublicKey
	GetPendingCommittee() []incognitokey.CommitteePublicKey
	GetPubKeyCommitteeIndex(string) int
	GetLastProposerIndex() int
	UnmarshalBlock(blockString []byte) (types.BlockInterface, error)

	InsertAndBroadcastBlock(block types.BlockInterface) error
	CreateNewBlock(
		version int,
		proposer string,
		round int,
		startTime int64,
		committees []incognitokey.CommitteePublicKey,
		hash common.Hash) (types.BlockInterface, error)
	ValidateBlockSignatures(block types.BlockInterface, committees []incognitokey.CommitteePublicKey) error
	ValidatePreSignBlock(block types.BlockInterface, signingCommittes, committees []incognitokey.CommitteePublicKey) error
	GetShardID() int

	//for new syncker
	GetBestViewHeight() uint64
	GetFinalViewHeight() uint64
	GetBestViewHash() string
	GetFinalViewHash() string
}

//Exported interfaces

//Actor
type Actor interface {
	// GetConsensusName - retrieve consensus name
	GetConsensusName() string
	GetChainKey() string
	GetChainID() int
	// Start - start consensus
	Start() error
	// Stop - stop consensus
	Stop() error
	Destroy()
	// IsOngoing - check whether consensus is currently voting on a block
	IsOngoing() bool
	IsStarted() bool
	// ProcessBFTMsg - process incoming BFT message
	ProcessBFTMsg(msg *wire.MessageBFT)
	// LoadUserKey - load user mining key
	LoadUserKeys(miningKey []signatureschemes2.MiningKey) error
	// GetUserPublicKey - get user public key of loaded mining key
	GetUserPublicKey() *incognitokey.CommitteePublicKey
	// ValidateData - validate data with this consensus signature scheme
	ValidateData(data []byte, sig string, publicKey string) error
	// SignData - sign data with this consensus signature scheme
	SignData(data []byte) (string, error)
}

///
