package wire

import (
	"encoding/json"

	"github.com/libp2p/go-libp2p-peer"
)

const (
	MaxBFTReplyPayload = 5000 // 5 Kb
)

type MessageBFTReply struct {
	BlockHash     string
	AggregatedSig string
	ValidatorsIdx []int
}

func (self *MessageBFTReply) MessageType() string {
	return CmdBFTReply
}

func (self *MessageBFTReply) MaxPayloadLength(pver int) int {
	return MaxBFTReplyPayload
}

func (self *MessageBFTReply) JsonSerialize() ([]byte, error) {
	jsonBytes, err := json.Marshal(self)
	return jsonBytes, err
}

func (self *MessageBFTReply) JsonDeserialize(jsonStr string) error {
	err := json.Unmarshal([]byte(jsonStr), self)
	return err
}

func (self *MessageBFTReply) SetSenderID(senderID peer.ID) error {
	return nil
}
