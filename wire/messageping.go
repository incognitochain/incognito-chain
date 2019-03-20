package wire

import (
	"encoding/hex"
	"encoding/json"

	"time"

	"github.com/constant-money/constant-chain/cashec"
	"github.com/constant-money/constant-chain/common"
	"github.com/libp2p/go-libp2p-peer"
)

const (
	MaxPingPayload = 1000 // 1 1Kb
)

type MessagePing struct {
	Timestamp time.Time
}

func (msg MessagePing) MessageType() string {
	return CmdPing
}

func (msg *MessagePing) Hash() string {
	rawBytes, err := msg.JsonSerialize()
	if err != nil {
		return ""
	}
	return common.HashH(rawBytes).String()
}

func (msg *MessagePing) MaxPayloadLength(pver int) int {
	return MaxPingPayload
}

func (msg *MessagePing) JsonSerialize() ([]byte, error) {
	jsonBytes, err := json.Marshal(msg)
	return jsonBytes, err
}

func (msg *MessagePing) JsonDeserialize(jsonStr string) error {
	jsonDecodeString, _ := hex.DecodeString(jsonStr)
	err := json.Unmarshal([]byte(jsonDecodeString), msg)
	return err
}
func (msg *MessagePing) SetSenderID(senderID peer.ID) error {
	return nil
}

func (msg *MessagePing) SignMsg(_ *cashec.KeySet) error {
	return nil
}

func (msg *MessagePing) VerifyMsgSanity() error {
	return nil
}
