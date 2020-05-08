package coin

import (
	"bytes"
	"encoding/json"
	"errors"
	"math/big"
	"strconv"

	"github.com/incognitochain/incognito-chain/incognitokey"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	errhandler "github.com/incognitochain/incognito-chain/privacy/errorhandler"
	"github.com/incognitochain/incognito-chain/privacy/key"
	"github.com/incognitochain/incognito-chain/privacy/operation"
	henc "github.com/incognitochain/incognito-chain/privacy/privacy_v1/hybridencryption"
)

// Coin represents a coin
type PlainCoinV1 struct {
	publicKey    *operation.Point
	commitment   *operation.Point
	snDerivator  *operation.Scalar
	serialNumber *operation.Point
	randomness   *operation.Scalar
	value        uint64
	info         []byte //256 bytes
}

func ArrayPlainCoinToPlainCoinV1(inputCoins []PlainCoin) []*PlainCoinV1 {
	res := make([]*PlainCoinV1, len(inputCoins))
	for i := 0; i < len(inputCoins); i += 1 {
		res[i] = inputCoins[i].(*PlainCoinV1)
	}
	return res
}

func ArrayCoinToCoinV1(inputCoins []Coin) []*CoinV1 {
	res := make([]*CoinV1, len(inputCoins))
	for i := 0; i < len(inputCoins); i += 1 {
		res[i] = inputCoins[i].(*CoinV1)
	}
	return res
}

// Init (Coin) initializes a coin
func (c *PlainCoinV1) Init() *PlainCoinV1 {
	if c == nil {
		c = new(PlainCoinV1)
	}
	c.value = 0
	c.randomness = new(operation.Scalar)
	c.publicKey = new(operation.Point).Identity()
	c.serialNumber = new(operation.Point).Identity()
	c.snDerivator = new(operation.Scalar).FromUint64(0)
	c.commitment = nil
	return c
}

func (*PlainCoinV1) GetVersion() uint8 { return 1 }
func (c *PlainCoinV1) GetShardID() (uint8, error) {
	if c.publicKey == nil {
		return 255, errors.New("Cannot get ShardID because PublicKey of PlainCoin is concealed")
	}
	pubKeyBytes := c.publicKey.ToBytes()
	lastByte := pubKeyBytes[operation.Ed25519KeySize-1]
	shardID := common.GetShardIDFromLastByte(lastByte)
	return shardID, nil
}

// ver1 does not need to care for index
func (c PlainCoinV1) GetIndex() uint8                   { return 0 }
func (c PlainCoinV1) GetCommitment() *operation.Point   { return c.commitment }
func (c PlainCoinV1) GetPublicKey() *operation.Point    { return c.publicKey }
func (c PlainCoinV1) GetSNDerivator() *operation.Scalar { return c.snDerivator }
func (c PlainCoinV1) GetKeyImage() *operation.Point     { return c.serialNumber }
func (c PlainCoinV1) GetRandomness() *operation.Scalar  { return c.randomness }
func (c PlainCoinV1) GetValue() uint64                  { return c.value }
func (c PlainCoinV1) GetInfo() []byte                   { return c.info }
func (c PlainCoinV1) IsEncrypted() bool                 { return false }

func (c *PlainCoinV1) SetPublicKey(v *operation.Point)    { c.publicKey = v }
func (c *PlainCoinV1) SetCommitment(v *operation.Point)   { c.commitment = v }
func (c *PlainCoinV1) SetSNDerivator(v *operation.Scalar) { c.snDerivator = v }
func (c *PlainCoinV1) SetKeyImage(v *operation.Point)     { c.serialNumber = v }
func (c *PlainCoinV1) SetRandomness(v *operation.Scalar)  { c.randomness = v }
func (c *PlainCoinV1) SetValue(v uint64)                  { c.value = v }
func (c *PlainCoinV1) SetInfo(v []byte) {
	c.info = make([]byte, len(v))
	copy(c.info, v)
}

// Conceal data leaving serialnumber
func (c *PlainCoinV1) ConcealData(additionalData interface{}) {
	c.SetCommitment(nil)
	c.SetValue(0)
	c.SetSNDerivator(nil)
	c.SetPublicKey(nil)
	c.SetRandomness(nil)
}

//MarshalJSON (CoinV1) converts coin to bytes array,
//base58 check encode that bytes array into string
//json.Marshal the string
func (c PlainCoinV1) MarshalJSON() ([]byte, error) {
	data := c.Bytes()
	temp := base58.Base58Check{}.Encode(data, common.ZeroByte)
	return json.Marshal(temp)
}

// UnmarshalJSON (Coin) receives bytes array of coin (it was be MarshalJSON before),
// json.Unmarshal the bytes array to string
// base58 check decode that string to bytes array
// and set bytes array to coin
func (c *PlainCoinV1) UnmarshalJSON(data []byte) error {
	dataStr := ""
	_ = json.Unmarshal(data, &dataStr)
	temp, _, err := base58.Base58Check{}.Decode(dataStr)
	if err != nil {
		return err
	}
	c.SetBytes(temp)
	return nil
}

// HashH returns the SHA3-256 hashing of coin bytes array
func (c *PlainCoinV1) HashH() *common.Hash {
	hash := common.HashH(c.Bytes())
	return &hash
}

//CommitAll commits a coin with 5 attributes include:
// public key, value, serial number derivator, shardID form last byte public key, randomness
func (c *PlainCoinV1) CommitAll() error {
	shardID, err := c.GetShardID()
	if err != nil {
		return err
	}
	values := []*operation.Scalar{
		new(operation.Scalar).FromUint64(0),
		new(operation.Scalar).FromUint64(c.value),
		c.snDerivator,
		new(operation.Scalar).FromUint64(uint64(shardID)),
		c.randomness,
	}
	c.commitment, err = operation.PedCom.CommitAll(values)
	if err != nil {
		return err
	}
	c.commitment.Add(c.commitment, c.publicKey)

	return nil
}

// Bytes converts a coin's details to a bytes array
// Each fields in coin is saved in len - body format
func (c *PlainCoinV1) Bytes() []byte {
	var coinBytes []byte

	if c.publicKey != nil {
		publicKey := c.publicKey.ToBytesS()
		coinBytes = append(coinBytes, byte(operation.Ed25519KeySize))
		coinBytes = append(coinBytes, publicKey...)
	} else {
		coinBytes = append(coinBytes, byte(0))
	}

	if c.commitment != nil {
		commitment := c.commitment.ToBytesS()
		coinBytes = append(coinBytes, byte(operation.Ed25519KeySize))
		coinBytes = append(coinBytes, commitment...)
	} else {
		coinBytes = append(coinBytes, byte(0))
	}

	if c.snDerivator != nil {
		coinBytes = append(coinBytes, byte(operation.Ed25519KeySize))
		coinBytes = append(coinBytes, c.snDerivator.ToBytesS()...)
	} else {
		coinBytes = append(coinBytes, byte(0))
	}

	if c.serialNumber != nil {
		serialNumber := c.serialNumber.ToBytesS()
		coinBytes = append(coinBytes, byte(operation.Ed25519KeySize))
		coinBytes = append(coinBytes, serialNumber...)
	} else {
		coinBytes = append(coinBytes, byte(0))
	}

	if c.randomness != nil {
		coinBytes = append(coinBytes, byte(operation.Ed25519KeySize))
		coinBytes = append(coinBytes, c.randomness.ToBytesS()...)
	} else {
		coinBytes = append(coinBytes, byte(0))
	}

	if c.value > 0 {
		value := new(big.Int).SetUint64(c.value).Bytes()
		coinBytes = append(coinBytes, byte(len(value)))
		coinBytes = append(coinBytes, value...)
	} else {
		coinBytes = append(coinBytes, byte(0))
	}

	if len(c.info) > 0 {
		byteLengthInfo := byte(getMin(len(c.info), MaxSizeInfoCoin))
		coinBytes = append(coinBytes, byteLengthInfo)
		infoBytes := c.info[0:byteLengthInfo]
		coinBytes = append(coinBytes, infoBytes...)
	} else {
		coinBytes = append(coinBytes, byte(0))
	}

	return coinBytes
}

// SetBytes receives a coinBytes (in bytes array), and
// reverts coinBytes to a Coin object
func (c *PlainCoinV1) SetBytes(coinBytes []byte) error {
	if len(coinBytes) == 0 {
		return errors.New("coinBytes is empty")
	}
	var err error

	offset := 0
	c.publicKey, err = parsePointForSetBytes(&coinBytes, &offset)
	if err != nil {
		return errors.New("SetBytes CoinV1 publicKey error: " + err.Error())
	}
	c.commitment, err = parsePointForSetBytes(&coinBytes, &offset)
	if err != nil {
		return errors.New("SetBytes CoinV1 commitment error: " + err.Error())
	}
	c.snDerivator, err = parseScalarForSetBytes(&coinBytes, &offset)
	if err != nil {
		return errors.New("SetBytes CoinV1 snDerivator error: " + err.Error())
	}
	c.serialNumber, err = parsePointForSetBytes(&coinBytes, &offset)
	if err != nil {
		return errors.New("SetBytes CoinV1 serialNumber error: " + err.Error())
	}
	c.randomness, err = parseScalarForSetBytes(&coinBytes, &offset)
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
		c.value = new(big.Int).SetBytes(coinBytes[offset : offset+int(lenField)]).Uint64()
		offset += int(lenField)
	}

	c.info, err = parseInfoForSetBytes(&coinBytes, &offset)
	if err != nil {
		return errors.New("SetBytes CoinV1 info error: " + err.Error())
	}
	return nil
}

// func (c *PlainCoinV1) GetCoinValue(privateKey *key.PrivateKey) uint64 {
// 	return c.GetValue()
// }

type CoinObject struct {
	PublicKey      string `json:"PublicKey"`
	CoinCommitment string `json:"CoinCommitment"`
	SNDerivator    string `json:"SNDerivator"`
	SerialNumber   string `json:"SerialNumber"`
	Randomness     string `json:"Randomness"`
	Value          string `json:"Value"`
	Info           string `json:"Info"`
}

// SetBytes (InputCoin) receives a coinBytes (in bytes array), and
// reverts coinBytes to a InputCoin object
func (pc *PlainCoinV1) ParseCoinObjectToInputCoin(coinObj CoinObject) error {
	if pc == nil {
		pc = new(PlainCoinV1).Init()
	}
	if coinObj.PublicKey != "" {
		publicKey, _, err := base58.Base58Check{}.Decode(coinObj.PublicKey)
		if err != nil {
			return err
		}

		publicKeyPoint, err := new(operation.Point).FromBytesS(publicKey)
		if err != nil {
			return err
		}
		pc.SetPublicKey(publicKeyPoint)
	}

	if coinObj.CoinCommitment != "" {
		coinCommitment, _, err := base58.Base58Check{}.Decode(coinObj.CoinCommitment)
		if err != nil {
			return err
		}

		coinCommitmentPoint, err := new(operation.Point).FromBytesS(coinCommitment)
		if err != nil {
			return err
		}
		pc.SetCommitment(coinCommitmentPoint)
	}

	if coinObj.SNDerivator != "" {
		snderivator, _, err := base58.Base58Check{}.Decode(coinObj.SNDerivator)
		if err != nil {
			return err
		}

		snderivatorScalar := new(operation.Scalar).FromBytesS(snderivator)
		if err != nil {
			return err
		}
		pc.SetSNDerivator(snderivatorScalar)
	}

	if coinObj.SerialNumber != "" {
		serialNumber, _, err := base58.Base58Check{}.Decode(coinObj.SerialNumber)
		if err != nil {
			return err
		}

		serialNumberPoint, err := new(operation.Point).FromBytesS(serialNumber)
		if err != nil {
			return err
		}
		pc.SetKeyImage(serialNumberPoint)
	}

	if coinObj.Randomness != "" {
		randomness, _, err := base58.Base58Check{}.Decode(coinObj.Randomness)
		if err != nil {
			return err
		}

		randomnessScalar := new(operation.Scalar).FromBytesS(randomness)
		if err != nil {
			return err
		}
		pc.SetRandomness(randomnessScalar)
	}

	if coinObj.Value != "" {
		value, err := strconv.ParseUint(coinObj.Value, 10, 64)
		if err != nil {
			return err
		}
		pc.SetValue(value)
	}

	if coinObj.Info != "" {
		infoBytes, _, err := base58.Base58Check{}.Decode(coinObj.Info)
		if err != nil {
			return err
		}
		pc.SetInfo(infoBytes)
	}
	return nil
}

// OutputCoin represents a output coin of transaction
// It contains CoinDetails and CoinDetailsEncrypted (encrypted value and randomness)
// CoinDetailsEncrypted is nil when you send tx without privacy
type CoinV1 struct {
	CoinDetails          *PlainCoinV1
	CoinDetailsEncrypted *henc.HybridCipherText
}

// CoinV1 does not have index so return 0
func (c CoinV1) GetIndex() uint8                   { return 0 }
func (c CoinV1) GetVersion() uint8                 { return 1 }
func (c CoinV1) GetPublicKey() *operation.Point    { return c.CoinDetails.GetPublicKey() }
func (c CoinV1) GetCommitment() *operation.Point   { return c.CoinDetails.GetCommitment() }
func (c CoinV1) GetKeyImage() *operation.Point     { return c.CoinDetails.GetKeyImage() }
func (c CoinV1) GetRandomness() *operation.Scalar  { return c.CoinDetails.GetRandomness() }
func (c CoinV1) GetSNDerivator() *operation.Scalar { return c.CoinDetails.GetSNDerivator() }
func (c CoinV1) GetShardID() (uint8, error)                 { return c.CoinDetails.GetShardID() }
func (c CoinV1) GetValue() uint64                  { return c.CoinDetails.GetValue() }
func (c CoinV1) GetInfo() []byte                   { return c.CoinDetails.GetInfo() }
func (c CoinV1) IsEncrypted() bool                 { return c.CoinDetailsEncrypted != nil }

// Init (OutputCoin) initializes a output coin
func (c *CoinV1) Init() *CoinV1 {
	if c == nil {
		c = new(CoinV1)
	}
	c.CoinDetails = new(PlainCoinV1).Init()
	c.CoinDetailsEncrypted = new(henc.HybridCipherText)
	return c
}

// For ver1, privateKey of coin is privateKey of user
func (c PlainCoinV1) ParsePrivateKeyOfCoin(privKey key.PrivateKey) (*operation.Scalar, error) {
	return new(operation.Scalar).FromBytesS(privKey), nil
}

func (c PlainCoinV1) ParseKeyImageWithPrivateKey(privKey key.PrivateKey) (*operation.Point, error) {
	k, _ := c.ParsePrivateKeyOfCoin(privKey)
	Hp := operation.HashToPoint(c.GetPublicKey().ToBytesS())
	return new(operation.Point).ScalarMult(Hp, k), nil
}

// Bytes (OutputCoin) converts a output coin's details to a bytes array
// Each fields in coin is saved in len - body format
func (c *CoinV1) Bytes() []byte {
	var outCoinBytes []byte

	if c.CoinDetailsEncrypted != nil {
		coinDetailsEncryptedBytes := c.CoinDetailsEncrypted.Bytes()
		outCoinBytes = append(outCoinBytes, byte(len(coinDetailsEncryptedBytes)))
		outCoinBytes = append(outCoinBytes, coinDetailsEncryptedBytes...)
	} else {
		outCoinBytes = append(outCoinBytes, byte(0))
	}

	coinDetailBytes := c.CoinDetails.Bytes()

	lenCoinDetailBytes := []byte{}
	if len(coinDetailBytes) <= 255 {
		lenCoinDetailBytes = []byte{byte(len(coinDetailBytes))}
	} else {
		lenCoinDetailBytes = common.IntToBytes(len(coinDetailBytes))
	}

	outCoinBytes = append(outCoinBytes, lenCoinDetailBytes...)
	outCoinBytes = append(outCoinBytes, coinDetailBytes...)
	return outCoinBytes
}

// SetBytes (OutputCoin) receives a coinBytes (in bytes array), and
// reverts coinBytes to a OutputCoin object
func (c *CoinV1) SetBytes(bytes []byte) error {
	if len(bytes) == 0 {
		return errors.New("coinBytes is empty")
	}

	offset := 0
	lenCoinDetailEncrypted := int(bytes[0])
	offset += 1

	if lenCoinDetailEncrypted > 0 {
		if offset+lenCoinDetailEncrypted > len(bytes) {
			// out of range
			return errors.New("out of range Parse CoinDetailsEncrypted")
		}
		c.CoinDetailsEncrypted = new(henc.HybridCipherText)
		err := c.CoinDetailsEncrypted.SetBytes(bytes[offset : offset+lenCoinDetailEncrypted])
		if err != nil {
			return err
		}
		offset += lenCoinDetailEncrypted
	}

	// try get 1-byte for len
	if offset > len(bytes) {
		// out of range
		return errors.New("out of range Parse CoinDetails")
	}
	lenOutputCoin := int(bytes[offset])
	c.CoinDetails = new(PlainCoinV1)
	if lenOutputCoin != 0 {
		offset += 1
		if offset+lenOutputCoin > len(bytes) {
			// out of range
			return errors.New("out of range Parse output coin details")
		}
		err := c.CoinDetails.SetBytes(bytes[offset : offset+lenOutputCoin])
		if err != nil {
			// 1-byte is wrong
			// try get 2-byte for len
			if offset+1 > len(bytes) {
				// out of range
				return errors.New("out of range Parse output coin details")
			}
			lenOutputCoin = common.BytesToInt(bytes[offset-1 : offset+1])
			offset += 1
			if offset+lenOutputCoin > len(bytes) {
				// out of range
				return errors.New("out of range Parse output coin details")
			}
			err1 := c.CoinDetails.SetBytes(bytes[offset : offset+lenOutputCoin])
			return err1
		}
	} else {
		// 1-byte is wrong
		// try get 2-byte for len
		if offset+2 > len(bytes) {
			// out of range
			return errors.New("out of range Parse output coin details")
		}
		lenOutputCoin = common.BytesToInt(bytes[offset : offset+2])
		offset += 2
		if offset+lenOutputCoin > len(bytes) {
			// out of range
			return errors.New("out of range Parse output coin details")
		}
		err1 := c.CoinDetails.SetBytes(bytes[offset : offset+lenOutputCoin])
		return err1
	}

	return nil
}

// Encrypt returns a ciphertext encrypting for a coin using a hybrid cryptosystem,
// in which AES encryption scheme is used as a data encapsulation scheme,
// and ElGamal cryptosystem is used as a key encapsulation scheme.
func (c *CoinV1) Encrypt(recipientTK key.TransmissionKey) *errhandler.PrivacyError {
	// 32-byte first: Randomness, the rest of msg is value of coin
	msg := append(c.CoinDetails.randomness.ToBytesS(), new(big.Int).SetUint64(c.CoinDetails.value).Bytes()...)

	pubKeyPoint, err := new(operation.Point).FromBytesS(recipientTK)
	if err != nil {
		return errhandler.NewPrivacyErr(errhandler.EncryptOutputCoinErr, err)
	}

	c.CoinDetailsEncrypted, err = henc.HybridEncrypt(msg, pubKeyPoint)
	if err != nil {
		return errhandler.NewPrivacyErr(errhandler.EncryptOutputCoinErr, err)
	}

	return nil
}

func (c CoinV1) Decrypt(keySet *incognitokey.KeySet) (PlainCoin, error) {
	if keySet == nil {
		err := errors.New("Cannot decrypt coinv1 with empty key")
		return nil, errhandler.NewPrivacyErr(errhandler.DecryptOutputCoinErr, err)
	}
	if len(keySet.ReadonlyKey.Rk) == 0 && len(keySet.PrivateKey) == 0 {
		err := errors.New("Cannot Decrypt CoinV1: Keyset does not contain viewkey or privatekey")
		return nil, errhandler.NewPrivacyErr(errhandler.DecryptOutputCoinErr, err)
	}

	if bytes.Equal(c.GetPublicKey().ToBytesS(), keySet.PaymentAddress.Pk[:]) {
		result := &CoinV1{
			CoinDetails:          c.CoinDetails,
			CoinDetailsEncrypted: c.CoinDetailsEncrypted,
		}
		if result.CoinDetailsEncrypted != nil && !result.CoinDetailsEncrypted.IsNil() {
			if len(keySet.ReadonlyKey.Rk) > 0 {
				msg, err := henc.HybridDecrypt(c.CoinDetailsEncrypted, new(operation.Scalar).FromBytesS(keySet.ReadonlyKey.Rk))
				if err != nil {
					return nil, errhandler.NewPrivacyErr(errhandler.DecryptOutputCoinErr, err)
				}
				// Assign randomness and value to outputCoin details
				result.CoinDetails.randomness = new(operation.Scalar).FromBytesS(msg[0:operation.Ed25519KeySize])
				result.CoinDetails.value = new(big.Int).SetBytes(msg[operation.Ed25519KeySize:]).Uint64()
			}
		}
		if len(keySet.PrivateKey) > 0 {
			// check spent with private key
			keyImage := new(operation.Point).Derive(
				operation.PedCom.G[operation.PedersenPrivateKeyIndex],
				new(operation.Scalar).FromBytesS(keySet.PrivateKey),
				result.CoinDetails.GetSNDerivator())
			result.CoinDetails.SetKeyImage(keyImage)
		}
		return result.CoinDetails, nil
	}
	err := errors.New("coin publicKey does not equal keyset paymentAddress")
	return nil, errhandler.NewPrivacyErr(errhandler.DecryptOutputCoinErr, err)
}
