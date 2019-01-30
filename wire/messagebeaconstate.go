package wire

import (
	"encoding/json"

	"github.com/ninjadotorg/constant/blockchain"

	"github.com/libp2p/go-libp2p-peer"
	"github.com/ninjadotorg/constant/cashec"
	"github.com/ninjadotorg/constant/common"
)

const (
	MaxBeaconStatePayload = 4000000 // 4 Mb
)

type MessageBeaconState struct {
	Timestamp int64
	ChainInfo blockchain.BeaconChainState
	SenderID  string
}

func (msg *MessageBeaconState) Hash() string {
	rawBytes, err := msg.JsonSerialize()
	if err != nil {
		return ""
	}
	return common.HashH(rawBytes).String()
}

func (msg *MessageBeaconState) MessageType() string {
	return CmdBeaconState
}

func (msg *MessageBeaconState) MaxPayloadLength(pver int) int {
	return MaxBeaconStatePayload
}

func (msg *MessageBeaconState) JsonSerialize() ([]byte, error) {
	jsonBytes, err := json.Marshal(msg)
	return jsonBytes, err
}

func (msg *MessageBeaconState) JsonDeserialize(jsonStr string) error {
	err := json.Unmarshal([]byte(jsonStr), msg)
	return err
}

func (msg *MessageBeaconState) SetSenderID(senderID peer.ID) error {
	msg.SenderID = senderID.Pretty()
	return nil
}

func (msg *MessageBeaconState) SignMsg(_ *cashec.KeySet) error {
	return nil
}

func (msg *MessageBeaconState) VerifyMsgSanity() error {
	return nil
}
