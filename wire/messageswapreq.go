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
//func (msg *MessageSwapRequest) Hash() string {
//	rawBytes, err := msg.JsonSerialize()
//	if err != nil {
//		return ""
//	}
//	return common.HashH(rawBytes).String()
//}
//
//func (msg *MessageSwapRequest) MessageType() string {
//	return CmdSwapRequest
//}
//
//func (msg *MessageSwapRequest) MaxPayloadLength(pver int) int {
//	return MaxSwapRequestPayload
//}
//
//func (msg *MessageSwapRequest) JsonSerialize() ([]byte, error) {
//	jsonBytes, err := json.Marshal(msg)
//	return jsonBytes, err
//}
//
//func (msg *MessageSwapRequest) JsonDeserialize(jsonStr string) error {
//	err := json.Unmarshal([]byte(jsonStr), msg)
//	return err
//}
//
//func (msg *MessageSwapRequest) SetSenderID(senderID peer.ID) error {
//	msg.SenderID = senderID.Pretty()
//	return nil
//}
//
//func (msg *MessageSwapRequest) GetMsgByte() []byte {
//	rawBytes := []byte{}
//	bLTime := make([]byte, 8)
//	binary.LittleEndian.PutUint64(bLTime, uint64(msg.LockTime))
//	rawBytes = append(rawBytes, bLTime...)
//	rawBytes = append(rawBytes, msg.shardID)
//	rawBytes = append(rawBytes, []byte(msg.Candidate)...)
//	rawBytes = append(rawBytes, []byte(msg.Requester)...)
//	rawBytes = append(rawBytes, []byte(msg.SenderID)...)
//	return rawBytes
//}
//
//func (msg *MessageSwapRequest) Verify() error {
//	msgBytes := msg.GetMsgByte()
//	err := cashec.ValidateDataB58(msg.Requester, msg.RequesterSig, msgBytes)
//
//	if err != nil {
//		return err
//	}
//	return nil
//}
