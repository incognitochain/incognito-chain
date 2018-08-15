package peer

import (
	"github.com/libp2p/go-libp2p-host"
	"github.com/libp2p/go-libp2p"
	"log"
	"io"
	"crypto/rand"
	mrand "math/rand"
	"github.com/libp2p/go-libp2p-crypto"
	"fmt"
	"context"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/libp2p/go-libp2p-net"
	"bufio"
	"sync"
	"github.com/internet-cash/prototype/wire"
	"github.com/davecgh/go-spew/spew"
	"strings"
	n "net"
	"github.com/libp2p/go-libp2p-peer"
	"sync/atomic"
)

const (
	LOCAL_HOST = "127.0.0.1"
)

type Peer struct {
	connected  int32
	disconnect int32

	Host             host.Host
	Multiaddr        ma.Multiaddr
	PeerId           peer.ID
	ListeningAddress n.Addr
	Seed             int64
	FlagMutex        sync.Mutex

	Config Config

	quit chan struct{}
}

type Config struct {
	MessageListeners MessageListeners
}

type MessageListeners struct {
	OnTx    func(p *Peer, msg *wire.MessageTx)
	OnBlock func(p *Peer, msg *wire.MessageBlock)
}

func (self Peer) NewPeer() (*Peer, error) {
	// If the seed is zero, use real cryptographic randomness. Otherwise, use a
	// deterministic randomness source to make generated keys stay the same
	// across multiple runs
	var r io.Reader
	if self.Seed == 0 {
		r = rand.Reader
	} else {
		r = mrand.New(mrand.NewSource(self.Seed))
	}

	// Generate a key pair for this Host. We will use it
	// to obtain a valid Host ID.
	priv, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r)
	if err != nil {
		return &self, err
	}

	ip := strings.Split(self.ListeningAddress.String(), ":")[0]
	if len(ip) == 0 {
		ip = LOCAL_HOST
	}
	port := strings.Split(self.ListeningAddress.String(), ":")[1]
	net := self.ListeningAddress.Network()
	listeningAddressString := fmt.Sprintf("/%s/%s/tcp/%s", net, ip, port)
	opts := []libp2p.Option{
		libp2p.ListenAddrStrings(listeningAddressString),
		libp2p.Identity(priv),
	}

	basicHost, err := libp2p.New(context.Background(), opts...)
	if err != nil {
		return &self, err
	}

	// Build Host multiaddress
	hostAddr, _ := ma.NewMultiaddr(fmt.Sprintf("/ipfs/%s", basicHost.ID().Pretty()))

	// Now we can build a full multiaddress to reach this Host
	// by encapsulating both addresses:
	addr := basicHost.Addrs()[0]
	fullAddr := addr.Encapsulate(hostAddr)
	log.Printf("I am listening on %s\n", fullAddr)
	pid, err := fullAddr.ValueForProtocol(ma.P_IPFS)
	if err != nil {
		log.Print(err)
		return &self, err
	}
	peerid, err := peer.IDB58Decode(pid)
	if err != nil {
		log.Print(err)
		return &self, err
	}

	self.Host = basicHost
	self.Multiaddr = fullAddr
	self.PeerId = peerid

	return &self, nil
}

func (self Peer) Start() (error) {
	self.Host.SetStreamHandler("/peer/1.0.0", self.HandleStream)
	// Hang forever
	<-make(chan struct{})
	return nil
}

// WaitForDisconnect waits until the peer has completely disconnected and all
// resources are cleaned up.  This will happen if either the local or remote
// side has been disconnected or the peer is forcibly disconnected via
// Disconnect.
func (p Peer) WaitForDisconnect() {
	<-p.quit
}

func (self Peer) HandleStream(s net.Stream) {
	// Remember to close the stream when we are done.
	defer s.Close()

	log.Println("Got a new stream!")

	// Create a buffer stream for non blocking read and write.
	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))

	go self.InHandler(rw)
	go self.InHandler(rw)
}

/**
Handle all in message
 */
func (self Peer) InHandler(rw *bufio.ReadWriter) {
	for {
		str, err := rw.ReadString('\n')
		if err != nil {
			log.Print(err)
			return
		}

		if str == "" {
			return
		}
		if str != "\n" {
			var message wire.Message
			message.JsonDeserialize(str)

			switch msg := message.(type) {
			case *wire.MessageTx:
				if self.Config.MessageListeners.OnTx != nil {
					self.FlagMutex.Lock()
					self.Config.MessageListeners.OnTx(&self, msg)
					self.FlagMutex.Unlock()
				}
			case *wire.MessageBlock:
				if self.Config.MessageListeners.OnBlock != nil {
					self.FlagMutex.Lock()
					self.Config.MessageListeners.OnBlock(&self, msg)
					self.FlagMutex.Unlock()
				}
			default:
				log.Printf("Received unhandled message of type %v "+
					"from %v", msg.MessageType(), self)
				spew.Dump(msg)
			}
		}
	}
}

// Disconnect disconnects the peer by closing the connection.  Calling this
// function when the peer is already disconnected or in the process of
// disconnecting will have no effect.
func (self Peer) Disconnect() {
	if atomic.AddInt32(&self.disconnect, 1) != 1 {
		return
	}

	log.Printf("Disconnecting %s", self)
	if atomic.LoadInt32(&self.connected) != 0 {
		self.Host.Close()
	}
	if self.quit != nil {
		close(self.quit)
	}
	self.disconnect = 1
}
