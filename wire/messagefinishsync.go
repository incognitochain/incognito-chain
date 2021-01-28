package wire

import (
	"encoding/hex"
	"encoding/json"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
	peer "github.com/libp2p/go-libp2p-peer"
)

type MessageFinishSync struct {
	CommitteePublicKey string
}

func (msg *MessageFinishSync) Hash() string {
	rawBytes, err := msg.JsonSerialize()
	if err != nil {
		return ""
	}
	return common.HashH(rawBytes).String()
}

func (msg *MessageFinishSync) MessageType() string {
	return CmdMsgFinishSync
}

func (msg *MessageFinishSync) MaxPayloadLength(pver int) int {
	return MaxTxPayload
}

func (msg *MessageFinishSync) JsonSerialize() ([]byte, error) {
	jsonBytes, err := json.Marshal(msg)
	return jsonBytes, err
}

func (msg *MessageFinishSync) JsonDeserialize(jsonStr string) error {
	jsonDecodeString, _ := hex.DecodeString(jsonStr)
	err := json.Unmarshal([]byte(jsonDecodeString), msg)
	return err
}

func (msg *MessageFinishSync) SetSenderID(senderID peer.ID) error {
	return nil
}

func (msg *MessageFinishSync) SignMsg(_ *incognitokey.KeySet) error {
	return nil
}

func (msg *MessageFinishSync) VerifyMsgSanity() error {
	return nil
}
