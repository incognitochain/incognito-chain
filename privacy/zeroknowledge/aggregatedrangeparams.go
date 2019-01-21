package zkp

import (
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/privacy"
	"github.com/pkg/errors"
	"math/big"
)

/***** Bullet proof params *****/

// BulletproofParams includes all generator for aggregated range proof
type BulletproofParams struct {
	G []*privacy.EllipticPoint
	H []*privacy.EllipticPoint
}

func setBulletproofParams() BulletproofParams {
	var gen BulletproofParams
	const capacity = 64 // fixed value
	gen.G = make([]*privacy.EllipticPoint, capacity, capacity)
	gen.H = make([]*privacy.EllipticPoint, capacity, capacity)

	for i := 0; i < capacity; i++ {
		gen.G[i] = new(privacy.EllipticPoint)
		gen.G[i].Set(privacy.PubParams.G[5 + i].X, privacy.PubParams.G[5 + i].Y)

		gen.H[i] = new(privacy.EllipticPoint)
		gen.H[i].Set(privacy.PubParams.G[69 + i].X, privacy.PubParams.G[69 + i].Y)
	}
	return gen
}

var AggParam = setBulletproofParams()

// CommitAll commits a list of PCM_CAPACITY value(s)
func EncodeVectors(a []*big.Int, b []*big.Int, g[]*privacy.EllipticPoint, h[]*privacy.EllipticPoint) (*privacy.EllipticPoint, error) {
	if len(a) != len(b) || len(g) != len(h) || len(a)!= len(g){
		return nil, errors.New("invalid input")
	}

	res := new(privacy.EllipticPoint).Zero()
	for i := 0; i < len(a); i++ {
		res = res.Add(g[i].ScalarMult(a[i]).Add(h[i].ScalarMult(b[i])))
	}
	return res, nil
}

func generateChallengeForAggRange(values []*privacy.EllipticPoint) *big.Int {
	bytes := AggParam.G[0].Compress()
	for i := 1; i < len(AggParam.G); i++ {
		bytes = append(bytes, AggParam.G[i].Compress()...)
	}

	for i := 0; i < len(AggParam.H); i++ {
		bytes = append(bytes, AggParam.H[i].Compress()...)
	}

	for i := 0; i < len(values); i++ {
		bytes = append(bytes, values[i].Compress()...)
	}

	hash := common.HashB(bytes)

	res := new(big.Int).SetBytes(hash)
	res.Mod(res, privacy.Curve.Params().N)
	return res
}

func generateChallengeForAggRangeFromBytes(values [][]byte) *big.Int {
	bytes := AggParam.G[0].Compress()
	for i := 1; i < len(AggParam.G); i++ {
		bytes = append(bytes, AggParam.G[i].Compress()...)
	}

	for i := 0; i < len(AggParam.H); i++ {
		bytes = append(bytes, AggParam.H[i].Compress()...)
	}

	for i := 0; i < len(values); i++ {
		bytes = append(bytes, values[i]...)
	}

	hash := common.HashB(bytes)

	res := new(big.Int).SetBytes(hash)
	res.Mod(res, privacy.Curve.Params().N)
	return res
}

//func generateChallengeForAggRangeFromBigInt(values []*big.Int) *big.Int {
//	bytes := AggParam.G[0].Compress()
//	for i := 1; i < len(AggParam.G); i++ {
//		bytes = append(bytes, AggParam.G[i].Compress()...)
//	}
//
//	for i := 0; i < len(AggParam.H); i++ {
//		bytes = append(bytes, AggParam.H[i].Compress()...)
//	}
//
//	for i := 0; i < len(values); i++ {
//		bytes = append(bytes, values[i].Bytes()...)
//	}
//
//	hash := common.HashB(bytes)
//
//	res := new(big.Int).SetBytes(hash)
//	res.Mod(res, privacy.Curve.Params().N)
//	return res
//}
