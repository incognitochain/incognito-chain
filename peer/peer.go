package peer

import (
	"github.com/libp2p/go-libp2p-host"
	"github.com/libp2p/go-libp2p"
	"log"
	"io"
	"crypto/rand"
	mrand "math/rand"
	"github.com/libp2p/go-libp2p-crypto"
	"fmt"
	"context"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/libp2p/go-libp2p-net"
	"bufio"
	"sync"
	"github.com/internet-cash/prototype/wire"
	"strings"
	n "net"
	"github.com/libp2p/go-libp2p-peer"
	"sync/atomic"
	"time"
	"encoding/hex"
	"encoding/json"
	"reflect"
	"bytes"
)

const (
	LOCAL_HOST = "127.0.0.1"
	// trickleTimeout is the duration of the ticker which trickles down the
	// inventory to a peer.
	trickleTimeout = 10 * time.Second
)

type Peer struct {
	connected  int32
	disconnect int32

	Host             host.Host
	TargetAddress    ma.Multiaddr
	PeerId           peer.ID
	ListeningAddress n.Addr
	Seed             int64
	FlagMutex        sync.Mutex

	Config Config

	sendMessageQueue chan outMsg
	quit             chan struct{}
}

// Config is the struct to hold configuration options useful to Peer.
type Config struct {
	MessageListeners MessageListeners
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
	OnTx    func(p *Peer, msg *wire.MessageTx)
	OnBlock func(p *Peer, msg *wire.MessageBlock)
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
	log.Printf("I am listening on %s\n", fullAddr)
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

	self.Host = basicHost
	self.TargetAddress = fullAddr
	self.PeerId = peerid
	self.quit = make(chan struct{})
	self.sendMessageQueue = make(chan outMsg, 1)
	return &self, nil
}

func (self Peer) Start() (error) {
	log.Println("Peer start")
	log.Println("Set stream handler and wait for connection from other peer")
	self.Host.SetStreamHandler("/blockchain/1.0.0", self.HandleStream)
	select {} // hang forever
	return nil
}

// WaitForDisconnect waits until the peer has completely disconnected and all
// resources are cleaned up.  This will happen if either the local or remote
// side has been disconnected or the peer is forcibly disconnected via
// Disconnect.
func (p Peer) WaitForDisconnect() {
	<-p.quit
}

func (self Peer) HandleStream(stream net.Stream) {
	// Remember to close the stream when we are done.
	//defer stream.Close()

	log.Printf("%s Received a new stream!", self.Host.ID().String())

	// TODO this code make EOF for libp2p
	//if !atomic.CompareAndSwapInt32(&self.connected, 0, 1) {
	//	return
	//}

	// Create a buffer stream for non blocking read and write.
	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

	go self.InMessageHandler(rw)
	go self.OutMessageHandler(rw)
}

/**
Handle all in message
 */
func (self Peer) InMessageHandler(rw *bufio.ReadWriter) {
	for {
		log.Printf("read stream")
		str, err := rw.ReadString('\n')
		if err != nil {
			log.Println(err)
			return
		}

		//if str == "" {
		//	return
		//}
		log.Printf("Message: %s \n", str)
		if str != "\n" {

			// Parse Message header
			jsonDecodeString, _ := hex.DecodeString(str)
			messageHeader := jsonDecodeString[len(jsonDecodeString)-wire.MessageHeaderSize:]
			messageHeader = bytes.Trim(messageHeader, "\x00")
			log.Println(string(messageHeader))

			messageType := string(messageHeader[:len(messageHeader)])
			var message, err = wire.MakeEmptyMessage(string(messageType))

			// Parse Message body
			messageBody := jsonDecodeString[:len(jsonDecodeString)-wire.MessageHeaderSize]
			log.Println(string(messageBody))
			if err != nil {
				log.Println(err)
				continue
			}
			err = json.Unmarshal(messageBody, &message)
			//temp := message.(map[string]interface{})
			if err != nil {
				log.Println(err)
				continue
			}
			realType := reflect.TypeOf(message)
			log.Print(realType)
			// check type of Message
			switch realType {
			case reflect.TypeOf(&wire.MessageTx{}):
				if self.Config.MessageListeners.OnTx != nil {
					self.FlagMutex.Lock()
					self.Config.MessageListeners.OnTx(&self, message.(*wire.MessageTx))
					self.FlagMutex.Unlock()
				}
			case reflect.TypeOf(&wire.MessageBlock{}):
				if self.Config.MessageListeners.OnBlock != nil {
					self.FlagMutex.Lock()
					self.Config.MessageListeners.OnBlock(&self, message.(*wire.MessageBlock))
					self.FlagMutex.Unlock()
				}
			default:
				log.Printf("Received unhandled message of type %v "+
					"from %v", realType, self)
			}
		}
	}
}

/**
// OutMessageHandler handles the queuing of outgoing data for the peer. This runs as
// a muxer for various sources of input so we can ensure that server and peer
// handlers will not block on us sending a message.  That data is then passed on
// to outHandler to be actually written.
*/
func (self Peer) OutMessageHandler(rw *bufio.ReadWriter) {
	/* for test message
	go func() {
		for {
			time.Sleep(1 * time.Second)
			a := fmt.Sprintf("%s hello \n", self.Host.ID().String())
			log.Printf("%s Write string %s", self.Host.ID().String(), a)
			self.FlagMutex.Lock()
			rw.Writer.WriteString(a)
			rw.Writer.Flush()
			self.FlagMutex.Unlock()
		}
	}()*/
	for {
		select {
		case outMsg := <-self.sendMessageQueue:
			{
				self.FlagMutex.Lock()
				// TODO
				// send message
				message, err := outMsg.msg.JsonSerialize()
				if err != nil {
					continue
				}
				message += "\n"
				log.Printf("Send a message: %s", message)
				rw.Writer.WriteString(message)
				rw.Writer.Flush()
				self.FlagMutex.Unlock()
			}
		case <-self.quit:
			break
			//default:
			//	log.Println("Wait for sending message")
		}
	}
}

// Connected returns whether or not the peer is currently connected.
//
// This function is safe for concurrent access.
func (self Peer) Connected() bool {
	return atomic.LoadInt32(&self.connected) != 0 &&
		atomic.LoadInt32(&self.disconnect) == 0
}

// Disconnect disconnects the peer by closing the connection.  Calling this
// function when the peer is already disconnected or in the process of
// disconnecting will have no effect.
func (self Peer) Disconnect() {
	if atomic.AddInt32(&self.disconnect, 1) != 1 {
		return
	}

	log.Printf("Disconnecting %s", self)
	if atomic.LoadInt32(&self.connected) != 0 {
		self.Host.Close()
	}
	if self.quit != nil {
		close(self.quit)
	}
	self.disconnect = 1
}

// QueueMessageWithEncoding adds the passed bitcoin message to the peer send
// queue. This function is identical to QueueMessage, however it allows the
// caller to specify the wire encoding type that should be used when
// encoding/decoding blocks and transactions.
//
// This function is safe for concurrent access.
func (self Peer) QueueMessageWithEncoding(msg wire.Message, doneChan chan<- struct{}) {
	//if !self.Connected() {
	//	if doneChan != nil {
	//		go func() {
	//			doneChan <- struct{}{}
	//		}()
	//	}
	//	return
	//}
	self.sendMessageQueue <- outMsg{msg: msg, doneChan: doneChan}
	//self.FlagMutex.Lock()
	//self.ReadWrite.WriteString("test test test")
	//self.ReadWrite.Flush()
	//self.FlagMutex.Unlock()
}
