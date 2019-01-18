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

	peer2 "github.com/libp2p/go-libp2p-peer"
	"github.com/ninjadotorg/constant/addrmanager"
	"github.com/ninjadotorg/constant/blockchain"
	"github.com/ninjadotorg/constant/cashec"
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/connmanager"
	"github.com/ninjadotorg/constant/consensus/ppos"
	"github.com/ninjadotorg/constant/database"
	"github.com/ninjadotorg/constant/mempool"
	"github.com/ninjadotorg/constant/netsync"
	"github.com/ninjadotorg/constant/peer"
	"github.com/ninjadotorg/constant/rewardagent"
	"github.com/ninjadotorg/constant/rpcserver"
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
	rpcServer       *rpcserver.RpcServer
	memPool         *mempool.TxPool
	waitGroup       sync.WaitGroup
	netSync         *netsync.NetSync
	addrManager     *addrmanager.AddrManager
	wallet          *wallet.Wallet
	consensusEngine *ppos.Engine
	blockgen        *blockchain.BlkTmplGenerator
	rewardAgent     *rewardagent.RewardAgent
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
		Logger.log.Info("Disable TLS for RPC is false")
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
func (serverObj *Server) NewServer(listenAddrs []string, db database.DatabaseInterface, chainParams *blockchain.Params, protocolVer string, interrupt <-chan struct{}) error {
	// Init data for Server
	serverObj.protocolVersion = protocolVer
	serverObj.chainParams = chainParams
	serverObj.cQuit = make(chan struct{})
	serverObj.cNewPeers = make(chan *peer.Peer)
	serverObj.dataBase = db

	var err error

	// Create a new block chain instance with the appropriate configuration.9
	if cfg.Light {
		if serverObj.wallet == nil {
			return errors.New("wallet NOT FOUND. LightMode Mode required Wallet with at least one child account")
		}
		if len(serverObj.wallet.MasterAccount.Child) < 1 {
			return errors.New("no child account in wallet. LightMode Mode required Wallet with at least one child account")
		}
	}
	serverObj.blockChain = &blockchain.BlockChain{}
	err = serverObj.blockChain.Init(&blockchain.Config{
		ChainParams: serverObj.chainParams,
		DataBase:    serverObj.dataBase,
		Interrupt:   interrupt,
		LightMode:   cfg.Light,
		Wallet:      serverObj.wallet,
	})
	if err != nil {
		return err
	}

	// Search for a feeEstimator state in the database. If none can be found
	// or if it cannot be loaded, create a new one.
	if cfg.FastMode {
		Logger.log.Info("Load chain dependencies from DB")
		serverObj.feeEstimator = make(map[byte]*mempool.FeeEstimator)
		for _, bestState := range serverObj.blockChain.BestState {
			chainID := bestState.BestBlock.Header.ChainID
			feeEstimatorData, err := serverObj.dataBase.GetFeeEstimator(chainID)
			if err == nil && len(feeEstimatorData) > 0 {
				feeEstimator, err := mempool.RestoreFeeEstimator(feeEstimatorData)
				if err != nil {
					Logger.log.Errorf("Failed to restore fee estimator %v", err)
					Logger.log.Info("Init NewFeeEstimator")
					serverObj.feeEstimator[chainID] = mempool.NewFeeEstimator(
						mempool.DefaultEstimateFeeMaxRollback,
						mempool.DefaultEstimateFeeMinRegisteredBlocks)
				} else {
					serverObj.feeEstimator[chainID] = feeEstimator
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
		BlockChain:   serverObj.blockChain,
		DataBase:     serverObj.dataBase,
		ChainParams:  chainParams,
		FeeEstimator: serverObj.feeEstimator,
	})

	serverObj.addrManager = addrmanager.New(cfg.DataDir)

	serverObj.rewardAgent, err = rewardagent.RewardAgent{}.Init(&rewardagent.RewardAgentConfig{
		BlockChain: serverObj.blockChain,
	})
	if err != nil {
		return err
	}

	serverObj.blockgen, err = blockchain.BlkTmplGenerator{}.Init(serverObj.memPool, serverObj.blockChain, serverObj.rewardAgent)
	if err != nil {
		return err
	}
	serverObj.consensusEngine, err = ppos.Engine{}.Init(&ppos.EngineConfig{
		ChainParams:  serverObj.chainParams,
		BlockChain:   serverObj.blockChain,
		ConnManager:  serverObj.connManager,
		MemPool:      serverObj.memPool,
		Server:       serverObj,
		FeeEstimator: serverObj.feeEstimator,
		BlockGen:     serverObj.blockgen,
	})
	if err != nil {
		return err
	}

	// Init Net Sync manager to process messages
	serverObj.netSync = netsync.NetSync{}.New(&netsync.NetSyncConfig{
		BlockChain:   serverObj.blockChain,
		ChainParam:   chainParams,
		MemTxPool:    serverObj.memPool,
		Server:       serverObj,
		Consensus:    serverObj.consensusEngine,
		FeeEstimator: serverObj.feeEstimator,
	})

	// Create a connection manager.
	var peers []*peer.Peer
	if !cfg.DisableListen {
		var err error
		peers, err = serverObj.InitListenerPeers(serverObj.addrManager, listenAddrs, cfg.MaxOutPeers, cfg.MaxInPeers)
		if err != nil {
			Logger.log.Error(err)
			return err
		}
	}

	connManager := connmanager.ConnManager{}.New(&connmanager.Config{
		OnInboundAccept:      serverObj.InboundPeerConnected,
		OnOutboundConnection: serverObj.OutboundPeerConnected,
		GetCurrentPbk:        serverObj.GetCurrentPbk,
		//GetCurrentShard:      serverObj.GetCurrentShard,
		//GetPbksOfShard:       serverObj.GetPbksOfShard,
		//GetShardByPbk:        serverObj.GetShardByPbk,
		//GetPbksOfBeacon:      serverObj.GetPbksOfBeacon,
		ListenerPeers:        peers,
		DiscoverPeers:        cfg.DiscoverPeers,
		DiscoverPeersAddress: cfg.DiscoverPeersAddress,
		ExternalAddress:      cfg.ExternalAddress,
		// config for connection of shard
		MaxPeerSameShard:  cfg.MaxPeerSameShard,
		MaxPeerOtherShard: cfg.MaxPeerOtherShard,
		MaxPeerOther:      cfg.MaxPeerOther,
		MaxPeerNoShard:    cfg.MaxPeerNoShard,
		MaxPeerBeacon:     cfg.MaxPeerBeacon,
	})
	serverObj.connManager = connManager

	// Start up persistent peers.
	permanentPeers := cfg.ConnectPeers
	if len(permanentPeers) == 0 {
		permanentPeers = cfg.AddPeers
	}

	for _, addr := range permanentPeers {
		go serverObj.connManager.Connect(addr, "")
	}

	fmt.Println("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA", cfg.DisableRPC)
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

		rpcConfig := rpcserver.RpcServerConfig{
			Listenters:    rpcListeners,
			RPCQuirks:     cfg.RPCQuirks,
			RPCMaxClients: cfg.RPCMaxClients,
			ChainParams:   chainParams,
			BlockChain:    serverObj.blockChain,
			TxMemPool:     serverObj.memPool,
			Server:        serverObj,
			Wallet:        serverObj.wallet,
			ConnMgr:       serverObj.connManager,
			AddrMgr:       serverObj.addrManager,
			RPCUser:       cfg.RPCUser,
			RPCPass:       cfg.RPCPass,
			RPCLimitUser:  cfg.RPCLimitUser,
			RPCLimitPass:  cfg.RPCLimitPass,
			DisableAuth:   cfg.RPCDisableAuth,
			//IsGenerateNode:  cfg.Generate,
			FeeEstimator:    serverObj.feeEstimator,
			ProtocolVersion: serverObj.protocolVersion,
			Database:        &serverObj.dataBase,
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
	Logger.log.Info("inbound connected")
}

/*
// outboundPeerConnected is invoked by the connection manager when a new
// outbound connection is established.  It initializes a new outbound server
// peer instance, associates it with the relevant state such as the connection
// request instance and the connection itserverObj, and finally notifies the address
// manager of the attempt.
*/
func (serverObj *Server) OutboundPeerConnected(peerConn *peer.PeerConn) {
	Logger.log.Info("Outbound PEER connected with PEER Id - " + peerConn.RemotePeerID.Pretty())
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
	for chainId, feeEstimator := range serverObj.feeEstimator {
		feeEstimatorData := feeEstimator.Save()
		if len(feeEstimatorData) > 0 {
			err := serverObj.dataBase.StoreFeeEstimator(feeEstimatorData, chainId)
			if err != nil {
				Logger.log.Errorf("Can't save fee estimator data on chain #%d: %v", chainId, err)
			} else {
				Logger.log.Infof("Save fee estimator data on chain #%d", chainId)
			}
		}
	}

	serverObj.consensusEngine.Stop()

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

	Logger.log.Info("Start peer handler")

	if len(cfg.ConnectPeers) == 0 {
		for _, addr := range serverObj.addrManager.AddressCache() {
			go serverObj.connManager.Connect(addr.RawAddress, addr.PublicKey)
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

	Logger.log.Info("Starting server")
	if cfg.TestNet {
		Logger.log.Critical("************************")
		Logger.log.Critical("* Testnet is active *")
		Logger.log.Critical("************************")
	}
	// Server startup time. Used for the uptime command for uptime calculation.
	serverObj.startupTime = time.Now().Unix()

	// Start the peer handler which in turn starts the address and block
	// managers.
	serverObj.waitGroup.Add(1)

	go serverObj.peerHandler()

	fmt.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!", cfg.DisableRPC, serverObj.rpcServer)
	if !cfg.DisableRPC && serverObj.rpcServer != nil {
		serverObj.waitGroup.Add(1)

		// Start the rebroadcastHandler, which ensures user tx received by
		// the RPC server are rebroadcast until being included in a block.
		//go serverObj.rebroadcastHandler()

		fmt.Println("START !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
		serverObj.rpcServer.Start()
	}

	// //creat mining
	// if cfg.Generate == true && (len(cfg.MiningAddrs) > 0) {
	// 	serverObj.Miner.Start()
	// }
	err := serverObj.consensusEngine.Start()
	if err != nil {
		Logger.log.Error(err)
		go serverObj.Stop()
		return
	}
	if cfg.Generate && (len(cfg.ProducerSpendingKey) > 0) {
		producerKeySet, err := cfg.GetProducerKeySet()
		if err != nil {
			Logger.log.Critical(err)
			return
		}
		serverObj.consensusEngine.StartProducer(*producerKeySet)
	}
}

/*
// initListeners initializes the configured net listeners and adds any bound
// addresses to the address manager. Returns the listeners and a NAT interface,
// which is non-nil if UPnP is in use.
*/
func (serverObj *Server) InitListenerPeers(amgr *addrmanager.AddrManager, listenAddrs []string, targetOutbound int, targetInbound int) ([]*peer.Peer, error) {
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
			Config:           *serverObj.NewPeerConfig(),
			PeerConns:        make(map[string]*peer.PeerConn),
			PendingPeers:     make(map[string]*peer.Peer),
		}.NewPeer()
		peer.Config.MaxInbound = targetInbound
		peer.Config.MaxOutbound = targetOutbound
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
func (serverObj *Server) NewPeerConfig() *peer.Config {
	producerKeySet, err := cfg.GetProducerKeySet()
	if err != nil {
		Logger.log.Critical("cfg GetProducerKeySet error", err)
	}
	config := &peer.Config{
		MessageListeners: peer.MessageListeners{
			OnBlock:     serverObj.OnBlock,
			OnTx:        serverObj.OnTx,
			OnVersion:   serverObj.OnVersion,
			OnGetBlocks: serverObj.OnGetBlocks,
			OnVerAck:    serverObj.OnVerAck,
			OnGetAddr:   serverObj.OnGetAddr,
			OnAddr:      serverObj.OnAddr,

			//ppos
			OnRequestSign:   serverObj.OnRequestSign,
			OnInvalidBlock:  serverObj.OnInvalidBlock,
			OnBlockSig:      serverObj.OnBlockSig,
			OnGetChainState: serverObj.OnGetChainState,
			OnChainState:    serverObj.OnChainState,
			//
			//OnRegistration: serverObj.OnRegistration,
			OnSwapRequest: serverObj.OnSwapRequest,
			OnSwapSig:     serverObj.OnSwapSig,
			OnSwapUpdate:  serverObj.OnSwapUpdate,
		},

		GetShardByPbk: serverObj.GetShardByPbk,
	}
	config.ProducerKeySet = producerKeySet

	return config
}

// OnBlock is invoked when a peer receives a block message.  It
// blocks until the coin block has been fully processed.
func (serverObj *Server) OnBlock(p *peer.PeerConn,
	msg *wire.MessageBlock) {
	Logger.log.Info("Receive a new block START")

	var txProcessed chan struct{}
	serverObj.netSync.QueueBlock(nil, msg, txProcessed)
	//<-txProcessed

	Logger.log.Info("Receive a new block END")
}

func (serverObj *Server) OnGetBlocks(_ *peer.PeerConn, msg *wire.MessageGetBlocks) {
	Logger.log.Info("Receive a " + msg.MessageType() + " message START")
	var txProcessed chan struct{}
	serverObj.netSync.QueueGetBlock(nil, msg, txProcessed)
	//<-txProcessed

	Logger.log.Info("Receive a " + msg.MessageType() + " message END")
}

// OnTx is invoked when a peer receives a tx message.  It blocks
// until the transaction has been fully processed.  Unlock the block
// handler this does not serialize all transactions through a single thread
// transactions don't rely on the previous one in a linear fashion like blocks.
func (serverObj *Server) OnTx(peer *peer.PeerConn, msg *wire.MessageTx) {
	Logger.log.Info("Receive a new transaction START")
	var txProcessed chan struct{}
	serverObj.netSync.QueueTx(nil, msg, txProcessed)
	//<-txProcessed

	Logger.log.Info("Receive a new transaction END")
}

/*func (serverObj *Server) OnRegistration(peer *peer.PeerConn, msg *wire.MessageRegistration) {
	Logger.log.Info("Receive a new registration START")
	var txProcessed chan struct{}
	serverObj.netSync.QueueRegisteration(nil, msg, txProcessed)
	//<-txProcessed

	Logger.log.Info("Receive a new registration END")
}*/

func (serverObj *Server) OnSwapRequest(peer *peer.PeerConn, msg *wire.MessageSwapRequest) {
	Logger.log.Info("Receive a new request swap START")
	var txProcessed chan struct{}
	serverObj.netSync.QueueMessage(nil, msg, txProcessed)
	Logger.log.Info("Receive a new request swap END")
}

func (serverObj *Server) OnSwapSig(peer *peer.PeerConn, msg *wire.MessageSwapSig) {
	Logger.log.Info("Receive a new sign swap START")
	var txProcessed chan struct{}
	serverObj.netSync.QueueMessage(nil, msg, txProcessed)
	Logger.log.Info("Receive a new sign swap END")
}

func (serverObj *Server) OnSwapUpdate(peer *peer.PeerConn, msg *wire.MessageSwapUpdate) {
	Logger.log.Info("Receive a new update swap START")
	var txProcessed chan struct{}
	serverObj.netSync.QueueMessage(nil, msg, txProcessed)
	Logger.log.Info("Receive a new update swap END")
}

/*
// OnVersion is invoked when a peer receives a version message
// and is used to negotiate the protocol version details as well as kick start
// the communications.
*/
func (serverObj *Server) OnVersion(peerConn *peer.PeerConn, msg *wire.MessageVersion) {
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

	serverObj.cNewPeers <- remotePeer
	valid := false
	if msg.ProtocolVersion == serverObj.protocolVersion {
		valid = true
	}

	// check for accept connection
	if !serverObj.connManager.CheckAcceptConn(peerConn) {
		peerConn.ForceClose()
		return
	}

	msgV, err := wire.MakeEmptyMessage(wire.CmdVerack)
	if err != nil {
		return
	}

	msgV.(*wire.MessageVerAck).Valid = valid
	msgV.(*wire.MessageVerAck).Timestamp = time.Now()

	peerConn.QueueMessageWithEncoding(msgV, nil)

	//	push version message again
	if !peerConn.VerAckReceived() {
		err := serverObj.PushVersionMessage(peerConn)
		if err != nil {
			Logger.log.Error(err)
		}
	}

	Logger.log.Info("Receive version message END")
}

/*
OnVerAck is invoked when a peer receives a version acknowlege message
*/
func (serverObj *Server) OnVerAck(peerConn *peer.PeerConn, msg *wire.MessageVerAck) {
	Logger.log.Info("Receive verack message START")

	if msg.Valid {
		peerConn.VerValid = true

		if peerConn.IsOutbound {
			serverObj.addrManager.Good(peerConn.RemotePeer)
		}

		// send message for get addr
		msgS, err := wire.MakeEmptyMessage(wire.CmdGetAddr)
		if err != nil {
			return
		}
		var dc chan<- struct{}
		peerConn.QueueMessageWithEncoding(msgS, dc)

		//	broadcast addr to all peer
		for _, listen := range serverObj.connManager.ListeningPeers {
			msgS, err := wire.MakeEmptyMessage(wire.CmdAddr)
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
			msgS.(*wire.MessageAddr).RawPeers = rawPeers
			var doneChan chan<- struct{}
			for _, _peerConn := range listen.PeerConns {
				go _peerConn.QueueMessageWithEncoding(msgS, doneChan)
			}
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

	Logger.log.Info("Receive verack message END")
}

func (serverObj *Server) OnGetAddr(peerConn *peer.PeerConn, msg *wire.MessageGetAddr) {
	Logger.log.Info("Receive getaddr message START")

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
	peerConn.QueueMessageWithEncoding(msgS, dc)

	Logger.log.Info("Receive getaddr message END")
}

func (serverObj *Server) OnAddr(peerConn *peer.PeerConn, msg *wire.MessageAddr) {
	Logger.log.Infof("Receive addr message %v", msg.RawPeers)
}

func (serverObj *Server) OnRequestSign(_ *peer.PeerConn, msg *wire.MessageBlockSigReq) {
	Logger.log.Info("Receive a requestsign START")
	var txProcessed chan struct{}
	serverObj.netSync.QueueMessage(nil, msg, txProcessed)
	Logger.log.Info("Receive a requestsign END")
}

func (serverObj *Server) OnInvalidBlock(_ *peer.PeerConn, msg *wire.MessageInvalidBlock) {
	Logger.log.Info("Receive a invalidblock START", msg)
	var txProcessed chan struct{}
	serverObj.netSync.QueueMessage(nil, msg, txProcessed)
	Logger.log.Info("Receive a invalidblock END", msg)
}

func (serverObj *Server) OnBlockSig(_ *peer.PeerConn, msg *wire.MessageBlockSig) {
	Logger.log.Info("Receive a BlockSig")
	var txProcessed chan struct{}
	serverObj.netSync.QueueMessage(nil, msg, txProcessed)
}

func (serverObj *Server) OnGetChainState(_ *peer.PeerConn, msg *wire.MessageGetChainState) {
	Logger.log.Info("Receive a getchainstate START")
	var txProcessed chan struct{}
	serverObj.netSync.QueueMessage(nil, msg, txProcessed)
	Logger.log.Info("Receive a getchainstate END")
}

func (serverObj *Server) OnChainState(_ *peer.PeerConn, msg *wire.MessageChainState) {
	Logger.log.Info("Receive a chainstate START")
	var txProcessed chan struct{}
	serverObj.netSync.QueueMessage(nil, msg, txProcessed)
	Logger.log.Info("Receive a chainstate END")
}

func (serverObj *Server) GetPeerIDsFromPublicKey(pubKey string) []peer2.ID {
	result := []peer2.ID{}

	for _, listener := range serverObj.connManager.Config.ListenerPeers {
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
func (serverObj *Server) PushMessageToAll(msg wire.Message) error {
	Logger.log.Info("Push msg to all peers")
	var dc chan<- struct{}
	for index := 0; index < len(serverObj.connManager.Config.ListenerPeers); index++ {
		msg.SetSenderID(serverObj.connManager.Config.ListenerPeers[index].PeerID)
		serverObj.connManager.Config.ListenerPeers[index].QueueMessageWithEncoding(msg, dc)
	}
	return nil
}

/*
PushMessageToPeer push msg to peer
*/
func (serverObj *Server) PushMessageToPeer(msg wire.Message, peerId peer2.ID) error {
	Logger.log.Infof("Push msg to peer %s", peerId.Pretty())
	var dc chan<- struct{}
	for index := 0; index < len(serverObj.connManager.Config.ListenerPeers); index++ {
		peerConn := serverObj.connManager.Config.ListenerPeers[index].GetPeerConnByPeerID(peerId.Pretty())
		if peerConn != nil {
			msg.SetSenderID(serverObj.connManager.Config.ListenerPeers[index].PeerID)
			peerConn.QueueMessageWithEncoding(msg, dc)
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
func (serverObj *Server) PushMessageToPbk(msg wire.Message, pbk string) error {
	Logger.log.Infof("Push msg to pbk %s", pbk)
	var dc chan<- struct{}
	for index := 0; index < len(serverObj.connManager.Config.ListenerPeers); index++ {
		peerConn := serverObj.connManager.Config.ListenerPeers[index].GetPeerConnByPbk(pbk)
		if peerConn != nil {
			msg.SetSenderID(serverObj.connManager.Config.ListenerPeers[index].PeerID)
			peerConn.QueueMessageWithEncoding(msg, dc)
			Logger.log.Infof("Pushed pbk %s", pbk)
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
func (serverObj *Server) PushMessageToShard(msg wire.Message, shard byte) error {
	Logger.log.Infof("Push msg to shard %d", shard)
	var dc chan<- struct{}
	for index := 0; index < len(serverObj.connManager.Config.ListenerPeers); index++ {
		peerConns := serverObj.connManager.Config.ListenerPeers[index].GetListPeerConnByShard(shard)
		if len(peerConns) > 0 {
			for _, peerConn := range peerConns {
				msg.SetSenderID(serverObj.connManager.Config.ListenerPeers[index].PeerID)
				peerConn.QueueMessageWithEncoding(msg, dc)
			}
			Logger.log.Infof("Pushed shard %d", shard)
			return nil
		} else {
			Logger.log.Error("RemotePeer of shard not exist!")
		}
	}
	return errors.New("RemotePeer of shard not found")
}

// handleAddPeerMsg deals with adding new peers.  It is invoked from the
// peerHandler goroutine.
func (serverObj *Server) handleAddPeerMsg(peer *peer.Peer) bool {
	if peer == nil {
		return false
	}
	Logger.log.Info("Zero peer have just sent a message version")
	Logger.log.Info(peer)
	return true
}

/*
GetChainState - send a getchainstate msg to connected peer
*/
func (serverObj *Server) PushMessageGetChainState() error {
	Logger.log.Infof("Send a GetChainState")
	for _, listener := range serverObj.connManager.Config.ListenerPeers {
		msg, err := wire.MakeEmptyMessage(wire.CmdGetChainState)
		if err != nil {
			return err
		}
		msg.(*wire.MessageGetChainState).Timestamp = time.Unix(time.Now().Unix(), 0)
		msg.SetSenderID(listener.PeerID)
		Logger.log.Infof("Send a GetChainState from %s", listener.RawAddress)
		listener.QueueMessageWithEncoding(msg, nil)
	}
	return nil
}

func (serverObj *Server) PushVersionMessage(peerConn *peer.PeerConn) error {
	// push message version
	msg, err := wire.MakeEmptyMessage(wire.CmdVersion)
	msg.(*wire.MessageVersion).Timestamp = time.Unix(time.Now().Unix(), 0)
	msg.(*wire.MessageVersion).LocalAddress = peerConn.ListenerPeer.ListeningAddress
	msg.(*wire.MessageVersion).RawLocalAddress = peerConn.ListenerPeer.RawAddress
	msg.(*wire.MessageVersion).LocalPeerId = peerConn.ListenerPeer.PeerID
	msg.(*wire.MessageVersion).RemoteAddress = peerConn.ListenerPeer.ListeningAddress
	msg.(*wire.MessageVersion).RawRemoteAddress = peerConn.ListenerPeer.RawAddress
	msg.(*wire.MessageVersion).RemotePeerId = peerConn.ListenerPeer.PeerID
	msg.(*wire.MessageVersion).ProtocolVersion = serverObj.protocolVersion

	// ValidateTransaction Public KeyWallet from ProducerPrvKey
	if peerConn.ListenerPeer.Config.ProducerKeySet != nil {
		msg.(*wire.MessageVersion).PublicKey = peerConn.ListenerPeer.Config.ProducerKeySet.GetPublicKeyB58()
		signDataB58, err := peerConn.ListenerPeer.Config.ProducerKeySet.SignDataB58([]byte{0x00})
		if err == nil {
			msg.(*wire.MessageVersion).SignDataB58 = signDataB58
		}
	}
	if err != nil {
		return err
	}
	peerConn.QueueMessageWithEncoding(msg, nil)
	return nil
}

func (serverObj *Server) GetShardByPbk(pbk string) *byte {
	if pbk == "" {
		return nil
	}
	shard, ok := mPBK[pbk]
	if ok {
		return &shard
	}
	return nil
}

func (serverObj *Server) GetCurrentPbk() string {
	ks, err := cfg.GetProducerKeySet()
	if err != nil {
		return ""
	}
	pbk := ks.GetPublicKeyB58()
	return pbk
}

//func (serverObj *Server) GetCurrentShard() *byte {
//	ks, err := cfg.GetProducerKeySet()
//	if err != nil {
//		return nil
//	}
//	pbk := ks.GetPublicKeyB58()
//	shard, ok := mPBK[pbk]
//	if ok {
//		return &shard
//	}
//	return nil
//}
//
//func (serverObj *Server) GetPbksOfShard(shard byte) []string {
//	pBKs := make([]string, 0)
//	for k, v := range mPBK {
//		if v == shard {
//			pBKs = append(pBKs, k)
//		}
//	}
//	return pBKs
//}
//
//func (serverObj *Server) GetPbksOfBeacon() []string {
//	pBKs := make([]string, 0)
//	return pBKs
//}
//
//func (serverObj *Server) getCurrentShardInfoByPbk() (*byte, string) {
//	ks, err := cfg.GetProducerKeySet()
//	if err != nil {
//		return nil, ""
//	}
//	pbk := ks.GetPublicKeyB58()
//	shard, ok := mPBK[pbk]
//	if ok {
//		return &shard, ""
//	}
//	return nil, ""
//}
//
//func (serverObj *Server) getShardInfoByPbk(pbk string) (*byte, string) {
//	shard, ok := mPBK[pbk]
//	if ok {
//		return &shard, ""
//	}
//	return nil, ""
//}
//
//func (serverObj *Server) shardChanged(oldShard *byte, newShard *byte) {
//	// update shard connection, random peers, drop peers and new peers
//}

var mPBK = map[string]byte{
	"15Z7uGSzG4ZR5ENDnBE6PuGcquNGYj7PqPFj4ojEEGk8QQNZoN6": 0,
	"15CfJ8vH78zw8PT2FbBeNssFWcHMW1sSxoJ6RKv2hZ6nztp4mCQ": 1,
	"17PnJ3sjHvFLp3Sck12FaHfvk4AghGctecTG54bdLNFVGygi8DN": 2,
	"17qiWdX7ubTHpVu5eMDxMCCwesYYcLWKE1KTP62LQK3ALrQ6A5T": 3,
	"18mxtXGaaRkfkLS9L7eNGTjawpxTnqZSBqKXLSDc4Un8VLGgVPg": 4,
	"17W59bSax64ykVeGPk8nnXQAoWmiDfPGtVQMVvqJSSep3Py2Jxn": 5,
	"15nvyVJvmrzp3KK7SF8xMcsffZyvV2BTBmnR4kx8XszdtXhqUm9": 6,
	"15VmkDTBgFs86h8fD7c9Bk41xndCGA3qXKmqMjy2dJpC6UVWNhZ": 7,
	"159DQTsMrzrKyF1787R2iK8RA9X8GMXjgwLqPsVR1a129RjSAi5": 8,
	"18fk4aLAT7F8aTf4Uo784DiGgEBJajC3u8SqcY766FcRPPLHPBz": 9,
	"15ma6n91BbgyCJNeWa9TzG5gQGCERLZ9F9jaYB1mMPGsJGKhmB7": 10,
	"18NwuP2PqDNcAWyhAgPpcRgFeS8h7LWv8LX7vzRgfaVmTzBERBZ": 11,
	"165RABeGBuYYX72S6w8wJqvSgZE7JZ32YVG8ApSwUW38Lm3RrEt": 12,
	"15yDGFUwf5r7rZcfEzEmpcNvMfC5zi1g454AeHMZNSGEiBFacvt": 13,
	"16C6356Xst2bKnAuXYM3Ezfz7ZwG9kiKmHAPTFMupQs3wzQfaoM": 14,
}
