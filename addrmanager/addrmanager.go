package addrmanager

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/incognitochain/incognito-chain/peer"
	peer2 "github.com/libp2p/go-libp2p-peer"
)

type AddrManager struct {
	mtx           sync.Mutex
	peersFilePath string
	key           [32]byte
	started       int32
	shutdown      int32
	waitGroup     sync.WaitGroup

	cQuit chan struct{}

	addrIndex map[string]*peer.Peer
}

// data structure of address which need to be saving in file
type serializedKnownAddress struct {
	Addr      string `json:"Addr"`
	Src       string `json:"Src"`
	PublicKey string `json:"PublicKey"`
}

// data structure of list address which need to be saving in file
type serializedAddrManager struct {
	Version   int                       `json:"Version"`
	Key       [32]byte                  `json:"Key"`
	Addresses []*serializedKnownAddress `json:"Addresses"`
}

func New(dataDir string) *AddrManager {
	addrManager := AddrManager{
		peersFilePath: filepath.Join(dataDir, "peer.json"), // path to file which is used for storing information in add manager
		cQuit:         make(chan struct{}),
		mtx:           sync.Mutex{},
	}
	addrManager.reset()
	return &addrManager
}

// savePeers saves all the known addresses to a file so they can be read back
// in at next run.
func (addrManager *AddrManager) savePeers() error {

	if len(addrManager.addrIndex) == 0 {
		return nil
	}

	storageData := new(serializedAddrManager)
	storageData.Version = 1
	copy(storageData.Key[:], addrManager.key[:])

	storageData.Addresses = make([]*serializedKnownAddress, len(addrManager.addrIndex))
	i := 0

	// get all good address in list of addresses manager
	for rawAddress, peerObj := range addrManager.addrIndex {
		// init address data to push into storage data
		addressData := new(serializedKnownAddress)
		addressData.Addr = rawAddress
		Logger.log.Info("PeerID", peerObj.PeerID.String(), len(peerObj.PeerID.String()))
		addressData.Src = peerObj.PeerID.Pretty()
		addressData.PublicKey = peerObj.PublicKey

		// push into array
		storageData.Addresses[i] = addressData
		i++
	}

	// Create file with file path
	writerFile, err := os.Create(addrManager.peersFilePath)
	if err != nil {
		Logger.log.Errorf("Error opening file %s: %+peerObj", addrManager.peersFilePath, err)
		return NewAddrManagerError(UnexpectedError, err)
	}
	// encode to json format
	enc := json.NewEncoder(writerFile)
	defer writerFile.Close()
	// write into file with json format
	if err := enc.Encode(&storageData); err != nil {
		Logger.log.Errorf("Failed to encode file %s: %+v", addrManager.peersFilePath, err)
		return NewAddrManagerError(UnexpectedError, err)
	}
	return nil
}

// loadPeers loads the known address from the saved file.  If empty, missing, or
// malformed file, just don't load anything and start fresh
func (addrManager *AddrManager) loadPeers() {
	err := addrManager.deserializePeers(addrManager.peersFilePath)
	if err != nil {
		Logger.log.Errorf("Failed to parse file %s: %+v", addrManager.peersFilePath, err)
		// if it is invalid we nuke the old one unconditionally.
		err = os.Remove(addrManager.peersFilePath)
		if err != nil {
			Logger.log.Errorf("Failed to remove corrupt peers file %s: %+v", addrManager.peersFilePath, err)
		}
		addrManager.reset()
	}
	Logger.log.Infof("Loaded %d addresses from file '%s'", addrManager.numAddresses(), addrManager.peersFilePath)
}

// NumAddresses returns the number of addresses known to the address manager.
func (addrManager *AddrManager) numAddresses() int {
	return len(addrManager.addrIndex)
}

// reset resets the address manager by reinitialising the random source
// and allocating fresh empty bucket storage.
func (addrManager *AddrManager) reset() {
	addrManager.addrIndex = make(map[string]*peer.Peer)
}

// deserializePeers - read storage data about Addresses manager and restore a object for it
// data are read from filePath which be used when saving
func (addrManager *AddrManager) deserializePeers(filePath string) error {
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

	if sam.Version != version {
		return fmt.Errorf("unknown version %+v in serialized addrmanager", sam.Version)
	}
	copy(addrManager.key[:], sam.Key[:])

	for _, v := range sam.Addresses {
		if len(v.Src) > 10000 {
			continue
		}
		peer := new(peer.Peer)
		peer.PeerID = peer2.ID(v.Src)
		peer.RawAddress = v.Addr
		peer.PublicKey = v.PublicKey

		addrManager.addrIndex[peer.RawAddress] = peer

	}
	return nil
}

// Start begins the core address handler which manages a pool of known
// addresses, timeouts, and interval based writes.
func (addrManager *AddrManager) Start() {
	// Already started?
	if atomic.AddInt32(&addrManager.started, 1) != 1 {
		return
	}

	Logger.log.Info("Starting address manager")

	// Load peers we already know about from file.
	addrManager.loadPeers()
	// Start the address ticker to save addresses periodically.
	addrManager.waitGroup.Add(1)
	go addrManager.addressHandler()

}

// Stop gracefully shuts down the address manager by stopping the main handler.
func (addrManager *AddrManager) Stop() error {
	if atomic.AddInt32(&addrManager.shutdown, 1) != 1 {
		Logger.log.Errorf("Address manager is already in the process of shutting down")
		return nil
	}

	Logger.log.Infof("Address manager shutting down")
	// close channel to break loop select channel
	close(addrManager.cQuit)
	addrManager.waitGroup.Wait()
	return nil
}

// addressHandler is the main handler for the address manager.  It must be run
// as a goroutine.
func (addrManager *AddrManager) addressHandler() {
	dumpAddressTicker := time.NewTicker(DumpAddressInterval)
	defer dumpAddressTicker.Stop()
out:
	for {
		select {
		case <-dumpAddressTicker.C:
			addrManager.savePeers()

		case <-addrManager.cQuit:
			// break out loop
			break out
		}
	}
	// saving before done everything
	addrManager.savePeers()
	// Done wait group
	addrManager.waitGroup.Done()
	// Log to notice
	Logger.log.Infof("Address handler done")
}

// Good marks the given address as good.  To be called after a successful
// connection and version exchange.  If the address is unknown to the address
// manager it will be ignored.
func (addrManager *AddrManager) Good(addr *peer.Peer) {
	addrManager.mtx.Lock()
	defer addrManager.mtx.Unlock()

	addrManager.addrIndex[addr.RawAddress] = addr
}

// AddressCache returns the current address cache.  It must be treated as
// read-only (but since it is a copy now, this is not as dangerous).
func (addrManager *AddrManager) AddressCache() []*peer.Peer {
	addrManager.mtx.Lock()
	defer addrManager.mtx.Unlock()

	addrIndexLen := len(addrManager.addrIndex)
	if addrIndexLen == 0 {
		return nil
	}
	allAddr := make([]*peer.Peer, 0, addrIndexLen)
	// Iteration order is undefined here, but we randomise it anyway.
	for _, v := range addrManager.addrIndex {
		allAddr = append(allAddr, v)
	}
	return allAddr
}
