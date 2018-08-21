package netsync

import (
	"github.com/ninjadotorg/money-prototype/blockchain"
	"sync"
	"github.com/ninjadotorg/money-prototype/peer"
)

type NetSync struct {
	started  int32
	shutdown int32

	Chain      *blockchain.BlockChain
	ChainParam *blockchain.Params

	wg   sync.WaitGroup
	quit chan struct{}

	//
	syncPeer *peer.Peer
}
