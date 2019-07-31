package peer

import (
	"bufio"
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"log"
	mrand "math/rand"
	"strings"
	"sync"
	"time"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/wire"
	"github.com/libp2p/go-libp2p"
	crypto "github.com/libp2p/go-libp2p-crypto"
	host "github.com/libp2p/go-libp2p-host"
	net "github.com/libp2p/go-libp2p-net"
	peer "github.com/libp2p/go-libp2p-peer"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/patrickmn/go-cache"
)

// ConnState represents the state of the requested connection.
type ConnState uint8

// RemotePeer is present for libp2p node data
type Peer struct {
	messagePoolNew *cache.Cache

	// channel
	cStop           chan struct{}
	cDisconnectPeer chan *PeerConn
	cNewConn        chan *newPeerMsg
	cNewStream      chan *newStreamMsg
	cStopConn       chan struct{}

	Host host.Host
	Port string

	TargetAddress    ma.Multiaddr
	PeerID           peer.ID
	RawAddress       string
	ListeningAddress common.SimpleAddr
	PublicKey        string

	Seed   int64
	Config Config
	Shard  byte

	PeerConns       map[string]*PeerConn
	PeerConnsMtx    sync.Mutex
	PendingPeers    map[string]*Peer
	pendingPeersMtx sync.Mutex

	HandleConnected    func(peerConn *PeerConn)
	HandleDisconnected func(peerConn *PeerConn)
	HandleFailed       func(peerConn *PeerConn)
}

// config is the struct to hold configuration options useful to RemotePeer.
type Config struct {
	MessageListeners MessageListeners
	UserKeySet       *incognitokey.KeySet
	MaxOutPeers      int
	MaxInPeers       int
	MaxPeers         int
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
	OnTx               func(p *PeerConn, msg *wire.MessageTx)
	OnTxToken          func(p *PeerConn, msg *wire.MessageTxToken)
	OnTxPrivacyToken   func(p *PeerConn, msg *wire.MessageTxPrivacyToken)
	OnBlockShard       func(p *PeerConn, msg *wire.MessageBlockShard)
	OnBlockBeacon      func(p *PeerConn, msg *wire.MessageBlockBeacon)
	OnCrossShard       func(p *PeerConn, msg *wire.MessageCrossShard)
	OnShardToBeacon    func(p *PeerConn, msg *wire.MessageShardToBeacon)
	OnGetBlockBeacon   func(p *PeerConn, msg *wire.MessageGetBlockBeacon)
	OnGetBlockShard    func(p *PeerConn, msg *wire.MessageGetBlockShard)
	OnGetCrossShard    func(p *PeerConn, msg *wire.MessageGetCrossShard)
	OnGetShardToBeacon func(p *PeerConn, msg *wire.MessageGetShardToBeacon)
	OnVersion          func(p *PeerConn, msg *wire.MessageVersion)
	OnVerAck           func(p *PeerConn, msg *wire.MessageVerAck)
	OnGetAddr          func(p *PeerConn, msg *wire.MessageGetAddr)
	OnAddr             func(p *PeerConn, msg *wire.MessageAddr)

	//PBFT
	OnBFTMsg             func(p *PeerConn, msg wire.Message)
	OnPeerState          func(p *PeerConn, msg *wire.MessagePeerState)
	PushRawBytesToShard  func(p *PeerConn, msgBytes *[]byte, shard byte) error
	PushRawBytesToBeacon func(p *PeerConn, msgBytes *[]byte) error
	GetCurrentRoleShard  func() (string, *byte)
}

func (peerObj *Peer) HashToPool(hash string) error {
	if peerObj.messagePoolNew == nil {
		peerObj.messagePoolNew = cache.New(messageLiveTime, messageCleanupInterval)
	}
	return peerObj.messagePoolNew.Add(hash, 1, messageLiveTime)
}

func (peerObj *Peer) CheckHashPool(hash string) bool {
	_, expiredT, exist := peerObj.messagePoolNew.GetWithExpiration(hash)
	if exist {
		if (expiredT != time.Time{}) {
			return true
		}
	}
	return false
}

/*
NewPeer - create a new peer with go libp2p
*/
func (peerObj Peer) NewPeer() (*Peer, error) {
	// If the seed is zero, use real cryptographic randomness. Otherwise, use a
	// deterministic randomness source to make generated keys stay the same
	// across multiple runs
	var r io.Reader
	if peerObj.Seed == 0 {
		r = rand.Reader
	} else {
		r = mrand.New(mrand.NewSource(peerObj.Seed))
	}

	// Generate a key pair for this Host. We will use it
	// to obtain a valid Host Id.
	priv, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r)
	if err != nil {
		return &peerObj, NewPeerError(PeerGenerateKeyPairError, err, &peerObj)
	}

	ip := strings.Split(peerObj.ListeningAddress.String(), ":")[0]
	if len(ip) == 0 {
		ip = localHost
	}
	Logger.log.Info(ip)
	port := strings.Split(peerObj.ListeningAddress.String(), ":")[1]
	net := peerObj.ListeningAddress.Network()
	listeningAddressString := fmt.Sprintf("/%s/%s/tcp/%s", net, ip, port)
	opts := []libp2p.Option{
		libp2p.ListenAddrStrings(listeningAddressString),
		libp2p.Identity(priv),
	}

	basicHost, err := libp2p.New(context.Background(), opts...)
	if err != nil {
		return &peerObj, NewPeerError(CreateP2PNodeError, err, &peerObj)
	}

	// Build Host multiaddress
	mulAddrStr := fmt.Sprintf("/ipfs/%s", basicHost.ID().Pretty())

	hostAddr, err := ma.NewMultiaddr(mulAddrStr)
	if err != nil {
		return &peerObj, NewPeerError(CreateP2PAddressError, err, &peerObj)
	}

	// Now we can build a full multiaddress to reach this Host
	// by encapsulating both addresses:
	addr := basicHost.Addrs()[0]
	fullAddr := addr.Encapsulate(hostAddr)
	rawAddress := fmt.Sprintf("%s%s", listeningAddressString, mulAddrStr)
	Logger.log.Infof("I am listening on %s with PEER Id - %s", rawAddress, basicHost.ID().Pretty())
	pid, err := fullAddr.ValueForProtocol(ma.P_IPFS)
	if err != nil {
		return &peerObj, NewPeerError(GetPeerIdFromProtocolError, err, &peerObj)
	}
	peerID, err := peer.IDB58Decode(pid)
	if err != nil {
		log.Print(err)
		return &peerObj, NewPeerError(GetPeerIdFromProtocolError, err, &peerObj)
	}

	peerObj.RawAddress = rawAddress
	peerObj.Host = basicHost
	peerObj.Port = port
	peerObj.TargetAddress = fullAddr
	peerObj.PeerID = peerID
	peerObj.cStop = make(chan struct{}, 1)
	peerObj.cDisconnectPeer = make(chan *PeerConn)
	peerObj.cNewConn = make(chan *newPeerMsg)
	peerObj.cNewStream = make(chan *newStreamMsg)
	peerObj.cStopConn = make(chan struct{})

	peerObj.PeerConnsMtx = sync.Mutex{}
	peerObj.pendingPeersMtx = sync.Mutex{}
	return &peerObj, nil
}

// Start - start peer to begin waiting for connections from other peers
func (peerObj *Peer) Start() {
	Logger.log.Info("RemotePeer start")
	// ping to bootnode for test env
	Logger.log.Info("Set stream handler and wait for connection from other peer")
	peerObj.Host.SetStreamHandler(protocolID, peerObj.PushStream)

	go peerObj.processConn()

	_, ok := <-peerObj.cStop
	if !ok { // stop
		close(peerObj.cStopConn)
		Logger.log.Criticalf("PEER server shutdown complete %s", peerObj.PeerID)
	}
}

func (peerObj *Peer) PushStream(stream net.Stream) {
	go func(stream net.Stream) {
		newStreamMsg := newStreamMsg{
			stream: stream,
			cConn:  nil,
		}
		peerObj.cNewStream <- &newStreamMsg
	}(stream)
}

func (peerObj *Peer) PushConn(peer *Peer, cConn chan *PeerConn) {
	go func(peer *Peer, cConn chan *PeerConn) {
		newPeerMsg := newPeerMsg{
			peer:  peer,
			cConn: cConn,
		}
		peerObj.cNewConn <- &newPeerMsg
	}(peer, cConn)
}

// processConn - control all channel which correspond to connection and process
func (peerObj *Peer) processConn() {
	for {
		select {
		case <-peerObj.cStopConn:
			Logger.log.Critical("ProcessConn QUIT")
			return
		case newPeerMsg := <-peerObj.cNewConn:
			Logger.log.Infof("ProcessConn START CONN %s %s", newPeerMsg.peer.PeerID.Pretty(), newPeerMsg.peer.RawAddress)
			cConn := make(chan *PeerConn)
			go func(peerObj *Peer) {
				peerConn, err := peerObj.handleNewConnectionOut(newPeerMsg.peer, cConn)
				if err != nil && peerConn == nil {
					Logger.log.Errorf("Fail in opening stream from PEER Id - %s with err: %s", peerObj.PeerID.Pretty(), err.Error())
				}
			}(peerObj)
			p := <-cConn
			if newPeerMsg.cConn != nil {
				newPeerMsg.cConn <- p
			}
			Logger.log.Infof("ProcessConn END CONN %s %s", newPeerMsg.peer.PeerID.Pretty(), newPeerMsg.peer.RawAddress)
			continue
		case newStreamMsg := <-peerObj.cNewStream:
			remotePeerID := newStreamMsg.stream.Conn().RemotePeer()
			Logger.log.Infof("ProcessConn START STREAM %s", remotePeerID.Pretty())
			cConn := make(chan *PeerConn)
			go peerObj.handleNewStreamIn(newStreamMsg.stream, cConn)
			p := <-cConn
			if newStreamMsg.cConn != nil {
				newStreamMsg.cConn <- p
			}
			Logger.log.Infof("ProcessConn END STREAM %s", remotePeerID.Pretty())
			continue
		}
	}
}

func (peerObj *Peer) connPending(peer *Peer) {
	peerObj.pendingPeersMtx.Lock()
	defer peerObj.pendingPeersMtx.Unlock()
	peerIDStr := peer.PeerID.Pretty()
	peerObj.PendingPeers[peerIDStr] = peer
}

func (peerObj *Peer) connEstablished(peer *Peer) {
	peerObj.pendingPeersMtx.Lock()
	defer peerObj.pendingPeersMtx.Unlock()
	peerIDStr := peer.PeerID.Pretty()
	_, ok := peerObj.PendingPeers[peerIDStr]
	if ok {
		delete(peerObj.PendingPeers, peerIDStr)
	}
}

func (peerObj *Peer) connCanceled(peer *Peer) {
	peerObj.PeerConnsMtx.Lock()
	peerObj.pendingPeersMtx.Lock()
	defer func() {
		peerObj.PeerConnsMtx.Unlock()
		peerObj.pendingPeersMtx.Unlock()
	}()
	peerIDStr := peer.PeerID.Pretty()
	_, ok := peerObj.PeerConns[peerIDStr]
	if ok {
		delete(peerObj.PeerConns, peerIDStr)
	}
	peerObj.PendingPeers[peerIDStr] = peer
}

func (peerObj *Peer) countOfInboundConn() int {
	peerObj.PeerConnsMtx.Lock()
	defer peerObj.PeerConnsMtx.Unlock()
	ret := int(0)
	for _, peerConn := range peerObj.PeerConns {
		if !peerConn.GetIsOutbound() {
			ret++
		}
	}
	return ret
}

func (peerObj *Peer) countOfOutboundConn() int {
	peerObj.PeerConnsMtx.Lock()
	defer peerObj.PeerConnsMtx.Unlock()
	ret := int(0)
	for _, peerConn := range peerObj.PeerConns {
		if peerConn.GetIsOutbound() {
			ret++
		}
	}
	return ret
}

func (peerObj *Peer) GetPeerConnByPeerID(peerID string) *PeerConn {
	peerObj.PeerConnsMtx.Lock()
	defer peerObj.PeerConnsMtx.Unlock()
	peerConn, ok := peerObj.PeerConns[peerID]
	if ok {
		return peerConn
	}
	return nil
}

func (peerObj *Peer) setPeerConn(peerConn *PeerConn) {
	peerObj.PeerConnsMtx.Lock()
	defer peerObj.PeerConnsMtx.Unlock()
	peerIDStr := peerConn.RemotePeer.PeerID.Pretty()
	internalConnPeer, ok := peerObj.PeerConns[peerIDStr]
	if ok {
		if internalConnPeer.GetIsConnected() {
			internalConnPeer.Close()
		}
		Logger.log.Infof("SetPeerConn and Remove %s %s", peerIDStr, internalConnPeer.RemotePeer.RawAddress)
	}
	peerObj.PeerConns[peerIDStr] = peerConn
}

func (peerObj *Peer) removePeerConn(peerConn *PeerConn) error {
	peerObj.PeerConnsMtx.Lock()
	defer peerObj.PeerConnsMtx.Unlock()
	peerIDStr := peerConn.RemotePeer.PeerID.Pretty()
	internalConnPeer, ok := peerObj.PeerConns[peerIDStr]
	if ok {
		if internalConnPeer.GetIsConnected() {
			internalConnPeer.Close()
		}
		delete(peerObj.PeerConns, peerIDStr)
		Logger.log.Infof("RemovePeerConn %s %s", peerIDStr, peerConn.RemotePeer.RawAddress)
		return nil
	} else {
		return NewPeerError(UnexpectedError, errors.New(fmt.Sprintf("Can not find %+v", peerIDStr)), nil)
	}
}

// handleNewConnectionOut - main process when receiving a new peer connection,
// this mean we want to connect out to other peer
func (peerObj *Peer) handleNewConnectionOut(peer *Peer, cConn chan *PeerConn) (*PeerConn, error) {
	Logger.log.Infof("Opening stream to PEER Id - %s", peer.RawAddress)

	peerIDStr := peer.PeerID.Pretty()

	_, ok := peerObj.PeerConns[peerIDStr]
	if ok {
		Logger.log.Infof("Checked Existed PEER Id - %s", peer.RawAddress)

		if cConn != nil {
			cConn <- nil
		}
		return nil, nil
	}

	if peerIDStr == peerObj.PeerID.Pretty() {
		Logger.log.Infof("Checked MypeerObj PEER Id - %s", peer.RawAddress)
		//peerObj.newPeerConnectionMutex.Unlock()

		if cConn != nil {
			cConn <- nil
		}
		return nil, nil
	}

	if peerObj.countOfOutboundConn() >= peerObj.Config.MaxOutPeers && peerObj.Config.MaxOutPeers > 0 && !ok {
		Logger.log.Infof("Checked Max Outbound Connection PEER Id - %s", peer.RawAddress)

		//push to pending peers
		peerObj.connPending(peer)

		if cConn != nil {
			cConn <- nil
		}
		return nil, nil
	}

	stream, err := peerObj.Host.NewStream(context.Background(), peer.PeerID, protocolID)
	Logger.log.Info(peer, stream, err)
	if err != nil {
		if cConn != nil {
			cConn <- nil
		}
		return nil, NewPeerError(OpeningStreamP2PError, err, peerObj)
	}

	remotePeerID := stream.Conn().RemotePeer()

	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

	peerConn := PeerConn{
		isOutbound:         true, // we are connecting to remote peer -> this is an outbound peer
		RemotePeer:         peer,
		RemotePeerID:       remotePeerID,
		RemoteRawAddress:   peer.RawAddress,
		ListenerPeer:       peerObj,
		Config:             peerObj.Config,
		RWStream:           rw,
		cDisconnect:        make(chan struct{}),
		cClose:             make(chan struct{}),
		cRead:              make(chan struct{}),
		cWrite:             make(chan struct{}),
		cMsgHash:           make(map[string]chan bool),
		sendMessageQueue:   make(chan outMsg),
		HandleConnected:    peerObj.handleConnected,
		HandleDisconnected: peerObj.handleDisconnected,
		HandleFailed:       peerObj.handleFailed,
	}

	go peerConn.InMessageHandler(rw)
	go peerConn.OutMessageHandler(rw)

	peerObj.setPeerConn(&peerConn)
	defer func() {
		stream.Close()
		err := peerObj.removePeerConn(&peerConn)
		Logger.log.Error(err)
	}()

	peerConn.RetryCount = 0
	peerConn.SetConnState(connEstablished)

	go peerObj.handleConnected(&peerConn)

	if cConn != nil {
		cConn <- &peerConn
	}

	for {
		select {
		case <-peerConn.cDisconnect:
			Logger.log.Infof("NewPeerConnection Disconnected Stream PEER Id %s", peerConn.RemotePeerID.Pretty())
			return &peerConn, nil
		case <-peerConn.cClose:
			Logger.log.Infof("NewPeerConnection closed stream PEER Id %s", peerConn.RemotePeerID.Pretty())
			go func() {
				_, ok := <-peerConn.cDisconnect
				if !ok {
					Logger.log.Infof("NewPeerConnection disconnected after closed stream PEER Id %s", peerConn.RemotePeerID.Pretty())
					return
				}
			}()
			return &peerConn, nil
		}
	}
	return &peerConn, nil
}

// handleNewStreamIn - this mean we have other peer want to be connect to us(an inbound peer)
// we need to create data about this inbound peer and handle our inbound stream
func (peerObj *Peer) handleNewStreamIn(stream net.Stream, cDone chan *PeerConn) {
	// Remember to close the stream when we are done.
	defer stream.Close()

	if peerObj.countOfInboundConn() >= peerObj.Config.MaxInPeers && peerObj.Config.MaxInPeers > 0 {
		Logger.log.Infof("Max RemotePeer Inbound Connection")

		if cDone != nil {
			close(cDone)
		}
		return
	}

	remotePeerID := stream.Conn().RemotePeer()
	Logger.log.Infof("PEER %s Received a new stream from OTHER PEER with Id %s", peerObj.Host.ID().String(), remotePeerID.Pretty())
	_, ok := peerObj.PeerConns[remotePeerID.Pretty()]
	if ok {
		Logger.log.Infof("Received a new stream existed PEER Id - %s", remotePeerID.Pretty())

		if cDone != nil {
			close(cDone)
		}
		return
	}

	// Create a buffer stream for non blocking read and write.
	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

	peerConn := PeerConn{
		isOutbound:   false, // we are connected from remote peer -> this is an inbound peer
		ListenerPeer: peerObj,
		RemotePeer: &Peer{
			PeerID: remotePeerID,
		},
		Config:             peerObj.Config,
		RemotePeerID:       remotePeerID,
		RWStream:           rw,
		cDisconnect:        make(chan struct{}),
		cClose:             make(chan struct{}),
		cRead:              make(chan struct{}),
		cWrite:             make(chan struct{}),
		cMsgHash:           make(map[string]chan bool),
		sendMessageQueue:   make(chan outMsg),
		HandleConnected:    peerObj.handleConnected,
		HandleDisconnected: peerObj.handleDisconnected,
		HandleFailed:       peerObj.handleFailed,
	}

	peerObj.setPeerConn(&peerConn)

	go peerConn.InMessageHandler(rw)
	go peerConn.OutMessageHandler(rw)

	peerConn.RetryCount = 0
	peerConn.SetConnState(connEstablished)

	go peerObj.handleConnected(&peerConn)

	if cDone != nil {
		close(cDone)
	}

	defer func() {
		stream.Close()
		err := peerObj.removePeerConn(&peerConn)
		Logger.log.Error(err)
	}()

	for {
		select {
		case <-peerConn.cDisconnect:
			Logger.log.Infof("HandleStream disconnected stream PEER Id %s", peerConn.RemotePeerID.Pretty())
			return
		case <-peerConn.cClose:
			Logger.log.Infof("HandleStream closed stream PEER Id %s", peerConn.RemotePeerID.Pretty())
			go func() {
				_, ok := <-peerConn.cDisconnect
				if !ok {
					Logger.log.Infof("HandleStream disconnected after closed stream PEER Id %s", peerConn.RemotePeerID.Pretty())
					return
				}
			}()
			return
		}
	}
}

// QueueMessageWithEncoding adds the passed Incognito message to the peer send
// queue. This function is identical to QueueMessage, however it allows the
// caller to specify the wire encoding type that should be used when
// encoding/decoding blocks and transactions.
//
// This function is safe for concurrent access.
func (peerObj *Peer) QueueMessageWithEncoding(msg wire.Message, doneChan chan<- struct{}, msgType byte, msgShard *byte) {
	for _, peerConnection := range peerObj.PeerConns {
		go peerConnection.QueueMessageWithEncoding(msg, doneChan, msgType, msgShard)
	}
}

// Stop - stop all features of peer,
// not connect,
// not stream,
// not read and write message on stream
func (peerObj *Peer) Stop() {
	Logger.log.Warnf("Stopping PEER %s", peerObj.PeerID.Pretty())

	peerObj.Host.Close()
	peerObj.PeerConnsMtx.Lock()
	defer peerObj.PeerConnsMtx.Unlock()
	for _, peerConn := range peerObj.PeerConns {
		peerConn.SetConnState(connCanceled)
	}

	close(peerObj.cStop)
	Logger.log.Criticalf("PEER %s stopped", peerObj.PeerID.Pretty())
}

// handleConnected - set established flag to a peer when being connected
func (peerObj *Peer) handleConnected(peerConn *PeerConn) {
	Logger.log.Infof("handleConnected %s", peerConn.RemotePeerID.Pretty())
	peerConn.RetryCount = 0
	peerConn.SetConnState(connEstablished)

	peerObj.connEstablished(peerConn.RemotePeer)

	if peerObj.HandleConnected != nil {
		peerObj.HandleConnected(peerConn)
	}
}

// handleDisconnected - handle connected peer when it is disconnected, remove and retry connection
func (peerObj *Peer) handleDisconnected(peerConn *PeerConn) {
	Logger.log.Infof("handleDisconnected %s", peerConn.RemotePeerID.Pretty())
	peerConn.SetConnState(connCanceled)
	if peerConn.GetIsOutbound() && !peerConn.GetIsForceClose() {
		go peerObj.retryPeerConnection(peerConn)
	}

	if peerObj.HandleDisconnected != nil {
		peerObj.HandleDisconnected(peerConn)
	}
}

// handleFailed - handle when connecting peer failure
func (peerObj *Peer) handleFailed(peerConn *PeerConn) {
	Logger.log.Infof("handleFailed %s", peerConn.RemotePeerID.String())

	peerObj.connCanceled(peerConn.RemotePeer)

	if peerObj.HandleFailed != nil {
		peerObj.HandleFailed(peerConn)
	}
}

// retryPeerConnection - retry to connect to peer when being disconnected
func (peerObj *Peer) retryPeerConnection(peerConn *PeerConn) {
	time.AfterFunc(retryConnDuration, func() {
		Logger.log.Infof("Retry Zero RemotePeer Connection %s", peerConn.RemoteRawAddress)
		peerConn.RetryCount += 1

		if peerConn.RetryCount < maxRetryConn {
			peerConn.SetConnState(connPending)
			cConn := make(chan *PeerConn)
			peerConn.ListenerPeer.PushConn(peerConn.RemotePeer, cConn)
			p := <-cConn
			if p == nil {
				peerConn.RetryCount++
				go peerObj.retryPeerConnection(peerConn)
			}
		}
	})
}

// GetPeerConnOfAll - return all Peer connection to other peers
func (peerObj *Peer) GetPeerConnOfAll() []*PeerConn {
	peerObj.PeerConnsMtx.Lock()
	defer peerObj.PeerConnsMtx.Unlock()
	peerConns := make([]*PeerConn, 0)
	for _, peerConn := range peerObj.PeerConns {
		peerConns = append(peerConns, peerConn)
	}
	return peerConns
}

type PeerMessageInOut struct {
	PeerID  peer.ID
	Message wire.Message
	Time    int64
}

var inboundPeerMessage = map[string][]PeerMessageInOut{}
var outboundPeerMessage = map[string][]PeerMessageInOut{}
var inMutex = &sync.Mutex{}
var outMutex = &sync.Mutex{}

func StoreInboundPeerMessage(msg wire.Message, time int64, peerID peer.ID) {
	messageType := msg.MessageType()
	inMutex.Lock()
	defer inMutex.Unlock()
	existingMessages := inboundPeerMessage[messageType]
	if len(existingMessages) == 0 {
		inboundPeerMessage[messageType] = []PeerMessageInOut{
			{Message: msg, Time: time, PeerID: peerID},
		}
		return
	}

	messages := []PeerMessageInOut{
		{peerID, msg, time},
	}
	for _, message := range existingMessages {
		if message.Time < time-10 {
			continue
		}
		messages = append(messages, message)
	}
	inboundPeerMessage[messageType] = messages
}
func GetInboundPeerMessages() map[string][]PeerMessageInOut {
	return inboundPeerMessage
}
func GetInboundPeerMessagesByType(messageType string) []PeerMessageInOut {
	messages, ok := inboundPeerMessage[messageType]
	if !ok {
		return []PeerMessageInOut{}
	}
	return messages
}

func GetInboundMessagesByPeer() map[string]int {
	result := map[string]int{}
	for _, inboundMessages := range inboundPeerMessage {
		for _, message := range inboundMessages {
			result[message.PeerID.Pretty()]++
		}
	}
	return result
}

func StoreOutboundPeerMessage(msg wire.Message, time int64, peerID peer.ID) {
	messageType := msg.MessageType()
	outMutex.Lock()
	defer outMutex.Unlock()
	existingMessages := outboundPeerMessage[messageType]
	if len(existingMessages) == 0 {
		outboundPeerMessage[messageType] = []PeerMessageInOut{
			{Message: msg, Time: time, PeerID: peerID},
		}
		return
	}
	messages := []PeerMessageInOut{
		{peerID, msg, time},
	}
	for _, message := range existingMessages {
		if message.Time < time-10 {
			continue
		}
		messages = append(messages, message)
	}
	outboundPeerMessage[messageType] = messages
}
func GetOutboundPeerMessages() map[string][]PeerMessageInOut {
	return outboundPeerMessage
}
func GetOutboundPeerMessagesByType(messageType string) []PeerMessageInOut {
	messages, ok := outboundPeerMessage[messageType]
	if !ok {
		return []PeerMessageInOut{}
	}
	return messages
}

func GetOutboundMessagesByPeer() map[string]int {
	result := map[string]int{}
	for _, outboundMessages := range outboundPeerMessage {
		for _, message := range outboundMessages {
			result[message.PeerID.Pretty()]++
		}
	}
	return result
}
