package wire

import (
	"encoding/json"

	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/common"
	peer "github.com/libp2p/go-libp2p-peer"
)

type MessageGetBlockShard struct {
	FromPool         bool
	ByHash           bool
	BySpecificHeight bool
	BlkHashes        []common.Hash
	BlkHeights       []uint64
	ShardID          byte
	SenderID         string
	Timestamp        int64
}

func (msg *MessageGetBlockShard) Hash() string {
	rawBytes, err := msg.JsonSerialize()
	if err != nil {
		return ""
	}
	return common.HashH(rawBytes).String()
}

func (msg *MessageGetBlockShard) MessageType() string {
	return CmdGetBlockShard
}

func (msg *MessageGetBlockShard) MaxPayloadLength(pver int) int {
	return MaxBlockPayload
}

func (msg *MessageGetBlockShard) JsonSerialize() ([]byte, error) {
	jsonBytes, err := json.Marshal(msg)
	return jsonBytes, err
}

func (msg *MessageGetBlockShard) JsonDeserialize(jsonStr string) error {
	err := json.Unmarshal([]byte(jsonStr), msg)
	return err
}

func (msg *MessageGetBlockShard) SetSenderID(senderID peer.ID) error {
	msg.SenderID = senderID.Pretty()
	return nil
}

func (msg *MessageGetBlockShard) SignMsg(_ *incognitokey.KeySet) error {
	return nil
}

func (msg *MessageGetBlockShard) VerifyMsgSanity() error {
	return nil
}
