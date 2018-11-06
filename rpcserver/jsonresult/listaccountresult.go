package jsonresult

type ListAccounts struct {
	WalletName string            `json:"WalletName"`
	Accounts   map[string]uint64 `json:"Accounts"`
}
