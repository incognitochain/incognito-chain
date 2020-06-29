package privacy

import (
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	C25519 "github.com/incognitochain/incognito-chain/privacy/curve25519"
	"math/big"
	"sort"
)

type Scalar struct {
	key C25519.Key
}

func (sc Scalar) GetKey() C25519.Key {
	return sc.key
}

func (sc Scalar) String() string {
	return fmt.Sprintf("%x", sc.key[:])
}

func (sc Scalar) MarshalText() []byte {
	return []byte(fmt.Sprintf("%x", sc.key[:]))
}

func (sc *Scalar) UnmarshalText(data []byte) (*Scalar, error) {
	if sc == nil {
		sc = new(Scalar)
	}

	byteSlice, _ := hex.DecodeString(string(data))
	if len(byteSlice) != Ed25519KeySize {
		return nil, errors.New("Incorrect key size")
	}
	copy(sc.key[:], byteSlice)
	return sc, nil
}

func (sc Scalar) ToBytes() [Ed25519KeySize]byte {
	return sc.key.ToBytes()
}

func (sc Scalar) ToBytesS() []byte {
	slice := sc.key.ToBytes()
	return slice[:]
}

func (sc *Scalar) FromBytes(b [Ed25519KeySize]byte) *Scalar {
	if sc == nil {
		sc = new(Scalar)
	}
	sc.key.FromBytes(b)
	//if !C25519.ScValid(&sc.key) {
	//	panic("Invalid Scalar Value")
	//}
	return sc
}

func (sc *Scalar) FromBytesS(b []byte) *Scalar {
	//if len(b) != Ed25519KeySize {
	//	panic("Invalid Ed25519 Key Size")
	//}
	if sc == nil {
		sc = new(Scalar)
	}
	var array [Ed25519KeySize]byte
	copy(array[:], b)
	sc.key.FromBytes(array)

	//if !C25519.ScValid(&sc.key) {
	//	panic("Invalid Scalar Value")
	//}
	return sc
}

func (sc *Scalar) SetKey(a *C25519.Key) (*Scalar, error) {
	if sc == nil {
		sc = new(Scalar)
	}
	sc.key = *a
	if sc.ScalarValid() == false {
		return nil, errors.New("Invalid key value")
	}
	return sc, nil
}

func (sc *Scalar) Set(a *Scalar) *Scalar {
	if sc == nil {
		sc = new(Scalar)
	}
	sc.key = a.key
	return sc
}

func RandomScalar() *Scalar {
	sc := new(Scalar)
	key := C25519.RandomScalar()
	sc.key = *key
	return sc
}

func HashToScalar(data []byte) *Scalar {
	key := C25519.HashToScalar(data)
	sc, error := new(Scalar).SetKey(key)
	if error != nil {
		return nil
	}
	return sc
}

func (sc *Scalar) FromUint64(i uint64) *Scalar {
	if sc == nil {
		sc = new(Scalar)
	}
	sc.SetKey(d2h(i))
	return sc
}

func (sc *Scalar) ToUint64() uint64 {
	if sc == nil {
		return 0
	}
	keyBN := new(big.Int).SetBytes(sc.ToBytesS())
	return keyBN.Uint64()
}

func (sc *Scalar) Add(a, b *Scalar) *Scalar {
	if sc == nil {
		sc = new(Scalar)
	}
	var res C25519.Key
	C25519.ScAdd(&res, &a.key, &b.key)
	sc.key = res
	return sc
}

func (sc *Scalar) Sub(a, b *Scalar) *Scalar {
	if sc == nil {
		sc = new(Scalar)
	}
	var res C25519.Key
	C25519.ScSub(&res, &a.key, &b.key)
	sc.key = res
	return sc
}

func (sc *Scalar) Mul(a, b *Scalar) *Scalar {
	if sc == nil {
		sc = new(Scalar)
	}
	var res C25519.Key
	C25519.ScMul(&res, &a.key, &b.key)
	sc.key = res
	return sc
}

// a*b + c % l
func (sc *Scalar) MulAdd(a, b, c *Scalar) *Scalar {
	if sc == nil {
		sc = new(Scalar)
	}
	var res C25519.Key
	C25519.ScMulAdd(&res, &a.key, &b.key, &c.key)
	sc.key = res
	return sc
}

func (sc *Scalar) Exp(a *Scalar, v uint64) *Scalar {
	if sc == nil {
		sc = new(Scalar)
	}

	var res C25519.Key
	C25519.ScMul(&res, &a.key, &a.key)
	for i := 0; i < int(v)-2; i++ {
		C25519.ScMul(&res, &res, &a.key)
	}

	sc.key = res
	return sc
}

func (sc *Scalar) ScalarValid() bool {
	if sc == nil {
		return false
	}
	return C25519.ScValid(&sc.key)
}

func (sc *Scalar) IsOne() bool {
	s := sc.key
	return ((int(s[0]|s[1]|s[2]|s[3]|s[4]|s[5]|s[6]|s[7]|s[8]|
		s[9]|s[10]|s[11]|s[12]|s[13]|s[14]|s[15]|s[16]|s[17]|
		s[18]|s[19]|s[20]|s[21]|s[22]|s[23]|s[24]|s[25]|s[26]|
		s[27]|s[28]|s[29]|s[30]|s[31])-1)>>8)+1 == 1
}

func IsScalarEqual(sc1, sc2 *Scalar) bool {
	tmpa := sc1.ToBytesS()
	tmpb := sc2.ToBytesS()

	return subtle.ConstantTimeCompare(tmpa, tmpb) == 1
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
	if sc == nil {
		return false
	}
	return C25519.ScIsZero(&sc.key)
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
	if sc == nil {
		sc = new(Scalar)
	}

	var inverse_result C25519.Key
	x := a.key

	reversex := Reverse(x)
	bigX := new(big.Int).SetBytes(reversex[:])

	reverseL := Reverse(C25519.CurveOrder()) // as speed improvements it can be made constant
	bigL := new(big.Int).SetBytes(reverseL[:])

	var inverse big.Int
	inverse.ModInverse(bigX, bigL)

	inverse_bytes := inverse.Bytes()

	if len(inverse_bytes) > Ed25519KeySize {
		panic("Inverse cannot be more than Ed25519KeySize bytes in this domain")
	}

	for i, j := 0, len(inverse_bytes)-1; i < j; i, j = i+1, j-1 {
		inverse_bytes[i], inverse_bytes[j] = inverse_bytes[j], inverse_bytes[i]
	}
	copy(inverse_result[:], inverse_bytes[:]) // copy the bytes  as they should be

	sc.key = inverse_result
	return sc
}

func Reverse(x C25519.Key) (result C25519.Key) {
	result = x
	// A key is in little-endian, but the big package wants the bytes in
	// big-endian, so Reverse them.
	blen := len(x) // its hardcoded 32 bytes, so why do len but lets do it
	for i := 0; i < blen/2; i++ {
		result[i], result[blen-1-i] = result[blen-1-i], result[i]
	}
	return
}

func d2h(val uint64) *C25519.Key {
	key := new(C25519.Key)
	for i := 0; val > 0; i++ {
		key[i] = byte(val & 0xFF)
		val /= 256
	}
	return key
}
