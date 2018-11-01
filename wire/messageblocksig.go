package wire

import (
	"encoding/json"

	"github.com/libp2p/go-libp2p-peer"
)

const (
	MaxBlockSigPayload = 1000 // 1 Kb
)

type MessageBlockSig struct {
	Validator string
	BlockSig  string
}

func (self *MessageBlockSig) MessageType() string {
	return CmdBlockSig
}

func (self *MessageBlockSig) MaxPayloadLength(pver int) int {
	return MaxBlockSigPayload
}

func (self *MessageBlockSig) JsonSerialize() ([]byte, error) {
	jsonBytes, err := json.Marshal(self)
	return jsonBytes, err
}

func (self *MessageBlockSig) JsonDeserialize(jsonStr string) error {
	err := json.Unmarshal([]byte(jsonStr), self)
	return err
}

func (self *MessageBlockSig) SetSenderID(senderID peer.ID) error {
	return nil
}
