package common

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/ethereum/go-ethereum/crypto"
)

const HashSize = 32

const MaxHashStringSize = HashSize * 2

var ErrHashStrSize = fmt.Errorf("max hash string length is %v bytes", MaxHashStringSize)

type Hash [HashSize]byte

// MarshalText to use Hash as map's key
func (hash Hash) MarshalText() ([]byte, error) {
	return []byte(hash.String()), nil
}

func (hash Hash) UnmarshalText(text []byte) error {
	copy(hash[:], text)
	return nil
}

func (hash *Hash) UnmarshalJSON(data []byte) error {
	hashString := ""
	_ = json.Unmarshal(data, &hashString)
	hash.Decode(hash, hashString)
	return nil
}

// Format writes first few bytes of hash for debugging
func (hash *Hash) Format(f fmt.State, c rune) {
	if c == 'h' {
		t := hash.String()
		f.Write([]byte(t[:8]))
	} else {
		m := "%"
		for i := 0; i < 128; i++ {
			if f.Flag(i) {
				m += string(i)
			}
		}
		m += string(c)
		fmt.Fprintf(f, m, hash[:])
	}
}

/*
String returns the Hash as the hexadecimal string of the byte-reversed
 hash.
*/
func (hash Hash) String() string {
	for i := 0; i < HashSize/2; i++ {
		hash[i], hash[HashSize-1-i] = hash[HashSize-1-i], hash[i]
	}
	return hex.EncodeToString(hash[:])
}

func (hash Hash) StringNotReverse() string {
	return hex.EncodeToString(hash[:])
}

/*
CloneBytes returns a copy of the bytes which represent the hash as a byte
slice.
NOTE: It is generally cheaper to just slice the hash directly thereby reusing the same bytes rather than calling this method.
*/
func (hash *Hash) CloneBytes() []byte {
	newHash := make([]byte, HashSize)
	copy(newHash, hash[:])

	return newHash
}

/*
SetBytes sets the bytes which represent the hash.  An error is returned if the number of bytes passed in is not HashSize.
*/
func (hash *Hash) SetBytes(newHash []byte) error {
	nhlen := len(newHash)
	if nhlen != HashSize {
		return fmt.Errorf("invalid hash length of %v, want %v", nhlen,
			HashSize)
	}
	copy(hash[:], newHash)

	return nil
}
func (hash *Hash) GetBytes() []byte {
	newBytes := []byte{}
	newBytes = make([]byte, len(hash))
	copy(newBytes, hash[:])
	return newBytes
}

// BytesToHash sets b to hash If b is larger than len(h), b will be cropped from the left.
func NewHash(b []byte) (*Hash, error) {
	var h Hash
	err := h.SetBytes(b)
	if err != nil {
		return nil, err
	}
	return &h, nil
}

/*
IsEqual returns true if target is the same as hash.
*/
func (hash *Hash) IsEqual(target *Hash) bool {
	if hash == nil && target == nil {
		return true
	}
	if hash == nil || target == nil {
		return false
	}
	return *hash == *target
}

/*
NewHash returns a new Hash from a byte slice.  An error is returned if the number of bytes passed in is not HashSize.
*/
func (hash Hash) NewHash(newHash []byte) (*Hash, error) {
	err := hash.SetBytes(newHash)
	if err != nil {
		return nil, err
	}
	return &hash, err
}

/*
// NewHashFromStr creates a Hash from a hash string.  The string should be
// the hexadecimal string of a byte-reversed hash, but any missing characters
// result in zero padding at the end of the Hash.
*/
func (hashObj Hash) NewHashFromStr(hash string) (*Hash, error) {
	err := hashObj.Decode(&hashObj, hash)
	if err != nil {
		return nil, err
	}
	return &hashObj, nil
}

func NewHashFromStr(s string) (*Hash, error) {
	var h Hash
	err := h.Decode(&h, s)
	if err != nil {
		return nil, err
	}
	return &h, err
}

/*
// Decode decodes the byte-reversed hexadecimal string encoding of a Hash to a
// destination.
*/
func (hashObj *Hash) Decode(dst *Hash, src string) error {
	// Return error if hash string is too long.
	if len(src) > MaxHashStringSize {
		return ErrHashStrSize
	}

	// Hex decoder expects the hash to be a multiple of two.  When not, pad
	// with a leading zero.
	var srcBytes []byte
	if len(src)%2 == 0 {
		srcBytes = []byte(src)
	} else {
		srcBytes = make([]byte, 1+len(src))
		srcBytes[0] = '0'
		copy(srcBytes[1:], src)
	}

	// Hex decode the source bytes to a temporary destination.
	var reversedHash Hash
	_, err := hex.Decode(reversedHash[HashSize-hex.DecodedLen(len(srcBytes)):], srcBytes)
	if err != nil {
		return err
	}

	// Reverse copy from the temporary hash to destination.  Because the
	// temporary was zeroed, the written result will be correctly padded.
	for i, b := range reversedHash[:HashSize/2] {
		dst[i], dst[HashSize-1-i] = reversedHash[HashSize-1-i], b
	}

	return nil
}

// Cmp compare two hash
// hash = target : return 0
// hash > target : return 1
// hash < target : return -1
func (hash *Hash) Cmp(target *Hash) int {
	for i := 0; i < HashSize; i++ {
		if hash[i] > target[i] {
			return 1
		}
		if hash[i] < target[i] {
			return -1
		}
	}
	return 0
}

func ConvertArrayStringToArrayHash(strs []string) ([]*Hash, error) {
	hashes := []*Hash{}
	for _, str := range strs {
		temp := Hash{}
		hash, err := temp.NewHashFromStr(str)
		if err != nil {
			return nil, err
		}
		hashes = append(hashes, hash)
	}
	return hashes, nil
}

// Keccak256 returns Keccak256 hash as a Hash object for storing and comparing
func Keccak256(data ...[]byte) Hash {
	h := crypto.Keccak256(data...)
	r := Hash{}
	copy(r[:], h)
	return r
}
