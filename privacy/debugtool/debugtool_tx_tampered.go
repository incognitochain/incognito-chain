package debugtool

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/privacy/coin"
	"github.com/incognitochain/incognito-chain/privacy/operation"
	"github.com/incognitochain/incognito-chain/privacy/privacy_v1/zeroknowledge/aggregatedrange"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
	"github.com/incognitochain/incognito-chain/transaction"
	"github.com/incognitochain/incognito-chain/wallet"
)

//Tx
func (tool *DebugTool) CreateTxNoPrivacyWithoutSignature(privateKey, tokenIDString, paymentString string, version int8) ([]byte, error) {
	keyWallet, _, pubkey, coinV1s, _, err := tool.PrepareTransaction(privateKey, tokenIDString)
	keySet := &keyWallet.KeySet

	tokenID, err := common.Hash{}.NewHashFromStr(tokenIDString)
	if err != nil {
		return nil, err
	}

	db, err := InitDatabase()
	if err != nil {
		return nil, err
	}

	if version == 1 {
		_, err := CreateAndSaveCoins(100, 0, keySet.PrivateKey, pubkey, db, 1, *tokenID)
		if err != nil {
			return nil, err
		}

		cmToBeSaved := [][]byte{}
		for _, output := range coinV1s {
			cmToBeSaved = append(cmToBeSaved, output.GetCommitment().ToBytesS())
		}

		err = statedb.StoreCommitments(db, *tokenID, cmToBeSaved, 0)
		if err != nil {
			return nil, err
		}

		inputCoins := []coin.PlainCoin{}
		for i := 0; i < len(coinV1s); i++ {
			tmpCoin := new(coin.PlainCoinV1)
			err := tmpCoin.SetBytes(coinV1s[i].Bytes())
			if err != nil {
				return nil, err
			}
			keyImage, err := tmpCoin.ParseKeyImageWithPrivateKey(keySet.PrivateKey)

			if err != nil {
				return nil, err
			}
			tmpCoin.SetKeyImage(keyImage)
			inputCoins = append(inputCoins, tmpCoin)
		}

		wallet, err := wallet.Base58CheckDeserialize(paymentString)
		if err != nil{
			return nil, err
		}

		fee := uint64(100)

		_, newTxPrivacyParam, err := CreateTxPrivacyInitParams(db, keySet, wallet.KeySet.PaymentAddress, inputCoins, false, fee, common.PRVCoinID)
		if err != nil {
			return nil, err
		}

		tx := new(transaction.TxVersion1)

		err = tx.Init(newTxPrivacyParam)
		if err != nil {
			return nil, err
		}

		//Remove the signature
		tx.Sig = nil

		txBytes, err := json.Marshal(tx)
		if err != nil {
			return nil, err
		}

		result := jsonresult.NewCreateTransactionResult(tx.Hash(), common.EmptyString, txBytes, 0)

		base58Result := result.Base58CheckData

		return []byte(base58Result), nil

	}

	return nil, nil

}

func (tool *DebugTool) CreateTxPrivacyWithoutSignature(privateKey, tokenIDString, paymentString string, version int8) ([]byte, error) {
	b, err := tool.CreateRawTx(privateKey, paymentString, uint64(10000))

	tx := new(transaction.TxVersion1)
	err = json.Unmarshal(b, &tx)
	if err != nil{
		tx2 := new(transaction.TxVersion2)
		err = json.Unmarshal(b, &tx2)
		if err != nil{
			return nil, err
		}

		fmt.Println("TxVersion 2")

		tx2.Sig = nil

		txBytes, err := json.Marshal(tx2)
		if err != nil {
			return nil, err
		}

		result := jsonresult.NewCreateTransactionResult(tx.Hash(), common.EmptyString, txBytes, 0)

		base58Result := result.Base58CheckData

		return []byte(base58Result), nil
	}

	fmt.Println("TxVersion 1")

	tx.Sig = nil

	txBytes, err := json.Marshal(tx)
	if err != nil {
		return nil, err
	}

	result := jsonresult.NewCreateTransactionResult(tx.Hash(), common.EmptyString, txBytes, 0)

	base58Result := result.Base58CheckData

	return []byte(base58Result), nil
}

func (tool *DebugTool) CreateTxPrivacyWithBulletProofCommitmentsTampered(privateKey, tokenIDString, paymentString string, version int8) ([]byte, error){
	keyWallet, senderPaymentAddress, _, coinV1s, _, err := tool.PrepareTransaction(privateKey, tokenIDString)
	keySet := &keyWallet.KeySet

	if err != nil {
		return nil, err
	}

	if version == 1{
		amount := uint64(RandIntInterval(0, 1000000000))
		coinsToSpend, err := ChooseCoinsToSpend(coinV1s, amount, false)
		if err != nil {
			return nil, err
		}

		inputCoins := []coin.PlainCoin{}
		for i := 0; i < len(coinsToSpend); i++ {
			tmpCoin := new(coin.PlainCoinV1)

			err = tmpCoin.SetBytes(coinsToSpend[i].Bytes())
			if err != nil {
				return nil, err
			}

			keyImage, err := tmpCoin.ParseKeyImageWithPrivateKey(keySet.PrivateKey)
			if err != nil {
				return nil, err
			}
			tmpCoin.SetKeyImage(keyImage)

			inputCoins = append(inputCoins, tmpCoin)
		}

		walletReceiver, err := wallet.Base58CheckDeserialize(paymentString)
		if err != nil{
			return nil, err
		}

		fee := uint64(100)

		paymentInfos, _, err := CreateTxPrivacyInitParams(nil, keySet, walletReceiver.KeySet.PaymentAddress, inputCoins, true, fee, common.PRVCoinID)
		if err != nil {
			return nil, err
		}

		tx := new(transaction.TxVersion1)

		//initializeTxAndParams
		InitParam(tx, fee, keySet, version)

		witness, err := tool.InitPaymentWitness(tokenIDString, senderPaymentAddress, inputCoins, paymentInfos, keySet, fee)
		if err != nil {
			return nil, err
		}

		paymentProof, err1 := witness.Prove(true, paymentInfos)
		if err1 != nil {
			return nil, err1
		}
		//Change commitments in the range proof
		bulletProof := paymentProof.GetAggregatedRangeProof()
		tmpBulletProof, ok := bulletProof.(*aggregatedrange.AggregatedRangeProof)
		if !ok {
			return nil, errors.New("cannot parse bullet proof")
		}

		cmsValue := tmpBulletProof.GetCommitments()
		cmsValue[0] = operation.RandomPoint()
		tmpBulletProof.SetCommitments(cmsValue)

		paymentProof.SetAggregatedRangeProof(tmpBulletProof)

		tx.Proof = paymentProof

		sigPrivate := append(keySet.PrivateKey, witness.GetRandSecretKey().ToBytesS()...)
		err = tx.Sign(sigPrivate)
		if err != nil {
			return nil, err
		}

		txBytes, err := json.Marshal(tx)
		if err != nil {
			return nil, err
		}

		result := jsonresult.NewCreateTransactionResult(tx.Hash(), common.EmptyString, txBytes, 0)

		base58Result := result.Base58CheckData

		return []byte(base58Result), nil
	}

	return nil, nil
}

func (tool *DebugTool) CreateTxPrivacyWithSNProofCommitmentTampered(privateKey, tokenIDString, paymentString string) ([]byte, error) {
	keyWallet, senderPaymentAddress, _, coinV1s, _, err := tool.PrepareTransaction(privateKey, tokenIDString)
	keySet := &keyWallet.KeySet

	if err != nil {
		return nil, err
	}

	amount := uint64(RandIntInterval(0, 1000000000))
	coinsToSpend, err := ChooseCoinsToSpend(coinV1s, amount, false)
	if err != nil {
		return nil, err
	}

	inputCoins := []coin.PlainCoin{}
	for i := 0; i < len(coinsToSpend); i++ {
		tmpCoin := new(coin.PlainCoinV1)

		err = tmpCoin.SetBytes(coinsToSpend[i].Bytes())
		if err != nil {
			return nil, err
		}

		keyImage, err := tmpCoin.ParseKeyImageWithPrivateKey(keySet.PrivateKey)
		if err != nil {
			return nil, err
		}
		tmpCoin.SetKeyImage(keyImage)

		inputCoins = append(inputCoins, tmpCoin)
	}

	walletReceiver, err := wallet.Base58CheckDeserialize(paymentString)
	if err != nil {
		return nil, err
	}

	fee := uint64(100)

	tx := new(transaction.TxVersion1)

	paymentInfos, _, err := CreateTxPrivacyInitParams(nil, keySet, walletReceiver.KeySet.PaymentAddress, inputCoins, true, fee, common.PRVCoinID)
	if err != nil {
		return nil, err
	}

	//initializeTxAndParams
	InitParam(tx, fee, keySet, 1)

	witness, err := tool.InitPaymentWitness(tokenIDString, senderPaymentAddress, inputCoins, paymentInfos, keySet, fee)
	if err != nil {
		return nil, err
	}

	paymentProof, err1 := witness.Prove(true, paymentInfos)
	if err1 != nil {
		return nil, err1
	}

	//Tamper with the serial number privacy proof
	snPrivacyProof := paymentProof.GetSerialNumberProof()
	snPrivacyProof[0] = snPrivacyProof[1]
	paymentProof.SetSerialNumberProof(snPrivacyProof)

	tx.Proof = paymentProof

	sigPrivate := append(keySet.PrivateKey, witness.GetRandSecretKey().ToBytesS()...)
	err = tx.Sign(sigPrivate)
	if err != nil {
		return nil, err
	}

	txBytes, err := json.Marshal(tx)
	if err != nil {
		return nil, err
	}

	result := jsonresult.NewCreateTransactionResult(tx.Hash(), common.EmptyString, txBytes, 0)

	base58Result := result.Base58CheckData

	return []byte(base58Result), nil
}

func (tool *DebugTool) CreateTxPrivacyWithOneOfManyProofTampered(privateKey, tokenIDString, paymentString string) ([]byte, error) {
	keyWallet, senderPaymentAddress, _, coinV1s, _, err := tool.PrepareTransaction(privateKey, tokenIDString)
	keySet := &keyWallet.KeySet

	if err != nil {
		return nil, err
	}

	amount := uint64(RandIntInterval(0, 1000000000))
	coinsToSpend, err := ChooseCoinsToSpend(coinV1s, amount, false)
	if err != nil {
		return nil, err
	}

	inputCoins := []coin.PlainCoin{}
	for i := 0; i < len(coinsToSpend); i++ {
		tmpCoin := new(coin.PlainCoinV1)

		err = tmpCoin.SetBytes(coinsToSpend[i].Bytes())
		if err != nil {
			return nil, err
		}

		keyImage, err := tmpCoin.ParseKeyImageWithPrivateKey(keySet.PrivateKey)
		if err != nil {
			return nil, err
		}
		tmpCoin.SetKeyImage(keyImage)

		inputCoins = append(inputCoins, tmpCoin)
	}

	walletReceiver, err := wallet.Base58CheckDeserialize(paymentString)
	if err != nil {
		return nil, err
	}

	fee := uint64(100)

	tx := new(transaction.TxVersion1)

	paymentInfos, _, err := CreateTxPrivacyInitParams(nil, keySet, walletReceiver.KeySet.PaymentAddress, inputCoins, true, fee, common.PRVCoinID)
	if err != nil {
		return nil, err
	}

	//initializeTxAndParams
	InitParam(tx, fee, keySet, 1)

	witness, err := tool.InitPaymentWitness(tokenIDString, senderPaymentAddress, inputCoins, paymentInfos, keySet, fee)
	if err != nil {
		return nil, err
	}

	paymentProof, err1 := witness.Prove(true, paymentInfos)
	if err1 != nil {
		return nil, err1
	}

	//Tamper with the one-of-many proof
	oneOfManyProof := paymentProof.GetOneOfManyProof()
	oneOfManyProof[0].Statement.Commitments[0] = operation.RandomPoint()
	paymentProof.SetOneOfManyProof(oneOfManyProof)

	tx.Proof = paymentProof

	sigPrivate := append(keySet.PrivateKey, witness.GetRandSecretKey().ToBytesS()...)
	err = tx.Sign(sigPrivate)
	if err != nil {
		return nil, err
	}

	txBytes, err := json.Marshal(tx)
	if err != nil {
		return nil, err
	}

	result := jsonresult.NewCreateTransactionResult(tx.Hash(), common.EmptyString, txBytes, 0)

	base58Result := result.Base58CheckData

	return []byte(base58Result), nil


}

//TxToken
func (tool *DebugTool) CreateTxTokenPrivacyWithoutSignature(privateKey, tokenIDString, paymentString string) ([]byte, error) {
	b, err := tool.CreateRawTxToken(privateKey, tokenIDString, paymentString, uint64(100), true)

	tx := new(transaction.TxTokenVersion1)
	err = json.Unmarshal(b, &tx)
	if err != nil{
		return nil, err
	}

	r := common.RandInt()
	if r % 3 == 0{
		tx.Tx.SetSig(nil)
	}else if r % 3 == 1{
		tx.TxTokenData.TxNormal.SetSig(nil)
	}else{
		tx.Tx.SetSig(nil)
		tx.TxTokenData.TxNormal.SetSig(nil)
	}

	txBytes, err := json.Marshal(tx)
	if err != nil {
		return nil, err
	}

	result := jsonresult.NewCreateTransactionResult(tx.Hash(), common.EmptyString, txBytes, 0)

	base58Result := result.Base58CheckData

	return []byte(base58Result), nil
}

////
func (tool *DebugTool) SendTxNoPrivacyFake(privateKey, tokenIDString, paymentString string, txType int64, version int8) ([]byte, error){
	var base58Bytes []byte

	switch txType{
	case 0: //Create txprivacy without signatures
		base58Bytes, err := tool.CreateTxNoPrivacyWithoutSignature(privateKey, tokenIDString, paymentString, version)
		if err!=nil {
			return nil, err
		}

		query := fmt.Sprintf(`{
			"jsonrpc": "1.0",
			"method": "sendtransaction",
			"params": [
				"%s"
			],
			"id": 1
		}`, string(base58Bytes))

		return tool.SendPostRequestWithQuery(query)
	default:
	}

	query := fmt.Sprintf(`{
		"jsonrpc": "1.0",
		"method": "sendtransaction",
		"params": [
			"%s"
		],
		"id": 1
	}`, string(base58Bytes))

	return tool.SendPostRequestWithQuery(query)
}

func (tool *DebugTool) SendTxPrivacyFake(privateKey, tokenIDString, paymentString string, txType int64, version int8) ([]byte, error){
	if tokenIDString != common.PRVIDStr{
		base58Bytes, err := tool.CreateTxTokenPrivacyWithoutSignature(privateKey, tokenIDString, paymentString)
		if err != nil{
			return nil, err
		}
		query := fmt.Sprintf(`{
			"jsonrpc": "1.0",
			"method": "sendrawprivacycustomtokentransaction",
			"params": [
				"%s"
			],
			"id": 1
		}`, string(base58Bytes))

		return tool.SendPostRequestWithQuery(query)
	}

	switch txType{
	case 0: //Create txprivacy without signatures
		base58Bytes, err := tool.CreateTxPrivacyWithoutSignature(privateKey, tokenIDString, paymentString, version)
		if err!=nil {
			return nil, err
		}

		query := fmt.Sprintf(`{
			"jsonrpc": "1.0",
			"method": "sendtransaction",
			"params": [
				"%s"
			],
			"id": 1
		}`, string(base58Bytes))

		return tool.SendPostRequestWithQuery(query)
	case 1://Create txprivacy with the bulletProof tampered
		base58Bytes, err := tool.CreateTxPrivacyWithBulletProofCommitmentsTampered(privateKey, tokenIDString, paymentString, version)
		if err!=nil {
			return nil, err
		}

		query := fmt.Sprintf(`{
			"jsonrpc": "1.0",
			"method": "sendtransaction",
			"params": [
				"%s"
			],
			"id": 1
		}`, string(base58Bytes))

		return tool.SendPostRequestWithQuery(query)
	case 2://Create txprivacy with snPrivacyProof tampered
		base58Bytes, err := tool.CreateTxPrivacyWithSNProofCommitmentTampered(privateKey, tokenIDString, paymentString)
		if err!=nil {
			return nil, err
		}

		query := fmt.Sprintf(`{
			"jsonrpc": "1.0",
			"method": "sendtransaction",
			"params": [
				"%s"
			],
			"id": 1
		}`, string(base58Bytes))

		return tool.SendPostRequestWithQuery(query)
	case 3://Create txprivacy with one-of-many proofs tampered
		base58Bytes, err := tool.CreateTxPrivacyWithOneOfManyProofTampered(privateKey, tokenIDString, paymentString)
		if err!=nil {
			return nil, err
		}

		query := fmt.Sprintf(`{
			"jsonrpc": "1.0",
			"method": "sendtransaction",
			"params": [
				"%s"
			],
			"id": 1
		}`, string(base58Bytes))

		return tool.SendPostRequestWithQuery(query)
	default:
		return nil, errors.New("Wrong txType")
	}
}

func (tool *DebugTool) CreateRawTxToken(privateKey, tokenIDString, paymentString string, amount uint64, isPrivacy bool) ([]byte, error){
	query := fmt.Sprintf(`{
		"id": 1,
		"jsonrpc": "1.0",
		"method": "createrawprivacycustomtokentransaction",
		"params": [
			"%s",
			null,
			10,
			1,
			{
				"Privacy": %v,
				"TokenID": "%s",
				"TokenName": "",
				"TokenSymbol": "",
				"TokenFee": 0,
				"TokenTxType": 1,
				"TokenAmount": 0,
				"TokenReceivers": {
					"%s": %d
				}
			}
			]
	}`, privateKey, isPrivacy, tokenIDString, paymentString, amount)

	respondInBytes, err := tool.SendPostRequestWithQuery(query)
	if err != nil {
		return nil, err
	}

	respond, err := ParseResponse(respondInBytes)
	if err != nil{
		return nil, err
	}

	if respond.Error != nil {
		return nil, respond.Error
	}

	var msg json.RawMessage
	err = json.Unmarshal(respond.Result, &msg)

	var result map[string]interface{}
	err = json.Unmarshal(msg, &result)

	base58Check, ok := result["Base58CheckData"]
	if !ok {
		fmt.Println(result)
		return nil, errors.New("cannot find base58CheckData")
	}

	tmp, _ := base58Check.(string)

	bytearrays, err := DecodeBase58Check(tmp)
	if err != nil{
		return nil, err
	}


	return bytearrays, nil
}

func (tool *DebugTool) CreateRawTx(privateKey, paymentString string, amount uint64) ([]byte, error){
	query := fmt.Sprintf(`{
		"jsonrpc": "1.0",
		"method": "createtransaction",
		"params": [
			"%s",
			{
				"%s":%d
			},
			1,
			1
		],
		"id": 1
	}`, privateKey, paymentString, amount)

	respondInBytes, err := tool.SendPostRequestWithQuery(query)
	if err != nil {
		return nil, err
	}

	respond, err := ParseResponse(respondInBytes)
	if err != nil{
		return nil, err
	}

	if respond.Error != nil {
		return nil, respond.Error
	}

	var msg json.RawMessage
	err = json.Unmarshal(respond.Result, &msg)

	var result map[string]interface{}
	err = json.Unmarshal(msg, &result)

	base58Check, ok := result["Base58CheckData"]
	if !ok {
		fmt.Println(result)
		return nil, errors.New("cannot find base58CheckData")
	}

	tmp, _ := base58Check.(string)

	bytearrays, err := DecodeBase58Check(tmp)
	if err != nil{
		return nil, err
	}


	return bytearrays, nil
}