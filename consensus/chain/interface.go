package chain

import (
	"time"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/wire"
)

type ConsensusInterface interface {
	GetInfo() string

	Start()
	Stop()
	IsRun() bool

	ReceiveProposeMsg(interface{})
	ReceivePrepareMsg(interface{})
}

type BlockInterface interface {
	GetHeight() uint64
	GetProducerPubKey() string
	// GetProducerSig() string
	Hash() *common.Hash
	AddValidationField(validateData string) error
	GetValidationField() string
}

type ChainInterface interface {
	GetConsensusEngine() ConsensusInterface
	PushMessageToValidator(wire.Message) error
	GetLastBlockTimeStamp() uint64
	GetBlkMinTime() time.Duration
	IsReady() bool
	GetHeight() uint64
	GetCommitteeSize() int
	GetNodePubKeyCommitteeIndex() int
	GetLastProposerIndex() int
	GetNodePubKey() string
	CreateNewBlock(round int) BlockInterface
	InsertBlk(interface{}, bool)
	ValidateBlock(interface{}) error
	ValidatePreSignBlock(interface{}) error
}

type Node interface {
	PushMessageToShard(wire.Message, byte) error
	PushMessageToBeacon(wire.Message) error
	GetNodePubKey() string
	GetUserKeySet() *incognitokey.KeySet
	GetActiveShardNumber() int
}
