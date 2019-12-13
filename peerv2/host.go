package peerv2

import (
	"bufio"
	"context"
	crypto2 "crypto"
	"fmt"

	p2pgrpc "github.com/incognitochain/go-libp2p-grpc"
	"github.com/libp2p/go-libp2p"
	core "github.com/libp2p/go-libp2p-core"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
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

func NewHost(version string, pubIP string, port int, privateKey string) *Host {
	// TODO(@bunyip): handle errors
	var privKey crypto.PrivKey
	if len(privateKey) == 0 {
		privKey, _, _ = crypto.GenerateKeyPair(crypto.ECDSA, 2048)
		m, _ := crypto.MarshalPrivateKey(privKey)
		encoded := crypto.ConfigEncodeKey(m)
		fmt.Println("encoded libp2p key:", encoded)
	} else {
		b, _ := crypto.ConfigDecodeKey(privateKey)
		privKey, _ = crypto.UnmarshalPrivateKey(b)
	}

	listenAddr, err := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/%d", pubIP, port))
	catchError(err)

	ctx := context.Background()
	opts := []libp2p.Option{
		libp2p.ConnectionManager(nil),
		libp2p.ListenAddrs(listenAddr),
		libp2p.Identity(privKey),
	}

	p2pHost, err := libp2p.New(ctx, opts...)
	if err != nil {
		Logger.Criticalf("Couldn't create libp2p host, err: %+v", err)
		catchError(err)
	}

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
		GRPC:     p2pgrpc.NewGRPCProtocol(ctx, p2pHost),
	}

	Logger.Infof("selfPeer: %v %v %v", selfPeer.PeerID.String(), selfPeer.IP, selfPeer.Port)
	return node
}

func catchError(err error) {
	if err != nil {
		panic(err)
	}
}
