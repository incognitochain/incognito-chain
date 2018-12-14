package addrmanager

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	peer2 "github.com/libp2p/go-libp2p-peer"
	"github.com/ninjadotorg/constant/peer"
)

type AddrManager struct {
	mtx       sync.Mutex
	peersFile string
	key       [32]byte
	started   int32
	shutdown  int32
	waitGroup sync.WaitGroup

	cQuit chan struct{}

	addrIndex map[string]*peer.Peer // address key to KnownAddress for all addrs.
}

type serializedKnownAddress struct {
	Addr      string
	Src       string
	PublicKey string
}

type serializedAddrManager struct {
	Version   int
	Key       [32]byte
	Addresses []*serializedKnownAddress
}

func New(dataDir string) *AddrManager {
	addrManager := AddrManager{
		peersFile: filepath.Join(dataDir, "peer.json"),
		cQuit:     make(chan struct{}),
		mtx:       sync.Mutex{},
	}
	addrManager.reset()
	return &addrManager
}

// savePeers saves all the known addresses to a file so they can be read back
// in at next run.
func (self *AddrManager) savePeers() error {

	if len(self.addrIndex) == 0 {
		return nil
	}

	sam := new(serializedAddrManager)
	sam.Version = 1
	copy(sam.Key[:], self.key[:])

	sam.Addresses = make([]*serializedKnownAddress, len(self.addrIndex))
	i := 0

	for k, v := range self.addrIndex {
		ska := new(serializedKnownAddress)
		ska.Addr = k
		ska.Src = v.PeerID.Pretty()
		ska.PublicKey = v.PublicKey

		sam.Addresses[i] = ska
		i++
	}

	w, err := os.Create(self.peersFile)
	if err != nil {
		Logger.log.Errorf("Error opening file %s: %+v", self.peersFile, err)
		return NewAddrManagerError(UnexpectedError, err)
	}
	enc := json.NewEncoder(w)
	defer w.Close()
	if err := enc.Encode(&sam); err != nil {
		Logger.log.Errorf("Failed to encode file %s: %+v", self.peersFile, err)
		return NewAddrManagerError(UnexpectedError, err)
	}
	return nil
}

// loadPeers loads the known address from the saved file.  If empty, missing, or
// malformed file, just don't load anything and start fresh
func (self *AddrManager) loadPeers() {
	//self.mtx.Lock()
	//defer self.mtx.Unlock()
	err := self.deserializePeers(self.peersFile)
	if err != nil {
		Logger.log.Errorf("Failed to parse file %s: %+v", self.peersFile, err)
		// if it is invalid we nuke the old one unconditionally.
		err = os.Remove(self.peersFile)
		if err != nil {
			Logger.log.Errorf("Failed to remove corrupt peers file %s: %+v", self.peersFile, err)
		}
		self.reset()
	}
	Logger.log.Infof("Loaded %d addresses from file '%s'", self.numAddresses(), self.peersFile)
}

// NumAddresses returns the number of addresses known to the address manager.
func (self *AddrManager) numAddresses() int {
	//return a.nTried + a.nNew
	return len(self.addrIndex)
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

	if sam.Version != Version {
		return fmt.Errorf("unknown version %+v in serialized addrmanager", sam.Version)
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
	self.waitGroup.Add(1)
	go self.addressHandler()

}

// Stop gracefully shuts down the address manager by stopping the main handler.
func (self *AddrManager) Stop() error {
	if atomic.AddInt32(&self.shutdown, 1) != 1 {
		Logger.log.Errorf("Address manager is already in the process of shutting down")
		return nil
	}

	Logger.log.Infof("Address manager shutting down")
	close(self.cQuit)
	self.waitGroup.Wait()
	return nil
}

// addressHandler is the main handler for the address manager.  It must be run
// as a goroutine.
func (self *AddrManager) addressHandler() {
	dumpAddressTicker := time.NewTicker(DumpAddressInterval)
	defer dumpAddressTicker.Stop()
out:
	for {
		select {
		case <-dumpAddressTicker.C:
			self.savePeers()

		case <-self.cQuit:
			break out
		}
	}
	self.savePeers()
	self.waitGroup.Done()
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
