package coin

import (
	"encoding/json"
	"errors"
	"math/big"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	errhandler "github.com/incognitochain/incognito-chain/privacy/errorhandler"
	"github.com/incognitochain/incognito-chain/privacy/key"
	"github.com/incognitochain/incognito-chain/privacy/operation"
	henc "github.com/incognitochain/incognito-chain/privacy/privacy_v1/hybridencryption"
)

// Coin represents a coin
type Coin_v1 struct {
	publicKey      *operation.Point
	coinCommitment *operation.Point
	snDerivator    *operation.Scalar
	serialNumber   *operation.Point
	randomness     *operation.Scalar
	value          uint64
	info           []byte //256 bytes
}

func (*Coin_v1) GetVersion() uint8                        { return 1 }
func (coin *Coin_v1) GetPublicKey() *operation.Point      { return coin.publicKey }
func (coin *Coin_v1) GetCoinCommitment() *operation.Point { return coin.coinCommitment }
func (coin *Coin_v1) GetSNDerivator() *operation.Scalar   { return coin.snDerivator }
func (coin *Coin_v1) GetSerialNumber() *operation.Point   { return coin.serialNumber }
func (coin *Coin_v1) GetRandomness() *operation.Scalar    { return coin.randomness }
func (coin *Coin_v1) GetValue() uint64                    { return coin.value }
func (coin *Coin_v1) GetInfo() []byte                     { return coin.info }

func (coin *Coin_v1) SetPublicKey(v *operation.Point)      { coin.publicKey = v }
func (coin *Coin_v1) SetCoinCommitment(v *operation.Point) { coin.coinCommitment = v }
func (coin *Coin_v1) SetSNDerivator(v *operation.Scalar)   { coin.snDerivator = v }
func (coin *Coin_v1) SetSerialNumber(v *operation.Point)   { coin.serialNumber = v }
func (coin *Coin_v1) SetRandomness(v *operation.Scalar)    { coin.randomness = v }
func (coin *Coin_v1) SetValue(v uint64)                    { coin.value = v }

func (coin *Coin_v1) SetInfo(v []byte) {
	coin.info = make([]byte, len(v))
	copy(coin.info, v)
}

// Init (Coin) initializes a coin
func (coin *Coin_v1) Init() *Coin_v1 {
	if coin == nil {
		coin = new(Coin_v1)
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
func (coin *Coin_v1) GetPubKeyLastByte() byte {
	pubKeyBytes := coin.publicKey.ToBytes()
	return pubKeyBytes[operation.Ed25519KeySize-1]
}

// MarshalJSON (Coin_v1) converts coin to bytes array,
// base58 check encode that bytes array into string
// json.Marshal the string
func (coin Coin_v1) MarshalJSON() ([]byte, error) {
	data := coin.Bytes()
	temp := base58.Base58Check{}.Encode(data, common.ZeroByte)
	return json.Marshal(temp)
}

// UnmarshalJSON (Coin) receives bytes array of coin (it was be MarshalJSON before),
// json.Unmarshal the bytes array to string
// base58 check decode that string to bytes array
// and set bytes array to coin
func (coin *Coin_v1) UnmarshalJSON(data []byte) error {
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
func (coin *Coin_v1) HashH() *common.Hash {
	hash := common.HashH(coin.Bytes())
	return &hash
}

//CommitAll commits a coin with 5 attributes include:
// public key, value, serial number derivator, shardID form last byte public key, randomness
func (coin *Coin_v1) CommitAll() error {
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
func (coin *Coin_v1) Bytes() []byte {
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
func (coin *Coin_v1) SetBytes(coinBytes []byte) error {
	if len(coinBytes) == 0 {
		return errors.New("coinBytes is empty")
	}
	var err error

	offset := 0
	coin.publicKey, err = parsePointForSetBytes(&coinBytes, &offset)
	if err != nil {
		return errors.New("SetBytes coin_v1 publicKey error: " + err.Error())
	}
	coin.coinCommitment, err = parsePointForSetBytes(&coinBytes, &offset)
	if err != nil {
		return errors.New("SetBytes coin_v1 coinCommitment error: " + err.Error())
	}
	coin.snDerivator, err = parseScalarForSetBytes(&coinBytes, &offset)
	if err != nil {
		return errors.New("SetBytes coin_v1 snDerivator error: " + err.Error())
	}
	coin.serialNumber, err = parsePointForSetBytes(&coinBytes, &offset)
	if err != nil {
		return errors.New("SetBytes coin_v1 serialNumber error: " + err.Error())
	}
	coin.randomness, err = parseScalarForSetBytes(&coinBytes, &offset)
	if err != nil {
		return errors.New("SetBytes coin_v1 serialNumber error: " + err.Error())
	}

	if offset >= len(coinBytes) {
		return errors.New("SetBytes coin_v1: out of range Parse value")
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
		return errors.New("SetBytes coin_v1 info error: " + err.Error())
	}
	return nil
}

// InputCoin represents a input coin of transaction
type InputCoin struct {
	CoinDetails *Coin_v1
}

// Init (InputCoin) initializes a input coin
func (inputCoin *InputCoin) Init() *InputCoin {
	if inputCoin.CoinDetails == nil {
		inputCoin.CoinDetails = new(Coin_v1).Init()
	}
	return inputCoin
}

// Bytes (InputCoin) converts a input coin's details to a bytes array
// Each fields in coin is saved in len - body format
func (inputCoin *InputCoin) Bytes() []byte {
	return inputCoin.CoinDetails.Bytes()
}

// SetBytes (InputCoin) receives a coinBytes (in bytes array), and
// reverts coinBytes to a InputCoin object
func (inputCoin *InputCoin) SetBytes(bytes []byte) error {
	inputCoin.CoinDetails = new(Coin_v1)
	return inputCoin.CoinDetails.SetBytes(bytes)
}

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
func (inputCoin *InputCoin) ParseCoinObjectToInputCoin(coinObj CoinObject) error {
	inputCoin.CoinDetails = new(Coin_v1).Init()

	if coinObj.PublicKey != "" {
		publicKey, _, err := base58.Base58Check{}.Decode(coinObj.PublicKey)
		if err != nil {
			return err
		}

		publicKeyPoint, err := new(operation.Point).FromBytesS(publicKey)
		if err != nil {
			return err
		}
		inputCoin.CoinDetails.SetPublicKey(publicKeyPoint)
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
		inputCoin.CoinDetails.SetCoinCommitment(coinCommitmentPoint)
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
		inputCoin.CoinDetails.SetSNDerivator(snderivatorScalar)
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
		inputCoin.CoinDetails.SetSerialNumber(serialNumberPoint)
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
		inputCoin.CoinDetails.SetRandomness(randomnessScalar)
	}

	if coinObj.Value != "" {
		value, err := strconv.ParseUint(coinObj.Value, 10, 64)
		if err != nil {
			return err
		}
		inputCoin.CoinDetails.SetValue(value)
	}

	if coinObj.Info != "" {
		infoBytes, _, err := base58.Base58Check{}.Decode(coinObj.Info)
		if err != nil {
			return err
		}
		inputCoin.CoinDetails.SetInfo(infoBytes)
	}
	return nil
}

// OutputCoin represents a output coin of transaction
// It contains CoinDetails and CoinDetailsEncrypted (encrypted value and randomness)
// CoinDetailsEncrypted is nil when you send tx without privacy
type OutputCoin struct {
	CoinDetails          *Coin_v1
	CoinDetailsEncrypted *henc.HybridCipherText
}

// Init (OutputCoin) initializes a output coin
func (outputCoin *OutputCoin) Init() *OutputCoin {
	outputCoin.CoinDetails = new(Coin_v1).Init()
	outputCoin.CoinDetailsEncrypted = new(henc.HybridCipherText)
	return outputCoin
}

// Bytes (OutputCoin) converts a output coin's details to a bytes array
// Each fields in coin is saved in len - body format
func (outputCoin *OutputCoin) Bytes() []byte {
	var outCoinBytes []byte

	if outputCoin.CoinDetailsEncrypted != nil {
		coinDetailsEncryptedBytes := outputCoin.CoinDetailsEncrypted.Bytes()
		outCoinBytes = append(outCoinBytes, byte(len(coinDetailsEncryptedBytes)))
		outCoinBytes = append(outCoinBytes, coinDetailsEncryptedBytes...)
	} else {
		outCoinBytes = append(outCoinBytes, byte(0))
	}

	coinDetailBytes := outputCoin.CoinDetails.Bytes()

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
func (outputCoin *OutputCoin) SetBytes(bytes []byte) error {
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
		outputCoin.CoinDetailsEncrypted = new(henc.HybridCipherText)
		err := outputCoin.CoinDetailsEncrypted.SetBytes(bytes[offset : offset+lenCoinDetailEncrypted])
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
	outputCoin.CoinDetails = new(Coin_v1)
	if lenOutputCoin != 0 {
		offset += 1
		if offset+lenOutputCoin > len(bytes) {
			// out of range
			return errors.New("out of range Parse output coin details")
		}
		err := outputCoin.CoinDetails.SetBytes(bytes[offset : offset+lenOutputCoin])
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
			err1 := outputCoin.CoinDetails.SetBytes(bytes[offset : offset+lenOutputCoin])
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
		err1 := outputCoin.CoinDetails.SetBytes(bytes[offset : offset+lenOutputCoin])
		return err1
	}

	return nil
}

// Encrypt returns a ciphertext encrypting for a coin using a hybrid cryptosystem,
// in which AES encryption scheme is used as a data encapsulation scheme,
// and ElGamal cryptosystem is used as a key encapsulation scheme.
func (outputCoin *OutputCoin) Encrypt(recipientTK key.TransmissionKey) *errhandler.PrivacyError {
	// 32-byte first: Randomness, the rest of msg is value of coin
	msg := append(outputCoin.CoinDetails.randomness.ToBytesS(), new(big.Int).SetUint64(outputCoin.CoinDetails.value).Bytes()...)

	pubKeyPoint, err := new(operation.Point).FromBytesS(recipientTK)
	if err != nil {
		return errhandler.NewPrivacyErr(errhandler.EncryptOutputCoinErr, err)
	}

	outputCoin.CoinDetailsEncrypted, err = henc.HybridEncrypt(msg, pubKeyPoint)
	if err != nil {
		return errhandler.NewPrivacyErr(errhandler.EncryptOutputCoinErr, err)
	}

	return nil
}

// Decrypt decrypts a ciphertext encrypting for coin with recipient's receiving key
func (outputCoin *OutputCoin) Decrypt(viewingKey key.ViewingKey) *errhandler.PrivacyError {
	msg, err := henc.HybridDecrypt(outputCoin.CoinDetailsEncrypted, new(operation.Scalar).FromBytesS(viewingKey.Rk))
	if err != nil {
		return errhandler.NewPrivacyErr(errhandler.DecryptOutputCoinErr, err)
	}

	// Assign randomness and value to outputCoin details
	outputCoin.CoinDetails.randomness = new(operation.Scalar).FromBytesS(msg[0:operation.Ed25519KeySize])
	outputCoin.CoinDetails.value = new(big.Int).SetBytes(msg[operation.Ed25519KeySize:]).Uint64()

	return nil
}
