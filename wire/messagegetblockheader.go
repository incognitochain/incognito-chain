package wire

import (
	"encoding/json"

	peer "github.com/libp2p/go-libp2p-peer"
)

type MessageGetBlockHeader struct {
}

func (self MessageGetBlockHeader) MessageType() string {
	return CmdGetBlockHeader
}

func (self MessageGetBlockHeader) MaxPayloadLength(pver int) int {
	return MaxHeaderPayload
}

func (self MessageGetBlockHeader) JsonSerialize() ([]byte, error) {
	jsonBytes, err := json.Marshal(self)
	return jsonBytes, err
}

func (self MessageGetBlockHeader) JsonDeserialize(jsonStr string) error {
	err := json.Unmarshal([]byte(jsonStr), self)
	return err
}

func (self MessageGetBlockHeader) SetSenderID(senderID peer.ID) error {
	return nil
}
