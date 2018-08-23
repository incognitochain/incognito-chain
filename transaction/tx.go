package transaction

import (
	"strconv"
	"fmt"

	"github.com/ninjadotorg/cash-prototype/common"
	//"encoding/json"
)

type Tx struct {
	Version  int
	Type     string // NORMAL / ACTION_PARAMS
	TxIn     []TxIn
	TxOut    []TxOut
	LockTime int
	txHash   *common.Hash
}

//func (self *Tx) MarshalJSON() ([]byte, error) {
//	m := make(map[string]interface{})
//	m["Version"] = self.Version
//	m["Type"] = self.Type
//	m["TxIn"] = self.TxIn
//	m["TxOut"] = self.TxOut
//	m["LockTime"] = self.LockTime
//	return json.Marshal(m)
//}
//
//func (self *Tx) UnmarshalJSON(data []byte) (error) {
//	m := make(map[string]interface{})
//	_ = json.Unmarshal(data, &m)
//
//	temp, ok := m["Type"]
//	if ok {
//		self.Type = temp.(string)
//	}
//	self.Version = int(m["Version"].(float64))
//	for _, v := range m["TxIn"].([]interface{}) {
//		kk := v.(map[string]interface{})
//		hash, _ := common.Hash{}.NewHashFromStr(kk["PreviousOutPoint"].(map[string]interface{})["Hash"].(string))
//		outp := OutPoint{
//			Hash: *hash,
//			Vout: int(kk["PreviousOutPoint"].(map[string]interface{})["Vout"].(float64)),
//		}
//		a := TxIn{
//			Sequence:         int(kk["Sequence"].(float64)),
//			SignatureScript:  []byte(kk["SignatureScript"].(string)),
//			PreviousOutPoint: outp,
//		}
//		self.TxIn = append(self.TxIn, a)
//	}
//	for _, v := range m["TxOut"].([]interface{}) {
//		ww := v.(map[string]interface{})
//		a := TxOut{
//			Value:    ww["Value"].(float64),
//			PkScript: []byte(ww["PkScript"].(string)),
//		}
//		txOutTemp, ok := ww["TxOutType"]
//		if ok {
//			a.TxOutType = txOutTemp.(string)
//		}
//		self.TxOut = append(self.TxOut, a)
//	}
//	self.LockTime = int(m["LockTime"].(float64))
//	return nil
//}

func (self *Tx) AddTxIn(ti TxIn) {
	self.TxIn = append(self.TxIn, ti)
}

func (self *Tx) GetTxIn() []TxIn {
	return self.TxIn
}

func (self *Tx) AddTxOut(to TxOut) {
	self.TxOut = append(self.TxOut, to)
}

func (self *Tx) GetTxOUt() []TxOut {
	return self.TxOut
}

func (self *Tx) Hash() (*common.Hash) {
	if self.txHash != nil {
		return self.txHash
	}
	record := strconv.Itoa(self.Version) + strconv.Itoa(self.Version)
	record += self.Type
	for _, txin := range self.TxIn {
		record += fmt.Sprint(txin.Sequence)
		record += string(txin.SignatureScript)
		record += fmt.Sprint(txin.PreviousOutPoint.Vout)
		record += fmt.Sprint(txin.PreviousOutPoint.Hash.String())
	}
	for _, txout := range self.TxOut {
		record += fmt.Sprint(txout.Value)
		record += string(txout.PkScript)
	}
	hash := common.DoubleHashH([]byte(record))
	self.txHash = &hash
	return self.txHash
}

func (self *Tx) ValidateTransaction() (bool) {
	return true
}

func (self *Tx) GetType() (string) {
	return self.Type
}
