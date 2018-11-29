package zkp

import (
	"fmt"
	"github.com/minio/blake2b-simd"
	"github.com/ninjadotorg/constant/privacy-protocol"
	"math"
	"math/big"
)
type PKComMultiRangeProof struct {
	Comms            []privacy.EllipticPoint
	A                privacy.EllipticPoint
	S                privacy.EllipticPoint
	T1               privacy.EllipticPoint
	T2               privacy.EllipticPoint
	Tau              *big.Int
	Th               *big.Int
	Mu               *big.Int
	IPP              InnerProdArg
	maxExp 					 int
	// challenges
	Cy *big.Int
	Cz *big.Int
	Cx *big.Int
}

type PKComMultiRangeWitness struct {
	Comms  [] privacy.EllipticPoint
	Values [] *big.Int
	Rands  [] *big.Int
	maxExp int
}
func pad(l int) int {
	deg:=0
	for l>0 {
		if (l%2==0){
			deg++
			l = l/2
		} else {break}
	}
	i:=0
	for {
		if (math.Pow(2,float64(i))< float64(l)){
			i++
		} else {
			l = int(math.Pow(2,float64(i+deg)))
			break
		}
	}
	return l
}
func InitCommonParams(l int, maxExp int){
	VecLength =  maxExp * pad(l)
	RangeProofParams = NewECPrimeGroupKey(VecLength)
}
func (wit *PKComMultiRangeWitness) Set(v []*big.Int, maxExp int){
	l:=pad(len(v)+1)
	wit.Values = make([]*big.Int,l)
	for i:=0;i<l;i++{
		wit.Values[i] = new(big.Int)
		wit.Values[i].SetInt64(0)
	}
	total:=new(big.Int).SetUint64(0)
	for i:=0;i<len(v);i++{
		wit.Values[i] = new(big.Int)
		*wit.Values[i] = *v[i]
		total.Add(total,v[i])
	}
	*wit.Values[l-1] = *total
	wit.maxExp = maxExp
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
func (wit *PKComMultiRangeWitness) Prove() (*PKComMultiRangeProof,error) {
	// RangeProofParams.V has the total number of values and bits we can support

	InitCommonParams(len(wit.Values),wit.maxExp)
	MRProof := PKComMultiRangeProof{}
	MRProof.maxExp = wit.maxExp
	m := len(wit.Values)
	bitsPerValue := RangeProofParams.V / m
	// we concatenate the binary representation of the values
	PowerOfTwos := PowerVector(bitsPerValue, big.NewInt(2))
	Comms := make([]privacy.EllipticPoint, m)
	gammas := make([]*big.Int, m)
	wit.Rands=make([]*big.Int, m)
	aLConcat := make([]*big.Int, RangeProofParams.V)
	aRConcat := make([]*big.Int, RangeProofParams.V)
	sumRand:=new (big.Int)
	sumRand.SetUint64(0)
	for j := range wit.Values {
		v := wit.Values[j]
		if v.Cmp(big.NewInt(0)) == -1 {
			fmt.Println("H is below range! Not proving")
		}
		if v.Cmp(new(big.Int).Exp(big.NewInt(2), big.NewInt(int64(wit.maxExp)), privacy.Curve.Params().N)) == 1 {
			fmt.Println("Value is above range! Not proving.")
			err:= fmt.Errorf("Value is above range! Not proving.")
			return new(PKComMultiRangeProof),err
		}
	gamma:= new(big.Int).SetBytes(privacy.RandBytes(32))
		gamma.Mod(gamma,privacy.Curve.Params().N)
		Comms[j] = RangeProofParams.G.ScalarMulPoint(v).AddPoint(RangeProofParams.H.ScalarMulPoint(gamma))
		gammas[j] = gamma
		wit.Rands[j] = gamma
		// break up v into its bitwise representation
		aL := reverse(StrToBigIntArray(PadLeft(fmt.Sprintf("%b", v), "0", bitsPerValue)))
		aR := VectorAddScalar(aL, big.NewInt(-1))
		for i := range aR {
			aLConcat[bitsPerValue*j+i] = aL[i]
			aRConcat[bitsPerValue*j+i] = aR[i]
		}
	}
	MRProof.Comms = Comms
	wit.Comms = Comms
	alpha := new(big.Int).SetBytes(privacy.RandBytes(32))
	alpha.Mod(alpha,privacy.Curve.Params().N)

	A := TwoVectorPCommitWithGens(RangeProofParams.BPG, RangeProofParams.BPH, aLConcat, aRConcat).AddPoint(RangeProofParams.H.ScalarMulPoint((alpha)))
	MRProof.A = A

	sL := RandVector(RangeProofParams.V)
	sR := RandVector(RangeProofParams.V)

	rho := new(big.Int).SetBytes(privacy.RandBytes(32))
	rho.Mod(alpha,privacy.Curve.Params().N)

	S := TwoVectorPCommitWithGens(RangeProofParams.BPG, RangeProofParams.BPH, sL, sR).AddPoint(RangeProofParams.H.ScalarMulPoint(rho))
	MRProof.S = S

	chal1s256 := blake2b.Sum256([]byte(A.X.String() + A.Y.String()))
	cy := new(big.Int).SetBytes(chal1s256[:])
	MRProof.Cy = cy

	chal2s256 := blake2b.Sum256([]byte(S.X.String() + S.Y.String()))
	cz := new(big.Int).SetBytes(chal2s256[:])
	MRProof.Cz = cz

	zPowersTimesTwoVec := make([]*big.Int, RangeProofParams.V)
	for j := 0; j < m; j++ {
		zp := new(big.Int).Exp(cz, big.NewInt(2+int64(j)), privacy.Curve.Params().N)
		for i := 0; i < bitsPerValue; i++ {
			zPowersTimesTwoVec[j*bitsPerValue+i] = new(big.Int).Mod(new(big.Int).Mul(PowerOfTwos[i], zp), privacy.Curve.Params().N)
		}
	}
	// need to generate l(X), r(X), and t(X)=<l(X),r(X)>
	PowerOfCY := PowerVector(RangeProofParams.V, cy)
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

	MRProof.T1 = T1
	MRProof.T2 = T2

	chal3s256 := blake2b.Sum256([]byte(T1.X.String() + T1.Y.String() + T2.X.String() + T2.Y.String()))
	cx := new(big.Int).SetBytes(chal3s256[:])

	MRProof.Cx = cx

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
	MRProof.Th = that
	vecRandomnessTotal := big.NewInt(0)
	for j := 0; j < m; j++ {
		zp := new(big.Int).Exp(cz, big.NewInt(2+int64(j)), privacy.Curve.Params().N)
		tmp1 := new(big.Int).Mul(gammas[j], zp)
		vecRandomnessTotal = new(big.Int).Mod(new(big.Int).Add(vecRandomnessTotal, tmp1), privacy.Curve.Params().N)
	}
	taux1 := new(big.Int).Mod(new(big.Int).Mul(tau2, new(big.Int).Mul(cx, cx)), privacy.Curve.Params().N)
	taux2 := new(big.Int).Mod(new(big.Int).Mul(tau1, cx), privacy.Curve.Params().N)
	taux := new(big.Int).Mod(new(big.Int).Add(taux1, new(big.Int).Add(taux2, vecRandomnessTotal)), privacy.Curve.Params().N)
	MRProof.Tau = taux
	mu := new(big.Int).Mod(new(big.Int).Add(alpha, new(big.Int).Mul(rho, cx)), privacy.Curve.Params().N)
	MRProof.Mu = mu
	HPrime := make([]privacy.EllipticPoint, len(RangeProofParams.BPH))
	for i := range HPrime {
		HPrime[i] = RangeProofParams.BPH[i].ScalarMulPoint(new(big.Int).ModInverse(PowerOfCY[i], privacy.Curve.Params().N))
	}
	P := TwoVectorPCommitWithGens(RangeProofParams.BPG, HPrime, left, right)
	MRProof.IPP = InnerProductProve(left, right, that, P, RangeProofParams.U, RangeProofParams.BPG, HPrime)
	return &MRProof,nil
}
func (wit *PKComMultiRangeWitness) ProveSum() (*PKComZeroProof, error){
	l := len(wit.Comms)
	if (l==0){
		fmt.Println("Witness for proving sum value is not valid")
		err := fmt.Errorf("Witness for proving sum value is not valid")
		return new(PKComZeroProof), err
	}
	sumComms := RangeProofParams.Zero()
	sumRand := new(big.Int).SetInt64(0)
	temp:=new(privacy.EllipticPoint)
	for i:=0;i<l-1;i++ {
		temp = privacy.PedCom.CommitAtIndex(wit.Values[i], wit.Rands[i],privacy.VALUE)
		sumComms = sumComms.AddPoint(*temp)
		sumRand.Add(sumRand,new(big.Int).Set(wit.Rands[i]))
	}
	sumRand.Sub(sumRand,wit.Rands[l-1])
	temp = privacy.PedCom.CommitAtIndex(wit.Values[l-1], wit.Rands[l-1],privacy.VALUE)
	temp,_ = temp.Inverse()
	sumComms = sumComms.AddPoint(*temp)
	sumRand.Mod(sumRand,privacy.Curve.Params().N)
	var zeroWit PKComZeroWitness
	idx := new(byte)
	*idx = privacy.VALUE
	zeroWit.Set(&sumComms,idx,sumRand)
	return zeroWit.Prove()
}
func (pro *PKComMultiRangeProof) VerifySum(zproof *PKComZeroProof) bool{
	if (zproof.commitmentValue==nil){
		fmt.Println("Proof for verifying sum value is not valid")
		return false
	}
	zeroCom:= RangeProofParams.Zero()
	for i:=0;i<len(pro.Comms)-1;i++{
		zeroCom = zeroCom.AddPoint(pro.Comms[i])
	}
	invSumCom,_:=pro.Comms[len(pro.Comms)-1].Inverse()
	zeroCom = zeroCom.AddPoint(*invSumCom)
	if (!zeroCom.IsEqual(*zproof.commitmentValue)){
		return false
	} else {
		return zproof.Verify()
	}
}
/*
PKComMultiRangeProof Verify
Takes in a PKComMultiRangeProof and verifies its correctness
*/
func (pro *PKComMultiRangeProof) Verify() bool {

	m := len(pro.Comms)
	InitCommonParams(m,pro.maxExp)
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



