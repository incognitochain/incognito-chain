package peer

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/wire"
	"github.com/stretchr/testify/assert"
	"reflect"
	"sync"
	"testing"
	"time"
)

var _ = func() (_ struct{}) {
	fmt.Println("This runs before init()!")
	Logger.Init(common.NewBackend(nil).Logger("test", true))
	return
}()

func TestPeerConn_ReadString(t *testing.T) {
	peerConn := PeerConn{}
	sample := bytes.NewBuffer(nil)
	rw := bufio.NewReadWriter(bufio.NewReader(sample), bufio.NewWriter(sample))
	text := "Hello World"
	bytes := make([]byte, 0)
	bytes = append(bytes, []byte(text)...)
	bytes = append(bytes, delimMessageByte)
	sample.Write(bytes)
	sample.WriteTo(rw) // In reality, my "decoder" would be reading from 'sample' and writing to 'target'
	rw.Writer.Flush()

	value, err := peerConn.readString(rw, delimMessageByte, spamMessageSize)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(value)
}

func TestPeerConn_ProcessInMessageStr(t *testing.T) {
	p1 := &Peer{}
	p1.SetPublicKey("abc1")
	peerConn := PeerConn{
		cWrite:      make(chan struct{}),
		cDisconnect: make(chan struct{}),
		cClose:      make(chan struct{}),
		isUnitTest:  true,
		remotePeer:  p1,
	}

	outMsg := outMsg{
		message: &wire.MessageVerAck{
			Timestamp: time.Now(),
			Valid:     true,
		},
	}
	messageBytes, err := outMsg.message.JsonSerialize()
	if err != nil {
		t.Error(err)
	}
	headerBytes := make([]byte, wire.MessageHeaderSize)
	cmdType, messageErr := wire.GetCmdType(reflect.TypeOf(outMsg.message))
	if messageErr != nil {
		t.Error(messageErr)
	}
	copy(headerBytes[:], []byte(cmdType))
	copy(headerBytes[wire.MessageCmdTypeSize:], []byte{outMsg.forwardType})
	if outMsg.forwardValue != nil {
		copy(headerBytes[wire.MessageCmdTypeSize+1:], []byte{*outMsg.forwardValue})
	}
	messageBytes = append(messageBytes, headerBytes...)
	//messageBytes = append(messageBytes, []byte(delimMessageStr)...)
	messageBytes, err = common.GZipFromBytes(messageBytes)
	if err != nil {
		t.Error(err)
	}
	messageHex := hex.EncodeToString(messageBytes)
	//messageHex += delimMessageStr

	err = peerConn.processInMessageString(messageHex)
	if err != nil {
		t.Error(err)
	}
}

func TestPeerConn_InMessageHandler(t *testing.T) {
	p1 := &Peer{}
	p1.SetPublicKey("abc1")
	peerConn := PeerConn{
		cWrite:      make(chan struct{}),
		cDisconnect: make(chan struct{}),
		cClose:      make(chan struct{}),
		isUnitTest:  true,
		remotePeer:  p1,
	}
	sample := bytes.NewBuffer(nil)
	rw := bufio.NewReadWriter(bufio.NewReader(sample), bufio.NewWriter(sample))
	outMsg := outMsg{
		message: &wire.MessageVerAck{
			Timestamp: time.Now(),
			Valid:     true,
		},
	}
	messageBytes, err := outMsg.message.JsonSerialize()
	if err != nil {
		t.Error(err)
	}
	headerBytes := make([]byte, wire.MessageHeaderSize)
	cmdType, messageErr := wire.GetCmdType(reflect.TypeOf(outMsg.message))
	if messageErr != nil {
		t.Error(messageErr)
	}
	copy(headerBytes[:], []byte(cmdType))
	copy(headerBytes[wire.MessageCmdTypeSize:], []byte{outMsg.forwardType})
	if outMsg.forwardValue != nil {
		copy(headerBytes[wire.MessageCmdTypeSize+1:], []byte{*outMsg.forwardValue})
	}
	messageBytes = append(messageBytes, headerBytes...)
	messageBytes, err = common.GZipFromBytes(messageBytes)
	if err != nil {
		t.Error(err)
	}
	messageHex := hex.EncodeToString(messageBytes)
	messageHex += delimMessageStr

	sample.Write([]byte(messageHex))
	sample.WriteTo(rw)
	rw.Writer.Flush()
	err = peerConn.inMessageHandler(rw)
	if err != nil {
		t.Error(err)
	}
}

func TestPeerConn_HandleMsgCheckResp(t *testing.T) {
	peerConn := PeerConn{
		cMsgHash:   make(map[string]chan bool),
		isUnitTest: true,
	}
	peerConn.cMsgHash["abc"] = make(chan bool)
	message := &wire.MessageMsgCheckResp{
		HashStr: "abc",
	}
	err := peerConn.handleMsgCheckResp(message)
	if err != nil {
		t.Error(err)
	}
}

func TestPeerConn_HandleMsgCheck(t *testing.T) {
	p1 := &Peer{}
	p1.SetPublicKey("abc1")
	p2 := &Peer{}
	p2.SetPublicKey("abc1")
	peerConn := PeerConn{
		cMsgHash:     make(map[string]chan bool),
		isUnitTest:   true,
		listenerPeer: p1,
		remotePeer:   p2,
	}
	peerConn.cMsgHash["abc"] = make(chan bool)
	message := &wire.MessageMsgCheck{
		HashStr: "abc",
	}
	peerConn.listenerPeer.HashToPool(message.HashStr)
	err := peerConn.handleMsgCheck(message)
	if err != nil {
		t.Error(err)
	}
}

func TestPeerConn_UpdateConnectionState(t *testing.T) {
	peerConn := PeerConn{
		stateMtx: sync.RWMutex{},
	}
	peerConn.setConnState(1)
	assert.Equal(t, uint8(1), uint8(peerConn.connState))
}
