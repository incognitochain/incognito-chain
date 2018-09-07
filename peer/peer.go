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

const (
	LOCAL_HOST = "127.0.0.1"
	// trickleTimeout is the duration of the ticker which trickles down the
	// inventory to a peer.
	trickleTimeout = 10 * time.Second

	maxRetryConn      = 25
	retryConnDuration = 10 * time.Second
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

type Peer struct {
	Host host.Host

	TargetAddress    ma.Multiaddr
	PeerId           peer.ID
	RawAddress       string
	ListeningAddress common.SimpleAddr

	Seed          int64
	outboundMutex sync.Mutex
	Config        Config
	MaxOutbound   int

	PeerConns    map[peer.ID]*PeerConn
	PendingPeers map[peer.ID]*Peer

	Peers map[string]string

	quit chan struct{}

	HandleConnected    func(peerConn *PeerConn)
	HandleDisconnected func(peerConn *PeerConn)
	HandleFailed       func(peerConn *PeerConn)
}

// Config is the struct to hold configuration options useful to Peer.
type Config struct {
	MessageListeners MessageListeners
	SealerPrvKey     string
}

type WrappedStream struct {
	Stream net.Stream
	Writer *bufio.Writer
	Reader *bufio.Reader
}

// MessageListeners defines callback function pointers to invoke with message
// listeners for a peer. Any listener which is not set to a concrete callback
// during peer initialization is ignored. Execution of multiple message
// listeners occurs serially, so one callback blocks the execution of the next.
//
// NOTE: Unless otherwise documented, these listeners must NOT directly call any
// blocking calls (such as WaitForShutdown) on the peer instance since the input
// handler goroutine blocks until the callback has completed.  Doing so will
// result in a deadlock.
type MessageListeners struct {
	OnTx        func(p *PeerConn, msg *wire.MessageTx)
	OnBlock     func(p *PeerConn, msg *wire.MessageBlock)
	OnGetBlocks func(p *PeerConn, msg *wire.MessageGetBlocks)
	OnVersion   func(p *PeerConn, msg *wire.MessageVersion)
	OnVerAck    func(p *PeerConn, msg *wire.MessageVerAck)
	OnGetAddr   func(p *PeerConn, msg *wire.MessageGetAddr)
	OnAddr      func(p *PeerConn, msg *wire.MessageAddr)
}

// outMsg is used to house a message to be sent along with a channel to signal
// when the message has been sent (or won't be sent due to things such as
// shutdown)
type outMsg struct {
	msg      wire.Message
	doneChan chan<- struct{}
	//encoding wire.MessageEncoding
}

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
	// to obtain a valid Host ID.
	priv, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r)
	if err != nil {
		return &self, err
	}

	ip := strings.Split(self.ListeningAddress.String(), ":")[0]
	if len(ip) == 0 {
		ip = LOCAL_HOST
	}
	port := strings.Split(self.ListeningAddress.String(), ":")[1]
	net := self.ListeningAddress.Network()
	listeningAddressString := fmt.Sprintf("/%s/%s/tcp/%s", net, ip, port)
	opts := []libp2p.Option{
		libp2p.ListenAddrStrings(listeningAddressString),
		libp2p.Identity(priv),
	}

	basicHost, err := libp2p.New(context.Background(), opts...)
	if err != nil {
		return &self, err
	}

	// Build Host multiaddress
	mulAddrStr := fmt.Sprintf("/ipfs/%s", basicHost.ID().Pretty())

	hostAddr, err := ma.NewMultiaddr(mulAddrStr)
	if err != nil {
		log.Print(err)
		return &self, err
	}

	// Now we can build a full multiaddress to reach this Host
	// by encapsulating both addresses:
	addr := basicHost.Addrs()[0]
	fullAddr := addr.Encapsulate(hostAddr)
	Logger.log.Infof("I am listening on %s with PEER ID - %s\n", fullAddr, basicHost.ID().String())
	pid, err := fullAddr.ValueForProtocol(ma.P_IPFS)
	if err != nil {
		log.Print(err)
		return &self, err
	}
	peerid, err := peer.IDB58Decode(pid)
	if err != nil {
		log.Print(err)
		return &self, err
	}

	self.RawAddress = fullAddr.String()
	self.Host = basicHost
	self.TargetAddress = fullAddr
	self.PeerId = peerid
	self.quit = make(chan struct{}, 1)
	self.Peers = make(map[string]string)

	self.outboundMutex = sync.Mutex{}

	return &self, nil
}

func (self *Peer) Start() error {
	Logger.log.Info("Peer start")
	// ping to bootnode for test env
	go self.HandleExchangePeers()
	Logger.log.Info("Set stream handler and wait for connection from other peer")
	self.Host.SetStreamHandler("/blockchain/1.0.0", self.HandleStream)
	select {
	case <-self.quit:
		Logger.log.Infof("PEER server shutdown complete %s", self.PeerId)
		break
	} // hang forever
	return nil
}

func (self *Peer) ConnPending(peer *Peer) {
	self.PendingPeers[peer.PeerId] = peer
}

func (self *Peer) ConnEstablished(peer *Peer) {
	_, ok := self.PendingPeers[peer.PeerId]
	if ok {
		delete(self.PendingPeers, peer.PeerId)
	}
}

func (self *Peer) ConnCanceled(peer *Peer) {
	_, ok := self.PeerConns[peer.PeerId]
	if ok {
		delete(self.PeerConns, peer.PeerId)
	}
	self.PendingPeers[peer.PeerId] = peer
}

func (self *Peer) NewPeerConnection(peer *Peer) (*PeerConn, error) {
	Logger.log.Infof("Opening stream to PEER ID - %s \n", peer.PeerId.String())

	self.outboundMutex.Lock()

	_peerConn, ok := self.PeerConns[peer.PeerId]
	if ok && _peerConn.State() == ConnEstablished {
		Logger.log.Infof("Existed PEER ID - %s", peer.PeerId.String())

		self.ConnPending(peer)

		self.outboundMutex.Unlock()
		return nil, nil
	}

	if peer.PeerId.Pretty() == self.PeerId.Pretty() {
		Logger.log.Infof("Myself PEER ID - %s", peer.PeerId.String())

		self.outboundMutex.Unlock()
		return nil, nil
	}

	if len(self.PeerConns) >= self.MaxOutbound && !ok {
		Logger.log.Infof("Max Peer Outbound Connection")

		self.ConnPending(peer)

		self.outboundMutex.Unlock()
		return nil, nil
	}

	stream, err := self.Host.NewStream(context.Background(), peer.PeerId, "/blockchain/1.0.0")
	if err != nil {
		Logger.log.Errorf("Fail in opening stream to PEER ID - %s with err: %s", self.PeerId.String(), err.Error())

		self.outboundMutex.Unlock()
		return nil, err
	}

	defer stream.Close()

	remotePeerId := stream.Conn().RemotePeer()

	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

	peerConn := PeerConn{
		IsOutbound:         true,
		Peer:               peer,
		ListenerPeer:       self,
		Config:             self.Config,
		PeerId:             remotePeerId,
		ReaderWriterStream: rw,
		quit:               make(chan struct{}),
		disconnect:         make(chan struct{}),
		sendMessageQueue:   make(chan outMsg, 1),
		HandleConnected:    self.handleConnected,
		HandleDisconnected: self.handleDisconnected,
		HandleFailed:       self.handleFailed,
	}

	self.PeerConns[peerConn.PeerId] = &peerConn

	go peerConn.InMessageHandler(rw)
	go peerConn.OutMessageHandler(rw)

	peerConn.RetryCount = 0
	peerConn.updateState(ConnEstablished)

	go self.handleConnected(&peerConn)

	self.outboundMutex.Unlock()

	for {
		select {
		case <-peerConn.disconnect:
			Logger.log.Infof("NewPeerConnection Close Stream PEER ID %s", peerConn.PeerId.String())
			break
		}
	}

	return &peerConn, nil
}

func (self *Peer) HandleStream(stream net.Stream) {
	// Remember to close the stream when we are done.
	defer stream.Close()

	remotePeerId := stream.Conn().RemotePeer()
	Logger.log.Infof("PEER %s Received a new stream from OTHER PEER with ID-%s", self.Host.ID().String(), remotePeerId.String())

	// TODO this code make EOF for libp2p
	//if !atomic.CompareAndSwapInt32(&self.connected, 0, 1) {
	//	return
	//}

	// Create a buffer stream for non blocking read and write.
	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

	peerConn := PeerConn{
		IsOutbound:         false,
		ListenerPeer:       self,
		Peer:               self,
		Config:             self.Config,
		PeerId:             remotePeerId,
		ReaderWriterStream: rw,
		quit:               make(chan struct{}),
		disconnect:         make(chan struct{}),
		sendMessageQueue:   make(chan outMsg, 1),
		HandleConnected:    self.handleConnected,
		HandleDisconnected: self.handleDisconnected,
		HandleFailed:       self.handleFailed,
	}

	self.PeerConns[peerConn.PeerId] = &peerConn

	go peerConn.InMessageHandler(rw)
	go peerConn.OutMessageHandler(rw)

	peerConn.RetryCount = 0
	peerConn.updateState(ConnEstablished)

	go self.handleConnected(&peerConn)

	for {
		select {
		case <-peerConn.disconnect:
			Logger.log.Infof("HandleStream Close Stream PEER ID %s", peerConn.PeerId.String())
			break
		}
	}
}

func (self Peer) HandleExchangePeers() {
	//	Logger.log.Infof("Start Exchange Peers")
	//	var client *rpc.Client
	//	var err error
	//
	//listen:
	//	for {
	//		Logger.log.Infof("Peers", self.Peers)
	//		if client == nil {
	//			client, err = rpc.Dial("tcp", "35.199.177.89:9339")
	//			if err != nil {
	//				Logger.log.Error("[Exchange Peers] re-connect:", err)
	//			}
	//		}
	//		if client != nil {
	//			Logger.log.Infof("[Exchange Peers] Ping")
	//			var response []server.RawPeer
	//			args := &server.PingArgs{self.RawAddress, self.Config.SealerPrvKey}
	//			Logger.log.Infof("[Exchange Peers] Ping", args)
	//			err := client.Call("Handler.Ping", args, &response)
	//			if err != nil {
	//				Logger.log.Error("[Exchange Peers] Ping:", err)
	//				client = nil
	//				time.Sleep(time.Second * 2)
	//
	//				goto listen
	//			}
	//			Logger.log.Infof("Ping response", response)
	//			for _, rawPeer := range response {
	//				if rawPeer.SealerPrvKey != "" && !strings.Contains(rawPeer.RawAddress, self.PeerId.String()) {
	//					keyPair := &cashec.KeyPair{}
	//					keyPair.Import([]byte(rawPeer.SealerPrvKey))
	//
	//					_, exist := self.Peers[string(keyPair.PublicKey)]
	//
	//					if !exist {
	//						self.Peers[string(keyPair.PublicKey)] = rawPeer.RawAddress
	//					}
	//				}
	//			}
	//			Logger.log.Infof("Ping response Peers", self.Peers)
	//		}
	//		time.Sleep(time.Second * 2)
	//	}
}

// QueueMessageWithEncoding adds the passed bitcoin message to the peer send
// queue. This function is identical to QueueMessage, however it allows the
// caller to specify the wire encoding type that should be used when
// encoding/decoding blocks and transactions.
//
// This function is safe for concurrent access.
func (self Peer) QueueMessageWithEncoding(msg wire.Message, doneChan chan<- struct{}) {
	for _, peerConnection := range self.PeerConns {
		peerConnection.QueueMessageWithEncoding(msg, doneChan)
	}
}

func (self *Peer) Stop() {
	Logger.log.Infof("PEER %s Stop", self.PeerId.String())

	self.Host.Close()
	for _, peerConn := range self.PeerConns {
		peerConn.updateState(ConnCanceled)
	}
	self.quit <- struct{}{}
}

func (self *Peer) handleConnected(peerConn *PeerConn) {
	Logger.log.Infof("handleConnected %s", peerConn.PeerId.String())
	peerConn.RetryCount = 0
	peerConn.updateState(ConnEstablished)

	self.ConnEstablished(peerConn.Peer)

	if self.HandleConnected != nil {
		self.HandleConnected(peerConn)
	}
}

func (self *Peer) handleDisconnected(peerConn *PeerConn) {
	Logger.log.Infof("handleDisconnected %s", peerConn.PeerId.String())

	if peerConn.IsOutbound {
		if peerConn.State() != ConnCanceled {
			go self.retryPeerConnection(peerConn)
		}
	} else {
		self.ConnCanceled(peerConn.Peer)
	}

	if self.HandleDisconnected != nil {
		self.HandleDisconnected(peerConn)
	}
}

func (self *Peer) handleFailed(peerConn *PeerConn) {
	Logger.log.Infof("handleFailed %s", peerConn.PeerId.String())

	self.ConnCanceled(peerConn.Peer)

	if self.HandleFailed != nil {
		self.HandleFailed(peerConn)
	}
}

func (self *Peer) retryPeerConnection(peerConn *PeerConn) {
	time.AfterFunc(retryConnDuration, func() {
		Logger.log.Infof("Retry New Peer Connection %s", peerConn.PeerId.String())
		peerConn.RetryCount += 1

		if peerConn.RetryCount < maxRetryConn {
			peerConn.updateState(ConnPending)
			_, err := peerConn.ListenerPeer.NewPeerConnection(peerConn.Peer)
			if err != nil {
				go self.retryPeerConnection(peerConn)
			}
		} else {
			peerConn.updateState(ConnCanceled)

			self.ConnCanceled(peerConn.Peer)
			self.newPeerConnection()
			self.ConnPending(peerConn.Peer)
		}
	})
}

func (self *Peer) newPeerConnection() {
	for _, peer := range self.PendingPeers {
		go self.NewPeerConnection(peer)
	}
}
