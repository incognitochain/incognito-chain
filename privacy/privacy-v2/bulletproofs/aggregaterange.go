package bulletproofs

import (
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/pkg/errors"
	"math"
)

const (
	maxExp               = 64
	numCommitValue       = 5
	maxOutputNumber      = 32
	maxOutputNumberParam = 256
)

type AggregatedRangeWitness struct {
	values []uint64
	rands  []*privacy.Scalar
}

type AggregatedRangeProof struct {
	cmsValue          []*privacy.Point
	a                 *privacy.Point
	s                 *privacy.Point
	t1                *privacy.Point
	t2                *privacy.Point
	tauX              *privacy.Scalar
	tHat              *privacy.Scalar
	mu                *privacy.Scalar
	innerProductProof *InnerProductProof
}

type bulletproofParams struct {
	g  []*privacy.Point
	h  []*privacy.Point
	u  *privacy.Point
	cs *privacy.Point
}

var AggParam = newBulletproofParams(maxOutputNumber)

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

func (proof *AggregatedRangeProof) Init() {
	proof.a = new(privacy.Point).Identity()
	proof.s = new(privacy.Point).Identity()
	proof.t1 = new(privacy.Point).Identity()
	proof.t2 = new(privacy.Point).Identity()
	proof.tauX = new(privacy.Scalar)
	proof.tHat = new(privacy.Scalar)
	proof.mu = new(privacy.Scalar)
	proof.innerProductProof = new(InnerProductProof)
}

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

func (proof AggregatedRangeProof) Bytes() []byte {
	var res []byte

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

func (proof *AggregatedRangeProof) SetBytes(bytes []byte) error {
	if len(bytes) == 0 {
		return nil
	}

	lenValues := int(bytes[0])
	offset := 1
	var err error

	proof.cmsValue = make([]*privacy.Point, lenValues)
	for i := 0; i < lenValues; i++ {
		proof.cmsValue[i], err = new(privacy.Point).FromBytesS(bytes[offset : offset+privacy.Ed25519KeySize])
		if err != nil {
			return err
		}
		offset += privacy.Ed25519KeySize
	}

	proof.a, err = new(privacy.Point).FromBytesS(bytes[offset : offset+privacy.Ed25519KeySize])
	if err != nil {
		return err
	}
	offset += privacy.Ed25519KeySize

	proof.s, err = new(privacy.Point).FromBytesS(bytes[offset : offset+privacy.Ed25519KeySize])
	if err != nil {
		return err
	}
	offset += privacy.Ed25519KeySize

	proof.t1, err = new(privacy.Point).FromBytesS(bytes[offset : offset+privacy.Ed25519KeySize])
	if err != nil {
		return err
	}
	offset += privacy.Ed25519KeySize

	proof.t2, err = new(privacy.Point).FromBytesS(bytes[offset : offset+privacy.Ed25519KeySize])
	if err != nil {
		return err
	}
	offset += privacy.Ed25519KeySize

	proof.tauX = new(privacy.Scalar).FromBytesS(bytes[offset : offset+privacy.Ed25519KeySize])
	offset += privacy.Ed25519KeySize

	proof.tHat = new(privacy.Scalar).FromBytesS(bytes[offset : offset+privacy.Ed25519KeySize])
	offset += privacy.Ed25519KeySize

	proof.mu = new(privacy.Scalar).FromBytesS(bytes[offset : offset+privacy.Ed25519KeySize])
	offset += privacy.Ed25519KeySize

	proof.innerProductProof = new(InnerProductProof)
	proof.innerProductProof.SetBytes(bytes[offset:])

	//privacy.Logger.Log.Debugf("AFTER SETBYTES ------------ %v\n", proof.Bytes())
	return nil
}

func (wit *AggregatedRangeWitness) Set(values []uint64, rands []*privacy.Scalar) {
	numValue := len(values)
	wit.values = make([]uint64, numValue)
	wit.rands = make([]*privacy.Scalar, numValue)

	for i := range values {
		wit.values[i] = values[i]
		wit.rands[i] = new(privacy.Scalar).Set(rands[i])
	}
}

func (wit AggregatedRangeWitness) Prove() (*AggregatedRangeProof, error) {
	proof := new(AggregatedRangeProof)
	numValue := len(wit.values)
	if numValue > maxOutputNumber {
		return nil, errors.New("Must less than maxOutputNumber")
	}
	numValuePad := pad(numValue)
	N := maxExp * numValuePad
	aggParam := setAggregateParams(N)

	values := make([]uint64, numValuePad)
	rands := make([]*privacy.Scalar, numValuePad)
	for i := range wit.values {
		values[i] = wit.values[i]
		rands[i] = new(privacy.Scalar).Set(wit.rands[i])
	}
	for i := numValue; i < numValuePad; i++ {
		values[i] = uint64(0)
		rands[i] = new(privacy.Scalar).FromUint64(0)
	}

	proof.cmsValue = make([]*privacy.Point, numValue)
	for i := 0; i < numValue; i++ {
		proof.cmsValue[i] = privacy.PedCom.CommitAtIndex(new(privacy.Scalar).FromUint64(values[i]), rands[i], privacy.PedersenValueIndex)
	}
	// Convert values to binary array
	aL := make([]*privacy.Scalar, N)
	aR := make([]*privacy.Scalar, N)
	sL := make([]*privacy.Scalar, N)
	sR := make([]*privacy.Scalar, N)

	for i, value := range values {
		tmp := privacy.ConvertUint64ToBinary(value, maxExp)
		for j := 0; j < maxExp; j++ {
			aL[i*maxExp+j] = tmp[j]
			aR[i*maxExp+j] = new(privacy.Scalar).Sub(tmp[j], new(privacy.Scalar).FromUint64(1))
			sL[i*maxExp+j] = privacy.RandomScalar()
			sR[i*maxExp+j] = privacy.RandomScalar()
		}
	}
	// LINE 40-50
	// Commitment to aL, aR: A = h^alpha * G^aL * H^aR
	// Commitment to sL, sR : S = h^rho * G^sL * H^sR
	var alpha, rho *privacy.Scalar
	if A, err := encodeVectors(aL, aR, aggParam.g, aggParam.h); err != nil {
		return nil, err
	} else if S, err := encodeVectors(sL, sR, aggParam.g, aggParam.h); err != nil {
		return nil, err
	} else {
		alpha = privacy.RandomScalar()
		rho = privacy.RandomScalar()
		A.Add(A, new(privacy.Point).ScalarMult(privacy.PedCom.G[privacy.PedersenRandomnessIndex], alpha))
		S.Add(S, new(privacy.Point).ScalarMult(privacy.PedCom.G[privacy.PedersenRandomnessIndex], rho))
		proof.a = A
		proof.s = S
	}
	// challenge y, z
	y := generateChallenge(aggParam.cs.ToBytesS(), []*privacy.Point{proof.a, proof.s})
	z := generateChallenge(y.ToBytesS(), []*privacy.Point{proof.a, proof.s})

	// LINE 51-54
	twoNumber := new(privacy.Scalar).FromUint64(2)
	twoVectorN := powerVector(twoNumber, maxExp)

	// HPrime = H^(y^(1-i)
	HPrime := computeHPrime(y, N, aggParam.h)

	// l(X) = (aL -z*1^n) + sL*X; r(X) = y^n hada (aR +z*1^n + sR*X) + z^2 * 2^n
	yVector := powerVector(y, N)
	hadaProduct, err := hadamardProduct(yVector, vectorAddScalar(aR, z))
	if err != nil {
		return nil, err
	}
	vectorSum := make([]*privacy.Scalar, N)
	zTmp := new(privacy.Scalar).Set(z)
	for j := 0; j < numValuePad; j++ {
		zTmp.Mul(zTmp, z)
		for i := 0; i < maxExp; i++ {
			vectorSum[j*maxExp+i] = new(privacy.Scalar).Mul(twoVectorN[i], zTmp)
		}
	}
	zNeg := new(privacy.Scalar).Sub(new(privacy.Scalar).FromUint64(0), z)
	l0 := vectorAddScalar(aL, zNeg)
	l1 := sL
	var r0, r1 []*privacy.Scalar
	if r0, err = vectorAdd(hadaProduct, vectorSum); err != nil {
		return nil, err
	} else {
		if r1, err = hadamardProduct(yVector, sR); err != nil {
			return nil, err
		}
	}

	// t(X) = <l(X), r(X)> = t0 + t1*X + t2*X^2
	// t1 = <l1, ro> + <l0, r1>, t2 = <l1, r1>
	var t1, t2 *privacy.Scalar
	if ip3, err := innerProduct(l1, r0); err != nil {
		return nil, err
	} else if ip4, err := innerProduct(l0, r1); err != nil {
		return nil, err
	} else {
		t1 = new(privacy.Scalar).Add(ip3, ip4)
		if t2, err = innerProduct(l1, r1); err != nil {
			return nil, err
		}
	}

	// commitment to t1, t2
	tau1 := privacy.RandomScalar()
	tau2 := privacy.RandomScalar()
	proof.t1 = privacy.PedCom.CommitAtIndex(t1, tau1, privacy.PedersenValueIndex)
	proof.t2 = privacy.PedCom.CommitAtIndex(t2, tau2, privacy.PedersenValueIndex)

	x := generateChallenge(z.ToBytesS(), []*privacy.Point{proof.t1, proof.t2})
	xSquare := new(privacy.Scalar).Mul(x, x)

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
	proof.tauX = new(privacy.Scalar).Mul(tau2, xSquare)
	proof.tauX.Add(proof.tauX, new(privacy.Scalar).Mul(tau1, x))
	zTmp = new(privacy.Scalar).Set(z)
	tmpBN := new(privacy.Scalar)
	for j := 0; j < numValuePad; j++ {
		zTmp.Mul(zTmp, z)
		proof.tauX.Add(proof.tauX, tmpBN.Mul(zTmp, rands[j]))
	}

	// alpha, rho blind A, S
	// mu = alpha + rho*x
	proof.mu = new(privacy.Scalar).Add(alpha, new(privacy.Scalar).Mul(rho, x))

	// instead of sending left vector and right vector, we use inner sum argument to reduce proof size from 2*n to 2(log2(n)) + 2
	innerProductWit := new(InnerProductWitness)
	innerProductWit.a = lVector
	innerProductWit.b = rVector
	innerProductWit.p, err = encodeVectors(lVector, rVector, aggParam.g, HPrime)
	if err != nil {
		return nil, err
	}
	uPrime := new(privacy.Point).ScalarMult(aggParam.u, privacy.HashToScalar(x.ToBytesS()))
	innerProductWit.p = innerProductWit.p.Add(innerProductWit.p, new(privacy.Point).ScalarMult(uPrime, proof.tHat))

	proof.innerProductProof, err = innerProductWit.Prove(aggParam.g, HPrime, uPrime, x.ToBytesS())
	if err != nil {
		return nil, err
	}

	return proof, nil
}

func (proof AggregatedRangeProof) Verify() (bool, error) {
	numValue := len(proof.cmsValue)
	if numValue > maxOutputNumber {
		return false, errors.New("Must less than maxOutputNumber")
	}
	numValuePad := pad(numValue)
	N := numValuePad * maxExp
	aggParam := setAggregateParams(N)

	cmsValue := proof.cmsValue
	for i := numValue; i < numValuePad; i++ {
		cmsValue = append(cmsValue, new(privacy.Point).Identity())
	}

	// recalculate challenge y, z
	y := generateChallenge(aggParam.cs.ToBytesS(), []*privacy.Point{proof.a, proof.s})
	z := generateChallenge(y.ToBytesS(), []*privacy.Point{proof.a, proof.s})
	zSquare := new(privacy.Scalar).Mul(z, z)

	x := generateChallenge(z.ToBytesS(), []*privacy.Point{proof.t1, proof.t2})
	xSquare := new(privacy.Scalar).Mul(x, x)

	// HPrime = H^(y^(1-i)
	HPrime := computeHPrime(y, N, aggParam.h)

	// g^tHat * h^tauX = V^(z^2) * g^delta(y,z) * T1^x * T2^(x^2)
	yVector := powerVector(y, N)
	deltaYZ, err := computeDeltaYZ(z, zSquare, yVector, N)
	if err != nil {
		return false, err
	}

	LHS := privacy.PedCom.CommitAtIndex(proof.tHat, proof.tauX, privacy.PedersenValueIndex)
	RHS := new(privacy.Point).ScalarMult(proof.t2, xSquare)
	RHS.Add(RHS, new(privacy.Point).AddPedersen(deltaYZ, privacy.PedCom.G[privacy.PedersenValueIndex], x, proof.t1))

	expVector := vectorMulScalar(powerVector(z, numValuePad), zSquare)
	RHS.Add(RHS, new(privacy.Point).MultiScalarMult(expVector, cmsValue))

	if !privacy.IsPointEqual(LHS, RHS) {
		privacy.Logger.Log.Errorf("verify aggregated range proof statement 1 failed")
		return false, errors.New("verify aggregated range proof statement 1 failed")
	}
	uPrime := new(privacy.Point).ScalarMult(aggParam.u, privacy.HashToScalar(x.ToBytesS()))
	innerProductArgValid := proof.innerProductProof.Verify(aggParam.g, HPrime, uPrime, x.ToBytesS())
	if !innerProductArgValid {
		privacy.Logger.Log.Errorf("verify aggregated range proof statement 2 failed")
		return false, errors.New("verify aggregated range proof statement 2 failed")
	}

	return true, nil
}

func (proof AggregatedRangeProof) VerifyFaster() (bool, error) {
	numValue := len(proof.cmsValue)
	if numValue > maxOutputNumber {
		return false, errors.New("Must less than maxOutputNumber")
	}
	numValuePad := pad(numValue)
	N := maxExp * numValuePad
	aggParam := setAggregateParams(N)

	cmsValue := proof.cmsValue
	for i := numValue; i < numValuePad; i++ {
		cmsValue = append(cmsValue, new(privacy.Point).Identity())
	}

	// recalculate challenge y, z
	y := generateChallenge(aggParam.cs.ToBytesS(), []*privacy.Point{proof.a, proof.s})
	z := generateChallenge(y.ToBytesS(), []*privacy.Point{proof.a, proof.s})
	zSquare := new(privacy.Scalar).Mul(z, z)

	x := generateChallenge(z.ToBytesS(), []*privacy.Point{proof.t1, proof.t2})
	xSquare := new(privacy.Scalar).Mul(x, x)

	// g^tHat * h^tauX = V^(z^2) * g^delta(y,z) * T1^x * T2^(x^2)
	yVector := powerVector(y, N)
	deltaYZ, err := computeDeltaYZ(z, zSquare, yVector, N)
	if err != nil {
		return false, err
	}

	// Verify the first argument
	LHS := privacy.PedCom.CommitAtIndex(proof.tHat, proof.tauX, privacy.PedersenValueIndex)
	RHS := new(privacy.Point).ScalarMult(proof.t2, xSquare)
	RHS.Add(RHS, new(privacy.Point).AddPedersen(deltaYZ, privacy.PedCom.G[privacy.PedersenValueIndex], x, proof.t1))
	expVector := vectorMulScalar(powerVector(z, numValuePad), zSquare)
	RHS.Add(RHS, new(privacy.Point).MultiScalarMult(expVector, cmsValue))
	if !privacy.IsPointEqual(LHS, RHS) {
		privacy.Logger.Log.Errorf("verify aggregated range proof statement 1 failed")
		return false, errors.New("verify aggregated range proof statement 1 failed")
	}

	// Verify the second argument
	hashCache := x.ToBytesS()
	L := proof.innerProductProof.l
	R := proof.innerProductProof.r
	s := make([]*privacy.Scalar, N)
	sInverse := make([]*privacy.Scalar, N)
	logN := int(math.Log2(float64(N)))
	vSquareList := make([]*privacy.Scalar, logN)
	vInverseSquareList := make([]*privacy.Scalar, logN)

	for i := 0; i < N; i ++ {
		s[i] = new(privacy.Scalar).Set(proof.innerProductProof.a)
		sInverse[i] = new(privacy.Scalar).Set(proof.innerProductProof.b)
	}

	for i := range L {
		v := generateChallenge(hashCache, []*privacy.Point{L[i], R[i]})
		hashCache = v.ToBytesS()
		vInverse := new(privacy.Scalar).Invert(v)
		vSquareList[i] = new(privacy.Scalar).Mul(v,v)
		vInverseSquareList[i] = new(privacy.Scalar).Mul(vInverse, vInverse)

		for j :=  0; j < N; j++ {
			if j&int(math.Pow(2, float64(logN-i-1))) != 0 {
				s[j] = new(privacy.Scalar).Mul(s[j], v)
				sInverse[j] = new(privacy.Scalar).Mul(sInverse[j], vInverse)
			} else {
				s[j] = new(privacy.Scalar).Mul(s[j], vInverse)
				sInverse[j] = new(privacy.Scalar).Mul(sInverse[j], v)
			}
		}
	}
	// HPrime = H^(y^(1-i)
	HPrime := computeHPrime(y, N, aggParam.h)
	uPrime := new(privacy.Point).ScalarMult(aggParam.u, privacy.HashToScalar(x.ToBytesS()))
	c := new(privacy.Scalar).Mul(proof.innerProductProof.a, proof.innerProductProof.b)
	tmp1 := new(privacy.Point).MultiScalarMult(s, aggParam.g)
	tmp2 := new(privacy.Point).MultiScalarMult(sInverse, HPrime)
	rightHS := new(privacy.Point).Add(tmp1, tmp2)
	rightHS.Add(rightHS, new(privacy.Point).ScalarMult(uPrime, c))

	tmp3 := new(privacy.Point).MultiScalarMult(vSquareList, L)
	tmp4 := new(privacy.Point).MultiScalarMult(vInverseSquareList, R)
	leftHS := new(privacy.Point).Add(tmp3, tmp4)
	leftHS.Add(leftHS, proof.innerProductProof.p)

	res := privacy.IsPointEqual(rightHS, leftHS)
	if !res {
		privacy.Logger.Log.Errorf("verify aggregated range proof statement 2 failed")
		return false, errors.New("verify aggregated range proof statement 2 failed")
	}

	return true, nil
}


func VerifyBatch(proofs []*AggregatedRangeProof) (bool, error, int) {
	baseG := privacy.PedCom.G[privacy.PedersenValueIndex]
	baseH := privacy.PedCom.G[privacy.PedersenRandomnessIndex]

	sum_tHat := new(privacy.Scalar).FromUint64(0)
	sum_tauX := new(privacy.Scalar).FromUint64(0)
	list_x_alpha := make([]*privacy.Scalar, 0)
	list_x_beta := make([]*privacy.Scalar, 0)
	list_xSquare := make([]*privacy.Scalar, 0)
	list_zSquare := make([]*privacy.Scalar,0)

	list_t1 := make([]*privacy.Point, 0)
	list_t2 := make([]*privacy.Point, 0)
	list_V := make([]*privacy.Point, 0)

	sum_mu := new(privacy.Scalar).FromUint64(0)
	sum_absubthat := new(privacy.Scalar).FromUint64(0)

	list_S := make([]*privacy.Point, 0)
	list_A := make([]*privacy.Point, 0)
	list_beta := make([]*privacy.Scalar, 0)
	list_LR := make([]*privacy.Point, 0)
	list_lVector := make([]*privacy.Scalar, 0)
	list_rVector := make([]*privacy.Scalar, 0)
	list_gVector := make([]*privacy.Point, 0)
	list_hVector := make([]*privacy.Point, 0)

	twoNumber := new(privacy.Scalar).FromUint64(2)
	twoVectorN := powerVector(twoNumber, maxExp)

	for k, proof := range proofs {
		numValue := len(proof.cmsValue)
		if numValue > maxOutputNumber {
			return false, errors.New("Must less than maxOutputNumber"), k
		}
		numValuePad := pad(numValue)
		N := maxExp * numValuePad
		aggParam := setAggregateParams(N)

		cmsValue := proof.cmsValue
		for i := numValue; i < numValuePad; i++ {
			identity := new(privacy.Point).Identity()
			cmsValue = append(cmsValue, identity)
		}

		// recalculate challenge y, z, x
		y := generateChallenge(aggParam.cs.ToBytesS(), []*privacy.Point{proof.a, proof.s})
		z := generateChallenge(y.ToBytesS(), []*privacy.Point{proof.a, proof.s})
		x := generateChallenge(z.ToBytesS(), []*privacy.Point{proof.t1, proof.t2})
		zSquare := new(privacy.Scalar).Mul(z, z)
		xSquare := new(privacy.Scalar).Mul(x, x)

		// Random alpha and beta for batch equations check
		alpha := privacy.RandomScalar()
		beta := privacy.RandomScalar()
		list_beta = append(list_beta, beta)

		// Compute first equation check
		yVector := powerVector(y, N)
		deltaYZ, err := computeDeltaYZ(z, zSquare, yVector, N)
		if err != nil {
			return false, err, k
		}
		sum_tHat.Add(sum_tHat, new(privacy.Scalar).Mul(alpha, new(privacy.Scalar).Sub(proof.tHat, deltaYZ)))
		sum_tauX.Add(sum_tauX, new(privacy.Scalar).Mul(alpha, proof.tauX))

		list_x_alpha =append(list_x_alpha, new(privacy.Scalar).Mul(x, alpha))
		list_x_beta = append(list_x_beta, new(privacy.Scalar).Mul(x, beta))
		list_xSquare = append(list_xSquare, new(privacy.Scalar).Mul(xSquare, alpha))
		tmp := vectorMulScalar(powerVector(z, numValuePad), new(privacy.Scalar).Mul(zSquare, alpha))
		list_zSquare = append(list_zSquare, tmp...)

		list_V = append(list_V, cmsValue...)
		list_t1 = append(list_t1, proof.t1)
		list_t2 = append(list_t2, proof.t2)

		// Verify the second argument
		hashCache := x.ToBytesS()
		L := proof.innerProductProof.l
		R := proof.innerProductProof.r
		s := make([]*privacy.Scalar, N)
		sInverse := make([]*privacy.Scalar, N)
		logN := int(math.Log2(float64(N)))
		vSquareList := make([]*privacy.Scalar, logN)
		vInverseSquareList := make([]*privacy.Scalar, logN)

		for i := 0; i < N; i ++ {
			s[i] = new(privacy.Scalar).Set(proof.innerProductProof.a)
			sInverse[i] = new(privacy.Scalar).Set(proof.innerProductProof.b)
		}

		for i := range L {
			v := generateChallenge(hashCache, []*privacy.Point{L[i], R[i]})
			hashCache = v.ToBytesS()
			vInverse := new(privacy.Scalar).Invert(v)
			vSquareList[i] = new(privacy.Scalar).Mul(v,v)
			vInverseSquareList[i] = new(privacy.Scalar).Mul(vInverse, vInverse)

			for j :=  0; j < N; j++ {
				if j&int(math.Pow(2, float64(logN-i-1))) != 0 {
					s[j] = new(privacy.Scalar).Mul(s[j], v)
					sInverse[j] = new(privacy.Scalar).Mul(sInverse[j], vInverse)
				} else {
					s[j] = new(privacy.Scalar).Mul(s[j], vInverse)
					sInverse[j] = new(privacy.Scalar).Mul(sInverse[j], v)
				}
			}
		}

		lVector := make([]*privacy.Scalar, N)
		rVector := make([]*privacy.Scalar, N)

		vectorSum := make([]*privacy.Scalar, N)
		zTmp := new(privacy.Scalar).Set(z)
		for j := 0; j < numValuePad; j++ {
			zTmp.Mul(zTmp, z)
			for i := 0; i < maxExp; i++ {
				vectorSum[j*maxExp+i] = new(privacy.Scalar).Mul(twoVectorN[i], zTmp)
			}
		}
		yInverse :=  new(privacy.Scalar).Invert(y)
		yTmp := new(privacy.Scalar).Set(y)
		for j:= 0; j < N; j++ {
			yTmp.Mul(yTmp, yInverse)
			lVector[j] = new(privacy.Scalar).Add(s[j], z)
			rVector[j] = new(privacy.Scalar).Sub(sInverse[j], vectorSum[j])
			rVector[j].Mul(rVector[j], yTmp)
			rVector[j].Sub(rVector[j], z)

			lVector[j].Mul(lVector[j], beta)
			rVector[j].Mul(rVector[j], beta)
		}

		list_lVector = append(list_lVector, lVector...)
		list_rVector = append(list_rVector, rVector...)

		tmp1 := new(privacy.Point).MultiScalarMult(vSquareList, L)
		tmp2 := new(privacy.Point).MultiScalarMult(vInverseSquareList, R)
		list_LR = append(list_LR, new(privacy.Point).Add(tmp1, tmp2))

		list_gVector = append(list_gVector, aggParam.g...)
		list_hVector = append(list_hVector, aggParam.h...)

		sum_mu.Add(sum_mu, new(privacy.Scalar).Mul(proof.mu, beta))
		ab := new(privacy.Scalar).Mul(proof.innerProductProof.a, proof.innerProductProof.b)
		absubthat := new(privacy.Scalar).Sub(ab, proof.tHat)
		absubthat.Mul(absubthat, privacy.HashToScalar(x.ToBytesS()))
		sum_absubthat.Add(sum_absubthat, new(privacy.Scalar).Mul(absubthat, beta))
		list_A = append(list_A, proof.a)
		list_S = append(list_S, proof.s)
	}

	tmp1 := new(privacy.Point).MultiScalarMult(list_lVector, list_gVector)
	tmp2 := new(privacy.Point).MultiScalarMult(list_rVector, list_hVector)
	tmp3 := new(privacy.Point).ScalarMult(AggParam.u, sum_absubthat)
	tmp4 := new(privacy.Point).ScalarMult(baseH, sum_mu)
	LHSPrime := new(privacy.Point).Add(tmp1, tmp2)
	LHSPrime.Add(LHSPrime, tmp3)
	LHSPrime.Add(LHSPrime, tmp4)

	LHS := new(privacy.Point).AddPedersen(sum_tHat, baseG ,sum_tauX, baseH)
	LHSPrime.Add(LHSPrime, LHS)

	tmp5 := new(privacy.Point).MultiScalarMult(list_beta, list_A)
	tmp6 := new(privacy.Point).MultiScalarMult(list_x_beta, list_S)
	RHSPrime := new(privacy.Point).Add(tmp5, tmp6)
	RHSPrime.Add(RHSPrime, new(privacy.Point). MultiScalarMult(list_beta, list_LR))

	part1 := new(privacy.Point).MultiScalarMult(list_x_alpha, list_t1)
	part2 := new(privacy.Point).MultiScalarMult(list_xSquare, list_t2)
	RHS := new(privacy.Point).Add(part1, part2)
	RHS.Add(RHS, new(privacy.Point).MultiScalarMult(list_zSquare, list_V))
	RHSPrime.Add(RHSPrime, RHS)
	//fmt.Println("Batch Verification ", LHSPrime)
	//fmt.Println("Batch Verification ", RHSPrime)

	if !privacy.IsPointEqual(LHSPrime, RHSPrime) {
		privacy.Logger.Log.Errorf("batch verify aggregated range proof failed")
		return false, errors.New("batch verify aggregated range proof failed"), -1
	}
	return true, nil, -1
}

func computeHPrime(y *privacy.Scalar, N int, H []*privacy.Point) []*privacy.Point {
	yInverse := new(privacy.Scalar).Invert(y)
	HPrime := make([]*privacy.Point, N)
	expyInverse := new(privacy.Scalar).FromUint64(1)
	for i := 0; i < N; i++ {
		HPrime[i] = new(privacy.Point).ScalarMult(H[i], expyInverse)
		expyInverse.Mul(expyInverse, yInverse)
	}
	return HPrime
}

func computeDeltaYZ (z, zSquare *privacy.Scalar, yVector []*privacy.Scalar, N int ) (*privacy.Scalar, error) {
	oneNumber := new(privacy.Scalar).FromUint64(1)
	twoNumber := new(privacy.Scalar).FromUint64(2)
	oneVectorN := powerVector(oneNumber, maxExp)
	twoVectorN := powerVector(twoNumber, maxExp)
	oneVector := powerVector(oneNumber, N)

	deltaYZ := new(privacy.Scalar).Sub(z, zSquare)
	// ip1 = <1^(n*m), y^(n*m)>
	var ip1, ip2 *privacy.Scalar
	var err error
	if ip1, err = innerProduct(oneVector, yVector); err != nil {
		return nil, err
	} else if ip2, err = innerProduct(oneVectorN, twoVectorN); err != nil {
		return nil, err
	} else {
		deltaYZ.Mul(deltaYZ, ip1)
		sum := new(privacy.Scalar).FromUint64(0)
		zTmp := new(privacy.Scalar).Set(zSquare)
		for j := 0; j < int(N/maxExp); j++ {
			zTmp.Mul(zTmp, z)
			sum.Add(sum, zTmp)
		}
		sum.Mul(sum, ip2)
		deltaYZ.Sub(deltaYZ, sum)
	}
	return deltaYZ, nil
}

func innerProduct(a []*privacy.Scalar, b []*privacy.Scalar) (*privacy.Scalar, error)  {
	if len(a) != len(b) {
		return nil, errors.New("Incompatible sizes of a and b")
	}
	result := new(privacy.Scalar).FromUint64(uint64(0))
	for i := range a {
		//res = a[i]*b[i] + res % l
		result.MulAdd(a[i], b[i], result)
	}
	return result, nil
}

func vectorAdd(a []*privacy.Scalar, b []*privacy.Scalar) ([]*privacy.Scalar, error) {
	if len(a) != len(b) {
		return nil, errors.New("Incompatible sizes of a and b")
	}
	result := make([]*privacy.Scalar, len(a))
	for i := range a {
		result[i] = new(privacy.Scalar).Add(a[i], b[i])
	}
	return result, nil
}

func setAggregateParams(N int) *bulletproofParams {
	aggParam := new(bulletproofParams)
	aggParam.g = AggParam.g[0 : N]
	aggParam.h = AggParam.h[0 : N]
	aggParam.u = AggParam.u
	aggParam.cs = AggParam.cs
	return aggParam
}


// bulletproofParams includes all generator for aggregated range proof
func newBulletproofParams(m int) *bulletproofParams {
	capacity := maxExp * m // fixed value
	param := new(bulletproofParams)
	param.g = make([]*privacy.Point, capacity)
	param.h = make([]*privacy.Point, capacity)
	csByte := []byte{}

	for i := 0; i < capacity; i++ {
		param.g[i] = privacy.HashToPointFromIndex(int64(numCommitValue + i), privacy.CStringBulletProof)
		param.h[i] = privacy.HashToPointFromIndex(int64(numCommitValue + i + maxOutputNumberParam*maxExp), privacy.CStringBulletProof)
		csByte = append(csByte, param.g[i].ToBytesS()...)
		csByte = append(csByte, param.h[i].ToBytesS()...)
	}

	param.u = new(privacy.Point)
	param.u = privacy.HashToPointFromIndex(int64(numCommitValue + 2*maxOutputNumberParam*maxExp), privacy.CStringBulletProof)
	csByte = append(csByte, param.u.ToBytesS()...)

	param.cs = privacy.HashToPoint(csByte)
	return param
}

func generateChallenge(hashCache []byte, values []*privacy.Point,) *privacy.Scalar {
	bytes := []byte{}
	bytes = append(bytes, hashCache...)
	for i := 0; i < len(values); i++ {
		bytes = append(bytes, values[i].ToBytesS()...)
	}
	hash := privacy.HashToScalar(bytes)
	return hash
}

func pad(num int) int {
	if num == 1 || num == 2 {
		return num
	}
	tmp := 2
	for i := 2; ; i++ {
		tmp *= 2
		if tmp >= num {
			num = tmp
			break
		}
	}
	return num
}

func hadamardProduct(a []*privacy.Scalar, b []*privacy.Scalar) ([]*privacy.Scalar, error) {
	if len(a) != len(b) {
		return nil, errors.New("Invalid input")
	}
	result := make([]*privacy.Scalar, len(a))
	for i := 0; i < len(result); i++ {
		result[i] = new(privacy.Scalar).Mul(a[i], b[i])
	}
	return result, nil
}

// powerVector calculates base^n
func powerVector(base *privacy.Scalar, n int) []*privacy.Scalar {
	result := make([]*privacy.Scalar, n)
	result[0] = new(privacy.Scalar).FromUint64(1)
	if n > 1 {
		result[1] = new(privacy.Scalar).Set(base)
		for i := 2; i < n; i++ {
			result[i] = new(privacy.Scalar).Mul(result[i-1], base)
		}
	}
	return result
}

// vectorAddScalar adds a vector to a big int, returns big int array
func vectorAddScalar(v []*privacy.Scalar, s *privacy.Scalar) []*privacy.Scalar {
	result := make([]*privacy.Scalar, len(v))
	for i := range v {
		result[i] = new(privacy.Scalar).Add(v[i], s)
	}
	return result
}

// vectorMulScalar mul a vector to a big int, returns a vector
func vectorMulScalar(v []*privacy.Scalar, s *privacy.Scalar) []*privacy.Scalar {
	result := make([]*privacy.Scalar, len(v))
	for i := range v {
		result[i] = new(privacy.Scalar).Mul(v[i], s)
	}
	return result
}

// CommitAll commits a list of PCM_CAPACITY value(s)
func encodeVectors(l []*privacy.Scalar, r []*privacy.Scalar, g []*privacy.Point, h []*privacy.Point) (*privacy.Point, error) {
	if len(l) != len(r) || len(g) != len(l) || len(h) != len(g) {
		return nil, errors.New("Invalid input")
	}
	tmp1 := new(privacy.Point).MultiScalarMult(l, g)
	tmp2 := new(privacy.Point).MultiScalarMult(r, h)
	res := new(privacy.Point).Add(tmp1, tmp2)
	return res, nil
}


// estimateMultiRangeProofSize estimate multi range proof size
func EstimateMultiRangeProofSize(nOutput int) uint64 {
	return uint64((nOutput+2*int(math.Log2(float64(maxExp*pad(nOutput))))+5)*privacy.Ed25519KeySize + 5*privacy.Ed25519KeySize + 2)
}
