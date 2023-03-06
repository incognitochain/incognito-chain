package jsonresult

import (
	"encoding/json"
	"errors"
	"log"
	"math/big"
	"strconv"

	"github.com/ethereum/go-ethereum/common/math"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/privacy/coin"
)

type ICoinInfo interface {
	GetVersion() uint8
	GetCommitment() *privacy.Point
	GetInfo() []byte
	GetPublicKey() *privacy.Point
	GetValue() uint64
	GetKeyImage() *privacy.Point
	GetRandomness() *privacy.Scalar
	GetShardID() (uint8, error)
	GetSNDerivator() *privacy.Scalar
	GetCoinDetailEncrypted() []byte
	IsEncrypted() bool
	GetTxRandom() *coin.TxRandom
	GetSharedRandom() *privacy.Scalar
	GetSharedConcealRandom() *privacy.Scalar
	GetAssetTag() *privacy.Point
	GetOTATag() *uint8
}
type ListOutputCoins struct {
	FromHeight uint64               `json:"FromHeight"`
	ToHeight   uint64               `json:"ToHeight"`
	Outputs    map[string][]OutCoin `json:"Outputs"`
}
type OutCoin struct {
	Version              string `json:"Version"`
	Index                string `json:"Index"`
	PublicKey            string `json:"PublicKey"`
	Commitment           string `json:"Commitment"`
	SNDerivator          string `json:"SNDerivator"`
	KeyImage             string `json:"KeyImage"`
	SerialNumber         string `json:"SerialNumber"`
	Randomness           string `json:"Randomness"`
	Value                string `json:"Value"`
	Info                 string `json:"Info"`
	SharedRandom         string `json:"SharedRandom"`
	SharedConcealRandom  string `json:"SharedConcealRandom"`
	TxRandom             string `json:"TxRandom"`
	CoinDetailsEncrypted string `json:"CoinDetailsEncrypted"`
	AssetTag             string `json:"AssetTag"`
	OTATag               string `json:"OTATag"`
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
		Version:      strconv.FormatUint(uint64(outCoin.GetVersion()), 10),
		PublicKey:    publicKey,
		Value:        strconv.FormatUint(outCoin.GetValue(), 10),
		Info:         EncodeBase58Check(outCoin.GetInfo()),
		Commitment:   commitment,
		SNDerivator:  snd,
		KeyImage:     keyImage,
		SerialNumber: keyImage,
		Randomness:   randomness,
	}
	if outCoin.GetCoinDetailEncrypted() != nil {
		result.CoinDetailsEncrypted = base58.Base58Check{}.Encode(outCoin.GetCoinDetailEncrypted(), common.ZeroByte)
	}
	if outCoin.GetSharedRandom() != nil {
		result.SharedRandom = base58.Base58Check{}.Encode(outCoin.GetSharedRandom().ToBytesS(), common.ZeroByte)
	}
	if outCoin.GetSharedConcealRandom() != nil {
		result.SharedRandom = base58.Base58Check{}.Encode(outCoin.GetSharedConcealRandom().ToBytesS(), common.ZeroByte)
	}
	if outCoin.GetTxRandom() != nil {
		result.TxRandom = base58.Base58Check{}.Encode(outCoin.GetTxRandom().Bytes(), common.ZeroByte)
	}
	if outCoin.GetAssetTag() != nil {
		result.AssetTag = base58.Base58Check{}.Encode(outCoin.GetAssetTag().ToBytesS(), common.ZeroByte)
	}
	if outCoin.GetOTATag() != nil {
		result.OTATag = strconv.Itoa(int(*outCoin.GetOTATag()))
	}
	return result
}
func NewCoinFromJsonOutCoin(jsonOutCoin OutCoin) (ICoinInfo, *big.Int, error) {
	var keyImage, pubkey, cm *privacy.Point
	var snd, randomness *privacy.Scalar
	var info []byte
	var err error
	var idx *big.Int
	var sharedRandom, sharedConcealRandom *privacy.Scalar
	var txRandom *coin.TxRandom
	var coinDetailEncrypted *privacy.HybridCipherText
	var assetTag *privacy.Point
	var otaTag *uint8
	value, ok := math.ParseUint64(jsonOutCoin.Value)
	if !ok {
		return nil, nil, errors.New("Cannot parse value")
	}
	if len(jsonOutCoin.KeyImage) == 0 {
		keyImage = nil
	} else {
		keyImageInBytes, _, err := base58.Base58Check{}.Decode(jsonOutCoin.KeyImage)
		if err != nil {
			return nil, nil, err
		}
		keyImage, err = new(privacy.Point).FromBytesS(keyImageInBytes)
		if err != nil {
			return nil, nil, err
		}
	}
	if len(jsonOutCoin.Commitment) == 0 {
		cm = nil
	} else {
		cmInbytes, _, err := base58.Base58Check{}.Decode(jsonOutCoin.Commitment)
		if err != nil {
			return nil, nil, err
		}
		cm, err = new(privacy.Point).FromBytesS(cmInbytes)
		if err != nil {
			return nil, nil, err
		}
	}
	if len(jsonOutCoin.PublicKey) == 0 {
		pubkey = nil
	} else {
		pubkeyInBytes, _, err := base58.Base58Check{}.Decode(jsonOutCoin.PublicKey)
		if err != nil {
			return nil, nil, err
		}
		pubkey, err = new(privacy.Point).FromBytesS(pubkeyInBytes)
		if err != nil {
			return nil, nil, err
		}
	}
	if len(jsonOutCoin.Randomness) == 0 {
		randomness = nil
	} else {
		randomnessInBytes, _, err := base58.Base58Check{}.Decode(jsonOutCoin.Randomness)
		if err != nil {
			return nil, nil, err
		}
		randomness = new(privacy.Scalar).FromBytesS(randomnessInBytes)
	}
	if len(jsonOutCoin.SNDerivator) == 0 {
		snd = nil
	} else {
		sndInBytes, _, err := base58.Base58Check{}.Decode(jsonOutCoin.SNDerivator)
		if err != nil {
			return nil, nil, err
		}
		snd = new(privacy.Scalar).FromBytesS(sndInBytes)
	}
	if len(jsonOutCoin.Info) == 0 {
		info = []byte{}
	} else {
		info, _, err = base58.Base58Check{}.Decode(jsonOutCoin.Info)
		if err != nil {
			return nil, nil, err
		}
	}
	if len(jsonOutCoin.SharedRandom) == 0 {
		sharedRandom = nil
	} else {
		sharedRandomInBytes, _, err := base58.Base58Check{}.Decode(jsonOutCoin.SharedRandom)
		if err != nil {
			return nil, nil, err
		}
		sharedRandom = new(privacy.Scalar).FromBytesS(sharedRandomInBytes)
	}
	if len(jsonOutCoin.SharedConcealRandom) == 0 {
		sharedRandom = nil
	} else {
		sharedConcealRandomInBytes, _, err := base58.Base58Check{}.Decode(jsonOutCoin.SharedConcealRandom)
		if err != nil {
			return nil, nil, err
		}
		sharedConcealRandom = new(privacy.Scalar).FromBytesS(sharedConcealRandomInBytes)
	}
	if len(jsonOutCoin.TxRandom) == 0 {
		sharedRandom = nil
	} else {
		txRandomInBytes, _, err := base58.Base58Check{}.Decode(jsonOutCoin.TxRandom)
		if err != nil {
			return nil, nil, err
		}
		txRandom = new(coin.TxRandom)
		err = txRandom.SetBytes(txRandomInBytes)
		if err != nil {
			return nil, nil, err
		}
	}
	if len(jsonOutCoin.AssetTag) == 0 {
		assetTag = nil
	} else {
		assetTagInBytes, _, err := base58.Base58Check{}.Decode(jsonOutCoin.AssetTag)
		if err != nil {
			return nil, nil, err
		}
		assetTag, err = new(privacy.Point).FromBytesS(assetTagInBytes)
		if err != nil {
			return nil, nil, err
		}
	}
	if len(jsonOutCoin.OTATag) == 0 {
		otaTag = nil
	} else {
		temp, err := strconv.Atoi(jsonOutCoin.OTATag)
		if err != nil {
			return nil, nil, err
		}
		tempUint8 := uint8(temp)
		otaTag = &tempUint8
	}
	if len(jsonOutCoin.Index) == 0 {
		idx = nil
	} else {
		idxInBytes, _, err := base58.Base58Check{}.Decode(jsonOutCoin.Index)
		if err != nil {
			return nil, nil, err
		}
		idx = new(big.Int).SetBytes(idxInBytes)
	}
	if jsonOutCoin.Version == "1" {
		pCoinV1 := new(coin.PlainCoinV1).Init()
		pCoinV1.SetRandomness(randomness)
		pCoinV1.SetPublicKey(pubkey)
		pCoinV1.SetCommitment(cm)
		pCoinV1.SetSNDerivator(snd)
		pCoinV1.SetKeyImage(keyImage)
		pCoinV1.SetInfo(info)
		pCoinV1.SetValue(value)
		if len(jsonOutCoin.CoinDetailsEncrypted) != 0 {
			coinDetailEncryptedInBytes, _, err := base58.Base58Check{}.Decode(jsonOutCoin.CoinDetailsEncrypted)
			if err != nil {
				return nil, nil, err
			}
			coinDetailEncrypted = new(privacy.HybridCipherText)
			err = coinDetailEncrypted.SetBytes(coinDetailEncryptedInBytes)
			if err != nil {
				return nil, nil, err
			}
			coinV1 := new(coin.CoinV1).Init()
			coinV1.CoinDetails = pCoinV1
			coinV1.CoinDetailsEncrypted = coinDetailEncrypted
			return coinV1, idx, nil
		}
		return pCoinV1, idx, nil
	} else if jsonOutCoin.Version == "2" || jsonOutCoin.Version == "3" {
		coinV2 := new(coin.CoinV2).Init()
		if len(jsonOutCoin.CoinDetailsEncrypted) != 0 {
			coinDetailEncryptedInBytes, _, err := base58.Base58Check{}.Decode(jsonOutCoin.CoinDetailsEncrypted)
			if err != nil {
				return nil, nil, err
			}
			amountEncrypted := new(privacy.Scalar).FromBytesS(coinDetailEncryptedInBytes)
			coinV2.SetAmount(amountEncrypted)
		} else {
			coinV2.SetValue(value)
		}
		coinV2.SetRandomness(randomness)
		coinV2.SetPublicKey(pubkey)
		coinV2.SetCommitment(cm)
		coinV2.SetKeyImage(keyImage)
		coinV2.SetInfo(info)
		coinV2.SetAssetTag(assetTag)
		if jsonOutCoin.Version == "3" {
			if otaTag == nil {
				return nil, nil, errors.New("coin missing ota tag")
			}
			coinV2.SetOTATag(*otaTag)
		}
		coinV2.SetSharedRandom(sharedRandom)
		coinV2.SetSharedConcealRandom(sharedConcealRandom)
		coinV2.SetTxRandom(txRandom)
		return coinV2, idx, nil
	}
	return nil, nil, errors.New("cannot find coin version")
}
