package connmanager

import (
	"fmt"
	"sync"
	"testing"

	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/peer"
	"github.com/stretchr/testify/assert"
)

var _ = func() (_ struct{}) {
	fmt.Println("This runs before init()!")
	Logger.Init(common.NewBackend(nil).Logger("test", true))
	return
}()

func TestConnManager_GetPeerId(t *testing.T) {
	connManager := New(&Config{})
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
	peer1 := peer.Peer{}
	peer1.SetPeerConnsMtx(&sync.Mutex{})
	mapPeerConnection := make(map[string]*peer.PeerConn)
	peerConn := peer.PeerConn{}
	mapPeerConnection[peerConn.GetRemotePeerID().String()] = &peerConn
	peer1.SetPeerConns(mapPeerConnection)
	connManager := New(&Config{
		ListenerPeer: &peer1,
	})
	result := make([]*peer.PeerConn, 0)
	result = connManager.GetPeerConnOfAll()
	if len(result) == 0 {
		t.Error("Error GetPeerConnOfAll")
	}
}

func TestConnManager_GetPeerConnOfPublicKey(t *testing.T) {
	peer1 := peer.Peer{}
	peer1.SetPeerConnsMtx(&sync.Mutex{})
	mapPeerConnection := make(map[string]*peer.PeerConn)

	peerConn1 := peer.PeerConn{}
	p1 := &peer.Peer{}
	p1.SetPublicKey("abc1", common.BlsConsensus)
	peerConn1.SetRemotePeer(p1)
	peerConn1.SetRemotePeerID("a")

	peerConn2 := peer.PeerConn{}
	p2 := &peer.Peer{}
	p2.SetPublicKey("abc2", common.BlsConsensus)
	peerConn2.SetRemotePeer(p2)
	peerConn2.SetRemotePeerID("b")

	mapPeerConnection[peerConn1.GetRemotePeerID().String()] = &peerConn1
	mapPeerConnection[peerConn2.GetRemotePeerID().String()] = &peerConn2
	peer1.SetPeerConns(mapPeerConnection)
	connManager := New(&Config{
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

	consensusState.beaconCommittee = make([]string, len(beaconCommittee))
	copy(consensusState.beaconCommittee, beaconCommittee)

	peer1 := peer.Peer{}
	peer1.SetPeerConnsMtx(&sync.Mutex{})
	mapPeerConnection := make(map[string]*peer.PeerConn)

	peerConn1 := peer.PeerConn{}
	p1 := &peer.Peer{}
	p1.SetPublicKey("abc1", common.BlsConsensus)
	peerConn1.SetRemotePeer(p1)
	peerConn1.SetRemotePeerID("a")

	peerConn2 := peer.PeerConn{}
	p2 := &peer.Peer{}
	p2.SetPublicKey("abc2", common.BlsConsensus)
	peerConn2.SetRemotePeer(p2)
	peerConn2.SetRemotePeerID("b")

	mapPeerConnection[peerConn1.GetRemotePeerID().String()] = &peerConn1
	mapPeerConnection[peerConn2.GetRemotePeerID().String()] = &peerConn2
	peer1.SetPeerConns(mapPeerConnection)
	connManager := New(&Config{
		ListenerPeer: &peer1,
	})
	connManager.UpdateConsensusState("", "", nil, consensusState.beaconCommittee, nil)
	result := make([]*peer.PeerConn, 0)
	result = connManager.GetPeerConnOfBeacon()
	if len(result) != 2 {
		assert.Equal(t, 2, len(result))
	} else {
		assert.Equal(t, 1, 1)
	}
}

func TestConnManager_GetPeerConnOfShard(t *testing.T) {
	consensusState := &ConsensusState{
		shardByCommittee: make(map[string]byte),
		committeeByShard: make(map[byte][]string),
	}

	shardCommittee := make(map[byte][]string)
	shardCommittee[0] = []string{"abc1", "b"}
	shardCommittee[2] = []string{"c", "abc2"}
	for shardID, committee := range shardCommittee {
		consensusState.committeeByShard[shardID] = make([]string, len(committee))
		copy(consensusState.committeeByShard[shardID], committee)
	}
	consensusState.rebuild()

	peer1 := peer.Peer{}
	peer1.SetPeerConnsMtx(&sync.Mutex{})
	mapPeerConnection := make(map[string]*peer.PeerConn)
	peerConn1 := peer.PeerConn{}
	p1 := &peer.Peer{}
	p1.SetPublicKey("abc1", common.BlsConsensus)
	peerConn1.SetRemotePeer(p1)
	peerConn1.SetRemotePeerID("a")

	peerConn2 := peer.PeerConn{}
	p2 := &peer.Peer{}
	p2.SetPublicKey("abc2", common.BlsConsensus)
	peerConn2.SetRemotePeer(p2)
	peerConn2.SetRemotePeerID("b")

	mapPeerConnection[peerConn1.GetRemotePeerID().String()] = &peerConn1
	mapPeerConnection[peerConn2.GetRemotePeerID().String()] = &peerConn2
	peer1.SetPeerConns(mapPeerConnection)
	connManager := New(&Config{
		ListenerPeer: &peer1,
	})
	connManager.UpdateConsensusState("", "", nil, nil, consensusState.committeeByShard)
	result := make([]*peer.PeerConn, 0)
	blockchain.NewBeaconBestStateWithConfig(&blockchain.Params{})
	bestState := blockchain.GetBeaconBestState()
	bestState.ShardCommittee[0] = []incognitokey.CommitteePublicKey{{MiningPubKey: map[string][]byte{common.BlsConsensus: []byte("abc1")}}}
	result = connManager.GetPeerConnOfShard(0)
	assert.Equal(t, 1, len(result))
	bestState.ShardCommittee[2] = []incognitokey.CommitteePublicKey{{MiningPubKey: map[string][]byte{common.BlsConsensus: []byte("abc2")}}}
	result = connManager.GetPeerConnOfShard(2)
	assert.Equal(t, 1, len(result))
}

func TestConnManager_Start(t *testing.T) {
	connManager := New(&Config{})
	connManager.Start("")
	err := connManager.Stop()
	assert.Equal(t, nil, err)
}
