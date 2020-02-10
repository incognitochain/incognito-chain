package syncker

import "github.com/incognitochain/incognito-chain/incognitokey"

type Server interface {
	RequestBlock(peerID string, sID int, currentFinalHeight uint64, currentBestHash string, nBlocks uint64) chan interface{}
	RequestCrossShardBlockByViewRange(peerID string, sID int, latestCrossShardBlockHeight uint64) chan interface{}
	RequestShard2BeaconBlockByViewRange(peerID string, sID int, latestS2BHeight uint64) chan interface{}
}

type ViewInterface interface {
	GetHeight() uint64
	GetHash() string
}

type Chain interface {
	GetShardBestView() ViewInterface
	GetShardFinalView() ViewInterface
	InsertBlock(block interface{})
}

type PeerStateMsg struct {
	PeerID string
}

type Syncker struct {
	PeerStateCh chan PeerStateMsg
	UserPk      chan incognitokey.CommitteePublicKey
}

// Everytime block is created, we update the committee list so that Syncker know if it is in Committee or not
func UpdateCommittee(chain string, committees []incognitokey.CommitteePublicKey) {

}
