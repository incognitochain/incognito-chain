package zkp

import (
	"fmt"
	"github.com/ninjadotorg/constant/privacy"
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
		//wit.a[i] = privacy.RandBigInt()
		//wit.b[i] = privacy.RandBigInt()
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
	wit := new(AggregatedRangeWitness)
	numValue := 1
	wit.values = make([]*big.Int, numValue)
	wit.rands = make([]*big.Int, numValue)

	for i := range wit.values {
		wit.values[i] = big.NewInt(10)
		wit.rands[i] = privacy.RandBigInt()
	}

	start := time.Now()
	proof, err := wit.Prove()
	if err != nil {
		fmt.Printf("Err: %v\n", err)
	}
	end := time.Since(start)
	fmt.Printf("Aggregated range proving time: %v\n", end)

	bytes := proof.Bytes()
	fmt.Printf("Aggregated range proof size: %v\n", len(bytes))

	proof2 := new(AggregatedRangeProof)
	proof2.SetBytes(bytes)

	start = time.Now()
	res := proof.Verify()
	end = time.Since(start)
	fmt.Printf("Aggregated range verification time: %v\n", end)

	assert.Equal(t, true, res)
}

func BenchmarkAggregatedRangeProve(b *testing.B) {
	wit := new(AggregatedRangeWitness)
	numValue := 1
	wit.values = make([]*big.Int, numValue)
	wit.rands = make([]*big.Int, numValue)

	for i := range wit.values {
		wit.values[i] = big.NewInt(10)
		wit.rands[i] = privacy.RandBigInt()
	}

	for i:=0; i<b.N; i++ {
		start := time.Now()
		proof, err := wit.Prove()
		if err != nil {
			fmt.Printf("Err: %v\n", err)
		}
		end := time.Since(start)
		fmt.Printf("Aggregated range proving time: %v\n", end)

		bytes := proof.Bytes()
		fmt.Printf("Len byte proof: %v\n", len(bytes))

		proof2 := new(AggregatedRangeProof)
		proof2.SetBytes(bytes)

		start = time.Now()
		res := proof.Verify()
		end = time.Since(start)
		fmt.Printf("Aggregated range verification time: %v\n", end)

		assert.Equal(b, true, res)
	}
}

func TestMultiExponentiation(t *testing.T){
	//exponents := []*big.Int{big.NewInt(5), big.NewInt(10),big.NewInt(5),big.NewInt(7), big.NewInt(5)}

	exponents := make([]*big.Int, 64)
	for i:= range exponents{
		exponents[i] = new(big.Int).SetBytes(privacy.RandBytes(2))
	}

	bases := newBulletproofParams(1)
	//fmt.Printf("Values: %v\n", exponents[0])

	start1 := time.Now()
	expectedRes := new(privacy.EllipticPoint).Zero()
	for i:= range exponents{
		expectedRes = expectedRes.Add(bases.G[i].ScalarMult(exponents[i]))
	}
	end1 := time.Since(start1)
	fmt.Printf("normal calculation time: %v\n", end1)
	fmt.Printf("Res from normal calculation: %+v\n", expectedRes)


	start2 := time.Now()
	testcase4, err := privacy.MultiScalarmult(bases.G, exponents)
	end2 := time.Since(start2)
	fmt.Printf("multi scalarmult time: %v\n", end2)
	fmt.Printf("Res from multi exponentiation alg: %+v\n", testcase4)

	start3 := time.Now()
	testcase5, err := privacy.MultiScalar2(bases.G, exponents)
	end3 := time.Since(start3)
	fmt.Printf("multi scalarmult 2 time: %v\n", end3)
	fmt.Printf("Res from multi exponentiation alg: %+v\n", testcase5)


	if err != nil{
		fmt.Printf("Error of multi-exponentiation algorithm")
	}

	assert.Equal(t, expectedRes, testcase4)
}

func TestPad (t*testing.T){
	num := 1000
	testcase1 := 1024

	start := time.Now()
	padNum := pad(num)
	end := time.Since(start)
	fmt.Printf("Pad 1: %v\n", end)

	assert.Equal(t, testcase1, padNum)
}

func TestPowerVector(t*testing.T){
	twoVector := powerVector(big.NewInt(2), 5)
	fmt.Printf("two vector : %v\n", twoVector)
}
