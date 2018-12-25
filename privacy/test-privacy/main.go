package main

import (
	"github.com/ninjadotorg/constant/privacy"
)

func main() {

	// fmt.Printf("N: %X\n", privacy.Curve.Params().N)
	//fmt.Printf("P: %X\n", privacy.Curve.Params().P)
	// fmt.Printf("B: %X\n", privacy.Curve.Params().B)
	// fmt.Printf("Gx: %x\n", privacy.Curve.Params().Gx)
	// fmt.Printf("Gy: %X\n", privacy.Curve.Params().Gy)
	// fmt.Printf("BitSize: %X\n", privacy.Curve.Params().BitSize)

	/*---------------------- TEST KEY SET ----------------------*/


	/*---------------------- TEST ZERO KNOWLEDGE ----------------------*/

	//privacy.TestProofIsZero()

	//privacy.TestProductCommitment()

	// privacy.TestProofIsZero()

	/*****************zkp.TestPKComZeroOne()****************/

	//zkp.TestPKOneOfMany()

	//zkp.TestPKComMultiRange()

	//zkp.TestOpeningsProtocol()



	/*---------------------- TEST ZERO KNOWLEDGE ----------------------*/

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
	//privacy.PedCom.Setup()
	//privacy.PedCom.TestFunction(00)

	//var zk zkp.ZKProtocols
	//
	//valueRand := privacy.RandBytes(32)
	//vInt := new(big.Int).SetBytes(valueRagit rend)
	//vInt.Mod(vInt, big.NewInt(2))
	//rand := privacy.RandBytes(32)
	//
	//partialCommitment := privacy.PedCom.CommitAtIndex(vInt.Bytes(), rand, privacy.VALUE)
	//
	//
	//var witness zkp.PKComZeroOneWitness
	//witness.commitedValue = vInt.Bytes()
	//witness.rand = rand
	//witness.commitment = partialCommitment
	//witness.index = privacy.VALUE
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
	//privacy.PedCom = privacy.GetPedersenParams()
	//fmt.Printf("a1: %p\n", &privacy.PedCom)
	//privacy.PedCom = privacy.GetPedersenParams()
	//fmt.Printf("a1: %p\n", &privacy.PedCom)

	/*----------------- TEST NEW PCM -----------------*/
	//var generators []privacy.EllipticPoint
	//generators := make([]privacy.EllipticPoint, 3)
	//generators[0] = privacy.EllipticPoint{big.NewInt(23), big.NewInt(0)}
	//generators[0].ComputeYCoord()
	//generators[1] = privacy.EllipticPoint{big.NewInt(12), big.NewInt(0)}
	//generators[1].ComputeYCoord()
	//generators[2] = privacy.EllipticPoint{big.NewInt(45), big.NewInt(0)}
	//generators[2].ComputeYCoord()
	//newPedCom := privacy.NewPedersenParams(generators)
	//fmt.Printf("Zero PedCom: %+v\n", newPedCom)

	/*----------------- TEST COMMITMENT -----------------*/
	//privacy.TestCommitment(01)

	/*----------------- TEST SIGNATURE -----------------*/
	//privacy.TestSchn()
	//zkp.PKComMultiRangeTest()
	//privacy.TestMultiSig()

	/*----------------- TEST RANDOM WITH MAXIMUM VALUE -----------------*/
	//for i :=0; i<1000; i++{
	//	fmt.Printf("N: %v\n",privacy.Curve.Params().N)
	//	rand, _ := rand.Int(rand.Reader, privacy.Curve.Params().N)
	//
	//	fmt.Printf("rand: %v\n", rand)
	//	fmt.Printf("Len rand: %v\n", len(rand.Bytes()))
	//}

	/*----------------- TEST AES -----------------*/
	//privacy.TestAESCTR()

	/*----------------- TEST ENCRYPT/DECRYPT COIN -----------------*/

	//coin := new(privacy.OutputCoin)
	//coin.CoinDetails = new(privacy.Coin)
	//coin.CoinDetails.Randomness = privacy.RandInt()
	//coin.CoinDetails.Value = 10
	//
	//fmt.Printf("Plain text 1: Radnomness : %v\n", coin.CoinDetails.Randomness)
	//
	//spendingKey := privacy.GenerateSpendingKey(new(big.Int).SetInt64(123).Bytes())
	//keySetSender := cashec.KeySet{}
	//keySetSender.ImportFromPrivateKey(&spendingKey)
	//coin.CoinDetails.PublicKey, _ = privacy.DecompressKey(keySetSender.PaymentAddress.Pk)
	//
	//err := coin.Encrypt(keySetSender.PaymentAddress.Tk)
	//if err != nil {
	//	fmt.Println(err)
	//}
	//
	//coinByte := coin.Bytes()
	//
	//fmt.Printf("Coin encrypt bytes: %v\n", coinByte)
	//coin2 := new(privacy.OutputCoin)
	//err = coin2.SetBytes(coinByte)
	//if err != nil {
	//	fmt.Printf("Coin encrypt setbytes: %+v\n", coin2)
	//}
	//
	//coin.Decrypt(keySetSender.ReadonlyKey)
	//
	//fmt.Printf("DEcrypted Plain text 1: Radnomness : %v\n", coin.CoinDetails.Randomness)
	//fmt.Printf("DEcrypted Plain text 1: Value : %v\n", coin.CoinDetails.Value)

	/*----------------- TEST NDH -----------------*/
	//fmt.Println(zkp.TestProofIsZero())
	//fmt.Println(zkp.TestOpeningsProtocol())
	//fmt.Println(zkp.TestPKEqualityOfCommittedVal())
	//fmt.Printf("ElGamal PublicKey Encryption Scheme test: %v", privacy.TestElGamalPubKeyEncryption())
	/*--------------------------------------------*/

	// keySetSender := new(cashec.KeySet)
	// //spendingKey := privacy.GenerateSpendingKey([]byte{0, 1, 23, 235})
	// spendingKey := privacy.GenerateSpendingKey([]byte{1, 1, 1, 1})
	// keySetSender.ImportFromPrivateKey(&spendingKey)

	// data := []byte{0}
	// signature, err := keySetSender.Sign(data)
	// if err != nil{
	// 	fmt.Println(err)
	// }
	// fmt.Println(hex.EncodeToString(signature))

	// //signature , _:= hex.DecodeString("5d9f5e9c350a877ddbbe227b40c19b00c040e715924740f2d92cc9dc02da5937ba433dbca431f2a0a447e21fd096d894f869a9e31b8217ee0cf9c33f8b032ade")
	// //
	// res, err := keySetSender.Verify(data, signature)
	// if err != nil{
	// 	fmt.Println(err)
	// }

	// fmt.Println(res)

	/*----------------- TEST TX SALARY -----------------*/

	// keySet := new(cashec.KeySet)
	// spendingKey := privacy.GenerateSpendingKey([]byte{1, 1, 1, 1})
	// keySet.ImportFromPrivateKey(&spendingKey)

	// var db database.DatabaseInterface

	// tx, err := transaction.InitTxSalary(10, &keySet.PaymentAddress, &keySet.PrivateKey, db)
	// if err != nil{
	// 	fmt.Println(err)
	// }
	// fmt.Printf("Tx: %+v\n", tx)

	// res := transaction.ValidateTxSalary(tx, db)

	// fmt.Printf("Res: %v\n", res)

	/*----------------- TEST IS NIL -----------------*/
	//zkp := new(zkp.PKOneOfManyProof)
	//fmt.Printf("len zkp.cl: %v\n", len(zkp.cl))
	//fmt.Println(zkp.IsNil())

	//coin := new(privacy.Coin).Init()
	//fmt.Println(coin.SerialNumber == nil)
	//fmt.Printf("coin.Serial numbre: %v\n", coin.SerialNumber)

	//num := 0
	//bytes := privacy.IntToByteArr(num)
	//fmt.Printf("bytes: %v\n", bytes)
	//
	//num2 := privacy.ByteArrToInt(bytes)
	//fmt.Printf("num2: %v\n", num2)

	/*----------------- TEST COIN BYTES -----------------*/

	//keySet := new(cashec.KeySet)
	//spendingKey := privacy.GenerateSpendingKey([]byte{1, 1, 1, 1})
	//keySet.ImportFromPrivateKey(&spendingKey)
	//
	//coin := new(privacy.Coin)
	//coin.PublicKey, _ = privacy.DecompressKey(keySet.PaymentAddress.Pk)
	//
	//coin.Value = 10
	//coin.SNDerivator = privacy.RandInt()
	//coin.Randomness = privacy.RandInt()
	//coin.CommitAll()
	//coin.Value = 0
	//
	//
	//outCoin := new(privacy.OutputCoin)
	//outCoin.CoinDetails = coin
	//outCoin.CoinDetailsEncrypted = new(privacy.CoinDetailsEncrypted)
	//outCoin.Encrypt(keySet.PaymentAddress.Tk)
	//coin.Randomness = nil
	//
	//outCoinBytes := outCoin.Bytes()
	//
	//fmt.Printf("Out coin bytes: %v\n", outCoinBytes)
	//fmt.Printf("Len Out coin bytes: %v\n", len(outCoinBytes))

	/*----------------- TEST SIGN TX -----------------*/
	//tx := new(transaction.Tx)
	//tx.Fee = 0
	//tx.Type = common.TxNormalType
	//
	//keySet := new(cashec.KeySet)
	//spendingKey := privacy.GenerateSpendingKey([]byte{1, 1, 1, 1})
	//fmt.Printf("spending key byte : %v\n", spendingKey)
	//keySet.ImportFromPrivateKey(&spendingKey)
	//
	//randSK := privacy.RandInt()
	//
	//tx.PubKeyLastByteSender = keySet.PaymentAddress.Pk[len(keySet.PaymentAddress.Pk) - 1]
	//sigPrivKeyBytes := tx.SetSigPrivKey(spendingKey, randSK)
	//
	//fmt.Printf("spending key byte : %v\n", spendingKey)
	//fmt.Printf("randSK byte : %v\n", randSK.Bytes())
	//fmt.Printf("Private key combine: %v\n", sigPrivKeyBytes)
	//
	//
	//
	//tx.SignTx(true)
	//
	//res, err := tx.VerifySigTx(true)
	//if err != nil{
	//	fmt.Printf("Err: %v\n", err)
	//}
	//
	//fmt.Println(res)

	/*----------------- TEST AddPaddingBigInt -----------------*/


	//keySet := new(cashec.KeySet)
	//spendingKey := privacy.GenerateSpendingKey([]byte{1, 1, 1, 1})
	//keySet.ImportFromPrivateKey(&spendingKey)
	//
	//coin := new(privacy.Coin)
	//coin.PublicKey, _ = privacy.DecompressKey(keySet.PaymentAddress.Pk)
	//
	//coin.Value = 10
	//coin.SNDerivator = privacy.RandInt()
	//coin.Randomness = privacy.RandInt()
	//coin.CommitAll()
	//
	//// recalculate coin commitment
	//fmt.Printf("coin info: %+v\n", coin)
	//com := new(privacy.EllipticPoint)
	//com.X, com.Y = big.NewInt(0), big.NewInt(0)
	//com.X.Set(coin.PublicKey.X)
	//com.Y.Set(coin.PublicKey.Y)
	//fmt.Printf("g_sK : %+v\n", privacy.PedCom.G[privacy.SK])
	//fmt.Printf("g^sK : %+v\n", com)
	//
	//tmp := new(privacy.EllipticPoint)
	//tmp.X, tmp.Y = big.NewInt(0), big.NewInt(0)
	//tmp.X.Set(privacy.PedCom.G[privacy.VALUE].X)
	//tmp.Y.Set(privacy.PedCom.G[privacy.VALUE].Y)
	//tmp = tmp.ScalarMult(new(big.Int).SetUint64(coin.Value))
	//com = com.Add(tmp)
	//fmt.Printf("g_VALUE : %+v\n", privacy.PedCom.G[privacy.VALUE])
	//fmt.Printf("g^Value : %+v\n", tmp)
	//
	//tmp = new(privacy.EllipticPoint)
	//tmp.X, tmp.Y = big.NewInt(0), big.NewInt(0)
	//tmp.X.Set(privacy.PedCom.G[privacy.SND].X)
	//tmp.Y.Set(privacy.PedCom.G[privacy.SND].Y)
	//tmp = tmp.ScalarMult(coin.SNDerivator)
	//com = com.Add(tmp)
	//fmt.Printf("g_SND : %+v\n", privacy.PedCom.G[privacy.SND])
	//fmt.Printf("g^SND : %+v\n", tmp)
	//
	//tmp = new(privacy.EllipticPoint)
	//tmp.X, tmp.Y = big.NewInt(0), big.NewInt(0)
	//tmp.X.Set(privacy.PedCom.G[privacy.SHARDID].X)
	//tmp.Y.Set(privacy.PedCom.G[privacy.SHARDID].Y)
	//tmp = tmp.ScalarMult(new(big.Int).SetBytes([]byte{coin.GetPubKeyLastByte()}))
	//com = com.Add(tmp)
	//fmt.Printf("g_SHARDID : %+v\n", privacy.PedCom.G[privacy.SHARDID])
	//fmt.Printf("g^Sharid : %+v\n", tmp)
	//
	//tmp = new(privacy.EllipticPoint)
	//tmp.X, tmp.Y = big.NewInt(0), big.NewInt(0)
	//tmp.X.Set(privacy.PedCom.G[privacy.RAND].X)
	//tmp.Y.Set(privacy.PedCom.G[privacy.RAND].Y)
	//tmp = tmp.ScalarMult(coin.Randomness)
	//com = com.Add(tmp)
	//fmt.Printf("g_RAND : %+v\n", privacy.PedCom.G[privacy.RAND])
	//fmt.Printf("g^Randomness : %+v\n", tmp)
	//
	//if !com.IsEqual(coin.CoinCommitment){
	//	fmt.Printf("wrong")
	//}

	//point := new(privacy.EllipticPoint)
	//point.Randomize()
	//fmt.Println(point)

	//fmt.Println(privacy.TestElGamalPubKeyEncryption())

	//aes := new(privacy.AES)
	//key := privacy.RandInt().Bytes()
	//
	//fmt.Printf("Key : %v\n", len(key))
	//aes.SetKey(key)
	//ciphertext, err := aes.Encrypt([]byte{123})
	//if err != nil{
	//	fmt.Println(err)
	//}
	//
	//
	//
	//fmt.Printf("Ciphertext: %v\n", ciphertext)
	//plaintext, _ := aes.Decrypt(ciphertext)
	//fmt.Printf("Plaintext: %v\n", plaintext)

	//a := big.NewInt(10)
	//
	////x := new(big.Int).Set(a)
	//x := &a
	//fmt.Printf("address of x: %v\n", &x)
	//fmt.Printf("value of x: %v\n", x)
	//
	//y := x
	//fmt.Printf("address of y: %v\n", &y)
	//fmt.Printf("value of y: %v\n", y)
	//
	//x = big.NewInt(2)
	//fmt.Printf("address of x: %v\n", &x)
	//fmt.Printf("value of x: %v\n", x)
	//fmt.Printf("address of y: %v\n", &y)
	//fmt.Printf("value of y: %v\n", y)


	//point1 := new(privacy.EllipticPoint).Zero()
	//point1.X.Set(privacy.RandInt())
	//point1.ComputeYCoord()
	//
	////point1.Randomize()
	//
	//point2 := new(privacy.EllipticPoint).Zero()
	//point2.X.Set(privacy.RandInt())
	//point2.ComputeYCoord()
	////point2.Randomize()
	//
	//fmt.Printf("Point 1: %p\n", (*point1).X)
	//fmt.Printf("Point 2: %p\n", (*point2).X)
	//fmt.Printf("\n\n")
	//
	////*(*point1).X = *(*point2).X
	//*(point1.X) = *(point2.X)
	//
	//
	////*(*point1).Y = *(*point2).Y
	//
	//fmt.Printf("Point 1: %p\n", (*point1).X)
	//fmt.Printf("Point 2: %p\n", (*point2).X)
	//fmt.Printf("Point 1 value: %v\n", *(*point1).X)
	//fmt.Printf("Point 2 value: %v\n", *(*point2).X)
	//fmt.Printf("\n\n")
	//
	////*(*point2).X = *privacy.RandInt()
	//tmp:= privacy.RandInt()
	//fmt.Printf("addr temp: %p\n", tmp)
	//point2.X = tmp
	////(*point2).X.Set(tmp)
	//
	//fmt.Printf("Point 1: %p\n", (*point1).X)
	//fmt.Printf("Point 2: %p\n", (*point2).X)
	//fmt.Printf("Point 1 value: %v\n", *(*point1).X)
	//fmt.Printf("Point 2 value: %v\n", *(*point2).X)
	//fmt.Printf("\n\n")

	privacy.TestSchn()

}
