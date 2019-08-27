package consensus

import (
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/wire"

	libp2p "github.com/libp2p/go-libp2p-peer"
)

type NodeInterface interface {
	PushMessageToShard(msg wire.Message, shard byte, exclusivePeerIDs map[libp2p.ID]bool) error
	PushMessageToBeacon(msg wire.Message, exclusivePeerIDs map[libp2p.ID]bool) error
	PushMessageToChain(msg wire.Message, chain blockchain.ChainInterface) error
	UpdateConsensusState(role string, userPbk string, currentShard *byte, beaconCommittee []string, shardCommittee map[byte][]string)
	IsEnableMining() bool
	GetMiningKeys() string
	GetPrivateKey() string
}

type ConsensusInterface interface {
	NewInstance(chain blockchain.ChainInterface, chainKey string, node NodeInterface, logger common.Logger) ConsensusInterface
	GetConsensusName() string

	Start()
	Stop()
	IsOngoing() bool

	ProcessBFTMsg(msg *wire.MessageBFT)

	ValidateProducerSig(block common.BlockInterface) error
	ValidateCommitteeSig(block common.BlockInterface, committee []incognitokey.CommitteePublicKey) error

	LoadUserKey(miningKey string) error
	LoadUserKeyFromIncPrivateKey(privateKey string) (string, error)
	GetUserPublicKey() *incognitokey.CommitteePublicKey
	ValidateData(data []byte, sig string, publicKey string) error
	SignData(data []byte) (string, error)
	ExtractBridgeValidationData(block common.BlockInterface) ([][]byte, []int, error)
}

type BeaconInterface interface {
	blockchain.ChainInterface
	GetAllCommittees() map[string]map[string][]incognitokey.CommitteePublicKey
}
