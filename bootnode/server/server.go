package server

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
	"sync"
	"time"

	"github.com/incognitochain/incognito-chain/cashec"
)

const (
	HeartbeatInterval = 10
	HeartbeatTimeout  = 60
)

// timeZeroVal is simply the zero value for a time.Time and is used to avoid
// creating multiple instances.
var timeZeroVal time.Time

// UsageFlag define flags that specify additional properties about the
// circumstances under which a command can be used.
type UsageFlag uint32

type Peer struct {
	ID         string
	RawAddress string
	PublicKey  string
	FirstPing  time.Time
	LastPing   time.Time
}

// rpcServer provides a concurrent safe RPC server to a bootnode server.
type RpcServer struct {
	Peers    map[string]*Peer // list peers which are still pinging to bootnode continuously
	peersMtx sync.Mutex
	server   *rpc.Server
	Config   RpcServerConfig // config for RPC server
}

type RpcServerConfig struct {
	Port int // rpc port
}

func (rpcServer *RpcServer) Init(config *RpcServerConfig) {
	// get config and init list Peers
	rpcServer.Config = *config
	rpcServer.Peers = make(map[string]*Peer)
	rpcServer.server = rpc.NewServer()
	// start go routin hertbeat to check invalid peers
	go rpcServer.PeerHeartBeat()
}

// Start - create handler and add into rpc server
// Listen and serve rpc server with config port
func (rpcServer *RpcServer) Start() {
	handler := &Handler{rpcServer}
	rpcServer.server.Register(handler)
	l, e := net.Listen("tcp", fmt.Sprintf(":%d", rpcServer.Config.Port))
	if e != nil {
		log.Fatal("listen error:", e)
	}
	rpcServer.server.Accept(l)
	l.Close()
}

// AddOrUpdatePeer - push a connected peer in to list of mem
func (rpcServer *RpcServer) AddOrUpdatePeer(rawAddress string, publicKeyB58 string, signDataB58 string) error {
	rpcServer.peersMtx.Lock()
	defer rpcServer.peersMtx.Unlock()
	if signDataB58 != "" && publicKeyB58 != "" && rawAddress != "" {
		err := cashec.ValidateDataB58(publicKeyB58, signDataB58, []byte(rawAddress))
		if err == nil {
			rpcServer.Peers[publicKeyB58] = &Peer{
				ID:         rpcServer.CombineID(rawAddress, publicKeyB58),
				RawAddress: rawAddress,
				PublicKey:  publicKeyB58,
				FirstPing:  time.Now().Local(),
				LastPing:   time.Now().Local(),
			}
		} else {
			log.Println("AddOrUpdatePeer error", err)
			return err
		}
	}
	return nil
}

func (rpcServer *RpcServer) RemovePeerByPbk(publicKey string) {
	delete(rpcServer.Peers, publicKey)
}

// CombineID - return string = rawAddress of peer + public key in base58check encode of node(run as committee)
// in case node is not running like a committee, we dont have public key of user who running node
// from this, we can check who is committee in network from bootnode if node provide data for bootnode about key
func (rpcServer *RpcServer) CombineID(rawAddress string, publicKey string) string {
	return rawAddress + publicKey
}

// PeerHeartBeat - loop forever after HeartbeatInterval to check peers
// which are not connected to remove from bootnode
// use Last Ping time to compare with time.now
func (rpcServer *RpcServer) PeerHeartBeat() {
	for {
		now := time.Now().Local()
		if len(rpcServer.Peers) > 0 {
		loop:
			for publicKey, peer := range rpcServer.Peers {
				if now.Sub(peer.LastPing).Seconds() > HeartbeatTimeout {
					rpcServer.RemovePeerByPbk(publicKey)
					goto loop
				}
			}
		}
		time.Sleep(HeartbeatInterval * time.Second)
	}
}
