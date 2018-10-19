package server

import (
	"fmt"
	"github.com/ninjadotorg/cash/wire"
)

type Handler struct {
	server *RpcServer
}

type PingArgs struct {
	RawAddress string
	PublicKey string
}
func (s Handler) Ping(args *PingArgs, peers *[]wire.RawPeer) error {
	fmt.Println("Ping", args)
	s.server.AddOrUpdatePeer(args.RawAddress, args.PublicKey)
	for _, p := range s.server.Peers {
		*peers = append(*peers, wire.RawPeer{p.RawAddress, p.PublicKey})
	}

	fmt.Println("Response", *peers)

	return nil
}