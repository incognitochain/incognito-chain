package addrmanager

import (
	"sync"
	"github.com/ninjadotorg/cash-prototype/peer"
	"os"
	"encoding/json"
	"log"
	"fmt"
	peer2 "github.com/libp2p/go-libp2p-peer"
	"sync/atomic"
	"time"
)

const (
	// dumpAddressInterval is the interval used to dump the address
	// cache to disk for future use.
	dumpAddressInterval = time.Minute * 10

	// newBucketCount is the number of buckets that we spread new addresses
	// over.
	newBucketCount = 1024

	// triedBucketCount is the number of buckets we split tried
	// addresses over.
	triedBucketCount = 64
)

type localAddress struct {
	na    *peer.Peer
	score AddressPriority
}

// AddressPriority type is used to describe the hierarchy of local address
// discovery methods.
type AddressPriority int

type AddrManager struct {
	mtx       sync.Mutex
	peersFile string
	key       [32]byte
	started   int32
	shutdown  int32
	waitgroup sync.WaitGroup
	quit      chan struct{}
	nTried    int
	nNew      int

	addrIndex map[string]*peer.Peer // address key to ka for all addrs.

	localAddresses map[string]*localAddress
}

type serializedKnownAddress struct {
	Addr        string
	Src         string
	Attempts    int
	TimeStamp   int64
	LastAttempt int64
	LastSuccess int64
	// no refcount or tried, that is available from context.
}

type serializedAddrManager struct {
	Version      int
	Key          [32]byte
	Addresses    []*serializedKnownAddress
	NewBuckets   [newBucketCount][]string // string is NetAddressKey
	TriedBuckets [triedBucketCount][]string
}

// savePeers saves all the known addresses to a file so they can be read back
// in at next run.
func (self *AddrManager) SavePeers() {
	self.mtx.Lock()
	defer self.mtx.Unlock()

	sam := new(serializedAddrManager)
	sam.Version = 1
	copy(sam.Key[:], self.key[:])

	sam.Addresses = make([]*serializedKnownAddress, len(self.addrIndex))
	i := 0

	for k, v := range self.addrIndex {
		ska := new(serializedKnownAddress)
		ska.Addr = k
		ska.Src = v.PeerId.String()

		sam.Addresses[i] = ska
		i++
	}

	w, err := os.Create(self.peersFile)
	if err != nil {
		log.Printf("Error opening file %s: %v", self.peersFile, err)
		return
	}
	enc := json.NewEncoder(w)
	defer w.Close()
	if err := enc.Encode(&sam); err != nil {
		log.Printf("Failed to encode file %s: %v", self.peersFile, err)
		return
	}
}

// loadPeers loads the known address from the saved file.  If empty, missing, or
// malformed file, just don't load anything and start fresh
func (self *AddrManager) loadPeers() {
	self.mtx.Lock()
	defer self.mtx.Unlock()

	err := self.deserializePeers(self.peersFile)
	if err != nil {
		log.Printf("Failed to parse file %s: %v", self.peersFile, err)
		// if it is invalid we nuke the old one unconditionally.
		err = os.Remove(self.peersFile)
		if err != nil {
			log.Printf("Failed to remove corrupt peers file %s: %v",
				self.peersFile, err)
		}
		self.reset()
		return
	}
	log.Printf("Loaded %d addresses from file '%s'", self.numAddresses(), self.peersFile)
}

// NumAddresses returns the number of addresses known to the address manager.
func (a *AddrManager) numAddresses() int {
	return a.nTried + a.nNew
}

// reset resets the address manager by reinitialising the random source
// and allocating fresh empty bucket storage.
func (self *AddrManager) reset() {
	self.addrIndex = make(map[string]*peer.Peer)
}

func (self *AddrManager) deserializePeers(filePath string) error {

	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return nil
	}
	r, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("%s error opening file: %v", filePath, err)
	}
	defer r.Close()

	var sam serializedAddrManager
	dec := json.NewDecoder(r)
	err = dec.Decode(&sam)
	if err != nil {
		return fmt.Errorf("error reading %s: %v", filePath, err)
	}

	if sam.Version != 1 {
		return fmt.Errorf("unknown version %v in serialized "+
			"addrmanager", sam.Version)
	}
	copy(self.key[:], sam.Key[:])

	for _, v := range sam.Addresses {
		peer := new(peer.Peer)
		peer.PeerId = peer2.ID(v.Src)
		peer.RawAddress = v.Addr
		self.addrIndex[peer.RawAddress] = peer
	}
	return nil
}

// Start begins the core address handler which manages a pool of known
// addresses, timeouts, and interval based writes.
func (self *AddrManager) Start() {
	// Already started?
	if atomic.AddInt32(&self.started, 1) != 1 {
		return
	}

	log.Printf("Starting address manager")

	// Load peers we already know about from file.
	self.loadPeers()
	// Start the address ticker to save addresses periodically.
	self.waitgroup.Add(1)
	go self.addressHandler()

}

// Stop gracefully shuts down the address manager by stopping the main handler.
func (self *AddrManager) Stop() error {
	if atomic.AddInt32(&self.shutdown, 1) != 1 {
		log.Printf("Address manager is already in the process of " +
			"shutting down")
		return nil
	}

	log.Printf("Address manager shutting down")
	close(self.quit)
	self.waitgroup.Wait()
	return nil
}

// addressHandler is the main handler for the address manager.  It must be run
// as a goroutine.
func (self *AddrManager) addressHandler() {
	dumpAddressTicker := time.NewTicker(dumpAddressInterval)
	defer dumpAddressTicker.Stop()
out:
	for {
		select {
		case <-dumpAddressTicker.C:
			self.SavePeers()

		case <-self.quit:
			break out
		}
	}
	self.SavePeers()
	self.waitgroup.Done()
	log.Printf("Address handler done")
}
