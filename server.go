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

	"github.com/internet-cash/prototype/blockchain"
	"github.com/internet-cash/prototype/connmanager"
	"github.com/internet-cash/prototype/database"
	"github.com/internet-cash/prototype/peer"
	"github.com/internet-cash/prototype/wire"
	"github.com/internet-cash/prototype/rpcserver"
	"github.com/internet-cash/prototype/mempool"
	"github.com/internet-cash/prototype/common"
	"github.com/internet-cash/prototype/mining/miner"
	"github.com/internet-cash/prototype/mining"
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

	ChainParams *blockchain.Params
	ConnManager *connmanager.ConnManager
	Chain       *blockchain.BlockChain
	Db          database.DB
	RpcServer   *rpcserver.RpcServer
	MemPool     mempool.TxPool
	Quit        chan struct{}
	WaitGroup   sync.WaitGroup
	Miner       *miner.Miner
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
	self.Quit = make(chan struct{})
	self.Db = db
	self.MemPool = *mempool.New(&mempool.Config{
		Policy: mempool.Policy{},
	})

	var peers []peer.Peer
	if !cfg.DisableListen {
		var err error
		peers, err = self.InitListenerPeers(listenAddrs)
		if err != nil {
			return nil, err
		}
	}

	// Create a new block chain instance with the appropriate configuration.9
	var err error
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
			TxMemPool:     &self.MemPool,
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

func (self Server) InboundPeerConnected(peer *peer.Peer) {

}

// outboundPeerConnected is invoked by the connection manager when a new
// outbound connection is established.  It initializes a new outbound server
// peer instance, associates it with the relevant state such as the connection
// request instance and the connection itself, and finally notifies the address
// manager of the attempt.
func (self Server) OutboundPeerConnected(connRequest *connmanager.ConnReq,
	peer *peer.Peer) {

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

	close(self.Quit)
	return nil
}

// PeerHandler is used to handle peer operations such as adding and removing
// peers to and from the server, banning peers, and broadcasting messages to
// peers.  It must be run in a goroutine.
func (self Server) PeerHandler() {
	log.Println("Start peer handler")
	go self.ConnManager.Start()
out:
	for {
		select {
		case <-self.Quit:
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
	go self.PeerHandler()

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
func (self Server) InitListenerPeers(listenAddrs []string) ([]peer.Peer, error) {
	netAddrs, err := parseListeners(listenAddrs, "ip")
	if err != nil {
		return nil, err
	}

	peers := make([]peer.Peer, 0, len(netAddrs))
	for _, addr := range netAddrs {
		peer, err := peer.Peer{
			Seed:             0,
			FlagMutex:        sync.Mutex{},
			ListeningAddress: addr,
			Config:           *self.NewPeerConfig(),
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
			OnBlock: self.OnBlock,
			OnTx:    self.OnTx,
		},
	}
}

// OnBlock is invoked when a peer receives a block message.  It
// blocks until the coin block has been fully processed.
func (self *Server) OnBlock(p *peer.Peer,
	msg *wire.MessageBlock) {
	// TODO get message block and process, Tuan Anh
}

// OnTx is invoked when a peer receives a tx message.  It blocks
// until the transaction has been fully processed.  Unlock the block
// handler this does not serialize all transactions through a single thread
// transactions don't rely on the previous one in a linear fashion like blocks.
func (self Server) OnTx(_ *peer.Peer,
	msg *wire.MessageTx) {
	log.Println("Receive a new transaction")
	// TODO get message tx and process, Tuan Anh
	hash, txDesc, error := self.MemPool.CanAcceptTransaction(msg.Transaction)

	if error != nil {
		fmt.Print(error)
	} else {
		fmt.Print("there is hash of transaction", hash)
		fmt.Print("there is priority of transaction in pool", txDesc.StartingPriority)
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
		listen.QueueMessageWithEncoding(msg, dc)
	}
}

func (self Server) PushBlockMessage() {
	// TODO push block message for connected peer
	//
}

func (self Server) handleMiner() error {

	return nil
}
