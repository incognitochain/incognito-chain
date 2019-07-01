package peer

import (
	"bufio"
	"bytes"
	"fmt"
	"testing"
)

func TestPeerConn_ReadString(t *testing.T) {
	peerConn := PeerConn{}
	sample := bytes.NewBuffer(nil)
	target := bytes.NewBuffer(nil)
	buf := bufio.NewReadWriter(bufio.NewReader(sample), bufio.NewWriter(target))
	sample.Write([]byte{6, 0, 3, DelimMessageByte})
	sample.WriteTo(buf) // In reality, my "decoder" would be reading from 'sample' and writing to 'target'
	buf.Writer.Flush()

	value, err := peerConn.readString(buf, DelimMessageByte, SpamMessageSize)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(value)
}
