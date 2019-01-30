package wire

import (
	"encoding/json"

	"github.com/ninjadotorg/constant/blockchain"

	"github.com/libp2p/go-libp2p-peer"
	"github.com/ninjadotorg/constant/cashec"
	"github.com/ninjadotorg/constant/common"
)

const (
	MaxShardStatePayload = 4000000 // 4 Mb
)

type MessageShardState struct {
	Timestamp int64
	ChainInfo blockchain.ShardChainState
	SenderID  string
}

func (msg *MessageShardState) Hash() string {
	rawBytes, err := msg.JsonSerialize()
	if err != nil {
		return ""
	}
	return common.HashH(rawBytes).String()
}

func (msg *MessageShardState) MessageType() string {
	return CmdShardState
}

func (msg *MessageShardState) MaxPayloadLength(pver int) int {
	return MaxShardStatePayload
}

func (msg *MessageShardState) JsonSerialize() ([]byte, error) {
	jsonBytes, err := json.Marshal(msg)
	return jsonBytes, err
}

func (msg *MessageShardState) JsonDeserialize(jsonStr string) error {
	err := json.Unmarshal([]byte(jsonStr), msg)
	return err
}

func (msg *MessageShardState) SetSenderID(senderID peer.ID) error {
	msg.SenderID = senderID.Pretty()
	return nil
}

func (msg *MessageShardState) SignMsg(_ *cashec.KeySet) error {
	return nil
}

func (msg *MessageShardState) VerifyMsgSanity() error {
	return nil
}
