package consensus

import (
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

// type KeyInterface interface{
// 	LoadKey(string) error
// 	GetPublicKey() string
// 	GetPrivateKey() string
// 	SigData(data []byte) (string,error)
// 	VerifyData(data []byte,sig []byte )
// }
