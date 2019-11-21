package privacy

import (
	"crypto/subtle"
	"fmt"
	"github.com/stretchr/testify/assert"
	"math/big"
	//C25519 "github.com/deroproject/derosuite/crypto"
	C25519 "github.com/incognitochain/incognito-chain/privacy/curve25519"
	"testing"
)

func TestCompare(t *testing.T) {
	a := new(Scalar).FromUint64(1001)
	b := new(Scalar).FromUint64(1001)
	fmt.Println(Compare(a, b))
}

func TestCheckDuplicateScalarArray(t *testing.T) {
	a := RandomScalar()
	b := RandomScalar()
	c := RandomScalar()
	d := RandomScalar()

	flag := CheckDuplicateScalarArray([]*Scalar{a, b, c, d, a, b})
	fmt.Println(flag)

	data := []struct {
		arr         []*Scalar
		isDuplicate bool
	}{
		{[]*Scalar{new(Scalar).FromUint64(uint64(100)), new(Scalar).FromUint64(uint64(1000)), new(Scalar).FromUint64(uint64(10000)), new(Scalar).FromUint64(uint64(100000)), new(Scalar).FromUint64(uint64(100000))}, true},
		{[]*Scalar{new(Scalar).FromUint64(uint64(100)), new(Scalar).FromUint64(uint64(1000)), new(Scalar).FromUint64(uint64(10000)), new(Scalar).FromUint64(uint64(100000)), new(Scalar).FromUint64(uint64(1000000))}, false},
	}

	for _, dataItem := range data {
		isDuplicate := CheckDuplicateScalarArray(dataItem.arr)
		assert.Equal(t, dataItem.isDuplicate, isDuplicate)
	}
}

func TestScalar_Mul(t *testing.T) {
	count := 0
	for i := 0; i < 100; i++ {

		a := RandomScalar()
		b := RandomScalar()
		c := RandomScalar()
		res := new(Scalar).Mul(a, b)
		res = res.Mul(res, c)

		curveOrder := C25519.CurveOrder()

		resBN := new(big.Int).SetBytes(res.ToBytesS())
		curveOrderBN := new(big.Int).SetBytes(ArrayToSlice(curveOrder.ToBytes()))

		if resBN.Cmp(curveOrderBN) == 1 {
			count++
			fmt.Printf("Wrong!!!!!\n")
		}

		var resPrime C25519.Key
		C25519.ScMul(&resPrime, &a.key, &b.key)
		C25519.ScMul(&resPrime, &resPrime, &c.key)
		tmp := resPrime.MarshalText()
		ok := subtle.ConstantTimeCompare(res.MarshalText(), tmp) == 1
		if !ok {
			t.Fatalf("expected Scalar Mul correct !")
		}
	}

	fmt.Printf("Count : %v\n", count)

}

func TestScalar_Add(t *testing.T) {
	count := 0
	for i := 0; i < 100; i++ {
		a := RandomScalar()
		b := RandomScalar()
		c := RandomScalar()

		res := new(Scalar).Add(a, b)
		res = res.Add(res, c)
		res = res.Add(res, a)

		var resPrime C25519.Key
		C25519.ScAdd(&resPrime, &a.key, &b.key)
		C25519.ScAdd(&resPrime, &resPrime, &c.key)
		C25519.ScAdd(&resPrime, &resPrime, &a.key)

		tmp := resPrime.MarshalText()
		ok := subtle.ConstantTimeCompare(res.MarshalText(), tmp) == 1
		if !ok {
			t.Fatalf("expected Scalar Mul correct !")
		}
	}

	fmt.Printf("Count : %v\n", count)
}

func TestScalar_Sub(t *testing.T) {

	for i := 0; i < 100; i++ {
		a := RandomScalar()
		b := RandomScalar()
		c := RandomScalar()

		res := new(Scalar).Sub(a, b)
		res = res.Sub(res, c)

		var resPrime C25519.Key
		C25519.ScSub(&resPrime, &a.key, &b.key)
		C25519.ScSub(&resPrime, &resPrime, &c.key)
		tmp := resPrime.MarshalText()
		ok := subtle.ConstantTimeCompare(res.MarshalText(), tmp) == 1
		if !ok {
			t.Fatalf("expected Scalar Mul correct !")
		}
	}
}
func TestScalar_Exp(t *testing.T) {
	for i := 0; i < 1; i++ {
		a := RandomScalar()
		b := uint64(15)

		res := new(Scalar).Exp(a, b)
		resPrime := new(Scalar).Mul(a, a)
		resPrime.Mul(resPrime, a)
		resPrime.Mul(resPrime, a)
		resPrime.Mul(resPrime, a)

		resPrime.Mul(resPrime, a)
		resPrime.Mul(resPrime, a)
		resPrime.Mul(resPrime, a)
		resPrime.Mul(resPrime, a)
		resPrime.Mul(resPrime, a)

		resPrime.Mul(resPrime, a)
		resPrime.Mul(resPrime, a)
		resPrime.Mul(resPrime, a)
		resPrime.Mul(resPrime, a)
		resPrime.Mul(resPrime, a)

		fmt.Println(resPrime)
		fmt.Println(res.key)
	}
}

func TestScalar_Invert(t *testing.T) {
	for i := 0; i < 100; i++ {
		a := RandomScalar()
		inv_a := new(Scalar).Invert(a)

		res := new(Scalar).Mul(a, inv_a)
		ok := res.IsOne()
		if !ok {
			t.Fatalf("expected Scalar Invert correct !")
		}
	}

	b := new(Scalar).FromUint64(1)
	bInverse := b.Invert(b)
	fmt.Printf("bInverse %v\n", bInverse)
}

//func Test(t *testing.T){
//	a := new(Scalar).SetUint64(253)
//	b := new(Scalar).SetUint64(253)
//	c := new(Scalar).Mul(a,b)
//	fmt.Println("c: ", c)
//	cPrime  := Reverse(c.key)
//
//	cB := cPrime.ToBytes()
//	fmt.Println("cB: ", cB)
//	aI := new(big.Int).SetBytes(ArrayToSlice(cB))
//	fmt.Println("aI: ", aI)
//	fmt.Println("aI.Bytes(): ", aI.Bytes())
//	fmt.Println("SliceToArray(aI.Bytes()): ", SliceToArray(aI.Bytes()))
//
//	key := new(C25519.Key)
//	key.FromBytes(SliceToArray(common.AddPaddingBigInt(aI, 32)))
//	fmt.Printf("Key: %v\n", key)
//
//	keyInverse := Reverse(*key)
//	fmt.Printf("keyInverse: %v\n", keyInverse)
//
//	sc, err := new(Scalar).SetKey(key)
//	fmt.Printf("sc: %v\n", sc)
//	fmt.Printf("err: %v\n", err)
//
//}
