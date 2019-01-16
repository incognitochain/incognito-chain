package netsync

import (
	"fmt"
	"sync"
	"sync/atomic"

	libp2p "github.com/libp2p/go-libp2p-peer"
	"github.com/ninjadotorg/constant/blockchain"
	"github.com/ninjadotorg/constant/mempool"
	"github.com/ninjadotorg/constant/peer"
	"github.com/ninjadotorg/constant/wire"
)

type NetSync struct {
	started   int32
	shutdown  int32
	waitgroup sync.WaitGroup

	cMessage chan interface{}
	cQuit    chan struct{}

	config *NetSyncConfig
}

type NetSyncConfig struct {
	BlockChain *blockchain.BlockChain
	ChainParam *blockchain.Params
	MemTxPool  *mempool.TxPool
	Server     interface {
		// list functions callback which are assigned from Server struct
		PushMessageToPeer(wire.Message, libp2p.ID) error
		PushMessageToAll(wire.Message) error
	}
	Consensus interface {
		OnBFTPropose(*wire.MessageBFTPropose)
		OnBFTPrepare(*wire.MessageBFTPrepare)
		OnBFTCommit(*wire.MessageBFTCommit)
		OnBFTReady(*wire.MessageBFTReady)
	}

	FeeEstimator map[byte]*mempool.FeeEstimator
}

func (self NetSync) New(cfg *NetSyncConfig) *NetSync {
	self.config = cfg
	self.cQuit = make(chan struct{})
	self.cMessage = make(chan interface{})
	return &self
}

func (self *NetSync) Start() {
	// Already started?
	if atomic.AddInt32(&self.started, 1) != 1 {
		return
	}
	Logger.log.Info("Starting sync manager")
	self.waitgroup.Add(1)
	go self.messageHandler()
}

// Stop gracefully shuts down the sync manager by stopping all asynchronous
// handlers and waiting for them to finish.
func (self *NetSync) Stop() {
	if atomic.AddInt32(&self.shutdown, 1) != 1 {
		Logger.log.Warn("Sync manager is already in the process of shutting down")
	}

	Logger.log.Warn("Sync manager shutting down")
	close(self.cQuit)
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
		case msgChan := <-self.cMessage:
			{
				go func(msgC interface{}) {
					switch msg := msgC.(type) {
					case *wire.MessageTx:
						{
							self.HandleMessageTx(msg)
						}
						//case *wire.MessageRegistration:
						//	{
						//		self.HandleMessageRegisteration(msg)
						//	}
					case *wire.MessageBFTPropose:
						{
							self.HandleMessageBFTPropose(msg)
						}
					case *wire.MessageBFTPrepare:
						{
							self.HandleMessageBFTPrepare(msg)
						}
					case *wire.MessageBFTCommit:
						{
							self.HandleMessageBFTCommit(msg)
						}
					case *wire.MessageBFTReady:
						{
							self.HandleMessageBFTReady(msg)
						}
					case *wire.MessageBlockBeacon:
						{
							self.HandleMessageBlockBeacon(msg)
						}
					case *wire.MessageBlockShard:
						{
							self.HandleMessageBlockShard(msg)
						}
					case *wire.MessageCrossShard:
						{
							self.HandleMessageCrossShard(msg)
						}
					case *wire.MessageShardToBeacon:
						{
							self.HandleMessageShardToBeacon(msg)
						}
					case *wire.MessageGetBlockBeacon:
						{
							self.HandleMessageGetBlockBeacon(msg)
						}
					case *wire.MessageGetBlockShard:
						{
							self.HandleMessageGetBlockShard(msg)
						}

					// case *wire.MessageInvalidBlock:
					// 	{
					// 		self.HandleMessageInvalidBlock(msg)
					// 	}
					case *wire.MessageGetBeaconState:
						{
							self.HandleMessageGetBeaconState(msg)
						}
					case *wire.MessageBeaconState:
						{
							self.HandleMessageBeaconState(msg)
						}
					case *wire.MessageGetShardState:
						{
							self.HandleMessageGetShardState(msg)
						}
					case *wire.MessageShardState:
						{
							self.HandleMessageShardState(msg)
						}
					// case *wire.MessageSwapRequest:
					// 	{
					// 		self.HandleMessageSwapRequest(msg)
					// 	}
					// case *wire.MessageSwapSig:
					// 	{
					// 		self.HandleMessageSwapSig(msg)
					// 	}
					// case *wire.MessageSwapUpdate:
					// 	{
					// 		self.HandleMessageSwapUpdate(msg)
					// 	}
					default:
						Logger.log.Infof("Invalid message type in block "+"handler: %T", msg)
					}
				}(msgChan)
			}
		case msgChan := <-self.cQuit:
			{
				Logger.log.Warn(msgChan)
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
/*func (self *NetSync) QueueRegisteration(peer *peer.Peer, msg *wire.MessageRegistration, done chan struct{}) {
	// Don't accept more transactions if we're shutting down.
	if atomic.LoadInt32(&self.shutdown) != 0 {
		done <- struct{}{}
		return
	}
	self.cMessage <- msg
}*/

func (self *NetSync) QueueTx(peer *peer.Peer, msg *wire.MessageTx, done chan struct{}) {
	// Don't accept more transactions if we're shutting down.
	if atomic.LoadInt32(&self.shutdown) != 0 {
		done <- struct{}{}
		return
	}
	self.cMessage <- msg
}

// handleTxMsg handles transaction messages from all peers.
func (self *NetSync) HandleMessageTx(msg *wire.MessageTx) {
	Logger.log.Info("Handling new message tx")
	hash, txDesc, err := self.config.MemTxPool.MaybeAcceptTransaction(msg.Transaction)

	if err != nil {
		Logger.log.Error(err)
	} else {
		Logger.log.Infof("there is hash of transaction %s", hash.String())
		Logger.log.Infof("there is priority of transaction in pool: %d", txDesc.StartingPriority)

		// Broadcast to network
		err := self.config.Server.PushMessageToAll(msg)
		if err != nil {
			Logger.log.Error(err)
		}
	}
}

// handleTxMsg handles transaction messages from all peers.
/*func (self *NetSync) HandleMessageRegisteration(msg *wire.MessageRegistration) {
	Logger.log.Info("Handling new message tx")
	hash, txDesc, err := self.config.MemTxPool.MaybeAcceptTransaction(msg.Transaction)

	if err != nil {
		Logger.log.Error(err)
	} else {
		Logger.log.Infof("there is hash of transaction %s", hash.String())
		Logger.log.Infof("there is priority of transaction in pool: %d", txDesc.StartingPriority)

		// Broadcast to network
		err := self.config.Server.PushMessageToAll(msg)
		if err != nil {
			Logger.log.Error(err)
		}
	}
}*/

// QueueBlock adds the passed block message and peer to the block handling
// queue. Responds to the done channel argument after the block message is
// processed.
func (self *NetSync) QueueBlock(_ *peer.Peer, msg wire.Message, done chan struct{}) {
	// Don't accept more transactions if we're shutting down.
	if atomic.LoadInt32(&self.shutdown) != 0 {
		done <- struct{}{}
		return
	}
	self.cMessage <- msg
}
func (self *NetSync) QueueGetBlockShard(peer *peer.Peer, msg *wire.MessageGetBlockShard, done chan struct{}) {
	// Don't accept more transactions if we're shutting down.
	if atomic.LoadInt32(&self.shutdown) != 0 {
		done <- struct{}{}
		return
	}
	self.cMessage <- msg
}

func (self *NetSync) QueueGetBlockBeacon(peer *peer.Peer, msg *wire.MessageGetBlockBeacon, done chan struct{}) {
	// Don't accept more transactions if we're shutting down.
	if atomic.LoadInt32(&self.shutdown) != 0 {
		done <- struct{}{}
		return
	}
	self.cMessage <- msg
}

func (self *NetSync) QueueMessage(peer *peer.Peer, msg wire.Message, done chan struct{}) {
	// Don't accept more transactions if we're shutting down.
	if atomic.LoadInt32(&self.shutdown) != 0 {
		done <- struct{}{}
		return
	}
	self.cMessage <- msg
}

func (self *NetSync) HandleMessageGetBlockShard(msg *wire.MessageGetBlockShard) {
	fmt.Println()
	Logger.log.Info("Handling new message - " + wire.CmdGetBlockShard)
	fmt.Println()
	peerID, err := libp2p.IDB58Decode(msg.SenderID)
	if err != nil {
		Logger.log.Error(err)
		return
	}
	for index := msg.From; index <= msg.To; index++ {
		fmt.Println()
		fmt.Println("GET BLOCK", index)
		fmt.Println()
		blk, err := self.config.BlockChain.GetShardBlockByHeight(index, msg.ShardID)
		if err != nil {
			fmt.Println(err)
			return
		}
		msgShardBlk, err := wire.MakeEmptyMessage(wire.CmdBlockShard)
		if err != nil {
			fmt.Println(err)
			return
		}
		msgShardBlk.(*wire.MessageBlockShard).Block = *blk
		err = self.config.Server.PushMessageToPeer(msgShardBlk, peerID)
		if err != nil {
			fmt.Println(err)
		}
	}

}
func (self *NetSync) HandleMessageGetBlockBeacon(msg *wire.MessageGetBlockBeacon) {
	fmt.Println()
	Logger.log.Info("Handling new message - " + wire.CmdGetBlockBeacon)
	fmt.Println()
	peerID, err := libp2p.IDB58Decode(msg.SenderID)
	if err != nil {
		Logger.log.Error(err)
		return
	}
	for index := msg.From; index <= msg.To; index++ {
		blk, err := self.config.BlockChain.GetBeaconBlockByHeight(index)
		if err != nil {
			fmt.Println(err)
			return
		}
		msgBeaconBlk, err := wire.MakeEmptyMessage(wire.CmdBlockBeacon)
		if err != nil {
			fmt.Println(err)
			return
		}
		msgBeaconBlk.(*wire.MessageBlockBeacon).Block = *blk
		err = self.config.Server.PushMessageToPeer(msgBeaconBlk, peerID)
		if err != nil {
			fmt.Println(err)
		}
	}

	// blockHash, _ := common.Hash{}.NewHashFromStr(msg.LastBlockHash)
	// senderBlockHeaderIndex, shardID, err := self.config.BlockChain.GetShardBlockHeightByHash(blockHash)
	// if err == nil {
	// 	bestHashStr := self.config.BlockChain.BestState[shardID].BestBlockHash.String()
	// 	Logger.log.Infof("Blockhash from message %s", blockHash.String())
	// 	Logger.log.Infof("Blockhash of bestChain in shardID %d - %s", shardID, bestHashStr)
	// 	Logger.log.Info("index of block %d \n", senderBlockHeaderIndex)
	// 	Logger.log.Info("shardID of block %d \n", shardID)
	// 	if bestHashStr != blockHash.String() {
	// 		// Send Blocks back to requestor
	// 		chainBlocks, _ := self.config.BlockChain.GetShardBlocks(shardID)
	// 		for index := int(senderBlockHeaderIndex) + 1; index <= len(chainBlocks); index++ {
	// 			block, _ := self.config.BlockChain.GetShardBlockByHeight(int32(index), shardID)
	// 			Logger.log.Info("Send block %s \n", block.Hash().String())

	// 			blockMsg, err := wire.MakeEmptyMessage(wire.CmdBlock)
	// 			if err != nil {
	// 				Logger.log.Error(err)
	// 				break
	// 			}

	// 			blockMsg.(*wire.MessageBlock).Block = *block
	// 			if msg.SenderID == "" {
	// 				Logger.log.Error("Sender ID is empty")
	// 				break
	// 			}
	// 			peerID, err := libp2p.IDB58Decode(msg.SenderID)
	// 			if err != nil {
	// 				Logger.log.Error(err)
	// 				break
	// 			}
	// 			self.config.Server.PushMessageToPeer(blockMsg, peerID)
	// 		}
	// 	}
	// } else {
	// 	Logger.log.Error(blockHash.String(), "----------")
	// 	Logger.log.Error(self.config.BlockChain.BestState[9].BestBlockHash.String())
	// 	chainBlocks, err2 := self.config.BlockChain.GetShardBlocks(9)
	// 	if err2 != nil {
	// 		Logger.log.Error(err2)
	// 	}
	// 	for _, block := range chainBlocks {
	// 		Logger.log.Error(block.Hash().String())
	// 	}
	// 	Logger.log.Error(err)
	// 	Logger.log.Error("No new blocks to return")
	// }
}

func (self *NetSync) HandleMessageBlockBeacon(msg *wire.MessageBlockBeacon) {
	Logger.log.Info("Handling new message BlockBeacon")
	self.config.BlockChain.OnBlockBeaconReceived(&msg.Block)
}
func (self *NetSync) HandleMessageBlockShard(msg *wire.MessageBlockShard) {
	Logger.log.Info("Handling new message BlockShard")
	self.config.BlockChain.OnBlockShardReceived(&msg.Block)
}
func (self *NetSync) HandleMessageCrossShard(msg *wire.MessageCrossShard) {
	Logger.log.Info("Handling new message CrossShard")
	self.config.BlockChain.OnCrossShardBlockReceived(msg.Block)

}
func (self *NetSync) HandleMessageShardToBeacon(msg *wire.MessageShardToBeacon) {
	Logger.log.Info("Handling new message ShardToBeacon")
	self.config.BlockChain.OnShardToBeaconBlockReceived(msg.Block)
}

func (self *NetSync) HandleMessageBFTPropose(msg *wire.MessageBFTPropose) {
	Logger.log.Info("Handling new message BFTPropose")
	self.config.Consensus.OnBFTPropose(msg)
}

func (self *NetSync) HandleMessageBFTPrepare(msg *wire.MessageBFTPrepare) {
	Logger.log.Info("Handling new message BFTPrepare")
	self.config.Consensus.OnBFTPrepare(msg)
}

func (self *NetSync) HandleMessageBFTCommit(msg *wire.MessageBFTCommit) {
	Logger.log.Info("Handling new message BFTCommit")
	self.config.Consensus.OnBFTCommit(msg)
}

func (self *NetSync) HandleMessageBFTReady(msg *wire.MessageBFTReady) {
	Logger.log.Info("Handling new message BFTReply")
	self.config.Consensus.OnBFTReady(msg)
}

// func (self *NetSync) HandleMessageInvalidBlock(msg *wire.MessageInvalidBlock) {
// 	Logger.log.Info("Handling new message invalidblock")
// 	self.config.Consensus.OnInvalidBlockReceived(msg.BlockHash, msg.shardID, msg.Reason)
// }

func (self *NetSync) HandleMessageGetBeaconState(msg *wire.MessageGetBeaconState) {
	Logger.log.Info("Handling new message getbeaconstate")
	peerID, err := libp2p.IDB58Decode(msg.SenderID)
	if err != nil {
		Logger.log.Error(err)
		return
	}
	beaconState, e := self.config.BlockChain.GetBeaconState()
	if e != nil {
		Logger.log.Error(e)
		return
	}
	msgBeaconState, err := wire.MakeEmptyMessage(wire.CmdBeaconState)
	if err != nil {
		Logger.log.Error(err)
		return
	}
	msgBeaconState.(*wire.MessageBeaconState).ChainInfo = *beaconState
	self.config.Server.PushMessageToPeer(msgBeaconState, peerID)
}

func (self *NetSync) HandleMessageBeaconState(msg *wire.MessageBeaconState) {
	Logger.log.Info("Handling new message beaconstate")
	peerID, err := libp2p.IDB58Decode(msg.SenderID)
	if err != nil {
		Logger.log.Error(err)
		return
	}
	self.config.BlockChain.OnBeaconStateReceived(&msg.ChainInfo, peerID)
}

func (self *NetSync) HandleMessageGetShardState(msg *wire.MessageGetShardState) {
	Logger.log.Info("Handling new message getshardstate")
	peerID, err := libp2p.IDB58Decode(msg.SenderID)
	if err != nil {
		Logger.log.Error(err)
		return
	}
	shardState := self.config.BlockChain.GetShardState(msg.ShardID)

	msgShardState, err := wire.MakeEmptyMessage(wire.CmdShardState)
	if err != nil {
		Logger.log.Error(err)
		return
	}
	msgShardState.(*wire.MessageShardState).ChainInfo = *shardState
	self.config.Server.PushMessageToPeer(msgShardState, peerID)
}

func (self *NetSync) HandleMessageShardState(msg *wire.MessageShardState) {
	Logger.log.Info("Handling new message shardstate")
	peerID, err := libp2p.IDB58Decode(msg.SenderID)
	if err != nil {
		Logger.log.Error(err)
		return
	}
	self.config.BlockChain.OnShardStateReceived(&msg.ChainInfo, peerID)
}

// func (self *NetSync) HandleMessageSwapRequest(msg *wire.MessageSwapRequest) {
// 	Logger.log.Info("Handling new message requestswap")
// 	self.config.Consensus.OnSwapRequest(msg)
// }

// func (self *NetSync) HandleMessageSwapSig(msg *wire.MessageSwapSig) {
// 	Logger.log.Info("Handling new message signswap")
// 	self.config.Consensus.OnSwapSig(msg)
// }

// func (self *NetSync) HandleMessageSwapUpdate(msg *wire.MessageSwapUpdate) {
// 	Logger.log.Info("Handling new message SwapUpdate")
// 	self.config.Consensus.OnSwapUpdate(msg)
// }
