package wire

import (
	"encoding/json"
	"github.com/libp2p/go-libp2p-peer"
)

const (
	MaxSignSwapPayload = 1000 // 1 Kb
)

type MessageSignSwap struct {
	ChainID   byte
	PublicKey string
	Sig       string
}

func (self MessageSignSwap) MessageType() string {
	return CmdSignSwap
}

func (self MessageSignSwap) MaxPayloadLength(pver int) int {
	return MaxSignSwapPayload
}

func (self MessageSignSwap) JsonSerialize() ([]byte, error) {
	jsonBytes, err := json.Marshal(self)
	return jsonBytes, err
}

func (self MessageSignSwap) JsonDeserialize(jsonStr string) error {
	err := json.Unmarshal([]byte(jsonStr), self)
	return err
}

func (self MessageSignSwap) SetSenderID(senderID peer.ID) error {
	return nil
}
