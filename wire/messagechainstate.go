package wire

import (
	"encoding/json"

	peer "github.com/libp2p/go-libp2p-peer"
)

type MessageChainState struct {
	ChainInfo interface{}
	SenderID  string
}

func (self *MessageChainState) MessageType() string {
	return CmdChainState
}

func (self *MessageChainState) MaxPayloadLength(pver int) int {
	return MaxBlockPayload
}

func (self *MessageChainState) JsonSerialize() ([]byte, error) {
	jsonBytes, err := json.Marshal(self)
	return jsonBytes, err
}

func (self *MessageChainState) JsonDeserialize(jsonStr string) error {
	err := json.Unmarshal([]byte(jsonStr), self)
	return err
}

func (self *MessageChainState) SetSenderID(senderID peer.ID) error {
	self.SenderID = senderID.Pretty()
	return nil
}
