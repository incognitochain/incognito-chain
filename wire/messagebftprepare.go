package wire

import (
	"encoding/json"

	"github.com/libp2p/go-libp2p-peer"
	"github.com/ninjadotorg/constant/cashec"
	"github.com/ninjadotorg/constant/common"
)

const (
	MaxBFTPreparePayload = 1000 // 1 Kb
)

type MessageBFTPrepare struct {
	BlkHash string
	Ri      []byte
	Pubkey  string
	MsgSig  string
}

func (msg *MessageBFTPrepare) Hash() string {
	rawBytes, err := msg.JsonSerialize()
	if err != nil {
		return ""
	}
	return common.HashH(rawBytes).String()
}

func (msg *MessageBFTPrepare) MessageType() string {
	return CmdBFTPrepare
}

func (msg *MessageBFTPrepare) MaxPayloadLength(pver int) int {
	return MaxBFTPreparePayload
}

func (msg *MessageBFTPrepare) JsonSerialize() ([]byte, error) {
	jsonBytes, err := json.Marshal(msg)
	return jsonBytes, err
}

func (msg *MessageBFTPrepare) JsonDeserialize(jsonStr string) error {
	err := json.Unmarshal([]byte(jsonStr), msg)
	return err
}

func (msg *MessageBFTPrepare) SetSenderID(senderID peer.ID) error {
	return nil
}

func (msg *MessageBFTPrepare) SignMsg(_ *cashec.KeySet) error {
	return nil
}

func (msg *MessageBFTPrepare) VerifyMsgSanity() error {
	return nil
}
