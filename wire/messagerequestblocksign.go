package wire

import (
	"encoding/json"

	"github.com/libp2p/go-libp2p-peer"
	"github.com/ninjadotorg/cash-prototype/blockchain"
)

const (
	MaxRequestBlockSign = 4000000 // 4Mb
)

type MessageRequestBlockSign struct {
	Block    blockchain.Block
	SenderID string
}

func (self *MessageRequestBlockSign) MessageType() string {
	return CmdRequestBlockSign
}

func (self *MessageRequestBlockSign) MaxPayloadLength(pver int) int {
	return MaxRequestBlockSign
}

func (self *MessageRequestBlockSign) JsonSerialize() ([]byte, error) {
	jsonBytes, err := json.Marshal(self)
	return jsonBytes, err
}

func (self *MessageRequestBlockSign) JsonDeserialize(jsonStr string) error {
	err := json.Unmarshal([]byte(jsonStr), self)
	return err
}

func (self *MessageRequestBlockSign) SetSenderID(senderID peer.ID) error {
	self.SenderID = senderID.Pretty()
	return nil
}
