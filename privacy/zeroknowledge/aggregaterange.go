package zkp

import (
	"fmt"
	"math/big"

	"github.com/ninjadotorg/constant/privacy"
)

// This protocol proves in zero-knowledge that a list of committed values falls in [0, 2^64)

type AggregatedRangeWitness struct {
	values []*big.Int
	rands  []*big.Int
}

type AggregatedRangeProof struct {
	cmsValue          []*privacy.EllipticPoint
	A                 *privacy.EllipticPoint
	S                 *privacy.EllipticPoint
	T1                *privacy.EllipticPoint
	T2                *privacy.EllipticPoint
	tauX              *big.Int
	tHat              *big.Int
	mu                *big.Int
	innerProductProof *InnerProductProof
}

func (proof *AggregatedRangeProof) Init() *AggregatedRangeProof {
	proof.A = new(privacy.EllipticPoint).Zero()
	proof.S = new(privacy.EllipticPoint).Zero()
	proof.T1 = new(privacy.EllipticPoint).Zero()
	proof.T2 = new(privacy.EllipticPoint).Zero()
	proof.tauX = new(big.Int)
	proof.tHat = new(big.Int)
	proof.mu = new(big.Int)
	proof.innerProductProof = new(InnerProductProof)
	return proof
}

func (proof *AggregatedRangeProof) IsNil() bool {
	if proof.A == nil {
		return true
	}
	if proof.S == nil {
		return true
	}
	if proof.T1 == nil {
		return true
	}
	if proof.T2 == nil {
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
		res = append(res, proof.cmsValue[i].Compress()...)
	}

	res = append(res, proof.A.Compress()...)
	res = append(res, proof.S.Compress()...)
	res = append(res, proof.T1.Compress()...)
	res = append(res, proof.T2.Compress()...)

	res = append(res, privacy.AddPaddingBigInt(proof.tauX, privacy.BigIntSize)...)
	res = append(res, privacy.AddPaddingBigInt(proof.tHat, privacy.BigIntSize)...)
	res = append(res, privacy.AddPaddingBigInt(proof.mu, privacy.BigIntSize)...)
	res = append(res, proof.innerProductProof.Bytes()...)

	//fmt.Printf("BYTES ------------ %v\n", res)
	return res

}

func (proof *AggregatedRangeProof) SetBytes(bytes []byte) error {
	if len(bytes) == 0 {
		return nil
	}

	//fmt.Printf("BEFORE SETBYTES ------------ %v\n", bytes)

	lenValues := int(bytes[0])
	offset := 1

	proof.cmsValue = make([]*privacy.EllipticPoint, lenValues)
	for i := 0; i < lenValues; i++ {
		proof.cmsValue[i] = new(privacy.EllipticPoint)
		err := proof.cmsValue[i].Decompress(bytes[offset : offset+privacy.CompressedPointSize])
		if err != nil {
			return err
		}
		offset += privacy.CompressedPointSize
	}

	proof.A = new(privacy.EllipticPoint)
	err := proof.A.Decompress(bytes[offset:])
	if err != nil {
		return err
	}
	offset += privacy.CompressedPointSize

	proof.S = new(privacy.EllipticPoint)
	err = proof.S.Decompress(bytes[offset:])
	if err != nil {
		return err
	}
	offset += privacy.CompressedPointSize

	proof.T1 = new(privacy.EllipticPoint)
	err = proof.T1.Decompress(bytes[offset:])
	if err != nil {
		return err
	}
	offset += privacy.CompressedPointSize

	proof.T2 = new(privacy.EllipticPoint)
	err = proof.T2.Decompress(bytes[offset:])
	if err != nil {
		return err
	}
	offset += privacy.CompressedPointSize

	proof.tauX = new(big.Int).SetBytes(bytes[offset : offset+privacy.BigIntSize])
	offset += privacy.BigIntSize

	proof.tHat = new(big.Int).SetBytes(bytes[offset : offset+privacy.BigIntSize])
	offset += privacy.BigIntSize

	proof.mu = new(big.Int).SetBytes(bytes[offset : offset+privacy.BigIntSize])
	offset += privacy.BigIntSize

	proof.innerProductProof = new(InnerProductProof)
	proof.innerProductProof.SetBytes(bytes[offset:])

	//fmt.Printf("AFTER SETBYTES ------------ %v\n", proof.Bytes())
	return nil
}

func (wit *AggregatedRangeWitness) Set(values []*big.Int, rands []*big.Int) {
	numValue := len(values)
	wit.values = make([]*big.Int, numValue)
	wit.rands = make([]*big.Int, numValue)

	for i := range values {
		wit.values[i] = new(big.Int).Set(values[i])
		wit.rands[i] = new(big.Int).Set(rands[i])
	}
}

func (wit *AggregatedRangeWitness) Prove() (*AggregatedRangeProof, error) {
	proof := new(AggregatedRangeProof)

	numValue := len(wit.values)
	numValuePad := pad(numValue)
	values := make([]*big.Int, numValuePad)
	rands := make([]*big.Int, numValuePad)

	for i := range wit.values {
		values[i] = new(big.Int).Set(wit.values[i])
		rands[i] = new(big.Int).Set(wit.rands[i])
	}

	for i := numValue; i < numValuePad; i++ {
		values[i] = big.NewInt(0)
		rands[i] = big.NewInt(0)
	}

	AggParam := newBulletproofParams(numValuePad)

	proof.cmsValue = make([]*privacy.EllipticPoint, numValue)
	for i := 0; i < numValue; i++ {
		proof.cmsValue[i] = privacy.PedCom.CommitAtIndex(values[i], rands[i], privacy.VALUE)
	}

	n := privacy.MaxExp
	// Convert values to binary array
	aL := make([]*big.Int, numValuePad*n)
	for i, value := range values {
		tmp := privacy.ConvertBigIntToBinary(value, n)
		for j := 0; j < n; j++ {
			aL[i*n+j] = tmp[j]
		}
	}

	oneNumber := big.NewInt(1)
	twoNumber := big.NewInt(2)
	oneVector := powerVector(oneNumber, n*numValuePad)
	oneVectorN := powerVector(oneNumber, n)
	twoVectorN := powerVector(twoNumber, n)

	aR, err := vectorSub(aL, oneVector)
	if err != nil {
		return nil, err
	}

	// random alpha
	alpha := privacy.RandScalar()

	// Commitment to aL, aR: A = h^alpha * G^aL * H^aR
	A, err := EncodeVectors(aL, aR, AggParam.G, AggParam.H)
	if err != nil {
		return nil, err
	}
	A = A.Add(privacy.PedCom.G[privacy.RAND].ScalarMult(alpha))
	proof.A = A

	// Random blinding vectors sL, sR
	sL := make([]*big.Int, n*numValuePad)
	sR := make([]*big.Int, n*numValuePad)
	for i := range sL {
		sL[i] = privacy.RandScalar()
		sR[i] = privacy.RandScalar()
	}

	// random rho
	rho := privacy.RandScalar()

	// Commitment to sL, sR : S = h^rho * G^sL * H^sR
	S, err := EncodeVectors(sL, sR, AggParam.G, AggParam.H)
	if err != nil {
		return nil, err
	}
	S = S.Add(privacy.PedCom.G[privacy.RAND].ScalarMult(rho))
	proof.S = S

	// challenge y, z
	y := generateChallengeForAggRange(AggParam, [][]byte{A.Compress(), S.Compress()})
	z := generateChallengeForAggRange(AggParam, [][]byte{A.Compress(), S.Compress(), y.Bytes()})
	zNeg := new(big.Int).Neg(z)
	zNeg.Mod(zNeg, privacy.Curve.Params().N)
	zSquare := new(big.Int).Mul(z, z)
	zSquare.Mod(zSquare, privacy.Curve.Params().N)

	// l(X) = (aL -z*1^n) + sL*X
	yVector := powerVector(y, n*numValuePad)

	l0 := vectorAddScalar(aL, zNeg)
	l1 := sL

	// r(X) = y^n hada (aR +z*1^n + sR*X) + z^2 * 2^n
	hadaProduct, err := hadamardProduct(yVector, vectorAddScalar(aR, z))
	if err != nil {
		return nil, err
	}

	vectorSum := make([]*big.Int, n*numValuePad)
	zTmp := new(big.Int).Set(z)
	for j := 0; j < numValuePad; j++ {
		zTmp.Mul(zTmp, z)
		zTmp.Mod(zTmp, privacy.Curve.Params().N)
		for i := 0; i < n; i++ {
			vectorSum[j*n+i] = new(big.Int).Mul(twoVectorN[i], zTmp)
			vectorSum[j*n+i].Mod(vectorSum[j*n+i], privacy.Curve.Params().N)
		}
	}

	r0, err := vectorAdd(hadaProduct, vectorSum)
	if err != nil {
		return nil, err
	}

	r1, err := hadamardProduct(yVector, sR)
	if err != nil {
		return nil, err
	}

	//t(X) = <l(X), r(X)> = t0 + t1*X + t2*X^2

	//calculate t0 = v*z^2 + delta(y, z)
	deltaYZ := new(big.Int).Sub(z, zSquare)

	// innerProduct1 = <1^(n*m), y^(n*m)>
	innerProduct1, err := innerProduct(oneVector, yVector)
	if err != nil {
		return nil, err
	}

	deltaYZ.Mul(deltaYZ, innerProduct1)

	// innerProduct2 = <1^n, 2^n>
	innerProduct2, err := innerProduct(oneVectorN, twoVectorN)
	if err != nil {
		return nil, err
	}

	sum := big.NewInt(0)
	zTmp = new(big.Int).Set(zSquare)
	for j := 0; j < numValuePad; j++ {
		zTmp.Mul(zTmp, z)
		zTmp.Mod(zTmp, privacy.Curve.Params().N)

		sum.Add(sum, zTmp)
	}
	sum.Mul(sum, innerProduct2)
	deltaYZ.Sub(deltaYZ, sum)
	deltaYZ.Mod(deltaYZ, privacy.Curve.Params().N)

	// Check whether t0 is computed correctedly or not
	//t0 := new(big.Int).Set(deltaYZ)
	//zTmp = new(big.Int).Set(z)
	//for i := range wit.values {
	//	zTmp.Mul(zTmp, z)
	//	t0.Add(t0, new(big.Int).Mul(wit.values[i], zTmp))
	//}
	//t0.Mod(t0, privacy.Curve.Params().N)
	//
	//tmp, _ := innerProduct(l0, r0)
	//
	//if t0.Cmp(tmp) == 0 {
	//	fmt.Printf("t0 is right\n")
	//} else {
	//	fmt.Printf("t0 is wrong\n")
	//}

	// t1 = <l1, r0> + <l0, r1>
	innerProduct3, err := innerProduct(l1, r0)
	if err != nil {
		return nil, err
	}

	innerProduct4, err := innerProduct(l0, r1)
	if err != nil {
		return nil, err
	}

	t1 := new(big.Int).Add(innerProduct3, innerProduct4)
	t1.Mod(t1, privacy.Curve.Params().N)

	// t2 = <l1, r1>
	t2, err := innerProduct(l1, r1)
	if err != nil {
		return nil, err
	}

	// commitment to t1, t2
	tau1 := privacy.RandScalar()
	tau2 := privacy.RandScalar()

	proof.T1 = privacy.PedCom.CommitAtIndex(t1, tau1, privacy.VALUE)
	proof.T2 = privacy.PedCom.CommitAtIndex(t2, tau2, privacy.VALUE)

	// challenge x = hash(G || H || A || S || T1 || T2)
	x := generateChallengeForAggRange(AggParam, [][]byte{proof.A.Compress(), proof.S.Compress(), proof.T1.Compress(), proof.T2.Compress()})
	xSquare := new(big.Int).Exp(x, twoNumber, privacy.Curve.Params().N)

	// lVector = aL - z*1^n + sL*x
	lVector, err := vectorAdd(vectorAddScalar(aL, zNeg), vectorMulScalar(sL, x))
	if err != nil {
		return nil, err
	}

	// rVector = y^n hada (aR +z*1^n + sR*x) + z^2*2^n
	tmpVector, err := vectorAdd(vectorAddScalar(aR, z), vectorMulScalar(sR, x))
	if err != nil {
		return nil, err
	}
	rVector, err := hadamardProduct(yVector, tmpVector)
	if err != nil {
		return nil, err
	}

	vectorSum = make([]*big.Int, n*numValuePad)
	zTmp = new(big.Int).Set(z)
	for j := 0; j < numValuePad; j++ {
		zTmp.Mul(zTmp, z)
		zTmp.Mod(zTmp, privacy.Curve.Params().N)
		for i := 0; i < n; i++ {
			vectorSum[j*n+i] = new(big.Int).Mul(twoVectorN[i], zTmp)
			vectorSum[j*n+i].Mod(vectorSum[j*n+i], privacy.Curve.Params().N)
		}
	}

	rVector, err = vectorAdd(rVector, vectorSum)
	if err != nil {
		return nil, err
	}

	// tHat = <lVector, rVector>
	proof.tHat, err = innerProduct(lVector, rVector)
	if err != nil {
		return nil, err
	}

	// blinding value for tHat: tauX = tau2*x^2 + tau1*x + z^2*rand
	proof.tauX = new(big.Int).Mul(tau2, xSquare)
	proof.tauX.Add(proof.tauX, new(big.Int).Mul(tau1, x))
	zTmp = new(big.Int).Set(z)
	tmpBN := new(big.Int)
	for j := 0; j < numValuePad; j++ {
		zTmp.Mul(zTmp, z)
		zTmp.Mod(zTmp, privacy.Curve.Params().N)

		proof.tauX.Add(proof.tauX, tmpBN.Mul(zTmp, rands[j]))
	}
	proof.tauX.Mod(proof.tauX, privacy.Curve.Params().N)

	// alpha, rho blind A, S
	// mu = alpha + rho*x
	proof.mu = new(big.Int).Mul(rho, x)
	proof.mu.Add(proof.mu, alpha)
	proof.mu.Mod(proof.mu, privacy.Curve.Params().N)

	// instead of sending left vector and right vector, we use inner sum argument to reduce proof size from 2*n to 2(log2(n)) + 2
	innerProductWit := new(InnerProductWitness)
	innerProductWit.a = lVector
	innerProductWit.b = rVector
	innerProductWit.p, err = EncodeVectors(lVector, rVector, AggParam.G, AggParam.H)
	if err != nil {
		return nil, err
	}
	innerProductWit.p = innerProductWit.p.Add(AggParam.U.ScalarMult(proof.tHat))

	proof.innerProductProof, err = innerProductWit.Prove(AggParam)
	if err != nil {
		return nil, err
	}

	return proof, nil
}

func (proof *AggregatedRangeProof) Verify() bool {
	numValue := len(proof.cmsValue)
	numValuePad := pad(numValue)

	tmpcmsValue := proof.cmsValue

	for i := numValue; i < numValuePad; i++ {
		tmpcmsValue = append(tmpcmsValue, new(privacy.EllipticPoint).Zero())
	}

	AggParam := newBulletproofParams(numValuePad)
	n := privacy.MaxExp
	oneNumber := big.NewInt(1)
	twoNumber := big.NewInt(2)
	oneVector := powerVector(oneNumber, n*numValuePad)
	oneVectorN := powerVector(oneNumber, n)
	twoVectorN := powerVector(twoNumber, n)

	// recalculate challenge y, z
	y := generateChallengeForAggRange(AggParam, [][]byte{proof.A.Compress(), proof.S.Compress()})
	z := generateChallengeForAggRange(AggParam, [][]byte{proof.A.Compress(), proof.S.Compress(), y.Bytes()})
	zNeg := new(big.Int).Neg(z)
	zNeg.Mod(zNeg, privacy.Curve.Params().N)
	zSquare := new(big.Int).Exp(z, twoNumber, privacy.Curve.Params().N)

	// challenge x = hash(G || H || A || S || T1 || T2)
	fmt.Printf("T2: %v\n", proof.T2)
	x := generateChallengeForAggRange(AggParam, [][]byte{proof.A.Compress(), proof.S.Compress(), proof.T1.Compress(), proof.T2.Compress()})
	xSquare := new(big.Int).Exp(x, twoNumber, privacy.Curve.Params().N)

	yVector := powerVector(y, n*numValuePad)

	// HPrime = H^(y^(1-i)
	tmp := new(big.Int)
	HPrime := make([]*privacy.EllipticPoint, n*numValuePad)
	for i := 0; i < n*numValuePad; i++ {
		HPrime[i] = AggParam.H[i].ScalarMult(tmp.Exp(y, big.NewInt(int64(-i)), privacy.Curve.Params().N))
	}

	// g^tHat * h^tauX = V^(z^2) * g^delta(y,z) * T1^x * T2^(x^2)
	deltaYZ := new(big.Int).Sub(z, zSquare)

	// innerProduct1 = <1^(n*m), y^(n*m)>
	innerProduct1, err := innerProduct(oneVector, yVector)
	if err != nil {
		return false
	}

	deltaYZ.Mul(deltaYZ, innerProduct1)

	// innerProduct2 = <1^n, 2^n>
	innerProduct2, err := innerProduct(oneVectorN, twoVectorN)
	if err != nil {
		return false
	}

	sum := big.NewInt(0)
	zTmp := new(big.Int).Set(zSquare)
	for j := 0; j < numValuePad; j++ {
		zTmp.Mul(zTmp, z)
		zTmp.Mod(zTmp, privacy.Curve.Params().N)

		sum.Add(sum, zTmp)
	}
	sum.Mul(sum, innerProduct2)
	deltaYZ.Sub(deltaYZ, sum)
	deltaYZ.Mod(deltaYZ, privacy.Curve.Params().N)

	left1 := privacy.PedCom.CommitAtIndex(proof.tHat, proof.tauX, privacy.VALUE)
	right1 := privacy.PedCom.G[privacy.VALUE].ScalarMult(deltaYZ).Add(proof.T1.ScalarMult(x)).Add(proof.T2.ScalarMult(xSquare))

	expVector := vectorMulScalar(powerVector(z, numValuePad), zSquare)
	for i, cm := range tmpcmsValue {
		right1 = right1.Add(cm.ScalarMult(expVector[i]))
	}

	if !left1.IsEqual(right1) {
		return false
	}

	return proof.innerProductProof.Verify(AggParam)
}
