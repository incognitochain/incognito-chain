package zkp

import (
	"fmt"
	"github.com/minio/blake2b-simd"
	"math/big"

	"github.com/ninjadotorg/constant/privacy-protocol"

)
type Helper interface {
	InitBasePoint() *BasePoint
}
type proofFactor privacy.EllipticPoint
type BasePoint struct {
	G privacy.EllipticPoint
	H privacy.EllipticPoint
}
type ProofOfProductCommitment struct {
	basePoint BasePoint
	D proofFactor
	D1 proofFactor
	E proofFactor
	f1 big.Int
	z1 big.Int
	f2 big.Int
	z2 big.Int
	z3 big.Int
	G1 privacy.EllipticPoint // G1 = bG + rb*H
	cmA 		 	[]byte
	cmB 		 	[]byte
	cmC 		 	[]byte
}
type InputCommitments struct {
	witnessA  []byte
	cmA       []byte
	randA     []byte
	witnessB  []byte
	cmB       []byte
	randB     []byte
	witnessAB []byte
	cmC       []byte
	randC     []byte
}
func (basePoint *BasePoint) InitBasePoint() {

	P:= new(privacy.EllipticPoint)
	P.X = privacy.Curve.Params().Gx
	P.Y = privacy.Curve.Params().Gy
	basePoint.G = privacy.HashGenerator(*P)
	basePoint.H = privacy.HashGenerator(basePoint.G)
}
// Random number modular N

func computeCommitmentPoint(pointG privacy.EllipticPoint, pointH privacy.EllipticPoint, val1 *big.Int, val2 *big.Int) proofFactor{
	factor:= new(proofFactor)
	factor.X, factor.Y= privacy.Curve.ScalarMult(pointG.X, pointG.Y, val1.Bytes())
	tmp:= new(proofFactor)
	tmp.X, tmp.Y = privacy.Curve.ScalarMult(pointH.X, pointH.Y, val2.Bytes())
	factor.X,factor.Y = privacy.Curve.Add(factor.X, factor.Y, tmp.X, tmp.Y)
	return *factor;
}
func computeHashString(data [][]byte) []byte{
	str:=make([]byte, 0)
	for i:=0;i<len(data);i++{
		str = append(str,data[i]...)
	}
	hashFunc := blake2b.New256()
	hashFunc.Write(str)
	hashValue := hashFunc.Sum(nil)
	return hashValue
}
func MultiScalarMul(factors  [] *big.Int, point privacy.EllipticPoint) *privacy.EllipticPoint{
	a:=new(big.Int)
	a.SetInt64(1)
	for i:=0;i<len(factors);i++{
			a.Mul(a,factors[i])
	}
	P:=new(privacy.EllipticPoint)
	P.X, P.Y = privacy.Curve.ScalarMult(point.X, point.Y,a.Bytes());
	return P
}
func ProveProductCommitment(ipCm InputCommitments)  ProofOfProductCommitment{
	proof :=  new(ProofOfProductCommitment)
	proof.basePoint.InitBasePoint();
	d := new(big.Int).SetBytes(privacy.RandBytes(32));
	e := new(big.Int).SetBytes(privacy.RandBytes(32));
	s := new(big.Int).SetBytes(privacy.RandBytes(32));
	s1 := new(big.Int).SetBytes(privacy.RandBytes(32));
	t := new(big.Int).SetBytes(privacy.RandBytes(32));
	ipCm.cmA = privacy.Pcm.CommitWithSpecPoint(proof.basePoint.G, proof.basePoint.H,ipCm.witnessA,ipCm.randA)
	ipCm.cmB = privacy.Pcm.CommitWithSpecPoint(proof.basePoint.G, proof.basePoint.H,ipCm.witnessB,ipCm.randB)
	ipCm.cmC = privacy.Pcm.CommitWithSpecPoint(proof.basePoint.G, proof.basePoint.H,ipCm.witnessAB,ipCm.randC)
	//Compute D factor of Proof
	D:= computeCommitmentPoint(proof.basePoint.G, proof.basePoint.H, d,s);

	//Compute D' factor of Proof
	G1 := new(privacy.EllipticPoint)
	G1,_= privacy.DecompressCommitment(ipCm.cmB);
	D1:= computeCommitmentPoint(*G1,proof.basePoint.H, d,s1);



	//Compute E factor of Proof
	E:= computeCommitmentPoint(proof.basePoint.G,proof.basePoint.H, e,t)
	proof.D = D;
	proof.E = E;
 	proof.D1 = D1;
	// x = hash(G||H||D||D1||E)
	data:=[][]byte{
		proof.basePoint.G.X.Bytes(),
		proof.basePoint.G.Y.Bytes(),
		proof.basePoint.H.Y.Bytes(),
		proof.basePoint.H.Y.Bytes(),
		D.X.Bytes(),
		D.Y.Bytes(),
		D1.X.Bytes(),
		D1.Y.Bytes(),
		E.X.Bytes(),
		E.Y.Bytes(),
	}
	x:=new(big.Int)
	x.SetBytes(computeHashString(data))

	//compute f1
	a:= new(big.Int)
	a.SetBytes(ipCm.witnessA)
	f1:= a.Mul(a,x)

	f1 = f1.Add(f1,d)

	f1 = f1.Mod(f1,privacy.Curve.Params().N);
	proof.f1 = *f1;

	//compute z1
	ra:= new(big.Int)
	ra.SetBytes(ipCm.randA)
	z1:= ra.Mul(ra,x)
	z1 = z1.Add(z1,s)
	z1 = z1.Mod(z1,privacy.Curve.Params().N)
	proof.z1 = *z1;
	//compute f2
	b:= new(big.Int)
	b.SetBytes(ipCm.witnessB)
	f2:= b.Mul(b,x)
	f2 = f2.Add(f2,e)
	f2 = f2.Mod(f2,privacy.Curve.Params().N)
	proof.f2 = *f2;
	//compute z2 = rb*x+t mod p
	rb:= new(big.Int)
	rb.SetBytes(ipCm.randB)
	z2:= rb.Mul(rb,x)
	z2 = z2.Add(z2,t)
	z2 = z2.Mod(z2,privacy.Curve.Params().N)
	proof.z2 = *z2;
	//compute z3 = (rc-a*rb) + s'
	rb_new:=new(big.Int)
	a_new:= new(big.Int)
	a_new.SetBytes(ipCm.witnessA)
	rb_new.SetBytes(ipCm.randB)
	rc:= new(big.Int)
	rc.SetBytes(ipCm.randC)
	rc = rc.Sub(rc,a_new.Mul(a_new,rb_new))
	z3:= rc.Mul(rc,x)
	z3 = z3.Add(z3,s1)
	z3 = z3.Mod(z3,privacy.Curve.Params().N)
	proof.z3 = *z3;
	proof.cmA = ipCm.cmA
	proof.cmB = ipCm.cmB
	proof.cmC = ipCm.cmC
	proof.G1 = *G1
	return *proof;
}

func VerifyProductCommitment (proof ProofOfProductCommitment) bool {
	pts1 := new(privacy.EllipticPoint)
	data:=[][]byte{
		proof.basePoint.G.X.Bytes(),
		proof.basePoint.G.Y.Bytes(),
		proof.basePoint.H.Y.Bytes(),
		proof.basePoint.H.Y.Bytes(),
		proof.D.X.Bytes(),
		proof.D.Y.Bytes(),
		proof.D1.X.Bytes(),
		proof.D1.Y.Bytes(),
		proof.E.X.Bytes(),
		proof.E.Y.Bytes(),
	}
	x:= computeHashString(data)
	checkFlag:=0;

	//Check witness 1: xA + D == 	Commit(f1,z1)
	A:= new(privacy.EllipticPoint)
	A,_ = privacy.DecompressCommitment(proof.cmA);
	pts1.X, pts1.Y = privacy.Curve.ScalarMult(A.X, A.Y, x)
	pts1.X, pts1.Y = privacy.Curve.Add(pts1.X, pts1.Y, proof.D.X,proof.D.Y);

	com1 := computeCommitmentPoint(proof.basePoint.G,proof.basePoint.H, &proof.f1, &proof.z1)

	if (com1.X.Cmp(pts1.X)==0 && com1.Y.Cmp(pts1.Y)==0){
		checkFlag +=1
		println("Passed test 1")
		}
	//Check witness 2: xB + E == 	Commit(f2,z2)
	B:= new(privacy.EllipticPoint)
	B,_ = privacy.DecompressCommitment(proof.cmB);
	pts1.X, pts1.Y = privacy.Curve.ScalarMult(B.X, B.Y, x)
	pts1.X, pts1.Y = privacy.Curve.Add(pts1.X, pts1.Y, proof.E.X,proof.E.Y);
	com2 := computeCommitmentPoint(proof.basePoint.G,proof.basePoint.H, &proof.f2, &proof.z2)

	if (com2.X.Cmp(pts1.X)==0 && com2.Y.Cmp(pts1.Y)==0){
		checkFlag +=1
		println("Passed test 2")
	}
	//  Check witness 3: xC + D1 == Commit'(f1,z3)
	C := new(privacy.EllipticPoint)
	C,_ = privacy.DecompressCommitment(proof.cmC);
	pts1.X, pts1.Y = privacy.Curve.ScalarMult(C.X, C.Y, x)
	pts1.X, pts1.Y = privacy.Curve.Add(pts1.X, pts1.Y, proof.D1.X,proof.D1.Y);
	com3 := computeCommitmentPoint(proof.G1,proof.basePoint.H, &proof.f1, &proof.z3)
	//fmt.Println(pts1)
	//fmt.Println(com3)
	if (com3.X.Cmp(pts1.X)==0 && com3.Y.Cmp(pts1.Y)==0){
		checkFlag +=1
		println("Passed test 3")
	}
	println(checkFlag)
	if(checkFlag == 3) {
		return true;
	}
	return false;
}
func TestProductCommitment() {

	res := true
	for res{
		witnessA := privacy.RandBytes(32)
		witnessB := privacy.RandBytes(32)
		C:= new(big.Int)
		C.SetBytes(witnessA)
		C.Mul(C, new(big.Int).SetBytes(witnessB))
		witnessC := C.Bytes()


		rA := privacy.RandBytes(32)
		rB := privacy.RandBytes(32)
		rC := privacy.RandBytes(32)
		r1Int := new(big.Int).SetBytes(rA)
		r2Int := new(big.Int).SetBytes(rB)
		r3Int := new(big.Int).SetBytes(rC)
		r1Int.Mod(r1Int, privacy.Curve.Params().N)
		r2Int.Mod(r2Int, privacy.Curve.Params().N)
		r3Int.Mod(r3Int, privacy.Curve.Params().N)

		rA = r1Int.Bytes()
		rB = r2Int.Bytes()
		rC = r3Int.Bytes()

		ipCm:= new(InputCommitments)
		ipCm.witnessA = witnessA
		ipCm.witnessB = witnessB
		ipCm.witnessAB = witnessC
		ipCm.randA = rA
		ipCm.randB = rB
		ipCm.randC = rC
		proof:= ProveProductCommitment(*ipCm)
		res = VerifyProductCommitment(proof)
		fmt.Println(res)
	}

}
