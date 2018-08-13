package transaction

type TxOut struct {
	Value    int64
	PkScript []byte
}

func (self TxOut) NewTxOut(value int64, pkScript []byte) *TxOut {
	self = TxOut{
		Value:    value,
		PkScript: pkScript,
	}
	return &self
}
