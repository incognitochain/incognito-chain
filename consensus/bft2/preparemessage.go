package bft2

import (
	"encoding/json"
	"github.com/constant-money/constant-chain/cashec"
	"github.com/constant-money/constant-chain/common"
	peer "github.com/libp2p/go-libp2p-peer"
)

func (msg *PrepareMsg) Hash() string {
	rawBytes, err := msg.JsonSerialize()
	if err != nil {
		return ""
	}
	return common.HashH(rawBytes).String()
}

func (msg *PrepareMsg) MessageType() string {
	return "CmdBFTReq"
}

func (msg *PrepareMsg) MaxPayloadLength(pver int) int {
	return 10000
}

func (msg *PrepareMsg) JsonSerialize() ([]byte, error) {
	jsonBytes, err := json.Marshal(msg)
	return jsonBytes, err
}

func (msg *PrepareMsg) JsonDeserialize(jsonStr string) error {
	err := json.Unmarshal([]byte(jsonStr), msg)
	return err
}

func (msg *PrepareMsg) SetSenderID(senderID peer.ID) error {
	return nil
}

func (msg *PrepareMsg) SignMsg(keySet *cashec.KeySet) error {
	return nil
}

func (msg *PrepareMsg) VerifyMsgSanity() error {
	return nil
}
