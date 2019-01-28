package wire

import (
	"encoding/json"

	"github.com/libp2p/go-libp2p-peer"
	"github.com/ninjadotorg/constant/cashec"
	"github.com/ninjadotorg/constant/common"
)

const (
	MaxBFTProposePayload = 1000 // 1 Kb
)

type MessageBFTPropose struct {
	Block  json.RawMessage
	MsgSig string
}

func (msg *MessageBFTPropose) Hash() string {
	rawBytes, err := msg.JsonSerialize()
	if err != nil {
		return ""
	}
	return common.HashH(rawBytes).String()
}

func (msg *MessageBFTPropose) MessageType() string {
	return CmdBFTPropose
}

func (msg *MessageBFTPropose) MaxPayloadLength(pver int) int {
	return MaxBFTProposePayload
}

func (msg *MessageBFTPropose) JsonSerialize() ([]byte, error) {
	jsonBytes, err := json.Marshal(msg)
	return jsonBytes, err
}

func (msg *MessageBFTPropose) JsonDeserialize(jsonStr string) error {
	err := json.Unmarshal([]byte(jsonStr), msg)
	return err
}

func (msg *MessageBFTPropose) SetSenderID(senderID peer.ID) error {
	return nil
}

func (msg *MessageBFTPropose) SignMsg(_ *cashec.KeySet) error {
	return nil
}

func (msg *MessageBFTPropose) VerifyMsgSanity() error {
	return nil
}
