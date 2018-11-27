package zkp

import (
	"fmt"
	"github.com/minio/blake2b-simd"
	"github.com/ninjadotorg/constant/privacy-protocol"
	"math/big"
)

type PKComMultiRangeProof struct {
	Comms []privacy.EllipticPoint
	A     privacy.EllipticPoint
	S     privacy.EllipticPoint
	T1    privacy.EllipticPoint
	T2    privacy.EllipticPoint
	Tau   *big.Int
	Th    *big.Int
	Mu    *big.Int
	IPP   InnerProdArg

	// challenges
	Cy *big.Int
	Cz *big.Int
	Cx *big.Int
}
type PKComMultiRangeWitness struct {
	Values [] *big.Int
}

func (wit *PKComMultiRangeWitness) Set(v []*big.Int){
	VecLength = 64*len(v)
	wit.Values = v
}


// Calculates (aL - z*1^n) + sL*x
func CalculateLMRP(aL, sL []*big.Int, z, x *big.Int) []*big.Int {
	result := make([]*big.Int, len(aL))

	tmp1 := VectorAddScalar(aL, new(big.Int).Neg(z))
	tmp2 := ScalarVectorMul(sL, x)

	result = VectorAdd(tmp1, tmp2)

	return result
}

func CalculateRMRP(aR, sR, y, zTimesTwo []*big.Int, z, x *big.Int) []*big.Int {
	if len(aR) != len(sR) || len(aR) != len(y) || len(y) != len(zTimesTwo) {
		fmt.Println("CalculateR: Uh oh! Arrays not of the same length")
		fmt.Printf("len(aR): %d\n", len(aR))
		fmt.Printf("len(sR): %d\n", len(sR))
		fmt.Printf("len(y): %d\n", len(y))
		fmt.Printf("len(po2): %d\n", len(zTimesTwo))
	}

	result := make([]*big.Int, len(aR))

	tmp11 := VectorAddScalar(aR, z)
	tmp12 := ScalarVectorMul(sR, x)
	tmp1 := VectorHadamard(y, VectorAdd(tmp11, tmp12))

	result = VectorAdd(tmp1, zTimesTwo)

	return result
}

/*
DeltaMRP is a helper function that is used in the multi range proof

\delta(y, z) = (z-z^2)<1^n, y^n> - \sum_j z^3+j<1^n, 2^n>
*/

func DeltaMRP(y []*big.Int, z *big.Int, m int) *big.Int {
	result := big.NewInt(0)

	// (z-z^2)<1^n, y^n>
	z2 := new(big.Int).Mod(new(big.Int).Mul(z, z), privacy.Curve.Params().N)
	t1 := new(big.Int).Mod(new(big.Int).Sub(z, z2), privacy.Curve.Params().N)
	t2 := new(big.Int).Mod(new(big.Int).Mul(t1, VectorSum(y)), privacy.Curve.Params().N)

	// \sum_j z^3+j<1^n, 2^n>
	// <1^n, 2^n> = 2^n - 1
	po2sum := new(big.Int).Sub(new(big.Int).Exp(big.NewInt(2), big.NewInt(int64(RangeProofParams.V/m)), privacy.Curve.Params().N), big.NewInt(1))
	t3 := big.NewInt(0)

	for j := 0; j < m; j++ {
		zp := new(big.Int).Exp(z, big.NewInt(3+int64(j)), privacy.Curve.Params().N)
		tmp1 := new(big.Int).Mod(new(big.Int).Mul(zp, po2sum), privacy.Curve.Params().N)
		t3 = new(big.Int).Mod(new(big.Int).Add(t3, tmp1), privacy.Curve.Params().N)
	}

	result = new(big.Int).Mod(new(big.Int).Sub(t2, t3), privacy.Curve.Params().N)

	return result
}

/*
PKComMultiRangeProof Prove
Takes in a list of values and provides an aggregate
range proof for all the values.

changes:
 all values are concatenated
 r(x) is computed differently
 tau_x calculation is different
 delta calculation is different

{(g, h \in G, \textbf{V} \in G^m ; \textbf{v, \gamma} \in Z_p^m) :
	V_j = h^{\gamma_j}g^{v_j} \wedge v_j \in [0, 2^n - 1] \forall j \in [1, m]}
*/
func (wit *PKComMultiRangeWitness) Prove() PKComMultiRangeProof {
	// RangeProofParams.V has the total number of values and bits we can support

	MRPResult := PKComMultiRangeProof{}

	m := len(wit.Values)
	bitsPerValue := RangeProofParams.V / m

	// we concatenate the binary representation of the values

	PowerOfTwos := PowerVector(bitsPerValue, big.NewInt(2))

	Comms := make([]privacy.EllipticPoint, m)
	gammas := make([]*big.Int, m)
	aLConcat := make([]*big.Int, RangeProofParams.V)
	aRConcat := make([]*big.Int, RangeProofParams.V)

	for j := range wit.Values {
		v := wit.Values[j]
		if v.Cmp(big.NewInt(0)) == -1 {
			fmt.Println("Value is below range! Not proving")
		}

		if v.Cmp(new(big.Int).Exp(big.NewInt(2), big.NewInt(int64(bitsPerValue)), privacy.Curve.Params().N)) == 1 {
			fmt.Println("Value is above range! Not proving.")
			return *new(PKComMultiRangeProof)
		}

		gamma:= new(big.Int).SetBytes(privacy.RandBytes(32))
		gamma.Mod(gamma,privacy.Curve.Params().N)

		Comms[j] = RangeProofParams.G.ScalarMulPoint(v).AddPoint(RangeProofParams.H.ScalarMulPoint(gamma))
		gammas[j] = gamma

		// break up v into its bitwise representation
		aL := reverse(StrToBigIntArray(PadLeft(fmt.Sprintf("%b", v), "0", bitsPerValue)))
		aR := VectorAddScalar(aL, big.NewInt(-1))

		for i := range aR {
			aLConcat[bitsPerValue*j+i] = aL[i]
			aRConcat[bitsPerValue*j+i] = aR[i]
		}
	}

	MRPResult.Comms = Comms

	alpha := new(big.Int).SetBytes(privacy.RandBytes(32))
	alpha.Mod(alpha,privacy.Curve.Params().N)

	A := TwoVectorPCommitWithGens(RangeProofParams.BPG, RangeProofParams.BPH, aLConcat, aRConcat).AddPoint(RangeProofParams.H.ScalarMulPoint((alpha)))
	MRPResult.A = A

	sL := RandVector(RangeProofParams.V)
	sR := RandVector(RangeProofParams.V)

	rho := new(big.Int).SetBytes(privacy.RandBytes(32))
	rho.Mod(alpha,privacy.Curve.Params().N)

	S := TwoVectorPCommitWithGens(RangeProofParams.BPG, RangeProofParams.BPH, sL, sR).AddPoint(RangeProofParams.H.ScalarMulPoint(rho))
	MRPResult.S = S

	chal1s256 := blake2b.Sum256([]byte(A.X.String() + A.Y.String()))
	cy := new(big.Int).SetBytes(chal1s256[:])
	MRPResult.Cy = cy

	chal2s256 := blake2b.Sum256([]byte(S.X.String() + S.Y.String()))
	cz := new(big.Int).SetBytes(chal2s256[:])
	MRPResult.Cz = cz

	zPowersTimesTwoVec := make([]*big.Int, RangeProofParams.V)
	for j := 0; j < m; j++ {
		zp := new(big.Int).Exp(cz, big.NewInt(2+int64(j)), privacy.Curve.Params().N)
		for i := 0; i < bitsPerValue; i++ {
			zPowersTimesTwoVec[j*bitsPerValue+i] = new(big.Int).Mod(new(big.Int).Mul(PowerOfTwos[i], zp), privacy.Curve.Params().N)
		}
	}

	//fmt.Println(zPowersTimesTwoVec)

	// need to generate l(X), r(X), and t(X)=<l(X),r(X)>
	PowerOfCY := PowerVector(RangeProofParams.V, cy)
	// fmt.Println(PowerOfCY)
	l0 := VectorAddScalar(aLConcat, new(big.Int).Neg(cz))
	l1 := sL
	r0 := VectorAdd(
		VectorHadamard(
			PowerOfCY,
			VectorAddScalar(aRConcat, cz)),
		zPowersTimesTwoVec)
	r1 := VectorHadamard(sR, PowerOfCY)

	//calculate t0
	vz2 := big.NewInt(0)
	z2 := new(big.Int).Mod(new(big.Int).Mul(cz, cz), privacy.Curve.Params().N)
	PowerOfCZ := PowerVector(m, cz)
	for j := 0; j < m; j++ {
		vz2 = new(big.Int).Add(vz2,
			new(big.Int).Mul(
				PowerOfCZ[j],
				new(big.Int).Mul(wit.Values[j], z2)))
		vz2 = new(big.Int).Mod(vz2, privacy.Curve.Params().N)
	}

	t0 := new(big.Int).Mod(new(big.Int).Add(vz2, DeltaMRP(PowerOfCY, cz, m)), privacy.Curve.Params().N)

	t1 := new(big.Int).Mod(new(big.Int).Add(InnerProduct(l1, r0), InnerProduct(l0, r1)), privacy.Curve.Params().N)
	t2 := InnerProduct(l1, r1)

	// given the t_i values, we can generate commitments to them
	tau1 := new(big.Int).SetBytes(privacy.RandBytes(32))
	tau1.Mod(tau1,privacy.Curve.Params().N)

	tau2 := new(big.Int).SetBytes(privacy.RandBytes(32))
	tau2.Mod(tau2,privacy.Curve.Params().N)

	T1 := RangeProofParams.G.ScalarMulPoint(t1).AddPoint(RangeProofParams.H.ScalarMulPoint(tau1)) //commitment to t1
	T2 := RangeProofParams.G.ScalarMulPoint(t2).AddPoint(RangeProofParams.H.ScalarMulPoint(tau2)) //commitment to t2

	MRPResult.T1 = T1
	MRPResult.T2 = T2

	chal3s256 := blake2b.Sum256([]byte(T1.X.String() + T1.Y.String() + T2.X.String() + T2.Y.String()))
	cx := new(big.Int).SetBytes(chal3s256[:])

	MRPResult.Cx = cx

	left := CalculateLMRP(aLConcat, sL, cz, cx)
	right := CalculateRMRP(aRConcat, sR, PowerOfCY, zPowersTimesTwoVec, cz, cx)

	thatPrime := new(big.Int).Mod( // t0 + t1*x + t2*x^2
		new(big.Int).Add(t0, new(big.Int).Add(new(big.Int).Mul(t1, cx), new(big.Int).Mul(new(big.Int).Mul(cx, cx), t2))), privacy.Curve.Params().N)

	that := InnerProduct(left, right) // NOTE: BP Java implementation calculates this from the t_i

	// thatPrime and that should be equal
	if thatPrime.Cmp(that) != 0 {
		fmt.Println("Proving -- Uh oh! Two diff ways to compute same value not working")
		fmt.Printf("\tthatPrime = %s\n", thatPrime.String())
		fmt.Printf("\tthat = %s \n", that.String())
	}

	MRPResult.Th = that

	vecRandomnessTotal := big.NewInt(0)
	for j := 0; j < m; j++ {
		zp := new(big.Int).Exp(cz, big.NewInt(2+int64(j)), privacy.Curve.Params().N)
		tmp1 := new(big.Int).Mul(gammas[j], zp)
		vecRandomnessTotal = new(big.Int).Mod(new(big.Int).Add(vecRandomnessTotal, tmp1), privacy.Curve.Params().N)
	}
	//fmt.Println(vecRandomnessTotal)
	taux1 := new(big.Int).Mod(new(big.Int).Mul(tau2, new(big.Int).Mul(cx, cx)), privacy.Curve.Params().N)
	taux2 := new(big.Int).Mod(new(big.Int).Mul(tau1, cx), privacy.Curve.Params().N)
	taux := new(big.Int).Mod(new(big.Int).Add(taux1, new(big.Int).Add(taux2, vecRandomnessTotal)), privacy.Curve.Params().N)

	MRPResult.Tau = taux

	mu := new(big.Int).Mod(new(big.Int).Add(alpha, new(big.Int).Mul(rho, cx)), privacy.Curve.Params().N)
	MRPResult.Mu = mu

	HPrime := make([]privacy.EllipticPoint, len(RangeProofParams.BPH))

	for i := range HPrime {
		HPrime[i] = RangeProofParams.BPH[i].ScalarMulPoint(new(big.Int).ModInverse(PowerOfCY[i], privacy.Curve.Params().N))
	}

	P := TwoVectorPCommitWithGens(RangeProofParams.BPG, HPrime, left, right)
	//fmt.Println(P)

	MRPResult.IPP = InnerProductProve(left, right, that, P, RangeProofParams.U, RangeProofParams.BPG, HPrime)

	return MRPResult
}

/*
PKComMultiRangeProof Verify
Takes in a PKComMultiRangeProof and verifies its correctness

*/
func (pro *PKComMultiRangeProof) Verify() bool {

	m := len(pro.Comms)
	if (m==0) {
		return false
	}
	bitsPerValue := RangeProofParams.V / m

	//changes:
	// check 1 changes since it includes all commitments
	// check 2 commitment generation is also different

	// verify the challenges
	chal1s256 := blake2b.Sum256([]byte(pro.A.X.String() + pro.A.Y.String()))
	cy := new(big.Int).SetBytes(chal1s256[:])
	if cy.Cmp(pro.Cy) != 0 {
		fmt.Println("MRPVerify - Challenge Cy failing!")
		return false
	}
	chal2s256 := blake2b.Sum256([]byte(pro.S.X.String() + pro.S.Y.String()))
	cz := new(big.Int).SetBytes(chal2s256[:])
	if cz.Cmp(pro.Cz) != 0 {
		fmt.Println("MRPVerify - Challenge Cz failing!")
		return false
	}
	chal3s256 := blake2b.Sum256([]byte(pro.T1.X.String() + pro.T1.Y.String() + pro.T2.X.String() + pro.T2.Y.String()))
	cx := new(big.Int).SetBytes(chal3s256[:])
	if cx.Cmp(pro.Cx) != 0 {
		fmt.Println("RPVerify - Challenge Cx failing!")
		return false
	}

	// given challenges are correct, very range proof
	PowersOfY := PowerVector(RangeProofParams.V, cy)

	// t_hat * G + tau * H
	lhs := RangeProofParams.G.ScalarMulPoint(pro.Th).AddPoint(RangeProofParams.H.ScalarMulPoint(pro.Tau))

	// z^2 * \bold{z}^m \bold{V} + delta(y,z) * G + x * T1 + x^2 * T2
	CommPowers := RangeProofParams.Zero()
	PowersOfZ := PowerVector(m, cz)
	z2 := new(big.Int).Mod(new(big.Int).Mul(cz, cz), privacy.Curve.Params().N)

	for j := 0; j < m; j++ {
		CommPowers = CommPowers.AddPoint(pro.Comms[j].ScalarMulPoint(new(big.Int).Mul(z2, PowersOfZ[j])))
	}

	rhs := RangeProofParams.G.ScalarMulPoint(DeltaMRP(PowersOfY, cz, m)).AddPoint(
		pro.T1.ScalarMulPoint(cx)).AddPoint(
		pro.T2.ScalarMulPoint(new(big.Int).Mul(cx, cx))).AddPoint(CommPowers)

	if !lhs.IsEqual(rhs) {
		fmt.Println("MRPVerify - Uh oh! Check line (63) of verification")
		fmt.Println(rhs)
		fmt.Println(lhs)
		return false
	}

	tmp1 := RangeProofParams.Zero()
	zneg := new(big.Int).Mod(new(big.Int).Neg(cz), privacy.Curve.Params().N)
	for i := range RangeProofParams.BPG {
		tmp1 = tmp1.AddPoint(RangeProofParams.BPG[i].ScalarMulPoint(zneg))
	}

	PowerOfTwos := PowerVector(bitsPerValue, big.NewInt(2))
	tmp2 := RangeProofParams.Zero()
	// generate h'
	HPrime := make([]privacy.EllipticPoint, len(RangeProofParams.BPH))

	for i := range HPrime {
		mi := new(big.Int).ModInverse(PowersOfY[i], privacy.Curve.Params().N)
		HPrime[i] = RangeProofParams.BPH[i].ScalarMulPoint(mi)
	}

	for j := 0; j < m; j++ {
		for i := 0; i < bitsPerValue; i++ {
			val1 := new(big.Int).Mul(cz, PowersOfY[j*bitsPerValue+i])
			zp := new(big.Int).Exp(cz, big.NewInt(2+int64(j)), privacy.Curve.Params().N)
			val2 := new(big.Int).Mod(new(big.Int).Mul(zp, PowerOfTwos[i]), privacy.Curve.Params().N)
			tmp2 = tmp2.AddPoint(HPrime[j*bitsPerValue+i].ScalarMulPoint(new(big.Int).Add(val1, val2)))
		}
	}

	// without subtracting this value should equal muCH + l[i]G[i] + r[i]H'[i]
	// we want to make sure that the innerproduct checks out, so we subtract it
	tmp,_:=RangeProofParams.H.ScalarMulPoint(pro.Mu).Inverse()
	P := pro.A.AddPoint(pro.S.ScalarMulPoint(cx)).AddPoint(tmp1).AddPoint(tmp2).AddPoint(*tmp)
	//fmt.Println(P)

	if !InnerProductVerifyFast(pro.Th, P, RangeProofParams.U, RangeProofParams.BPG, HPrime, pro.IPP) {
		fmt.Println("MRPVerify - Uh oh! Check line (65) of verification!")
		return false
	}
	return true
}
func TestPKComMultiRange() {
	test:=8
	values := make([]*big.Int,test)
	for i:=0;i<test;i++{
		values[i] = new(big.Int)
		x:=new(big.Int).SetBytes(privacy.RandBytes(8))
		fmt.Println(x)
		values[i]=x
	}
	//values := []*big.Int{big.NewInt(5136325419070411678), big.NewInt()}
	var witness PKComMultiRangeWitness
	witness.Values = values
		RangeProofParams = NewECPrimeGroupKey(64 * len(values))
	// Testing smallest number in range
	proof:=witness.Prove()
	if proof.Verify() {
		fmt.Println("Multi Range Proof Verification works")
	} else {
		fmt.Println("***** Multi Range Proof FAILURE")
	}
}



