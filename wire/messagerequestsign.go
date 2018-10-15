package wire

import (
	"encoding/json"

	"github.com/libp2p/go-libp2p-peer"
	"github.com/ninjadotorg/cash-prototype/blockchain"
)

const (
	MaxRequestBlockSign = 4000000 // 4Mb
)

type MessageRequestSign struct {
	Block    blockchain.Block
	SenderID string
}

func (self *MessageRequestSign) MessageType() string {
	return CmdRequestSign
}

func (self *MessageRequestSign) MaxPayloadLength(pver int) int {
	return MaxRequestBlockSign
}

func (self *MessageRequestSign) JsonSerialize() ([]byte, error) {
	jsonBytes, err := json.Marshal(self)
	return jsonBytes, err
}

func (self *MessageRequestSign) JsonDeserialize(jsonStr string) error {
	err := json.Unmarshal([]byte(jsonStr), self)
	return err
}

func (self *MessageRequestSign) SetSenderID(senderID peer.ID) error {
	self.SenderID = senderID.Pretty()
	return nil
}
