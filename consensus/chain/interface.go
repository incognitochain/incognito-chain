package chain

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/wire"
	"time"
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
	Hash() *common.Hash
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
	ValidateBlock(interface{}) int
	ValidateSignature(interface{}, string) bool
	InsertBlk(interface{}, bool)
}

type Node interface {
	PushMessageToShard(wire.Message, byte) error
	PushMessageToBeacon(wire.Message) error
	GetNodePubKey() string
	GetUserKeySet() *incognitokey.KeySet
	GetActiveShardNumber() int
}
