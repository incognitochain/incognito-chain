package netsync

import (
	"github.com/ninjadotorg/cash-prototype/blockchain"
	"sync"
	"github.com/ninjadotorg/cash-prototype/peer"
	"sync/atomic"
	"log"
)

type NetSync struct {
	started  int32
	shutdown int32

	Chain      *blockchain.BlockChain
	ChainParam *blockchain.Params

	msgChan   chan interface{}
	waitgroup sync.WaitGroup
	quit      chan struct{}

	//
	syncPeer *peer.Peer
}

func (self *NetSync) Start() {
	// Already started?
	if atomic.AddInt32(&self.started, 1) != 1 {
		return
	}
	log.Print("Starting sync manager")
	self.waitgroup.Add(1)
	go self.blockHandler()

}

// blockHandler is the main handler for the sync manager.  It must be run as a
// goroutine.  It processes block and inv messages in a separate goroutine
// from the peer handlers so the block (MsgBlock) messages are handled by a
// single thread without needing to lock memory data structures.  This is
// important because the sync manager controls which blocks are needed and how
// the fetching should proceed.
func (self *NetSync) blockHandler() {
out:
	for {
		select {
		case msgChan := <-self.msgChan:
			{
				switch msg := msgChan.(type) {
				default:
					log.Print("Invalid message type in block "+
						"handler: %T", msg)
				}
			}
		case <-self.quit:
			{
				break out
			}
		}
	}

	self.waitgroup.Done()
	log.Print("Block handler done")
}
