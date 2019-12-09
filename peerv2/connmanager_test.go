package peerv2

import (
	"reflect"
	"testing"
	"time"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/mocks"
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
			got, err := DiscoverHighWay(tt.args.discoverPeerAddress, tt.args.shardsStr)
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

	cm := &ConnManager{
		HighwayAddress:   testHighwayAddress,
		LocalHost:        &Host{Host: h},
		stop:             make(chan int),
		registerRequests: make(chan int, 1),
	}
	go cm.keepHighwayConnection()
	time.Sleep(2 * time.Second)
	close(cm.stop)

	assert.Equal(t, 1, len(cm.registerRequests), "not connect at startup")
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

	cm := &ConnManager{
		HighwayAddress:   testHighwayAddress,
		LocalHost:        &Host{Host: h},
		stop:             make(chan int),
		registerRequests: make(chan int, 10),
	}
	go cm.keepHighwayConnection()
	time.Sleep(4 * time.Second)
	close(cm.stop)

	assert.Equal(t, 2, len(cm.registerRequests), "not reconnect")
}

func setupHost() (*mocks.Host, *mocks.Network) {
	net := &mocks.Network{}
	h := &mocks.Host{}
	h.On("Network").Return(net)
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
