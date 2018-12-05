package wire

import (
	"encoding/hex"
	"encoding/json"

	"github.com/libp2p/go-libp2p-peer"
)

type MessageMsgCheckResp struct {
	Hash   string
	Accept bool
}

func (self MessageMsgCheckResp) MessageType() string {
	return CmdMsgCheckResp
}

func (self MessageMsgCheckResp) MaxPayloadLength(pver int) int {
	return MaxBlockPayload
}

func (self MessageMsgCheckResp) JsonSerialize() ([]byte, error) {
	jsonBytes, err := json.Marshal(self)
	return jsonBytes, err
}

func (self MessageMsgCheckResp) JsonDeserialize(jsonStr string) error {
	jsonDecodeString, _ := hex.DecodeString(jsonStr)
	err := json.Unmarshal([]byte(jsonDecodeString), self)
	return err
}

func (self MessageMsgCheckResp) SetSenderID(senderID peer.ID) error {
	return nil
}
