package transaction

type TxOut struct {
	Value    float64
	PkScript []byte
}

func (self TxOut) NewTxOut(value float64, pkScript []byte) *TxOut {
	self = TxOut{
		Value:    value,
		PkScript: pkScript,
	}
	return &self
}
