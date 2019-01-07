package wire

import (
	"encoding/json"

	"github.com/libp2p/go-libp2p-peer"
	"github.com/ninjadotorg/constant/common"
)

const (
	MaxSwapSigPayload = 1000 // 1 Kb
)

type MessageSwapSig struct {
	Validator string
	SwapSig   string
}

func (self *MessageSwapSig) Hash() string {
	rawBytes, err := self.JsonSerialize()
	if err != nil {
		return ""
	}
	return common.HashH(rawBytes).String()
}

func (self *MessageSwapSig) MessageType() string {
	return CmdSwapSig
}

func (self *MessageSwapSig) MaxPayloadLength(pver int) int {
	return MaxSwapSigPayload
}

func (self *MessageSwapSig) JsonSerialize() ([]byte, error) {
	jsonBytes, err := json.Marshal(self)
	return jsonBytes, err
}

func (self *MessageSwapSig) JsonDeserialize(jsonStr string) error {
	err := json.Unmarshal([]byte(jsonStr), self)
	return err
}

func (self *MessageSwapSig) SetSenderID(senderID peer.ID) error {
	return nil
}
