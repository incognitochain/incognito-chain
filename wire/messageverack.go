package wire

import (
	"encoding/hex"
	"encoding/json"

	"time"

	"github.com/libp2p/go-libp2p-peer"
	"github.com/ninjadotorg/constant/cashec"
)

type MessageVerAck struct {
	Valid     bool
	Timestamp time.Time
}

func (self MessageVerAck) MessageType() string {
	return CmdVerack
}

func (self MessageVerAck) MaxPayloadLength(pver int) int {
	return MaxBlockPayload
}

func (self MessageVerAck) JsonSerialize() ([]byte, error) {
	jsonBytes, err := json.Marshal(self)
	return jsonBytes, err
}

func (self MessageVerAck) JsonDeserialize(jsonStr string) error {
	jsonDecodeString, _ := hex.DecodeString(jsonStr)
	err := json.Unmarshal([]byte(jsonDecodeString), self)
	return err
}

func (self MessageVerAck) SetSenderID(senderID peer.ID) error {
	return nil
}

func (self *MessageVerAck) SignMsg(_ *cashec.KeySet) error {
	return nil
}

func (self *MessageVerAck) VerifyMsgSanity() error {
	return nil
}
