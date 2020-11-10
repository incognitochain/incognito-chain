package jsonresult

import (
	"encoding/json"
	"github.com/incognitochain/incognito-chain/privacy/coin"
	"github.com/incognitochain/incognito-chain/privacy/operation"
	"log"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
)

type ListOutputCoins struct {
	FromHeight uint64 `json:"FromHeight"`
	ToHeight uint64 `json:"ToHeight"`
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
	SharedRandom		 string `json:"SharedRandom"`
	SharedConcealRandom  string `json:"SharedConcealRandom"`
	TxRandom	         string	`json:"TxRandom"`
	CoinDetailsEncrypted string `json:"CoinDetailsEncrypted"`
	AssetTag			 string `json:"AssetTag"`
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

type ICoinInfo interface {
	GetVersion() uint8
	GetCommitment() *operation.Point
	GetInfo() []byte
	GetPublicKey() *operation.Point
	GetValue() uint64
	GetKeyImage() *operation.Point
	GetRandomness() *operation.Scalar
	GetShardID() (uint8, error)
	GetSNDerivator() *operation.Scalar
	GetCoinDetailEncrypted() []byte
	IsEncrypted() bool
	GetTxRandom() *coin.TxRandom
	GetSharedRandom() *operation.Scalar
	GetSharedConcealRandom() *operation.Scalar
	GetAssetTag() *operation.Point
}


func NewOutCoin(outCoin ICoinInfo) OutCoin {
	keyImage := ""
	if outCoin.GetKeyImage() != nil && !outCoin.GetKeyImage().IsIdentity() {
		keyImage = base58.Base58Check{}.Encode(outCoin.GetKeyImage().ToBytesS(), common.ZeroByte)
	}

	publicKey := ""
	if outCoin.GetPublicKey() != nil {
		publicKey = base58.Base58Check{}.Encode(outCoin.GetPublicKey().ToBytesS(), common.ZeroByte)
	}

	commitment := ""
	if outCoin.GetCommitment() != nil {
		commitment = base58.Base58Check{}.Encode(outCoin.GetCommitment().ToBytesS(), common.ZeroByte)
	}

	snd := ""
	if outCoin.GetSNDerivator() != nil {
		snd = base58.Base58Check{}.Encode(outCoin.GetSNDerivator().ToBytesS(), common.ZeroByte)
	}

	randomness := ""
	if outCoin.GetRandomness() != nil {
		randomness = base58.Base58Check{}.Encode(outCoin.GetRandomness().ToBytesS(), common.ZeroByte)
	}

	result := OutCoin{
		Version: 		strconv.FormatUint(uint64(outCoin.GetVersion()), 10),
		PublicKey:      publicKey,
		Value:          strconv.FormatUint(outCoin.GetValue(), 10),
		Info:           EncodeBase58Check(outCoin.GetInfo()),
		Commitment: 	commitment,
		SNDerivator:    snd,
		KeyImage:   	keyImage,
		Randomness: 	randomness,
	}

	if outCoin.GetCoinDetailEncrypted() != nil {
		result.CoinDetailsEncrypted = base58.Base58Check{}.Encode(outCoin.GetCoinDetailEncrypted(), common.ZeroByte)
	}

	if outCoin.GetSharedRandom() != nil{
		result.SharedRandom = base58.Base58Check{}.Encode(outCoin.GetSharedRandom().ToBytesS(), common.ZeroByte)
	}
	if outCoin.GetSharedConcealRandom() != nil{
		result.SharedRandom = base58.Base58Check{}.Encode(outCoin.GetSharedConcealRandom().ToBytesS(), common.ZeroByte)
	}
	if outCoin.GetTxRandom() != nil{
		result.TxRandom = base58.Base58Check{}.Encode(outCoin.GetTxRandom().Bytes(), common.ZeroByte)
	}
	if outCoin.GetAssetTag() != nil{
		result.AssetTag = base58.Base58Check{}.Encode(outCoin.GetAssetTag().ToBytesS(), common.ZeroByte)
	}

	return result
}
