package jsonrpc

type ListAccounts struct {
	Accounts map[string]uint64 `json:"Accounts"`
}
