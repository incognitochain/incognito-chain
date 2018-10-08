package peer

import (
	"bufio"
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"log"
	mrand "math/rand"
	"strings"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-crypto"
	"github.com/libp2p/go-libp2p-host"
	"github.com/libp2p/go-libp2p-net"
	"github.com/libp2p/go-libp2p-peer"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/ninjadotorg/cash-prototype/common"
	"github.com/ninjadotorg/cash-prototype/wire"
)

// ConnState represents the state of the requested connection.
type ConnState uint8

// Peer is present for libp2p node data
type Peer struct {
	Host host.Host

	TargetAddress    ma.Multiaddr
	PeerID           peer.ID
	RawAddress       string
	ListeningAddress common.SimpleAddr
	PublicKey        string

	Seed        int64
	Config      Config
	MaxOutbound int
	MaxInbound  int

	PeerConns         map[string]*PeerConn
	peerConnsMutex    sync.Mutex
	PendingPeers      map[string]*Peer
	pendingPeersMutex sync.Mutex

	cStop           chan struct{}
	cDisconnectPeer chan *PeerConn
	cNewConn        chan *NewPeerMsg
	cNewStream      chan *NewStreamMsg
	cStopConn       chan struct{}

	HandleConnected    func(peerConn *PeerConn)
	HandleDisconnected func(peerConn *PeerConn)
	HandleFailed       func(peerConn *PeerConn)
}

type NewPeerMsg struct {
	Peer *Peer
	Done chan struct{}
}

type NewStreamMsg struct {
	Stream net.Stream
	Done   chan struct{}
}

// config is the struct to hold configuration options useful to Peer.
type Config struct {
	MessageListeners MessageListeners
	SealerPrvKey     string
}

type WrappedStream struct {
	Stream net.Stream
	Writer *bufio.Writer
	Reader *bufio.Reader
}

/*
// MessageListeners defines callback function pointers to invoke with message
// listeners for a peer. Any listener which is not set to a concrete callback
// during peer initialization is ignored. Execution of multiple message
// listeners occurs serially, so one callback blocks the execution of the next.
//
// NOTE: Unless otherwise documented, these listeners must NOT directly call any
// blocking calls (such as WaitForShutdown) on the peer instance since the input
// handler goroutine blocks until the callback has completed.  Doing so will
// result in a deadlock.
*/
type MessageListeners struct {
	OnTx        func(p *PeerConn, msg *wire.MessageTx)
	OnBlock     func(p *PeerConn, msg *wire.MessageBlock)
	OnGetBlocks func(p *PeerConn, msg *wire.MessageGetBlocks)
	OnVersion   func(p *PeerConn, msg *wire.MessageVersion)
	OnVerAck    func(p *PeerConn, msg *wire.MessageVerAck)
	OnGetAddr   func(p *PeerConn, msg *wire.MessageGetAddr)
	OnAddr      func(p *PeerConn, msg *wire.MessageAddr)

	//PoS
	OnRequestSign   func(p *PeerConn, msg *wire.MessageRequestSign)
	OnInvalidBlock  func(p *PeerConn, msg *wire.MessageInvalidBlock)
	OnBlockSig      func(p *PeerConn, msg *wire.MessageBlockSig)
	OnGetChainState func(p *PeerConn, msg *wire.MessageGetChainState)
	OnChainState    func(p *PeerConn, msg *wire.MessageChainState)
}

// outMsg is used to house a message to be sent along with a channel to signal
// when the message has been sent (or won't be sent due to things such as
// shutdown)
type outMsg struct {
	message  wire.Message
	doneChan chan<- struct{}
	//encoding wire.MessageEncoding
}

/*
NewPeer - create a new peer with go libp2p
 */
func (self Peer) NewPeer() (*Peer, error) {
	// If the seed is zero, use real cryptographic randomness. Otherwise, use a
	// deterministic randomness source to make generated keys stay the same
	// across multiple runs
	var r io.Reader
	if self.Seed == 0 {
		r = rand.Reader
	} else {
		r = mrand.New(mrand.NewSource(self.Seed))
	}

	// Generate a key pair for this Host. We will use it
	// to obtain a valid Host Id.
	priv, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r)
	if err != nil {
		return &self, NewPeerError(PeerGenerateKeyPairErr, err, &self)
	}

	ip := strings.Split(self.ListeningAddress.String(), ":")[0]
	if len(ip) == 0 {
		ip = LocalHost
	}
	Logger.log.Info(ip)
	port := strings.Split(self.ListeningAddress.String(), ":")[1]
	net := self.ListeningAddress.Network()
	listeningAddressString := fmt.Sprintf("/%s/%s/tcp/%s", net, ip, port)
	opts := []libp2p.Option{
		libp2p.ListenAddrStrings(listeningAddressString),
		libp2p.Identity(priv),
	}

	basicHost, err := libp2p.New(context.Background(), opts...)
	if err != nil {
		return &self, NewPeerError(CreateP2PNodeErr, err, &self)
	}

	// Build Host multiaddress
	mulAddrStr := fmt.Sprintf("/ipfs/%s", basicHost.ID().Pretty())

	hostAddr, err := ma.NewMultiaddr(mulAddrStr)
	if err != nil {
		return &self, NewPeerError(CreateP2PAddressErr, err, &self)
	}

	// Now we can build a full multiaddress to reach this Host
	// by encapsulating both addresses:
	addr := basicHost.Addrs()[0]
	fullAddr := addr.Encapsulate(hostAddr)
	Logger.log.Infof("I am listening on %s with PEER Id - %s\n", fullAddr, basicHost.ID().String())
	pid, err := fullAddr.ValueForProtocol(ma.P_IPFS)
	if err != nil {
		return &self, NewPeerError(GetPeerIdFromProtocolErr, err, &self)
	}
	peerID, err := peer.IDB58Decode(pid)
	if err != nil {
		log.Print(err)
		return &self, NewPeerError(GetPeerIdFromProtocolErr, err, &self)
	}

	self.RawAddress = fullAddr.String()
	self.Host = basicHost
	self.TargetAddress = fullAddr
	self.PeerID = peerID
	self.cStop = make(chan struct{}, 1)
	self.cDisconnectPeer = make(chan *PeerConn)
	self.cNewConn = make(chan *NewPeerMsg)
	self.cNewStream = make(chan *NewStreamMsg)
	self.cStopConn = make(chan struct{})

	self.peerConnsMutex = sync.Mutex{}
	return &self, nil
}

/*
Start - start peer to begin waiting for connections from other peers
 */
func (self *Peer) Start() {
	Logger.log.Info("Peer start")
	// ping to bootnode for test env
	Logger.log.Info("Set stream handler and wait for connection from other peer")
	self.Host.SetStreamHandler(ProtocolId, self.PushStream)

	go self.ProcessConn()

	select {
	case <-self.cStop:
		close(self.cStopConn)
		Logger.log.Infof("PEER server shutdown complete %s", self.PeerID)
		break
	}
	return
}

func (self *Peer) PushStream(stream net.Stream) {
	newStreamMsg := NewStreamMsg{
		Stream: stream,
		Done:   nil,
	}
	self.cNewStream <- &newStreamMsg
}

func (self *Peer) PushConn(peer *Peer, done chan struct{}) {
	newPeerMsg := NewPeerMsg{
		Peer: peer,
		Done: done,
	}
	self.cNewConn <- &newPeerMsg
}

func (self *Peer) ProcessConn() {
	for {
		select {
		case <-self.cStopConn:
			Logger.log.Info("ProcessConn QUIT")
			return
		case newPeerMsg := <-self.cNewConn:
			Logger.log.Infof("ProcessConn START CONN %s %s", newPeerMsg.Peer.PeerID, newPeerMsg.Peer.RawAddress)
			cDone := make(chan struct{})
			go func(self *Peer) {
				peerConn, err := self.handleConn(newPeerMsg.Peer, cDone)
				if err != nil {
					Logger.log.Errorf("Fail in opening stream from PEER Id - %s with err: %s", self.PeerID.String(), err.Error())
				} else {
					Logger.log.Errorf("Success in opening stream from PEER Id - %s to %s", self.PeerID.String(), peerConn.RawAddress)
				}
			}(self)
			<-cDone
			if newPeerMsg.Done != nil {
				close(newPeerMsg.Done)
			}
			Logger.log.Infof("ProcessConn END CONN %s %s", newPeerMsg.Peer.PeerID, newPeerMsg.Peer.RawAddress)
			continue
		case newStreamMsg := <-self.cNewStream:
			remotePeerID := newStreamMsg.Stream.Conn().RemotePeer()
			Logger.log.Infof("ProcessConn START STREAM %s", remotePeerID)
			cDone := make(chan struct{})
			go self.HandleStream(newStreamMsg.Stream, cDone)
			<-cDone
			if newStreamMsg.Done != nil {
				close(newStreamMsg.Done)
			}
			Logger.log.Infof("ProcessConn END STREAM %s", remotePeerID)
			continue
		}
	}
	return
}

func (self *Peer) ConnPending(peer *Peer) {
	self.pendingPeersMutex.Lock()
	self.PendingPeers[peer.PeerID.String()] = peer
	self.pendingPeersMutex.Unlock()
}

func (self *Peer) ConnEstablished(peer *Peer) {
	self.pendingPeersMutex.Lock()
	_, ok := self.PendingPeers[peer.PeerID.String()]
	if ok {
		delete(self.PendingPeers, peer.PeerID.String())
	}
	self.pendingPeersMutex.Unlock()
}

func (self *Peer) ConnCanceled(peer *Peer) {
	_, ok := self.PeerConns[peer.PeerID.String()]
	if ok {
		delete(self.PeerConns, peer.PeerID.String())
	}
	self.pendingPeersMutex.Lock()
	self.PendingPeers[peer.PeerID.String()] = peer
	self.pendingPeersMutex.Unlock()
}

func (self *Peer) NumInbound() int {
	ret := int(0)
	self.peerConnsMutex.Lock()
	for _, peerConn := range self.PeerConns {
		if !peerConn.IsOutbound {
			ret++
		}
	}
	self.peerConnsMutex.Unlock()
	return ret
}

func (self *Peer) NumOutbound() int {
	ret := int(0)
	self.peerConnsMutex.Lock()
	for _, peerConn := range self.PeerConns {
		if peerConn.IsOutbound {
			ret++
		}
	}
	self.peerConnsMutex.Unlock()
	return ret
}

func (self *Peer) SetPeerConn(peerConn *PeerConn) {
	_peerConn, ok := self.PeerConns[peerConn.Peer.PeerID.String()]
	if ok && _peerConn != peerConn {
		if _peerConn.IsConnected {
			_peerConn.Close()
		}
		Logger.log.Infof("SetPeerConn and Remove %s %s", _peerConn.Peer.PeerID, _peerConn.Peer.RawAddress)
	}
	self.PeerConns[peerConn.Peer.PeerID.String()] = peerConn
}

func (self *Peer) RemovePeerConn(peerConn *PeerConn) {
	_peerConn, ok := self.PeerConns[peerConn.Peer.PeerID.String()]
	if ok {
		if _peerConn.IsConnected {
			_peerConn.Close()
		}
		delete(self.PeerConns, peerConn.Peer.PeerID.String())
		Logger.log.Infof("RemovePeerConn %s %s", peerConn.Peer.PeerID, peerConn.Peer.RawAddress)
	}
}

func (self *Peer) handleConn(peer *Peer, cDone chan struct{}) (*PeerConn, error) {
	Logger.log.Infof("Opening stream to PEER Id - %s \n", peer.RawAddress)

	_, ok := self.PeerConns[peer.PeerID.String()]
	if ok {
		Logger.log.Infof("Checked Existed PEER Id - %s", peer.RawAddress)

		if cDone != nil {
			close(cDone)
		}
		return nil, nil
	}

	if peer.PeerID.Pretty() == self.PeerID.Pretty() {
		Logger.log.Infof("Checked Myself PEER Id - %s", peer.RawAddress)
		//self.newPeerConnectionMutex.Unlock()

		if cDone != nil {
			close(cDone)
		}
		return nil, nil
	}

	if self.NumOutbound() >= self.MaxOutbound && self.MaxOutbound > 0 && !ok {
		Logger.log.Infof("Checked Max Outbound Connection PEER Id - %s", peer.RawAddress)

		//push to pending peers
		self.ConnPending(peer)

		if cDone != nil {
			close(cDone)
		}
		return nil, nil
	}

	stream, err := self.Host.NewStream(context.Background(), peer.PeerID, ProtocolId)
	Logger.log.Info(peer, stream, err)
	if err != nil {
		if cDone != nil {
			close(cDone)
		}
		return nil, NewPeerError(OpeningStreamP2PErr, err, self)
	}

	defer stream.Close()

	remotePeerID := stream.Conn().RemotePeer()

	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

	peerConn := PeerConn{
		IsOutbound:         true,
		Peer:               peer,
		ListenerPeer:       self,
		Config:             self.Config,
		PeerID:             remotePeerID,
		ReaderWriterStream: rw,
		cDisconnect:        make(chan struct{}),
		cClose:             make(chan struct{}),
		cRead:              make(chan struct{}),
		cWrite:             make(chan struct{}),
		sendMessageQueue:   make(chan outMsg),
		HandleConnected:    self.handleConnected,
		HandleDisconnected: self.handleDisconnected,
		HandleFailed:       self.handleFailed,
	}

	self.SetPeerConn(&peerConn)

	go peerConn.InMessageHandler(rw)
	go peerConn.OutMessageHandler(rw)

	peerConn.RetryCount = 0
	peerConn.updateState(ConnEstablished)

	go self.handleConnected(&peerConn)

	if cDone != nil {
		close(cDone)
	}

	for {
		select {
		case <-peerConn.cDisconnect:
			Logger.log.Infof("NewPeerConnection Disconnected Stream PEER Id %s", peerConn.PeerID.String())
			return &peerConn, nil
		case <-peerConn.cClose:
			Logger.log.Infof("NewPeerConnection closed stream PEER Id %s", peerConn.PeerID.String())
			go func() {
				select {
				case <-peerConn.cDisconnect:
					Logger.log.Infof("NewPeerConnection disconnected after closed stream PEER Id %s", peerConn.PeerID.String())
					return
				}
			}()
			return &peerConn, nil
		}
	}

	return &peerConn, nil
}

func (self *Peer) HandleStream(stream net.Stream, cDone chan struct{}) {
	// Remember to close the stream when we are done.
	defer stream.Close()

	if self.NumInbound() >= self.MaxInbound && self.MaxInbound > 0 {
		Logger.log.Infof("Max Peer Inbound Connection")

		if cDone != nil {
			close(cDone)
		}
		return
	}

	remotePeerID := stream.Conn().RemotePeer()
	Logger.log.Infof("PEER %s Received a new stream from OTHER PEER with Id %s", self.Host.ID().String(), remotePeerID.String())
	_, ok := self.PeerConns[remotePeerID.String()]
	if ok {
		Logger.log.Infof("Received a new stream existed PEER Id - %s", remotePeerID)
		//self.newPeerConnectionMutex.Unlock()

		if cDone != nil {
			close(cDone)
		}
		return
	}

	// Create a buffer stream for non blocking read and write.
	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

	peerConn := PeerConn{
		IsOutbound:   false,
		ListenerPeer: self,
		Peer: &Peer{
			PeerID: remotePeerID,
		},
		Config:             self.Config,
		PeerID:             remotePeerID,
		ReaderWriterStream: rw,
		cDisconnect:        make(chan struct{}),
		cClose:             make(chan struct{}),
		cRead:              make(chan struct{}),
		cWrite:             make(chan struct{}),
		sendMessageQueue:   make(chan outMsg),
		HandleConnected:    self.handleConnected,
		HandleDisconnected: self.handleDisconnected,
		HandleFailed:       self.handleFailed,
	}

	self.SetPeerConn(&peerConn)
	//self.newPeerConnectionMutex.Unlock()

	go peerConn.InMessageHandler(rw)
	go peerConn.OutMessageHandler(rw)

	peerConn.RetryCount = 0
	peerConn.updateState(ConnEstablished)

	go self.handleConnected(&peerConn)

	if cDone != nil {
		close(cDone)
	}

	for {
		select {
		case <-peerConn.cDisconnect:
			Logger.log.Infof("HandleStream disconnected stream PEER Id %s", peerConn.PeerID.String())
			return
		case <-peerConn.cClose:
			Logger.log.Infof("HandleStream closed stream PEER Id %s", peerConn.PeerID.String())
			go func() {
				select {
				case <-peerConn.cDisconnect:
					Logger.log.Infof("HandleStream disconnected after closed stream PEER Id %s", peerConn.PeerID.String())
					return
				}
			}()
			return
		}
	}
}

// QueueMessageWithEncoding adds the passed bitcoin message to the peer send
// queue. This function is identical to QueueMessage, however it allows the
// caller to specify the wire encoding type that should be used when
// encoding/decoding blocks and transactions.
//
// This function is safe for concurrent access.
func (self *Peer) QueueMessageWithEncoding(msg wire.Message, doneChan chan<- struct{}) {
	for _, peerConnection := range self.PeerConns {
		go peerConnection.QueueMessageWithEncoding(msg, doneChan)
	}
}

func (self *Peer) Stop() {
	Logger.log.Infof("PEER %s Stop", self.PeerID.String())

	self.Host.Close()
	self.peerConnsMutex.Lock()
	for _, peerConn := range self.PeerConns {
		peerConn.updateState(ConnCanceled)
	}
	self.peerConnsMutex.Unlock()

	close(self.cStop)
}

/*
handleConnected - set established flag to a peer when being connected
 */
func (self *Peer) handleConnected(peerConn *PeerConn) {
	Logger.log.Infof("handleConnected %s", peerConn.PeerID.String())
	peerConn.RetryCount = 0
	peerConn.updateState(ConnEstablished)

	self.ConnEstablished(peerConn.Peer)

	if self.HandleConnected != nil {
		self.HandleConnected(peerConn)
	}
}

/*
handleDisconnected - handle connected peer when it is disconnected, remove and retry connection
 */
func (self *Peer) handleDisconnected(peerConn *PeerConn) {
	Logger.log.Infof("handleDisconnected %s", peerConn.PeerID.String())
	peerConn.updateState(ConnCanceled)
	self.RemovePeerConn(peerConn)
	if peerConn.IsOutbound {
		go self.retryPeerConnection(peerConn)
	}

	if self.HandleDisconnected != nil {
		self.HandleDisconnected(peerConn)
	}
}

/*
handleFailed - handle when connecting peer failure
 */
func (self *Peer) handleFailed(peerConn *PeerConn) {
	Logger.log.Infof("handleFailed %s", peerConn.PeerID.String())

	self.ConnCanceled(peerConn.Peer)

	if self.HandleFailed != nil {
		self.HandleFailed(peerConn)
	}
}

/*
retryPeerConnection - retry to connect to peer when being disconnected
 */
func (self *Peer) retryPeerConnection(peerConn *PeerConn) {
	time.AfterFunc(RetryConnDuration, func() {
		Logger.log.Infof("Retry New Peer Connection %s", peerConn.RawAddress)
		peerConn.RetryCount += 1

		if peerConn.RetryCount < MaxRetryConn {
			peerConn.updateState(ConnPending)
			peerConn.ListenerPeer.PushConn(peerConn.Peer, nil)
		} else {
			peerConn.updateState(ConnCanceled)
			self.ConnCanceled(peerConn.Peer)
			self.renewPeerConnection()
			self.ConnPending(peerConn.Peer)
		}
	})
}

/*
renewPeerConnection - create peer conn by goroutines for pending peers(reconnect)
 */
func (self *Peer) renewPeerConnection() {
	if len(self.PendingPeers) > 0 {
		self.pendingPeersMutex.Lock()
		Logger.log.Infof("*start - Creating peer conn to %d pending peers", len(self.PendingPeers))
		for _, peer := range self.PendingPeers {
			Logger.log.Infof("---> Peer: ", peer.RawAddress)
			go self.PushConn(peer, nil)
		}
		Logger.log.Infof("*end - Creating peer conn to %d pending peers", len(self.PendingPeers))
		self.pendingPeersMutex.Unlock()
	}
}
