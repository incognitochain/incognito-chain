package gomobile

import (
	"encoding/base64"
	"encoding/json"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/consensus/signatureschemes/blsmultisig"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/transaction"
	"github.com/incognitochain/incognito-chain/wallet"
	"github.com/pkg/errors"
	"math/big"
	"strconv"
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
func GenerateBLSKeyPairFromSeed(args string) string {
	// convert seed from string to bytes array
	//fmt.Printf("args: %v\n", args)
	seed, _ := base64.StdEncoding.DecodeString(args)
	//fmt.Printf("bls seed: %v\n", seed)

	// generate  bls key
	privateKey, publicKey := blsmultisig.KeyGen(seed)

	// append key pair to one bytes array
	keyPairBytes := []byte{}
	keyPairBytes = append(keyPairBytes, common.AddPaddingBigInt(privateKey, common.BigIntSize)...)
	keyPairBytes = append(keyPairBytes, blsmultisig.CmprG2(publicKey)...)

	//  base64.StdEncoding.EncodeToString()
	keyPairEncode := base64.StdEncoding.EncodeToString(keyPairBytes)

	return keyPairEncode
}

// args: seed
func GenerateKeyFromSeed(seedB64Encoded string) (string, error) {
	seed, err := base64.StdEncoding.DecodeString(seedB64Encoded)
	if err != nil {
		return "", nil
	}

	println("[Go] Seed: ", seed)

	key := privacy.GeneratePrivateKey(seed)
	println("[Go] key: ", key)

	res := base64.StdEncoding.EncodeToString(key)
	println("[Go] res: ", res)
	return res, nil
}

func ScalarMultBase(scalarB64Encode string) (string, error) {
	scalar, err := base64.StdEncoding.DecodeString(scalarB64Encode)
	if err != nil {
		return "", nil
	}

	point := new(privacy.Point).ScalarMultBase(new(privacy.Scalar).FromBytesS(scalar))
	res := base64.StdEncoding.EncodeToString(point.ToBytesS())
	return res, nil
}

func DeriveSerialNumber(args string) (string, error) {
	// parse data
	bytes := []byte(args)
	println("Bytes: %v\n", bytes)

	paramMaps := make(map[string]interface{})

	err := json.Unmarshal(bytes, &paramMaps)
	if err != nil {
		println("Error can not unmarshal data : %v\n", err)
		return "", err
	}

	privateKeyStr, ok := paramMaps["privateKey"].(string)
	if !ok {
		println("Invalid private key")
		return "", errors.New("Invalid private key")
	}

	keyWallet, err := wallet.Base58CheckDeserialize(privateKeyStr)
	if err != nil {
		println("Can not decode private key")
		return "", errors.New("Can not decode private key")
	}
	privateKeyScalar := new(privacy.Scalar).FromBytesS(keyWallet.KeySet.PrivateKey)

	snds, ok := paramMaps["snds"].([]interface{})
	if !ok {
		println("Invalid list of serial number derivator")
		return "", errors.New("Invalid list of serial number derivator")

	}
	sndScalars := make([]*privacy.Scalar, len(snds))

	for i := 0; i < len(snds); i++ {
		tmp, ok := snds[i].(string)
		println("tmp: ", tmp)
		if !ok {
			println("Invalid serial number derivator")
			return "", errors.New("Invalid serial number derivator")

		}
		sndBytes, _, err := base58.Base58Check{}.Decode(tmp)
		println("sndBytes: ", sndBytes)
		if err != nil {
			println("Can not decode serial number derivator")
			return "", errors.New("Can not decode serial number derivator")
		}
		sndScalars[i] = new(privacy.Scalar).FromBytesS(sndBytes)
	}

	// calculate serial number and return result

	serialNumberPoint := make([]*privacy.Point, len(sndScalars))
	serialNumberStr := make([]string, len(serialNumberPoint))

	serialNumberBytes := make([]byte, 0)

	for i := 0; i < len(sndScalars); i++ {
		serialNumberPoint[i] = new(privacy.Point).Derive(privacy.PedCom.G[privacy.PedersenPrivateKeyIndex], privateKeyScalar, sndScalars[i])
		println("serialNumberPoint[i]: ", serialNumberPoint[i])

		serialNumberStr[i] = base58.Base58Check{}.Encode(serialNumberPoint[i].ToBytesS(), 0x00)
		println("serialNumberStr[i]: ", serialNumberStr[i])
		serialNumberBytes = append(serialNumberBytes, serialNumberPoint[i].ToBytesS()...)
	}

	result := base64.StdEncoding.EncodeToString(serialNumberBytes)

	return result, nil
}

func InitPrivacyTx(args string) (string, error) {
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
		return "", errors.New("Invalid sender private key")
	}
	println("senderSKParam: %v\n", senderSKParam)

	keyWallet, err := wallet.Base58CheckDeserialize(senderSKParam)
	if err != nil {
		println("Error can not decode sender private key : %v\n", err)
		return "", err
	}
	senderSK := keyWallet.KeySet.PrivateKey
	println("senderSK: ", senderSK)

	//get payment infos
	println(paramMaps["paramPaymentInfos"])
	paymentInfoParams := paramMaps["paramPaymentInfos"].([]interface{})
	//if !ok {
	//	println("Invalid payment info params!")
	//	return "", errors.New("Invalid payment info params")
	//}

	paymentInfo := make([]*privacy.PaymentInfo, 0)
	for i := 0; i < len(paymentInfoParams); i++ {
		tmp := paymentInfoParams[i].(map[string]interface{})
		paymentAddrStr, ok := tmp["paymentAddressStr"].(string)
		if !ok {
			println("Invalid payment info params!")
			return "", err
		}

		amount, ok := tmp["amount"].(float64)

		paymentInfoTmp := new(privacy.PaymentInfo)
		keyWallet, err := wallet.Base58CheckDeserialize(paymentAddrStr)
		if err != nil {
			println("Error can not decode sender private key : %v\n", err)
			return "", err
		}
		paymentInfoTmp.PaymentAddress = keyWallet.KeySet.PaymentAddress
		paymentInfoTmp.Amount = uint64(amount)

		paymentInfo = append(paymentInfo, paymentInfoTmp)
	}

	//get fee
	fee := paramMaps["fee"].(float64)
	println("fee: ", fee)

	// get has Privacy
	hasPrivacy := paramMaps["isPrivacy"].(bool)

	println("hasPrivacy: ", hasPrivacy)

	inputCoinStrs, _ := paramMaps["inputCoinStrs"].([]interface{})
	println("inputCoinStrs: ", inputCoinStrs)

	inputCoins := make([]*privacy.InputCoin, len(inputCoinStrs))
	for i := 0; i < len(inputCoins); i++ {
		tmp := inputCoinStrs[i].(map[string]interface{})
		coinObjTmp := new(privacy.CoinObject)
		coinObjTmp.PublicKey = tmp["PublicKey"].(string)
		coinObjTmp.CoinCommitment = tmp["CoinCommitment"].(string)
		coinObjTmp.SNDerivator = tmp["SNDerivator"].(string)
		coinObjTmp.SerialNumber = tmp["SerialNumber"].(string)
		coinObjTmp.Randomness = tmp["Randomness"].(string)
		coinObjTmp.Value = tmp["Value"].(string)
		coinObjTmp.Info = tmp["Info"].(string)

		inputCoins[i] = new(privacy.InputCoin).Init()
		inputCoins[i].ParseCoinObjectToInputCoin(*coinObjTmp)
	}

	println("inputCoins: ", inputCoins)

	commitmentIndicesParam, ok := paramMaps["commitmentIndices"].([]interface{})
	if !ok {
		return "", errors.New("invalid commitment indices param")
	}
	commitmentStrsParam, ok := paramMaps["commitmentStrs"].([]interface{})
	if !ok {
		return "", errors.New("invalid commitment strings param")
	}

	myCommitmentIndicesParam, ok := paramMaps["myCommitmentIndices"].([]interface{})
	if !ok {
		return "", errors.New("invalid my commitment indices param")
	}

	sndOutputsParam, ok := paramMaps["sndOutputs"].([]interface{})
	if !ok {
		return "", errors.New("invalid snd outputs param")
	}

	println("sndOutputsParam: ", sndOutputsParam)

	commitmentIndices := make([]uint64, len(commitmentIndicesParam))
	commitmentStrs := make([]string, len(commitmentStrsParam))
	myCommitmentIndices := make([]uint64, len(myCommitmentIndicesParam))
	sndOutputs := make([]*privacy.Scalar, len(sndOutputsParam))

	commitmentBytes := make([][]byte, len(commitmentStrsParam))
	for i := 0; i < len(commitmentIndices); i++ {
		commitmentIndices[i] = uint64(commitmentIndicesParam[i].(float64))
		commitmentStrs[i] = commitmentStrsParam[i].(string)

		commitmentBytes[i], _, err = base58.Base58Check{}.Decode(commitmentStrs[i])
		if err != nil {
			return "", nil
		}
	}

	for i := 0; i < len(myCommitmentIndices); i++ {
		myCommitmentIndices[i] = uint64(myCommitmentIndicesParam[i].(float64))
	}

	for i := 0; i < len(sndOutputs); i++ {

		println("sndOutputsParam[i].(string): ", sndOutputsParam[i].(string))
		tmp, _, err := base58.Base58Check{}.Decode(sndOutputsParam[i].(string))
		if err != nil {
			return "", nil
		}

		sndOutputs[i] = new(privacy.Scalar).FromBytesS(tmp)
	}

	paramCreateTx := transaction.NewTxPrivacyInitParamsForASM(&senderSK, paymentInfo, inputCoins, uint64(fee), hasPrivacy, nil, nil, nil, commitmentIndices, commitmentBytes, myCommitmentIndices, sndOutputs)
	println("paramCreateTx: ", paramCreateTx)

	tx := new(transaction.Tx)
	err = tx.InitForASM(paramCreateTx)

	if err != nil {
		println("Can not create tx: ", err)
		return "", err
	}

	// serialize tx json
	txJson, err := json.Marshal(tx)
	if err != nil {
		println("Can not marshal tx: ", err)
		return "", err
	}

	lockTimeBytes := common.AddPaddingBigInt(new(big.Int).SetInt64(tx.LockTime), 8)
	resBytes := append(txJson, lockTimeBytes...)

	B64Res := base64.StdEncoding.EncodeToString(resBytes)

	return B64Res, nil
}

func InitPrivacyTokenTx(args string) (string, error) {
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
		return "", errors.New("Invalid sender private key")
	}
	println("senderSKParam: %v\n", senderSKParam)

	keyWallet, err := wallet.Base58CheckDeserialize(senderSKParam)
	if err != nil {
		println("Error can not decode sender private key : %v\n", err)
		return "", err
	}
	senderSK := keyWallet.KeySet.PrivateKey
	println("senderSK: ", senderSK)

	//get payment infos
	println(paramMaps["paramPaymentInfos"])
	paymentInfoParams := paramMaps["paramPaymentInfos"].([]interface{})
	//if !ok {
	//	println("Invalid payment info params!")
	//	return "", errors.New("Invalid payment info params")
	//}

	paymentInfo := make([]*privacy.PaymentInfo, 0)
	for i := 0; i < len(paymentInfoParams); i++ {
		tmp := paymentInfoParams[i].(map[string]interface{})
		paymentAddrStr, ok := tmp["paymentAddressStr"].(string)
		if !ok {
			println("Invalid payment info params!")
			return "", err
		}

		amount, ok := tmp["amount"].(float64)

		paymentInfoTmp := new(privacy.PaymentInfo)
		keyWallet, err := wallet.Base58CheckDeserialize(paymentAddrStr)
		if err != nil {
			println("Error can not decode sender private key : %v\n", err)
			return "", err
		}
		paymentInfoTmp.PaymentAddress = keyWallet.KeySet.PaymentAddress
		paymentInfoTmp.Amount = uint64(amount)

		paymentInfo = append(paymentInfo, paymentInfoTmp)
	}

	//get fee
	fee := paramMaps["fee"].(float64)
	println("fee: ", fee)

	// get has Privacy
	hasPrivacy := paramMaps["isPrivacy"].(bool)

	println("hasPrivacy: ", hasPrivacy)

	inputCoinStrs, _ := paramMaps["inputCoinStrs"].([]interface{})
	println("inputCoinStrs: ", inputCoinStrs)

	inputCoins := make([]*privacy.InputCoin, len(inputCoinStrs))
	for i := 0; i < len(inputCoins); i++ {
		tmp := inputCoinStrs[i].(map[string]interface{})
		coinObjTmp := new(privacy.CoinObject)
		coinObjTmp.PublicKey = tmp["PublicKey"].(string)
		coinObjTmp.CoinCommitment = tmp["CoinCommitment"].(string)
		coinObjTmp.SNDerivator = tmp["SNDerivator"].(string)
		coinObjTmp.SerialNumber = tmp["SerialNumber"].(string)
		coinObjTmp.Randomness = tmp["Randomness"].(string)
		coinObjTmp.Value = tmp["Value"].(string)
		coinObjTmp.Info = tmp["Info"].(string)

		inputCoins[i] = new(privacy.InputCoin).Init()
		inputCoins[i].ParseCoinObjectToInputCoin(*coinObjTmp)
	}

	println("inputCoins: ", inputCoins)

	// for native token
	commitmentIndicesParamForNativeToken, ok := paramMaps["commitmentIndicesForNativeToken"].([]interface{})
	if !ok {
		return "", errors.New("invalid commitment indices param")
	}
	commitmentStrsParamForNativeToken, ok := paramMaps["commitmentStrsForNativeToken"].([]interface{})
	if !ok {
		return "", errors.New("invalid commitment strings param")
	}

	myCommitmentIndicesParamForNativeToken, ok := paramMaps["myCommitmentIndicesForNativeToken"].([]interface{})
	if !ok {
		return "", errors.New("invalid my commitment indices param")
	}

	sndOutputsParamForNativeToken, ok := paramMaps["sndOutputsForNativeToken"].([]interface{})
	if !ok {
		return "", errors.New("invalid snd outputs param")
	}

	println("sndOutputsParamForNativeToken: ", sndOutputsParamForNativeToken)

	commitmentIndicesForNativeToken := make([]uint64, len(commitmentIndicesParamForNativeToken))
	commitmentStrsForNativeToken := make([]string, len(commitmentStrsParamForNativeToken))
	myCommitmentIndicesForNativeToken := make([]uint64, len(myCommitmentIndicesParamForNativeToken))
	sndOutputsForNativeToken := make([]*privacy.Scalar, len(sndOutputsParamForNativeToken))

	commitmentBytesForNativeToken := make([][]byte, len(commitmentStrsParamForNativeToken))
	for i := 0; i < len(commitmentIndicesForNativeToken); i++ {
		commitmentIndicesForNativeToken[i] = uint64(commitmentIndicesParamForNativeToken[i].(float64))
		commitmentStrsForNativeToken[i] = commitmentStrsParamForNativeToken[i].(string)

		commitmentBytesForNativeToken[i], _, err = base58.Base58Check{}.Decode(commitmentStrsForNativeToken[i])
		if err != nil {
			return "", nil
		}
	}

	for i := 0; i < len(myCommitmentIndicesForNativeToken); i++ {
		myCommitmentIndicesForNativeToken[i] = uint64(myCommitmentIndicesParamForNativeToken[i].(float64))
	}

	for i := 0; i < len(sndOutputsForNativeToken); i++ {

		println("sndOutputsParamForNativeToken[i].(string): ", sndOutputsParamForNativeToken[i].(string))
		tmp, _, err := base58.Base58Check{}.Decode(sndOutputsParamForNativeToken[i].(string))
		if err != nil {
			return "", nil
		}

		sndOutputsForNativeToken[i] = new(privacy.Scalar).FromBytesS(tmp)
	}

	// for privacy token
	commitmentIndicesParamForPToken, ok := paramMaps["commitmentIndicesForNativeToken"].([]interface{})
	if !ok {
		return "", errors.New("invalid commitment indices param")
	}
	commitmentStrsParamForPToken, ok := paramMaps["commitmentStrsForNativeToken"].([]interface{})
	if !ok {
		return "", errors.New("invalid commitment strings param")
	}

	myCommitmentIndicesParamForPToken, ok := paramMaps["myCommitmentIndicesForNativeToken"].([]interface{})
	if !ok {
		return "", errors.New("invalid my commitment indices param")
	}

	sndOutputsParamForPToken, ok := paramMaps["sndOutputsForNativeToken"].([]interface{})
	if !ok {
		return "", errors.New("invalid snd outputs param")
	}

	println("sndOutputsParamForPToken: ", sndOutputsParamForPToken)

	commitmentIndicesForPToken := make([]uint64, len(commitmentIndicesParamForPToken))
	commitmentStrsForPToken := make([]string, len(commitmentStrsParamForPToken))
	myCommitmentIndicesForPToken := make([]uint64, len(myCommitmentIndicesParamForPToken))
	sndOutputsForPToken := make([]*privacy.Scalar, len(sndOutputsParamForPToken))

	commitmentBytesForPToken := make([][]byte, len(commitmentStrsParamForPToken))
	for i := 0; i < len(commitmentIndicesForPToken); i++ {
		commitmentIndicesForPToken[i] = uint64(commitmentIndicesParamForPToken[i].(float64))
		commitmentStrsForPToken[i] = commitmentStrsParamForPToken[i].(string)

		commitmentBytesForPToken[i], _, err = base58.Base58Check{}.Decode(commitmentStrsForPToken[i])
		if err != nil {
			return "", nil
		}
	}

	for i := 0; i < len(myCommitmentIndicesForPToken); i++ {
		myCommitmentIndicesForPToken[i] = uint64(myCommitmentIndicesParamForPToken[i].(float64))
	}

	for i := 0; i < len(sndOutputsForPToken); i++ {
		println("sndOutputsParamForNativeToken[i].(string): ", sndOutputsParamForPToken[i].(string))
		tmp, _, err := base58.Base58Check{}.Decode(sndOutputsParamForPToken[i].(string))
		if err != nil {
			return "", nil
		}

		sndOutputsForPToken[i] = new(privacy.Scalar).FromBytesS(tmp)
	}

	//todo:
	privacyTokenParam := new(transaction.CustomTokenPrivacyParamTx)
	hasPrivacyForPToken := false
	shardID := byte(1)
	info := []byte{}

	paramCreateTx := transaction.NewTxPrivacyTokenInitParamsForASM(
		&senderSK, paymentInfo, inputCoins, uint64(fee), privacyTokenParam, nil, hasPrivacy, hasPrivacyForPToken, shardID, info,
		commitmentIndicesForNativeToken, commitmentBytesForNativeToken, myCommitmentIndicesForNativeToken, sndOutputsForNativeToken,
		commitmentIndicesForPToken, commitmentBytesForPToken, myCommitmentIndicesForPToken, sndOutputsForPToken)
	println("paramCreateTx: ", paramCreateTx)

	tx := new(transaction.TxCustomTokenPrivacy)
	err = tx.InitForASM(paramCreateTx)

	if err != nil {
		println("Can not create tx: ", err)
		return "", err
	}

	// serialize tx json
	txJson, err := json.Marshal(tx)
	if err != nil {
		println("Can not marshal tx: ", err)
		return "", err
	}

	lockTimeBytes := common.AddPaddingBigInt(new(big.Int).SetInt64(tx.LockTime), 8)
	resBytes := append(txJson, lockTimeBytes...)

	B64Res := base64.StdEncoding.EncodeToString(resBytes)

	return B64Res, nil
}


func RandomScalars(n string) (string, error) {
	nInt, err := strconv.ParseUint(n, 10, 64)
	println("nInt: ", nInt)
	if err != nil {
		return "", nil

	}

	scalars := make([]byte, 0)
	for i := 0; i < int(nInt); i++ {
		scalars = append(scalars, privacy.RandomScalar().ToBytesS()...)
	}

	res := base64.StdEncoding.EncodeToString(scalars)

	println("res scalars: ", res)

	return res, nil
}
