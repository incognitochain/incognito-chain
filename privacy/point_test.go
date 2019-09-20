package privacy

import (
	"crypto/subtle"
	"fmt"
	C25519 "github.com/deroproject/derosuite/crypto"
	"testing"
)

func TestPoint_IsZero(t *testing.T) {
	p := new(Point).Zero()
	fmt.Println(p.IsZero())
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
	for i:=0; i< 1000; i++ {
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

		tmpresPrime, _ := resPrime.MarshalText()
		ok := subtle.ConstantTimeCompare(tmpres, tmpresPrime) == 1
		if !ok {
			t.Fatalf("expected Scalar Mul Base correct !")
		}
	}
}

func TestScalarMulBase(t *testing.T) {
	for i:=0; i< 1000; i++ {
		a := RandomScalar()
		b := RandomScalar()

		res1 := new(Point).ScalarMultBase(a)
		res2 := new(Point).ScalarMultBase(b)
		res := new(Point).Add(res1, res2)
		tmpres := res.MarshalText()

		resPrime1 := C25519.ScalarmultBase(a.key)
		resPrime2 := C25519.ScalarmultBase(b.key)
		var resPrime C25519.Key

		C25519.AddKeys(&resPrime, &resPrime1, &resPrime2)

		tmpresPrime, _ := resPrime.MarshalText()
		ok := subtle.ConstantTimeCompare(tmpres, tmpresPrime) == 1
		if !ok {
			t.Fatalf("expected Scalar Mul Base correct !")
		}
	}
}

func TestPoint_Add(t *testing.T) {
	for i:=0; i< 1000; i++ {
		pa := RandomPoint()
		pb := RandomPoint()
		pc := RandomPoint()

		res := new(Point).Add(pa, pb)
		res.Add(res, pc)
		tmpres := res.MarshalText()

		var resPrime C25519.Key
		C25519.AddKeys(&resPrime, &pa.key, &pb.key)
		C25519.AddKeys(&resPrime, &resPrime, &pc.key)

		tmpresPrime, _ := resPrime.MarshalText()
		ok := subtle.ConstantTimeCompare(tmpres, tmpresPrime) == 1
		if !ok {
			t.Fatalf("expected Add correct !")
		}
		resPrimePrime, _ := new(Point).SetKey(&resPrime)
		okk := IsEqual(res, resPrimePrime)
		if !okk {
			t.Fatalf("expected Add correct !")
		}
	}
}

func TestPoint_Sub(t *testing.T) {
	for i:=0; i< 1000; i++ {
		pa := RandomPoint()
		pb := RandomPoint()
		pc := RandomPoint()

		res := new(Point).Sub(pa, pb)
		res.Sub(res, pc)
		tmpres := res.MarshalText()

		var resPrime C25519.Key
		C25519.SubKeys(&resPrime, &pa.key, &pb.key)
		C25519.SubKeys(&resPrime, &resPrime, &pc.key)

		tmpresPrime, _ := resPrime.MarshalText()
		ok := subtle.ConstantTimeCompare(tmpres, tmpresPrime) == 1
		if !ok {
			t.Fatalf("expected Sub correct !")
		}
		resPrimePrime, _ := new(Point).SetKey(&resPrime)
		okk := IsEqual(res, resPrimePrime)
		if !okk {
			t.Fatalf("expected Sub correct !")
		}
	}
}

func TestPoint_InvertScalarMul(t *testing.T) {
	for i:=0; i< 1000; i++ {
		a := RandomScalar()
		pa := RandomPoint()

		// compute (pa^a)^1/a = pa
		res := new(Point).ScalarMult(pa, a)
		res.InvertScalarMult(res, a)
		tmpres:= res.MarshalText()

		tmpresPrime := pa.MarshalText()
		ok := subtle.ConstantTimeCompare(tmpres, tmpresPrime) == 1
		if !ok {
			t.Fatalf("expected Invert Scalar Mul correct !")
		}
	}
}

func TestPoint_InvertScalarMultBase(t *testing.T) {
	for i:=0; i< 1000; i++ {
		a := RandomScalar()

		// compute (g^1/a)^a = g
		res := new(Point).InvertScalarMultBase(a)
		res.ScalarMult(res, a)
		tmpres := res.MarshalText()

		tmpresPrime, _ := C25519.GBASE.MarshalText()
		ok := subtle.ConstantTimeCompare(tmpres, tmpresPrime) == 1
		if !ok {
			t.Fatalf("expected Invert Scalar Mul Base correct !")
		}
	}
}

func TestHashToPoint(t *testing.T) {
	for i:=0; i< 10; i++ {
		for j:= 0; j< 6; j++ {
			p := HashToPoint(int64(j))
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