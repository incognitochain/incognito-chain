package wire

import (
	"encoding/hex"
	"encoding/json"

	"time"

	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/libp2p/go-libp2p-peer"
)

type MessageVerAck struct {
	Valid     bool
	Timestamp time.Time
}

func (msg *MessageVerAck) Hash() string {
	rawBytes, err := msg.JsonSerialize()
	if err != nil {
		return ""
	}
	return common.HashH(rawBytes).String()
}

func (msg *MessageVerAck) MessageType() string {
	return CmdVerack
}

func (msg *MessageVerAck) MaxPayloadLength(pver int) int {
	return MaxBlockPayload
}

func (msg *MessageVerAck) JsonSerialize() ([]byte, error) {
	jsonBytes, err := json.Marshal(msg)
	return jsonBytes, err
}

func (msg *MessageVerAck) JsonDeserialize(jsonStr string) error {
	jsonDecodeString, _ := hex.DecodeString(jsonStr)
	err := json.Unmarshal([]byte(jsonDecodeString), msg)
	return err
}

func (msg *MessageVerAck) SetSenderID(senderID peer.ID) error {
	return nil
}

func (msg *MessageVerAck) SignMsg(_ *incognitokey.KeySet) error {
	return nil
}

func (msg *MessageVerAck) VerifyMsgSanity() error {
	return nil
}
