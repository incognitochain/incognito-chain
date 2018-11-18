package privacy

import (
	"crypto/elliptic"
	"fmt"
	"math/big"
)

// Curve P256
var Curve = elliptic.P256()

const (
	PointBytesLenCompressed int  = 33
	PointCompressed         byte = 0
)

//EllipticPointHelper contain some function for elliptic point
type EllipticPointHelper interface {
	InversePoint() (*EllipticPoint, error)
	RandPoint(x, y *big.Int) *EllipticPoint
	CompressPoint() []byte
	DecompressPoint(x *big.Int, ybit bool) (*big.Int, error)
	IsSafe() bool
}

// EllipticPoint represents an point of elliptic curve
type EllipticPoint struct {
	X, Y *big.Int
}

// InversePoint return inverse point of ECC Point input
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
	//resPoint.X.SetBytes(eccPoint.X.Bytes())
	//resPoint.Y.SetBytes(eccPoint.Y.Bytes())

	*(resPoint.X) = *(eccPoint.X)
	*(resPoint.Y) = *(eccPoint.Y)

	resPoint.Y.Sub(Curve.Params().P, resPoint.Y)
	resPoint.Y.Mod(resPoint.Y, Curve.Params().P)

	return resPoint, nil
}

// RandPoint return pseudorandom point calculated by function F(x,y) = (x + y)^-1 * G, with G is genertor in Secp256k1
func (eccPoint *EllipticPoint) RandPoint(x, y *big.Int) *EllipticPoint {
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

// IsSafe return true if eccPoint*eccPoint is not at infinity
func (eccPoint EllipticPoint) IsSafe() bool {
	var res EllipticPoint
	res.X, res.Y = Curve.Double(eccPoint.X, eccPoint.Y)
	if (res.X == nil) || (res.Y == nil) {
		return false
	}
	return true
}

// CompressPoint compresses key from 64 bytes to PointBytesLenCompressed bytes
func (eccPoint EllipticPoint) CompressPoint() []byte {
	if Curve.IsOnCurve(eccPoint.X, eccPoint.Y) {
		b := make([]byte, 0, PointBytesLenCompressed)
		format := PointCompressed
		if isOdd(eccPoint.Y) {
			format |= 0x1
		}
		b = append(b, format)
		return paddedAppend(32, b, eccPoint.X.Bytes())
	}
	return nil
}

// DecompressPoint decompresses a byte array, which was created by CompressPoint func,
// to a point on the given curve.
func (eccPoint *EllipticPoint) DecompressPoint(compressPointBytes []byte) error {
	format := compressPointBytes[0]
	format &= ^byte(0x1)

	if format != PointCompressed {
		return fmt.Errorf("invalid magic in compressed "+
			"compressPoint bytes: %d", compressPointBytes[0])
	}
	ybit := (format & 0x1) == 0x1
	var err error
	if eccPoint.X == nil {
		eccPoint.X = new(big.Int).SetBytes(compressPointBytes[1:33])
	} else {
		eccPoint.X.SetBytes(compressPointBytes[1:33])
	}
	eccPoint.Y, err = DecompressPoint(eccPoint.X, ybit)
	return err
}

// DecompressPoint decompresses a point on the given curve given the X point and
// the solution to use.
func DecompressPoint(x *big.Int, ybit bool) (*big.Int, error) {
	Q := Curve.Params().P
	temp := new(big.Int)
	xTemp := new(big.Int)

	// Y = +-sqrt(x^3 - 3*x + B)
	xCube := new(big.Int).Mul(x, x)
	xCube.Mul(xCube, x)
	xCube.Add(xCube, Curve.Params().B)
	xCube.Sub(xCube, xTemp.Mul(x, new(big.Int).SetInt64(3)))
	xCube.Mod(xCube, Curve.Params().P)

	//check P = 3 mod 4?
	if temp.Mod(Q, new(big.Int).SetInt64(4)).Cmp(new(big.Int).SetInt64(3)) != 0 {
		return nil, fmt.Errorf("parameter P must be congruent to 3 mod 4")
	}

	// Now calculate sqrt mod p of x^3 - 3*x + B
	// This code used to do a full sqrt based on tonelli/shanks,
	// but this was replaced by the algorithms referenced in
	// https://bitcointalk.org/index.php?topic=162805.msg1712294#msg1712294
	y := new(big.Int).Exp(xCube, PAdd1Div4(Q), Q)

	if ybit != isOdd(y) {
		y.Sub(Curve.Params().P, y)
	}

	// Check that y is a square root of x^3  - 3*x + B.
	ySquare := new(big.Int).Mul(y, y)
	ySquare.Mod(ySquare, Curve.Params().P)
	if ySquare.Cmp(xCube) != 0 {
		return nil, fmt.Errorf("invalid square root")
	}

	//fmt.Println(Curve.IsOnCurve(x, y))

	// Verify that y-coord has expected parity.
	if ybit != isOdd(y) {
		return nil, fmt.Errorf("ybit doesn't match oddness")
	}

	return y, nil
}
