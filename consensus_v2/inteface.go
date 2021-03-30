package consensus_v2

import (
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/pubsub"
	"github.com/incognitochain/incognito-chain/wire"
	peer "github.com/libp2p/go-libp2p-peer"
)

type EngineConfig struct {
	Node          NodeInterface
	Blockchain    *blockchain.BlockChain
	PubSubManager *pubsub.PubSubManager
}

//Used interfaces
//NodeInterface
type NodeInterface interface {
	PushMessageToChain(msg wire.Message, chain common.ChainInterface) error
	IsEnableMining() bool
	GetMiningKeys() string
	GetPrivateKey() string
	GetUserMiningState() (role string, chainID int)
	GetPubkeyMiningState(*incognitokey.CommitteePublicKey) (role string, chainID int)
	RequestMissingViewViaStream(peerID string, hashes [][]byte, fromCID int, chainName string) (err error)
	GetSelfPeerID() peer.ID
}

//
