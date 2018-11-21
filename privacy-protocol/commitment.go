package privacy

import (
	"math/big"
)

// PedersenCommitment represents a commitment that includes 4 generators
type Commitment interface {
	// Setup initialize the commitment parameters
	Setup()
	// Commit commits openings and return result to receiver
	Commit([]Opening)
	// Open return the openings of commitment
	Open() []Opening

	//CommitSpecValue([]byte, []byte, byte) []byte
}

type Opening struct{
	Value []byte
	Random []byte
}

const (
	//PCM_CAPACITY ...
	PCM_CAPACITY = 4
	ECM_CAPACITY = 5
)

type PedersenCommitment struct{
	// Commitment value
	Commitment []EllipticPoint
	// Openings of commitment
	Openings []Opening
	// Generators of commitment
	G []EllipticPoint

}


func (pcm *PedersenCommitment) Setup()  {
	pcm.G = make([]EllipticPoint, PCM_CAPACITY)

	pcm.G[0] = EllipticPoint{new(big.Int).SetBytes(Curve.Params().Gx.Bytes()), new(big.Int).SetBytes(Curve.Params().Gy.Bytes())}

	for i := 1; i < PCM_CAPACITY; i++ {
		pcm.G[i] = pcm.G[i-1].HashPoint()
	}
}


//func TestCommitment(){
//	var pcm PedersenCommitment
//	res := pcm.Setup()
//	fmt.Println(res)
//}

