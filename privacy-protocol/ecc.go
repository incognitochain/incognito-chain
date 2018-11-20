package privacy

import (
	"crypto/elliptic"
	"fmt"
	"math/big"
)

// Curve P256
var Curve = elliptic.P256()

const (
	PointBytesLenCompressed      = 33
	PointCompressed         byte = 0x2
)

//EllipticPointHelper contain some function for elliptic point
type EllipticPointHelper interface {
	InversePoint() (*EllipticPoint, error)
	RandPoint(x, y *big.Int) *EllipticPoint
	Rand()
	CompressPoint() []byte
	DecompressPoint(compressPointBytes []byte) error
	IsSafe() bool
	ComputeYCoord(x *big.Int) *big.Int
}

// EllipticPoint represents an point of elliptic curve
type EllipticPoint struct {
	X, Y *big.Int
}

//ComputeYCoord calculates Y coord from X
func (eccPoint *EllipticPoint) ComputeYCoord() error {
	if eccPoint.Y == nil {
		eccPoint.Y = big.NewInt(0)
	}

	xTemp := new(big.Int)

	// Y = +-sqrt(x^3 - 3*x + B)
	x3 := new(big.Int).Mul(eccPoint.X, eccPoint.X)
	x3.Mul(x3, eccPoint.X)
	x3.Add(x3, Curve.Params().B)
	x3.Sub(x3, xTemp.Mul(eccPoint.X, big.NewInt(3)))
	x3.Mod(x3, Curve.Params().P)

	// //check P = 3 mod 4?
	// if temp.Mod(Q, new(big.Int).SetInt64(4)).Cmp(new(big.Int).SetInt64(3)) == 0 {
	// 	//		fmt.Println("Ok!!!")
	// }

	// Now calculate sqrt mod p of x^3 - 3*x + B
	// This code used to do a full sqrt based on tonelli/shanks,
	// but this was replaced by the algorithms referenced in
	// https://bitcointalk.org/index.php?topic=162805.msg1712294#msg1712294
	eccPoint.Y = new(big.Int).Exp(x3, PAdd1Div4(Curve.Params().P), Curve.Params().P)
	// Check that y is a square root of x^3  - 3*x + B.
	y2 := new(big.Int).Mul(eccPoint.Y, eccPoint.Y)
	y2.Mod(y2, Curve.Params().P)
	//fmt.Printf("y2: %X\n", y2)
	//fmt.Printf("x3: %X\n", x3)
	if y2.Cmp(x3) != 0 {
		return fmt.Errorf("Cant compute y")
	}
	return nil
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



//Rand make object's value to random
func (eccPoint *EllipticPoint) Rand() {
	if eccPoint.X == nil {
		eccPoint.X = big.NewInt(0)
	}
	if eccPoint.Y == nil {
		eccPoint.Y = big.NewInt(0)
	}

	for {
		eccPoint.X.SetBytes(RandBytes(32))
		err := eccPoint.ComputeYCoord()
		if Curve.IsOnCurve(eccPoint.X, eccPoint.Y) && (err == nil) && (eccPoint.IsSafe()) {
			break
		}
	}

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
	ybit := (format & 0x1) == 0x1
	format &= ^byte(0x1)

	if format != PointCompressed {
		return fmt.Errorf("invalid magic in compressed "+
			"compressPoint bytes: %d", compressPointBytes[0])
	}

	var err error
	if eccPoint.X == nil {
		eccPoint.X = new(big.Int).SetBytes(compressPointBytes[1:33])
	} else {
		eccPoint.X.SetBytes(compressPointBytes[1:33])
	}
	eccPoint.Y, err = decompPoint(eccPoint.X, ybit)
	return err
}

// DecompPoint decompresses a point on the given curve given the X point and
// the solution to use.
func decompPoint(x *big.Int, ybit bool) (*big.Int, error) {
	Q := Curve.Params().P
	// temp := new(big.Int)
	xTemp := new(big.Int)

	// Y = +-sqrt(x^3 - 3*x + B)
	xCube := new(big.Int).Mul(x, x)
	xCube.Mul(xCube, x)
	xCube.Add(xCube, Curve.Params().B)
	xCube.Mod(xCube, Curve.Params().P)
	xCube.Sub(xCube, xTemp.Mul(x, new(big.Int).SetInt64(3)))
	xCube.Mod(xCube, Curve.Params().P)

	//check P = 3 mod 4?
	// if temp.Mod(Q, new(big.Int).SetInt64(4)).Cmp(new(big.Int).SetInt64(3)) != 0 {
	// 	return nil, fmt.Errorf("parameter P must be congruent to 3 mod 4")
	// }

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

	// Verify that y-coord has expected parity.
	if ybit != isOdd(y) {
		return nil, fmt.Errorf("ybit doesn't match oddness")
	}

	return y, nil
}

func TestECC() bool {
	//Test compress && decompress
	eccPoint := new(EllipticPoint)
	eccPoint.Rand()
	if !Curve.IsOnCurve(eccPoint.X, eccPoint.Y) {
		return false
	}
	fmt.Printf("On curve!")
	if !eccPoint.IsSafe() {
		return false
	}
	fmt.Printf("Safe!")
	compressBytes := eccPoint.CompressPoint()
	eccPointDecompressed := new(EllipticPoint)
	err := eccPointDecompressed.DecompressPoint(compressBytes)
	if err != nil {
		return false
	}
	return true
}
