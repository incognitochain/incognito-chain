package wire

import (
	"encoding/json"

	"github.com/incognitochain/incognito-chain/incognitokey"
	peer "github.com/libp2p/go-libp2p-peer"

	"time"

	"github.com/incognitochain/incognito-chain/common"
)

const (
	MaxGetAddressPayload = 100000 // 1 Kb
)

type RawPeer struct {
	RawAddress string
	PublicKey  string
}

type MessageAddr struct {
	Timestamp time.Time
	RawPeers  []RawPeer
}

func (msg *MessageAddr) Hash() string {
	rawBytes, err := msg.JsonSerialize()
	if err != nil {
		return ""
	}
	return common.HashH(rawBytes).String()
}

func (msg *MessageAddr) MessageType() string {
	return CmdAddr
}

func (msg *MessageAddr) MaxPayloadLength(pver int) int {
	return MaxGetAddressPayload
}

func (msg *MessageAddr) JsonSerialize() ([]byte, error) {
	jsonBytes, err := json.Marshal(msg)
	return jsonBytes, err
}

func (msg *MessageAddr) JsonDeserialize(jsonStr string) error {
	err := json.Unmarshal([]byte(jsonStr), msg)
	return err
}

func (msg *MessageAddr) SetSenderID(senderID peer.ID) error {
	return nil
}

func (msg *MessageAddr) SignMsg(_ *incognitokey.KeySet) error {
	return nil
}

func (msg *MessageAddr) VerifyMsgSanity() error {
	return nil
}
