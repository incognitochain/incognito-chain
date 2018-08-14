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

func (self Server) OutboundPeerConnected(peer *peer.Peer) {

}

// WaitForShutdown blocks until the main listener and peer handlers are stopped.
func (self Server) WaitForShutdown() {
	self.WaitGroup.Wait()
}

func (self Server) Stop() error {
	close(self.Quit)
	return nil
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
}
