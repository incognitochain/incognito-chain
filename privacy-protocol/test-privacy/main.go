package main

import (
	//"fmt"
	"github.com/ninjadotorg/constant/privacy-protocol"
	//"math/big"

	privacy "github.com/ninjadotorg/constant/privacy-protocol"
)

func main() {

	// fmt.Printf("N: %X\n", privacy-protocol.Curve.Params().N)
	//fmt.Printf("P: %X\n", privacy-protocol.Curve.Params().P)
	// fmt.Printf("B: %X\n", privacy-protocol.Curve.Params().B)
	// fmt.Printf("Gx: %x\n", privacy-protocol.Curve.Params().Gx)
	// fmt.Printf("Gy: %X\n", privacy-protocol.Curve.Params().Gy)
	// fmt.Printf("BitSize: %X\n", privacy-protocol.Curve.Params().BitSize)

	/*---------------------- TEST KEY SET ----------------------*/
	//spendingKey := privacy.GenerateSpendingKey(new(big.Int).SetInt64(123).Bytes())
	//fmt.Printf("\nSpending key: %v\n", spendingKey)
	//fmt.Println(len(spendingKey))

	// publicKey is compressed
	//publicKey := privacy.GeneratePublicKey(spendingKey)
	//fmt.Printf("\nPublic key: %v\n", publicKey)
	//fmt.Printf("Len public key: %v\n", len(publicKey))
	//point, err := privacy.DecompressKey(publicKey)
	//if err != nil {
	//fmt.Println(err)
	//}
	//fmt.Printf("Public key decompress: %v\n", point)
	//
	//receivingKey := privacy.GenerateReceivingKey(spendingKey)
	//fmt.Printf("\nReceiving key: %v\n", receivingKey)
	//fmt.Println(len(receivingKey))
	//
	//transmissionKey := privacy.GenerateTransmissionKey(receivingKey)
	//fmt.Printf("\nTransmission key: %v\n", transmissionKey)
	//fmt.Println(len(transmissionKey))
	//
	//point, err = privacy.DecompressKey(transmissionKey)
	//if err != nil {
	//	fmt.Println(err)
	//}
	//fmt.Printf("Transmission key point decompress: %+v\n ", point)

	//paymentAddress := privacy.GeneratePaymentAddress(spendingKey)
	//fmt.Println(paymentAddress.ToBytes())
	//fmt.Printf("tk: %v\n", paymentAddress.Tk)
	//fmt.Printf("pk: %v\n", paymentAddress.Pk)
	//
	//fmt.Printf("spending key bytes: %v\n", spendingKey.String())

	/*---------------------- TEST ZERO KNOWLEDGE ----------------------*/

	//privacy.TestProofIsZero()

	//privacy.TestProductCommitment()

	// privacy.TestProofIsZero()

	// privacy.TestPKComZeroOne()

	// privacy.TestPKOneOfMany()

	//i := 0
	//runtime.GOMAXPROCS(runtime.NumCPU())
	//privacy.Elcm.InitCommitment()
	//n := 500
	//for i = 0; i < n; i++ {
	//
	//	//  zkp.TestPKComZeroOne()
	//
	//	//for i := 0; i < 500; i++ {
	//
	//	if !zkp.TestPKOneOfMany() {
	//		break
	//	}
	//	if !zkpoptimization.TestPKOneOfMany() {
	//		break
	//	}
	//
	//	fmt.Println("----------------------")
	//}
	////}
	//if i == n {
	//	fmt.Println("Well done")
	//} else {
	//	fmt.Println("ewww")
	//}

	//privacy.TestCommitment()

	//zkp.TestPKMaxValue()
	//privacy.Pcm.Setup()
	//privacy.Pcm.TestFunction(00)

	//var zk zkp.ZKProtocols
	//
	//valueRand := privacy.RandBytes(32)
	//vInt := new(big.Int).SetBytes(valueRagit rend)
	//vInt.Mod(vInt, big.NewInt(2))
	//rand := privacy.RandBytes(32)
	//
	//partialCommitment := privacy.Pcm.CommitAtIndex(vInt.Bytes(), rand, privacy.VALUE)
	//
	//
	//var witness zkp.PKComZeroOneWitness
	//witness.CommitedValue = vInt.Bytes()
	//witness.Rand = rand
	//witness.Commitment = partialCommitment
	//witness.Index = privacy.VALUE
	//
	//zk.SetWitness(witness)
	//proof, _ := zk.Prove()
	//zk.SetProof(proof)
	//fmt.Println(zk.Verify())
	//fmt.Printf("%v", privacy.TestECC())
	//spendingKey := privacy.GenerateSpendingKey(new(big.Int).SetInt64(123).Bytes())
	//
	//// publicKey is compressed
	//publicKey := privacy.GeneratePublicKey(spendingKey)
	//fmt.Printf("\nPublic key: %v\n", publicKey)
	//fmt.Printf("Len public key: %v\n", len(publicKey))
	//point := new(privacy.EllipticPoint)
	//point, err := privacy.DecompressKey(publicKey)
	//if err != nil {
	//	fmt.Println(err)
	//}
	//fmt.Printf("Public key decompress: %v %v\n", point.X.Bytes(), point.Y.Bytes())
	//fmt.Printf("\n %v\n", point.CompressPoint())
	//point, err = privacy.DecompressKey(point.CompressPoint())
	//if err != nil {
	//	fmt.Println(err)
	//}
	//fmt.Printf("Public key decompress: %v %v\n", point.X.Bytes(), point.Y.Bytes())
	//fmt.Printf("\n %v\n", point.CompressPoint())
	//zkp.TestPKComProduct()
	//privacy.Curve.Params().G

	//zkp.TestPKComZeroOne()

	/*----------------- TEST PCM SINGLETON -----------------*/
	//privacy.Pcm = privacy.GetPedersenParams()
	//fmt.Printf("a1: %p\n", &privacy.Pcm)
	//privacy.Pcm = privacy.GetPedersenParams()
	//fmt.Printf("a1: %p\n", &privacy.Pcm)

	/*----------------- TEST NEW PCM -----------------*/
	//var generators []privacy.EllipticPoint
	//generators := make([]privacy.EllipticPoint, 3)
	//generators[0] = privacy.EllipticPoint{big.NewInt(23), big.NewInt(0)}
	//generators[0].ComputeYCoord()
	//generators[1] = privacy.EllipticPoint{big.NewInt(12), big.NewInt(0)}
	//generators[1].ComputeYCoord()
	//generators[2] = privacy.EllipticPoint{big.NewInt(45), big.NewInt(0)}
	//generators[2].ComputeYCoord()
	//newPcm := privacy.NewPedersenParams(generators)
	//fmt.Printf("New Pcm: %+v\n", newPcm)

	/*----------------- TEST COMMITMENT -----------------*/
	//privacy.TestCommitment(01)

	/*----------------- TEST SIGNATURE -----------------*/
	privacy.TestSchn()

}
