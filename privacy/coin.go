package privacy

import (
	"encoding/json"
	"errors"
	"math/big"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
)

// Coin represents a coin
type Coin struct {
	publicKey      *Point
	coinCommitment *Point
	snDerivator    *Scalar
	serialNumber   *Point
	randomness     *Scalar
	value          uint64
	info           []byte //256 bytes
}

// Start GET/SET
func (coin Coin) GetPublicKey() *Point {
	return coin.publicKey
}

func (coin *Coin) SetPublicKey(v *Point) {
	coin.publicKey = v
}

func (coin Coin) GetCoinCommitment() *Point {
	return coin.coinCommitment
}

func (coin *Coin) SetCoinCommitment(v *Point) {
	coin.coinCommitment = v
}

func (coin Coin) GetSNDerivator() *Scalar {
	return coin.snDerivator
}

func (coin *Coin) SetSNDerivator(v *Scalar) {
	coin.snDerivator = v
}

func (coin Coin) GetSerialNumber() *Point {
	return coin.serialNumber
}

func (coin *Coin) SetSerialNumber(v *Point) {
	coin.serialNumber = v
}

func (coin Coin) GetRandomness() *Scalar {
	return coin.randomness
}

func (coin *Coin) SetRandomness(v *Scalar) {
	coin.randomness = v
}

func (coin Coin) GetValue() uint64 {
	return coin.value
}

func (coin *Coin) SetValue(v uint64) {
	coin.value = v
}

func (coin Coin) GetInfo() []byte {
	return coin.info
}

func (coin *Coin) SetInfo(v []byte) {
	coin.info = make([]byte, len(v))
	copy(coin.info, v)
}

// Init (Coin) initializes a coin
func (coin *Coin) Init() *Coin {
	coin.publicKey = new(Point).Identity()

	coin.coinCommitment = new(Point).Identity()

	coin.snDerivator = new(Scalar).FromUint64(0)

	coin.serialNumber = new(Point).Identity()

	coin.randomness = new(Scalar)

	coin.value = 0

	return coin
}

// GetPubKeyLastByte returns the last byte of public key
func (coin *Coin) GetPubKeyLastByte() byte {
	pubKeyBytes := coin.publicKey.ToBytes()
	return pubKeyBytes[Ed25519KeySize-1]
}

// MarshalJSON (Coin) converts coin to bytes array,
// base58 check encode that bytes array into string
// json.Marshal the string
func (coin Coin) MarshalJSON() ([]byte, error) {
	data := coin.Bytes()
	temp := base58.Base58Check{}.Encode(data, common.ZeroByte)
	return json.Marshal(temp)
}

// UnmarshalJSON (Coin) receives bytes array of coin (it was be MarshalJSON before),
// json.Unmarshal the bytes array to string
// base58 check decode that string to bytes array
// and set bytes array to coin
func (coin *Coin) UnmarshalJSON(data []byte) error {
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
func (coin *Coin) HashH() *common.Hash {
	hash := common.HashH(coin.Bytes())
	return &hash
}

//CommitAll commits a coin with 5 attributes include:
// public key, value, serial number derivator, shardID form last byte public key, randomness
func (coin *Coin) CommitAll() error {
	shardID := common.GetShardIDFromLastByte(coin.GetPubKeyLastByte())
	values := []*Scalar{new(Scalar).FromUint64(0), new(Scalar).FromUint64(coin.value), coin.snDerivator, new(Scalar).FromUint64(uint64(shardID)), coin.randomness}
	commitment, err := PedCom.commitAll(values)
	if err != nil {
		return err
	}
	coin.coinCommitment = commitment
	coin.coinCommitment.Add(coin.coinCommitment, coin.publicKey)

	return nil
}

// Bytes converts a coin's details to a bytes array
// Each fields in coin is saved in len - body format
func (coin *Coin) Bytes() []byte {
	var coinBytes []byte

	if coin.publicKey != nil {
		publicKey := coin.publicKey.ToBytesS()
		coinBytes = append(coinBytes, byte(Ed25519KeySize))
		coinBytes = append(coinBytes, publicKey...)
	} else {
		coinBytes = append(coinBytes, byte(0))
	}

	if coin.coinCommitment != nil {
		coinCommitment := coin.coinCommitment.ToBytesS()
		coinBytes = append(coinBytes, byte(Ed25519KeySize))
		coinBytes = append(coinBytes, coinCommitment...)
	} else {
		coinBytes = append(coinBytes, byte(0))
	}

	if coin.snDerivator != nil {
		coinBytes = append(coinBytes, byte(Ed25519KeySize))
		coinBytes = append(coinBytes, coin.snDerivator.ToBytesS()...)
	} else {
		coinBytes = append(coinBytes, byte(0))
	}

	if coin.serialNumber != nil {
		serialNumber := coin.serialNumber.ToBytesS()
		coinBytes = append(coinBytes, byte(Ed25519KeySize))
		coinBytes = append(coinBytes, serialNumber...)
	} else {
		coinBytes = append(coinBytes, byte(0))
	}

	if coin.randomness != nil {
		coinBytes = append(coinBytes, byte(Ed25519KeySize))
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
		byteLengthInfo := byte(0)
		if len(coin.info) > MaxSizeInfoCoin {
			// only get 255 byte of info
			byteLengthInfo = byte(MaxSizeInfoCoin)
		} else {
			lengthInfo := len(coin.info)
			byteLengthInfo = byte(lengthInfo)
		}
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
func (coin *Coin) SetBytes(coinBytes []byte) error {
	if len(coinBytes) == 0 {
		return errors.New("coinBytes is empty")
	}

	offset := 0
	var err error

	// Parse PublicKey
	lenField := coinBytes[offset]
	offset++
	if lenField != 0 {
		if offset+int(lenField) > len(coinBytes) {
			// out of range
			return errors.New("out of range Parse PublicKey")
		}
		data := coinBytes[offset : offset+int(lenField)]
		coin.publicKey, err = new(Point).FromBytesS(data)
		if err != nil {
			return err
		}
		offset += int(lenField)
	}

	// Parse CoinCommitment
	if offset > len(coinBytes) {
		// out of range
		return errors.New("out of range Parse CoinCommitment")
	}
	lenField = coinBytes[offset]
	offset++
	if lenField != 0 {
		if offset+int(lenField) > len(coinBytes) {
			// out of range
			return errors.New("out of range Parse CoinCommitment")
		}
		data := coinBytes[offset : offset+int(lenField)]
		coin.coinCommitment, err = new(Point).FromBytesS(data)
		if err != nil {
			return err
		}
		offset += int(lenField)
	}

	// Parse SNDerivator
	if offset > len(coinBytes) {
		// out of range
		return errors.New("out of range Parse SNDerivator")
	}
	lenField = coinBytes[offset]
	offset++
	if lenField != 0 {
		if offset+int(lenField) > len(coinBytes) {
			// out of range
			return errors.New("out of range Parse SNDerivator")
		}
		data := coinBytes[offset : offset+int(lenField)]
		coin.snDerivator = new(Scalar).FromBytesS(data)

		offset += int(lenField)
	}

	//Parse sn
	if offset > len(coinBytes) {
		// out of range
		return errors.New("out of range Parse sn")
	}
	lenField = coinBytes[offset]
	offset++
	if lenField != 0 {
		if offset+int(lenField) > len(coinBytes) {
			// out of range
			return errors.New("out of range Parse sn")
		}
		data := coinBytes[offset : offset+int(lenField)]
		coin.serialNumber, err = new(Point).FromBytesS(data)
		if err != nil {
			return err
		}
		offset += int(lenField)
	}

	// Parse Randomness
	if offset > len(coinBytes) {
		// out of range
		return errors.New("out of range Parse Randomness")
	}
	lenField = coinBytes[offset]
	offset++
	if lenField != 0 {
		if offset+int(lenField) > len(coinBytes) {
			// out of range
			return errors.New("out of range Parse Randomness")
		}
		data := coinBytes[offset : offset+int(lenField)]
		coin.randomness = new(Scalar).FromBytesS(data)
		offset += int(lenField)
	}

	// Parse Value
	if offset > len(coinBytes) {
		// out of range
		return errors.New("out of range Parse PublicKey")
	}
	lenField = coinBytes[offset]
	offset++
	if lenField != 0 {
		if offset+int(lenField) > len(coinBytes) {
			// out of range
			return errors.New("out of range Parse PublicKey")
		}
		coin.value = new(big.Int).SetBytes(coinBytes[offset : offset+int(lenField)]).Uint64()
		offset += int(lenField)
	}

	// Parse Info
	if offset > len(coinBytes) {
		// out of range
		return errors.New("out of range Parse Info")
	}
	lenField = coinBytes[offset]
	offset++
	if lenField != 0 {
		if offset+int(lenField) > len(coinBytes) {
			// out of range
			return errors.New("out of range Parse Info")
		}
		coin.info = make([]byte, lenField)
		copy(coin.info, coinBytes[offset:offset+int(lenField)])
	}
	return nil
}

// InputCoin represents a input coin of transaction
type InputCoin struct {
	CoinDetails *Coin
}

// Init (InputCoin) initializes a input coin
func (inputCoin *InputCoin) Init() *InputCoin {
	if inputCoin.CoinDetails == nil {
		inputCoin.CoinDetails = new(Coin).Init()
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
	inputCoin.CoinDetails = new(Coin)
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
	inputCoin.CoinDetails = new(Coin).Init()

	if coinObj.PublicKey != "" {
		publicKey, _, err := base58.Base58Check{}.Decode(coinObj.PublicKey)
		if err != nil {
			return err
		}

		publicKeyPoint, err := new(Point).FromBytesS(publicKey)
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

		coinCommitmentPoint, err := new(Point).FromBytesS(coinCommitment)
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

		snderivatorScalar := new(Scalar).FromBytesS(snderivator)
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

		serialNumberPoint, err := new(Point).FromBytesS(serialNumber)
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

		randomnessScalar := new(Scalar).FromBytesS(randomness)
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
	CoinDetails          *Coin
	CoinDetailsEncrypted *HybridCipherText
}

// Init (OutputCoin) initializes a output coin
func (outputCoin *OutputCoin) Init() *OutputCoin {
	outputCoin.CoinDetails = new(Coin).Init()
	outputCoin.CoinDetailsEncrypted = new(HybridCipherText)
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
		outputCoin.CoinDetailsEncrypted = new(HybridCipherText)
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
	outputCoin.CoinDetails = new(Coin)
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
func (outputCoin *OutputCoin) Encrypt(recipientTK TransmissionKey) *PrivacyError {
	// 32-byte first: Randomness, the rest of msg is value of coin
	msg := append(outputCoin.CoinDetails.randomness.ToBytesS(), new(big.Int).SetUint64(outputCoin.CoinDetails.value).Bytes()...)

	pubKeyPoint, err := new(Point).FromBytesS(recipientTK)
	if err != nil {
		return NewPrivacyErr(EncryptOutputCoinErr, err)
	}

	outputCoin.CoinDetailsEncrypted, err = HybridEncrypt(msg, pubKeyPoint)
	if err != nil {
		return NewPrivacyErr(EncryptOutputCoinErr, err)
	}

	return nil
}

// Decrypt decrypts a ciphertext encrypting for coin with recipient's receiving key
func (outputCoin *OutputCoin) Decrypt(viewingKey ViewingKey) *PrivacyError {
	msg, err := HybridDecrypt(outputCoin.CoinDetailsEncrypted, new(Scalar).FromBytesS(viewingKey.Rk))
	if err != nil {
		return NewPrivacyErr(DecryptOutputCoinErr, err)
	}

	// Assign randomness and value to outputCoin details
	outputCoin.CoinDetails.randomness = new(Scalar).FromBytesS(msg[0:Ed25519KeySize])
	outputCoin.CoinDetails.value = new(big.Int).SetBytes(msg[Ed25519KeySize:]).Uint64()

	return nil
}
