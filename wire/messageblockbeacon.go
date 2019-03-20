package wire

import (
	"encoding/json"

	"github.com/constant-money/constant-chain/blockchain"
	"github.com/constant-money/constant-chain/cashec"
	"github.com/constant-money/constant-chain/common"
	"github.com/libp2p/go-libp2p-peer"
)

// const (
// 	MaxBlockPayload = 1000000 // 1 Mb
// )

type MessageBlockBeacon struct {
	Block blockchain.BeaconBlock
}

func (msg *MessageBlockBeacon) Hash() string {
	rawBytes, err := msg.JsonSerialize()
	if err != nil {
		return ""
	}
	return common.HashH(rawBytes).String()
}

func (msg *MessageBlockBeacon) MessageType() string {
	return CmdBlockBeacon
}

func (msg *MessageBlockBeacon) MaxPayloadLength(pver int) int {
	return MaxBlockPayload
}

func (msg *MessageBlockBeacon) JsonSerialize() ([]byte, error) {
	jsonBytes, err := json.Marshal(msg)
	return jsonBytes, err
}

func (msg *MessageBlockBeacon) JsonDeserialize(jsonStr string) error {
	err := json.Unmarshal([]byte(jsonStr), msg)
	return err
}

func (msg *MessageBlockBeacon) SetSenderID(senderID peer.ID) error {
	return nil
}

func (msg *MessageBlockBeacon) SignMsg(_ *cashec.KeySet) error {
	return nil
}

func (msg *MessageBlockBeacon) VerifyMsgSanity() error {
	return nil
}
