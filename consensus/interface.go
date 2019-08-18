package consensus

import (
	"time"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/wire"

	libp2p "github.com/libp2p/go-libp2p-peer"
)

type nodeInterface interface {
	PushMessageToShard(wire.Message, byte, map[libp2p.ID]bool) error
	PushMessageToBeacon(wire.Message, map[libp2p.ID]bool) error
	IsEnableMining() bool
	GetMiningKeys() string
}

type ConsensusInterface interface {
	NewInstance() ConsensusInterface
	GetConsensusName() string

	Start()
	Stop()
	IsOngoing() bool

	ProcessBFTMsg(msg *wire.MessageBFT)

	// ValidateBlock(block common.BlockInterface) error

	// ValidateProducerPosition(block common.BlockInterface) error
	ValidateProducerSig(blockHash *common.Hash, validationData string) error
	ValidateCommitteeSig(blockHash *common.Hash, committee []string, validationData string) error

	LoadUserKey(string) error
	GetUserPublicKey() string
	GetUserPrivateKey() string
	// SignData(data []byte) (string, error)
	// ValidateAggregatedSig(dataHash *common.Hash, aggSig string, validatorPubkeyList []string) error
	// ValidateSingleSig(dataHash *common.Hash, sig string, pubkey string) error
}

type ChainInterface interface {
	GetChainName() string
	GetConsensusType() string
	GetLastBlockTimeStamp() int64
	GetMinBlkInterval() time.Duration
	GetMaxBlkCreateTime() time.Duration
	IsReady() bool
	GetActiveShardNumber() int

	GetPubkeyRole(pubkey string, round int) (string, byte)
	CurrentHeight() uint64
	GetCommitteeSize() int
	GetCommittee() []string
	GetPubKeyCommitteeIndex(string) int
	GetLastProposerIndex() int

	CreateNewBlock(round int) common.BlockInterface
	InsertBlk(common.BlockInterface, bool)
	ValidateBlock(common.BlockInterface) error
	ValidateBlockSanity(common.BlockInterface) error
	ValidateBlockWithBlockChain(common.BlockInterface) error
	GetShardID() int
}

type BeaconInterface interface {
	ChainInterface
	GetAllCommittees() map[string]map[string][]string
}

// type MultisigSchemeInterface interface {
// 	LoadUserKey(string) error
// 	GetUserPublicKey() string
// 	GetUserPrivateKey() string
// 	SignData(data []byte) (string, error)
// 	ValidateAggSig(dataHash *common.Hash, aggSig string, validatorPubkeyList []string) error
// 	ValidateSingleSig(dataHash *common.Hash, sig string, pubkey string) error
// }
