package zkp

import (
	"fmt"
	"github.com/ninjadotorg/constant/privacy"
	"math/big"
)

//type AggregatedRangeWitness struct {
//	values []*big.Int
//	randomnesses []*big.Int
//
//	commitments []*privacy.EllipticPoint
//	n byte
//
//
//}
//
//
//type AggregatedRangeProof struct {
//
//}
//
//
//func (wit * AggregatedRangeWitness) Prove() (*AggregatedRangeProof, error){
//	aL :=
//
//	proof:= new(AggregatedRangeProof)
//	return proof, nil
//}

type SingleRangeWitness struct {
	value *big.Int
	rand *big.Int

	//commitments []*privacy.EllipticPoint
	n byte
}


type SingleRangeProof struct {
	cmValue *privacy.EllipticPoint
	A *privacy.EllipticPoint
	S *privacy.EllipticPoint
	T1 *privacy.EllipticPoint
	T2 *privacy.EllipticPoint
	tauX *big.Int
	tHat *big.Int
	lVector []*big.Int
	rVector []*big.Int
	mu *big.Int


	n byte

}


func (wit * SingleRangeWitness) Prove() (*SingleRangeProof, error){
	proof:= new(SingleRangeProof)
	proof.n = wit.n

	proof.cmValue = privacy.PedCom.CommitAtIndex(wit.value, wit.rand, privacy.VALUE)

	n := int(wit.n)
	// Convert value to binary array
	aL := privacy.ConvertBigIntToBinary(wit.value, n)

	fmt.Printf("aL: %v\n", aL)

	oneNumber := big.NewInt(1)
	twoNumber := big.NewInt(2)

	// aR = aL - 1
	aR := make([]*big.Int, n)
	for i:= range aR{
		aR[i] = new(big.Int).Sub(aL[i], oneNumber)
		aR[i].Mod(aR[i], twoNumber)
	}

	fmt.Printf("aR: %v\n", aR)

	aLConcatAR := make([]*big.Int, 0)
	for i := range aL{
		aLConcatAR = append(aLConcatAR, aL[i])
		aLConcatAR = append(aLConcatAR, aR[i])
	}

	fmt.Printf("Len ALconcatAR : %v\n", len(aLConcatAR))

	// random alpha
	alpha := privacy.RandInt()

	// Commitment to aL, aR: A = h^x * G^aL * H^aR
	A, err := EncodeVectors(aL, aR, AggParam.G, AggParam.H)
	if err != nil{
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
	if err != nil{
		return nil, err
	}
	S = S.Add(privacy.PedCom.G[privacy.RAND].ScalarMult(rho))
	proof.S = S

	// challenge y, z
	y := generateChallengeForAggRange([]*privacy.EllipticPoint{A, S})
	z := generateChallengeForAggRangeFromBytes([][]byte{A.Compress(), S.Compress(), y.Bytes()})

	// l(X) = (aL -z*1^n) + sL*X
	yVector := powerVector(y, n)
	l0 := vectorAddScalar(aL, new(big.Int).Neg(z))
	l1 := sL

	// r(X) = y^n hada (aR +z*1^n + sR*X) + z^2 * 2^n

	zSquare := new(big.Int).Exp(z, twoNumber, privacy.Curve.Params().N)
	zCube := new(big.Int).Exp(z, big.NewInt(3), privacy.Curve.Params().N)
	tmp := new(big.Int)

	zSquareMulTwoVec := make([]*big.Int, n)
	for i := 0; i < n; i++ {
		zSquareMulTwoVec[i] = new(big.Int).Set(zSquare)
		zSquareMulTwoVec[i].Mul(zSquareMulTwoVec[i], tmp.Exp(twoNumber, big.NewInt(int64(i)), nil))
	}
	hadaProduct, err := hadamardProduct(yVector,  vectorAddScalar(aR, z))
	if err != nil{
		return nil, err
	}

	r0, err := vectorAdd(hadaProduct, zSquareMulTwoVec)
	if err != nil{
		return nil, err
	}

	r1, err := hadamardProduct(sR, yVector)
	if err != nil{
		return nil, err
	}

	//t(X) = <l(X), r(X)> = t0 + t1*X + t2*X^2

	//calculate t0 = v*z^2 + delta(y, z)
	vMulZSquare := new(big.Int).Mul(wit.value, zSquare)

	oneVector := powerVector(oneNumber, n)
	twoVector := powerVector(twoNumber, n)
	deltaYZ := new(big.Int).Sub(z, zSquare)

	// innerProduct1 = <1^n, y^n>
	innerProduct1, err := innerProduct(oneVector, yVector)
	if err != nil{
		return nil, err
	}
	deltaYZ.Mul(deltaYZ, innerProduct1)

	// innerProduct2 = <1^n, 2^n>
	innerProduct2, err := innerProduct(oneVector, twoVector)
	if err != nil{
		return nil, err
	}

	deltaYZ.Sub(deltaYZ, new(big.Int).Mul(zCube, innerProduct2))
	deltaYZ.Mod(deltaYZ, privacy.Curve.Params().N)

	t0 := new(big.Int).Add(vMulZSquare, deltaYZ)
	t0.Mod(t0,privacy.Curve.Params().N)

	// t1 = <l1, r0> + <l0, r1>
	innerProduct3, err := innerProduct(l1, r0)
	if err != nil{
		return nil, err
	}

	innerProduct4, err := innerProduct(l0, r1)
	if err != nil{
		return nil, err
	}

	t1 := new(big.Int).Add(innerProduct3, innerProduct4)
	t1.Mod(t1,privacy.Curve.Params().N)

	// t2 = <l1, r1>
	innerProduct5, err := innerProduct(l1, r1)
	if err != nil{
		return nil, err
	}

	t2 := new(big.Int).Set(innerProduct5)

	// commitment to t1, t2
	tau1 := privacy.RandInt()
	tau2 := privacy.RandInt()

	proof.T1 = privacy.PedCom.CommitAtIndex(t1, tau1, privacy.VALUE)
	proof.T2 = privacy.PedCom.CommitAtIndex(t2, tau2, privacy.VALUE)

	// challenge x = hash(G || H || A || S || T1 || T2)
	x := generateChallengeForAggRange([]*privacy.EllipticPoint{proof.A, proof.S, proof.T1, proof.T2})
	xSquare := new(big.Int).Mul(x, x)
	xSquare.Mod(xSquare, privacy.Curve.Params().N)

	// lVector = aL - z*1^n + sL*x
	proof.lVector, err = vectorAdd(vectorAddScalar(aL, new(big.Int).Neg(z)), vectorMulScalar(sL, x))
	if err != nil{
		return nil, err
	}

	// rVector = y^n hada (aR +z*1^n + sR*x) + z^2*2^n
	vectorSum, err := vectorAdd(vectorAddScalar(aR, z), vectorMulScalar(sR, x))
	if err != nil{
		return nil, err
	}
	proof.rVector, err = hadamardProduct(yVector, vectorSum)
	if err != nil{
		return nil, err
	}
	proof.rVector, err = vectorAdd(proof.rVector, vectorMulScalar(twoVector, zSquare))
	if err != nil{
		return nil, err
	}

	// tHat = <lVector, rVector>
	proof.tHat, err = innerProduct(proof.lVector, proof.rVector)
	if err != nil{
		return nil, err
	}

	// blinding value for tHat: tauX = tau2*x^2 + tau1*x + z^2*rand
	proof.tauX = new(big.Int).Mul(tau2, xSquare)
	proof.tauX.Add(proof.tauX, new(big.Int).Mul(tau1, x))
	proof.tauX.Add(proof.tauX, new(big.Int).Mul(zSquare, wit.rand))

	// alpha, rho blind A, S
	// mu = alpha + rho*x
	proof.mu = new(big.Int).Mul(rho, x)
	proof.mu.Add(proof.mu, alpha)

	return proof, nil
}


func (proof *SingleRangeProof) Verify() bool{
	n := int(proof.n)
	oneNumber := big.NewInt(1)
	twoNumber := big.NewInt(2)
	oneVector := powerVector(oneNumber, n)
	twoVector := powerVector(twoNumber, n)

	// recalculate challenge y, z
	y := generateChallengeForAggRange([]*privacy.EllipticPoint{proof.A, proof.S})
	z := generateChallengeForAggRangeFromBytes([][]byte{proof.A.Compress(), proof.S.Compress(), y.Bytes()})
	zSquare := new(big.Int).Exp(z, twoNumber, privacy.Curve.Params().N)
	zCube := new(big.Int).Exp(z, big.NewInt(3), privacy.Curve.Params().N)

	// challenge x = hash(G || H || A || S || T1 || T2)
	x := generateChallengeForAggRange([]*privacy.EllipticPoint{proof.A, proof.S, proof.T1, proof.T2})
	xSquare := new(big.Int).Mul(x, x)
	xSquare.Mod(xSquare, privacy.Curve.Params().N)

	yVector := powerVector(y, n)

	// HPrime = H^(y^(1-i)
	HPrime := make([]*privacy.EllipticPoint, n)
	for i := range HPrime{
		HPrime[i] = AggParam.H[i].ScalarMult(new(big.Int).Exp(y, big.NewInt(int64(1-i)), nil))
	}

	// g^tHat * h^tauX = V^(z^2) * g^delta(y,z) * T1^x * T2^(x^2)
	deltaYZ := new(big.Int).Sub(z, zSquare)

	// innerProduct1 = <1^n, y^n>
	innerProduct1, err := innerProduct(oneVector, yVector)
	if err != nil{
		fmt.Printf("Err 1\n")
		return false
	}
	deltaYZ.Mul(deltaYZ, innerProduct1)

	// innerProduct2 = <1^n, 2^n>
	innerProduct2, err := innerProduct(oneVector, twoVector)
	if err != nil{
		fmt.Printf("Err 2\n")
		return false
	}

	deltaYZ.Sub(deltaYZ, new(big.Int).Mul(zCube, innerProduct2))
	deltaYZ.Mod(deltaYZ, privacy.Curve.Params().N)


	left1 := privacy.PedCom.CommitAtIndex(proof.tHat, proof.tauX, privacy.VALUE)
	right1 := proof.cmValue.ScalarMult(zSquare).Add(privacy.PedCom.G[privacy.VALUE].ScalarMult(deltaYZ)).Add(proof.T1.ScalarMult(x)).Add(proof.T2.ScalarMult(xSquare))

	if !left1.IsEqual(right1){
		fmt.Printf("Err 3\n")
		return false
	}

	// A * S^x * G^(-z) * HPrime^(z*y^n + z^2*2^n) = h^mu * G^l * HPrime^r
	zNeg := new(big.Int).Neg(z)
	expVector, err := vectorAdd(vectorMulScalar(yVector, z), vectorMulScalar(twoVector, zSquare))
	if err != nil{
		fmt.Printf("Err 4\n")
		return false
	}

	left2 := proof.A.Add(proof.S.ScalarMult(x))
	for i:= range HPrime{
		left2 = left2.Add(AggParam.G[i].ScalarMult(zNeg)).Add(HPrime[i].ScalarMult(expVector[i]))
	}

	right2 := privacy.PedCom.G[privacy.RAND].ScalarMult(proof.mu)
	for i:= range HPrime{
		right2 = right2.Add(AggParam.G[i].ScalarMult(proof.lVector[i])).Add(HPrime[i].ScalarMult(proof.rVector[i]))
	}

	if !left2.IsEqual(right2){
		fmt.Printf("Err 5\n")
		return false
	}

	right3, err := innerProduct(proof.lVector, proof.rVector)
	if err != nil{
		fmt.Printf("Err 6\n")
		return false
	}
	return proof.tHat.Cmp(right3) == 0
}
