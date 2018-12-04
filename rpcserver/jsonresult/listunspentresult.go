package jsonresult

import "math/big"

type ListUnspentResult struct {
	ListUnspentResultItems map[string]map[byte][]ListUnspentResultItem `json:"ListUnspentResultItems"`
}

type ListUnspentResultItem struct {
<<<<<<< HEAD
	TxId          string          `json:"TxIndex"`
	JoinSplitDesc []JoinSplitDesc `json:"JoinSplitDesc"`
=======
	TxId     string    `json:"TxIndex"`
	OutCoins []OutCoin `json:"JoinSplitDesc"`
>>>>>>> 11987acd2d9351ec7e456c55fe5c9ba23f2cef2d
}

/*func (self *ListUnspentResultItem) Init(data interface{}) {
	mapData := data.(map[string]interface{})
	self.TxId = mapData["TxIndex"].(string)
<<<<<<< HEAD
	self.JoinSplitDesc = make([]JoinSplitDesc, 0)
=======
	self.JoinSplitDesc = make([]OutCoin, 0)
>>>>>>> 11987acd2d9351ec7e456c55fe5c9ba23f2cef2d
	temps := mapData["JoinSplitDesc"].([]interface{})
	for _, temp := range temps {
		item := OutCoin{}
		item.Init(temp)
		self.JoinSplitDesc = append(self.JoinSplitDesc, item)
	}
}*/

type OutCoin struct {
	PublicKey      string
	CoinCommitment string
	SNDerivator    big.Int
	SerialNumber   string
	Randomness     big.Int
	Value          uint64
	Info           string
}

/*func (self *OutCoin) Init(data interface{}) {
	mapData := data.(map[string]interface{})

}*/
