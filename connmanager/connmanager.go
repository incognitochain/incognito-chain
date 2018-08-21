package connmanager

import (
	"github.com/internet-cash/prototype/peer"
	pstore "github.com/libp2p/go-libp2p-peerstore"
	ma "github.com/multiformats/go-multiaddr"
	libpeer "github.com/libp2p/go-libp2p-peer"
	"sync"
	"log"
	"sync/atomic"
	"context"
	"bufio"
	"fmt"
)

const (
	// defaultTargetOutbound is the default number of outbound connections to
	// maintain.
	defaultTargetOutbound = uint32(8)
)

// ConnState represents the state of the requested connection.
type ConnState uint8

// ConnState can be either pending, established, disconnected or failed.  When
// a new connection is requested, it is attempted and categorized as
// established or failed depending on the connection result.  An established
// connection which was disconnected is categorized as disconnected.
const (
	ConnPending      ConnState = iota
	ConnFailing
	ConnCanceled
	ConnEstablished
	ConnDisconnected
)

// ConnReq is the connection request to a network address. If permanent, the
// connection will be retried on disconnection.
type ConnReq struct {
	Id uint64

	Peer       peer.Peer
	Permanent  bool
	stateMtx   sync.RWMutex
	ConnState  ConnState
	retryCount uint32
}

// UpdateState updates the state of the connection request.
func (self *ConnReq) UpdateState(state ConnState) {
	self.stateMtx.Lock()
	self.ConnState = state
	self.stateMtx.Unlock()
}

type ConnManager struct {
	connReqCount uint64
	start        int32
	stop         int32

	Config Config
	// Pending Connection
	Pending map[uint64]*ConnReq

	// Connected Connection
	Connected map[uint64]*ConnReq

	WaitGroup sync.WaitGroup

	// Request channel
	Requests chan interface{}
	// Quit channel
	Quit chan struct{}

	FailedAttempts uint32
}

type Config struct {
	// ListenerPeers defines a slice of listeners for which the connection
	// manager will take ownership of and accept connections.  When a
	// connection is accepted, the OnAccept handler will be invoked with the
	// connection.  Since the connection manager takes ownership of these
	// listeners, they will be closed when the connection manager is
	// stopped.
	//
	// This field will not have any effect if the OnAccept field is not
	// also specified.  It may be nil if the caller does not wish to listen
	// for incoming connections.
	ListenerPeers []peer.Peer

	// OnInboundAccept is a callback that is fired when an inbound connection is accepted
	OnInboundAccept func(*peer.Peer)

	//OnOutboundConnection is a callback that is fired when an outbound connection is established
	OnOutboundConnection func(*ConnReq, *peer.Peer)

	//OnOutboundDisconnection is a callback that is fired when an outbound connection is disconnected
	OnOutboundDisconnection func(*ConnReq)

	// TargetOutbound is the number of outbound network connections to
	// maintain. Defaults to 8.
	TargetOutbound uint32
}

// registerPending is used to register a pending connection attempt. By
// registering pending connection attempts we allow callers to cancel pending
// connection attempts before their successful or in the case they're not
// longer wanted.
type registerPending struct {
	connRequest *ConnReq
	done        chan struct{}
}

// handleConnected is used to queue a successful connection.
type handleConnected struct {
	connRequest *ConnReq
	Peer        peer.Peer
}

// handleDisconnected is used to remove a connection.
type handleDisconnected struct {
	id    uint64
	retry bool
}

// handleFailed is used to remove a pending connection.
type handleFailed struct {
	c   *ConnReq
	err error
}

// Stop gracefully shuts down the connection manager.
func (self ConnManager) Stop() {
	if atomic.AddInt32(&self.stop, 1) != 1 {
		log.Println("Connection manager already stopped")
		return
	}
	log.Println("Stop connection manager")

	// Stop all the listeners.  There will not be any listeners if
	// listening is disabled.
	for _, listener := range self.Config.ListenerPeers {
		listener.Disconnect()
	}

	close(self.Quit)
	log.Println("Connection manager stopped")
}

func (self ConnManager) New(cfg *Config) (*ConnManager, error) {
	if cfg.TargetOutbound == 0 {
		cfg.TargetOutbound = defaultTargetOutbound
	}
	self.Config = *cfg
	self.Quit = make(chan struct{})
	self.Requests = make(chan interface{})
	return &self, nil
}

// HandleFailedConn handles a connection failed due to a disconnect or any
// other failure. If permanent, it retries the connection after the configured
// retry duration. Otherwise, if required, it makes a new connection request.
// After maxFailedConnectionAttempts new connections will be retried after the
// configured retry duration.
func (self *ConnManager) HandleFailedConn(c *ConnReq) {
	// TODO
}

// connHandler handles all connection related requests.  It must be run as a
// goroutine.
//
// The connection handler makes sure that we maintain a pool of active outbound
// connections so that we remain connected to the network.  Connection requests
// are processed and mapped by their assigned ids.
func (self ConnManager) connHandler() {
	// pending holds all registered conn requests that have yet to
	// succeed.
	self.Pending = make(map[uint64]*ConnReq)

	// conns represents the set of all actively connected peers.
	self.Connected = make(map[uint64]*ConnReq, self.Config.TargetOutbound)

out:
	for {
		select {
		case req := <-self.Requests:
			switch msg := req.(type) {
			case registerPending:
				{
					connReq := msg.connRequest
					connReq.UpdateState(ConnPending)
					self.Pending[msg.connRequest.Id] = connReq
					close(msg.done)
				}
			case handleConnected:
				{
					connReq := msg.connRequest
					if _, ok := self.Pending[connReq.Id]; !ok {
						if msg.connRequest != nil {
							//msg.conn.Close()
						}
						//log.Debugf("Ignoring connection for "+
						//	"canceled connreq=%v", connReq)
						continue
					}
					connReq.UpdateState(ConnEstablished)
					connReq.Peer = msg.Peer
					self.Connected[connReq.Id] = connReq
					//spew.Dump("Connected to ", connReq)
					connReq.retryCount = 0
					self.FailedAttempts = 0

					delete(self.Pending, connReq.Id)

					if self.Config.OnOutboundConnection != nil {
						go self.Config.OnOutboundConnection(connReq, &msg.Peer)
					}
				}
			case handleDisconnected:
				{
					connReq, ok := self.Connected[msg.id]
					if !ok {
						connReq, ok = self.Pending[msg.id]
						if !ok {
							log.Printf("Unknown connid=%d",
								msg.id)
							continue
						}

						// Pending connection was found, remove
						// it from pending map if we should
						// ignore a later, successful
						// connection.
						connReq.UpdateState(ConnCanceled)
						log.Printf("Canceling: %v", connReq)
						delete(self.Pending, msg.id)
						continue
					}
					// An existing connection was located, mark as
					// disconnected and execute disconnection
					// callback.
					log.Printf("Disconnected from %v", connReq)
					delete(self.Connected, msg.id)

					//if connReq.Peer != nil {
					//connReq.conn.Close()
					//}

					if self.Config.OnOutboundDisconnection != nil {
						go self.Config.OnOutboundDisconnection(connReq)
					}

					// All internal state has been cleaned up, if
					// this connection is being removed, we will
					// make no further attempts with this request.
					if !msg.retry {
						connReq.UpdateState(ConnDisconnected)
						continue
					}

					// Otherwise, we will attempt a reconnection if
					// we do not have enough peers, or if this is a
					// persistent peer. The connection request is
					// re added to the pending map, so that
					// subsequent processing of connections and
					// failures do not ignore the request.
					if uint32(len(self.Connected)) < self.Config.TargetOutbound ||
						connReq.Permanent {

						connReq.UpdateState(ConnPending)
						log.Printf("Reconnecting to %v",
							connReq)
						self.Pending[msg.id] = connReq
						self.HandleFailedConn(connReq)
					}
				}
			case handleFailed:
				{
					connReq := msg.c

					if _, ok := self.Pending[connReq.Id]; !ok {
						log.Printf("Ignoring connection for "+
							"canceled conn req: %v", connReq)
						continue
					}

					connReq.UpdateState(ConnFailing)
					log.Printf("Failed to connect to %v: %v",
						connReq, msg.err)
					self.HandleFailedConn(connReq)
				}
			}
		case <-self.Quit:
			break out
		}
	}
	self.WaitGroup.Done()
	log.Printf("Connection handler done")
}

// Connect assigns an id and dials a connection to the address of the
// connection request.
func (self ConnManager) Connect(addr string) {
	if atomic.LoadInt32(&self.stop) != 0 {
		return
	}
	// The following code extracts target's peer ID from the
	// given multiaddress
	ipfsaddr, err := ma.NewMultiaddr(addr)
	if err != nil {
		log.Print(err)
		return
	}

	pid, err := ipfsaddr.ValueForProtocol(ma.P_IPFS)
	if err != nil {
		log.Print(err)
		return
	}

	peerId, err := libpeer.IDB58Decode(pid)
	if err != nil {
		log.Print(err)
		return
	}

	// Decapsulate the /ipfs/<peerID> part from the target
	// /ip4/<a.b.c.d>/ipfs/<peer> becomes /ip4/<a.b.c.d>
	targetPeerAddr, _ := ma.NewMultiaddr(
		fmt.Sprintf("/ipfs/%s", libpeer.IDB58Encode(peerId)))
	targetAddr := ipfsaddr.Decapsulate(targetPeerAddr)

	connReq := ConnReq{
		Permanent: true,
		Peer: peer.Peer{
			TargetAddress: targetAddr,
			PeerId:        peerId,
		},
	}
	if atomic.LoadUint64(&connReq.Id) == 0 {
		atomic.StoreUint64(&connReq.Id, atomic.AddUint64(&self.connReqCount, 1))

		// Submit a request of a pending connection attempt to the
		// connection manager. By registering the id before the
		// connection is even established, we'll be able to later
		// cancel the connection via the Remove method.
		done := make(chan struct{})
		select {
		case self.Requests <- registerPending{&connReq, done}:
		case <-self.Quit:
			return
		}

		// Wait for the registration to successfully add the pending
		// conn req to the conn manager's internal state.
		select {
		case <-done:
		case <-self.Quit:
			return
		}
	}

	//spew.Dump("Attempting to connect to", connRequest.Peer.TargetAddress.String())

	for _, listen := range self.Config.ListenerPeers {
		listen.Host.Peerstore().AddAddr(connReq.Peer.PeerId, connReq.Peer.TargetAddress, pstore.PermanentAddrTTL)
		log.Printf("opening stream %s \n", connReq.Peer.PeerId.String())
		// make a new stream from host B to host A
		// it should be handled on host A by the handler we set above because
		// we use the same /peer/1.0.0 protocol
		stream, err := listen.Host.NewStream(context.Background(), connReq.Peer.PeerId, "/blockchain/1.0.0")
		if err != nil {
			log.Println(err)
			continue
		}
		// Create a buffered stream so that read and writes are non blocking.
		rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

		// Create a thread to read and write data.
		go listen.InMessageHandler(rw)
		go listen.OutMessageHandler(rw)
	}

	select {
	case self.Requests <- handleConnected{&connReq, connReq.Peer}:
	case <-self.Quit:
	}
}

// Disconnect disconnects the connection corresponding to the given connection
// id. If permanent, the connection will be retried with an increasing backoff
// duration.
func (cm *ConnManager) Disconnect(id uint64) {
	if atomic.LoadInt32(&cm.stop) != 0 {
		return
	}

	select {
	case cm.Requests <- handleDisconnected{id, true}:
	case <-cm.Quit:
	}
}

func (self ConnManager) Start() {
	// Already started?
	if atomic.AddInt32(&self.start, 1) != 1 {
		return
	}

	log.Println("Connection manager started")
	self.WaitGroup.Add(1)
	// Start handler to listent channel from connection peer
	go self.connHandler()

	// Start all the listeners so long as the caller requested them and
	// provided a callback to be invoked when connections are accepted.
	if self.Config.OnInboundAccept != nil {
		for _, listner := range self.Config.ListenerPeers {
			self.WaitGroup.Add(1)
			go self.listenHandler(listner)
		}
	}
}

// listenHandler accepts incoming connections on a given listener.  It must be
// run as a goroutine.
func (self ConnManager) listenHandler(listen peer.Peer) {
	listen.Start()
}
