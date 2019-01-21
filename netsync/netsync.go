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
							netSync.HandleMessageBFTPropose(msg)
						}
					case *wire.MessageBFTPrepare:
						{
							netSync.HandleMessageBFTPrepare(msg)
						}
					case *wire.MessageBFTCommit:
						{
							netSync.HandleMessageBFTCommit(msg)
						}
					case *wire.MessageBFTReady:
						{
							netSync.HandleMessageBFTReady(msg)
						}
					case *wire.MessageBlockBeacon:
						{
							netSync.HandleMessageBlockBeacon(msg)
						}
					case *wire.MessageBlockShard:
						{
							netSync.HandleMessageBlockShard(msg)
						}
					case *wire.MessageCrossShard:
						{
							netSync.HandleMessageCrossShard(msg)
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

					// case *wire.MessageInvalidBlock:
					// 	{
					// 		netSync.HandleMessageInvalidBlock(msg)
					// 	}
					case *wire.MessageGetBeaconState:
						{
							netSync.HandleMessageGetBeaconState(msg)
						}
					case *wire.MessageBeaconState:
						{
							netSync.HandleMessageBeaconState(msg)
						}
					case *wire.MessageGetShardState:
						{
							netSync.HandleMessageGetShardState(msg)
						}
					case *wire.MessageShardState:
						{
							netSync.HandleMessageShardState(msg)
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

func (netSync *NetSync) HandleMessageGetBlockShard(msg *wire.MessageGetBlockShard) {
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
		blk, err := netSync.config.BlockChain.GetShardBlockByHeight(index, msg.ShardID)
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
		err = netSync.config.Server.PushMessageToPeer(msgShardBlk, peerID)
		if err != nil {
			fmt.Println(err)
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
	for index := msg.From; index <= msg.To; index++ {
		blk, err := netSync.config.BlockChain.GetBeaconBlockByHeight(index)
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
		err = netSync.config.Server.PushMessageToPeer(msgBeaconBlk, peerID)
		if err != nil {
			fmt.Println(err)
		}
	}

	// blockHash, _ := common.Hash{}.NewHashFromStr(msg.LastBlockHash)
	// senderBlockHeaderIndex, shardID, err := netSync.config.BlockChain.GetShardBlockHeightByHash(blockHash)
	// if err == nil {
	// 	bestHashStr := netSync.config.BlockChain.BestState[shardID].BestBlockHash.String()
	// 	Logger.log.Infof("Blockhash from message %s", blockHash.String())
	// 	Logger.log.Infof("Blockhash of bestChain in shardID %d - %s", shardID, bestHashStr)
	// 	Logger.log.Info("index of block %d \n", senderBlockHeaderIndex)
	// 	Logger.log.Info("shardID of block %d \n", shardID)
	// 	if bestHashStr != blockHash.String() {
	// 		// Send Blocks back to requestor
	// 		chainBlocks, _ := netSync.config.BlockChain.GetShardBlocks(shardID)
	// 		for index := int(senderBlockHeaderIndex) + 1; index <= len(chainBlocks); index++ {
	// 			block, _ := netSync.config.BlockChain.GetShardBlockByHeight(int32(index), shardID)
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
	// 			netSync.config.Server.PushMessageToPeer(blockMsg, peerID)
	// 		}
	// 	}
	// } else {
	// 	Logger.log.Error(blockHash.String(), "----------")
	// 	Logger.log.Error(netSync.config.BlockChain.BestState[9].BestBlockHash.String())
	// 	chainBlocks, err2 := netSync.config.BlockChain.GetShardBlocks(9)
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

func (netSync *NetSync) HandleMessageBFTPropose(msg *wire.MessageBFTPropose) {
	Logger.log.Info("Handling new message BFTPropose")
	netSync.config.Consensus.OnBFTPropose(msg)
}

func (netSync *NetSync) HandleMessageBFTPrepare(msg *wire.MessageBFTPrepare) {
	Logger.log.Info("Handling new message BFTPrepare")
	netSync.config.Consensus.OnBFTPrepare(msg)
}

func (netSync *NetSync) HandleMessageBFTCommit(msg *wire.MessageBFTCommit) {
	Logger.log.Info("Handling new message BFTCommit")
	netSync.config.Consensus.OnBFTCommit(msg)
}

func (netSync *NetSync) HandleMessageBFTReady(msg *wire.MessageBFTReady) {
	Logger.log.Info("Handling new message BFTReady")
	netSync.config.Consensus.OnBFTReady(msg)
}

// func (netSync *NetSync) HandleMessageInvalidBlock(msg *wire.MessageInvalidBlock) {
// 	Logger.log.Info("Handling new message invalidblock")
// 	netSync.config.Consensus.OnInvalidBlockReceived(msg.BlockHash, msg.shardID, msg.Reason)
// }

func (netSync *NetSync) HandleMessageGetBeaconState(msg *wire.MessageGetBeaconState) {
	Logger.log.Info("Handling new message getbeaconstate")
	peerID, err := libp2p.IDB58Decode(msg.SenderID)
	if err != nil {
		Logger.log.Error(err)
		return
	}
	beaconState, e := netSync.config.BlockChain.GetBeaconState()
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
	netSync.config.Server.PushMessageToPeer(msgBeaconState, peerID)
}

func (netSync *NetSync) HandleMessageBeaconState(msg *wire.MessageBeaconState) {
	Logger.log.Info("Handling new message beaconstate")
	peerID, err := libp2p.IDB58Decode(msg.SenderID)
	if err != nil {
		Logger.log.Error(err)
		return
	}
	netSync.config.BlockChain.OnBeaconStateReceived(&msg.ChainInfo, peerID)
}

func (netSync *NetSync) HandleMessageGetShardState(msg *wire.MessageGetShardState) {
	Logger.log.Info("Handling new message getshardstate")
	peerID, err := libp2p.IDB58Decode(msg.SenderID)
	if err != nil {
		Logger.log.Error(err)
		return
	}
	shardState := netSync.config.BlockChain.GetShardState(msg.ShardID)

	msgShardState, err := wire.MakeEmptyMessage(wire.CmdShardState)
	if err != nil {
		Logger.log.Error(err)
		return
	}
	msgShardState.(*wire.MessageShardState).ChainInfo = *shardState
	netSync.config.Server.PushMessageToPeer(msgShardState, peerID)
}

func (netSync *NetSync) HandleMessageShardState(msg *wire.MessageShardState) {
	Logger.log.Info("Handling new message shardstate")
	peerID, err := libp2p.IDB58Decode(msg.SenderID)
	if err != nil {
		Logger.log.Error(err)
		return
	}
	netSync.config.BlockChain.OnShardStateReceived(&msg.ChainInfo, peerID)
}

// func (netSync *NetSync) HandleMessageSwapRequest(msg *wire.MessageSwapRequest) {
// 	Logger.log.Info("Handling new message requestswap")
// 	netSync.config.Consensus.OnSwapRequest(msg)
// }

// func (netSync *NetSync) HandleMessageSwapSig(msg *wire.MessageSwapSig) {
// 	Logger.log.Info("Handling new message signswap")
// 	netSync.config.Consensus.OnSwapSig(msg)
// }

// func (netSync *NetSync) HandleMessageSwapUpdate(msg *wire.MessageSwapUpdate) {
// 	Logger.log.Info("Handling new message SwapUpdate")
// 	netSync.config.Consensus.OnSwapUpdate(msg)
// }
