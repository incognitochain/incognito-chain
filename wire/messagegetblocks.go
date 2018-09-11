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

func (self MessageGetBlocks) JsonSerialize() ([]byte, error) {
	jsonBytes, err := json.Marshal(self)
	return jsonBytes, err
}

func (self MessageGetBlocks) JsonDeserialize(jsonStr string) error {
	err := json.Unmarshal([]byte(jsonStr), self)
	return err
}

func (self MessageGetBlocks) SetSenderID(senderID peer.ID) error {
	self.SenderID = senderID
	return nil
}
