package transaction

type TxOut struct {
	Value    int
	PkScript []byte
}

func (self TxOut) NewTxOut(value int, pkScript []byte) *TxOut {
	self = TxOut{
		Value:    value,
		PkScript: pkScript,
	}
	return &self
}
