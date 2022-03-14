package bulletproofs

import (
	"fmt"

	"github.com/incognitochain/incognito-chain/privacy/operation/v1"
	"github.com/incognitochain/incognito-chain/privacy/privacy_util"
)

// ConvertIntToBinary represents a integer number in binary
func ConvertUint64ToBinary(number uint64, n int) []*operation.Scalar {
	if number == 0 {
		res := make([]*operation.Scalar, n)
		for i := 0; i < n; i++ {
			res[i] = new(operation.Scalar).FromUint64(0)
		}
		return res
	}

	binary := make([]*operation.Scalar, n)

	for i := 0; i < n; i++ {
		binary[i] = new(operation.Scalar).FromUint64(number % 2)
		number /= 2
	}
	return binary
}

//nolint:gocritic // This function uses capitalized variable name
func computeHPrime(y *operation.Scalar, N int, H []*operation.Point) []*operation.Point {
	yInverse := new(operation.Scalar).Invert(y)
	HPrime := make([]*operation.Point, N)
	expyInverse := new(operation.Scalar).FromUint64(1)
	for i := 0; i < N; i++ {
		HPrime[i] = new(operation.Point).ScalarMult(H[i], expyInverse)
		expyInverse.Mul(expyInverse, yInverse)
	}
	return HPrime
}

//nolint:gocritic // This function uses capitalized variable name
func computeDeltaYZ(z, zSquare *operation.Scalar, yVector []*operation.Scalar, N int) (*operation.Scalar, error) {
	oneNumber := new(operation.Scalar).FromUint64(1)
	twoNumber := new(operation.Scalar).FromUint64(2)
	oneVectorN := powerVector(oneNumber, privacy_util.MaxExp)
	twoVectorN := powerVector(twoNumber, privacy_util.MaxExp)
	oneVector := powerVector(oneNumber, N)

	deltaYZ := new(operation.Scalar).Sub(z, zSquare)
	// ip1 = <1^(n*m), y^(n*m)>
	var ip1, ip2 *operation.Scalar
	var err error
	if ip1, err = innerProduct(oneVector, yVector); err != nil {
		return nil, err
	} else if ip2, err = innerProduct(oneVectorN, twoVectorN); err != nil {
		return nil, err
	} else {
		deltaYZ.Mul(deltaYZ, ip1)
		sum := new(operation.Scalar).FromUint64(0)
		zTmp := new(operation.Scalar).Set(zSquare)
		for j := 0; j < int(N/privacy_util.MaxExp); j++ {
			zTmp.Mul(zTmp, z)
			sum.Add(sum, zTmp)
		}
		sum.Mul(sum, ip2)
		deltaYZ.Sub(deltaYZ, sum)
	}
	return deltaYZ, nil
}

func innerProduct(a []*operation.Scalar, b []*operation.Scalar) (*operation.Scalar, error) {
	if len(a) != len(b) {
		return nil, fmt.Errorf("incompatible sizes of a and b")
	}
	result := new(operation.Scalar).FromUint64(uint64(0))
	for i := range a {
		// res = a[i]*b[i] + res % l
		result.MulAdd(a[i], b[i], result)
	}
	return result, nil
}

func vectorAdd(a []*operation.Scalar, b []*operation.Scalar) ([]*operation.Scalar, error) {
	if len(a) != len(b) {
		return nil, fmt.Errorf("incompatible sizes of a and b")
	}
	result := make([]*operation.Scalar, len(a))
	for i := range a {
		result[i] = new(operation.Scalar).Add(a[i], b[i])
	}
	return result, nil
}

//nolint:gocritic // This function uses capitalized variable name
func setAggregateParams(N int) *bulletproofParams {
	aggParam := new(bulletproofParams)
	aggParam.g = AggParam.g[0:N]
	aggParam.h = AggParam.h[0:N]
	aggParam.u = AggParam.u
	aggParam.cs = AggParam.cs
	return aggParam
}

func roundUpPowTwo(v int) int {
	if v == 0 {
		return 1
	}
	v--
	v |= v >> 1
	v |= v >> 2
	v |= v >> 4
	v |= v >> 8
	v |= v >> 16
	v++
	return v
}

func hadamardProduct(a []*operation.Scalar, b []*operation.Scalar) ([]*operation.Scalar, error) {
	if len(a) != len(b) {
		return nil, fmt.Errorf("invalid input")
	}
	result := make([]*operation.Scalar, len(a))
	for i := 0; i < len(result); i++ {
		result[i] = new(operation.Scalar).Mul(a[i], b[i])
	}
	return result, nil
}

// powerVector calculates base^n
func powerVector(base *operation.Scalar, n int) []*operation.Scalar {
	result := make([]*operation.Scalar, n)
	result[0] = new(operation.Scalar).FromUint64(1)
	if n > 1 {
		result[1] = new(operation.Scalar).Set(base)
		for i := 2; i < n; i++ {
			result[i] = new(operation.Scalar).Mul(result[i-1], base)
		}
	}
	return result
}

// vectorAddScalar adds a vector to a big int, returns big int array
func vectorAddScalar(v []*operation.Scalar, s *operation.Scalar) []*operation.Scalar {
	result := make([]*operation.Scalar, len(v))
	for i := range v {
		result[i] = new(operation.Scalar).Add(v[i], s)
	}
	return result
}

// vectorMulScalar mul a vector to a big int, returns a vector
func vectorMulScalar(v []*operation.Scalar, s *operation.Scalar) []*operation.Scalar {
	result := make([]*operation.Scalar, len(v))
	for i := range v {
		result[i] = new(operation.Scalar).Mul(v[i], s)
	}
	return result
}

// CommitAll commits a list of PCM_CAPACITY value(s)
func encodeVectors(l []*operation.Scalar, r []*operation.Scalar, g []*operation.Point, h []*operation.Point) (*operation.Point, error) {
	if len(l) != len(r) || len(g) != len(l) || len(h) != len(g) {
		return nil, fmt.Errorf("invalid input")
	}
	tmp1 := new(operation.Point).MultiScalarMult(l, g)
	tmp2 := new(operation.Point).MultiScalarMult(r, h)
	res := new(operation.Point).Add(tmp1, tmp2)
	return res, nil
}

// bulletproofParams includes all generator for aggregated range proof
func newBulletproofParams(m int) *bulletproofParams {
	maxExp := privacy_util.MaxExp
	numCommitValue := privacy_util.NumBase
	maxOutputCoin := privacy_util.MaxOutputCoin
	capacity := maxExp * m // fixed value
	param := new(bulletproofParams)
	param.g = make([]*operation.Point, capacity)
	param.h = make([]*operation.Point, capacity)
	csByte := []byte{}

	for i := 0; i < capacity; i++ {
		param.g[i] = operation.HashToPointFromIndex(int64(numCommitValue+i), operation.CStringBulletProof)
		param.h[i] = operation.HashToPointFromIndex(int64(numCommitValue+i+maxOutputCoin*maxExp), operation.CStringBulletProof)
		csByte = append(csByte, param.g[i].ToBytesS()...)
		csByte = append(csByte, param.h[i].ToBytesS()...)
	}

	param.u = new(operation.Point)
	param.u = operation.HashToPointFromIndex(int64(numCommitValue+2*maxOutputCoin*maxExp), operation.CStringBulletProof)
	csByte = append(csByte, param.u.ToBytesS()...)

	param.cs = operation.HashToPoint(csByte)
	return param
}

func generateChallenge(hashCache []byte, values []*operation.Point) *operation.Scalar {
	bytes := []byte{}
	bytes = append(bytes, hashCache...)
	for i := 0; i < len(values); i++ {
		bytes = append(bytes, values[i].ToBytesS()...)
	}
	hash := operation.HashToScalar(bytes)
	return hash
}
