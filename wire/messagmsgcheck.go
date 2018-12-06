package wire

import (
	"encoding/hex"
	"encoding/json"

	"github.com/libp2p/go-libp2p-peer"
)

type MessageMsgCheck struct {
	Hash string
}

func (self MessageMsgCheck) MessageType() string {
	return CmdMsgCheck
}

func (self MessageMsgCheck) MaxPayloadLength(pver int) int {
	return MaxBlockPayload
}

func (self MessageMsgCheck) JsonSerialize() ([]byte, error) {
	jsonBytes, err := json.Marshal(self)
	return jsonBytes, err
}

func (self MessageMsgCheck) JsonDeserialize(jsonStr string) error {
	jsonDecodeString, _ := hex.DecodeString(jsonStr)
	err := json.Unmarshal([]byte(jsonDecodeString), self)
	return err
}

func (self MessageMsgCheck) SetSenderID(senderID peer.ID) error {
	return nil
}
