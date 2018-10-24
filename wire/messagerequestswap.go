package wire

import (
	"encoding/json"
	"github.com/libp2p/go-libp2p-peer"
)

const (
	MaxRequestSwapPayload = 1000 // 1 Kb
)

type MessageRequestSwap struct {
	SenderID     string
	RequesterPbk string
	ChainID      byte
	SealerPbk    string
}

func (self MessageRequestSwap) MessageType() string {
	return CmdRequestSwap
}

func (self MessageRequestSwap) MaxPayloadLength(pver int) int {
	return MaxRequestSwapPayload
}

func (self MessageRequestSwap) JsonSerialize() ([]byte, error) {
	jsonBytes, err := json.Marshal(self)
	return jsonBytes, err
}

func (self MessageRequestSwap) JsonDeserialize(jsonStr string) error {
	err := json.Unmarshal([]byte(jsonStr), self)
	return err
}

func (self *MessageRequestSwap) SetSenderID(senderID peer.ID) error {
	self.SenderID = senderID.Pretty()
	return nil
}
