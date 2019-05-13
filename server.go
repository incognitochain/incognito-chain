package main

import (
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/constant-money/constant-chain/databasemp"
	"github.com/constant-money/constant-chain/transaction"
	"log"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/constant-money/constant-chain/addrmanager"
	"github.com/constant-money/constant-chain/blockchain"
	"github.com/constant-money/constant-chain/cashec"
	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/connmanager"
	"github.com/constant-money/constant-chain/consensus/constantbft"
	"github.com/constant-money/constant-chain/database"
	"github.com/constant-money/constant-chain/mempool"
	"github.com/constant-money/constant-chain/netsync"
	"github.com/constant-money/constant-chain/peer"
	"github.com/constant-money/constant-chain/rpcserver"
	"github.com/constant-money/constant-chain/wallet"
	"github.com/constant-money/constant-chain/wire"
	libp2p "github.com/libp2p/go-libp2p-peer"
)

type Server struct {
	started     int32
	startupTime int64

	protocolVersion string
	chainParams     *blockchain.Params
	connManager     *connmanager.ConnManager
	blockChain      *blockchain.BlockChain
	dataBase        database.DatabaseInterface
	rpcServer       *rpcserver.RpcServer

	memPool           *mempool.TxPool
	tempMemPool       *mempool.TxPool
	beaconPool        *mempool.BeaconPool
	shardPool         map[byte]blockchain.ShardPool
	shardToBeaconPool *mempool.ShardToBeaconPool
	crossShardPool    map[byte]blockchain.CrossShardPool

	waitGroup       sync.WaitGroup
	netSync         *netsync.NetSync
	addrManager     *addrmanager.AddrManager
	userKeySet      *cashec.KeySet
	wallet          *wallet.Wallet
	consensusEngine *constantbft.Engine
	blockgen        *blockchain.BlkTmplGenerator
	// The fee estimator keeps track of how long transactions are left in
	// the mempool before they are mined into blocks.
	feeEstimator map[byte]*mempool.FeeEstimator

	cQuit     chan struct{}
	cNewPeers chan *peer.Peer
}

// setupRPCListeners returns a slice of listeners that are configured for use
// with the RPC server depending on the configuration settings for listen
// addresses and TLS.
func (serverObj *Server) setupRPCListeners() ([]net.Listener, error) {
	// Setup TLS if not disabled.
	listenFunc := net.Listen
	if !cfg.DisableTLS {
		Logger.log.Debug("Disable TLS for RPC is false")
		// Generate the TLS cert and key file if both don't already
		// exist.
		if !fileExists(cfg.RPCKey) && !fileExists(cfg.RPCCert) {
			err := rpcserver.GenCertPair(cfg.RPCCert, cfg.RPCKey)
			if err != nil {
				return nil, err
			}
		}
		keyPair, err := tls.LoadX509KeyPair(cfg.RPCCert, cfg.RPCKey)
		if err != nil {
			return nil, err
		}

		tlsConfig := tls.Config{
			Certificates: []tls.Certificate{keyPair},
			MinVersion:   tls.VersionTLS12,
		}

		// Change the standard net.Listen function to the tls one.
		listenFunc = func(net string, laddr string) (net.Listener, error) {
			return tls.Listen(net, laddr, &tlsConfig)
		}
	} else {
		Logger.log.Debug("Disable TLS for RPC is true")
	}

	netAddrs, err := common.ParseListeners(cfg.RPCListeners, "tcp")
	if err != nil {
		return nil, err
	}

	listeners := make([]net.Listener, 0, len(netAddrs))
	for _, addr := range netAddrs {
		listener, err := listenFunc(addr.Network(), addr.String())
		if err != nil {
			log.Printf("Can't listen on %s: %v", addr, err)
			continue
		}
		listeners = append(listeners, listener)
	}

	return listeners, nil
}

/*
NewServer - create server object which control all process of node
*/
func (serverObj *Server) NewServer(listenAddrs string, db database.DatabaseInterface, dbmp databasemp.DatabaseInterface, chainParams *blockchain.Params, protocolVer string, interrupt <-chan struct{}) error {
	// Init data for Server
	serverObj.protocolVersion = protocolVer
	serverObj.chainParams = chainParams
	serverObj.cQuit = make(chan struct{})
	serverObj.cNewPeers = make(chan *peer.Peer)
	serverObj.dataBase = db

	var err error

	serverObj.userKeySet, err = cfg.GetUserKeySet()
	if err != nil {
		if cfg.NodeMode == common.NODEMODE_AUTO || cfg.NodeMode == common.NODEMODE_BEACON || cfg.NodeMode == common.NODEMODE_SHARD {
			Logger.log.Critical(err)
			return err
		} else {
			Logger.log.Error(err)
		}
	}

	serverObj.beaconPool = mempool.GetBeaconPool()
	serverObj.shardToBeaconPool = mempool.GetShardToBeaconPool()

	serverObj.crossShardPool = make(map[byte]blockchain.CrossShardPool)
	serverObj.shardPool = make(map[byte]blockchain.ShardPool)

	serverObj.blockChain = &blockchain.BlockChain{}

	relayShards := []byte{}
	if cfg.RelayShards == "all" {
		for index := 0; index < common.MAX_SHARD_NUMBER; index++ {
			relayShards = append(relayShards, byte(index))
		}
	} else {
		var validPath = regexp.MustCompile(`(?s)[[:digit:]]+`)
		relayShardsStr := validPath.FindAllString(cfg.RelayShards, -1)
		for index := 0; index < len(relayShardsStr); index++ {
			s, err := strconv.Atoi(relayShardsStr[index])
			if err == nil {
				relayShards = append(relayShards, byte(s))
			}
		}
	}

	err = serverObj.blockChain.Init(&blockchain.Config{
		ChainParams:       serverObj.chainParams,
		DataBase:          serverObj.dataBase,
		Interrupt:         interrupt,
		RelayShards:       relayShards,
		BeaconPool:        serverObj.beaconPool,
		ShardPool:         serverObj.shardPool,
		ShardToBeaconPool: serverObj.shardToBeaconPool,
		CrossShardPool:    serverObj.crossShardPool,
		Server:            serverObj,
		UserKeySet:        serverObj.userKeySet,
		NodeMode:          cfg.NodeMode,
	})

	if err != nil {
		return err
	}
	//init beacon pol
	mempool.InitBeaconPool()
	//init shard pool
	mempool.InitShardPool(serverObj.shardPool)
	//init cross shard pool
	mempool.InitCrossShardPool(serverObj.crossShardPool, db)

	//init shard to beacon bool
	mempool.InitShardToBeaconPool()

	// or if it cannot be loaded, create a new one.
	if cfg.FastStartup {
		Logger.log.Debug("Load chain dependencies from DB")
		serverObj.feeEstimator = make(map[byte]*mempool.FeeEstimator)
		for shardID, bestState := range serverObj.blockChain.BestState.Shard {
			_ = bestState
			feeEstimatorData, err := serverObj.dataBase.GetFeeEstimator(shardID)
			if err == nil && len(feeEstimatorData) > 0 {
				feeEstimator, err := mempool.RestoreFeeEstimator(feeEstimatorData)
				if err != nil {
					Logger.log.Errorf("Failed to restore fee estimator %v", err)
					Logger.log.Debug("Init NewFeeEstimator")
					serverObj.feeEstimator[shardID] = mempool.NewFeeEstimator(
						mempool.DefaultEstimateFeeMaxRollback,
						mempool.DefaultEstimateFeeMinRegisteredBlocks)
				} else {
					serverObj.feeEstimator[shardID] = feeEstimator
				}
			}
		}
	} else {
		err := serverObj.dataBase.CleanCommitments()
		if err != nil {
			Logger.log.Error(err)
			return err
		}
		err = serverObj.dataBase.CleanSerialNumbers()
		if err != nil {
			Logger.log.Error(err)
			return err
		}
		err = serverObj.dataBase.CleanFeeEstimator()
		if err != nil {
			Logger.log.Error(err)
			return err
		}

		serverObj.feeEstimator = make(map[byte]*mempool.FeeEstimator)
	}
	// create mempool tx
	serverObj.memPool = &mempool.TxPool{}
	serverObj.memPool.Init(&mempool.Config{
		BlockChain:        serverObj.blockChain,
		DataBase:          serverObj.dataBase,
		ChainParams:       chainParams,
		FeeEstimator:      serverObj.feeEstimator,
		TxLifeTime:        cfg.TxPoolTTL,
		MaxTx:             cfg.TxPoolMaxTx,
		DataBaseMempool:   dbmp,
		IsLoadFromMempool: cfg.LoadMempool,
		PersistMempool:    cfg.PersistMempool,
	})
	serverObj.memPool.AnnouncePersisDatabaseMempool()
	//add tx pool
	serverObj.blockChain.AddTxPool(serverObj.memPool)

	//==============Temp mem pool only used for validation
	serverObj.tempMemPool = &mempool.TxPool{}
	serverObj.tempMemPool.Init(&mempool.Config{
		BlockChain:   serverObj.blockChain,
		DataBase:     serverObj.dataBase,
		ChainParams:  chainParams,
		FeeEstimator: serverObj.feeEstimator,
		MaxTx:        cfg.TxPoolMaxTx,
	})
	serverObj.blockChain.AddTempTxPool(serverObj.tempMemPool)
	//===============
	serverObj.addrManager = addrmanager.New(cfg.DataDir)
	// Init block template generator
	serverObj.blockgen, err = blockchain.BlkTmplGenerator{}.Init(serverObj.memPool, serverObj.blockChain, serverObj.shardToBeaconPool, serverObj.crossShardPool)
	if err != nil {
		return err
	}

	// Init consensus engine
	serverObj.consensusEngine, err = constantbft.Engine{}.Init(&constantbft.EngineConfig{
		CrossShardPool:    serverObj.crossShardPool,
		ShardToBeaconPool: serverObj.shardToBeaconPool,
		ChainParams:       serverObj.chainParams,
		BlockChain:        serverObj.blockChain,
		Server:            serverObj,
		BlockGen:          serverObj.blockgen,
		NodeMode:          cfg.NodeMode,
		UserKeySet:        serverObj.userKeySet,
	})
	if err != nil {
		return err
	}

	// Init Net Sync manager to process messages
	serverObj.netSync = netsync.NetSync{}.New(&netsync.NetSyncConfig{
		BlockChain: serverObj.blockChain,
		ChainParam: chainParams,
		TxMemPool:  serverObj.memPool,
		Server:     serverObj,
		Consensus:  serverObj.consensusEngine,

		ShardToBeaconPool: serverObj.shardToBeaconPool,
		CrossShardPool:    serverObj.crossShardPool,
	})
	// Create a connection manager.
	var peer *peer.Peer
	if !cfg.DisableListen {
		var err error
		peer, err = serverObj.InitListenerPeer(serverObj.addrManager, listenAddrs, cfg.MaxPeers, cfg.MaxOutPeers, cfg.MaxInPeers)
		if err != nil {
			Logger.log.Error(err)
			return err
		}
	}
	connManager := connmanager.ConnManager{}.New(&connmanager.Config{
		OnInboundAccept:      serverObj.InboundPeerConnected,
		OnOutboundConnection: serverObj.OutboundPeerConnected,
		ListenerPeer:         peer,
		DiscoverPeers:        cfg.DiscoverPeers,
		DiscoverPeersAddress: cfg.DiscoverPeersAddress,
		ExternalAddress:      cfg.ExternalAddress,
		// config for connection of shard
		MaxPeersSameShard:  cfg.MaxPeersSameShard,
		MaxPeersOtherShard: cfg.MaxPeersOtherShard,
		MaxPeersOther:      cfg.MaxPeersOther,
		MaxPeersNoShard:    cfg.MaxPeersNoShard,
		MaxPeersBeacon:     cfg.MaxPeersBeacon,
	})
	serverObj.connManager = connManager

	// Start up persistent peers.
	permanentPeers := cfg.ConnectPeers
	if len(permanentPeers) == 0 {
		permanentPeers = cfg.AddPeers
	}

	for _, addr := range permanentPeers {
		go serverObj.connManager.Connect(addr, "", nil)
	}

	if !cfg.DisableRPC {
		// Setup listeners for the configured RPC listen addresses and
		// TLS settings.
		fmt.Println("settingup RPCListeners")
		rpcListeners, err := serverObj.setupRPCListeners()
		if err != nil {
			return err
		}
		if len(rpcListeners) == 0 {
			return errors.New("RPCS: No valid listen address")
		}

		miningPubkeyB58 := ""
		if serverObj.userKeySet != nil {
			miningPubkeyB58 = serverObj.userKeySet.GetPublicKeyB58()
		}
		rpcConfig := rpcserver.RpcServerConfig{
			Listenters:      rpcListeners,
			RPCQuirks:       cfg.RPCQuirks,
			RPCMaxClients:   cfg.RPCMaxClients,
			ChainParams:     chainParams,
			BlockChain:      serverObj.blockChain,
			TxMemPool:       serverObj.memPool,
			Server:          serverObj,
			Wallet:          serverObj.wallet,
			ConnMgr:         serverObj.connManager,
			AddrMgr:         serverObj.addrManager,
			RPCUser:         cfg.RPCUser,
			RPCPass:         cfg.RPCPass,
			RPCLimitUser:    cfg.RPCLimitUser,
			RPCLimitPass:    cfg.RPCLimitPass,
			DisableAuth:     cfg.RPCDisableAuth,
			NodeMode:        cfg.NodeMode,
			FeeEstimator:    serverObj.feeEstimator,
			ProtocolVersion: serverObj.protocolVersion,
			Database:        &serverObj.dataBase,
			IsMiningNode:    cfg.NodeMode != common.NODEMODE_RELAY && miningPubkeyB58 != "", // a node is mining if it constains this condiction when runing
			MiningPubKeyB58: miningPubkeyB58,
		}
		serverObj.rpcServer = &rpcserver.RpcServer{}
		serverObj.rpcServer.Init(&rpcConfig)

		// Signal process shutdown when the RPC server requests it.
		go func() {
			<-serverObj.rpcServer.RequestedProcessShutdown()
			shutdownRequestChannel <- struct{}{}
		}()
	}

	return nil
}

/*
// InboundPeerConnected is invoked by the connection manager when a new
// inbound connection is established.
*/
func (serverObj *Server) InboundPeerConnected(peerConn *peer.PeerConn) {
	Logger.log.Debug("inbound connected")
}

/*
// outboundPeerConnected is invoked by the connection manager when a new
// outbound connection is established.  It initializes a new outbound server
// peer instance, associates it with the relevant state such as the connection
// request instance and the connection itserverObj, and finally notifies the address
// manager of the attempt.
*/
func (serverObj *Server) OutboundPeerConnected(peerConn *peer.PeerConn) {
	Logger.log.Debug("Outbound PEER connected with PEER Id - " + peerConn.RemotePeerID.Pretty())
	err := serverObj.PushVersionMessage(peerConn)
	if err != nil {
		Logger.log.Error(err)
	}
}

/*
// WaitForShutdown blocks until the main listener and peer handlers are stopped.
*/
func (serverObj *Server) WaitForShutdown() {
	serverObj.waitGroup.Wait()
}

/*
// Stop gracefully shuts down the connection manager.
*/
func (serverObj *Server) Stop() error {
	// stop connManager
	serverObj.connManager.Stop()

	// Shutdown the RPC server if it's not disabled.
	if !cfg.DisableRPC && serverObj.rpcServer != nil {
		serverObj.rpcServer.Stop()
	}

	// Save fee estimator in the db
	for shardID, feeEstimator := range serverObj.feeEstimator {
		feeEstimatorData := feeEstimator.Save()
		if len(feeEstimatorData) > 0 {
			err := serverObj.dataBase.StoreFeeEstimator(feeEstimatorData, shardID)
			if err != nil {
				Logger.log.Errorf("Can't save fee estimator data on chain #%d: %v", shardID, err)
			} else {
				Logger.log.Debugf("Save fee estimator data on chain #%d", shardID)
			}
		}
	}

	serverObj.consensusEngine.Stop()
	serverObj.blockChain.StopSync()
	// Signal the remaining goroutines to cQuit.
	close(serverObj.cQuit)
	return nil
}

/*
// peerHandler is used to handle peer operations such as adding and removing
// peers to and from the server, banning peers, and broadcasting messages to
// peers.  It must be run in a goroutine.
*/
func (serverObj *Server) peerHandler() {
	// Start the address manager and sync manager, both of which are needed
	// by peers.  This is done here since their lifecycle is closely tied
	// to this handler and rather than adding more channels to sychronize
	// things, it's easier and slightly faster to simply start and stop them
	// in this handler.
	serverObj.addrManager.Start()
	serverObj.netSync.Start()

	Logger.log.Debug("Start peer handler")

	if len(cfg.ConnectPeers) == 0 {
		for _, addr := range serverObj.addrManager.AddressCache() {
			go serverObj.connManager.Connect(addr.RawAddress, addr.PublicKey, nil)
		}
	}

	go serverObj.connManager.Start(cfg.DiscoverPeersAddress)

out:
	for {
		select {
		case p := <-serverObj.cNewPeers:
			serverObj.handleAddPeerMsg(p)
		case <-serverObj.cQuit:
			{
				break out
			}
		}
	}
	serverObj.netSync.Stop()
	serverObj.addrManager.Stop()
	serverObj.connManager.Stop()
}

/*
// Start begins accepting connections from peers.
*/
func (serverObj Server) Start() {
	// Already started?
	if atomic.AddInt32(&serverObj.started, 1) != 1 {
		return
	}

	Logger.log.Debug("Starting server")
	serverObj.CheckForceUpdateSourceCode()
	if cfg.TestNet {
		Logger.log.Critical("************************" +
			"* Testnet is active *" +
			"************************")
	}
	// Server startup time. Used for the uptime command for uptime calculation.
	serverObj.startupTime = time.Now().Unix()

	// Start the peer handler which in turn starts the address and block
	// managers.
	serverObj.waitGroup.Add(1)

	go serverObj.peerHandler()
	if !cfg.DisableRPC && serverObj.rpcServer != nil {
		serverObj.waitGroup.Add(1)

		// Start the rebroadcastHandler, which ensures user tx received by
		// the RPC server are rebroadcast until being included in a block.
		//go serverObj.rebroadcastHandler()

		serverObj.rpcServer.Start()
	}
	go serverObj.blockChain.StartSyncBlk()

	if cfg.NodeMode != common.NODEMODE_RELAY {
		err := serverObj.consensusEngine.Start()
		if err != nil {
			Logger.log.Error(err)
			go serverObj.Stop()
			return
		}
	}

	if serverObj.memPool != nil {
		txDescs := serverObj.memPool.LoadOrResetDatabaseMP()
		for _, txDesc := range txDescs {
			<-time.Tick(50 * time.Millisecond)
			if !txDesc.IsFowardMessage {
				tx := txDesc.Desc.Tx
				switch tx.GetType() {
				case common.TxNormalType:
					{
						txMsg, err := wire.MakeEmptyMessage(wire.CmdTx)
						if err != nil {
							continue
						}
						normalTx := tx.(*transaction.Tx)
						txMsg.(*wire.MessageTx).Transaction = normalTx
						err = serverObj.PushMessageToAll(txMsg)
						if err == nil {
							serverObj.memPool.MarkFowardedTransaction(*tx.Hash())
						}
					}
				case common.TxCustomTokenType:
					{
						txMsg, err := wire.MakeEmptyMessage(wire.CmdCustomToken)
						if err != nil {
							continue
						}
						customTokenTx := tx.(*transaction.TxCustomToken)
						txMsg.(*wire.MessageTxToken).Transaction = customTokenTx
						err = serverObj.PushMessageToAll(txMsg)
						if err == nil {
							serverObj.memPool.MarkFowardedTransaction(*tx.Hash())
						}
					}
				case common.TxCustomTokenPrivacyType:
					{
						txMsg, err := wire.MakeEmptyMessage(wire.CmdPrivacyCustomToken)
						if err != nil {
							continue
						}
						customPrivacyTokenTx := tx.(*transaction.TxCustomTokenPrivacy)
						txMsg.(*wire.MessageTxPrivacyToken).Transaction = customPrivacyTokenTx
						err = serverObj.PushMessageToAll(txMsg)
						if err == nil {
							serverObj.memPool.MarkFowardedTransaction(*tx.Hash())
						}
					}
				}
			}
		}
		go mempool.TxPoolMainLoop(serverObj.memPool)
	}
}

func (serverObject Server) CheckForceUpdateSourceCode() {
	if common.NextForceUpdate == "" {
		return
	}
	Logger.log.Warn("\n*********************************************************************************\n" +
		"* Detected a Force Updating Time for this source code from https://github.com/constant-money/constant-chain at " + common.NextForceUpdate + " *" +
		"\n*********************************************************************************\n")
	go func() {
		for true {
			now := time.Now()
			forceTime, _ := time.ParseInLocation(common.DateInputFormat, common.NextForceUpdate, time.Local)
			fmt.Println(now)
			fmt.Println(forceTime)
			forced := now.After(forceTime)
			if forced {
				Logger.log.Error("\n*********************************************************************************\n" +
					"We're exited because having a force update on this souce code." +
					"\nPlease Update source code at https://github.com/constant-money/constant-chain" +
					"\n*********************************************************************************\n")
				os.Exit(common.ExitCodeForceUpdate)
			}
			Logger.log.Debug("Check time to force update source code from https://github.com/constant-money/constant-chain after " + common.NextForceUpdate)
			time.Sleep(time.Second * 60) // each minute
		}
	}()
}

/*
// initListeners initializes the configured net listeners and adds any bound
// addresses to the address manager. Returns the listeners and a NAT interface,
// which is non-nil if UPnP is in use.
*/
func (serverObj *Server) InitListenerPeer(amgr *addrmanager.AddrManager, listenAddrs string, maxPeers int, maxOutPeers int, maxInPeers int) (*peer.Peer, error) {
	netAddr, err := common.ParseListener(listenAddrs, "ip")
	if err != nil {
		return nil, err
	}

	// use keycache to save listener peer into file, this will make peer id of listener not change after turn off node
	kc := KeyCache{}
	kc.Load(filepath.Join(cfg.DataDir, "listenerpeer.json"))

	// load seed of libp2p from keycache file, if not exist -> save a new data into keycache file
	seed := int64(0)
	seedC, _ := strconv.ParseInt(os.Getenv("LISTENER_PEER_SEED"), 10, 64)
	if seedC == 0 {
		key := "LISTENER_PEER_SEED"
		seedT := kc.Get(key)
		if seedT == nil {
			seed = common.RandInt64()
			kc.Set(key, seed)
		} else {
			seed = int64(seedT.(float64))
		}
	} else {
		seed = seedC
	}

	peer, err := peer.Peer{
		Seed:             seed,
		ListeningAddress: *netAddr,
		Config:           *serverObj.NewPeerConfig(),
		PeerConns:        make(map[string]*peer.PeerConn),
		PendingPeers:     make(map[string]*peer.Peer),
	}.NewPeer()
	peer.Config.MaxInPeers = maxInPeers
	peer.Config.MaxOutPeers = maxOutPeers
	peer.Config.MaxPeers = maxPeers
	if err != nil {
		return nil, err
	}

	kc.Save()
	return peer, nil
}

/*
// newPeerConfig returns the configuration for the listening RemotePeer.
*/
func (serverObj *Server) NewPeerConfig() *peer.Config {
	KeySetUser := serverObj.userKeySet
	config := &peer.Config{
		MessageListeners: peer.MessageListeners{
			OnBlockShard:       serverObj.OnBlockShard,
			OnBlockBeacon:      serverObj.OnBlockBeacon,
			OnCrossShard:       serverObj.OnCrossShard,
			OnShardToBeacon:    serverObj.OnShardToBeacon,
			OnTx:               serverObj.OnTx,
			OnTxToken:          serverObj.OnTxToken,
			OnTxPrivacyToken:   serverObj.OnTxPrivacyToken,
			OnVersion:          serverObj.OnVersion,
			OnGetBlockBeacon:   serverObj.OnGetBlockBeacon,
			OnGetBlockShard:    serverObj.OnGetBlockShard,
			OnGetCrossShard:    serverObj.OnGetCrossShard,
			OnGetShardToBeacon: serverObj.OnGetShardToBeacon,
			OnVerAck:           serverObj.OnVerAck,
			OnGetAddr:          serverObj.OnGetAddr,
			OnAddr:             serverObj.OnAddr,

			//constantbft
			OnBFTMsg: serverObj.OnBFTMsg,
			// OnInvalidBlock:  serverObj.OnInvalidBlock,
			OnPeerState: serverObj.OnPeerState,
			//
			PushRawBytesToShard:  serverObj.PushRawBytesToShard,
			PushRawBytesToBeacon: serverObj.PushRawBytesToBeacon,
			GetCurrentRoleShard:  serverObj.GetCurrentRoleShard,
		},
	}
	if KeySetUser != nil && len(KeySetUser.PrivateKey) != 0 {
		config.UserKeySet = KeySetUser
	}
	return config
}

// OnBlock is invoked when a peer receives a block message.  It
// blocks until the coin block has been fully processed.
func (serverObj *Server) OnBlockShard(p *peer.PeerConn,
	msg *wire.MessageBlockShard) {
	Logger.log.Debug("Receive a new blockshard START")

	var txProcessed chan struct{}
	serverObj.netSync.QueueBlock(nil, msg, txProcessed)
	//<-txProcessed

	Logger.log.Debug("Receive a new blockshard END")
}

func (serverObj *Server) OnBlockBeacon(p *peer.PeerConn,
	msg *wire.MessageBlockBeacon) {
	Logger.log.Debug("Receive a new blockbeacon START")

	var txProcessed chan struct{}
	serverObj.netSync.QueueBlock(nil, msg, txProcessed)
	//<-txProcessed

	Logger.log.Debug("Receive a new blockbeacon END")
}

func (serverObj *Server) OnCrossShard(p *peer.PeerConn,
	msg *wire.MessageCrossShard) {
	Logger.log.Debug("Receive a new crossshard START")

	var txProcessed chan struct{}
	serverObj.netSync.QueueBlock(nil, msg, txProcessed)
	//<-txProcessed

	Logger.log.Debug("Receive a new crossshard END")
}

func (serverObj *Server) OnShardToBeacon(p *peer.PeerConn,
	msg *wire.MessageShardToBeacon) {
	Logger.log.Debug("Receive a new shardToBeacon START")

	var txProcessed chan struct{}
	serverObj.netSync.QueueBlock(nil, msg, txProcessed)
	//<-txProcessed

	Logger.log.Debug("Receive a new shardToBeacon END")
}

func (serverObj *Server) OnGetBlockBeacon(_ *peer.PeerConn, msg *wire.MessageGetBlockBeacon) {
	Logger.log.Debug("Receive a " + msg.MessageType() + " message START")
	var txProcessed chan struct{}
	serverObj.netSync.QueueGetBlockBeacon(nil, msg, txProcessed)
	//<-txProcessed

	Logger.log.Debug("Receive a " + msg.MessageType() + " message END")
}
func (serverObj *Server) OnGetBlockShard(_ *peer.PeerConn, msg *wire.MessageGetBlockShard) {
	Logger.log.Debug("Receive a " + msg.MessageType() + " message START")
	var txProcessed chan struct{}
	serverObj.netSync.QueueGetBlockShard(nil, msg, txProcessed)
	//<-txProcessed

	Logger.log.Debug("Receive a " + msg.MessageType() + " message END")
}

func (serverObj *Server) OnGetCrossShard(_ *peer.PeerConn, msg *wire.MessageGetCrossShard) {
	Logger.log.Debug("Receive a getcrossshard START")
	var txProcessed chan struct{}
	serverObj.netSync.QueueMessage(nil, msg, txProcessed)
	Logger.log.Debug("Receive a getcrossshard END")
}

func (serverObj *Server) OnGetShardToBeacon(_ *peer.PeerConn, msg *wire.MessageGetShardToBeacon) {
	Logger.log.Debug("Receive a getshardtobeacon START")
	var txProcessed chan struct{}
	serverObj.netSync.QueueMessage(nil, msg, txProcessed)
	Logger.log.Debug("Receive a getshardtobeacon END")
}

// OnTx is invoked when a peer receives a tx message.  It blocks
// until the transaction has been fully processed.  Unlock the block
// handler this does not serialize all transactions through a single thread
// transactions don't rely on the previous one in a linear fashion like blocks.
func (serverObj *Server) OnTx(peer *peer.PeerConn, msg *wire.MessageTx) {
	Logger.log.Debug("Receive a new transaction START")
	var txProcessed chan struct{}
	serverObj.netSync.QueueTx(nil, msg, txProcessed)
	//<-txProcessed

	Logger.log.Debug("Receive a new transaction END")
}

func (serverObj *Server) OnTxToken(peer *peer.PeerConn, msg *wire.MessageTxToken) {
	Logger.log.Debug("Receive a new transaction(normal token) START")
	var txProcessed chan struct{}
	serverObj.netSync.QueueTxToken(nil, msg, txProcessed)
	//<-txProcessed

	Logger.log.Debug("Receive a new transaction(normal token) END")
}

func (serverObj *Server) OnTxPrivacyToken(peer *peer.PeerConn, msg *wire.MessageTxPrivacyToken) {
	Logger.log.Debug("Receive a new transaction(privacy token) START")
	var txProcessed chan struct{}
	serverObj.netSync.QueueTxPrivacyToken(nil, msg, txProcessed)
	//<-txProcessed

	Logger.log.Debug("Receive a new transaction(privacy token) END")
}

/*
// OnVersion is invoked when a peer receives a version message
// and is used to negotiate the protocol version details as well as kick start
// the communications.
*/
func (serverObj *Server) OnVersion(peerConn *peer.PeerConn, msg *wire.MessageVersion) {
	Logger.log.Debug("Receive version message START")

	pbk := ""
	if msg.PublicKey != "" {
		err := cashec.ValidateDataB58(msg.PublicKey, msg.SignDataB58, []byte(peerConn.ListenerPeer.PeerID.Pretty()))
		if err == nil {
			pbk = msg.PublicKey
		} else {
			peerConn.ForceClose()
			return
		}
	}

	remotePeer := &peer.Peer{
		ListeningAddress: msg.LocalAddress,
		RawAddress:       msg.RawLocalAddress,
		PeerID:           msg.LocalPeerId,
		PublicKey:        pbk,
	}
	peerConn.RemotePeer.PublicKey = pbk

	serverObj.cNewPeers <- remotePeer
	valid := false
	if msg.ProtocolVersion == serverObj.protocolVersion {
		valid = true
	}

	// check for accept connection
	if !serverObj.connManager.CheckForAcceptConn(peerConn) {
		peerConn.ForceClose()
		return
	}

	msgV, err := wire.MakeEmptyMessage(wire.CmdVerack)
	if err != nil {
		return
	}

	msgV.(*wire.MessageVerAck).Valid = valid
	msgV.(*wire.MessageVerAck).Timestamp = time.Now()

	peerConn.QueueMessageWithEncoding(msgV, nil, peer.MESSAGE_TO_PEER, nil)

	//	push version message again
	if !peerConn.VerAckReceived() {
		err := serverObj.PushVersionMessage(peerConn)
		if err != nil {
			Logger.log.Error(err)
		}
	}

	Logger.log.Debug("Receive version message END")
}

/*
OnVerAck is invoked when a peer receives a version acknowlege message
*/
func (serverObj *Server) OnVerAck(peerConn *peer.PeerConn, msg *wire.MessageVerAck) {
	Logger.log.Debug("Receive verack message START")

	if msg.Valid {
		peerConn.VerValid = true

		if peerConn.GetIsOutbound() {
			serverObj.addrManager.Good(peerConn.RemotePeer)
		}

		// send message for get addr
		msgSG, err := wire.MakeEmptyMessage(wire.CmdGetAddr)
		if err != nil {
			return
		}
		var dc chan<- struct{}
		peerConn.QueueMessageWithEncoding(msgSG, dc, peer.MESSAGE_TO_PEER, nil)

		//	broadcast addr to all peer
		listen := serverObj.connManager.ListeningPeer
		msgSA, err := wire.MakeEmptyMessage(wire.CmdAddr)
		if err != nil {
			return
		}

		rawPeers := []wire.RawPeer{}
		peers := serverObj.addrManager.AddressCache()
		for _, peer := range peers {
			if peerConn.RemotePeerID.Pretty() != serverObj.connManager.GetPeerId(peer.RawAddress) {
				rawPeers = append(rawPeers, wire.RawPeer{peer.RawAddress, peer.PublicKey})
			}
		}
		msgSA.(*wire.MessageAddr).RawPeers = rawPeers
		var doneChan chan<- struct{}
		for _, _peerConn := range listen.PeerConns {
			go _peerConn.QueueMessageWithEncoding(msgSA, doneChan, peer.MESSAGE_TO_PEER, nil)
		}

		// send message get blocks

		//msgNew, err := wire.MakeEmptyMessage(wire.CmdGetBlocks)
		//msgNew.(*wire.MessageGetBlocks).LastBlockHash = *serverObj.blockChain.BestState.BestBlockHash
		//println(peerConn.ListenerPeer.PeerId.String())
		//msgNew.(*wire.MessageGetBlocks).SenderID = peerConn.ListenerPeer.PeerId.String()
		//if err != nil {
		//	return
		//}
		//peerConn.QueueMessageWithEncoding(msgNew, nil)
	} else {
		peerConn.VerValid = true
	}

	Logger.log.Debug("Receive verack message END")
}

func (serverObj *Server) OnGetAddr(peerConn *peer.PeerConn, msg *wire.MessageGetAddr) {
	Logger.log.Debug("Receive getaddr message START")

	// send message for addr
	msgS, err := wire.MakeEmptyMessage(wire.CmdAddr)
	if err != nil {
		return
	}

	peers := serverObj.addrManager.AddressCache()
	rawPeers := []wire.RawPeer{}
	for _, peer := range peers {
		if peerConn.RemotePeerID.Pretty() != serverObj.connManager.GetPeerId(peer.RawAddress) {
			rawPeers = append(rawPeers, wire.RawPeer{peer.RawAddress, peer.PublicKey})
		}
	}
	msgS.(*wire.MessageAddr).RawPeers = rawPeers
	var dc chan<- struct{}
	peerConn.QueueMessageWithEncoding(msgS, dc, peer.MESSAGE_TO_PEER, nil)

	Logger.log.Debug("Receive getaddr message END")
}

func (serverObj *Server) OnAddr(peerConn *peer.PeerConn, msg *wire.MessageAddr) {
	Logger.log.Debugf("Receive addr message %v", msg.RawPeers)
}

func (serverObj *Server) OnBFTMsg(_ *peer.PeerConn, msg wire.Message) {
	Logger.log.Debug("Receive a BFTMsg START")
	var txProcessed chan struct{}
	serverObj.netSync.QueueMessage(nil, msg, txProcessed)
	Logger.log.Debug("Receive a BFTMsg END")
}

func (serverObj *Server) OnPeerState(_ *peer.PeerConn, msg *wire.MessagePeerState) {
	Logger.log.Debug("Receive a peerstate START")
	var txProcessed chan struct{}
	serverObj.netSync.QueueMessage(nil, msg, txProcessed)
	Logger.log.Debug("Receive a peerstate END")
}

func (serverObj *Server) GetPeerIDsFromPublicKey(pubKey string) []libp2p.ID {
	result := []libp2p.ID{}

	listener := serverObj.connManager.Config.ListenerPeer
	for _, peerConn := range listener.PeerConns {
		// Logger.log.Debug("Test PeerConn", peerConn.RemotePeer.PaymentAddress)
		if peerConn.RemotePeer.PublicKey == pubKey {
			exist := false
			for _, item := range result {
				if item.Pretty() == peerConn.RemotePeer.PeerID.Pretty() {
					exist = true
				}
			}

			if !exist {
				result = append(result, peerConn.RemotePeer.PeerID)
			}
		}
	}

	return result
}

/*
PushMessageToAll broadcast msg
*/
func (serverObj *Server) PushMessageToAll(msg wire.Message) error {
	Logger.log.Debug("Push msg to all peers")
	var dc chan<- struct{}
	msg.SetSenderID(serverObj.connManager.Config.ListenerPeer.PeerID)
	serverObj.connManager.Config.ListenerPeer.QueueMessageWithEncoding(msg, dc, peer.MESSAGE_TO_ALL, nil)
	return nil
}

/*
PushMessageToPeer push msg to peer
*/
func (serverObj *Server) PushMessageToPeer(msg wire.Message, peerId libp2p.ID) error {
	Logger.log.Debugf("Push msg to peer %s", peerId.Pretty())
	var dc chan<- struct{}
	peerConn := serverObj.connManager.Config.ListenerPeer.GetPeerConnByPeerID(peerId.Pretty())
	if peerConn != nil {
		msg.SetSenderID(serverObj.connManager.Config.ListenerPeer.PeerID)
		peerConn.QueueMessageWithEncoding(msg, dc, peer.MESSAGE_TO_PEER, nil)
		Logger.log.Debugf("Pushed peer %s", peerId.Pretty())
		return nil
	} else {
		Logger.log.Error("RemotePeer not exist!")
	}
	return errors.New("RemotePeer not found")
}

/*
PushMessageToPeer push msg to pbk
*/
func (serverObj *Server) PushMessageToPbk(msg wire.Message, pbk string) error {
	Logger.log.Debugf("Push msg to pbk %s", pbk)
	peerConns := serverObj.connManager.GetPeerConnOfPbk(pbk)
	if len(peerConns) > 0 {
		for _, peerConn := range peerConns {
			msg.SetSenderID(peerConn.ListenerPeer.PeerID)
			peerConn.QueueMessageWithEncoding(msg, nil, peer.MESSAGE_TO_PEER, nil)
		}
		Logger.log.Debugf("Pushed pbk %s", pbk)
		return nil
	} else {
		Logger.log.Error("RemotePeer not exist!")
	}
	return errors.New("RemotePeer not found")
}

/*
PushMessageToPeer push msg to pbk
*/
func (serverObj *Server) PushMessageToShard(msg wire.Message, shard byte) error {
	Logger.log.Debugf("Push msg to shard %d", shard)
	peerConns := serverObj.connManager.GetPeerConnOfShard(shard)
	if len(peerConns) > 0 {
		for _, peerConn := range peerConns {
			msg.SetSenderID(peerConn.ListenerPeer.PeerID)
			peerConn.QueueMessageWithEncoding(msg, nil, peer.MESSAGE_TO_SHARD, &shard)
		}
		Logger.log.Debugf("Pushed shard %d", shard)
	} else {
		Logger.log.Error("RemotePeer of shard not exist!")
		listener := serverObj.connManager.Config.ListenerPeer
		listener.QueueMessageWithEncoding(msg, nil, peer.MESSAGE_TO_SHARD, &shard)
	}
	return nil
}

func (serverObj *Server) PushRawBytesToShard(p *peer.PeerConn, msgBytes *[]byte, shard byte) error {
	Logger.log.Debugf("Push raw bytes to shard %d", shard)
	peerConns := serverObj.connManager.GetPeerConnOfShard(shard)
	if len(peerConns) > 0 {
		for _, peerConn := range peerConns {
			if p == nil || peerConn != p {
				peerConn.QueueMessageWithBytes(msgBytes, nil)
			}
		}
		Logger.log.Debugf("Pushed shard %d", shard)
	} else {
		Logger.log.Error("RemotePeer of shard not exist!")
		peerConns := serverObj.connManager.GetPeerConnOfAll()
		for _, peerConn := range peerConns {
			if p == nil || peerConn != p {
				peerConn.QueueMessageWithBytes(msgBytes, nil)
			}
		}
	}
	return nil
}

/*
PushMessageToPeer push msg to beacon node
*/
func (serverObj *Server) PushMessageToBeacon(msg wire.Message) error {
	Logger.log.Debugf("Push msg to beacon")
	peerConns := serverObj.connManager.GetPeerConnOfBeacon()
	if len(peerConns) > 0 {
		fmt.Println("BFT:", len(peerConns))
		for _, peerConn := range peerConns {
			msg.SetSenderID(peerConn.ListenerPeer.PeerID)
			peerConn.QueueMessageWithEncoding(msg, nil, peer.MESSAGE_TO_BEACON, nil)
		}
		Logger.log.Debugf("Pushed beacon done")
		return nil
	} else {
		Logger.log.Error("RemotePeer of beacon not exist!")
		listener := serverObj.connManager.Config.ListenerPeer
		listener.QueueMessageWithEncoding(msg, nil, peer.MESSAGE_TO_BEACON, nil)
	}
	return errors.New("RemotePeer of beacon not found")
}

func (serverObj *Server) PushRawBytesToBeacon(p *peer.PeerConn, msgBytes *[]byte) error {
	Logger.log.Debugf("Push raw bytes to beacon")
	peerConns := serverObj.connManager.GetPeerConnOfBeacon()
	if len(peerConns) > 0 {
		for _, peerConn := range peerConns {
			if p == nil || peerConn != p {
				peerConn.QueueMessageWithBytes(msgBytes, nil)
			}
		}
		Logger.log.Debugf("Pushed raw bytes beacon done")
	} else {
		Logger.log.Error("RemotePeer of beacon raw bytes not exist!")
		peerConns := serverObj.connManager.GetPeerConnOfAll()
		for _, peerConn := range peerConns {
			if p == nil || peerConn != p {
				peerConn.QueueMessageWithBytes(msgBytes, nil)
			}
		}
	}
	return nil
}

// handleAddPeerMsg deals with adding new peers.  It is invoked from the
// peerHandler goroutine.
func (serverObj *Server) handleAddPeerMsg(peer *peer.Peer) bool {
	if peer == nil {
		return false
	}
	Logger.log.Debug("Zero peer have just sent a message version")
	//Logger.log.Debug(peer)
	return true
}

func (serverObj *Server) PushVersionMessage(peerConn *peer.PeerConn) error {
	// push message version
	msg, err := wire.MakeEmptyMessage(wire.CmdVersion)
	msg.(*wire.MessageVersion).Timestamp = time.Now().UnixNano()
	msg.(*wire.MessageVersion).LocalAddress = peerConn.ListenerPeer.ListeningAddress
	msg.(*wire.MessageVersion).RawLocalAddress = peerConn.ListenerPeer.RawAddress
	msg.(*wire.MessageVersion).LocalPeerId = peerConn.ListenerPeer.PeerID
	msg.(*wire.MessageVersion).RemoteAddress = peerConn.ListenerPeer.ListeningAddress
	msg.(*wire.MessageVersion).RawRemoteAddress = peerConn.ListenerPeer.RawAddress
	msg.(*wire.MessageVersion).RemotePeerId = peerConn.ListenerPeer.PeerID
	msg.(*wire.MessageVersion).ProtocolVersion = serverObj.protocolVersion

	// ValidateTransaction Public Key from ProducerPrvKey
	if peerConn.ListenerPeer.Config.UserKeySet != nil {
		msg.(*wire.MessageVersion).PublicKey = peerConn.ListenerPeer.Config.UserKeySet.GetPublicKeyB58()
		signDataB58, err := peerConn.ListenerPeer.Config.UserKeySet.SignDataB58([]byte(peerConn.RemotePeer.PeerID.Pretty()))
		if err == nil {
			msg.(*wire.MessageVersion).SignDataB58 = signDataB58
		}
	}
	if err != nil {
		return err
	}
	peerConn.QueueMessageWithEncoding(msg, nil, peer.MESSAGE_TO_PEER, nil)
	return nil
}

func (serverObj *Server) GetCurrentRoleShard() (string, *byte) {
	return serverObj.connManager.GetCurrentRoleShard()
}

func (serverObj *Server) UpdateConsensusState(role string, userPbk string, currentShard *byte, beaconCommittee []string, shardCommittee map[byte][]string) {
	serverObj.connManager.UpdateConsensusState(role, userPbk, currentShard, beaconCommittee, shardCommittee)
}

func (serverObj *Server) PushMessageGetBlockBeaconByHeight(from uint64, to uint64, peerID libp2p.ID) error {
	msg, err := wire.MakeEmptyMessage(wire.CmdGetBlockBeacon)
	if err != nil {
		return err
	}
	msg.(*wire.MessageGetBlockBeacon).BlkHeights = append(msg.(*wire.MessageGetBlockBeacon).BlkHeights, from)
	msg.(*wire.MessageGetBlockBeacon).BlkHeights = append(msg.(*wire.MessageGetBlockBeacon).BlkHeights, to)
	if peerID != "" {
		return serverObj.PushMessageToPeer(msg, peerID)
	}
	return serverObj.PushMessageToAll(msg)
}

func (serverObj *Server) PushMessageGetBlockBeaconByHash(blkHashes []common.Hash, getFromPool bool, peerID libp2p.ID) error {
	msg, err := wire.MakeEmptyMessage(wire.CmdGetBlockBeacon)
	if err != nil {
		return err
	}
	msg.(*wire.MessageGetBlockBeacon).ByHash = true
	msg.(*wire.MessageGetBlockBeacon).FromPool = getFromPool
	msg.(*wire.MessageGetBlockBeacon).BlkHashes = blkHashes
	if peerID != "" {
		return serverObj.PushMessageToPeer(msg, peerID)
	}
	return serverObj.PushMessageToAll(msg)
}

func (serverObj *Server) PushMessageGetBlockShardByHeight(shardID byte, from uint64, to uint64, peerID libp2p.ID) error {
	msg, err := wire.MakeEmptyMessage(wire.CmdGetBlockShard)
	if err != nil {
		return err
	}
	msg.(*wire.MessageGetBlockShard).BlkHeights = append(msg.(*wire.MessageGetBlockShard).BlkHeights, from)
	msg.(*wire.MessageGetBlockShard).BlkHeights = append(msg.(*wire.MessageGetBlockShard).BlkHeights, to)
	msg.(*wire.MessageGetBlockShard).ShardID = shardID
	if peerID == "" {
		return serverObj.PushMessageToShard(msg, shardID)
	}
	return serverObj.PushMessageToPeer(msg, peerID)

}

func (serverObj *Server) PushMessageGetBlockShardByHash(shardID byte, blksHash []common.Hash, getFromPool bool, peerID libp2p.ID) error {
	msg, err := wire.MakeEmptyMessage(wire.CmdGetBlockShard)
	if err != nil {
		return err
	}
	msg.(*wire.MessageGetBlockShard).ByHash = true
	msg.(*wire.MessageGetBlockShard).FromPool = getFromPool
	msg.(*wire.MessageGetBlockShard).BlksHash = blksHash
	msg.(*wire.MessageGetBlockShard).ShardID = shardID
	if peerID == "" {
		return serverObj.PushMessageToShard(msg, shardID)
	}
	return serverObj.PushMessageToPeer(msg, peerID)

}

func (serverObj *Server) PushMessageGetBlockShardToBeaconByHeight(shardID byte, from uint64, to uint64, peerID libp2p.ID) error {
	Logger.log.Debugf("Send a GetShardToBeacon")
	listener := serverObj.connManager.Config.ListenerPeer
	msg, err := wire.MakeEmptyMessage(wire.CmdGetShardToBeacon)
	if err != nil {
		return err
	}
	msg.(*wire.MessageGetShardToBeacon).ShardID = shardID
	msg.(*wire.MessageGetShardToBeacon).BlkHeights = append(msg.(*wire.MessageGetShardToBeacon).BlkHeights, from)
	msg.(*wire.MessageGetShardToBeacon).BlkHeights = append(msg.(*wire.MessageGetShardToBeacon).BlkHeights, to)
	msg.(*wire.MessageGetShardToBeacon).Timestamp = time.Now().Unix()
	msg.SetSenderID(listener.PeerID)
	Logger.log.Debugf("Send a GetCrossShard from %s", listener.RawAddress)
	if peerID == "" {
		return serverObj.PushMessageToShard(msg, shardID)
	}
	return serverObj.PushMessageToPeer(msg, peerID)

}

func (serverObj *Server) PushMessageGetBlockShardToBeaconByHash(shardID byte, blkHashes []common.Hash, getFromPool bool, peerID libp2p.ID) error {
	Logger.log.Debugf("Send a GetShardToBeacon")
	listener := serverObj.connManager.Config.ListenerPeer
	msg, err := wire.MakeEmptyMessage(wire.CmdGetShardToBeacon)
	if err != nil {
		return err
	}
	msg.(*wire.MessageGetShardToBeacon).ByHash = true
	msg.(*wire.MessageGetShardToBeacon).FromPool = getFromPool
	msg.(*wire.MessageGetShardToBeacon).ShardID = shardID
	msg.(*wire.MessageGetShardToBeacon).BlkHashes = blkHashes
	msg.(*wire.MessageGetShardToBeacon).Timestamp = time.Now().Unix()
	msg.SetSenderID(listener.PeerID)
	Logger.log.Debugf("Send a GetCrossShard from %s", listener.RawAddress)
	if peerID == "" {
		return serverObj.PushMessageToShard(msg, shardID)
	}
	return serverObj.PushMessageToPeer(msg, peerID)
}

func (serverObj *Server) PushMessageGetBlockShardToBeaconBySpecificHeight(shardID byte, blkHeights []uint64, getFromPool bool, peerID libp2p.ID) error {
	Logger.log.Debugf("Send a GetShardToBeacon")
	listener := serverObj.connManager.Config.ListenerPeer
	msg, err := wire.MakeEmptyMessage(wire.CmdGetShardToBeacon)
	if err != nil {
		return err
	}
	msg.(*wire.MessageGetShardToBeacon).BySpecificHeight = true
	msg.(*wire.MessageGetShardToBeacon).FromPool = getFromPool
	msg.(*wire.MessageGetShardToBeacon).ShardID = shardID
	msg.(*wire.MessageGetShardToBeacon).BlkHeights = blkHeights
	msg.(*wire.MessageGetShardToBeacon).Timestamp = time.Now().Unix()
	msg.SetSenderID(listener.PeerID)
	Logger.log.Debugf("Send a GetShardToBeacon from %s", listener.RawAddress)
	if peerID == "" {
		return serverObj.PushMessageToShard(msg, shardID)
	}
	return serverObj.PushMessageToPeer(msg, peerID)
}

func (serverObj *Server) PushMessageGetBlockCrossShardByHash(fromShard byte, toShard byte, blkHashes []common.Hash, getFromPool bool, peerID libp2p.ID) error {
	Logger.log.Debugf("Send a GetCrossShard")
	listener := serverObj.connManager.Config.ListenerPeer
	msg, err := wire.MakeEmptyMessage(wire.CmdGetCrossShard)
	if err != nil {
		return err
	}
	msg.(*wire.MessageGetCrossShard).ByHash = true
	msg.(*wire.MessageGetCrossShard).FromPool = getFromPool
	msg.(*wire.MessageGetCrossShard).FromShardID = fromShard
	msg.(*wire.MessageGetCrossShard).ToShardID = toShard
	msg.(*wire.MessageGetCrossShard).BlkHashes = blkHashes
	msg.(*wire.MessageGetCrossShard).Timestamp = time.Now().Unix()
	msg.SetSenderID(listener.PeerID)
	Logger.log.Debugf("Send a GetCrossShard from %s", listener.RawAddress)
	if peerID == "" {
		return serverObj.PushMessageToShard(msg, fromShard)
	}
	return serverObj.PushMessageToPeer(msg, peerID)

}

func (serverObj *Server) PushMessageGetBlockCrossShardBySpecificHeight(fromShard byte, toShard byte, blkHeights []uint64, getFromPool bool, peerID libp2p.ID) error {
	Logger.log.Debugf("Send a GetCrossShard")
	listener := serverObj.connManager.Config.ListenerPeer
	msg, err := wire.MakeEmptyMessage(wire.CmdGetCrossShard)
	if err != nil {
		return err
	}
	msg.(*wire.MessageGetCrossShard).FromPool = getFromPool
	msg.(*wire.MessageGetCrossShard).BySpecificHeight = true
	msg.(*wire.MessageGetCrossShard).FromShardID = fromShard
	msg.(*wire.MessageGetCrossShard).ToShardID = toShard
	msg.(*wire.MessageGetCrossShard).BlkHeights = blkHeights
	msg.(*wire.MessageGetCrossShard).Timestamp = time.Now().Unix()
	msg.SetSenderID(listener.PeerID)
	Logger.log.Debugf("Send a GetCrossShard from %s", listener.RawAddress)
	if peerID == "" {
		return serverObj.PushMessageToShard(msg, fromShard)
	}
	return serverObj.PushMessageToPeer(msg, peerID)
}

func (serverObj *Server) BoardcastNodeState() error {
	listener := serverObj.connManager.Config.ListenerPeer
	msg, err := wire.MakeEmptyMessage(wire.CmdPeerState)
	if err != nil {
		return err
	}
	msg.(*wire.MessagePeerState).Beacon = blockchain.ChainState{
		serverObj.blockChain.BestState.Beacon.BeaconHeight,
		serverObj.blockChain.BestState.Beacon.BestBlockHash,
		serverObj.blockChain.BestState.Beacon.Hash(),
	}
	for _, shardID := range serverObj.blockChain.GetCurrentSyncShards() {
		msg.(*wire.MessagePeerState).Shards[shardID] = blockchain.ChainState{
			serverObj.blockChain.BestState.Shard[shardID].ShardHeight,
			serverObj.blockChain.BestState.Shard[shardID].BestBlockHash,
			serverObj.blockChain.BestState.Shard[shardID].Hash(),
		}
	}
	msg.(*wire.MessagePeerState).ShardToBeaconPool = serverObj.shardToBeaconPool.GetValidPendingBlockHeight()
	if serverObj.userKeySet != nil {
		userRole, shardID := serverObj.blockChain.BestState.Beacon.GetPubkeyRole(serverObj.userKeySet.GetPublicKeyB58(), serverObj.blockChain.BestState.Beacon.BestBlock.Header.Round)
		if (cfg.NodeMode == common.NODEMODE_AUTO || cfg.NodeMode == common.NODEMODE_SHARD) && userRole == common.NODEMODE_SHARD {
			userRole = serverObj.blockChain.BestState.Shard[shardID].GetPubkeyRole(serverObj.userKeySet.GetPublicKeyB58(), serverObj.blockChain.BestState.Shard[shardID].BestBlock.Header.Round)
			if userRole == "shard-proposer" || userRole == "shard-validator" {
				msg.(*wire.MessagePeerState).CrossShardPool[shardID] = serverObj.crossShardPool[shardID].GetValidBlockHeight()
			}
		}
	}
	msg.SetSenderID(listener.PeerID)
	Logger.log.Debugf("Boardcast peerstate from %s", listener.RawAddress)
	serverObj.PushMessageToAll(msg)
	return nil
}
