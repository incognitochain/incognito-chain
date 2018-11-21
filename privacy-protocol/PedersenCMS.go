package privacy

import (
"fmt"
"math/big"

"github.com/minio/blake2b-simd"
)

const (
	PK    = 0
	VALUE = 1
	SN    = 2
	RAND  = 3
)

// PedersenCommitment represents a commitment that includes 4 generators
//type PedersenCommitment interface {
//	// Params returns the parameters for the commitment
//	Params() *PCParams
//	// Setup initialize the parameters
//	Setup() *PCParams
//	// CommitAll commits
//	Commit([PCM_CAPACITY][]byte) []byte
//	getHashOfValues([]byte) []byte
//	CommitSpecValue([]byte, []byte, byte) []byte
//	TestFunction(byte) bool
//}

// PCParams represents the parameters for the commitment
type PCParams struct {
	G [PCM_CAPACITY]EllipticPoint // generators
	// G[0]: public key
	// G[1]: Value
	// G[2]: SerialNumber
	// G[3]: Random
}


//PCParams ...
var Pcm PCParams

//GetHashOfValues get hash of n points in G append with input values
//return blake_2b(G[0]||G[1]||...||G[PCM_CAPACITY-1]||<values>)
func (com PCParams) GetHashOfValues(values [][]byte) []byte {
	appendStr := Pcm.G[0].CompressPoint()
	for i := 1; i < PCM_CAPACITY; i++ {
		appendStr = append(appendStr, Pcm.G[i].CompressPoint()...)
	}
	for i := 0; i < len(values); i++ {
		appendStr = append(appendStr, values[i]...)
	}
	hashFunc := blake2b.New256()
	hashFunc.Write(appendStr)
	hashValue := hashFunc.Sum(nil)
	return hashValue
}

// Params returns parameters of commitment
func (com PCParams) Params() PCParams {
	return com
}

// Setup initializes parameters of Pedersen commitment
func (com *PCParams) InitCommitment() {

	// G0 is the base point of curve P256
	// G1 is the G0's hash
	// G2 is the G1's hash
	// G3 is the G2's hash
	//com.G[0] = EllipticPoint{new(big.Int).SetBytes(107 23 209 242 225 44 66 71 248 188 230 229 99 164 64 242 119 3 125 129 45 235 51 160 244 161 57 69 216 152 194 150)}
	//com.G[1] = EllipticPoint{big.NewInt(23958808978146169065002618177085267768449753078050893909214942889381335385516),
	//						big.NewInt(7683342728189100387735055495645379680280002577849876913406403613392655773246)}
	//com.G[2] = EllipticPoint{big.NewInt(3815079413196168640531957554999832314172554264133353679019997712191027719881),
	//						big.NewInt(77272455227631396963147866972129356047185376243416151568214292723259718584837)}
	//com.G[3] = EllipticPoint{big.NewInt(96857018922751703925448770729453114406629947956893140191061131561860088449600),
	//						big.NewInt(91021963404625493633337876713358362238547713052118385945840294627046689931308)}

	com.G[0] = EllipticPoint{Curve.Params().Gx, Curve.Params().Gy}
	//fmt.Printf("G0.X: %#v\n", com.G[0].X.Bytes())
	//fmt.Printf("G0.Y: %#v\n", com.G[0].Y.Bytes())
	for i := 1; i < PCM_CAPACITY; i++ {
		com.G[i] = com.G[i-1].HashPoint()
		//fmt.Printf("G%v.X: %#v\n", i, com.G[i].X.Bytes())
		//fmt.Printf("G%v.Y: %#v\n", i, com.G[i].Y.Bytes())
	}

	//TODO: hard code parameters
}

// Commit commits a list of PCM_CAPACITY value(s)
func (com PCParams) Commit(values [PCM_CAPACITY][]byte) []byte {
	temp := EllipticPoint{big.NewInt(0), big.NewInt(0)}
	commitment := EllipticPoint{big.NewInt(0), big.NewInt(0)}
	for i := 0; i < PCM_CAPACITY; i++ {
		temp.X, temp.Y = Curve.ScalarMult(com.G[i].X, com.G[i].Y, values[i])
		commitment.X, commitment.Y = Curve.Add(commitment.X, commitment.Y, temp.X, temp.Y)
	}

	// convert result from Elliptic to bytes array
	// append type commitment into the first byte
	res := CompressCommitment(commitment, FULL_CM)
	return res
}

// CommitSpecValue commits specific value with index and returns 34 bytes
func (com PCParams) CommitSpecValue(value, sRnd []byte, index byte) []byte {
	var commitment, temp EllipticPoint
	commitment = EllipticPoint{big.NewInt(0), big.NewInt(0)}
	temp = EllipticPoint{big.NewInt(0), big.NewInt(0)}
	temp.X, temp.Y = Curve.ScalarMult(com.G[PCM_CAPACITY-1].X, com.G[PCM_CAPACITY-1].Y, sRnd)
	commitment.X, commitment.Y = Curve.Add(commitment.X, commitment.Y, temp.X, temp.Y)
	temp.X, temp.Y = Curve.ScalarMult(com.G[index].X, com.G[index].Y, value)
	commitment.X, commitment.Y = Curve.Add(commitment.X, commitment.Y, temp.X, temp.Y)

	// commit commitment from a elliptic point to 34 bytes
	res := CompressCommitment(commitment, index)
	return res
}

func (com PCParams) CommitWithSpecPoint(G EllipticPoint, H EllipticPoint, value, sRnd []byte) []byte {
	var commitment, temp EllipticPoint
	commitment = EllipticPoint{big.NewInt(0), big.NewInt(0)}
	temp = EllipticPoint{big.NewInt(0), big.NewInt(0)}
	temp.X, temp.Y = Curve.ScalarMult(G.X, G.Y, value)
	commitment.X, commitment.Y = Curve.Add(commitment.X, commitment.Y, temp.X, temp.Y)
	temp.X, temp.Y = Curve.ScalarMult(H.X, H.Y, sRnd)
	commitment.X, commitment.Y = Curve.Add(commitment.X, commitment.Y, temp.X, temp.Y)
	//fmt.Println("cmt Point:", commitment)
	//append type commitment into the first byte
	var res []byte
	var idx byte
	idx = 0
	res = append(res, idx)
	res = append(res, commitment.CompressPoint()...)
	return res
}

// CommitBitByBit commits value bit by bit and commits (nBitsThreshold - nBits) zero bits as padding
func (com PCParams) CommitBitByBit(value uint64, nBits int, nBitsThreshold int, rands [][]byte, index byte) ([][]byte, error) {
	if len(rands) != nBitsThreshold {
		return nil, fmt.Errorf("do not have enough random number to commit")
	}
	Pcm.InitCommitment()

	commitments := make([][]byte, nBitsThreshold)
	commitmentPoints := make([]EllipticPoint, nBitsThreshold)
	for i := 0; value > 0; i++ {
		commitmentPoints[i] = EllipticPoint{big.NewInt(0), big.NewInt(0)}
		commitments[i] = make([]byte, 34)
		//commitmentPoints[i].X, commitmentPoints[i].Y = Curve.ScalarMult(com.G[RAND].X, com.G[RAND].Y, rands[i])
		//
		//bit := value % 2
		//if bit == 1 {
		//	commitmentPoints[i].X, commitmentPoints[i].Y = Curve.Add(commitmentPoints[i].X, commitmentPoints[i].Y, com.G[index].X, com.G[index].Y)
		//}

		bit := value % 2
		if bit == 1 {
			commitments[i] = Pcm.CommitSpecValue(big.NewInt(1).Bytes(), rands[i], index)
		} else {
			commitments[i] = Pcm.CommitSpecValue(big.NewInt(0).Bytes(), rands[i], index)
		}

		//Compress commitment to byte array
		//commitments[i] = CompressCommitment(commitmentPoints[i], index)
		value = value / 2
	}

	// commit padding bits
	for j := nBits; j < nBitsThreshold; j++ {
		commitmentPoints[j] = EllipticPoint{big.NewInt(0), big.NewInt(0)}
		commitments[j] = make([]byte, 34)
		commitmentPoints[j].X, commitmentPoints[j].Y = Curve.ScalarMult(com.G[RAND].X, com.G[RAND].Y, rands[j])
		//Compress commitment to byte array
		commitments[j] = CompressCommitment(commitmentPoints[j], index)
	}

	return commitments, nil
}

//testFunction allow we test each of function for PedersenCommitment
//00: Test generate commitment for four random value and show that on console
//01: Test generate commitment for special value and its random value in special index
func (com PCParams) TestFunction(testCode byte) bool {
	switch testCode {
	case 0: //Generate commitment for 4 random value
		//Generate 4 random value
		value1 := RandBytes(32)
		value2 := RandBytes(32)
		value3 := RandBytes(32)
		valuer := RandBytes(32)
		fmt.Println("Value 1: ", value1)
		fmt.Println("Value 2: ", value2)
		fmt.Println("Value 3: ", value3)
		fmt.Println("Value r: ", valuer)

		//Compute commitment for all value, 4 is value of constant PCM_CAPACITY
		commitmentAll := Pcm.Commit([4][]byte{value1, value2, value3, valuer})

		fmt.Println("Commitment value: ", commitmentAll)
		break
	case 1: //Generate commitment for special value and its random value
		//Generate 2 random value
		value1 := RandBytes(32)
		valuer := RandBytes(32)
		fmt.Println("Value 1: ", value1)
		fmt.Println("Value r: ", valuer)

		//Compute commitment for special value with index 0
		commitmentSpec := Pcm.CommitSpecValue(value1, valuer, 0)

		fmt.Println("Commitment value: ", commitmentSpec)
		break
	}
	return true
}
