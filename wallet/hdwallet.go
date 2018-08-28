package wallet

import (
	"crypto/hmac"
	"crypto/sha512"
	"errors"
	"github.com/ninjadotorg/cash-prototype/cashec"
	"bytes"
)

const (
// FirstHardenedChild is the index of the firxt "harded" child key as per the
// bip32 spec
//FirstHardenedChild = uint32(0x80000000)

// PublicKeyCompressedLength is the byte count of a compressed public key
//PublicKeyCompressedLength = 33
)

var (
	// PrivateWalletVersion is the version flag for serialized private keys
	//PrivateWalletVersion, _ = hex.DecodeString("0488ADE4")

	// PublicWalletVersion is the version flag for serialized private keys
	//PublicWalletVersion, _ = hex.DecodeString("0488B21E")

	// ErrSerializedKeyWrongSize is returned when trying to deserialize a key that
	// has an incorrect length
	ErrSerializedKeyWrongSize = errors.New("Serialized keys should by exactly 82 bytes")

	// ErrHardnedChildPublicKey is returned when trying to create a harded child
	// of the public key
	ErrHardnedChildPublicKey = errors.New("Can't create hardened child for public key")

	// ErrInvalidChecksum is returned when deserializing a key with an incorrect
	// checksum
	ErrInvalidChecksum = errors.New("Checksum doesn't match")

	// ErrInvalidPrivateKey is returned when a derived private key is invalid
	ErrInvalidPrivateKey = errors.New("Invalid private key")

	// ErrInvalidPublicKey is returned when a derived public key is invalid
	ErrInvalidPublicKey = errors.New("Invalid public key")
)

// KeyPair represents a bip32 extended key
type Key struct {
	Depth       byte   // 1 bytes
	ChildNumber []byte // 4 bytes
	ChainCode   []byte // 32 bytes
	KeyPair     cashec.KeyPair
}

// NewMasterKey creates a new master extended key from a seed
func NewMasterKey(seed []byte) (*Key, error) {
	// Generate key and chaincode
	hmac := hmac.New(sha512.New, []byte("Bitcoin seed"))
	_, err := hmac.Write(seed)
	if err != nil {
		return nil, err
	}
	intermediary := hmac.Sum(nil)

	// Split it into our key and chain code
	keyBytes := intermediary[:32]  // use to create master private/public keypair
	chainCode := intermediary[32:] // be used with public key (in keypair) for new child keys

	// Validate key
	/*err = validatePrivateKey(keyBytes)
	if err != nil {
		return nil, err
	}*/

	keyPair, err := (&cashec.KeyPair{}).GenerateKey(keyBytes)
	if err != nil {
		return nil, err
	}

	// Create the key struct
	key := &Key{
		ChainCode:   chainCode,
		KeyPair:     *keyPair,
		Depth:       0x00,
		ChildNumber: []byte{0x00, 0x00, 0x00, 0x00},
	}

	return key, nil
}

// NewChildKey derives a child key from a given parent as outlined by bip32
func (key *Key) NewChildKey(childIdx uint32) (*Key, error) {
	intermediary, err := key.getIntermediary(childIdx)
	if err != nil {
		return nil, err
	}

	newSeed := key.KeyPair.PrivateKey
	newSeed = append(newSeed, intermediary[:32]...)
	newKeypair, err := (&cashec.KeyPair{}).GenerateKey(newSeed)
	// Create child KeyPair with data common to all both scenarios
	childKey := &Key{
		ChildNumber: uint32Bytes(childIdx),
		ChainCode:   intermediary[32:],
		Depth:       key.Depth + 1,
		KeyPair:     *newKeypair,
	}

	return childKey, nil
}

func (key *Key) getIntermediary(childIdx uint32) ([]byte, error) {
	// Get intermediary to create key and chaincode from
	// Hardened children are based on the private key
	// NonHardened children are based on the public key
	childIndexBytes := uint32Bytes(childIdx)

	var data []byte
	//if childIdx >= FirstHardenedChild {
	//	data = append([]byte{0x0}, key.KeyPair.PrivateKey...)
	//} else {
	data = key.KeyPair.PublicKey
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
		keyBytes := key.KeyPair.PrivateKey
		keyBytes = append([]byte{byte(len(key.KeyPair.PrivateKey))}, keyBytes...)
		keyBytes = append([]byte{0x0}, keyBytes...)
		buffer.Write(keyBytes)
	} else {
		keyBytes := key.KeyPair.PublicKey
		keyBytes = append([]byte{byte(len(key.KeyPair.PublicKey))}, keyBytes...)
		keyBytes = append([]byte{0x1}, keyBytes...)
		buffer.Write(keyBytes)
	}

	// Append the standard doublesha256 checksum
	serializedKey, err := addChecksumToBytes(buffer.Bytes())
	if err != nil {
		return nil, err
	}

	return serializedKey, nil
}

// B58Serialize encodes the KeyPair in the standard Bitcoin base58 encoding
func (key *Key) B58Serialize(private bool) string {
	serializedKey, err := key.Serialize(private)
	if err != nil {
		return ""
	}

	return base58Encode(serializedKey)
}

// String encodes the KeyPair in the standard Bitcoin base58 encoding
func (key *Key) String(private bool) string {
	return key.B58Serialize(private)
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
		key.KeyPair.PrivateKey = data[39:keyLength]
	} else {
		key.KeyPair.PublicKey = data[39:keyLength]
	}

	// validate checksum
	cs1, err := checksum(data[0: len(data)-4])
	if err != nil {
		return nil, err
	}

	cs2 := data[len(data)-4:]
	for i := range cs1 {
		if cs1[i] != cs2[i] {
			return nil, ErrInvalidChecksum
		}
	}
	return key, nil
}

// B58Deserialize deserializes a KeyPair encoded in base58 encoding
func B58Deserialize(data string) (*Key, error) {
	b, _, err := base58Decode(data)
	if err != nil {
		return nil, err
	}
	return Deserialize(b)
}
