package privacy

import (
	"crypto/subtle"
	"fmt"

	//C25519 "github.com/deroproject/derosuite/crypto"
	C25519 "github.com/incognitochain/incognito-chain/privacy/curve25519"
	"testing"
)

func BenchmarkPoint_AddPedersen(b *testing.B) {
	a := RandomScalar()
	c := RandomScalar()

	A := RandomPoint()
	C := RandomPoint()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		new(Point).AddPedersen(a, A, c, C)
	}
}

func TestPoint_ScalarMultPRIME(t *testing.T) {
	for i := 0; i < 10000; i++ {
		a := RandomScalar()
		pa := RandomPoint()
		b := RandomScalar()

		res := new(Point).ScalarMult(pa, a)
		res.ScalarMult(res, b)
		tmpres := res.MarshalText()

		tmp := new(Scalar).Mul(a, b)
		tmpP := new(Point).ScalarMult(pa, tmp)

		resPrime := C25519.ScalarMultKey(&pa.key, &a.key)
		resPrime = C25519.ScalarMultKey(resPrime, &b.key)

		tmpresPrime := resPrime.MarshalText()
		ok := subtle.ConstantTimeCompare(tmpres, tmpresPrime) == 1
		if !ok {
			t.Fatalf("expected Scalar Mul Base correct !")
		}

		ok1 := subtle.ConstantTimeCompare(tmpP.MarshalText(), tmpresPrime) == 1
		if !ok1 {
			t.Fatalf("expected Scalar Mul Base correct !")
		}
	}
}

func TestPoint_MarshalText(t *testing.T) {
	p := RandomPoint()
	fmt.Println(p)
	pByte := p.MarshalText()
	fmt.Println(len(pByte))
	pPrime, _ := new(Point).UnmarshalText(pByte)
	fmt.Println(pPrime)
}

func TestScalarMul(t *testing.T) {
	for i := 0; i < 1000; i++ {
		a := RandomScalar()
		pa := RandomPoint()
		b := RandomScalar()

		res := new(Point).ScalarMult(pa, a)
		res.ScalarMult(res, b)
		res = new(Point).ScalarMult(res, a)
		tmpres := res.MarshalText()

		resPrime := C25519.ScalarMultKey(&pa.key, &a.key)
		resPrime = C25519.ScalarMultKey(resPrime, &b.key)
		resPrime = C25519.ScalarMultKey(resPrime, &a.key)

		tmpresPrime := resPrime.MarshalText()
		ok := subtle.ConstantTimeCompare(tmpres, tmpresPrime) == 1
		if !ok {
			t.Fatalf("expected Scalar Mul Base correct !")
		}
	}
}

func TestScalarMulBase(t *testing.T) {

	Gbytes := PedCom.G[0].ToBytesS()
	fmt.Printf("Gbytes: %v\n", Gbytes)

	array := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}
	aScalar := new(Scalar)
	aScalar.FromBytesS(array)
	res1 := new(Point).ScalarMultBase(aScalar)
	res2 := new(Point).ScalarMult(PedCom.G[0], aScalar)
	fmt.Printf("Res1: %v\n", res1.ToBytesS())
	fmt.Printf("Res2: %v\n", res2.ToBytesS())

	for i := 0; i < 1000; i++ {
		a := RandomScalar()
		b := RandomScalar()

		res1 := new(Point).ScalarMultBase(a)
		res2 := new(Point).ScalarMultBase(b)
		res := new(Point).Add(res1, res2)
		tmpres := res.MarshalText()

		resPrime1 := C25519.ScalarmultBase(&a.key)
		resPrime2 := C25519.ScalarmultBase(&b.key)
		var resPrime C25519.Key

		C25519.AddKeys(&resPrime, resPrime1, resPrime2)

		tmpresPrime := resPrime.MarshalText()
		ok := subtle.ConstantTimeCompare(tmpres, tmpresPrime) == 1
		if !ok {
			t.Fatalf("expected Scalar Mul Base correct !")
		}
	}
}

func TestPoint_Add(t *testing.T) {
	count := 0
	for i := 0; i < 1000; i++ {
		pa := RandomPoint()
		pb := RandomPoint()
		pc := RandomPoint()

		res := new(Point).Add(pa, pb)
		res.Add(res, pc)

		tmpres := res.MarshalText()

		var resPrime C25519.Key
		C25519.AddKeys(&resPrime, &pa.key, &pb.key)
		C25519.AddKeys(&resPrime, &resPrime, &pc.key)

		tmpresPrime := resPrime.MarshalText()
		ok := subtle.ConstantTimeCompare(tmpres, tmpresPrime) == 1
		if !ok {
			t.Fatalf("expected Add correct !")
		}
		resPrimePrime, _ := new(Point).SetKey(&resPrime)
		okk := IsPointEqual(res, resPrimePrime)
		if !okk {
			t.Fatalf("expected Add correct !")
		}
	}

	fmt.Printf("Count wrong: %v\n", count)
}

func TestPoint_Sub(t *testing.T) {
	for i := 0; i < 1000; i++ {
		pa := RandomPoint()
		pb := RandomPoint()
		pc := RandomPoint()

		res := new(Point).Sub(pa, pb)
		res.Sub(res, pc)
		tmpres := res.MarshalText()

		var resPrime C25519.Key
		C25519.SubKeys(&resPrime, &pa.key, &pb.key)
		C25519.SubKeys(&resPrime, &resPrime, &pc.key)

		tmpresPrime := resPrime.MarshalText()
		ok := subtle.ConstantTimeCompare(tmpres, tmpresPrime) == 1
		if !ok {
			t.Fatalf("expected Sub correct !")
		}
		resPrimePrime, _ := new(Point).SetKey(&resPrime)
		okk := IsPointEqual(res, resPrimePrime)
		if !okk {
			t.Fatalf("expected Sub correct !")
		}
	}
}

func TestPoint_InvertScalarMul(t *testing.T) {
	for i := 0; i < 1000; i++ {
		a := RandomScalar()
		pa := RandomPoint()

		// compute (pa^a)^1/a = pa
		res := new(Point).ScalarMult(pa, a)
		res.InvertScalarMult(res, a)
		tmpres := res.MarshalText()

		tmpresPrime := pa.MarshalText()
		ok := subtle.ConstantTimeCompare(tmpres, tmpresPrime) == 1
		if !ok {
			t.Fatalf("expected Invert Scalar Mul correct !")
		}
	}
}

func TestPoint_InvertScalarMultBase(t *testing.T) {
	for i := 0; i < 1000; i++ {
		a := RandomScalar()

		// compute (g^1/a)^a = g
		res := new(Point).InvertScalarMultBase(a)
		res.ScalarMult(res, a)
		tmpres := res.MarshalText()

		tmpresPrime := C25519.GBASE.MarshalText()
		ok := subtle.ConstantTimeCompare(tmpres, tmpresPrime) == 1
		if !ok {
			t.Fatalf("expected Invert Scalar Mul Base correct !")
		}
	}
}

func TestHashToPoint(t *testing.T) {
	for i := 0; i < 10; i++ {
		for j := 0; j < 6; j++ {
			p := HashToPointFromIndex(int64(j), CStringBulletProof)
			fmt.Println(p.key)
		}
		fmt.Println()
	}
}

func TestPoint_FromBytes(t *testing.T) {
	//bytes := [32]byte{}
	//bytes[0] = 12
	//point, err := new(Point).FromBytes(bytes)
	//
	//ok := point.PointValid()
	//
	//if !ok {
	//	t.Fatalf("expected point is valid!")
	//}
}
