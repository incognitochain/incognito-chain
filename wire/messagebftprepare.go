package wire

import (
	"encoding/json"

	"github.com/libp2p/go-libp2p-peer"
	"github.com/ninjadotorg/constant/blockchain"
)

const (
	MaxBFTPreparePayload = 1000 // 1 Kb
)

type MessageBFTPrepare struct {
	Phase string
	Block blockchain.BlockV2
}

func (self *MessageBFTPrepare) MessageType() string {
	return CmdBlockSig
}

func (self *MessageBFTPrepare) MaxPayloadLength(pver int) int {
	return MaxBlockSigPayload
}

func (self *MessageBFTPrepare) JsonSerialize() ([]byte, error) {
	jsonBytes, err := json.Marshal(self)
	return jsonBytes, err
}

func (self *MessageBFTPrepare) JsonDeserialize(jsonStr string) error {
	err := json.Unmarshal([]byte(jsonStr), self)
	return err
}

func (self *MessageBFTPrepare) SetSenderID(senderID peer.ID) error {
	return nil
}
