package privacy

import (
	"encoding/json"
	"errors"
	"github.com/ninjadotorg/constant/common/base58"
	"math/big"
)

type SerialNumber []byte   // 33 bytes
type CoinCommitment []byte // 33 bytes
type Random []byte         // 32 bytes
type Value []byte          // 32 bytes
type SNDerivator []byte    // 32 bytes

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

// Bytes converts a coin's details to a byte array
func (coin *Coin) Bytes() []byte {
	var coinBytes []byte

	if coin.PublicKey != nil {
		PublicKey := coin.PublicKey.Compress()
		coinBytes = append(coinBytes, byte(len(PublicKey)))
		coinBytes = append(coinBytes, PublicKey...)
	} else {
		coinBytes = append(coinBytes, byte(0))
	}

	if coin.CoinCommitment != nil {
		CoinCommitment := coin.CoinCommitment.Compress()
		coinBytes = append(coinBytes, byte(len(CoinCommitment)))
		coinBytes = append(coinBytes, CoinCommitment...)
	} else {
		coinBytes = append(coinBytes, byte(0))
	}

	if coin.SNDerivator != nil {
		SNDerivator := coin.SNDerivator.Bytes()
		coinBytes = append(coinBytes, byte(len(SNDerivator)))
		coinBytes = append(coinBytes, SNDerivator...)
	} else {
		coinBytes = append(coinBytes, byte(0))
	}

	if coin.SerialNumber != nil {
		SerialNumber := coin.SerialNumber.Compress()
		coinBytes = append(coinBytes, byte(len(SerialNumber)))
		coinBytes = append(coinBytes, SerialNumber...)
	} else {
		coinBytes = append(coinBytes, byte(0))
	}

	if coin.Randomness != nil {
		Randomness := coin.Randomness.Bytes()
		coinBytes = append(coinBytes, byte(len(Randomness)))
		coinBytes = append(coinBytes, Randomness...)
	} else {
		coinBytes = append(coinBytes, byte(0))
	}

	if (coin.Value > 0) {
		Value := new(big.Int).SetUint64(coin.Value).Bytes()
		coinBytes = append(coinBytes, byte(len(Value)))
		coinBytes = append(coinBytes, Value...)
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
	//Parse PubKey
	lenField := coinBytes[offset]
	offset++
	if (lenField != 0) {
		coin.PublicKey = new(EllipticPoint)
		coin.PublicKey, err = DecompressKey(coinBytes[offset : offset+int(lenField)])
		if err != nil {
			return err
		}
		offset += int(lenField)
	}

	// Parse CoinCommitment
	lenField = coinBytes[offset]
	offset++
	if (lenField != 0) {
		coin.CoinCommitment = new(EllipticPoint)
		coin.CoinCommitment, err = DecompressKey(coinBytes[offset : offset+int(lenField)])
		if err != nil {
			return err
		}
		offset += int(lenField)
	}

	// Parse SNDerivator
	lenField = coinBytes[offset]
	offset++
	if (lenField != 0) {
		coin.SNDerivator = new(big.Int)
		coin.SNDerivator.SetBytes(coinBytes[offset : offset+int(lenField)])
		offset += int(lenField)
	}

	//Parse SN
	lenField = coinBytes[offset]
	offset++
	if (lenField != 0) {
		coin.SerialNumber = new(EllipticPoint)
		coin.SerialNumber, err = DecompressKey(coinBytes[offset : offset+int(lenField)])
		if err != nil {
			return err
		}
		offset += int(lenField)
	}

	// Parse Randomness
	lenField = coinBytes[offset]
	offset++
	if (lenField != 0) {
		coin.Randomness = new(big.Int)
		coin.Randomness.SetBytes(coinBytes[offset : offset+int(lenField)])
		offset += int(lenField)
	}

	// Parse Value
	lenField = coinBytes[offset]
	offset++
	if (lenField != 0) {
		x := new(big.Int)
		x.SetBytes(coinBytes[offset : offset+int(lenField)])
		coin.Value = x.Uint64()
		offset += int(lenField)
	}

	// Parse Info
	lenField = coinBytes[offset]
	offset++
	if (lenField != 0) {
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
	if (inputCoin.CoinDetails != nil) {
		inputCoin.CoinDetails.Init()
	}
	return inputCoin
}

func (inputCoin *InputCoin) Bytes() []byte {
	return inputCoin.CoinDetails.Bytes()
}
func (inputCoin *InputCoin) SetBytes(bytes []byte) {
	if len(bytes) == 0 {
		return
	}
	inputCoin.CoinDetails = new(Coin)
	inputCoin.CoinDetails.SetBytes(bytes)
}

type OutputCoin struct {
	CoinDetails          *Coin
	CoinDetailsEncrypted *CoinDetailsEncrypted
}

func (outputCoin *OutputCoin) Init() *OutputCoin {
	if (outputCoin.CoinDetails != nil) {
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
		return nil
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
	outputCoin.CoinDetails.SetBytes(bytes[offset : offset+lenCoinDetail])
	return nil
}

type CoinDetailsEncrypted struct {
	RandomEncrypted []byte // 48 bytes
	ValueEncrypted  []byte // min: 17 bytes, max: 24 bytes
	SymKeyEncrypted []byte // 66 bytes
}

func (self *CoinDetailsEncrypted) Init() *CoinDetailsEncrypted {
	self.RandomEncrypted = []byte{}
	self.ValueEncrypted = []byte{}
	self.SymKeyEncrypted = []byte{}
	return self
}
func (coinDetailsEncrypted *CoinDetailsEncrypted) IsNil() bool {
	if coinDetailsEncrypted.SymKeyEncrypted == nil {
		return true
	}
	if coinDetailsEncrypted.RandomEncrypted == nil {
		return true
	}
	if coinDetailsEncrypted.ValueEncrypted == nil {
		return true
	}
	return false
}

func (coinDetailsEncrypted *CoinDetailsEncrypted) Bytes() [] byte {
	if coinDetailsEncrypted.IsNil() {
		return []byte{}
	}
	var res []byte
	res = append(res, coinDetailsEncrypted.RandomEncrypted...)
	res = append(res, coinDetailsEncrypted.SymKeyEncrypted...)
	res = append(res, coinDetailsEncrypted.ValueEncrypted...)

	return res
}
func (coinDetailsEncrypted *CoinDetailsEncrypted) SetBytes(bytes []byte) error {
	if len(bytes) == 0 {
		return nil
	}

	coinDetailsEncrypted.RandomEncrypted = bytes[0:48]
	coinDetailsEncrypted.SymKeyEncrypted = bytes[48 : 48+66]
	coinDetailsEncrypted.ValueEncrypted = bytes[48+66:]

	return nil
}

// Encrypt encrypts a coin using a hybrid cryptosystem, in which AES encryption scheme is used as a data encapsulation scheme,
// and ElGamal cryptosystem is used as a key encapsulation scheme.
func (coin *OutputCoin) Encrypt(recipientTK TransmissionKey) error {
	// Generate a AES key as the abscissa of a random elliptic point
	secretPoint := new(EllipticPoint)
	secretPoint.Randomize()
	aesKeyByte := secretPoint.X.Bytes()

	// Encrypt coin details using aesKeyByte
	aesScheme := new(AES)
	aesScheme.SetKey(aesKeyByte)

	// encrypt coin's Randomness
	coin.CoinDetailsEncrypted = new(CoinDetailsEncrypted)
	var err error

	randomnessBytes := coin.CoinDetails.Randomness.Bytes()
	coin.CoinDetailsEncrypted.RandomEncrypted, err =  aesScheme.Encrypt(randomnessBytes)
	if err != nil{
		return err
	}

	// encrypt coin's value
	valueBytes := new(big.Int).SetUint64(coin.CoinDetails.Value).Bytes()
	coin.CoinDetailsEncrypted.ValueEncrypted, err =aesScheme.Encrypt(valueBytes)
	if err != nil{
		return err
	}

	// Encrypt aesKeyByte using Transmission key's receiver with ElGamal cryptosystem
	// prepare public key for ElGamal cryptosystem
	pubKey := new(ElGamalPubKey)
	pubKey.H, err = DecompressKey(recipientTK)
	if err != nil {
		return err
	}

	pubKey.Curve = &Curve

	coin.CoinDetailsEncrypted.SymKeyEncrypted = pubKey.Encrypt(secretPoint).Bytes()
	if err != nil {
		return err
	}

	return nil
}

func (coin *OutputCoin) Decrypt(viewingKey ViewingKey) error {
	if coin.CoinDetailsEncrypted.IsNil() {
		return errors.New("encrypted coin details is nil")
	}
	// Decrypt symKeyEncrypted using receiver's receiving key to get symKey
	// prepare private key for Elgamal cryptosystem
	privKey := new(ElGamalPrivKey)
	privKey.Set(&Curve, new(big.Int).SetBytes(viewingKey.Rk))

	// convert byte array to ElGamalCiphertext
	symKeyCipher := new(ElGamalCiphertext)
	symKeyCipher.SetBytes(coin.CoinDetailsEncrypted.SymKeyEncrypted)
	symKeyPoint := privKey.Decrypt(symKeyCipher)

	/*** Decrypt Encrypted using aes key to get coin details (Randomness) ***/
	aesScheme := new(AES)
	aesScheme.SetKey(symKeyPoint.X.Bytes())

	randomness, err := aesScheme.Decrypt(coin.CoinDetailsEncrypted.RandomEncrypted)
	if err != nil {
		return err
	}

	value, err := aesScheme.Decrypt(coin.CoinDetailsEncrypted.ValueEncrypted)
	if err != nil {
		return err
	}

	// assign public key to coin detail
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
