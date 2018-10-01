package wire

import (
	"encoding/hex"
	"encoding/json"
	"time"

	"github.com/libp2p/go-libp2p-peer"
	"github.com/ninjadotorg/cash-prototype/common"
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
	PublicKey        string
}

func (self MessageVersion) MessageType() string {
	return CmdVersion
}

func (self MessageVersion) MaxPayloadLength(pver int) int {
	return MaxBlockPayload
}

func (self MessageVersion) JsonSerialize() ([]byte, error) {
	jsonBytes, err := json.Marshal(self)
	return jsonBytes, err
}

func (self MessageVersion) JsonDeserialize(jsonStr string) error {
	jsonDecodeString, _ := hex.DecodeString(jsonStr)
	err := json.Unmarshal([]byte(jsonDecodeString), self)
	return err
}

func (self MessageVersion) SetSenderID(senderID peer.ID) error {
	return nil
}
