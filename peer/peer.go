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

var MAX_RETRIES_CHECK_HASH_MESSAGE = 5
var HEAVY_MESSAGE_SIZE = 512 * 1024
var SPAM_MESSAGE_SIZE = 50 * 1024 * 1024
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

	GetShardByPbk func(pbk string) *byte
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

func (peerObj *Peer) ReceivedHashMessage(hash string) {
	if peerObj.messagePool == nil {
		peerObj.messagePool = make(map[string]bool)
	}
	peerObj.messagePool[hash] = true
	if len(peerObj.messagePool) > MESSAGE_HASH_POOL_SIZE {
		for k, _ := range peerObj.messagePool {
			delete(peerObj.messagePool, k)
			break
		}
	}
}

func (peerObj *Peer) CheckHashMessage(hash string) (bool) {
	if peerObj.messagePool == nil {
		peerObj.messagePool = make(map[string]bool)
	}
	ok, _ := peerObj.messagePool[hash]
	return ok
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
		return &peerObj, NewPeerError(PeerGenerateKeyPairErr, err, &peerObj)
	}

	ip := strings.Split(peerObj.ListeningAddress.String(), ":")[0]
	if len(ip) == 0 {
		ip = LocalHost
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
		return &peerObj, NewPeerError(CreateP2PNodeErr, err, &peerObj)
	}

	// Build Host multiaddress
	mulAddrStr := fmt.Sprintf("/ipfs/%s", basicHost.ID().Pretty())

	hostAddr, err := ma.NewMultiaddr(mulAddrStr)
	if err != nil {
		return &peerObj, NewPeerError(CreateP2PAddressErr, err, &peerObj)
	}

	// Now we can build a full multiaddress to reach this Host
	// by encapsulating both addresses:
	addr := basicHost.Addrs()[0]
	fullAddr := addr.Encapsulate(hostAddr)
	Logger.log.Infof("I am listening on %s with PEER Id - %s", fullAddr, basicHost.ID().String())
	pid, err := fullAddr.ValueForProtocol(ma.P_IPFS)
	if err != nil {
		return &peerObj, NewPeerError(GetPeerIdFromProtocolErr, err, &peerObj)
	}
	peerID, err := peer.IDB58Decode(pid)
	if err != nil {
		log.Print(err)
		return &peerObj, NewPeerError(GetPeerIdFromProtocolErr, err, &peerObj)
	}

	peerObj.RawAddress = fmt.Sprintf("%s%s", listeningAddressString, mulAddrStr)
	peerObj.Host = basicHost
	peerObj.TargetAddress = fullAddr
	peerObj.PeerID = peerID
	peerObj.cStop = make(chan struct{}, 1)
	peerObj.cDisconnectPeer = make(chan *PeerConn)
	peerObj.cNewConn = make(chan *NewPeerMsg)
	peerObj.cNewStream = make(chan *NewStreamMsg)
	peerObj.cStopConn = make(chan struct{})

	peerObj.peerConnMutex = sync.Mutex{}
	return &peerObj, nil
}

/*
Start - start peer to begin waiting for connections from other peers
*/
func (peerObj *Peer) Start() {
	Logger.log.Info("RemotePeer start")
	// ping to bootnode for test env
	Logger.log.Info("Set stream handler and wait for connection from other peer")
	peerObj.Host.SetStreamHandler(ProtocolId, peerObj.PushStream)

	go peerObj.processConn()

	select {
	case <-peerObj.cStop:
		close(peerObj.cStopConn)
		Logger.log.Warnf("PEER server shutdown complete %s", peerObj.PeerID)
		break
	}
	return
}

func (peerObj *Peer) PushStream(stream net.Stream) {
	newStreamMsg := NewStreamMsg{
		Stream: stream,
		CConn:  nil,
	}
	peerObj.cNewStream <- &newStreamMsg
}

func (peerObj *Peer) PushConn(peer *Peer, cConn chan *PeerConn) {
	newPeerMsg := NewPeerMsg{
		Peer:  peer,
		CConn: cConn,
	}
	peerObj.cNewConn <- &newPeerMsg
}

func (peerObj *Peer) processConn() {
	for {
		select {
		case <-peerObj.cStopConn:
			Logger.log.Info("ProcessConn QUIT")
			return
		case newPeerMsg := <-peerObj.cNewConn:
			Logger.log.Infof("ProcessConn START CONN %s %s", newPeerMsg.Peer.PeerID, newPeerMsg.Peer.RawAddress)
			cDone := make(chan *PeerConn)
			go func(peerObj *Peer) {
				peerConn, err := peerObj.handleConn(newPeerMsg.Peer, cDone)
				if err != nil && peerConn == nil {
					Logger.log.Errorf("Fail in opening stream from PEER Id - %s with err: %s", peerObj.PeerID.Pretty(), err.Error())
				}
			}(peerObj)
			p := <-cDone
			if newPeerMsg.CConn != nil {
				newPeerMsg.CConn <- p
			}
			Logger.log.Infof("ProcessConn END CONN %s %s", newPeerMsg.Peer.PeerID, newPeerMsg.Peer.RawAddress)
			continue
		case newStreamMsg := <-peerObj.cNewStream:
			remotePeerID := newStreamMsg.Stream.Conn().RemotePeer()
			Logger.log.Infof("ProcessConn START STREAM %s", remotePeerID)
			cConn := make(chan *PeerConn)
			go peerObj.handleStream(newStreamMsg.Stream, cConn)
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

func (peerObj *Peer) ConnPending(peer *Peer) {
	peerObj.pendingPeersMutex.Lock()
	peerObj.PendingPeers[peer.PeerID.Pretty()] = peer
	peerObj.pendingPeersMutex.Unlock()
}

func (peerObj *Peer) ConnEstablished(peer *Peer) {
	peerObj.pendingPeersMutex.Lock()
	_, ok := peerObj.PendingPeers[peer.PeerID.Pretty()]
	if ok {
		delete(peerObj.PendingPeers, peer.PeerID.Pretty())
	}
	peerObj.pendingPeersMutex.Unlock()
}

func (peerObj *Peer) ConnCanceled(peer *Peer) {
	_, ok := peerObj.PeerConns[peer.PeerID.Pretty()]
	if ok {
		delete(peerObj.PeerConns, peer.PeerID.Pretty())
	}
	peerObj.pendingPeersMutex.Lock()
	peerObj.PendingPeers[peer.PeerID.Pretty()] = peer
	peerObj.pendingPeersMutex.Unlock()
}

func (peerObj *Peer) NumInbound() int {
	ret := int(0)
	peerObj.peerConnMutex.Lock()
	for _, peerConn := range peerObj.PeerConns {
		if !peerConn.IsOutbound {
			ret++
		}
	}
	peerObj.peerConnMutex.Unlock()
	return ret
}

func (peerObj *Peer) NumOutbound() int {
	ret := int(0)
	peerObj.peerConnMutex.Lock()
	for _, peerConn := range peerObj.PeerConns {
		if peerConn.IsOutbound {
			ret++
		}
	}
	peerObj.peerConnMutex.Unlock()
	return ret
}

func (peerObj *Peer) GetPeerConnByPeerID(peerID string) (*PeerConn) {
	peerConn, ok := peerObj.PeerConns[peerID]
	if ok {
		return peerConn
	}
	return nil
}

func (peerObj *Peer) GetPeerConnByPbk(pbk string) (*PeerConn) {
	for _, peerConn := range peerObj.PeerConns {
		if peerConn.RemotePeer.PublicKey == pbk {
			return peerConn
		}
	}
	return nil
}

func (peerObj *Peer) GetListPeerConnByShard(shard byte) ([]*PeerConn) {
	peerConns := make([]*PeerConn, 0)
	for _, peerConn := range peerObj.PeerConns {
		shardT := peerObj.Config.GetShardByPbk(peerConn.RemotePeer.PublicKey)
		if shardT != nil && *shardT == shard {
			peerConns = append(peerConns, peerConn)
		}
	}
	return peerConns
}

func (peerObj *Peer) UpdateShardForPeerConn() {
	for _, peerConn := range peerObj.PeerConns {
		_ = peerConn
	}
}

func (peerObj *Peer) SetPeerConn(peerConn *PeerConn) {
	internalConnPeer, ok := peerObj.PeerConns[peerConn.RemotePeer.PeerID.Pretty()]
	if ok && internalConnPeer != peerConn {
		if internalConnPeer.IsConnected {
			internalConnPeer.Close()
		}
		Logger.log.Infof("SetPeerConn and Remove %s %s", internalConnPeer.RemotePeer.PeerID, internalConnPeer.RemotePeer.RawAddress)
	}
	peerObj.PeerConns[peerConn.RemotePeer.PeerID.Pretty()] = peerConn
}

func (peerObj *Peer) RemovePeerConn(peerConn *PeerConn) {
	internalConnPeer, ok := peerObj.PeerConns[peerConn.RemotePeer.PeerID.Pretty()]
	if ok {
		if internalConnPeer.IsConnected {
			internalConnPeer.Close()
		}
		delete(peerObj.PeerConns, peerConn.RemotePeer.PeerID.Pretty())
		Logger.log.Infof("RemovePeerConn %s %s", peerConn.RemotePeer.PeerID, peerConn.RemotePeer.RawAddress)
	}
}

func (peerObj *Peer) handleConn(peer *Peer, cConn chan *PeerConn) (*PeerConn, error) {
	Logger.log.Infof("Opening stream to PEER Id - %s", peer.RawAddress)

	_, ok := peerObj.PeerConns[peer.PeerID.Pretty()]
	if ok {
		Logger.log.Infof("Checked Existed PEER Id - %s", peer.RawAddress)

		if cConn != nil {
			cConn <- nil
		}
		return nil, nil
	}

	if peer.PeerID.Pretty() == peerObj.PeerID.Pretty() {
		Logger.log.Infof("Checked MypeerObj PEER Id - %s", peer.RawAddress)
		//peerObj.newPeerConnectionMutex.Unlock()

		if cConn != nil {
			cConn <- nil
		}
		return nil, nil
	}

	if peerObj.NumOutbound() >= peerObj.Config.MaxOutbound && peerObj.Config.MaxOutbound > 0 && !ok {
		Logger.log.Infof("Checked Max Outbound Connection PEER Id - %s", peer.RawAddress)

		//push to pending peers
		peerObj.ConnPending(peer)

		if cConn != nil {
			cConn <- nil
		}
		return nil, nil
	}

	stream, err := peerObj.Host.NewStream(context.Background(), peer.PeerID, ProtocolId)
	Logger.log.Info(peer, stream, err)
	if err != nil {
		if cConn != nil {
			cConn <- nil
		}
		return nil, NewPeerError(OpeningStreamP2PErr, err, peerObj)
	}

	defer stream.Close()

	remotePeerID := stream.Conn().RemotePeer()

	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

	peerConn := PeerConn{
		IsOutbound:         true,
		RemotePeer:         peer,
		RemotePeerID:       remotePeerID,
		RemoteRawAddress:   peer.RawAddress,
		ListenerPeer:       peerObj,
		Config:             peerObj.Config,
		ReaderWriterStream: rw,
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

	peerObj.SetPeerConn(&peerConn)

	go peerConn.InMessageHandler(rw)
	go peerConn.OutMessageHandler(rw)

	peerConn.RetryCount = 0
	peerConn.updateConnState(ConnEstablished)

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
				select {
				case <-peerConn.cDisconnect:
					Logger.log.Infof("NewPeerConnection disconnected after closed stream PEER Id %s", peerConn.RemotePeerID.Pretty())
					return
				}
			}()
			return &peerConn, nil
		}
	}

	return &peerConn, nil
}

func (peerObj *Peer) handleStream(stream net.Stream, cDone chan *PeerConn) {
	// Remember to close the stream when we are done.
	defer stream.Close()

	if peerObj.NumInbound() >= peerObj.Config.MaxInbound && peerObj.Config.MaxInbound > 0 {
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
		IsOutbound:   false,
		ListenerPeer: peerObj,
		RemotePeer: &Peer{
			PeerID: remotePeerID,
		},
		Config:             peerObj.Config,
		RemotePeerID:       remotePeerID,
		ReaderWriterStream: rw,
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

	peerObj.SetPeerConn(&peerConn)

	go peerConn.InMessageHandler(rw)
	go peerConn.OutMessageHandler(rw)

	peerConn.RetryCount = 0
	peerConn.updateConnState(ConnEstablished)

	go peerObj.handleConnected(&peerConn)

	if cDone != nil {
		close(cDone)
	}

	for {
		select {
		case <-peerConn.cDisconnect:
			Logger.log.Infof("HandleStream disconnected stream PEER Id %s", peerConn.RemotePeerID.Pretty())
			return
		case <-peerConn.cClose:
			Logger.log.Infof("HandleStream closed stream PEER Id %s", peerConn.RemotePeerID.Pretty())
			go func() {
				select {
				case <-peerConn.cDisconnect:
					Logger.log.Infof("HandleStream disconnected after closed stream PEER Id %s", peerConn.RemotePeerID.Pretty())
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
func (peerObj *Peer) QueueMessageWithEncoding(msg wire.Message, doneChan chan<- struct{}) {
	for _, peerConnection := range peerObj.PeerConns {
		go peerConnection.QueueMessageWithEncoding(msg, doneChan)
	}
}

func (peerObj *Peer) Stop() {
	Logger.log.Infof("Stopping PEER %s", peerObj.PeerID.Pretty())

	peerObj.Host.Close()
	peerObj.peerConnMutex.Lock()
	for _, peerConn := range peerObj.PeerConns {
		peerConn.updateConnState(ConnCanceled)
	}
	peerObj.peerConnMutex.Unlock()

	close(peerObj.cStop)
	Logger.log.Infof("PEER %s stopped", peerObj.PeerID.Pretty())
}

/*
handleConnected - set established flag to a peer when being connected
*/
func (peerObj *Peer) handleConnected(peerConn *PeerConn) {
	Logger.log.Infof("handleConnected %s", peerConn.RemotePeerID.Pretty())
	peerConn.RetryCount = 0
	peerConn.updateConnState(ConnEstablished)

	peerObj.ConnEstablished(peerConn.RemotePeer)

	if peerObj.HandleConnected != nil {
		peerObj.HandleConnected(peerConn)
	}
}

/*
handleDisconnected - handle connected peer when it is disconnected, remove and retry connection
*/
func (peerObj *Peer) handleDisconnected(peerConn *PeerConn) {
	Logger.log.Infof("handleDisconnected %s", peerConn.RemotePeerID.Pretty())
	peerConn.updateConnState(ConnCanceled)
	peerObj.RemovePeerConn(peerConn)
	if peerConn.IsOutbound && !peerConn.isForceClose {
		go peerObj.retryPeerConnection(peerConn)
	}

	if peerObj.HandleDisconnected != nil {
		peerObj.HandleDisconnected(peerConn)
	}
}

/*
handleFailed - handle when connecting peer failure
*/
func (peerObj *Peer) handleFailed(peerConn *PeerConn) {
	Logger.log.Infof("handleFailed %s", peerConn.RemotePeerID.String())

	peerObj.ConnCanceled(peerConn.RemotePeer)

	if peerObj.HandleFailed != nil {
		peerObj.HandleFailed(peerConn)
	}
}

/*
retryPeerConnection - retry to connect to peer when being disconnected
*/
func (peerObj *Peer) retryPeerConnection(peerConn *PeerConn) {
	time.AfterFunc(RetryConnDuration, func() {
		Logger.log.Infof("Retry Zero RemotePeer Connection %s", peerConn.RemoteRawAddress)
		peerConn.RetryCount += 1

		if peerConn.RetryCount < MaxRetryConn {
			peerConn.updateConnState(ConnPending)
			cConn := make(chan *PeerConn)
			peerConn.ListenerPeer.PushConn(peerConn.RemotePeer, cConn)
			p := <-cConn
			if p == nil {
				peerConn.RetryCount++
				go peerObj.retryPeerConnection(peerConn)
			}
		} else {
			peerConn.updateConnState(ConnCanceled)
			peerObj.ConnCanceled(peerConn.RemotePeer)
			peerObj.renewPeerConnection()
			peerObj.ConnPending(peerConn.RemotePeer)
		}
	})
}

/*
renewPeerConnection - create peer conn by goroutines for pending peers(reconnect)
*/
func (peerObj *Peer) renewPeerConnection() {
	if len(peerObj.PendingPeers) > 0 {
		peerObj.pendingPeersMutex.Lock()
		Logger.log.Infof("*start - Creating peer conn to %d pending peers", len(peerObj.PendingPeers))
		for _, peer := range peerObj.PendingPeers {
			Logger.log.Infof("---> RemotePeer: ", peer.RawAddress)
			go peerObj.PushConn(peer, nil)
		}
		Logger.log.Infof("*end - Creating peer conn to %d pending peers", len(peerObj.PendingPeers))
		peerObj.pendingPeersMutex.Unlock()
	}
}

func (peerObj *Peer) ClosePeerConnsOfShard(shard byte) {
	for _, peerConn := range peerObj.PeerConns {
		sh := peerObj.Config.GetShardByPbk(peerConn.RemotePeer.PublicKey)
		if sh != nil && *sh == shard {
			peerConn.ForceClose()
		}
	}
}

func (peerObj *Peer) CountPeerConnOfShard(shard *byte) int {
	c := 0
	for _, peerConn := range peerObj.PeerConns {
		sh := peerObj.Config.GetShardByPbk(peerConn.RemotePeer.PublicKey)
		if (shard == nil && sh == nil) || (sh != nil && shard != nil && *sh == *shard) {
			c++
		}
	}
	return c
}

func (peerObj *Peer) GetPeerConnOfShard(shard *byte) []*PeerConn {
	c := make([]*PeerConn, 0)
	for _, peerConn := range peerObj.PeerConns {
		sh := peerObj.Config.GetShardByPbk(peerConn.RemotePeer.PublicKey)
		if (shard == nil && sh == nil) || (sh != nil && shard != nil && *sh == *shard) {
			c = append(c, peerConn)
		}
	}
	return c
}
