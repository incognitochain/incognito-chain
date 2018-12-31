package main

import (
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	libp2p "github.com/libp2p/go-libp2p-peer"
	"github.com/ninjadotorg/constant/addrmanager"
	"github.com/ninjadotorg/constant/blockchain"
	"github.com/ninjadotorg/constant/cashec"
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/connmanager"
	"github.com/ninjadotorg/constant/consensus/constantpos"
	"github.com/ninjadotorg/constant/database"
	"github.com/ninjadotorg/constant/mempool"
	"github.com/ninjadotorg/constant/netsync"
	"github.com/ninjadotorg/constant/peer"
	"github.com/ninjadotorg/constant/rewardagent"
	"github.com/ninjadotorg/constant/wallet"
	"github.com/ninjadotorg/constant/wire"
)

type Server struct {
	started     int32
	startupTime int64

	protocolVersion string
	chainParams     *blockchain.Params
	connManager     *connmanager.ConnManager
	blockChain      *blockchain.BlockChain
	dataBase        database.DatabaseInterface
	// rpcServer       *rpcserver.RpcServer
	memPool           *mempool.TxPool
	shardToBeaconPool *mempool.ShardToBeaconPool
	waitGroup         sync.WaitGroup
	netSync           *netsync.NetSync
	addrManager       *addrmanager.AddrManager
	wallet            *wallet.Wallet
	consensusEngine   *constantpos.Engine
	blockgen          *blockchain.BlkTmplGenerator
	rewardAgent       *rewardagent.RewardAgent
	// The fee estimator keeps track of how long transactions are left in
	// the mempool before they are mined into blocks.
	feeEstimator map[byte]*mempool.FeeEstimator

	cQuit     chan struct{}
	cNewPeers chan *peer.Peer
}

// setupRPCListeners returns a slice of listeners that are configured for use
// with the RPC server depending on the configuration settings for listen
// addresses and TLS.
func (self *Server) setupRPCListeners() ([]net.Listener, error) {
	// Setup TLS if not disabled.
	listenFunc := net.Listen
	if !cfg.DisableTLS {
		Logger.log.Info("Disable TLS for RPC is false")
		// Generate the TLS cert and key file if both don't already
		// exist.
		// if !fileExists(cfg.RPCKey) && !fileExists(cfg.RPCCert) {
		// 	err := rpcserver.GenCertPair(cfg.RPCCert, cfg.RPCKey)
		// 	if err != nil {
		// 		return nil, err
		// 	}
		// }
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
		Logger.log.Info("Disable TLS for RPC is true")
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
func (self *Server) NewServer(listenAddrs []string, db database.DatabaseInterface, chainParams *blockchain.Params, protocolVer string, interrupt <-chan struct{}) error {
	// Init data for Server
	self.protocolVersion = protocolVer
	self.chainParams = chainParams
	self.cQuit = make(chan struct{})
	self.cNewPeers = make(chan *peer.Peer)
	self.dataBase = db

	var err error

	// Create a new block chain instance with the appropriate configuration.9
	// if cfg.Light {
	// 	if self.wallet == nil {
	// 		return errors.New("Wallet NOT FOUND. Light Mode required Wallet with at least one child account")
	// 	}
	// 	if len(self.wallet.MasterAccount.Child) < 1 {
	// 		return errors.New("No child account in wallet. Light Mode required Wallet with at least one child account")
	// 	}
	// }
	userKeySet, err := cfg.GetUserKeySet()
	if err != nil {
		Logger.log.Error(err)
		return err
	}
	self.shardToBeaconPool = &mempool.ShardToBeaconPool{}
	self.blockChain = &blockchain.BlockChain{}
	err = self.blockChain.Init(&blockchain.Config{
		ChainParams: self.chainParams,
		DataBase:    self.dataBase,
		Interrupt:   interrupt,
		// Light:       cfg.Light,
		Wallet:            self.wallet,
		ShardToBeaconPool: self.shardToBeaconPool,
	})
	if err != nil {
		return err
	}

	// Search for a feeEstimator state in the database. If none can be found
	// or if it cannot be loaded, create a new one.
	// if false {
	// 	Logger.log.Info("Load chain dependencies from DB")
	// 	self.feeEstimator = make(map[byte]*mempool.FeeEstimator)
	// 	for _, bestState := range self.blockChain.BestState.Shard {
	// 		shardID := bestState.BestBlock.Header.ShardID
	// 		feeEstimatorData, err := self.dataBase.GetFeeEstimator(shardID)
	// 		if err == nil && len(feeEstimatorData) > 0 {
	// 			feeEstimator, err := mempool.RestoreFeeEstimator(feeEstimatorData)
	// 			if err != nil {
	// 				Logger.log.Errorf("Failed to restore fee estimator %v", err)
	// 				Logger.log.Info("Init NewFeeEstimator")
	// 				self.feeEstimator[shardID] = mempool.NewFeeEstimator(
	// 					mempool.DefaultEstimateFeeMaxRollback,
	// 					mempool.DefaultEstimateFeeMinRegisteredBlocks)
	// 			} else {
	// 				self.feeEstimator[shardID] = feeEstimator
	// 			}
	// 		}
	// 	}
	// } else {
	// 	err := self.dataBase.CleanCommitments()
	// 	if err != nil {
	// 		Logger.log.Error(err)
	// 		return err
	// 	}
	// 	err = self.dataBase.CleanSerialNumbers()
	// 	if err != nil {
	// 		Logger.log.Error(err)
	// 		return err
	// 	}
	// 	err = self.dataBase.CleanFeeEstimator()
	// 	if err != nil {
	// 		Logger.log.Error(err)
	// 		return err
	// 	}

	// 	self.feeEstimator = make(map[byte]*mempool.FeeEstimator)
	// }
	self.feeEstimator = make(map[byte]*mempool.FeeEstimator)
	// create mempool tx
	self.memPool = &mempool.TxPool{}
	self.memPool.Init(&mempool.Config{
		BlockChain:   self.blockChain,
		DataBase:     self.dataBase,
		ChainParams:  chainParams,
		FeeEstimator: self.feeEstimator,
	})
	self.addrManager = addrmanager.New(cfg.DataDir)

	self.rewardAgent, err = rewardagent.RewardAgent{}.Init(&rewardagent.RewardAgentConfig{
		BlockChain: self.blockChain,
	})
	if err != nil {
		return err
	}

	self.blockgen, err = blockchain.BlkTmplGenerator{}.Init(self.memPool, self.blockChain, self.rewardAgent, self.shardToBeaconPool)
	if err != nil {
		return err
	}

	self.consensusEngine, err = constantpos.Engine{}.Init(&constantpos.EngineConfig{
		ChainParams: self.chainParams,
		BlockChain:  self.blockChain,
		ConnManager: self.connManager,
		Server:      self,
		BlockGen:    self.blockgen,
		NodeMode:    cfg.NodeMode,
		UserKeySet:  userKeySet,
	})
	if err != nil {
		return err
	}
	// Init Net Sync manager to process messages
	self.netSync = netsync.NetSync{}.New(&netsync.NetSyncConfig{
		BlockChain:        self.blockChain,
		ChainParam:        chainParams,
		MemTxPool:         self.memPool,
		Server:            self,
		Consensus:         self.consensusEngine,
		FeeEstimator:      self.feeEstimator,
		ShardToBeaconPool: self.shardToBeaconPool,
	})
	// Create a connection manager.
	var peers []*peer.Peer
	if !cfg.DisableListen {
		var err error
		peers, err = self.InitListenerPeers(self.addrManager, listenAddrs, cfg.MaxPeers, cfg.MaxOutPeers, cfg.MaxInPeers)
		if err != nil {
			Logger.log.Error(err)
			return err
		}
	}
	connManager := connmanager.ConnManager{}.New(&connmanager.Config{
		OnInboundAccept:      self.InboundPeerConnected,
		OnOutboundConnection: self.OutboundPeerConnected,
		ListenerPeers:        peers,
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

	// for testing
	//currentShard := byte(0)
	//userPbk := ""
	//ks, err := cfg.GetUserKeySet()
	//if err == nil {
	//	userPbk = ks.GetPublicKeyB58()
	//}
	//shardCommittee := map[byte][]string{
	//	byte(0): []string{"15Z7uGSzG4ZR5ENDnBE6PuGcquNGYj7PqPFj4ojEEGk8QQNZoN6"},
	//	byte(1): []string{"15CfJ8vH78zw8PT2FbBeNssFWcHMW1sSxoJ6RKv2hZ6nztp4mCQ"},
	//	byte(2): []string{"17PnJ3sjHvFLp3Sck12FaHfvk4AghGctecTG54bdLNFVGygi8DN"},
	//	byte(3): []string{"17qiWdX7ubTHpVu5eMDxMCCwesYYcLWKE1KTP62LQK3ALrQ6A5T"},
	//}
	//beaconCommittee := []string{
	//	"15yDGFUwf5r7rZcfEzEmpcNvMfC5zi1g454AeHMZNSGEiBFacvt",
	//	"16C6356Xst2bKnAuXYM3Ezfz7ZwG9kiKmHAPTFMupQs3wzQfaoM",
	//}
	//connManager.UpdateConsensusState("", userPbk, &currentShard, beaconCommittee, shardCommittee)

	self.connManager = connManager
	// Start up persistent peers.
	permanentPeers := cfg.ConnectPeers
	if len(permanentPeers) == 0 {
		permanentPeers = cfg.AddPeers
	}

	for _, addr := range permanentPeers {
		go self.connManager.Connect(addr, "")
	}

	if !cfg.DisableRPC {
		// Setup listeners for the configured RPC listen addresses and
		// TLS settings.
		rpcListeners, err := self.setupRPCListeners()
		if err != nil {
			return err
		}
		if len(rpcListeners) == 0 {
			return errors.New("RPCS: No valid listen address")
		}

		// rpcConfig := rpcserver.RpcServerConfig{
		// 	Listenters:    rpcListeners,
		// 	RPCQuirks:     cfg.RPCQuirks,
		// 	RPCMaxClients: cfg.RPCMaxClients,
		// 	ChainParams:   chainParams,
		// 	BlockChain:    self.blockChain,
		// 	TxMemPool:     self.memPool,
		// 	Server:        self,
		// 	Wallet:        self.wallet,
		// 	ConnMgr:       self.connManager,
		// 	AddrMgr:       self.addrManager,
		// 	RPCUser:       cfg.RPCUser,
		// 	RPCPass:       cfg.RPCPass,
		// 	RPCLimitUser:  cfg.RPCLimitUser,
		// 	RPCLimitPass:  cfg.RPCLimitPass,
		// 	DisableAuth:   cfg.RPCDisableAuth,
		// 	// IsGenerateNode:  cfg.Generate,
		// 	NodeRole:        cfg.NodeRole,
		// 	FeeEstimator:    self.feeEstimator,
		// 	ProtocolVersion: self.protocolVersion,
		// 	Database:        &self.dataBase,
		// }
		// self.rpcServer = &rpcserver.RpcServer{}
		// self.rpcServer.Init(&rpcConfig)

		// // Signal process shutdown when the RPC server requests it.
		// go func() {
		// 	<-self.rpcServer.RequestedProcessShutdown()
		// 	shutdownRequestChannel <- struct{}{}
		// }()
	}

	return nil
}

/*
// InboundPeerConnected is invoked by the connection manager when a new
// inbound connection is established.
*/
func (self *Server) InboundPeerConnected(peerConn *peer.PeerConn) {
	Logger.log.Info("inbound connected")
}

/*
// outboundPeerConnected is invoked by the connection manager when a new
// outbound connection is established.  It initializes a new outbound server
// peer instance, associates it with the relevant state such as the connection
// request instance and the connection itself, and finally notifies the address
// manager of the attempt.
*/
func (self *Server) OutboundPeerConnected(peerConn *peer.PeerConn) {
	Logger.log.Info("Outbound PEER connected with PEER Id - " + peerConn.RemotePeerID.Pretty())
	err := self.PushVersionMessage(peerConn)
	if err != nil {
		Logger.log.Error(err)
	}
}

/*
// WaitForShutdown blocks until the main listener and peer handlers are stopped.
*/
func (self *Server) WaitForShutdown() {
	self.waitGroup.Wait()
}

/*
// Stop gracefully shuts down the connection manager.
*/
func (self *Server) Stop() error {
	// stop connManager
	self.connManager.Stop()

	// Shutdown the RPC server if it's not disabled.
	// if !cfg.DisableRPC && self.rpcServer != nil {
	// 	self.rpcServer.Stop()
	// }

	// Save fee estimator in the db
	// for shardID, feeEstimator := range self.feeEstimator {
	// 	feeEstimatorData := feeEstimator.Save()
	// 	if len(feeEstimatorData) > 0 {
	// 		err := self.dataBase.StoreFeeEstimator(feeEstimatorData, shardID)
	// 		if err != nil {
	// 			Logger.log.Errorf("Can't save fee estimator data on chain #%d: %v", shardID, err)
	// 		} else {
	// 			Logger.log.Infof("Save fee estimator data on chain #%d", shardID)
	// 		}
	// 	}
	// }

	self.consensusEngine.Stop()

	// Signal the remaining goroutines to cQuit.
	close(self.cQuit)
	return nil
}

/*
// peerHandler is used to handle peer operations such as adding and removing
// peers to and from the server, banning peers, and broadcasting messages to
// peers.  It must be run in a goroutine.
*/
func (self *Server) peerHandler() {
	// Start the address manager and sync manager, both of which are needed
	// by peers.  This is done here since their lifecycle is closely tied
	// to this handler and rather than adding more channels to sychronize
	// things, it's easier and slightly faster to simply start and stop them
	// in this handler.
	self.addrManager.Start()
	self.netSync.Start()

	Logger.log.Info("Start peer handler")

	if len(cfg.ConnectPeers) == 0 {
		for _, addr := range self.addrManager.AddressCache() {
			go self.connManager.Connect(addr.RawAddress, addr.PublicKey)
		}
	}

	go self.connManager.Start(cfg.DiscoverPeersAddress)

out:
	for {
		select {
		case p := <-self.cNewPeers:
			self.handleAddPeerMsg(p)
		case <-self.cQuit:
			{
				break out
			}
		}
	}
	self.netSync.Stop()
	self.addrManager.Stop()
	self.connManager.Stop()
}

/*
// Start begins accepting connections from peers.
*/
func (self Server) Start() {
	// Already started?
	if atomic.AddInt32(&self.started, 1) != 1 {
		return
	}

	Logger.log.Info("Starting server")
	if cfg.TestNet {
		Logger.log.Critical("************************")
		Logger.log.Critical("* Testnet is active *")
		Logger.log.Critical("************************")
	}
	// Server startup time. Used for the uptime command for uptime calculation.
	self.startupTime = time.Now().Unix()

	// Start the peer handler which in turn starts the address and block
	// managers.
	self.waitGroup.Add(1)

	go self.peerHandler()
	// if !cfg.DisableRPC && self.rpcServer != nil {
	// 	self.waitGroup.Add(1)

	// 	// Start the rebroadcastHandler, which ensures user tx received by
	// 	// the RPC server are rebroadcast until being included in a block.
	// 	//go self.rebroadcastHandler()

	// 	self.rpcServer.Start()
	// }

	err := self.consensusEngine.Start()
	if err != nil {
		Logger.log.Error(err)
		go self.Stop()
		return
	}

}

/*
// initListeners initializes the configured net listeners and adds any bound
// addresses to the address manager. Returns the listeners and a NAT interface,
// which is non-nil if UPnP is in use.
*/
func (self *Server) InitListenerPeers(amgr *addrmanager.AddrManager, listenAddrs []string, maxPeers int, maxOutPeers int, maxInPeers int) ([]*peer.Peer, error) {
	netAddrs, err := common.ParseListeners(listenAddrs, "ip")
	if err != nil {
		return nil, err
	}

	// use keycache to save listener peer into file, this will make peer id of listener not change after turn off node
	kc := KeyCache{}
	kc.Load(filepath.Join(cfg.DataDir, "listenerpeer.json"))

	peers := make([]*peer.Peer, 0, len(netAddrs))
	for _, addr := range netAddrs {
		// load seed of libp2p from keycache file, if not exist -> save a new data into keycache file
		seed := int64(0)
		seedC, _ := strconv.ParseInt(os.Getenv("NODE_SEED"), 10, 64)
		if seedC == 0 {
			key := fmt.Sprintf("%s_seed", addr.String())
			seedT := kc.Get(key)
			if seedT == nil {
				seed = time.Now().UnixNano()
				kc.Set(key, seed)
			} else {
				seed = int64(seedT.(float64))
			}
		} else {
			seed = seedC
		}

		peer, err := peer.Peer{
			Seed:             seed,
			ListeningAddress: addr,
			Config:           *self.NewPeerConfig(),
			PeerConns:        make(map[string]*peer.PeerConn),
			PendingPeers:     make(map[string]*peer.Peer),
		}.NewPeer()
		peer.Config.MaxInPeers = maxInPeers
		peer.Config.MaxOutPeers = maxOutPeers
		peer.Config.MaxPeers = maxPeers
		if err != nil {
			return nil, err
		}
		peers = append(peers, peer)
	}

	kc.Save()

	return peers, nil
}

/*
// newPeerConfig returns the configuration for the listening RemotePeer.
*/
func (self *Server) NewPeerConfig() *peer.Config {
	KeySetUser, err := cfg.GetUserKeySet()
	if err != nil {
		Logger.log.Critical(err)
	}
	config := &peer.Config{
		MessageListeners: peer.MessageListeners{
			OnBlockShard:    self.OnBlockShard,
			OnBlockBeacon:   self.OnBlockBeacon,
			OnCrossShard:    self.OnCrossShard,
			OnShardToBeacon: self.OnShardToBeacon,
			OnTx:            self.OnTx,
			OnVersion:       self.OnVersion,
			OnGetBlocks:     self.OnGetBlocks,
			OnVerAck:        self.OnVerAck,
			OnGetAddr:       self.OnGetAddr,
			OnAddr:          self.OnAddr,

			//constantpos
			OnBFTPropose: self.OnBFTPropose,
			OnBFTPrepare: self.OnBFTPrepare,
			OnBFTCommit:  self.OnBFTCommit,
			OnBFTReply:   self.OnBFTReply,
			// OnInvalidBlock:  self.OnInvalidBlock,
			OnGetChainState: self.OnGetChainState,
			OnChainState:    self.OnChainState,
			//
			//OnRegistration: self.OnRegistration,
			// OnSwapRequest: self.OnSwapRequest,
			// OnSwapSig:     self.OnSwapSig,
			// OnSwapUpdate:  self.OnSwapUpdate,
			PushRawBytesToShard:  self.PushRawBytesToShard,
			PushRawBytesToBeacon: self.PushRawBytesToBeacon,
			GetCurrentRoleShard:  self.GetCurrentRoleShard,
		},
	}
	if len(KeySetUser.PrivateKey) != 0 {
		config.UserKeySet = KeySetUser
	}
	return config
}

// OnBlock is invoked when a peer receives a block message.  It
// blocks until the coin block has been fully processed.
func (self *Server) OnBlockShard(p *peer.PeerConn,
	msg *wire.MessageBlockShard) {
	Logger.log.Info("Receive a new blockshard START")

	var txProcessed chan struct{}
	self.netSync.QueueBlock(nil, msg, txProcessed)
	//<-txProcessed

	Logger.log.Info("Receive a new blockshard END")
}

func (self *Server) OnBlockBeacon(p *peer.PeerConn,
	msg *wire.MessageBlockBeacon) {
	Logger.log.Info("Receive a new blockbeacon START")

	var txProcessed chan struct{}
	self.netSync.QueueBlock(nil, msg, txProcessed)
	//<-txProcessed

	Logger.log.Info("Receive a new blockbeacon END")
}

func (self *Server) OnCrossShard(p *peer.PeerConn,
	msg *wire.MessageCrossShard) {
	Logger.log.Info("Receive a new crossshard START")

	var txProcessed chan struct{}
	self.netSync.QueueBlock(nil, msg, txProcessed)
	//<-txProcessed

	Logger.log.Info("Receive a new crossshard END")
}

func (self *Server) OnShardToBeacon(p *peer.PeerConn,
	msg *wire.MessageShardToBeacon) {
	Logger.log.Info("Receive a new shardToBeacon START")

	var txProcessed chan struct{}
	self.netSync.QueueBlock(nil, msg, txProcessed)
	//<-txProcessed

	Logger.log.Info("Receive a new shardToBeacon END")
}

func (self *Server) OnGetBlocks(_ *peer.PeerConn, msg *wire.MessageGetBlocks) {
	Logger.log.Info("Receive a " + msg.MessageType() + " message START")
	var txProcessed chan struct{}
	self.netSync.QueueGetBlock(nil, msg, txProcessed)
	//<-txProcessed

	Logger.log.Info("Receive a " + msg.MessageType() + " message END")
}

// OnTx is invoked when a peer receives a tx message.  It blocks
// until the transaction has been fully processed.  Unlock the block
// handler this does not serialize all transactions through a single thread
// transactions don't rely on the previous one in a linear fashion like blocks.
func (self *Server) OnTx(peer *peer.PeerConn, msg *wire.MessageTx) {
	Logger.log.Info("Receive a new transaction START")
	var txProcessed chan struct{}
	self.netSync.QueueTx(nil, msg, txProcessed)
	//<-txProcessed

	Logger.log.Info("Receive a new transaction END")
}

/*func (self *Server) OnRegistration(peer *peer.PeerConn, msg *wire.MessageRegistration) {
	Logger.log.Info("Receive a new registration START")
	var txProcessed chan struct{}
	self.netSync.QueueRegisteration(nil, msg, txProcessed)
	//<-txProcessed

	Logger.log.Info("Receive a new registration END")
}*/

// func (self *Server) OnSwapRequest(peer *peer.PeerConn, msg *wire.MessageSwapRequest) {
// 	Logger.log.Info("Receive a new request swap START")
// 	var txProcessed chan struct{}
// 	self.netSync.QueueMessage(nil, msg, txProcessed)
// 	Logger.log.Info("Receive a new request swap END")
// }

// func (self *Server) OnSwapSig(peer *peer.PeerConn, msg *wire.MessageSwapSig) {
// 	Logger.log.Info("Receive a new sign swap START")
// 	var txProcessed chan struct{}
// 	self.netSync.QueueMessage(nil, msg, txProcessed)
// 	Logger.log.Info("Receive a new sign swap END")
// }

// func (self *Server) OnSwapUpdate(peer *peer.PeerConn, msg *wire.MessageSwapUpdate) {
// 	Logger.log.Info("Receive a new update swap START")
// 	var txProcessed chan struct{}
// 	self.netSync.QueueMessage(nil, msg, txProcessed)
// 	Logger.log.Info("Receive a new update swap END")
// }

/*
// OnVersion is invoked when a peer receives a version message
// and is used to negotiate the protocol version details as well as kick start
// the communications.
*/
func (self *Server) OnVersion(peerConn *peer.PeerConn, msg *wire.MessageVersion) {
	Logger.log.Info("Receive version message START")

	pbk := ""
	err := cashec.ValidateDataB58(msg.PublicKey, msg.SignDataB58, []byte{0x00})
	if err == nil {
		pbk = msg.PublicKey
	} else {
		peerConn.ForceClose()
		return
	}
	remotePeer := &peer.Peer{
		ListeningAddress: msg.LocalAddress,
		RawAddress:       msg.RawLocalAddress,
		PeerID:           msg.LocalPeerId,
		PublicKey:        pbk,
	}
	peerConn.RemotePeer.PublicKey = pbk

	self.cNewPeers <- remotePeer
	valid := false
	if msg.ProtocolVersion == self.protocolVersion {
		valid = true
	}

	// check for accept connection
	if !self.connManager.CheckForAcceptConn(peerConn) {
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
		err := self.PushVersionMessage(peerConn)
		if err != nil {
			Logger.log.Error(err)
		}
	}

	Logger.log.Info("Receive version message END")
}

/*
OnVerAck is invoked when a peer receives a version acknowlege message
*/
func (self *Server) OnVerAck(peerConn *peer.PeerConn, msg *wire.MessageVerAck) {
	Logger.log.Info("Receive verack message START")

	if msg.Valid {
		peerConn.VerValid = true

		if peerConn.IsOutbound {
			self.addrManager.Good(peerConn.RemotePeer)
		}

		// send message for get addr
		msgS, err := wire.MakeEmptyMessage(wire.CmdGetAddr)
		if err != nil {
			return
		}
		var dc chan<- struct{}
		peerConn.QueueMessageWithEncoding(msgS, dc, peer.MESSAGE_TO_PEER, nil)

		//	broadcast addr to all peer
		for _, listen := range self.connManager.ListeningPeers {
			msgS, err := wire.MakeEmptyMessage(wire.CmdAddr)
			if err != nil {
				return
			}

			rawPeers := []wire.RawPeer{}
			peers := self.addrManager.AddressCache()
			for _, peer := range peers {
				if peerConn.RemotePeerID.Pretty() != self.connManager.GetPeerId(peer.RawAddress) {
					rawPeers = append(rawPeers, wire.RawPeer{peer.RawAddress, peer.PublicKey})
				}
			}
			msgS.(*wire.MessageAddr).RawPeers = rawPeers
			var doneChan chan<- struct{}
			for _, _peerConn := range listen.PeerConns {
				go _peerConn.QueueMessageWithEncoding(msgS, doneChan, peer.MESSAGE_TO_PEER, nil)
			}
		}

		// send message get blocks

		//msgNew, err := wire.MakeEmptyMessage(wire.CmdGetBlocks)
		//msgNew.(*wire.MessageGetBlocks).LastBlockHash = *self.blockChain.BestState.BestBlockHash
		//println(peerConn.ListenerPeer.PeerId.String())
		//msgNew.(*wire.MessageGetBlocks).SenderID = peerConn.ListenerPeer.PeerId.String()
		//if err != nil {
		//	return
		//}
		//peerConn.QueueMessageWithEncoding(msgNew, nil)
	} else {
		peerConn.VerValid = true
	}

	Logger.log.Info("Receive verack message END")
}

func (self *Server) OnGetAddr(peerConn *peer.PeerConn, msg *wire.MessageGetAddr) {
	Logger.log.Info("Receive getaddr message START")

	// send message for addr
	msgS, err := wire.MakeEmptyMessage(wire.CmdAddr)
	if err != nil {
		return
	}

	addresses := []string{}
	peers := self.addrManager.AddressCache()
	for _, peer := range peers {
		if peerConn.RemotePeerID.Pretty() != self.connManager.GetPeerId(peer.RawAddress) {
			addresses = append(addresses, peer.RawAddress)
		}
	}

	rawPeers := []wire.RawPeer{}
	for _, peer := range peers {
		if peerConn.RemotePeerID.Pretty() != self.connManager.GetPeerId(peer.RawAddress) {
			rawPeers = append(rawPeers, wire.RawPeer{peer.RawAddress, peer.PublicKey})
		}
	}
	msgS.(*wire.MessageAddr).RawPeers = rawPeers
	var dc chan<- struct{}
	peerConn.QueueMessageWithEncoding(msgS, dc, peer.MESSAGE_TO_PEER, nil)

	Logger.log.Info("Receive getaddr message END")
}

func (self *Server) OnAddr(peerConn *peer.PeerConn, msg *wire.MessageAddr) {
	Logger.log.Infof("Receive addr message %v", msg.RawPeers)
}

func (self *Server) OnBFTPropose(_ *peer.PeerConn, msg *wire.MessageBFTPropose) {
	Logger.log.Info("Receive a BFTPropose START")
	var txProcessed chan struct{}
	self.netSync.QueueMessage(nil, msg, txProcessed)
	Logger.log.Info("Receive a BFTPropose END")
}

func (self *Server) OnBFTPrepare(_ *peer.PeerConn, msg *wire.MessageBFTPrepare) {
	Logger.log.Info("Receive a BFTPrepare START")
	var txProcessed chan struct{}
	self.netSync.QueueMessage(nil, msg, txProcessed)
	Logger.log.Info("Receive a BFTPrepare END")
}

func (self *Server) OnBFTCommit(_ *peer.PeerConn, msg *wire.MessageBFTCommit) {
	Logger.log.Info("Receive a BFTCommit START")
	var txProcessed chan struct{}
	self.netSync.QueueMessage(nil, msg, txProcessed)
	Logger.log.Info("Receive a BFTCommit END")
}

func (self *Server) OnBFTReply(_ *peer.PeerConn, msg *wire.MessageBFTReply) {
	Logger.log.Info("Receive a BFTReply START")
	var txProcessed chan struct{}
	self.netSync.QueueMessage(nil, msg, txProcessed)
	Logger.log.Info("Receive a BFTReply END")
}

// func (self *Server) OnInvalidBlock(_ *peer.PeerConn, msg *wire.MessageInvalidBlock) {
// 	Logger.log.Info("Receive a invalidblock START", msg)
// 	var txProcessed chan struct{}
// 	self.netSync.QueueMessage(nil, msg, txProcessed)
// 	Logger.log.Info("Receive a invalidblock END", msg)
// }

func (self *Server) OnGetChainState(_ *peer.PeerConn, msg *wire.MessageGetChainState) {
	Logger.log.Info("Receive a getchainstate START")
	var txProcessed chan struct{}
	self.netSync.QueueMessage(nil, msg, txProcessed)
	Logger.log.Info("Receive a getchainstate END")
}

func (self *Server) OnChainState(_ *peer.PeerConn, msg *wire.MessageChainState) {
	Logger.log.Info("Receive a chainstate START")
	var txProcessed chan struct{}
	self.netSync.QueueMessage(nil, msg, txProcessed)
	Logger.log.Info("Receive a chainstate END")
}

func (self *Server) GetPeerIDsFromPublicKey(pubKey string) []libp2p.ID {
	result := []libp2p.ID{}

	for _, listener := range self.connManager.Config.ListenerPeers {
		for _, peerConn := range listener.PeerConns {
			// Logger.log.Info("Test PeerConn", peerConn.RemotePeer.PaymentAddress)
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
	}

	return result
}

/*
PushMessageToAll broadcast msg
*/
func (self *Server) PushMessageToAll(msg wire.Message) error {
	Logger.log.Info("Push msg to all peers")
	var dc chan<- struct{}
	for index := 0; index < len(self.connManager.Config.ListenerPeers); index++ {
		msg.SetSenderID(self.connManager.Config.ListenerPeers[index].PeerID)
		self.connManager.Config.ListenerPeers[index].QueueMessageWithEncoding(msg, dc, peer.MESSAGE_TO_ALL, nil)
	}
	return nil
}

/*
PushMessageToPeer push msg to peer
*/
func (self *Server) PushMessageToPeer(msg wire.Message, peerId libp2p.ID) error {
	Logger.log.Infof("Push msg to peer %s", peerId.String())
	var dc chan<- struct{}
	for index := 0; index < len(self.connManager.Config.ListenerPeers); index++ {
		peerConn := self.connManager.Config.ListenerPeers[index].GetPeerConnByPeerID(peerId.Pretty())
		if peerConn != nil {
			msg.SetSenderID(self.connManager.Config.ListenerPeers[index].PeerID)
			peerConn.QueueMessageWithEncoding(msg, dc, peer.MESSAGE_TO_PEER, nil)
			Logger.log.Infof("Pushed peer %s", peerId.Pretty())
			return nil
		} else {
			Logger.log.Error("RemotePeer not exist!")
		}
	}
	return errors.New("RemotePeer not found")
}

/*
PushMessageToPeer push msg to pbk
*/
func (self *Server) PushMessageToPbk(msg wire.Message, pbk string) error {
	Logger.log.Infof("Push msg to pbk %s", pbk)
	peerConns := self.connManager.GetPeerConnOfPbk(pbk)
	if peerConns != nil && len(peerConns) > 0 {
		for _, peerConn := range peerConns {
			msg.SetSenderID(peerConn.ListenerPeer.PeerID)
			peerConn.QueueMessageWithEncoding(msg, nil, peer.MESSAGE_TO_PEER, nil)
		}
		Logger.log.Infof("Pushed pbk %s", pbk)
		return nil
	} else {
		Logger.log.Error("RemotePeer not exist!")
	}
	return errors.New("RemotePeer not found")
}

/*
PushMessageToPeer push msg to pbk
*/
func (self *Server) PushMessageToShard(msg wire.Message, shard byte) error {
	Logger.log.Infof("Push msg to shard %d", shard)
	peerConns := self.connManager.GetPeerConnOfShard(shard)
	if peerConns != nil && len(peerConns) > 0 {
		for _, peerConn := range peerConns {
			msg.SetSenderID(peerConn.ListenerPeer.PeerID)
			peerConn.QueueMessageWithEncoding(msg, nil, peer.MESSAGE_TO_SHARD, &shard)
		}
		Logger.log.Infof("Pushed shard %d", shard)
	} else {
		Logger.log.Error("RemotePeer of shard not exist!")
		for _, listener := range self.connManager.Config.ListenerPeers {
			listener.QueueMessageWithEncoding(msg, nil, peer.MESSAGE_TO_SHARD, &shard)
		}
	}
	return nil
}

func (self *Server) PushRawBytesToShard(p *peer.PeerConn, msgBytes *[]byte, shard byte) error {
	Logger.log.Infof("Push raw bytes to shard %d", shard)
	peerConns := self.connManager.GetPeerConnOfShard(shard)
	if peerConns != nil && len(peerConns) > 0 {
		for _, peerConn := range peerConns {
			if p == nil || peerConn != p {
				peerConn.QueueMessageWithBytes(msgBytes, nil)
			}
		}
		Logger.log.Infof("Pushed shard %d", shard)
	} else {
		Logger.log.Error("RemotePeer of shard not exist!")
		peerConns := self.connManager.GetPeerConnOfAll()
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
func (self *Server) PushMessageToBeacon(msg wire.Message) error {
	Logger.log.Infof("Push msg to beacon")
	peerConns := self.connManager.GetPeerConnOfBeacon()
	if peerConns != nil && len(peerConns) > 0 {
		// fmt.Println(len(peerConns))
		for _, peerConn := range peerConns {
			msg.SetSenderID(peerConn.ListenerPeer.PeerID)
			peerConn.QueueMessageWithEncoding(msg, nil, peer.MESSAGE_TO_BEACON, nil)
		}
		Logger.log.Infof("Pushed beacon done")
		return nil
	} else {
		Logger.log.Error("RemotePeer of beacon not exist!")
		for _, listener := range self.connManager.Config.ListenerPeers {
			listener.QueueMessageWithEncoding(msg, nil, peer.MESSAGE_TO_BEACON, nil)
		}
	}
	return errors.New("RemotePeer of beacon not found")
}

func (self *Server) PushRawBytesToBeacon(p *peer.PeerConn, msgBytes *[]byte) error {
	Logger.log.Infof("Push raw bytes to beacon")
	peerConns := self.connManager.GetPeerConnOfBeacon()
	if peerConns != nil && len(peerConns) > 0 {
		for _, peerConn := range peerConns {
			if p == nil || peerConn != p {
				peerConn.QueueMessageWithBytes(msgBytes, nil)
			}
		}
		Logger.log.Infof("Pushed raw bytes beacon done")
	} else {
		Logger.log.Error("RemotePeer of beacon not exist!")
		peerConns := self.connManager.GetPeerConnOfAll()
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
func (self *Server) handleAddPeerMsg(peer *peer.Peer) bool {
	if peer == nil {
		return false
	}
	Logger.log.Info("Zero peer have just sent a message version")
	//Logger.log.Info(peer)
	return true
}

/*
GetChainState - send a getchainstate msg to connected peer
*/
func (self *Server) PushMessageGetChainState() error {
	Logger.log.Infof("Send a GetChainState")
	for _, listener := range self.connManager.Config.ListenerPeers {
		msg, err := wire.MakeEmptyMessage(wire.CmdGetChainState)
		if err != nil {
			return err
		}
		msg.(*wire.MessageGetChainState).Timestamp = time.Unix(time.Now().Unix(), 0)
		msg.SetSenderID(listener.PeerID)
		Logger.log.Infof("Send a GetChainState from %s", listener.RawAddress)
		listener.QueueMessageWithEncoding(msg, nil, peer.MESSAGE_TO_PEER, nil)
	}
	return nil
}

func (self *Server) PushVersionMessage(peerConn *peer.PeerConn) error {
	// push message version
	msg, err := wire.MakeEmptyMessage(wire.CmdVersion)
	msg.(*wire.MessageVersion).Timestamp = time.Unix(time.Now().Unix(), 0)
	msg.(*wire.MessageVersion).LocalAddress = peerConn.ListenerPeer.ListeningAddress
	msg.(*wire.MessageVersion).RawLocalAddress = peerConn.ListenerPeer.RawAddress
	msg.(*wire.MessageVersion).LocalPeerId = peerConn.ListenerPeer.PeerID
	msg.(*wire.MessageVersion).RemoteAddress = peerConn.ListenerPeer.ListeningAddress
	msg.(*wire.MessageVersion).RawRemoteAddress = peerConn.ListenerPeer.RawAddress
	msg.(*wire.MessageVersion).RemotePeerId = peerConn.ListenerPeer.PeerID
	msg.(*wire.MessageVersion).ProtocolVersion = self.protocolVersion
	msg.(*wire.MessageVersion).PublicKey = peerConn.ListenerPeer.Config.UserKeySet.GetPublicKeyB58()
	// Validate Public Key from UserPrvKey
	// if peerConn.ListenerPeer.Config.UserKeySet != "" {
	// 	// keySet, err := cfg.GetUserKeySet()
	// 	// if err != nil {
	// 	// 	Logger.log.Critical("Invalid producer's private key")
	// 	// 	return err
	// 	// }
	// 	msg.(*wire.MessageVersion).PublicKey = peerConn.ListenerPeer.Config.UserKeySet.GetPublicKeyB58()
	// }

	// ValidateTransaction Public Key from ProducerPrvKey
	if peerConn.ListenerPeer.Config.UserKeySet != nil {
		msg.(*wire.MessageVersion).PublicKey = peerConn.ListenerPeer.Config.UserKeySet.GetPublicKeyB58()
		signDataB58, err := peerConn.ListenerPeer.Config.UserKeySet.SignDataB58([]byte{0x00})
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

func (self *Server) GetCurrentRoleShard() (string, *byte) {
	return self.connManager.GetCurrentRoleShard()
}

func (self *Server) UpdateConsensusState(role string, userPbk string, currentShard *byte, beaconCommittee []string, shardCommittee map[byte][]string) {
	self.connManager.UpdateConsensusState(role, userPbk, currentShard, beaconCommittee, shardCommittee)
}
