package zkp

import (
	"fmt"
	"github.com/minio/blake2b-simd"
	"github.com/ninjadotorg/constant/privacy-protocol"
	"math/big"
)



/*------------------------------------------------------*/
/*-------DECLARE INNER INGREDIENT FOR THE PROTOCOL------*/

type Helper interface {
	InitBasePoint() *BasePoint
}
type proofFactor privacy.EllipticPoint
type BasePoint struct {
	G privacy.EllipticPoint
	H privacy.EllipticPoint
}
type PKComProductProof struct {
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
type PKComProductWitness struct {
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
/*-----------------END OF DECLARATION-------------------*/
/*------------------------------------------------------*/

type PKComProductProtocol struct {
	Witness PKComProductWitness
	Proof   PKComProductProof
}

/*------------------------------------------------------*/
/*------IMPLEMENT INNER INGREDIENT FOR THE PROTOCOL-----*/
func (basePoint *BasePoint) InitBasePoint() {

	P:= new(privacy.EllipticPoint)
	P.X = privacy.Curve.Params().Gx
	P.Y = privacy.Curve.Params().Gy
	basePoint.G = privacy.HashGenerator(*P)
	basePoint.H = privacy.HashGenerator(basePoint.G)
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


func (pro *PKComProductProtocol) SetWitness(witness PKComProductWitness) {
	pro.Witness = witness
}
func (pro *PKComProductProtocol) SetProof(proof PKComProductProof) {
	pro.Proof = proof
}


func (pro *PKComProductProtocol) Prove() (*PKComProductProof,error) {
	proof :=  new(PKComProductProof)
	proof.basePoint.InitBasePoint();
	d := new(big.Int).SetBytes(privacy.RandBytes(32));
	e := new(big.Int).SetBytes(privacy.RandBytes(32));
	s := new(big.Int).SetBytes(privacy.RandBytes(32));
	s1 := new(big.Int).SetBytes(privacy.RandBytes(32));
	t := new(big.Int).SetBytes(privacy.RandBytes(32));
	pro.Witness.cmA = privacy.Elcm.CommitWithSpecPoint(proof.basePoint.G, proof.basePoint.H,pro.Witness.witnessA,pro.Witness.randA)
	pro.Witness.cmB = privacy.Elcm.CommitWithSpecPoint(proof.basePoint.G, proof.basePoint.H,pro.Witness.witnessB,pro.Witness.randB)
	pro.Witness.cmC = privacy.Elcm.CommitWithSpecPoint(proof.basePoint.G, proof.basePoint.H,pro.Witness.witnessAB,pro.Witness.randC)
	//Compute D factor of Proof
	D:= privacy.Elcm.CommitWithSpecPoint(proof.basePoint.G, proof.basePoint.H, d.Bytes(),s.Bytes());

	//Compute D' factor of Proof
	G1 := new(privacy.EllipticPoint)
	G1,_= privacy.DecompressCommitment(pro.Witness.cmB);
	D1:= privacy.Elcm.CommitWithSpecPoint(*G1,proof.basePoint.H, d.Bytes(),s1.Bytes());



	//Compute E factor of Proof
	E:= privacy.Elcm.CommitWithSpecPoint(proof.basePoint.G,proof.basePoint.H, e.Bytes(),t.Bytes())

	D_temp,_ :=privacy.DecompressCommitment(D)
	proof.D.X,proof.D.Y = D_temp.X, D_temp.Y
	E_temp,_ :=privacy.DecompressCommitment(E);
	proof.E.X, proof.E.Y = E_temp.X, E_temp.Y

	D1_temp,_ :=privacy.DecompressCommitment(D1);
	proof.D1.X, proof.D1.Y = D1_temp.X, D1_temp.Y
	// x = hash(G||H||D||D1||E)
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
	x:=new(big.Int)
	x.SetBytes(computeHashString(data))

	//compute f1
	a:= new(big.Int)
	a.SetBytes(pro.Witness.witnessA)
	f1:= a.Mul(a,x)

	f1 = f1.Add(f1,d)

	f1 = f1.Mod(f1,privacy.Curve.Params().N);
	proof.f1 = *f1;

	//compute z1
	ra:= new(big.Int)
	ra.SetBytes(pro.Witness.randA)
	z1:= ra.Mul(ra,x)
	z1 = z1.Add(z1,s)
	z1 = z1.Mod(z1,privacy.Curve.Params().N)
	proof.z1 = *z1;
	//compute f2
	b:= new(big.Int)
	b.SetBytes(pro.Witness.witnessB)
	f2:= b.Mul(b,x)
	f2 = f2.Add(f2,e)
	f2 = f2.Mod(f2,privacy.Curve.Params().N)
	proof.f2 = *f2;
	//compute z2 = rb*x+t mod p
	rb:= new(big.Int)
	rb.SetBytes(pro.Witness.randB)
	z2:= rb.Mul(rb,x)
	z2 = z2.Add(z2,t)
	z2 = z2.Mod(z2,privacy.Curve.Params().N)
	proof.z2 = *z2;
	//compute z3 = (rc-a*rb) + s'
	rb_new:=new(big.Int)
	a_new:= new(big.Int)
	a_new.SetBytes(pro.Witness.witnessA)
	rb_new.SetBytes(pro.Witness.randB)
	rc:= new(big.Int)
	rc.SetBytes(pro.Witness.randC)
	rc = rc.Sub(rc,a_new.Mul(a_new,rb_new))
	z3:= rc.Mul(rc,x)
	z3 = z3.Add(z3,s1)
	z3 = z3.Mod(z3,privacy.Curve.Params().N)
	proof.z3 = *z3;
	proof.cmA = pro.Witness.cmA
	proof.cmB = pro.Witness.cmB
	proof.cmC = pro.Witness.cmC
	proof.G1 = *G1
	return proof,nil;
}

func (pro *PKComProductProtocol) Verify () bool {
	pts1 := new(privacy.EllipticPoint)
	data:=[][]byte{
		pro.Proof.basePoint.G.X.Bytes(),
		pro.Proof.basePoint.G.Y.Bytes(),
		pro.Proof.basePoint.H.Y.Bytes(),
		pro.Proof.basePoint.H.Y.Bytes(),
		pro.Proof.D.X.Bytes(),
		pro.Proof.D.Y.Bytes(),
		pro.Proof.D1.X.Bytes(),
		pro.Proof.D1.Y.Bytes(),
		pro.Proof.E.X.Bytes(),
		pro.Proof.E.Y.Bytes(),
	}
	x:= computeHashString(data)
	checkFlag:=0;

	//Check witness 1: xA + D == 	Commit(f1,z1)
	A:= new(privacy.EllipticPoint)
	A,_ = privacy.DecompressCommitment(pro.Proof.cmA);
	pts1.X, pts1.Y = privacy.Curve.ScalarMult(A.X, A.Y, x)
	pts1.X, pts1.Y = privacy.Curve.Add(pts1.X, pts1.Y, pro.Proof.D.X,pro.Proof.D.Y);

	com1 := privacy.Elcm.CommitWithSpecPoint(pro.Proof.basePoint.G,pro.Proof.basePoint.H, pro.Proof.f1.Bytes(), pro.Proof.z1.Bytes())
	com1_temp,_:=privacy.DecompressCommitment(com1)





	//fmt.Printf("Com1: %+v\n",com1_temp)
	//fmt.Printf("pts1: %+v\n",pts1)

	if (com1_temp.X.Cmp(pts1.X)==0 && com1_temp.Y.Cmp(pts1.Y)==0){
		checkFlag +=1
		fmt.Println("Passed test 1")
	}
	//Check witness 2: xB + E == 	Commit(f2,z2)
	B:= new(privacy.EllipticPoint)
	B,_ = privacy.DecompressCommitment(pro.Proof.cmB);
	pts1.X, pts1.Y = privacy.Curve.ScalarMult(B.X, B.Y, x)
	pts1.X, pts1.Y = privacy.Curve.Add(pts1.X, pts1.Y, pro.Proof.E.X,pro.Proof.E.Y);
	com2 := privacy.Elcm.CommitWithSpecPoint(pro.Proof.basePoint.G,pro.Proof.basePoint.H, pro.Proof.f2.Bytes(), pro.Proof.z2.Bytes())
	com2_temp,_:=privacy.DecompressCommitment(com2)


	//fmt.Printf("Com2: %+v\n",com2_temp)
	//fmt.Printf("pts2: %+v\n",pts1)



	if (com2_temp.X.Cmp(pts1.X)==0 && com2_temp.Y.Cmp(pts1.Y)==0){
		checkFlag +=1
		fmt.Println("Passed test 2")
	}
	//  Check witness 3: xC + D1 == Commit'(f1,z3)
	C := new(privacy.EllipticPoint)
	C,_ = privacy.DecompressCommitment(pro.Proof.cmC);
	pts1.X, pts1.Y = privacy.Curve.ScalarMult(C.X, C.Y, x)
	pts1.X, pts1.Y = privacy.Curve.Add(pts1.X, pts1.Y, pro.Proof.D1.X,pro.Proof.D1.Y);
	com3 := privacy.Elcm.CommitWithSpecPoint(pro.Proof.G1,pro.Proof.basePoint.H, pro.Proof.f1.Bytes(), pro.Proof.z3.Bytes())
	com3_temp,_:=privacy.DecompressCommitment(com3)



	//fmt.Printf("Com3: %+v\n",com3_temp)
	//fmt.Printf("pts3: %+v\n",pts1)

	if (com3_temp.X.Cmp(pts1.X)==0 && com3_temp.Y.Cmp(pts1.Y)==0){
		checkFlag +=1
		fmt.Println("Passed test 3")
	}
	//println(checkFlag)
	if(checkFlag == 3) {
		fmt.Println("Passed all test. This proof is valid")
		return true;
	}
	return false;
}
func TestPKComProduct() {

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

		ipCm:= new(PKComProductWitness)
		ipCm.witnessA = witnessA
		ipCm.witnessB = witnessB
		ipCm.witnessAB = witnessC
		ipCm.randA = rA
		ipCm.randB = rB
		ipCm.randC = rC
		var zk PKComProductProtocol
		zk.SetWitness(*ipCm)
		proof,_:= zk.Prove()
		zk.SetProof(*proof)
		res = zk.Verify();
		fmt.Println(res)
	}

}
