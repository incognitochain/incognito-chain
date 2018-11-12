package main

import(
	"github.com/ninjadotorg/constant/privacy"
)

func main() {

	// fmt.Printf("N: %X\n", privacy.Curve.Params().N)
	//fmt.Printf("P: %X\n", privacy.Curve.Params().P)
	// fmt.Printf("B: %X\n", privacy.Curve.Params().B)
	// fmt.Printf("Gx: %x\n", privacy.Curve.Params().Gx)
	// fmt.Printf("Gy: %X\n", privacy.Curve.Params().Gy)
	// fmt.Printf("BitSize: %X\n", privacy.Curve.Params().BitSize)

	//spendingKey := privacy.GenerateSpendingKey(new(big.Int).SetInt64(123).Bytes())
	//fmt.Printf("\nSpending key: %v\n", spendingKey)
	//fmt.Println(len(spendingKey))
	//
	//address := privacy.GeneratePublicKey(spendingKey)
	//fmt.Printf("\nAddress: %v\n", address)
	//fmt.Println(len(address))
	//point, err := privacy.DecompressKey(address)
	//if err != nil {
	//	fmt.Println(err)
	//}
	//fmt.Printf("Pk decom: %v\n", point)
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
	//
	//msg := "hello, world"
	//hash := sha256.Sum256([]byte(msg))
	//
	//signature, err := privacy.Sign(hash[:], spendingKey)
	//if err != nil {
	//	panic(err)
	//}
	//fmt.Printf("signature: %v\n", signature)
	//
	//valid := privacy.Verify(signature, hash[:], address)
	//fmt.Println("\nsignature verified:", valid)
	//
	//tx, _ := transaction.CreateEmptyTxs()
	//fmt.Printf("Transaction: %+v\n", tx)

	//privacy.Pcm.InitCommitment()
	//privacy.TestProofIsZero()
	//// fmt.Println("Done")
	//a:= new(privacy.InputCommitments)
	//
	//privacy.TestPKComZeroOne()
	//privacy.TestProductCommitment()
	//privacy.Pcm.InitCommitment()
	// privacy.TestProofIsZero()
	// fmt.Println("Done")
	// privacy.TestPKComZeroOne()
	//privacy.TestPKOneOfMany()

	//type Proof interface{
	//	InitProof()
	//}
	//
	//type Protocol interface{
	//	Prove() Proof
	//}
	//
	//type Proof1 struct{
	//	a int
	//}
	//
	//func (p*Proof1) InitProof() {
	//
	//}

	//var p1, p2 privacy.Poly
	//p1 = privacy.RandomPoly(1, 3)
	//fmt.Println(p1.String())
	//p2 = privacy.RandomPoly(1, 3)
	//fmt.Println(p2.String())
	//p3 := p1.Mul(p2, nil)
	//fmt.Println(p3.String())

	//privacy.TestSchn()

	privacy.TestPKOneOfMany()


}
