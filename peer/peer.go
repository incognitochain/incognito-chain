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
)

type Peer struct {
	Host          host.Host
	ListeningPort int
	Seed          int64
	flagMutex     sync.Mutex
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

	opts := []libp2p.Option{
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", self.ListeningPort)),
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
	log.Printf("I am %s\n", fullAddr)
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
			chain := make([]Block, 0)
			if err := json.Unmarshal([]byte(str), &chain); err != nil {
				log.Fatal(err)
			}

			self.flagMutex.Lock()
			if len(chain) > len(Blockchain) {
				Blockchain = chain
				bytes, err := json.MarshalIndent(Blockchain, "", "  ")
				if err != nil {

					log.Fatal(err)
				}
				// Green console color: 	\x1b[32m
				// Reset console color: 	\x1b[0m
				fmt.Printf("\x1b[32m%s\x1b[0m> ", string(bytes))
			}
			self.flagMutex.Unlock()
		}
	}
}
