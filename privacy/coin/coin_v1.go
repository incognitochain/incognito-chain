package coin

import (
	"encoding/json"
	"errors"
	"math/big"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/privacy/operation"
)

// Coin represents a coin
type CoinV1 struct {
	publicKey      *operation.Point
	coinCommitment *operation.Point
	snDerivator    *operation.Scalar
	serialNumber   *operation.Point
	randomness     *operation.Scalar
	value          uint64
	info           []byte //256 bytes
}

func (*CoinV1) GetVersion() uint8                       { return 1 }
func (coin CoinV1) GetPublicKey() *operation.Point      { return coin.publicKey }
func (coin CoinV1) GetCoinCommitment() *operation.Point { return coin.coinCommitment }
func (coin CoinV1) GetSNDerivator() *operation.Scalar   { return coin.snDerivator }
func (coin CoinV1) GetSerialNumber() *operation.Point   { return coin.serialNumber }
func (coin CoinV1) GetRandomness() *operation.Scalar    { return coin.randomness }
func (coin CoinV1) GetValue() uint64                    { return coin.value }
func (coin CoinV1) GetInfo() []byte                     { return coin.info }

func (coin *CoinV1) SetPublicKey(v *operation.Point)      { coin.publicKey = v }
func (coin *CoinV1) SetCoinCommitment(v *operation.Point) { coin.coinCommitment = v }
func (coin *CoinV1) SetSNDerivator(v *operation.Scalar)   { coin.snDerivator = v }
func (coin *CoinV1) SetSerialNumber(v *operation.Point)   { coin.serialNumber = v }
func (coin *CoinV1) SetRandomness(v *operation.Scalar)    { coin.randomness = v }
func (coin *CoinV1) SetValue(v uint64)                    { coin.value = v }

func (coin *CoinV1) SetInfo(v []byte) {
	coin.info = make([]byte, len(v))
	copy(coin.info, v)
}

// Init (Coin) initializes a coin
func (coin *CoinV1) Init() *CoinV1 {
	if coin == nil {
		coin = new(CoinV1)
	}
	coin.value = 0
	coin.randomness = new(operation.Scalar)
	coin.publicKey = new(operation.Point).Identity()
	coin.serialNumber = new(operation.Point).Identity()
	coin.coinCommitment = new(operation.Point).Identity()
	coin.snDerivator = new(operation.Scalar).FromUint64(0)
	return coin
}

// GetPubKeyLastByte returns the last byte of public key
func (coin *CoinV1) GetPubKeyLastByte() byte {
	pubKeyBytes := coin.publicKey.ToBytes()
	return pubKeyBytes[operation.Ed25519KeySize-1]
}

// MarshalJSON (CoinV1) converts coin to bytes array,
// base58 check encode that bytes array into string
// json.Marshal the string
func (coin CoinV1) MarshalJSON() ([]byte, error) {
	data := coin.Bytes()
	temp := base58.Base58Check{}.Encode(data, common.ZeroByte)
	return json.Marshal(temp)
}

// UnmarshalJSON (Coin) receives bytes array of coin (it was be MarshalJSON before),
// json.Unmarshal the bytes array to string
// base58 check decode that string to bytes array
// and set bytes array to coin
func (coin *CoinV1) UnmarshalJSON(data []byte) error {
	dataStr := ""
	_ = json.Unmarshal(data, &dataStr)
	temp, _, err := base58.Base58Check{}.Decode(dataStr)
	if err != nil {
		return err
	}
	coin.SetBytes(temp)
	return nil
}

// HashH returns the SHA3-256 hashing of coin bytes array
func (coin *CoinV1) HashH() *common.Hash {
	hash := common.HashH(coin.Bytes())
	return &hash
}

//CommitAll commits a coin with 5 attributes include:
// public key, value, serial number derivator, shardID form last byte public key, randomness
func (coin *CoinV1) CommitAll() error {
	shardID := common.GetShardIDFromLastByte(coin.GetPubKeyLastByte())
	values := []*operation.Scalar{new(operation.Scalar).FromUint64(0), new(operation.Scalar).FromUint64(coin.value), coin.snDerivator, new(operation.Scalar).FromUint64(uint64(shardID)), coin.randomness}
	commitment, err := operation.PedCom.CommitAll(values)
	if err != nil {
		return err
	}
	coin.coinCommitment = commitment
	coin.coinCommitment.Add(coin.coinCommitment, coin.publicKey)

	return nil
}

// Bytes converts a coin's details to a bytes array
// Each fields in coin is saved in len - body format
func (coin *CoinV1) Bytes() []byte {
	var coinBytes []byte

	if coin.publicKey != nil {
		publicKey := coin.publicKey.ToBytesS()
		coinBytes = append(coinBytes, byte(operation.Ed25519KeySize))
		coinBytes = append(coinBytes, publicKey...)
	} else {
		coinBytes = append(coinBytes, byte(0))
	}

	if coin.coinCommitment != nil {
		coinCommitment := coin.coinCommitment.ToBytesS()
		coinBytes = append(coinBytes, byte(operation.Ed25519KeySize))
		coinBytes = append(coinBytes, coinCommitment...)
	} else {
		coinBytes = append(coinBytes, byte(0))
	}

	if coin.snDerivator != nil {
		coinBytes = append(coinBytes, byte(operation.Ed25519KeySize))
		coinBytes = append(coinBytes, coin.snDerivator.ToBytesS()...)
	} else {
		coinBytes = append(coinBytes, byte(0))
	}

	if coin.serialNumber != nil {
		serialNumber := coin.serialNumber.ToBytesS()
		coinBytes = append(coinBytes, byte(operation.Ed25519KeySize))
		coinBytes = append(coinBytes, serialNumber...)
	} else {
		coinBytes = append(coinBytes, byte(0))
	}

	if coin.randomness != nil {
		coinBytes = append(coinBytes, byte(operation.Ed25519KeySize))
		coinBytes = append(coinBytes, coin.randomness.ToBytesS()...)
	} else {
		coinBytes = append(coinBytes, byte(0))
	}

	if coin.value > 0 {
		value := new(big.Int).SetUint64(coin.value).Bytes()
		coinBytes = append(coinBytes, byte(len(value)))
		coinBytes = append(coinBytes, value...)
	} else {
		coinBytes = append(coinBytes, byte(0))
	}

	if len(coin.info) > 0 {
		byteLengthInfo := byte(getMin(len(coin.info), MaxSizeInfoCoin))
		coinBytes = append(coinBytes, byteLengthInfo)
		infoBytes := coin.info[0:byteLengthInfo]
		coinBytes = append(coinBytes, infoBytes...)
	} else {
		coinBytes = append(coinBytes, byte(0))
	}

	return coinBytes
}

// SetBytes receives a coinBytes (in bytes array), and
// reverts coinBytes to a Coin object
func (coin *CoinV1) SetBytes(coinBytes []byte) error {
	if len(coinBytes) == 0 {
		return errors.New("coinBytes is empty")
	}
	var err error

	offset := 0
	coin.publicKey, err = parsePointForSetBytes(&coinBytes, &offset)
	if err != nil {
		return errors.New("SetBytes CoinV1 publicKey error: " + err.Error())
	}
	coin.coinCommitment, err = parsePointForSetBytes(&coinBytes, &offset)
	if err != nil {
		return errors.New("SetBytes CoinV1 coinCommitment error: " + err.Error())
	}
	coin.snDerivator, err = parseScalarForSetBytes(&coinBytes, &offset)
	if err != nil {
		return errors.New("SetBytes CoinV1 snDerivator error: " + err.Error())
	}
	coin.serialNumber, err = parsePointForSetBytes(&coinBytes, &offset)
	if err != nil {
		return errors.New("SetBytes CoinV1 serialNumber error: " + err.Error())
	}
	coin.randomness, err = parseScalarForSetBytes(&coinBytes, &offset)
	if err != nil {
		return errors.New("SetBytes CoinV1 serialNumber error: " + err.Error())
	}

	if offset >= len(coinBytes) {
		return errors.New("SetBytes CoinV1: out of range Parse value")
	}
	lenField := coinBytes[offset]
	offset++
	if lenField != 0 {
		if offset+int(lenField) > len(coinBytes) {
			// out of range
			return errors.New("out of range Parse PublicKey")
		}
		coin.value = new(big.Int).SetBytes(coinBytes[offset : offset+int(lenField)]).Uint64()
		offset += int(lenField)
	}

	coin.info, err = parseInfoForSetBytes(&coinBytes, &offset)
	if err != nil {
		return errors.New("SetBytes CoinV1 info error: " + err.Error())
	}
	return nil
}

// func (coin *CoinV1) GetCoinValue(privateKey *key.PrivateKey) uint64 {
// 	return coin.GetValue()
// }
