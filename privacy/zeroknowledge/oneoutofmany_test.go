package zkp

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"log"
	"math/big"
	"testing"
)

func TestMain(m *testing.M) {
	log.SetOutput(ioutil.Discard)
	m.Run()
}

var _ = func() (_ struct{}) {
	fmt.Println("This runs before init()!")
	privacy.Logger.Init(common.NewBackend(nil).Logger("test", true))
	return
}()

//TestPKOneOfMany test protocol for one of many Commitment is Commitment to zero
func TestPKOneOfMany(t *testing.T) {
	//witness := new(OneOutOfManyWitness)
	//
	//indexIsZero := 2
	//
	//// list of commitments
	//commitments := make([]*privacy.EllipticPoint, privacy.CMRingSize)
	//snDerivators := make([]*big.Int, privacy.CMRingSize)
	//randoms := make([]*big.Int, privacy.CMRingSize)
	//
	//for i := 0; i < privacy.CMRingSize; i++ {
	//	snDerivators[i] = privacy.RandScalar()
	//	randoms[i] = privacy.RandScalar()
	//	commitments[i] = privacy.PedCom.CommitAtIndex(snDerivators[i], randoms[i], privacy.SND)
	//}
	//
	//// create Commitment to zero at indexIsZero
	//snDerivators[indexIsZero] = big.NewInt(0)
	//commitments[indexIsZero] = privacy.PedCom.CommitAtIndex(snDerivators[indexIsZero], randoms[indexIsZero], privacy.SND)
	//
	//witness.Set(commitments, randoms[indexIsZero], uint64(indexIsZero))
	//start := time.Now()
	//proof, err := witness.Prove()
	//end := time.Since(start)
	//privacy.Logger.Log.Infof("One out of many proving time: %v\n", end)
	//if err != nil {
	//	privacy.Logger.Log.Error(err)
	//}
	//
	////Convert proof to bytes array
	//proofBytes := proof.Bytes()
	//
	//privacy.Logger.Log.Infof("One out of many proof size: %v\n", len(proofBytes))
	//
	////revert bytes array to proof
	//proof2 := new(OneOutOfManyProof).Init()
	//proof2.SetBytes(proofBytes)
	//proof2.stmt.commitments = commitments
	//
	//start = time.Now()
	//
	//res := proof.Verify()
	//end = time.Since(start)
	//privacy.Logger.Log.Infof("One out of many verification time: %v\n", end)
	//
	//assert.Equal(t, true, res)

	//test for JS
	//hCommitments := make([]*privacy.EllipticPoint, privacy.CMRingSize)
	//hCommitments[0] = new(privacy.EllipticPoint)
	//hCommitments[0].Decompress([]byte{3, 225, 145, 19, 109, 28, 119, 32, 54, 143, 94, 108, 235, 211, 15, 196, 220, 174, 19, 96, 115, 86, 50, 41, 0, 121, 123, 23, 241, 231, 126, 44, 124})
	//
	//hCommitments[1] = new(privacy.EllipticPoint)
	//hCommitments[1].Decompress([]byte{3, 61, 175, 99, 143, 220, 215, 180, 105, 134, 244, 90, 204, 108, 93, 0, 176, 93, 250, 102, 73, 91, 69, 42, 125, 7, 227, 35, 176, 30, 91, 117, 173})
	//
	//hCommitments[2] = new(privacy.EllipticPoint)
	//hCommitments[2].Decompress([]byte{3, 131, 74, 242, 235, 166, 189, 189, 108, 197, 76, 104, 145, 18, 180, 201, 70, 168, 40, 234, 56, 128, 109, 46, 96, 6, 198, 207, 10, 148, 69, 175, 22})
	//
	//hCommitments[3] = new(privacy.EllipticPoint)
	//hCommitments[3].Decompress([]byte{3, 158, 17, 107, 212, 6, 7, 198, 47, 35, 121, 78, 205, 225, 89, 10, 212, 110, 80, 42, 65, 84, 208, 177, 158, 216, 213, 252, 6, 186, 47, 112, 71})
	//
	//hCommitments[4] = new(privacy.EllipticPoint)
	//hCommitments[4].Decompress([]byte{2, 167, 93, 55, 90, 102, 118, 80, 111, 118, 145, 48, 78, 181, 214, 249, 81, 56, 89, 106, 173, 41, 87, 104, 249, 85, 103, 245, 115, 48, 171, 50, 166})
	//
	//hCommitments[5] = new(privacy.EllipticPoint)
	//hCommitments[5].Decompress([]byte{2, 232, 218, 17, 250, 193, 207, 155, 20, 217, 68, 12, 192, 177, 238, 240, 181, 138, 216, 226, 242, 138, 62, 8, 146, 231, 7, 121, 178, 19, 147, 207, 181})
	//
	//hCommitments[6] = new(privacy.EllipticPoint)
	//hCommitments[6].Decompress([]byte{3, 33, 30, 16, 170, 189, 229, 38, 67, 132, 147, 243, 195, 60, 150, 70, 185, 41, 149, 150, 113, 19, 195, 31, 93, 7, 153, 30, 232, 255, 202, 28, 31})
	//
	//hCommitments[7] = new(privacy.EllipticPoint)
	//hCommitments[7].Decompress([]byte{3, 126, 80, 209, 45, 86, 216, 198, 49, 250, 206, 154, 70, 142, 41, 160, 173, 116, 55, 253, 120, 171, 120, 243, 154, 243, 211, 223, 39, 11, 227, 247, 4})

	//hCommitments[3] = privacy.PedCom.CommitAtIndex(big.NewInt(0), big.NewInt(100), privacy.SK)

	commitments := make([]*privacy.EllipticPoint, privacy.CMRingSize)
	commitments[0] = new(privacy.EllipticPoint)
	commitments[0].Decompress([]byte{2, 63, 242, 198, 114, 250, 36, 102, 85, 80, 173, 148, 153, 247, 78, 215, 30, 54, 40, 193, 40, 190, 206, 73, 198, 39, 23, 48, 56, 136, 58, 91, 167})

	commitments[1] = new(privacy.EllipticPoint)
	commitments[1].Decompress([]byte{2, 203, 30, 129, 126, 123, 135, 125, 29, 43, 137, 52, 148, 146, 17, 87, 85, 237, 67, 191, 175, 241, 86, 102, 239, 183, 114, 78, 11, 127, 116, 16, 143})

	commitments[2] = new(privacy.EllipticPoint)
	commitments[2].Decompress([]byte{2, 123, 251, 169, 31, 79, 237, 122, 212, 173, 208, 175, 20, 111, 140, 19, 185, 72, 17, 229, 163, 84, 255, 63, 157, 51, 251, 209, 160, 122, 250, 30, 116})

	commitments[3] = new(privacy.EllipticPoint)
	commitments[3].Decompress([]byte{2, 174, 247, 205, 128, 120, 191, 95, 219, 186, 227, 95, 10, 157, 200, 224, 109, 152, 179, 5, 188, 162, 125, 167, 214, 127, 178, 173, 246, 109, 18, 23, 254})

	commitments[4] = new(privacy.EllipticPoint)
	commitments[4].Decompress([]byte{2, 8, 49, 76, 243, 238, 108, 171, 35, 55, 118, 239, 95, 214, 43, 88, 155, 4, 152, 62, 74, 15, 62, 203, 158, 189, 163, 62, 150, 255, 220, 14, 170})

	commitments[5] = new(privacy.EllipticPoint)
	commitments[5].Decompress([]byte{3, 205, 17, 244, 179, 44, 154, 114, 20, 78, 113, 196, 20, 133, 98, 165, 111, 74, 139, 53, 74, 224, 153, 41, 66, 224, 190, 220, 179, 136, 193, 241, 218})

	commitments[6] = new(privacy.EllipticPoint)
	commitments[6].Decompress([]byte{3, 191, 145, 66, 202, 76, 92, 64, 185, 89, 85, 149, 239, 190, 231, 208, 214, 25, 0, 218, 142, 114, 18, 188, 122, 111, 213, 6, 108, 128, 129, 122, 109})

	commitments[7] = new(privacy.EllipticPoint)
	commitments[7].Decompress([]byte{2, 186, 15, 36, 170, 79, 9, 118, 9, 249, 10, 215, 114, 5, 80, 9, 156, 206, 217, 242, 156, 30, 210, 169, 109, 221, 103, 37, 186, 24, 88, 47, 121})

	commitments[3] = privacy.PedCom.CommitAtIndex(big.NewInt(0), big.NewInt(100), privacy.SK)

	//hWit := new(OneOutOfManyWitness)
	//hWit.Set(hCommitments, big.NewInt(100), 3)
	//hProof1, _ := hWit.Prove()
	//res := hProof1.Verify()
	//fmt.Println(res)

	hProof := new(OneOutOfManyProof).Init()
	hProof.SetBytes([]byte{2, 252, 129, 242, 143, 200, 121, 32, 63, 4, 37, 248, 52, 208, 1, 131, 68, 219, 226, 104, 7, 122, 29, 72, 133, 107, 9, 40, 15, 228, 158, 214, 221, 2, 222, 146, 133, 245, 79, 118, 168, 82, 2, 191, 122, 39, 39, 20, 4, 177, 155, 90, 127, 81, 19, 22, 247, 4, 123, 233, 106, 152, 180, 113, 60, 12, 2, 133, 200, 39, 72, 31, 216, 112, 63, 162, 62, 169, 73, 160, 57, 177, 78, 135, 91, 0, 192, 254, 40, 7, 125, 230, 191, 166, 144, 154, 99, 212, 86, 2, 43, 24, 179, 196, 182, 191, 180, 242, 71, 29, 32, 167, 238, 182, 182, 114, 211, 13, 178, 238, 119, 21, 71, 233, 151, 119, 25, 251, 59, 4, 206, 78, 3, 63, 197, 122, 186, 102, 162, 22, 232, 197, 253, 163, 244, 48, 180, 189, 70, 189, 188, 233, 191, 60, 237, 215, 195, 187, 64, 124, 245, 246, 140, 50, 196, 2, 96, 114, 90, 5, 184, 21, 182, 146, 209, 7, 218, 236, 236, 198, 188, 48, 23, 161, 179, 99, 14, 203, 250, 57, 189, 77, 102, 47, 57, 141, 204, 221, 2, 38, 181, 212, 41, 75, 247, 193, 123, 39, 187, 182, 204, 53, 120, 13, 171, 26, 198, 76, 231, 251, 185, 23, 177, 210, 232, 169, 24, 76, 120, 172, 184, 3, 71, 156, 156, 242, 42, 237, 66, 77, 93, 28, 173, 196, 241, 201, 162, 217, 9, 245, 151, 47, 42, 210, 80, 9, 171, 186, 39, 192, 119, 64, 45, 123, 2, 61, 83, 62, 204, 249, 69, 136, 101, 8, 202, 52, 117, 52, 118, 152, 79, 196, 164, 109, 172, 239, 2, 166, 220, 100, 167, 27, 64, 170, 233, 54, 163, 2, 247, 126, 170, 208, 26, 156, 136, 223, 94, 93, 167, 42, 223, 247, 236, 72, 0, 91, 117, 196, 212, 118, 200, 68, 228, 39, 18, 203, 95, 43, 212, 166, 2, 97, 48, 62, 10, 220, 196, 116, 121, 121, 12, 210, 15, 62, 161, 0, 141, 232, 215, 23, 253, 198, 64, 181, 45, 244, 0, 150, 25, 194, 223, 114, 204, 2, 86, 48, 227, 172, 10, 149, 73, 110, 144, 37, 174, 81, 90, 188, 127, 135, 6, 212, 221, 145, 110, 212, 163, 126, 126, 176, 90, 77, 95, 67, 136, 13, 185, 223, 154, 106, 217, 197, 94, 158, 10, 69, 97, 23, 73, 221, 132, 49, 3, 103, 215, 38, 174, 51, 199, 72, 174, 127, 8, 177, 239, 148, 198, 64, 191, 37, 49, 129, 10, 179, 181, 89, 213, 222, 179, 64, 231, 201, 40, 28, 63, 73, 69, 142, 49, 209, 26, 232, 46, 43, 105, 15, 107, 93, 145, 171, 184, 180, 120, 102, 99, 255, 203, 142, 244, 62, 28, 153, 220, 145, 103, 53, 167, 69, 82, 46, 76, 110, 108, 121, 34, 21, 152, 128, 79, 117, 62, 7, 27, 141, 62, 227, 50, 157, 183, 191, 183, 248, 43, 235, 105, 204, 129, 136, 1, 99, 157, 146, 43, 178, 77, 175, 107, 103, 203, 9, 185, 181, 192, 57, 176, 3, 66, 170, 255, 18, 118, 121, 184, 80, 25, 254, 105, 234, 207, 31, 111, 17, 97, 132, 234, 6, 25, 230, 223, 133, 170, 152, 93, 233, 200, 169, 3, 137, 31, 154, 85, 23, 76, 201, 156, 183, 185, 102, 216, 126, 254, 171, 224, 68, 255, 9, 249, 127, 208, 11, 90, 188, 133, 252, 231, 5, 139, 59, 35, 130, 162, 122, 251, 11, 11, 188, 124, 93, 14, 45, 52, 93, 170, 131, 4, 103, 189, 210, 159, 122, 145, 133, 28, 197, 199, 11, 178, 241, 158, 215, 145, 190, 167, 247, 70, 1, 216, 166, 52, 153, 15, 229, 32, 80, 116, 136, 248, 116, 172, 221, 165, 235, 221, 31, 68, 245, 72, 211, 160, 234, 70, 242, 127, 136, 246, 239, 225, 116, 104, 191, 190, 32, 81, 126, 247, 126, 13, 77, 170, 61, 127, 188, 16, 82, 0, 25, 0, 178, 160, 103, 123, 72, 117, 117, 215, 247, 192, 229, 24, 221, 28, 127, 26, 188, 5, 130, 122, 28, 76, 197, 216, 203, 177, 165, 246, 192, 78, 87, 13, 99, 177, 14, 78, 72, 115, 118})
	hProof.stmt = new(OneOutOfManyStatement)
	hProof.stmt.Set(commitments)

	for i := 0; i < 3; i++ {
		privacy.Logger.Log.Infof("cl: %v: %v\n", i, hProof.cl[i].Compress())
		privacy.Logger.Log.Infof("ca: %v: %v\n", i, hProof.ca[i].Compress())
		privacy.Logger.Log.Infof("cb: %v: %v\n", i, hProof.cb[i].Compress())

		privacy.Logger.Log.Infof("cd point: %v: %+v\n", i, hProof.cd[i].X.Bytes())
		privacy.Logger.Log.Infof("cd: %v: %v\n", i, hProof.cd[i].Compress())

		privacy.Logger.Log.Infof("f: %v: %v\n", i, hProof.f[i].Bytes())
		privacy.Logger.Log.Infof("za: %v: %v\n", i, hProof.za[i].Bytes())
		privacy.Logger.Log.Infof("zb: %v: %v\n", i, hProof.zb[i].Bytes())

		privacy.Logger.Log.Infof("commitment: %v: %v\n\n", i, hProof.stmt.commitments[i].Compress())

	}
	privacy.Logger.Log.Infof("zd: %v\n", hProof.zd.Bytes())

	res2 := hProof.Verify()
	assert.Equal(t, true, res2)
}

func TestGetCoefficient(t *testing.T) {

	a := make([]*big.Int, 3)

	a[0] = new(big.Int).SetBytes([]byte{28, 30, 162, 177, 161, 127, 119, 10, 195, 106, 31, 125, 252, 56, 111, 229, 236, 245, 202, 172, 27, 54, 110, 9, 9, 8, 56, 189, 248, 100, 190, 129})
	a[1] = new(big.Int).SetBytes([]byte{144, 245, 78, 232, 93, 155, 71, 49, 175, 154, 78, 81, 146, 120, 171, 74, 88, 99, 196, 61, 124, 156, 35, 55, 39, 22, 189, 111, 108, 236, 3, 131})
	a[2] = new(big.Int).SetBytes([]byte{224, 15, 114, 83, 56, 148, 202, 7, 187, 99, 242, 4, 2, 168, 169, 168, 44, 174, 215, 111, 119, 162, 172, 44, 225, 97, 236, 240, 242, 233, 148, 49})

	res := GetCoefficient([]byte{0, 1, 1}, 3, 3, a, []byte{0, 1, 1})

	//expectedRes := big.NewInt(-6)
	//expectedRes.Mod(expectedRes, privacy.Curve.Params().N)
	privacy.Logger.Log.Infof("res: %v\n", res.Bytes())

	//assert.Equal(t, expectedRes, res)
}

func TestCd(t *testing.T) {

	indexIsZeroBinary := privacy.ConvertIntToBinary(3, 3)

	commitments := make([]*privacy.EllipticPoint, privacy.CMRingSize)
	commitments[0] = new(privacy.EllipticPoint)
	commitments[0].Decompress([]byte{2, 63, 242, 198, 114, 250, 36, 102, 85, 80, 173, 148, 153, 247, 78, 215, 30, 54, 40, 193, 40, 190, 206, 73, 198, 39, 23, 48, 56, 136, 58, 91, 167})

	commitments[1] = new(privacy.EllipticPoint)
	commitments[1].Decompress([]byte{2, 203, 30, 129, 126, 123, 135, 125, 29, 43, 137, 52, 148, 146, 17, 87, 85, 237, 67, 191, 175, 241, 86, 102, 239, 183, 114, 78, 11, 127, 116, 16, 143})

	commitments[2] = new(privacy.EllipticPoint)
	commitments[2].Decompress([]byte{2, 123, 251, 169, 31, 79, 237, 122, 212, 173, 208, 175, 20, 111, 140, 19, 185, 72, 17, 229, 163, 84, 255, 63, 157, 51, 251, 209, 160, 122, 250, 30, 116})

	commitments[3] = new(privacy.EllipticPoint)
	commitments[3].Decompress([]byte{2, 174, 247, 205, 128, 120, 191, 95, 219, 186, 227, 95, 10, 157, 200, 224, 109, 152, 179, 5, 188, 162, 125, 167, 214, 127, 178, 173, 246, 109, 18, 23, 254})

	commitments[4] = new(privacy.EllipticPoint)
	commitments[4].Decompress([]byte{2, 8, 49, 76, 243, 238, 108, 171, 35, 55, 118, 239, 95, 214, 43, 88, 155, 4, 152, 62, 74, 15, 62, 203, 158, 189, 163, 62, 150, 255, 220, 14, 170})

	commitments[5] = new(privacy.EllipticPoint)
	commitments[5].Decompress([]byte{3, 205, 17, 244, 179, 44, 154, 114, 20, 78, 113, 196, 20, 133, 98, 165, 111, 74, 139, 53, 74, 224, 153, 41, 66, 224, 190, 220, 179, 136, 193, 241, 218})

	commitments[6] = new(privacy.EllipticPoint)
	commitments[6].Decompress([]byte{3, 191, 145, 66, 202, 76, 92, 64, 185, 89, 85, 149, 239, 190, 231, 208, 214, 25, 0, 218, 142, 114, 18, 188, 122, 111, 213, 6, 108, 128, 129, 122, 109})

	commitments[7] = new(privacy.EllipticPoint)
	commitments[7].Decompress([]byte{2, 186, 15, 36, 170, 79, 9, 118, 9, 249, 10, 215, 114, 5, 80, 9, 156, 206, 217, 242, 156, 30, 210, 169, 109, 221, 103, 37, 186, 24, 88, 47, 121})

	commitments[3] = privacy.PedCom.CommitAtIndex(big.NewInt(0), big.NewInt(100), privacy.SK)

	u := make([]*big.Int, 3)
	u[0] = new(big.Int).SetBytes([]byte{115, 84, 227, 241, 75, 118, 237, 82, 97, 108, 155, 112, 54, 220, 78, 198, 225, 208, 171, 130, 246, 237, 35, 27, 35, 151, 155, 0, 185, 179, 85, 177})
	u[1] = new(big.Int).SetBytes([]byte{253, 131, 9, 69, 18, 229, 241, 89, 154, 121, 114, 55, 195, 14, 140, 18, 86, 221, 247, 122, 215, 152, 89, 163, 219, 18, 201, 182, 219, 36, 1, 186})
	u[2] = new(big.Int).SetBytes([]byte{164, 247, 85, 18, 97, 44, 48, 254, 16, 39, 92, 150, 194, 113, 110, 241, 149, 134, 133, 139, 181, 73, 194, 167, 207, 46, 172, 12, 126, 197, 227, 64})

	a := make([]*big.Int, 3)

	a[0] = new(big.Int).SetBytes([]byte{164, 164, 236, 47, 101, 132, 122, 99, 215, 140, 145, 0, 228, 144, 12, 101, 204, 70, 147, 3, 210, 131, 51, 8, 83, 123, 63, 162, 156, 144, 212, 64})
	a[1] = new(big.Int).SetBytes([]byte{100, 84, 45, 170, 63, 89, 221, 197, 41, 18, 146, 22, 144, 213, 176, 102, 28, 171, 216, 195, 24, 136, 195, 190, 199, 223, 101, 151, 11, 136, 150, 137})
	a[2] = new(big.Int).SetBytes([]byte{119, 150, 136, 147, 232, 252, 45, 188, 140, 27, 139, 171, 10, 237, 146, 229, 60, 117, 231, 20, 198, 32, 98, 28, 10, 97, 189, 107, 230, 121, 55, 76})

	cd := make([]*privacy.EllipticPoint, 3)
	for k := 0; k < 3; k++ {
		// Calculate pi,k which is coefficient of x^k in polynomial pi(x)
		cd[k] = new(privacy.EllipticPoint).Zero()

		for i := 0; i < 8; i++ {
			privacy.Logger.Log.Infof("k : %v\n", k)
			privacy.Logger.Log.Infof("i : %v\n", i)

			iBinary := privacy.ConvertIntToBinary(i, 3)
			privacy.Logger.Log.Infof("iBinary : %v\n", iBinary)
			//privacy.Logger.Log.Infof("n : %v\n", n)
			privacy.Logger.Log.Infof("a0 : %v\n", a[0].Bytes())
			privacy.Logger.Log.Infof("a1 : %v\n", a[1].Bytes())
			privacy.Logger.Log.Infof("a2 : %v\n", a[2].Bytes())
			privacy.Logger.Log.Infof("indexIsZeroBinary : %v\n", indexIsZeroBinary)

			pik := GetCoefficient(iBinary, k, 3, a, indexIsZeroBinary)
			privacy.Logger.Log.Infof("pik: %v: %v\n", i, pik.Bytes())

			cd[k] = cd[k].Add(commitments[i].ScalarMult(pik))
			fmt.Println()
			//privacy.Logger.Log.Infof("cd %v: %v\n", i, cd[k].Compress())
		}

		//fmt.Println()

		cd[k] = cd[k].Add(privacy.PedCom.CommitAtIndex(big.NewInt(0), u[k], privacy.SK))

		privacy.Logger.Log.Infof("cd %v: %v\n", k, cd[k].Compress())
	}
}
