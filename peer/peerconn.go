package peer

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"encoding/json"
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
	connState  ConnState
	stateMtx   sync.RWMutex

	ReaderWriterStream *bufio.ReadWriter
	verAckReceived     bool
	VerValid           bool
	IsConnected        bool

	TargetAddress    ma.Multiaddr
	PeerID           peer.ID
	RawAddress       string
	ListeningAddress common.SimpleAddr

	Config Config

	sendMessageQueue chan outMsg
	cDisconnect      chan struct{}
	cRead            chan struct{}
	cWrite           chan struct{}
	cClose           chan struct{}

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
		Logger.log.Infof("PEER %s (address: %s) Reading stream", self.Peer.PeerID.String(), self.Peer.RawAddress)
		str, err := rw.ReadString(DelimMessageByte)
		if err != nil {
			self.IsConnected = false
			Logger.log.Error("---------------------------------------------------------------------")
			Logger.log.Errorf("InMessageHandler ERROR %s %s", self.PeerID, self.Peer.RawAddress)
			Logger.log.Error(err)
			Logger.log.Errorf("InMessageHandler QUIT %s %s", self.PeerID, self.Peer.RawAddress)
			Logger.log.Error("---------------------------------------------------------------------")
			close(self.cWrite)
			return
		}

		if str != DelimMessageStr {
			go func(msgStr string) {
				// Parse Message header from last 24 bytes header message
				jsonDecodeString, _ := hex.DecodeString(msgStr)
				messageHeader := jsonDecodeString[len(jsonDecodeString)-wire.MessageHeaderSize:]

				// get cmd type in header message
				commandInHeader := messageHeader[:wire.MessageCmdTypeSize]
				commandInHeader = bytes.Trim(messageHeader, "\x00")
				Logger.log.Infof("Received message type %s from %s", string(commandInHeader), self.PeerID)
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
				case reflect.TypeOf(&wire.MessageRequestSign{}):
					if self.Config.MessageListeners.OnRequestSign != nil {
						self.Config.MessageListeners.OnRequestSign(self, message.(*wire.MessageRequestSign))
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
				Logger.log.Infof("Content: %s", string(messageByte))
				message := hex.EncodeToString(messageByte)
				// add end character to message (delim '\n')
				message += DelimMessageStr
				Logger.log.Infof("Content in hex encode: %s", string(message))

				// send on p2p stream
				Logger.log.Infof("Send a message %s to %s", outMsg.message.MessageType(), self.Peer.PeerID.String())
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
			Logger.log.Infof("OutMessageHandler QUIT %s %s", self.PeerID, self.Peer.RawAddress)

			self.IsConnected = false

			close(self.cDisconnect)

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
	if self.IsConnected {
		self.sendMessageQueue <- outMsg{message: msg, doneChan: doneChan}
	}
}

func (p *PeerConn) VerAckReceived() bool {
	verAckReceived := p.verAckReceived
	return verAckReceived
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
