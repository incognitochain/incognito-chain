package wire

//import (
//	"encoding/json"
//
//	"github.com/libp2p/go-libp2p-peer"
//	"github.com/ninjadotorg/constant/common"
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
//func (self *MessageSwapUpdate) Hash() string {
//	rawBytes, err := self.JsonSerialize()
//	if err != nil {
//		return ""
//	}
//	return common.HashH(rawBytes).String()
//}
//
//func (self *MessageSwapUpdate) MessageType() string {
//	return CmdSwapUpdate
//}
//
//func (self *MessageSwapUpdate) MaxPayloadLength(pver int) int {
//	return MaxSwapUpdatePayload
//}
//
//func (self *MessageSwapUpdate) JsonSerialize() ([]byte, error) {
//	jsonBytes, err := json.Marshal(self)
//	return jsonBytes, err
//}
//
//func (self *MessageSwapUpdate) JsonDeserialize(jsonStr string) error {
//	err := json.Unmarshal([]byte(jsonStr), self)
//	return err
//}
//
//func (self *MessageSwapUpdate) SetSenderID(senderID peer.ID) error {
//	return nil
//}
