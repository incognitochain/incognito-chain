package wire

import (
	"encoding/json"

	peer "github.com/libp2p/go-libp2p-peer"
	"github.com/ninjadotorg/cash-prototype/blockchain"
)

const (
	MaxHeaderPayload = 4000000
)

type MessageBlockHeader struct {
	Header blockchain.BlockHeader
}

func (self MessageBlockHeader) MessageType() string {
	return CmdBlock
}

func (self MessageBlockHeader) MaxPayloadLength(pver int) int {
	return MaxHeaderPayload
}

func (self MessageBlockHeader) JsonSerialize() ([]byte, error) {
	jsonBytes, err := json.Marshal(self)
	return jsonBytes, err
}

func (self MessageBlockHeader) JsonDeserialize(jsonStr string) error {
	err := json.Unmarshal([]byte(jsonStr), self)
	return err
}
func (self MessageBlockHeader) SetSenderID(senderID peer.ID) error {
	return nil
}
