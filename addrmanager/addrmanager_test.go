package addrmanager

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/peer"
	"os"
	"testing"
)

var dataDir string
var addrManager *AddrManager
var _ = func() (_ struct{}) {
	fmt.Println("This runs before init()!")
	dataDir, _ = os.Getwd()
	addrManager = New(dataDir)
	Logger.Init(common.NewBackend(nil).Logger("test", true))
	return
}()

func TestAddrManager_New(t *testing.T) {
	if addrManager.peersFilePath != dataDir+"/peer.json" {
		t.Error("Wrong addrManager.peersFilePath")
	}
	if addrManager.cQuit == nil {
		t.Error("Wrong addrManager.cQuit")
	}
	if len(addrManager.addrIndex) > 0 {
		t.Error("Wrong addrManager.addrIndex")
	}
	if addrManager.started == 1 {
		t.Error("Wrong addrManager.started")
	}
	if addrManager.shutdown == 1 {
		t.Error("Wrong addrManager.shutdown")
	}
}

func TestAddrManager_Good(t *testing.T) {
	rawAddress := "localhost:9333"
	addr := peer.Peer{
		RawAddress: rawAddress,
	}
	addrManager.Good(&addr)
	if len(addrManager.addrIndex) == 0 {
		t.Error("Wrong addrManager.addrIndex")
	}
	if _, ok := addrManager.addrIndex[rawAddress]; !ok {
		t.Error("Wrong addrManager.addrIndex[0] with ", rawAddress)
	}
}

func TestAddrManager_Start(t *testing.T) {
	addrManager = New(dataDir)
	addrManager.Start()
	if addrManager.started != 1 {
		t.Error("Can not start")
	}
	addrManager.Stop()
}

func TestAddrManager_Stop(t *testing.T) {
	addrManager = New(dataDir)
	addrManager.Start()
	err := addrManager.Stop()
	if err != nil {
		t.Error(err)
	}
	if addrManager.shutdown != 1 {
		t.Error("Can not stop")
	}
}

func TestAddrManager_AddressCache(t *testing.T) {
	// init
	addrManager = New(dataDir)
	rawAddress := "localhost:9333"
	// get cache
	cached := addrManager.AddressCache()
	if len(cached) > 0 {
		t.Error("Cache should be empty")
	}
	// add peer into addr manager
	addr := peer.Peer{
		RawAddress: rawAddress,
	}
	addrManager.Good(&addr)
	cached = addrManager.AddressCache()
	if len(cached) == 0 {
		t.Error("Cache should be not empty")
	}
}

func TestNewAddrManager_SavePeer(t *testing.T) {
	addrManager = New(dataDir)
	rawAddress := "localhost:9333"
	// add peer into addr manager
	addr := peer.Peer{
		RawAddress: rawAddress,
	}
	addrManager.Good(&addr)
	err := addrManager.savePeers()

	defer func() { os.Remove(addrManager.peersFilePath) }()

	if err != nil {
		t.Error(err)
	} else {
		addrManager.reset()
		addrManager.loadPeers()
		cached := addrManager.AddressCache()
		if len(cached) == 0 {
			t.Error("Cache should be not empty")
		}
	}
}
