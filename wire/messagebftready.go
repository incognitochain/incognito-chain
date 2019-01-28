package wire

import (
	"encoding/json"

	"github.com/libp2p/go-libp2p-peer"
	"github.com/ninjadotorg/constant/cashec"
	"github.com/ninjadotorg/constant/common"
)

const (
	MaxBFTReadyPayload = 5000 // 5 Kb
)

type MessageBFTReady struct {
	BestStateHash common.Hash
	Timestamp     int64
}

func (msg *MessageBFTReady) Hash() string {
	rawBytes, err := msg.JsonSerialize()
	if err != nil {
		return ""
	}
	return common.HashH(rawBytes).String()
}

func (msg *MessageBFTReady) MessageType() string {
	return CmdBFTReady
}

func (msg *MessageBFTReady) MaxPayloadLength(pver int) int {
	return MaxBFTReadyPayload
}

func (msg *MessageBFTReady) JsonSerialize() ([]byte, error) {
	jsonBytes, err := json.Marshal(msg)
	return jsonBytes, err
}

func (msg *MessageBFTReady) JsonDeserialize(jsonStr string) error {
	err := json.Unmarshal([]byte(jsonStr), msg)
	return err
}

func (msg *MessageBFTReady) SetSenderID(senderID peer.ID) error {
	return nil
}

func (msg *MessageBFTReady) SignMsg(_ *cashec.KeySet) error {
	return nil
}

func (msg *MessageBFTReady) VerifyMsgSanity() error {
	return nil
}
