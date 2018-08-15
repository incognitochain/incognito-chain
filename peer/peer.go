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
)

const (
	LOCAL_HOST = "127.0.0.1"
)

type Peer struct {
	Host             host.Host
	ListeningAddress n.Addr
	Seed             int64
	FlagMutex        sync.Mutex

	Config Config
}

type Config struct {
	MessageListeners MessageListeners
}

type MessageListeners struct {
	OnTx    func(p *Peer, msg *wire.MessageTransaction)
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
	if net == "tcp4" {
		net = "ip4"
	} else {
		net = "ip6"
	}
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
	self.Host = basicHost
	return &self, nil
}

func (self Peer) Start() (error) {
	self.Host.SetStreamHandler("/peer/1.0.0", self.HandleStream)
	return nil
}

func (self Peer) HandleStream(s net.Stream) {
	log.Println("Got a new stream!")

	// Create a buffer stream for non blocking read and write.
	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))

	go self.InHandler(rw)
}

/**
Handle all in message
 */
func (self Peer) InHandler(rw *bufio.ReadWriter) {

	for {
		str, err := rw.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		if str == "" {
			return
		}
		if str != "\n" {
			var message wire.Message
			message.JsonDeserialize(str)

			switch msg := message.(type) {
			case *wire.MessageTransaction:
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
