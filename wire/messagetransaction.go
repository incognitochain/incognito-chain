package wire

import (
	"encoding/hex"
	"encoding/json"

	"github.com/libp2p/go-libp2p-peer"
	"github.com/ninjadotorg/constant/transaction"
)

const (
	MaxTxPayload = 4000000 // 4 Mb
)

//func (self *MessageTx) UnmarshalJSON(data []byte) error {
//	tmp := make(map[string]interface{})
//	err := json.Unmarshal(data, &tmp)
//	if err != nil {
//		return err
//	}
//	if tmp["Transaction"].(map[string]interface{})["Type"] == common.TxNormalType {
//		self.Transaction = &transaction.Tx{}
//	} else if tmp["Transaction"].(map[string]interface{})["Type"] == common.TxVotingType {
//		self.Transaction = &transaction.TxVoting{}
//	}
//	err = json.Unmarshal(data, self)
//	if err != nil {
//		return err
//	}
//	return nil
//}

type MessageTx struct {
	Transaction transaction.Transaction
}

func (self MessageTx) MessageType() string {
	return CmdTx
}

func (self MessageTx) MaxPayloadLength(pver int) int {
	return MaxTxPayload
}

func (self MessageTx) JsonSerialize() ([]byte, error) {
	jsonBytes, err := json.Marshal(self)
	return jsonBytes, err
}

func (self MessageTx) JsonDeserialize(jsonStr string) error {
	jsonDecodeString, _ := hex.DecodeString(jsonStr)
	err := json.Unmarshal([]byte(jsonDecodeString), self)
	return err
}

func (self MessageTx) SetSenderID(senderID peer.ID) error {
	return nil
}
