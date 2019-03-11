package wire

import (
	"encoding/json"

	"github.com/big0t/constant-chain/blockchain"

	"github.com/big0t/constant-chain/cashec"
	"github.com/big0t/constant-chain/common"
	"github.com/libp2p/go-libp2p-peer"
)

const (
	MaxPeerStatePayload = 4000000 // 4 Mb
)

type MessagePeerState struct {
	Beacon            blockchain.ChainState
	Shards            map[byte]blockchain.ChainState
	ShardToBeaconPool map[byte][]common.Hash
	CrossShardPool    map[byte]map[byte][]common.Hash
	Timestamp         int64
	SenderID          string
}

func (msg *MessagePeerState) Hash() string {
	rawBytes, err := msg.JsonSerialize()
	if err != nil {
		return ""
	}
	return common.HashH(rawBytes).String()
}

func (msg *MessagePeerState) MessageType() string {
	return CmdPeerState
}

func (msg *MessagePeerState) MaxPayloadLength(pver int) int {
	return MaxPeerStatePayload
}

func (msg *MessagePeerState) JsonSerialize() ([]byte, error) {
	jsonBytes, err := json.Marshal(msg)
	return jsonBytes, err
}

func (msg *MessagePeerState) JsonDeserialize(jsonStr string) error {
	err := json.Unmarshal([]byte(jsonStr), msg)
	return err
}

func (msg *MessagePeerState) SetSenderID(senderID peer.ID) error {
	msg.SenderID = senderID.Pretty()
	return nil
}

func (msg *MessagePeerState) SignMsg(_ *cashec.KeySet) error {
	return nil
}

func (msg *MessagePeerState) VerifyMsgSanity() error {
	return nil
}
