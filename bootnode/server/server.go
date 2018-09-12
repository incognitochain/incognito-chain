package server

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
	"sort"
	"time"
)

const (
	heartbeatInterval = 5
	heartbeatTimeout = 60
)

// timeZeroVal is simply the zero value for a time.Time and is used to avoid
// creating multiple instances.
var timeZeroVal time.Time

// UsageFlag define flags that specify additional properties about the
// circumstances under which a command can be used.
type UsageFlag uint32

type Peer struct {
	ID string
	RawAddress string
	PublicKey string
	FirstPing time.Time
	LastPing time.Time
}

// rpcServer provides a concurrent safe RPC server to a chain server.
type RpcServer struct {
	Peers []*Peer

	Config RpcServerConfig
}

type RpcServerConfig struct {
	Port int
}

func (self *RpcServer) Init(config *RpcServerConfig) (error) {
	self.Config = *config
	self.Peers = make([]*Peer, 0)
	go self.PeerHeartBeat()
	return nil
}

func (self *RpcServer) Start() {
	handler := &Handler{self}
	server := rpc.NewServer()
	server.Register(handler)
	l, e := net.Listen("tcp", fmt.Sprintf(":%d", self.Config.Port))
	if e != nil {
		log.Fatal("listen error:", e)
	}
	server.Accept(l)
}

func (self *RpcServer) AddOrUpdatePeer(rawAddress string, publicKey string) {
	exist := false
	for _, peer := range self.Peers {
		if self.CombineID(rawAddress, publicKey) == peer.ID {
			exist = true
			peer.LastPing = time.Now().Local()
		}
	}

	if !exist {
		self.Peers = append(self.Peers, &Peer{self.CombineID(rawAddress, publicKey), rawAddress, publicKey,time.Now().Local(), time.Now().Local()})
		sort.Slice(self.Peers, func(i, j int) bool {
			return self.Peers[i].FirstPing.Sub(self.Peers[j].FirstPing) <= 0
		})
	}
}

func (self *RpcServer) RemovePeer(ID string) {
	removeIdx := -1
	for idx, peer := range self.Peers {
		if peer.ID == ID {
			removeIdx = idx
		}
	}

	if removeIdx != -1 {
		self.RemovePeerByIdx(removeIdx)
	}
}

func (self *RpcServer) RemovePeerByIdx(idx int) {
	self.Peers = append(self.Peers[:idx], self.Peers[idx+1:]...)
}

func (self *RpcServer) CombineID(rawAddress string, publicKey string) string {
	return rawAddress + publicKey
}

func (self *RpcServer) PeerHeartBeat() {
	for {
		now := time.Now().Local()
		if len(self.Peers) > 0 {
		loop:
			for idx, peer := range self.Peers {
				if now.Sub(peer.LastPing).Seconds() > heartbeatTimeout {
					self.RemovePeerByIdx(idx)
					goto loop
				}
			}
		}
		time.Sleep(heartbeatInterval * time.Second)
	}
}
