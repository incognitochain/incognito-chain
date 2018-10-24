package addrmanager

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	peer2 "github.com/libp2p/go-libp2p-peer"
	"github.com/ninjadotorg/cash/peer"
)

const (
	// dumpAddressInterval is the interval used to dump the address
	// cache to disk for future use.
	dumpAddressInterval = time.Second * 10

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
	mtx        sync.Mutex
	peersFile  string
	lookupFunc func(string) ([]string, error)
	rand       *rand.Rand
	key        [32]byte
	started    int32
	shutdown   int32
	waitgroup  sync.WaitGroup
	quit       chan struct{}
	nTried     int
	nNew       int

	addrIndex map[string]*peer.Peer // address key to KnownAddress for all addrs.

	localAddresses map[string]*localAddress
}

type serializedKnownAddress struct {
	Addr        string
	Src         string
	PublicKey   string
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

func New(dataDir string, lookupFunc func(string) ([]string, error)) *AddrManager {
	addrManager := AddrManager{
		peersFile:      filepath.Join(dataDir, "peer.json"),
		lookupFunc:     lookupFunc,
		rand:           rand.New(rand.NewSource(time.Now().UnixNano())),
		quit:           make(chan struct{}),
		localAddresses: make(map[string]*localAddress),
		mtx:            sync.Mutex{},
	}
	addrManager.reset()
	return &addrManager
}

// savePeers saves all the known addresses to a file so they can be read back
// in at next run.
func (self *AddrManager) savePeers() {
	//self.mtx.Lock()
	//defer self.mtx.Unlock()

	if len(self.addrIndex) == 0 {
		return
	}

	sam := new(serializedAddrManager)
	sam.Version = 1
	copy(sam.Key[:], self.key[:])

	sam.Addresses = make([]*serializedKnownAddress, len(self.addrIndex))
	i := 0

	for k, v := range self.addrIndex {
		ska := new(serializedKnownAddress)
		ska.Addr = k
		ska.Src = v.PeerID.String()
		ska.PublicKey = v.PublicKey

		sam.Addresses[i] = ska
		i++
	}

	w, err := os.Create(self.peersFile)
	if err != nil {
		Logger.log.Infof("Error opening file %s: %+v", self.peersFile, err)
		return
	}
	enc := json.NewEncoder(w)
	defer w.Close()
	if err := enc.Encode(&sam); err != nil {
		Logger.log.Infof("Failed to encode file %s: %+v", self.peersFile, err)
		return
	}
}

// loadPeers loads the known address from the saved file.  If empty, missing, or
// malformed file, just don't load anything and start fresh
func (self *AddrManager) loadPeers() {
	//self.mtx.Lock()
	//defer self.mtx.Unlock()
	err := self.deserializePeers(self.peersFile)
	if err != nil {
		Logger.log.Infof("Failed to parse file %s: %+v", self.peersFile, err)
		// if it is invalid we nuke the old one unconditionally.
		err = os.Remove(self.peersFile)
		if err != nil {
			Logger.log.Infof("Failed to remove corrupt peers file %s: %+v",
				self.peersFile, err)
		}
		self.reset()
		return
	}
	Logger.log.Infof("Loaded %d addresses from file '%s'", self.numAddresses(), self.peersFile)
}

// NumAddresses returns the number of addresses known to the address manager.
func (a *AddrManager) numAddresses() int {
	//return a.nTried + a.nNew
	return len(a.addrIndex)
}

// reset resets the address manager by reinitialising the random source
// and allocating fresh empty bucket storage.
func (self *AddrManager) reset() {
	//self.mtx.Lock()
	//defer self.mtx.Unlock()

	self.addrIndex = make(map[string]*peer.Peer)
}

func (self *AddrManager) deserializePeers(filePath string) error {
	//self.mtx.Lock()
	//defer self.mtx.Unlock()

	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return nil
	}
	r, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("%s error opening file: %+v", filePath, err)
	}
	defer r.Close()

	var sam serializedAddrManager
	dec := json.NewDecoder(r)
	err = dec.Decode(&sam)
	if err != nil {
		return fmt.Errorf("error reading %s: %+v", filePath, err)
	}

	if sam.Version != 1 {
		return fmt.Errorf("unknown version %+v in serialized "+
			"addrmanager", sam.Version)
	}
	copy(self.key[:], sam.Key[:])

	for _, v := range sam.Addresses {
		peer := new(peer.Peer)
		peer.PeerID = peer2.ID(v.Src)
		peer.RawAddress = v.Addr
		peer.PublicKey = v.PublicKey

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

	Logger.log.Info("Starting address manager")

	// Load peers we already know about from file.
	self.loadPeers()
	// Start the address ticker to save addresses periodically.
	self.waitgroup.Add(1)
	go self.addressHandler()

}

// Stop gracefully shuts down the address manager by stopping the main handler.
func (self *AddrManager) Stop() error {
	if atomic.AddInt32(&self.shutdown, 1) != 1 {
		Logger.log.Infof("Address manager is already in the process of " +
			"shutting down")
		return nil
	}

	Logger.log.Infof("Address manager shutting down")
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
			self.savePeers()

		case <-self.quit:
			break out
		}
	}
	self.savePeers()
	self.waitgroup.Done()
	Logger.log.Infof("Address handler done")
}

// Good marks the given address as good.  To be called after a successful
// connection and version exchange.  If the address is unknown to the address
// manager it will be ignored.
func (self *AddrManager) Good(addr *peer.Peer) {
	self.mtx.Lock()
	defer self.mtx.Unlock()

	self.addrIndex[addr.RawAddress] = addr
}

func (self *AddrManager) AddAddresses(addr []*peer.Peer) {
	self.mtx.Lock()
	defer self.mtx.Unlock()

	for _, peer := range addr {
		self.addrIndex[peer.RawAddress] = peer
	}
}

func (self *AddrManager) AddAddressesStr(addrs []string) {
	self.mtx.Lock()
	defer self.mtx.Unlock()

	for _, addr := range addrs {
		peer := peer.Peer{
			RawAddress: addr,
		}
		self.addrIndex[addr] = &peer
	}
}

// AddressCache returns the current address cache.  It must be treated as
// read-only (but since it is a copy now, this is not as dangerous).
func (self *AddrManager) AddressCache() []*peer.Peer {
	self.mtx.Lock()
	defer self.mtx.Unlock()

	addrIndexLen := len(self.addrIndex)
	if addrIndexLen == 0 {
		return nil
	}
	allAddr := make([]*peer.Peer, 0, addrIndexLen)
	// Iteration order is undefined here, but we randomise it anyway.
	for _, v := range self.addrIndex {
		allAddr = append(allAddr, v)
	}
	return allAddr
}

func (self *AddrManager) ExistedAddr(addr string) bool {
	self.mtx.Lock()
	defer self.mtx.Unlock()

	_, ok := self.addrIndex[addr]
	return ok
}
