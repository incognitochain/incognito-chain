package zkp
//
import (
	"fmt"
	"github.com/minio/blake2b-simd"
	"github.com/ninjadotorg/constant/privacy-protocol"
	"math/big"
)
/*The protocol is used for opening the commitments of product of 2 values*/
/*------------------------------------------------------*/
/*-------DECLARE INNER INGREDIENT FOR THE PROTOCOL------*/

type Helper interface {
	InitBasePoint() *BasePoint
}
type proofFactor *privacy.EllipticPoint
type BasePoint struct {
	G privacy.EllipticPoint
	H privacy.EllipticPoint
}
type PKComProductProof struct {
	basePoint BasePoint
	D proofFactor
	D1 proofFactor
	E proofFactor
	f1 *big.Int
	z1 *big.Int
	f2 *big.Int
	z2 *big.Int
	z3 *big.Int
	G1 *privacy.EllipticPoint // G1 = bG + rb*H
	cmA 		 	*privacy.EllipticPoint
	cmB 		 	*privacy.EllipticPoint
	cmC 		 	*privacy.EllipticPoint
	index      byte
}
type PKComProductWitness struct {
	witnessA  *big.Int
	randA     *big.Int
	witnessB  *big.Int
	randB     *big.Int
	witnessAB *big.Int
	randC     *big.Int
}

/*-----------------END OF DECLARATION-------------------*/
/*------------------------------------------------------*/
func (wit *PKComProductWitness) Set(
	witA *big.Int,
	randA *big.Int,
	cmA *privacy.EllipticPoint,
	witB *big.Int,
	randB *big.Int,
	cmB *privacy.EllipticPoint,
	witAB *big.Int,
	randC *big.Int,
	cmC *privacy.EllipticPoint,
	index byte){
		wit.witnessA=witA
		wit.witnessB=witB
		wit.witnessAB = witAB
		wit.randA = randA
		wit.randB = randB
		wit.randC = randC
}
func (pro *PKComProductProof)  Init(){
	pro.basePoint.InitBasePoint()
	pro.D = new(privacy.EllipticPoint)
	pro.E = new(privacy.EllipticPoint)
	pro.D1 = new(privacy.EllipticPoint)
	pro.cmA = new(privacy.EllipticPoint)
	pro.cmB = new(privacy.EllipticPoint)
	pro.cmC = new(privacy.EllipticPoint)
	pro.G1 = new(privacy.EllipticPoint)
	pro.f1 = new(big.Int)
	pro.f2 = new(big.Int)
	pro.z1 = new(big.Int)
	pro.z2 = new(big.Int)
	pro.z3 = new(big.Int)
}
func (wit *PKComProductWitness) Get() *PKComProductWitness {
	return wit
}


/*------------------------------------------------------*/
/*------IMPLEMENT INNER INGREDIENT FOR THE PROTOCOL-----*/
/*Init 2 point G and H for calculate the commitment*/
func (basePoint *BasePoint) InitBasePoint() {
	P:= new(privacy.EllipticPoint)
	P.X = privacy.Curve.Params().Gx
	P.Y = privacy.Curve.Params().Gy
	basePoint.G = P.Hash(0)
	basePoint.H = basePoint.G.Hash(0)
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

func (wit *PKComProductWitness) Prove() (*PKComProductProof,error) {

	/*---------------------------------------------------------------------------------|
	| INPUT: ck: PedersenCommitment Key                                                |
	|				A : Commiment of value a                                                   |
	|				(a,ra): value and its random                                               |
	|				B : Commiment of value b                                                   |
	|				(b,rb): value and its random														                   |
	|				C : Commiment of value a*b														                     |
	|				(ab,rc): product of 2 values and its random														     |
	| OUTPUT: The proof for proving the statement: 														         |
	|         "A,B and C is the commitment of a,b and a*b"														 |
	|---------------------------------------------------------------------------------*/
	/*--------------This Prove() function work under the following scheme--------------|
	|	Choose random d, e, s, s', t in Zp																							 |
	|	Let ck = (G, p, q, H)																														 |
	|	Let ck' = (G, p, B, H), which B = b*G																						 |
	|	Compute D = Com(d,s) under ck																										 |
	|	Compute D' = Com(d,s') under ck'																								 |
	|	Compute E = Com(e,t) under ck																		 								 |
	|	Send D, D, E to the verifier (included in the Proof)														 |
	|	Compute x = hash(G||H||D||D1||E) then:																					 |
	|		Compute f1 = a*x+d mod p																										   |
	|		Compute z1 = ra*x +s mod p																										 |
	|		Compute f2 = bx+e mod p																										     |
	|		Compute z2 = rb*x+t mod p																										   |
	|		Compute z3 = (rc - a*rb) + s' mod p																						 |
	|	Send f1, f2, z1, z2, z3 to the verifier (included in the Proof)                  |
	|---------------------------------------------------------------------------------*/
	/* ---------------------------------------------------------------------------------------------------------------|
 |		THE LINK OF ORIGINAL PAPER FOR THIS PROTOCOL: https://link.springer.com/chapter/10.1007%2F978-3-319-43005-8_1 |
 |   WE CHANGE FROM INTERACTIVE TO NON-INTERACTIVE SCHEME VIA FIAT-SHAMIR HEURISTIC														      |
 |----------------------------------------------------------------------------------------------------------------*/
	proof :=  new(PKComProductProof)
	proof.Init()
	SpecCom1:=privacy.PCParams{[]privacy.EllipticPoint{proof.basePoint.G,
																											  proof.basePoint.H},
																											  2}
	d := new(big.Int).SetBytes(privacy.RandBytes(32));
	e := new(big.Int).SetBytes(privacy.RandBytes(32));
	s := new(big.Int).SetBytes(privacy.RandBytes(32));
	s1 := new(big.Int).SetBytes(privacy.RandBytes(32));
	t := new(big.Int).SetBytes(privacy.RandBytes(32));

	proof.cmA = SpecCom1.CommitAtIndex(wit.witnessA, wit.randA,0)
	proof.cmB = SpecCom1.CommitAtIndex(wit.witnessB, wit.randB,0)
	proof.cmC = SpecCom1.CommitAtIndex(wit.witnessAB,wit.randC,0)
	//Compute D factor of Proof
	D:= SpecCom1.CommitAtIndex(d,s,0);
	//Compute D' factor of Proof
	SpecCom2:=privacy.PCParams{[]privacy.EllipticPoint{*proof.cmB, proof.basePoint.H},
		                         2}
	D1:= SpecCom2.CommitAtIndex(d,s1,0);
	//Compute E factor of Proof
	E:= SpecCom1.CommitAtIndex(e,t,0)
	*proof.D = *D
	*proof.E = *E
	*proof.D1 = *D1
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
	a.Set(wit.witnessA)
	f1:= a.Mul(a,x)
	f1 = f1.Add(f1,d)
	f1 = f1.Mod(f1,privacy.Curve.Params().N);
	*proof.f1 = *f1;

	//compute z1
	ra:= new(big.Int)
	ra.Set(wit.randA)
	z1:= ra.Mul(ra,x)
	z1 = z1.Add(z1,s)
	z1 = z1.Mod(z1,privacy.Curve.Params().N)
	*proof.z1 = *z1;
	//compute f2
	b:= new(big.Int)
	b.Set(wit.witnessB)
	f2:= b.Mul(b,x)
	f2 = f2.Add(f2,e)
	f2 = f2.Mod(f2,privacy.Curve.Params().N)
	*proof.f2 = *f2;
	//compute z2 = rb*x+t mod p
	rb:= new(big.Int)
	rb.Set(wit.randB)
	z2:= rb.Mul(rb,x)
	z2 = z2.Add(z2,t)
	z2 = z2.Mod(z2,privacy.Curve.Params().N)
	*proof.z2 = *z2;
	//compute z3 = (rc-a*rb) + s'
	rb_new:=new(big.Int)
	a_new:= new(big.Int)
	a_new.Set(wit.witnessA)
	rb_new.Set(wit.randB)
	rc:= new(big.Int)
	rc.Set(wit.randC)
	rc = rc.Sub(rc,a_new.Mul(a_new,rb_new))
	z3:= rc.Mul(rc,x)
	z3 = z3.Add(z3,s1)
	z3 = z3.Mod(z3,privacy.Curve.Params().N)
	*proof.z3 = *z3;
	*proof.G1 = *proof.cmB
	proof.cmA = SpecCom1.CommitAtIndex(wit.witnessA, wit.randA,0)
	proof.cmB = SpecCom1.CommitAtIndex(wit.witnessB, wit.randB,0)
	proof.cmC = SpecCom1.CommitAtIndex(wit.witnessAB,wit.randC,0)
	
	return proof,nil;
}

func (pro *PKComProductProof) Verify () bool {

	/*------------------------------------------------------|
	|	INPUT: The Proof received from the prover							|
	| OUTPUT: True if the proof is valid, False otherwise		|
	|------------------------------------------------------*/
	/*--------------This Verify() function work under the following scheme--------------|
	|	Check if D, D', E is points on Curve or not							 												  |
	|	Check if f1, f2, z1, z2, z3 in Zp or not																				  |
	|	Follow 3 test:																					                          |
	|		Check if Com(f1,z1) under ck equals to x*A + D or not														|
	|	  Check if Com(f2,z2) under ck equals to x*B + E or not														|
	|   Check if Com(f1,z3) under ck' equals to x*C + D' or not												  |
	|----------------------------------------------------------------------------------*/

	var pts_cmp privacy.EllipticPoint
	data:=[][]byte{
		pro.basePoint.G.X.Bytes(),
		pro.basePoint.G.Y.Bytes(),
		pro.basePoint.H.Y.Bytes(),
		pro.basePoint.H.Y.Bytes(),
		pro.D.X.Bytes(),
		pro.D.Y.Bytes(),
		pro.D1.X.Bytes(),
		pro.D1.Y.Bytes(),
		pro.E.X.Bytes(),
		pro.E.Y.Bytes(),
	}
	x:= new(big.Int).SetBytes(computeHashString(data))
	//Check if D,D',E is on Curve
	if !(privacy.Curve.IsOnCurve(pro.D.X, pro.D.Y) &&
		privacy.Curve.IsOnCurve(pro.D1.X, pro.D1.Y) &&
		privacy.Curve.IsOnCurve(pro.E.X, pro.E.Y)){
		return false;
	}
	//Check if f1,f2,z1,z2,z3 in Zp
	if (pro.f1.Cmp(privacy.Curve.Params().P)==1 ||
		pro.f2.Cmp(privacy.Curve.Params().P)==1 ||
		pro.z1.Cmp(privacy.Curve.Params().P)==1 ||
		pro.z2.Cmp(privacy.Curve.Params().P)==1 ||
		pro.z3.Cmp(privacy.Curve.Params().P)==1){
		return false;
	}
	//Check witness 1: xA + D == 	CommitAll(f1,z1)
	A:= new(privacy.EllipticPoint)
	A = pro.cmA;
	pts_cmp = A.ScalarMulPoint(x).AddPoint(*pro.D)
	SpecCom1:=privacy.PCParams{[]privacy.EllipticPoint{pro.basePoint.G, pro.basePoint.H},
														 2}
	com1 := SpecCom1.CommitAtIndex(pro.f1, pro.z1,0)
	if (com1.IsEqual(pts_cmp)){
		fmt.Println("Passed test 1")
	} else {
		fmt.Println("Failed test 1")
		return false
	}
	//Check witness 2: xB + E == 	CommitAll(f2,z2)
	pts_cmp = pro.cmB.ScalarMulPoint(x).AddPoint(*pro.E)
	com2 := SpecCom1.CommitAtIndex(pro.f2, pro.z2,0)
	if (com2.IsEqual(pts_cmp)){
		fmt.Println("Passed test 2")
	}	else {
		fmt.Println("Failed test 2")
		return false
	}
	//  Check witness 3: xC + D1 == CommitAll'(f1,z3)
	SpecCom2:=privacy.PCParams{[]privacy.EllipticPoint{*pro.G1, pro.basePoint.H},
		2}
	pts_cmp = pro.cmC.ScalarMulPoint(x).AddPoint(*pro.D1)
	com3 := SpecCom2.CommitAtIndex(pro.f1, pro.z3,0)
	if (com3.IsEqual(pts_cmp)){
		fmt.Println("Passed test 3")
		fmt.Println("Passed all test. This proof is valid.")
	}	else {
		fmt.Println("Failed test 3")
		return false
	}
	return true;
}
//func TestPKComProduct() {
//	res := true
//	for i:=0;i<100;i++ {
//		witnessA := privacy.RandBytes(32)
//		witnessB := privacy.RandBytes(32)
//		C:= new(big.Int)
//		C.SetBytes(witnessA)
//		C.Mul(C, new(big.Int).SetBytes(witnessB))
//		witnessC := C.Bytes()
//		rA := privacy.RandBytes(32)
//		rB := privacy.RandBytes(32)
//		rC := privacy.RandBytes(32)
//		r1Int := new(big.Int).SetBytes(rA)
//		r2Int := new(big.Int).SetBytes(rB)
//		r3Int := new(big.Int).SetBytes(rC)
//		r1Int.Mod(r1Int, privacy.Curve.Params().N)
//		r2Int.Mod(r2Int, privacy.Curve.Params().N)
//		r3Int.Mod(r3Int, privacy.Curve.Params().N)
//
//		rA = r1Int.Bytes()
//		rB = r2Int.Bytes()
//		rC = r3Int.Bytes()
//		ipCm:= new(PKComProductWitness)
//		ipCm.witnessA = new(big.Int).SetBytes(witnessA)
//		ipCm.witnessB = new(big.Int).SetBytes(witnessB)
//		ipCm.witnessAB = new(big.Int).SetBytes(witnessC)
//		ipCm.randA = new(big.Int).SetBytes(rA)
//		ipCm.randB = new(big.Int).SetBytes(rB)
//		ipCm.randC = new(big.Int).SetBytes(rC)
//
//		proof,_:= ipCm.Prove()
//		res = proof.Verify();
//		fmt.Printf("Test %d is %t\n",i,res)
//	}
//}
