package gomobile

import (
	"encoding/base64"
	"encoding/json"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/consensus/signatureschemes/blsmultisig"
	"github.com/incognitochain/incognito-chain/metadata"
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
		return "", err
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

func InitParamCreatePrivacyTx(args string) (*transaction.TxPrivacyInitParamsForASM, error) {
	bytes := []byte(args)
	println("Bytes: %v\n", bytes)

	paramMaps := make(map[string]interface{})

	err := json.Unmarshal(bytes, &paramMaps)
	if err != nil {
		println("Error can not unmarshal data : %v\n", err)
		return nil, err
	}

	println("paramMaps:", paramMaps)

	// sender's private key
	senderSKParam, ok := paramMaps["senderSK"].(string)
	if !ok {
		println("Invalid sender private key!")
		return nil, errors.New("Invalid sender private key")
	}
	println("senderSKParam: %v\n", senderSKParam)

	keyWallet, err := wallet.Base58CheckDeserialize(senderSKParam)
	if err != nil {
		println("Error can not decode sender private key : %v\n", err)
		return nil, err
	}
	senderSK := keyWallet.KeySet.PrivateKey
	println("senderSK: ", senderSK)

	//get payment infos
	println(paramMaps["paramPaymentInfos"])
	paymentInfoParams, ok := paramMaps["paramPaymentInfos"].([]interface{})
	if !ok {
		println("Invalid payment info params!")
		return nil, errors.New("Invalid payment info params")
	}

	paymentInfo := make([]*privacy.PaymentInfo, 0)
	for i := 0; i < len(paymentInfoParams); i++ {
		tmp, ok := paymentInfoParams[i].(map[string]interface{})
		if !ok {
			println("Invalid payment info param!")
			return nil, errors.New("Invalid payment info param")
		}
		paymentAddrStr, ok := tmp["paymentAddressStr"].(string)
		if !ok {
			println("Invalid payment info param payment address string")
			return nil, errors.New("Invalid payment info param payment address string")
		}

		amount, ok := tmp["amount"].(float64)
		if !ok {
			println("Invalid payment info param amount")
			return nil, errors.New("Invalid payment info param amount")
		}

		paymentInfoTmp := new(privacy.PaymentInfo)
		keyWallet, err := wallet.Base58CheckDeserialize(paymentAddrStr)
		if err != nil {
			println("Error can not decode sender private key : %v\n", err)
			return nil, err
		}
		paymentInfoTmp.PaymentAddress = keyWallet.KeySet.PaymentAddress
		paymentInfoTmp.Amount = uint64(amount)
		paymentInfo = append(paymentInfo, paymentInfoTmp)
	}

	//get fee
	fee, ok := paramMaps["fee"].(float64)
	if !ok {
		println("Invalid fee param!")
		return nil, errors.New("Invalid fee param")
	}
	println("fee: ", fee)

	// get has Privacy
	hasPrivacy, ok := paramMaps["isPrivacy"].(bool)
	if !ok {
		println("Invalid is privacy param!")
		return nil, errors.New("Invalid is privacy param")
	}
	println("hasPrivacy: ", hasPrivacy)

	// get has Privacy
	info, ok := paramMaps["info"].(string)
	if !ok {
		println("Invalid info param!")
		return nil, errors.New("Invalid info param")
	}
	infoBytes := []byte(info)
	println("infoBytes: ", infoBytes)

	inputCoinStrs, ok := paramMaps["inputCoinStrs"].([]interface{})
	if !ok {
		println("Invalid input coin string params!")
		return nil, errors.New("Invalid input coin string params")
	}
	println("inputCoinStrs: ", inputCoinStrs)

	inputCoins := make([]*privacy.InputCoin, len(inputCoinStrs))
	for i := 0; i < len(inputCoins); i++ {
		tmp, ok := inputCoinStrs[i].(map[string]interface{})
		if !ok {
			println("Invalid input coin string param!")
			return nil, errors.New("Invalid input coin string param")
		}
		coinObjTmp := new(privacy.CoinObject)
		coinObjTmp.PublicKey, ok = tmp["PublicKey"].(string)
		if !ok {
			println("Invalid input coin public key param!")
			return nil, errors.New("Invalid input coin public key param")
		}
		coinObjTmp.CoinCommitment, ok = tmp["CoinCommitment"].(string)
		if !ok {
			println("Invalid input coin coin commitment param!")
			return nil, errors.New("Invalid input coin coin commitment param")
		}
		coinObjTmp.SNDerivator, ok = tmp["SNDerivator"].(string)
		if !ok {
			println("Invalid input coin snderivator param!")
			return nil, errors.New("Invalid input coin snderivator param")
		}
		coinObjTmp.SerialNumber, ok = tmp["SerialNumber"].(string)
		if !ok {
			println("Invalid input coin serial number param!")
			return nil, errors.New("Invalid input coin serial number param")
		}
		coinObjTmp.Randomness, ok = tmp["Randomness"].(string)
		if !ok {
			println("Invalid input coin randomness param!")
			return nil, errors.New("Invalid input coin randomness param")
		}
		coinObjTmp.Value, ok = tmp["Value"].(string)
		if !ok {
			println("Invalid input coin value param!")
			return nil, errors.New("Invalid input coin value param")
		}
		coinObjTmp.Info, ok = tmp["Info"].(string)
		if !ok {
			println("Invalid input coin info param!")
			return nil, errors.New("Invalid input coin info param")
		}

		inputCoins[i] = new(privacy.InputCoin).Init()
		inputCoins[i].ParseCoinObjectToInputCoin(*coinObjTmp)
	}

	println("inputCoins: ", inputCoins)

	commitmentIndicesParam, ok := paramMaps["commitmentIndices"].([]interface{})
	if !ok {
		return nil, errors.New("invalid commitment indices param")
	}
	commitmentStrsParam, ok := paramMaps["commitmentStrs"].([]interface{})
	if !ok {
		return nil, errors.New("invalid commitment strings param")
	}

	myCommitmentIndicesParam, ok := paramMaps["myCommitmentIndices"].([]interface{})
	if !ok {
		return nil, errors.New("invalid my commitment indices param")
	}

	sndOutputsParam, ok := paramMaps["sndOutputs"].([]interface{})
	if !ok {
		return nil, errors.New("invalid snd outputs param")
	}

	println("sndOutputsParam: ", sndOutputsParam)

	commitmentIndices := make([]uint64, len(commitmentIndicesParam))
	commitmentStrs := make([]string, len(commitmentStrsParam))
	myCommitmentIndices := make([]uint64, len(myCommitmentIndicesParam))
	sndOutputs := make([]*privacy.Scalar, len(sndOutputsParam))

	commitmentBytes := make([][]byte, len(commitmentStrsParam))
	for i := 0; i < len(commitmentIndices); i++ {
		tmp, ok := commitmentIndicesParam[i].(float64)
		if !ok {
			return nil, errors.New("invalid commitment indices param")
		}
		commitmentIndices[i] = uint64(tmp)
		commitmentStrs[i], ok = commitmentStrsParam[i].(string)
		if !ok {
			return nil, errors.New("invalid commitment string param")
		}

		commitmentBytes[i], _, err = base58.Base58Check{}.Decode(commitmentStrs[i])
		if err != nil {
			return nil, nil
		}
	}

	for i := 0; i < len(myCommitmentIndices); i++ {
		tmp, ok :=  myCommitmentIndicesParam[i].(float64)
		if !ok {
			return nil, errors.New("invalid my commitment index param")
		}
		myCommitmentIndices[i] = uint64(tmp)
	}

	for i := 0; i < len(sndOutputs); i++ {
		println("sndOutputsParam[i].(string): ", sndOutputsParam[i].(string))
		tmp, _, err := base58.Base58Check{}.Decode(sndOutputsParam[i].(string))
		if err != nil {
			return nil, err
		}

		sndOutputs[i] = new(privacy.Scalar).FromBytesS(tmp)
	}

	paramCreateTx := transaction.NewTxPrivacyInitParamsForASM(&senderSK, paymentInfo, inputCoins, uint64(fee), hasPrivacy, nil, nil, nil, commitmentIndices, commitmentBytes, myCommitmentIndices, sndOutputs)
	println("paramCreateTx: ", paramCreateTx)

	return paramCreateTx, nil
}

func InitPrivacyTx(args string) (string, error) {
	paramCreateTx, err := InitParamCreatePrivacyTx(args)
	if err!= nil{
		return "", err
	}

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

func Staking(args string) (string, error) {
	// parse meta data
	bytes := []byte(args)
	println("Bytes: %v\n", bytes)

	paramMaps := make(map[string]interface{})

	err := json.Unmarshal(bytes, &paramMaps)
	if err != nil {
		println("Error can not unmarshal data : %v\n", err)
		return "", err
	}

	println("paramMaps:", paramMaps)

	metaDataParam, ok := paramMaps["metaData"].(map[string]interface{})
	if !ok {
		return "", errors.New("Invalid meta data param")
	}

	metaDataType, ok := metaDataParam["Type"].(float64)
	if !ok {
		println("Invalid meta data type param")
		return "", errors.New("Invalid meta data type param")
	}

	funderPaymentAddress, ok := metaDataParam["FunderPaymentAddress"].(string)
	if !ok {
		println("Invalid meta data funder payment address param")
		return "", errors.New("Invalid meta data funder payment address param")
	}
	rewardReceiverPaymentAddress, ok := metaDataParam["RewardReceiverPaymentAddress"].(string)
	if !ok {
		println("Invalid meta data reward receiver payment address param")
		return "", errors.New("Invalid meta data reward receiver payment address param")
	}
	stakingAmountShard, ok := metaDataParam["StakingAmountShard"].(float64)
	if !ok {
		println("Invalid meta data staking amount param")
		return "", errors.New("Invalid meta data staking amount param")
	}
	committeePublicKey, ok := metaDataParam["CommitteePublicKey"].(string)
	if !ok {
		println("Invalid meta data committee public key param")
		return "", errors.New("Invalid meta data committee public key param")
	}
	autoReStaking, ok := metaDataParam["AutoReStaking"].(bool)
	if !ok {
		println("Invalid meta data auto restaking param")
		return "", errors.New("Invalid meta data auto restaking param")
	}

	metaData, err := metadata.NewStakingMetadata(int(metaDataType), funderPaymentAddress, rewardReceiverPaymentAddress, uint64(stakingAmountShard), committeePublicKey, autoReStaking)
	if err!= nil{
		return "", err
	}

	paramCreateTx, err := InitParamCreatePrivacyTx(args)
	if err!= nil{
		return "", err
	}

	paramCreateTx.SetMetaData(metaData)

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

func InitParamCreatePrivacyTokenTx(args string) (*transaction.TxPrivacyTokenInitParamsForASM, error){
	bytes := []byte(args)
	println("Bytes: %v\n", bytes)

	paramMaps := make(map[string]interface{})

	err := json.Unmarshal(bytes, &paramMaps)
	if err != nil {
		println("Error can not unmarshal data : %v\n", err)
		return nil, err
	}

	println("paramMaps:", paramMaps)

	// sender's private key
	senderSKParam, ok := paramMaps["senderSK"].(string)
	if !ok {
		println("Invalid sender private key!")
		return nil, errors.New("Invalid sender private key")
	}
	println("senderSKParam: %v\n", senderSKParam)

	keyWallet, err := wallet.Base58CheckDeserialize(senderSKParam)
	if err != nil {
		println("Error can not decode sender private key : %v\n", err)
		return nil, err
	}
	senderSK := keyWallet.KeySet.PrivateKey
	println("senderSK: ", senderSK)

	keyWallet.KeySet.InitFromPrivateKeyByte(keyWallet.KeySet.PrivateKey)
	publicKey := keyWallet.KeySet.PaymentAddress.Pk
	shardID := common.GetShardIDFromLastByte(publicKey[len(publicKey) - 1])
	println("shardID: ", shardID)

	//get payment infos
	println(paramMaps["paramPaymentInfos"])
	paymentInfoParams, ok := paramMaps["paramPaymentInfos"].([]interface{})
	if !ok {
		println("Invalid payment info params!")
		return nil, errors.New("Invalid payment info params")
	}

	paymentInfo := make([]*privacy.PaymentInfo, 0)
	for i := 0; i < len(paymentInfoParams); i++ {
		tmp, ok := paymentInfoParams[i].(map[string]interface{})
		if !ok {
			println("Invalid payment info param!")
			return nil, errors.New("Invalid payment info param")
		}
		paymentAddrStr, ok := tmp["paymentAddressStr"].(string)
		if !ok {
			println("Invalid payment info param payment address string")
			return nil, errors.New("Invalid payment info param payment address string")
		}

		amount, ok := tmp["amount"].(float64)
		if !ok {
			println("Invalid payment info param amount")
			return nil, errors.New("Invalid payment info param amount")
		}

		paymentInfoTmp := new(privacy.PaymentInfo)
		keyWallet, err := wallet.Base58CheckDeserialize(paymentAddrStr)
		if err != nil {
			println("Error can not decode sender private key : %v\n", err)
			return nil, err
		}
		paymentInfoTmp.PaymentAddress = keyWallet.KeySet.PaymentAddress
		paymentInfoTmp.Amount = uint64(amount)
		paymentInfo = append(paymentInfo, paymentInfoTmp)
	}

	//get fee
	fee, ok := paramMaps["fee"].(float64)
	if !ok {
		println("Invalid fee param!")
		return nil, errors.New("Invalid fee param")
	}
	println("fee: ", fee)

	// get has Privacy
	hasPrivacy, ok := paramMaps["isPrivacy"].(bool)
	if !ok {
		println("Invalid is privacy param!")
		return nil, errors.New("Invalid is privacy param")
	}
	println("hasPrivacy: ", hasPrivacy)

	// get has Privacy for ptoken
	hasPrivacyForPToken, ok := paramMaps["isPrivacyForPToken"].(bool)
	if !ok {
		println("Invalid is privacy for ptoken param!")
		return nil, errors.New("Invalid is privacy for ptoken param")
	}
	println("hasPrivacyForPToken: ", hasPrivacyForPToken)

	// get info
	info, ok := paramMaps["info"].(string)
	if !ok {
		println("Invalid info param!")
		return nil, errors.New("Invalid info param")
	}
	infoBytes := []byte(info)
	println("infoBytes: ", infoBytes)

	inputCoinStrs, ok := paramMaps["inputCoinStrs"].([]interface{})
	if !ok {
		println("Invalid input coin string params!")
		return nil, errors.New("Invalid input coin string params")
	}
	println("inputCoinStrs: ", inputCoinStrs)

	inputCoins := make([]*privacy.InputCoin, len(inputCoinStrs))
	for i := 0; i < len(inputCoins); i++ {
		tmp, ok := inputCoinStrs[i].(map[string]interface{})
		if !ok {
			println("Invalid input coin string param!")
			return nil, errors.New("Invalid input coin string param")
		}
		coinObjTmp := new(privacy.CoinObject)
		coinObjTmp.PublicKey, ok = tmp["PublicKey"].(string)
		if !ok {
			println("Invalid input coin public key param!")
			return nil, errors.New("Invalid input coin public key param")
		}
		coinObjTmp.CoinCommitment, ok = tmp["CoinCommitment"].(string)
		if !ok {
			println("Invalid input coin coin commitment param!")
			return nil, errors.New("Invalid input coin coin commitment param")
		}
		coinObjTmp.SNDerivator, ok = tmp["SNDerivator"].(string)
		if !ok {
			println("Invalid input coin snderivator param!")
			return nil, errors.New("Invalid input coin snderivator param")
		}
		coinObjTmp.SerialNumber, ok = tmp["SerialNumber"].(string)
		if !ok {
			println("Invalid input coin serial number param!")
			return nil, errors.New("Invalid input coin serial number param")
		}
		coinObjTmp.Randomness, ok = tmp["Randomness"].(string)
		if !ok {
			println("Invalid input coin randomness param!")
			return nil, errors.New("Invalid input coin randomness param")
		}
		coinObjTmp.Value, ok = tmp["Value"].(string)
		if !ok {
			println("Invalid input coin value param!")
			return nil, errors.New("Invalid input coin value param")
		}
		coinObjTmp.Info, ok = tmp["Info"].(string)
		if !ok {
			println("Invalid input coin info param!")
			return nil, errors.New("Invalid input coin info param")
		}

		inputCoins[i] = new(privacy.InputCoin).Init()
		inputCoins[i].ParseCoinObjectToInputCoin(*coinObjTmp)
	}

	println("inputCoins: ", inputCoins)


	// for native token
	commitmentIndicesParamForNativeToken, ok := paramMaps["commitmentIndicesForNativeToken"].([]interface{})
	if !ok {
		return nil, errors.New("invalid commitment indices param")
	}
	commitmentStrsParamForNativeToken, ok := paramMaps["commitmentStrsForNativeToken"].([]interface{})
	if !ok {
		return nil, errors.New("invalid commitment strings param")
	}

	myCommitmentIndicesParamForNativeToken, ok := paramMaps["myCommitmentIndicesForNativeToken"].([]interface{})
	if !ok {
		return nil, errors.New("invalid my commitment indices param")
	}

	sndOutputsParamForNativeToken, ok := paramMaps["sndOutputsForNativeToken"].([]interface{})
	if !ok {
		return nil, errors.New("invalid snd outputs param")
	}

	println("sndOutputsParamForNativeToken: ", sndOutputsParamForNativeToken)

	commitmentIndicesForNativeToken := make([]uint64, len(commitmentIndicesParamForNativeToken))
	commitmentStrsForNativeToken := make([]string, len(commitmentStrsParamForNativeToken))
	myCommitmentIndicesForNativeToken := make([]uint64, len(myCommitmentIndicesParamForNativeToken))
	sndOutputsForNativeToken := make([]*privacy.Scalar, len(sndOutputsParamForNativeToken))

	commitmentBytesForNativeToken := make([][]byte, len(commitmentStrsParamForNativeToken))
	for i := 0; i < len(commitmentIndicesForNativeToken); i++ {
		tmp, ok := commitmentIndicesParamForNativeToken[i].(float64)
		if !ok {
			return nil, errors.New("invalid commitment indices for native token param")
		}
		commitmentIndicesForNativeToken[i] = uint64(tmp)
		commitmentStrsForNativeToken[i], ok = commitmentStrsParamForNativeToken[i].(string)
		if !ok {
			return nil, errors.New("invalid commitment string for native token param")
		}

		commitmentBytesForNativeToken[i], _, err = base58.Base58Check{}.Decode(commitmentStrsForNativeToken[i])
		if err != nil {
			return nil, nil
		}
	}

	for i := 0; i < len(myCommitmentIndicesForNativeToken); i++ {
		tmp, ok := myCommitmentIndicesParamForNativeToken[i].(float64)
		if !ok {
			return nil, errors.New("invalid my commitment index for native token param")
		}
		myCommitmentIndicesForNativeToken[i] = uint64(tmp)
	}

	for i := 0; i < len(sndOutputsForNativeToken); i++ {

		println("sndOutputsParamForNativeToken[i].(string): ", sndOutputsParamForNativeToken[i].(string))
		tmp, _, err := base58.Base58Check{}.Decode(sndOutputsParamForNativeToken[i].(string))
		if err != nil {
			return nil, nil
		}

		sndOutputsForNativeToken[i] = new(privacy.Scalar).FromBytesS(tmp)
	}

	// for privacy token
	commitmentIndicesParamForPToken, ok := paramMaps["commitmentIndicesForPToken"].([]interface{})
	if !ok {
		return nil, errors.New("invalid commitment indices for ptoken param")
	}
	commitmentStrsParamForPToken, ok := paramMaps["commitmentStrsForPToken"].([]interface{})
	if !ok {
		return nil, errors.New("invalid commitment strings for ptoken param")
	}

	myCommitmentIndicesParamForPToken, ok := paramMaps["myCommitmentIndicesForPToken"].([]interface{})
	if !ok {
		return nil, errors.New("invalid my commitment indices for ptoken param")
	}

	sndOutputsParamForPToken, ok := paramMaps["sndOutputsForPToken"].([]interface{})
	if !ok {
		return nil, errors.New("invalid snd outputs for ptoken param")
	}

	println("sndOutputsParamForPToken: ", sndOutputsParamForPToken)

	commitmentIndicesForPToken := make([]uint64, len(commitmentIndicesParamForPToken))
	commitmentStrsForPToken := make([]string, len(commitmentStrsParamForPToken))
	myCommitmentIndicesForPToken := make([]uint64, len(myCommitmentIndicesParamForPToken))
	sndOutputsForPToken := make([]*privacy.Scalar, len(sndOutputsParamForPToken))

	commitmentBytesForPToken := make([][]byte, len(commitmentStrsParamForPToken))
	for i := 0; i < len(commitmentIndicesForPToken); i++ {
		tmp, ok := commitmentIndicesParamForPToken[i].(float64)
		if !ok {
			return nil, errors.New("invalid commitment indices for privacy token param")
		}
		commitmentIndicesForPToken[i] = uint64(tmp)
		commitmentStrsForPToken[i] = commitmentStrsParamForPToken[i].(string)

		commitmentBytesForPToken[i], _, err = base58.Base58Check{}.Decode(commitmentStrsForPToken[i])
		if err != nil {
			return nil, err
		}
	}

	println("commitmentBytesForPToken: ", commitmentBytesForPToken)
	println("commitmentIndicesForPToken: ", commitmentIndicesForPToken)

	for i := 0; i < len(myCommitmentIndicesForPToken); i++ {
		tmp, ok := myCommitmentIndicesParamForPToken[i].(float64)
		if !ok {
			return nil, errors.New("invalid commitment indices for privacy token param")
		}
		myCommitmentIndicesForPToken[i] = uint64(tmp)
	}
	println("myCommitmentIndicesForPToken: ", myCommitmentIndicesForPToken)


	for i := 0; i < len(sndOutputsForPToken); i++ {
		tmp, _, err := base58.Base58Check{}.Decode(sndOutputsParamForPToken[i].(string))
		if err != nil {
			return nil, err
		}

		sndOutputsForPToken[i] = new(privacy.Scalar).FromBytesS(tmp)
	}
	println("sndOutputsForPToken: ", sndOutputsForPToken)

	// get privacy token param
	privacyTokenParam := new(transaction.CustomTokenPrivacyParamTx)
	pTokenParam, ok := paramMaps["privacyTokenParam"].(map[string]interface{})
	if !ok {
		println("Invalid privacy token param")
		return nil, errors.New("Invalid privacy token param")
	}
	privacyTokenParam.PropertyID, ok = pTokenParam["propertyID"].(string)
	if !ok {
		println("Invalid token ID param")
		return nil, errors.New("Invalid token ID param")
	}
	privacyTokenParam.PropertyName, ok = pTokenParam["propertyName"].(string)
	if !ok {
		println("Invalid token name param")
		return nil, errors.New("Invalid token name param")
	}
	privacyTokenParam.PropertySymbol, ok = pTokenParam["propertySymbol"].(string)
	if !ok {
		println("Invalid token symbol param")
		return nil, errors.New("Invalid token symbol param")
	}
	tmpAmount, ok := pTokenParam["amount"].(float64)
	if !ok {
		println("Invalid amount param")
		return nil, errors.New("Invalid amount param")
	}
	privacyTokenParam.Amount = uint64(tmpAmount)
	tmpTokenTxType, ok := pTokenParam["tokenTxType"].(float64)
	if !ok {
		println("Invalid token tx type param")
		return nil, errors.New("Invalid token tx type param")
	}
	privacyTokenParam.TokenTxType = int(tmpTokenTxType)
	tmpFeePToken, ok := pTokenParam["fee"].(float64)
	if !ok {
		println("Invalid fee token param")
		return nil, errors.New("Invalid fee token param")
	}
	privacyTokenParam.Fee = uint64(tmpFeePToken)
	paymentInfoForPTokenParam, ok := pTokenParam["paymentInfoForPToken"].([]interface{})
	if !ok {
		println("Invalid payment info params!")
		return nil, errors.New("Invalid payment info params")
	}

	paymentInfoForPToken := make([]*privacy.PaymentInfo, 0)
	for i := 0; i < len(paymentInfoForPTokenParam); i++ {
		tmp, ok := paymentInfoForPTokenParam[i].(map[string]interface{})
		if !ok {
			println("Invalid payment info param!")
			return nil, errors.New("Invalid payment info param")
		}
		paymentAddrStr, ok := tmp["paymentAddressStr"].(string)
		if !ok {
			println("Invalid payment info param payment address string")
			return nil, errors.New("Invalid payment info param payment address string")
		}

		amount, ok := tmp["amount"].(float64)
		if !ok {
			println("Invalid payment info param amount")
			return nil, errors.New("Invalid payment info param amount")
		}

		paymentInfoTmp := new(privacy.PaymentInfo)
		keyWallet, err := wallet.Base58CheckDeserialize(paymentAddrStr)
		if err != nil {
			println("Error can not decode sender private key : %v\n", err)
			return nil, err
		}
		paymentInfoTmp.PaymentAddress = keyWallet.KeySet.PaymentAddress
		println("PK receiver token: ", paymentInfoTmp.PaymentAddress.Pk)
		paymentInfoTmp.Amount = uint64(amount)
		paymentInfoForPToken = append(paymentInfoForPToken, paymentInfoTmp)
	}

	privacyTokenParam.Receiver = paymentInfoForPToken

	tokenInputsParam, ok := pTokenParam["tokenInputs"].([]interface{})
	if !ok {
		println("Invalid token input coin string params!")
		return nil, errors.New("Invalid token input coin string params")
	}
	println("tokenInputs: ", tokenInputsParam)

	tokenInputs := make([]*privacy.InputCoin, len(tokenInputsParam))
	for i := 0; i < len(tokenInputs); i++ {
		tmp, ok := tokenInputsParam[i].(map[string]interface{})
		if !ok {
			println("Invalid input coin string param!")
			return nil, errors.New("Invalid input coin string param")
		}
		coinObjTmp := new(privacy.CoinObject)
		coinObjTmp.PublicKey, ok = tmp["PublicKey"].(string)
		if !ok {
			println("Invalid input coin public key param!")
			return nil, errors.New("Invalid input coin public key param")
		}
		coinObjTmp.CoinCommitment, ok = tmp["CoinCommitment"].(string)
		if !ok {
			println("Invalid input coin coin commitment param!")
			return nil, errors.New("Invalid input coin coin commitment param")
		}
		coinObjTmp.SNDerivator, ok = tmp["SNDerivator"].(string)
		if !ok {
			println("Invalid input coin snderivator param!")
			return nil, errors.New("Invalid input coin snderivator param")
		}
		coinObjTmp.SerialNumber, ok = tmp["SerialNumber"].(string)
		if !ok {
			println("Invalid input coin serial number param!")
			return nil, errors.New("Invalid input coin serial number param")
		}
		coinObjTmp.Randomness, ok = tmp["Randomness"].(string)
		if !ok {
			println("Invalid input coin randomness param!")
			return nil, errors.New("Invalid input coin randomness param")
		}
		coinObjTmp.Value, ok = tmp["Value"].(string)
		if !ok {
			println("Invalid input coin value param!")
			return nil, errors.New("Invalid input coin value param")
		}
		coinObjTmp.Info, ok = tmp["Info"].(string)
		if !ok {
			println("Invalid input coin info param!")
			return nil, errors.New("Invalid input coin info param")
		}

		tokenInputs[i] = new(privacy.InputCoin).Init()
		tokenInputs[i].ParseCoinObjectToInputCoin(*coinObjTmp)
	}

	println("tokenInputs: ", tokenInputs)
	privacyTokenParam.TokenInput = tokenInputs


	println("privacyTokenParam: ", len(privacyTokenParam.Receiver))
	println("privacyTokenParam.PropertyName: ", privacyTokenParam.PropertyName)
	println("privacyTokenParam.PropertySymbol: ", privacyTokenParam.PropertySymbol)
	println("privacyTokenParam.TokenInput: ", len(privacyTokenParam.TokenInput))
	println("privacyTokenParam.TokenTxType: ", privacyTokenParam.TokenTxType)
	println("privacyTokenParam.Amount: ", privacyTokenParam.Amount)
	//println("privacyTokenParam.PropertySymbol: ", len(privacyTokenParam.PropertySymbol))

	paramCreateTx := transaction.NewTxPrivacyTokenInitParamsForASM(
		&senderSK, paymentInfo, inputCoins, uint64(fee), privacyTokenParam, nil, hasPrivacy, hasPrivacyForPToken, shardID, infoBytes,
		commitmentIndicesForNativeToken, commitmentBytesForNativeToken, myCommitmentIndicesForNativeToken, sndOutputsForNativeToken,
		commitmentIndicesForPToken, commitmentBytesForPToken, myCommitmentIndicesForPToken, sndOutputsForPToken)
	println("paramCreateTx: ", paramCreateTx)

	return paramCreateTx, nil
}

func InitPrivacyTokenTx(args string) (string, error) {
	paramCreateTx, err := InitParamCreatePrivacyTokenTx(args)
	if err != nil{
		return "", err
	}

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

	tokenIDBytes := tx.TxPrivacyTokenData.PropertyID.GetBytes()

	lockTimeBytes := common.AddPaddingBigInt(new(big.Int).SetInt64(tx.LockTime), 8)
	resBytes := append(txJson, lockTimeBytes...)
	resBytes = append(resBytes, tokenIDBytes...)

	B64Res := base64.StdEncoding.EncodeToString(resBytes)

	return B64Res, nil
}

func InitBurningRequestTx(args string) (string, error) {
	// parse meta data
	bytes := []byte(args)
	println("Bytes: %v\n", bytes)

	paramMaps := make(map[string]interface{})

	err := json.Unmarshal(bytes, &paramMaps)
	if err != nil {
		println("Error can not unmarshal data : %v\n", err)
		return "", err
	}

	println("paramMaps:", paramMaps)

	metaDataParam, ok := paramMaps["metaData"].(map[string]interface{})
	if !ok {
		return "", errors.New("Invalid meta data param")
	}

	metaDataType, ok := metaDataParam["Type"].(float64)
	if !ok {
		println("Invalid meta data type param")
		return "", errors.New("Invalid meta data type param")
	}

	burnerAddressParam, ok := metaDataParam["BurnerAddress"].(string)
	if !ok {
		println("Invalid meta data burner payment address param")
		return "", errors.New("Invalid meta data burner payment address param")
	}
	keyWalletBurner, err := wallet.Base58CheckDeserialize(burnerAddressParam)
	if err != nil{
		return "", nil
	}
	burnerAddress := keyWalletBurner.KeySet.PaymentAddress

	burningAmount, ok := metaDataParam["BurningAmount"].(float64)
	if !ok {
		println("Invalid meta data burning amount param")
		return "", errors.New("Invalid meta data burning amount param")
	}
	tokenID, ok := metaDataParam["TokenID"].(string)
	if !ok {
		println("Invalid meta data token id param")
		return "", errors.New("Invalid meta data token id param")
	}
	tokenIDHash, err := new(common.Hash).NewHashFromStr(tokenID)
	if err != nil{
		return "", err
	}

	tokenName, ok := metaDataParam["TokenName"].(string)
	if !ok {
		println("Invalid meta data token name param")
		return "", errors.New("Invalid meta data token name param")
	}
	remoteAddress, ok := metaDataParam["RemoteAddress"].(string)
	if !ok {
		println("Invalid meta data remote address param")
		return "", errors.New("Invalid meta data remote address param")
	}

	metaData, err := metadata.NewBurningRequest(burnerAddress, uint64(burningAmount), *tokenIDHash, tokenName, remoteAddress, int(metaDataType))
	if err!= nil{
		return "", err
	}

	paramCreateTx, err := InitParamCreatePrivacyTokenTx(args)
	if err != nil{
		return "", err
	}

	paramCreateTx.SetMetaData(metaData)

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

	tokenIDBytes := tx.TxPrivacyTokenData.PropertyID.GetBytes()

	lockTimeBytes := common.AddPaddingBigInt(new(big.Int).SetInt64(tx.LockTime), 8)
	resBytes := append(txJson, lockTimeBytes...)
	resBytes = append(resBytes, tokenIDBytes...)

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


func InitWithdrawRewardTx(args string) (string, error) {
	// parse meta data
	bytes := []byte(args)
	println("Bytes: %v\n", bytes)

	paramMaps := make(map[string]interface{})

	err := json.Unmarshal(bytes, &paramMaps)
	if err != nil {
		println("Error can not unmarshal data : %v\n", err)
		return "", err
	}

	println("paramMaps:", paramMaps)

	metaDataParam, ok := paramMaps["metaData"].(map[string]interface{})
	if !ok {
		return "", errors.New("Invalid meta data param")
	}

	metaDataType, ok := metaDataParam["Type"].(float64)
	if !ok {
		println("Invalid meta data type param")
		return "", errors.New("Invalid meta data type param")
	}

	paymentAddressParam, ok := metaDataParam["PaymentAddress"].(string)
	if !ok {
		println("Invalid meta data payment address param")
		return "", errors.New("Invalid meta data payment address param")
	}
	keyWallet, err := wallet.Base58CheckDeserialize(paymentAddressParam)
	if err != nil{
		return "", nil
	}
	paymentAddress := keyWallet.KeySet.PaymentAddress



	tokenIDParam, ok := metaDataParam["TokenID"].(string)
	if !ok {
		println("Invalid meta data token id param")
		return "", errors.New("Invalid meta data token id param")
	}

	tokenId, err := new(common.Hash).NewHashFromStr(tokenIDParam)
	if err != nil {
		return "", err
	}

	tmp := &metadata.WithDrawRewardRequest{
		PaymentAddress: paymentAddress,
		MetadataBase: *metadata.NewMetadataBase(int(metaDataType)),
		TokenID: *tokenId,
	}

	paramCreateTx, err := InitParamCreatePrivacyTx(args)
	if err!= nil{
		return "", err
	}

	paramCreateTx.SetMetaData(tmp)

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