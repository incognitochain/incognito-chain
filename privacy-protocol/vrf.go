package privacy

import (
	"math/big"
)

// Eval returns a pseudo-random elliptic curve point P = F(seed, derivator), where
// F is a pseudo-random function defined by F(x, y) = 1/(x + y)*G, where x, y are integers,
// seed and derivator are integers of size at least 32 bytes,
// G is a generating point of the group of points of an elliptic curve.
func Eval(seed, derivator *big.Int, generator *EllipticPoint) *EllipticPoint {
	// generator must be on the curve
	if !generator.IsSafe() {
		return nil
	}

	// res stores the resulting point
	res := EllipticPoint{big.NewInt(0), big.NewInt(0)}
	res.X, res.Y = Curve.ScalarMult(generator.X, generator.Y, (new(big.Int).ModInverse(new(big.Int).Add(seed, derivator), Curve.Params().N)).Bytes())

	return &res
}
