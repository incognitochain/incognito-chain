package transaction

type TxOut struct {
	Value     float64
	PkScript  []byte
	TxOutType string // TXOUT_COIN / TXOUT_COIN
}

func (self TxOut) NewTxOut(value float64, pkScript []byte, txOutType string) *TxOut {
	self = TxOut{
		Value:     value,
		PkScript:  pkScript,
		TxOutType: txOutType,
	}
	return &self
}
