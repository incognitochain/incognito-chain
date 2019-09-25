package peerv2

import (
	"context"
	"fmt"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
)

type ConnManager struct {
	LocalHost            *Host
	DiscoverPeersAddress string
}

func (s *ConnManager) Start() {

	//connect to proxy node
	proxyIP, proxyPort := ParseListenner(s.DiscoverPeersAddress, "127.0.0.1", 9300)
	ipfsaddr, err := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/%d", proxyIP, proxyPort))
	if err != nil {
		panic(err)
	}
	peerid, err := peer.IDB58Decode("QmbV4AAHWFFEtE67qqmNeEYXs5Yw5xNMS75oEKtdBvfoKN")
	must(s.LocalHost.Host.Connect(context.Background(), peer.AddrInfo{peerid, append([]multiaddr.Multiaddr{}, ipfsaddr)}))

	//server service on this node
	gRPCService := GRPCService_Server{}
	gRPCService.registerServices(s.LocalHost.GRPC.GetGRPCServer())

	//client on this node
	client := GRPCService_Client{s.LocalHost.GRPC}
	res, err := client.ProxyRegister(context.Background(), peerid, "mypub")

	fmt.Println(res, err)
}

//go run *.go --listen "127.0.0.1:9433" --externaladdress "127.0.0.1:9433" --datadir "/data/fullnode" --discoverpeersaddress "127.0.0.1:9330" --loglevel debug
