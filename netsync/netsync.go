package netsync

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/metadata"

	"github.com/incognitochain/incognito-chain/syncker"

	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/mempool"
	"github.com/incognitochain/incognito-chain/peer"
	"github.com/incognitochain/incognito-chain/pubsub"
	"github.com/incognitochain/incognito-chain/wire"
	libp2p "github.com/libp2p/go-libp2p-peer"
	"github.com/patrickmn/go-cache"
)

// NetSync is a gate for message to enter node from network (after Peerconn),
// all message must be process by NetSync before proccessed by other package in node
// NetSync parses message from other peer, identifies type of message
// After parsing, it will detect if message is duplicate or not
// If message is duplicate then discard it, otherwise pass it to the right handler
// NetSync start when node start and run all the time while node is alive
// and it will stop when received quit signal
type NetSync struct {
	started  int32
	shutdown int32

	cMessage chan interface{}
	cQuit    chan struct{}

	cTxMessage chan wire.Message
	cpuUsage   float64

	config *NetSyncConfig
	cache  *NetSyncCache

	usingNewPool bool
	temp         uint64
}

type NetSyncConfig struct {
	Syncker          *syncker.SynckerManager
	BlockChain       *blockchain.BlockChain
	TxMemPool        *mempool.TxPool
	PubSubManager    *pubsub.PubSubManager
	TransactionEvent pubsub.EventChannel // transaction event
	// RoleInCommitteesEvent pubsub.EventChannel // role in committees event
	BeaconBlockEvent pubsub.EventChannel // beacon block event
	ShardBlockEvent  pubsub.EventChannel // shard block event
	RelayShard       []byte
	// RoleInCommittees      int
	// roleInCommitteesMtx   sync.RWMutex
	Server interface {
		// list functions callback which are assigned from Server struct
		PushMessageToPeer(wire.Message, libp2p.ID) error
		PushMessageToAll(wire.Message) error
		PushMessageToShard(msg wire.Message, shard byte) error
	}
	Consensus interface {
		OnBFTMsg(*wire.MessageBFT)
	}
}

type NetSyncCache struct {
	blockCache    *cache.Cache
	txCache       *cache.Cache
	txCacheMtx    sync.Mutex
	blockCacheMtx sync.Mutex
}

func (netSync *NetSync) Init(cfg *NetSyncConfig) {
	netSync.config = cfg
	netSync.cQuit = make(chan struct{})
	netSync.cMessage = make(chan interface{}, 1000)
	netSync.cTxMessage = make(chan wire.Message, 1000)

	// init cache
	blockCache := cache.New(messageLiveTime, messageCleanupInterval)
	txCache := cache.New(messageLiveTime, messageCleanupInterval)
	netSync.cache = &NetSyncCache{
		txCache:    txCache,
		blockCache: blockCache,
	}

	txPoolVersion := config.Param().TxPoolVersion
	if txPoolVersion == 0 {
		netSync.usingNewPool = false
	} else {
		netSync.usingNewPool = true
	}

	// register pubsub channel
	_, subChanTx, err := netSync.config.PubSubManager.RegisterNewSubscriber(pubsub.TransactionHashEnterNodeTopic)
	if err != nil {
		Logger.log.Error(err)
	}
	netSync.config.TransactionEvent = subChanTx
	// _, subChanRole, err := netSync.config.PubSubManager.RegisterNewSubscriber(pubsub.ShardRoleTopic)
	// if err != nil {
	// 	Logger.log.Error(err)
	// }
	// netSync.config.RoleInCommitteesEvent = subChanRole
	_, subChanBeaconBlock, err := netSync.config.PubSubManager.RegisterNewSubscriber(pubsub.NewBeaconBlockTopic)
	if err != nil {
		Logger.log.Error(err)
	}
	netSync.config.BeaconBlockEvent = subChanBeaconBlock
	_, subChanShardBlock, err := netSync.config.PubSubManager.RegisterNewSubscriber(pubsub.NewShardblockTopic)
	if err != nil {
		Logger.log.Error(err)
	}
	netSync.config.ShardBlockEvent = subChanShardBlock
}

func (netSync *NetSync) Start() error {
	// Already started?
	if atomic.AddInt32(&netSync.started, 1) != 1 {
		return NewNetSyncError(AlreadyStartError, errors.New("Already started"))
	}
	Logger.log.Debug("Starting sync manager")
	//netSync.waitgroup.Add(1)
	go netSync.transactionHandler()
	go netSync.messageHandler()
	go netSync.GetCPUUsage()
	go netSync.cacheLoop()
	return nil
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
					// go metrics.AnalyzeTimeSeriesMetricData(map[string]interface{}{
					// 	metrics.Measurement:      metrics.HandleAllMessage,
					// 	metrics.MeasurementValue: float64(1),
					// 	metrics.Tag:              metrics.ShardIDTag,
					// 	metrics.TagValue:         fmt.Sprintf("shardid-%+v", netSync.config.RoleInCommittees)})
					// go metrics.AnalyzeTimeSeriesMetricData(map[string]interface{}{
					// 	metrics.Measurement:      metrics.HandleAllMessageSize,
					// 	metrics.MeasurementValue: float64(reflect.TypeOf(msgC).Size()),
					// 	metrics.Tag:              metrics.ShardIDTag,
					// 	metrics.TagValue:         fmt.Sprintf("shardid-%+v", netSync.config.RoleInCommittees)})
					switch msg := msgC.(type) {
					case *wire.MessageTx, *wire.MessageTxPrivacyToken:
						{
							beaconHeight := netSync.config.BlockChain.GetBeaconBestState().BestBlock.GetHeight()
							var tx metadata.Transaction
							switch msg := msgC.(type) {
							case *wire.MessageTx:
								tx = msg.Transaction
								netSync.handleMessageTx(msg, tx, int64(beaconHeight))
							case *wire.MessageTxPrivacyToken:
								tx = msg.Transaction
								netSync.handleMessageTx(msg, tx, int64(beaconHeight))
							}
						}
					case *wire.MessageBFT:
						{
							netSync.handleMessageBFTMsg(msg)
						}

					default:
						Logger.log.Debugf("Invalid message type in block "+"handler: %T", msg)
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

	//netSync.waitgroup.Done()
	close(netSync.cMessage)
	Logger.log.Debug("Block handler done")
}

func (netSync *NetSync) GetCPUUsage() {
	ticker := time.NewTicker(2 * time.Second)
	idle0, total0 := common.GetCPUSample()
	defer ticker.Stop()
	for _ = range ticker.C {
		idle1, total1 := common.GetCPUSample()
		idle, total := idle1-idle0, total1-total0
		netSync.cpuUsage = 100 * float64(total-idle) / float64(total)
		Logger.log.Debugf("TestCPU: CPU Usage %v %v %v", idle, total, netSync.cpuUsage)
		idle0, total0 = common.GetCPUSample()
	}
}

func (netSync *NetSync) transactionHandler() {
	ticker := time.NewTicker(1 * time.Second)
	txsrps := 100 // total transaction release per second
	defer ticker.Stop()
	queueTx := []wire.Message{}
	for {
		select {
		case msgTx := <-netSync.cTxMessage:
			queueTx = append(queueTx, msgTx)
		case <-ticker.C:
			i := 0
			if netSync.cpuUsage < 80 {
				for ; (i < txsrps) && (i < len(queueTx)); i++ {
					Logger.log.Infof("netSync.cpuUsage %v is low now, send tx to mempool", netSync.cpuUsage)
					netSync.cMessage <- queueTx[i]
				}
				queueTx = queueTx[i:]
			}
		}
	}
}

func (netSync *NetSync) QueueTx(peer *peer.Peer, msg *wire.MessageTx, done chan struct{}) error {
	// Don't accept more transactions if we're shutting down.

	if atomic.LoadInt32(&netSync.shutdown) != 0 {
		done <- struct{}{}
		return NewNetSyncError(AlreadyShutdownError, errors.New("We're shutting down"))
	}
	/*if netSync.cpuUsage > 80 {*/
	/*Logger.log.Infof("Received tx %v, cpuUsage too high %v, send to another pool", msg.Transaction.Hash().String(), netSync.cpuUsage)*/
	/*go func(msg *wire.MessageTx) {*/
	/*netSync.cTxMessage <- msg*/
	/*}(msg)*/
	/*} else {*/
	/*Logger.log.Infof("Received tx %v, send to mempool", msg.Transaction.Hash().String(), netSync.cpuUsage)*/
	/*netSync.cMessage <- msg*/
	/*}*/
	netSync.cMessage <- msg
	return nil
}

func (netSync *NetSync) QueueTxPrivacyToken(peer *peer.Peer, msg *wire.MessageTxPrivacyToken, done chan struct{}) error {
	// Don't accept more transactions if we're shutting down.
	if atomic.LoadInt32(&netSync.shutdown) != 0 {
		done <- struct{}{}
		return NewNetSyncError(AlreadyShutdownError, errors.New("We're shutting down"))
	}
	if netSync.cpuUsage > 80 {
		go func(msg *wire.MessageTx) {
			netSync.cTxMessage <- msg
		}((*wire.MessageTx)(msg))
	} else {
		netSync.cMessage <- msg
	}
	return nil
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

// handleTxMsg handles transaction messages from all peers.
func (netSync *NetSync) handleMessageTx(msg wire.Message, tx metadata.Transaction, beaconHeight int64) {
	netSync.temp++
	Logger.log.Infof("[collect-txs] number of transactions receive:", netSync.temp)
	txHash := tx.Hash().String()
	Logger.log.Debugf(" Handling new message tx %v", txHash)
	if isAdded := netSync.handleCacheTx(*tx.Hash()); !isAdded {
		sID := common.GetShardIDFromLastByte(tx.GetSenderAddrLastByte())
		sBState := netSync.config.BlockChain.GetBestStateShard(sID)
		if sBState == nil {
			Logger.log.Infof("[collect-txs] Received tx from sID %v but cannot get shard Beststate", sID)
			return
		}
		if (sBState.BestBlock == nil) || (sBState.BestBlock.GetProposeTime() < (time.Now().Unix() - int64(time.Hour.Seconds()))) {
			Logger.log.Infof("[collect-txs] 0")
			return
		}
		if !netSync.usingNewPool {
			txDB := sBState.GetCopiedTransactionStateDB()
			err := tx.LoadData(txDB)
			if err == nil {
				hash, _, err := netSync.config.TxMemPool.MaybeAcceptTransaction(tx, beaconHeight)
				if err != nil {
					Logger.log.Infof("[collect-txs] Validate tx %v return err %v", tx.Hash().String(), err)
				} else {
					Logger.log.Debugf("Node got hash of transaction %s", hash.String())
				}
			} else {
				Logger.log.Infof("[collect-txs] Validate tx %v return err %v", tx.Hash().String(), err)
			}

		} else {
			tp, err := netSync.config.BlockChain.GetConfig().PoolManager.GetShardTxsPool(sID)
			if err != nil {
				Logger.log.Infof("[collect-txs] Cannot get tx pool of shard %v, got err %v", sID, err)
			} else {
				if !tp.IsRunning() {
					Logger.log.Infof("[collect-txs] 1")
					return
				}
				tp.GetInbox() <- tx
				Logger.log.Infof("[collect-txs] Sent transaction %+v to pool", txHash)
			}
		}
	}
	Logger.log.Infof("Transaction %+v found in cache", txHash)
}

// handleTxMsg handles transaction messages from all peers.
func (netSync *NetSync) handleMessageTxPrivacyToken(msg *wire.MessageTxPrivacyToken, beaconHeight int64) {
	txHash := msg.Transaction.Hash().String()
	Logger.log.Debugf("Handling new message tx %v", txHash)
	// if !netSync.handleTxWithRole(msg.Transaction) {
	// 	return
	// }
	if isAdded := netSync.handleCacheTx(*msg.Transaction.Hash()); !isAdded {
		tx := msg.Transaction
		sID := common.GetShardIDFromLastByte(tx.GetSenderAddrLastByte())
		tp, err := netSync.config.BlockChain.GetConfig().PoolManager.GetShardTxsPool(sID)
		if err != nil {
			Logger.log.Errorf("Cannot get tx pool of shard %v, got err %v", sID, err)
		} else {
			if !tp.IsRunning() {
				return
			}
			tp.GetInbox() <- tx
			// Broadcast to network
			/*go metrics.AnalyzeTimeSeriesMetricData(map[string]interface{}{
				metrics.Measurement:      metrics.TxEnterNetSyncSuccess,
				metrics.MeasurementValue: float64(1),
				metrics.Tag:              metrics.TxHashTag,
				metrics.TagValue:         msg.Transaction.Hash().String(),
			})*/
			// Logger.log.Debugf("there is hash of transaction %s", tx.Hash().String())
			// err := netSync.config.Server.PushMessageToAll(msg)
			// if err != nil {
			// 	Logger.log.Error(err)
			// } else {
			// netSync.config.TxMemPool.MarkForwardedTransaction(*msg.Transaction.Hash())
			// }
		}
	}
	Logger.log.Infof("Transaction %+v found in cache", txHash)
}

func (netSync *NetSync) handleMessageBFTMsg(msg *wire.MessageBFT) {
	// go metrics.AnalyzeTimeSeriesMetricData(map[string]interface{}{
	// 	metrics.Measurement:      metrics.HandleMessageBFTMsg,
	// 	metrics.MeasurementValue: float64(1),
	// 	metrics.Tag:              metrics.ShardIDTag,
	// 	metrics.TagValue:         fmt.Sprintf("shardid-%+v", netSync.config.RoleInCommittees),
	// })
	Logger.log.Info("Handling new message BFTMsg")
	// startTime := time.Now()
	if err := msg.VerifyMsgSanity(); err != nil {
		Logger.log.Error(err)
		return
	}
	netSync.config.Consensus.OnBFTMsg(msg)
	// go metrics.AnalyzeTimeSeriesMetricData(map[string]interface{}{
	// 	metrics.Measurement:      metrics.HandleMessageBFTMsgTime,
	// 	metrics.MeasurementValue: float64(time.Since(startTime).Seconds()),
	// 	metrics.Tag:              metrics.ShardIDTag,
	// 	metrics.TagValue:         fmt.Sprintf("shardid-%+v", netSync.config.RoleInCommittees),
	// })
}

func (netSync *NetSync) handleCacheBlock(blockHash string) bool {
	netSync.cache.blockCacheMtx.Lock()
	defer netSync.cache.blockCacheMtx.Unlock()
	_, ok := netSync.cache.blockCache.Get(blockHash)
	if ok {
		return true
	}
	err := netSync.cache.blockCache.Add(blockHash, 1, messageLiveTime)
	if err != nil {
		Logger.log.Error(err)
	}
	return false
}

// handleCacheTx - check txHash and cache
func (netSync *NetSync) handleCacheTx(txHash common.Hash) bool {
	netSync.cache.txCacheMtx.Lock()
	defer netSync.cache.txCacheMtx.Unlock()
	_, ok := netSync.cache.txCache.Get(txHash.String())
	if ok {
		return true
	}
	err := netSync.cache.txCache.Add(txHash.String(), 1, messageLiveTime)
	if err != nil {
		Logger.log.Error(err)
	}
	return false
}

// handleTxWithRole - check tx and make decision is processed or not
// func (netSync *NetSync) handleTxWithRole(tx metadata.Transaction) bool {
// 	senderShardID := common.GetShardIDFromLastByte(tx.GetSenderAddrLastByte())
// 	for _, shardID := range netSync.config.RelayShard {
// 		if senderShardID == shardID {
// 			return true
// 		}
// 	}
// 	netSync.config.roleInCommitteesMtx.RLock()
// 	if netSync.config.RoleInCommittees > -1 && byte(netSync.config.RoleInCommittees) == senderShardID {
// 		netSync.config.roleInCommitteesMtx.RUnlock()
// 		return true
// 	} else {
// 		netSync.config.roleInCommitteesMtx.RUnlock()
// 		return false
// 	}
// }

func (netSync *NetSync) cacheLoop() {
	for w := 0; w < workers; w++ {
		go netSync.handleCacheTxHashWorker(netSync.config.TransactionEvent)
	}
	for {
		select {
		case msg := <-netSync.config.ShardBlockEvent:
			{
				if shardBlock, ok := msg.Value.(*types.ShardBlock); !ok {
					continue
				} else {
					go netSync.handleCacheBlock("s" + shardBlock.Header.Hash().String())
				}
			}
		case msg := <-netSync.config.BeaconBlockEvent:
			{
				if beaconBlock, ok := msg.Value.(*types.BeaconBlock); !ok {
					continue
				} else {
					go netSync.handleCacheBlock("b" + beaconBlock.Header.Hash().String())
				}
			}
			// case msg := <-netSync.config.RoleInCommitteesEvent:
			// 	{
			// 		if shardID, ok := msg.Value.(int); !ok {
			// 			continue
			// 		} else {
			// 			netSync.config.roleInCommitteesMtx.Lock()
			// 			netSync.config.RoleInCommittees = shardID
			// 			netSync.config.roleInCommitteesMtx.Unlock()
			// 		}
			// 	}
		}
	}
}

func (netSync *NetSync) handleCacheTxHashWorker(event pubsub.EventChannel) {
	for msg := range event {
		value, ok := msg.Value.(common.Hash)
		if !ok {
			continue
		}
		go netSync.handleCacheTx(value)
		time.Sleep(time.Nanosecond)
	}
}
