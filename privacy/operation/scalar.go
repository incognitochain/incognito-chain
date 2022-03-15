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

type Scalar struct {
	sc edwards25519.Scalar
}

func NewScalar() *Scalar {
	return &Scalar{*edwards25519.NewScalar()}
}

var ScZero = NewScalar()
var ScOne = NewScalar().FromUint64(1)
var ScMinusOne = NewScalar().Negate(ScOne)

func (sc Scalar) String() string {
	return hex.EncodeToString(sc.ToBytesS())
}

func (sc Scalar) MarshalText() []byte {
	return []byte(sc.String())
}

func (sc *Scalar) UnmarshalText(data []byte) (*Scalar, error) {
	byteSlice, _ := hex.DecodeString(string(data))
	if len(byteSlice) != Ed25519KeySize {
		return nil, fmt.Errorf("invalid scalar byte size")
	}
	return sc.FromBytesS(byteSlice), nil
}

func (sc Scalar) ToBytesS() []byte {
	return sc.sc.Bytes()
}

func (sc *Scalar) FromBytesS(b []byte) *Scalar {
	var temp [32]byte
	copy(temp[:], b)
	sc.sc.SetUnreducedBytes(temp[:]) // pad & reduce the input bytes
	return sc
}

func (sc Scalar) ToBytes() (result [32]byte) {
	copy(result[:], sc.sc.Bytes())
	return result
}

func (sc *Scalar) FromBytes(bArr [32]byte) *Scalar {
	sc.sc.SetUnreducedBytes(bArr[:])
	return sc
}

// SetKey is a legacy function
func (sc *Scalar) SetKey(a *[32]byte) (*Scalar, error) {
	_, err := sc.sc.SetCanonicalBytes(a[:])
	return sc, err
}

func (sc *Scalar) Set(a *Scalar) *Scalar {
	sc.sc.Set(&a.sc)
	return sc
}

func RandomScalar() *Scalar {
	b := make([]byte, 64)
	rand.Read(b)
	res, _ := edwards25519.NewScalar().SetUniformBytes(b)
	return &Scalar{*res}
}

func HashToScalar(data []byte) *Scalar {
	h := common.Keccak256(data)
	sc := NewScalar()
	sc.sc.SetUnreducedBytes(h[:])
	return sc
}

func (sc *Scalar) FromUint64(i uint64) *Scalar {
	bn := big.NewInt(0).SetUint64(i)
	bSlice := bn.FillBytes(make([]byte, 32))
	var b [32]byte
	copy(b[:], bSlice)
	rev := Reverse(b)
	sc.sc.SetCanonicalBytes(rev[:])
	return sc
}

func (sc *Scalar) ToUint64Little() uint64 {
	var b [32]byte
	copy(b[:], sc.sc.Bytes())
	rev := Reverse(b)
	bn := big.NewInt(0).SetBytes(rev[:])
	return bn.Uint64()
}

func (sc *Scalar) Add(a, b *Scalar) *Scalar {
	sc.sc.Add(&a.sc, &b.sc)
	return sc
}

func (sc *Scalar) Sub(a, b *Scalar) *Scalar {
	sc.sc.Subtract(&a.sc, &b.sc)
	return sc
}

func (sc *Scalar) Negate(a *Scalar) *Scalar {
	sc.sc.Negate(&a.sc)
	return sc
}

func (sc *Scalar) Mul(a, b *Scalar) *Scalar {
	sc.sc.Multiply(&a.sc, &b.sc)
	return sc
}

// a*b + c % l
func (sc *Scalar) MulAdd(a, b, c *Scalar) *Scalar {
	sc.sc.MultiplyAdd(&a.sc, &b.sc, &c.sc)
	return sc
}

func (sc *Scalar) ScalarValid() bool {
	return edwards25519.IsReduced(&sc.sc)
}

func IsScalarEqual(sc1, sc2 *Scalar) bool {
	return sc1.sc.Equal(&sc2.sc) == 1
}

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

func (sc *Scalar) IsZero() bool {
	return IsScalarEqual(sc, ScZero)
}

func CheckDuplicateScalarArray(arr []*Scalar) bool {
	sort.Slice(arr, func(i, j int) bool {
		return Compare(arr[i], arr[j]) == -1
	})

	for i := 0; i < len(arr)-1; i++ {
		if IsScalarEqual(arr[i], arr[i+1]) == true {
			return true
		}
	}
	return false
}

func (sc *Scalar) Invert(a *Scalar) *Scalar {
	sc.sc.Invert(&a.sc)
	return sc
}

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

func d2h(val uint64) [32]byte {
	var key [32]byte
	for i := 0; val > 0; i++ {
		key[i] = byte(val & 0xFF)
		val /= 256
	}
	return key
}
