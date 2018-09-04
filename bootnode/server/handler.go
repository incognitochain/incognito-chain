package server


type Handler struct {
	server *RpcServer
}

func (s Handler) Ping(ID string, peers *[]string) error {
	s.server.AddPeer(ID)

	for _, peer := range s.server.Peers {
		*peers = append(*peers, peer.ID)
	}

	return nil
}