package privacy

import (
	"crypto/elliptic"
	"fmt"
	"math/big"
)

// Curve P256
var Curve = elliptic.P256()

//EllipticPointHelper contain some function for elliptic point
type EllipticPointHelper interface {
	InversePoint() (*EllipticPoint, error)
	RandSecp256k1Point(x, y *big.Int) *EllipticPoint
	IsSafe() bool
}

// EllipticPoint represents an point of elliptic curve
type EllipticPoint struct {
	X, Y *big.Int
}

//InversePoint return inverse point of ECC Point input
func (eccPoint EllipticPoint) InversePoint() (*EllipticPoint, error) {
	//Check that input is ECC point
	if !Curve.IsOnCurve(eccPoint.X, eccPoint.Y) {
		return nil, fmt.Errorf("Input is not ECC Point")
	}
	//Create result point
	resPoint := new(EllipticPoint)
	resPoint.X = big.NewInt(0)
	resPoint.Y = big.NewInt(0)

	//inverse point of A(x,y) in ECC is A'(x, P - y) with P is order of Curve
	resPoint.X.SetBytes(eccPoint.X.Bytes())
	resPoint.Y.SetBytes(eccPoint.Y.Bytes())
	resPoint.Y.Sub(Curve.Params().P, resPoint.Y)

	return resPoint, nil
}

//RandSecp256k1Point return pseudorandom point calculated by function F(x,y) = (x + y)^-1 * G, with G is genertor in Secp256k1
func (eccPoint *EllipticPoint) RandSecp256k1Point(x, y *big.Int) *EllipticPoint {
	//Generate result point
	res := new(EllipticPoint)
	res.X = big.NewInt(0)
	res.Y = big.NewInt(0)
	//res <- G, we'll calculate res * (x + y)^-1 after.
	res.X.SetBytes(Curve.Params().Gx.Bytes())
	res.Y.SetBytes(Curve.Params().Gy.Bytes())

	//intTemp is x + y (mod N)
	intTemp := big.NewInt(0)
	intTemp.SetBytes(x.Bytes())
	intTemp.Add(intTemp, y)
	intTemp.Mod(intTemp, Curve.Params().N)

	//intTempInverse is (x + y)^-1 in (Zn)*
	//Cuz intTempInverse * intTemp = 1 (int (Zn)*), so we use Euclid Theorem: a*x + b*y = GCD(a,b)
	//in this case, we calculate GCD(intTemp,n) = intTemp*x + n*y
	//if GCD(intTemp,n) = 1, we have intTemp*x + n*y = 1 (*)
	//Mod (*) by n, we have intTemp*x = 1 (mod n)
	//so intTempInverse = x
	intTempInverse := big.NewInt(0)
	intY := big.NewInt(0)
	intGCD := big.NewInt(0)
	intGCD = intGCD.GCD(intTempInverse, intY, intTemp, Curve.Params().N)

	if intGCD.Cmp(big.NewInt(1)) != 0 {
		//if GCD return value != 1, it mean we dont have (x + y)^-1 in Zn
		return nil
	}

	//res = res * (x+y)^-1
	res.X, res.Y = Curve.ScalarMult(res.X, res.Y, intTempInverse.Bytes())

	if eccPoint.X == nil {
		eccPoint = res
	} else {
		eccPoint.X = res.X
		eccPoint.Y = res.Y
	}
	return res
}

//IsSafe return true if eccPoint*eccPoint is not at infinity
func (eccPoint EllipticPoint) IsSafe() bool {
	var res EllipticPoint
	res.X, res.Y = Curve.Double(eccPoint.X, eccPoint.Y)
	if (res.X == nil) || (res.Y == nil) {
		return false
	}
	return true
}
