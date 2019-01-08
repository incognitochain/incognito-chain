package peer

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"reflect"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p-peer"
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/wire"
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
	isOutbound       bool
	isOutboundMtx    sync.Mutex
	isForceClose     bool
	isForceCloseMtx  sync.Mutex

	RWStream       *bufio.ReadWriter
	VerValid       bool
	isConnected    bool
	isConnectedMtx sync.Mutex

	Config Config

	ListenerPeer *Peer

	HandleConnected    func(peerConn *PeerConn)
	HandleDisconnected func(peerConn *PeerConn)
	HandleFailed       func(peerConn *PeerConn)
}

func (self *PeerConn) GetIsOutbound() bool {
	self.isOutboundMtx.Lock()
	defer self.isOutboundMtx.Unlock()
	return self.isOutbound
}

func (self *PeerConn) SetIsOutbound(v bool) {
	self.isOutboundMtx.Lock()
	defer self.isOutboundMtx.Unlock()
	self.isOutbound = v
}

func (self *PeerConn) GetIsForceClose() bool {
	self.isForceCloseMtx.Lock()
	defer self.isForceCloseMtx.Unlock()
	return self.isForceClose
}

func (self *PeerConn) SetIsForceClose(v bool) {
	self.isForceCloseMtx.Lock()
	defer self.isForceCloseMtx.Unlock()
	self.isForceClose = v
}

func (self *PeerConn) GetIsConnected() bool {
	self.isConnectedMtx.Lock()
	defer self.isConnectedMtx.Unlock()
	return self.isConnected
}

func (self *PeerConn) SetIsConnected(v bool) {
	self.isConnectedMtx.Lock()
	defer self.isConnectedMtx.Unlock()
	self.isConnected = v
}

func (self *PeerConn) ReadString(rw *bufio.ReadWriter, delim byte, maxReadBytes int) (string, error) {
	buf := make([]byte, 0)
	bufL := 0
	for {
		b, err := rw.ReadByte()
		if err != nil {
			return "", err
		}
		if b == delim {
			break
		}
		buf = append(buf, b)
		bufL++
		if bufL > maxReadBytes {
			return "", errors.New("Limit bytes for message")
		}
	}

	return string(buf), nil
}

/*
Handle all in message
*/
func (self *PeerConn) InMessageHandler(rw *bufio.ReadWriter) {
	self.SetIsConnected(true)
	for {
		Logger.log.Infof("PEER %s (address: %s) Reading stream", self.RemotePeer.PeerID.Pretty(), self.RemotePeer.RawAddress)
		//str, errR := rw.ReadString(DelimMessageByte)
		//if errR != nil {
		//	self.SetIsConnected(false)
		//	Logger.log.Error("---------------------------------------------------------------------")
		//	Logger.log.Errorf("InMessageHandler ERROR %s %s", self.RemotePeerID.Pretty(), self.RemotePeer.RawAddress)
		//	Logger.log.Error(errR)
		//	Logger.log.Errorf("InMessageHandler QUIT %s %s", self.RemotePeerID.Pretty(), self.RemotePeer.RawAddress)
		//	Logger.log.Error("---------------------------------------------------------------------")
		//	close(self.cWrite)
		//	return
		//}
		str, errR := self.ReadString(rw, DelimMessageByte, SPAM_MESSAGE_SIZE)
		if errR != nil {
			self.SetIsConnected(false)
			Logger.log.Error("---------------------------------------------------------------------")
			Logger.log.Errorf("InMessageHandler ERROR %s %s", self.RemotePeerID.Pretty(), self.RemotePeer.RawAddress)
			Logger.log.Error(errR)
			Logger.log.Errorf("InMessageHandler QUIT %s %s", self.RemotePeerID.Pretty(), self.RemotePeer.RawAddress)
			Logger.log.Error("---------------------------------------------------------------------")
			close(self.cWrite)
			return
		}

		if str != DelimMessageStr {
			go func(msgStr string) {
				// Parse Message header from last 24 bytes header message
				jsonDecodeBytes, _ := hex.DecodeString(msgStr)

				// unzip data before process
				jsonDecodeBytes, err := common.GZipFromBytes(jsonDecodeBytes)
				if err != nil {
					Logger.log.Error("Can unzip from message")
					Logger.log.Error(err)
					return
				}
				// disconnect when received spam message
				//if len(jsonDecodeBytes) >= SPAM_MESSAGE_SIZE {
				//	Logger.log.Error("InMessageHandler received spam message")
				//	self.ForceClose()
				//	return
				//}

				Logger.log.Infof("In message content : %s", string(jsonDecodeBytes))

				// Parse Message body
				messageBody := jsonDecodeBytes[:len(jsonDecodeBytes)-wire.MessageHeaderSize]

				messageHeader := jsonDecodeBytes[len(jsonDecodeBytes)-wire.MessageHeaderSize:]
				// check forward
				if self.Config.MessageListeners.GetCurrentRoleShard != nil {
					cRole, cShard := self.Config.MessageListeners.GetCurrentRoleShard()
					if cShard != nil {
						fT := messageHeader[wire.MessageCmdTypeSize]
						if fT == MESSAGE_TO_SHARD {
							fS := messageHeader[wire.MessageCmdTypeSize+1]
							if *cShard != fS {
								if self.Config.MessageListeners.PushRawBytesToShard != nil {
									self.Config.MessageListeners.PushRawBytesToShard(self, &jsonDecodeBytes, *cShard)
								}
								return
							}
						}
					}
					if cRole != "" {
						fT := messageHeader[wire.MessageCmdTypeSize]
						if fT == MESSAGE_TO_BEACON && cRole != "beacon" {
							if self.Config.MessageListeners.PushRawBytesToBeacon != nil {
								self.Config.MessageListeners.PushRawBytesToBeacon(self, &jsonDecodeBytes)
							}
							return
						}
					}
				}

				// get cmd type in header message
				commandInHeader := bytes.Trim(messageHeader[:wire.MessageCmdTypeSize], "\x00")
				commandType := string(messageHeader[:len(commandInHeader)])
				// convert to particular message from message cmd type
				message, err := wire.MakeEmptyMessage(string(commandType))
				if err != nil {
					Logger.log.Error("Can not find particular message for message cmd type")
					Logger.log.Error(err)
					return
				}

				err = json.Unmarshal(messageBody, &message)
				if err != nil {
					Logger.log.Error("Can not parse struct from json message")
					Logger.log.Error(err)
					return
				}
				realType := reflect.TypeOf(message)
				Logger.log.Infof("Cmd message type of struct %s", realType.String())

				// cache message hash S
				hashMsg := message.Hash()
				if self.ListenerPeer.CheckHashPool(hashMsg) {
					Logger.log.Infof("InMessageHandler existed hash message %s", hashMsg)
					return
				}
				self.ListenerPeer.HashToPool(hashMsg)
				// cache message hash E

				// process message for each of message type
				switch realType {
				case reflect.TypeOf(&wire.MessageTx{}):
					if self.Config.MessageListeners.OnTx != nil {
						self.Config.MessageListeners.OnTx(self, message.(*wire.MessageTx))
					}
				case reflect.TypeOf(&wire.MessageBlockShard{}):
					if self.Config.MessageListeners.OnBlockShard != nil {
						self.Config.MessageListeners.OnBlockShard(self, message.(*wire.MessageBlockShard))
					}
				case reflect.TypeOf(&wire.MessageBlockBeacon{}):
					if self.Config.MessageListeners.OnBlockBeacon != nil {
						self.Config.MessageListeners.OnBlockBeacon(self, message.(*wire.MessageBlockBeacon))
					}
				case reflect.TypeOf(&wire.MessageCrossShard{}):
					if self.Config.MessageListeners.OnCrossShard != nil {
						self.Config.MessageListeners.OnCrossShard(self, message.(*wire.MessageCrossShard))
					}
				case reflect.TypeOf(&wire.MessageShardToBeacon{}):
					if self.Config.MessageListeners.OnShardToBeacon != nil {
						self.Config.MessageListeners.OnShardToBeacon(self, message.(*wire.MessageShardToBeacon))
					}
				case reflect.TypeOf(&wire.MessageGetBlockBeacon{}):
					if self.Config.MessageListeners.OnGetBlockBeacon != nil {
						self.Config.MessageListeners.OnGetBlockBeacon(self, message.(*wire.MessageGetBlockBeacon))
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
				case reflect.TypeOf(&wire.MessageBFTPropose{}):
					if self.Config.MessageListeners.OnBFTPropose != nil {
						self.Config.MessageListeners.OnBFTPropose(self, message.(*wire.MessageBFTPropose))
					}
				case reflect.TypeOf(&wire.MessageBFTPrepare{}):
					if self.Config.MessageListeners.OnBFTPrepare != nil {
						self.Config.MessageListeners.OnBFTPrepare(self, message.(*wire.MessageBFTPrepare))
					}
				case reflect.TypeOf(&wire.MessageBFTCommit{}):
					if self.Config.MessageListeners.OnBFTCommit != nil {
						self.Config.MessageListeners.OnBFTCommit(self, message.(*wire.MessageBFTCommit))
					}
				case reflect.TypeOf(&wire.MessageBFTReply{}):
					if self.Config.MessageListeners.OnBFTReply != nil {
						self.Config.MessageListeners.OnBFTReply(self, message.(*wire.MessageBFTReply))
					}
					// case reflect.TypeOf(&wire.MessageInvalidBlock{}):
					// 	if self.Config.MessageListeners.OnInvalidBlock != nil {
					// 		self.Config.MessageListeners.OnInvalidBlock(self, message.(*wire.MessageInvalidBlock))
					// 	}
				case reflect.TypeOf(&wire.MessageGetBeaconState{}):
					if self.Config.MessageListeners.OnGetBeaconState != nil {
						self.Config.MessageListeners.OnGetBeaconState(self, message.(*wire.MessageGetBeaconState))
					}
				case reflect.TypeOf(&wire.MessageBeaconState{}):
					if self.Config.MessageListeners.OnBeaconState != nil {
						self.Config.MessageListeners.OnBeaconState(self, message.(*wire.MessageBeaconState))
					}
				case reflect.TypeOf(&wire.MessageGetShardState{}):
					if self.Config.MessageListeners.OnGetShardState != nil {
						self.Config.MessageListeners.OnGetShardState(self, message.(*wire.MessageGetShardState))
					}
				case reflect.TypeOf(&wire.MessageShardState{}):
					if self.Config.MessageListeners.OnShardState != nil {
						self.Config.MessageListeners.OnShardState(self, message.(*wire.MessageShardState))
					}
					/*case reflect.TypeOf(&wire.MessageRegistration{}):
					  if self.Config.MessageListeners.OnRegistration != nil {
						  self.Config.MessageListeners.OnRegistration(self, message.(*wire.MessageRegistration))
					  }*/
					// case reflect.TypeOf(&wire.MessageSwapRequest{}):
					// 	if self.Config.MessageListeners.OnSwapRequest != nil {
					// 		self.Config.MessageListeners.OnSwapRequest(self, message.(*wire.MessageSwapRequest))
					// 	}
					// case reflect.TypeOf(&wire.MessageSwapSig{}):
					// 	if self.Config.MessageListeners.OnSwapSig != nil {
					// 		self.Config.MessageListeners.OnSwapSig(self, message.(*wire.MessageSwapSig))
					// 	}
					// case reflect.TypeOf(&wire.MessageSwapUpdate{}):
					// 	if self.Config.MessageListeners.OnSwapUpdate != nil {
					// 		self.Config.MessageListeners.OnSwapUpdate(self, message.(*wire.MessageSwapUpdate))
					// 	}
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
				if outMsg.rawBytes != nil && len(*outMsg.rawBytes) > 0 {
					Logger.log.Infof("OutMessageHandler with raw bytes")
					message := hex.EncodeToString(*outMsg.rawBytes)
					message += DelimMessageStr
					_, err := rw.Writer.WriteString(message)
					if err != nil {
						Logger.log.Critical("DM ERROR", err)
						continue
					}
					err = rw.Writer.Flush()
					if err != nil {
						Logger.log.Critical("DM ERROR", err)
						continue
					}
				} else {
					// Create and send messageHex
					messageBytes, err := outMsg.message.JsonSerialize()
					if err != nil {
						Logger.log.Error("Can not serialize json format for messageHex:" + outMsg.message.MessageType())
						Logger.log.Error(err)
						continue
					}

					// add 24 bytes headerBytes into messageHex
					headerBytes := make([]byte, wire.MessageHeaderSize)
					cmdType, _ := wire.GetCmdType(reflect.TypeOf(outMsg.message))
					copy(headerBytes[:], []byte(cmdType))
					copy(headerBytes[wire.MessageCmdTypeSize:], []byte{outMsg.forwardType})
					if outMsg.forwardValue != nil {
						copy(headerBytes[wire.MessageCmdTypeSize+1:], []byte{*outMsg.forwardValue})
					}
					messageBytes = append(messageBytes, headerBytes...)
					Logger.log.Infof("Out messageHex TYPE %s CONTENT %s", cmdType, string(messageBytes))

					// zip data before send
					messageBytes, err = common.GZipToBytes(messageBytes)
					if err != nil {
						Logger.log.Error("Can not gzip for messageHex:" + outMsg.message.MessageType())
						Logger.log.Error(err)
						continue
					}
					messageHex := hex.EncodeToString(messageBytes)
					//Logger.log.Infof("Content in hex encode: %s", string(messageHex))
					// add end character to messageHex (delim '\n')
					messageHex += DelimMessageStr

					// send on p2p stream
					Logger.log.Infof("Send a messageHex %s to %s", outMsg.message.MessageType(), self.RemotePeer.PeerID.Pretty())
					_, err = rw.Writer.WriteString(messageHex)
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
				continue
			}
		case <-self.cWrite:
			Logger.log.Infof("OutMessageHandler QUIT %s %s", self.RemotePeerID.Pretty(), self.RemotePeer.RawAddress)

			self.SetIsConnected(false)

			close(self.cDisconnect)

			if self.HandleDisconnected != nil {
				go self.HandleDisconnected(self)
			}

			return
		}
	}
}

func (self *PeerConn) checkMessageHashBeforeSend(hash string) bool {
	numRetries := 0
BeginCheckHashMessage:
	numRetries++
	bTimeOut := false
	// new model for received response
	self.cMsgHash[hash] = make(chan bool)
	cTimeOut := make(chan struct{})
	bCheck := false
	// send msg for check has
	go func() {
		msgCheck, err := wire.MakeEmptyMessage(wire.CmdMsgCheck)
		if err != nil {
			Logger.log.Error("checkMessageHashBeforeSend error", err)
			return
		}
		msgCheck.(*wire.MessageMsgCheck).HashStr = hash
		self.QueueMessageWithEncoding(msgCheck, nil, MESSAGE_TO_PEER, nil)
	}()
	// set time out for check message
	go func() {
		select {
		case <-time.NewTimer(MAX_TIMEOUT_CHECK_HASH_MESSAGE * time.Second).C:
			if cTimeOut != nil {
				Logger.log.Infof("checkMessageHashBeforeSend TIMER time out %s", hash)
				bTimeOut = true
				close(cTimeOut)
			}
			return
		}
	}()
	Logger.log.Infof("checkMessageHashBeforeSend WAIT result check hash %s", hash)
	select {
	case bCheck = <-self.cMsgHash[hash]:
		Logger.log.Infof("checkMessageHashBeforeSend RECEIVED hash %s bAccept %s", hash, bCheck)
		cTimeOut = nil
		break
	case <-cTimeOut:
		Logger.log.Infof("checkMessageHashBeforeSend RECEIVED time out %d", numRetries)
		cTimeOut = nil
		bTimeOut = true
		break
	}
	if cTimeOut == nil {
		delete(self.cMsgHash, hash)
	}
	Logger.log.Infof("checkMessageHashBeforeSend FINISHED check hash %s %s", hash, bCheck)
	if bTimeOut && numRetries < MAX_RETRIES_CHECK_HASH_MESSAGE {
		goto BeginCheckHashMessage
	}
	return bCheck
}

// QueueMessageWithEncoding adds the passed Constant message to the peer send
// queue. This function is identical to QueueMessage, however it allows the
// caller to specify the wire encoding type that should be used when
// encoding/decoding blocks and transactions.
//
// This function is safe for concurrent access.
func (self *PeerConn) QueueMessageWithEncoding(msg wire.Message, doneChan chan<- struct{}, forwardType byte, forwardValue *byte) {
	Logger.log.Infof("QueueMessageWithEncoding %s %s", self.RemotePeer.PeerID.Pretty(), msg.MessageType())
	go func() {
		if self.GetIsConnected() {
			data, _ := msg.JsonSerialize()
			if len(data) >= HEAVY_MESSAGE_SIZE && msg.MessageType() != wire.CmdMsgCheck && msg.MessageType() != wire.CmdMsgCheckResp {
				hash := msg.Hash()
				Logger.log.Infof("QueueMessageWithEncoding HEAVY_MESSAGE_SIZE %s %s", hash, msg.MessageType())

				if self.checkMessageHashBeforeSend(hash) {
					self.sendMessageQueue <- outMsg{
						message:      msg,
						doneChan:     doneChan,
						forwardType:  forwardType,
						forwardValue: forwardValue,
					}
				}
			} else {
				self.sendMessageQueue <- outMsg{
					message:      msg,
					doneChan:     doneChan,
					forwardType:  forwardType,
					forwardValue: forwardValue,
				}
			}
		}
	}()
}

func (self *PeerConn) QueueMessageWithBytes(msgBytes *[]byte, doneChan chan<- struct{}) {
	Logger.log.Infof("QueueMessageWithBytes %s", self.RemotePeer.PeerID.Pretty())
	if msgBytes == nil || len(*msgBytes) <= 0 {
		return
	}
	go func() {
		if self.GetIsConnected() {
			data := (*msgBytes)[wire.MessageHeaderSize:]
			if len(data) >= HEAVY_MESSAGE_SIZE {
				hash := common.HashH(data).String()
				Logger.log.Infof("QueueMessageWithBytes HEAVY_MESSAGE_SIZE %s", hash)

				if self.checkMessageHashBeforeSend(hash) {
					self.sendMessageQueue <- outMsg{
						rawBytes: msgBytes,
						doneChan: doneChan,
					}
				}
			} else {
				self.sendMessageQueue <- outMsg{
					rawBytes: msgBytes,
					doneChan: doneChan,
				}
			}
		}
	}()
}

func (p *PeerConn) handleMsgCheck(msg *wire.MessageMsgCheck) {
	Logger.log.Infof("handleMsgCheck %s", msg.HashStr)
	msgResp, err := wire.MakeEmptyMessage(wire.CmdMsgCheckResp)
	if err != nil {
		Logger.log.Error("handleMsgCheck error", err)
		return
	}
	if p.ListenerPeer.CheckHashPool(msg.HashStr) {
		msgResp.(*wire.MessageMsgCheckResp).HashStr = msg.HashStr
		msgResp.(*wire.MessageMsgCheckResp).Accept = false
	} else {
		msgResp.(*wire.MessageMsgCheckResp).HashStr = msg.HashStr
		msgResp.(*wire.MessageMsgCheckResp).Accept = true
	}
	p.QueueMessageWithEncoding(msgResp, nil, MESSAGE_TO_PEER, nil)
}

func (p *PeerConn) handleMsgCheckResp(msg *wire.MessageMsgCheckResp) {
	Logger.log.Infof("handleMsgCheckResp %s", msg.HashStr)
	m, ok := p.cMsgHash[msg.HashStr]
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
	defer p.stateMtx.Unlock()
	p.connState = connState
}

// State is the connection state of the requested connection.
func (p *PeerConn) ConnState() ConnState {
	p.stateMtx.RLock()
	defer p.stateMtx.RUnlock()
	connState := p.connState
	return connState
}

func (p *PeerConn) Close() {
	close(p.cClose)
}

func (p *PeerConn) ForceClose() {
	p.SetIsForceClose(true)
	close(p.cClose)
}
