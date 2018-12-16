package wire

import (
	"encoding/json"

	"github.com/ninjadotorg/constant/cashec"

	"github.com/libp2p/go-libp2p-peer"
)

const (
	MaxGetAddressPayload = 1000 // 1 Kb
)

type RawPeer struct {
	RawAddress string
	PublicKey  string
}

type MessageAddr struct {
	RawPeers []RawPeer
}

func (self MessageAddr) MessageType() string {
	return CmdAddr
}

func (self MessageAddr) MaxPayloadLength(pver int) int {
	return MaxGetAddressPayload
}

func (self MessageAddr) JsonSerialize() ([]byte, error) {
	jsonBytes, err := json.Marshal(self)
	return jsonBytes, err
}

func (self MessageAddr) JsonDeserialize(jsonStr string) error {
	err := json.Unmarshal([]byte(jsonStr), self)
	return err
}

func (self MessageAddr) SetSenderID(senderID peer.ID) error {
	return nil
}

func (self MessageAddr) SetIntendedReceiver(_ string) error {
	return nil
}

func (self MessageAddr) SignMsg(_ *cashec.KeySet) error {
	return nil
}

func (self MessageAddr) VerifyMsgSanity() error {
	return nil
}
