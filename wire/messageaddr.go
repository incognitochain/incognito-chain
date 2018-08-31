package wire

import (
	"encoding/json"
	"github.com/ninjadotorg/cash-prototype/common"
	"github.com/libp2p/go-libp2p-peer"
)

type MessageAddr struct {
	RemoteAddress    common.SimpleAddr
	RawRemoteAddress string
	RemotePeerId     peer.ID
}

func (self MessageAddr) MessageType() string {
	return CmdGetAddr
}

func (self MessageAddr) MaxPayloadLength(pver int) int {
	return MaxBlockPayload
}

func (self MessageAddr) JsonSerialize() ([]byte, error) {
	jsonBytes, err := json.Marshal(self)
	return jsonBytes, err
}

func (self MessageAddr) JsonDeserialize(jsonStr string) error {
	err := json.Unmarshal([]byte(jsonStr), self)
	return err
}
