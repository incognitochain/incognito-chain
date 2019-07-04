package zkp

import (
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
	privacy.Logger.Log.Info("Time encode vector: %v\n", end)
	if err != nil {
		privacy.Logger.Log.Info("Err: %v\n", err)
	}
	start = time.Now()
	expectedRes := new(privacy.EllipticPoint).Zero()
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
		privacy.Logger.Log.Info("Err: %v\n", err)
	}

	for i := range wit.a {
		wit.p = wit.p.Add(AggParam.G[i].ScalarMult(wit.a[i]))
		wit.p = wit.p.Add(AggParam.H[i].ScalarMult(wit.b[i]))
	}
	wit.p = wit.p.Add(AggParam.U.ScalarMult(c))

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

	point := new(privacy.EllipticPoint).Zero()
	privacy.Logger.Log.Info("testt: %v\n", point.Compress())
	wit := new(AggregatedRangeWitness)
	numValue := 3
	wit.values = make([]*big.Int, numValue)
	wit.rands = make([]*big.Int, numValue)

	for i := range wit.values {
		wit.values[i] = big.NewInt(10)
		wit.rands[i] = privacy.RandScalar()
	}

	start := time.Now()
	proof, err := wit.Prove()
	if err != nil {
		privacy.Logger.Log.Info("Err: %v\n", err)
	}
	end := time.Since(start)
	privacy.Logger.Log.Info("Aggregated range proving time: %v\n", end)

	bytes := proof.Bytes()
	privacy.Logger.Log.Info("Aggregated range proof size: %v\n", len(bytes))

	proof2 := new(AggregatedRangeProof)
	proof2.SetBytes(bytes)

	start = time.Now()
	res := proof2.Verify()
	end = time.Since(start)
	privacy.Logger.Log.Info("Aggregated range verification time: %v\n", end)

	assert.Equal(t, true, res)
}

func BenchmarkAggregatedRangeProve(b *testing.B) {
	wit := new(AggregatedRangeWitness)
	numValue := 1
	wit.values = make([]*big.Int, numValue)
	wit.rands = make([]*big.Int, numValue)

	for i := range wit.values {
		wit.values[i] = big.NewInt(10)
		wit.rands[i] = privacy.RandScalar()
	}

	for i := 0; i < b.N; i++ {
		start := time.Now()
		proof, err := wit.Prove()
		if err != nil {
			privacy.Logger.Log.Info("Err: %v\n", err)
		}
		end := time.Since(start)
		privacy.Logger.Log.Info("Aggregated range proving time: %v\n", end)

		bytes := proof.Bytes()
		privacy.Logger.Log.Info("Len byte proof: %v\n", len(bytes))

		proof2 := new(AggregatedRangeProof)
		proof2.SetBytes(bytes)

		start = time.Now()
		res := proof.Verify()
		end = time.Since(start)
		privacy.Logger.Log.Info("Aggregated range verification time: %v\n", end)

		assert.Equal(b, true, res)
	}
}

func TestMultiExponentiation(t *testing.T) {
	//exponents := []*big.Int{big.NewInt(5), big.NewInt(10),big.NewInt(5),big.NewInt(7), big.NewInt(5)}

	exponents := make([]*big.Int, 64)
	for i := range exponents {
		exponents[i] = new(big.Int).SetBytes(privacy.RandBytes(2))
	}

	bases := newBulletproofParams(1)
	//privacy.Logger.Log.Info("Values: %v\n", exponents[0])

	start1 := time.Now()
	expectedRes := new(privacy.EllipticPoint).Zero()
	for i := range exponents {
		expectedRes = expectedRes.Add(bases.G[i].ScalarMult(exponents[i]))
	}
	end1 := time.Since(start1)
	privacy.Logger.Log.Info("normal calculation time: %v\n", end1)
	privacy.Logger.Log.Info("Res from normal calculation: %+v\n", expectedRes)

	start2 := time.Now()
	testcase4, err := privacy.MultiScalarmult(bases.G, exponents)
	if err != nil {
		privacy.Logger.Log.Info("Error of multi-exponentiation algorithm")
	}
	end2 := time.Since(start2)
	privacy.Logger.Log.Info("multi scalarmult time: %v\n", end2)
	privacy.Logger.Log.Info("Res from multi exponentiation alg: %+v\n", testcase4)

	start3 := time.Now()
	testcase5, err := privacy.MultiScalar2(bases.G, exponents)
	if err != nil {
		privacy.Logger.Log.Info("Error of multi-exponentiation algorithm")
	}
	end3 := time.Since(start3)
	privacy.Logger.Log.Info("multi scalarmult 2 time: %v\n", end3)
	privacy.Logger.Log.Info("Res from multi exponentiation alg: %+v\n", testcase5)

	assert.Equal(t, expectedRes, testcase4)
}

func TestPad(t *testing.T) {
	num := 1000
	testcase1 := 1024

	start := time.Now()
	padNum := pad(num)
	end := time.Since(start)
	privacy.Logger.Log.Info("Pad 1: %v\n", end)

	assert.Equal(t, testcase1, padNum)
}

func TestPowerVector(t *testing.T) {
	twoVector := powerVector(big.NewInt(2), 5)
	privacy.Logger.Log.Info("two vector : %v\n", twoVector)
}
