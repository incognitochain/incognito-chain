package privacy

//
//import (
//	"fmt"
//
//	"math/big"
//
//	"github.com/minio/blake2b-simd"
//)
//
//// Check type of commitments
//
//
//const (
//	//CM_CAPACITY ...
//	CM_CAPACITY = 5
//)
//
//// ElGamalCommitment represents a commitment that includes 4 generators
////type ElGamalCommitment interface {
////	// Params returns the parameters for the commitment
////	Params() *ElCParams
////	// Setup initialize the parameters
////	InitCommitment() *ElCParams
////	// CommitAll commits
////	CommitAll([CM_CAPACITY][]byte) []byte
////	GetHashOfValues([]byte) []byte
////	CommitAtIndex([]byte, []byte, byte) []byte
////	TestFunction(byte) bool
////}
//
//// ElCParams represents the parameters for the commitment
//type ElCParams struct {
//	G [CM_CAPACITY]EllipticPoint // generators
//	// G[0]: the first component's commitment
//	// G[1]: public key
//	// G[2]: H
//	// G[3]: SNDerivator
//	// G[4]: Random
//}
//
////Elcm ...
//var Elcm ElCParams
//
//
//

//
////ComputeYCoord calculates Y coord from X
//func computeYCoord(x *big.Int) *big.Int {
//	Q := Curve.Params().P
//	temp := new(big.Int)
//	xTemp := new(big.Int)
//
//	// Y = +-sqrt(x^3 - 3*x + B)
//	x3 := new(big.Int).Mul(x, x)
//	x3.Mul(x3, x)
//	x3.Add(x3, Curve.Params().B)
//	x3.Sub(x3, xTemp.Mul(x, new(big.Int).SetInt64(3)))
//	x3.Mod(x3, Curve.Params().P)
//
//	//check P = 3 mod 4?
//	if temp.Mod(Q, new(big.Int).SetInt64(4)).Cmp(new(big.Int).SetInt64(3)) == 0 {
//		//		fmt.Println("Ok!!!")
//	}
//
//	// Now calculate sqrt mod p of x^3 - 3*x + B
//	// This code used to do a full sqrt based on tonelli/shanks,
//	// but this was replaced by the algorithms referenced in
//	// https://bitcointalk.org/index.php?topic=162805.msg1712294#msg1712294
//	y := new(big.Int).Exp(x3, PAdd1Div4(Q), Q)
//	// Check that y is a square root of x^3  - 3*x + B.
//	y2 := new(big.Int).Mul(y, y)
//	y2.Mod(y2, Curve.Params().P)
//	//fmt.Printf("y2: %X\n", y2)
//	//fmt.Printf("x3: %X\n", x3)
//	if y2.Cmp(x3) != 0 {
//		return nil
//	}
//	return y
//}
//
//// Params returns parameters of commitment
//func (com ElCParams) Params() ElCParams {
//	return com
//}
//
//// Setup initializes parameters of Pedersen commitment
//func (com *ElCParams) InitCommitment() {
//
//	// G0 is the base point of curve P256
//	// G1 is the G0's hash
//	// G2 is the G1's hash
//	// G3 is the G2's hash
//	// G4 is the G3's hash
//	//com.G[0] = EllipticPoint{new(big.Int).SetBytes(107 23 209 242 225 44 66 71 248 188 230 229 99 164 64 242 119 3 125 129 45 235 51 160 244 161 57 69 216 152 194 150)}
//	//com.G[1] = EllipticPoint{big.NewInt(23958808978146169065002618177085267768449753078050893909214942889381335385516),
//	//						big.NewInt(7683342728189100387735055495645379680280002577849876913406403613392655773246)}
//	//com.G[2] = EllipticPoint{big.NewInt(3815079413196168640531957554999832314172554264133353679019997712191027719881),
//	//						big.NewInt(77272455227631396963147866972129356047185376243416151568214292723259718584837)}
//	//com.G[3] = EllipticPoint{big.NewInt(96857018922751703925448770729453114406629947956893140191061131561860088449600),
//	//						big.NewInt(91021963404625493633337876713358362238547713052118385945840294627046689931308)}
//
//	com.G[0] = EllipticPoint{Curve.Params().Gx, Curve.Params().Gy}
//	for i := 1; i < CM_CAPACITY; i++ {
//		com.G[i] = com.G[i-1].HashPoint()
//	}
//
//	//TODO: hard code parameters
//}
//
//// CommitAll commits a list of CM_CAPACITY value(s)
//// PedersenCommitment includes 2 components:
//// Component 1: g[0]^r
//// Component 2: g[1]^v1 * g[2]^v2 * g[3]^v3 * g[4]^r
//// Component 1's r and component 2's r are the same
//func (com ElCParams) Commit(values [CM_CAPACITY - 1][]byte) []byte {
//	temp := EllipticPoint{big.NewInt(0), big.NewInt(0)}
//
//	component1 := new(EllipticPoint)
//	component1.X, component1.Y = Curve.ScalarMult(com.G[0].X, com.G[0].Y, values[CM_CAPACITY-2])
//
//	component2 := EllipticPoint{big.NewInt(0), big.NewInt(0)}
//
//	for i := 1; i < CM_CAPACITY-1; i++ {
//		temp.X, temp.Y = Curve.ScalarMult(com.G[i].X, com.G[i].Y, values[i])
//		component2.X, component2.Y = Curve.Add(component2.X, component2.Y, temp.X, temp.Y)
//	}
//
//	//convert Component1 from Elliptic to bytes array 33 bytes
//	componentBytes1 := component1.CompressPoint()
//	//convert Component2 from Elliptic to bytes array 34 bytes
//	// the first byte is commitment's type
//	componentBytes2 := CompressCommitment(component2, FULL)
//
//	var commitment []byte
//	commitment = append(commitment, componentBytes1...)
//	commitment = append(commitment, componentBytes2...)
//
//	return commitment
//}
//
//// CommitAtIndex commits specific value with index and returns 34 bytes
//func (com ElCParams) CommitAtIndex(value, sRnd []byte, index byte) []byte {
//	component1 := new(EllipticPoint)
//	component1.X, component1.Y = Curve.ScalarMult(com.G[0].X, com.G[0].Y, sRnd)
//
//	var component2, temp EllipticPoint
//	component2 = EllipticPoint{big.NewInt(0), big.NewInt(0)}
//	temp = EllipticPoint{big.NewInt(0), big.NewInt(0)}
//	temp.X, temp.Y = Curve.ScalarMult(com.G[CM_CAPACITY-1].X, com.G[CM_CAPACITY-1].Y, sRnd)
//	component2.X, component2.Y = Curve.Add(component2.X, component2.Y, temp.X, temp.Y)
//	temp.X, temp.Y = Curve.ScalarMult(com.G[index].X, com.G[index].Y, value)
//	component2.X, component2.Y = Curve.Add(component2.X, component2.Y, temp.X, temp.Y)
//
//	//convert Component1 from Elliptic to bytes array 33 bytes
//	componentBytes1 := component1.CompressPoint()
//	//convert Component2 from Elliptic to bytes array 34 bytes
//	// the first byte is commitment's type
//	componentBytes2 := CompressCommitment(component2, index)
//
//	var commitment []byte
//	commitment = append(commitment, componentBytes1...)
//	commitment = append(commitment, componentBytes2...)
//
//	return commitment
//}
//
//func (com ElCParams) CommitWithSpecPoint(G EllipticPoint, H EllipticPoint, value, sRnd []byte) []byte {
//
//	component1 := new(EllipticPoint)
//	component1.X, component1.Y = Curve.ScalarMult(H.X, H.Y, sRnd)
//
//	var component2, temp EllipticPoint
//	component2 = EllipticPoint{big.NewInt(0), big.NewInt(0)}
//	temp = EllipticPoint{big.NewInt(0), big.NewInt(0)}
//	temp.X, temp.Y = Curve.ScalarMult(G.X, G.Y, value)
//	component2.X, component2.Y = Curve.Add(component2.X, component2.Y, temp.X, temp.Y)
//	temp.X, temp.Y = Curve.ScalarMult(H.X, H.Y, sRnd)
//	component2.X, component2.Y = Curve.Add(component2.X, component2.Y, temp.X, temp.Y)
//
//	//append type component2 into the first byte
//	var componentBytes2 []byte
//	var idx byte
//	idx = 0
//	componentBytes2 = append(componentBytes2, idx)
//	componentBytes2 = append(componentBytes2, component2.CompressPoint()...)
//
//	componentBytes1 := component1.CompressPoint()
//
//	var commitment []byte
//	commitment = append(commitment, componentBytes1...)
//	commitment = append(commitment, componentBytes2...)
//
//	return commitment
//
//}
//
//// CommitBitByBit commits value bit by bit and commits (nBitsThreshold - nBits) zero bits as padding
////func (com ElCParams) CommitBitByBit(value uint64, nBits int, nBitsThreshold int, rands[][]byte, index byte) ([][]byte, error){
////	if len(rands) != nBitsThreshold{
////		return nil, fmt.Errorf("do not have enough random number to commit")
////	}
////	Elcm.Setup()
////
////	commitments := make([][]byte, nBitsThreshold)
////	commitmentPoints := make([]EllipticPoint, nBitsThreshold)
////	for i:=0; value > 0; i++{
////		commitmentPoints[i] = EllipticPoint{big.NewInt(0), big.NewInt(0)}
////		commitments[i] = make([]byte, 34)
////		//commitmentPoints[i].X, commitmentPoints[i].Y = Curve.ScalarMult(com.G[RAND].X, com.G[RAND].Y, rands[i])
////		//
////		//bit := value % 2
////		//if bit == 1 {
////		//	commitmentPoints[i].X, commitmentPoints[i].Y = Curve.Add(commitmentPoints[i].X, commitmentPoints[i].Y, com.G[index].X, com.G[index].Y)
////		//}
////
////		bit := value % 2
////		if bit == 1 {
////			commitments[i] = Elcm.CommitAtIndex(big.NewInt(1).Bytes(), rands[i], index)
////		} else{
////			commitments[i] = Elcm.CommitAtIndex(big.NewInt(0).Bytes(), rands[i], index)
////		}
////
////		//Compress commitment to byte array
////		//commitments[i] = CompressCommitment(commitmentPoints[i], index)
////		value = value / 2
////	}
////
////	// commit padding bits
////	for j := nBits; j < nBitsThreshold; j++{
////		commitmentPoints[j] = EllipticPoint{big.NewInt(0), big.NewInt(0)}
////		commitments[j] = make([]byte, 34)
////		commitmentPoints[j].X, commitmentPoints[j].Y = Curve.ScalarMult(com.G[RAND].X, com.G[RAND].Y, rands[j])
////		//Compress commitment to byte array
////		commitments[j] = CompressCommitment(commitmentPoints[j], index)
////	}
////
////	return commitments, nil
////}
//
////testFunction allow we test each of function for ElGamalCommitment
////00: Test generate commitment for four random value and show that on console
////01: Test generate commitment for special value and its random value in special index
//func (com ElCParams) TestFunction(testCode byte) bool {
//	switch testCode {
//	case 0: //Generate commitment for 4 random value
//		//Generate 4 random value
//		value1 := RandBytes(32)
//		value2 := RandBytes(32)
//		value3 := RandBytes(32)
//		valuer := RandBytes(32)
//		fmt.Println("H 1: ", value1)
//		fmt.Println("H 2: ", value2)
//		fmt.Println("H 3: ", value3)
//		fmt.Println("H r: ", valuer)
//
//		//Compute commitment for all value, 4 is value of constant CM_CAPACITY
//		commitmentAll := Elcm.Commit([CM_CAPACITY - 1][]byte{value1, value2, value3, valuer})
//
//		fmt.Println("PedersenCommitment value: ", commitmentAll)
//		break
//	case 1: //Generate commitment for special value and its random value
//		//Generate 2 random value
//		value1 := RandBytes(32)
//		valuer := RandBytes(32)
//		fmt.Println("H 1: ", value1)
//		fmt.Println("H r: ", valuer)
//
//		//Compute commitment for special value with index 0
//		commitmentSpec := Elcm.CommitAtIndex(value1, valuer, 0)
//
//		fmt.Println("PedersenCommitment value: ", commitmentSpec)
//		break
//	}
//	return true
//}
//
////CompressCommitment from 64 bytes to 34 bytes (include bytes index)
//func CompressCommitment(cmPoint EllipticPoint, typeCommitment byte) []byte {
//	var commitment []byte
//	commitment = append(commitment, typeCommitment)
//	commitment = append(commitment, cmPoint.CompressPoint()...)
//	return commitment
//}
//
//// DecompressCommitment decompress commitment byte array
//func DecompressCommitment(commitment []byte) (point *EllipticPoint, err error) {
//	//typeCommitment := commitment[0]
//	//fmt.Printf("Type Commmitment: %v\n", typeCommitment)
//	//Decompress the second component's commitment
//	return DecompressKey(commitment[34:67])
//}
