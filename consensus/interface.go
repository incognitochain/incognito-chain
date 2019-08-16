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
	GetConsensusEngine() ConsensusInterface
	PushMessageToValidators(wire.Message) error
	GetLastBlockTimeStamp() uint64
	GetBlkMinTime() time.Duration
	IsReady() bool
	GetHeight() uint64
	GetCommitteeSize() int
	GetCommittee() []string
	GetPubKeyCommitteeIndex(string) int
	GetLastProposerIndex() int
	// GetNodePubKey() string
	CreateNewBlock(round int) common.BlockInterface
	InsertBlk(interface{}, bool)
	ValidateBlock(interface{}) error
	ValidateBlockSanity(interface{}) error
	ValidateBlockWithBlockChain(interface{}) error
	GetActiveShardNumber() int
	GetPubkeyRole(pubkey string, round int) (string, byte)
	GetShardID() byte
	GetConsensusType() string
}

// type MultisigSchemeInterface interface {
// 	LoadUserKey(string) error
// 	GetUserPublicKey() string
// 	GetUserPrivateKey() string
// 	SignData(data []byte) (string, error)
// 	ValidateAggSig(dataHash *common.Hash, aggSig string, validatorPubkeyList []string) error
// 	ValidateSingleSig(dataHash *common.Hash, sig string, pubkey string) error
// }
