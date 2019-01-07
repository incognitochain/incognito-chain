package wire

import (
	"encoding/hex"
	"encoding/json"

	"github.com/libp2p/go-libp2p-peer"
	"github.com/ninjadotorg/constant/cashec"
	"time"
	"github.com/ninjadotorg/constant/common"
)

type MessageMsgCheckResp struct {
	HashStr   string
	Accept    bool
	Timestamp time.Time
}

func (self *MessageMsgCheckResp) Hash() string {
	rawBytes := make([]byte, 0)
	rawBytes = append(rawBytes, []byte(self.MessageType())...)
	rawBytes = append(rawBytes, []byte(self.HashStr)...)
	rawBytes = append(rawBytes, common.BoolToByte(self.Accept))
	rawBytes = append(rawBytes, common.Int64ToBytes(self.Timestamp.UnixNano())...)
	return common.HashH(rawBytes).String()
}

func (self *MessageMsgCheckResp) MessageType() string {
	return CmdMsgCheckResp
}

func (self *MessageMsgCheckResp) MaxPayloadLength(pver int) int {
	return MaxBlockPayload
}

func (self *MessageMsgCheckResp) JsonSerialize() ([]byte, error) {
	jsonBytes, err := json.Marshal(self)
	return jsonBytes, err
}

func (self *MessageMsgCheckResp) JsonDeserialize(jsonStr string) error {
	jsonDecodeString, _ := hex.DecodeString(jsonStr)
	err := json.Unmarshal([]byte(jsonDecodeString), self)
	return err
}

func (self *MessageMsgCheckResp) SetSenderID(senderID peer.ID) error {
	return nil
}

func (self *MessageMsgCheckResp) SignMsg(_ *cashec.KeySet) error {
	return nil
}

func (self *MessageMsgCheckResp) VerifyMsgSanity() error {
	return nil
}
