package peerv2

import (
	"net/rpc"

	"github.com/pkg/errors"
)

func DiscoverHighWay(
	discoverPeerAddress string,
	shardsStr []string,
) (
	map[string][]HighwayAddr,
	error,
) {
	if discoverPeerAddress == "" {
		return nil, errors.Errorf("empty address")
	}
	client := new(rpc.Client)
	var err error
	client, err = rpc.Dial("tcp", discoverPeerAddress)
	Logger.Info("Dialing...")
	if err != nil {
		return nil, errors.Errorf("Connect to discover peer %v return error %v:", discoverPeerAddress, err)
	}
	defer client.Close()

	Logger.Infof("Connected to %v", discoverPeerAddress)
	req := Request{Shard: shardsStr}
	var res Response
	Logger.Infof("Start dialing RPC server with param %v", req)

	err = client.Call("Handler.GetPeers", req, &res)

	if err != nil {
		return nil, errors.Errorf("Call Handler.GetPeers return error %v", err)
	}
	Logger.Infof("Bootnode return %v", res.PeerPerShard)
	return res.PeerPerShard, nil
}
