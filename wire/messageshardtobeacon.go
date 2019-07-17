package wire

import (
	"encoding/json"

	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/libp2p/go-libp2p-peer"
)

// const (
// 	MaxBlockPayload = 1000000 // 1 Mb
// )

type MessageShardToBeacon struct {
	Block *blockchain.ShardToBeaconBlock
}

func (msg *MessageShardToBeacon) Hash() string {
	rawBytes, err := msg.JsonSerialize()
	if err != nil {
		return ""
	}
	return common.HashH(rawBytes).String()
}

func (msg *MessageShardToBeacon) MessageType() string {
	return CmdBlkShardToBeacon
}

func (msg *MessageShardToBeacon) MaxPayloadLength(pver int) int {
	return MaxBlockPayload
}

func (msg *MessageShardToBeacon) JsonSerialize() ([]byte, error) {
	jsonBytes, err := json.Marshal(msg)
	return jsonBytes, err
}

func (msg *MessageShardToBeacon) JsonDeserialize(jsonStr string) error {
	err := json.Unmarshal([]byte(jsonStr), msg)
	return err
}

func (msg *MessageShardToBeacon) SetSenderID(senderID peer.ID) error {
	return nil
}

func (msg *MessageShardToBeacon) SignMsg(_ *incognitokey.KeySet) error {
	return nil
}

func (msg *MessageShardToBeacon) VerifyMsgSanity() error {
	return nil
}
