package rpcclient

type HighwayAddr struct {
	Libp2pAddr string
	RPCUrl     string
}

type Response struct {
	PeerPerShard map[string][]HighwayAddr
}

type Request struct {
	Shard []string
}
