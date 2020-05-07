package jsonresult

import (
	"encoding/json"
	"github.com/incognitochain/incognito-chain/privacy/coin"
	"log"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
)

type ListOutputCoins struct {
	Outputs map[string][]OutCoin `json:"Outputs"`
}

type OutCoin struct {
	Version 			 string `json:"Version"`
	Index 				 string `json:"Index"`
	PublicKey            string `json:"PublicKey"`
	Commitment       	 string `json:"Commitment"`
	SNDerivator          string `json:"SNDerivator"`
	KeyImage         	 string `json:"KeyImage"`
	Randomness           string `json:"Randomness"`
	Value                string `json:"Value"`
	Info                 string `json:"Info"`
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

func NewOutCoin(outCoin coin.PlainCoin) OutCoin {
	keyImage := ""
	if outCoin.GetKeyImage() != nil && !outCoin.GetKeyImage().IsIdentity() {
		keyImage = base58.Base58Check{}.Encode(outCoin.GetKeyImage().ToBytesS(), common.ZeroByte)
	}

	result := OutCoin{
		Version: 		strconv.FormatUint(uint64(outCoin.GetVersion()), 10),
		Index: 			strconv.FormatUint(uint64(outCoin.GetIndex()), 10),
		PublicKey:      base58.Base58Check{}.Encode(outCoin.GetPublicKey().ToBytesS(), common.ZeroByte),
		Value:          strconv.FormatUint(outCoin.GetValue(), 10),
		Info:           base58.Base58Check{}.Encode(outCoin.GetInfo()[:], common.ZeroByte),
		Commitment: base58.Base58Check{}.Encode(outCoin.GetCommitment().ToBytesS(), common.ZeroByte),
		SNDerivator:    base58.Base58Check{}.Encode(outCoin.GetSNDerivator().ToBytesS(), common.ZeroByte),
		KeyImage:   keyImage,
		Randomness: base58.Base58Check{}.Encode(outCoin.GetRandomness().ToBytesS(), common.ZeroByte),
	}
	return result
}
