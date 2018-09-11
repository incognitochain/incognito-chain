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

func (self *ListUnspentResultItem) Init(data interface{}) {
	mapData := data.(map[string]interface{})
	self.TxId = mapData["TxId"].(string)
	self.JoinSplitDesc = make([]JoinSplitDesc, 0)
	temps := mapData["JoinSplitDesc"].([]interface{})
	for _, temp := range temps {
		item := JoinSplitDesc{}
		item.Init(temp)
		self.JoinSplitDesc = append(self.JoinSplitDesc, item)
	}
}

type JoinSplitDesc struct {
	Commitments [][]byte `json:"Commitments"`
	Amount      uint64   `json:"Amount"`
	Anchor      []byte   `json:"Anchor"`
}

func (self *JoinSplitDesc) Init(data interface{}) {
	mapData := data.(map[string]interface{})
	self.Amount = uint64(mapData["Amount"].(float64))
	self.Anchor = []byte(mapData["Anchor"].(string))
	self.Commitments = make([][]byte, 0)
	temps := mapData["Commitments"].([]interface{})
	for _, temp := range temps {
		self.Commitments = append(self.Commitments, []byte(temp.(string)))
	}
}

// end
