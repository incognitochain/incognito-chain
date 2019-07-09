package zkp

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
	"time"
)

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
		G[i].Set(AggParam.G[i].X, AggParam.G[i].Y)

		H[i] = new(privacy.EllipticPoint)
		H[i].Set(AggParam.H[i].X, AggParam.H[i].Y)
	}
	start := time.Now()
	actualRes, err := EncodeVectors(a, b, G, H)
	end := time.Since(start)
	fmt.Printf("Time encode vector: %v\n", end)
	if err != nil {
		fmt.Printf("Err: %v\n", err)
	}
	start = time.Now()
	expectedRes := new(privacy.EllipticPoint).Zero()
	for i := 0; i < n; i++ {
		expectedRes = expectedRes.Add(G[i].ScalarMult(a[i]))
		expectedRes = expectedRes.Add(H[i].ScalarMult(b[i]))
	}

	end = time.Since(start)
	fmt.Printf("Time normal encode vector: %v\n", end)

	assert.Equal(t, expectedRes, actualRes)
}

func TestInnerProductProve(t *testing.T) {
	var AggParam = newBulletproofParams(1)
	wit := new(InnerProductWitness)
	n := privacy.MaxExp
	wit.a = make([]*big.Int, n)
	wit.b = make([]*big.Int, n)

	for i := range wit.a {
		//wit.a[i] = privacy.RandScalar()
		//wit.b[i] = privacy.RandScalar()
		tmp := privacy.RandBytes(3)

		wit.a[i] = new(big.Int).SetBytes(tmp)
		wit.b[i] = new(big.Int).SetBytes(tmp)
	}

	wit.p = new(privacy.EllipticPoint).Zero()
	c, err := innerProduct(wit.a, wit.b)
	if err != nil {
		fmt.Printf("Err: %v\n", err)
	}

	for i := range wit.a {
		wit.p = wit.p.Add(AggParam.G[i].ScalarMult(wit.a[i]))
		wit.p = wit.p.Add(AggParam.H[i].ScalarMult(wit.b[i]))
	}
	wit.p = wit.p.Add(AggParam.U.ScalarMult(c))

	proof, err := wit.Prove(AggParam)
	if err != nil {
		fmt.Printf("Err: %v\n", err)
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

	for i := range values {
		values[i] = new(big.Int).SetBytes(privacy.RandBytes(2))
		rands[i] = privacy.RandScalar()
	}
	wit.Set(values, rands)

	// proving
	start := time.Now()
	proof, err := wit.Prove()
	assert.Equal(t, nil, err)
	end := time.Since(start)
	fmt.Printf("Aggregated range proving time: %v\n", end)

	// validate sanity for proof
	isValidSanity := proof.ValidateSanity()
	assert.Equal(t, true, isValidSanity)

	// convert proof to bytes array
	bytes := proof.Bytes()
	expectProofSize := estimateMultiRangeProofSize(numValue)
	assert.Equal(t, int(expectProofSize), len(bytes))
	fmt.Printf("Aggregated range proof size: %v\n", len(bytes))

	// new AggregatedRangeProof from bytes array
	proof2 := new(AggregatedRangeProof)
	proof2.SetBytes(bytes)

	// verify the proof
	start = time.Now()
	res := proof2.Verify()
	end = time.Since(start)
	fmt.Printf("Aggregated range verification time: %v\n", end)

	assert.Equal(t, true, res)
}

func TestPad(t *testing.T) {
	data := []struct{
		number int
		paddedNumber int
	}{
		{1000, 1024},
		{3, 4},
		{5, 8},
	}

	for _, item := range data{
		num := pad(item.number)
		assert.Equal(t, item.paddedNumber, num)
	}
}

func TestPowerVector(t *testing.T) {
	twoVector := powerVector(big.NewInt(2), 5)
	assert.Equal(t, 5, len(twoVector))
}
