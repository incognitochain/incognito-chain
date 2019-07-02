package wire

import (
	"encoding/hex"
	"encoding/json"

	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/libp2p/go-libp2p-peer"
)

const (
	MaxVersionPayload = 1000 // 1 1Kb
)

type MessageVersion struct {
	ProtocolVersion  string
	Timestamp        int64
	RemoteAddress    common.SimpleAddr
	RawRemoteAddress string
	RemotePeerId     peer.ID
	LocalAddress     common.SimpleAddr
	RawLocalAddress  string
	LocalPeerId      peer.ID
	PublicKey        string
	SignDataB58      string
}

func (msg *MessageVersion) Hash() string {
	rawBytes, err := msg.JsonSerialize()
	if err != nil {
		return ""
	}
	return common.HashH(rawBytes).String()
}

func (msg *MessageVersion) MessageType() string {
	return CmdVersion
}

func (msg *MessageVersion) MaxPayloadLength(pver int) int {
	return MaxVersionPayload
}

func (msg *MessageVersion) JsonSerialize() ([]byte, error) {
	jsonBytes, err := json.Marshal(msg)
	return jsonBytes, err
}

func (msg *MessageVersion) JsonDeserialize(jsonStr string) error {
	jsonDecodeString, _ := hex.DecodeString(jsonStr)
	err := json.Unmarshal([]byte(jsonDecodeString), msg)
	return err
}

func (msg *MessageVersion) SetSenderID(senderID peer.ID) error {
	return nil
}

func (msg *MessageVersion) SignMsg(_ *incognitokey.KeySet) error {
	return nil
}

func (msg *MessageVersion) VerifyMsgSanity() error {
	return nil
}
