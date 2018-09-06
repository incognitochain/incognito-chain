package server

type Handler struct {
	server *RpcServer
}

type RawPeer struct {
	RawAddress string
	SealerPrvKey string
}

type PingArgs struct {
	ID string
	SealerPrvKey string
}
func (s Handler) Ping(args *PingArgs, peers *[]RawPeer) error {
	s.server.AddPeer(args.ID, args.SealerPrvKey)

	for _, p := range s.server.Peers {
		*peers = append(*peers, RawPeer{p.ID, p.SealerPrvKey})
	}

	return nil
}