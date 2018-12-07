package privacy

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"

	"io"
	"math/big"
	"github.com/ninjadotorg/constant/common/base58"
	"encoding/json"
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
	PubKeyLastByte byte
}

func (self Coin) MarshalJSON() ([]byte, error) {
	data := self.Bytes()
	temp := base58.Base58Check{}.Encode(data, byte(0x00))
	return json.Marshal(temp)
}

func (self *Coin) UnmarshalJSON(data []byte) error {
	dataStr := ""
	_ = json.Unmarshal(data, &dataStr)
	temp, _, err := base58.Base58Check{}.Decode(dataStr)
	if err != nil {
		return err
	}
	// TODO

	_ = temp
	return nil
}

func (coin *Coin) Bytes() []byte {
	var coin_bytes []byte
	coin_bytes = append(coin_bytes, coin.PublicKey.Compress()...)
	coin_bytes = append(coin_bytes, coin.CoinCommitment.Compress()...)
	coin_bytes = append(coin_bytes, PadBigInt(coin.SNDerivator, BigIntSize)...)
	coin_bytes = append(coin_bytes, coin.SerialNumber.Compress()...)
	coin_bytes = append(coin_bytes, PadBigInt(coin.Randomness, 2*BigIntSize)...)
	coin_bytes = append(coin_bytes, PadBigInt(new(big.Int).SetUint64(coin.Value),2*BigIntSize)...)
	coin_bytes = append(coin_bytes, coin.Info...)
	coin_bytes = append(coin_bytes, coin.PubKeyLastByte)
	return coin_bytes
}
func (coin *Coin) SetBytes(coin_byte []byte){
	offset:=0
	coin.PublicKey = new(EllipticPoint)
	coin.PublicKey.Decompress(coin_byte[offset:])
	offset+=CompressedPointSize

	coin.CoinCommitment = new(EllipticPoint)
	coin.CoinCommitment.Decompress(coin_byte[offset:])
	offset+=CompressedPointSize

	coin.SNDerivator = new(big.Int)
	coin.SNDerivator.SetBytes(coin_byte[offset:offset+BigIntSize])
	offset+=BigIntSize

	coin.SerialNumber = new(EllipticPoint)
	coin.SerialNumber.Decompress(coin_byte[offset:])
	offset+=CompressedPointSize

	coin.SNDerivator = new(big.Int)
	coin.SNDerivator.SetBytes(coin_byte[offset:offset+2*BigIntSize])
	offset+=2*BigIntSize

	x := new(big.Int)
	x.SetBytes(coin_byte[offset:offset+2*BigIntSize])
	coin.Value = x.Uint64()
	offset+=2*BigIntSize


	coin.Info = coin_byte[offset:offset+InfoLength]
	offset+=InfoLength
	coin.PubKeyLastByte = coin_byte[offset]
}
// InputCoin represents a input coin of transaction
type InputCoin struct {
	//ShardId *big.Int
	//BlockHeight *big.Int
	CoinDetails *Coin
}

func (inputCoin *InputCoin) Bytes() []byte {
	return inputCoin.CoinDetails.Bytes()
}

type OutputCoin struct {
	CoinDetails            *Coin
	CoinDetailsEncrypted   *CoinDetailsEncrypted
}

func (outputCoin *OutputCoin) Bytes() []byte {
	var out_coin_bytes []byte
	out_coin_bytes = append(out_coin_bytes, outputCoin.CoinDetails.Bytes()...)
	out_coin_bytes = append(out_coin_bytes, outputCoin.CoinDetailsEncrypted.Bytes()...)
	return out_coin_bytes
}

func (outputCoin *OutputCoin) SetBytes() {

}

type CoinDetailsEncrypted struct {
	RandomEncrypted []byte
	SymKeyEncrypted *ElGamalCipherText
}

func (coinDetailsEncrypted *CoinDetailsEncrypted) Bytes() [] byte {
	var res []byte
	res = append(res, coinDetailsEncrypted.RandomEncrypted...)
	res = append(res, coinDetailsEncrypted.SymKeyEncrypted.Bytes()...)
	return res
}

func (coin *OutputCoin) Encrypt(receiverTK TransmissionKey) error {
	/**** Generate symmetric key of AES cryptosystem,
				it is used for encryption coin details ****/
	symKeyPoint := new(EllipticPoint)
	symKeyPoint.Randomize()
	symKeyByte := symKeyPoint.X.Bytes()
	//fmt.Printf("Plain text 2: symKey byte: %v\n", symKeyByte)

	/**** Encrypt coin details using symKeyByte ****/
	// just encrypt Randomness of coin
	randomCoin := coin.CoinDetails.Randomness.Bytes()

	block, err := aes.NewCipher(symKeyByte)

	if err != nil {
		return err
	}

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	coin.CoinDetailsEncrypted = new(CoinDetailsEncrypted)
	coin.CoinDetailsEncrypted.RandomEncrypted = make([]byte, aes.BlockSize+len(randomCoin))
	iv := coin.CoinDetailsEncrypted.RandomEncrypted[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic(err)
	}

	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(coin.CoinDetailsEncrypted.RandomEncrypted[aes.BlockSize:], randomCoin)

	/****** Encrypt symKeyByte using Transmission key's receiver with ElGamal cryptosystem ****/
	// prepare public key for ElGamal cryptosystem
	pubKey := new(ElGamalPubKey)
	pubKey.H, _ = DecompressKey(receiverTK)
	pubKey.Curve = &Curve

	coin.CoinDetailsEncrypted.SymKeyEncrypted = pubKey.ElGamalEnc(symKeyPoint)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func (coin *OutputCoin) Decrypt(receivingKey ReceivingKey) error {
	/*** Decrypt symKeyEncrypted using receiver's receiving key to get symKey ***/
	// prepare private key for Elgamal cryptosystem
	privKey := new(ElGamalPrivKey)
	privKey.Set(&Curve, new(big.Int).SetBytes(receivingKey))

	symKeyPoint := privKey.ElGamalDec(coin.CoinDetailsEncrypted.SymKeyEncrypted)

	//fmt.Printf("Decrypted plaintext 2: SymKey : %v\n", symKeyPoint.X.Bytes())

	/*** Decrypt Encrypted using receiver's receiving key to get coin details (Randomness) ***/
	randomness := make([]byte, 32)
	// Set key to decrypt
	block, err := aes.NewCipher(symKeyPoint.X.Bytes())
	if err != nil {
		return err
	}

	iv := coin.CoinDetailsEncrypted.RandomEncrypted[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return err
	}

	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(randomness, coin.CoinDetailsEncrypted.RandomEncrypted[aes.BlockSize:])
	return nil
}

//CommitAll commits a coin with 5 attributes (public key, value, serial number derivator, last byte pk, r)
func (coin *Coin) CommitAll() {
	values := []*big.Int{big.NewInt(0), big.NewInt(int64(coin.Value)), coin.SNDerivator, new(big.Int).SetBytes([]byte{coin.PubKeyLastByte}), coin.Randomness}
	//fmt.Printf("coin info: %v\n", values)
	coin.CoinCommitment = PedCom.CommitAll(values)
	coin.CoinCommitment = coin.CoinCommitment.Add(coin.PublicKey)
}

//// CommitPublicKey commits a public key's coin
//func (coin *Coin) CommitPublicKey() []byte {
//	var values [PCM_CAPACITY-1][]byte
//	values = [PCM_CAPACITY-1][]byte{coin.PublicKey, nil, nil, coin.Randomness}
//
//
//	var commitment []byte
//	commitment = append(commitment, PK)
//	commitment = append(commitment, PedCom.Commit(values)...)
//	return commitment
//}
//
//// CommitValue commits a value's coin
//func (coin *Coin) CommitValue() []byte {
//	var values [PCM_CAPACITY-1][]byte
//	values = [PCM_CAPACITY-1][]byte{nil, coin.H, nil, coin.Randomness}
//
//	var commitment []byte
//	commitment = append(commitment, VALUE)
//	commitment = append(commitment, PedCom.Commit(values)...)
//	return commitment
//}
//
//// CommitSNDerivator commits a serial number's coin
//func (coin *Coin) CommitSNDerivator() []byte {
//	var values [PCM_CAPACITY-1][]byte
//	values = [PCM_CAPACITY-1][]byte{nil, nil, coin.SNDerivator, coin.Randomness}
//
//	var commitment []byte
//	commitment = append(commitment, SND)
//	commitment = append(commitment, PedCom.Commit(values)...)
//	return commitment
//}

// UnspentCoin represents a list of coins to be spent corresponding to spending key
//type UnspentCoin struct {
//	SpendingKey SpendingKey
//	UnspentCoinList map[Coin]big.Int
//}
