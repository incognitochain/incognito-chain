package netsync

import (
	"github.com/ninjadotorg/cash-prototype/blockchain"
	"sync"
	"github.com/ninjadotorg/cash-prototype/peer"
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
