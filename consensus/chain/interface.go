package chain

import (
	"time"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/wire"
)

type ConsensusEngineInterface interface {
	Start()
	Stop()
	IsOngoing(chainkey string) bool

	ProcessBFTMsg(msg *wire.MessageBFT)
	ValidateBlockWithConsensus(block BlockInterface, chainName string, consensusType string) error
}

type ConsensusInterface interface {
	NewInstance() ConsensusInterface
	GetConsensusName() string

	Start()
	Stop()
	IsOngoing() bool

	ProcessBFTMsg(msg *wire.MessageBFT)

	ValidateBlock(block BlockInterface) error
}

type BlockInterface interface {
	GetHeight() uint64
	Hash() *common.Hash
	AddValidationField(validateData string) error
	GetValidationField() string
	GetRound() int
	GetRoundKey() string
}

type ChainInterface interface {
	GetConsensusEngine() ConsensusEngineInterface
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
	CreateNewBlock(round int) BlockInterface
	InsertBlk(interface{}, bool)
	ValidateBlock(interface{}) error
	ValidateBlockSanity(interface{}) error
	ValidateBlockWithBlockChain(interface{}) error
	GetActiveShardNumber() int
}

type Node interface {
	PushMessageToShard(wire.Message, byte) error
	PushMessageToBeacon(wire.Message) error
	IsEnableMining() bool
	GetMiningKey() string
}
