package consensus

import (
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/wire"
)

type NodeInterface interface {
	PushMessageToChain(msg wire.Message, chain blockchain.ChainInterface) error
	// PushMessageToBlockToAll(msg wire.Message) error
	UpdateConsensusState(role string, userPbk string, currentShard *byte, beaconCommittee []string, shardCommittee map[byte][]string)
	IsEnableMining() bool
	GetMiningKeys() string
	GetPrivateKey() string
	DropAllConnections()
}

type ConsensusInterface interface {
	// NewInstance - Create a new instance of this consensus
	NewInstance(chain blockchain.ChainInterface, chainKey string, node NodeInterface, logger common.Logger) ConsensusInterface
	// GetConsensusName - retrieve consensus name
	GetConsensusName() string
	// Start - start consensus
	Start() error
	// Stop - stop consensus
	Stop() error
	// IsOngoing - check whether consensus is currently voting on a block
	IsOngoing() bool
	// ProcessBFTMsg - process incoming BFT message
	ProcessBFTMsg(msg *wire.MessageBFT)
	// ValidateProducerSig - validate a block producer signature
	ValidateProducerSig(block common.BlockInterface) error
	// ValidateCommitteeSig - validate a block committee signature
	ValidateCommitteeSig(block common.BlockInterface, committee []incognitokey.CommitteePublicKey) error
	// LoadUserKey - load user mining key
	LoadUserKey(miningKey string) error
	// LoadUserKeyFromIncPrivateKey - load user mining key from incognito privatekey
	LoadUserKeyFromIncPrivateKey(privateKey string) (string, error)
	// GetUserPublicKey - get user public key of loaded mining key
	GetUserPublicKey() *incognitokey.CommitteePublicKey
	// ValidateData - validate data with this consensus signature scheme
	ValidateData(data []byte, sig string, publicKey string) error
	// SignData - sign data with this consensus signature scheme
	SignData(data []byte) (string, error)
	// ExtractBridgeValidationData - extract bridge related field in validation data of block
	ExtractBridgeValidationData(block common.BlockInterface) ([][]byte, []int, error)
}

type BeaconInterface interface {
	blockchain.ChainInterface
	GetAllCommittees() map[string]map[string][]incognitokey.CommitteePublicKey
	GetBeaconPendingList() []incognitokey.CommitteePublicKey
	GetShardsPendingList() map[string]map[string][]incognitokey.CommitteePublicKey
	GetShardsWaitingList() []incognitokey.CommitteePublicKey
	GetBeaconWaitingList() []incognitokey.CommitteePublicKey
}
