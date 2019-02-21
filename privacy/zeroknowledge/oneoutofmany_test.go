package zkp

import (
	"fmt"
	"github.com/ninjadotorg/constant/privacy"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
)

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
	//fmt.Printf("One out of many proving time: %v\n", end)
	//if err != nil {
	//	privacy.Logger.Log.Error(err)
	//}
	//
	////Convert proof to bytes array
	//proofBytes := proof.Bytes()
	//
	//fmt.Printf("One out of many proof size: %v\n", len(proofBytes))
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
	//fmt.Printf("One out of many verification time: %v\n", end)
	//
	//assert.Equal(t, true, res)

	//test for JS
	hCommitments := make([]*privacy.EllipticPoint, privacy.CMRingSize)
	hCommitments[0] = new(privacy.EllipticPoint)
	hCommitments[0].Decompress([]byte{3, 225, 145, 19, 109, 28, 119, 32, 54, 143, 94, 108, 235, 211, 15, 196, 220, 174, 19, 96, 115, 86, 50, 41, 0, 121, 123, 23, 241, 231, 126, 44, 124})

	hCommitments[1] = new(privacy.EllipticPoint)
	hCommitments[1].Decompress([]byte{3, 61, 175, 99, 143, 220, 215, 180, 105, 134, 244, 90, 204, 108, 93, 0, 176, 93, 250, 102, 73, 91, 69, 42, 125, 7, 227, 35, 176, 30, 91, 117, 173})

	hCommitments[2] = new(privacy.EllipticPoint)
	hCommitments[2].Decompress([]byte{3, 131, 74, 242, 235, 166, 189, 189, 108, 197, 76, 104, 145, 18, 180, 201, 70, 168, 40, 234, 56, 128, 109, 46, 96, 6, 198, 207, 10, 148, 69, 175, 22})

	hCommitments[3] = new(privacy.EllipticPoint)
	hCommitments[3].Decompress([]byte{3, 158, 17, 107, 212, 6, 7, 198, 47, 35, 121, 78, 205, 225, 89, 10, 212, 110, 80, 42, 65, 84, 208, 177, 158, 216, 213, 252, 6, 186, 47, 112, 71})

	hCommitments[4] = new(privacy.EllipticPoint)
	hCommitments[4].Decompress([]byte{2, 167, 93, 55, 90, 102, 118, 80, 111, 118, 145, 48, 78, 181, 214, 249, 81, 56, 89, 106, 173, 41, 87, 104, 249, 85, 103, 245, 115, 48, 171, 50, 166})

	hCommitments[5] = new(privacy.EllipticPoint)
	hCommitments[5].Decompress([]byte{2, 232, 218, 17, 250, 193, 207, 155, 20, 217, 68, 12, 192, 177, 238, 240, 181, 138, 216, 226, 242, 138, 62, 8, 146, 231, 7, 121, 178, 19, 147, 207, 181})

	hCommitments[6] = new(privacy.EllipticPoint)
	hCommitments[6].Decompress([]byte{3, 33, 30, 16, 170, 189, 229, 38, 67, 132, 147, 243, 195, 60, 150, 70, 185, 41, 149, 150, 113, 19, 195, 31, 93, 7, 153, 30, 232, 255, 202, 28, 31})

	hCommitments[7] = new(privacy.EllipticPoint)
	hCommitments[7].Decompress([]byte{3, 126, 80, 209, 45, 86, 216, 198, 49, 250, 206, 154, 70, 142, 41, 160, 173, 116, 55, 253, 120, 171, 120, 243, 154, 243, 211, 223, 39, 11, 227, 247, 4})

	//hCommitments[3] = privacy.PedCom.CommitAtIndex(big.NewInt(0), big.NewInt(100), privacy.SK)

	//hWit := new(OneOutOfManyWitness)
	//hWit.Set(hCommitments, big.NewInt(100), 3)
	//hProof1, _ := hWit.Prove()
	//res := hProof1.Verify()
	//fmt.Println(res)

	hProof := new(OneOutOfManyProof).Init()
	hProof.SetBytes([]byte{3, 10, 121, 95, 78, 228, 21, 63, 151, 84, 79, 225, 211, 30, 203, 82, 111, 142, 241, 235, 42, 108, 141, 248, 161, 241, 115, 53, 100, 40, 45, 112, 194, 3, 131, 61, 152, 228, 205, 102, 237, 22, 202, 109, 50, 15, 190, 230, 55, 120, 139, 229, 111, 133, 28, 177, 236, 38, 71, 65, 33, 194, 64, 136, 215, 214, 3, 133, 104, 134, 214, 98, 148, 183, 119, 140, 7, 65, 145, 91, 224, 202, 153, 12, 118, 98, 88, 174, 114, 72, 181, 169, 223, 15, 134, 2, 135, 46, 209, 2, 126, 167, 113, 1, 102, 147, 101, 153, 160, 135, 253, 122, 85, 66, 184, 115, 5, 204, 102, 44, 95, 21, 205, 144, 6, 129, 103, 43, 122, 143, 216, 236, 3, 221, 24, 216, 125, 213, 209, 140, 242, 140, 93, 33, 105, 123, 29, 212, 139, 40, 146, 249, 236, 178, 127, 9, 164, 120, 170, 245, 105, 46, 49, 82, 252, 3, 158, 34, 3, 112, 119, 73, 246, 172, 141, 19, 12, 110, 87, 71, 155, 157, 81, 251, 25, 253, 216, 37, 224, 52, 211, 129, 125, 165, 41, 203, 97, 242, 2, 205, 159, 146, 82, 169, 30, 106, 125, 173, 246, 239, 176, 202, 169, 51, 22, 61, 96, 23, 158, 43, 38, 38, 188, 230, 68, 133, 38, 243, 15, 245, 88, 2, 60, 217, 13, 228, 249, 54, 123, 47, 250, 167, 68, 196, 254, 253, 255, 144, 204, 112, 23, 20, 61, 112, 39, 178, 167, 49, 19, 250, 236, 168, 8, 24, 3, 75, 131, 79, 68, 32, 107, 84, 199, 249, 210, 179, 5, 105, 32, 16, 248, 12, 26, 65, 80, 223, 227, 250, 204, 66, 248, 109, 237, 149, 31, 156, 211, 3, 16, 248, 221, 19, 149, 39, 237, 211, 53, 249, 51, 10, 159, 42, 157, 134, 212, 174, 128, 246, 107, 133, 197, 198, 156, 202, 37, 94, 105, 229, 205, 184, 2, 92, 140, 106, 168, 113, 11, 194, 217, 102, 224, 70, 246, 51, 224, 112, 123, 236, 94, 222, 3, 89, 34, 245, 57, 76, 154, 243, 172, 143, 38, 212, 229, 2, 186, 6, 12, 145, 94, 70, 158, 184, 43, 5, 68, 222, 95, 89, 78, 107, 61, 147, 107, 23, 45, 151, 244, 199, 168, 63, 92, 59, 66, 78, 234, 246, 134, 10, 31, 108, 58, 138, 81, 160, 125, 0, 158, 223, 107, 50, 76, 150, 13, 105, 209, 156, 102, 221, 159, 223, 140, 115, 134, 178, 199, 13, 103, 62, 255, 44, 102, 142, 215, 228, 194, 136, 230, 225, 21, 68, 39, 37, 151, 255, 220, 12, 225, 15, 9, 136, 103, 155, 171, 118, 221, 6, 133, 173, 136, 140, 6, 199, 115, 244, 61, 119, 37, 7, 36, 24, 231, 233, 114, 138, 166, 145, 98, 20, 221, 108, 32, 16, 30, 56, 121, 113, 199, 65, 121, 125, 120, 37, 158, 88, 77, 228, 16, 46, 95, 183, 168, 133, 255, 242, 148, 153, 28, 86, 162, 146, 73, 93, 188, 145, 176, 42, 175, 162, 188, 34, 209, 68, 68, 50, 135, 190, 0, 207, 240, 193, 13, 116, 215, 217, 234, 66, 167, 238, 161, 4, 27, 245, 69, 18, 83, 238, 32, 105, 33, 184, 123, 130, 176, 18, 208, 198, 121, 247, 151, 32, 177, 155, 133, 139, 96, 125, 138, 125, 110, 184, 234, 29, 208, 8, 65, 234, 238, 73, 60, 221, 212, 5, 184, 200, 12, 90, 165, 252, 112, 95, 10, 14, 187, 33, 30, 212, 208, 236, 199, 42, 19, 144, 88, 99, 34, 97, 98, 66, 242, 169, 49, 218, 23, 205, 134, 178, 96, 34, 217, 211, 24, 75, 121, 130, 128, 210, 10, 167, 139, 181, 189, 197, 2, 32, 172, 71, 193, 68, 132, 75, 164, 210, 94, 12, 75, 124, 157, 231, 170, 194, 255, 99, 183, 143, 237, 115, 91, 76, 87, 207, 171, 188, 177, 120, 154, 63, 111, 165, 187, 47, 254, 176, 94, 43, 73, 183, 243, 244, 126, 180, 253, 101, 249, 207, 28, 205, 254, 217, 194, 200, 151, 21, 0, 62, 155, 107, 58, 83, 89, 240, 93, 225, 232, 106, 0, 2, 55, 84, 144, 139, 255, 175, 231, 67, 170, 167})
	hProof.stmt = new(OneOutOfManyStatement)
	hProof.stmt.Set(hCommitments)


	for i:=0; i<3; i++{
		fmt.Printf("cl: %v: %v\n", i, hProof.cl[i].Compress())
		fmt.Printf("ca: %v: %v\n", i, hProof.ca[i].Compress())
		fmt.Printf("cb: %v: %v\n", i, hProof.cb[i].Compress())

		fmt.Printf("cd point: %v: %+v\n", i, hProof.cd[i].X.Bytes())
		fmt.Printf("cd: %v: %v\n", i, hProof.cd[i].Compress())

		fmt.Printf("f: %v: %v\n", i, hProof.f[i].Bytes())
		fmt.Printf("za: %v: %v\n", i, hProof.za[i].Bytes())
		fmt.Printf("zb: %v: %v\n", i, hProof.zb[i].Bytes())

		fmt.Printf("commitment: %v: %v\n\n", i, hProof.stmt.commitments[i].Compress())

	}
	fmt.Printf("zd: %v\n", hProof.zd.Bytes())

	res2 := hProof.Verify()
	assert.Equal(t, true, res2)
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
