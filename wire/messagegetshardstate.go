package wire

import (
	"encoding/json"

	"github.com/libp2p/go-libp2p-peer"
	"github.com/ninjadotorg/constant/cashec"
	"github.com/ninjadotorg/constant/common"
)

const (
	MaxGetShardStatePayload = 1000 // 1kb
)

type MessageGetShardState struct {
	ShardID   byte
	Timestamp int64
	SenderID  string
}

func (msg *MessageGetShardState) Hash() string {
	rawBytes, err := msg.JsonSerialize()
	if err != nil {
		return ""
	}
	return common.HashH(rawBytes).String()
}

func (msg *MessageGetShardState) MessageType() string {
	return CmdGetShardState
}

func (msg *MessageGetShardState) MaxPayloadLength(pver int) int {
	return MaxGetShardStatePayload
}

func (msg *MessageGetShardState) JsonSerialize() ([]byte, error) {
	jsonBytes, err := json.Marshal(msg)
	return jsonBytes, err
}

func (msg *MessageGetShardState) JsonDeserialize(jsonStr string) error {
	err := json.Unmarshal([]byte(jsonStr), msg)
	return err
}

func (msg *MessageGetShardState) SetSenderID(senderID peer.ID) error {
	msg.SenderID = senderID.Pretty()
	return nil
}

func (msg *MessageGetShardState) SignMsg(_ *cashec.KeySet) error {
	return nil
}

func (msg *MessageGetShardState) VerifyMsgSanity() error {
	return nil
}
