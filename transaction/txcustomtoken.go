package transaction

import "hash"

type TxTokenVin struct {
	Vout      int
	Signature string
	PubKey    string
}

type TxTokenVout struct {
	Value       uint64
	ScripPubKey string
}

type TxToken struct {
	Version         byte
	TxType          byte
	PropertyName    string
	PropertySymbol  string
	Vin             []TxTokenVin
	Vout            []TxTokenVout
	TxCustomTokenID hash.Hash
}

type TxCustomToken struct {
	Tx

	TxToken TxToken
}
