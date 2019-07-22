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
	Value          string
	Info           string
}

func (outcoin *OutCoin) Init(data interface{}) error {
	temp, err := json.Marshal(data)
	if err != nil {
		log.Print(err)
		return err
	}

	err = json.Unmarshal(temp, &outcoin)
	if err != nil {
		log.Print(err)
		return err
	}
	return nil
}
