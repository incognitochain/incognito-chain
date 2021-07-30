package aggregatedrange

import (
	"errors"
	"math"

	"github.com/incognitochain/incognito-chain/privacy/operation"
)

const (
	maxExp               = 64
	numOutputParam       = 32
	maxOutputNumber      = 32
	numCommitValue       = 5
	maxOutputNumberParam = 256
)

// bulletproofParams includes all generator for aggregated range proof
type bulletproofParams struct {
	g  []*operation.Point
	h  []*operation.Point
	u  *operation.Point
	cs []byte
}

var AggParam = newBulletproofParams(numOutputParam)

func newBulletproofParams(m int) *bulletproofParams {
	gen := new(bulletproofParams)
	gen.cs = []byte{}
	capacity := maxExp * m // fixed value
	gen.g = make([]*operation.Point, capacity)
	gen.h = make([]*operation.Point, capacity)
	csByteH := []byte{}
	csByteG := []byte{}
	for i := 0; i < capacity; i++ {
		gen.g[i] = operation.HashToPointFromIndex(int64(numCommitValue+i), operation.CStringBulletProof)
		gen.h[i] = operation.HashToPointFromIndex(int64(numCommitValue+i+maxOutputNumberParam*maxExp), operation.CStringBulletProof)
		csByteG = append(csByteG, gen.g[i].ToBytesS()...)
		csByteH = append(csByteH, gen.h[i].ToBytesS()...)
	}

	gen.u = new(operation.Point)
	gen.u = operation.HashToPointFromIndex(int64(numCommitValue+2*maxOutputNumberParam*maxExp), operation.CStringBulletProof)

	gen.cs = append(gen.cs, csByteG...)
	gen.cs = append(gen.cs, csByteH...)
	gen.cs = append(gen.cs, gen.u.ToBytesS()...)

	return gen
}

func generateChallenge(values [][]byte) *operation.Scalar {
	bytes := []byte{}
	for i := 0; i < len(values); i++ {
		bytes = append(bytes, values[i]...)
	}
	hash := operation.HashToScalar(bytes)
	return hash
}

func generateChallengeOld(AggParam *bulletproofParams, values [][]byte) *operation.Scalar {
	bytes := []byte{}
	for i := 0; i < len(AggParam.g); i++ {
		bytes = append(bytes, AggParam.g[i].ToBytesS()...)
	}

	for i := 0; i < len(AggParam.h); i++ {
		bytes = append(bytes, AggParam.h[i].ToBytesS()...)
	}

	bytes = append(bytes, AggParam.u.ToBytesS()...)

	for i := 0; i < len(values); i++ {
		bytes = append(bytes, values[i]...)
	}

	hash := operation.HashToScalar(bytes)
	return hash
}

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
func vectorAdd(a []*operation.Scalar, b []*operation.Scalar) ([]*operation.Scalar, error) {
	if len(a) != len(b) {
		return nil, errors.New("VectorAdd: Arrays not of the same length")
	}

	res := make([]*operation.Scalar, len(a))
	for i := range a {
		res[i] = new(operation.Scalar).Add(a[i], b[i])
	}
	return res, nil
}

// innerProduct calculates inner product between two vectors a and b
func innerProduct(a []*operation.Scalar, b []*operation.Scalar) (*operation.Scalar, error) {
	if len(a) != len(b) {
		return nil, errors.New("InnerProduct: Arrays not of the same length")
	}
	res := new(operation.Scalar).FromUint64(uint64(0))
	for i := range a {
		//res = a[i]*b[i] + res % l
		res.MulAdd(a[i], b[i], res)
	}
	return res, nil
}

// hadamardProduct calculates hadamard product between two vectors a and b
func hadamardProduct(a []*operation.Scalar, b []*operation.Scalar) ([]*operation.Scalar, error) {
	if len(a) != len(b) {
		return nil, errors.New("InnerProduct: Arrays not of the same length")
	}

	res := make([]*operation.Scalar, len(a))
	for i := 0; i < len(res); i++ {
		res[i] = new(operation.Scalar).Mul(a[i], b[i])
	}
	return res, nil
}

// powerVector calculates base^n
func powerVector(base *operation.Scalar, n int) []*operation.Scalar {
	res := make([]*operation.Scalar, n)
	res[0] = new(operation.Scalar).FromUint64(1)
	if n > 1 {
		res[1] = new(operation.Scalar).Set(base)
		for i := 2; i < n; i++ {
			res[i] = new(operation.Scalar).Mul(res[i-1], base)
		}
	}
	return res
}

// vectorAddScalar adds a vector to a big int, returns big int array
func vectorAddScalar(v []*operation.Scalar, s *operation.Scalar) []*operation.Scalar {
	res := make([]*operation.Scalar, len(v))

	for i := range v {
		res[i] = new(operation.Scalar).Add(v[i], s)
	}
	return res
}

// vectorMulScalar mul a vector to a big int, returns a vector
func vectorMulScalar(v []*operation.Scalar, s *operation.Scalar) []*operation.Scalar {
	res := make([]*operation.Scalar, len(v))

	for i := range v {
		res[i] = new(operation.Scalar).Mul(v[i], s)
	}
	return res
}

// CommitAll commits a list of PCM_CAPACITY value(s)
func encodeVectors(l []*operation.Scalar, r []*operation.Scalar, g []*operation.Point, h []*operation.Point) (*operation.Point, error) {
	if len(l) != len(r) || len(g) != len(l) || len(h) != len(g) {
		return nil, errors.New("invalid input")
	}
	tmp1 := new(operation.Point).MultiScalarMult(l, g)
	tmp2 := new(operation.Point).MultiScalarMult(r, h)

	res := new(operation.Point).Add(tmp1, tmp2)
	return res, nil
}

// estimateMultiRangeProofSize estimate multi range proof size
func EstimateMultiRangeProofSize(nOutput int) uint64 {
	return uint64((nOutput+2*int(math.Log2(float64(maxExp*pad(nOutput))))+5)*operation.Ed25519KeySize + 5*operation.Ed25519KeySize + 2)
}
