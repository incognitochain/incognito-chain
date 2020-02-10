package syncker

import (
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/wire"
)

type Server interface {
	//Request block from "peerID" of shard "fromSID" with param currentFinalHeight and currentBestHash
	//Function return channel of each block, and a stop channel to tell sender side to stop send block
	RequestBlock(peerID string, fromSID int, currentFinalHeight uint64, currentBestHash string) (blockCh chan interface{}, stopCh chan int)

	//Request cross block from "peerID" for shard "toShardID" with param latestCrossShardBlockHeight in current pool
	//Function return channel of each block, and a stop channel to tell sender side to stop send block
	RequestCrossShardBlockPool(peerID string, toShardID int, latestCrossShardBlockHeight uint64) (blockCh chan interface{}, stopCh chan int)

	//Request s2b block from "peerID" of shard "fromSID" with param latestS2BHeight in current pool
	//Function return channel of each block, and a stop channel to tell sender side to stop send block
	RequestS2BBlockPool(peerID string, fromSID int, latestS2BHeight uint64) (blockCh chan interface{}, stopCh chan int)

	GetCrossShardPool(sid byte) Pool
	GetS2BPool(sid byte) Pool
}

type Pool interface {
	GetLatestCrossShardFinalHeight(byte) uint64
	GetLatestFinalHeight() uint64
	AddBlock(block interface{}) error
}

type ViewInterface interface {
	GetHeight() uint64
	GetHash() string
}

type Chain interface {
	GetBestView() ViewInterface
	GetFinalView() ViewInterface
	InsertBlock(block interface{}) error
}

type PeerStateMsg struct {
	PeerID string
}

type Syncker struct {
	PeerStateCh         chan wire.MessagePeerState
	UserPk              incognitokey.CommitteePublicKey
	BeaconPeerState     map[string]BeaconPeerState         //sender -> state
	ShardPeerState      map[byte]map[string]ShardPeerState //sid -> sender -> state
	S2BPeerState        map[byte]map[string]ShardPeerState //sid -> sender -> state
	CrossShardPeerState map[byte]map[string]ShardPeerState //sid -> sender -> state
}

// Everytime block is created, we update the committee list so that Syncker know if it is in Committee or not
func UpdateCommittee(chain string, committees []incognitokey.CommitteePublicKey) {

}

func NewSyncker(userPk incognitokey.CommitteePublicKey) *Syncker {
	s := &Syncker{
		PeerStateCh: make(chan wire.MessagePeerState),
		UserPk:      userPk,
	}
	go func() {
		for {
			select {
			case PeerState := <-s.PeerStateCh:

			}
		}

	}()
	return s
}
