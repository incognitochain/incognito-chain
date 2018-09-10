package jsonrpc

// ListUnspentResult models a successful response from the listunspent request.
type ListUnspentResult struct {
	/*TxID          string  `json:"TxID"`
	Vout          int     `json:"Vout"`
	Address       string  `json:"Address"`
	Account       string  `json:"Account"`
	ScriptPubKey  string  `json:"ScriptPubKey"`
	RedeemScript  string  `json:"RedeemScript,omitempty"`
	Amount        float64 `json:"Amount"`
	Confirmations int64   `json:"Confirmations"`
	Spendable     bool    `json:"Spendable"`
	TxOutType     string  `json:"TxOutType"`*/
	ListUnspentResultItems map[string][]ListUnspentResultItem `json:"ListUnspentResultItems"`
}

type ListUnspentResultItem struct {
	TxId          string          `json:"TxId"`
	JoinSplitDesc []JoinSplitDesc `json:"JoinSplitDesc"`
}

type JoinSplitDesc struct {
	Commitments [][]byte `json:"Commitments"`
	Amount      uint64   `json:"Amount"`
	Anchor      []byte   `json:"Anchor"`
}

// end
