package addrmanager

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/peer"
	peer2 "github.com/libp2p/go-libp2p-peer"
	"github.com/pkg/errors"
)

type AddrManager struct {
	mtx           sync.Mutex
	peersFilePath string
	key           common.Hash
	started       int32
	shutdown      int32
	waitGroup     sync.WaitGroup

	cQuit chan struct{}

	addrIndex map[string]*peer.Peer
}

// data structure of address which need to be saving in file
type serializedKnownAddress struct {
	Addr          string `json:"Addr"`
	Src           string `json:"Src"`
	PublicKey     string `json:"PublicKey"`
	PublicKeyType string `json:"PublicKeyType"`
}

// data structure of list address which need to be saving in file
type serializedAddrManager struct {
	Version   int                       `json:"Version"`
	Key       common.Hash               `json:"Key"`
	Addresses []*serializedKnownAddress `json:"Addresses"`
}

// NewAddrManager - init a AddrManager object,
// set config and return pointer to object
func NewAddrManager(dataDir string, key common.Hash) *AddrManager {
	addrManager := AddrManager{
		peersFilePath: filepath.Join(dataDir, dataFile), // path to file which is used for storing information in add manager
		cQuit:         make(chan struct{}),
		mtx:           sync.Mutex{},
		key:           key,
	}
	addrManager.reset()
	return &addrManager
}

// savePeers saves all the known addresses to a file so they can be read back
// in at next run.
func (addrManager *AddrManager) savePeers() error {
	addrManager.mtx.Lock()
	defer addrManager.mtx.Unlock()

	// check len of addrIndex
	if len(addrManager.addrIndex) == 0 {
		// dont have anything to save into file data
		return nil
	}

	storageData := new(serializedAddrManager)
	storageData.Version = version
	copy(storageData.Key[:], addrManager.key[:])

	storageData.Addresses = []*serializedKnownAddress{}

	// get all good address in list of addresses manager
	for rawAddress, peerObj := range addrManager.addrIndex {
		peerID := peerObj.GetPeerID()
		pretty := peerID.Pretty()
		if len(pretty) > maxLengthPeerPretty {
			continue
		}
		// init address data to push into storage data
		addressData := new(serializedKnownAddress)
		addressData.Addr = rawAddress
		Logger.log.Debug("PeerID", peerID.String(), len(peerID.String()))
		addressData.Src = pretty
		addressData.PublicKey, addressData.PublicKeyType = peerObj.GetPublicKey()

		// push into array
		storageData.Addresses = append(storageData.Addresses, addressData)
	}

	// Create file with file path
	writerFile, err := os.Create(addrManager.peersFilePath)
	if err != nil {
		Logger.log.Errorf("Error opening file %s: %+peerObj", addrManager.peersFilePath, err)
		return NewAddrManagerError(CreateDataFileError, err)
	}
	// encode to json format
	enc := json.NewEncoder(writerFile)
	defer writerFile.Close()
	// write into file with json format
	if err := enc.Encode(&storageData); err != nil {
		Logger.log.Errorf("Failed to encode file %s: %+v", addrManager.peersFilePath, err)
		return NewAddrManagerError(EncodeDataFileError, err)
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
// data are read from dataFile which be used when saving
func (addrManager *AddrManager) deserializePeers(filePath string) error {
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return nil
	}
	r, err := os.Open(filePath)
	if err != nil {
		return NewAddrManagerError(OpenDataFileError, errors.New(fmt.Sprintf("%s error opening file: %+v", filePath, err)))
	}
	defer r.Close()

	var storageData serializedAddrManager
	dec := json.NewDecoder(r)
	err = dec.Decode(&storageData)
	if err != nil {
		return NewAddrManagerError(DecodeDataFileError, errors.New(fmt.Sprintf("error reading %s: %+v", filePath, err)))
	}

	if storageData.Version != version {
		return NewAddrManagerError(WrongVersionError, errors.New(fmt.Sprintf("unknown Version %+v in serialized addrmanager", storageData.Version)))
	}
	copy(addrManager.key[:], storageData.Key[:])

	for _, storagePeer := range storageData.Addresses {
		if len(storagePeer.Src) > maxLengthPeerPretty {
			continue
		}
		peer := new(peer.Peer)
		peer.SetPeerID(peer2.ID(storagePeer.Src))
		peer.SetRawAddress(storagePeer.Addr)
		peer.SetPublicKey(storagePeer.PublicKey, storagePeer.PublicKeyType)

		addrManager.addrIndex[peer.GetRawAddress()] = peer

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
	addrManager.shutdown = 0
	// Load peers we already know about from file.
	addrManager.loadPeers()
	// Start the address ticker to save addresses periodically.
	addrManager.waitGroup.Add(1)
	go addrManager.addressHandler()

}

// Stop gracefully shuts down the address manager by stopping the main handler.
func (addrManager *AddrManager) Stop() error {
	if atomic.AddInt32(&addrManager.shutdown, 1) != 1 {
		errStr := fmt.Sprint("Address manager is already in the process of shutting down")
		Logger.log.Error(errStr)
		return NewAddrManagerError(StopError, errors.New(errStr))
	}
	addrManager.started = 0
	Logger.log.Infof("Address manager shutting down")
	// close channel to break loop select channel
	close(addrManager.cQuit)
	addrManager.waitGroup.Wait()
	return nil
}

// addressHandler is the main handler for the address manager.  It must be run
// as a goroutine.
func (addrManager *AddrManager) addressHandler() {
	dumpAddressTicker := time.NewTicker(dumpAddressInterval)
	defer dumpAddressTicker.Stop()
out:
	for {
		select {
		case <-dumpAddressTicker.C:
			err := addrManager.savePeers()
			if err != nil {
				Logger.log.Error(err)
			}

		case <-addrManager.cQuit:
			// break out loop
			break out
		}
	}
	// saving before done everything
	err := addrManager.savePeers()
	if err != nil {
		Logger.log.Error(err)
	}
	// Done wait group
	addrManager.waitGroup.Done()
	// Log to notice
	Logger.log.Infof("Address handler done")
}

// Good marks the given address as good.  To be called after a successful
// connection and Version exchange.  If the address is unknown to the address
// manager it will be ignored.
func (addrManager *AddrManager) Good(addr *peer.Peer) {
	addrManager.mtx.Lock()
	defer addrManager.mtx.Unlock()

	addrManager.addrIndex[addr.GetRawAddress()] = addr
}

// AddressCache returns the current address cache.  It must be treated as
// read-only (but since it is a copy now, this is not as dangerous).
func (addrManager *AddrManager) AddressCache() []*peer.Peer {
	addrManager.mtx.Lock()
	defer addrManager.mtx.Unlock()

	addrIndexLen := len(addrManager.addrIndex)
	if addrIndexLen == 0 {
		Logger.log.Debug("Address is empty")
		return nil
	}
	allAddr := make([]*peer.Peer, 0, addrIndexLen)
	// Iteration order is undefined here, but we randomise it anyway.
	for _, index := range addrManager.addrIndex {
		allAddr = append(allAddr, index)
	}
	return allAddr
}
