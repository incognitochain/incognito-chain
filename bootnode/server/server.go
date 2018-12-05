package server

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
	"time"
	"github.com/ninjadotorg/constant/cashec"
)

const (
	heartbeatInterval = 5
	heartbeatTimeout  = 60
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

// rpcServer provides a concurrent safe RPC server to a chain server.
type RpcServer struct {
	Peers map[string]*Peer

	Config RpcServerConfig
}

type RpcServerConfig struct {
	Port int
}

func (self *RpcServer) Init(config *RpcServerConfig) (error) {
	self.Config = *config
	self.Peers = make(map[string]*Peer)
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

func (self *RpcServer) AddOrUpdatePeer(rawAddress string, publicKeyB58 string, signDataB58 string) {
	if signDataB58 != "" && publicKeyB58 != "" && rawAddress != "" {
		err := cashec.ValidateDataB58(publicKeyB58, signDataB58, []byte{0x00})
		if err == nil {
			self.Peers[publicKeyB58] = &Peer{
				ID:         self.CombineID(rawAddress, publicKeyB58),
				RawAddress: rawAddress,
				PublicKey:  publicKeyB58,
				FirstPing:  time.Now().Local(),
				LastPing:   time.Now().Local(),
			}
		} else {
			log.Println("AddOrUpdatePeer error", err)
		}
	}
}

func (self *RpcServer) RemovePeerByPbk(publicKey string) {
	delete(self.Peers, publicKey)
}

func (self *RpcServer) CombineID(rawAddress string, publicKey string) string {
	return rawAddress + publicKey
}

func (self *RpcServer) PeerHeartBeat() {
	for {
		now := time.Now().Local()
		if len(self.Peers) > 0 {
		loop:
			for publicKey, peer := range self.Peers {
				if now.Sub(peer.LastPing).Seconds() > heartbeatTimeout {
					self.RemovePeerByPbk(publicKey)
					goto loop
				}
			}
		}
		time.Sleep(heartbeatInterval * time.Second)
	}
}
