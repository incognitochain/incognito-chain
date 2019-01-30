package wire

import (
	"encoding/json"

	"github.com/libp2p/go-libp2p-peer"
	"github.com/ninjadotorg/constant/cashec"
	"github.com/ninjadotorg/constant/common"
)

type MessageGetShardToBeacons struct {
	From      uint64
	To        uint64
	ShardID   byte
	SenderID  string
	Timestamp int64
}

func (msg *MessageGetShardToBeacons) Hash() string {
	rawBytes, err := msg.JsonSerialize()
	if err != nil {
		return ""
	}
	return common.HashH(rawBytes).String()
}

func (msg *MessageGetShardToBeacons) MessageType() string {
	return CmdGetShardToBeacon
}

func (msg *MessageGetShardToBeacons) MaxPayloadLength(pver int) int {
	return MaxBlockPayload
}

func (msg *MessageGetShardToBeacons) JsonSerialize() ([]byte, error) {
	jsonBytes, err := json.Marshal(msg)
	return jsonBytes, err
}

func (msg *MessageGetShardToBeacons) JsonDeserialize(jsonStr string) error {
	err := json.Unmarshal([]byte(jsonStr), msg)
	return err
}

func (msg *MessageGetShardToBeacons) SetSenderID(senderID peer.ID) error {
	msg.SenderID = senderID.Pretty()
	return nil
}

func (msg *MessageGetShardToBeacons) SignMsg(_ *cashec.KeySet) error {
	return nil
}

func (msg *MessageGetShardToBeacons) VerifyMsgSanity() error {
	return nil
}
