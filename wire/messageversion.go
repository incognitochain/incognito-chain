package wire

import (
	"encoding/json"
	"time"
	"encoding/hex"
	"github.com/ninjadotorg/cash-prototype/common"
	"github.com/libp2p/go-libp2p-peer"
)

type MessageVersion struct {
	ProtocolVersion  int
	Timestamp        time.Time
	RemoteAddress    common.SimpleAddr
	RawRemoteAddress string
	RemotePeerId     peer.ID
	LocalAddress     common.SimpleAddr
	RawLocalAddress  string
	LocalPeerId      peer.ID
	LastBlock        int
}

func (self MessageVersion) MessageType() string {
	return CmdVersion
}

func (self MessageVersion) MaxPayloadLength(pver int) int {
	return MaxBlockPayload
}

func (self MessageVersion) JsonSerialize() (string, error) {
	jsonStr, err := json.Marshal(self)
	header := make([]byte, MessageHeaderSize)
	copy(header[:], self.MessageType())
	jsonStr = append(jsonStr, header...)
	return hex.EncodeToString(jsonStr), err
}

func (self MessageVersion) JsonDeserialize(jsonStr string) error {
	jsonDecodeString, _ := hex.DecodeString(jsonStr)
	err := json.Unmarshal([]byte(jsonDecodeString), self)
	return err
}
