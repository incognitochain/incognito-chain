package server

import "github.com/ninjadotorg/cash-prototype/peer"

type Handler struct {
	server *RpcServer
}

type PingArgs struct {
	ID string
	SealerPrvKey string
}
func (s Handler) Ping(args *PingArgs, peers *[]peer.RawPeer) error {
	s.server.AddPeer(args.ID, args.SealerPrvKey)

	for _, p := range s.server.Peers {
		*peers = append(*peers, peer.RawPeer{p.ID, p.SealerPrvKey})
	}

	return nil
}