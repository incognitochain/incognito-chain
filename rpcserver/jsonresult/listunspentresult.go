package jsonresult

import (
	"encoding/json"
	"log"
	"math/big"
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

func (outcoin *OutCoin) Init(data interface{}) error {
	temp, err := json.Marshal(data)
	if err != nil {
		log.Print(err)
		return err
	}

	type Alias OutCoin
	temp1 := &struct {
		Value string
		*Alias
	}{
		Alias: (*Alias)(outcoin),
	}
	err = json.Unmarshal(temp, temp1)
	temp2 := big.Int{}
	temp2.SetString(temp1.Value, 10)
	outcoin.Value = temp2.Uint64()
	if err != nil {
		log.Print(err)
		return err
	}
}
