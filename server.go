package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	peer2 "github.com/libp2p/go-libp2p-peer"
	"github.com/ninjadotorg/cash-prototype/addrmanager"
	"github.com/ninjadotorg/cash-prototype/blockchain"
	"github.com/ninjadotorg/cash-prototype/common"
	"github.com/ninjadotorg/cash-prototype/connmanager"
	"github.com/ninjadotorg/cash-prototype/consensus/pos"
	"github.com/ninjadotorg/cash-prototype/database"
	"github.com/ninjadotorg/cash-prototype/mempool"
	"github.com/ninjadotorg/cash-prototype/mining"
	"github.com/ninjadotorg/cash-prototype/netsync"
	"github.com/ninjadotorg/cash-prototype/peer"
	"github.com/ninjadotorg/cash-prototype/rpcserver"
	"github.com/ninjadotorg/cash-prototype/wallet"
	"github.com/ninjadotorg/cash-prototype/wire"
	"github.com/ninjadotorg/cash-prototype/wallet"
	"path/filepath"
	"os"
	"strconv"
	"crypto/tls"
)

const (
	defaultNumberOfTargetOutbound = 8
)

// onionAddr implements the net.Addr interface and represents a tor address.
type onionAddr struct {
	addr string
}

type Server struct {
	started     int32
	startupTime int64

	donePeers chan *peer.Peer
	quit      chan struct{}
	newPeers  chan *peer.Peer

	chainParams *blockchain.Params
	ConnManager *connmanager.ConnManager
	BlockChain  *blockchain.BlockChain
	Db          database.DB
	RpcServer   *rpcserver.RpcServer
	MemPool     *mempool.TxPool
	WaitGroup   sync.WaitGroup
	// Miner       *miner.Miner
	NetSync     *netsync.NetSync
	AddrManager *addrmanager.AddrManager
	Wallet      *wallet.Wallet

	ConsensusEngine *pos.Engine
}

// setupRPCListeners returns a slice of listeners that are configured for use
// with the RPC server depending on the configuration settings for listen
// addresses and TLS.
func setupRPCListeners() ([]net.Listener, error) {
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
		keypair, err := tls.LoadX509KeyPair(cfg.RPCCert, cfg.RPCKey)
		if err != nil {
			return nil, err
		}

		tlsConfig := tls.Config{
			Certificates: []tls.Certificate{keypair},
			MinVersion:   tls.VersionTLS12,
		}

		// Change the standard net.Listen function to the tls one.
		listenFunc = func(net string, laddr string) (net.Listener, error) {
			return tls.Listen(net, laddr, &tlsConfig)
		}
	} else {
		Logger.log.Info("Disable TLS for RPC is true")
	}

	netAddrs, err := parseListeners(cfg.RPCListeners, "tcp")
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

func (self *Server) NewServer(listenAddrs []string, db database.DB, chainParams *blockchain.Params, interrupt <-chan struct{}) error {

	// Init data for Server
	self.chainParams = chainParams
	self.quit = make(chan struct{})
	self.donePeers = make(chan *peer.Peer)
	self.newPeers = make(chan *peer.Peer)
	self.Db = db

	var err error

	// Create a new block chain instance with the appropriate configuration.9
	self.BlockChain = &blockchain.BlockChain{}
	err = self.BlockChain.Init(&blockchain.Config{
		ChainParams: self.chainParams,
		DataBase:    self.Db,
		Interrupt:   interrupt,
	})
	if err != nil {
		return err
	}

	self.MemPool = mempool.New(&mempool.Config{
		Policy:      mempool.Policy{},
		BlockChain:  self.BlockChain,
		ChainParams: chainParams,
	})

	self.AddrManager = addrmanager.New(cfg.DataDir, nil)

	blockTemplateGenerator := mining.NewBlkTmplGenerator(self.MemPool, self.BlockChain)

	// self.Miner = miner.New(&miner.Config{
	// 	ChainParams:            self.chainParams,
	// 	BlockTemplateGenerator: blockTemplateGenerator,
	// 	MiningAddrs:            cfg.MiningAddrs,
	// 	Chain:                  self.BlockChain,
	// 	Server:                 self,
	// })

	self.ConsensusEngine = pos.New(&pos.Config{
		ChainParams: self.chainParams,
		Chain:       self.BlockChain,
		MemPool:     self.MemPool,
		BlockGen:    blockTemplateGenerator,
		Server:      self,
	})

	// Init Net Sync manager to process messages
	self.NetSync, err = netsync.NetSync{}.New(&netsync.NetSyncConfig{
		BlockChain: self.BlockChain,
		ChainParam: chainParams,
		MemPool:    self.MemPool,
		Server:     self,
	})
	if err != nil {
		return err
	}

	var peers []peer.Peer
	if !cfg.DisableListen {
		var err error
		peers, err = self.InitListenerPeers(self.AddrManager, listenAddrs)
		if err != nil {
			return err
		}
	}

	// Create a connection manager.
	targetOutbound := defaultNumberOfTargetOutbound
	if cfg.MaxPeers < targetOutbound {
		targetOutbound = cfg.MaxPeers
	}

	connManager, err := connmanager.ConnManager{}.New(&connmanager.Config{
		OnInboundAccept:      self.InboundPeerConnected,
		OnOutboundConnection: self.OutboundPeerConnected,
		ListenerPeers:        peers,
		TargetOutbound:       uint32(targetOutbound),
	})
	if err != nil {
		return err
	}
	self.ConnManager = connManager

	// Start up persistent peers.
	permanentPeers := cfg.ConnectPeers
	if len(permanentPeers) == 0 {
		permanentPeers = cfg.AddPeers
	}

	for _, addr := range permanentPeers {
		go self.ConnManager.Connect(addr)
	}

	if !cfg.DisableRPC {
		// Setup listeners for the configured RPC listen addresses and
		// TLS settings.
		rpcListeners, err := setupRPCListeners()
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
			BlockChain:    self.BlockChain,
			TxMemPool:     self.MemPool,
			Server:        self,
			Wallet:        self.Wallet,
			ConnMgr:       self.ConnManager,
			AddrMgr:       self.AddrManager,
			RPCUser:       cfg.RPCUser,
			RPCPass:       cfg.RPCPass,
			RPCLimitUser:  cfg.RPCLimitUser,
			RPCLimitPass:  cfg.RPCLimitPass,
			DisableAuth:   cfg.RPCDisableAuth,
		}
		self.RpcServer = &rpcserver.RpcServer{}
		err = self.RpcServer.Init(&rpcConfig)
		if err != nil {
			return err
		}

		// Signal process shutdown when the RPC server requests it.
		go func() {
			<-self.RpcServer.RequestedProcessShutdown()
			shutdownRequestChannel <- struct{}{}
		}()
	}

	return nil
}

func (self *Server) InboundPeerConnected(peerConn *peer.PeerConn) {
	Logger.log.Info("inbound connected")
}

// outboundPeerConnected is invoked by the connection manager when a new
// outbound connection is established.  It initializes a new outbound server
// peer instance, associates it with the relevant state such as the connection
// request instance and the connection itself, and finally notifies the address
// manager of the attempt.
func (self *Server) OutboundPeerConnected(peerConn *peer.PeerConn) {
	Logger.log.Info("Outbound PEER connected with PEER ID - " + peerConn.PeerId.String())
	// TODO:
	// call address manager to process new outbound peer
	// push message version
	// if message version is compatible -> add outbound peer to address manager
	//for _, listen := range self.ConnManager.Config.ListenerPeers {
	//	listen.NegotiateOutboundProtocol(peer)
	//}
	//go self.peerDoneHandler(peer)
	//
	//msgNew, err := wire.MakeEmptyMessage(wire.CmdGetBlocks)
	//msgNew.(*wire.MessageGetBlocks).LastBlockHash = *self.BlockChain.BestState.BestBlock.Hash()
	//msgNew.(*wire.MessageGetBlocks).SenderID = self.ConnManager.Config.ListenerPeers[0].PeerId
	//if err != nil {
	//	return
	//}
	//self.ConnManager.Config.ListenerPeers[0].QueueMessageWithEncoding(msgNew, nil)

	// push message version
	msg, err := wire.MakeEmptyMessage(wire.CmdVersion)
	msg.(*wire.MessageVersion).Timestamp = time.Unix(time.Now().Unix(), 0)
	msg.(*wire.MessageVersion).LocalAddress = peerConn.ListenerPeer.ListeningAddress
	msg.(*wire.MessageVersion).RawLocalAddress = peerConn.ListenerPeer.RawAddress
	msg.(*wire.MessageVersion).LocalPeerId = peerConn.ListenerPeer.PeerId
	msg.(*wire.MessageVersion).RemoteAddress = peerConn.ListenerPeer.ListeningAddress
	msg.(*wire.MessageVersion).RawRemoteAddress = peerConn.ListenerPeer.RawAddress
	msg.(*wire.MessageVersion).RemotePeerId = peerConn.ListenerPeer.PeerId
	msg.(*wire.MessageVersion).LastBlock = 0
	msg.(*wire.MessageVersion).ProtocolVersion = 1
	if err != nil {
		return
	}
	dc := make(chan struct{})
	peerConn.QueueMessageWithEncoding(msg, dc)
}

// peerDoneHandler handles peer disconnects by notifiying the server that it's
// done along with other performing other desirable cleanup.
func (self *Server) peerDoneHandler(peer *peer.Peer) {
	//peer.WaitForDisconnect()
	self.donePeers <- peer
}

// WaitForShutdown blocks until the main listener and peer handlers are stopped.
func (self Server) WaitForShutdown() {
	self.WaitGroup.Wait()
}

// Stop gracefully shuts down the connection manager.
func (self Server) Stop() error {
	// stop connection manager
	self.ConnManager.Stop()

	// Shutdown the RPC server if it's not disabled.
	if !cfg.DisableRPC && self.RpcServer != nil {
		self.RpcServer.Stop()
	}

	self.Miner.Stop()

	close(self.quit)
	return nil
}

// peerHandler is used to handle peer operations such as adding and removing
// peers to and from the server, banning peers, and broadcasting messages to
// peers.  It must be run in a goroutine.
func (self Server) peerHandler() {
	// Start the address manager and sync manager, both of which are needed
	// by peers.  This is done here since their lifecycle is closely tied
	// to this handler and rather than adding more channels to sychronize
	// things, it's easier and slightly faster to simply start and stop them
	// in this handler.
	self.AddrManager.Start()
	self.NetSync.Start()

	Logger.log.Info("Start peer handler")

	if !cfg.DisableDNSSeed {
		// TODO load peer from seed DNS
		// add to address manager
		//self.AddrManager.AddAddresses(make([]*peer.Peer, 0))

		self.ConnManager.SeedFromDNS(self.chainParams.DNSSeeds, func(addrs []string) {
			// Bitcoind uses a lookup of the dns seeder here. This
			// is rather strange since the values looked up by the
			// DNS seed lookups will vary quite a lot.
			// to replicate this behaviour we put all addresses as
			// having come from the first one.
			self.AddrManager.AddAddressesStr(addrs)
		})
	}

	if len(cfg.ConnectPeers) == 0 {
		// TODO connect with peer in file
		for _, addr := range self.AddrManager.AddressCache() {
			go self.ConnManager.Connect(addr.RawAddress)
		}
	}

	go self.ConnManager.Start()

out:
	for {
		select {
		case p := <-self.donePeers:
			self.handleDonePeerMsg(p)
		case p := <-self.newPeers:
			self.handleAddPeerMsg(p)
		case <-self.quit:
			{
				// Disconnect all peers on server shutdown.
				//state.forAllPeers(func(sp *serverPeer) {
				//	srvrLog.Tracef("Shutdown peer %s", sp)
				//	sp.Disconnect()
				//})
				break out
			}
		}
	}
	self.NetSync.Stop()
	self.AddrManager.Stop()
	self.ConnManager.Stop()
}

// Start begins accepting connections from peers.
func (self Server) Start() {
	// Already started?
	if atomic.AddInt32(&self.started, 1) != 1 {
		return
	}

	Logger.log.Info("Starting server")
	// Server startup time. Used for the uptime command for uptime calculation.
	self.startupTime = time.Now().Unix()

	// Start the peer handler which in turn starts the address and block
	// managers.
	self.WaitGroup.Add(1)

	go self.peerHandler()

	if !cfg.DisableRPC && self.RpcServer != nil {
		self.WaitGroup.Add(1)

		// Start the rebroadcastHandler, which ensures user tx received by
		// the RPC server are rebroadcast until being included in a block.
		//go self.rebroadcastHandler()

		self.RpcServer.Start()
	}

	//creat mining
	if cfg.Generate == true && (len(cfg.MiningAddrs) > 0) {
		self.Miner.Start()
	}

	// test, print length of chain
	/*go func(server Server) {
		for {
			time.Sleep(time.Second * 3)
			log.Printf("\n --- BlockChain length: %d ---- \n", len(server.BlockChain.Blocks))
		}
	}(self)*/
}

// parseListeners determines whether each listen address is IPv4 and IPv6 and
// returns a slice of appropriate net.Addrs to listen on with TCP. It also
// properly detects addresses which apply to "all interfaces" and adds the
// address as both IPv4 and IPv6.
func parseListeners(addrs []string, netType string) ([]common.SimpleAddr, error) {
	netAddrs := make([]common.SimpleAddr, 0, len(addrs)*2)
	for _, addr := range addrs {
		host, _, err := net.SplitHostPort(addr)
		if err != nil {
			// Shouldn't happen due to already being normalized.
			return nil, err
		}

		// Empty host or host of * on plan9 is both IPv4 and IPv6.
		if host == "" || (host == "*" && runtime.GOOS == "plan9") {
			netAddrs = append(netAddrs, common.SimpleAddr{Net: netType + "4", Addr: addr})
			//netAddrs = append(netAddrs, simpleAddr{net: netType + "6", addr: addr})
			continue
		}

		// Strip IPv6 zone id if present since net.ParseIP does not
		// handle it.
		zoneIndex := strings.LastIndex(host, "%")
		if zoneIndex > 0 {
			host = host[:zoneIndex]
		}

		// Parse the IP.
		ip := net.ParseIP(host)
		if ip == nil {
			return nil, fmt.Errorf("'%s' is not a valid IP address", host)
		}

		// To4 returns nil when the IP is not an IPv4 address, so use
		// this determine the address type.
		if ip.To4() == nil {
			//netAddrs = append(netAddrs, simpleAddr{net: netType + "6", addr: addr})
		} else {
			netAddrs = append(netAddrs, common.SimpleAddr{Net: netType + "4", Addr: addr})
		}
	}
	return netAddrs, nil
}

// initListeners initializes the configured net listeners and adds any bound
// addresses to the address manager. Returns the listeners and a NAT interface,
// which is non-nil if UPnP is in use.
func (self *Server) InitListenerPeers(amgr *addrmanager.AddrManager, listenAddrs []string) ([]peer.Peer, error) {
	netAddrs, err := parseListeners(listenAddrs, "ip")
	if err != nil {
		return nil, err
	}

	kc := KeyCache{}
	kc.Load(filepath.Join(cfg.DataDir, "kc.json"))

	peers := make([]peer.Peer, 0, len(netAddrs))
	for _, addr := range netAddrs {
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
			FlagMutex:        sync.Mutex{},
			ListeningAddress: addr,
			Config:           *self.NewPeerConfig(),
			PeerConns:        make(map[string]*peer.PeerConn),
			//OutboundReaderWriterStreams: make(map[peer2.ID]*bufio.ReadWriter),
			//InboundReaderWriterStreams:  make(map[peer2.ID]*bufio.ReadWriter),
		}.NewPeer()
		if err != nil {
			return nil, err
		}
		peers = append(peers, *peer)
	}

	kc.Save()

	return peers, nil
}

/**
// newPeerConfig returns the configuration for the listening Peer.
*/
func (self *Server) NewPeerConfig() *peer.Config {
	return &peer.Config{
		MessageListeners: peer.MessageListeners{
			OnBlock:     self.OnBlock,
			OnTx:        self.OnTx,
			OnVersion:   self.OnVersion,
			OnGetBlocks: self.OnGetBlocks,
			OnVerAck:    self.OnVerAck,
			OnGetAddr:   self.OnGetAddr,
			OnAddr:      self.OnAddr,

			//PoS
			OnRequestSign:  self.OnRequestSign,
			OnInvalidBlock: self.OnInvalidBlock,
			OnSignedBlock:  self.OnSignedBlock,
		},
	}
}

// OnBlock is invoked when a peer receives a block message.  It
// blocks until the coin block has been fully processed.
func (self *Server) OnBlock(p *peer.PeerConn,
	msg *wire.MessageBlock) {
	Logger.log.Info("Receive a new block")
	var txProcessed chan struct{}
	self.NetSync.QueueBlock(nil, msg, txProcessed)
	//<-txProcessed
}

func (self *Server) OnGetBlocks(_ *peer.PeerConn, msg *wire.MessageGetBlocks) {
	Logger.log.Info("Receive a get-block message")
	var txProcessed chan struct{}
	self.NetSync.QueueGetBlock(nil, msg, txProcessed)
	//<-txProcessed
}

// OnTx is invoked when a peer receives a tx message.  It blocks
// until the transaction has been fully processed.  Unlock the block
// handler this does not serialize all transactions through a single thread
// transactions don't rely on the previous one in a linear fashion like blocks.
func (self Server) OnTx(peer *peer.PeerConn,
	msg *wire.MessageTx) {
	Logger.log.Info("Receive a new transaction")
	var txProcessed chan struct{}
	self.NetSync.QueueTx(nil, msg, txProcessed)
	//<-txProcessed
}

// OnVersion is invoked when a peer receives a version bitcoin message
// and is used to negotiate the protocol version details as well as kick start
// the communications.
func (self *Server) OnVersion(peerConn *peer.PeerConn, msg *wire.MessageVersion) {
	remotePeer := &peer.Peer{
		ListeningAddress: msg.LocalAddress,
		RawAddress:       msg.RawLocalAddress,
		PeerId:           msg.LocalPeerId,
	}
	self.newPeers <- remotePeer
	// TODO check version message
	//

	// if version message is ok -> add to addManager
	//self.AddrManager.Good(remotePeer)

	// TODO push message again for remote peer
	//var dc chan<- struct{}
	//for _, listen := range self.ConnManager.Config.ListenerPeers {
	//	msg, err := wire.MakeEmptyMessage(wire.CmdVerack)
	//	if err != nil {
	//		continue
	//	}
	//	listen.QueueMessageWithEncoding(msg, dc)
	//}

	msgV, err := wire.MakeEmptyMessage(wire.CmdVerack)
	if err != nil {
		return
	}
	var dc chan<- struct{}
	peerConn.QueueMessageWithEncoding(msgV, dc)

}

func (self *Server) OnVerAck(peerConn *peer.PeerConn, msg *wire.MessageVerAck) {
	// TODO for onverack message
	log.Printf("Receive verack message")
	self.AddrManager.Good(peerConn.Peer)

	msgS, err := wire.MakeEmptyMessage(wire.CmdGetAddr)
	if err != nil {
		return
	}
	dc := make(chan struct{})
	peerConn.QueueMessageWithEncoding(msgS, dc)
}

func (self *Server) OnGetAddr(_ *peer.PeerConn, msg *wire.MessageGetAddr) {
	// TODO for ongetaddr message
	log.Printf("Receive getaddr message")
}

func (self *Server) OnAddr(_ *peer.PeerConn, msg *wire.MessageAddr) {
	// TODO for onaddr message
	log.Printf("Receive addr message")
}

func (self *Server) OnRequestSign(_ *peer.PeerConn, msg *wire.MessageRequestSign) {

}

func (self *Server) OnInvalidBlock(_ *peer.PeerConn, msg *wire.MessageInvalidBlock) {

}

func (self *Server) OnSignedBlock(_ *peer.PeerConn, msg *wire.MessageSignedBlock) {

}
func (self Server) PushTxMessage(hashTx *common.Hash) {
	var dc chan<- struct{}
	tx, _ := self.MemPool.GetTx(hashTx)
	for _, listen := range self.ConnManager.Config.ListenerPeers {
		msg, err := wire.MakeEmptyMessage(wire.CmdTx)
		if err != nil {
			return
		}
		msg.(*wire.MessageTx).Transaction = tx
		listen.QueueMessageWithEncoding(msg, dc)
	}
}

func (self Server) PushBlockMessageWithPeerId(block *blockchain.Block, peerID peer2.ID) error {
	var dc chan<- struct{}
	msg, err := wire.MakeEmptyMessage(wire.CmdBlock)
	msg.(*wire.MessageBlock).Block = *block
	if err != nil {
		return err
	}
	self.ConnManager.Config.ListenerPeers[0].QueueMessageWithEncoding(msg, dc)
	return nil
}

func (self *Server) PushBlockMessage(block *blockchain.Block) error {
	// TODO push block message for connected peer
	//@todo got error here
	var dc chan<- struct{}
	for _, listen := range self.ConnManager.Config.ListenerPeers {
		msg, err := wire.MakeEmptyMessage(wire.CmdBlock)
		if err != nil {
			return err
		}
		msg.(*wire.MessageBlock).Block = *block
		listen.QueueMessageWithEncoding(msg, dc)
	}
	return nil
}

func (self *Server) PushRequestSignBlock(block *blockchain.Block, peerID peer2.ID) error {
	//TODO push block to specific peerID
	msg, err := wire.MakeEmptyMessage(wire.CmdRequestSign)
	if err != nil {
		return err
	}
	msg.(*wire.MessageRequestSign).Block = *block
	return nil
}

func (self *Server) PushBlockSignature(msg *wire.MessageSignedBlock) error {
	var dc chan<- struct{}
	for _, listen := range self.ConnManager.Config.ListenerPeers {
		listen.QueueMessageWithEncoding(msg, dc)
	}
	return nil
}

func (self *Server) PushInvalidBlockMessage(msg *wire.MessageInvalidBlock) error {
	var dc chan<- struct{}
	for _, listen := range self.ConnManager.Config.ListenerPeers {
		listen.QueueMessageWithEncoding(msg, dc)
	}
	return nil
}

// handleDonePeerMsg deals with peers that have signalled they are done.  It is
// invoked from the peerHandler goroutine.
func (self *Server) handleDonePeerMsg(sp *peer.Peer) {
	//self.AddrManager.
	// TODO
}

// handleAddPeerMsg deals with adding new peers.  It is invoked from the
// peerHandler goroutine.
func (self *Server) handleAddPeerMsg(peer *peer.Peer) bool {
	if peer == nil {
		return false
	}

	// TODO:
	return true
}

func (self *Server) UpdateChain(block *blockchain.Block) {
	// save block
	self.BlockChain.StoreBlock(block)

	// save best state
	newBestState := &blockchain.BestState{}
	numTxns := uint64(len(block.Transactions))
	newBestState.Init(block, 0, 0, numTxns, numTxns, time.Unix(block.Header.Timestamp.Unix(), 0))
	self.BlockChain.BestState = newBestState
	self.BlockChain.StoreBestState()

	// save index of block
	self.BlockChain.StoreBlockIndex(block)
}
