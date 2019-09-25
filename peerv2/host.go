package peerv2

import (
	"bufio"
	"bytes"
	"context"
	crypto2 "crypto"
	"fmt"
	"github.com/libp2p/go-libp2p"
	core "github.com/libp2p/go-libp2p-core"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
	p2pgrpc "github.com/paralin/go-libp2p-grpc"
)

type PeerConn struct {
	RemotePeer *Peer
	RW         *bufio.ReadWriter
}

type Peer struct {
	IP            string
	Port          int
	TargetAddress []core.Multiaddr
	PeerID        peer.ID
	PublicKey     crypto2.PublicKey
}

type HostConfig struct {
	MaxConnection int
	PublicIP      string
	Port          int
	PrivateKey    crypto.PrivKey
}

type Host struct {
	Version  string
	Host     host.Host
	SelfPeer *Peer
	GRPC     *p2pgrpc.GRPCProtocol
}

func NewHost(version string, pubIP string, port int, rand []byte) *Host {
	if len(rand) == 0 {
		rand = generateRand()
	}
	if len(rand) < 40 {
		panic("Rand bytes is less than 40")
	}
	buf := bytes.NewBuffer(rand)
	privKey, _, err := crypto.GenerateKeyPairWithReader(crypto.ECDSA, 256, buf)
	catchError(err)

	listenAddr, err := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/%d", pubIP, port))
	catchError(err)

	ctx := context.Background()
	opts := []libp2p.Option{
		libp2p.ConnectionManager(nil),
		libp2p.ListenAddrs(listenAddr),
		libp2p.Identity(privKey),
	}

	p2pHost, err := libp2p.New(ctx, opts...)

	selfPeer := &Peer{
		PeerID:        p2pHost.ID(),
		IP:            pubIP,
		Port:          port,
		TargetAddress: append([]multiaddr.Multiaddr{}, listenAddr),
	}

	node := &Host{
		Host:     p2pHost,
		SelfPeer: selfPeer,
		Version:  version,
		GRPC:     p2pgrpc.NewGRPCProtocol(context.Background(), p2pHost),
	}

	fmt.Println(selfPeer)
	return node
}

func catchError(err error) {
	if err != nil {
		panic(err)
	}
}
