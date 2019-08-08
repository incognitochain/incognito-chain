package main

import (
	"github.com/incognitochain/incognito-chain/wire"
	"net/rpc"
)

func main() {
	client, err := rpc.Dial("tcp", "127.0.0.1:9330")
	if err != nil {
		panic(err)
	}
	if client != nil {
		defer client.Close()
		var response []wire.RawPeer
		err := client.Call("Handler.GetPeers", "", &response)
		if err != nil {
			panic(err)
		}
	}
}
