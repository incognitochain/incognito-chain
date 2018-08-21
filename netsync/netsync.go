package netsync

import (
	"github.com/ninjadotorg/cash-prototype/blockchain"
	"sync"
	"github.com/ninjadotorg/cash-prototype/peer"
	"sync/atomic"
	"log"
	"github.com/ninjadotorg/cash-prototype/mempool"
	"github.com/ninjadotorg/cash-prototype/wire"
	"fmt"
)

type NetSync struct {
	started  int32
	shutdown int32

	msgChan   chan interface{}
	waitgroup sync.WaitGroup
	quit      chan struct{}

	//
	syncPeer *peer.Peer

	Config *NetSyncConfig
}

type NetSyncConfig struct {
	Chain      *blockchain.BlockChain
	ChainParam *blockchain.Params
	MemPool    *mempool.TxPool
}

func (self NetSync) New(cfg *NetSyncConfig) (*NetSync, error) {
	self.Config = cfg
	self.quit = make(chan struct{})
	self.msgChan = make(chan interface{})
	return &self, nil
}

func (self *NetSync) Start() {
	// Already started?
	if atomic.AddInt32(&self.started, 1) != 1 {
		return
	}
	log.Print("Starting sync manager")
	self.waitgroup.Add(1)
	go self.messageHandler()

}

// Stop gracefully shuts down the sync manager by stopping all asynchronous
// handlers and waiting for them to finish.
func (self *NetSync) Stop() error {
	if atomic.AddInt32(&self.shutdown, 1) != 1 {
		log.Print("Sync manager is already in the process of " +
			"shutting down")
		return nil
	}

	log.Print("Sync manager shutting down")
	close(self.quit)
	self.waitgroup.Wait()
	return nil
}

// messageHandler is the main handler for the sync manager.  It must be run as a
// goroutine.  It processes block and inv messages in a separate goroutine
// from the peer handlers so the block (MsgBlock) messages are handled by a
// single thread without needing to lock memory data structures.  This is
// important because the sync manager controls which blocks are needed and how
// the fetching should proceed.
func (self *NetSync) messageHandler() {
out:
	for {
		select {
		case msgChan := <-self.msgChan:
			{
				switch msg := msgChan.(type) {
				case *wire.MessageTx:
					{
						self.HandleMessageTx(msg)
					}
				default:
					log.Printf("Invalid message type in block "+"handler: %T", msg)
				}
			}
		case msgChan := <-self.quit:
			{
				log.Println(msgChan)
				break out
			}
		}
	}

	self.waitgroup.Done()
	log.Print("Block handler done")
}

// QueueTx adds the passed transaction message and peer to the block handling
// queue. Responds to the done channel argument after the tx message is
// processed.
func (self *NetSync) QueueTx(_ *peer.Peer, msg *wire.MessageTx, done chan struct{}) {
	// Don't accept more transactions if we're shutting down.
	if atomic.LoadInt32(&self.shutdown) != 0 {
		done <- struct{}{}
		return
	}
	self.msgChan <- msg
}

// handleTxMsg handles transaction messages from all peers.
func (self *NetSync) HandleMessageTx(msg *wire.MessageTx) {
	// TODO get message tx and process, Tuan Anh
	hash, txDesc, error := self.Config.MemPool.CanAcceptTransaction(msg.Transaction)

	if error != nil {
		fmt.Print(error)
	} else {
		fmt.Print("there is hash of transaction", hash)
		fmt.Print("there is priority of transaction in pool", txDesc.StartingPriority)
	}
}
