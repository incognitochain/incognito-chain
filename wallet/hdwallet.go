package wallet

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha512"

	"github.com/incognitochain/incognito-chain/cashec"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
)

// KeyWallet represents with bip32 standard
type KeyWallet struct {
	Depth       byte   // 1 bytes
	ChildNumber []byte // 4 bytes
	ChainCode   []byte // 32 bytes
	KeySet      cashec.KeySet
}

// NewMasterKey creates a new master extended PubKey from a Seed
func NewMasterKey(seed []byte) (*KeyWallet, error) {
	// Generate PubKey and chaincode
	hmac := hmac.New(sha512.New, []byte("Incognito Seed"))
	_, err := hmac.Write(seed)
	if err != nil {
		Logger.log.Error(err)
		return nil, err
	}
	intermediary := hmac.Sum(nil)

	// Split it into our PubKey and chain code
	keyBytes := intermediary[:32]  // use to create master private/public keypair
	chainCode := intermediary[32:] // be used with public PubKey (in keypair) for new Child keys

	keySet := (&cashec.KeySet{}).GenerateKey(keyBytes)

	// Create the PubKey struct
	key := &KeyWallet{
		ChainCode:   chainCode,
		KeySet:      *keySet,
		Depth:       0x00,
		ChildNumber: []byte{0x00, 0x00, 0x00, 0x00},
	}

	return key, nil
}

// NewChildKey derives a Child KeyWallet from a given parent as outlined by bip32
func (key *KeyWallet) NewChildKey(childIdx uint32) (*KeyWallet, error) {
	intermediary, err := key.getIntermediary(childIdx)
	if err != nil {
		return nil, err
	}

	newSeed := []byte{}
	newSeed = append(newSeed[:], intermediary[:32]...)
	newKeyset := (&cashec.KeySet{}).GenerateKey(newSeed)
	// Create Child KeySet with data common to all both scenarios
	childKey := &KeyWallet{
		ChildNumber: uint32Bytes(childIdx),
		ChainCode:   intermediary[32:],
		Depth:       key.Depth + 1,
		KeySet:      *newKeyset,
	}

	return childKey, nil
}

func (key *KeyWallet) getIntermediary(childIdx uint32) ([]byte, error) {
	childIndexBytes := uint32Bytes(childIdx)

	var data []byte
	data = append(data, childIndexBytes...)

	hmac := hmac.New(sha512.New, key.ChainCode)
	_, err := hmac.Write(data)
	if err != nil {
		return nil, err
	}
	return hmac.Sum(nil), nil
}

// Serialize a KeySet to a 78 byte byte slice
func (key *KeyWallet) Serialize(keyType byte) ([]byte, error) {
	// Write fields to buffer in order
	buffer := new(bytes.Buffer)
	buffer.WriteByte(keyType)
	if keyType == PriKeyType {

		buffer.WriteByte(key.Depth)
		buffer.Write(key.ChildNumber)
		buffer.Write(key.ChainCode)

		// Private keys should be prepended with a single null byte
		keyBytes := make([]byte, 0)
		keyBytes = append(keyBytes, byte(len(key.KeySet.PrivateKey))) // set length
		keyBytes = append(keyBytes, key.KeySet.PrivateKey[:]...)      // set pri-key
		buffer.Write(keyBytes)
	} else if keyType == PaymentAddressType {
		keyBytes := make([]byte, 0)
		keyBytes = append(keyBytes, byte(len(key.KeySet.PaymentAddress.Pk))) // set length PaymentAddress
		keyBytes = append(keyBytes, key.KeySet.PaymentAddress.Pk[:]...)      // set PaymentAddress

		keyBytes = append(keyBytes, byte(len(key.KeySet.PaymentAddress.Tk))) // set length Pkenc
		keyBytes = append(keyBytes, key.KeySet.PaymentAddress.Tk[:]...)      // set Pkenc
		buffer.Write(keyBytes)
	} else if keyType == ReadonlyKeyType {
		keyBytes := make([]byte, 0)
		keyBytes = append(keyBytes, byte(len(key.KeySet.ReadonlyKey.Pk))) // set length PaymentAddress
		keyBytes = append(keyBytes, key.KeySet.ReadonlyKey.Pk[:]...)      // set PaymentAddress

		keyBytes = append(keyBytes, byte(len(key.KeySet.ReadonlyKey.Rk))) // set length Skenc
		keyBytes = append(keyBytes, key.KeySet.ReadonlyKey.Rk[:]...)      // set Pkenc
		buffer.Write(keyBytes)
	}

	// Append the standard doublesha256 checksum
	serializedKey, err := addChecksumToBytes(buffer.Bytes())
	if err != nil {
		Logger.log.Error(err)
		return nil, err
	}

	return serializedKey, nil
}

// Base58CheckSerialize encodes the KeySet in the standard Incognito base58 encoding
func (key *KeyWallet) Base58CheckSerialize(keyType byte) string {
	serializedKey, err := key.Serialize(keyType)
	if err != nil {
		return ""
	}

	return base58.Base58Check{}.Encode(serializedKey, common.ZeroByte)
}

// Deserialize a byte slice into a KeySet
func Deserialize(data []byte) (*KeyWallet, error) {
	var key = &KeyWallet{}
	keyType := data[0]
	if keyType == PriKeyType {
		key.Depth = data[1]
		key.ChildNumber = data[2:6]
		key.ChainCode = data[6:38]
		keyLength := int(data[38])
		key.KeySet.PrivateKey = make([]byte, keyLength)
		copy(key.KeySet.PrivateKey[:], data[39:39+keyLength])
	} else if keyType == PaymentAddressType {
		apkKeyLength := int(data[1])
		pkencKeyLength := int(data[apkKeyLength+2])
		key.KeySet.PaymentAddress.Pk = make([]byte, apkKeyLength)
		key.KeySet.PaymentAddress.Tk = make([]byte, pkencKeyLength)
		copy(key.KeySet.PaymentAddress.Pk[:], data[2:2+apkKeyLength])
		copy(key.KeySet.PaymentAddress.Tk[:], data[3+apkKeyLength:3+apkKeyLength+pkencKeyLength])
	} else if keyType == ReadonlyKeyType {
		apkKeyLength := int(data[1])
		skencKeyLength := int(data[apkKeyLength+2])
		key.KeySet.ReadonlyKey.Pk = make([]byte, apkKeyLength)
		key.KeySet.ReadonlyKey.Rk = make([]byte, skencKeyLength)
		copy(key.KeySet.ReadonlyKey.Pk[:], data[2:2+apkKeyLength])
		copy(key.KeySet.ReadonlyKey.Rk[:], data[3+apkKeyLength:3+apkKeyLength+skencKeyLength])
	}

	// validate checksum
	cs1 := base58.ChecksumFirst4Bytes(data[0 : len(data)-4])
	cs2 := data[len(data)-4:]
	for i := range cs1 {
		if cs1[i] != cs2[i] {
			return nil, NewWalletError(InvalidChecksumErr, nil)
		}
	}
	return key, nil
}

// Base58CheckDeserialize deserializes a KeySet encoded in base58 encoding
// because data contained serialized key type -> return object can contain data of keyset with only one of data keyset
func Base58CheckDeserialize(data string) (*KeyWallet, error) {
	b, _, err := base58.Base58Check{}.Decode(data)
	if err != nil {
		return nil, err
	}
	return Deserialize(b)
}

func GetPublicKeyFromPrivateKeyStr(privateKeyStr string) ([]byte, error) {
	keyWallet, err := Base58CheckDeserialize(privateKeyStr)
	if err != nil {
		return []byte{}, err
	}

	keyWallet.KeySet.ImportFromPrivateKeyByte(keyWallet.KeySet.PrivateKey)

	return keyWallet.KeySet.PaymentAddress.Pk, nil
}
