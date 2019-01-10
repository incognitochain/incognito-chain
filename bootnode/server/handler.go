package server

import (
	"fmt"
	"github.com/ninjadotorg/constant/wire"
)

type Handler struct {
	server *RpcServer
}

type PingArgs struct {
	RawAddress string
	PublicKey  string
	SignData   string
}

func (s Handler) Ping(args *PingArgs, peers *[]wire.RawPeer) error {
	fmt.Println("Ping", args)
	// update peer information to server
	s.server.AddOrUpdatePeer(args.RawAddress, args.PublicKey, args.SignData)
	// return note list
	for _, p := range s.server.Peers {
		*peers = append(*peers, wire.RawPeer{p.RawAddress, p.PublicKey})
	}

	fmt.Println("Response", *peers)

	return nil
}
