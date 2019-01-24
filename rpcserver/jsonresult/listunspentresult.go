package jsonresult

import (
	"encoding/json"
	"log"
	"math/big"
)

type ListUnspentResult struct {
	ListUnspentResultItems map[string][]ListUnspentResultItem `json:"ListUnspentResultItems"`
}

type ListUnspentResultItem struct {
	OutCoins []OutCoin `json:"OutCoins"`
}

func (listUnspentResultItem *ListUnspentResultItem) Init(data interface{}) {
	listUnspentResultItem.OutCoins = []OutCoin{}
	for _, item := range data.([]interface{}) {
		i := OutCoin{}
		i.Init(item)
		listUnspentResultItem.OutCoins = append(listUnspentResultItem.OutCoins, i)
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

func (outcoin *OutCoin) Init(data interface{}) {
	temp, err := json.Marshal(data)
	if err != nil {
		log.Print(err)
		return
	}
	err = json.Unmarshal(temp, outcoin)
	if err != nil {
		log.Print(err)
		return
	}
}
