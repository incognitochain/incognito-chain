package wire

import (
	"encoding/json"
	"github.com/libp2p/go-libp2p-peer"
)

const (
	MaxUpdateSwapPayload = 1000 // 1 Kb
)

type MessageUpdateSwap struct {
	LockTime     int64
	RequesterPbk string
	ChainID      byte
	SealerPbk    string
	Signatures   map[string]string
}

func (self MessageUpdateSwap) MessageType() string {
	return CmdUpdateSwap
}

func (self MessageUpdateSwap) MaxPayloadLength(pver int) int {
	return MaxUpdateSwapPayload
}

func (self MessageUpdateSwap) JsonSerialize() ([]byte, error) {
	jsonBytes, err := json.Marshal(self)
	return jsonBytes, err
}

func (self MessageUpdateSwap) JsonDeserialize(jsonStr string) error {
	err := json.Unmarshal([]byte(jsonStr), self)
	return err
}

func (self MessageUpdateSwap) SetSenderID(senderID peer.ID) error {
	return nil
}
