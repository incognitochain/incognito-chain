package privacy

import (
	"math/big"
)

// PedersenCommitment represents a commitment that includes 4 generators
type PedersenCommitment interface {
	// Params returns the parameters for the commitment
	Params() *PCParams
	// CommitAll commits
	CommitAll(openings []*big.Int) *EllipticPoint
	// CommitAtIndex commits value at index
	CommitAtIndex(value *big.Int, rand *big.Int, index byte) *EllipticPoint
	//CommitEllipticPoint commit a elliptic point
	CommitEllipticPoint(point *EllipticPoint, rand *big.Int) *EllipticPoint
}

// PCParams represents the parameters for the commitment
type PCParams struct {
	G        []*EllipticPoint // generators
	Capacity int
	// G[0]: public key
	// G[1]: Value
	// G[2]: SNDerivator
	// G[3]: ShardID
	// G[4]: Random
}

// newPedersenParams creates new generators
func newPedersenParams() PCParams {
	var pcm PCParams
	pcm.Capacity = 5
	pcm.G = make([]*EllipticPoint, pcm.Capacity)
	//pcm.G[0] := EllipticPoint{new(big.Int).SetBytes(Curve.Params().Gx.Bytes()), new(big.Int).SetBytes(Curve.Params().Gy.Bytes())}
	pcm.G[0] = new(EllipticPoint)
	pcm.G[0].X, pcm.G[0].Y = Curve.Params().Gx, Curve.Params().Gy
	for i := 1; i < pcm.Capacity; i++ {
		pcm.G[i] = pcm.G[0].Hash(i)
	}
	return pcm
}

var PedCom = newPedersenParams()

// Params returns parameters of commitment
func (com PCParams) Params() PCParams {
	return com
}

// CommitAll commits a list of PCM_CAPACITY value(s)
func (com PCParams) CommitAll(openings []*big.Int) *EllipticPoint {
	if len(openings) != com.Capacity {
		return nil
	}
	commitment := new(EllipticPoint)
	commitment.X = big.NewInt(0)
	commitment.Y = big.NewInt(0)
	for i := 0; i < com.Capacity; i++ {
		commitment = commitment.Add(com.G[i].ScalarMul(openings[i]))
	}
	return commitment
}

// CommitAtIndex commits specific value with index and returns 34 bytes
func (com PCParams) CommitAtIndex(value, rand *big.Int, index byte) *EllipticPoint {
	commitment := com.G[com.Capacity-1].ScalarMul(rand).Add(com.G[index].ScalarMul(value))
	return commitment
}

//func (com PCParams) CommitWithSpecPoint(G EllipticPoint, H EllipticPoint, value, sRnd []byte) []byte {
//	var commitment, temp EllipticPoint
//	commitment = EllipticPoint{big.NewInt(0), big.NewInt(0)}
//	temp = EllipticPoint{big.NewInt(0), big.NewInt(0)}
//	temp.X, temp.Y = Curve.ScalarMult(G.X, G.Y, value)
//	commitment.X, commitment.Y = Curve.Add(commitment.X, commitment.Y, temp.X, temp.Y)
//	temp.X, temp.Y = Curve.ScalarMult(H.X, H.Y, sRnd)
//	commitment.X, commitment.Y = Curve.Add(commitment.X, commitment.Y, temp.X, temp.Y)
//	//fmt.Println("cmt Point:", commitment)
//	//append type commitment into the first byte
//	var res []byte
//	var idx byte
//	idx = 0
//	res = append(res, idx)
//	res = append(res, commitment.CompressPoint()...)
//	return res
//}

// CommitBitByBit commits value bit by bit and commits (nBitsThreshold - nBits) zero bits as padding
//func (com PCParams) CommitBitByBit(value uint64, nBits int, nBitsThreshold int, rands [][]byte, index byte) ([][]byte, error) {
//	if len(rands) != nBitsThreshold {
//		return nil, fmt.Errorf("do not have enough random number to commit")
//	}
//	PedCom.InitCommitment()
//
//	commitments := make([][]byte, nBitsThreshold)
//	commitmentPoints := make([]EllipticPoint, nBitsThreshold)
//	for i := 0; value > 0; i++ {
//		commitmentPoints[i] = EllipticPoint{big.NewInt(0), big.NewInt(0)}
//		commitments[i] = make([]byte, 34)
//		//commitmentPoints[i].X, commitmentPoints[i].Y = Curve.ScalarMult(com.G[RAND].X, com.G[RAND].Y, rands[i])
//		//
//		//bit := value % 2
//		//if bit == 1 {
//		//	commitmentPoints[i].X, commitmentPoints[i].Y = Curve.Add(commitmentPoints[i].X, commitmentPoints[i].Y, com.G[index].X, com.G[index].Y)
//		//}
//
//		bit := value % 2
//		if bit == 1 {
//			commitments[i] = PedCom.CommitAtIndex(big.NewInt(1).Bytes(), rands[i], index)
//		} else {
//			commitments[i] = PedCom.CommitAtIndex(big.NewInt(0).Bytes(), rands[i], index)
//		}
//
//		//Compress commitment to byte array
//		//commitments[i] = CompressCommitment(commitmentPoints[i], index)
//		value = value / 2
//	}
//
//	// commit padding bits
//	for j := nBits; j < nBitsThreshold; j++ {
//		commitmentPoints[j] = EllipticPoint{big.NewInt(0), big.NewInt(0)}
//		commitments[j] = make([]byte, 34)
//		commitmentPoints[j].X, commitmentPoints[j].Y = Curve.ScalarMult(com.G[RAND].X, com.G[RAND].Y, rands[j])
//		//Compress commitment to byte array
//		commitments[j] = CompressCommitment(commitmentPoints[j], index)
//	}
//
//	return commitments, nil
//}

//testFunction allow we test each of function for PedersenCommitment
//00: Test generate commitment for four random value and show that on console
//01: Test generate commitment for special value and its random value in special index
// func TestCommitment(testCode byte) bool {

// 	PedCom := NewPedersenParams()
// 	switch testCode {
// 	case 0: //Generate commitment for 4 random value
// 		//Generate 4 random value
// 		value1 := new(big.Int).SetBytes(RandBytes(32))
// 		value2 := new(big.Int).SetBytes(RandBytes(32))
// 		value3 := new(big.Int).SetBytes(RandBytes(32))
// 		valuer := new(big.Int).SetBytes(RandBytes(32))
// 		fmt.Println("H 1: ", value1)
// 		fmt.Println("H 2: ", value2)
// 		fmt.Println("H 3: ", value3)
// 		fmt.Println("H r: ", valuer)

// 		//Compute commitment for all value, 4 is value of constant PCM_CAPACITY
// 		commitmentAll := PedCom.CommitAll([PCM_CAPACITY]big.Int{*value1, *value2, *value3, *valuer})
// 		fmt.Println("Pedersen commitment point: ", commitmentAll)

// 		cmBytes := commitmentAll.Compress()
// 		fmt.Printfunc GenerateChallenge(values [][]byte) []byte {
// 	appendStr := Elcm.G[0].CompressPoint()
// 	for i := 1; i < CM_CAPACITY; i++ {
// 		appendStr = append(appendStr, Elcm.G[i].CompressPoint()...)
// 	}
// 	for i := 0; i < len(values); i++ {
// 		appendStr = append(appendStr, values[i]...)
// 	}
// 	hashFunc := blake2b.New256()
// 	hashFunc.Write(appendStr)
// 	hashValue := hashFunc.Sum(nil)
// 	return hashValue
// }("Pedersen commitment bytes: ", cmBytes)

// 		cmPoint :func GenerateChallenge(values [][]byte) []byte {
// 	appendStr := Elcm.G[0].CompressPoint()
// 	for i := 1; i < CM_CAPACITY; i++ {
// 		appendStr = append(appendStr, Elcm.G[i].CompressPoint()...)
// 	}
// 	for i := 0; i < len(values); i++ {
// 		appendStr = append(appendStr, values[i]...)
// 	}
// 	hashFunc := blake2b.New256()
// 	hashFunc.Write(appendStr)
// 	hashValue := hashFunc.Sum(nil)
// 	return hashValue
// }new(EllipticPoint)
// 		cmPoint.Dfunc GenerateChallenge(values [][]byte) []byte {
// 	appendStr := Elcm.G[0].CompressPoint()
// 	for i := 1; i < CM_CAPACITY; i++ {
// 		appendStr = append(appendStr, Elcm.G[i].CompressPoint()...)
// 	}
// 	for i := 0; i < len(values); i++ {
// 		appendStr = append(appendStr, values[i]...)
// 	}
// 	hashFunc := blake2b.New256()
// 	hashFunc.Write(appendStr)
// 	hashValue := hashFunc.Sum(nil)
// 	return hashValue
// }ompress(cmBytes)
// 		fmt.Printfunc GenerateChallenge(values [][]byte) []byte {
// 	appendStr := Elcm.G[0].CompressPoint()
// 	for i := 1; i < CM_CAPACITY; i++ {
// 		appendStr = append(appendStr, Elcm.G[i].CompressPoint()...)
// 	}
// 	for i := 0; i < len(values); i++ {
// 		appendStr = append(appendStr, values[i]...)
// 	}
// 	hashFunc := blake2b.New256()
// 	hashFunc.Write(appendStr)
// 	hashValue := hashFunc.Sum(nil)
// 	return hashValue
// }("Pedersen commitment decompress: ", cmPoint)

// 		break
// 	case 1: //Generate commitment for special value and its random value
// 		//Generate 2 random value
// 		value1 := new(big.Int).SetBytes(RandBytes(32))
// 		valuer := new(big.Int).SetBytes(RandBytes(32))
// 		fmt.Println("H 1: ", value1)
// 		fmt.Println("H r: ", valuer)

// 		//Compute commitment for special value with index 0
// 		commitmentSpec := PedCom.CommitAtIndex(*value1, *valuer, 0)

// 		fmt.Println("Pedersen commitment value: ", commitmentSpec)

// 		cmBytes := commitmentSpec.Compress()
// 		fmt.Println("Pedersen commitment bytes: ", cmBytes)

// 		cmPoint := new(EllipticPoint)
// 		cmPoint.Decompress(cmBytes)
// 		fmt.Println("Pedersen commitment decompress: ", cmPoint)
// 		break
// 	}
// 	return true
// }
