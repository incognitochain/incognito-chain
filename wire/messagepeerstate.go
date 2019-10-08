package wire

import (
	"encoding/json"

	"github.com/incognitochain/incognito-chain/blockchain"
	peer "github.com/libp2p/go-libp2p-peer"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
)

const (
	MaxPeerStatePayload = 4000000 // 4 Mb
)

type MessagePeerState struct {
	Beacon                                                  blockchain.ChainState
	Shards                                                  map[byte]blockchain.ChainState
	ShardToBeaconPool                                       map[byte][]uint64
	CrossShardPool                                          map[byte]map[byte][]uint64
	Timestamp                                               int64
	SenderID                                                string
	SenderMiningPublicKeyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyy string
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

func (msg *MessagePeerState) SignMsg(_ *incognitokey.KeySet) error {
	return nil
}

func (msg *MessagePeerState) VerifyMsgSanity() error {
	return nil
}
