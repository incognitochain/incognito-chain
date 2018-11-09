package jsonresult

import "github.com/ninjadotorg/constant/transaction"

type CustomToken struct {
	ID      string   `json:"ID"`
	Name    string   `json:"Name"`
	Symbol  string   `json:"Symbol"`
	Amount  uint64   `json:"Amount"`
	ListTxs []string `json:"ListTxs"`
}

func (self *CustomToken) Init(obj transaction.TxCustomToken) {
	self.ID = obj.TxTokenData.PropertyID.String()
	self.Symbol = obj.TxTokenData.PropertySymbol
	self.Name = obj.TxTokenData.PropertyName
	self.Amount = obj.TxTokenData.Amount
}

type ListCustomToken struct {
	ListCustomToken []CustomToken `json:"ListCustomToken"`
}
