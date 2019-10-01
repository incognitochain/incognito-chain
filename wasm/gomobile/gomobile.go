package gomobile

import (
	"encoding/json"
	"github.com/incognitochain/incognito-chain/database"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/transaction"
	"github.com/incognitochain/incognito-chain/wallet"
)

//args {
//     "values": valueStrs,
//     "rands": randStrs
//   }
//convert object to JSON string (JSON.stringify)
//func AggregatedRangeProve(args string) string {
//	println("args:", args)
//	bytes := []byte(args)
//	println("Bytes:", bytes)
//	temp := make(map[string][]string)
//
//	err := json.Unmarshal(bytes, &temp)
//	if err != nil {
//		println("Can not unmarshal", err)
//		return ""
//	}
//	println("temp values", temp["values"])
//	println("temp rands", temp["rands"])
//
//	if len(temp["values"]) != len(temp["rands"]) {
//		println("Wrong args")
//	}
//
//	values := make([]uint64, len(temp["values"]))
//	rands := make([]*privacy.Scalar, len(temp["values"]))
//
//	//todo
//	for i := 0; i < len(temp["values"]); i++ {
//		values[i] = temp["values"][i]
//		rands[i], _ = new(privacy.Scalar).SetString(temp["rands"][i], 10)
//	}
//
//	wit := new(aggregaterange.AggregatedRangeWitness)
//	wit.Set(values, rands)
//
//	start := time.Now()
//	proof, err := wit.Prove()
//	if err != nil {
//		println("Err: %v\n", err)
//	}
//	end := time.Since(start)
//	println("Aggregated range proving time: %v\n", end)
//
//	proofBytes := proof.Bytes()
//	println("Proof bytes: ", proofBytes)
//
//	proofBase64 := base64.StdEncoding.EncodeToString(proofBytes)
//	println("proofBase64: %v\n", proofBase64)
//
//	return proofBase64
//}
//
//// args {
////      "commitments": commitments,   // list of bytes arrays
////      "rand": rand,					// string
//// 		"indexiszero" 					//number
////    }
//// convert object to JSON string (JSON.stringify)
//func OneOutOfManyProve(args string) (string, error) {
//	bytes := []byte(args)
//	//println("Bytes:", bytes)
//	temp := make(map[string][]string)
//
//	err := json.Unmarshal(bytes, &temp)
//	if err != nil {
//		println(err)
//		return "", err
//	}
//
//	// list of commitments
//	commitmentStrs := temp["commitments"]
//	//fmt.Printf("commitmentStrs: %v\n", commitmentStrs)
//
//	if len(commitmentStrs) != privacy.CommitmentRingSize {
//		println(err)
//		return "", errors.New("the number of Commitment list's elements must be equal to CMRingSize")
//	}
//
//	commitmentPoints := make([]*privacy.Point, len(commitmentStrs))
//
//	for i := 0; i < len(commitmentStrs); i++ {
//		//fmt.Printf("commitments %v: %v\n", i,  commitmentStrs[i])
//		tmp, _ := new(big.Int).SetString(commitmentStrs[i], 16)
//		tmpByte := tmp.Bytes()
//		//fmt.Printf("tmpByte %v: %v\n", i, tmpByte)
//
//		commitmentPoints[i] = new(privacy.Point)
//		commitmentPoints[i].FromBytesS(tmpByte)
//	}
//
//	// rand
//	//randBN, _ := new(privacy.Scalar).FromUint64(temp["rand"][0])
//	randBN := privacy.RandomScalar()
//	//println("randBN: ", randBN)
//
//	// indexIsZero
//	indexIsZero, _ := new(big.Int).SetString(temp["indexiszero"][0], 10)
//	indexIsZeroUint64 := indexIsZero.Uint64()
//
//	//println("indexIsZeroUint64: ", indexIsZeroUint64)
//
//	// set witness for One out of many protocol
//	wit := new(oneoutofmany.OneOutOfManyWitness)
//	wit.Set(commitmentPoints, randBN, indexIsZeroUint64)
//	println("Wit: ", wit)
//	// proving
//	//start := time.Now()
//	proof, err := wit.Prove()
//	//fmt.Printf("Proof go: %v\n", proof)
//	if err != nil {
//		println("Err: %v\n", err)
//	}
//	//end := time.Since(start)
//	//fmt.Printf("One out of many proving time: %v\n", end)
//
//	// convert proof to bytes array
//	proofBytes := proof.Bytes()
//	//println("Proof bytes: ", proofBytes)
//
//	proofBase64 := base64.StdEncoding.EncodeToString(proofBytes)
//	//println("proofBase64: %v\n", proofBase64)
//
//	return proofBase64, nil
//}

// GenerateBLSKeyPairFromSeed generates BLS key pair from seed
//func GenerateBLSKeyPairFromSeed(args string) string {
//	// convert seed from string to bytes array
//	//fmt.Printf("args: %v\n", args)
//	seed, _ := base64.StdEncoding.DecodeString(args)
//	//fmt.Printf("bls seed: %v\n", seed)
//
//	// generate  bls key
//	privateKey, publicKey := blsmultisig.KeyGen(seed)
//
//	// append key pair to one bytes array
//	keyPairBytes := []byte{}
//	keyPairBytes = append(keyPairBytes, privateKey.Bytes()...)
//	keyPairBytes = append(keyPairBytes, blsmultisig.CmprG2(publicKey)...)
//
//	//  base64.StdEncoding.EncodeToString()
//	keyPairEncode := base64.StdEncoding.EncodeToString(keyPairBytes)
//
//	return keyPairEncode
//}
//
//

func InitPrivacyTx(args string) (string, error){
	bytes := []byte(args)
	println("Bytes: %v\n", bytes)

	paramMaps := make(map[string]interface{})

	err := json.Unmarshal(bytes, &paramMaps)
	if err != nil {
		println("Error can not unmarshal data : %v\n", err)
		return "", err
	}

	println("paramMaps:", paramMaps)

	// sender's private key
	senderSKParam, ok := paramMaps["senderSK"].(string)
	if !ok {
		println("Invalid sender private key!")
		return "", err
	}
	println("senderSKParam: %v\n", senderSKParam)

	keyWallet, err := wallet.Base58CheckDeserialize(senderSKParam)
	if err != nil{
		println("Error can not decode sender private key : %v\n", err)
		return "", err
	}
	senderSK := keyWallet.KeySet.PrivateKey
	println("senderSK: ", senderSK)

	//get payment infos
	println(paramMaps["paramPaymentInfos"])
	paymentInfoParams, ok := paramMaps["paramPaymentInfos"].(map[string]uint64)
	if !ok {
		println("Invalid payment info params!")
		return "", err
	}

	paymentInfo := make([]*privacy.PaymentInfo, 0)
	for key, value := range paymentInfoParams{
		paymentInfoTmp := new(privacy.PaymentInfo)
		keyWallet, err := wallet.Base58CheckDeserialize(key)
		if err != nil{
			println("Error can not decode sender private key : %v\n", err)
			return "", err
		}
		paymentInfoTmp.PaymentAddress = keyWallet.KeySet.PaymentAddress
		paymentInfoTmp.Amount = value

		paymentInfo = append(paymentInfo, paymentInfoTmp)
	}

	//get fee
	fee := paramMaps["fee"].(uint64)
	println("fee: ", fee)

	// get has Privacy
	hasPrivacy := paramMaps["hasPrivacy"].(bool)
	println("hasPrivacy: ", hasPrivacy)

	inputCoins := make([]*privacy.InputCoin, 0)
	println("inputCoins: ", inputCoins)

	db := new(database.DatabaseInterface)
	println("db: ", db)

	paramCreateTx := transaction.NewTxPrivacyInitParams(&senderSK, paymentInfo, inputCoins, fee, hasPrivacy, *db, nil, nil, nil )
	println("paramCreateTx: ", paramCreateTx)


	//tx:= new(transaction.Tx)
	//err = tx.Init(paramCreateTx)
	//if err != nil{
	//	println("Can not create tx: ", err)
	//	return "", err
	//}

	return "", nil

}
