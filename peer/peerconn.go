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

	"github.com/libp2p/go-libp2p-peer"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/ninjadotorg/cash-prototype/common"
	"github.com/ninjadotorg/cash-prototype/wire"
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
	IsConnected		   bool

	TargetAddress    ma.Multiaddr
	PeerID           peer.ID
	RawAddress       string
	ListeningAddress common.SimpleAddr

	// flagMutex sync.Mutex
	Config Config

	sendMessageQueue chan outMsg
	quit             chan struct{}
	disconnect       chan struct{}

	Peer             *Peer
	ValidatorAddress string
	ListenerPeer     *Peer

	HandleConnected    func(peerConn *PeerConn)
	HandleDisconnected func(peerConn *PeerConn)
	HandleFailed       func(peerConn *PeerConn)
}

/*
Handle all in message
*/
func (self *PeerConn) InMessageHandler(rw *bufio.ReadWriter) {
	self.IsConnected = true
	for {
		Logger.log.Infof("PEER %s Reading stream", self.PeerID.String())
		str, err := rw.ReadString('\n')
		if err != nil {
			self.IsConnected = false

			Logger.log.Infof("PEER %s quit IN message handler", self.PeerID)
			Logger.log.Errorf("InMessageHandler PEER %s ERROR %s", self.Peer.RawAddress, err)
			Logger.log.Infof("InMessageHandler PEER %s quit IN message handler", self.Peer.RawAddress)
			Logger.log.Info("PEER %s quit IN message handler")
			self.quit <- struct{}{}
			return
		}

		//Logger.log.Infof("Received message: %s \n", str)
		if str != "\n" {
			go func(msgStr string) {
				// Parse Message header
				jsonDecodeString, _ := hex.DecodeString(msgStr)
				messageHeader := jsonDecodeString[len(jsonDecodeString)-wire.MessageHeaderSize:]

				//Logger.log.Infof("Received message: %s \n", jsonDecodeString)

				commandInHeader := messageHeader[:12]
				commandInHeader = bytes.Trim(messageHeader, "\x00")
				Logger.log.Infof("Received message Type - %s %s", string(commandInHeader), self.PeerID)
				commandType := string(messageHeader[:len(commandInHeader)])
				var message, err = wire.MakeEmptyMessage(string(commandType))

				// Parse Message body
				messageBody := jsonDecodeString[:len(jsonDecodeString)-wire.MessageHeaderSize]
				//Logger.log.Infof("Message Body - %s %s", string(messageBody), self.PeerID)
				if err != nil {
					Logger.log.Error(err)
					return
				}
				if commandType != wire.CmdBlock {
					err = json.Unmarshal(messageBody, &message)
				} else {
					err = json.Unmarshal(messageBody, &message)
				}
				//temp := message.(map[string]interface{})
				if err != nil {
					Logger.log.Error(err)
					return
				}
				realType := reflect.TypeOf(message)
				log.Print(realType)
				// check type of Message
				switch realType {
				case reflect.TypeOf(&wire.MessageTx{}):
					if self.Config.MessageListeners.OnTx != nil {
						//self.flagMutex.Lock()
						self.Config.MessageListeners.OnTx(self, message.(*wire.MessageTx))
						//self.flagMutex.Unlock()
					}
				case reflect.TypeOf(&wire.MessageBlock{}):
					if self.Config.MessageListeners.OnBlock != nil {
						//self.flagMutex.Lock()
						self.Config.MessageListeners.OnBlock(self, message.(*wire.MessageBlock))
						//self.flagMutex.Unlock()
					}
				case reflect.TypeOf(&wire.MessageGetBlocks{}):
					if self.Config.MessageListeners.OnGetBlocks != nil {
						//self.flagMutex.Lock()
						self.Config.MessageListeners.OnGetBlocks(self, message.(*wire.MessageGetBlocks))
						//self.flagMutex.Unlock()
					}
				case reflect.TypeOf(&wire.MessageVersion{}):
					if self.Config.MessageListeners.OnVersion != nil {
						//self.flagMutex.Lock()
						versionMessage := message.(*wire.MessageVersion)
						self.Config.MessageListeners.OnVersion(self, versionMessage)
						//self.flagMutex.Unlock()
					}
				case reflect.TypeOf(&wire.MessageVerAck{}):
					//self.flagMutex.Lock()
					self.verAckReceived = true
					if self.Config.MessageListeners.OnVerAck != nil {
						self.Config.MessageListeners.OnVerAck(self, message.(*wire.MessageVerAck))
					}
					//self.flagMutex.Unlock()
				case reflect.TypeOf(&wire.MessageGetAddr{}):
					//self.flagMutex.Lock()
					if self.Config.MessageListeners.OnGetAddr != nil {
						self.Config.MessageListeners.OnGetAddr(self, message.(*wire.MessageGetAddr))
					}
					//self.flagMutex.Unlock()
				case reflect.TypeOf(&wire.MessageAddr{}):
					//self.flagMutex.Lock()
					if self.Config.MessageListeners.OnGetAddr != nil {
						self.Config.MessageListeners.OnAddr(self, message.(*wire.MessageAddr))
					}
					//self.flagMutex.Unlock()
				case reflect.TypeOf(&wire.MessageRequestSign{}):
					if self.Config.MessageListeners.OnRequestSign != nil {
						//self.flagMutex.Lock()
						self.Config.MessageListeners.OnRequestSign(self, message.(*wire.MessageRequestSign))
						//self.flagMutex.Unlock()
					}
				case reflect.TypeOf(&wire.MessageInvalidBlock{}):
					if self.Config.MessageListeners.OnInvalidBlock != nil {
						//self.flagMutex.Lock()
						self.Config.MessageListeners.OnInvalidBlock(self, message.(*wire.MessageInvalidBlock))
						//self.flagMutex.Unlock()
					}
				case reflect.TypeOf(&wire.MessageBlockSig{}):
					if self.Config.MessageListeners.OnBlockSig != nil {
						//self.flagMutex.Lock()
						self.Config.MessageListeners.OnBlockSig(self, message.(*wire.MessageBlockSig))
						//self.flagMutex.Unlock()
					}
				case reflect.TypeOf(&wire.MessageGetChainState{}):
					if self.Config.MessageListeners.OnGetChainState != nil {
						//self.flagMutex.Lock()
						self.Config.MessageListeners.OnGetChainState(self, message.(*wire.MessageGetChainState))
						//self.flagMutex.Unlock()
					}
				case reflect.TypeOf(&wire.MessageChainState{}):
					if self.Config.MessageListeners.OnChainState != nil {
						//self.flagMutex.Lock()
						self.Config.MessageListeners.OnChainState(self, message.(*wire.MessageChainState))
						//self.flagMutex.Unlock()
					}
				default:
					Logger.log.Warnf("InMessageHandler Received unhandled message of type %v "+
						"from %v", realType, self)
				}
			}(str)
		}
	}
}

/*
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
				// self.flagMutex.Lock()
				// TODO
				// send message
				messageByte, err := outMsg.msg.JsonSerialize()
				if err != nil {
					fmt.Println(err)
					// self.flagMutex.Unlock()
					continue
				}
				header := make([]byte, wire.MessageHeaderSize)
				CmdType, _ := wire.GetCmdType(reflect.TypeOf(outMsg.msg))
				copy(header[:], []byte(CmdType))
				messageByte = append(messageByte, header...)
				message := hex.EncodeToString(messageByte)
				message += "\n"
				Logger.log.Infof("Send a message %s to %s", outMsg.msg.MessageType(), self.Peer.RawAddress) // , string(messageByte)
				_, err = rw.Writer.WriteString(message)
				if err != nil {
					Logger.log.Critical("DM ERROR", err)
					// self.flagMutex.Unlock()
					return
				}
				err = rw.Writer.Flush()
				if err != nil {
					Logger.log.Critical("DM ERROR", err)
					// self.flagMutex.Unlock()
					return
				}
				// self.flagMutex.Unlock()

			}
		case <-self.quit:
			Logger.log.Infof("PEER %s quit OUT message handler", self.PeerID)

			self.IsConnected = false

			close(self.disconnect)

			if self.HandleDisconnected != nil {
				go self.HandleDisconnected(self)
			}

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
func (self *PeerConn) QueueMessageWithEncoding(msg wire.Message, doneChan chan<- struct{}) {
	if self.state == ConnEstablished {
		self.sendMessageQueue <- outMsg{msg: msg, doneChan: doneChan}
	}
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

func (p *PeerConn) Close() {
	close(p.quit)
}
