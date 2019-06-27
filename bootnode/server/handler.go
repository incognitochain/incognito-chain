package server

import (
	"fmt"

	"github.com/incognitochain/incognito-chain/wire"
)

type Handler struct {
	rpcServer *RpcServer
}

type PingArgs struct {
	RawAddress string
	PublicKey  string
	SignData   string
}

// Ping - handler func which receive data from rpc client,
// add into list current peers and response all of them to client
func (s Handler) Ping(args *PingArgs, responseMessagePeers *[]wire.RawPeer) error {
	fmt.Println("Receive ```Ping``` method from ```RPC client``` with data", args)

	// update peer which have just send information to our rpc server
	err := s.rpcServer.AddOrUpdatePeer(args.RawAddress, args.PublicKey, args.SignData)
	if err != nil {
		return err
	}

	s.rpcServer.peersMtx.Lock()
	defer s.rpcServer.peersMtx.Unlock()
	// return note list
	for _, p := range s.rpcServer.Peers {
		*responseMessagePeers = append(*responseMessagePeers, wire.RawPeer{p.RawAddress, p.PublicKey})
	}
	fmt.Println("Response", *responseMessagePeers)
	return nil
}
