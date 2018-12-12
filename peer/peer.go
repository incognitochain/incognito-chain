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
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/wire"
	"github.com/ninjadotorg/constant/cashec"
)

// ConnState represents the state of the requested connection.
type ConnState uint8

var HEAVY_MESSAGE_SIZE = 512 * 1024
var MESSAGE_HASH_POOL_SIZE = 1000

// RemotePeer is present for libp2p node data
type Peer struct {
	messagePool       map[string]bool
	peerConnMutex     sync.Mutex
	pendingPeersMutex sync.Mutex

	// channel
	cStop           chan struct{}
	cDisconnectPeer chan *PeerConn
	cNewConn        chan *NewPeerMsg
	cNewStream      chan *NewStreamMsg
	cStopConn       chan struct{}

	Host host.Host

	TargetAddress    ma.Multiaddr
	PeerID           peer.ID
	RawAddress       string
	ListeningAddress common.SimpleAddr
	PublicKey        string

	Seed   int64
	Config Config
	Shard  byte

	PeerConns    map[string]*PeerConn
	PendingPeers map[string]*Peer

	HandleConnected    func(peerConn *PeerConn)
	HandleDisconnected func(peerConn *PeerConn)
	HandleFailed       func(peerConn *PeerConn)
}

type NewPeerMsg struct {
	Peer  *Peer
	CConn chan *PeerConn
}

type NewStreamMsg struct {
	Stream net.Stream
	CConn  chan *PeerConn
}

// config is the struct to hold configuration options useful to RemotePeer.
type Config struct {
	MessageListeners MessageListeners
	ProducerKeySet   *cashec.KeySet
	MaxOutbound      int
	MaxInbound       int
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
	OnRequestSign   func(p *PeerConn, msg *wire.MessageBlockSigReq)
	OnInvalidBlock  func(p *PeerConn, msg *wire.MessageInvalidBlock)
	OnBlockSig      func(p *PeerConn, msg *wire.MessageBlockSig)
	OnGetChainState func(p *PeerConn, msg *wire.MessageGetChainState)
	OnChainState    func(p *PeerConn, msg *wire.MessageChainState)
	//OnRegistration  func(p *PeerConn, msg *wire.MessageRegistration)
	OnSwapRequest func(p *PeerConn, msg *wire.MessageSwapRequest)
	OnSwapSig     func(p *PeerConn, msg *wire.MessageSwapSig)
	OnSwapUpdate  func(p *PeerConn, msg *wire.MessageSwapUpdate)
}

// outMsg is used to house a message to be sent along with a channel to signal
// when the message has been sent (or won't be sent due to things such as
// shutdown)
type outMsg struct {
	message  wire.Message
	doneChan chan<- struct{}
	//encoding wire.MessageEncoding
}

func (self *Peer) ReceivedHashMessage(hash string) {
	if self.messagePool == nil {
		self.messagePool = make(map[string]bool)
	}
	self.messagePool[hash] = true
	if len(self.messagePool) > MESSAGE_HASH_POOL_SIZE {
		for k, _ := range self.messagePool {
			delete(self.messagePool, k)
			break
		}
	}
}

func (self *Peer) CheckHashMessage(hash string) (bool) {
	if self.messagePool == nil {
		self.messagePool = make(map[string]bool)
	}
	ok, _ := self.messagePool[hash]
	return ok
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
	Logger.log.Infof("I am listening on %s with PEER Id - %s", fullAddr, basicHost.ID().String())
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

	self.peerConnMutex = sync.Mutex{}
	return &self, nil
}

/*
Start - start peer to begin waiting for connections from other peers
*/
func (self *Peer) Start() {
	Logger.log.Info("RemotePeer start")
	// ping to bootnode for test env
	Logger.log.Info("Set stream handler and wait for connection from other peer")
	self.Host.SetStreamHandler(ProtocolId, self.PushStream)

	go self.processConn()

	select {
	case <-self.cStop:
		close(self.cStopConn)
		Logger.log.Warnf("PEER server shutdown complete %s", self.PeerID)
		break
	}
	return
}

func (self *Peer) PushStream(stream net.Stream) {
	newStreamMsg := NewStreamMsg{
		Stream: stream,
		CConn:  nil,
	}
	self.cNewStream <- &newStreamMsg
}

func (self *Peer) PushConn(peer *Peer, cConn chan *PeerConn) {
	newPeerMsg := NewPeerMsg{
		Peer:  peer,
		CConn: cConn,
	}
	self.cNewConn <- &newPeerMsg
}

func (self *Peer) processConn() {
	for {
		select {
		case <-self.cStopConn:
			Logger.log.Info("ProcessConn QUIT")
			return
		case newPeerMsg := <-self.cNewConn:
			Logger.log.Infof("ProcessConn START CONN %s %s", newPeerMsg.Peer.PeerID, newPeerMsg.Peer.RawAddress)
			cDone := make(chan *PeerConn)
			go func(self *Peer) {
				peerConn, err := self.handleConn(newPeerMsg.Peer, cDone)
				if err != nil && peerConn == nil {
					Logger.log.Errorf("Fail in opening stream from PEER Id - %s with err: %s", self.PeerID.String(), err.Error())
				}
			}(self)
			p := <-cDone
			if newPeerMsg.CConn != nil {
				newPeerMsg.CConn <- p
			}
			Logger.log.Infof("ProcessConn END CONN %s %s", newPeerMsg.Peer.PeerID, newPeerMsg.Peer.RawAddress)
			continue
		case newStreamMsg := <-self.cNewStream:
			remotePeerID := newStreamMsg.Stream.Conn().RemotePeer()
			Logger.log.Infof("ProcessConn START STREAM %s", remotePeerID)
			cConn := make(chan *PeerConn)
			go self.handleStream(newStreamMsg.Stream, cConn)
			p := <-cConn
			if newStreamMsg.CConn != nil {
				newStreamMsg.CConn <- p
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
	self.peerConnMutex.Lock()
	for _, peerConn := range self.PeerConns {
		if !peerConn.IsOutbound {
			ret++
		}
	}
	self.peerConnMutex.Unlock()
	return ret
}

func (self *Peer) NumOutbound() int {
	ret := int(0)
	self.peerConnMutex.Lock()
	for _, peerConn := range self.PeerConns {
		if peerConn.IsOutbound {
			ret++
		}
	}
	self.peerConnMutex.Unlock()
	return ret
}

func (self *Peer) GetPeerConnByPeerID(peerID string) (*PeerConn) {
	peerConn, ok := self.PeerConns[peerID]
	if ok {
		return peerConn
	}
	return nil
}

func (self *Peer) GetPeerConnByPbk(pbk string) (*PeerConn) {
	for _, peerConn := range self.PeerConns {
		if peerConn.RemotePeer.PublicKey == pbk {
			return peerConn
		}
	}
	return nil
}

func (self *Peer) GetListPeerConnByShard(shard byte) ([]*PeerConn) {
	peerConns := make([]*PeerConn, 0)
	for _, peerConn := range self.PeerConns {
		if peerConn.RemotePeer.Shard == shard {
			peerConns = append(peerConns, peerConn)
		}
	}
	return peerConns
}

func (self *Peer) UpdateShardForPeerConn() {
	for _, peerConn := range self.PeerConns {
		_ = peerConn
	}
}

func (self *Peer) SetPeerConn(peerConn *PeerConn) {
	internalConnPeer, ok := self.PeerConns[peerConn.RemotePeer.PeerID.String()]
	if ok && internalConnPeer != peerConn {
		if internalConnPeer.IsConnected {
			internalConnPeer.Close()
		}
		Logger.log.Infof("SetPeerConn and Remove %s %s", internalConnPeer.RemotePeer.PeerID, internalConnPeer.RemotePeer.RawAddress)
	}
	self.PeerConns[peerConn.RemotePeer.PeerID.String()] = peerConn
}

func (self *Peer) RemovePeerConn(peerConn *PeerConn) {
	internalConnPeer, ok := self.PeerConns[peerConn.RemotePeer.PeerID.String()]
	if ok {
		if internalConnPeer.IsConnected {
			internalConnPeer.Close()
		}
		delete(self.PeerConns, peerConn.RemotePeer.PeerID.String())
		Logger.log.Infof("RemovePeerConn %s %s", peerConn.RemotePeer.PeerID, peerConn.RemotePeer.RawAddress)
	}
}

func (self *Peer) handleConn(peer *Peer, cConn chan *PeerConn) (*PeerConn, error) {
	Logger.log.Infof("Opening stream to PEER Id - %s", peer.RawAddress)

	_, ok := self.PeerConns[peer.PeerID.String()]
	if ok {
		Logger.log.Infof("Checked Existed PEER Id - %s", peer.RawAddress)

		if cConn != nil {
			cConn <- nil
		}
		return nil, nil
	}

	if peer.PeerID.Pretty() == self.PeerID.Pretty() {
		Logger.log.Infof("Checked Myself PEER Id - %s", peer.RawAddress)
		//self.newPeerConnectionMutex.Unlock()

		if cConn != nil {
			cConn <- nil
		}
		return nil, nil
	}

	if self.NumOutbound() >= self.Config.MaxOutbound && self.Config.MaxOutbound > 0 && !ok {
		Logger.log.Infof("Checked Max Outbound Connection PEER Id - %s", peer.RawAddress)

		//push to pending peers
		self.ConnPending(peer)

		if cConn != nil {
			cConn <- nil
		}
		return nil, nil
	}

	stream, err := self.Host.NewStream(context.Background(), peer.PeerID, ProtocolId)
	Logger.log.Info(peer, stream, err)
	if err != nil {
		if cConn != nil {
			cConn <- nil
		}
		return nil, NewPeerError(OpeningStreamP2PErr, err, self)
	}

	defer stream.Close()

	remotePeerID := stream.Conn().RemotePeer()

	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

	peerConn := PeerConn{
		IsOutbound:         true,
		RemotePeer:         peer,
		RemotePeerID:       remotePeerID,
		RemoteRawAddress:   peer.RawAddress,
		ListenerPeer:       self,
		Config:             self.Config,
		ReaderWriterStream: rw,
		cDisconnect:        make(chan struct{}),
		cClose:             make(chan struct{}),
		cRead:              make(chan struct{}),
		cWrite:             make(chan struct{}),
		cMsgHash:           make(map[string]chan bool),
		sendMessageQueue:   make(chan outMsg),
		HandleConnected:    self.handleConnected,
		HandleDisconnected: self.handleDisconnected,
		HandleFailed:       self.handleFailed,
	}

	self.SetPeerConn(&peerConn)

	go peerConn.InMessageHandler(rw)
	go peerConn.OutMessageHandler(rw)

	peerConn.RetryCount = 0
	peerConn.updateConnState(ConnEstablished)

	go self.handleConnected(&peerConn)

	if cConn != nil {
		cConn <- &peerConn
	}

	for {
		select {
		case <-peerConn.cDisconnect:
			Logger.log.Infof("NewPeerConnection Disconnected Stream PEER Id %s", peerConn.RemotePeerID.String())
			return &peerConn, nil
		case <-peerConn.cClose:
			Logger.log.Infof("NewPeerConnection closed stream PEER Id %s", peerConn.RemotePeerID.String())
			go func() {
				select {
				case <-peerConn.cDisconnect:
					Logger.log.Infof("NewPeerConnection disconnected after closed stream PEER Id %s", peerConn.RemotePeerID.String())
					return
				}
			}()
			return &peerConn, nil
		}
	}

	return &peerConn, nil
}

func (self *Peer) handleStream(stream net.Stream, cDone chan *PeerConn) {
	// Remember to close the stream when we are done.
	defer stream.Close()

	if self.NumInbound() >= self.Config.MaxInbound && self.Config.MaxInbound > 0 {
		Logger.log.Infof("Max RemotePeer Inbound Connection")

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
		RemotePeer: &Peer{
			PeerID: remotePeerID,
		},
		Config:             self.Config,
		RemotePeerID:       remotePeerID,
		ReaderWriterStream: rw,
		cDisconnect:        make(chan struct{}),
		cClose:             make(chan struct{}),
		cRead:              make(chan struct{}),
		cWrite:             make(chan struct{}),
		cMsgHash:           make(map[string]chan bool),
		sendMessageQueue:   make(chan outMsg),
		HandleConnected:    self.handleConnected,
		HandleDisconnected: self.handleDisconnected,
		HandleFailed:       self.handleFailed,
	}

	self.SetPeerConn(&peerConn)

	go peerConn.InMessageHandler(rw)
	go peerConn.OutMessageHandler(rw)

	peerConn.RetryCount = 0
	peerConn.updateConnState(ConnEstablished)

	go self.handleConnected(&peerConn)

	if cDone != nil {
		close(cDone)
	}

	for {
		select {
		case <-peerConn.cDisconnect:
			Logger.log.Infof("HandleStream disconnected stream PEER Id %s", peerConn.RemotePeerID.String())
			return
		case <-peerConn.cClose:
			Logger.log.Infof("HandleStream closed stream PEER Id %s", peerConn.RemotePeerID.String())
			go func() {
				select {
				case <-peerConn.cDisconnect:
					Logger.log.Infof("HandleStream disconnected after closed stream PEER Id %s", peerConn.RemotePeerID.String())
					return
				}
			}()
			return
		}
	}
}

// QueueMessageWithEncoding adds the passed Constant message to the peer send
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
	Logger.log.Infof("Stopping PEER %s", self.PeerID.String())

	self.Host.Close()
	self.peerConnMutex.Lock()
	for _, peerConn := range self.PeerConns {
		peerConn.updateConnState(ConnCanceled)
	}
	self.peerConnMutex.Unlock()

	close(self.cStop)
	Logger.log.Infof("PEER %s stopped", self.PeerID.String())
}

/*
handleConnected - set established flag to a peer when being connected
*/
func (self *Peer) handleConnected(peerConn *PeerConn) {
	Logger.log.Infof("handleConnected %s", peerConn.RemotePeerID.String())
	peerConn.RetryCount = 0
	peerConn.updateConnState(ConnEstablished)

	self.ConnEstablished(peerConn.RemotePeer)

	if self.HandleConnected != nil {
		self.HandleConnected(peerConn)
	}
}

/*
handleDisconnected - handle connected peer when it is disconnected, remove and retry connection
*/
func (self *Peer) handleDisconnected(peerConn *PeerConn) {
	Logger.log.Infof("handleDisconnected %s", peerConn.RemotePeerID.String())
	peerConn.updateConnState(ConnCanceled)
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
	Logger.log.Infof("handleFailed %s", peerConn.RemotePeerID.String())

	self.ConnCanceled(peerConn.RemotePeer)

	if self.HandleFailed != nil {
		self.HandleFailed(peerConn)
	}
}

/*
retryPeerConnection - retry to connect to peer when being disconnected
*/
func (self *Peer) retryPeerConnection(peerConn *PeerConn) {
	time.AfterFunc(RetryConnDuration, func() {
		Logger.log.Infof("Retry New RemotePeer Connection %s", peerConn.RemoteRawAddress)
		peerConn.RetryCount += 1

		if peerConn.RetryCount < MaxRetryConn {
			peerConn.updateConnState(ConnPending)
			cConn := make(chan *PeerConn)
			peerConn.ListenerPeer.PushConn(peerConn.RemotePeer, cConn)
			p := <-cConn
			if p == nil {
				peerConn.RetryCount++
				go self.retryPeerConnection(peerConn)
			}
		} else {
			peerConn.updateConnState(ConnCanceled)
			self.ConnCanceled(peerConn.RemotePeer)
			self.renewPeerConnection()
			self.ConnPending(peerConn.RemotePeer)
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
			Logger.log.Infof("---> RemotePeer: ", peer.RawAddress)
			go self.PushConn(peer, nil)
		}
		Logger.log.Infof("*end - Creating peer conn to %d pending peers", len(self.PendingPeers))
		self.pendingPeersMutex.Unlock()
	}
}
