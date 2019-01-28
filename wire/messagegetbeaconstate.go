package wire

import (
	"encoding/json"

	"github.com/libp2p/go-libp2p-peer"
	"github.com/ninjadotorg/constant/cashec"
	"github.com/ninjadotorg/constant/common"
)

const (
	MaxGetBeaconStatePayload = 1000 // 1kb
)

type MessageGetBeaconState struct {
	Timestamp int64
	SenderID  string
}

func (msg *MessageGetBeaconState) Hash() string {
	rawBytes, err := msg.JsonSerialize()
	if err != nil {
		return ""
	}
	return common.HashH(rawBytes).String()
}

func (msg *MessageGetBeaconState) MessageType() string {
	return CmdGetBeaconState
}

func (msg *MessageGetBeaconState) MaxPayloadLength(pver int) int {
	return MaxGetBeaconStatePayload
}

func (msg *MessageGetBeaconState) JsonSerialize() ([]byte, error) {
	jsonBytes, err := json.Marshal(msg)
	return jsonBytes, err
}

func (msg *MessageGetBeaconState) JsonDeserialize(jsonStr string) error {
	err := json.Unmarshal([]byte(jsonStr), msg)
	return err
}

func (msg *MessageGetBeaconState) SetSenderID(senderID peer.ID) error {
	msg.SenderID = senderID.Pretty()
	return nil
}

func (msg *MessageGetBeaconState) SignMsg(_ *cashec.KeySet) error {
	return nil
}

func (msg *MessageGetBeaconState) VerifyMsgSanity() error {
	return nil
}
