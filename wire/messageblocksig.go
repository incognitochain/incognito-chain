package wire

import (
	"encoding/json"

	peer "github.com/libp2p/go-libp2p-peer"
)

type MessageBlockSig struct {
	BlockHash    string
	Validator    string
	ValidatorSig string
}

func (self *MessageBlockSig) MessageType() string {
	return CmdBlockSig
}

func (self *MessageBlockSig) MaxPayloadLength(pver int) int {
	return MaxBlockPayload
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
