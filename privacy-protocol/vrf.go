package privacy

import "math/big"


// PRF returns pseudorandom point calculated by function F(x,y) = (x + y)^-1 * G, with G is generator in P256
// x = seed, y = derivator
func Eval(seed, derivator *big.Int) *EllipticPoint {
	//Generate result point
	res := EllipticPoint{big.NewInt(0), big.NewInt(0)}

	//res <- G, we'll calculate res * (seed + derivator)^-1 after.
	res.X.SetBytes(Curve.Params().Gx.Bytes())
	res.Y.SetBytes(Curve.Params().Gy.Bytes())
	//res.X = Curve.Params().Gx
	//res.Y = Curve.Params().Gy

	//intTemp is seed + derivator (mod N)
	intTemp := big.NewInt(0)
	intTemp.SetBytes(seed.Bytes())
	intTemp.Add(intTemp, derivator)
	intTemp.Mod(intTemp, Curve.Params().N)

	//intTempInverse is (seed + derivator)^-1 in (Zn)*
	//Cuz intTempInverse * intTemp = 1 (int (Zn)*), so we use Euclid Theorem: a*seed + b*derivator = GCD(a,b)
	//in this case, we calculate GCD(intTemp,n) = intTemp*seed + n*derivator
	//if GCD(intTemp,n) = 1, we have intTemp*seed + n*derivator = 1 (*)
	//Mod (*) by n, we have intTemp*seed = 1 (mod n)
	//so intTempInverse = seed
	intTempInverse := big.NewInt(0)
	intY := big.NewInt(0)
	intGCD := big.NewInt(0)
	intGCD = intGCD.GCD(intTempInverse, intY, intTemp, Curve.Params().N)

	if intGCD.Cmp(big.NewInt(1)) != 0 {
		//if GCD return value != 1, it mean we dont have (seed + derivator)^-1 in Zn
		return nil
	}

	//res = res * (seed+derivator)^-1
	res.X, res.Y = Curve.ScalarMult(res.X, res.Y, intTempInverse.Bytes())

	return &res
}
