package wire
//
//import (
//	"encoding/binary"
//	"encoding/json"
//
//	"github.com/libp2p/go-libp2p-peer"
//	"github.com/ninjadotorg/constant/cashec"
//	"github.com/ninjadotorg/constant/common"
//)
//
//const (
//	MaxSwapRequestPayload = 1000 // 1 Kb
//)
//
//type MessageSwapRequest struct {
//	LockTime     int64
//	shardID      byte
//	Candidate    string
//	Requester    string
//	RequesterSig string
//	SenderID     string
//}
//
//func (self *MessageSwapRequest) Hash() string {
//	rawBytes, err := self.JsonSerialize()
//	if err != nil {
//		return ""
//	}
//	return common.HashH(rawBytes).String()
//}
//
//func (self *MessageSwapRequest) MessageType() string {
//	return CmdSwapRequest
//}
//
//func (self *MessageSwapRequest) MaxPayloadLength(pver int) int {
//	return MaxSwapRequestPayload
//}
//
//func (self *MessageSwapRequest) JsonSerialize() ([]byte, error) {
//	jsonBytes, err := json.Marshal(self)
//	return jsonBytes, err
//}
//
//func (self *MessageSwapRequest) JsonDeserialize(jsonStr string) error {
//	err := json.Unmarshal([]byte(jsonStr), self)
//	return err
//}
//
//func (self *MessageSwapRequest) SetSenderID(senderID peer.ID) error {
//	self.SenderID = senderID.Pretty()
//	return nil
//}
//
//func (self *MessageSwapRequest) GetMsgByte() []byte {
//	rawBytes := []byte{}
//	bLTime := make([]byte, 8)
//	binary.LittleEndian.PutUint64(bLTime, uint64(self.LockTime))
//	rawBytes = append(rawBytes, bLTime...)
//	rawBytes = append(rawBytes, self.shardID)
//	rawBytes = append(rawBytes, []byte(self.Candidate)...)
//	rawBytes = append(rawBytes, []byte(self.Requester)...)
//	rawBytes = append(rawBytes, []byte(self.SenderID)...)
//	return rawBytes
//}
//
//func (self *MessageSwapRequest) Verify() error {
//	msgBytes := self.GetMsgByte()
//	err := cashec.ValidateDataB58(self.Requester, self.RequesterSig, msgBytes)
//
//	if err != nil {
//		return err
//	}
//	return nil
//}
