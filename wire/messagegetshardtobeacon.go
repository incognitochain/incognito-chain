package wire

import (
	"encoding/json"

	"github.com/libp2p/go-libp2p-peer"
	"github.com/ninjadotorg/constant/cashec"
	"github.com/ninjadotorg/constant/common"
)

type MessageGetShardToBeacon struct {
	FromPool  bool
	ByHash    bool
	BlksHash  []common.Hash
	From      uint64
	To        uint64
	ShardID   byte
	SenderID  string
	Timestamp int64
}

func (msg *MessageGetShardToBeacon) Hash() string {
	rawBytes, err := msg.JsonSerialize()
	if err != nil {
		return ""
	}
	return common.HashH(rawBytes).String()
}

func (msg *MessageGetShardToBeacon) MessageType() string {
	return CmdGetShardToBeacon
}

func (msg *MessageGetShardToBeacon) MaxPayloadLength(pver int) int {
	return MaxBlockPayload
}

func (msg *MessageGetShardToBeacon) JsonSerialize() ([]byte, error) {
	jsonBytes, err := json.Marshal(msg)
	return jsonBytes, err
}

func (msg *MessageGetShardToBeacon) JsonDeserialize(jsonStr string) error {
	err := json.Unmarshal([]byte(jsonStr), msg)
	return err
}

func (msg *MessageGetShardToBeacon) SetSenderID(senderID peer.ID) error {
	msg.SenderID = senderID.Pretty()
	return nil
}

func (msg *MessageGetShardToBeacon) SignMsg(_ *cashec.KeySet) error {
	return nil
}

func (msg *MessageGetShardToBeacon) VerifyMsgSanity() error {
	return nil
}
