package wire

//
//import (
//	"encoding/json"
//
//	"github.com/libp2p/go-libp2p-peer"
//	"github.com/big0t/constant-chain/common"
//)
//
//const (
//	MaxSwapSigPayload = 1000 // 1 Kb
//)
//
//type MessageSwapSig struct {
//	Validator string
//	SwapSig   string
//}
//
//func (msg *MessageSwapSig) Hash() string {
//	rawBytes, err := msg.JsonSerialize()
//	if err != nil {
//		return ""
//	}
//	return common.HashH(rawBytes).String()
//}
//
//func (msg *MessageSwapSig) MessageType() string {
//	return CmdSwapSig
//}
//
//func (msg *MessageSwapSig) MaxPayloadLength(pver int) int {
//	return MaxSwapSigPayload
//}
//
//func (msg *MessageSwapSig) JsonSerialize() ([]byte, error) {
//	jsonBytes, err := json.Marshal(msg)
//	return jsonBytes, err
//}
//
//func (msg *MessageSwapSig) JsonDeserialize(jsonStr string) error {
//	err := json.Unmarshal([]byte(jsonStr), msg)
//	return err
//}
//
//func (msg *MessageSwapSig) SetSenderID(senderID peer.ID) error {
//	return nil
//}
