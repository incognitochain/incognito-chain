package peerv2

import (
	"reflect"
	"testing"
	"time"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/peerv2/mocks"
	"github.com/incognitochain/incognito-chain/peerv2/rpcclient"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var testHighwayAddress = "/ip4/0.0.0.0/tcp/7337/p2p/QmSPa4gxx6PRmoNRu6P2iFwEwmayaoLdR5By3i3MgM9gMv"

func TestDiscoverHighWay(t *testing.T) {
	type args struct {
		discoverPeerAddress string
		shardsStr           []string
	}
	tests := []struct {
		name    string
		args    args
		want    map[string][]string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	// x, err := DiscoverHighWay("0.0.0.0:9330", []string{"all"})
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rpcClient := rpcclient.RPCClient{}
			got, err := rpcClient.DiscoverHighway(tt.args.discoverPeerAddress, tt.args.shardsStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("DiscoverHighWay() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DiscoverHighWay() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestConnectAtStart makes sure connection is established at start-up time
func TestConnectAtStart(t *testing.T) {
	h, net := setupHost()
	// net.On("Connectedness", mock.Anything).Return(network.NotConnected).Return(network.Connected)
	setupConnectedness(net, []network.Connectedness{network.NotConnected, network.Connected})
	var err error
	h.On("Connect", mock.Anything, mock.Anything).Return(err)

	hwAddrs := map[string][]rpcclient.HighwayAddr{"all": []rpcclient.HighwayAddr{rpcclient.HighwayAddr{Libp2pAddr: testHighwayAddress}}}
	discoverer := &mocks.HighwayDiscoverer{}
	discoverer.On("DiscoverHighway", mock.Anything, mock.Anything).Return(hwAddrs, nil)
	cm := &ConnManager{
		LocalHost:        &Host{Host: h},
		discoverer:       discoverer,
		stop:             make(chan int),
		registerRequests: make(chan peer.ID, 1),
	}
	go cm.keepHighwayConnection()
	time.Sleep(11 * time.Second)
	close(cm.stop)

	assert.Equal(t, 1, len(cm.registerRequests), "not connect at startup")
}

// TestConnectWhenMaxedRetry checks if new highway is picked when failing to connect to old highway for some number of times
func TestConnectWhenMaxedRetry(t *testing.T) {
	h, net := setupHost()
	setupConnectedness(net, []network.Connectedness{network.NotConnected, network.Connected, network.NotConnected, network.NotConnected, network.NotConnected, network.NotConnected, network.NotConnected, network.NotConnected, network.NotConnected, network.NotConnected, network.NotConnected, network.NotConnected, network.NotConnected, network.NotConnected, network.NotConnected, network.NotConnected})
	var err error
	h.On("Connect", mock.Anything, mock.Anything).Return(err)

	hwAddrs := map[string][]rpcclient.HighwayAddr{"all": []rpcclient.HighwayAddr{rpcclient.HighwayAddr{Libp2pAddr: testHighwayAddress}}}
	discoverer := &mocks.HighwayDiscoverer{}
	discoverer.On("DiscoverHighway", mock.Anything, mock.Anything).Return(hwAddrs, nil).Times(2)
	cm := &ConnManager{
		LocalHost:        &Host{Host: h},
		discoverer:       discoverer,
		stop:             make(chan int),
		registerRequests: make(chan peer.ID, 1),
	}
	go cm.keepHighwayConnection()
	time.Sleep(10 * ReconnectHighwayTimestep)
	close(cm.stop)

	discoverer.AssertNumberOfCalls(t, "DiscoverHighway", 2)
}

// TestReconnect checks if connection is re-established after being disconnected
func TestReconnect(t *testing.T) {
	h, net := setupHost()
	// Not -> Con -> Not -> Con
	setupConnectedness(
		net,
		[]network.Connectedness{
			network.NotConnected,
			network.Connected,
			network.NotConnected,
			network.Connected,
		},
	)
	var err error
	h.On("Connect", mock.Anything, mock.Anything).Return(err)

	cm := ConnManager{
		DiscoverPeersAddress: testHighwayAddress,
		LocalHost:            &Host{Host: h},
		registerRequests:     make(chan peer.ID, 5),
	}
	for i := 0; i < 4; i++ {
		maxed := cm.checkConnection(&peer.AddrInfo{})
		assert.False(t, maxed)
	}

	assert.Equal(t, 2, len(cm.registerRequests), "not reconnect")
}

func TestPeriodicManageSub(t *testing.T) {
	sc := new(subscribeCounter)
	cm := ConnManager{
		Requester:        &BlockRequester{},
		stop:             make(chan int),
		registerRequests: make(chan peer.ID, 10),
		subscriber:       sc,
	}
	go cm.manageRoleSubscription()
	time.Sleep(2 * time.Second)
	close(cm.stop)

	assert.Equal(t, 1, sc.normal, "not subbed")
}

func TestForcedSub(t *testing.T) {
	sc := new(subscribeCounter)
	cm := ConnManager{
		Requester:        &BlockRequester{},
		stop:             make(chan int),
		registerRequests: make(chan peer.ID, 10),
		subscriber:       sc,
	}
	cm.registerRequests <- peer.ID("") // Sent forced, must sub with forced = True next time
	go cm.manageRoleSubscription()
	time.Sleep(2 * time.Second)
	close(cm.stop)

	assert.Equal(t, 1, sc.forced, "not subbed")
}

func TestChooseHighwayFiltered(t *testing.T) {
	hwAddrs := []rpcclient.HighwayAddr{
		rpcclient.HighwayAddr{Libp2pAddr: testHighwayAddress},
		rpcclient.HighwayAddr{Libp2pAddr: ""},
	}
	pid := peer.ID("")
	_, err := chooseHighway(hwAddrs, pid)
	assert.Nil(t, err)
}

func TestChooseHighwaySorted(t *testing.T) {
	addr1 := "/ip4/0.0.0.0/tcp/7337/p2p/QmSPa4gxx6PRmoNRu6P2iFwEwmayaoLdR5By3i3MgM9gMv"
	addr2 := "/ip4/0.0.0.1/tcp/7337/p2p/QmSPa4gxx6PRmoNRu6P2iFwEwmayaoLdR5By3i3MgM9gMv"
	addr3 := "/ip4/0.0.1.0/tcp/7337/p2p/QmSPa4gxx6PRmoNRu6P2iFwEwmayaoLdR5By3i3MgM9gMv"
	hwAddrs1 := []rpcclient.HighwayAddr{
		rpcclient.HighwayAddr{Libp2pAddr: addr1},
		rpcclient.HighwayAddr{Libp2pAddr: addr2},
		rpcclient.HighwayAddr{Libp2pAddr: addr3},
	}
	pid := peer.ID("")
	info1, err := chooseHighway(hwAddrs1, pid)
	assert.Nil(t, err)

	hwAddrs2 := []rpcclient.HighwayAddr{
		rpcclient.HighwayAddr{Libp2pAddr: addr3},
		rpcclient.HighwayAddr{Libp2pAddr: addr2},
		rpcclient.HighwayAddr{Libp2pAddr: addr1},
	}
	info2, err := chooseHighway(hwAddrs2, pid)
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
	assert.Equal(t, hwAddrs[1], info)
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
	assert.Equal(t, hwAddrs["all"], hws)
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
	for i := 0; i < 100; i++ {
		_, err := getHighwayAddrs(discoverer, hwAddrs)
		assert.Nil(t, err)
	}
	assert.Len(t, rpcUsed, 2)
}

type subscribeCounter struct {
	normal int
	forced int
}

func (subCounter *subscribeCounter) Subscribe(forced bool) error {
	if forced {
		subCounter.forced++
	} else {
		subCounter.normal++
	}
	return nil
}

func (subCounter *subscribeCounter) GetMsgToTopics() msgToTopics {
	return msgToTopics{}
}

func setupHost() (*mocks.Host, *mocks.Network) {
	net := &mocks.Network{}
	h := &mocks.Host{}
	h.On("Network").Return(net)
	h.On("ID").Return(peer.ID(""))
	return h, net
}

func setupConnectedness(net *mocks.Network, values []network.Connectedness) {
	idx := -1
	net.On("Connectedness", mock.Anything).Return(func(_ peer.ID) network.Connectedness {
		if idx+1 < len(values) {
			idx += 1
		}
		return values[idx]
	})
}

func init() {
	Logger.Init(common.NewBackend(nil).Logger("test", true))
}
