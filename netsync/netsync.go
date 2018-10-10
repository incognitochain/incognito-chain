package netsync

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	peer2 "github.com/libp2p/go-libp2p-peer"
	"github.com/ninjadotorg/cash-prototype/blockchain"
	"github.com/ninjadotorg/cash-prototype/common"
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

	config *NetSyncConfig
}

type NetSyncConfig struct {
	BlockChain *blockchain.BlockChain
	ChainParam *blockchain.Params
	MemPool    *mempool.TxPool
	Server     interface {
		// list functions callback which are assigned from Server struct
		PushMessageToPeer(wire.Message, peer2.ID) error
	}
	Consensus interface {
		OnBlockReceived(*blockchain.Block)
		OnRequestSign(*wire.MessageRequestSign)
		OnBlockSigReceived(string, string, string)
		OnInvalidBlockReceived(string, byte, string)
		OnGetChainState(*wire.MessageGetChainState)
		OnChainStateReceived(*wire.MessageChainState)
	}
	FeeEstimator map[byte]*mempool.FeeEstimator
}

func (self NetSync) New(cfg *NetSyncConfig) (*NetSync, error) {
	self.config = cfg
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
	// self.waitgroup.Wait()
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
				case *wire.MessageBlockSig:
					{
						self.HandleMessageBlockSig(msg)
					}
				case *wire.MessageInvalidBlock:
					{
						self.HandleMessageInvalidBlock(msg)
					}
				case *wire.MessageRequestSign:
					{
						self.HandleMessageRequestSign(msg)
					}
				case *wire.MessageGetChainState:
					{
						self.HandleMessageGetChainState(msg)
					}
				case *wire.MessageChainState:
					{
						self.HandleMessageChainState(msg)
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
	// TODO get message tx and process
	hash, txDesc, error := self.config.MemPool.MaybeAcceptTransaction(msg.Transaction)

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

func (self *NetSync) QueueMessage(peer *peer.Peer, msg wire.Message, done chan struct{}) {
	// Don't accept more transactions if we're shutting down.
	if atomic.LoadInt32(&self.shutdown) != 0 {
		done <- struct{}{}
		return
	}
	self.msgChan <- msg
}

func (self *NetSync) HandleMessageGetBlocks(msg *wire.MessageGetBlocks) {
	Logger.log.Info("Handling new message - " + wire.CmdGetBlocks)
	blockHash, _ := common.Hash{}.NewHashFromStr(msg.LastBlockHash)
	senderBlockHeaderIndex, chainID, err := self.config.BlockChain.GetBlockHeightByBlockHash(blockHash)
	if err == nil {
		bestHashStr := self.config.BlockChain.BestState[chainID].BestBlockHash.String()
		Logger.log.Infof("Blockhash from message %s", blockHash.String())
		Logger.log.Infof("Blockhash of bestChain in chainID %d - %s", chainID, bestHashStr)
		Logger.log.Info("Index of block %d \n", senderBlockHeaderIndex)
		Logger.log.Info("chainId of block %d \n", chainID)
		if bestHashStr != blockHash.String() {
			// Send Blocks back to requestor
			chainBlocks, _ := self.config.BlockChain.GetChainBlocks(chainID)
			for index := int(senderBlockHeaderIndex) + 1; index <= len(chainBlocks); index++ {
				block, _ := self.config.BlockChain.GetBlockByBlockHeight(int32(index), chainID)
				Logger.log.Info("Send block %s \n", block.Hash().String())

				blockMsg, err := wire.MakeEmptyMessage(wire.CmdBlock)
				if err != nil {
					Logger.log.Error(err)
					break
				}

				blockMsg.(*wire.MessageBlock).Block = *block
				if msg.SenderID == "" {
					Logger.log.Error("Sender ID is empty")
					break
				}
				peerID, err := peer2.IDB58Decode(msg.SenderID)
				if err != nil {
					Logger.log.Error(err)
					break
				}
				self.config.Server.PushMessageToPeer(blockMsg, peerID)
			}
		}
	} else {
		Logger.log.Error(blockHash.String(), "----------")
		Logger.log.Error(self.config.BlockChain.BestState[9].BestBlockHash.String())
		chainBlocks, err2 := self.config.BlockChain.GetChainBlocks(9)
		if err2 != nil {
			Logger.log.Error(err2)
		}
		for _, block := range chainBlocks {
			Logger.log.Error(block.Hash().String())
		}
		Logger.log.Error(err)
		Logger.log.Error("No new blocks to return")
	}

	// Logger.log.Infof("Send a msgVersion: %s", msgNewJSON)
	// rw := self.syncPeer.OutboundReaderWriterStreams[msg.SenderID]
	// self.syncPeer.flagMutex.Lock()
	// rw.Writer.WriteString(msgNewJSON)
	// rw.Writer.Flush()
	// self.syncPeer.flagMutex.Unlock()
}

func (self *NetSync) HandleMessageBlock(msg *wire.MessageBlock) {
	Logger.log.Info("Handling new message BlockSig")
	self.config.Consensus.OnBlockReceived(&msg.Block)
}

func (self *NetSync) HandleMessageBlockSig(msg *wire.MessageBlockSig) {
	Logger.log.Info("Handling new message BlockSig")
	self.config.Consensus.OnBlockSigReceived(msg.BlockHash, msg.Validator, msg.ValidatorSig)
}
func (self *NetSync) HandleMessageInvalidBlock(msg *wire.MessageInvalidBlock) {
	Logger.log.Info("Handling new message invalidblock")
	self.config.Consensus.OnInvalidBlockReceived(msg.BlockHash, msg.ChainID, msg.Reason)
}
func (self *NetSync) HandleMessageRequestSign(msg *wire.MessageRequestSign) {
	Logger.log.Info("Handling new message requestsign")
	self.config.Consensus.OnRequestSign(msg)
}

func (self *NetSync) HandleMessageGetChainState(msg *wire.MessageGetChainState) {
	Logger.log.Info("Handling new message getchainstate")
	self.config.Consensus.OnGetChainState(msg)
}

func (self *NetSync) HandleMessageChainState(msg *wire.MessageChainState) {
	Logger.log.Info("Handling new message chainstate")
	self.config.Consensus.OnChainStateReceived(msg)
}
