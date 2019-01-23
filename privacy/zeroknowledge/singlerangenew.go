package zkp

import (
	"fmt"
	"github.com/ninjadotorg/constant/privacy"
	"math/big"
)

type SingleRangeWitness struct {
	value *big.Int
	rand  *big.Int

	n byte
}

type SingleRangeProof struct {
	cmValue *privacy.EllipticPoint
	A       *privacy.EllipticPoint
	S       *privacy.EllipticPoint
	T1      *privacy.EllipticPoint
	T2      *privacy.EllipticPoint
	tauX    *big.Int
	tHat    *big.Int
	mu      *big.Int
	innerProductProof *InnerProductProof

	n byte
}

func (wit *SingleRangeWitness) Prove() (*SingleRangeProof, error) {
	var AggParam = newBulletproofParams(1)
	proof := new(SingleRangeProof)
	proof.n = wit.n

	proof.cmValue = privacy.PedCom.CommitAtIndex(wit.value, wit.rand, privacy.VALUE)

	n := int(wit.n)
	// Convert value to binary array
	aL := privacy.ConvertBigIntToBinary(wit.value, n)

	oneNumber := big.NewInt(1)
	twoNumber := big.NewInt(2)
	oneVector := powerVector(oneNumber, n)
	twoVector := powerVector(twoNumber, n)

	aR, err := vectorSub(aL, oneVector)
	if err != nil {
		return nil, err
	}

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
	sL := make([]*big.Int, n)
	sR := make([]*big.Int, n)
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
	zCube := new(big.Int).Mul(z, zSquare)
	//zCube.Mod(zCube, privacy.Curve.Params().N)

	// l(X) = (aL -z*1^n) + sL*X
	yVector := powerVector(y, n)

	l0 := vectorAddScalar(aL, zNeg)
	l1 := sL

	// r(X) = y^n hada (aR +z*1^n + sR*X) + z^2 * 2^n
	hadaProduct, err := hadamardProduct(yVector, vectorAddScalar(aR, z))
	if err != nil {
		return nil, err
	}

	r0, err := vectorAdd(hadaProduct, vectorMulScalar(twoVector, zSquare))
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

	// innerProduct1 = <1^n, y^n>
	innerProduct1, err := innerProduct(oneVector, yVector)
	if err != nil {
		return nil, err
	}
	deltaYZ.Mul(deltaYZ, innerProduct1)

	// innerProduct2 = <1^n, 2^n>
	innerProduct2, err := innerProduct(oneVector, twoVector)
	if err != nil {
		return nil, err
	}

	deltaYZ.Sub(deltaYZ, new(big.Int).Mul(zCube, innerProduct2))
	deltaYZ.Mod(deltaYZ, privacy.Curve.Params().N)

	//t0 := new(big.Int).Mul(wit.value, zSquare)
	//t0.Add(t0, deltaYZ)
	//t0.Mod(t0, privacy.Curve.Params().N)
	//
	//test, _ := innerProduct(l0, r0)
	//if test.Cmp(t0) ==0 {
	//	fmt.Printf("t0 is right!!!!\n")
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

	// lVector = aL - z*1^n + sL*x
	lVector, err := vectorAdd(vectorAddScalar(aL, zNeg), vectorMulScalar(sL, x))
	if err != nil {
		return nil, err
	}

	// rVector = y^n hada (aR +z*1^n + sR*x) + z^2*2^n
	vectorSum, err := vectorAdd(vectorAddScalar(aR, z), vectorMulScalar(sR, x))
	if err != nil {
		return nil, err
	}
	rVector, err := hadamardProduct(yVector, vectorSum)
	if err != nil {
		return nil, err
	}
	rVector, err = vectorAdd(rVector, vectorMulScalar(twoVector, zSquare))
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
	proof.tauX.Add(proof.tauX, new(big.Int).Mul(zSquare, wit.rand))
	proof.tauX.Mod(proof.tauX, privacy.Curve.Params().N)

	// alpha, rho blind A, S
	// mu = alpha + rho*x
	proof.mu = new(big.Int).Mul(rho, x)
	proof.mu.Add(proof.mu, alpha)
	proof.mu.Mod(proof.mu, privacy.Curve.Params().N)

	// instead of sending left vector and right vector, we use inner product argument to reduce proof size from 2*n to 2(log2(n)) + 2
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

func (proof *SingleRangeProof) Verify() bool {
	var AggParam = newBulletproofParams(1)
	n := int(proof.n)
	oneNumber := big.NewInt(1)
	twoNumber := big.NewInt(2)
	oneVector := powerVector(oneNumber, n)
	twoVector := powerVector(twoNumber, n)

	// recalculate challenge y, z
	y := generateChallengeForAggRange(AggParam, []*privacy.EllipticPoint{proof.A, proof.S})
	z := generateChallengeForAggRangeFromBytes(AggParam, [][]byte{proof.A.Compress(), proof.S.Compress(), y.Bytes()})
	zNeg := new(big.Int).Neg(z)
	zNeg.Mod(zNeg, privacy.Curve.Params().N)
	zSquare := new(big.Int).Exp(z, twoNumber, privacy.Curve.Params().N)
	zCube := new(big.Int).Exp(z, big.NewInt(3), privacy.Curve.Params().N)

	// challenge x = hash(G || H || A || S || T1 || T2)
	x := generateChallengeForAggRange(AggParam, []*privacy.EllipticPoint{proof.A, proof.S, proof.T1, proof.T2})
	xSquare := new(big.Int).Exp(x, twoNumber, privacy.Curve.Params().N)

	yVector := powerVector(y, n)

	// HPrime = H^(y^(1-i)
	tmp := new(big.Int)
	HPrime := make([]*privacy.EllipticPoint, n)
	for i :=0; i<n; i++ {
		HPrime[i] = AggParam.H[i].ScalarMult(tmp.Exp(y, big.NewInt(int64(-i)), privacy.Curve.Params().N))
	}

	// g^tHat * h^tauX = V^(z^2) * g^delta(y,z) * T1^x * T2^(x^2)
	deltaYZ := new(big.Int).Sub(z, zSquare)

	// innerProduct1 = <1^n, y^n>
	innerProduct1, err := innerProduct(oneVector, yVector)
	if err != nil {
		return false
	}
	deltaYZ.Mul(deltaYZ, innerProduct1)

	// innerProduct2 = <1^n, 2^n>
	innerProduct2, err := innerProduct(oneVector, twoVector)
	if err != nil {
		return false
	}

	deltaYZ.Sub(deltaYZ, new(big.Int).Mul(zCube, innerProduct2))
	deltaYZ.Mod(deltaYZ, privacy.Curve.Params().N)


	left1 := privacy.PedCom.CommitAtIndex(proof.tHat, proof.tauX, privacy.VALUE)
	right1 := proof.cmValue.ScalarMult(zSquare).Add(privacy.PedCom.G[privacy.VALUE].ScalarMult(deltaYZ)).Add(proof.T1.ScalarMult(x)).Add(proof.T2.ScalarMult(xSquare))

	if !left1.IsEqual(right1) {
		fmt.Printf("Err 3\n")
		return false
	}

	// A * S^x * G^(-z) * HPrime^(z*y^n + z^2*2^n) = h^mu * G^l * HPrime^r
	//expVector, err := vectorAdd(vectorMulScalar(yVector, z), vectorMulScalar(twoVector, zSquare))
	//if err != nil {
	//	fmt.Printf("Err 4\n")
	//	return false
	//}

	//left2 := proof.A.Add(proof.S.ScalarMult(x))
	//for i := range AggParam.G {
	//	left2 = left2.Add(AggParam.G[i].ScalarMult(zNeg)).Add(HPrime[i].ScalarMult(expVector[i]))
	//}
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
	//
	//right3, err := innerProduct(proof.lVector, proof.rVector)
	//if err != nil {
	//	fmt.Printf("Err 6\n")
	//	return false
	//}

	return proof.innerProductProof.Verify()
}
