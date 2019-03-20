package wire

import (
	"encoding/json"
	"fmt"

	"github.com/constant-money/constant-chain/cashec"
	"github.com/constant-money/constant-chain/common"
	peer "github.com/libp2p/go-libp2p-peer"
)

const (
	MaxBFTReqPayload = 1000 // 1 Kb
)

type MessageBFTReq struct {
	BestStateHash  common.Hash
	ProposerOffset int
	Pubkey         string
	ContentSig     string
	Timestamp      int64
}

func (msg *MessageBFTReq) Hash() string {
	rawBytes, err := msg.JsonSerialize()
	if err != nil {
		return ""
	}
	return common.HashH(rawBytes).String()
}

func (msg *MessageBFTReq) MessageType() string {
	return CmdBFTReq
}

func (msg *MessageBFTReq) MaxPayloadLength(pver int) int {
	return MaxBFTReadyPayload
}

func (msg *MessageBFTReq) JsonSerialize() ([]byte, error) {
	jsonBytes, err := json.Marshal(msg)
	return jsonBytes, err
}

func (msg *MessageBFTReq) JsonDeserialize(jsonStr string) error {
	err := json.Unmarshal([]byte(jsonStr), msg)
	return err
}

func (msg *MessageBFTReq) SetSenderID(senderID peer.ID) error {
	return nil
}

func (msg *MessageBFTReq) SignMsg(keySet *cashec.KeySet) error {
	dataBytes := []byte{}
	dataBytes = append(dataBytes, msg.BestStateHash.GetBytes()...)
	dataBytes = append(dataBytes, []byte(fmt.Sprint(msg.ProposerOffset))...)
	dataBytes = append(dataBytes, []byte(msg.Pubkey)...)
	dataBytes = append(dataBytes, []byte(fmt.Sprint(msg.Timestamp))...)
	var err error
	msg.ContentSig, err = keySet.SignDataB58(dataBytes)
	return err
}

func (msg *MessageBFTReq) VerifyMsgSanity() error {
	dataBytes := []byte{}
	dataBytes = append(dataBytes, msg.BestStateHash.GetBytes()...)
	dataBytes = append(dataBytes, []byte(fmt.Sprint(msg.ProposerOffset))...)
	dataBytes = append(dataBytes, []byte(msg.Pubkey)...)
	dataBytes = append(dataBytes, []byte(fmt.Sprint(msg.Timestamp))...)
	err := cashec.ValidateDataB58(msg.Pubkey, msg.ContentSig, dataBytes)
	return err
}
