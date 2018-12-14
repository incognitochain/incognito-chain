package wire

import (
	"encoding/json"

	"github.com/ninjadotorg/constant/privacy-protocol"

	"github.com/libp2p/go-libp2p-peer"
)

const (
	MaxBFTPreparePayload = 1000 // 1 Kb
)

type MessageBFTPrepare struct {
	Ri     privacy.EllipticPoint
	Pubkey string
	MsgSig string
}

func (self *MessageBFTPrepare) MessageType() string {
	return CmdBFTPrepare
}

func (self *MessageBFTPrepare) MaxPayloadLength(pver int) int {
	return MaxBFTPreparePayload
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
