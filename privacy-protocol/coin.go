package privacy

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ninjadotorg/constant/common/base58"
	"io"
	"math/big"
)

type SerialNumber []byte   //33 bytes
type CoinCommitment []byte //67 bytes
type Random []byte         //32 bytes
type Value []byte          //32 bytes
type SNDerivator []byte

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

func (coin *Coin) Init() *Coin {
	coin.PublicKey = new(EllipticPoint).Zero()
	coin.CoinCommitment = new(EllipticPoint).Zero()
	coin.SNDerivator = new(big.Int)
	coin.SerialNumber = new(EllipticPoint).Zero()
	coin.Randomness = new(big.Int)
	coin.Value = 0
	return coin
}

func (coin *Coin) Bytes() []byte {
	var coin_bytes []byte
	if coin.PublicKey != nil {
		PublicKey := coin.PublicKey.Compress()
		coin_bytes = append(coin_bytes, byte(len(PublicKey)))
		coin_bytes = append(coin_bytes, PublicKey...)
	} else {
		coin_bytes = append(coin_bytes, byte(0))
	}
	if coin.CoinCommitment != nil {
		CoinCommitment := coin.CoinCommitment.Compress()
		coin_bytes = append(coin_bytes, byte(len(CoinCommitment)))
		coin_bytes = append(coin_bytes, CoinCommitment...)
	} else {
		coin_bytes = append(coin_bytes, byte(0))
	}

	if coin.SNDerivator != nil {
		SNDerivator := coin.SNDerivator.Bytes()
		coin_bytes = append(coin_bytes, byte(len(SNDerivator)))
		coin_bytes = append(coin_bytes, SNDerivator...)
	} else {
		coin_bytes = append(coin_bytes, byte(0))
	}

	if coin.SerialNumber != nil {
		SerialNumber := coin.SerialNumber.Compress()
		coin_bytes = append(coin_bytes, byte(len(SerialNumber)))
		coin_bytes = append(coin_bytes, SerialNumber...)
	} else {
		coin_bytes = append(coin_bytes, byte(0))
	}
	if coin.Randomness != nil {
		Randomness := coin.Randomness.Bytes()
		coin_bytes = append(coin_bytes, byte(len(Randomness)))
		coin_bytes = append(coin_bytes, Randomness...)
	} else {
		coin_bytes = append(coin_bytes, byte(0))
	}
	if (coin.Value > 0) {
		Value := new(big.Int).SetUint64(coin.Value).Bytes()
		coin_bytes = append(coin_bytes, byte(len(Value)))
		coin_bytes = append(coin_bytes, Value...)
	} else {
		coin_bytes = append(coin_bytes, byte(0))
	}
	Info := coin.Info
	if len(Info) > 0{
		coin_bytes = append(coin_bytes, byte(len(Info)))
		coin_bytes = append(coin_bytes, Info...)
	} else{
		coin_bytes = append(coin_bytes, byte(0))
	}

	return coin_bytes
}

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
		coin.PublicKey, err = DecompressKey(coinBytes[offset:offset+int(lenField)])
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
		coin.CoinCommitment, err = DecompressKey(coinBytes[offset:offset+int(lenField)])
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
		coin.SNDerivator.SetBytes(coinBytes[offset: offset+int(lenField)])
		offset += int(lenField)
	}

	//Parse SN
	lenField = coinBytes[offset]
	offset++
	if (lenField != 0) {
		coin.SerialNumber = new(EllipticPoint)
		coin.SerialNumber, err = DecompressKey(coinBytes[offset:offset+int(lenField)])
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
		coin.Randomness.SetBytes(coinBytes[offset: offset+int(lenField)])
		offset += int(lenField)
	}

	// Parse Value
	lenField = coinBytes[offset]
	offset++
	if (lenField != 0) {
		x := new(big.Int)
		x.SetBytes(coinBytes[offset: offset+int(lenField)])
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
		err := outputCoin.CoinDetailsEncrypted.SetBytes(bytes[offset:offset+lenCoinDetailEncrypted])
		if err != nil{
			return err
		}
		offset += lenCoinDetailEncrypted
	}

	lenCoinDetail := int(bytes[offset])
	offset += 1
	outputCoin.CoinDetails = new(Coin)
	outputCoin.CoinDetails.SetBytes(bytes[offset:offset+lenCoinDetail])
	return nil
}

type CoinDetailsEncrypted struct {
	RandomEncrypted []byte // 48 bytes
	ValueEncrypted []byte  // min: 17 bytes, max: 24 bytes
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
func (coinDetailsEncrypted *CoinDetailsEncrypted) SetBytes(bytes []byte) error{
	if len(bytes) == 0 {
		return nil
	}

	coinDetailsEncrypted.RandomEncrypted = bytes[0:48]
	coinDetailsEncrypted.SymKeyEncrypted = bytes[48:48+66]
	coinDetailsEncrypted.ValueEncrypted = bytes[48+66:]

	return nil
}

func (coin *OutputCoin) Encrypt(receiverTK TransmissionKey) error {
	/**** Generate symmetric key of AES cryptosystem,
				it is used for encryption coin details ****/
	symKeyPoint := new(EllipticPoint)
	symKeyPoint.Randomize()
	symKeyByte := symKeyPoint.X.Bytes()

	/**** Encrypt coin details using symKeyByte ****/
	// encrypt coin's Randomness
	randomnessBytes := coin.CoinDetails.Randomness.Bytes()

	block, err := aes.NewCipher(symKeyByte)
	if err != nil {
		return err
	}

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	coin.CoinDetailsEncrypted = new(CoinDetailsEncrypted)
	coin.CoinDetailsEncrypted.RandomEncrypted = make([]byte, aes.BlockSize+len(randomnessBytes))
	iv := coin.CoinDetailsEncrypted.RandomEncrypted[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return err
	}

	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(coin.CoinDetailsEncrypted.RandomEncrypted[aes.BlockSize:], randomnessBytes)

	// encrypt coin's Value
	ValueBytes := new(big.Int).SetUint64(coin.CoinDetails.Value).Bytes()
	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	coin.CoinDetailsEncrypted.ValueEncrypted = make([]byte, aes.BlockSize+len(ValueBytes))

	stream = cipher.NewCTR(block, iv)
	stream.XORKeyStream(coin.CoinDetailsEncrypted.ValueEncrypted[aes.BlockSize:], ValueBytes)

	/****** Encrypt symKeyByte using Transmission key's receiver with ElGamal cryptosystem ****/
	// prepare public key for ElGamal cryptosystem
	pubKey := new(ElGamalPubKey)
	pubKey.H, _ = DecompressKey(receiverTK)
	pubKey.Curve = &Curve

	coin.CoinDetailsEncrypted.SymKeyEncrypted = pubKey.ElGamalEnc(symKeyPoint).Bytes()
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func (coin *OutputCoin) Decrypt(viewingKey ViewingKey) error {
	if coin.CoinDetailsEncrypted.IsNil() {
		return errors.New("coin details encrypted is nil")
	}
	/*** Decrypt symKeyEncrypted using receiver's receiving key to get symKey ***/
	// prepare private key for Elgamal cryptosystem
	privKey := new(ElGamalPrivKey)
	privKey.Set(&Curve, new(big.Int).SetBytes(viewingKey.Rk))

	// convert byte array to ElGamalCipherText
	symKeyCipher := new(ElGamalCipherText)
	symKeyCipher.SetBytes(coin.CoinDetailsEncrypted.SymKeyEncrypted)
	symKeyPoint := privKey.ElGamalDec(symKeyCipher)

	/*** Decrypt Encrypted using receiver's receiving key to get coin details (Randomness) ***/
	randomness := make([]byte, 32)
	// Set key to decrypt
	block, err := aes.NewCipher(symKeyPoint.X.Bytes())
	if err != nil {
		return err
	}

	iv := coin.CoinDetailsEncrypted.RandomEncrypted[:aes.BlockSize]
	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(randomness, coin.CoinDetailsEncrypted.RandomEncrypted[aes.BlockSize:])

	/*** Decrypt Encrypted using receiver's receiving key to get coin details (Value) ***/
	value := make([]byte, len(coin.CoinDetailsEncrypted.ValueEncrypted[aes.BlockSize:]))
	stream = cipher.NewCTR(block, iv)
	stream.XORKeyStream(value, coin.CoinDetailsEncrypted.ValueEncrypted[aes.BlockSize:])

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
