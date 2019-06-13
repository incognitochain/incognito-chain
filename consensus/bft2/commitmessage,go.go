package bft2

import (
	"encoding/json"
	"github.com/constant-money/constant-chain/cashec"
	"github.com/constant-money/constant-chain/common"
	peer "github.com/libp2p/go-libp2p-peer"
)

func (msg *CommitMsg) Hash() string {
	rawBytes, err := msg.JsonSerialize()
	if err != nil {
		return ""
	}
	return common.HashH(rawBytes).String()
}

func (msg *CommitMsg) MessageType() string {
	return "CmdBFTReq"
}

func (msg *CommitMsg) MaxPayloadLength(pver int) int {
	return 10000
}

func (msg *CommitMsg) JsonSerialize() ([]byte, error) {
	jsonBytes, err := json.Marshal(msg)
	return jsonBytes, err
}

func (msg *CommitMsg) JsonDeserialize(jsonStr string) error {
	err := json.Unmarshal([]byte(jsonStr), msg)
	return err
}

func (msg *CommitMsg) SetSenderID(senderID peer.ID) error {
	return nil
}

func (msg *CommitMsg) SignMsg(keySet *cashec.KeySet) error {
	return nil
}

func (msg *CommitMsg) VerifyMsgSanity() error {
	return nil
}
