package wallet

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha512"
	"errors"

	"github.com/ninjadotorg/cash-prototype/cashec"
	"github.com/ninjadotorg/cash-prototype/common/base58"
	"github.com/ninjadotorg/cash-prototype/privacy/client"
)

const (
	// FirstHardenedChild is the index of the firxt "harded" Child Key as per the
	// bip32 spec
	//FirstHardenedChild = uint32(0x80000000)

	// PublicKeyCompressedLength is the byte count of a compressed public Key
	PublicKeyCompressedLength = 33
)

var (
	// PrivateWalletVersion is the version flag for serialized private keys
	//PrivateWalletVersion, _ = hex.DecodeString("0488ADE4")

	// PublicWalletVersion is the version flag for serialized private keys
	//PublicWalletVersion, _ = hex.DecodeString("0488B21E")

	// ErrSerializedKeyWrongSize is returned when trying to deserialize a Key that
	// has an incorrect length
	ErrSerializedKeyWrongSize = errors.New("Serialized keys should by exactly 82 bytes")

	// ErrHardnedChildPublicKey is returned when trying to create a harded Child
	// of the public Key
	ErrHardnedChildPublicKey = errors.New("Can't create hardened Child for public Key")

	// ErrInvalidChecksum is returned when deserializing a Key with an incorrect
	// checksum
	ErrInvalidChecksum = errors.New("Checksum doesn't match")

	// ErrInvalidPrivateKey is returned when a derived private Key is invalid
	ErrInvalidPrivateKey = errors.New("Invalid private Key")

	// ErrInvalidPublicKey is returned when a derived public Key is invalid
	ErrInvalidPublicKey = errors.New("Invalid public Key")
)

// KeyPair represents a bip32 extended Key
type Key struct {
	Depth       byte   // 1 bytes
	ChildNumber []byte // 4 bytes
	ChainCode   []byte // 32 bytes
	KeyPair     cashec.KeyPair
}

// NewMasterKey creates a new master extended Key from a Seed
func NewMasterKey(seed []byte) (*Key, error) {
	// Generate Key and chaincode
	hmac := hmac.New(sha512.New, []byte("Bitcoin Seed"))
	_, err := hmac.Write(seed)
	if err != nil {
		Logger.log.Error(err)
		return nil, err
	}
	intermediary := hmac.Sum(nil)

	// Split it into our Key and chain code
	keyBytes := intermediary[:32]  // use to create master private/public keypair
	chainCode := intermediary[32:] // be used with public Key (in keypair) for new Child keys

	// Validate Key
	/*err = validatePrivateKey(keyBytes)
	if err != nil {
		return nil, err
	}*/

	keyPair, err := (&cashec.KeyPair{}).GenerateKey(keyBytes)
	if err != nil {
		Logger.log.Error(err)
		return nil, err
	}

	// Create the Key struct
	key := &Key{
		ChainCode:   chainCode,
		KeyPair:     *keyPair,
		Depth:       0x00,
		ChildNumber: []byte{0x00, 0x00, 0x00, 0x00},
	}

	return key, nil
}

// NewChildKey derives a Child Key from a given parent as outlined by bip32
func (key *Key) NewChildKey(childIdx uint32) (*Key, error) {
	intermediary, err := key.getIntermediary(childIdx)
	if err != nil {
		return nil, err
	}

	newSeed := []byte{}
	newSeed = append(newSeed[:], intermediary[:32]...)
	newKeypair, err := (&cashec.KeyPair{}).GenerateKey(newSeed)
	// Create Child KeyPair with data common to all both scenarios
	childKey := &Key{
		ChildNumber: uint32Bytes(childIdx),
		ChainCode:   intermediary[32:],
		Depth:       key.Depth + 1,
		KeyPair:     *newKeypair,
	}

	return childKey, nil
}

func (key *Key) getIntermediary(childIdx uint32) ([]byte, error) {
	// Get intermediary to create Key and chaincode from
	// Hardened children are based on the private Key
	// NonHardened children are based on the public Key
	childIndexBytes := uint32Bytes(childIdx)

	var data []byte
	//if childIdx >= FirstHardenedChild {
	//	data = append([]byte{0x0}, Key.KeyPair.PrivateKey...)
	//} else {
	// data = key.KeyPair.PublicKey
	//}
	data = append(data, childIndexBytes...)

	hmac := hmac.New(sha512.New, key.ChainCode)
	_, err := hmac.Write(data)
	if err != nil {
		return nil, err
	}
	return hmac.Sum(nil), nil
}

// Serialize a KeyPair to a 78 byte byte slice
func (key *Key) Serialize(privateKey bool) ([]byte, error) {
	// Write fields to buffer in order
	buffer := new(bytes.Buffer)
	buffer.WriteByte(key.Depth)
	buffer.Write(key.ChildNumber)
	buffer.Write(key.ChainCode)
	if privateKey {
		// Private keys should be prepended with a single null byte
		keyBytes := key.KeyPair.PrivateKey[:]
		keyBytes = append([]byte{byte(len(key.KeyPair.PrivateKey))}, keyBytes...)
		keyBytes = append([]byte{0x0}, keyBytes...)
		buffer.Write(keyBytes)
	} else {
		keyBytes := append(key.KeyPair.PublicKey.Apk[:], key.KeyPair.PublicKey.Pkenc[:]...)
		keyBytes = append([]byte{byte(len(key.KeyPair.PublicKey.Apk) + len(key.KeyPair.PublicKey.Pkenc))}, keyBytes...)
		keyBytes = append([]byte{0x1}, keyBytes...)
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

// Base58CheckSerialize encodes the KeyPair in the standard Bitcoin base58 encoding
func (key *Key) Base58CheckSerialize(private bool) string {
	serializedKey, err := key.Serialize(private)
	if err != nil {
		return ""
	}

	return base58.Base58Check{}.Encode(serializedKey, byte(0x00))
}

// Deserialize a byte slice into a KeyPair
func Deserialize(data []byte) (*Key, error) {
	//if len(data) != 101 {
	//	return nil, ErrSerializedKeyWrongSize
	//}
	var key = &Key{}
	key.Depth = data[0]
	key.ChildNumber = data[1:5]
	key.ChainCode = data[5:37]
	keyType := data[37]
	keyLength := data[38]
	if keyType == byte(0) {
		copy(key.KeyPair.PrivateKey[:], data[39:39+keyLength])
		// key.KeyPair.PrivateKey = data[39 : 39+keyLength]
	} else {
		apkEndByte := 39 + client.SpendingAddressLength
		copy(key.KeyPair.PublicKey.Apk[:], data[39:apkEndByte])
		copy(key.KeyPair.PublicKey.Pkenc[:], data[apkEndByte:apkEndByte+(int(keyLength)-client.SpendingAddressLength)])
		// key.KeyPair.PublicKey = data[39 : 39+keyLength]
	}

	// validate checksum
	cs1 := base58.ChecksumFirst4Bytes(data[0 : len(data)-4])
	cs2 := data[len(data)-4:]
	for i := range cs1 {
		if cs1[i] != cs2[i] {
			return nil, ErrInvalidChecksum
		}
	}
	return key, nil
}

// Base58CheckDeserialize deserializes a KeyPair encoded in base58 encoding
func Base58CheckDeserialize(data string) (*Key, error) {
	b, _, err := base58.Base58Check{}.Decode(data)
	if err != nil {
		return nil, err
	}
	return Deserialize(b)
}
