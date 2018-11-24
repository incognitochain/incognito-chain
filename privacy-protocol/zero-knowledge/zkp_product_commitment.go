package zkp
//
//import (
//	"fmt"
//	"github.com/minio/blake2b-simd"
//	"github.com/ninjadotorg/constant/privacy-protocol"
//	"math/big"
//)
///*The protocol is used for opening the commitments of product of 2 values*/
///*------------------------------------------------------*/
///*-------DECLARE INNER INGREDIENT FOR THE PROTOCOL------*/
//
//type Helper interface {
//	InitBasePoint() *BasePoint
//}
//type proofFactor privacy.EllipticPoint
//type BasePoint struct {
//	G privacy.EllipticPoint
//	H privacy.EllipticPoint
//}
//type PKComProductProof struct {
//	basePoint BasePoint
//	D proofFactor
//	D1 proofFactor
//	E proofFactor
//	f1 big.Int
//	z1 big.Int
//	f2 big.Int
//	z2 big.Int
//	z3 big.Int
//	G1 privacy.EllipticPoint // G1 = bG + rb*H
//	cmA 		 	[]byte
//	cmB 		 	[]byte
//	cmC 		 	[]byte
//}
//type PKComProductWitness struct {
//	witnessA  []byte
//	cmA       []byte
//	randA     []byte
//	witnessB  []byte
//	cmB       []byte
//	randB     []byte
//	witnessAB []byte
//	cmC       []byte
//	randC     []byte
//}
///*-----------------END OF DECLARATION-------------------*/
///*------------------------------------------------------*/
//
//type PKComProductProtocol struct {
//	Witness PKComProductWitness
//	Proof   PKComProductProof
//}
//
///*------------------------------------------------------*/
///*------IMPLEMENT INNER INGREDIENT FOR THE PROTOCOL-----*/
///*Init 2 point G and H for calculate the commitment*/
//func (basePoint *BasePoint) InitBasePoint() {
//	P:= new(privacy.EllipticPoint)
//	P.X = privacy.Curve.Params().Gx
//	P.Y = privacy.Curve.Params().Gy
//	basePoint.G = P.HashPoint()
//	basePoint.H = basePoint.G.HashPoint()
//}
//func computeHashString(data [][]byte) []byte{
//	str:=make([]byte, 0)
//	for i:=0;i<len(data);i++{
//		str = append(str,data[i]...)
//	}
//	hashFunc := blake2b.New256()
//	hashFunc.Write(str)
//	hashValue := hashFunc.Sum(nil)
//	return hashValue
//}
//func (pro *PKComProductProtocol) SetWitness(witness PKComProductWitness) {
//	pro.Witness = witness
//}
//func (pro *PKComProductProtocol) SetProof(proof PKComProductProof) {
//	pro.Proof = proof
//}
//func (pro *PKComProductProtocol) Prove() (*PKComProductProof,error) {
//
//	/*---------------------------------------------------------------------------------|
//	| INPUT: ck: PedersenCommitment Key                                                         |
//	|				A : Commiment of value a                                                   |
//	|				(a,ra): value and its random                                               |
//	|				B : Commiment of value b                                                   |
//	|				(b,rb): value and its random														                   |
//	|				C : Commiment of value a*b														                     |
//	|				(ab,rc): product of 2 values and its random														     |
//	| OUTPUT: The proof for proving the statement: 														         |
//	|         "A,B and C is the commitment of a,b and a*b"														 |
//	|---------------------------------------------------------------------------------*/
//	/*--------------This Prove() function work under the following scheme--------------|
//	|	Choose random d, e, s, s', t in Zp																							 |
//	|	Let ck = (G, p, q, H)																														 |
//	|	Let ck' = (G, p, B, H), which B = b*G																						 |
//	|	Compute D = Com(d,s) under ck																										 |
//	|	Compute D' = Com(d,s') under ck'																								 |
//	|	Compute E = Com(e,t) under ck																		 								 |
//	|	Send D, D, E to the verifier (included in the Proof)														 |
//	|	Compute x = hash(G||H||D||D1||E) then:																					 |
//	|		Compute f1 = a*x+d mod p																										   |
//	|		Compute z1 = ra*x +s mod p																										 |
//	|		Compute f2 = bx+e mod p																										     |
//	|		Compute z2 = rb*x+t mod p																										   |
//	|		Compute z3 = (rc - a*rb) + s' mod p																						 |
//	|	Send f1, f2, z1, z2, z3 to the verifier (included in the Proof)                  |
//	|---------------------------------------------------------------------------------*/
//	/* ---------------------------------------------------------------------------------------------------------------|
//  |		THE LINK OF ORIGINAL PAPER FOR THIS PROTOCOL: https://link.springer.com/chapter/10.1007%2F978-3-319-43005-8_1 |
//  |   WE CHANGE FROM INTERACTIVE TO NON-INTERACTIVE SCHEME VIA FIAT-SHAMIR HEURISTIC														    |
//  |----------------------------------------------------------------------------------------------------------------*/
//	proof :=  new(PKComProductProof)
//	proof.basePoint.InitBasePoint();
//	d := new(big.Int).SetBytes(privacy.RandBytes(32));
//	e := new(big.Int).SetBytes(privacy.RandBytes(32));
//	s := new(big.Int).SetBytes(privacy.RandBytes(32));
//	s1 := new(big.Int).SetBytes(privacy.RandBytes(32));
//	t := new(big.Int).SetBytes(privacy.RandBytes(32));
//	pro.Witness.cmA = privacy.PedCom.CommitWithSpecPoint(proof.basePoint.G, proof.basePoint.H,pro.Witness.witnessA,pro.Witness.randA)
//	pro.Witness.cmB = privacy.PedCom.CommitWithSpecPoint(proof.basePoint.G, proof.basePoint.H,pro.Witness.witnessB,pro.Witness.randB)
//	pro.Witness.cmC = privacy.PedCom.CommitWithSpecPoint(proof.basePoint.G, proof.basePoint.H,pro.Witness.witnessAB,pro.Witness.randC)
//	//Compute D factor of Proof
//	D:= privacy.PedCom.CommitWithSpecPoint(proof.basePoint.G, proof.basePoint.H, d.Bytes(),s.Bytes());
//	//Compute D' factor of Proof
//	G1 := new(privacy.EllipticPoint)
//	G1,_= privacy.DecompressCommitment(pro.Witness.cmB);
//	D1:= privacy.PedCom.CommitWithSpecPoint(*G1,proof.basePoint.H, d.Bytes(),s1.Bytes());
//	//Compute E factor of Proof
//	E:= privacy.PedCom.CommitWithSpecPoint(proof.basePoint.G,proof.basePoint.H, e.Bytes(),t.Bytes())
//	D_temp,_ :=privacy.DecompressCommitment(D)
//	proof.D.X,proof.D.Y = D_temp.X, D_temp.Y
//	E_temp,_ :=privacy.DecompressCommitment(E);
//	proof.E.X, proof.E.Y = E_temp.X, E_temp.Y
//
//	D1_temp,_ :=privacy.DecompressCommitment(D1);
//	proof.D1.X, proof.D1.Y = D1_temp.X, D1_temp.Y
//	// x = hash(G||H||D||D1||E)
//	data:=[][]byte{
//		proof.basePoint.G.X.Bytes(),
//		proof.basePoint.G.Y.Bytes(),
//		proof.basePoint.H.Y.Bytes(),
//		proof.basePoint.H.Y.Bytes(),
//		proof.D.X.Bytes(),
//		proof.D.Y.Bytes(),
//		proof.D1.X.Bytes(),
//		proof.D1.Y.Bytes(),
//		proof.E.X.Bytes(),
//		proof.E.Y.Bytes(),
//	}
//	x:=new(big.Int)
//	x.SetBytes(computeHashString(data))
//
//	//compute f1
//	a:= new(big.Int)
//	a.SetBytes(pro.Witness.witnessA)
//	f1:= a.Mul(a,x)
//	f1 = f1.Add(f1,d)
//	f1 = f1.Mod(f1,privacy.Curve.Params().N);
//	proof.f1 = *f1;
//
//	//compute z1
//	ra:= new(big.Int)
//	ra.SetBytes(pro.Witness.randA)
//	z1:= ra.Mul(ra,x)
//	z1 = z1.Add(z1,s)
//	z1 = z1.Mod(z1,privacy.Curve.Params().N)
//	proof.z1 = *z1;
//	//compute f2
//	b:= new(big.Int)
//	b.SetBytes(pro.Witness.witnessB)
//	f2:= b.Mul(b,x)
//	f2 = f2.Add(f2,e)
//	f2 = f2.Mod(f2,privacy.Curve.Params().N)
//	proof.f2 = *f2;
//	//compute z2 = rb*x+t mod p
//	rb:= new(big.Int)
//	rb.SetBytes(pro.Witness.randB)
//	z2:= rb.Mul(rb,x)
//	z2 = z2.Add(z2,t)
//	z2 = z2.Mod(z2,privacy.Curve.Params().N)
//	proof.z2 = *z2;
//	//compute z3 = (rc-a*rb) + s'
//	rb_new:=new(big.Int)
//	a_new:= new(big.Int)
//	a_new.SetBytes(pro.Witness.witnessA)
//	rb_new.SetBytes(pro.Witness.randB)
//	rc:= new(big.Int)
//	rc.SetBytes(pro.Witness.randC)
//	rc = rc.Sub(rc,a_new.Mul(a_new,rb_new))
//	z3:= rc.Mul(rc,x)
//	z3 = z3.Add(z3,s1)
//	z3 = z3.Mod(z3,privacy.Curve.Params().N)
//	proof.z3 = *z3;
//	proof.cmA = pro.Witness.cmA
//	proof.cmB = pro.Witness.cmB
//	proof.cmC = pro.Witness.cmC
//	proof.G1 = *G1
//	return proof,nil;
//}
//
//func (pro *PKComProductProtocol) Verify () bool {
//
//	/*------------------------------------------------------|
//	|	INPUT: The Proof received from the prover							|
//	| OUTPUT: True if the proof is valid, False otherwise		|
//	|------------------------------------------------------*/
//	/*--------------This Verify() function work under the following scheme--------------|
//	|	Check if D, D', E is points on Curve or not							 												  |
//	|	Check if f1, f2, z1, z2, z3 in Zp or not																				  |
//	|	Follow 3 test:																					                          |
//	|		Check if Com(f1,z1) under ck equals to x*A + D or not														|
//	|	  Check if Com(f2,z2) under ck equals to x*B + E or not														|
//	|   Check if Com(f1,z3) under ck' equals to x*C + D' or not												  |
//	|----------------------------------------------------------------------------------*/
//	pts1 := new(privacy.EllipticPoint)
//	data:=[][]byte{
//		pro.Proof.basePoint.G.X.Bytes(),
//		pro.Proof.basePoint.G.Y.Bytes(),
//		pro.Proof.basePoint.H.Y.Bytes(),
//		pro.Proof.basePoint.H.Y.Bytes(),
//		pro.Proof.D.X.Bytes(),
//		pro.Proof.D.Y.Bytes(),
//		pro.Proof.D1.X.Bytes(),
//		pro.Proof.D1.Y.Bytes(),
//		pro.Proof.E.X.Bytes(),
//		pro.Proof.E.Y.Bytes(),
//	}
//	x:= computeHashString(data)
//	checkFlag:=0;
//
//	//Check if D,D',E is on Curve
//	if !(privacy.Curve.IsOnCurve(pro.Proof.D.X, pro.Proof.D.Y) &&
//		privacy.Curve.IsOnCurve(pro.Proof.D1.X, pro.Proof.D1.Y) &&
//		privacy.Curve.IsOnCurve(pro.Proof.E.X, pro.Proof.E.Y)){
//		return false;
//	}
//	//Check if f1,f2,z1,z2,z3 in Zp
//	if (pro.Proof.f1.Cmp(privacy.Curve.Params().P)==1 ||
//		pro.Proof.f2.Cmp(privacy.Curve.Params().P)==1 ||
//		pro.Proof.z1.Cmp(privacy.Curve.Params().P)==1 ||
//		pro.Proof.z2.Cmp(privacy.Curve.Params().P)==1 ||
//		pro.Proof.z3.Cmp(privacy.Curve.Params().P)==1){
//		return false;
//	}
//	//Check witness 1: xA + D == 	CommitAll(f1,z1)
//	A:= new(privacy.EllipticPoint)
//	A,_ = privacy.DecompressCommitment(pro.Proof.cmA);
//	pts1.X, pts1.Y = privacy.Curve.ScalarMult(A.X, A.Y, x)
//	pts1.X, pts1.Y = privacy.Curve.Add(pts1.X, pts1.Y, pro.Proof.D.X,pro.Proof.D.Y);
//
//	com1 := privacy.PedCom.CommitWithSpecPoint(pro.Proof.basePoint.G,pro.Proof.basePoint.H, pro.Proof.f1.Bytes(), pro.Proof.z1.Bytes())
//	com1_temp,_:=privacy.DecompressCommitment(com1)
//
//	if (com1_temp.X.Cmp(pts1.X)==0 && com1_temp.Y.Cmp(pts1.Y)==0){
//		checkFlag +=1
//		fmt.Println("Passed test 1")
//	}
//	//Check witness 2: xB + E == 	CommitAll(f2,z2)
//	B:= new(privacy.EllipticPoint)
//	B,_ = privacy.DecompressCommitment(pro.Proof.cmB);
//	pts1.X, pts1.Y = privacy.Curve.ScalarMult(B.X, B.Y, x)
//	pts1.X, pts1.Y = privacy.Curve.Add(pts1.X, pts1.Y, pro.Proof.E.X,pro.Proof.E.Y);
//	com2 := privacy.PedCom.CommitWithSpecPoint(pro.Proof.basePoint.G,pro.Proof.basePoint.H, pro.Proof.f2.Bytes(), pro.Proof.z2.Bytes())
//	com2_temp,_:=privacy.DecompressCommitment(com2)
//
//	if (com2_temp.X.Cmp(pts1.X)==0 && com2_temp.Y.Cmp(pts1.Y)==0){
//		checkFlag +=1
//		fmt.Println("Passed test 2")
//	}
//	//  Check witness 3: xC + D1 == CommitAll'(f1,z3)
//	C := new(privacy.EllipticPoint)
//	C,_ = privacy.DecompressCommitment(pro.Proof.cmC);
//	pts1.X, pts1.Y = privacy.Curve.ScalarMult(C.X, C.Y, x)
//	pts1.X, pts1.Y = privacy.Curve.Add(pts1.X, pts1.Y, pro.Proof.D1.X,pro.Proof.D1.Y);
//	com3 := privacy.PedCom.CommitWithSpecPoint(pro.Proof.G1,pro.Proof.basePoint.H, pro.Proof.f1.Bytes(), pro.Proof.z3.Bytes())
//	com3_temp,_:=privacy.DecompressCommitment(com3)
//	if (com3_temp.X.Cmp(pts1.X)==0 && com3_temp.Y.Cmp(pts1.Y)==0){
//		checkFlag +=1
//		fmt.Println("Passed test 3")
//	}
//	//println(checkFlag)
//	if(checkFlag == 3) {
//		fmt.Println("Passed all test. This proof is valid.")
//		return true;
//	}
//	return false;
//}
//func TestPKComProduct() {
//
//	res := true
//	for res{
//		witnessA := privacy.RandBytes(32)
//		witnessB := privacy.RandBytes(32)
//		C:= new(big.Int)
//		C.SetBytes(witnessA)
//		C.Mul(C, new(big.Int).SetBytes(witnessB))
//		witnessC := C.Bytes()
//
//
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
//		ipCm.witnessA = witnessA
//		ipCm.witnessB = witnessB
//		ipCm.witnessAB = witnessC
//		ipCm.randA = rA
//		ipCm.randB = rB
//		ipCm.randC = rC
//		var zk PKComProductProtocol
//		zk.SetWitness(*ipCm)
//		proof,_:= zk.Prove()
//		zk.SetProof(*proof)
//		zk.Verify();
//	}
//}
