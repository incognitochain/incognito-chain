package zkp

import (
	"fmt"
	"github.com/ninjadotorg/constant/privacy"
	"math/big"
	"testing"
	"github.com/stretchr/testify/assert"
	"time"
)

//TestPKOneOfMany test protocol for one of many Commitment is Commitment to zero
func TestPKOneOfMany(t *testing.T) {
	witness := new(OneOutOfManyWitness)

	indexIsZero := 2

	// list of commitments
	commitments := make([]*privacy.EllipticPoint, privacy.CMRingSize)
	snDerivators := make([]*big.Int, privacy.CMRingSize)
	randoms := make([]*big.Int, privacy.CMRingSize)

	for i := 0; i < privacy.CMRingSize; i++ {
		snDerivators[i] = privacy.RandScalar()
		randoms[i] = privacy.RandScalar()
		commitments[i] = privacy.PedCom.CommitAtIndex(snDerivators[i], randoms[i], privacy.SND)
	}

	// create Commitment to zero at indexIsZero
	snDerivators[indexIsZero] = big.NewInt(0)
	commitments[indexIsZero] = privacy.PedCom.CommitAtIndex(snDerivators[indexIsZero], randoms[indexIsZero], privacy.SND)

	witness.Set(commitments, randoms[indexIsZero], uint64(indexIsZero))
	start := time.Now()
	proof, err := witness.Prove()
	end := time.Since(start)
	fmt.Printf("One out of many proving time: %v\n", end)
	if err != nil {
		privacy.Logger.Log.Error(err)
	}

	//Convert proof to bytes array
	proofBytes := proof.Bytes()

	fmt.Printf("One out of many proof size: %v\n", len(proofBytes))

	//revert bytes array to proof
	proof2 := new(OneOutOfManyProof).Init()
	proof2.SetBytes(proofBytes)
	proof2.stmt.commitments = commitments

	start = time.Now()

	res := proof.Verify()
	end = time.Since(start)
	fmt.Printf("One out of many verification time: %v\n", end)

	assert.Equal(t, true, res)

	//test for JS
	//hCommitments := make([]*privacy.EllipticPoint, privacy.CMRingSize)
	//hCommitments[0] = new(privacy.EllipticPoint)
	//hCommitments[0].Decompress([]byte{2, 63, 242, 198, 114, 250, 36, 102, 85, 80, 173, 148, 153, 247, 78, 215, 30, 54, 40, 193, 40, 190, 206, 73, 198, 39, 23, 48, 56, 136, 58, 91, 167})
	//
	//hCommitments[1] = new(privacy.EllipticPoint)
	//hCommitments[1].Decompress([]byte{2, 203, 30, 129, 126, 123, 135, 125, 29, 43, 137, 52, 148, 146, 17, 87, 85, 237, 67, 191, 175, 241, 86, 102, 239, 183, 114, 78, 11, 127, 116, 16, 143})
	//
	//hCommitments[2] = new(privacy.EllipticPoint)
	//hCommitments[2].Decompress([]byte{2, 123, 251, 169, 31, 79, 237, 122, 212, 173, 208, 175, 20, 111, 140, 19, 185, 72, 17, 229, 163, 84, 255, 63, 157, 51, 251, 209, 160, 122, 250, 30, 116})
	//
	//hCommitments[3] = new(privacy.EllipticPoint)
	//hCommitments[3].Decompress([]byte{2, 174, 247, 205, 128, 120, 191, 95, 219, 186, 227, 95, 10, 157, 200, 224, 109, 152, 179, 5, 188, 162, 125, 167, 214, 127, 178, 173, 246, 109, 18, 23, 254})
	//
	//hCommitments[4] = new(privacy.EllipticPoint)
	//hCommitments[4].Decompress([]byte{2, 8, 49, 76, 243, 238, 108, 171, 35, 55, 118, 239, 95, 214, 43, 88, 155, 4, 152, 62, 74, 15, 62, 203, 158, 189, 163, 62, 150, 255, 220, 14, 170})
	//
	//hCommitments[5] = new(privacy.EllipticPoint)
	//hCommitments[5].Decompress([]byte{3, 205, 17, 244, 179, 44, 154, 114, 20, 78, 113, 196, 20, 133, 98, 165, 111, 74, 139, 53, 74, 224, 153, 41, 66, 224, 190, 220, 179, 136, 193, 241, 218})
	//
	//hCommitments[6] = new(privacy.EllipticPoint)
	//hCommitments[6].Decompress([]byte{3, 191, 145, 66, 202, 76, 92, 64, 185, 89, 85, 149, 239, 190, 231, 208, 214, 25, 0, 218, 142, 114, 18, 188, 122, 111, 213, 6, 108, 128, 129, 122, 109})
	//
	//hCommitments[7] = new(privacy.EllipticPoint)
	//hCommitments[7].Decompress([]byte{2, 186, 15, 36, 170, 79, 9, 118, 9, 249, 10, 215, 114, 5, 80, 9, 156, 206, 217, 242, 156, 30, 210, 169, 109, 221, 103, 37, 186, 24, 88, 47, 121})
	//
	//hCommitments[3] = privacy.PedCom.CommitAtIndex(big.NewInt(0), big.NewInt(100), privacy.SK)
	//
	////hWit := new(OneOutOfManyWitness)
	////hWit.Set(hCommitments, big.NewInt(100), 3)
	////hProof1, _ := hWit.Prove()
	////res := hProof1.Verify()
	////fmt.Println(res)
	//
	//hProof := new(OneOutOfManyProof).Init()
	//hProof.SetBytes([]byte{3, 33, 59, 108, 16, 125, 209, 19, 158, 64, 30, 253, 139, 106, 153, 100, 236, 67, 31, 116, 106, 131, 129, 182, 220, 66, 215, 2, 43, 102, 153, 183, 178, 3, 216, 204, 16, 91, 194, 133, 111, 183, 46, 117, 180, 37, 175, 63, 253, 180, 28, 194, 154, 30, 191, 13, 42, 75, 16, 244, 193, 102, 12, 69, 251, 125, 2, 2, 31, 136, 46, 223, 130, 107, 77, 42, 146, 163, 203, 59, 14, 137, 63, 120, 69, 42, 157, 239, 60, 199, 11, 185, 238, 152, 111, 54, 255, 190, 241, 2, 65, 38, 87, 26, 176, 109, 107, 211, 117, 156, 26, 11, 80, 235, 60, 89, 78, 105, 180, 14, 165, 188, 214, 96, 246, 209, 147, 90, 105, 98, 204, 57, 3, 99, 97, 67, 131, 213, 19, 134, 215, 143, 54, 115, 172, 21, 38, 80, 166, 147, 174, 146, 189, 0, 91, 21, 98, 113, 220, 63, 85, 169, 107, 126, 205, 2, 237, 183, 33, 80, 38, 238, 8, 73, 104, 158, 107, 244, 248, 97, 180, 151, 49, 91, 47, 121, 152, 201, 99, 137, 252, 104, 74, 57, 158, 82, 45, 177, 2, 31, 232, 234, 213, 125, 83, 103, 106, 142, 121, 189, 204, 64, 135, 88, 251, 87, 12, 138, 166, 30, 181, 114, 195, 141, 25, 169, 73, 53, 17, 51, 181, 2, 205, 162, 172, 130, 207, 165, 13, 217, 143, 70, 34, 92, 200, 18, 193, 114, 232, 90, 244, 161, 198, 198, 143, 252, 15, 36, 254, 36, 255, 90, 195, 8, 2, 64, 43, 33, 238, 77, 159, 95, 147, 70, 35, 151, 155, 107, 223, 212, 164, 232, 147, 12, 177, 112, 248, 42, 4, 29, 41, 128, 175, 12, 233, 208, 207, 3, 111, 197, 49, 7, 95, 148, 25, 58, 50, 1, 110, 80, 14, 251, 132, 72, 160, 67, 196, 250, 57, 108, 254, 78, 237, 76, 23, 245, 193, 13, 158, 225, 2, 6, 193, 44, 27, 209, 98, 85, 43, 238, 85, 116, 95, 223, 228, 21, 193, 93, 104, 166, 111, 149, 176, 14, 152, 182, 53, 249, 112, 98, 188, 109, 163, 3, 73, 145, 149, 118, 206, 37, 140, 75, 168, 93, 120, 53, 39, 64, 168, 150, 59, 179, 114, 210, 64, 68, 6, 209, 33, 143, 0, 249, 17, 219, 192, 71, 146, 86, 236, 160, 161, 202, 128, 254, 4, 118, 29, 204, 36, 247, 14, 64, 55, 145, 167, 231, 183, 129, 113, 206, 139, 145, 58, 251, 151, 95, 156, 110, 136, 184, 249, 27, 177, 150, 128, 219, 126, 165, 184, 244, 143, 229, 252, 152, 120, 136, 142, 241, 255, 89, 33, 208, 56, 137, 168, 91, 3, 44, 48, 176, 75, 141, 138, 240, 3, 151, 231, 91, 182, 88, 142, 163, 13, 212, 52, 138, 133, 217, 220, 148, 38, 73, 157, 60, 199, 92, 111, 219, 226, 218, 140, 26, 240, 142, 178, 36, 21, 27, 96, 178, 106, 102, 7, 15, 19, 33, 24, 21, 79, 90, 140, 88, 26, 91, 233, 227, 96, 153, 46, 156, 63, 228, 167, 130, 96, 180, 147, 129, 141, 58, 219, 78, 162, 191, 217, 2, 8, 115, 201, 227, 104, 60, 175, 63, 14, 169, 69, 149, 118, 55, 19, 93, 134, 17, 129, 55, 59, 208, 233, 22, 171, 31, 127, 204, 58, 100, 156, 194, 86, 218, 196, 93, 131, 79, 27, 145, 72, 73, 222, 15, 248, 7, 72, 79, 183, 157, 113, 110, 229, 232, 18, 89, 212, 52, 201, 158, 79, 15, 178, 6, 228, 244, 249, 102, 231, 78, 140, 176, 31, 63, 114, 147, 12, 177, 225, 95, 188, 133, 69, 49, 212, 145, 178, 129, 203, 76, 105, 142, 240, 82, 161, 110, 47, 8, 210, 157, 42, 170, 46, 249, 184, 30, 49, 173, 44, 14, 150, 97, 79, 1, 7, 67, 77, 190, 186, 78, 72, 99, 66, 42, 176, 150, 3, 20, 23, 48, 93, 7, 71, 105, 220, 40, 132, 160, 115, 93, 62, 127, 89, 75, 149, 131, 150, 65, 195, 240, 102, 126, 173, 81, 157, 254, 156, 201, 113, 24, 161, 46, 24, 59, 58, 237, 175, 142, 219, 253, 57, 137, 236, 53, 83, 218, 59, 12, 145, 206})
	//hProof.stmt = new(OneOutOfManyStatement)
	//hProof.stmt.Set(hCommitments)
	//
	//
	//for i:=0; i<3; i++{
	//	fmt.Printf("cl: %v: %v\n", i, hProof.cl[i].Compress())
	//	fmt.Printf("ca: %v: %v\n", i, hProof.ca[i].Compress())
	//	fmt.Printf("cb: %v: %v\n", i, hProof.cb[i].Compress())
	//
	//	fmt.Printf("cd point: %v: %+v\n", i, hProof.cd[i].X.Bytes())
	//	fmt.Printf("cd: %v: %v\n", i, hProof.cd[i].Compress())
	//
	//	fmt.Printf("f: %v: %v\n", i, hProof.f[i].Bytes())
	//	fmt.Printf("za: %v: %v\n", i, hProof.za[i].Bytes())
	//	fmt.Printf("zb: %v: %v\n", i, hProof.zb[i].Bytes())
	//
	//	fmt.Printf("commitment: %v: %v\n\n", i, hProof.stmt.commitments[i].Compress())
	//
	//}
	//fmt.Printf("zd: %v\n", hProof.zd.Bytes())
	//
	//res2 := hProof.Verify()
	//assert.Equal(t, true, res2)
}


func TestGetCoefficient(t *testing.T) {

	a := make([]*big.Int, 3)

	a[0] = new(big.Int).SetBytes([]byte{28, 30, 162, 177, 161, 127, 119, 10, 195, 106, 31, 125, 252, 56, 111, 229, 236, 245, 202, 172, 27, 54, 110, 9, 9, 8, 56, 189, 248, 100, 190, 129})
	a[1] = new(big.Int).SetBytes([]byte{144, 245, 78, 232, 93, 155, 71, 49, 175, 154, 78, 81, 146, 120, 171, 74, 88, 99, 196, 61, 124, 156, 35, 55, 39, 22, 189, 111, 108, 236, 3, 131})
	a[2] = new(big.Int).SetBytes([]byte{224, 15, 114, 83, 56, 148, 202, 7, 187, 99, 242, 4, 2, 168, 169, 168, 44, 174, 215, 111, 119, 162, 172, 44, 225, 97, 236, 240, 242, 233, 148, 49})

	res := GetCoefficient([]byte{0,1,1}, 3, 3, a, []byte{0,1,1})

	//expectedRes := big.NewInt(-6)
	//expectedRes.Mod(expectedRes, privacy.Curve.Params().N)
	fmt.Printf("res: %v\n", res.Bytes())

	//assert.Equal(t, expectedRes, res)
}


func TestCd(t *testing.T){

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
			fmt.Printf("k : %v\n", k)
			fmt.Printf("i : %v\n", i)

			iBinary := privacy.ConvertIntToBinary(i, 3)
			fmt.Printf("iBinary : %v\n", iBinary)
			//fmt.Printf("n : %v\n", n)
			fmt.Printf("a0 : %v\n", a[0].Bytes())
			fmt.Printf("a1 : %v\n", a[1].Bytes())
			fmt.Printf("a2 : %v\n", a[2].Bytes())
			fmt.Printf("indexIsZeroBinary : %v\n", indexIsZeroBinary)

			pik := GetCoefficient(iBinary, k, 3, a, indexIsZeroBinary)
			fmt.Printf("pik: %v: %v\n", i, pik.Bytes())

			cd[k] = cd[k].Add(commitments[i].ScalarMult(pik))
			fmt.Println()
			//fmt.Printf("cd %v: %v\n", i, cd[k].Compress())
		}

		//fmt.Println()

		cd[k] = cd[k].Add(privacy.PedCom.CommitAtIndex(big.NewInt(0), u[k], privacy.SK))

		fmt.Printf("cd %v: %v\n", k, cd[k].Compress())
	}
}
