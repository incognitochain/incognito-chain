package aggregaterange

import (
	"errors"
	"math"
	"sync"

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
func vectorAdd(a []*privacy.Scalar, b []*privacy.Scalar) ([]*privacy.Scalar, error) {
	if len(a) != len(b) {
		return nil, errors.New("VectorAdd: Arrays not of the same length")
	}

	res := make([]*privacy.Scalar, len(a))
	var wg sync.WaitGroup
	wg.Add(len(a))
	for i := range a {
		go func(i int, wg *sync.WaitGroup) {
			defer wg.Done()
			res[i] = new(privacy.Scalar).Add(a[i], b[i])
		}(i, &wg)
	}
	wg.Wait()
	return res, nil
}

// innerProduct calculates inner product between two vectors a and b
func innerProduct(a []*privacy.Scalar, b []*privacy.Scalar) (*privacy.Scalar, error) {
	if len(a) != len(b) {
		return nil, errors.New("InnerProduct: Arrays not of the same length")
	}

	res := new(privacy.Scalar).SetUint64(uint64(0))
	tmp := new(privacy.Scalar)

	for i := range a {
		res.Add(res, tmp.Mul(a[i], b[i]))
	}

	return res, nil
}

// hadamardProduct calculates hadamard product between two vectors a and b
func hadamardProduct(a []*privacy.Scalar, b []*privacy.Scalar) ([]*privacy.Scalar, error) {
	if len(a) != len(b) {
		return nil, errors.New("InnerProduct: Arrays not of the same length")
	}

	res := make([]*privacy.Scalar, len(a))
	var wg sync.WaitGroup
	wg.Add(len(a))
	for i := 0; i < len(res); i++ {
		go func(i int, wg *sync.WaitGroup) {
			defer wg.Done()
			res[i] = new(privacy.Scalar).Mul(a[i], b[i])
		}(i, &wg)
	}
	wg.Wait()

	return res, nil
}

// powerVector calculates base^n
func powerVector(base *privacy.Scalar, n int) []*privacy.Scalar {
	res := make([]*privacy.Scalar, n)
	res[0] = new(privacy.Scalar).SetUint64(1)

	var wg sync.WaitGroup
	wg.Add(n - 1)
	for i := 1; i < n; i++ {
		go func(i int, wg *sync.WaitGroup) {
			defer wg.Done()
			res[i] = new(privacy.Scalar).Exp(base, new(privacy.Scalar).SetUint64(uint64(i)))
		}(i, &wg)
	}
	wg.Wait()
	return res
}

// vectorAddScalar adds a vector to a big int, returns big int array
func vectorAddScalar(v []*privacy.Scalar, s *privacy.Scalar) []*privacy.Scalar {
	res := make([]*privacy.Scalar, len(v))

	var wg sync.WaitGroup
	wg.Add(len(v))
	for i := range v {
		go func(i int, wg *sync.WaitGroup) {
			defer wg.Done()
			res[i] = new(privacy.Scalar).Add(v[i], s)
		}(i, &wg)
	}
	wg.Wait()
	return res
}

// vectorMulScalar mul a vector to a big int, returns a vector
func vectorMulScalar(v []*privacy.Scalar, s *privacy.Scalar) []*privacy.Scalar {
	res := make([]*privacy.Scalar, len(v))

	var wg sync.WaitGroup
	wg.Add(len(v))
	for i := range v {
		go func(i int, wg *sync.WaitGroup) {
			defer wg.Done()
			res[i] = new(privacy.Scalar).Mul(v[i], s)
		}(i, &wg)
	}
	wg.Wait()
	return res
}

// estimateMultiRangeProofSize estimate multi range proof size
func EstimateMultiRangeProofSize(nOutput int) uint64 {
	return uint64((nOutput+2*int(math.Log2(float64(maxExp*pad(nOutput))))+5)*privacy.Ed25519KeySize + 5*privacy.Ed25519KeySize + 2)
}

// CommitAll commits a list of PCM_CAPACITY value(s)
func encodeVectors(l []*privacy.Scalar, r []*privacy.Scalar, g []*privacy.Point, h []*privacy.Point) (*privacy.Point, error) {
	if len(l) != len(r) || len(g) != len(l) || len(h) != len(g) {
		return nil, errors.New("invalid input")
	}

	res := new(privacy.Point)
	res.Zero()
	var wg sync.WaitGroup
	var tmp1, tmp2 *privacy.Point

	for i := 0; i < len(l); i++ {
		wg.Add(2)
		go func(i int, wg *sync.WaitGroup) {
			defer wg.Done()
			tmp1 = new(privacy.Point).ScalarMult(g[i], l[i])
		}(i, &wg)

		go func(i int, wg *sync.WaitGroup) {
			defer wg.Done()
			tmp2 = new(privacy.Point).ScalarMult(h[i], r[i])
		}(i, &wg)

		wg.Wait()

		res = res.Add(res, tmp1)
		res = res.Add(res, tmp2)
	}
	return res, nil
}
