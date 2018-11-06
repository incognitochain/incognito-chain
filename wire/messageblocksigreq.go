package wire

import (
	"encoding/json"

	"github.com/libp2p/go-libp2p-peer"
	"github.com/ninjadotorg/constant/blockchain"
)

const (
	MaxBlockSigReq = 4000000 // 4Mb
)

type MessageBlockSigReq struct {
	Block    blockchain.Block
	SenderID string
}

func (self *MessageBlockSigReq) MessageType() string {
	return CmdBlockSigReq
}

func (self *MessageBlockSigReq) MaxPayloadLength(pver int) int {
	return MaxBlockSigReq
}

func (self *MessageBlockSigReq) JsonSerialize() ([]byte, error) {
	jsonBytes, err := json.Marshal(self)
	return jsonBytes, err
}

func (self *MessageBlockSigReq) JsonDeserialize(jsonStr string) error {
	err := json.Unmarshal([]byte(jsonStr), self)
	return err
}

func (self *MessageBlockSigReq) SetSenderID(senderID peer.ID) error {
	self.SenderID = senderID.Pretty()
	return nil
}
