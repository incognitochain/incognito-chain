package jsonrpc

import "github.com/ninjadotorg/cash-prototype/common"

// ListUnspentResult models a successful response from the listunspent request.
type ListUnspentResult struct {
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
	Amounts     []uint64 `json:"Amounts"`
	Anchor      []byte   `json:"Anchor"`
}

func (self *JoinSplitDesc) Init(data interface{}) {
	mapData := data.(map[string]interface{})
	self.Amounts = make([]uint64, 0)
	tempsAmounts := mapData["Amounts"].([]interface{})
	for _, temp := range tempsAmounts {
		self.Amounts = append(self.Amounts, uint64(temp.(float64)))
	}

	self.Anchor = common.JsonUnmarshallByteArray(mapData["Anchor"].(string))

	self.Commitments = make([][]byte, 0)
	temps := mapData["Commitments"].([]interface{})
	for _, temp := range temps {
		self.Commitments = append(self.Commitments, []byte(temp.(string)))
	}
}

// end
