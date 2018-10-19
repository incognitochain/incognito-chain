package wire

import (
	"encoding/json"
	"github.com/libp2p/go-libp2p-peer"
)

const (
	MaxRequestSwapPeriodPayload = 1000 // 1 Kb
)

type MessageRequestSwapPeriod struct {
	ChainID  byte
	NodeAddr string
}

func (self MessageRequestSwapPeriod) MessageType() string {
	return CmdRequestSwapPeriod
}

func (self MessageRequestSwapPeriod) MaxPayloadLength(pver int) int {
	return MaxRequestSwapPeriodPayload
}

func (self MessageRequestSwapPeriod) JsonSerialize() ([]byte, error) {
	jsonBytes, err := json.Marshal(self)
	return jsonBytes, err
}

func (self MessageRequestSwapPeriod) JsonDeserialize(jsonStr string) error {
	err := json.Unmarshal([]byte(jsonStr), self)
	return err
}

func (self MessageRequestSwapPeriod) SetSenderID(senderID peer.ID) error {
	return nil
}
