package jsonresult

type UnspentCustomToken struct {
	Value          uint64 // Amount to transfer
	PaymentAddress string // payment address of receiver

	Index           int
	TxCustomTokenID string
}
