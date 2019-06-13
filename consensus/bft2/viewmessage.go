package bft2

import (
	"encoding/json"
	"github.com/constant-money/constant-chain/cashec"
	"github.com/constant-money/constant-chain/common"
	peer "github.com/libp2p/go-libp2p-peer"
)

func (msg *View) Hash() string {
	rawBytes, err := msg.JsonSerialize()
	if err != nil {
		return ""
	}
	return common.HashH(rawBytes).String()
}

func (msg *View) MessageType() string {
	return "CmdBFTReq"
}

func (msg *View) MaxPayloadLength(pver int) int {
	return 10000
}

func (msg *View) JsonSerialize() ([]byte, error) {
	jsonBytes, err := json.Marshal(msg)
	return jsonBytes, err
}

func (msg *View) JsonDeserialize(jsonStr string) error {
	err := json.Unmarshal([]byte(jsonStr), msg)
	return err
}

func (msg *View) SetSenderID(senderID peer.ID) error {
	return nil
}

func (msg *View) SignMsg(keySet *cashec.KeySet) error {
	return nil
}

func (msg *View) VerifyMsgSanity() error {
	return nil
}
