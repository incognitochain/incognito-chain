package aggregaterange

import (
	"crypto/rand"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"testing"
	"time"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	log.SetOutput(ioutil.Discard)
	m.Run()
}

var _ = func() (_ struct{}) {
	fmt.Println("This runs before init()!")
	privacy.Logger.Init(common.NewBackend(nil).Logger("test", true))
	return
}()

//TestInnerProduct test inner product calculation
func TestInnerProduct(t *testing.T) {
	n := 2
	a := make([]*big.Int, n)
	b := make([]*big.Int, n)

	for i := 0; i < n; i++ {
		a[i] = big.NewInt(10)
		b[i] = big.NewInt(20)
	}

	c, _ := innerProduct(a, b)
	assert.Equal(t, big.NewInt(400), c)

	bytes := privacy.RandBytes(33)

	num1 := new(big.Int).SetBytes(bytes)
	num1Inverse := new(big.Int).ModInverse(num1, privacy.Curve.Params().N)

	num2 := new(big.Int).SetBytes(bytes)
	num2 = num2.Mod(num2, privacy.Curve.Params().N)
	num2Inverse := new(big.Int).ModInverse(num2, privacy.Curve.Params().N)

	assert.Equal(t, num1Inverse, num2Inverse)
}

func TestEncodeVectors(t *testing.T) {
	var AggParam = newBulletproofParams(1)
	n := 64
	a := make([]*big.Int, n)
	b := make([]*big.Int, n)
	G := make([]*privacy.EllipticPoint, n)
	H := make([]*privacy.EllipticPoint, n)

	for i := range a {
		a[i] = big.NewInt(10)
		b[i] = big.NewInt(10)

		G[i] = new(privacy.EllipticPoint)
		G[i].Set(AggParam.g[i].GetX(), AggParam.g[i].GetY())

		H[i] = new(privacy.EllipticPoint)
		H[i].Set(AggParam.h[i].GetX(), AggParam.h[i].GetY())
	}
	start := time.Now()
	actualRes, err := encodeVectors(a, b, G, H)
	end := time.Since(start)
	privacy.Logger.Log.Info("Time encode vector: %v\n", end)
	if err != nil {
		privacy.Logger.Log.Info("Err: %v\n", err)
	}
	start = time.Now()
	expectedRes := new(privacy.EllipticPoint)
	expectedRes.Zero()
	for i := 0; i < n; i++ {
		expectedRes = expectedRes.Add(G[i].ScalarMult(a[i]))
		expectedRes = expectedRes.Add(H[i].ScalarMult(b[i]))
	}

	end = time.Since(start)
	privacy.Logger.Log.Info("Time normal encode vector: %v\n", end)

	assert.Equal(t, expectedRes, actualRes)
}

func TestInnerProductProve(t *testing.T) {
	var AggParam = newBulletproofParams(1)
	wit := new(InnerProductWitness)
	n := maxExp
	wit.a = make([]*big.Int, n)
	wit.b = make([]*big.Int, n)

	for i := range wit.a {
		//wit.a[i] = privacy.RandScalar()
		//wit.b[i] = privacy.RandScalar()
		tmp := privacy.RandBytes(3)

		wit.a[i] = new(big.Int).SetBytes(tmp)
		wit.b[i] = new(big.Int).SetBytes(tmp)
	}

	wit.p = new(privacy.EllipticPoint)
	wit.p.Zero()
	c, err := innerProduct(wit.a, wit.b)
	if err != nil {
		privacy.Logger.Log.Info("Err: %v\n", err)
	}

	for i := range wit.a {
		wit.p = wit.p.Add(AggParam.g[i].ScalarMult(wit.a[i]))
		wit.p = wit.p.Add(AggParam.h[i].ScalarMult(wit.b[i]))
	}
	wit.p = wit.p.Add(AggParam.u.ScalarMult(c))

	proof, err := wit.Prove(AggParam)
	if err != nil {
		privacy.Logger.Log.Info("Err: %v\n", err)
	}

	bytes := proof.Bytes()

	proof2 := new(InnerProductProof)
	proof2.SetBytes(bytes)

	res := proof2.Verify(AggParam)

	assert.Equal(t, true, res)
}

func TestAggregatedRangeProve(t *testing.T) {
	// prepare witness for Aggregated range protocol
	wit := new(AggregatedRangeWitness)
	numValue := 3
	values := make([]*big.Int, numValue)
	rands := make([]*big.Int, numValue)

	var r = rand.Reader

	for i := range values {
		values[i] = new(big.Int).SetBytes(privacy.RandBytes(2))
		rands[i] = privacy.RandScalar(r)
	}
	wit.Set(values, rands)

	// proving
	start := time.Now()
	proof, err := wit.Prove()
	assert.Equal(t, nil, err)

	end := time.Since(start)
	privacy.Logger.Log.Info("Aggregated range proving time: %v\n", end)

	// validate sanity for proof
	isValidSanity := proof.ValidateSanity()
	assert.Equal(t, true, isValidSanity)

	// convert proof to bytes array
	bytes := proof.Bytes()
	expectProofSize := EstimateMultiRangeProofSize(numValue)
	assert.Equal(t, int(expectProofSize), len(bytes))
	fmt.Printf("Aggregated range proof size: %v\n", len(bytes))

	// new aggregatedRangeProof from bytes array
	proof2 := new(AggregatedRangeProof)
	proof2.SetBytes(bytes)

	// verify the proof
	start = time.Now()
	res, err := proof2.Verify()
	end = time.Since(start)
	privacy.Logger.Log.Info("Aggregated range verification time: %v\n", end)

	assert.Equal(t, true, res)
	assert.Equal(t, nil, err)
}

func TestPad(t *testing.T) {
	data := []struct {
		number       int
		paddedNumber int
	}{
		{1000, 1024},
		{3, 4},
		{5, 8},
	}

	for _, item := range data {
		num := pad(item.number)
		assert.Equal(t, item.paddedNumber, num)
	}
}

func TestPowerVector(t *testing.T) {
	twoVector := powerVector(big.NewInt(2), 5)
	assert.Equal(t, 5, len(twoVector))
}

func TestJS(t *testing.T) {
	proofBytes := []byte{2, 3, 3, 118, 16, 36, 196, 69, 53, 37, 222, 255, 2, 92, 197, 84, 176, 71, 147, 95, 31, 9, 50, 23, 231, 137, 175, 236, 207, 196, 103, 170, 51, 86, 2, 140, 188, 58, 43, 218, 166, 121, 238, 13, 69, 125, 24, 127, 202, 64, 179, 204, 235, 139, 124, 76, 6, 187, 20, 110, 219, 251, 161, 7, 12, 155, 87, 2, 49, 231, 184, 220, 111, 118, 240, 153, 249, 229, 56, 221, 205, 216, 1, 150, 164, 170, 167, 87, 26, 208, 115, 189, 30, 76, 51, 246, 7, 190, 7, 251, 3, 4, 127, 182, 95, 215, 46, 119, 189, 34, 67, 64, 165, 204, 214, 115, 127, 171, 151, 90, 182, 28, 164, 162, 161, 28, 232, 197, 23, 124, 235, 163, 139, 3, 140, 46, 205, 51, 241, 135, 21, 205, 177, 95, 210, 104, 158, 32, 224, 240, 81, 156, 196, 22, 118, 36, 207, 220, 200, 117, 89, 220, 230, 192, 146, 73, 3, 185, 41, 210, 210, 254, 2, 18, 117, 39, 202, 31, 82, 166, 246, 100, 188, 88, 110, 13, 254, 217, 68, 118, 27, 27, 226, 52, 15, 25, 124, 148, 172, 175, 14, 146, 154, 192, 131, 192, 35, 236, 201, 239, 243, 28, 94, 114, 254, 107, 216, 84, 114, 208, 36, 253, 207, 135, 73, 83, 245, 153, 174, 148, 142, 246, 123, 44, 251, 195, 118, 28, 80, 30, 213, 10, 155, 7, 99, 83, 209, 83, 77, 68, 46, 229, 140, 202, 242, 153, 154, 219, 20, 202, 124, 183, 112, 248, 252, 247, 142, 83, 170, 164, 47, 179, 255, 247, 39, 237, 231, 177, 61, 55, 81, 219, 123, 225, 87, 228, 209, 101, 54, 192, 163, 27, 76, 12, 183, 7, 2, 113, 232, 135, 114, 161, 128, 64, 241, 112, 121, 24, 6, 248, 239, 26, 142, 160, 61, 13, 165, 58, 134, 247, 102, 61, 24, 138, 222, 152, 58, 252, 160, 3, 21, 19, 4, 97, 214, 79, 209, 193, 250, 250, 96, 110, 127, 26, 244, 239, 36, 172, 54, 12, 252, 194, 159, 82, 167, 117, 29, 5, 242, 144, 251, 141, 2, 111, 251, 232, 116, 39, 163, 221, 218, 12, 143, 64, 161, 53, 58, 100, 219, 195, 5, 174, 66, 36, 181, 62, 111, 135, 48, 163, 11, 24, 236, 236, 45, 3, 148, 1, 188, 103, 147, 14, 22, 235, 151, 199, 190, 4, 167, 78, 142, 44, 38, 40, 241, 177, 17, 222, 199, 113, 130, 188, 47, 84, 243, 171, 76, 227, 2, 67, 159, 232, 38, 21, 138, 155, 136, 240, 122, 253, 248, 14, 134, 128, 21, 81, 233, 63, 117, 246, 42, 53, 188, 23, 47, 154, 27, 147, 37, 91, 30, 3, 97, 55, 150, 176, 5, 26, 247, 147, 128, 227, 162, 12, 147, 128, 158, 57, 230, 77, 142, 84, 150, 170, 133, 20, 229, 112, 232, 27, 60, 67, 252, 73, 2, 46, 115, 87, 129, 167, 253, 246, 12, 187, 83, 186, 72, 35, 201, 174, 94, 162, 153, 64, 105, 79, 145, 108, 196, 138, 178, 144, 205, 95, 7, 97, 212, 3, 238, 70, 84, 148, 169, 58, 255, 94, 58, 237, 34, 115, 144, 42, 226, 55, 89, 117, 2, 182, 75, 141, 60, 55, 184, 54, 29, 89, 71, 107, 254, 93, 3, 15, 154, 9, 65, 59, 164, 33, 116, 159, 212, 188, 156, 113, 232, 83, 65, 181, 17, 134, 141, 181, 242, 79, 136, 90, 16, 240, 187, 73, 184, 225, 25, 3, 42, 96, 34, 61, 131, 66, 74, 167, 244, 224, 138, 175, 106, 161, 161, 91, 83, 192, 113, 255, 117, 164, 128, 72, 214, 189, 38, 112, 86, 50, 209, 239, 3, 189, 106, 139, 246, 17, 252, 123, 2, 111, 154, 122, 113, 151, 18, 214, 205, 110, 134, 170, 1, 120, 102, 198, 246, 141, 247, 217, 121, 11, 27, 180, 64, 2, 175, 2, 148, 236, 54, 105, 163, 123, 190, 197, 26, 102, 233, 254, 207, 130, 53, 190, 191, 3, 83, 198, 62, 168, 147, 229, 172, 187, 31, 36, 255, 180, 3, 72, 195, 169, 38, 118, 218, 93, 7, 106, 19, 232, 99, 12, 56, 254, 136, 66, 20, 231, 98, 146, 246, 229, 210, 155, 116, 140, 69, 129, 54, 200, 85, 2, 5, 12, 151, 130, 149, 89, 94, 36, 227, 39, 10, 14, 68, 94, 123, 189, 61, 53, 138, 34, 154, 55, 116, 94, 170, 50, 227, 29, 135, 204, 115, 59, 199, 62, 230, 19, 175, 197, 125, 67, 63, 200, 112, 11, 117, 109, 199, 138, 116, 20, 244, 114, 237, 147, 223, 63, 223, 22, 150, 108, 49, 85, 57, 248, 72, 187, 204, 142, 26, 55, 17, 29, 101, 218, 254, 208, 130, 182, 189, 110, 38, 163, 196, 83, 62, 225, 207, 6, 179, 16, 188, 15, 35, 74, 254, 142, 3, 121, 25, 14, 138, 27, 86, 228, 94, 83, 130, 87, 208, 48, 17, 4, 125, 80, 0, 82, 244, 204, 19, 170, 74, 53, 93, 115, 172, 69, 160, 141, 225}

	proof := new(AggregatedRangeProof)
	proof.SetBytes(proofBytes)

	res, _ := proof.Verify()
	fmt.Println(res)
}
