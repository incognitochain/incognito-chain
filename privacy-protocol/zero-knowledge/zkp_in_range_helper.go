package zkp

import (
	"crypto/elliptic"
	"fmt"
	"github.com/minio/blake2b-simd"
	"github.com/ninjadotorg/constant/privacy-protocol"
	"math"
	"math/big"
	"strconv"
)
var RangeProofParams CryptoParams

/* ------------ Inner Product Functions ---------------*/
type InnerProdArg struct {
	L []privacy.EllipticPoint
	R []privacy.EllipticPoint
	A *big.Int
	B *big.Int

	Challenges []*big.Int
}

func GenerateNewParams(G, H []privacy.EllipticPoint, x *big.Int, L, R, P privacy.EllipticPoint) ([]privacy.EllipticPoint, []privacy.EllipticPoint, privacy.EllipticPoint) {
	nprime := len(G) / 2

	Gprime := make([]privacy.EllipticPoint, nprime)
	Hprime := make([]privacy.EllipticPoint, nprime)

	xinv := new(big.Int).ModInverse(x, privacy.Curve.Params().N)

	// Gprime = xinv * G[:nprime] + x*G[nprime:]
	// Hprime = x * H[:nprime] + xinv*H[nprime:]

	for i := range Gprime {
		//fmt.Printf("i: %d && i+nprime: %d\n", i, i+nprime)
		Gprime[i] = G[i].ScalarMulPoint(xinv).AddPoint(G[i+nprime].ScalarMulPoint(x))
		Hprime[i] = H[i].ScalarMulPoint(x).AddPoint(H[i+nprime].ScalarMulPoint(xinv))
	}

	x2 := new(big.Int).Mod(new(big.Int).Mul(x, x), privacy.Curve.Params().N)
	xinv2 := new(big.Int).ModInverse(x2, privacy.Curve.Params().N)

	Pprime := L.ScalarMulPoint(x2).AddPoint(P).AddPoint(R.ScalarMulPoint(xinv2)) // x^2 * L + P + xinv^2 * R

	return Gprime, Hprime, Pprime
}

/* Inner Product Argument

Proves that <a,b>=c

This is a building block for BulletProofs

*/
func InnerProductProveSub(proof InnerProdArg, G, H []privacy.EllipticPoint, a []*big.Int, b []*big.Int, u privacy.EllipticPoint, P privacy.EllipticPoint) InnerProdArg {
	//fmt.Printf("Proof so far: %s\n", proof)
	if len(a) == 1 {
		// Prover sends a & b
		//fmt.Printf("a: %d && b: %d\n", a[0], b[0])
		proof.A = a[0]
		proof.B = b[0]
		return proof
	}

	curIt := int(math.Log2(float64(len(a)))) - 1

	nprime := len(a) / 2
	//fmt.Println(nprime)
	//fmt.Println(len(H))
	cl := InnerProduct(a[:nprime], b[nprime:]) // either this line
	cr := InnerProduct(a[nprime:], b[:nprime]) // or this line
	L := TwoVectorPCommitWithGens(G[nprime:], H[:nprime], a[:nprime], b[nprime:]).AddPoint(u.ScalarMulPoint(cl))
	R := TwoVectorPCommitWithGens(G[:nprime], H[nprime:], a[nprime:], b[:nprime]).AddPoint(u.ScalarMulPoint(cr))

	proof.L[curIt] = L
	proof.R[curIt] = R

	// prover sends L & R and gets a challenge
	s256 := blake2b.Sum256([]byte(
		L.X.String() + L.Y.String() +
			R.X.String() + R.Y.String()))

	x := new(big.Int).SetBytes(s256[:])

	proof.Challenges[curIt] = x

	Gprime, Hprime, Pprime := GenerateNewParams(G, H, x, L, R, P)
	//fmt.Printf("Prover - Intermediate Pprime value: %s \n", Pprime)
	xinv := new(big.Int).ModInverse(x, privacy.Curve.Params().N)

	// or these two lines
	aprime := VectorAdd(
		ScalarVectorMul(a[:nprime], x),
		ScalarVectorMul(a[nprime:], xinv))
	bprime := VectorAdd(
		ScalarVectorMul(b[:nprime], xinv),
		ScalarVectorMul(b[nprime:], x))

	return InnerProductProveSub(proof, Gprime, Hprime, aprime, bprime, u, Pprime)
}

func InnerProductProve(a []*big.Int, b []*big.Int, c *big.Int, P, U privacy.EllipticPoint, G, H []privacy.EllipticPoint) InnerProdArg {
	loglen := int(math.Log2(float64(len(a))))

	challenges := make([]*big.Int, loglen+1)
	Lvals := make([]privacy.EllipticPoint, loglen)
	Rvals := make([]privacy.EllipticPoint, loglen)

	runningProof := InnerProdArg{
		Lvals,
		Rvals,
		big.NewInt(0),
		big.NewInt(0),
		challenges}

	// randomly generate an x value from public data
	x := blake2b.Sum256([]byte(P.X.String() + P.Y.String()))

	runningProof.Challenges[loglen] = new(big.Int).SetBytes(x[:])

	Pprime := P.AddPoint(U.ScalarMulPoint(new(big.Int).Mul(new(big.Int).SetBytes(x[:]), c)))
	ux := U.ScalarMulPoint(new(big.Int).SetBytes(x[:]))
	//fmt.Printf("Prover Pprime value to run sub off of: %s\n", Pprime)
	return InnerProductProveSub(runningProof, G, H, a, b, ux, Pprime)
}

/* Inner Product Verify
Given a inner product proof, verifies the correctness of the proof

Since we're using the Fiat-Shamir transform, we need to verify all x hash computations,
all g' and h' computations

P : the Pedersen commitment we are verifying is a commitment to the innner product
ipp : the proof

*/
func InnerProductVerify(c *big.Int, P, U privacy.EllipticPoint, G, H []privacy.EllipticPoint, ipp InnerProdArg) bool {
	//fmt.Println("Verifying Inner Product Argument")
	//fmt.Printf("Commitment Value: %s \n", P)
	s1 := blake2b.Sum256([]byte(P.X.String() + P.Y.String()))
	chal1 := new(big.Int).SetBytes(s1[:])
	ux := U.ScalarMulPoint(chal1)
	curIt := len(ipp.Challenges) - 1

	if ipp.Challenges[curIt].Cmp(chal1) != 0 {
		fmt.Println("IPVerify - Initial Challenge Failed")
		return false
	}

	curIt -= 1

	Gprime := G
	Hprime := H
	Pprime := P.AddPoint(ux.ScalarMulPoint(c)) // line 6 from protocol 1
	//fmt.Printf("New Commitment value with u^cx: %s \n", Pprime)

	for curIt >= 0 {
		Lval := ipp.L[curIt]
		Rval := ipp.R[curIt]

		// prover sends L & R and gets a challenge
		s256 := blake2b.Sum256([]byte(
			Lval.X.String() + Lval.Y.String() +
				Rval.X.String() + Rval.Y.String()))

		chal2 := new(big.Int).SetBytes(s256[:])

		if ipp.Challenges[curIt].Cmp(chal2) != 0 {
			fmt.Println("IPVerify - Challenge verification failed at index " + strconv.Itoa(curIt))
			return false
		}

		Gprime, Hprime, Pprime = GenerateNewParams(Gprime, Hprime, chal2, Lval, Rval, Pprime)
		curIt -= 1
	}
	ccalc := new(big.Int).Mod(new(big.Int).Mul(ipp.A, ipp.B), privacy.Curve.Params().N)

	Pcalc1 := Gprime[0].ScalarMulPoint(ipp.A)
	Pcalc2 := Hprime[0].ScalarMulPoint(ipp.B)
	Pcalc3 := ux.ScalarMulPoint(ccalc)
	Pcalc := Pcalc1.AddPoint(Pcalc2).AddPoint(Pcalc3)

	if !Pprime.IsEqual(Pcalc) {
		fmt.Println("IPVerify - Final Commitment checking failed")
		fmt.Printf("Final Pprime value: %s \n", Pprime)
		fmt.Printf("Calculated Pprime value to check against: %s \n", Pcalc)
		return false
	}

	return true
}

/* Inner Product Verify Fast
Given a inner product proof, verifies the correctness of the proof. Does the same as above except
we replace n separate exponentiations with a single ScalarMulPointi-exponentiation.
*/

func InnerProductVerifyFast(c *big.Int, P, U privacy.EllipticPoint, G, H []privacy.EllipticPoint, ipp InnerProdArg) bool {
	//fmt.Println("Verifying Inner Product Argument")
	//fmt.Printf("Commitment Value: %s \n", P)
	s1 := blake2b.Sum256([]byte(P.X.String() + P.Y.String()))
	chal1 := new(big.Int).SetBytes(s1[:])
	ux := U.ScalarMulPoint(chal1)
	curIt := len(ipp.Challenges) - 1

	// check all challenges
	if ipp.Challenges[curIt].Cmp(chal1) != 0 {
		fmt.Println("IPVerify - Initial Challenge Failed")
		return false
	}

	for j := curIt - 1; j >= 0; j-- {
		Lval := ipp.L[j]
		Rval := ipp.R[j]

		// prover sends L & R and gets a challenge
		s256 := blake2b.Sum256([]byte(
			Lval.X.String() + Lval.Y.String() +
				Rval.X.String() + Rval.Y.String()))

		chal2 := new(big.Int).SetBytes(s256[:])

		if ipp.Challenges[j].Cmp(chal2) != 0 {
			fmt.Println("IPVerify - Challenge verification failed at index " + strconv.Itoa(j))
			return false
		}
	}
	// begin computing

	curIt -= 1
	Pprime := P.AddPoint(ux.ScalarMulPoint(c)) // line 6 from protocol 1

	tmp1 := RangeProofParams.Zero()
	for j := curIt; j >= 0; j-- {
		x2 := new(big.Int).Exp(ipp.Challenges[j], big.NewInt(2), privacy.Curve.Params().N)
		x2i := new(big.Int).ModInverse(x2, privacy.Curve.Params().N)
		//fmt.Println(tmp1)
		tmp1 = ipp.L[j].ScalarMulPoint(x2).AddPoint(ipp.R[j].ScalarMulPoint(x2i)).AddPoint(tmp1)
		//fmt.Println(tmp1)
	}
	rhs := Pprime.AddPoint(tmp1)

	sScalars := make([]*big.Int, RangeProofParams.V)
	invsScalars := make([]*big.Int, RangeProofParams.V)

	for i := 0; i < RangeProofParams.V; i++ {
		si := big.NewInt(1)
		for j := curIt; j >= 0; j-- {
			// original challenge if the jth bit of i is 1, inverse challenge otherwise
			chal := ipp.Challenges[j]
			if big.NewInt(int64(i)).Bit(j) == 0 {
				chal = new(big.Int).ModInverse(chal, privacy.Curve.Params().N)
			}
			// fmt.Printf("Challenge raised to value: %d\n", chal)
			si = new(big.Int).Mod(new(big.Int).Mul(si, chal), privacy.Curve.Params().N)
		}
		//fmt.Printf("Si value: %d\n", si)
		sScalars[i] = si
		invsScalars[i] = new(big.Int).ModInverse(si, privacy.Curve.Params().N)
	}

	ccalc := new(big.Int).Mod(new(big.Int).Mul(ipp.A, ipp.B), privacy.Curve.Params().N)
	lhs := TwoVectorPCommitWithGens(G, H, ScalarVectorMul(sScalars, ipp.A), ScalarVectorMul(invsScalars, ipp.B)).AddPoint(ux.ScalarMulPoint(ccalc))

	if !rhs.IsEqual(lhs) {
		fmt.Println("IPVerify - Final Commitment checking failed")
		fmt.Printf("Final rhs value: %s \n", rhs)
		fmt.Printf("Final lhs value: %s \n", lhs)
		return false
	}
	return true
}
/*-----------------------------Vector Functions-----------------------------*/
// The length here always has to be a power of two
func InnerProduct(a []*big.Int, b []*big.Int) *big.Int {
	if len(a) != len(b) {
		fmt.Println("InnerProduct: Uh oh! Arrays not of the same length")
		fmt.Printf("len(a): %d\n", len(a))
		fmt.Printf("len(b): %d\n", len(b))
	}

	c := big.NewInt(0)

	for i := range a {
		tmp1 := new(big.Int).Mul(a[i], b[i])
		c = new(big.Int).Add(c, new(big.Int).Mod(tmp1, privacy.Curve.Params().N))
	}

	return new(big.Int).Mod(c, privacy.Curve.Params().N)
}

func VectorAdd(v []*big.Int, w []*big.Int) []*big.Int {
	if len(v) != len(w) {
		fmt.Println("VectorAddPoint: Uh oh! Arrays not of the same length")
		fmt.Printf("len(v): %d\n", len(v))
		fmt.Printf("len(w): %d\n", len(w))
	}
	result := make([]*big.Int, len(v))

	for i := range v {
		result[i] = new(big.Int).Mod(new(big.Int).Add(v[i], w[i]), privacy.Curve.Params().N)
	}

	return result
}

func VectorHadamard(v, w []*big.Int) []*big.Int {
	if len(v) != len(w) {
		fmt.Println("VectorHadamard: Uh oh! Arrays not of the same length")
		fmt.Printf("len(v): %d\n", len(w))
		fmt.Printf("len(w): %d\n", len(v))
	}

	result := make([]*big.Int, len(v))

	for i := range v {
		result[i] = new(big.Int).Mod(new(big.Int).Mul(v[i], w[i]), privacy.Curve.Params().N)
	}

	return result
}

func VectorAddScalar(v []*big.Int, s *big.Int) []*big.Int {
	result := make([]*big.Int, len(v))

	for i := range v {
		result[i] = new(big.Int).Mod(new(big.Int).Add(v[i], s), privacy.Curve.Params().N)
	}

	return result
}

func ScalarVectorMul(v []*big.Int, s *big.Int) []*big.Int {
	result := make([]*big.Int, len(v))

	for i := range v {
		result[i] = new(big.Int).Mod(new(big.Int).Mul(v[i], s), privacy.Curve.Params().N)
	}

	return result
}

// from here: https://play.golang.org/p/zciRZvD0Gr with a fix
func PadLeft(str, pad string, l int) string {
	strCopy := str
	for len(strCopy) < l {
		strCopy = pad + strCopy
	}

	return strCopy
}

func STRNot(str string) string {
	result := ""

	for _, i := range str {
		if i == '0' {
			result += "1"
		} else {
			result += "0"
		}
	}
	return result
}

func StrToBigIntArray(str string) []*big.Int {
	result := make([]*big.Int, len(str))

	for i := range str {
		t, success := new(big.Int).SetString(string(str[i]), 10)
		if success {
			result[i] = t
		}
	}

	return result
}

func reverse(l []*big.Int) []*big.Int {
	result := make([]*big.Int, len(l))

	for i := range l {
		result[i] = l[len(l)-i-1]
	}

	return result
}

func PowerVector(l int, base *big.Int) []*big.Int {
	result := make([]*big.Int, l)

	for i := 0; i < l; i++ {
		result[i] = new(big.Int).Exp(base, big.NewInt(int64(i)), privacy.Curve.Params().N)
	}

	return result
}

func RandVector(l int) []*big.Int {
	result := make([]*big.Int, l)

	for i := 0; i < l; i++ {
		x:= new(big.Int).SetBytes(privacy.RandBytes(32))
		x.Mod(x,privacy.Curve.Params().N)
		result[i] = x
	}

	return result
}

func VectorSum(y []*big.Int) *big.Int {
	result := big.NewInt(0)

	for _, j := range y {
		result = new(big.Int).Mod(new(big.Int).Add(result, j), privacy.Curve.Params().N)
	}

	return result
}
/*-----------------------Crypto Params Functions------------------*/

var VecLength int

type CryptoParams struct {
	C   elliptic.Curve      // curve
	BPG []privacy.EllipticPoint           // slice of gen 1 for BP
	BPH []privacy.EllipticPoint           // slice of gen 2 for BP
	N   *big.Int            // scalar prime
	U   privacy.EllipticPoint             // a point that is a fixed group element with an unknown discrete-log relative to g,h
	V   int                 // Vector length
	G   privacy.EllipticPoint             // G value for commitments of a single value
	H   privacy.EllipticPoint             // H value for commitments of a single value
}

func (c CryptoParams) Zero() privacy.EllipticPoint {
	return privacy.EllipticPoint{big.NewInt(0), big.NewInt(0)}
}


// NewECPrimeGroupKey returns the curve (field),
// Generator 1 x&y, Generator 2 x&y, order of the generators
func NewECPrimeGroupKey(n int) CryptoParams {

	gen1Vals := make([]privacy.EllipticPoint, n)
	gen2Vals := make([]privacy.EllipticPoint, n)
	u := privacy.EllipticPoint{big.NewInt(0), big.NewInt(0)}

	G:=privacy.PedCom.G[privacy.VALUE]
	H:=privacy.PedCom.G[privacy.RAND]

	for i:=0;i<n;i++{
		gen1Vals[i]= G.Hash(0)
		G=G.Hash(0)
		gen2Vals[i]= H.Hash(0)
		H = H.Hash(0)
	}
	u	= G.AddPoint(H).Hash(0)
	return CryptoParams{
		privacy.Curve,
		gen1Vals,
		gen2Vals,
		privacy.Curve.Params().N,
		u,
		n,
		privacy.PedCom.G[privacy.VALUE],
		privacy.PedCom.G[privacy.RAND]}
}

/*Perdersen commit for 2 vector*/
func TwoVectorPCommitWithGens(G, H []privacy.EllipticPoint, a, b []*big.Int) privacy.EllipticPoint {
	if len(G) != len(H) || len(G) != len(a) || len(a) != len(b) {
		fmt.Println("TwoVectorPCommitWithGens: Uh oh! Arrays not of the same length")
		fmt.Printf("len(G): %d\n", len(G))
		fmt.Printf("len(H): %d\n", len(H))
		fmt.Printf("len(a): %d\n", len(a))
		fmt.Printf("len(b): %d\n", len(b))
	}

	commitment := privacy.EllipticPoint{big.NewInt(0), big.NewInt(0)}

	for i := 0; i < len(G); i++ {
		modA := new(big.Int).Mod(a[i], privacy.Curve.Params().N)
		modB := new(big.Int).Mod(b[i], privacy.Curve.Params().N)
		commitment = commitment.AddPoint(G[i].ScalarMulPoint(modA)).AddPoint(H[i].ScalarMulPoint(modB))
	}

	return commitment
}
