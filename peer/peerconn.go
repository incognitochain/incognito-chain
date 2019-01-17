package peer

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"encoding/json"
	"reflect"
	"sync"

	"github.com/libp2p/go-libp2p-peer"
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/wire"
	"time"
)

type PeerConn struct {
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
	isForceClose     bool

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
func (peerConn *PeerConn) InMessageHandler(rw *bufio.ReadWriter) {
	peerConn.IsConnected = true
	for {
		Logger.log.Infof("PEER %s (address: %s) Reading stream", peerConn.RemotePeer.PeerID.Pretty(), peerConn.RemotePeer.RawAddress)
		str, err := rw.ReadString(DelimMessageByte)
		if err != nil {
			peerConn.IsConnected = false
			Logger.log.Error("---------------------------------------------------------------------")
			Logger.log.Errorf("InMessageHandler ERROR %s %s", peerConn.RemotePeerID.Pretty(), peerConn.RemotePeer.RawAddress)
			Logger.log.Error(err)
			Logger.log.Errorf("InMessageHandler QUIT %s %s", peerConn.RemotePeerID.Pretty(), peerConn.RemotePeer.RawAddress)
			Logger.log.Error("---------------------------------------------------------------------")
			close(peerConn.cWrite)
			return
		}

		// disconnect when received spam message
		if len(str) >= SPAM_MESSAGE_SIZE {
			Logger.log.Errorf("InMessageHandler received spam message")
			peerConn.ForceClose()
			continue
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
				peerConn.ListenerPeer.ReceivedHashMessage(hashMsg)
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
					if peerConn.Config.MessageListeners.OnTx != nil {
						peerConn.Config.MessageListeners.OnTx(peerConn, message.(*wire.MessageTx))
					}
				case reflect.TypeOf(&wire.MessageBlock{}):
					if peerConn.Config.MessageListeners.OnBlock != nil {
						peerConn.Config.MessageListeners.OnBlock(peerConn, message.(*wire.MessageBlock))
					}
				case reflect.TypeOf(&wire.MessageGetBlocks{}):
					if peerConn.Config.MessageListeners.OnGetBlocks != nil {
						peerConn.Config.MessageListeners.OnGetBlocks(peerConn, message.(*wire.MessageGetBlocks))
					}
				case reflect.TypeOf(&wire.MessageVersion{}):
					if peerConn.Config.MessageListeners.OnVersion != nil {
						versionMessage := message.(*wire.MessageVersion)
						peerConn.Config.MessageListeners.OnVersion(peerConn, versionMessage)
					}
				case reflect.TypeOf(&wire.MessageVerAck{}):
					peerConn.verAckReceived = true
					if peerConn.Config.MessageListeners.OnVerAck != nil {
						peerConn.Config.MessageListeners.OnVerAck(peerConn, message.(*wire.MessageVerAck))
					}
				case reflect.TypeOf(&wire.MessageGetAddr{}):
					if peerConn.Config.MessageListeners.OnGetAddr != nil {
						peerConn.Config.MessageListeners.OnGetAddr(peerConn, message.(*wire.MessageGetAddr))
					}
				case reflect.TypeOf(&wire.MessageAddr{}):
					if peerConn.Config.MessageListeners.OnGetAddr != nil {
						peerConn.Config.MessageListeners.OnAddr(peerConn, message.(*wire.MessageAddr))
					}
				case reflect.TypeOf(&wire.MessageBlockSigReq{}):
					if peerConn.Config.MessageListeners.OnRequestSign != nil {
						peerConn.Config.MessageListeners.OnRequestSign(peerConn, message.(*wire.MessageBlockSigReq))
					}
				case reflect.TypeOf(&wire.MessageInvalidBlock{}):
					if peerConn.Config.MessageListeners.OnInvalidBlock != nil {
						peerConn.Config.MessageListeners.OnInvalidBlock(peerConn, message.(*wire.MessageInvalidBlock))
					}
				case reflect.TypeOf(&wire.MessageBlockSig{}):
					if peerConn.Config.MessageListeners.OnBlockSig != nil {
						peerConn.Config.MessageListeners.OnBlockSig(peerConn, message.(*wire.MessageBlockSig))
					}
				case reflect.TypeOf(&wire.MessageGetChainState{}):
					if peerConn.Config.MessageListeners.OnGetChainState != nil {
						peerConn.Config.MessageListeners.OnGetChainState(peerConn, message.(*wire.MessageGetChainState))
					}
				case reflect.TypeOf(&wire.MessageChainState{}):
					if peerConn.Config.MessageListeners.OnChainState != nil {
						peerConn.Config.MessageListeners.OnChainState(peerConn, message.(*wire.MessageChainState))
					}
					/*case reflect.TypeOf(&wire.MessageRegistration{}):
					  if peerConn.Config.MessageListeners.OnRegistration != nil {
						  peerConn.Config.MessageListeners.OnRegistration(peerConn, message.(*wire.MessageRegistration))
					  }*/
				case reflect.TypeOf(&wire.MessageSwapRequest{}):
					if peerConn.Config.MessageListeners.OnSwapRequest != nil {
						peerConn.Config.MessageListeners.OnSwapRequest(peerConn, message.(*wire.MessageSwapRequest))
					}
				case reflect.TypeOf(&wire.MessageSwapSig{}):
					if peerConn.Config.MessageListeners.OnSwapSig != nil {
						peerConn.Config.MessageListeners.OnSwapSig(peerConn, message.(*wire.MessageSwapSig))
					}
				case reflect.TypeOf(&wire.MessageSwapUpdate{}):
					if peerConn.Config.MessageListeners.OnSwapUpdate != nil {
						peerConn.Config.MessageListeners.OnSwapUpdate(peerConn, message.(*wire.MessageSwapUpdate))
					}
				case reflect.TypeOf(&wire.MessageMsgCheck{}):
					peerConn.handleMsgCheck(message.(*wire.MessageMsgCheck))
				case reflect.TypeOf(&wire.MessageMsgCheckResp{}):
					peerConn.handleMsgCheckResp(message.(*wire.MessageMsgCheckResp))
				default:
					Logger.log.Warnf("InMessageHandler Received unhandled message of type % from %v", realType, peerConn)
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
func (peerConn *PeerConn) OutMessageHandler(rw *bufio.ReadWriter) {
	for {
		select {
		case outMsg := <-peerConn.sendMessageQueue:
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
				Logger.log.Infof("Send a message %s to %s", outMsg.message.MessageType(), peerConn.RemotePeer.PeerID.Pretty())
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
		case <-peerConn.cWrite:
			Logger.log.Infof("OutMessageHandler QUIT %s %s", peerConn.RemotePeerID.Pretty(), peerConn.RemotePeer.RawAddress)

			peerConn.IsConnected = false

			close(peerConn.cDisconnect)

			if peerConn.HandleDisconnected != nil {
				go peerConn.HandleDisconnected(peerConn)
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
func (peerConn *PeerConn) QueueMessageWithEncoding(msg wire.Message, doneChan chan<- struct{}) {
	go func() {
		if peerConn.IsConnected {
			data, _ := msg.JsonSerialize()
			if len(data) >= HEAVY_MESSAGE_SIZE && msg.MessageType() != wire.CmdMsgCheck && msg.MessageType() != wire.CmdMsgCheckResp {
				hash := common.HashH(data).String()

				Logger.log.Infof("QueueMessageWithEncoding HEAVY_MESSAGE_SIZE %s %s", hash, msg.MessageType())
				numRetries := 0

			BeginCheckHashMessage:
				numRetries++
				bTimeOut := false
				// new model for received response
				peerConn.cMsgHash[hash] = make(chan bool)
				cTimeOut := make(chan struct{})
				bOk := false
				// send msg for check has
				go func() {
					msgCheck, _ := wire.MakeEmptyMessage(wire.CmdMsgCheck)
					msgCheck.(*wire.MessageMsgCheck).Hash = hash
					peerConn.QueueMessageWithEncoding(msgCheck, nil)
				}()
				Logger.log.Infof("QueueMessageWithEncoding WAIT result check hash %s", hash)
				for {
					select {
					case accept := <-peerConn.cMsgHash[hash]:
						Logger.log.Infof("QueueMessageWithEncoding RECEIVED hash %s accept %s", hash, accept)
						bOk = accept
						cTimeOut = nil
						break
					case <-cTimeOut:
						Logger.log.Infof("QueueMessageWithEncoding RECEIVED time out")
						cTimeOut = nil
						bTimeOut = true
						break
					}
					if cTimeOut == nil {
						delete(peerConn.cMsgHash, hash)
						break
					}
				}
				// set time out for check message
				go func() {
					select {
					case <-time.NewTimer(10 * time.Second).C:
						if cTimeOut != nil {
							Logger.log.Infof("QueueMessageWithEncoding TIMER time out %s", hash)
							bTimeOut = true
							close(cTimeOut)
						}
						return
					}
				}()
				Logger.log.Infof("QueueMessageWithEncoding FINISHED check hash %s %s", hash, bOk)
				if bTimeOut && numRetries >= MAX_RETRIES_CHECK_HASH_MESSAGE {
					goto BeginCheckHashMessage
				}

				if bOk {
					peerConn.sendMessageQueue <- outMsg{message: msg, doneChan: doneChan}
				}
			} else {
				peerConn.sendMessageQueue <- outMsg{message: msg, doneChan: doneChan}
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

func (p *PeerConn) ForceClose() {
	p.isForceClose = true
	close(p.cClose)
}

func (p *PeerConn) CheckAccepted() bool {
	// check max conn
	// check max shard conn
	// check max unknown shard conn
	return true
}
