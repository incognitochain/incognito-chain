package wire

import (
	"encoding/hex"
	"encoding/json"
	"time"

	"github.com/libp2p/go-libp2p-peer"
	"github.com/ninjadotorg/constant/common"
)

const (
	MaxVersionPayload = 1000 // 1 1Kb
)

type MessageVersion struct {
	ProtocolVersion  string
	Timestamp        time.Time
	RemoteAddress    common.SimpleAddr
	RawRemoteAddress string
	RemotePeerId     peer.ID
	LocalAddress     common.SimpleAddr
	RawLocalAddress  string
	LocalPeerId      peer.ID
	PublicKey        string
	SignDataB58      string
}

func (self MessageVersion) MessageType() string {
	return CmdVersion
}

func (self MessageVersion) MaxPayloadLength(pver int) int {
	return MaxVersionPayload
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
