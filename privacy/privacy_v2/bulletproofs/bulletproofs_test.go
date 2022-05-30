package bulletproofs

import (
	cryptoRand "crypto/rand"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/privacy/operation"
	operationV1 "github.com/incognitochain/incognito-chain/privacy/operation/v1"
	"github.com/incognitochain/incognito-chain/privacy/privacy_util"
	bulletproofsV1 "github.com/incognitochain/incognito-chain/privacy/privacy_v2/bulletproofs/v1"
	. "github.com/stretchr/testify/assert"
)

var _ = func() (_ struct{}) {
	// seed the rng for test, not production
	rand.Seed(time.Now().UnixNano())
	Logger.Init(common.NewBackend(nil).Logger("test", true))
	return
}()

var (
	batchLen   = []int{2, 4, 8, 16, 32}[rand.Int()%5]
	batchLenCA = batchLen / 2
	batchBases = func() []*operation.Point {
		result := make([]*operation.Point, batchLen)
		for i := range result {
			if i < batchLenCA {
				result[i] = operation.RandomPoint()
			} else {
				result[i] = nil
			}
		}
		return result
	}()
	rangeProof   *AggregatedRangeProof
	rangeProofV1 *bulletproofsV1.AggregatedRangeProof

	batchedProofs, batchedProofsV1 = func() ([]*AggregatedRangeProof, []*bulletproofsV1.AggregatedRangeProof) {
		result := make([]*AggregatedRangeProof, batchLen)
		resultV1 := make([]*bulletproofsV1.AggregatedRangeProof, batchLen)
		fmt.Printf("batch %d with %d CA proofs\n", batchLen, batchLenCA)
		for i := 0; i < batchLen; i++ {
			numOutputs := []int{1, 2, 4}[rand.Int()%3] // can use other distribution
			// fmt.Printf("%d outputs\n", numOutputs)
			values := make([]uint64, numOutputs)
			rands := make([]*operation.Scalar, numOutputs)
			for i := range values {
				values[i] = uint64(rand.Uint64())
				rands[i] = operation.RandomScalar()
			}
			wit := new(AggregatedRangeWitness)
			wit.Set(values, rands)
			var err error
			if batchBases[i] == nil {
				result[i], err = wit.Prove()
				if err != nil {
					panic(err)
				}
			} else {
				result[i], err = wit.ProveUsingBase(batchBases[i])
				if err != nil {
					panic(err)
				}
			}
			resultV1[i] = new(bulletproofsV1.AggregatedRangeProof)
			err = resultV1[i].SetBytes(result[i].Bytes())
			if err != nil {
				panic(err)
			}
		}
		return result, resultV1
	}()
)

type fnProve = func(values []uint64, rands []*operation.Scalar, rands2 []*operationV1.Scalar)

var provers = map[string]fnProve{
	"Go&old-curve-impl": func(values []uint64, rands []*operation.Scalar, rands2 []*operationV1.Scalar) {
		wit := new(bulletproofsV1.AggregatedRangeWitness)
		wit.Set(values, rands2)
		proof, err := wit.Prove()
		if err != nil {
			panic(err)
		}
		rangeProofV1 = proof
	},
	"Go&new-curve-impl": func(values []uint64, rands []*operation.Scalar, rands2 []*operationV1.Scalar) {
		wit := new(AggregatedRangeWitness)
		wit.Set(values, rands)
		proof, err := wit.Prove()
		if err != nil {
			panic(err)
		}
		rangeProof = proof
	},
}

type fnProveVerify = func()

var pverifiers = map[string]fnProveVerify{
	"Go&old-curve-impl": func() {
		valid, err := rangeProofV1.Verify()
		if !valid || err != nil {
			panic(err)
		}
	},
	"Go&new-curve-impl": func() {
		valid, err := rangeProof.Verify()
		if !valid || err != nil {
			panic(err)
		}
	},
	"batch-old": func() {
		valid, err, _ := bulletproofsV1.VerifyBatch(batchedProofsV1[batchLenCA:])
		if !valid || err != nil {
			panic(err)
		}

		for i, proof := range batchedProofsV1[:batchLenCA] {
			tmpbase, _ := new(operationV1.Point).FromBytesS(batchBases[i].ToBytesS())
			valid, err := proof.VerifyUsingBase(tmpbase)
			if !valid || err != nil {
				panic(err)
			}
		}
	},
	"batch-new": func() {
		valid, err := VerifyBatch(batchedProofs, batchBases)
		if !valid || err != nil {
			panic(err)
		}
	},
}

type fnRandomScalarMult = func()

var pointMults = map[string]fnRandomScalarMult{
	"old-base": func() {
		sc := operationV1.RandomScalar()
		(&operationV1.Point{}).ScalarMultBase(sc)
	},
	"new-base": func() {
		sc := operation.RandomScalar()
		(&operation.Point{}).ScalarMultBase(sc)
	},
	"old-point": func() {
		sc := operationV1.RandomScalar()
		(&operationV1.Point{}).ScalarMult(operationV1.PedCom.G[0], sc)
	},
	"new-point": func() {
		sc := operation.RandomScalar()
		(&operation.Point{}).ScalarMult(operation.NewGeneratorPoint(), sc)
	},
	"legacy-map-to-point": func() {
		b := make([]byte, 32)
		cryptoRand.Read(b)
		operationV1.HashToPoint(b)
	},
}

type fnRandomMultiScalarMult = func([][]byte)

var multiScalarMults = map[string]fnRandomMultiScalarMult{
	"old-multi": func(points [][]byte) {
		var sLst []*operationV1.Scalar
		var pLst []*operationV1.Point
		for _, rawPoint := range points {
			sLst = append(sLst, operationV1.RandomScalar())
			var temp operationV1.Point
			_, err := temp.FromBytesS(rawPoint)
			if err != nil {
				panic(err)
			}
			pLst = append(pLst, &temp)
		}
		p := &operationV1.Point{}
		p.MultiScalarMult(sLst, pLst)
	},
	"new-multi": func(points [][]byte) {
		var sLst []*operation.Scalar
		var pLst []*operation.Point
		for _, rawPoint := range points {
			sLst = append(sLst, operation.RandomScalar())
			var temp operation.Point
			_, err := temp.FromBytesS(rawPoint)
			if err != nil {
				panic(err)
			}
			pLst = append(pLst, &temp)
		}
		p := operation.NewGeneratorPoint()
		p.MultiScalarMult(sLst, pLst)
	},
	"new-multi-vartime": func(points [][]byte) {
		var sLst []*operation.Scalar
		var pLst []*operation.Point
		for _, rawPoint := range points {
			sLst = append(sLst, operation.RandomScalar())
			var temp operation.Point
			_, err := temp.FromBytesS(rawPoint)
			if err != nil {
				panic(err)
			}
			pLst = append(pLst, &temp)
		}
		p := operation.NewGeneratorPoint()
		p.VarTimeMultiScalarMult(sLst, pLst)
	},
}

type fnRandomAddPedersen = func([]byte, []byte)

var pedAdds = map[string]fnRandomAddPedersen{
	"old": func(raw_A, raw_B []byte) {
		sc_a := operationV1.RandomScalar()
		sc_b := operationV1.RandomScalar()
		pA := &operationV1.Point{}
		pA.FromBytesS(raw_A)
		pB := &operationV1.Point{}
		pB.FromBytesS(raw_B)
		(&operationV1.Point{}).AddPedersen(sc_a, pA, sc_b, pB)
	},
	"new": func(raw_A, raw_B []byte) {
		sc_a := operation.RandomScalar()
		sc_b := operation.RandomScalar()
		pA := &operation.Point{}
		pA.FromBytesS(raw_A)
		pB := &operation.Point{}
		pB.FromBytesS(raw_B)
		(&operation.Point{}).AddPedersen(sc_a, pA, sc_b, pB)
	},
}

func BenchmarkBPProve(b *testing.B) {
	benchmarks := []struct {
		prover     string
		numOutputs int
	}{
		{"Go&old-curve-impl", 1},
		{"Go&old-curve-impl", 2},
		{"Go&old-curve-impl", 4},
		{"Go&old-curve-impl", 8},
		{"Go&old-curve-impl", 16},
		{"Go&new-curve-impl", 1},
		{"Go&new-curve-impl", 2},
		{"Go&new-curve-impl", 4},
		{"Go&new-curve-impl", 8},
		{"Go&new-curve-impl", 16},
	}

	for _, bm := range benchmarks {
		// prepare prover inputs
		values := make([]uint64, bm.numOutputs)
		rands := make([]*operation.Scalar, bm.numOutputs)
		rands2 := make([]*operationV1.Scalar, bm.numOutputs)
		for i := range values {
			values[i] = uint64(rand.Uint64())
			rands[i] = operation.RandomScalar()
			rands2[i] = (&operationV1.Scalar{}).FromBytesS(rands[i].ToBytesS())
		}

		b.ResetTimer()
		b.Run(fmt.Sprintf("%s proving %d outputs", bm.prover, bm.numOutputs), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				provers[bm.prover](values, rands, rands2)
			}
		})
	}
}

func BenchmarkBPVerify(b *testing.B) {
	benchmarks := []struct {
		verifier   string
		numOutputs int
	}{
		{"Go&old-curve-impl", 1},
		{"Go&old-curve-impl", 2},
		{"Go&old-curve-impl", 4},
		{"Go&old-curve-impl", 8},
		{"Go&old-curve-impl", 16},
		{"Go&old-curve-impl", 32},
		{"Go&new-curve-impl", 1},
		{"Go&new-curve-impl", 2},
		{"Go&new-curve-impl", 4},
		{"Go&new-curve-impl", 8},
		{"Go&new-curve-impl", 16},
		{"Go&new-curve-impl", 32},
	}

	for _, bm := range benchmarks {
		// prepare prover inputs
		values := make([]uint64, bm.numOutputs)
		rands := make([]*operation.Scalar, bm.numOutputs)
		rands2 := make([]*operationV1.Scalar, bm.numOutputs)
		for i := range values {
			values[i] = uint64(rand.Uint64())
			rands[i] = operation.RandomScalar()
			rands2[i] = (&operationV1.Scalar{}).FromBytesS(rands[i].ToBytesS())
		}
		provers["Go&old-curve-impl"](values, rands, rands2)
		provers["Go&new-curve-impl"](values, rands, rands2)

		b.ResetTimer()
		b.Run(fmt.Sprintf("%s verify %d outputs", bm.verifier, bm.numOutputs), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				pverifiers[bm.verifier]()
			}
		})
	}
}

func BenchmarkBPBatchVerify(b *testing.B) {
	benchmarks := []struct {
		verifier string
	}{
		{"batch-old"},
		{"batch-new"},
	}

	for _, bm := range benchmarks {
		b.ResetTimer()
		b.Run(fmt.Sprintf("%s batch-verify outputs", bm.verifier), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				pverifiers[bm.verifier]()
			}
		})
	}
}

func BenchmarkCurvePointMult(b *testing.B) {
	benchmarks := []struct {
		curvelib string
	}{
		{"old-base"},
		{"new-base"},
		{"old-point"},
		{"new-point"},
		{"legacy-map-to-point"},
	}

	for _, bm := range benchmarks {
		b.ResetTimer()
		b.Run(fmt.Sprintf("%s scalar-mult", bm.curvelib), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				pointMults[bm.curvelib]()
			}
		})
	}
}

func BenchmarkMultiScalarMult(b *testing.B) {
	benchmarks := []struct {
		curvelib  string
		numPoints int
	}{
		{"old-multi", 4},
		{"old-multi", 8},
		{"old-multi", 16},
		{"new-multi", 4},
		{"new-multi", 8},
		{"new-multi", 16},
		{"new-multi-vartime", 4},
	}

	for _, bm := range benchmarks {
		pointsRaw := make([][]byte, bm.numPoints)
		for i := 0; i < bm.numPoints; i++ {
			pointsRaw[i] = operation.RandomPoint().ToBytesS()
		}
		b.ResetTimer()
		b.Run(fmt.Sprintf("%s multi-scalar-mult-%d", bm.curvelib, bm.numPoints), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				multiScalarMults[bm.curvelib](pointsRaw)
			}
		})
	}
}

func BenchmarkAddPedersen(b *testing.B) {
	benchmarks := []struct {
		curvelib string
	}{
		{"old"},
		{"new"},
	}

	for _, bm := range benchmarks {
		pointsRaw := make([][]byte, 2)
		for i := 0; i < 2; i++ {
			pointsRaw[i] = operation.RandomPoint().ToBytesS()
		}
		b.ResetTimer()
		b.Run(fmt.Sprintf("%s add-pedersen", bm.curvelib), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				pedAdds[bm.curvelib](pointsRaw[0], pointsRaw[1])
			}
		})
	}
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
		num := roundUpPowTwo(item.number)
		Equal(t, item.paddedNumber, num)
	}
}

func TestPowerVector(t *testing.T) {
	twoVector := powerVector(new(operation.Scalar).FromUint64(2), 5)
	Equal(t, 5, len(twoVector))
}

func TestInnerProduct(t *testing.T) {
	for j := 0; j < 5; j++ {
		n := privacy_util.MaxExp
		a := make([]*operation.Scalar, n)
		b := make([]*operation.Scalar, n)
		uinta := make([]uint64, n)
		uintb := make([]uint64, n)
		uintc := uint64(0)
		for i := 0; i < n; i++ {
			uinta[i] = uint64(rand.Intn(100000000))
			uintb[i] = uint64(rand.Intn(100000000))
			a[i] = new(operation.Scalar).FromUint64(uinta[i])
			b[i] = new(operation.Scalar).FromUint64(uintb[i])
			uintc += uinta[i] * uintb[i]
		}

		c, _ := innerProduct(a, b)
		Equal(t, new(operation.Scalar).FromUint64(uintc), c)
	}
}

func TestEncodeVectors(t *testing.T) {
	for i := 0; i < 5; i++ {
		var AggParam = newBulletproofParams(1)
		n := privacy_util.MaxExp
		a := make([]*operation.Scalar, n)
		b := make([]*operation.Scalar, n)
		G := make([]*operation.Point, n)
		H := make([]*operation.Point, n)

		for i := range a {
			a[i] = operation.RandomScalar()
			b[i] = operation.RandomScalar()
			G[i] = new(operation.Point).Set(AggParam.g[i])
			H[i] = new(operation.Point).Set(AggParam.h[i])
		}
		expectedRes := new(operation.Point).Identity()
		for i := 0; i < n; i++ {
			expectedRes.Add(expectedRes, new(operation.Point).ScalarMult(G[i], a[i]))
			expectedRes.Add(expectedRes, new(operation.Point).ScalarMult(H[i], b[i]))
		}

		mbuilder := operation.NewMultBuilder(false)
		_, err := encodeVectors(a, b, G, H, mbuilder)
		actualRes := mbuilder.Eval()
		Equal(t, err, nil)
		True(t, operation.IsPointEqual(expectedRes, actualRes))

		mbuilder = operation.NewMultBuilder(true)
		_, err = encodeVectors(a, b, G, H, mbuilder)
		actualRes = mbuilder.Eval()
		Equal(t, err, nil)
		True(t, operation.IsPointEqual(expectedRes, actualRes))
	}
}

func TestInnerProductProveVerify(t *testing.T) {
	for k := 0; k < 4; k++ {
		numValue := rand.Intn(privacy_util.MaxOutputCoin)
		numValuePad := roundUpPowTwo(numValue)
		aggParam := new(bulletproofParams)
		aggParam.g = AggParam.g[0 : numValuePad*privacy_util.MaxExp]
		aggParam.h = AggParam.h[0 : numValuePad*privacy_util.MaxExp]
		aggParam.u = AggParam.u
		aggParam.cs = AggParam.cs

		wit := new(InnerProductWitness)
		n := privacy_util.MaxExp * numValuePad
		wit.a = make([]*operation.Scalar, n)
		wit.b = make([]*operation.Scalar, n)

		for i := range wit.a {
			//wit.a[i] = privacy.RandomScalar()
			//wit.b[i] = privacy.RandomScalar()
			wit.a[i] = new(operation.Scalar).FromUint64(uint64(rand.Intn(100000)))
			wit.b[i] = new(operation.Scalar).FromUint64(uint64(rand.Intn(100000)))
		}

		c, _ := innerProduct(wit.a, wit.b)
		wit.p = new(operation.Point).ScalarMult(aggParam.u, c)

		for i := range wit.a {
			wit.p.Add(wit.p, new(operation.Point).ScalarMult(aggParam.g[i], wit.a[i]))
			wit.p.Add(wit.p, new(operation.Point).ScalarMult(aggParam.h[i], wit.b[i]))
		}

		proof, err := wit.Prove(aggParam.g, aggParam.h, aggParam.u, aggParam.cs.ToBytesS())
		if err != nil {
			fmt.Printf("Err: %v\n", err)
			return
		}
		res2 := proof.Verify(aggParam.g, aggParam.h, aggParam.u, aggParam.cs.ToBytesS())
		Equal(t, true, res2)
		// res2prime := proof.VerifyFaster(aggParam.g, aggParam.h, aggParam.u, aggParam.cs.ToBytesS())
		// Equal(t, true, res2prime)

		bytes := proof.Bytes()
		proof2 := new(InnerProductProof)
		proof2.SetBytes(bytes)
		res3 := proof2.Verify(aggParam.g, aggParam.h, aggParam.u, aggParam.cs.ToBytesS())
		Equal(t, true, res3)
		res3prime := proof.Verify(aggParam.g, aggParam.h, aggParam.u, aggParam.cs.ToBytesS())
		Equal(t, true, res3prime)
	}
}

func TestAggregatedRangeProveVerifyTampered(t *testing.T) {
	count := 10
	for i := 0; i < count; i++ {
		//prepare witness for Aggregated range protocol
		wit := new(AggregatedRangeWitness)
		numValue := rand.Intn(privacy_util.MaxOutputCoin)
		values := make([]uint64, numValue)
		rands := make([]*operation.Scalar, numValue)

		for i := range values {
			values[i] = uint64(rand.Uint64())
			rands[i] = operation.RandomScalar()
		}
		wit.Set(values, rands)

		// proving
		proof, err := wit.Prove()
		Equal(t, nil, err)

		testAggregatedRangeProofTampered(proof, t)
	}
}

func testAggregatedRangeProofTampered(proof *AggregatedRangeProof, t *testing.T) {
	saved := proof.a
	// tamper with one field
	proof.a = operation.RandomPoint()
	// verify using the fast variant
	res, err := proof.Verify()
	Equal(t, false, res)
	NotEqual(t, nil, err)
	proof.a = saved

	saved = proof.s
	// tamper with one field
	proof.s = operation.RandomPoint()
	// verify using the fast variant
	res, err = proof.Verify()
	Equal(t, false, res)
	NotEqual(t, nil, err)
	proof.s = saved

	saved = proof.t1
	// tamper with one field
	proof.t1 = operation.RandomPoint()
	// verify using the fast variant
	res, err = proof.Verify()
	Equal(t, false, res)
	NotEqual(t, nil, err)
	proof.t1 = saved

	saved = proof.t2
	// tamper with one field
	proof.t2 = operation.RandomPoint()
	// verify using the fast variant
	res, err = proof.Verify()
	Equal(t, false, res)
	NotEqual(t, nil, err)
	proof.t2 = saved

	savedScalar := proof.tauX
	// tamper with one field
	proof.tauX = operation.RandomScalar()
	// verify using the fast variant
	res, err = proof.Verify()
	Equal(t, false, res)
	NotEqual(t, nil, err)
	proof.tauX = savedScalar

	savedScalar = proof.tHat
	// tamper with one field
	proof.tHat = operation.RandomScalar()
	// verify using the fast variant
	res, err = proof.Verify()
	Equal(t, false, res)
	NotEqual(t, nil, err)
	proof.tHat = savedScalar

	savedScalar = proof.innerProductProof.a
	// tamper with one field
	proof.innerProductProof.a = operation.RandomScalar()
	// verify using the fast variant
	res, err = proof.Verify()
	Equal(t, false, res)
	NotEqual(t, nil, err)
	proof.innerProductProof.a = savedScalar

	savedScalar = proof.innerProductProof.b
	// tamper with one field
	proof.innerProductProof.b = operation.RandomScalar()
	// verify using the fast variant
	res, err = proof.Verify()
	Equal(t, false, res)
	NotEqual(t, nil, err)
	proof.innerProductProof.b = savedScalar

	saved = proof.innerProductProof.p
	// tamper with one field
	proof.innerProductProof.p = operation.RandomPoint()
	// verify using the fast variant
	res, err = proof.Verify()
	Equal(t, false, res)
	NotEqual(t, nil, err)
	proof.innerProductProof.p = saved

	for i := 0; i < len(proof.cmsValue); i++ {
		saved := proof.cmsValue[i]
		// tamper with one field
		proof.cmsValue[i] = operation.RandomPoint()
		// verify using the fast variant
		res, err = proof.Verify()
		Equal(t, false, res)
		NotEqual(t, nil, err)
		proof.cmsValue[i] = saved
	}

	for i := 0; i < len(proof.innerProductProof.l); i++ {
		saved := proof.innerProductProof.l[i]
		// tamper with one field
		proof.innerProductProof.l[i] = operation.RandomPoint()
		// verify using the fast variant
		res, err = proof.Verify()
		Equal(t, false, res)
		NotEqual(t, nil, err)
		proof.innerProductProof.l[i] = saved
	}

	for i := 0; i < len(proof.innerProductProof.r); i++ {
		saved := proof.innerProductProof.r[i]
		// tamper with one field
		proof.innerProductProof.r[i] = operation.RandomPoint()
		// verify using the fast variant
		res, err = proof.Verify()
		Equal(t, false, res)
		NotEqual(t, nil, err)
		proof.innerProductProof.r[i] = saved
	}
}

func TestAggregatedRangeProveVerifyBatch(t *testing.T) {
	count := 10
	proofs := make([]*AggregatedRangeProof, 0)
	bases := make([]*operation.Point, count)
	for i := range bases[:count/2] {
		bases[i] = operation.RandomPoint()
	}

	for i := 0; i < count; i++ {
		//prepare witness for Aggregated range protocol
		wit := new(AggregatedRangeWitness)
		numValue := rand.Intn(privacy_util.MaxOutputCoin)
		values := make([]uint64, numValue)
		rands := make([]*operation.Scalar, numValue)

		for i := range values {
			values[i] = uint64(rand.Uint64())
			rands[i] = operation.RandomScalar()
		}
		wit.Set(values, rands)
		var proof *AggregatedRangeProof
		var err error
		if bases[i] == nil {
			proof, err = wit.Prove()
			Equal(t, nil, err)
			res, err := proof.Verify()
			Equal(t, true, res)
			Equal(t, nil, err)
		} else {
			proof, err = wit.ProveUsingBase(bases[i])
			Equal(t, nil, err)
			res, err := proof.VerifyUsingBase(bases[i])
			Equal(t, true, res)
			Equal(t, nil, err)
		}

		proofs = append(proofs, proof)
	}
	// verify the proof faster
	res, err := VerifyBatch(proofs, bases)
	Equal(t, true, res)
	Equal(t, nil, err)
}

func TestProveVerifyRangeProof(t *testing.T) {
	numOutputs := int(rand.Uint64()%16) + 1
	values := make([]uint64, numOutputs)
	rands := make([]*operation.Scalar, numOutputs)
	rands2 := make([]*operationV1.Scalar, numOutputs)
	for i := range values {
		values[i] = uint64(rand.Uint64())
		rands[i] = operation.RandomScalar()
		rands2[i] = (&operationV1.Scalar{}).FromBytesS(rands[i].ToBytesS())
	}
	// old prover + new verifier
	{
		wit := new(bulletproofsV1.AggregatedRangeWitness)
		wit.Set(values, rands2)
		proof, err := wit.Prove()
		Nil(t, err)
		valid, err := proof.VerifyFaster()
		Nil(t, err)
		True(t, valid)

		proofAgain := &AggregatedRangeProof{}

		// fmt.Printf("proof 1 %x\n", proof.Bytes())
		err = proofAgain.SetBytes(proof.Bytes())
		// fmt.Printf("proof 2 %x\n", proofAgain.Bytes())
		Nil(t, err)
		valid, err = proofAgain.Verify()
		Nil(t, err)
		True(t, valid)
	}

	// new prover + old verifier
	{
		wit := new(AggregatedRangeWitness)
		wit.Set(values, rands)
		proof, err := wit.Prove()
		Nil(t, err)
		valid, err := proof.Verify()
		Nil(t, err)
		True(t, valid)

		proofAgain := &bulletproofsV1.AggregatedRangeProof{}
		// fmt.Printf("proof 1 %x\n", proof.Bytes())
		err = proofAgain.SetBytes(proof.Bytes())
		// fmt.Printf("proof 2 %x\n", proofAgain.Bytes())
		Nil(t, err)
		valid, err = proofAgain.Verify()
		Nil(t, err)
		True(t, valid)
		valid, err = proofAgain.VerifyFaster()
		Nil(t, err)
		True(t, valid)
	}
}
