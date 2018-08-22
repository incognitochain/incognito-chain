package netsync

import (
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ninjadotorg/cash-prototype/blockchain"
	"github.com/ninjadotorg/cash-prototype/mempool"
	"github.com/ninjadotorg/cash-prototype/peer"
	"github.com/ninjadotorg/cash-prototype/wire"
)

type NetSync struct {
	started  int32
	shutdown int32

	msgChan   chan interface{}
	waitgroup sync.WaitGroup
	quit      chan struct{}
	Server    interface {
		// Push Tx message
		PushBlockMessage(*blockchain.Block)
	}
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
	time.AfterFunc(2*time.Second, func() {

	})
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
				case *wire.MessageBlock:
					{
						self.HandleMessageBlock(msg)
					}
				case *wire.MessageGetBlocks:
					{
						self.HandleMessageGetBlock(msg)
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
	log.Println("Handling new message tx")
	// TODO get message tx and process, Tuan Anh
	hash, txDesc, error := self.Config.MemPool.CanAcceptTransaction(msg.Transaction)

	if error != nil {
		fmt.Print(error)
	} else {
		fmt.Print("there is hash of transaction", hash)
		fmt.Print("there is priority of transaction in pool", txDesc.StartingPriority)
	}
}

// QueueBlock adds the passed block message and peer to the block handling
// queue. Responds to the done channel argument after the block message is
// processed.
func (self *NetSync) QueueBlock(_ *peer.Peer, msg *wire.MessageBlock, done chan struct{}) {
	// Don't accept more transactions if we're shutting down.
	if atomic.LoadInt32(&self.shutdown) != 0 {
		done <- struct{}{}
		return
	}
	self.msgChan <- msg
}

func (self *NetSync) QueueGetBlock(peer *peer.Peer, msg *wire.MessageGetBlocks, done chan struct{}) {
	// Don't accept more transactions if we're shutting down.
	if atomic.LoadInt32(&self.shutdown) != 0 {
		done <- struct{}{}
		return
	}
	self.msgChan <- msg
}

func (self *NetSync) HandleMessageBlock(msg *wire.MessageBlock) {
	log.Println("Handling new message block")
	// TODO get message block and process, Tuan Anh

	// Skip verify and insert directly to local blockchain
	// There should be a method in blockchain.go to insert block to prevent data-race if we read from memory
	if msg.Block.Header.PrevBlockHash == *self.Config.Chain.BestBlock.Hash() {
		newBlock := msg.Block
		self.Config.Chain.Blocks = append(self.Config.Chain.Blocks, &newBlock)
		self.Config.Chain.Headers[*msg.Block.Hash()] = len(self.Config.Chain.Blocks) - 1
		self.Config.Chain.BestBlock = &newBlock
	}

}

func (self *NetSync) HandleMessageGetBlock(msg *wire.MessageGetBlocks) {
	log.Println("Handling new message getblock")
	if senderBlockHeaderIndex, ok := self.Config.Chain.Headers[msg.LastBlockHash]; ok {
		if self.Config.Chain.BestBlock.Hash() != &msg.LastBlockHash {
			// Send Block to requestor
			for index := senderBlockHeaderIndex; index < len(self.Config.Chain.Blocks); index++ {
				msgNew, err := wire.MakeEmptyMessage(wire.CmdBlock)
				msgNew.(*wire.MessageBlock).Block = *self.Config.Chain.Blocks[index]
				if err != nil {
					return
				}
				// msgNewJSON, err := msg.JsonSerialize()
				// if err != nil {
				// 	return
				// }

				time.Sleep(10 * time.Millisecond)
			}
		}
	}

	// log.Printf("Send a msgVersion: %s", msgNewJSON)
	// rw := self.syncPeer.OutboundReaderWriterStreams[msg.SenderID]
	// self.syncPeer.FlagMutex.Lock()
	// rw.Writer.WriteString(msgNewJSON)
	// rw.Writer.Flush()
	// self.syncPeer.FlagMutex.Unlock()
}
