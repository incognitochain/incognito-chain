package zkp

import (
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/privacy"
	"github.com/pkg/errors"
	"math/big"
)

/***** Bullet proof component *****/

// BulletproofParams includes all generator for aggregated range proof
type BulletproofParams struct {
	G []*privacy.EllipticPoint
	H []*privacy.EllipticPoint
	U *privacy.EllipticPoint
}

func newBulletproofParams(m int) *BulletproofParams {
	gen := new(BulletproofParams)
	capacity := 64 * m // fixed value
	gen.G = make([]*privacy.EllipticPoint, capacity)
	gen.H = make([]*privacy.EllipticPoint, capacity)

	for i := 0; i < capacity; i++ {
		gen.G[i] = privacy.PedCom.G[0].Hash(5 + i)
		gen.H[i] = privacy.PedCom.G[0].Hash(5 + i + capacity)
	}
	gen.U = new(privacy.EllipticPoint)
	gen.U = gen.H[0].Hash(5 + 2*capacity)

	return gen
}

// CommitAll commits a list of PCM_CAPACITY value(s)
func EncodeVectors(a []*big.Int, b []*big.Int, g []*privacy.EllipticPoint, h []*privacy.EllipticPoint) (*privacy.EllipticPoint, error) {
	if len(a) != len(b) || len(g) != len(h) || len(a) != len(g) {
		return nil, errors.New("invalid input")
	}

	res := new(privacy.EllipticPoint).Zero()
	for i := 0; i < len(a); i++ {
		res = res.Add(g[i].ScalarMult(a[i])).Add(h[i].ScalarMult(b[i]))
	}
	return res, nil
}

func generateChallengeForAggRange(AggParam *BulletproofParams, values [][]byte) *big.Int {
	bytes := AggParam.G[0].Compress()
	for i := 1; i < len(AggParam.G); i++ {
		bytes = append(bytes, AggParam.G[i].Compress()...)
	}

	for i := 0; i < len(AggParam.H); i++ {
		bytes = append(bytes, AggParam.H[i].Compress()...)
	}

	bytes = append(bytes, AggParam.U.Compress()...)

	for i := 0; i < len(values); i++ {
		bytes = append(bytes, values[i]...)
	}

	hash := common.HashB(bytes)

	res := new(big.Int).SetBytes(hash)
	res.Mod(res, privacy.Curve.Params().N)
	return res
}
