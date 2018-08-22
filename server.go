package main

import (
	"sync"
	"log"
	"net"
	"fmt"
	"strings"
	"sync/atomic"
	"time"
	"runtime"
	"errors"

	"github.com/ninjadotorg/cash-prototype/blockchain"
	"github.com/ninjadotorg/cash-prototype/connmanager"
	"github.com/ninjadotorg/cash-prototype/database"
	"github.com/ninjadotorg/cash-prototype/peer"
	"github.com/ninjadotorg/cash-prototype/wire"
	"github.com/ninjadotorg/cash-prototype/rpcserver"
	"github.com/ninjadotorg/cash-prototype/mempool"
	"github.com/ninjadotorg/cash-prototype/common"
	"github.com/ninjadotorg/cash-prototype/mining/miner"
	"github.com/ninjadotorg/cash-prototype/mining"
	"github.com/ninjadotorg/cash-prototype/netsync"
	"github.com/ninjadotorg/cash-prototype/addrmanager"
	"bufio"
	peer2 "github.com/libp2p/go-libp2p-peer"
)

const (
	defaultNumberOfTargetOutbound = 8
)

// onionAddr implements the net.Addr interface with two struct fields
type simpleAddr struct {
	net, addr string
}

// String returns the address.
// This is part of the net.Addr interface.
func (a simpleAddr) String() string {
	return a.addr
}

// Network returns the network.
// This is part of the net.Addr interface.
func (a simpleAddr) Network() string {
	return a.net
}

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

	ChainParams *blockchain.Params
	ConnManager *connmanager.ConnManager
	Chain       *blockchain.BlockChain
	Db          database.DB
	RpcServer   *rpcserver.RpcServer
	MemPool     *mempool.TxPool
	WaitGroup   sync.WaitGroup
	Miner       *miner.Miner
	NetSync     *netsync.NetSync
	AddrManager *addrmanager.AddrManager
}

// setupRPCListeners returns a slice of listeners that are configured for use
// with the RPC server depending on the configuration settings for listen
// addresses and TLS.
func setupRPCListeners() ([]net.Listener, error) {
	// Setup TLS if not disabled.
	listenFunc := net.Listen
	if !cfg.DisableTLS {
		// Generate the TLS cert and key file if both don't already
		// exist.
		//if !fileExists(cfg.RPCKey) && !fileExists(cfg.RPCCert) {
		//	err := genCertPair(cfg.RPCCert, cfg.RPCKey)
		//	if err != nil {
		//		return nil, err
		//	}
		//}
		//keypair, err := tls.LoadX509KeyPair(cfg.RPCCert, cfg.RPCKey)
		//if err != nil {
		//	return nil, err
		//}
		//
		//tlsConfig := tls.Config{
		//	Certificates: []tls.Certificate{keypair},
		//	MinVersion:   tls.VersionTLS12,
		//}
		//
		//// Change the standard net.Listen function to the tls one.
		//listenFunc = func(net string, laddr string) (net.Listener, error) {
		//	return tls.Listen(net, laddr, &tlsConfig)
		//}
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

func (self Server) NewServer(listenAddrs []string, db database.DB, chainParams *blockchain.Params, interrupt <-chan struct{}) (*Server, error) {

	// Init data for Server
	self.ChainParams = chainParams
	self.quit = make(chan struct{})
	self.Db = db
	self.MemPool = mempool.New(&mempool.Config{
		Policy: mempool.Policy{},
	})

	self.AddrManager = addrmanager.New(cfg.DataDir, nil)

	var err error

	// Create a new block chain instance with the appropriate configuration.9
	self.Chain, err = blockchain.BlockChain{}.New(&blockchain.Config{
		ChainParams: self.ChainParams,
		Db:          self.Db,
		Interrupt:   interrupt,
	})
	if err != nil {
		return nil, err
	}

	blockTemplateGenerator := mining.NewBlkTmplGenerator(self.MemPool.MiningDescs(), self.Chain)

	self.Miner = miner.New(&miner.Config{
		ChainParams:            self.ChainParams,
		BlockTemplateGenerator: blockTemplateGenerator,
		MiningAddrs:            cfg.MiningAddrs,
		Chain:                  self.Chain,
	})

	// Init Net Sync manager to process messages
	self.NetSync, err = netsync.NetSync{}.New(&netsync.NetSyncConfig{
		Chain:      self.Chain,
		ChainParam: chainParams,
		MemPool:    self.MemPool,
	})
	if err != nil {
		return nil, err
	}

	var peers []peer.Peer
	if !cfg.DisableListen {
		var err error
		peers, err = self.InitListenerPeers(self.AddrManager, listenAddrs)
		if err != nil {
			return nil, err
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
		return nil, err
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
			return nil, err
		}
		if len(rpcListeners) == 0 {
			return nil, errors.New("RPCS: No valid listen address")
		}

		rpcConfig := rpcserver.RpcServerConfig{
			Listenters:    rpcListeners,
			RPCQuirks:     cfg.RPCQuirks,
			RPCMaxClients: cfg.RPCMaxClients,
			ChainParams:   chainParams,
			Chain:         self.Chain,
			TxMemPool:     self.MemPool,
			Server:        self,
		}
		self.RpcServer, err = rpcserver.RpcServer{}.Init(&rpcConfig)
		if err != nil {
			return nil, err
		}

		// Signal process shutdown when the RPC server requests it.
		go func() {
			<-self.RpcServer.RequestedProcessShutdown()
			shutdownRequestChannel <- struct{}{}
		}()
	}

	return &self, nil
}

func (self *Server) InboundPeerConnected(peer *peer.Peer) {
	log.Println("inbound connected")
}

// outboundPeerConnected is invoked by the connection manager when a new
// outbound connection is established.  It initializes a new outbound server
// peer instance, associates it with the relevant state such as the connection
// request instance and the connection itself, and finally notifies the address
// manager of the attempt.
func (self *Server) OutboundPeerConnected(connRequest *connmanager.ConnReq,
	peer *peer.Peer) {
	log.Println("outbound connected")

	// TODO:
	// call address manager to process new outbound peer
	// push message version
	// if message version is compatible -> add outbound peer to address manager
	for _, listen := range self.ConnManager.Config.ListenerPeers {
		listen.NegotiateOutboundProtocol(peer)
	}
	go self.peerDoneHandler(peer)
}

// peerDoneHandler handles peer disconnects by notifiying the server that it's
// done along with other performing other desirable cleanup.
func (self *Server) peerDoneHandler(peer *peer.Peer) {
	peer.WaitForDisconnect()
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

	log.Println("Start peer handler")
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

	log.Println("Starting server")
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
}

// parseListeners determines whether each listen address is IPv4 and IPv6 and
// returns a slice of appropriate net.Addrs to listen on with TCP. It also
// properly detects addresses which apply to "all interfaces" and adds the
// address as both IPv4 and IPv6.
func parseListeners(addrs []string, netType string) ([]net.Addr, error) {
	netAddrs := make([]net.Addr, 0, len(addrs)*2)
	for _, addr := range addrs {
		host, _, err := net.SplitHostPort(addr)
		if err != nil {
			// Shouldn't happen due to already being normalized.
			return nil, err
		}

		// Empty host or host of * on plan9 is both IPv4 and IPv6.
		if host == "" || (host == "*" && runtime.GOOS == "plan9") {
			netAddrs = append(netAddrs, simpleAddr{net: netType + "4", addr: addr})
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
			netAddrs = append(netAddrs, simpleAddr{net: netType + "4", addr: addr})
		}
	}
	return netAddrs, nil
}

// initListeners initializes the configured net listeners and adds any bound
// addresses to the address manager. Returns the listeners and a NAT interface,
// which is non-nil if UPnP is in use.
func (self Server) InitListenerPeers(amgr *addrmanager.AddrManager, listenAddrs []string) ([]peer.Peer, error) {
	netAddrs, err := parseListeners(listenAddrs, "ip")
	if err != nil {
		return nil, err
	}

	peers := make([]peer.Peer, 0, len(netAddrs))
	for _, addr := range netAddrs {
		peer, err := peer.Peer{
			Seed:                0,
			FlagMutex:           sync.Mutex{},
			ListeningAddress:    addr,
			Config:              *self.NewPeerConfig(),
			ReaderWritersStream: make(map[peer2.ID]*bufio.ReadWriter),
		}.NewPeer()
		if err != nil {
			return nil, err
		}
		peers = append(peers, *peer)
	}
	return peers, nil
}

/**
// newPeerConfig returns the configuration for the listening Peer.
*/
func (self *Server) NewPeerConfig() *peer.Config {
	return &peer.Config{
		MessageListeners: peer.MessageListeners{
			OnBlock:   self.OnBlock,
			OnTx:      self.OnTx,
			OnVersion: self.OnVersion,
		},
	}
}

// OnBlock is invoked when a peer receives a block message.  It
// blocks until the coin block has been fully processed.
func (self *Server) OnBlock(p *peer.Peer,
	msg *wire.MessageBlock) {
	log.Println("Receive a new block")
	var txProcessed chan struct{}
	self.NetSync.QueueBlock(nil, msg, txProcessed)
	<-txProcessed
}

// OnTx is invoked when a peer receives a tx message.  It blocks
// until the transaction has been fully processed.  Unlock the block
// handler this does not serialize all transactions through a single thread
// transactions don't rely on the previous one in a linear fashion like blocks.
func (self Server) OnTx(peer *peer.Peer,
	msg *wire.MessageTx) {
	log.Println("Receive a new transaction")
	var txProcessed chan struct{}
	self.NetSync.QueueTx(nil, msg, txProcessed)
	<-txProcessed
}

// OnVersion is invoked when a peer receives a version bitcoin message
// and is used to negotiate the protocol version details as well as kick start
// the communications.
func (self *Server) OnVersion(_ *peer.Peer, msg *wire.MessageVersion) {
	self.newPeers <- &peer.Peer{
		ListeningAddress: msg.LocalAddress,
	}
	// TODO push message again for remote peer
	var dc chan<- struct{}
	for _, listen := range self.ConnManager.Config.ListenerPeers {
		msg, err := wire.MakeEmptyMessage(wire.Cmdveack)
		if err != nil {
			return
		}
		listen.QueueMessageWithEncoding(msg, dc, )
	}
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
		listen.QueueMessageWithEncoding(msg, dc, )
	}
}

func (self Server) PushBlockMessage() {
	// TODO push block message for connected peer
	//
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
