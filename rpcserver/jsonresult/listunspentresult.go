package jsonresult

import (
	"encoding/json"
	"log"
)

type ListOutputCoins struct {
	Outputs map[string][]OutCoin `json:"Outputs"`
}

type OutCoin struct {
	PublicKey      string
	CoinCommitment string
	SNDerivator    string
	SerialNumber   string
	Randomness     string
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
