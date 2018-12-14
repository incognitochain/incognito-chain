package jsonresult

import (
	"math/big"
	"encoding/json"
)

type ListUnspentResult struct {
	ListUnspentResultItems map[string][]ListUnspentResultItem `json:"ListUnspentResultItems"`
}

type ListUnspentResultItem struct {
	OutCoins []OutCoin `json:"OutCoins"`
}

func (self *ListUnspentResultItem) Init(data interface{}) {
	self.OutCoins = []OutCoin{}
	for _, item := range data.([]interface{}) {
		i := OutCoin{}
		i.Init(item)
		self.OutCoins = append(self.OutCoins, i)
	}
}

type OutCoin struct {
	PublicKey      string
	CoinCommitment string
	SNDerivator    big.Int
	SerialNumber   string
	Randomness     big.Int
	Value          uint64
	Info           string
}

func (self *OutCoin) Init(data interface{}) {
	temp, err := json.Marshal(data)
	if err != nil {
		return
	}
	err = json.Unmarshal(temp, self)
}
