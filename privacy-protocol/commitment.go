package privacy

//
//import (
//	"fmt"
//	"math/big"
//	"sync"
//)
//
//const (
//	SK    = byte(0x00)
//	VALUE = byte(0x01)
//	SND   = byte(0x02)
//	RAND  = byte(0x03)
//	FULL  = byte(0x04)
//)
//
//const (
//	//PCM_CAPACITY ...
//	PCM_CAPACITY = 4
//	//ECM_CAPACITY = 5
//)
//
//// PedersenCommitment represents a commitment that includes 4 generators
////type PedersenCommitment interface {
////	//// Setup initialize the commitment parameters
////	//Setup()
////	// CommitAll commits openings and return result to receiver
////	CommitAll(openings []big.Int) error
////	// Open return the openings of commitment
////	Decommit() []Opening
////
////	//CommitAtIndex([]byte, []byte, byte) []byte
////}
////
////type Opening struct{
////	H []byte
////	index byte
////	// index = 0: PublicKey
////	// index = 1: H
////	// index = 2: Serial Number Derivator
////	// index = 3: Random
////}
//
//type PedersenParams struct {
//	// Generators of commitment
//	G [PCM_CAPACITY]EllipticPoint
//}
//
//// GetPedersenParams generate one-time generators of commitment
//// G[0] = base point of curve
//// G[1] = hash of G[0]
//// G[2] = hash of G[1]
//// G[3] = hash of G[2]
//
//var instance *PedersenParams
//var once sync.Once
//func GetPedersenParams() *PedersenParams {
//	once.Do(func() {
//		var pcm PedersenParams
//		pcm.G[0] = EllipticPoint{new(big.Int).SetBytes(Curve.Params().Gx.Bytes()), new(big.Int).SetBytes(Curve.Params().Gy.Bytes())}
//		for i := 1; i < PCM_CAPACITY; i++ {
//			pcm.G[i] = pcm.G[i-1].HashPoint()
//		}
//		instance = &pcm
//	})
//	return instance
//}
//
//type PedersenCommitment struct{
//	// PedersenCommitment value
//	commitment EllipticPoint
//	// Openings of commitment
//	openings []big.Int
//
//	Type byte
//}
//
//// CommitAll returns Pedersen commitment value
//func (pcm *PedersenCommitment) Commit(openings []Opening) error{
//	if len(openings) > 4{
//		return fmt.Errorf("Length of openings must be less or equal to 4")
//	}
//
//	// Set openings to pcm
//	pcm.Openings = openings
//
//	// Get pedersen params
//	pcmParams := GetPedersenParams()
//
//	// Create commitment
//	pcm.commitment = EllipticPoint{big.NewInt(0), big.NewInt(0)}
//	temp := EllipticPoint{big.NewInt(0), big.NewInt(0)}
//
//	for _, opening := range openings{
//		temp.X, temp.Y = Curve.ScalarMult(pcmParams.G[opening.index].X, pcmParams.G[opening.index].Y, opening.H)
//		pcm.commitment.X, pcm.commitment.Y = Curve.Add(pcm.commitment.X, pcm.commitment.Y, temp.X, temp.Y)
//	}
//	return nil
//}
//
//func (pcm *PedersenCommitment) Decommit() []Opening {
//	return pcm.Openings
//}
//
////// Compress compresses commitment from Elliptic point to 33 bytes array
////func (pcm PedersenCommitment) Compress() []byte {
////	return pcm.PedersenCommitment.CompressPoint()
////}
////
////func (pcm *PedersenCommitment) Decompress(commitmentBytes []byte) {
////	pcm.PedersenCommitment.DecompressPoint(commitmentBytes)
////}
//
//
////type ElGamalCommitment struct{
////	// PedersenCommitment value
////	PedersenCommitment []EllipticPoint
////	// Openings of commitment
////	Openings []Opening
////}
////
////type ElGamalParams struct {
////	// Generators of commitment
////	G [ECM_CAPACITY]EllipticPoint
////}
////
////// GetElGamalParams generate generators of commitment
////// G[0] = base point of curve
////// G[1] = hash of G[0]
////// G[2] = hash of G[1]
////// G[3] = hash of G[2]
////// G[4] = hash of G[3]
////
////var ecmParams *ElGamalParams
////var once1 sync.Once
////func GetElGamalParams() *ElGamalParams {
////	once1.Do(func() {
////		var ecm ElGamalParams
////		ecm.G[0] = EllipticPoint{new(big.Int).SetBytes(Curve.Params().Gx.Bytes()), new(big.Int).SetBytes(Curve.Params().Gy.Bytes())}
////		for i := 1; i < ECM_CAPACITY; i++ {
////			ecm.G[i] = ecm.G[i-1].HashPoint()
////		}
////		ecmParams = &ecm
////	})
////	return ecmParams
////}
////
////
////// CommitAll returns Pedersen commitment value
////func (ecm *ElGamalCommitment) CommitAll(openings []Opening) error{
////	if len(openings) > 4{
////		return fmt.Errorf("Length of openings must be less than or equal to 4")
////	}
////
////	// Set openings to ecm
////	ecm.Openings = openings
////
////	// Get ElGamalParams
////	ecmParams := GetElGamalParams()
////
////	// Create commitment include 2 components
////	ecm.PedersenCommitment = make([]EllipticPoint, 2)
////	// first component: G[4]^rand
////	ecm.PedersenCommitment[0] = EllipticPoint{big.NewInt(0), big.NewInt(0)}
////	ecm.PedersenCommitment[1] = EllipticPoint{big.NewInt(0), big.NewInt(0)}
////	temp := EllipticPoint{big.NewInt(0), big.NewInt(0)}
////
////	for _, opening := range openings{
////		temp.X, temp.Y = Curve.ScalarMult(ecmParams.G[opening.index].X, ecmParams.G[opening.index].Y, opening.H)
////		ecm.PedersenCommitment[0].X, ecm.PedersenCommitment[0].Y = Curve.Add(ecm.PedersenCommitment[0].X, ecm.PedersenCommitment[0].Y, temp.X, temp.Y)
////
////		if opening.index == RAND{
////			ecm.PedersenCommitment[0].X, ecm.PedersenCommitment[0].Y = Curve.ScalarMult(ecmParams.G[ECM_CAPACITY-1].X, ecmParams.G[ECM_CAPACITY-1].Y, opening.H)
////		}
////	}
////	return nil
////}
////
////func (ecm ElGamalCommitment) Decommit() []Opening {
////	return ecm.Openings
////}
////
////// Compress compresses commitment from Elliptic point to 33 bytes array
////func (ecm ElGamalCommitment) Compress() []byte {
////	var commitment []byte
////	commitmentBytes1 := ecm.PedersenCommitment[0].CompressPoint()
////	commitmentBytes2 := ecm.PedersenCommitment[1].CompressPoint()
////	commitment = append(commitment, commitmentBytes1...)
////	commitment = append(commitment, commitmentBytes2...)
////	return commitment
////}
////
////func (ecm * ElGamalCommitment) Decompress(commitmentBytes []byte) {
////	ecm.PedersenCommitment = make([]EllipticPoint, 2)
////	ecm.PedersenCommitment[0].DecompressPoint(commitmentBytes[0:33])
////	ecm.PedersenCommitment[1].DecompressPoint(commitmentBytes[34:66])
////}
////
////func DEcompressPedersen
////
////func DEcompressElgamal
//
//func TestCommitment(cm ElGamalCommitment){
//	var pcm PedersenCommitment
//
//	openings := []Opening{
//		{H: big.NewInt(1).Bytes(), index: VALUE},
//		{H: RandBytes(32), index: RAND},
//	}
//
//	pcm.Commit(openings)
//
//	fmt.Printf("PedersenCommitment point: %+v\n", pcm.commitment)
//	commitBytes := pcm.Compress()
//
//	fmt.Printf("PedersenCommitment  bytes: %v\n", commitBytes)
//
//	cm.Commit()
//	cm.Decommit()
//	cm.Compress()
//
//	var cmBytes []byte
//	var cmEl ElGamalCommitment
//	cmEl.Decompress(cmBytes)
//
//	//DEcompressElgamal()
//
//
//	//pcmParam := GetPedersenParams()
//
//
//}
////
