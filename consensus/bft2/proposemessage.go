package bft2

import (
	"encoding/json"
	"github.com/constant-money/constant-chain/cashec"
	"github.com/constant-money/constant-chain/common"
	peer "github.com/libp2p/go-libp2p-peer"
)

func (msg *ProposeMsg) Hash() string {
	rawBytes, err := msg.JsonSerialize()
	if err != nil {
		return ""
	}
	return common.HashH(rawBytes).String()
}

func (msg *ProposeMsg) MessageType() string {
	return "CmdBFTReq"
}

func (msg *ProposeMsg) MaxPayloadLength(pver int) int {
	return 10000
}

func (msg *ProposeMsg) JsonSerialize() ([]byte, error) {
	jsonBytes, err := json.Marshal(msg)
	return jsonBytes, err
}

func (msg *ProposeMsg) JsonDeserialize(jsonStr string) error {
	err := json.Unmarshal([]byte(jsonStr), msg)
	return err
}

func (msg *ProposeMsg) SetSenderID(senderID peer.ID) error {
	return nil
}

func (msg *ProposeMsg) SignMsg(keySet *cashec.KeySet) error {
	return nil
}

func (msg *ProposeMsg) VerifyMsgSanity() error {
	return nil
}
