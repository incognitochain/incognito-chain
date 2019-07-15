package connmanager

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/peer"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

var _ = func() (_ struct{}) {
	fmt.Println("This runs before init()!")
	Logger.Init(common.NewBackend(nil).Logger("test", true))
	return
}()

func TestConnManager_GetPeerId(t *testing.T) {
	connManager := ConnManager{}
	connManager.New(&Config{})
	peerID, err := connManager.GetPeerId("/ip4/172.105.115.134/tcp/10008/ipfs/Qmawc4fM5VqzeeUMBb8PAEYVvENFK1DLATQsZpQFMskxJq")
	fmt.Println("PeerID 1", peerID)
	if err != nil {
		t.Error("Error GetPeerId")
	}
	if peerID == "" {
		t.Error("Error GetPeerId")
	}

	peerID, err = connManager.GetPeerId("")
	fmt.Println("PeerID 2", peerID)
	if err == nil {
		t.Error("Error GetPeerId")
	}
	if peerID != "" {
		t.Error("Error GetPeerId")
	}
}

func TestConnManager_GetPeerConnOfAll(t *testing.T) {
	peer1 := peer.Peer{
		PeerConnsMtx: sync.Mutex{},
	}
	mapPeerConnection := make(map[string]*peer.PeerConn)
	peerConn := peer.PeerConn{}
	mapPeerConnection[peerConn.RemotePeerID.String()] = &peerConn
	peer1.PeerConns = mapPeerConnection
	connManager := ConnManager{}.New(&Config{
		ListenerPeer: &peer1,
	})
	result := make([]*peer.PeerConn, 0)
	result = connManager.GetPeerConnOfAll()
	if len(result) == 0 {
		t.Error("Error GetPeerConnOfAll")
	}
}

func TestConnManager_GetPeerConnOfPublicKey(t *testing.T) {
	peer1 := peer.Peer{
		PeerConnsMtx: sync.Mutex{},
	}
	mapPeerConnection := make(map[string]*peer.PeerConn)
	peerConn1 := peer.PeerConn{RemotePeer: &peer.Peer{
		PublicKey: "abc1",
	},
		RemotePeerID: "a"}
	peerConn2 := peer.PeerConn{RemotePeer: &peer.Peer{
		PublicKey: "abc2",
	},
		RemotePeerID: "b"}
	mapPeerConnection[peerConn1.RemotePeerID.String()] = &peerConn1
	mapPeerConnection[peerConn2.RemotePeerID.String()] = &peerConn2
	peer1.PeerConns = mapPeerConnection
	connManager := ConnManager{}.New(&Config{
		ListenerPeer: &peer1,
	})
	result := make([]*peer.PeerConn, 0)
	pbk := "abc1"
	result = connManager.GetPeerConnOfPublicKey(pbk)
	if len(result) != 1 {
		t.Error("Error GetPeerConnOfPbk")
	}
}

func TestConnManager_GetPeerConnOfBeacon(t *testing.T) {
	consensusState := &ConsensusState{}
	beaconCommittee := []string{"abc1", "abc2"}

	consensusState.BeaconCommittee = make([]string, len(beaconCommittee))
	copy(consensusState.BeaconCommittee, beaconCommittee)

	peer1 := peer.Peer{
		PeerConnsMtx: sync.Mutex{},
	}
	mapPeerConnection := make(map[string]*peer.PeerConn)
	peerConn1 := peer.PeerConn{RemotePeer: &peer.Peer{
		PublicKey: "abc1",
	},
		RemotePeerID: "a"}
	peerConn2 := peer.PeerConn{RemotePeer: &peer.Peer{
		PublicKey: "abc2",
	},
		RemotePeerID: "b"}
	mapPeerConnection[peerConn1.RemotePeerID.String()] = &peerConn1
	mapPeerConnection[peerConn2.RemotePeerID.String()] = &peerConn2
	peer1.PeerConns = mapPeerConnection
	connManager := ConnManager{}.New(&Config{
		ListenerPeer: &peer1,
	})
	connManager.Config.ConsensusState = consensusState
	result := make([]*peer.PeerConn, 0)
	result = connManager.GetPeerConnOfBeacon()
	if len(result) != 2 {
		t.Error("Error GetPeerConnOfPbk")
	} else {
		assert.Equal(t, 1, 1)
	}
}

func TestConnManager_GetPeerConnOfShard(t *testing.T) {
	consensusState := &ConsensusState{
		ShardByCommittee: make(map[string]byte),
		CommitteeByShard: make(map[byte][]string),
	}

	shardCommittee := make(map[byte][]string)
	shardCommittee[0] = []string{"abc1", "b"}
	shardCommittee[2] = []string{"c", "abc2"}
	for shardID, committee := range shardCommittee {
		consensusState.CommitteeByShard[shardID] = make([]string, len(committee))
		copy(consensusState.CommitteeByShard[shardID], committee)
	}
	consensusState.rebuild()

	peer1 := peer.Peer{
		PeerConnsMtx: sync.Mutex{},
	}
	mapPeerConnection := make(map[string]*peer.PeerConn)
	peerConn1 := peer.PeerConn{RemotePeer: &peer.Peer{
		PublicKey: "abc1",
	},
		RemotePeerID: "a"}
	peerConn2 := peer.PeerConn{RemotePeer: &peer.Peer{
		PublicKey: "abc2",
	},
		RemotePeerID: "b"}
	mapPeerConnection[peerConn1.RemotePeerID.String()] = &peerConn1
	mapPeerConnection[peerConn2.RemotePeerID.String()] = &peerConn2
	peer1.PeerConns = mapPeerConnection
	connManager := ConnManager{}.New(&Config{
		ListenerPeer: &peer1,
	})
	connManager.Config.ConsensusState = consensusState
	result := make([]*peer.PeerConn, 0)
	result = connManager.GetPeerConnOfShard(0)
	if len(result) != 1 {
		t.Error("Error GetPeerConnOfPbk")
	} else {
		assert.Equal(t, 1, 1)
	}
	result = connManager.GetPeerConnOfShard(2)
	if len(result) != 1 {
		t.Error("Error GetPeerConnOfPbk")
	} else {
		assert.Equal(t, 1, 1)
	}
}

func TestConnManager_Start(t *testing.T) {
	connManager := ConnManager{}
	connManager.New(&Config{})
	connManager.Start("")
	connManager.Stop()
}
