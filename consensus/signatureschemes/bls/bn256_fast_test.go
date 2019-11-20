package bls

import (
	"bytes"
	"crypto/rand"
	"fmt"
	bn256 "github.com/go-ethereum-master/crypto/bn256/cloudflare"
	"math/big"
	"testing"
)

func TestCompressG1(t *testing.T) {
	fmt.Println(bn256.P)
	for i := 0; i < 2000; i++ {
		_, X, _ := bn256.RandomG1(rand.Reader)
		tmp := CompressG1(X)
		XPrime, _ := DecompressG1(tmp)
		if !bytes.Equal(X.Marshal(), XPrime.Marshal()) {
			t.Error("Compress and Decompress failed")
		}
	}
}

func TestMultiScalarMultG2(t *testing.T) {
	for i := 0; i < 100; i++ {

		len := 64

		pointLs := make([]*bn256.G2, len)
		scalarLs := make([]*big.Int, len)

		for j := 0; j < len; j++ {
			x, X, _ := bn256.RandomG2(rand.Reader)
			pointLs[j] = new(bn256.G2).Set(X)
			scalarLs[j] = new(big.Int).Set(x)
		}

		// MultiscalarMult
		res := MultiScalarMultG2(pointLs, scalarLs)

		// Add list of ScalarMult
		resPrime := new(bn256.G2).ScalarMult(pointLs[0], scalarLs[0])
		for j := 1; j < len; j++ {
			tmp := new(bn256.G2).ScalarMult(pointLs[j], scalarLs[j])
			resPrime.Add(resPrime, tmp)
		}

		if !bytes.Equal(res.Marshal(), resPrime.Marshal()) {
			t.Error("MultiScalarMultG1 failed")
		}
	}
}

func BenchmarkG2_MultiScalarMulNormal(b *testing.B) {
	len := 64

	pointLs := make([]*bn256.G2, len)
	scalarLs := make([]*big.Int, len)

	for j := 0; j < len; j++ {
		x, X, _ := bn256.RandomG2(rand.Reader)
		pointLs[j] = new(bn256.G2).Set(X)
		scalarLs[j] = new(big.Int).Set(x)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Add list of Multiscalar
		resPrime := new(bn256.G2).ScalarMult(pointLs[0], scalarLs[0])
		for j := 1; j < len; j++ {
			tmp := new(bn256.G2).ScalarMult(pointLs[j], scalarLs[j])
			resPrime.Add(resPrime, tmp)
		}
	}
}

func BenchmarkMultiScalarMultG2(b *testing.B) {
	len := 64

	pointLs := make([]*bn256.G2, len)
	scalarLs := make([]*big.Int, len)

	for j := 0; j < len; j++ {
		x, X, _ := bn256.RandomG2(rand.Reader)
		pointLs[j] = new(bn256.G2).Set(X)
		scalarLs[j] = new(big.Int).Set(x)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		MultiScalarMultG2(pointLs, scalarLs)
	}
}

func BenchmarkMultiScalarMultG1(b *testing.B) {
	len := 64

	pointLs := make([]*bn256.G1, len)
	scalarLs := make([]*big.Int, len)

	for j := 0; j < len; j++ {
		x, X, _ := bn256.RandomG1(rand.Reader)
		pointLs[j] = new(bn256.G1).Set(X)
		scalarLs[j] = new(big.Int).Set(x)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		MultiScalarMultG1(pointLs, scalarLs)
	}
}
