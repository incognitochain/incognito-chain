package privacy

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
)

var _ = func() (_ struct{}) {
	Logger.Init(common.NewBackend(nil).Logger("test", true))
	Logger.Log.Info("This runs before init()!")
	return
}()

/*
	Unit test for MarshalJSON/UnmarshalJSON EllipticPoint
*/

func TestECMarshalJSON(t *testing.T) {
	// random a elliptic point
	point := new(EllipticPoint)
	point.Randomize()

	// marshalJSON point
	bytesJSON, err := point.MarshalJSON()
	assert.Equal(t, nil, err)
	assert.Greater(t, len(bytesJSON), 0)

	//unmarshalJSON
	point2 := new(EllipticPoint)
	err2 := point2.UnmarshalJSON(bytesJSON)
	assert.Equal(t, nil, err2)
	assert.Equal(t, point, point2)
}

/*
	Unit test for ComputeYCoord EllipticPoint
*/
func TestECComputeYCoord(t *testing.T) {
	points := make([]*EllipticPoint, 4)
	for i := 0; i < len(points); i++ {
		points[i] = new(EllipticPoint)
		points[i].Randomize()
	}

	data := []struct {
		X   *big.Int
		Y   *big.Int
		err error
	}{
		{points[0].x, points[0].y, nil},
		{points[1].x, points[1].y, nil},
		{points[2].x, points[2].y, nil},
		{points[3].x, points[3].y, nil},
		{new(big.Int).SetBytes([]byte("17575166438094688464157431909385935670362228351757383795768436485341155942033")), nil, InvalidXCoordErr},
		{new(big.Int).SetBytes([]byte("96168539034483116404758217466223875298688790819305130644610716258571307712734")), nil, InvalidXCoordErr},
	}

	for _, item := range data {
		pointTmp := new(EllipticPoint)
		pointTmp.x = item.X

		err := pointTmp.computeYCoord()
		assert.Equal(t, item.err, err)
		assert.Equal(t, item.Y, pointTmp.y)
	}
}

/*
	Unit test for Inverse EllipticPoint
*/

func TestECInverse(t *testing.T) {
	points := make([]*EllipticPoint, 10)
	for i := 0; i < len(points); i++ {
		points[i] = new(EllipticPoint)
		points[i].Randomize()
	}

	for _, item := range points {
		itemInv, err := item.inverse()

		invY := new(big.Int).Sub(Curve.Params().P, item.y)
		invY.Mod(invY, Curve.Params().P)

		assert.Equal(t, nil, err)
		assert.Equal(t, item.x, itemInv.x)
		assert.Equal(t, invY, itemInv.y)
	}
}

func TestECInverseWithInvalidPoint(t *testing.T) {
	points := make([]*EllipticPoint, 10)
	for i := 0; i < len(points); i++ {
		points[i] = new(EllipticPoint)
		points[i].Randomize()
	}

	for _, item := range points {
		item.x = new(big.Int).SetBytes(RandBytes(common.BigIntSize))
		itemInv, err := item.inverse()

		assert.Equal(t, IsNotAnEllipticPointErr, err)
		assert.Equal(t, (*EllipticPoint)(nil), itemInv)
	}
}

/*
	Unit test for Randomize EllipticPoint
*/

func TestECRandomize(t *testing.T) {
	for i := 0; i < 10; i++ {
		point := new(EllipticPoint)
		point.Randomize()

		assert.Equal(t, true, point.IsSafe())
	}
}

/*
	Unit test for IsSafe EllipticPoint
*/

func TestECIsSafeWithZeroPoint(t *testing.T) {
	point := new(EllipticPoint)
	point.Zero()
	assert.Equal(t, false, point.IsSafe())
}

/*
	Unit test for Compress/Decompress EllipticPoint
*/

func TestECCompressDecompress(t *testing.T) {
	for i := 0; i < 10; i++ {
		// random elliptic point
		point := new(EllipticPoint)
		point.Randomize()

		// compress the point
		pointBytes := point.Compress()
		assert.Equal(t, CompressedEllipticPointSize, len(pointBytes))

		// decompress from bytes array
		point2 := new(EllipticPoint)
		err := point2.Decompress(pointBytes)
		assert.Equal(t, point, point2)
		assert.Equal(t, nil, err)
	}
}

func TestECCompressWithInvalidPoint(t *testing.T) {
	for i := 0; i < 10; i++ {
		point := new(EllipticPoint)
		point.Randomize()

		// edit point
		point.x.Add(point.x, big.NewInt(1))

		pointBytes := point.Compress()

		assert.Equal(t, 0, len(pointBytes))
	}
}
