package peer

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"encoding/json"
	"reflect"
	"sync"

	"github.com/libp2p/go-libp2p-peer"
	"github.com/ninjadotorg/constant/wire"
	"github.com/ninjadotorg/constant/common"
	"time"
)

type PeerConn struct {
	connected      int32
	connState      ConnState
	stateMtx       sync.RWMutex
	verAckReceived bool

	// channel
	sendMessageQueue chan outMsg
	cDisconnect      chan struct{}
	cRead            chan struct{}
	cWrite           chan struct{}
	cClose           chan struct{}
	cMsgHash         map[string]chan bool

	RetryCount int32

	// remote peer info
	RemotePeer       *Peer
	RemotePeerID     peer.ID
	RemoteRawAddress string
	IsOutbound       bool

	ReaderWriterStream *bufio.ReadWriter
	VerValid           bool
	IsConnected        bool

	Config Config

	ListenerPeer *Peer

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
		Logger.log.Infof("PEER %s (address: %s) Reading stream", self.RemotePeer.PeerID.String(), self.RemotePeer.RawAddress)
		str, err := rw.ReadString(DelimMessageByte)
		if err != nil {
			self.IsConnected = false
			Logger.log.Error("---------------------------------------------------------------------")
			Logger.log.Errorf("InMessageHandler ERROR %s %s", self.RemotePeerID, self.RemotePeer.RawAddress)
			Logger.log.Error(err)
			Logger.log.Errorf("InMessageHandler QUIT %s %s", self.RemotePeerID, self.RemotePeer.RawAddress)
			Logger.log.Error("---------------------------------------------------------------------")
			close(self.cWrite)
			return
		}

		if str != DelimMessageStr {
			go func(msgStr string) {
				// Parse Message header from last 24 bytes header message
				jsonDecodeString, _ := hex.DecodeString(msgStr)
				Logger.log.Infof("In message content : %s", string(jsonDecodeString))
				messageHeader := jsonDecodeString[len(jsonDecodeString)-wire.MessageHeaderSize:]

				// get cmd type in header message
				commandInHeader := messageHeader[:wire.MessageCmdTypeSize]
				commandInHeader = bytes.Trim(messageHeader, "\x00")
				commandType := string(messageHeader[:len(commandInHeader)])
				// convert to particular message from message cmd type
				var message, err = wire.MakeEmptyMessage(string(commandType))
				if err != nil {
					Logger.log.Error("Can not find particular message for message cmd type")
					Logger.log.Error(err)
					return
				}

				// Parse Message body
				messageBody := jsonDecodeString[:len(jsonDecodeString)-wire.MessageHeaderSize]
				// cache message hash S
				hashMsg := common.HashH(messageBody).String()
				self.ListenerPeer.ReceivedHashMessage(hashMsg)
				// cache message hash E
				err = json.Unmarshal(messageBody, &message)
				if err != nil {
					Logger.log.Error("Can not parse struct from json message")
					Logger.log.Error(err)
					return
				}
				realType := reflect.TypeOf(message)
				Logger.log.Infof("Cmd message type of struct %s", realType.String())

				// process message for each of message type
				switch realType {
				case reflect.TypeOf(&wire.MessageTx{}):
					if self.Config.MessageListeners.OnTx != nil {
						self.Config.MessageListeners.OnTx(self, message.(*wire.MessageTx))
					}
				case reflect.TypeOf(&wire.MessageBlock{}):
					if self.Config.MessageListeners.OnBlock != nil {
						self.Config.MessageListeners.OnBlock(self, message.(*wire.MessageBlock))
					}
				case reflect.TypeOf(&wire.MessageGetBlocks{}):
					if self.Config.MessageListeners.OnGetBlocks != nil {
						self.Config.MessageListeners.OnGetBlocks(self, message.(*wire.MessageGetBlocks))
					}
				case reflect.TypeOf(&wire.MessageVersion{}):
					if self.Config.MessageListeners.OnVersion != nil {
						versionMessage := message.(*wire.MessageVersion)
						self.Config.MessageListeners.OnVersion(self, versionMessage)
					}
				case reflect.TypeOf(&wire.MessageVerAck{}):
					self.verAckReceived = true
					if self.Config.MessageListeners.OnVerAck != nil {
						self.Config.MessageListeners.OnVerAck(self, message.(*wire.MessageVerAck))
					}
				case reflect.TypeOf(&wire.MessageGetAddr{}):
					if self.Config.MessageListeners.OnGetAddr != nil {
						self.Config.MessageListeners.OnGetAddr(self, message.(*wire.MessageGetAddr))
					}
				case reflect.TypeOf(&wire.MessageAddr{}):
					if self.Config.MessageListeners.OnGetAddr != nil {
						self.Config.MessageListeners.OnAddr(self, message.(*wire.MessageAddr))
					}
				case reflect.TypeOf(&wire.MessageBlockSigReq{}):
					if self.Config.MessageListeners.OnRequestSign != nil {
						self.Config.MessageListeners.OnRequestSign(self, message.(*wire.MessageBlockSigReq))
					}
				case reflect.TypeOf(&wire.MessageInvalidBlock{}):
					if self.Config.MessageListeners.OnInvalidBlock != nil {
						self.Config.MessageListeners.OnInvalidBlock(self, message.(*wire.MessageInvalidBlock))
					}
				case reflect.TypeOf(&wire.MessageBlockSig{}):
					if self.Config.MessageListeners.OnBlockSig != nil {
						self.Config.MessageListeners.OnBlockSig(self, message.(*wire.MessageBlockSig))
					}
				case reflect.TypeOf(&wire.MessageGetChainState{}):
					if self.Config.MessageListeners.OnGetChainState != nil {
						self.Config.MessageListeners.OnGetChainState(self, message.(*wire.MessageGetChainState))
					}
				case reflect.TypeOf(&wire.MessageChainState{}):
					if self.Config.MessageListeners.OnChainState != nil {
						self.Config.MessageListeners.OnChainState(self, message.(*wire.MessageChainState))
					}
					/*case reflect.TypeOf(&wire.MessageRegistration{}):
					  if self.Config.MessageListeners.OnRegistration != nil {
						  self.Config.MessageListeners.OnRegistration(self, message.(*wire.MessageRegistration))
					  }*/
				case reflect.TypeOf(&wire.MessageSwapRequest{}):
					if self.Config.MessageListeners.OnSwapRequest != nil {
						self.Config.MessageListeners.OnSwapRequest(self, message.(*wire.MessageSwapRequest))
					}
				case reflect.TypeOf(&wire.MessageSwapSig{}):
					if self.Config.MessageListeners.OnSwapSig != nil {
						self.Config.MessageListeners.OnSwapSig(self, message.(*wire.MessageSwapSig))
					}
				case reflect.TypeOf(&wire.MessageSwapUpdate{}):
					if self.Config.MessageListeners.OnSwapUpdate != nil {
						self.Config.MessageListeners.OnSwapUpdate(self, message.(*wire.MessageSwapUpdate))
					}
				case reflect.TypeOf(&wire.MessageMsgCheck{}):
					self.handleMsgCheck(message.(*wire.MessageMsgCheck))
				case reflect.TypeOf(&wire.MessageMsgCheckResp{}):
					self.handleMsgCheckResp(message.(*wire.MessageMsgCheckResp))
				default:
					Logger.log.Warnf("InMessageHandler Received unhandled message of type % from %v", realType, self)
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
				// Create and send message
				messageByte, err := outMsg.message.JsonSerialize()
				if err != nil {
					Logger.log.Error("Can not serialize json format for message:" + outMsg.message.MessageType())
					Logger.log.Error(err)
					continue
				}

				// add 24 bytes header into message
				header := make([]byte, wire.MessageHeaderSize)
				cmdType, _ := wire.GetCmdType(reflect.TypeOf(outMsg.message))
				copy(header[:], []byte(cmdType))
				messageByte = append(messageByte, header...)
				Logger.log.Infof("Out message TYPE %s CONTENT %s", cmdType, string(messageByte))
				message := hex.EncodeToString(messageByte)
				//Logger.log.Infof("Content in hex encode: %s", string(message))
				// add end character to message (delim '\n')
				message += DelimMessageStr

				// send on p2p stream
				Logger.log.Infof("Send a message %s to %s", outMsg.message.MessageType(), self.RemotePeer.PeerID.String())
				_, err = rw.Writer.WriteString(message)
				if err != nil {
					Logger.log.Critical("DM ERROR", err)
					continue
				}
				err = rw.Writer.Flush()
				if err != nil {
					Logger.log.Critical("DM ERROR", err)
					continue
				}
			}
		case <-self.cWrite:
			Logger.log.Infof("OutMessageHandler QUIT %s %s", self.RemotePeerID, self.RemotePeer.RawAddress)

			self.IsConnected = false

			close(self.cDisconnect)

			if self.HandleDisconnected != nil {
				go self.HandleDisconnected(self)
			}

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
func (self *PeerConn) QueueMessageWithEncoding(msg wire.Message, doneChan chan<- struct{}) {
	go func() {
		if self.IsConnected {
			data, _ := msg.JsonSerialize()
			if len(data) >= HEAVY_MESSAGE_SIZE && msg.MessageType() != wire.CmdMsgCheck && msg.MessageType() != wire.CmdMsgCheckResp {
				hash := common.HashH(data).String()

				Logger.log.Infof("QueueMessageWithEncoding HEAVY_MESSAGE_SIZE %s %s", hash, msg.MessageType())

				self.cMsgHash[hash] = make(chan bool)
				cTimeOut := make(chan struct{})
				bOk := false
				// send msg for check hash
				go func() {
					msgCheck, _ := wire.MakeEmptyMessage(wire.CmdMsgCheck)
					msgCheck.(*wire.MessageMsgCheck).Hash = hash
					self.QueueMessageWithEncoding(msgCheck, nil)
				}()
				Logger.log.Infof("QueueMessageWithEncoding WAIT result check hash %s", hash)
				for {
					select {
					case accept := <-self.cMsgHash[hash]:
						Logger.log.Infof("QueueMessageWithEncoding RECEIVED hash %s accept %s", hash, accept)
						bOk = accept
						cTimeOut = nil
						break
					case <-cTimeOut:
						Logger.log.Infof("QueueMessageWithEncoding RECEIVED time out")
						cTimeOut = nil
						break
					}
					if cTimeOut == nil {
						delete(self.cMsgHash, hash)
						break
					}
				}
				Logger.log.Infof("QueueMessageWithEncoding FINISHED check hash %s %s", hash, bOk)
				// set time out for check message
				go func() {
					select {
					case <-time.NewTimer(10 * time.Second).C:
						if cTimeOut != nil {
							Logger.log.Infof("QueueMessageWithEncoding TIMER time out %s", hash)
							close(cTimeOut)
						}
						return
					}
				}()
				if bOk {
					self.sendMessageQueue <- outMsg{message: msg, doneChan: doneChan}
				}
			} else {
				self.sendMessageQueue <- outMsg{message: msg, doneChan: doneChan}
			}
		}
	}()
}

func (p *PeerConn) handleMsgCheck(msg *wire.MessageMsgCheck) {
	Logger.log.Infof("handleMsgCheck %s", msg.Hash)
	msgResp, _ := wire.MakeEmptyMessage(wire.CmdMsgCheckResp)
	if p.ListenerPeer.CheckHashMessage(msg.Hash) {
		msgResp.(*wire.MessageMsgCheckResp).Hash = msg.Hash
		msgResp.(*wire.MessageMsgCheckResp).Accept = false
	} else {
		msgResp.(*wire.MessageMsgCheckResp).Hash = msg.Hash
		msgResp.(*wire.MessageMsgCheckResp).Accept = true
	}
	p.QueueMessageWithEncoding(msgResp, nil)
}

func (p *PeerConn) handleMsgCheckResp(msg *wire.MessageMsgCheckResp) {
	Logger.log.Infof("handleMsgCheckResp %s", msg.Hash)
	m, ok := p.cMsgHash[msg.Hash]
	if ok {
		m <- msg.Accept
	}
}

func (p *PeerConn) VerAckReceived() bool {
	return p.verAckReceived
}

// updateState updates the state of the connection request.
func (p *PeerConn) updateConnState(connState ConnState) {
	p.stateMtx.Lock()
	p.connState = connState
	p.stateMtx.Unlock()
}

// State is the connection state of the requested connection.
func (p *PeerConn) ConnState() ConnState {
	p.stateMtx.RLock()
	connState := p.connState
	p.stateMtx.RUnlock()
	return connState
}

func (p *PeerConn) Close() {
	close(p.cClose)
}
