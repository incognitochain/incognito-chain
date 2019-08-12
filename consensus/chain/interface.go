package chain

import (
	"time"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/wire"

	libp2p "github.com/libp2p/go-libp2p-peer"
)

type ConsensusEngineInterface interface {
	Start()
	Stop()
	IsOngoing(chainkey string) bool

	ProcessBFTMsg(msg *wire.MessageBFT)

	LoadMiningKeys(keys string) error
	GetMiningPublicKey() (publickey string, keyType string)
	SignDataWithMiningKey(data []byte) (string, error)

	ValidateProducerPosition(block BlockInterface, chain ChainInterface) error
	ValidateProducerSig(block BlockInterface, chain ChainInterface) error
	ValidateCommitteeSig(block BlockInterface, chain ChainInterface) error

	VerifyData(data []byte, sig string, publicKey string, consensusType string) error

	SwitchConsensus(chainkey string, consensus string) error
}

type ConsensusInterface interface {
	NewInstance() ConsensusInterface
	GetConsensusName() string

	Start()
	Stop()
	IsOngoing() bool

	ProcessBFTMsg(msg *wire.MessageBFT)

	ValidateBlock(block BlockInterface) error

	ValidateProducerPosition(block BlockInterface) error
	ValidateProducerSig(block BlockInterface) error
	ValidateCommitteeSig(block BlockInterface) error
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
	GetPubkeyRole(pubkey string, round int) (string, byte)
}

type Node interface {
	PushMessageToShard(wire.Message, byte, map[libp2p.ID]bool) error
	PushMessageToBeacon(wire.Message, map[libp2p.ID]bool) error
	IsEnableMining() bool
	GetMiningKeys() string
}
