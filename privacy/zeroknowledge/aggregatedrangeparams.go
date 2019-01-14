package zkp

import (
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
func (param BulletproofParams) EncodeVectors(a []*big.Int, b []*big.Int) (*privacy.EllipticPoint, error) {
	if len(a) != len(b) || len(a) != len(param.G) {
		return nil, errors.New("invalid input")
	}

	res := new(privacy.EllipticPoint).Zero()
	for i := 0; i < len(param.G); i++ {
		res = res.Add(param.G[i].ScalarMult(a[i]).Add(param.H[i].ScalarMult(b[i])))
	}
	return res, nil
}
