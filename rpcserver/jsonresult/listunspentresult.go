package jsonresult

import (
	"encoding/json"
	"log"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/privacy"
)

type ListOutputCoins struct {
	Outputs map[string][]OutCoin `json:"Outputs"`
}

type OutCoin struct {
	PublicKey      string `json:"PublicKey"`
	CoinCommitment string `json:"CoinCommitment"`
	SNDerivator    string `json:"SNDerivator"`
	SerialNumber   string `json:"SerialNumber"`
	Randomness     string `json:"Randomness"`
	Value          string `json:"Value"`
	Info           string `json:"Info"`
}

func NewOutcoinFromInterface(data interface{}) (*OutCoin, error) {
	outcoin := OutCoin{}
	temp, err := json.Marshal(data)
	if err != nil {
		log.Print(err)
		return nil, err
	}

	err = json.Unmarshal(temp, &outcoin)
	if err != nil {
		log.Print(err)
		return nil, err
	}
	return &outcoin, nil
}

func NewOutCoin(outCoin *privacy.OutputCoin) OutCoin {
	serialNumber := ""

	if outCoin.CoinDetails.GetSerialNumber() != nil && !outCoin.CoinDetails.GetSerialNumber().IsIdentity() {
		serialNumber = base58.Base58Check{}.Encode(outCoin.CoinDetails.GetSerialNumber().ToBytesS(), common.ZeroByte)
	}

	result := OutCoin{
		PublicKey:      base58.Base58Check{}.Encode(outCoin.CoinDetails.GetPublicKey().ToBytesS(), common.ZeroByte),
		Value:          strconv.FormatUint(outCoin.CoinDetails.GetValue(), 10),
		Info:           base58.Base58Check{}.Encode(outCoin.CoinDetails.GetInfo()[:], common.ZeroByte),
		CoinCommitment: base58.Base58Check{}.Encode(outCoin.CoinDetails.GetCoinCommitment().ToBytesS(), common.ZeroByte),
		Randomness:     base58.Base58Check{}.Encode(outCoin.CoinDetails.GetRandomness().ToBytesS(), common.ZeroByte),
		SNDerivator:    base58.Base58Check{}.Encode(outCoin.CoinDetails.GetSNDerivator().ToBytesS(), common.ZeroByte),
		SerialNumber:   serialNumber,
	}

	return result
}
