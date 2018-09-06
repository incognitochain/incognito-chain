package netsync

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	peer2 "github.com/libp2p/go-libp2p-peer"
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
	//
	syncPeer *peer.Peer

	Config *NetSyncConfig
}

type NetSyncConfig struct {
	BlockChain *blockchain.BlockChain
	ChainParam *blockchain.Params
	MemPool    *mempool.TxPool
	Server interface {
		// list functions callback which are assigned from Server struct
		PushBlockMessageWithPeerId(*blockchain.Block, peer2.ID) error
		UpdateChain(*blockchain.Block)
	}
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
	Logger.log.Info("Starting sync manager")
	self.waitgroup.Add(1)
	go self.messageHandler()
	time.AfterFunc(2*time.Second, func() {

	})
}

// Stop gracefully shuts down the sync manager by stopping all asynchronous
// handlers and waiting for them to finish.
func (self *NetSync) Stop() error {
	if atomic.AddInt32(&self.shutdown, 1) != 1 {
		Logger.log.Info("Sync manager is already in the process of " +
			"shutting down")
		return nil
	}

	Logger.log.Info("Sync manager shutting down")
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
						self.HandleMessageGetBlocks(msg)
					}

				default:
					Logger.log.Infof("Invalid message type in block "+"handler: %T", msg)
				}
			}
		case msgChan := <-self.quit:
			{
				Logger.log.Info(msgChan)
				break out
			}
		}
	}

	self.waitgroup.Done()
	Logger.log.Info("Block handler done")
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
	Logger.log.Info("Handling new message tx")
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
	Logger.log.Info("Handling new message block")
	// TODO get message block and process
	newBlock := msg.Block

	// Skip verify and insert directly to local blockchain
	// There should be a method in blockchain.go to insert block to prevent data-race if we read from memory

	isMainChain, isOrphanBlock, err := self.Config.BlockChain.ProcessBlock(&newBlock)
	_ = isMainChain
	_ = isOrphanBlock
	_ = err

	a := self.Config.BlockChain.BestState.BestBlock.Hash().String()
	Logger.log.Infof(a)
	//if msg.Block.Header.PrevBlockHash == a {
	self.Config.Server.UpdateChain(&newBlock)
	//}
}

func (self *NetSync) HandleMessageGetBlocks(msg *wire.MessageGetBlocks) {
	Logger.log.Info("Handling new message getblock")
	if senderBlockHeaderIndex, err := self.Config.BlockChain.GetBlockHeightByBlockHash(&msg.LastBlockHash); err == nil {
		if self.Config.BlockChain.BestState.BestBlock.Hash() != &msg.LastBlockHash {
			// Send Blocks back to requestor
			allBlocks, _ := self.Config.BlockChain.GetAllBlocks()
			for index := int(senderBlockHeaderIndex) + 1; index < len(allBlocks); index++ {
				block, _ := self.Config.BlockChain.GetBlockByBlockHeight(int32(index))
				fmt.Printf("Send block %x \n", *block.Hash())
				self.Config.Server.PushBlockMessageWithPeerId(block, msg.SenderID)
				time.Sleep(time.Second * 3)
			}
		}
	}

	// Logger.log.Infof("Send a msgVersion: %s", msgNewJSON)
	// rw := self.syncPeer.OutboundReaderWriterStreams[msg.SenderID]
	// self.syncPeer.flagMutex.Lock()
	// rw.Writer.WriteString(msgNewJSON)
	// rw.Writer.Flush()
	// self.syncPeer.flagMutex.Unlock()
}
