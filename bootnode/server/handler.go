package server

import "fmt"

type Handler struct {
	server *RpcServer
}

type RawPeer struct {
	RawAddress string
	PublicKey string
}

type PingArgs struct {
	RawAddress string
	PublicKey string
}
func (s Handler) Ping(args *PingArgs, peers *[]RawPeer) error {
	fmt.Println("Ping", args)
	s.server.AddOrUpdatePeer(args.RawAddress, args.PublicKey)

	for _, p := range s.server.Peers {
		*peers = append(*peers, RawPeer{p.RawAddress, p.PublicKey})
	}

	fmt.Println("Response", *peers)

	return nil
}