package netsync

import (
	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/metadata"
	lru "github.com/hashicorp/golang-lru"
	"sync"
	"sync/atomic"
	
	"github.com/constant-money/constant-chain/blockchain"
	"github.com/constant-money/constant-chain/mempool"
	"github.com/constant-money/constant-chain/peer"
	"github.com/constant-money/constant-chain/wire"
	libp2p "github.com/libp2p/go-libp2p-peer"
)

const (
	beaconBlockCache = 1000
	shardBlockCache = 1000
	crossShardBlockCache = 500
	shardToBeaconBlockCache = 500
	txCache = 10000
)
type NetSync struct {
	started   int32
	shutdown  int32
	waitgroup sync.WaitGroup

	cMessage chan interface{}
	cQuit    chan struct{}

	config *NetSyncConfig
	Cache  *NetSyncCache
}

type NetSyncConfig struct {
	BlockChain        *blockchain.BlockChain
	ChainParam        *blockchain.Params
	TxMemPool         *mempool.TxPool
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
type NetSyncCache struct {
	beaconBlockCache  *lru.Cache
	shardBlockCache   *lru.Cache
	shardToBeaconBlockCache *lru.Cache
	crossShardBlockCache *lru.Cache
	txCache      *lru.Cache
	txCacheMtx    sync.RWMutex
	CTxCache      chan common.Hash
}
func (netSync NetSync) New(cfg *NetSyncConfig, cTxCache chan common.Hash) *NetSync {
	netSync.config = cfg
	netSync.cQuit = make(chan struct{})
	netSync.cMessage = make(chan interface{})
	beaconBlockCache, _ := lru.New(beaconBlockCache)
	shardBlockCache, _ := lru.New(shardBlockCache)
	txCache, _ := lru.New(txCache)
	shardToBeaconBlockCache, _ := lru.New(shardToBeaconBlockCache)
	crossShardBlockCache, _ := lru.New(crossShardBlockCache)
	netSync.Cache = &NetSyncCache{
		beaconBlockCache: beaconBlockCache,
		shardBlockCache: shardBlockCache,
		txCache: txCache,
		shardToBeaconBlockCache: shardToBeaconBlockCache,
		crossShardBlockCache: crossShardBlockCache,
	}
	netSync.Cache.CTxCache = cTxCache
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
	go netSync.cacheLoop()
}

// Stop gracefully shuts down the sync manager by stopping all asynchronous
// handlers and waiting for them to finish.
func (netSync *NetSync) Stop() {
	if atomic.AddInt32(&netSync.shutdown, 1) != 1 {
		Logger.log.Warn("Sync manager is already in the process of shutting down")
	}

	Logger.log.Warn("Sync manager shutting down")
	close(netSync. cQuit)
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
					case *wire.MessageTx, *wire.MessageTxToken, *wire.MessageTxPrivacyToken:
						{
							switch msg := msgC.(type) {
							case *wire.MessageTx:
								{
									netSync.HandleMessageTx(msg)
								}
							case *wire.MessageTxToken:
								{
									netSync.HandleMessageTxToken(msg)
								}
							case *wire.MessageTxPrivacyToken:
								{
									netSync.HandleMessageTxPrivacyToken(msg)
								}
							}
						}
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
							netSync.HandleMessageBeaconBlock(msg)
						}
					case *wire.MessageBlockShard:
						{
							netSync.HandleMessageShardBlock(msg)
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

func (netSync *NetSync) QueueTxToken(peer *peer.Peer, msg *wire.MessageTxToken, done chan struct{}) {
	// Don't accept more transactions if we're shutting down.
	if atomic.LoadInt32(&netSync.shutdown) != 0 {
		done <- struct{}{}
		return
	}
	netSync.cMessage <- msg
}

func (netSync *NetSync) QueueTxPrivacyToken(peer *peer.Peer, msg *wire.MessageTxPrivacyToken, done chan struct{}) {
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
	if isAdded := netSync.HandleCacheTx(msg.Transaction); !isAdded {
		hash, _, err := netSync.config.TxMemPool.MaybeAcceptTransaction(msg.Transaction)
		if err != nil {
			Logger.log.Error(err)
			
			// Broadcast to network
		} else {
			Logger.log.Infof("there is hash of transaction %s", hash.String())
			err := netSync.config.Server.PushMessageToAll(msg)
			if err != nil {
				Logger.log.Error(err)
			} else {
				netSync.config.TxMemPool.MarkFowardedTransaction(*msg.Transaction.Hash())
			}
		}
	}
}

// handleTxMsg handles transaction messages from all peers.
func (netSync *NetSync) HandleMessageTxToken(msg *wire.MessageTxToken) {
	Logger.log.Info("Handling new message tx")
	if isAdded := netSync.HandleCacheTx(msg.Transaction); !isAdded {
		hash, _, err := netSync.config.TxMemPool.MaybeAcceptTransaction(msg.Transaction)
		
		if err != nil {
			Logger.log.Error(err)
		} else {
			Logger.log.Infof("there is hash of transaction %s", hash.String())
			// Broadcast to network
			err := netSync.config.Server.PushMessageToAll(msg)
			if err != nil {
				Logger.log.Error(err)
			} else {
				netSync.config.TxMemPool.MarkFowardedTransaction(*msg.Transaction.Hash())
			}
		}
	}
}

// handleTxMsg handles transaction messages from all peers.
func (netSync *NetSync) HandleMessageTxPrivacyToken(msg *wire.MessageTxPrivacyToken) {
	Logger.log.Info("Handling new message tx")
	if isAdded := netSync.HandleCacheTx(msg.Transaction); !isAdded {
		hash, _, err := netSync.config.TxMemPool.MaybeAcceptTransaction(msg.Transaction)
		if err != nil {
			Logger.log.Error(err)
		} else {
			Logger.log.Infof("Node got hash of transaction %s", hash.String())
			// Broadcast to network
			err := netSync.config.Server.PushMessageToAll(msg)
			if err != nil {
				Logger.log.Error(err)
			} else {
				netSync.config.TxMemPool.MarkFowardedTransaction(*msg.Transaction.Hash())
			}
		}
	}
}

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

func (netSync *NetSync) HandleMessageBeaconBlock(msg *wire.MessageBlockBeacon) {
	Logger.log.Info("Handling new message BlockBeacon")
	if isAdded := netSync.HandleCacheBeaconBlock(&msg.Block); !isAdded {
		netSync.config.BlockChain.OnBlockBeaconReceived(&msg.Block)
	}
}
func (netSync *NetSync) HandleMessageShardBlock(msg *wire.MessageBlockShard) {
	Logger.log.Info("Handling new message BlockShard")
	if isAdded := netSync.HandleCacheShardBlock(&msg.Block); !isAdded {
		netSync.config.BlockChain.OnBlockShardReceived(&msg.Block)
	}
}
func (netSync *NetSync) HandleMessageCrossShard(msg *wire.MessageCrossShard) {
	Logger.log.Info("Handling new message CrossShard")
	if isAdded := netSync.HandleCacheCrossShardBlock(&msg.Block); !isAdded {
		netSync.config.BlockChain.OnCrossShardBlockReceived(msg.Block)
	}

}
func (netSync *NetSync) HandleMessageShardToBeacon(msg *wire.MessageShardToBeacon) {
	Logger.log.Info("Handling new message ShardToBeacon")
	if isAdded := netSync.HandleCacheShardToBeaconBlock(&msg.Block); !isAdded {
		netSync.config.BlockChain.OnShardToBeaconBlockReceived(msg.Block)
	}
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
	Logger.log.Info("Handling new message peerstate", msg.SenderID)
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
		netSync.GetBlkShardByHashAndSend(peerID, 0, msg.BlksHash, 0)
	} else {
		netSync.GetBlkShardByHeightAndSend(peerID, msg.FromPool, 0, msg.BySpecificHeight, msg.ShardID, msg.BlkHeights, 0)
	}
}

func (netSync *NetSync) HandleMessageGetBlockBeacon(msg *wire.MessageGetBlockBeacon) {
	Logger.log.Info("Handling new message - " + wire.CmdGetBlockBeacon)
	peerID, err := libp2p.IDB58Decode(msg.SenderID)
	if err != nil {
		Logger.log.Error(err)
		return
	}
	if msg.ByHash {
		netSync.GetBlkBeaconByHashAndSend(peerID, msg.BlkHashes)
	} else {
		netSync.GetBlkBeaconByHeightAndSend(peerID, msg.FromPool, msg.BySpecificHeight, msg.BlkHeights)
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
		netSync.GetBlkShardByHashAndSend(peerID, 2, msg.BlkHashes, 0)
	} else {
		netSync.GetBlkShardByHeightAndSend(peerID, msg.FromPool, 2, msg.BySpecificHeight, msg.ShardID, msg.BlkHeights, 0)
	}
}

func (netSync *NetSync) HandleMessageGetCrossShard(msg *wire.MessageGetCrossShard) {
	Logger.log.Info("Handling new message getcrossshard")
	peerID, err := libp2p.IDB58Decode(msg.SenderID)
	if err != nil {
		Logger.log.Error(err)
		return
	}
	if msg.ByHash {
		netSync.GetBlkShardByHashAndSend(peerID, 1, msg.BlkHashes, msg.ToShardID)
	} else {
		netSync.GetBlkShardByHeightAndSend(peerID, msg.FromPool, 1, msg.BySpecificHeight, msg.FromShardID, msg.BlkHeights, msg.ToShardID)
	}
}
func (netSync *NetSync) HandleCacheBeaconBlock(block *blockchain.BeaconBlock) bool {
	_, ok := netSync.Cache.beaconBlockCache.Get(block.Header.Hash())
	if ok {
		return true
	}
	netSync.Cache.beaconBlockCache.Add(block.Header.Hash(), true)
	return false
}

func (netSync *NetSync) HandleCacheShardBlock(block *blockchain.ShardBlock) bool {
	_, ok := netSync.Cache.shardBlockCache.Get(block.Header.Hash())
	if ok {
		return true
	}
	netSync.Cache.shardBlockCache.Add(block.Header.Hash(), true)
	return false
}

func (netSync *NetSync) HandleCacheTx(transaction metadata.Transaction) bool {
	netSync.Cache.txCacheMtx.RLock()
	defer netSync.Cache.txCacheMtx.RUnlock()
	_, ok := netSync.Cache.txCache.Get(*transaction.Hash())
	if ok {
		return true
	}
	return false
}

func (netSync *NetSync) HandleCacheTxHash(txHash common.Hash) {
	_, ok := netSync.Cache.txCache.Get(txHash)
	if !ok {
		netSync.Cache.txCache.Add(txHash, true)
	}
}
func (netSync *NetSync) HandleCacheShardToBeaconBlock(block *blockchain.ShardToBeaconBlock) bool {
	_, ok := netSync.Cache.shardToBeaconBlockCache.Get(block.Header.Hash())
	if ok {
		return true
	}
	netSync.Cache.shardToBeaconBlockCache.Add(block.Header.Hash(), true)
	return false
}
func (netSync *NetSync) HandleCacheCrossShardBlock(block *blockchain.CrossShardBlock) bool {
	_, ok := netSync.Cache.crossShardBlockCache.Get(block.Header.Hash())
	if ok {
		return true
	}
	netSync.Cache.crossShardBlockCache.Add(block.Header.Hash(), true)
	return false
}

func (netSync *NetSync) cacheLoop() {
	for {
		select {
			case txHash := <- netSync.Cache.CTxCache: {
				go func() {
					netSync.Cache.txCacheMtx.Lock()
					defer netSync.Cache.txCacheMtx.Unlock()
					_, ok := netSync.Cache.txCache.Get(txHash)
					if !ok {
						netSync.Cache.txCache.Add(txHash, true)
					}
				}()
			}
		}
	}
}
