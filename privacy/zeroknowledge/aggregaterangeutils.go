package zkp

import (
	"errors"
	"math"
	"math/big"
	"sync"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/privacy"
)

// pad returns number has format 2^k that it is the nearest number to num
func pad(num int) int {
	if num == 1 || num == 2 {
		return num
	}
	tmp := 2
	for i := 2; ; i++ {
		tmp *= 2
		if tmp >= num {
			num = tmp
			break
		}
	}
	return num
}

/*-----------------------------Vector Functions-----------------------------*/
// The length here always has to be a power of two

//vectorAdd adds two vector and returns result vector
func vectorAdd(a []*big.Int, b []*big.Int) ([]*big.Int, error) {
	if len(a) != len(b) {
		return nil, errors.New("VectorAdd: Arrays not of the same length")
	}

	res := make([]*big.Int, len(a))
	var wg sync.WaitGroup
	wg.Add(len(a))
	for i := range a {
		go func(i int, wg *sync.WaitGroup) {
			defer wg.Done()
			res[i] = new(big.Int).Add(a[i], b[i])
			res[i] = res[i].Mod(res[i], privacy.Curve.Params().N)
		}(i, &wg)
	}
	wg.Wait()
	return res, nil
}

//vectorAdd adds two vector and returns result vector
func vectorSub(a []*big.Int, b []*big.Int) ([]*big.Int, error) {
	if len(a) != len(b) {
		return nil, errors.New("VectorSub: Arrays not of the same length")
	}

	res := make([]*big.Int, len(a))
	for i := range a {
		res[i] = new(big.Int).Sub(a[i], b[i])
		res[i].Mod(res[i], privacy.Curve.Params().N)
	}
	return res, nil
}

// innerProduct calculates inner product between two vectors a and b
func innerProduct(a []*big.Int, b []*big.Int) (*big.Int, error) {
	if len(a) != len(b) {
		return nil, errors.New("InnerProduct: Arrays not of the same length")
	}

	res := big.NewInt(0)
	tmp := new(big.Int)

	for i := range a {
		res.Add(res, tmp.Mul(a[i], b[i]))
		res.Mod(res, privacy.Curve.Params().N)
	}

	return res, nil
}

// hadamardProduct calculates hadamard product between two vectors a and b
func hadamardProduct(a []*big.Int, b []*big.Int) ([]*big.Int, error) {
	if len(a) != len(b) {
		return nil, errors.New("InnerProduct: Arrays not of the same length")
	}

	res := make([]*big.Int, len(a))
	var wg sync.WaitGroup
	wg.Add(len(a))
	for i := 0; i < len(res); i++ {
		go func(i int, wg *sync.WaitGroup) {
			defer wg.Done()
			res[i] = new(big.Int).Mul(a[i], b[i])
			res[i].Mod(res[i], privacy.Curve.Params().N)
		}(i, &wg)
	}
	wg.Wait()

	return res, nil
}

// powerVector calculates base^n
func powerVector(base *big.Int, n int) []*big.Int {
	res := make([]*big.Int, n)
	res[0] = big.NewInt(1)

	var wg sync.WaitGroup
	wg.Add(n - 1)
	for i := 1; i < n; i++ {
		go func(i int, wg *sync.WaitGroup) {
			defer wg.Done()
			res[i] = new(big.Int).Exp(base, new(big.Int).SetInt64(int64(i)), privacy.Curve.Params().N)
		}(i, &wg)
	}
	wg.Wait()
	return res
}

// vectorAddScalar adds a vector to a big int, returns big int array
func vectorAddScalar(v []*big.Int, s *big.Int) []*big.Int {
	res := make([]*big.Int, len(v))

	var wg sync.WaitGroup
	wg.Add(len(v))
	for i := range v {
		go func(i int, wg *sync.WaitGroup) {
			defer wg.Done()
			res[i] = new(big.Int).Add(v[i], s)
			res[i].Mod(res[i], privacy.Curve.Params().N)
		}(i, &wg)
	}
	wg.Wait()
	return res
}

// vectorMulScalar mul a vector to a big int, returns a vector
func vectorMulScalar(v []*big.Int, s *big.Int) []*big.Int {
	res := make([]*big.Int, len(v))

	var wg sync.WaitGroup
	wg.Add(len(v))
	for i := range v {
		go func(i int, wg *sync.WaitGroup) {
			defer wg.Done()
			res[i] = new(big.Int).Mul(v[i], s)
			res[i].Mod(res[i], privacy.Curve.Params().N)
		}(i, &wg)
	}
	wg.Wait()
	return res
}

// estimateMultiRangeProofSize estimate multi range proof size
func estimateMultiRangeProofSize(nOutput int) uint64 {
	return uint64((nOutput+2*int(math.Log2(float64(maxExp*pad(nOutput))))+5)*privacy.CompressedEllipticPointSize + 5*common.BigIntSize + 2)
}

// CommitAll commits a list of PCM_CAPACITY value(s)
func encodeVectors(a []*big.Int, b []*big.Int, g []*privacy.EllipticPoint, h []*privacy.EllipticPoint) (*privacy.EllipticPoint, error) {
	if len(a) != len(b) || len(g) != len(h) || len(a) != len(g) {
		return nil, errors.New("invalid input")
	}

	res := new(privacy.EllipticPoint).Zero()
	var wg sync.WaitGroup
	var tmp1, tmp2 *privacy.EllipticPoint

	for i := 0; i < len(a); i++ {
		wg.Add(2)
		go func(i int, wg *sync.WaitGroup) {
			defer wg.Done()
			tmp1 = g[i].ScalarMult(a[i])
		}(i, &wg)

		go func(i int, wg *sync.WaitGroup) {
			defer wg.Done()
			tmp2 = h[i].ScalarMult(b[i])
		}(i, &wg)

		wg.Wait()

		res = res.Add(tmp1).Add(tmp2)
	}
	return res, nil
}
