package wire

import (
	"encoding/hex"
	"encoding/json"

	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/libp2p/go-libp2p-peer"
)

type MessageMsgCheck struct {
	HashStr   string
	Timestamp int64
}

func (msg *MessageMsgCheck) Hash() string {
	rawBytes := make([]byte, 0)
	rawBytes = append(rawBytes, []byte(msg.MessageType())...)
	rawBytes = append(rawBytes, []byte(msg.HashStr)...)
	rawBytes = append(rawBytes, common.Int64ToBytes(msg.Timestamp)...)
	return common.HashH(rawBytes).String()
}

func (msg *MessageMsgCheck) MessageType() string {
	return CmdMsgCheck
}

func (msg *MessageMsgCheck) MaxPayloadLength(pver int) int {
	return MaxBlockPayload
}

func (msg *MessageMsgCheck) JsonSerialize() ([]byte, error) {
	jsonBytes, err := json.Marshal(msg)
	return jsonBytes, err
}

func (msg *MessageMsgCheck) JsonDeserialize(jsonStr string) error {
	jsonDecodeString, _ := hex.DecodeString(jsonStr)
	err := json.Unmarshal([]byte(jsonDecodeString), msg)
	return err
}

func (msg *MessageMsgCheck) SetSenderID(senderID peer.ID) error {
	return nil
}

func (msg *MessageMsgCheck) SignMsg(_ *incognitokey.KeySet) error {
	return nil
}

func (msg *MessageMsgCheck) VerifyMsgSanity() error {
	return nil
}
