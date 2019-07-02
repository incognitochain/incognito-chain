package peer

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/wire"
	"reflect"
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
	bytes = append(bytes, DelimMessageByte)
	sample.Write(bytes)
	sample.WriteTo(rw) // In reality, my "decoder" would be reading from 'sample' and writing to 'target'
	rw.Writer.Flush()

	value, err := peerConn.readString(rw, DelimMessageByte, SpamMessageSize)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(value)
}

func TestPeerConn_InMessageHandler(t *testing.T) {
	peerConn := PeerConn{
		RemotePeer: &Peer{
			PublicKey: "abc1",
		},
		cWrite:      make(chan struct{}),
		cDisconnect: make(chan struct{}),
		cClose:      make(chan struct{}),
		isUnitTest:  true,
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
	messageBytes, err = common.GZipToBytes(messageBytes)
	if err != nil {
		t.Error(err)
	}
	messageHex := hex.EncodeToString(messageBytes)
	messageHex += DelimMessageStr

	sample.Write([]byte(messageHex))
	sample.WriteTo(rw)
	rw.Writer.Flush()
	err = peerConn.InMessageHandler(rw)
	if err != nil {
		t.Error(err)
	}
}
