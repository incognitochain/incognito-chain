package bulletproofs

import (
	"fmt"
	"math"

	// "github.com/incognitochain/incognito-chain/privacy/coin"
	"github.com/incognitochain/incognito-chain/privacy/operation/v1"
	"github.com/incognitochain/incognito-chain/privacy/privacy_util"
)

// CACommitmentScheme defines the Pedersen Commitment Scheme used for Confidential Asset feature.
// var CACommitmentScheme operation.PedersenCommitment = CopyPedersenCommitmentScheme(operation.PedCom)

// CopyPedersenCommitmentScheme is called upon package initialization to make a clone of generators.
func CopyPedersenCommitmentScheme(sch operation.PedersenCommitment) operation.PedersenCommitment {
	var result operation.PedersenCommitment
	var generators []*operation.Point
	for _, gen := range sch.G {
		generators = append(generators, new(operation.Point).Set(gen))
	}
	result.G = generators
	return result
}

// // GetFirstAssetTag is a helper that returns the asset tag field of the first coin from the input.
// // That will be used as base when proving.
// func GetFirstAssetTag(coins []*coin.CoinV2) (*operation.Point, error) {
// 	if len(coins) == 0 {
// 		return nil, fmt.Errorf("cannot get asset tag from empty input")
// 	}
// 	result := coins[0].GetAssetTag()
// 	if result == nil {
// 		return nil, fmt.Errorf("the coin does not have an asset tag")
// 	}
// 	return result, nil
// }

// ProveUsingBase runs like the Bulletproof Prove function, except it sets a Pederson base point before proving.
func (wit AggregatedRangeWitness) ProveUsingBase(anAssetTag *operation.Point) (*AggregatedRangeProof, error) {
	CACommitmentScheme := CopyPedersenCommitmentScheme(operation.PedCom)
	CACommitmentScheme.G[operation.PedersenValueIndex] = anAssetTag
	proof := new(AggregatedRangeProof)
	numValue := len(wit.values)
	if numValue > privacy_util.MaxOutputCoin {
		return nil, fmt.Errorf("output count exceeds MaxOutputCoin")
	}
	numValuePad := roundUpPowTwo(numValue)
	maxExp := privacy_util.MaxExp
	N := maxExp * numValuePad

	aggParam := setAggregateParams(N)

	values := make([]uint64, numValuePad)
	rands := make([]*operation.Scalar, numValuePad)
	for i := range wit.values {
		values[i] = wit.values[i]
		rands[i] = new(operation.Scalar).Set(wit.rands[i])
	}
	for i := numValue; i < numValuePad; i++ {
		values[i] = uint64(0)
		rands[i] = new(operation.Scalar).FromUint64(0)
	}

	// Convert values to binary array
	aL := make([]*operation.Scalar, N)
	aR := make([]*operation.Scalar, N)
	sL := make([]*operation.Scalar, N)
	sR := make([]*operation.Scalar, N)

	for i, value := range values {
		tmp := ConvertUint64ToBinary(value, maxExp)
		for j := 0; j < maxExp; j++ {
			aL[i*maxExp+j] = tmp[j]
			aR[i*maxExp+j] = new(operation.Scalar).Sub(tmp[j], new(operation.Scalar).FromUint64(1))
			sL[i*maxExp+j] = operation.RandomScalar()
			sR[i*maxExp+j] = operation.RandomScalar()
		}
	}
	// LINE 40-50
	// Commitment to aL, aR: A = h^alpha * G^aL * H^aR
	// Commitment to sL, sR : S = h^rho * G^sL * H^sR
	var alpha, rho *operation.Scalar
	if A, err := encodeVectors(aL, aR, aggParam.g, aggParam.h); err != nil {
		return nil, err
	} else if S, err := encodeVectors(sL, sR, aggParam.g, aggParam.h); err != nil {
		return nil, err
	} else {
		alpha = operation.RandomScalar()
		rho = operation.RandomScalar()
		A.Add(A, new(operation.Point).ScalarMult(CACommitmentScheme.G[operation.PedersenRandomnessIndex], alpha))
		S.Add(S, new(operation.Point).ScalarMult(CACommitmentScheme.G[operation.PedersenRandomnessIndex], rho))
		proof.a = A
		proof.s = S
	}
	// challenge y, z
	y := generateChallenge(aggParam.cs.ToBytesS(), []*operation.Point{proof.a, proof.s})
	z := generateChallenge(y.ToBytesS(), []*operation.Point{proof.a, proof.s})

	// LINE 51-54
	twoNumber := new(operation.Scalar).FromUint64(2)
	twoVectorN := powerVector(twoNumber, maxExp)

	// HPrime = H^(y^(1-i)
	HPrime := computeHPrime(y, N, aggParam.h)

	// l(X) = (aL -z*1^n) + sL*X; r(X) = y^n hada (aR +z*1^n + sR*X) + z^2 * 2^n
	yVector := powerVector(y, N)
	hadaProduct, err := hadamardProduct(yVector, vectorAddScalar(aR, z))
	if err != nil {
		return nil, err
	}
	vectorSum := make([]*operation.Scalar, N)
	zTmp := new(operation.Scalar).Set(z)
	for j := 0; j < numValuePad; j++ {
		zTmp.Mul(zTmp, z)
		for i := 0; i < maxExp; i++ {
			vectorSum[j*maxExp+i] = new(operation.Scalar).Mul(twoVectorN[i], zTmp)
		}
	}
	zNeg := new(operation.Scalar).Sub(new(operation.Scalar).FromUint64(0), z)
	l0 := vectorAddScalar(aL, zNeg)
	l1 := sL
	var r0, r1 []*operation.Scalar
	if r0, err = vectorAdd(hadaProduct, vectorSum); err != nil {
		return nil, err
	} else if r1, err = hadamardProduct(yVector, sR); err != nil {
		return nil, err
	}

	// t(X) = <l(X), r(X)> = t0 + t1*X + t2*X^2
	// t1 = <l1, ro> + <l0, r1>, t2 = <l1, r1>
	var t1, t2 *operation.Scalar
	if ip3, err := innerProduct(l1, r0); err != nil {
		return nil, err
	} else if ip4, err := innerProduct(l0, r1); err != nil {
		return nil, err
	} else {
		t1 = new(operation.Scalar).Add(ip3, ip4)
		if t2, err = innerProduct(l1, r1); err != nil {
			return nil, err
		}
	}

	// commitment to t1, t2
	tau1 := operation.RandomScalar()
	tau2 := operation.RandomScalar()
	proof.t1 = CACommitmentScheme.CommitAtIndex(t1, tau1, operation.PedersenValueIndex)
	proof.t2 = CACommitmentScheme.CommitAtIndex(t2, tau2, operation.PedersenValueIndex)

	x := generateChallenge(z.ToBytesS(), []*operation.Point{proof.t1, proof.t2})
	xSquare := new(operation.Scalar).Mul(x, x)

	// lVector = aL - z*1^n + sL*x
	// rVector = y^n hada (aR +z*1^n + sR*x) + z^2*2^n
	// tHat = <lVector, rVector>
	lVector, err := vectorAdd(vectorAddScalar(aL, zNeg), vectorMulScalar(sL, x))
	if err != nil {
		return nil, err
	}
	tmpVector, err := vectorAdd(vectorAddScalar(aR, z), vectorMulScalar(sR, x))
	if err != nil {
		return nil, err
	}
	rVector, err := hadamardProduct(yVector, tmpVector)
	if err != nil {
		return nil, err
	}
	rVector, err = vectorAdd(rVector, vectorSum)
	if err != nil {
		return nil, err
	}
	proof.tHat, err = innerProduct(lVector, rVector)
	if err != nil {
		return nil, err
	}

	// blinding value for tHat: tauX = tau2*x^2 + tau1*x + z^2*rand
	proof.tauX = new(operation.Scalar).Mul(tau2, xSquare)
	proof.tauX.Add(proof.tauX, new(operation.Scalar).Mul(tau1, x))
	zTmp = new(operation.Scalar).Set(z)
	tmpBN := new(operation.Scalar)
	for j := 0; j < numValuePad; j++ {
		zTmp.Mul(zTmp, z)
		proof.tauX.Add(proof.tauX, tmpBN.Mul(zTmp, rands[j]))
	}

	// alpha, rho blind A, S
	// mu = alpha + rho*x
	proof.mu = new(operation.Scalar).Add(alpha, new(operation.Scalar).Mul(rho, x))

	// instead of sending left vector and right vector, we use inner sum argument to reduce proof size from 2*n to 2(log2(n)) + 2
	innerProductWit := new(InnerProductWitness)
	innerProductWit.a = lVector
	innerProductWit.b = rVector
	innerProductWit.p, err = encodeVectors(lVector, rVector, aggParam.g, HPrime)
	if err != nil {
		return nil, err
	}
	uPrime := new(operation.Point).ScalarMult(aggParam.u, operation.HashToScalar(x.ToBytesS()))
	innerProductWit.p = innerProductWit.p.Add(innerProductWit.p, new(operation.Point).ScalarMult(uPrime, proof.tHat))

	proof.innerProductProof, err = innerProductWit.Prove(aggParam.g, HPrime, uPrime, x.ToBytesS())
	if err != nil {
		return nil, err
	}

	return proof, nil
}

// VerifyUsingBase runs like the Bulletproof Verify function, except it sets a Pederson base point before verifying.
func (proof AggregatedRangeProof) VerifyUsingBase(anAssetTag *operation.Point) (bool, error) {
	CACommitmentScheme := CopyPedersenCommitmentScheme(operation.PedCom)
	CACommitmentScheme.G[operation.PedersenValueIndex] = anAssetTag
	numValue := len(proof.cmsValue)
	if numValue > privacy_util.MaxOutputCoin {
		return false, fmt.Errorf("output count exceeds MaxOutputCoin")
	}
	numValuePad := roundUpPowTwo(numValue)
	maxExp := privacy_util.MaxExp
	N := numValuePad * maxExp
	aggParam := setAggregateParams(N)

	cmsValue := proof.cmsValue
	for i := numValue; i < numValuePad; i++ {
		cmsValue = append(cmsValue, new(operation.Point).Identity())
	}

	// recalculate challenge y, z
	y := generateChallenge(aggParam.cs.ToBytesS(), []*operation.Point{proof.a, proof.s})
	z := generateChallenge(y.ToBytesS(), []*operation.Point{proof.a, proof.s})
	zSquare := new(operation.Scalar).Mul(z, z)

	x := generateChallenge(z.ToBytesS(), []*operation.Point{proof.t1, proof.t2})
	xSquare := new(operation.Scalar).Mul(x, x)

	// HPrime = H^(y^(1-i)
	HPrime := computeHPrime(y, N, aggParam.h)

	// g^tHat * h^tauX = V^(z^2) * g^delta(y,z) * T1^x * T2^(x^2)
	yVector := powerVector(y, N)
	deltaYZ, err := computeDeltaYZ(z, zSquare, yVector, N)
	if err != nil {
		return false, err
	}

	LHS := CACommitmentScheme.CommitAtIndex(proof.tHat, proof.tauX, operation.PedersenValueIndex)
	RHS := new(operation.Point).ScalarMult(proof.t2, xSquare)
	RHS.Add(RHS, new(operation.Point).AddPedersen(deltaYZ, CACommitmentScheme.G[operation.PedersenValueIndex], x, proof.t1))

	expVector := vectorMulScalar(powerVector(z, numValuePad), zSquare)
	RHS.Add(RHS, new(operation.Point).MultiScalarMult(expVector, cmsValue))

	if !operation.IsPointEqual(LHS, RHS) {
		Logger.Log.Errorf("verify aggregated range proof statement 1 failed")
		return false, fmt.Errorf("verify aggregated range proof statement 1 failed")
	}
	uPrime := new(operation.Point).ScalarMult(aggParam.u, operation.HashToScalar(x.ToBytesS()))
	innerProductArgValid := proof.innerProductProof.Verify(aggParam.g, HPrime, uPrime, x.ToBytesS())
	if !innerProductArgValid {
		Logger.Log.Errorf("verify aggregated range proof statement 2 failed")
		return false, fmt.Errorf("verify aggregated range proof statement 2 failed")
	}

	return true, nil
}

func (proof AggregatedRangeProof) VerifyFasterUsingBase(anAssetTag *operation.Point) (bool, error) {
	CACommitmentScheme := CopyPedersenCommitmentScheme(operation.PedCom)
	CACommitmentScheme.G[operation.PedersenValueIndex] = anAssetTag
	numValue := len(proof.cmsValue)
	if numValue > privacy_util.MaxOutputCoin {
		return false, fmt.Errorf("output count exceeds MaxOutputCoin")
	}
	numValuePad := roundUpPowTwo(numValue)
	maxExp := privacy_util.MaxExp
	N := maxExp * numValuePad
	aggParam := setAggregateParams(N)

	cmsValue := proof.cmsValue
	for i := numValue; i < numValuePad; i++ {
		cmsValue = append(cmsValue, new(operation.Point).Identity())
	}

	// recalculate challenge y, z
	y := generateChallenge(aggParam.cs.ToBytesS(), []*operation.Point{proof.a, proof.s})
	z := generateChallenge(y.ToBytesS(), []*operation.Point{proof.a, proof.s})
	zSquare := new(operation.Scalar).Mul(z, z)

	x := generateChallenge(z.ToBytesS(), []*operation.Point{proof.t1, proof.t2})
	xSquare := new(operation.Scalar).Mul(x, x)

	// g^tHat * h^tauX = V^(z^2) * g^delta(y,z) * T1^x * T2^(x^2)
	yVector := powerVector(y, N)
	deltaYZ, err := computeDeltaYZ(z, zSquare, yVector, N)
	if err != nil {
		return false, err
	}

	// Verify the first argument
	LHS := CACommitmentScheme.CommitAtIndex(proof.tHat, proof.tauX, operation.PedersenValueIndex)
	RHS := new(operation.Point).ScalarMult(proof.t2, xSquare)
	RHS.Add(RHS, new(operation.Point).AddPedersen(deltaYZ, CACommitmentScheme.G[operation.PedersenValueIndex], x, proof.t1))
	expVector := vectorMulScalar(powerVector(z, numValuePad), zSquare)
	RHS.Add(RHS, new(operation.Point).MultiScalarMult(expVector, cmsValue))
	if !operation.IsPointEqual(LHS, RHS) {
		Logger.Log.Errorf("verify aggregated range proof statement 1 failed")
		return false, fmt.Errorf("verify aggregated range proof statement 1 failed")
	}

	// Verify the second argument
	hashCache := x.ToBytesS()
	L := proof.innerProductProof.l
	R := proof.innerProductProof.r
	s := make([]*operation.Scalar, N)
	sInverse := make([]*operation.Scalar, N)
	logN := int(math.Log2(float64(N)))
	vSquareList := make([]*operation.Scalar, logN)
	vInverseSquareList := make([]*operation.Scalar, logN)

	for i := 0; i < N; i++ {
		s[i] = new(operation.Scalar).Set(proof.innerProductProof.a)
		sInverse[i] = new(operation.Scalar).Set(proof.innerProductProof.b)
	}

	for i := range L {
		v := generateChallenge(hashCache, []*operation.Point{L[i], R[i]})
		hashCache = v.ToBytesS()
		vInverse := new(operation.Scalar).Invert(v)
		vSquareList[i] = new(operation.Scalar).Mul(v, v)
		vInverseSquareList[i] = new(operation.Scalar).Mul(vInverse, vInverse)

		for j := 0; j < N; j++ {
			if j&int(math.Pow(2, float64(logN-i-1))) != 0 {
				s[j] = new(operation.Scalar).Mul(s[j], v)
				sInverse[j] = new(operation.Scalar).Mul(sInverse[j], vInverse)
			} else {
				s[j] = new(operation.Scalar).Mul(s[j], vInverse)
				sInverse[j] = new(operation.Scalar).Mul(sInverse[j], v)
			}
		}
	}
	// HPrime = H^(y^(1-i)
	HPrime := computeHPrime(y, N, aggParam.h)
	uPrime := new(operation.Point).ScalarMult(aggParam.u, operation.HashToScalar(x.ToBytesS()))
	c := new(operation.Scalar).Mul(proof.innerProductProof.a, proof.innerProductProof.b)
	tmp1 := new(operation.Point).MultiScalarMult(s, aggParam.g)
	tmp2 := new(operation.Point).MultiScalarMult(sInverse, HPrime)
	rightHS := new(operation.Point).Add(tmp1, tmp2)
	rightHS.Add(rightHS, new(operation.Point).ScalarMult(uPrime, c))

	tmp3 := new(operation.Point).MultiScalarMult(vSquareList, L)
	tmp4 := new(operation.Point).MultiScalarMult(vInverseSquareList, R)
	leftHS := new(operation.Point).Add(tmp3, tmp4)
	leftHS.Add(leftHS, proof.innerProductProof.p)

	res := operation.IsPointEqual(rightHS, leftHS)
	if !res {
		Logger.Log.Errorf("verify aggregated range proof statement 2 failed")
		return false, fmt.Errorf("verify aggregated range proof statement 2 failed")
	}

	return true, nil
}

// TransformWitnessToCAWitness does base transformation.
// Our Bulletproof(G_r) scheme is parameterized by a base G_r.
// PRV transfers' Bulletproofs use a fixed N.U.M.S point for G_r.
//
// Confidential Asset transfers use G_r = G_at, which is a blinded asset tag.
// This function will return a suitable witness for Bulletproof(G_at).
func TransformWitnessToCAWitness(wit *AggregatedRangeWitness, assetTagBlinders []*operation.Scalar) (*AggregatedRangeWitness, error) {
	if len(assetTagBlinders) != len(wit.values) || len(assetTagBlinders) != len(wit.rands) {
		return nil, fmt.Errorf("cannot transform witness; parameter lengths mismatch")
	}
	newRands := make([]*operation.Scalar, len(wit.values))

	for i := range wit.values {
		temp := new(operation.Scalar).Sub(assetTagBlinders[i], assetTagBlinders[0])
		temp.Mul(temp, new(operation.Scalar).FromUint64(wit.values[i]))
		temp.Add(temp, wit.rands[i])
		newRands[i] = temp
	}
	result := new(AggregatedRangeWitness)
	result.Set(wit.values, newRands)
	return result, nil
}
