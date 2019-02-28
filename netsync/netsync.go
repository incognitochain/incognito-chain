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
	BlockChain        *blockchain.BlockChain
	ChainParam        *blockchain.Params
	MemTxPool         *mempool.TxPool
	ShardToBeaconPool blockchain.ShardToBeaconPool
	CrossShardPool    map[byte]blockchain.CrossShardPool
	Server            interface {
		// list functions callback which are assigned from Server struct
		PushMessageToPeer(wire.Message, libp2p.ID) error
		PushMessageToAll(wire.Message) error
	}
	Consensus interface {
		OnBFTMsg(wire.Message)
	}
}

func (netSync NetSync) New(cfg *NetSyncConfig) *NetSync {
	netSync.config = cfg
	netSync.cQuit = make(chan struct{})
	netSync.cMessage = make(chan interface{})
	return &netSync
}

func (netSync *NetSync) Start() {
	// Already started?
	if atomic.AddInt32(&netSync.started, 1) != 1 {
		return
	}
	Logger.log.Info("Starting sync manager")
	netSync.waitgroup.Add(1)
	go netSync.messageHandler()
}

// Stop gracefully shuts down the sync manager by stopping all asynchronous
// handlers and waiting for them to finish.
func (netSync *NetSync) Stop() {
	if atomic.AddInt32(&netSync.shutdown, 1) != 1 {
		Logger.log.Warn("Sync manager is already in the process of shutting down")
	}

	Logger.log.Warn("Sync manager shutting down")
	close(netSync.cQuit)
}

// messageHandler is the main handler for the sync manager.  It must be run as a
// goroutine.  It processes block and inv messages in a separate goroutine
// from the peer handlers so the block (MsgBlock) messages are handled by a
// single thread without needing to lock memory data structures.  This is
// important because the sync manager controls which blocks are needed and how
// the fetching should proceed.
func (netSync *NetSync) messageHandler() {
out:
	for {
		select {
		case msgChan := <-netSync.cMessage:
			{
				go func(msgC interface{}) {
					switch msg := msgC.(type) {
					case *wire.MessageTx:
						{
							netSync.HandleMessageTx(msg)
						}
						//case *wire.MessageRegistration:
						//	{
						//		netSync.HandleMessageRegisteration(msg)
						//	}
					case *wire.MessageBFTPropose:
						{
							netSync.HandleMessageBFTMsg(msg)
						}
					case *wire.MessageBFTPrepare:
						{
							netSync.HandleMessageBFTMsg(msg)
						}
					case *wire.MessageBFTCommit:
						{
							netSync.HandleMessageBFTMsg(msg)
						}
					case *wire.MessageBFTReady:
						{
							netSync.HandleMessageBFTMsg(msg)
						}
					case *wire.MessageBFTReq:
						{
							netSync.HandleMessageBFTMsg(msg)
						}
					case *wire.MessageBlockBeacon:
						{
							netSync.HandleMessageBlockBeacon(msg)
						}
					case *wire.MessageBlockShard:
						{
							netSync.HandleMessageBlockShard(msg)
						}
					case *wire.MessageGetCrossShard:
						{
							netSync.HandleMessageGetCrossShard(msg)
						}
					case *wire.MessageCrossShard:
						{
							netSync.HandleMessageCrossShard(msg)
						}
					case *wire.MessageGetShardToBeacon:
						{
							netSync.HandleMessageGetShardToBeacon(msg)
						}
					case *wire.MessageShardToBeacon:
						{
							netSync.HandleMessageShardToBeacon(msg)
						}
					case *wire.MessageGetBlockBeacon:
						{
							netSync.HandleMessageGetBlockBeacon(msg)
						}
					case *wire.MessageGetBlockShard:
						{
							netSync.HandleMessageGetBlockShard(msg)
						}
					case *wire.MessagePeerState:
						{
							netSync.HandleMessagePeerState(msg)
						}
					// case *wire.MessageSwapRequest:
					// 	{
					// 		netSync.HandleMessageSwapRequest(msg)
					// 	}
					// case *wire.MessageSwapSig:
					// 	{
					// 		netSync.HandleMessageSwapSig(msg)
					// 	}
					// case *wire.MessageSwapUpdate:
					// 	{
					// 		netSync.HandleMessageSwapUpdate(msg)
					// 	}
					default:
						Logger.log.Infof("Invalid message type in block "+"handler: %T", msg)
					}
				}(msgChan)
			}
		case msgChan := <-netSync.cQuit:
			{
				Logger.log.Warn(msgChan)
				break out
			}
		}
	}

	netSync.waitgroup.Done()
	Logger.log.Info("Block handler done")
}

// QueueTx adds the passed transaction message and peer to the block handling
// queue. Responds to the done channel argument after the tx message is
// processed.
/*func (netSync *NetSync) QueueRegisteration(peer *peer.Peer, msg *wire.MessageRegistration, done chan struct{}) {
	// Don't accept more transactions if we're shutting down.
	if atomic.LoadInt32(&netSync.shutdown) != 0 {
		done <- struct{}{}
		return
	}
	netSync.cMessage <- msg
}*/

func (netSync *NetSync) QueueTx(peer *peer.Peer, msg *wire.MessageTx, done chan struct{}) {
	// Don't accept more transactions if we're shutting down.
	if atomic.LoadInt32(&netSync.shutdown) != 0 {
		done <- struct{}{}
		return
	}
	netSync.cMessage <- msg
}

// handleTxMsg handles transaction messages from all peers.
func (netSync *NetSync) HandleMessageTx(msg *wire.MessageTx) {
	Logger.log.Info("Handling new message tx")
	hash, txDesc, err := netSync.config.MemTxPool.MaybeAcceptTransaction(msg.Transaction)

	if err != nil {
		Logger.log.Error(err)
	} else {
		Logger.log.Infof("there is hash of transaction %s", hash.String())
		Logger.log.Infof("there is priority of transaction in pool: %d", txDesc.StartingPriority)

		// Broadcast to network
		err := netSync.config.Server.PushMessageToAll(msg)
		if err != nil {
			Logger.log.Error(err)
		}
	}
}

// handleTxMsg handles transaction messages from all peers.
/*func (netSync *NetSync) HandleMessageRegisteration(msg *wire.MessageRegistration) {
	Logger.log.Info("Handling new message tx")
	hash, txDesc, err := netSync.config.MemTxPool.MaybeAcceptTransaction(msg.Transaction)

	if err != nil {
		Logger.log.Error(err)
	} else {
		Logger.log.Infof("there is hash of transaction %s", hash.String())
		Logger.log.Infof("there is priority of transaction in pool: %d", txDesc.StartingPriority)

		// Broadcast to network
		err := netSync.config.Server.PushMessageToAll(msg)
		if err != nil {
			Logger.log.Error(err)
		}
	}
}*/

// QueueBlock adds the passed block message and peer to the block handling
// queue. Responds to the done channel argument after the block message is
// processed.
func (netSync *NetSync) QueueBlock(_ *peer.Peer, msg wire.Message, done chan struct{}) {
	// Don't accept more transactions if we're shutting down.
	if atomic.LoadInt32(&netSync.shutdown) != 0 {
		done <- struct{}{}
		return
	}
	netSync.cMessage <- msg
}
func (netSync *NetSync) QueueGetBlockShard(peer *peer.Peer, msg *wire.MessageGetBlockShard, done chan struct{}) {
	// Don't accept more transactions if we're shutting down.
	if atomic.LoadInt32(&netSync.shutdown) != 0 {
		done <- struct{}{}
		return
	}
	netSync.cMessage <- msg
}

func (netSync *NetSync) QueueGetBlockBeacon(peer *peer.Peer, msg *wire.MessageGetBlockBeacon, done chan struct{}) {
	// Don't accept more transactions if we're shutting down.
	if atomic.LoadInt32(&netSync.shutdown) != 0 {
		done <- struct{}{}
		return
	}
	netSync.cMessage <- msg
}

func (netSync *NetSync) QueueMessage(peer *peer.Peer, msg wire.Message, done chan struct{}) {
	// Don't accept more transactions if we're shutting down.
	if atomic.LoadInt32(&netSync.shutdown) != 0 {
		done <- struct{}{}
		return
	}
	netSync.cMessage <- msg
}

func (netSync *NetSync) HandleMessageBlockBeacon(msg *wire.MessageBlockBeacon) {
	Logger.log.Info("Handling new message BlockBeacon")
	netSync.config.BlockChain.OnBlockBeaconReceived(&msg.Block)
}
func (netSync *NetSync) HandleMessageBlockShard(msg *wire.MessageBlockShard) {
	Logger.log.Info("Handling new message BlockShard")
	netSync.config.BlockChain.OnBlockShardReceived(&msg.Block)
}
func (netSync *NetSync) HandleMessageCrossShard(msg *wire.MessageCrossShard) {
	Logger.log.Info("Handling new message CrossShard")
	netSync.config.BlockChain.OnCrossShardBlockReceived(msg.Block)

}
func (netSync *NetSync) HandleMessageShardToBeacon(msg *wire.MessageShardToBeacon) {
	Logger.log.Info("Handling new message ShardToBeacon")
	netSync.config.BlockChain.OnShardToBeaconBlockReceived(msg.Block)
}

func (netSync *NetSync) HandleMessageBFTMsg(msg wire.Message) {
	Logger.log.Info("Handling new message BFTMsg")
	if err := msg.VerifyMsgSanity(); err != nil {
		Logger.log.Error(err)
		return
	}
	netSync.config.Consensus.OnBFTMsg(msg)
}

// func (netSync *NetSync) HandleMessageInvalidBlock(msg *wire.MessageInvalidBlock) {
// 	Logger.log.Info("Handling new message invalidblock")
// 	netSync.config.Consensus.OnInvalidBlockReceived(msg.BlockHash, msg.shardID, msg.Reason)
// }

func (netSync *NetSync) HandleMessagePeerState(msg *wire.MessagePeerState) {
	Logger.log.Info("Handling new message peerstate")
	peerID, err := libp2p.IDB58Decode(msg.SenderID)
	if err != nil {
		Logger.log.Error(err)
		return
	}

	netSync.config.BlockChain.OnPeerStateReceived(&msg.Beacon, &msg.Shards, &msg.ShardToBeaconPool, &msg.CrossShardPool, peerID)
}

func (netSync *NetSync) HandleMessageGetBlockShard(msg *wire.MessageGetBlockShard) {
	Logger.log.Info("Handling new message - " + wire.CmdGetBlockShard)
	peerID, err := libp2p.IDB58Decode(msg.SenderID)
	if err != nil {
		Logger.log.Error(err)
		return
	}
	if msg.ByHash {
		for _, blkHash := range msg.BlksHash {
			blk, err := netSync.config.BlockChain.GetShardBlockByHash(&blkHash)
			if err != nil {
				Logger.log.Error(err)
				return
			}
			newMsg, err := wire.MakeEmptyMessage(wire.CmdBlockShard)
			if err != nil {
				Logger.log.Error(err)
				return
			}
			newMsg.(*wire.MessageBlockShard).Block = *blk
			netSync.config.Server.PushMessageToPeer(newMsg, peerID)
		}
	} else {
		for index := msg.From; index <= msg.To; index++ {
			if index == 1 {
				continue
			}
			blk, err := netSync.config.BlockChain.GetShardBlockByHeight(index, msg.ShardID)
			if err != nil {
				Logger.log.Error(err)
				return
			}
			msgShardBlk, err := wire.MakeEmptyMessage(wire.CmdBlockShard)
			if err != nil {
				Logger.log.Error(err)
				return
			}
			msgShardBlk.(*wire.MessageBlockShard).Block = *blk
			err = netSync.config.Server.PushMessageToPeer(msgShardBlk, peerID)
			if err != nil {
				Logger.log.Error(err)
			}
		}
	}
}

func (netSync *NetSync) HandleMessageGetBlockBeacon(msg *wire.MessageGetBlockBeacon) {
	fmt.Println()
	Logger.log.Info("Handling new message - " + wire.CmdGetBlockBeacon)
	fmt.Println()
	peerID, err := libp2p.IDB58Decode(msg.SenderID)
	if err != nil {
		Logger.log.Error(err)
		return
	}
	if msg.ByHash {
		for _, blkHash := range msg.BlksHash {
			blk, err := netSync.config.BlockChain.GetBeaconBlockByHash(&blkHash)
			if err != nil {
				Logger.log.Error(err)
				return
			}
			newMsg, err := wire.MakeEmptyMessage(wire.CmdBlockBeacon)
			if err != nil {
				Logger.log.Error(err)
				return
			}
			newMsg.(*wire.MessageBlockBeacon).Block = *blk
			netSync.config.Server.PushMessageToPeer(newMsg, peerID)
		}
	} else {
		for index := msg.From; index <= msg.To; index++ {
			if index == 1 {
				continue
			}
			blk, err := netSync.config.BlockChain.GetBeaconBlockByHeight(index)
			if err != nil {
				Logger.log.Error(err)
				return
			}
			msgBeaconBlk, err := wire.MakeEmptyMessage(wire.CmdBlockBeacon)
			if err != nil {
				Logger.log.Error(err)
				return
			}
			msgBeaconBlk.(*wire.MessageBlockBeacon).Block = *blk
			err = netSync.config.Server.PushMessageToPeer(msgBeaconBlk, peerID)
			if err != nil {
				Logger.log.Error(err)
			}
		}
	}
}

func (netSync *NetSync) HandleMessageGetShardToBeacon(msg *wire.MessageGetShardToBeacon) {
	Logger.log.Info("Handling new message getshardtobeacon")
	peerID, err := libp2p.IDB58Decode(msg.SenderID)
	if err != nil {
		Logger.log.Error(err)
		return
	}
	if msg.ByHash {
		if msg.FromPool {
			// netSync.config.ShardToBeaconPool.
		} else {
			for _, blkHash := range msg.BlksHash {
				blk, err := netSync.config.BlockChain.GetShardBlockByHash(&blkHash)
				if err != nil {
					Logger.log.Error(err)
					return
				}
				shardToBeaconBlk := blk.CreateShardToBeaconBlock(netSync.config.BlockChain)
				newMsg, err := wire.MakeEmptyMessage(wire.CmdBlkShardToBeacon)
				if err != nil {
					Logger.log.Error(err)
					return
				}
				newMsg.(*wire.MessageShardToBeacon).Block = *shardToBeaconBlk
				netSync.config.Server.PushMessageToPeer(newMsg, peerID)
			}
		}
	} else {
		for index := msg.From; index <= msg.To; index++ {
			if index == 1 {
				continue
			}
			blk, err := netSync.config.BlockChain.GetShardBlockByHeight(index, msg.ShardID)
			if err != nil {
				Logger.log.Error(err)
				return
			}
			shardToBeaconBlk := blk.CreateShardToBeaconBlock(netSync.config.BlockChain)
			msgShardBlk, err := wire.MakeEmptyMessage(wire.CmdBlkShardToBeacon)
			if err != nil {
				Logger.log.Error(err)
				return
			}
			msgShardBlk.(*wire.MessageShardToBeacon).Block = *shardToBeaconBlk
			err = netSync.config.Server.PushMessageToPeer(msgShardBlk, peerID)
			if err != nil {
				Logger.log.Error(err)
			}
		}
	}
}

func (netSync *NetSync) HandleMessageGetCrossShard(msg *wire.MessageGetCrossShard) {
	Logger.log.Info("Handling new message getshardtobeacon")
	peerID, err := libp2p.IDB58Decode(msg.SenderID)
	if err != nil {
		Logger.log.Error(err)
		return
	}
	if msg.FromPool {
		// netSync.config.CrossShardPool.GetBlock()
	} else {
		if msg.ByHash {
			for _, blkHash := range msg.BlksHash {
				blk, err := netSync.config.BlockChain.GetShardBlockByHash(&blkHash)
				if err != nil {
					Logger.log.Error(err)
					return
				}

				crossShardBlk, err := blk.CreateCrossShardBlock(msg.ToShardID)
				if err != nil {
					Logger.log.Error(err)
					return
				}
				newMsg, err := wire.MakeEmptyMessage(wire.CmdCrossShard)
				if err != nil {
					Logger.log.Error(err)
					return
				}
				newMsg.(*wire.MessageCrossShard).Block = *crossShardBlk
				netSync.config.Server.PushMessageToPeer(newMsg, peerID)
			}
		} else {
			for _, blkHeight := range msg.BlksHeight {
				blk, err := netSync.config.BlockChain.GetShardBlockByHeight(blkHeight, msg.FromShardID)
				if err != nil {
					Logger.log.Error(err)
					continue
				}

				crossShardBlk, err := blk.CreateCrossShardBlock(msg.ToShardID)
				if err != nil {
					Logger.log.Error(err)
					continue
				}
				newMsg, err := wire.MakeEmptyMessage(wire.CmdCrossShard)
				if err != nil {
					Logger.log.Error(err)
					continue
				}
				newMsg.(*wire.MessageCrossShard).Block = *crossShardBlk
				netSync.config.Server.PushMessageToPeer(newMsg, peerID)
			}
		}
	}
}
