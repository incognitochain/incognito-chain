package zkp

import (
	"crypto/elliptic"
	"fmt"
	"github.com/minio/blake2b-simd"
	"github.com/ninjadotorg/constant/privacy"
	"github.com/pkg/errors"
	"math"
	"math/big"
)
var RangeProofParams CryptoParams

/* ------------ Inner Product Functions ---------------*/
type InnerProdArg struct {
	L []*privacy.EllipticPoint
	R []*privacy.EllipticPoint
	A *big.Int
	B *big.Int
	Challenges []*big.Int
}
func (IPA *InnerProdArg) Init(l int) {
	IPA.L= make([]*privacy.EllipticPoint,l)
	IPA.R= make([]*privacy.EllipticPoint,l)
	for i:=0;i<l;i++ {
		IPA.L[i] = new(privacy.EllipticPoint)
		IPA.R[i] = new(privacy.EllipticPoint)
	}
	IPA.A = new(big.Int)
	IPA.B = new(big.Int)
	IPA.Challenges = make([]*big.Int, l+1)
	for i:=0;i<l+1;i++{
		IPA.Challenges[i] = new(big.Int)
	}
}
func makeBigIntArray(l int) []*big.Int{
	result:=make([]*big.Int, l)
	for i:=0;i<l;i++{
		result[i] = new(big.Int).SetInt64(0)
	}
	return result
}
func (IPA InnerProdArg) Print(){
	for i:=0;i<len(IPA.Challenges);i++{
		fmt.Println(IPA.Challenges[i])
	}
}
func (IPA *InnerProdArg) Bytes() []byte{
	var res []byte
	for i:=0;i<len(IPA.L);i++{
		res =append(res, IPA.L[i].Compress()...)
	}
	for i:=0;i<len(IPA.R);i++{
		res = append(res, IPA.R[i].Compress()...)
	}
	for i:=0;i<len(IPA.Challenges);i++{
		res = append(res, privacy.AddPaddingBigInt(IPA.Challenges[i],privacy.BigIntSize)...)
	}
	res = append(res,privacy.AddPaddingBigInt(IPA.A,privacy.BigIntSize)...)
	res = append(res,privacy.AddPaddingBigInt(IPA.B,privacy.BigIntSize)...)
	return res
}
func (IPA *InnerProdArg) SetBytes(IPA_byte []byte){
	offset:=0
	l:=(len(IPA_byte) - 96)/98
	IPA.Init(l)
	L_array_length:=l*privacy.CompressedPointSize
	R_array_length:=L_array_length
	C_array_length:=privacy.BigIntSize*(l+1)
	L_array:=IPA_byte[0:L_array_length]
	offset = L_array_length
	R_array:=IPA_byte[offset:offset+R_array_length]
	offset+=R_array_length
	C_array:=IPA_byte[offset:offset+C_array_length]
	offset+=C_array_length
	offsetL:=0

	for i:=0;i<l;i++{
		IPA.L[i].Decompress(L_array[offsetL:])
		offsetL+=privacy.CompressedPointSize
	}
	offsetR := 0
	for i:=0;i<l;i++{
		IPA.R[i].Decompress(R_array[offsetR:])
		offsetR+=privacy.CompressedPointSize
	}
	offsetC := 0
	for i:=0;i<l+1;i++{
		IPA.Challenges[i].SetBytes(C_array[offsetC:offsetC+privacy.BigIntSize])
		offsetC+=privacy.BigIntSize
	}
	IPA.A.SetBytes(IPA_byte[offset:offset+privacy.BigIntSize])
	IPA.B.SetBytes(IPA_byte[offset+privacy.BigIntSize:offset+2*privacy.BigIntSize])
}
func GenerateNewParams(G, H []*privacy.EllipticPoint, x *big.Int, L, R, P *privacy.EllipticPoint) ([]*privacy.EllipticPoint, []*privacy.EllipticPoint, *privacy.EllipticPoint) {
	nprime := len(G) / 2

	Gprime := make([]*privacy.EllipticPoint, nprime)
	Hprime := make([]*privacy.EllipticPoint, nprime)

	xinv := new(big.Int).ModInverse(x, privacy.Curve.Params().N)

	for i := range Gprime {
		Gprime[i] = G[i].ScalarMult(xinv).Add(G[i+nprime].ScalarMult(x))
		Hprime[i] = H[i].ScalarMult(x).Add(H[i+nprime].ScalarMult(xinv))
	}

	x2 := new(big.Int).Mod(new(big.Int).Mul(x, x), privacy.Curve.Params().N)
	xinv2 := new(big.Int).ModInverse(x2, privacy.Curve.Params().N)

	Pprime := L.ScalarMult(x2).Add(P).Add(R.ScalarMult(xinv2)) // x^2 * L + P + xinv^2 * C1
	return Gprime, Hprime, Pprime
}

/* Inner Product Argument

Proves that <a,b>=c

This is a building block for BulletProofs

*/
func InnerProductProveSub(proof InnerProdArg, G, H []*privacy.EllipticPoint, a []*big.Int, b []*big.Int, u *privacy.EllipticPoint, P *privacy.EllipticPoint) InnerProdArg {
	if len(a) == 1 {
		proof.A = a[0]
		proof.B = b[0]
		return proof
	}
	curIt := int(math.Log2(float64(len(a)))) - 1
	nprime := len(a) / 2
	cl := InnerProduct(a[:nprime], b[nprime:]) // either this line
	cr := InnerProduct(a[nprime:], b[:nprime]) // or this line
	L := TwoVectorPCommitWithGens(G[nprime:], H[:nprime], a[:nprime], b[nprime:]).Add(u.ScalarMult(cl))
	R := TwoVectorPCommitWithGens(G[:nprime], H[nprime:], a[nprime:], b[:nprime]).Add(u.ScalarMult(cr))
	proof.L[curIt] = L
	proof.R[curIt] = R

	// prover sends L & C1 and gets a challenge
	s256 := blake2b.Sum256([]byte(
		L.X.String() + L.Y.String() +
			R.X.String() + R.Y.String()))

	x := new(big.Int).SetBytes(s256[:])

	proof.Challenges[curIt] = x

	Gprime, Hprime, Pprime := GenerateNewParams(G, H, x, L, R, P)
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

func InnerProductProve(a []*big.Int, b []*big.Int, c *big.Int, P, U *privacy.EllipticPoint, G, H []*privacy.EllipticPoint) InnerProdArg {
	loglen := int(math.Log2(float64(len(a))))

	challenges := make([]*big.Int, loglen+1)
	Lvals := make([]*privacy.EllipticPoint, loglen)
	Rvals := make([]*privacy.EllipticPoint, loglen)
	runningProof := InnerProdArg{
		Lvals,
		Rvals,
		big.NewInt(0),
		big.NewInt(0),
		challenges}
	// randomly generate an x value from public data
	x := blake2b.Sum256([]byte(P.X.String() + P.Y.String()))

	runningProof.Challenges[loglen] = new(big.Int).SetBytes(x[:])

	Pprime := P.Add(U.ScalarMult(new(big.Int).Mul(new(big.Int).SetBytes(x[:]), c)))
	ux := U.ScalarMult(new(big.Int).SetBytes(x[:]))
	//fmt.Printf("Prover Pprime value to run sub off of: %s\n", Pprime)
	return InnerProductProveSub(runningProof, G, H, a, b, ux, Pprime)
}

/* Inner Product Verify Fast
Given a inner product proof, verifies the correctness of the proof. Does the same as above except
we replace n separate exponentiations with a single ScalarMulPointi-exponentiation.
*/

func InnerProductVerifyFast(c *big.Int, P, U *privacy.EllipticPoint, G, H []*privacy.EllipticPoint, ipp InnerProdArg) bool {
	s1 := blake2b.Sum256([]byte(P.X.String() + P.Y.String()))
	chal1 := new(big.Int).SetBytes(s1[:])
	ux := U.ScalarMult(chal1)
	curIt := len(ipp.Challenges) - 1

	// check all challenges
	if ipp.Challenges[curIt].Cmp(chal1) != 0 {
		return false
	}
	for j := curIt - 1; j >= 0; j-- {
		Lval := ipp.L[j]
		Rval := ipp.R[j]
		// prover sends L & C1 and gets a challenge
		s256 := blake2b.Sum256([]byte(
			Lval.X.String() + Lval.Y.String() +
				Rval.X.String() + Rval.Y.String()))
		chal2 := new(big.Int).SetBytes(s256[:])

		if ipp.Challenges[j].Cmp(chal2) != 0 {
			return false
		}
	}
	// begin computing

	curIt -= 1
	Pprime := P.Add(ux.ScalarMult(c))

	tmp1 := RangeProofParams.Zero()
	for j := curIt; j >= 0; j-- {
		x2 := new(big.Int).Exp(ipp.Challenges[j], big.NewInt(2), privacy.Curve.Params().N)
		x2i := new(big.Int).ModInverse(x2, privacy.Curve.Params().N)
		tmp1 = ipp.L[j].ScalarMult(x2).Add(ipp.R[j].ScalarMult(x2i)).Add(tmp1)
	}
	rhs := Pprime.Add(tmp1)

	sScalars:= makeBigIntArray(RangeProofParams.V)
	invsScalars := makeBigIntArray(RangeProofParams.V)

	for i := 0; i < RangeProofParams.V; i++ {
		si := big.NewInt(1)
		for j := curIt; j >= 0; j-- {
			// original challenge if the jth bit of i is 1, inverse challenge otherwise
			chal := ipp.Challenges[j]
			if big.NewInt(int64(i)).Bit(j) == 0 {
				chal = new(big.Int).ModInverse(chal, privacy.Curve.Params().N)
			}
			si = new(big.Int).Mod(new(big.Int).Mul(si, chal), privacy.Curve.Params().N)
		}
		sScalars[i] = si
		invsScalars[i] = new(big.Int).ModInverse(si, privacy.Curve.Params().N)
	}

	ccalc := new(big.Int).Mod(new(big.Int).Mul(ipp.A, ipp.B), privacy.Curve.Params().N)
	lhs := TwoVectorPCommitWithGens(G, H, ScalarVectorMul(sScalars, ipp.A), ScalarVectorMul(invsScalars, ipp.B)).Add(ux.ScalarMult(ccalc))

	if !rhs.IsEqual(lhs) {
		return false
	}
	return true
}
/*-----------------------------Vector Functions-----------------------------*/
// The length here always has to be a power of two
func InnerProduct(a []*big.Int, b []*big.Int) *big.Int {
	if len(a) != len(b) {
		privacy.NewPrivacyErr(privacy.UnexpectedErr, errors.New("InnerProduct: Uh oh! Arrays not of the same length"))
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
		privacy.NewPrivacyErr(privacy.UnexpectedErr, errors.New("VectorAddPoint: Uh oh! Arrays not of the same length"))
	}
	result := make([]*big.Int, len(v))
	for i := range v {
		result[i] = new(big.Int).Mod(new(big.Int).Add(v[i], w[i]), privacy.Curve.Params().N)
	}
	return result
}

func VectorHadamard(v, w []*big.Int) []*big.Int {
	if len(v) != len(w) {
		privacy.NewPrivacyErr(privacy.UnexpectedErr, errors.New("VectorHadamard: Uh oh! Arrays not of the same length"))
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
	C   elliptic.Curve      							 // curve
	BPG []*privacy.EllipticPoint           // slice of gen 1 for BP
	BPH []*privacy.EllipticPoint           // slice of gen 2 for BP
	N   *big.Int            							 // scalar prime
	U   *privacy.EllipticPoint             // a point that is a fixed group element with an unknown discrete-log relative to g,h
	V   int                 							 // Vector length
	G   *privacy.EllipticPoint             // G value for commitments of a single value
	H   *privacy.EllipticPoint             // H value for commitments of a single value
}

func (c CryptoParams) Zero() *privacy.EllipticPoint {
	zeroPoint:=new(privacy.EllipticPoint)
	zeroPoint.X= new(big.Int).SetInt64(0)
	zeroPoint.Y= new(big.Int).SetInt64(0)
	return zeroPoint
}
// NewECPrimeGroupKey returns the curve (field),
// Generator 1 x&y, Generator 2 x&y, order of the generators
func NewECPrimeGroupKey(n int) CryptoParams {

	gen1Vals := make([]*privacy.EllipticPoint, n)
	gen2Vals := make([]*privacy.EllipticPoint, n)
	u := CryptoParams{}.Zero()
	G:=privacy.PedCom.G[privacy.VALUE]
	H:=privacy.PedCom.G[privacy.RAND]

	for i:=0;i<n;i++{
		gen1Vals[i]= G.Hash(0)
		G=G.Hash(0)
		gen2Vals[i]= H.Hash(0)
		H = H.Hash(0)
	}
	u	= G.Add(H).Hash(0)
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
func TwoVectorPCommitWithGens(G, H []*privacy.EllipticPoint, a, b []*big.Int) *privacy.EllipticPoint {
	if len(G) != len(H) || len(G) != len(a) || len(a) != len(b) {
		return nil
	}
	commitment := CryptoParams{}.Zero()
	for i := 0; i < len(G); i++ {
		modA := new(big.Int).Mod(a[i], privacy.Curve.Params().N)
		modB := new(big.Int).Mod(b[i], privacy.Curve.Params().N)
		commitment = commitment.Add(G[i].ScalarMult(modA)).Add(H[i].ScalarMult(modB))
	}
	return commitment
}
