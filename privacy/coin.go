package privacy

import (
	"encoding/json"
	"errors"
	"github.com/ninjadotorg/constant/common/base58"
	"math/big"
)

// Coin represents a coin
type Coin struct {
	PublicKey      *EllipticPoint
	CoinCommitment *EllipticPoint
	SNDerivator    *big.Int
	SerialNumber   *EllipticPoint
	Randomness     *big.Int
	Value          uint64
	Info           []byte //512 bytes
}

// GetPubKeyLastByte returns the last byte of public key
func (coin *Coin) GetPubKeyLastByte() byte {
	pubKeyBytes := coin.PublicKey.Compress()
	return pubKeyBytes[len(pubKeyBytes)-1]
}

func (coin Coin) MarshalJSON() ([]byte, error) {
	data := coin.Bytes()
	temp := base58.Base58Check{}.Encode(data, byte(0x00))
	return json.Marshal(temp)
}

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

// Init initializes a coin
func (coin *Coin) Init() *Coin {
	coin.PublicKey = new(EllipticPoint).Zero()
	coin.CoinCommitment = new(EllipticPoint).Zero()
	coin.SNDerivator = new(big.Int)
	coin.SerialNumber = new(EllipticPoint).Zero()
	coin.Randomness = new(big.Int)
	coin.Value = 0
	return coin
}

// Bytes converts a coin's details to a bytes array
func (coin *Coin) Bytes() []byte {
	var coinBytes []byte

	if coin.PublicKey != nil {
		publicKey := coin.PublicKey.Compress()
		coinBytes = append(coinBytes, byte(len(publicKey)))
		coinBytes = append(coinBytes, publicKey...)
	} else {
		coinBytes = append(coinBytes, byte(0))
	}

	if coin.CoinCommitment != nil {
		coinCommitment := coin.CoinCommitment.Compress()
		coinBytes = append(coinBytes, byte(len(coinCommitment)))
		coinBytes = append(coinBytes, coinCommitment...)
	} else {
		coinBytes = append(coinBytes, byte(0))
	}

	if coin.SNDerivator != nil {
		snDerivator := coin.SNDerivator.Bytes()
		coinBytes = append(coinBytes, byte(len(snDerivator)))
		coinBytes = append(coinBytes, snDerivator...)
	} else {
		coinBytes = append(coinBytes, byte(0))
	}

	if coin.SerialNumber != nil {
		serialNumber := coin.SerialNumber.Compress()
		coinBytes = append(coinBytes, byte(len(serialNumber)))
		coinBytes = append(coinBytes, serialNumber...)
	} else {
		coinBytes = append(coinBytes, byte(0))
	}

	if coin.Randomness != nil {
		randomness := coin.Randomness.Bytes()
		coinBytes = append(coinBytes, byte(len(randomness)))
		coinBytes = append(coinBytes, randomness...)
	} else {
		coinBytes = append(coinBytes, byte(0))
	}

	if coin.Value > 0 {
		value := new(big.Int).SetUint64(coin.Value).Bytes()
		coinBytes = append(coinBytes, byte(len(value)))
		coinBytes = append(coinBytes, value...)
	} else {
		coinBytes = append(coinBytes, byte(0))
	}

	if len(coin.Info) > 0 {
		coinBytes = append(coinBytes, byte(len(coin.Info)))
		coinBytes = append(coinBytes, coin.Info...)
	} else {
		coinBytes = append(coinBytes, byte(0))
	}

	return coinBytes
}

// SetBytes reverts a byte array to a Coin-type
func (coin *Coin) SetBytes(coinBytes []byte) error {
	if len(coinBytes) == 0 {
		return nil
	}

	var err error
	offset := 0

	// Parse PublicKey
	lenField := coinBytes[offset]
	offset++
	if lenField != 0 {
		coin.PublicKey, err = DecompressKey(coinBytes[offset : offset+int(lenField)])
		if err != nil {
			return err
		}
		offset += int(lenField)
	}

	// Parse CoinCommitment
	lenField = coinBytes[offset]
	offset++
	if lenField != 0 {
		coin.CoinCommitment, err = DecompressKey(coinBytes[offset : offset+int(lenField)])
		if err != nil {
			return err
		}
		offset += int(lenField)
	}

	// Parse SNDerivator
	lenField = coinBytes[offset]
	offset++
	if lenField != 0 {
		coin.SNDerivator = new(big.Int)
		coin.SNDerivator.SetBytes(coinBytes[offset : offset+int(lenField)])
		offset += int(lenField)
	}

	//Parse SN
	lenField = coinBytes[offset]
	offset++
	if lenField != 0 {
		coin.SerialNumber, err = DecompressKey(coinBytes[offset : offset+int(lenField)])
		if err != nil {
			return err
		}
		offset += int(lenField)
	}

	// Parse Randomness
	lenField = coinBytes[offset]
	offset++
	if lenField != 0 {
		coin.Randomness = new(big.Int)
		coin.Randomness.SetBytes(coinBytes[offset : offset+int(lenField)])
		offset += int(lenField)
	}

	// Parse Value
	lenField = coinBytes[offset]
	offset++
	if lenField != 0 {
		x := new(big.Int)
		x.SetBytes(coinBytes[offset : offset+int(lenField)])
		coin.Value = x.Uint64()
		offset += int(lenField)
	}

	// Parse Info
	lenField = coinBytes[offset]
	offset++
	if lenField != 0 {
		lenField = coinBytes[offset]
		copy(coin.Info, coinBytes[offset:offset+int(lenField)])
		offset += int(lenField)
	}
	return nil
}

// InputCoin represents a input coin of transaction
type InputCoin struct {
	CoinDetails *Coin
}

func (inputCoin *InputCoin) Init() *InputCoin {
	if inputCoin.CoinDetails == nil {
		inputCoin.CoinDetails = new(Coin)
		inputCoin.CoinDetails.Init()
	}
	return inputCoin
}

func (inputCoin *InputCoin) Bytes() []byte {
	return inputCoin.CoinDetails.Bytes()
}

func (inputCoin *InputCoin) SetBytes(bytes []byte) error {
	if len(bytes) == 0 {
		return errors.New("Bytes array is empty")
	}

	inputCoin.CoinDetails = new(Coin)
	return inputCoin.CoinDetails.SetBytes(bytes)
}

type OutputCoin struct {
	CoinDetails          *Coin
	CoinDetailsEncrypted *CoinDetailsEncrypted
}

func (outputCoin *OutputCoin) Init() *OutputCoin {
	if outputCoin.CoinDetails == nil {
		outputCoin.CoinDetails = new(Coin)
		outputCoin.CoinDetails.Init()
	}

	outputCoin.CoinDetailsEncrypted = new(CoinDetailsEncrypted)
	return outputCoin
}

func (outputCoin *OutputCoin) Bytes() []byte {
	var outCoinBytes []byte

	if outputCoin.CoinDetailsEncrypted != nil {
		coinDetailsEncryptedBytes := outputCoin.CoinDetailsEncrypted.Bytes()
		outCoinBytes = append(outCoinBytes, byte(len(coinDetailsEncryptedBytes))) //114 + ? bytes
		outCoinBytes = append(outCoinBytes, coinDetailsEncryptedBytes...)
	} else {
		outCoinBytes = append(outCoinBytes, byte(0))
	}

	coinDetailBytes := outputCoin.CoinDetails.Bytes()
	outCoinBytes = append(outCoinBytes, byte(len(coinDetailBytes)))
	outCoinBytes = append(outCoinBytes, coinDetailBytes...)
	return outCoinBytes
}

func (outputCoin *OutputCoin) SetBytes(bytes []byte) error {
	if len(bytes) == 0 {
		return errors.New("Bytes array is empty")
	}

	offset := 0
	lenCoinDetailEncrypted := int(bytes[0])
	offset += 1

	if lenCoinDetailEncrypted > 0 {
		outputCoin.CoinDetailsEncrypted = new(CoinDetailsEncrypted)
		err := outputCoin.CoinDetailsEncrypted.SetBytes(bytes[offset : offset+lenCoinDetailEncrypted])
		if err != nil {
			return err
		}
		offset += lenCoinDetailEncrypted
	}

	lenCoinDetail := int(bytes[offset])
	offset += 1
	outputCoin.CoinDetails = new(Coin)
	return outputCoin.CoinDetails.SetBytes(bytes[offset : offset+lenCoinDetail])
}

type CoinDetailsEncrypted struct {
	EncryptedRandomness []byte // 48 bytes
	EncryptedValue      []byte // 17 -> 24 bytes
	EncryptedSymKey     []byte // 66 bytes
}

func (coinDetailsEncrypted *CoinDetailsEncrypted) Init() *CoinDetailsEncrypted {
	coinDetailsEncrypted.EncryptedRandomness = []byte{}
	coinDetailsEncrypted.EncryptedValue = []byte{}
	coinDetailsEncrypted.EncryptedSymKey = []byte{}
	return coinDetailsEncrypted
}
func (coinDetailsEncrypted *CoinDetailsEncrypted) IsNil() bool {
	if coinDetailsEncrypted.EncryptedSymKey == nil {
		return true
	}
	if coinDetailsEncrypted.EncryptedRandomness == nil {
		return true
	}
	if coinDetailsEncrypted.EncryptedValue == nil {
		return true
	}
	return false
}

func (coinDetailsEncrypted *CoinDetailsEncrypted) Bytes() [] byte {
	if coinDetailsEncrypted.IsNil() {
		return []byte{}
	}

	var res []byte
	res = append(res, coinDetailsEncrypted.EncryptedRandomness...)
	res = append(res, coinDetailsEncrypted.EncryptedSymKey...)
	res = append(res, coinDetailsEncrypted.EncryptedValue...)

	return res
}
func (coinDetailsEncrypted *CoinDetailsEncrypted) SetBytes(bytes []byte) error {
	if len(bytes) == 0 {
		return nil
	}

	coinDetailsEncrypted.EncryptedRandomness = bytes[0:48]
	coinDetailsEncrypted.EncryptedSymKey = bytes[48:48+66]
	coinDetailsEncrypted.EncryptedValue = bytes[48+66:]

	return nil
}

// Encrypt returns a ciphertext encrypting for a coin using a hybrid cryptosystem,
// in which AES encryption scheme is used as a data encapsulation scheme,
// and ElGamal cryptosystem is used as a key encapsulation scheme.
func (coin *OutputCoin) Encrypt(recipientTK TransmissionKey) error {
	// Generate a AES key as the abscissa of a random elliptic point
	aesKeyPoint := new(EllipticPoint)
	aesKeyPoint.Randomize()
	aesKeyByte := aesKeyPoint.X.Bytes()

	// Encrypt coin details using aesKeyByte
	aesScheme := new(AES)
	aesScheme.SetKey(aesKeyByte)

	// Encrypt coin randomness
	coin.CoinDetailsEncrypted = new(CoinDetailsEncrypted)
	var err error

	randomnessBytes := coin.CoinDetails.Randomness.Bytes()
	coin.CoinDetailsEncrypted.EncryptedRandomness, err = aesScheme.Encrypt(randomnessBytes)
	if err != nil {
		return err
	}

	// Encrypt coin value
	valueBytes := new(big.Int).SetUint64(coin.CoinDetails.Value).Bytes()
	coin.CoinDetailsEncrypted.EncryptedValue, err = aesScheme.Encrypt(valueBytes)
	if err != nil {
		return err
	}

	// Get transmission key, which is a public key of ElGamal cryptosystem
	transmissionKey := new(ElGamalPubKey)
	transmissionKey.H, err = DecompressKey(recipientTK)
	if err != nil {
		return err
	}
	transmissionKey.Curve = &Curve

	// Encrypt aesKeyByte under recipient's transmission key using ElGamal cryptosystem
	coin.CoinDetailsEncrypted.EncryptedSymKey = transmissionKey.Encrypt(aesKeyPoint).Bytes()

	return nil
}

// Decrypt decrypts a ciphertext encrypting for coin with recipient's receiving key
func (coin *OutputCoin) Decrypt(viewingKey ViewingKey) error {
	// Validate ciphertext
	if coin.CoinDetailsEncrypted.IsNil() {
		return errors.New("Ciphertext must not be nil")
	}

	// Get receiving key, which is a private key of ElGamal cryptosystem
	receivingKey := new(ElGamalPrivKey)
	receivingKey.Set(&Curve, new(big.Int).SetBytes(viewingKey.Rk))

	// Parse encrypted AES key encoded as an elliptic point from EncryptedSymKey
	encryptedAESKey := new(ElGamalCiphertext)
	err := encryptedAESKey.SetBytes(coin.CoinDetailsEncrypted.EncryptedSymKey)
	if err != nil{
		return err
	}

	// Decrypt encryptedAESKey using recipient's receiving key
	aesKeyPoint := receivingKey.Decrypt(encryptedAESKey)

	// Get AES key
	aesScheme := new(AES)
	aesScheme.SetKey(aesKeyPoint.X.Bytes())

	// Decrypt encrypted coin randomness using AES key
	randomness, err := aesScheme.Decrypt(coin.CoinDetailsEncrypted.EncryptedRandomness)
	if err != nil {
		return err
	}

	// Decrypt encrypted coin value using AES key
	value, err := aesScheme.Decrypt(coin.CoinDetailsEncrypted.EncryptedValue)
	if err != nil {
		return err
	}

	// Assign randomness and value to coin details
	coin.CoinDetails.Randomness = new(big.Int).SetBytes(randomness)
	coin.CoinDetails.Value = new(big.Int).SetBytes(value).Uint64()

	return nil
}

//CommitAll commits a coin with 5 attributes (public key, value, serial number derivator, last byte pk, r)
func (coin *Coin) CommitAll() {
	values := []*big.Int{big.NewInt(0), new(big.Int).SetUint64(coin.Value), coin.SNDerivator, new(big.Int).SetBytes([]byte{coin.GetPubKeyLastByte()}), coin.Randomness}
	coin.CoinCommitment = PedCom.CommitAll(values)
	coin.CoinCommitment = coin.CoinCommitment.Add(coin.PublicKey)
}
