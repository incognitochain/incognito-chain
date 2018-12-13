package wire

import (
	"encoding/json"

	"github.com/libp2p/go-libp2p-peer"
	"github.com/ninjadotorg/constant/blockchain"
)

const (
	MaxBFTReplyPayload = 1000 // 1 Kb
)

type MessageBFTReply struct {
	Phase string
	Block blockchain.BlockV2
}

func (self *MessageBFTReply) MessageType() string {
	return CmdBlockSig
}

func (self *MessageBFTReply) MaxPayloadLength(pver int) int {
	return MaxBlockSigPayload
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
