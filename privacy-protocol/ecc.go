package privacy

import (
	"crypto/elliptic"
	"math/big"

	"github.com/ninjadotorg/constant/common"

	"github.com/pkg/errors"
	"encoding/json"
	"github.com/ninjadotorg/constant/common/base58"
)

// Curve P256
// We only use P256 Curve in our protocol
var Curve = elliptic.P256()

//EllipticPointHelper contain some function for elliptic point
type EllipticPointHelper interface {
	// <0xakk0r0kamui>
	Inverse() (*EllipticPoint, error)
	Randomize()
	Compress() []byte
	Decompress(compressPointBytes []byte) error
	IsSafe() bool
	ComputeYCoord()
	Hash() *EllipticPoint
	Set(x, y *big.Int)

	// </0xakk0r0kamui>

	// <PTD>
	AddPoint(EllipticPoint) *EllipticPoint
	ScalarMulPoint(*big.Int) *EllipticPoint
	IsEqual(EllipticPoint) bool
	// </PTD>
}

// EllipticPoint represents an point of elliptic curve,
// which contains X, Y. X is Abscissa, Y is Ordinate
type EllipticPoint struct {
	X, Y *big.Int
}

// Zero initializes elliptic point with X = 0, Y = 0
func (point *EllipticPoint) Zero() *EllipticPoint{
	point.X = new(big.Int).SetInt64(0)
	point.Y = new(big.Int).SetInt64(0)
	return point
}

// UnmarshalJSON unmarshal from byte array to elliptic point
func (point *EllipticPoint) UnmarshalJSON(data []byte) error {
	dataStr := ""
	_ = json.Unmarshal(data, &dataStr)
	temp, _, err := base58.Base58Check{}.Decode(dataStr)
	if err != nil {
		return err
	}
	point.Decompress(temp)
	return nil
}

// MarshalJSON marshal from elliptic point to byte array
func (point EllipticPoint) MarshalJSON() ([]byte, error) {
	data := point.Compress()
	temp := base58.Base58Check{}.Encode(data, byte(0x00))
	return json.Marshal(temp)
}

//ComputeYCoord calculates Y coord from X
func (point *EllipticPoint) ComputeYCoord() error {
	if point.Y == nil {
		point.Y = big.NewInt(0)
	}
	xTemp := new(big.Int)
	// Y = +-sqrt(x^3 - 3*x + B)
	x3 := new(big.Int).Mul(point.X, point.X)
	x3.Mul(x3, point.X)
	x3.Add(x3, Curve.Params().B)
	x3.Sub(x3, xTemp.Mul(point.X, big.NewInt(3)))
	x3.Mod(x3, Curve.Params().P)

	// //check P = 3 mod 4?
	// if temp.Mod(Q, new(big.Int).SetInt64(4)).Cmp(new(big.Int).SetInt64(3)) == 0 {
	// 	//		fmt.Println("Ok!!!")
	// }

	// Now calculate sqrt mod p of x^3 - 3*x + B
	// This code used to do a full sqrt based on tonelli/shanks,
	// but this was replaced by the algorithms referenced in
	// https://bitcointalk.org/index.php?topic=162805.msg1712294#msg1712294
	point.Y = new(big.Int).Exp(x3, PAdd1Div4(Curve.Params().P), Curve.Params().P)
	// Check that y is a square root of x^3  - 3*x + B.
	y2 := new(big.Int).Mul(point.Y, point.Y)
	y2.Mod(y2, Curve.Params().P)
	if y2.Cmp(x3) != 0 {
		return errors.New("Cant compute y")
	}
	return nil
}

// Inverse return inverse point of ECC Point input
func (point EllipticPoint) Inverse() (*EllipticPoint, error) {
	//Check that input is ECC point
	if !Curve.IsOnCurve(point.X, point.Y) {
		return nil, errors.New("Input is not ECC Point")
	}

	//Create result point
	resPoint := new(EllipticPoint)
	resPoint.X = big.NewInt(0)
	resPoint.Y = big.NewInt(0)

	//inverse point of A(x,y) in ECC is A'(x, P - y) with P is order of Curve
	resPoint.X.Set(point.X)
	resPoint.Y.Set(point.Y)
	resPoint.Y.Sub(Curve.Params().P, resPoint.Y)
	resPoint.Y.Mod(resPoint.Y, Curve.Params().P)

	return resPoint, nil
}

// Randomize make object's value to random
func (point *EllipticPoint) Randomize() {
	if point.X == nil {
		point.X = big.NewInt(0)
	}
	if point.Y == nil {
		point.Y = big.NewInt(0)
	}

	for {
		point.X.SetBytes(RandBytes(32))
		err := point.ComputeYCoord()
		if Curve.IsOnCurve(point.X, point.Y) && (err == nil) && (point.IsSafe()) {
			break
		}
	}

	//Logger.Log.Infof("Privacy log: Randomize elliptic point: %+v\n", point)
}

// IsSafe return true if eccPoint*eccPoint is not at infinity
func (point EllipticPoint) IsSafe() bool {
	var res EllipticPoint
	if !Curve.IsOnCurve(point.X, point.Y) {
		return false
	}
	res.X, res.Y = Curve.Double(point.X, point.Y)
	if res.X.Cmp(big.NewInt(0)) == 0 && res.Y.Cmp(big.NewInt(0)) == 0 {
		return false
	}
	return true
}

// Compress compresses key from 64 bytes to PointBytesLenCompressed bytes
func (point EllipticPoint) Compress() []byte {
	if Curve.IsOnCurve(point.X, point.Y) {
		b := make([]byte, 0, CompressedPointSize)
		format := PointCompressed
		if isOdd(point.Y) {
			format |= 0x1
		}
		b = append(b, format)
		return paddedAppend(32, b, point.X.Bytes())
	}
	return nil
}

// Decompress decompresses a byte array, which was created by CompressPoint func,
// to a point on the given curve.
func (point *EllipticPoint) Decompress(compressPointBytes []byte) error {
	format := compressPointBytes[0]
	ybit := (format & 0x1) == 0x1
	format &= ^byte(0x1)

	if format != PointCompressed {
		return errors.New("invalid magic in compressed compressPoint bytes")
	}
	var err error
	if point.X == nil {
		point.X = new(big.Int).SetBytes(compressPointBytes[1:33])
	} else {
		point.X.SetBytes(compressPointBytes[1:33])
	}
	point.Y, err = decompPoint(point.X, ybit)
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
	// 	return nil, errors.Zero("parameter P must be congruent to 3 mod 4")
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
		return nil, errors.New("invalid square root")
	}

	// Verify that y-coord has expected parity.
	if ybit != isOdd(y) {
		return nil, errors.New("ybit doesn't match oddness")
	}

	return y, nil
}

// Hash derives new elliptic point from another elliptic point using hash function
func (point EllipticPoint) Hash(index int) *EllipticPoint {
	// res.X = hash(g.X || index), res.Y = sqrt(res.X^3 - 3X + B)
	var res = new(EllipticPoint)
	res.X = big.NewInt(0)
	res.Y = big.NewInt(0)
	res.X.SetBytes(point.X.Bytes())
	res.X.Add(res.X, big.NewInt(int64(index)))
	for {
		res.X.SetBytes(common.DoubleHashB(res.X.Bytes()))
		res.ComputeYCoord()
		if (res.Y != nil) && (Curve.IsOnCurve(res.X, res.Y)) && (res.IsSafe()) {
			break
		}
	}
	return res
}

func (point * EllipticPoint) Set(x, y *big.Int) {
	point.X.Set(x)
	point.Y.Set(y)
}

func TestECC() bool {
	// //Test compress && decompress
	// eccPoint := new(EllipticPoint)
	// eccPoint.Randomize()
	// neg := 0
	// for i := 0; i < 2000; i++ {
	// 	eccPoint1 := new(EllipticPoint)
	// 	eccPoint1.Randomize()
	// 	eccPoint2 := new(EllipticPoint)
	// 	eccPoint2.Randomize()
	// 	eccPointX := new(EllipticPoint)
	// 	// eccPointX.Randomize()
	// 	start := time.Now()
	// 	eccPointX.X, eccPointX.Y = Curve.Add(eccPoint1.X, eccPoint1.Y, eccPoint2.X, eccPoint2.Y)
	// 	end := time.Now()
	// 	time1 := end.Sub(start)
	// 	start = time.Now()
	// 	eccPointX.X = big.NewInt(0)
	// 	eccPointX.Y = big.NewInt(0)
	// 	*eccPointX = (*eccPoint1).Add(*eccPoint2)
	// 	end = time.Now()
	// 	time2 := end.Sub(start)
	// 	// fmt.Printf("%v %v \n", time1, time2)
	// 	if time1 > time2 {
	// 		neg++
	// 	}
	// }
	// fmt.Println(neg)

	// if !Curve.IsOnCurve(eccPoint.X, eccPoint.Y) {
	// 	return false
	// }
	// fmt.Printf("On curve!")
	// if !eccPoint.IsSafe() {
	// 	return false
	// }
	// fmt.Printf("Safe!")
	// compressBytes := eccPoint.Compress()
	// eccPointDecompressed := new(EllipticPoint)
	// err := eccPointDecompressed.Decompress(compressBytes)
	// if err != nil {
	// 	return false
	// }
	return true
}

func (point EllipticPoint) Add(targetPoint *EllipticPoint) *EllipticPoint {
	res := new(EllipticPoint)
	res.X, res.Y = Curve.Add(point.X, point.Y, targetPoint.X, targetPoint.Y)
	return res
}

func (point EllipticPoint) Sub(targetPoint *EllipticPoint) *EllipticPoint {
	invPoint, err := targetPoint.Inverse()

	if err != nil {
		return nil
	}

	res := new(EllipticPoint).Zero()
	res = point.Add(invPoint)
	return res
}

func (point EllipticPoint) IsEqual(p *EllipticPoint) bool {
	if point.X.Cmp(p.X) == 0 && point.Y.Cmp(p.Y) == 0 {
		return true
	}
	return false
}
func (point EllipticPoint) ScalarMult(factor *big.Int) *EllipticPoint {
	res := new(EllipticPoint)
	res.X, res.Y = Curve.ScalarMult(point.X, point.Y, factor.Bytes())
	return res
}

/*******************************************************************************************/
