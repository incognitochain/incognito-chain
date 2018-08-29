package peer

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"sync"
	"sync/atomic"
	"github.com/libp2p/go-libp2p-host"
	"github.com/libp2p/go-libp2p-peer"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/ninjadotorg/cash-prototype/common"
	"github.com/ninjadotorg/cash-prototype/wire"
)

type PeerConn struct {
	connected  int32
	disconnect int32

	RetryCount int32
	IsOutbound bool
	state      ConnState
	stateMtx   sync.RWMutex

	Host               host.Host
	ReaderWriterStream *bufio.ReadWriter
	verAckReceived     bool

	TargetAddress    ma.Multiaddr
	PeerId           peer.ID
	RawAddress       string
	ListeningAddress common.SimpleAddr

	Seed      int64
	FlagMutex sync.Mutex
	Config    Config

	sendMessageQueue chan outMsg
	quit             chan struct{}

	Peer *Peer

	HandleConnected    func(peerConn *PeerConn)
	HandleDisconnected func(peerConn *PeerConn)
	HandleFailed       func(peerConn *PeerConn)
}

/**
Handle all in message
*/
func (self *PeerConn) InMessageHandler(rw *bufio.ReadWriter) {
	for {
		Logger.log.Infof("Reading stream")
		str, err := rw.ReadString('\n')
		if err != nil {
			Logger.log.Error(err)

			Logger.log.Infof("PEER %s quit IN message handler", self.PeerId)
			self.quit <- struct{}{}

			if self.HandleDisconnected != nil {
				self.HandleDisconnected(self)
			}

			return
		}

		//if str == "" {
		//	return
		//}
		Logger.log.Infof("Received message: %s \n", str)
		if str != "\n" {

			// Parse Message header
			jsonDecodeString, _ := hex.DecodeString(str)
			messageHeader := jsonDecodeString[len(jsonDecodeString)-wire.MessageHeaderSize:]

			commandInHeader := messageHeader[:12]
			commandInHeader = bytes.Trim(messageHeader, "\x00")
			Logger.log.Info("Message Type - " + string(commandInHeader))
			commandType := string(messageHeader[:len(commandInHeader)])
			var message, err = wire.MakeEmptyMessage(string(commandType))

			// Parse Message body
			messageBody := jsonDecodeString[:len(jsonDecodeString)-wire.MessageHeaderSize]
			Logger.log.Info("Message Body - " + string(messageBody))
			if err != nil {
				Logger.log.Error(err)
				continue
			}
			if commandType != wire.CmdBlock {
				err = json.Unmarshal(messageBody, &message)
			} else {
				err = json.Unmarshal(messageBody, &message)
			}
			//temp := message.(map[string]interface{})
			if err != nil {
				Logger.log.Error(err)
				continue
			}
			realType := reflect.TypeOf(message)
			log.Print(realType)
			// check type of Message
			switch realType {
			case reflect.TypeOf(&wire.MessageTx{}):
				if self.Config.MessageListeners.OnTx != nil {
					self.FlagMutex.Lock()
					self.Config.MessageListeners.OnTx(self, message.(*wire.MessageTx))
					self.FlagMutex.Unlock()
				}
			case reflect.TypeOf(&wire.MessageBlock{}):
				if self.Config.MessageListeners.OnBlock != nil {
					self.FlagMutex.Lock()
					self.Config.MessageListeners.OnBlock(self, message.(*wire.MessageBlock))
					self.FlagMutex.Unlock()
				}
			case reflect.TypeOf(&wire.MessageGetBlocks{}):
				if self.Config.MessageListeners.OnGetBlocks != nil {
					self.FlagMutex.Lock()
					self.Config.MessageListeners.OnGetBlocks(self, message.(*wire.MessageGetBlocks))
					self.FlagMutex.Unlock()
				}
			case reflect.TypeOf(&wire.MessageVersion{}):
				if self.Config.MessageListeners.OnVersion != nil {
					self.FlagMutex.Lock()
					versionMessage := message.(*wire.MessageVersion)
					self.Config.MessageListeners.OnVersion(self, versionMessage)
					self.FlagMutex.Unlock()
				}
			case reflect.TypeOf(&wire.MessageVerAck{}):
				self.FlagMutex.Lock()
				self.verAckReceived = true
				self.FlagMutex.Unlock()
				if self.Config.MessageListeners.OnVerAck != nil {
					self.FlagMutex.Lock()
					self.Config.MessageListeners.OnVerAck(self, message.(*wire.MessageVerAck))
					self.FlagMutex.Unlock()
				}
			default:
				Logger.log.Warnf("Received unhandled message of type %v "+
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
func (self PeerConn) OutMessageHandler(rw *bufio.ReadWriter) {
	for {
		select {
		case outMsg := <-self.sendMessageQueue:
			{
				self.FlagMutex.Lock()
				// TODO
				// send message
				messageByte, err := outMsg.msg.JsonSerialize()
				if err != nil {
					fmt.Println(err)
					continue
				}
				header := make([]byte, wire.MessageHeaderSize)
				CmdType, _ := wire.GetCmdType(reflect.TypeOf(outMsg.msg))
				copy(header[:], []byte(CmdType))
				messageByte = append(messageByte, header...)
				message := hex.EncodeToString(messageByte)
				message += "\n"
				Logger.log.Infof("Send a message %s %s: %s", self.PeerId.String(), outMsg.msg.MessageType(), message)
				rw.Writer.WriteString(message)
				rw.Writer.Flush()

				self.FlagMutex.Unlock()
			}
		case <-self.quit:
			Logger.log.Infof("PEER %s quit OUT message handler", self.PeerId)
			break
			//default:
			//	Logger.log.Info("Wait for sending message")
		}
	}
}

// Connected returns whether or not the peer is currently connected.
//
// This function is safe for concurrent access.
func (self PeerConn) Connected() bool {
	return atomic.LoadInt32(&self.connected) != 0 &&
		atomic.LoadInt32(&self.disconnect) == 0
}

// Disconnect disconnects the peer by closing the connection.  Calling this
// function when the peer is already disconnected or in the process of
// disconnecting will have no effect.
func (self PeerConn) Disconnect() {
	if atomic.AddInt32(&self.disconnect, 1) != 1 {
		return
	}

	Logger.log.Infof("Disconnecting %s", self)
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
func (self PeerConn) QueueMessageWithEncoding(msg wire.Message, doneChan chan<- struct{}) {
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

func (p *PeerConn) VerAckReceived() bool {
	p.FlagMutex.Lock()
	verAckReceived := p.verAckReceived
	p.FlagMutex.Unlock()

	return verAckReceived
}

// updateState updates the state of the connection request.
func (p *PeerConn) updateState(state ConnState) {
	p.stateMtx.Lock()
	p.state = state
	p.stateMtx.Unlock()
}

// State is the connection state of the requested connection.
func (p *PeerConn) State() ConnState {
	p.stateMtx.RLock()
	state := p.state
	p.stateMtx.RUnlock()
	return state
}
