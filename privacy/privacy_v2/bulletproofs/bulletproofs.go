// Package bulletproofs manages the creation, proving & verification of Bulletproofs.
// This is a class of compact-sized range proof that require no trusted setup.
package bulletproofs

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/blockchain"
	"math"

	"github.com/incognitochain/incognito-chain/privacy/operation"
	"github.com/incognitochain/incognito-chain/privacy/privacy_util"
)

// AggregatedRangeWitness contains the prover's secret data (the actual values to be proven & the generated random blinders)
// needed for creating a range proof.
type AggregatedRangeWitness struct {
	values []uint64
	rands  []*operation.Scalar
}

// AggregatedRangeProof is the struct for Bulletproof.
// The statement being proven is that output coins' values are in the uint64 range.
type AggregatedRangeProof struct {
	version           uint8
	cmsValue          []*operation.Point
	a                 *operation.Point
	s                 *operation.Point
	t1                *operation.Point
	t2                *operation.Point
	tauX              *operation.Scalar
	tHat              *operation.Scalar
	mu                *operation.Scalar
	innerProductProof *InnerProductProof
}

type bulletproofParams struct {
	g  []*operation.Point
	h  []*operation.Point
	u  *operation.Point
	cs *operation.Point

	precomps []operation.PrecomputedPoint
}

// AggParam contains global Bulletproofs parameters `g, h, u, cs`
var AggParam = newBulletproofParams(privacy_util.MaxOutputCoin)

func (proof AggregatedRangeProof) GetVersion() uint8 {
	return proof.version
}

// ValidateSanity performs sanity checks for this proof.
func (proof AggregatedRangeProof) ValidateSanity() bool {
	for i := 0; i < len(proof.cmsValue); i++ {
		if !proof.cmsValue[i].PointValid() {
			return false
		}
	}
	if !proof.a.PointValid() || !proof.s.PointValid() || !proof.t1.PointValid() || !proof.t2.PointValid() {
		return false
	}
	if !proof.tauX.ScalarValid() || !proof.tHat.ScalarValid() || !proof.mu.ScalarValid() {
		return false
	}

	return proof.innerProductProof.ValidateSanity()
}

// Init allocates all members of an AggregatedRangeProof object
func (proof *AggregatedRangeProof) Init() {
	proof.a = operation.NewIdentityPoint()
	proof.s = operation.NewIdentityPoint()
	proof.t1 = operation.NewIdentityPoint()
	proof.t2 = operation.NewIdentityPoint()
	proof.tauX = new(operation.Scalar)
	proof.tHat = new(operation.Scalar)
	proof.mu = new(operation.Scalar)
	proof.innerProductProof = new(InnerProductProof).Init()
	proof.version = 2
}

// IsNil returns true if any field in this proof is nil
func (proof AggregatedRangeProof) IsNil() bool {
	if proof.a == nil {
		return true
	}
	if proof.s == nil {
		return true
	}
	if proof.t1 == nil {
		return true
	}
	if proof.t2 == nil {
		return true
	}
	if proof.tauX == nil {
		return true
	}
	if proof.tHat == nil {
		return true
	}
	if proof.mu == nil {
		return true
	}
	return proof.innerProductProof == nil
}

func (proof AggregatedRangeProof) GetCommitments() []*operation.Point { return proof.cmsValue }

func (proof *AggregatedRangeProof) SetCommitments(cmsValue []*operation.Point) {
	proof.cmsValue = cmsValue
}

// Bytes marshals the proof into a byte slice
func (proof AggregatedRangeProof) Bytes() []byte {
	var res []byte
	if proof.version >= 2 {
		res = append(res, byte(0))
		res = append(res, byte(proof.version))
	}

	if proof.IsNil() {
		return []byte{}
	}

	res = append(res, byte(len(proof.cmsValue)))
	for i := 0; i < len(proof.cmsValue); i++ {
		res = append(res, proof.cmsValue[i].ToBytesS()...)
	}

	res = append(res, proof.a.ToBytesS()...)
	res = append(res, proof.s.ToBytesS()...)
	res = append(res, proof.t1.ToBytesS()...)
	res = append(res, proof.t2.ToBytesS()...)

	res = append(res, proof.tauX.ToBytesS()...)
	res = append(res, proof.tHat.ToBytesS()...)
	res = append(res, proof.mu.ToBytesS()...)
	res = append(res, proof.innerProductProof.Bytes()...)

	return res
}

// SetBytes unmarshals the proof from a byte slice
func (proof *AggregatedRangeProof) SetBytes(bytes []byte) error {
	if len(bytes) == 0 {
		return nil
	}

	if blockchain.EnableFixBulletProofv2 {
		if bytes[0] == 0 && len(bytes)%operation.Ed25519KeySize == 4 {
			// parse version
			proof.version = uint8(bytes[1])
			bytes = bytes[2:]
		} else {
			proof.version = 1
		}
	} else {
		if bytes[0] == 0 {
			// parse version
			proof.version = uint8(bytes[1])
			bytes = bytes[2:]
		} else {
			proof.version = 1
		}
	}

	lenValues := int(bytes[0])
	offset := 1
	var err error

	proof.cmsValue = make([]*operation.Point, lenValues)
	for i := 0; i < lenValues; i++ {
		if offset+operation.Ed25519KeySize > len(bytes) {
			return fmt.Errorf("range-proof byte unmarshaling failed")
		}
		proof.cmsValue[i], err = new(operation.Point).FromBytesS(bytes[offset : offset+operation.Ed25519KeySize])
		if err != nil {
			return err
		}
		offset += operation.Ed25519KeySize
	}

	if offset+operation.Ed25519KeySize > len(bytes) {
		return fmt.Errorf("range-proof byte unmarshaling failed")
	}
	proof.a, err = new(operation.Point).FromBytesS(bytes[offset : offset+operation.Ed25519KeySize])
	if err != nil {
		return err
	}
	offset += operation.Ed25519KeySize

	if offset+operation.Ed25519KeySize > len(bytes) {
		return fmt.Errorf("range-proof byte unmarshaling failed")
	}
	proof.s, err = new(operation.Point).FromBytesS(bytes[offset : offset+operation.Ed25519KeySize])
	if err != nil {
		return err
	}
	offset += operation.Ed25519KeySize

	if offset+operation.Ed25519KeySize > len(bytes) {
		return fmt.Errorf("range-proof byte unmarshaling failed")
	}
	proof.t1, err = new(operation.Point).FromBytesS(bytes[offset : offset+operation.Ed25519KeySize])
	if err != nil {
		return err
	}
	offset += operation.Ed25519KeySize

	if offset+operation.Ed25519KeySize > len(bytes) {
		return fmt.Errorf("range-proof byte unmarshaling failed")
	}
	proof.t2, err = new(operation.Point).FromBytesS(bytes[offset : offset+operation.Ed25519KeySize])
	if err != nil {
		return err
	}
	offset += operation.Ed25519KeySize

	if offset+operation.Ed25519KeySize > len(bytes) {
		return fmt.Errorf("range-proof byte unmarshaling failed")
	}
	proof.tauX = new(operation.Scalar).FromBytesS(bytes[offset : offset+operation.Ed25519KeySize])
	offset += operation.Ed25519KeySize

	if offset+operation.Ed25519KeySize > len(bytes) {
		return fmt.Errorf("range-proof byte unmarshaling failed")
	}
	proof.tHat = new(operation.Scalar).FromBytesS(bytes[offset : offset+operation.Ed25519KeySize])
	offset += operation.Ed25519KeySize

	if offset+operation.Ed25519KeySize > len(bytes) {
		return fmt.Errorf("range-proof byte unmarshaling failed")
	}
	proof.mu = new(operation.Scalar).FromBytesS(bytes[offset : offset+operation.Ed25519KeySize])
	offset += operation.Ed25519KeySize

	if offset >= len(bytes) {
		return fmt.Errorf("range-proof byte unmarshaling failed")
	}

	proof.innerProductProof = new(InnerProductProof)
	return proof.innerProductProof.SetBytes(bytes[offset:])
}

// Set sets the values of both `wit`'s members
func (wit *AggregatedRangeWitness) Set(values []uint64, rands []*operation.Scalar) {
	numValue := len(values)
	wit.values = make([]uint64, numValue)
	wit.rands = make([]*operation.Scalar, numValue)

	for i := range values {
		wit.values[i] = values[i]
		wit.rands[i] = new(operation.Scalar).Set(rands[i])
	}
}

func (wit AggregatedRangeWitness) Prove() (*AggregatedRangeProof, error) {
	proof := new(AggregatedRangeProof)
	proof.Init()
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

	// Pedersen commitments: V = g^v * h^r
	proof.cmsValue = make([]*operation.Point, numValue)
	initChal := aggParam.cs.ToBytesS()
	for i := 0; i < numValue; i++ {
		proof.cmsValue[i] = operation.PedCom.CommitAtIndex(new(operation.Scalar).FromUint64(values[i]), rands[i], operation.PedersenValueIndex)
		if proof.version >= 2 {
			initChal = append(initChal, proof.cmsValue[i].ToBytesS()...)
		}
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
	// Commitment to aL, aR:  A = h^alpha * G^aL * H^aR
	// Commitment to sL, sR : S = h^rho   * G^sL * H^sR
	var alpha, rho *operation.Scalar
	alpha = operation.RandomScalar()
	rho = operation.RandomScalar()
	mbuilder := operation.NewMultBuilder(false)
	_, err := encodeVectors(aL, aR, aggParam.g, aggParam.h, mbuilder)
	if err != nil {
		return nil, err
	}
	mbuilder.AppendSingle(alpha, operation.HBase)
	proof.a = mbuilder.Eval() // evaluate & clear builder
	_, err = encodeVectors(sL, sR, aggParam.g, aggParam.h, mbuilder)
	if err != nil {
		return nil, err
	}
	mbuilder.AppendSingle(rho, operation.HBase)
	proof.s = mbuilder.Eval()

	// challenge y, z
	y := generateChallenge(initChal, []*operation.Point{proof.a, proof.s})
	z := generateChallenge(y.ToBytesS(), []*operation.Point{proof.a, proof.s})

	// LINE 51-54
	twoNumber := new(operation.Scalar).FromUint64(2)
	twoVectorN := powerVector(twoNumber, maxExp)

	// HPrime = H^(y^(1-i))
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
	proof.t1 = operation.PedCom.CommitAtIndex(t1, tau1, operation.PedersenValueIndex)
	proof.t2 = operation.PedCom.CommitAtIndex(t2, tau2, operation.PedersenValueIndex)

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

	// blinding value for tHat (agg): tauX = tau2 * x^2 + tau1 * x + <z^(1+m), rands>
	proof.tauX = new(operation.Scalar).Mul(tau2, xSquare)
	proof.tauX.Add(proof.tauX, new(operation.Scalar).Mul(tau1, x))
	zTmp = new(operation.Scalar).Set(z)
	tmpBN := new(operation.Scalar)
	for j := 0; j < numValuePad; j++ {
		zTmp.Mul(zTmp, z)
		proof.tauX.Add(proof.tauX, tmpBN.Mul(zTmp, rands[j]))
	}

	// alpha, rho blind A, S
	// mu = alpha + rho * x
	proof.mu = new(operation.Scalar).Add(alpha, new(operation.Scalar).Mul(rho, x))

	// instead of sending left vector and right vector, we use inner sum argument to reduce proof size from 2*n to 2(log2(n)) + 2
	innerProductWit := new(InnerProductWitness)
	innerProductWit.a = lVector
	innerProductWit.b = rVector
	// u' = u^x
	uPrime := new(operation.Point).ScalarMult(aggParam.u, operation.HashToScalar(x.ToBytesS()))

	// for inner-product proof, compute P = g^l * h^r * u ^ (l*r); subsitute h <- h', u <- u'
	_, err = encodeVectors(lVector, rVector, aggParam.g, HPrime, mbuilder)
	if err != nil {
		return nil, err
	}
	mbuilder.AppendSingle(proof.tHat, uPrime)
	innerProductWit.p = mbuilder.Eval()

	proof.innerProductProof, err = innerProductWit.Prove(aggParam.g, HPrime, uPrime, x.ToBytesS())
	if err != nil {
		return nil, err
	}

	return proof, nil
}

//nolint:unused // legacy function for reference
func (proof AggregatedRangeProof) simpleVerify() (bool, error) {
	numValue := len(proof.cmsValue)
	if numValue > privacy_util.MaxOutputCoin {
		return false, fmt.Errorf("output count exceeds MaxOutputCoin")
	}
	numValuePad := roundUpPowTwo(numValue)
	maxExp := privacy_util.MaxExp
	N := numValuePad * maxExp
	twoVectorN := powerVector(new(operation.Scalar).FromUint64(2), maxExp)
	aggParam := setAggregateParams(N)

	cmsValue := proof.cmsValue
	initChal := aggParam.cs.ToBytesS()
	for i := 0; i < numValue; i++ {
		if proof.version >= 2 {
			initChal = append(initChal, proof.cmsValue[i].ToBytesS()...)
		}
	}
	for i := numValue; i < numValuePad; i++ {
		cmsValue = append(cmsValue, operation.NewIdentityPoint())
	}

	// recalculate challenge y, z
	y := generateChallenge(initChal, []*operation.Point{proof.a, proof.s})
	z := generateChallenge(y.ToBytesS(), []*operation.Point{proof.a, proof.s})
	zSquare := new(operation.Scalar).Mul(z, z)
	zNeg := new(operation.Scalar).Sub(new(operation.Scalar).FromUint64(0), z)

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

	LHS := operation.PedCom.CommitAtIndex(proof.tHat, proof.tauX, operation.PedersenValueIndex)
	RHS := new(operation.Point).ScalarMult(proof.t2, xSquare)
	RHS.Add(RHS, new(operation.Point).AddPedersen(deltaYZ, operation.PedCom.G[operation.PedersenValueIndex], x, proof.t1))

	expVector := vectorMulScalar(powerVector(z, numValuePad), zSquare)
	RHS.Add(RHS, new(operation.Point).VarTimeMultiScalarMult(expVector, cmsValue))

	if !operation.IsPointEqual(LHS, RHS) {
		Logger.Log.Errorf("verify aggregated range proof statement 1 failed")
		return false, fmt.Errorf("verify aggregated range proof statement 1 failed")
	}

	// verify eq (66)
	uPrime := new(operation.Point).ScalarMult(aggParam.u, operation.HashToScalar(x.ToBytesS()))

	vectorSum := make([]*operation.Scalar, N)
	zTmp := new(operation.Scalar).Set(z)
	for j := 0; j < numValuePad; j++ {
		zTmp.Mul(zTmp, z)
		for i := 0; i < maxExp; i++ {
			vectorSum[j*maxExp+i] = new(operation.Scalar).Mul(twoVectorN[i], zTmp)
			vectorSum[j*maxExp+i].Add(vectorSum[j*maxExp+i], new(operation.Scalar).Mul(z, yVector[j*maxExp+i]))
		}
	}
	tmpHPrime := new(operation.Point).VarTimeMultiScalarMult(vectorSum, HPrime)
	tmpG := new(operation.Point).Set(aggParam.g[0])
	for i := 1; i < N; i++ {
		tmpG.Add(tmpG, aggParam.g[i])
	}

	ASx := new(operation.Point).Add(proof.a, new(operation.Point).ScalarMult(proof.s, x))
	P := new(operation.Point).Add(new(operation.Point).ScalarMult(tmpG, zNeg), tmpHPrime)
	P.Add(P, ASx)
	P.Add(P, new(operation.Point).ScalarMult(uPrime, proof.tHat))
	PPrime := new(operation.Point).Add(proof.innerProductProof.p, new(operation.Point).ScalarMult(operation.HBase, proof.mu))
	if !operation.IsPointEqual(P, PPrime) {
		Logger.Log.Errorf("verify aggregated range proof statement 2-1 failed")
		return false, fmt.Errorf("verify aggregated range proof statement 2-1 failed")
	}

	// verify eq (68)
	innerProductArgValid := proof.innerProductProof.Verify(aggParam.g, HPrime, uPrime, x.ToBytesS())
	if !innerProductArgValid {
		Logger.Log.Errorf("verify aggregated range proof statement 2 failed")
		return false, fmt.Errorf("verify aggregated range proof statement 2 failed")
	}

	return true, nil
}

// Verify this Bulletproof using an optimized algorithm.
// No view into chain data is needed.
func (proof AggregatedRangeProof) Verify() (bool, error) {
	multBuilder, err := proof.buildVerify(nil)
	if err != nil {
		return false, err
	}
	if !multBuilder.Eval().IsIdentity() {
		Logger.Log.Errorf("Verify aggregated range proof failed")
		return false, fmt.Errorf("bulletproofs: range proof invalid")
	}
	return true, nil
}

func (proof AggregatedRangeProof) buildVerify(gval *operation.Point) (*operation.MultiScalarMultBuilder, error) {
	numValue := len(proof.cmsValue)
	if numValue > privacy_util.MaxOutputCoin {
		return nil, fmt.Errorf("output count exceeds MaxOutputCoin")
	}
	numValuePad := roundUpPowTwo(numValue)
	maxExp := privacy_util.MaxExp
	N := maxExp * numValuePad
	aggParam := setAggregateParams(N)
	twoVectorN := powerVector(new(operation.Scalar).FromUint64(2), maxExp)

	// pad commitment list to have power of 2 for length
	cmsValue := proof.cmsValue
	for i := numValue; i < numValuePad; i++ {
		cmsValue = append(cmsValue, operation.NewIdentityPoint())
	}
	initChal := aggParam.cs.ToBytesS()
	for i := 0; i < numValue; i++ {
		if proof.version >= 2 {
			initChal = append(initChal, proof.cmsValue[i].ToBytesS()...)
		}
	}

	// recalculate challenge y, z
	y := generateChallenge(initChal, []*operation.Point{proof.a, proof.s})
	z := generateChallenge(y.ToBytesS(), []*operation.Point{proof.a, proof.s})
	zSquare := new(operation.Scalar).Mul(z, z)
	zNeg := new(operation.Scalar).Sub(new(operation.Scalar).FromUint64(0), z)

	x := generateChallenge(z.ToBytesS(), []*operation.Point{proof.t1, proof.t2})
	xSquare := new(operation.Scalar).Mul(x, x)

	// compute delta(y,z) = (z-z^2) * <1, y^N> - z^3 * <1, 2^N>
	yVector := powerVector(y, N)
	deltaYZ, err := computeDeltaYZ(z, zSquare, yVector, N)
	if err != nil {
		return nil, err
	}

	// Verify eq (65) (agg): g^tHat * h^tauX = V^(z^2 * z^m) * g^delta(y,z) * T1^x * T2^(x^2)
	eq65Builder := operation.NewMultBuilder(true)
	eq65Builder.WithStaticPoints(AggParam.precomps)
	// set builder to V^(z^2 * z^m) * T1^x * T2^(x^2)
	eq65Builder.AppendSingle(xSquare, proof.t2)
	eq65Builder.AppendSingle(x, proof.t1)
	expVector := vectorMulScalar(powerVector(z, numValuePad), zSquare)
	eq65Builder.MustAppend(expVector, cmsValue)

	// g <- gval. Unite terms to RHS
	if gval != nil {
		eq65Builder.AppendSingle(operation.NewScalar().Sub(deltaYZ, proof.tHat), gval)
	} else {
		eq65Builder.MustSetStatic(precompPedGValIndex, operation.NewScalar().Sub(deltaYZ, proof.tHat))
	}
	eq65Builder.MustSetStatic(precompPedGRandIndex, operation.NewScalar().Negate(proof.tauX))
	// identity check will be combined

	// Verify eq (67): A * S^x * g^(-z) * h' ^ (z * y^n + z^2 * 2^n) = h^mu * g^l * h'^r
	eq66Builder := operation.NewMultBuilder(true)
	eq66Builder.WithStaticPoints(AggParam.precomps)

	vectorSum := make([]*operation.Scalar, N)
	zTmp := new(operation.Scalar).Set(z)
	for j := 0; j < numValuePad; j++ {
		zTmp.Mul(zTmp, z)
		for i := 0; i < maxExp; i++ {
			vectorSum[j*maxExp+i] = new(operation.Scalar).Mul(twoVectorN[i], zTmp)
			vectorSum[j*maxExp+i].Add(vectorSum[j*maxExp+i], new(operation.Scalar).Mul(z, yVector[j*maxExp+i]))
		}
	}
	// add h' ^ vector_sum to builder by substituting h' = h^(y^(1-i)) and using h as bases
	mulPowerVector(vectorSum, operation.NewScalar().Invert(y))
	eq66Builder.MustSetStatic(precompHIndex(aggParamNMax), vectorSum...)
	// g^(-z)
	for i := 0; i < N; i++ {
		eq66Builder.MustSetStatic(precompGIndex+i, zNeg)
	}

	eq66Builder.MustAppend([]*operation.Scalar{operation.NewScalar().FromUint64(1), x}, []*operation.Point{proof.a, proof.s}) // AS^x
	// add u' ^ t_hat to both sides
	// RHS is then simplified using g^l * h'^r * u' ^ t_hat = P (of inner product proof)
	eq66Builder.MustSetStatic(precompUIndex, operation.NewScalar().Mul(proof.tHat, operation.HashToScalar(x.ToBytesS()))) // tHat.U'
	// unite terms to LHS
	eq66Builder.MustSetStatic(precompPedGRandIndex, operation.NewScalar().Negate(proof.mu))
	eq66Builder.AppendSingle(operation.NewScalar().Set(operation.ScMinusOne), proof.innerProductProof.p)
	// identity check will be combined

	// Verify eq (68): inner product proof for t_hat = l * r
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

	// LHS = g^s * h' ^ s_inv * u' ^ c, substituting h' = h^(y^(1-i))
	ippBuilder := operation.NewMultBuilder(true)
	ippBuilder.WithStaticPoints(AggParam.precomps)
	ippBuilder.MustSetStatic(precompGIndex, s...)
	mulPowerVector(sInverse, operation.NewScalar().Invert(y))
	ippBuilder.MustSetStatic(precompHIndex(aggParamNMax), sInverse...)
	c := new(operation.Scalar).Mul(proof.innerProductProof.a, proof.innerProductProof.b)
	ippBuilder.MustSetStatic(precompUIndex, operation.NewScalar().Mul(c, operation.HashToScalar(x.ToBytesS())))

	// compute RHS = L ^ (v^2) * P * R * (v^-2)
	rhsBuilder := operation.NewMultBuilder(true)
	rhsBuilder.MustAppend(vSquareList, L)
	rhsBuilder.MustAppend(vInverseSquareList, R)
	rhsBuilder.AppendSingle(operation.NewScalar().FromUint64(1), proof.innerProductProof.p)

	err = ippBuilder.ConcatScaled(rhsBuilder, operation.ScMinusOne)
	if err != nil {
		return nil, err
	}
	// perform many identity checks simultaneously by multplying each one with a random scalar
	check := eq65Builder
	err = check.ConcatScaled(eq66Builder, operation.RandomScalar())
	if err != nil {
		return nil, err
	}
	err = check.ConcatScaled(ippBuilder, operation.RandomScalar())
	if err != nil {
		return nil, err
	}

	return check, nil
}

// VerifyBatch verifies a list of Bulletproofs in batched fashion.
// It saves time by using a multi-exponent operation.
func VerifyBatch(proofs []*AggregatedRangeProof, gvalLst []*operation.Point) (bool, error) {
	var check *operation.MultiScalarMultBuilder = nil
	for i := range proofs {
		multBuilder, err := proofs[i].buildVerify(gvalLst[i])
		if err != nil {
			return false, err
		}
		if check == nil {
			check = multBuilder
		} else {
			// perform many identity checks simultaneously by multplying each one with a random scalar
			err := check.ConcatScaled(multBuilder, operation.RandomScalar())
			if err != nil {
				return false, err
			}
		}
	}
	if check != nil && !check.Eval().IsIdentity() {
		Logger.Log.Errorf("Verify batch aggregated range proof failed")
		return false, fmt.Errorf("bulletproofs: batch range proof invalid")
	}
	return true, nil
}

// EstimateMultiRangeProofSize returns the upper bound of Bulletproof size given the number of output coins.
func EstimateMultiRangeProofSize(nOutput int) uint64 {
	return uint64((nOutput+2*int(math.Log2(float64(privacy_util.MaxExp*roundUpPowTwo(nOutput))))+5)*operation.Ed25519KeySize + 5*operation.Ed25519KeySize + 2)
}
