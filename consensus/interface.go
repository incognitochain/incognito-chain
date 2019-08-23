package consensus

import (
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/wire"

	libp2p "github.com/libp2p/go-libp2p-peer"
)

type NodeInterface interface {
	PushMessageToShard(wire.Message, byte, map[libp2p.ID]bool) error
	PushMessageToBeacon(wire.Message, map[libp2p.ID]bool) error
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

	// ValidateBlock(block common.BlockInterface) error

	// ValidateProducerPosition(block common.BlockInterface) error
	ValidateProducerSig(block common.BlockInterface) error
	ValidateCommitteeSig(block common.BlockInterface, committee []incognitokey.CommitteePublicKey) error

	LoadUserKey(string) error
	LoadUserKeyFromIncPrivateKey(privateKey string) (string, error)
	GetUserPublicKey() *incognitokey.CommitteePublicKey
	// GetUserPrivateKey() string
	ValidateData(data []byte, sig string, publicKey string) error
	SignData(data []byte) (string, error)
	// ValidateAggregatedSig(dataHash *common.Hash, aggSig string, validatorPubkeyList []string) error
	// ValidateSingleSig(dataHash *common.Hash, sig string, pubkey string) error
}

// type ChainInterface interface {
// 	GetChainName() string
// 	GetConsensusType() string
// 	GetLastBlockTimeStamp() int64
// 	GetMinBlkInterval() time.Duration
// 	GetMaxBlkCreateTime() time.Duration
// 	IsReady() bool
// 	GetActiveShardNumber() int

// 	GetPubkeyRole(pubkey string, round int) (string, byte)
// 	CurrentHeight() uint64
// 	GetCommitteeSize() int
// 	GetCommittee() []string
// 	GetPubKeyCommitteeIndex(string) int
// 	GetLastProposerIndex() int

// 	CreateNewBlock(round int) common.BlockInterface
// 	PushMessageToValidators(wire.Message) error
// 	InsertBlk(common.BlockInterface, bool)
// 	ValidateBlock(common.BlockInterface) error
// 	ValidateBlockSanity(common.BlockInterface) error
// 	ValidateBlockWithBlockChain(common.BlockInterface) error
// 	GetShardID() int
// }

type BeaconInterface interface {
	blockchain.ChainInterface
	GetAllCommittees() map[string]map[string][]incognitokey.CommitteePublicKey
}
