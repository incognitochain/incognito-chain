package coin

import (
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

// InputCoin represents a input coin of transaction
type InputCoin struct {
	CoinDetails *CoinV1
}

// Init (InputCoin) initializes a input coin
func (inputCoin *InputCoin) Init() *InputCoin {
	if inputCoin.CoinDetails == nil {
		inputCoin.CoinDetails = new(CoinV1).Init()
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
	inputCoin.CoinDetails = new(CoinV1)
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
	inputCoin.CoinDetails = new(CoinV1).Init()

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
	CoinDetails          *CoinV1
	CoinDetailsEncrypted *henc.HybridCipherText
}

// Init (OutputCoin) initializes a output coin
func (outputCoin *OutputCoin) Init() *OutputCoin {
	outputCoin.CoinDetails = new(CoinV1).Init()
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
	outputCoin.CoinDetails = new(CoinV1)
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
