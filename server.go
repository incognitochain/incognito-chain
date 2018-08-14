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
	"strings"
	"net"
	"runtime"
	"fmt"
)

const (
	defaultNumberOfTargetOutbound = 8
)

// onionAddr implements the net.Addr interface with two struct fields
type simpleAddr struct {
	net, addr string
}

// String returns the address.
//
// This is part of the net.Addr interface.
func (a simpleAddr) String() string {
	return a.addr
}

// Network returns the network.
//
// This is part of the net.Addr interface.
func (a simpleAddr) Network() string {
	return a.net
}

type Server struct {
	started     int32
	startupTime int64

	ChainParams *blockchain.Params
	ConnManager *connmanager.ConnManager

	Quit      chan struct{}
	WaitGroup sync.WaitGroup
}

func (self Server) NewServer(listenAddrs []string, db database.DB, chainParams *blockchain.Params, interrupt <-chan struct{}) (*Server, error) {

	var peers []peer.Peer
	if !cfg.DisableListen {
		// TODO with error
		peers, _ = self.InitListenerPeers(listenAddrs)
	}

	self.ChainParams = chainParams
	self.Quit = make(chan struct{})

	// Create a connection manager.
	targetOutbound := defaultNumberOfTargetOutbound
	if cfg.MaxPeers < targetOutbound {
		targetOutbound = cfg.MaxPeers
	}
	connmanager, err := connmanager.ConnManager{}.New(&connmanager.Config{
		OnInboundAccept:      self.InboundPeerConnected,
		OnOutboundConnection: self.OutboundPeerConnected,
		ListenerPeers:        peers,
		TargetOutbound:       uint32(targetOutbound),
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

// parseListeners determines whether each listen address is IPv4 and IPv6 and
// returns a slice of appropriate net.Addrs to listen on with TCP. It also
// properly detects addresses which apply to "all interfaces" and adds the
// address as both IPv4 and IPv6.
func parseListeners(addrs []string) ([]net.Addr, error) {
	netAddrs := make([]net.Addr, 0, len(addrs)*2)
	for _, addr := range addrs {
		host, _, err := net.SplitHostPort(addr)
		if err != nil {
			// Shouldn't happen due to already being normalized.
			return nil, err
		}

		// Empty host or host of * on plan9 is both IPv4 and IPv6.
		if host == "" || (host == "*" && runtime.GOOS == "plan9") {
			netAddrs = append(netAddrs, simpleAddr{net: "tcp4", addr: addr})
			netAddrs = append(netAddrs, simpleAddr{net: "tcp6", addr: addr})
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
			netAddrs = append(netAddrs, simpleAddr{net: "tcp6", addr: addr})
		} else {
			netAddrs = append(netAddrs, simpleAddr{net: "tcp4", addr: addr})
		}
	}
	return netAddrs, nil
}

// initListeners initializes the configured net listeners and adds any bound
// addresses to the address manager. Returns the listeners and a NAT interface,
// which is non-nil if UPnP is in use.
func (self Server) InitListenerPeers(listenAddrs []string) ([]peer.Peer, error) {
	netAddrs, err := parseListeners(listenAddrs)
	if err != nil {
		return nil, err
	}

	peers := make([]peer.Peer, 0, len(netAddrs))
	for _, addr := range netAddrs {
		peer, err := peer.Peer{
			Seed:               0,
			FlagMutex:          sync.Mutex{},
			ListeningPort:      strings.Split(addr.String(), ":")[0],
			ListeningIpAddress: strings.Split(addr.String(), ":")[1],
		}.NewPeer()
		if err != nil {
			return nil, err
		}
		peers = append(peers, *peer)
	}
	return peers, nil
}
