package peer

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/libp2p/go-libp2p-peer"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/ninjadotorg/cash-prototype/common"
	"github.com/ninjadotorg/cash-prototype/wire"
	"log"
	"reflect"
	"sync"
)

type PeerConn struct {
	connected int32

	RetryCount int32
	IsOutbound bool
	state      ConnState
	stateMtx   sync.RWMutex

	ReaderWriterStream *bufio.ReadWriter
	verAckReceived     bool
	VerValid           bool

	TargetAddress    ma.Multiaddr
	PeerId           peer.ID
	RawAddress       string
	ListeningAddress common.SimpleAddr

	flagMutex sync.Mutex
	Config    Config

	sendMessageQueue chan outMsg
	quit             chan struct{}
	disconnect       chan struct{}
	timeoutVerack    chan struct{}

	Peer             *Peer
	ValidatorAddress string
	ListenerPeer     *Peer

	HandleConnected    func(peerConn *PeerConn)
	HandleDisconnected func(peerConn *PeerConn)
	HandleFailed       func(peerConn *PeerConn)
}

/**
Handle all in message
*/
func (self *PeerConn) InMessageHandler(rw *bufio.ReadWriter) {
	for {
		Logger.log.Infof("PEER %s Reading stream", self.PeerId.String())
		str, err := rw.ReadString('\n')
		if err != nil {
			Logger.log.Error(err)

			Logger.log.Infof("PEER %s quit IN message handler", self.PeerId)
			self.quit <- struct{}{}
			return
		}

		//Logger.log.Infof("Received message: %s \n", str)
		if str != "\n" {

			// Parse Message header
			jsonDecodeString, _ := hex.DecodeString(str)
			messageHeader := jsonDecodeString[len(jsonDecodeString)-wire.MessageHeaderSize:]

			Logger.log.Infof("Received message: %s \n", jsonDecodeString)

			commandInHeader := messageHeader[:12]
			commandInHeader = bytes.Trim(messageHeader, "\x00")
			Logger.log.Infof("Message Type - %s %s", string(commandInHeader), self.PeerId)
			commandType := string(messageHeader[:len(commandInHeader)])
			var message, err = wire.MakeEmptyMessage(string(commandType))

			// Parse Message body
			messageBody := jsonDecodeString[:len(jsonDecodeString)-wire.MessageHeaderSize]
			Logger.log.Infof("Message Body - %s %s", string(messageBody), self.PeerId)
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
					self.flagMutex.Lock()
					self.Config.MessageListeners.OnTx(self, message.(*wire.MessageTx))
					self.flagMutex.Unlock()
				}
			case reflect.TypeOf(&wire.MessageBlock{}):
				if self.Config.MessageListeners.OnBlock != nil {
					self.flagMutex.Lock()
					self.Config.MessageListeners.OnBlock(self, message.(*wire.MessageBlock))
					self.flagMutex.Unlock()
				}
			case reflect.TypeOf(&wire.MessageGetBlocks{}):
				if self.Config.MessageListeners.OnGetBlocks != nil {
					self.flagMutex.Lock()
					self.Config.MessageListeners.OnGetBlocks(self, message.(*wire.MessageGetBlocks))
					self.flagMutex.Unlock()
				}
			case reflect.TypeOf(&wire.MessageVersion{}):
				if self.Config.MessageListeners.OnVersion != nil {
					self.flagMutex.Lock()
					versionMessage := message.(*wire.MessageVersion)
					self.Config.MessageListeners.OnVersion(self, versionMessage)
					self.flagMutex.Unlock()
				}
			case reflect.TypeOf(&wire.MessageVerAck{}):
				self.flagMutex.Lock()
				self.verAckReceived = true
				if self.Config.MessageListeners.OnVerAck != nil {
					self.Config.MessageListeners.OnVerAck(self, message.(*wire.MessageVerAck))
				}
				self.flagMutex.Unlock()
			case reflect.TypeOf(&wire.MessageGetAddr{}):
				self.flagMutex.Lock()
				if self.Config.MessageListeners.OnGetAddr != nil {
					self.Config.MessageListeners.OnGetAddr(self, message.(*wire.MessageGetAddr))
				}
				self.flagMutex.Unlock()
			case reflect.TypeOf(&wire.MessageAddr{}):
				self.flagMutex.Lock()
				if self.Config.MessageListeners.OnGetAddr != nil {
					self.Config.MessageListeners.OnAddr(self, message.(*wire.MessageAddr))
				}
				self.flagMutex.Unlock()
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
func (self *PeerConn) OutMessageHandler(rw *bufio.ReadWriter) {
	for {
		select {
		case outMsg := <-self.sendMessageQueue:
			{
				self.flagMutex.Lock()
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
				Logger.log.Infof("Send a message %s %s: %s", self.PeerId.String(), outMsg.msg.MessageType(), string(messageByte))
				rw.Writer.WriteString(message)
				rw.Writer.Flush()

				self.flagMutex.Unlock()

			}
		case <-self.quit:
			Logger.log.Infof("PEER %s quit OUT message handler", self.PeerId)

			self.disconnect <- struct{}{}

			if self.HandleDisconnected != nil {
				go self.HandleDisconnected(self)
			}

			break
		}
	}
}

// QueueMessageWithEncoding adds the passed bitcoin message to the peer send
// queue. This function is identical to QueueMessage, however it allows the
// caller to specify the wire encoding type that should be used when
// encoding/decoding blocks and transactions.
//
// This function is safe for concurrent access.
func (self *PeerConn) QueueMessageWithEncoding(msg wire.Message, doneChan chan<- struct{}) {
	self.sendMessageQueue <- outMsg{msg: msg, doneChan: doneChan}
}

func (p *PeerConn) VerAckReceived() bool {
	verAckReceived := p.verAckReceived
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
