package transaction

import (
	"github.com/internet-cash/prototype/common"
	"strconv"
	"fmt"
)

type Tx struct {
	Version  int
	TxIn     []*TxIn
	TxOut    []*TxOut
	LockTime int
}

func (self Tx) AddTxIn(ti *TxIn) {
	self.TxIn = append(self.TxIn, ti)
}

func (self Tx) AddTxOut(ti *TxIn) {
	self.TxIn = append(self.TxIn, ti)
}

func (self Tx) TxHash() (common.Hash) {
	record := strconv.Itoa(self.Version) + strconv.Itoa(self.Version)
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
	return common.DoubleHashH([]byte(record))
}
