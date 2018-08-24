package transaction

import (
	"strconv"
	"fmt"

	"github.com/ninjadotorg/cash-prototype/common"
	//"encoding/json"
)

type Tx struct {
	Version  int     `json:"Version"`
	Type     string  `json:"Type"` // NORMAL / ACTION_PARAMS
	TxIn     []TxIn  `json:"TxIn"`
	TxOut    []TxOut `json:"TxOut"`
	LockTime int     `json:"LockTime"`
}

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
	return &hash
}

func (self *Tx) ValidateTransaction() (bool) {
	return true
}

func (self *Tx) GetType() (string) {
	return self.Type
}
