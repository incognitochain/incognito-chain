package netsync

import (
	"sync"
	"sync/atomic"

	peer2 "github.com/libp2p/go-libp2p-peer"
	"github.com/ninjadotorg/constant/blockchain"
	"github.com/ninjadotorg/constant/common"
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
	Server interface {
		// list functions callback which are assigned from Server struct
		PushMessageToPeer(wire.Message, peer2.ID) error
		PushMessageToAll(wire.Message) error
	}
	Consensus interface {
		OnBlockReceived(*blockchain.Block)
		OnRequestSign(*wire.MessageBlockSigReq)
		OnBlockSigReceived(string, string)
		OnInvalidBlockReceived(string, byte, string)
		OnGetChainState(*wire.MessageGetChainState)
		OnChainStateReceived(*wire.MessageChainState)
		OnSwapRequest(swap *wire.MessageSwapRequest)
		OnSwapSig(swap *wire.MessageSwapSig)
		OnSwapUpdate(swap *wire.MessageSwapUpdate)
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
				switch msg := msgChan.(type) {
				case *wire.MessageTx:
					{
						netSync.HandleMessageTx(msg)
					}
					//case *wire.MessageRegistration:
					//	{
					//		netSync.HandleMessageRegisteration(msg)
					//	}
				case *wire.MessageBlock:
					{
						netSync.HandleMessageBlock(msg)
					}
				case *wire.MessageGetBlocks:
					{
						netSync.HandleMessageGetBlocks(msg)
					}
				case *wire.MessageBlockSig:
					{
						netSync.HandleMessageBlockSig(msg)
					}
				case *wire.MessageInvalidBlock:
					{
						netSync.HandleMessageInvalidBlock(msg)
					}
				case *wire.MessageBlockSigReq:
					{
						netSync.HandleMessageRequestSign(msg)
					}
				case *wire.MessageGetChainState:
					{
						netSync.HandleMessageGetChainState(msg)
					}
				case *wire.MessageChainState:
					{
						netSync.HandleMessageChainState(msg)
					}
				case *wire.MessageSwapRequest:
					{
						netSync.HandleMessageSwapRequest(msg)
					}
				case *wire.MessageSwapSig:
					{
						netSync.HandleMessageSwapSig(msg)
					}
				case *wire.MessageSwapUpdate:
					{
						netSync.HandleMessageSwapUpdate(msg)
					}
				default:
					Logger.log.Infof("Invalid message type in block "+"handler: %T", msg)
				}
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
func (netSync *NetSync) QueueBlock(_ *peer.Peer, msg *wire.MessageBlock, done chan struct{}) {
	// Don't accept more transactions if we're shutting down.
	if atomic.LoadInt32(&netSync.shutdown) != 0 {
		done <- struct{}{}
		return
	}
	netSync.cMessage <- msg
}

func (netSync *NetSync) QueueGetBlock(peer *peer.Peer, msg *wire.MessageGetBlocks, done chan struct{}) {
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

func (netSync *NetSync) HandleMessageGetBlocks(msg *wire.MessageGetBlocks) {
	Logger.log.Info("Handling new message - " + wire.CmdGetBlocks)
	blockHash, _ := common.Hash{}.NewHashFromStr(msg.LastBlockHash)
	senderBlockHeaderIndex, chainID, err := netSync.config.BlockChain.GetBlockHeightByBlockHash(blockHash)
	if err == nil {
		bestHashStr := netSync.config.BlockChain.BestState[chainID].BestBlockHash.String()
		Logger.log.Infof("Blockhash from message %s", blockHash.String())
		Logger.log.Infof("Blockhash of bestChain in chainID %d - %s", chainID, bestHashStr)
		Logger.log.Info("index of block %d \n", senderBlockHeaderIndex)
		Logger.log.Info("chainId of block %d \n", chainID)
		if bestHashStr != blockHash.String() {
			// Send Blocks back to requestor
			chainBlocks, _ := netSync.config.BlockChain.GetChainBlocks(chainID)
			for index := int(senderBlockHeaderIndex) + 1; index <= len(chainBlocks); index++ {
				block, _ := netSync.config.BlockChain.GetBlockByBlockHeight(int32(index), chainID)
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
				netSync.config.Server.PushMessageToPeer(blockMsg, peerID)
			}
		}
	} else {
		Logger.log.Error(blockHash.String(), "----------")
		Logger.log.Error(netSync.config.BlockChain.BestState[9].BestBlockHash.String())
		chainBlocks, err2 := netSync.config.BlockChain.GetChainBlocks(9)
		if err2 != nil {
			Logger.log.Error(err2)
		}
		for _, block := range chainBlocks {
			Logger.log.Error(block.Hash().String())
		}
		Logger.log.Error(err)
		Logger.log.Error("No new blocks to return")
	}
}

func (netSync *NetSync) HandleMessageBlock(msg *wire.MessageBlock) {
	Logger.log.Info("Handling new message BlockSig")
	netSync.config.Consensus.OnBlockReceived(&msg.Block)
}

func (netSync *NetSync) HandleMessageBlockSig(msg *wire.MessageBlockSig) {
	Logger.log.Info("Handling new message BlockSig")
	netSync.config.Consensus.OnBlockSigReceived(msg.Validator, msg.BlockSig)
}
func (netSync *NetSync) HandleMessageInvalidBlock(msg *wire.MessageInvalidBlock) {
	Logger.log.Info("Handling new message invalidblock")
	netSync.config.Consensus.OnInvalidBlockReceived(msg.BlockHash, msg.ChainID, msg.Reason)
}

func (netSync *NetSync) HandleMessageRequestSign(msg *wire.MessageBlockSigReq) {
	Logger.log.Info("Handling new message requestsign")
	netSync.config.Consensus.OnRequestSign(msg)
}

func (netSync *NetSync) HandleMessageGetChainState(msg *wire.MessageGetChainState) {
	Logger.log.Info("Handling new message getchainstate")
	netSync.config.Consensus.OnGetChainState(msg)
}

func (netSync *NetSync) HandleMessageChainState(msg *wire.MessageChainState) {
	Logger.log.Info("Handling new message chainstate")
	netSync.config.Consensus.OnChainStateReceived(msg)
}

func (netSync *NetSync) HandleMessageSwapRequest(msg *wire.MessageSwapRequest) {
	Logger.log.Info("Handling new message requestswap")
	netSync.config.Consensus.OnSwapRequest(msg)
}

func (netSync *NetSync) HandleMessageSwapSig(msg *wire.MessageSwapSig) {
	Logger.log.Info("Handling new message signswap")
	netSync.config.Consensus.OnSwapSig(msg)
}

func (netSync *NetSync) HandleMessageSwapUpdate(msg *wire.MessageSwapUpdate) {
	Logger.log.Info("Handling new message SwapUpdate")
	netSync.config.Consensus.OnSwapUpdate(msg)
}
