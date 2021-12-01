package wire

import (
	"encoding/hex"
	"encoding/json"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
	peer "github.com/libp2p/go-libp2p-peer"
)

type MessageFeature struct {
	Timestamp          int
	Feature            []string
	CommitteePublicKey []string
	Signature          [][]byte
}

func NewMessageFeature(timestamp int, committeePublicKey []string, signature [][]byte, feature []string) *MessageFeature {
	return &MessageFeature{Timestamp: timestamp, CommitteePublicKey: committeePublicKey, Signature: signature, Feature: feature}
}

func (msg *MessageFeature) Hash() string {
	rawBytes, err := msg.JsonSerialize()
	if err != nil {
		return ""
	}
	return common.HashH(rawBytes).String()
}

func (msg *MessageFeature) MessageType() string {
	return CmdMsgFeatureStat
}

func (msg *MessageFeature) MaxPayloadLength(pver int) int {
	return MaxTxPayload
}

func (msg *MessageFeature) JsonSerialize() ([]byte, error) {
	jsonBytes, err := json.Marshal(msg)
	return jsonBytes, err
}

func (msg *MessageFeature) JsonDeserialize(jsonStr string) error {
	jsonDecodeString, _ := hex.DecodeString(jsonStr)
	err := json.Unmarshal([]byte(jsonDecodeString), msg)
	return err
}

func (msg *MessageFeature) SetSenderID(senderID peer.ID) error {
	return nil
}

func (msg *MessageFeature) SignMsg(_ *incognitokey.KeySet) error {
	return nil
}

func (msg *MessageFeature) VerifyMsgSanity() error {
	return nil
}
