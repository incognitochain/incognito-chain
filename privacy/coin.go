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
	PublicKey      *EllipticPoint
	CoinCommitment *EllipticPoint
	SNDerivator    *big.Int
	SerialNumber   *EllipticPoint
	Randomness     *big.Int
	Value          uint64
	Info           []byte //256 bytes
}

// Init (Coin) initializes a coin
func (coin *Coin) Init() *Coin {
	coin.PublicKey = new(EllipticPoint).Zero()
	coin.CoinCommitment = new(EllipticPoint).Zero()
	coin.SNDerivator = new(big.Int)
	coin.SerialNumber = new(EllipticPoint).Zero()
	coin.Randomness = new(big.Int)
	coin.Value = 0
	return coin
}

// GetPubKeyLastByte returns the last byte of public key
func (coin *Coin) GetPubKeyLastByte() byte {
	pubKeyBytes := coin.PublicKey.Compress()
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
	values := []*big.Int{big.NewInt(0), new(big.Int).SetUint64(coin.Value), coin.SNDerivator, new(big.Int).SetBytes([]byte{shardID}), coin.Randomness}
	commitment, err := PedCom.commitAll(values)
	if err != nil {
		return err
	}
	coin.CoinCommitment = commitment
	coin.CoinCommitment = coin.CoinCommitment.Add(coin.PublicKey)
	return nil
}

// Bytes converts a coin's details to a bytes array
// Each fields in coin is saved in len - body format
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
		coinBytes = append(coinBytes, byte(common.BigIntSize))
		coinBytes = append(coinBytes, common.AddPaddingBigInt(coin.SNDerivator, common.BigIntSize)...)
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
		coinBytes = append(coinBytes, byte(common.BigIntSize))
		coinBytes = append(coinBytes, common.AddPaddingBigInt(coin.Randomness, common.BigIntSize)...)
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
		coin.PublicKey = new(EllipticPoint)
		err = coin.PublicKey.Decompress(coinBytes[offset : offset+int(lenField)])
		if err != nil {
			return err
		}
		offset += int(lenField)
	}

	// Parse CoinCommitment
	lenField = coinBytes[offset]
	offset++
	if lenField != 0 {
		coin.CoinCommitment = new(EllipticPoint)
		err = coin.CoinCommitment.Decompress(coinBytes[offset : offset+int(lenField)])
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

	//Parse sn
	lenField = coinBytes[offset]
	offset++
	if lenField != 0 {
		coin.SerialNumber = new(EllipticPoint)
		err = coin.SerialNumber.Decompress(coinBytes[offset : offset+int(lenField)])
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
		coin.Value = new(big.Int).SetBytes(coinBytes[offset : offset+int(lenField)]).Uint64()
		offset += int(lenField)
	}

	// Parse Info
	lenField = coinBytes[offset]
	offset++
	if lenField != 0 {
		coin.Info = make([]byte, lenField)
		copy(coin.Info, coinBytes[offset:offset+int(lenField)])
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
	CoinDetailsEncrypted *Ciphertext
}

// Init (OutputCoin) initializes a output coin
func (outputCoin *OutputCoin) Init() *OutputCoin {
	outputCoin.CoinDetails = new(Coin).Init()
	outputCoin.CoinDetailsEncrypted = new(Ciphertext)
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
		outputCoin.CoinDetailsEncrypted = new(Ciphertext)
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
	msg := append(common.AddPaddingBigInt(outputCoin.CoinDetails.Randomness, common.BigIntSize), new(big.Int).SetUint64(outputCoin.CoinDetails.Value).Bytes()...)

	pubKeyPoint := new(EllipticPoint)
	err := pubKeyPoint.Decompress(recipientTK)
	if err != nil {
		return NewPrivacyErr(DecompressTransmissionKeyErr, err)
	}

	outputCoin.CoinDetailsEncrypted, err = HybridEncrypt(msg, pubKeyPoint)
	if err != nil {
		return NewPrivacyErr(EncryptOutputCoinErr, err)
	}

	return nil
}

// Decrypt decrypts a ciphertext encrypting for coin with recipient's receiving key
func (outputCoin *OutputCoin) Decrypt(viewingKey ViewingKey) *PrivacyError {
	msg, err := HybridDecrypt(outputCoin.CoinDetailsEncrypted, new(big.Int).SetBytes(viewingKey.Rk))
	if err != nil {
		return NewPrivacyErr(DecryptOutputCoinErr, err)
	}

	// Assign randomness and value to outputCoin details
	outputCoin.CoinDetails.Randomness = new(big.Int).SetBytes(msg[0:common.BigIntSize])
	outputCoin.CoinDetails.Value = new(big.Int).SetBytes(msg[common.BigIntSize:]).Uint64()

	return nil
}
