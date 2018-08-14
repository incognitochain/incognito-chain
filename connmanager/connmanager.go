package connmanager

import (
	"github.com/internet-cash/prototype/peer"
	"sync"
	"log"
	"sync/atomic"
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
	start int32
	stop  int32

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

func (self ConnManager) New(cfg *Config) (*ConnManager, error) {
	cm := ConnManager{
		Config: *cfg, // Copy so caller can't mutate
		Quit:   make(chan struct{}),
	}
	return &cm, nil
}

// HandleFailedConn handles a connection failed due to a disconnect or any
// other failure. If permanent, it retries the connection after the configured
// retry duration. Otherwise, if required, it makes a new connection request.
// After maxFailedConnectionAttempts new connections will be retried after the
// configured retry duration.
func (self *ConnManager) HandleFailedConn(c *ConnReq) {
	// TODO
}

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
					log.Printf("Connected to %v", connReq)
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

func (self ConnManager) Start() {
	// Already started?
	if atomic.AddInt32(&self.start, 1) != 1 {
		return
	}

	log.Println("Connection manager started")
	self.WaitGroup.Add(1)
	// Start handler to listent channel from connection peer
	go self.connHandler()
}
