package peer

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	swarm "github.com/libp2p/go-libp2p-swarm"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	log.SetOutput(ioutil.Discard)
	m.Run()
}

var _ = func() (_ struct{}) {
	fmt.Println("This runs before init()!")
	Logger.Init(common.NewBackend(nil).Logger("test", true))
	return
}()

func TestPeer_HashToPool(t *testing.T) {
	peerObj := Peer{}
	peerObj.HashToPool("abc")
	value, ok := peerObj.messagePoolNew.Get("abc")
	if !ok {
		t.Error("Err HashToPool")
	}
	assert.Equal(t, 1, value)
}

func TestPeer_CheckHashPool(t *testing.T) {
	peerObj := Peer{}
	peerObj.HashToPool("abc")
	ok := peerObj.CheckHashPool("abc")
	assert.Equal(t, ok, true)
	time.Sleep(messageLiveTime + 1)
	ok = peerObj.CheckHashPool("abc")
	assert.Equal(t, ok, false)
}

func TestPeer_NewPeer(t *testing.T) {
	seed, _ := strconv.ParseInt(os.Getenv("LISTENER_PEER_SEED"), 10, 64)
	netAddr, err := common.ParseListener("127.0.0.1:9333", "ip")
	if err != nil {
		t.Error(err)
	}
	peerObj, err := Peer{
		Seed:             seed,
		ListeningAddress: *netAddr,
	}.NewPeer()
	if err != nil {
		t.Error(err)
	}
	assert.NotEmpty(t, peerObj.TargetAddress.String())
}

func TestPeer_Start(t *testing.T) {
	seed, _ := strconv.ParseInt(os.Getenv("LISTENER_PEER_SEED"), 10, 64)
	netAddr, err := common.ParseListener("127.0.0.1:9333", "ip")
	if err != nil {
		t.Error(err)
	}
	peerObj, err := Peer{
		Seed:             seed,
		ListeningAddress: *netAddr,
	}.NewPeer()
	close(peerObj.cStop)
	peerObj.Start()
}

func TestPeer_Stop(t *testing.T) {
	seed, _ := strconv.ParseInt(os.Getenv("LISTENER_PEER_SEED"), 10, 64)
	netAddr, err := common.ParseListener("127.0.0.1:9333", "ip")
	if err != nil {
		t.Error(err)
	}
	peerObj, err := Peer{
		Seed:             seed,
		ListeningAddress: *netAddr,
	}.NewPeer()
	go peerObj.Start()
	peerObj.Stop()
}

func TestPeer_PushConn(t *testing.T) {
	seed, _ := strconv.ParseInt(os.Getenv("LISTENER_PEER_SEED"), 10, 64)
	netAddr, err := common.ParseListener("127.0.0.1:9333", "ip")
	if err != nil {
		t.Error(err)
	}
	peerObj, err := Peer{
		Seed:             seed,
		ListeningAddress: *netAddr,
	}.NewPeer()

	peerConn := PeerConn{
		cMsgHash:   make(map[string]chan bool),
		isUnitTest: true,
		ListenerPeer: &Peer{
			PublicKey: "abc1",
		},
		RemotePeer: &Peer{
			PublicKey: "abc1",
		},
	}
	cConn := make(chan *PeerConn)
	peerObj.PushConn(peerConn.ListenerPeer, cConn)
	for {
		fmt.Print(111)
		select {
		case newPeerMsg := <-peerObj.cNewConn:
			{
				assert.Equal(t, newPeerMsg.Peer.PublicKey, "abc1")
				return
			}
		}
	}
}

func TestPeer_PushStream(t *testing.T) {
	seed, _ := strconv.ParseInt(os.Getenv("LISTENER_PEER_SEED"), 10, 64)
	netAddr, err := common.ParseListener("127.0.0.1:9333", "ip")
	if err != nil {
		t.Error(err)
	}
	peerObj, err := Peer{
		Seed:             seed,
		ListeningAddress: *netAddr,
	}.NewPeer()

	stream := &swarm.Stream{}
	peerObj.PushStream(stream)
	for {
		fmt.Print(111)
		select {
		case newStream := <-peerObj.cNewStream:
			{
				assert.Equal(t, newStream.Stream, stream)
				return
			}
		}
	}
}

func TestPeer_RemovePeerConn(t *testing.T) {
	seed, _ := strconv.ParseInt(os.Getenv("LISTENER_PEER_SEED"), 10, 64)
	netAddr, err := common.ParseListener("127.0.0.1:9333", "ip")
	if err != nil {
		t.Error(err)
	}
	peerObj, err := Peer{
		Seed:             seed,
		ListeningAddress: *netAddr,
		PeerConns:        make(map[string]*PeerConn),
	}.NewPeer()

	peerConn := &PeerConn{
		cMsgHash:   make(map[string]chan bool),
		isUnitTest: true,
		ListenerPeer: &Peer{
			PublicKey: "abc1",
		},
		RemotePeer: &Peer{
			PublicKey: "abc1",
		},
	}
	peerObj.setPeerConn(peerConn)
	assert.Equal(t, len(peerObj.PeerConns), 1)
	peerObj.removePeerConn(peerConn)
	assert.Equal(t, len(peerObj.PeerConns), 0)
}
