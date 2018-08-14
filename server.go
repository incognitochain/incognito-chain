package main

import (
	"github.com/internet-cash/prototype/blockchain"
	"github.com/internet-cash/prototype/connmanager"
	"github.com/internet-cash/prototype/database"
	"github.com/internet-cash/prototype/peer"
	"sync"
	"sync/atomic"
	"log"
	"time"
)

type Server struct {
	started     int32
	startupTime int64

	ChainParams *blockchain.Params
	ConnManager *connmanager.ConnManager

	Quit      chan struct{}
	WaitGroup sync.WaitGroup
}

func (self Server) NewServer(listenAddrs []string, db database.DB, chainParams *blockchain.Params, interrupt <-chan struct{}) (*Server, error) {

	if !cfg.DisableListen {

	}

	self.ChainParams = chainParams
	self.Quit = make(chan struct{})

	// Create a connection manager.
	connmanager, err := connmanager.ConnManager{}.New(&connmanager.Config{
		OnInboundAccept:      self.InboundPeerConnected,
		OnOutboundConnection: self.OutboundPeerConnected,
	})
	if err != nil {
		return nil, err
	}
	self.ConnManager = connmanager

	// Start up persistent peers.
	permanentPeers := cfg.ConnectPeers
	if len(permanentPeers) == 0 {
		permanentPeers = cfg.AddPeers
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
func (self Server) OutboundPeerConnected(connRequest *connmanager.ConnReq, peer *peer.Peer) {

}

// WaitForShutdown blocks until the main listener and peer handlers are stopped.
func (self Server) WaitForShutdown() {
	self.WaitGroup.Wait()
}

// Stop gracefully shuts down the connection manager.
func (self Server) Stop() error {
	close(self.Quit)
	return nil
}

// peerHandler is used to handle peer operations such as adding and removing
// peers to and from the server, banning peers, and broadcasting messages to
// peers.  It must be run in a goroutine.
func (self Server) peerHandler() {
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
	go self.peerHandler()
}
