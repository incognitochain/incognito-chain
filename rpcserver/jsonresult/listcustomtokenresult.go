package jsonresult

import "github.com/ninjadotorg/constant/transaction"

type CustomToken struct {
	ID        string   `json:"ID"`
	Name      string   `json:"Name"`
	Symbol    string   `json:"Symbol"`
	Amount    uint64   `json:"Amount"`
	IsPrivacy bool     `json:"IsPrivacy"`
	ListTxs   []string `json:"ListTxs"`
}

func (customToken *CustomToken) Init(obj transaction.TxCustomToken) {
	customToken.ID = obj.TxTokenData.PropertyID.String()
	customToken.Symbol = obj.TxTokenData.PropertySymbol
	customToken.Name = obj.TxTokenData.PropertyName
	customToken.Amount = obj.TxTokenData.Amount
}

func (customToken *CustomToken) InitPrivacy(obj transaction.TxCustomTokenPrivacy) {
	customToken.ID = obj.TxTokenPrivacyData.PropertyID.String()
	customToken.Symbol = obj.TxTokenPrivacyData.PropertySymbol
	customToken.Name = obj.TxTokenPrivacyData.PropertyName
	customToken.Amount = obj.TxTokenPrivacyData.Amount
	customToken.IsPrivacy = true
}

type ListCustomToken struct {
	ListCustomToken []CustomToken `json:"ListCustomToken"`
}
