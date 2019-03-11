package privacy

import (
	"crypto/elliptic"
	"math/big"

	"github.com/big0t/constant-chain/common"

	"encoding/json"

	"github.com/big0t/constant-chain/common/base58"
	"github.com/pkg/errors"
)

// The NIST curve P-256 will be used in the whole protocol
// https://csrc.nist.gov/publications/detail/fips/186/3/archive/2009-06-25
var Curve = elliptic.P256()

// EllipticPoint represents a point (X, y) on the elliptic curve
type EllipticPoint struct {
	X, Y *big.Int
}

// Zero returns the elliptic point (0, 0)
func (point *EllipticPoint) Zero() *EllipticPoint {
	point.X = big.NewInt(0)
	point.Y = big.NewInt(0)
	return point
}

// UnmarshalJSON returns an elliptic point from a byte array encoding for that point
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

// MarshalJSON returns a byte array encoding for an elliptic point
func (point EllipticPoint) MarshalJSON() ([]byte, error) {
	data := point.Compress()
	temp := base58.Base58Check{}.Encode(data, common.ZeroByte)
	return json.Marshal(temp)
}

// ComputeYCoord returns Y-coordinate from X-coordinate
func (point *EllipticPoint) ComputeYCoord() error {
	// Y = +-sqrt(x^3 - 3*x + B)
	xCube := new(big.Int).Exp(point.X, big.NewInt(3), Curve.Params().P)
	xCube.Add(xCube, Curve.Params().B)
	xCube.Sub(xCube, new(big.Int).Mul(point.X, big.NewInt(3)))
	xCube.Mod(xCube, Curve.Params().P)

	// compute sqrt(x^3 - 3*x + B) mod p
	// https://bitcointalk.org/index.php?topic=162805.msg1712294#msg1712294
	point.Y = new(big.Int).Exp(xCube, PAdd1Div4(Curve.Params().P), Curve.Params().P)

	// check if y is a square root of x^3 - 3*x + B.
	ySquared := new(big.Int).Mul(point.Y, point.Y)
	ySquared.Mod(ySquared, Curve.Params().P)

	// check if (X, Y) is a point on the curve
	if ySquared.Cmp(xCube) != 0 {
		return errors.New("X is not an abscissa of a point on the elliptic curve")
	}
	return nil
}

// Inverse returns the inverse point of an input elliptic point
func (point EllipticPoint) Inverse() (*EllipticPoint, error) {
	// check if point is on the curve
	if !Curve.IsOnCurve(point.X, point.Y) {
		return nil, errors.New("The input point is not an elliptic point")
	}

	resPoint := new(EllipticPoint).Zero()

	// the inverse of the point (x, y) mod P is the point (x, -y) mod P
	resPoint.X.Set(point.X)
	resPoint.Y.Sub(Curve.Params().P, point.Y)
	resPoint.Y.Mod(resPoint.Y, Curve.Params().P)

	return resPoint, nil
}

// Randomize generates a random elliptic point
func (point *EllipticPoint) Randomize() {
	for {
		point.X = RandScalar()
		err := point.ComputeYCoord()
		if (err == nil) && (point.IsSafe()) {
			break
		}
	}
}

// IsSafe returns true if an input elliptic point is on the curve and has order not equal to 2
func (point EllipticPoint) IsSafe() bool {
	if !Curve.IsOnCurve(point.X, point.Y) {
		return false
	}

	var doublePoint EllipticPoint
	doublePoint.X, doublePoint.Y = Curve.Double(point.X, point.Y)

	return !doublePoint.IsEqual(new(EllipticPoint).Zero())
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
		return paddedAppend(BigIntSize, b, point.X.Bytes())
	}
	return nil
}

// Decompress decompresses a byte array, which was created by CompressPoint func,
// to a point on the given curve.
func (point *EllipticPoint) Decompress(compressPointBytes []byte) error {
	format := compressPointBytes[0]
	yBit := (format & 0x1) == 0x1
	format &= ^byte(0x1)

	if format != PointCompressed {
		return errors.New("invalid magic in compressed compressPoint bytes")
	}

	point.X = new(big.Int).SetBytes(compressPointBytes[1:33])

	err := point.ComputeYCoord()
	if err != nil {
		return err
	}

	if yBit != isOdd(point.Y) {
		point.Y.Sub(Curve.Params().P, point.Y)
	}

	return nil
}

// Hash derives a new elliptic point from an elliptic point and an index using hash function
func (point EllipticPoint) Hash(index int) *EllipticPoint {
	res := new(EllipticPoint).Zero()
	tmp := AddPaddingBigInt(point.X, BigIntSize)
	tmp = append(tmp, byte(index))

	for {
		tmp = common.HashB(tmp)
		res.X.SetBytes(tmp)
		err := res.ComputeYCoord()

		if (err == nil) && (res.IsSafe()) {
			break
		}
	}

	return res
}

// Set sets two coordinates to an elliptic point
func (point *EllipticPoint) Set(x, y *big.Int) {
	if point.X == nil {
		point.X = new(big.Int)
	}
	if point.Y == nil {
		point.Y = new(big.Int)
	}

	point.X.Set(x)
	point.Y.Set(y)
}

// Add adds an elliptic point to another elliptic point
func (point EllipticPoint) Add(targetPoint *EllipticPoint) *EllipticPoint {
	res := new(EllipticPoint)
	res.X, res.Y = Curve.Add(point.X, point.Y, targetPoint.X, targetPoint.Y)
	return res
}

// Sub subtracts an elliptic point to another elliptic point
func (point EllipticPoint) Sub(targetPoint *EllipticPoint) (*EllipticPoint, error) {
	invPoint, err := targetPoint.Inverse()
	if err != nil {
		return nil, err
	}

	res := point.Add(invPoint)
	return res, nil
}

// IsEqual returns true if two input elliptic points are equal, false otherwise
func (point EllipticPoint) IsEqual(p *EllipticPoint) bool {
	return point.X.Cmp(p.X) == 0 && point.Y.Cmp(p.Y) == 0
}

// ScalarMult returns x*P for x in Z_N and P in E(Z_P)
func (point EllipticPoint) ScalarMult(factor *big.Int) *EllipticPoint {
	res := new(EllipticPoint).Zero()
	res.X, res.Y = Curve.ScalarMult(point.X, point.Y, factor.Bytes())
	return res
}

// Derive returns a pseudo-random elliptic curve point P = F(seed, derivator), where
// F is a pseudo-random function defined by F(x, y) = 1/(x + y)*G, where x, y are integers,
// seed and derivator are integers of size at least 32 bytes,
// G is a generating point of the group of points of an elliptic curve.
func (point *EllipticPoint) Derive(seed, derivator *big.Int) *EllipticPoint {
	// point must be on the curve
	if !point.IsSafe() {
		return nil
	}
	return point.ScalarMult(new(big.Int).ModInverse(new(big.Int).Add(seed, derivator), Curve.Params().N))
}
