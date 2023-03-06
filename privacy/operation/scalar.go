package operation

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
	"sort"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/privacy/operation/edwards25519"
)

// Scalar is a wrapper for `edwards25519.Scalar`, representing an integer modulo group order
type Scalar struct {
	sc edwards25519.Scalar
}

// NewScalar returns a new zero scalar
func NewScalar() *Scalar {
	return &Scalar{*edwards25519.NewScalar()}
}

var ScZero = NewScalar()
var ScOne = NewScalar().FromUint64(1)
var ScMinusOne = NewScalar().Negate(ScOne)

// String returns the hex string representation of `sc`
func (sc Scalar) String() string {
	return hex.EncodeToString(sc.ToBytesS())
}

// MarshalText returns the hex string representation of `sc` but as bytes
func (sc Scalar) MarshalText() []byte {
	return []byte(sc.String())
}

// UnmarshalText decodes a Scalar from its hex string form and sets `sc` to it, then returns `sc`
func (sc *Scalar) UnmarshalText(data []byte) (*Scalar, error) {
	byteSlice, _ := hex.DecodeString(string(data))
	if len(byteSlice) != Ed25519KeySize {
		return nil, fmt.Errorf("invalid scalar byte size")
	}
	return sc.FromBytesS(byteSlice), nil
}

// ToBytesS marshals `sc` into a byte slice
func (sc Scalar) ToBytesS() []byte {
	return sc.sc.Bytes()
}

// FromBytesS unmarshals `sc` from a byte slice, then returns `sc`
func (sc *Scalar) FromBytesS(b []byte) *Scalar {
	var temp [32]byte
	copy(temp[:], b)
	// pad & reduce the input bytes
	sc.sc.SetUnreducedBytes(temp[:]) //nolint:errcheck // valid range is ensured
	return sc
}

// ToBytes marshals `sc` into a byte array
func (sc Scalar) ToBytes() (result [32]byte) {
	copy(result[:], sc.sc.Bytes())
	return result
}

// FromBytes unmarshals `sc` from a byte array, then returns `sc`
func (sc *Scalar) FromBytes(bArr [32]byte) *Scalar {
	sc.sc.SetUnreducedBytes(bArr[:]) //nolint:errcheck // valid range is ensured
	return sc
}

// SetKey is a legacy function
func (sc *Scalar) SetKey(a *[32]byte) (*Scalar, error) {
	_, err := sc.sc.SetCanonicalBytes(a[:])
	return sc, err
}

// Set sets the value of `sc` to that of `a`, then returns `sc`
func (sc *Scalar) Set(a *Scalar) *Scalar {
	sc.sc.Set(&a.sc)
	return sc
}

// RandomScalar returns a random Scalar, generated using `crypto/`'s randomness
func RandomScalar() *Scalar {
	b := make([]byte, 64)
	rand.Read(b) //nolint // no recover from RNG error
	res, _ := edwards25519.NewScalar().SetUniformBytes(b)
	return &Scalar{*res}
}

// HashToScalar creates a Scalar using the bytes from `data` by first hashing it
func HashToScalar(data []byte) *Scalar {
	h := common.Keccak256(data)
	sc := NewScalar()
	sc.sc.SetUnreducedBytes(h[:]) //nolint:errcheck // valid range is ensured
	return sc
}

// FromUint64 sets the value of `sc` to that of `i`
func (sc *Scalar) FromUint64(i uint64) *Scalar {
	bn := big.NewInt(0).SetUint64(i)
	bSlice := common.AddPaddingBigInt(bn, Ed25519KeySize)
	var b [32]byte
	copy(b[:], bSlice)
	rev := Reverse(b)
	sc.sc.SetCanonicalBytes(rev[:]) //nolint:errcheck // valid range is ensured
	return sc
}

// ToUint64Little returns the value of `sc` as uint64.
// The value needs to be in range or this outputs zero.
func (sc *Scalar) ToUint64Little() uint64 {
	var b [32]byte
	copy(b[:], sc.sc.Bytes())
	rev := Reverse(b)
	bn := big.NewInt(0).SetBytes(rev[:])
	return bn.Uint64()
}

// Add sets `sc = a + b` (modulo l), then returns `sc`
func (sc *Scalar) Add(a, b *Scalar) *Scalar {
	sc.sc.Add(&a.sc, &b.sc)
	return sc
}

// Sub sets `sc = a - b` (modulo l), then returns `sc`
func (sc *Scalar) Sub(a, b *Scalar) *Scalar {
	sc.sc.Subtract(&a.sc, &b.sc)
	return sc
}

// Negate sets `sc` to `-a` (modulo l), then returns `sc`
func (sc *Scalar) Negate(a *Scalar) *Scalar {
	sc.sc.Negate(&a.sc)
	return sc
}

// Mul sets `sc = a * b` (modulo l), then returns `sc`
func (sc *Scalar) Mul(a, b *Scalar) *Scalar {
	sc.sc.Multiply(&a.sc, &b.sc)
	return sc
}

// MulAdd sets `sc = a * b + c` (modulo l), then returns `sc`
func (sc *Scalar) MulAdd(a, b, c *Scalar) *Scalar {
	sc.sc.MultiplyAdd(&a.sc, &b.sc, &c.sc)
	return sc
}

// ScalarValid checks the validity of `sc` for use in group operations
func (sc *Scalar) ScalarValid() bool {
	return edwards25519.IsReduced(&sc.sc)
}

// IsScalarEqual checks if two Scalars are equal
func IsScalarEqual(sc1, sc2 *Scalar) bool {
	return sc1.sc.Equal(&sc2.sc) == 1
}

// Compare the value of two Scalars (as non-negative integers)
func Compare(sca, scb *Scalar) int {
	tmpa := sca.ToBytesS()
	tmpb := scb.ToBytesS()

	for i := Ed25519KeySize - 1; i >= 0; i-- {
		if uint64(tmpa[i]) > uint64(tmpb[i]) {
			return 1
		}

		if uint64(tmpa[i]) < uint64(tmpb[i]) {
			return -1
		}
	}
	return 0
}

// IsZero checks if `sc` equals zero
func (sc *Scalar) IsZero() bool {
	return IsScalarEqual(sc, ScZero)
}

// CheckDuplicateScalarArray is a legacy function
func CheckDuplicateScalarArray(arr []*Scalar) bool {
	sort.Slice(arr, func(i, j int) bool {
		return Compare(arr[i], arr[j]) == -1
	})

	for i := 0; i < len(arr)-1; i++ {
		if IsScalarEqual(arr[i], arr[i+1]) {
			return true
		}
	}
	return false
}

func (sc *Scalar) Invert(a *Scalar) *Scalar {
	sc.sc.Invert(&a.sc)
	return sc
}

// Reverse returns a 32-byte array in reverse order (does not change the input)
func Reverse(x [32]byte) (result [32]byte) {
	result = x
	// A key is in little-endian, but the big package wants the bytes in
	// big-endian, so Reverse them.
	blen := len(x) // its hardcoded 32 bytes, so why do len but lets do it
	for i := 0; i < blen/2; i++ {
		result[i], result[blen-1-i] = result[blen-1-i], result[i]
	}
	return
}

// nolint // d2h is a legacy function
func d2h(val uint64) [32]byte {
	var key [32]byte
	for i := 0; val > 0; i++ {
		key[i] = byte(val & 0xFF)
		val /= 256
	}
	return key
}
