package wire

import (
	"encoding/json"

	peer "github.com/libp2p/go-libp2p-peer"
	"github.com/ninjadotorg/cash-prototype/common"
)

type MessageGetBlocks struct {
	LastBlockHash common.Hash
	SenderID      peer.ID
}

func (self MessageGetBlocks) MessageType() string {
	return CmdGetBlocks
}

func (self MessageGetBlocks) MaxPayloadLength(pver int) int {
	return MaxBlockPayload
}

func (self MessageGetBlocks) JsonSerialize() (string, error) {
	jsonStr, err := json.Marshal(self)
	header := make([]byte, MessageHeaderSize)
	copy(header[:], self.MessageType())
	jsonStr = append(jsonStr, header...)
	return string(jsonStr), err
}

func (self MessageGetBlocks) JsonDeserialize(jsonStr string) error {
	err := json.Unmarshal([]byte(jsonStr), self)
	return err
}
