package wire

import (
	"encoding/hex"
	"encoding/json"

	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/libp2p/go-libp2p-peer"
)

type MessageMsgCheckResp struct {
	HashStr   string
	Accept    bool
	Timestamp int64
}

func (msg *MessageMsgCheckResp) Hash() string {
	rawBytes := make([]byte, 0)
	rawBytes = append(rawBytes, []byte(msg.MessageType())...)
	rawBytes = append(rawBytes, []byte(msg.HashStr)...)
	rawBytes = append(rawBytes, common.BoolToByte(msg.Accept))
	rawBytes = append(rawBytes, common.Int64ToBytes(msg.Timestamp)...)
	return common.HashH(rawBytes).String()
}

func (msg *MessageMsgCheckResp) MessageType() string {
	return CmdMsgCheckResp
}

func (msg *MessageMsgCheckResp) MaxPayloadLength(pver int) int {
	return MaxBlockPayload
}

func (msg *MessageMsgCheckResp) JsonSerialize() ([]byte, error) {
	jsonBytes, err := json.Marshal(msg)
	return jsonBytes, err
}

func (msg *MessageMsgCheckResp) JsonDeserialize(jsonStr string) error {
	jsonDecodeString, _ := hex.DecodeString(jsonStr)
	err := json.Unmarshal([]byte(jsonDecodeString), msg)
	return err
}

func (msg *MessageMsgCheckResp) SetSenderID(senderID peer.ID) error {
	return nil
}

func (msg *MessageMsgCheckResp) SignMsg(_ *incognitokey.KeySet) error {
	return nil
}

func (msg *MessageMsgCheckResp) VerifyMsgSanity() error {
	return nil
}
