package privacy

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"golang.org/x/crypto/openpgp/elgamal"
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
	Randomness     *big.Int
	Value          uint64
	Info           []byte
}

// InputCoin represents a input coin of transaction
type InputCoin struct {
	BlockHeight *big.Int
	CoinDetails *Coin
}

type OutputCoin struct{
	CoinDetails   *Coin
	CoinDetailsEncrypted CoinDetailsEncrypted
}

type CoinDetailsEncrypted struct{
	RandomEncrypted []byte
	SymKeyEncrypted []byte
}


func (coin *Coin) Encrypt(receiverTK TransmissionKey) (*CoinDetailsEncrypted, error){
	/**** Generate symmetric key of AES cryptosystem,
				it is used for encryption coin details ****/
	var point EllipticPoint
	point.Randomize()
	symKey := point.X.Bytes()

	/**** Encrypt coin details using symKey ****/
	// just encrypt Randomness
	// Load your secret key from a safe place and reuse it across multiple
	// NewCipher calls. (Obviously don't use this example key for anything
	// real.) If you want to convert a passphrase to a key, use a suitable
	// package like bcrypt or scrypt.
	//key, _ := hex.DecodeString("6368616e676520746869732070617373")
	plaintext := []byte("some plaintext")

	block, err := aes.NewCipher(symKey)
	if err != nil {
		return nil, err
	}

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic(err)
	}

	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)

	encryptedCoin := new(CoinDetailsEncrypted)
	encryptedCoin.RandomEncrypted = ciphertext

	// Encrypt symKey using Transmission key's receiver with ElGamal cryptosystem
	// prepare public key for ElGamal cryptosystem
	pubKey := new(elgamal.PublicKey)

	encryptedCoin.SymKeyEncrypted = elgamal.Encrypt(rand.Reader, receiverTK, symKey)


	return encryptedCoin, nil

}

//CommitAll commits a coin with 4 attributes (public key, value, serial number, r)
//func (coin *Coin) CommitAll() {
//	//var values [PCM_CAPACITY-1][]byte
//	values := [PCM_CAPACITY][]byte{coin.PublicKey, coin.Value, coin.SNDerivator, coin.Randomness}
//	fmt.Printf("coin info: %v\n", values)
//	coin.CoinCommitment = append(coin.CoinCommitment, FULL)
//	coin.CoinCommitment = append(coin.CoinCommitment, PedCom.Commit(values)...)
//}

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
//	values = [PCM_CAPACITY-1][]byte{nil, coin.Value, nil, coin.Randomness}
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



