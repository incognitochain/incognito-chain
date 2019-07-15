package wire

import (
	"encoding/json"
	"fmt"

	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/common"
	peer "github.com/libp2p/go-libp2p-peer"
)

const (
	MaxBFTReadyPayload = 2000 // 1 Kb
)

type MessageBFTReady struct {
	// PoolState     map[byte]uint64
	BestStateHash common.Hash
	Round         int
	Pubkey        string
	ContentSig    string
	Timestamp     int64
}

func (msg *MessageBFTReady) Hash() string {
	rawBytes, err := msg.JsonSerialize()
	if err != nil {
		return ""
	}
	return common.HashH(rawBytes).String()
}

func (msg *MessageBFTReady) MessageType() string {
	return CmdBFTReady
}

func (msg *MessageBFTReady) MaxPayloadLength(pver int) int {
	return MaxBFTReadyPayload
}

func (msg *MessageBFTReady) JsonSerialize() ([]byte, error) {
	jsonBytes, err := json.Marshal(msg)
	return jsonBytes, err
}

func (msg *MessageBFTReady) JsonDeserialize(jsonStr string) error {
	err := json.Unmarshal([]byte(jsonStr), msg)
	return err
}

func (msg *MessageBFTReady) SetSenderID(senderID peer.ID) error {
	return nil
}

func (msg *MessageBFTReady) SignMsg(keySet *incognitokey.KeySet) error {
	dataBytes := []byte{}
	// dataBytes = append(dataBytes, []byte(fmt.Sprint(msg.PoolState))...)
	dataBytes = append(dataBytes, msg.BestStateHash.GetBytes()...)
	dataBytes = append(dataBytes, []byte(fmt.Sprint(msg.Round))...)
	dataBytes = append(dataBytes, []byte(msg.Pubkey)...)
	dataBytes = append(dataBytes, []byte(fmt.Sprint(msg.Timestamp))...)
	var err error
	msg.ContentSig, err = keySet.SignDataB58(dataBytes)
	return err
}

func (msg *MessageBFTReady) VerifyMsgSanity() error {
	dataBytes := []byte{}
	// dataBytes = append(dataBytes, []byte(fmt.Sprint(msg.PoolState))...)
	dataBytes = append(dataBytes, msg.BestStateHash.GetBytes()...)
	dataBytes = append(dataBytes, []byte(fmt.Sprint(msg.Round))...)
	dataBytes = append(dataBytes, []byte(msg.Pubkey)...)
	dataBytes = append(dataBytes, []byte(fmt.Sprint(msg.Timestamp))...)
	err := incognitokey.ValidateDataB58(msg.Pubkey, msg.ContentSig, dataBytes)
	return err
}
