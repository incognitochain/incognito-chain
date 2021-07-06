package peerv2

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/incognitochain/incognito-chain/peerv2/mocks"
	"github.com/incognitochain/incognito-chain/peerv2/rpcclient"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p/p2p/protocol/ping"
	cache "github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	swarmt "github.com/libp2p/go-libp2p-swarm/testing"
	bhost "github.com/libp2p/go-libp2p/p2p/host/basic"
)

// Checks if ignored addresses are skipped when choosing one to connect
// We ignore 1 address out of 2, randomly choose 100 times and make sure all runs return the same one
func TestChooseHighwayFilterIgnore(t *testing.T) {
	hwAddrs := []rpcclient.HighwayAddr{
		rpcclient.HighwayAddr{Libp2pAddr: testHighwayAddress},
		rpcclient.HighwayAddr{Libp2pAddr: testHighwayAddress2},
	}

	keeper := NewAddrKeeper()
	keeper.addrs = hwAddrs
	keeper.IgnoreAddress(hwAddrs[1])

	pid := peer.ID("")
	for i := 0; i < 100; i++ {
		chosen, err := keeper.chooseHighwayFromList(pid)
		assert.Nil(t, err)
		assert.Equal(t, hwAddrs[0], chosen)
	}
}

// Checks the edge case when all addresses are ignored when choosing 1 to connect.
// In this case, we randomly pick one from the full list.
// This test ignores all addresses in the list and runs once to make sure an address is returned.
func TestChooseHighwayIgnoredAll(t *testing.T) {
	hwAddrs := []rpcclient.HighwayAddr{
		rpcclient.HighwayAddr{Libp2pAddr: testHighwayAddress},
		rpcclient.HighwayAddr{Libp2pAddr: testHighwayAddress2},
	}

	keeper := NewAddrKeeper()
	keeper.addrs = hwAddrs
	keeper.IgnoreAddress(hwAddrs[0])
	keeper.IgnoreAddress(hwAddrs[1])

	pid := peer.ID("")
	chosen, err := keeper.chooseHighwayFromList(pid)
	assert.Nil(t, err)
	assert.NotNil(t, chosen)
}

// Makes sure ignored addresses had their timing reset when the new list doesn't contain them
func TestResetRPCIgnoreTiming(t *testing.T) {
	hwAddrs := []rpcclient.HighwayAddr{
		rpcclient.HighwayAddr{Libp2pAddr: testHighwayAddress},
		rpcclient.HighwayAddr{Libp2pAddr: testHighwayAddress2},
	}
	keeper := NewAddrKeeper()
	keeper.addrs = hwAddrs[:1]
	keeper.IgnoreAddress(hwAddrs[0])

	keeper.updateAddrs(hwAddrs[1:])
	assert.Equal(t, 0, len(keeper.ignoreRPCUntil))
	assert.Equal(t, 0, len(keeper.ignoreHWUntil))
}

// Makes sure an address is ignored when the RPC call to it failed
func TestRPCIgnoreWhenFail(t *testing.T) {
	var resultAddrs map[string][]rpcclient.HighwayAddr
	discoverer := &mocks.HighwayDiscoverer{}
	discoverer.On("DiscoverHighway", mock.Anything, mock.Anything).Return(resultAddrs, fmt.Errorf("dummy"))

	hwAddrs := []rpcclient.HighwayAddr{
		rpcclient.HighwayAddr{Libp2pAddr: testHighwayAddress},
		rpcclient.HighwayAddr{Libp2pAddr: testHighwayAddress2},
	}
	keeper := NewAddrKeeper()
	keeper.addrs = hwAddrs

	_, err := keeper.getHighwayAddrs(discoverer)
	assert.NotNil(t, err)
	assert.Equal(t, 1, len(keeper.ignoreRPCUntil))
	assert.Equal(t, 0, len(keeper.ignoreHWUntil))
}

// Makes sure we choose a random address to RPC when all addresses are ignored.
// To do this, we ignore all addresses and run once to make sure one is returned.
func TestRPCIgnoredAll(t *testing.T) {
	resultAddrs := map[string][]rpcclient.HighwayAddr{}
	discoverer := &mocks.HighwayDiscoverer{}
	discoverer.On("DiscoverHighway", mock.Anything, mock.Anything).Return(resultAddrs, nil)

	hwAddrs := []rpcclient.HighwayAddr{
		rpcclient.HighwayAddr{Libp2pAddr: testHighwayAddress},
		rpcclient.HighwayAddr{Libp2pAddr: testHighwayAddress2},
	}
	keeper := NewAddrKeeper()
	keeper.addrs = hwAddrs
	keeper.IgnoreAddress(hwAddrs[0])
	keeper.IgnoreAddress(hwAddrs[1])

	_, err := keeper.getHighwayAddrs(discoverer)
	assert.Nil(t, err)
}

func TestChooseHighwayFiltered(t *testing.T) {
	hwAddrs := []rpcclient.HighwayAddr{
		rpcclient.HighwayAddr{Libp2pAddr: testHighwayAddress},
		rpcclient.HighwayAddr{Libp2pAddr: ""},
	}
	keeper := NewAddrKeeper()
	keeper.addrs = hwAddrs

	pid := peer.ID("")
	_, err := keeper.chooseHighwayFromList(pid)
	assert.Nil(t, err)
}

func TestChooseHighwayFromSortedList(t *testing.T) {
	addr1 := "/ip4/0.0.0.0/tcp/7337/p2p/QmSPa4gxx6PRmoNRu6P2iFwEwmayaoLdR5By3i3MgM9gMv"
	addr2 := "/ip4/0.0.0.1/tcp/7337/p2p/QmSPa4gxx6PRmoNRu6P2iFwEwmayaoLdR5By3i3MgM9gMv"
	addr3 := "/ip4/0.0.1.0/tcp/7337/p2p/QmSPa4gxx6PRmoNRu6P2iFwEwmayaoLdR5By3i3MgM9gMv"
	hwAddrs1 := []rpcclient.HighwayAddr{
		rpcclient.HighwayAddr{Libp2pAddr: addr1},
		rpcclient.HighwayAddr{Libp2pAddr: addr2},
		rpcclient.HighwayAddr{Libp2pAddr: addr3},
	}
	keeper := NewAddrKeeper()
	keeper.addrs = hwAddrs1

	pid := peer.ID("")
	info1, err := keeper.chooseHighwayFromList(pid)
	assert.Nil(t, err)

	hwAddrs2 := []rpcclient.HighwayAddr{
		rpcclient.HighwayAddr{Libp2pAddr: addr3},
		rpcclient.HighwayAddr{Libp2pAddr: addr2},
		rpcclient.HighwayAddr{Libp2pAddr: addr1},
	}
	keeper.addrs = hwAddrs2

	info2, err := keeper.chooseHighwayFromList(pid)
	assert.Nil(t, err)
	assert.Equal(t, info1, info2)
}

func TestChoosePeerConsistent(t *testing.T) {
	addr1 := "/ip4/0.0.0.0/tcp/7337/p2p/QmSPa4gxx6PRmoNRu6P2iFwEwmayaoLdR5By3i3MgM9gMv"
	addr2 := "/ip4/0.0.0.0/tcp/7337/p2p/QmRWYJ1E6uXzBuY93iMkSDTSdF9XMzLhYcZKwQLLjKV2LW"
	addr3 := "/ip4/0.0.0.0/tcp/7337/p2p/QmQT92nmuhYbRHn6pbrHF2naWSerVaqmWFrEk8p5NfFWST"
	hwAddrs := []rpcclient.HighwayAddr{
		rpcclient.HighwayAddr{Libp2pAddr: addr1},
		rpcclient.HighwayAddr{Libp2pAddr: addr2},
		rpcclient.HighwayAddr{Libp2pAddr: addr3},
	}
	pid := peer.ID("")
	info, err := choosePeer(hwAddrs, pid)
	assert.Nil(t, err)
	assert.Equal(t, hwAddrs[2], info)
}

func TestGetHighwayAddrsRandomly(t *testing.T) {
	resultAddrs := map[string][]rpcclient.HighwayAddr{}
	discoverer := &mocks.HighwayDiscoverer{}
	rpcUsed := map[string]int{}
	discoverer.On("DiscoverHighway", mock.Anything, mock.Anything).Return(resultAddrs, nil).Run(
		func(args mock.Arguments) {
			rpcUsed[args.Get(0).(string)] = 1
		},
	)
	hwAddrs := []rpcclient.HighwayAddr{rpcclient.HighwayAddr{RPCUrl: "abc"}, rpcclient.HighwayAddr{RPCUrl: "xyz"}}
	keeper := NewAddrKeeper()
	keeper.addrs = hwAddrs

	for i := 0; i < 100; i++ {
		_, err := keeper.getHighwayAddrs(discoverer)
		assert.Nil(t, err)
	}
	assert.Len(t, rpcUsed, 2)
}

func TestGetAllHighways(t *testing.T) {
	hwAddrs := map[string][]rpcclient.HighwayAddr{
		"all": []rpcclient.HighwayAddr{rpcclient.HighwayAddr{Libp2pAddr: "abc"}, rpcclient.HighwayAddr{Libp2pAddr: "xyz"}},
		"1":   []rpcclient.HighwayAddr{rpcclient.HighwayAddr{Libp2pAddr: "123"}},
	}
	discoverer := &mocks.HighwayDiscoverer{}
	discoverer.On("DiscoverHighway", mock.Anything, mock.Anything).Return(hwAddrs, nil)
	hws, err := getAllHighways(discoverer, "")
	assert.Nil(t, err)
	assert.Equal(t, addresses(hwAddrs["all"]), hws)
}

func TestAddrKeeper_UpdateRTTData(t *testing.T) {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	h1 := bhost.New(swarmt.GenSwarm(t, ctx))
	h2 := bhost.New(swarmt.GenSwarm(t, ctx))
	ReEstimatedRTTTimestep = 1 * time.Second
	fmt.Println(h2.ID())
	fmt.Println(h2.Addrs())
	ps1 := ping.NewPingService(h1)
	_ = ping.NewPingService(h2)

	keeper := &AddrKeeper{
		currentHW: rpcclient.HighwayAddr{Libp2pAddr: h1.Addrs()[0].String() + "/p2p/" + h1.ID().String() + "/"},
		addrs: addresses{
			rpcclient.HighwayAddr{Libp2pAddr: h2.Addrs()[0].String() + "/p2p/" + h2.ID().String() + "/"},
		},
		addrsByRPCUrl:  map[string]*rpcclient.HighwayAddr{},
		locker:         &sync.RWMutex{},
		ignoreRPCUntil: map[rpcclient.HighwayAddr]time.Time{},
		ignoreHWUntil:  map[rpcclient.HighwayAddr]time.Time{},
		ignoreHW:       &cache.Cache{},
		lastRTT:        map[rpcclient.HighwayAddr]*RTTInfo{},
	}
	stopCh := make(chan interface{})
	go func() {
		time.Sleep(10 * time.Second)
		stopCh <- nil
	}()
	keeper.UpdateRTTData(&Host{Host: h1}, ps1, stopCh)
}

func TestPing(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	h1 := bhost.New(swarmt.GenSwarm(t, ctx))
	h2 := bhost.New(swarmt.GenSwarm(t, ctx))

	err := h1.Connect(ctx, peer.AddrInfo{
		ID:    h2.ID(),
		Addrs: h2.Addrs(),
	})

	if err != nil {
		t.Fatal(err)
	}

	ps1 := ping.NewPingService(h1)
	_ = ping.NewPingService(h2)

	testPing(t, ps1, h2.ID())
	// testPing(t, ps2, h1.ID())
}

func testPing(t *testing.T, ps *ping.PingService, p peer.ID) {
	pctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ts := ps.Ping(pctx, p)

	for i := 0; i < 5; i++ {
		select {
		case res := <-ts:
			if res.Error != nil {
				t.Fatal(res.Error)
			}
			t.Log("ping took: ", res.RTT)
		case <-time.After(time.Second * 4):
			t.Fatal("failed to receive ping")
		}
	}

}
