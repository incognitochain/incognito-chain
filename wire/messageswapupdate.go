package wire

//import (
//	"encoding/json"
//
//	"github.com/libp2p/go-libp2p-peer"
//	"github.com/constant-money/constant-chain/common"
//)
//
//const (
//	MaxSwapUpdatePayload = 1000 // 1 Kb
//)
//
//type MessageSwapUpdate struct {
//	LockTime   int64
//	Requester  string
//	shardID    byte
//	Candidate  string
//	Signatures map[string]string
//}
//
//func (msg *MessageSwapUpdate) Hash() string {
//	rawBytes, err := msg.JsonSerialize()
//	if err != nil {
//		return ""
//	}
//	return common.HashH(rawBytes).String()
//}
//
//func (msg *MessageSwapUpdate) MessageType() string {
//	return CmdSwapUpdate
//}
//
//func (msg *MessageSwapUpdate) MaxPayloadLength(pver int) int {
//	return MaxSwapUpdatePayload
//}
//
//func (msg *MessageSwapUpdate) JsonSerialize() ([]byte, error) {
//	jsonBytes, err := json.Marshal(msg)
//	return jsonBytes, err
//}
//
//func (msg *MessageSwapUpdate) JsonDeserialize(jsonStr string) error {
//	err := json.Unmarshal([]byte(jsonStr), msg)
//	return err
//}
//
//func (msg *MessageSwapUpdate) SetSenderID(senderID peer.ID) error {
//	return nil
//}
