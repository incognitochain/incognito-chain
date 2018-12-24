package zkp

import (
	"fmt"
	"math/big"

	"github.com/minio/blake2b-simd"
	"github.com/ninjadotorg/constant/privacy"
)

/*The protocol is used for opening the commitments of product of 2 values*/
/*------------------------------------------------------*/
/*-------DECLARE INNER INGREDIENT FOR THE PROTOCOL------*/

type PKComProductProof struct {
	//basePoint BasePoint
	D     *privacy.EllipticPoint
	E     *privacy.EllipticPoint
	f     *big.Int
	z     *big.Int
	cmA   *privacy.EllipticPoint
	cmB   *privacy.EllipticPoint
	index byte
}
type PKComProductWitness struct {
	witnessA *big.Int
	randA    *big.Int
	cmB      *privacy.EllipticPoint
	index    byte
}

/*-----------------END OF DECLARATION-------------------*/
/*------------------------------------------------------*/
func (wit *PKComProductWitness) Set(
	witA *big.Int,
	randA *big.Int,
	cmB *privacy.EllipticPoint,
	idx *byte) {
	if wit == nil {
		wit = new(PKComProductWitness)
	}
	wit.witnessA = new(big.Int)
	wit.randA = new(big.Int)
	wit.cmB = new(privacy.EllipticPoint)
	*wit.witnessA = *witA
	*wit.randA = *randA
	*wit.cmB = *cmB
	wit.index = *idx
}

func (pro *PKComProductProof) IsNull() bool {
	if (pro.D == nil) {
		return true
	}
	if (pro.E == nil) {
		return true
	}
	if (pro.f == nil) {
		return true
	}
	if (pro.z == nil) {
		return true
	}
	if (pro.cmA == nil) {
		return true
	}
	if (pro.cmB == nil) {
		return true
	}
	return false
}

func (pro *PKComProductProof) Init() *PKComProductProof {
	pro.D = new(privacy.EllipticPoint).Zero()
	pro.E = new(privacy.EllipticPoint).Zero()
	pro.f = new(big.Int)
	pro.z = new(big.Int)
	pro.cmA = new(privacy.EllipticPoint).Zero()
	pro.cmB = new(privacy.EllipticPoint).Zero()
	return pro
}

func (pro *PKComProductProof) Print() {
	fmt.Println(pro.D)
	fmt.Println(pro.E)
	fmt.Println(len(pro.f.Bytes()))
	fmt.Println(len(pro.z.Bytes()))
	fmt.Println(pro.cmA)
	fmt.Println(pro.cmB)
}

func (pro PKComProductProof) Bytes() []byte {
	if pro.D.IsEqual(new(privacy.EllipticPoint).Zero()) {
		return []byte{}
	}
	var proofbytes []byte
	//if pro.cmA == nil || pro.cmB == nil || {
	//
	//}
	proofbytes = append(proofbytes, pro.cmA.Compress()...)                                  // 33 bytes
	proofbytes = append(proofbytes, pro.cmB.Compress()...)                                  // 33 bytes
	proofbytes = append(proofbytes, pro.D.Compress()...)                                    // 33 bytes
	proofbytes = append(proofbytes, pro.E.Compress()...)                                    // 33 bytes
	proofbytes = append(proofbytes, privacy.AddPaddingBigInt(pro.f, privacy.BigIntSize)...) // 32 bytes
	proofbytes = append(proofbytes, privacy.AddPaddingBigInt(pro.z, privacy.BigIntSize)...) // 32 bytes
	proofbytes = append(proofbytes, pro.index)
	return proofbytes
}

func (pro *PKComProductProof) SetBytes(proofBytes []byte) error {
	pro = pro.Init()
	if len(proofBytes) == 0 {
		return nil
	}
	offset := 0
	pro.cmA.Decompress(proofBytes[offset:])
	offset += privacy.CompressedPointSize
	pro.cmB.Decompress(proofBytes[offset:])
	offset += privacy.CompressedPointSize
	pro.D.Decompress(proofBytes[offset:])
	offset += privacy.CompressedPointSize
	pro.E.Decompress(proofBytes[offset:])
	offset += privacy.CompressedPointSize
	pro.f.SetBytes(proofBytes[offset:offset+32])
	offset += 32
	pro.z.SetBytes(proofBytes[offset:offset+32])
	offset += 32
	pro.index = proofBytes[offset]
	return nil
}
func (wit *PKComProductWitness) Get() *PKComProductWitness {
	return wit
}

/*------------------------------------------------------*/
/*------IMPLEMENT INNER INGREDIENT FOR THE PROTOCOL-----*/
/*Init 2 point G and H for calculate the commitment*/
func computeHashString(data [][]byte) []byte {
	str := make([]byte, 0)
	for i := 0; i < len(data); i++ {
		str = append(str, data[i]...)
	}
	hashFunc := blake2b.New256()
	hashFunc.Write(str)
	hashValue := hashFunc.Sum(nil)
	return hashValue
}

func (wit *PKComProductWitness) Prove() (*PKComProductProof, error) {

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
	|		Compute z = ra*x +s mod p																										 |
	|		Compute f2 = bx+e mod p																										     |
	|		Compute z2 = rb*x+t mod p																										   |
	|		Compute z3 = (rc - a*rb) + s' mod p																						 |
	|	Send f1, f2, z, z2, z3 to the verifier (included in the Proof)                  |
	|---------------------------------------------------------------------------------*/
	/* ---------------------------------------------------------------------------------------------------------------|
	|		THE LINK OF ORIGINAL PAPER FOR THIS PROTOCOL: https://link.springer.com/chapter/10.1007%2F978-3-319-43005-8_1 |
	|   WE CHANGE FROM INTERACTIVE TO NON-INTERACTIVE SCHEME VIA FIAT-SHAMIR HEURISTIC														      |
	|----------------------------------------------------------------------------------------------------------------*/
	proof := new(PKComProductProof)
	proof.Init()
	proof.index = wit.index
	//proof.Init()
	//SpecCom1:=privacy.PCParams{[]privacy.EllipticPoint{proof.basePoint.G,
	//																										  proof.basePoint.H},
	//																										  2}
	d := new(big.Int).SetBytes(privacy.RandBytes(32))
	s := new(big.Int).SetBytes(privacy.RandBytes(32))

	proof.cmA = privacy.PedCom.CommitAtIndex(wit.witnessA, wit.randA, wit.index)
	D := privacy.PedCom.CommitAtIndex(d, s, wit.index)
	E := wit.cmB.ScalarMult(d)
	*proof.D = *D
	*proof.E = *E

	// x = hash(G||H||D||D1||E)
	data := [][]byte{
		privacy.PedCom.G[wit.index].X.Bytes(),
		privacy.PedCom.G[wit.index].Y.Bytes(),
		privacy.PedCom.G[privacy.RAND].Y.Bytes(),
		privacy.PedCom.G[privacy.RAND].Y.Bytes(),
		proof.D.X.Bytes(),
		proof.D.Y.Bytes(),
		proof.E.X.Bytes(),
		proof.E.Y.Bytes(),
	}
	x := new(big.Int)
	x.SetBytes(computeHashString(data))

	//compute f
	a := new(big.Int)
	a.Set(wit.witnessA)
	f := a.Mul(a, x)
	f = f.Add(f, d)
	f = f.Mod(f, privacy.Curve.Params().N)
	*proof.f = *f

	//compute z
	ra := new(big.Int)
	ra.Set(wit.randA)
	z := ra.Mul(ra, x)
	z = z.Add(z, s)
	z = z.Mod(z, privacy.Curve.Params().N)
	*proof.z = *z

	proof.cmA = privacy.PedCom.CommitAtIndex(wit.witnessA, wit.randA, wit.index)
	proof.cmB = wit.cmB

	return proof, nil
}

func (pro *PKComProductProof) Verify() bool {

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

	pts_cmp := new(privacy.EllipticPoint)
	data := [][]byte{
		privacy.PedCom.G[pro.index].X.Bytes(),
		privacy.PedCom.G[pro.index].Y.Bytes(),
		privacy.PedCom.G[privacy.RAND].Y.Bytes(),
		privacy.PedCom.G[privacy.RAND].Y.Bytes(),
		pro.D.X.Bytes(),
		pro.D.Y.Bytes(),
		pro.E.X.Bytes(),
		pro.E.Y.Bytes(),
	}
	x := new(big.Int).SetBytes(computeHashString(data))

	//Check if D,E is on Curve
	if !(privacy.Curve.IsOnCurve(pro.D.X, pro.D.Y) &&
		privacy.Curve.IsOnCurve(pro.E.X, pro.E.Y)) {
		return false
	}
	//Check if f,z in Zp
	if pro.f.Cmp(privacy.Curve.Params().P) == 1 ||
		pro.z.Cmp(privacy.Curve.Params().P) == 1 {
		return false
	}
	//Check witness 1: xA + D == 	CommitAll(f,z)
	A := pro.cmA
	pts_cmp = A.ScalarMult(x).Add(pro.D)
	com1 := privacy.PedCom.CommitAtIndex(pro.f, pro.z, pro.index)
	if !com1.IsEqual(pts_cmp) {
		return false
	}

	//Check witness 2: xB + E == 	CommitAll(f2,z2)
	com2 := pro.cmB.ScalarMult(pro.f)
	pts_cmp = privacy.PedCom.G[pro.index].ScalarMult(x).Add(pro.E)
	if !com2.IsEqual(pts_cmp) {
		return false
	}
	return true
}
