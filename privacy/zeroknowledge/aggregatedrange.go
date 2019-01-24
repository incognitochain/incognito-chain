package zkp

import (
	"fmt"
	"github.com/ninjadotorg/constant/privacy"
	"math/big"
)

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
	//lVector           []*big.Int
	//rVector           []*big.Int
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

	//if proof.IsNil() == true {
	//	return []byte{}
	//}
	//
	//for i := 0; i < int(proof.Counter); i++ {
	//	res = append(res, proof.Comms[i].Compress()...)
	//}
	//
	//res = append(res, proof.A.Compress()...)
	//res = append(res, proof.S.Compress()...)
	//res = append(res, proof.T1.Compress()...)
	//res = append(res, proof.T2.Compress()...)
	//
	//res = append(res, privacy.AddPaddingBigInt(proof.Tau, privacy.BigIntSize)...)
	//res = append(res, privacy.AddPaddingBigInt(proof.Th, privacy.BigIntSize)...)
	//res = append(res, privacy.AddPaddingBigInt(proof.Mu, privacy.BigIntSize)...)
	//res = append(res, privacy.AddPaddingBigInt(proof.Cx, privacy.BigIntSize)...)
	//res = append(res, privacy.AddPaddingBigInt(proof.Cy, privacy.BigIntSize)...)
	//res = append(res, privacy.AddPaddingBigInt(proof.Cz, privacy.BigIntSize)...)
	//res = append(res, proof.IPP.bytes()...)
	return res

}

func (proof *AggregatedRangeProof) SetBytes(proofbytes []byte) error {
	if proof.IsNil() {
		proof = proof.Init()
	}

	//if len(proofbytes) == 0 {
	//	return nil
	//}
	//
	//proof.Counter = proofbytes[0]
	//proof.maxExp = proofbytes[1]
	//offset := 2
	//
	//proof.Comms = make([]*privacy.EllipticPoint, proof.Counter)
	//for i := 0; i < int(proof.Counter); i++ {
	//	proof.Comms[i] = new(privacy.EllipticPoint)
	//	err := proof.Comms[i].Decompress(proofbytes[offset: offset+privacy.CompressedPointSize])
	//	if err != nil {
	//		return err
	//	}
	//	offset += privacy.CompressedPointSize
	//}
	//
	//proof.A = new(privacy.EllipticPoint)
	//err := proof.A.Decompress(proofbytes[offset:])
	//if err != nil {
	//	return err
	//}
	//offset += privacy.CompressedPointSize
	//
	//proof.S = new(privacy.EllipticPoint)
	//err = proof.S.Decompress(proofbytes[offset:])
	//if err != nil {
	//	return err
	//}
	//offset += privacy.CompressedPointSize
	//
	//proof.T1 = new(privacy.EllipticPoint)
	//err = proof.T1.Decompress(proofbytes[offset:])
	//if err != nil {
	//	return err
	//}
	//offset += privacy.CompressedPointSize
	//
	//proof.T2 = new(privacy.EllipticPoint)
	//err = proof.T2.Decompress(proofbytes[offset:])
	//if err != nil {
	//	return err
	//}
	//offset += privacy.CompressedPointSize
	//
	//proof.Tau = new(big.Int).SetBytes(proofbytes[offset: offset+privacy.BigIntSize])
	//offset += privacy.BigIntSize
	//
	//proof.Th = new(big.Int).SetBytes(proofbytes[offset: offset+privacy.BigIntSize])
	//offset += privacy.BigIntSize
	//
	//proof.Mu = new(big.Int).SetBytes(proofbytes[offset: offset+privacy.BigIntSize])
	//offset += privacy.BigIntSize
	//
	//proof.Cx = new(big.Int).SetBytes(proofbytes[offset: offset+privacy.BigIntSize])
	//offset += privacy.BigIntSize
	//
	//proof.Cy = new(big.Int).SetBytes(proofbytes[offset: offset+privacy.BigIntSize])
	//offset += privacy.BigIntSize
	//
	//proof.Cz = new(big.Int).SetBytes(proofbytes[offset: offset+privacy.BigIntSize])
	//offset += privacy.BigIntSize
	//
	//end := len(proofbytes)
	//proof.IPP.setBytes(proofbytes[offset:end])
	return nil
}

func (wit *AggregatedRangeWitness) Set(v []*big.Int, maxExp byte) {

}


func (wit *AggregatedRangeWitness) Prove() (*AggregatedRangeProof, error) {
	proof := new(AggregatedRangeProof)

	numValue := len(wit.values)
	AggParam := newBulletproofParams(numValue)

	proof.cmsValue = make([]*privacy.EllipticPoint, numValue)
	for i := range proof.cmsValue {
		proof.cmsValue[i] = privacy.PedCom.CommitAtIndex(wit.values[i], wit.rands[i], privacy.VALUE)
	}

	n := privacy.MaxExp
	// Convert values to binary array
	aL := make([]*big.Int, 0)
	for _, value := range wit.values {
		tmp := privacy.ConvertBigIntToBinary(value, n)
		aL = append(aL, tmp...)
	}

	//fmt.Printf("aL: %v\n", aL)

	oneNumber := big.NewInt(1)
	twoNumber := big.NewInt(2)
	oneVector := powerVector(oneNumber, n*numValue)
	oneVectorN := powerVector(oneNumber, n)
	twoVectorN := powerVector(twoNumber, n)

	aR, err := vectorSub(aL, oneVector)
	if err != nil {
		return nil, err
	}

	//fmt.Printf("aR: %v\n", aR)

	// random alpha
	alpha := privacy.RandInt()

	// Commitment to aL, aR: A = h^alpha * G^aL * H^aR
	A, err := EncodeVectors(aL, aR, AggParam.G, AggParam.H)
	if err != nil {
		return nil, err
	}
	A = A.Add(privacy.PedCom.G[privacy.RAND].ScalarMult(alpha))
	proof.A = A

	// Random blinding vectors sL, sR
	sL := make([]*big.Int, n*numValue)
	sR := make([]*big.Int, n*numValue)
	for i := range sL {
		sL[i] = privacy.RandInt()
		sR[i] = privacy.RandInt()
	}

	// random rho
	rho := privacy.RandInt()

	// Commitment to sL, sR : S = h^rho * G^sL * H^sR
	S, err := EncodeVectors(sL, sR, AggParam.G, AggParam.H)
	if err != nil {
		return nil, err
	}
	S = S.Add(privacy.PedCom.G[privacy.RAND].ScalarMult(rho))
	proof.S = S

	// challenge y, z
	y := generateChallengeForAggRange(AggParam, []*privacy.EllipticPoint{A, S})
	z := generateChallengeForAggRangeFromBytes(AggParam, [][]byte{A.Compress(), S.Compress(), y.Bytes()})
	zNeg := new(big.Int).Neg(z)
	zNeg.Mod(zNeg, privacy.Curve.Params().N)
	zSquare := new(big.Int).Exp(z, twoNumber, privacy.Curve.Params().N)

	//fmt.Printf("Prove y: %v\n", y)
	//fmt.Printf("Prove z: %v\n", z)

	// l(X) = (aL -z*1^n) + sL*X
	yVector := powerVector(y, n*numValue)

	l0 := vectorAddScalar(aL, zNeg)
	l1 := sL

	// r(X) = y^n hada (aR +z*1^n + sR*X) + z^2 * 2^n
	hadaProduct, err := hadamardProduct(yVector, vectorAddScalar(aR, z))
	if err != nil {
		return nil, err
	}

	vectorSum := make([]*big.Int, n*numValue)
	zTmp := new(big.Int).Set(z)
	for j := 0; j < numValue; j++ {
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
	for j := 0; j < numValue; j++ {
		zTmp.Mul(zTmp, z)
		zTmp.Mod(zTmp, privacy.Curve.Params().N)

		sum.Add(sum, zTmp)
	}
	sum.Mul(sum, innerProduct2)
	deltaYZ.Sub(deltaYZ, sum)
	deltaYZ.Mod(deltaYZ, privacy.Curve.Params().N)

	//fmt.Printf("Prove delta: %v\n", deltaYZ)

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
	//	fmt.Printf("AAAAAAAAAAA\n")
	//} else {
	//	fmt.Printf("BBBBBBBBBBB\n")
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
	tau1 := privacy.RandInt()
	tau2 := privacy.RandInt()

	proof.T1 = privacy.PedCom.CommitAtIndex(t1, tau1, privacy.VALUE)
	proof.T2 = privacy.PedCom.CommitAtIndex(t2, tau2, privacy.VALUE)

	// challenge x = hash(G || H || A || S || T1 || T2)
	x := generateChallengeForAggRange(AggParam, []*privacy.EllipticPoint{proof.A, proof.S, proof.T1, proof.T2})
	xSquare := new(big.Int).Exp(x, twoNumber, privacy.Curve.Params().N)

	fmt.Printf("Prove x: %v\n", x)

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

	vectorSum = make([]*big.Int, n*numValue)
	zTmp = new(big.Int).Set(z)
	for j := 0; j < numValue; j++ {
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
	tmp := new(big.Int)
	for j := 0; j < numValue; j++ {
		zTmp.Mul(zTmp, z)
		zTmp.Mod(zTmp, privacy.Curve.Params().N)

		proof.tauX.Add(proof.tauX, tmp.Mul(zTmp, wit.rands[j]))
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
	AggParam := newBulletproofParams(numValue)
	n := privacy.MaxExp
	oneNumber := big.NewInt(1)
	twoNumber := big.NewInt(2)
	oneVector := powerVector(oneNumber, n*numValue)
	oneVectorN := powerVector(oneNumber, n)
	twoVectorN := powerVector(twoNumber, n)

	// recalculate challenge y, z
	y := generateChallengeForAggRange(AggParam, []*privacy.EllipticPoint{proof.A, proof.S})
	z := generateChallengeForAggRangeFromBytes(AggParam, [][]byte{proof.A.Compress(), proof.S.Compress(), y.Bytes()})
	zNeg := new(big.Int).Neg(z)
	zNeg.Mod(zNeg, privacy.Curve.Params().N)
	zSquare := new(big.Int).Exp(z, twoNumber, privacy.Curve.Params().N)

	// challenge x = hash(G || H || A || S || T1 || T2)
	x := generateChallengeForAggRange(AggParam, []*privacy.EllipticPoint{proof.A, proof.S, proof.T1, proof.T2})
	xSquare := new(big.Int).Exp(x, twoNumber, privacy.Curve.Params().N)

	yVector := powerVector(y, n*numValue)

	// HPrime = H^(y^(1-i)
	tmp := new(big.Int)
	HPrime := make([]*privacy.EllipticPoint, n*numValue)
	for i := 0; i < n*numValue; i++ {
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

	for j := 1; j <= numValue; j++ {
		zTmp.Mul(zTmp, z)
		zTmp.Mod(zTmp, privacy.Curve.Params().N)

		sum.Add(sum, zTmp)
	}
	sum.Mul(sum, innerProduct2)
	deltaYZ.Sub(deltaYZ, sum)
	deltaYZ.Mod(deltaYZ, privacy.Curve.Params().N)

	left1 := privacy.PedCom.CommitAtIndex(proof.tHat, proof.tauX, privacy.VALUE)
	right1 := privacy.PedCom.G[privacy.VALUE].ScalarMult(deltaYZ).Add(proof.T1.ScalarMult(x)).Add(proof.T2.ScalarMult(xSquare))

	expVector := vectorMulScalar(powerVector(z, numValue), zSquare)
	for i, cm := range proof.cmsValue {
		right1 = right1.Add(cm.ScalarMult(expVector[i]))
	}

	if !left1.IsEqual(right1) {
		fmt.Printf("Err 3\n")
		return false
	}

	//A * S^x * G^(-z) * HPrime^(z*y^n + z^2*2^n) = h^mu * G^l * HPrime^r
	//expVector = vectorMulScalar(yVector, z)
	//left2 := proof.A.Add(proof.S.ScalarMult(x))
	//for i := range AggParam.G {
	//	left2 = left2.Add(AggParam.G[i].ScalarMult(zNeg)).Add(HPrime[i].ScalarMult(expVector[i]))
	//}
	//
	//twoVectorN := powerVector(twoNumber, n)
	//sum := new(privacy.EllipticPoint).Zero()
	//for i := range proof.cmsValue{
	//	for j:=0; j<n; j++{
	//		tmp := HPrime[i*n + j].ScalarMult(new(big.Int).Mul(new(big.Int).Exp(z, big.NewInt(int64(i+2)), privacy.Curve.Params().N), twoVectorN[j]))
	//		sum = sum.Add( tmp)
	//	}
	//}
	//
	//left2 = left2.Add(sum)
	//
	//right2 := privacy.PedCom.G[privacy.RAND].ScalarMult(proof.mu)
	//for i := range AggParam.G {
	//	right2 = right2.Add(AggParam.G[i].ScalarMult(proof.lVector[i])).Add(HPrime[i].ScalarMult(proof.rVector[i]))
	//}
	//
	//if !left2.IsEqual(right2) {
	//	fmt.Printf("Err 5\n")
	//	return false
	//}

	//right3, err := innerProduct(proof.lVector, proof.rVector)
	//if err != nil {
	//	fmt.Printf("Err 6\n")
	//	return false
	//}

	return proof.innerProductProof.Verify(AggParam)
}
