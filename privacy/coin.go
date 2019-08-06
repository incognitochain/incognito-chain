package privacy

import (
	"encoding/json"
	"errors"
	"math/big"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
)

// Coin represents a coin
type Coin struct {
	publicKey      *EllipticPoint
	coinCommitment *EllipticPoint
	snDerivator    *big.Int
	serialNumber   *EllipticPoint
	randomness     *big.Int
	value          uint64
	info           []byte //256 bytes
}

// Start GET/SET
func (coin Coin) GetPublicKey() *EllipticPoint {
	return coin.publicKey
}

func (coin *Coin) SetPublicKey(v *EllipticPoint) {
	coin.publicKey = v
}

func (coin Coin) GetCoinCommitment() *EllipticPoint {
	return coin.coinCommitment
}

func (coin *Coin) SetCoinCommitment(v *EllipticPoint) {
	coin.coinCommitment = v
}

func (coin Coin) GetSNDerivator() *big.Int {
	return coin.snDerivator
}

func (coin *Coin) SetSNDerivator(v *big.Int) {
	coin.snDerivator = v
}

func (coin Coin) GetSerialNumber() *EllipticPoint {
	return coin.serialNumber
}

func (coin *Coin) SetSerialNumber(v *EllipticPoint) {
	coin.serialNumber = v
}

func (coin Coin) GetRandomness() *big.Int {
	return coin.randomness
}

func (coin *Coin) SetRandomness(v *big.Int) {
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
	copy(coin.info, v)
}

// END Get/Set

// Init (Coin) initializes a coin
func (coin *Coin) Init() *Coin {
	coin.publicKey = new(EllipticPoint).Zero()
	coin.coinCommitment = new(EllipticPoint).Zero()
	coin.snDerivator = new(big.Int)
	coin.serialNumber = new(EllipticPoint).Zero()
	coin.randomness = new(big.Int)
	coin.value = 0
	return coin
}

// GetPubKeyLastByte returns the last byte of public key
func (coin *Coin) GetPubKeyLastByte() byte {
	pubKeyBytes := coin.publicKey.Compress()
	return pubKeyBytes[len(pubKeyBytes)-1]
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
	values := []*big.Int{big.NewInt(0), new(big.Int).SetUint64(coin.value), coin.snDerivator, new(big.Int).SetBytes([]byte{shardID}), coin.randomness}
	commitment, err := PedCom.commitAll(values)
	if err != nil {
		return err
	}
	coin.coinCommitment = commitment
	coin.coinCommitment = coin.coinCommitment.Add(coin.publicKey)
	return nil
}

// Bytes converts a coin's details to a bytes array
// Each fields in coin is saved in len - body format
func (coin *Coin) Bytes() []byte {
	var coinBytes []byte

	if coin.publicKey != nil {
		publicKey := coin.publicKey.Compress()
		coinBytes = append(coinBytes, byte(len(publicKey)))
		coinBytes = append(coinBytes, publicKey...)
	} else {
		coinBytes = append(coinBytes, byte(0))
	}

	if coin.coinCommitment != nil {
		coinCommitment := coin.coinCommitment.Compress()
		coinBytes = append(coinBytes, byte(len(coinCommitment)))
		coinBytes = append(coinBytes, coinCommitment...)
	} else {
		coinBytes = append(coinBytes, byte(0))
	}

	if coin.snDerivator != nil {
		coinBytes = append(coinBytes, byte(common.BigIntSize))
		coinBytes = append(coinBytes, common.AddPaddingBigInt(coin.snDerivator, common.BigIntSize)...)
	} else {
		coinBytes = append(coinBytes, byte(0))
	}

	if coin.serialNumber != nil {
		serialNumber := coin.serialNumber.Compress()
		coinBytes = append(coinBytes, byte(len(serialNumber)))
		coinBytes = append(coinBytes, serialNumber...)
	} else {
		coinBytes = append(coinBytes, byte(0))
	}

	if coin.randomness != nil {
		coinBytes = append(coinBytes, byte(common.BigIntSize))
		coinBytes = append(coinBytes, common.AddPaddingBigInt(coin.randomness, common.BigIntSize)...)
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
		coinBytes = append(coinBytes, byte(len(coin.info)))
		coinBytes = append(coinBytes, coin.info...)
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

	var err error
	offset := 0

	// Parse PublicKey
	lenField := coinBytes[offset]
	offset++
	if lenField != 0 {
		coin.publicKey = new(EllipticPoint)
		err = coin.publicKey.Decompress(coinBytes[offset : offset+int(lenField)])
		if err != nil {
			return err
		}
		offset += int(lenField)
	}

	// Parse CoinCommitment
	lenField = coinBytes[offset]
	offset++
	if lenField != 0 {
		coin.coinCommitment = new(EllipticPoint)
		err = coin.coinCommitment.Decompress(coinBytes[offset : offset+int(lenField)])
		if err != nil {
			return err
		}
		offset += int(lenField)
	}

	// Parse SNDerivator
	lenField = coinBytes[offset]
	offset++
	if lenField != 0 {
		coin.snDerivator = new(big.Int)
		coin.snDerivator.SetBytes(coinBytes[offset : offset+int(lenField)])
		offset += int(lenField)
	}

	//Parse sn
	lenField = coinBytes[offset]
	offset++
	if lenField != 0 {
		coin.serialNumber = new(EllipticPoint)
		err = coin.serialNumber.Decompress(coinBytes[offset : offset+int(lenField)])
		if err != nil {
			return err
		}
		offset += int(lenField)
	}

	// Parse Randomness
	lenField = coinBytes[offset]
	offset++
	if lenField != 0 {
		coin.randomness = new(big.Int)
		coin.randomness.SetBytes(coinBytes[offset : offset+int(lenField)])
		offset += int(lenField)
	}

	// Parse Value
	lenField = coinBytes[offset]
	offset++
	if lenField != 0 {
		coin.value = new(big.Int).SetBytes(coinBytes[offset : offset+int(lenField)]).Uint64()
		offset += int(lenField)
	}

	// Parse Info
	lenField = coinBytes[offset]
	offset++
	if lenField != 0 {
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

// OutputCoin represents a output coin of transaction
// It contains CoinDetails and CoinDetailsEncrypted (encrypted value and randomness)
// CoinDetailsEncrypted is nil when you send tx without privacy
type OutputCoin struct {
	CoinDetails          *Coin
	CoinDetailsEncrypted *hybridCiphertext
}

// Init (OutputCoin) initializes a output coin
func (outputCoin *OutputCoin) Init() *OutputCoin {
	outputCoin.CoinDetails = new(Coin).Init()
	outputCoin.CoinDetailsEncrypted = new(hybridCiphertext)
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
	outCoinBytes = append(outCoinBytes, byte(len(coinDetailBytes)))
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
		outputCoin.CoinDetailsEncrypted = new(hybridCiphertext)
		err := outputCoin.CoinDetailsEncrypted.SetBytes(bytes[offset : offset+lenCoinDetailEncrypted])
		if err != nil {
			return err
		}
		offset += lenCoinDetailEncrypted
	}

	lenCoinDetail := int(bytes[offset])
	offset += 1

	if lenCoinDetail > 0 {
		outputCoin.CoinDetails = new(Coin)
		err := outputCoin.CoinDetails.SetBytes(bytes[offset : offset+lenCoinDetail])
		if err != nil {
			return err
		}
	}

	return nil
}

// Encrypt returns a ciphertext encrypting for a coin using a hybrid cryptosystem,
// in which AES encryption scheme is used as a data encapsulation scheme,
// and ElGamal cryptosystem is used as a key encapsulation scheme.
func (outputCoin *OutputCoin) Encrypt(recipientTK TransmissionKey) *PrivacyError {
	// 32-byte first: Randomness, the rest of msg is value of coin
	msg := append(common.AddPaddingBigInt(outputCoin.CoinDetails.randomness, common.BigIntSize), new(big.Int).SetUint64(outputCoin.CoinDetails.value).Bytes()...)

	pubKeyPoint := new(EllipticPoint)
	err := pubKeyPoint.Decompress(recipientTK)
	if err != nil {
		return NewPrivacyErr(DecompressTransmissionKeyErr, err)
	}

	outputCoin.CoinDetailsEncrypted, err = hybridEncrypt(msg, pubKeyPoint)
	if err != nil {
		return NewPrivacyErr(EncryptOutputCoinErr, err)
	}

	return nil
}

// Decrypt decrypts a ciphertext encrypting for coin with recipient's receiving key
func (outputCoin *OutputCoin) Decrypt(viewingKey ViewingKey) *PrivacyError {
	msg, err := hybridDecrypt(outputCoin.CoinDetailsEncrypted, new(big.Int).SetBytes(viewingKey.Rk))
	if err != nil {
		return NewPrivacyErr(DecryptOutputCoinErr, err)
	}

	// Assign randomness and value to outputCoin details
	outputCoin.CoinDetails.randomness = new(big.Int).SetBytes(msg[0:common.BigIntSize])
	outputCoin.CoinDetails.value = new(big.Int).SetBytes(msg[common.BigIntSize:]).Uint64()

	return nil
}
