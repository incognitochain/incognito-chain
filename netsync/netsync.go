package netsync

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/pubsub"
	"sync"
	"sync/atomic"
	"time"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/patrickmn/go-cache"

	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/mempool"
	"github.com/incognitochain/incognito-chain/peer"
	"github.com/incognitochain/incognito-chain/wire"
	libp2p "github.com/libp2p/go-libp2p-peer"
)

const (
	workers             = 5
	MsgLiveTime         = 40 * time.Second  // in second
	MsgsCleanupInterval = 300 * time.Second //in second
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
	BlockChain            *blockchain.BlockChain
	ChainParam            *blockchain.Params
	TxMemPool             *mempool.TxPool
	ShardToBeaconPool     blockchain.ShardToBeaconPool
	CrossShardPool        map[byte]blockchain.CrossShardPool
	PubsubManager         *pubsub.PubSubManager
	TransactionEvent      pubsub.EventChannel
	RoleInCommitteesEvent pubsub.EventChannel
	RelayShard            []byte
	RoleInCommittees      int
	roleInCommitteesMtx   sync.RWMutex
	Server                interface {
		// list functions callback which are assigned from Server struct
		PushMessageToPeer(wire.Message, libp2p.ID) error
		PushMessageToAll(wire.Message) error
	}
	Consensus interface {
		OnBFTMsg(wire.Message)
	}
}
type NetSyncCache struct {
	blockCache    *cache.Cache
	txCache       *cache.Cache
	txCacheMtx    sync.Mutex
	blockCacheMtx sync.Mutex
}

func (netSync NetSync) New(cfg *NetSyncConfig) *NetSync {
	netSync.config = cfg
	netSync.cQuit = make(chan struct{})
	netSync.cMessage = make(chan interface{})
	blockCache := cache.New(MsgLiveTime, MsgsCleanupInterval)
	txCache := cache.New(MsgLiveTime, MsgsCleanupInterval)
	netSync.Cache = &NetSyncCache{
		txCache:    txCache,
		blockCache: blockCache,
	}
	_, subChanTx, _ := netSync.config.PubsubManager.RegisterNewSubscriber(pubsub.TransactionHashEnterNodeTopic)
	netSync.config.TransactionEvent = subChanTx
	_, subChanRole, _ := netSync.config.PubsubManager.RegisterNewSubscriber(pubsub.ShardRoleTopic)
	netSync.config.RoleInCommitteesEvent = subChanRole
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
					case *wire.MessageBFTAgree:
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
	if !netSync.HandleTxWithRole(msg.Transaction) {
		return
	}
	if isAdded := netSync.HandleCacheTx(*msg.Transaction.Hash()); !isAdded {
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
				netSync.config.TxMemPool.MarkForwardedTransaction(*msg.Transaction.Hash())
			}
		}
	}
}

// handleTxMsg handles transaction messages from all peers.
func (netSync *NetSync) HandleMessageTxToken(msg *wire.MessageTxToken) {
	Logger.log.Info("Handling new message tx")
	if !netSync.HandleTxWithRole(msg.Transaction) {
		return
	}
	if isAdded := netSync.HandleCacheTx(*msg.Transaction.Hash()); !isAdded {
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
				netSync.config.TxMemPool.MarkForwardedTransaction(*msg.Transaction.Hash())
			}
		}
	}
}

// handleTxMsg handles transaction messages from all peers.
func (netSync *NetSync) HandleMessageTxPrivacyToken(msg *wire.MessageTxPrivacyToken) {
	Logger.log.Info("Handling new message tx")
	if !netSync.HandleTxWithRole(msg.Transaction) {
		return
	}
	if isAdded := netSync.HandleCacheTx(*msg.Transaction.Hash()); !isAdded {
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
				netSync.config.TxMemPool.MarkForwardedTransaction(*msg.Transaction.Hash())
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
	//if oldBlock := netSync.IsOldBeaconBlock(msg.Block.Header.Height); !oldBlock {
	if isAdded := netSync.HandleCacheBlock(msg.Block.Header.Hash()); !isAdded {
		netSync.config.BlockChain.OnBlockBeaconReceived(&msg.Block)
	}
	//}
}
func (netSync *NetSync) HandleMessageShardBlock(msg *wire.MessageBlockShard) {
	Logger.log.Info("Handling new message BlockShard")
	//if oldBlock := netSync.IsOldShardBlock(msg.Block.Header.ShardID, msg.Block.Header.Height); !oldBlock {
	fmt.Println("Shard Block Received In net Sync: ", msg.Block.Header.Height, msg.Block.Header.ShardID, msg.Block.Header.Hash())
	if isAdded := netSync.HandleCacheBlock(msg.Block.Header.Hash()); !isAdded {
		fmt.Println("Shard Block NO Duplicate net Sync: ", msg.Block.Header.Height, msg.Block.Header.ShardID, msg.Block.Header.Hash())
		netSync.config.BlockChain.OnBlockShardReceived(&msg.Block)
		return
	}
	fmt.Println("Shard Block Duplicate net Sync: ", msg.Block.Header.Height, msg.Block.Header.ShardID, msg.Block.Header.Hash())
	//}
}
func (netSync *NetSync) HandleMessageCrossShard(msg *wire.MessageCrossShard) {
	Logger.log.Info("Handling new message CrossShard")
	if isAdded := netSync.HandleCacheBlock(msg.Block.Header.Hash()); !isAdded {
		netSync.config.BlockChain.OnCrossShardBlockReceived(msg.Block)
	}

}
func (netSync *NetSync) HandleMessageShardToBeacon(msg *wire.MessageShardToBeacon) {
	Logger.log.Info("Handling new message ShardToBeacon")
	if isAdded := netSync.HandleCacheBlock(msg.Block.Header.Hash()); !isAdded {
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
func (netSync *NetSync) HandleCacheBlock(blockHash common.Hash) bool {
	netSync.Cache.blockCacheMtx.Lock()
	defer netSync.Cache.blockCacheMtx.Unlock()
	_, ok := netSync.Cache.blockCache.Get(blockHash.String())
	if ok {
		return true
	}
	netSync.Cache.blockCache.Add(blockHash.String(), 1, MsgLiveTime)
	return false
}

func (netSync *NetSync) HandleCacheTx(txHash common.Hash) bool {
	netSync.Cache.txCacheMtx.Lock()
	defer netSync.Cache.txCacheMtx.Unlock()
	_, ok := netSync.Cache.txCache.Get(txHash.String())
	if ok {
		return true
	}
	netSync.Cache.txCache.Add(txHash.String(), 1, MsgLiveTime)
	return false
}

func (netSync *NetSync) HandleCacheTxHash(txHash common.Hash) {
	netSync.Cache.txCacheMtx.Lock()
	defer netSync.Cache.txCacheMtx.Unlock()
	netSync.Cache.txCache.Add(txHash.String(), 1, MsgLiveTime)
}

func (netSync *NetSync) HandleTxWithRole(tx metadata.Transaction) bool {
	senderShardID := common.GetShardIDFromLastByte(tx.GetSenderAddrLastByte())
	for _, shardID := range netSync.config.RelayShard {
		if senderShardID == shardID {
			return true
		}
	}
	netSync.config.roleInCommitteesMtx.RLock()
	if netSync.config.RoleInCommittees > -1 && byte(netSync.config.RoleInCommittees) == senderShardID {
		netSync.config.roleInCommitteesMtx.RUnlock()
		return true
	} else {
		netSync.config.roleInCommitteesMtx.RUnlock()
		return false
	}
}

func (netSync *NetSync) cacheLoop() {
	for w := 0; w < workers; w++ {
		go netSync.HandleCacheTxHashWorker(netSync.config.TransactionEvent)
	}
	for {
		select {
		case msg := <-netSync.config.RoleInCommitteesEvent:
			{
				if shardID, ok := msg.Value.(int); !ok {
					continue
				} else {
					netSync.config.roleInCommitteesMtx.Lock()
					netSync.config.RoleInCommittees = shardID
					netSync.config.roleInCommitteesMtx.Unlock()
				}
			}
		}
	}
}

func (netSync *NetSync) HandleCacheTxHashWorker(event pubsub.EventChannel) {
	for msg := range event {
		value, ok := msg.Value.(common.Hash)
		if !ok {
			continue
		}
		go netSync.HandleCacheTxHash(value)
		time.Sleep(time.Nanosecond)
	}
}

//if old block return true, otherwise return false
func (netSync *NetSync) IsOldShardBlock(shardID byte, blockHeight uint64) bool {
	shardBestState, err := netSync.config.BlockChain.GetShardBestState(shardID)
	// ignore if can't get shard best state
	if err != nil {
		return false
	}
	return shardBestState.ShardHeight >= blockHeight
}
func (netSync *NetSync) IsOldBeaconBlock(blockHeight uint64) bool {
	return netSync.config.BlockChain.GetBeaconHeight() >= blockHeight
}
