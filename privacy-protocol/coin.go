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

	//SerialNumber := []byte{}
	//if coin.SerialNumber != nil {
	//	SerialNumber = coin.SerialNumber.Compress()
	//}
	//if len(SerialNumber) == 0 {
	//	SerialNumber := [33]byte{}
	//	coin_bytes = append(coin_bytes, SerialNumber[:]...)
	//} else {
	//	coin_bytes = append(coin_bytes, SerialNumber...)
	//}
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
	coin_bytes = append(coin_bytes, byte(len(Info)))
	coin_bytes = append(coin_bytes, Info...)
	return coin_bytes
}

func (coin *Coin) SetBytes(coin_byte []byte) {
	offset := 0
	//Parse PubKey
	lenField := coin_byte[offset]
	offset++
	if (lenField != 0) {
		coin.PublicKey = new(EllipticPoint)
		coin.PublicKey.Decompress(coin_byte[offset:offset+int(lenField)])
	}
	offset += int(lenField)

	// Parse CoinCommitment
	lenField = coin_byte[offset]
	offset++
	if (lenField != 0) {
		coin.CoinCommitment = new(EllipticPoint)
		coin.CoinCommitment.Decompress(coin_byte[offset:offset+int(lenField)])
	}
	offset += int(lenField)

	// Parse SNDerivator
	lenField = coin_byte[offset]
	offset++
	if (lenField != 0) {
		coin.SNDerivator = new(big.Int)
		coin.SNDerivator.SetBytes(coin_byte[offset: offset+int(lenField)])
	}
	offset += int(lenField)

	//Parse SN
	lenField = coin_byte[offset]
	offset++
	if (lenField != 0) {
		coin.SerialNumber = new(EllipticPoint)
		coin.SerialNumber.Decompress(coin_byte[offset:offset+int(lenField)])
	}
	offset += int(lenField)
	// Parse Randomness
	lenField = coin_byte[offset]
	offset++
	if (lenField != 0) {
		coin.Randomness = new(big.Int)
		coin.Randomness.SetBytes(coin_byte[offset: offset+int(lenField)])
	}
	offset += int(lenField)

	// Parse Value
	lenField = coin_byte[offset]
	offset++
	if (lenField != 0) {
		x := new(big.Int)
		x.SetBytes(coin_byte[offset:offset+int(lenField)])
		coin.Value = x.Uint64()
	}
	offset += int(lenField)

	// Parse Info
	lenField = coin_byte[offset]
	offset++
	if (lenField != 0) {
		lenField = coin_byte[offset]
		copy(coin.Info, coin_byte[offset:offset+int(lenField)])
		offset += int(lenField)
	}
}

// InputCoin represents a input coin of transaction
type InputCoin struct {
	//ShardId *big.Int
	//BlockHeight *big.Int
	CoinDetails *Coin
}

func (inputCoin *InputCoin) Init() *InputCoin {
	//Todo:
	return inputCoin
}

func (inputCoin *InputCoin) Bytes() []byte {
	return inputCoin.CoinDetails.Bytes()
}
func (inputCoin *InputCoin) SetBytes(bytes []byte) {
	inputCoin.CoinDetails.SetBytes(bytes)
}


type OutputCoin struct {
	CoinDetails          *Coin
	CoinDetailsEncrypted *CoinDetailsEncrypted
}

func (outputCoin *OutputCoin) Init() *OutputCoin {
	//Todo:
	return outputCoin
}

func (outputCoin *OutputCoin) Bytes() []byte {
	var out_coin_bytes []byte
	out_coin_bytes = append(out_coin_bytes, byte(len(outputCoin.CoinDetailsEncrypted.Bytes()))) //112 bytes
	out_coin_bytes = append(out_coin_bytes, outputCoin.CoinDetailsEncrypted.Bytes()...)
	out_coin_bytes = append(out_coin_bytes, outputCoin.CoinDetails.Bytes()...)
	return out_coin_bytes
}

func (outputCoin *OutputCoin) SetBytes(b []byte) {
	length:=int(b[0])
	outputCoin.CoinDetailsEncrypted.SetBytes(b[0:length])
	outputCoin.CoinDetails.SetBytes(b[length:])
}


type CoinDetailsEncrypted struct {
	RandomEncrypted []byte // 48 bytes
	SymKeyEncrypted []byte // 66 bytes
}

func (coinDetailsEncrypted *CoinDetailsEncrypted) Bytes() [] byte {
	var res []byte
	res = append(res, coinDetailsEncrypted.RandomEncrypted...)
	res = append(res, coinDetailsEncrypted.SymKeyEncrypted...)
	return res
}
func(coinDetailsEncrypted *CoinDetailsEncrypted) SetBytes(bytes []byte){
	coinDetailsEncrypted.RandomEncrypted = bytes[0:48]
	coinDetailsEncrypted.SymKeyEncrypted = bytes[48:48+66]
}

func (coin *OutputCoin) Encrypt(receiverTK TransmissionKey) error {
	/**** Generate symmetric key of AES cryptosystem,
				it is used for encryption coin details ****/
	symKeyPoint := new(EllipticPoint)
	symKeyPoint.Randomize()
	symKeyByte := symKeyPoint.X.Bytes()

	/**** Encrypt coin details using symKeyByte ****/
	// just encrypt Randomness of coin
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
		panic(err)
	}

	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(coin.CoinDetailsEncrypted.RandomEncrypted[aes.BlockSize:], randomnessBytes)

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
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return err
	}

	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(randomness, coin.CoinDetailsEncrypted.RandomEncrypted[aes.BlockSize:])

	coin.CoinDetails.Randomness = new(big.Int).SetBytes(randomness)
	// Calculate value of coin
	gRandom := PedCom.G[RAND].ScalarMul(coin.CoinDetails.Randomness)
	gRandomInverse, _ := gRandom.Inverse()
	gShardID := PedCom.G[SHARDID].ScalarMul(big.NewInt(int64(coin.CoinDetails.GetPubKeyLastByte())))
	gShardIDInverse, _ := gShardID.Inverse()
	gSND := PedCom.G[SND].ScalarMul(coin.CoinDetails.SNDerivator)
	gSNDInverse, _ := gSND.Inverse()

	PublicKeyPoint, _ := DecompressKey(viewingKey.Pk)
	PublicKeyPointInverse, _ := PublicKeyPoint.Inverse()

	gValue := coin.CoinDetails.CoinCommitment.Add(gRandomInverse)
	gValue = gValue.Add(gShardIDInverse)
	gValue = gValue.Add(gSNDInverse)
	gValue = gValue.Add(PublicKeyPointInverse)

	// brute force to find value
	for v := 0; ; v++ {
		gv := PedCom.G[VALUE].ScalarMul(big.NewInt(int64(v)))
		if gv.IsEqual(gValue) {
			coin.CoinDetails.Value = uint64(v)
			break
		}
	}

	// assign public key to coin detail
	coin.CoinDetails.PublicKey = PublicKeyPoint

	return nil
}

//CommitAll commits a coin with 5 attributes (public key, value, serial number derivator, last byte pk, r)
func (coin *Coin) CommitAll() {
	values := []*big.Int{big.NewInt(0), big.NewInt(int64(coin.Value)), coin.SNDerivator, new(big.Int).SetBytes([]byte{coin.GetPubKeyLastByte()}), coin.Randomness}
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
