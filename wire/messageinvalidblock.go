package wire

import (
	"encoding/json"

	"github.com/big0t/constant-chain/common"
	"github.com/libp2p/go-libp2p-peer"
)

const (
	MaxInvalidBlockPayload = 1000 // 1 kb
)

type MessageInvalidBlock struct {
	Reason       string //the reason it's invalid could be in
	BlockHash    string
	shardID      byte
	Validator    string
	ValidatorSig string
}

func (msg *MessageInvalidBlock) Hash() string {
	rawBytes, err := msg.JsonSerialize()
	if err != nil {
		return ""
	}
	return common.HashH(rawBytes).String()
}

func (msg *MessageInvalidBlock) MessageType() string {
	return CmdInvalidBlock
}

func (msg *MessageInvalidBlock) MaxPayloadLength(pver int) int {
	return MaxInvalidBlockPayload
}

func (msg *MessageInvalidBlock) JsonSerialize() ([]byte, error) {
	jsonBytes, err := json.Marshal(msg)
	return jsonBytes, err
}

func (msg *MessageInvalidBlock) JsonDeserialize(jsonStr string) error {
	err := json.Unmarshal([]byte(jsonStr), msg)
	return err
}

func (msg *MessageInvalidBlock) SetSenderID(senderID peer.ID) error {
	return nil
}
