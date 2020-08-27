package debugtool

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/privacy/coin"
	"github.com/incognitochain/incognito-chain/privacy/operation"
	zkp "github.com/incognitochain/incognito-chain/privacy/privacy_v1/zeroknowledge"
	"github.com/incognitochain/incognito-chain/privacy/privacy_v1/zeroknowledge/aggregatedrange"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
	"github.com/incognitochain/incognito-chain/transaction"
	"github.com/incognitochain/incognito-chain/wallet"
)

//For both txnormal ver1 + ver2
func (tool *DebugTool) CreateTxNoPrivacyWithoutSignature(privateKey, tokenIDString, paymentString string, version int8) ([]byte, error) {
	keyWallet, _, pubkey, coinV1s, _, _, _, err := tool.PrepareTransaction(privateKey, tokenIDString)
	keySet := &keyWallet.KeySet

	tokenID, err := common.Hash{}.NewHashFromStr(tokenIDString)
	if err != nil {
		return nil, err
	}

	db, err := InitDatabase()
	if err != nil {
		return nil, err
	}

	amount := uint64(RandIntInterval(0, 1000000000))
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
		if err != nil {
			return nil, err
		}

		fee := uint64(100)

		_, newTxPrivacyParam, err := CreateTxPrivacyInitParams(db, keySet, wallet.KeySet.PaymentAddress, inputCoins, false, fee, amount, common.PRVCoinID)
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
	b, err := tool.CreateRawTx(privateKey, paymentString, uint64(10000), true)

	tx := new(transaction.TxVersion1)
	err = json.Unmarshal(b, &tx)
	if err != nil {
		tx2 := new(transaction.TxVersion2)
		err = json.Unmarshal(b, &tx2)
		if err != nil {
			return nil, err
		}

		tx2.Sig = nil

		txBytes, err := json.Marshal(tx2)
		if err != nil {
			return nil, err
		}

		result := jsonresult.NewCreateTransactionResult(tx.Hash(), common.EmptyString, txBytes, 0)

		base58Result := result.Base58CheckData

		return []byte(base58Result), nil
	}


	tx.Sig = nil

	txBytes, err := json.Marshal(tx)
	if err != nil {
		return nil, err
	}

	result := jsonresult.NewCreateTransactionResult(tx.Hash(), common.EmptyString, txBytes, 0)

	base58Result := result.Base58CheckData

	return []byte(base58Result), nil
}

func (tool *DebugTool) CreateTxPrivacyWithBulletProofCommitmentsTampered(privateKey, tokenIDString, paymentString string, version int8) ([]byte, error) {
	b, err := tool.CreateRawTx(privateKey, paymentString, uint64(10000), true)

	keyWallet, err  := wallet.Base58CheckDeserialize(privateKey)
	if err != nil{
		return nil, err
	}

	keySet := &keyWallet.KeySet

	tx := new(transaction.TxVersion1)
	err = json.Unmarshal(b, &tx)
	if err != nil {
		tx2 := new(transaction.TxVersion2)
		err = json.Unmarshal(b, &tx2)
		if err != nil {
			return nil, err
		}

		tx.Sig = nil

		txBytes, err := json.Marshal(tx2)
		if err != nil {
			return nil, err
		}

		result := jsonresult.NewCreateTransactionResult(tx.Hash(), common.EmptyString, txBytes, 0)

		base58Result := result.Base58CheckData

		return []byte(base58Result), nil
	}

	paymentProof, ok := tx.Proof.(*zkp.PaymentProof)
	if !ok{
		return nil, errors.New("cannot parse paymentproof")
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

	sigPrivate := append(keySet.PrivateKey, []byte{0,0,0,0}...)

	tx.Sig = nil
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

//func (tool *DebugTool) CreateTxPrivacyWithBulletProofCommitmentsTampered(privateKey, tokenIDString, paymentString string, version int8) ([]byte, error) {
//	keyWallet, senderPaymentAddress, _, coinV1s, coinV2s, listIndicesV1, listIndicesV2, err := tool.PrepareTransaction(privateKey, tokenIDString)
//	if err != nil {
//		return nil, err
//	}
//
//	keySet := &keyWallet.KeySet
//
//	walletReceiver, err := wallet.Base58CheckDeserialize(paymentString)
//	if err != nil {
//		return nil, err
//	}
//
//	amount := uint64(RandIntInterval(0, 1000000000))
//
//	if version == 1 {
//		_, inputCoins, err := PrepareInputCoins(coinV1s, amount, listIndicesV1, keySet)
//		if err != nil {
//			return nil, err
//		}
//
//		fee := uint64(100)
//
//		paymentInfos, _, err := CreateTxPrivacyInitParams(nil, keySet, walletReceiver.KeySet.PaymentAddress, inputCoins, true, fee, amount, common.PRVCoinID)
//		if err != nil {
//			return nil, err
//		}
//
//		tx := new(transaction.TxVersion1)
//
//		//initializeTxAndParams
//		InitParam(tx, fee, keySet, version)
//
//		witness, err := tool.InitPaymentWitness(tokenIDString, senderPaymentAddress, inputCoins, paymentInfos, keySet, fee)
//		if err != nil {
//			return nil, err
//		}
//
//		paymentProof, err1 := witness.Prove(true, paymentInfos)
//		if err1 != nil {
//			return nil, err1
//		}
//		//Change commitments in the range proof
//		bulletProof := paymentProof.GetAggregatedRangeProof()
//		tmpBulletProof, ok := bulletProof.(*aggregatedrange.AggregatedRangeProof)
//		if !ok {
//			return nil, errors.New("cannot parse bullet proof")
//		}
//
//		cmsValue := tmpBulletProof.GetCommitments()
//		cmsValue[0] = operation.RandomPoint()
//		tmpBulletProof.SetCommitments(cmsValue)
//
//		paymentProof.SetAggregatedRangeProof(tmpBulletProof)
//
//		tx.Proof = paymentProof
//
//		sigPrivate := append(keySet.PrivateKey, witness.GetRandSecretKey().ToBytesS()...)
//		err = tx.Sign(sigPrivate)
//		if err != nil {
//			return nil, err
//		}
//
//		txBytes, err := json.Marshal(tx)
//		if err != nil {
//			return nil, err
//		}
//
//		result := jsonresult.NewCreateTransactionResult(tx.Hash(), common.EmptyString, txBytes, 0)
//
//		base58Result := result.Base58CheckData
//
//		return []byte(base58Result), nil
//	} else {
//		//Create new transaction and init param
//		tx := new(transaction.TxVersion2)
//
//		myListIndicesV2, inputCoins, err := PrepareInputCoins(coinV2s, amount, listIndicesV2, nil)
//		if err != nil {
//			return nil, err
//		}
//
//		fee := uint64(100)
//
//		paymentInfos, _, err := CreateTxPrivacyInitParams(nil, keySet, walletReceiver.KeySet.PaymentAddress, inputCoins, true, fee, amount, common.PRVCoinID)
//		if err != nil {
//			return nil, err
//		}
//
//		outputCoins, err := ParseOutputCoinV2s(paymentInfos)
//		if err != nil {
//			return nil, err
//		}
//
//		InitParam(tx, fee, keySet, version)
//
//		//Create bulletproof
//		tx.Proof, err = privacy_v2.Prove(inputCoins, outputCoins, true, paymentInfos)
//		if err != nil {
//			return nil, err
//		}
//
//		//Attempt to alter the bulletproof
//		//Change commitments in the range proof
//		paymentProof, ok := tx.Proof.(*privacy_v2.PaymentProofV2)
//		if !ok {
//			return nil, errors.New("cannot parse PaymentProofV2")
//		}
//
//		bulletProof := paymentProof.GetAggregatedRangeProof()
//		tmpBulletProof, ok := bulletProof.(*bulletproofs.AggregatedRangeProof)
//		if !ok {
//			return nil, errors.New("cannot parse bullet proof")
//		}
//
//		cmsValue := tmpBulletProof.GetCommitments()
//		cmsValue[0] = operation.RandomPoint()
//		tmpBulletProof.SetCommitments(cmsValue)
//
//		paymentProof.SetAggregatedRangeProof(tmpBulletProof)
//
//		//Sign transaction by mlsag
//		err2 := tool.SignTransactionV2(tx, tokenIDString, senderPaymentAddress, inputCoins, outputCoins, fee, myListIndicesV2, keySet)
//		if err2 != nil {
//			return nil, err2
//		}
//
//		return MarshalTransaction(tx)
//	}
//}

//TxVer1 Only
func (tool *DebugTool) CreateTxPrivacyWithSNProofCommitmentTampered(privateKey, tokenIDString, paymentString string) ([]byte, error) {
	b, err := tool.CreateRawTx(privateKey, paymentString, uint64(10000), true)

	keyWallet, err  := wallet.Base58CheckDeserialize(privateKey)
	if err != nil{
		return nil, err
	}

	keySet := &keyWallet.KeySet

	tx := new(transaction.TxVersion1)
	err = json.Unmarshal(b, &tx)
	if err != nil {
		tx2 := new(transaction.TxVersion2)
		err = json.Unmarshal(b, &tx2)
		if err != nil {
			return nil, err
		}

		tx.Sig = nil

		txBytes, err := json.Marshal(tx2)
		if err != nil {
			return nil, err
		}

		result := jsonresult.NewCreateTransactionResult(tx.Hash(), common.EmptyString, txBytes, 0)

		base58Result := result.Base58CheckData

		return []byte(base58Result), nil
	}

	paymentProof, ok := tx.Proof.(*zkp.PaymentProof)
	if !ok{
		return nil, errors.New("cannot parse paymentproof")
	}

	//Tamper with the serial number privacy proof
	snPrivacyProof := paymentProof.GetSerialNumberProof()

	snPrivacyProofBytes := snPrivacyProof[0].Bytes()
	for {
		r := RandIntInterval(0, len(snPrivacyProofBytes)-1)
		tmpByte := snPrivacyProofBytes[r]
		tmpRandByte := uint8(common.RandIntInterval(0, 256))
		if tmpByte == tmpRandByte{
			continue
		}

		snPrivacyProofBytes[r] = tmpRandByte
		err = snPrivacyProof[0].SetBytes(snPrivacyProofBytes)
		if err != nil{
			continue
		}else{
			break
		}
	}

	paymentProof.SetSerialNumberProof(snPrivacyProof)

	tx.Proof = paymentProof

	sigPrivate := append(keySet.PrivateKey, []byte{0,0,0,0}...)

	tx.Sig = nil
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
	b, err := tool.CreateRawTx(privateKey, paymentString, uint64(10000), true)

	keyWallet, err  := wallet.Base58CheckDeserialize(privateKey)
	if err != nil{
		return nil, err
	}

	keySet := &keyWallet.KeySet

	tx := new(transaction.TxVersion1)
	err = json.Unmarshal(b, &tx)
	if err != nil {
		tx2 := new(transaction.TxVersion2)
		err = json.Unmarshal(b, &tx2)
		if err != nil {
			return nil, err
		}

		tx.Sig = nil

		txBytes, err := json.Marshal(tx2)
		if err != nil {
			return nil, err
		}

		result := jsonresult.NewCreateTransactionResult(tx.Hash(), common.EmptyString, txBytes, 0)

		base58Result := result.Base58CheckData

		return []byte(base58Result), nil
	}

	paymentProof, ok := tx.Proof.(*zkp.PaymentProof)
	if !ok{
		return nil, errors.New("cannot parse paymentproof")
	}

	//Tamper with the one-of-many proof
	oneOfManyProof := paymentProof.GetOneOfManyProof()
	if len(oneOfManyProof) > 0{
		oneOfManyProof[0].Statement = nil
		paymentProof.SetOneOfManyProof(oneOfManyProof)
		tx.Proof = paymentProof
	}




	sigPrivate := append(keySet.PrivateKey, []byte{0,0,0,0}...)

	tx.Sig = nil
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

//func (tool *DebugTool) CreateTxPrivacyWithSNProofCommitmentTampered(privateKey, tokenIDString, paymentString string) ([]byte, error) {
//	keyWallet, senderPaymentAddress, _, coinV1s, _, _, _, err := tool.PrepareTransaction(privateKey, tokenIDString)
//	if err != nil {
//		return nil, err
//	}
//	keySet := &keyWallet.KeySet
//
//	amount := uint64(RandIntInterval(0, 1000000000))
//	coinsToSpend, err := ChooseCoinsToSpend(coinV1s, amount, false)
//	if err != nil {
//		return nil, err
//	}
//
//	inputCoins := []coin.PlainCoin{}
//	for i := 0; i < len(coinsToSpend); i++ {
//		tmpCoin := new(coin.PlainCoinV1)
//
//		err = tmpCoin.SetBytes(coinsToSpend[i].Bytes())
//		if err != nil {
//			return nil, err
//		}
//
//		keyImage, err := tmpCoin.ParseKeyImageWithPrivateKey(keySet.PrivateKey)
//		if err != nil {
//			return nil, err
//		}
//		tmpCoin.SetKeyImage(keyImage)
//
//		inputCoins = append(inputCoins, tmpCoin)
//	}
//
//	walletReceiver, err := wallet.Base58CheckDeserialize(paymentString)
//	if err != nil {
//		return nil, err
//	}
//
//	fee := uint64(100)
//
//	tx := new(transaction.TxVersion1)
//
//	paymentInfos, _, err := CreateTxPrivacyInitParams(nil, keySet, walletReceiver.KeySet.PaymentAddress, inputCoins, true, fee, amount, common.PRVCoinID)
//	if err != nil {
//		return nil, err
//	}
//
//	//initializeTxAndParams
//	InitParam(tx, fee, keySet, 1)
//
//	witness, err := tool.InitPaymentWitness(tokenIDString, senderPaymentAddress, inputCoins, paymentInfos, keySet, fee)
//	if err != nil {
//		return nil, err
//	}
//
//	paymentProof, err1 := witness.Prove(true, paymentInfos)
//	if err1 != nil {
//		return nil, err1
//	}
//
//	//Tamper with the serial number privacy proof
//	snPrivacyProof := paymentProof.GetSerialNumberProof()
//
//	snPrivacyProofBytes := snPrivacyProof[0].Bytes()
//	for {
//		r := RandIntInterval(0, len(snPrivacyProofBytes)-1)
//		tmpByte := snPrivacyProofBytes[r]
//		tmpRandByte := uint8(common.RandIntInterval(0, 256))
//		if tmpByte == tmpRandByte{
//			continue
//		}
//
//		snPrivacyProofBytes[r] = tmpRandByte
//		err = snPrivacyProof[0].SetBytes(snPrivacyProofBytes)
//		if err != nil{
//			continue
//		}else{
//			break
//		}
//	}
//
//	paymentProof.SetSerialNumberProof(snPrivacyProof)
//	tx.Proof = paymentProof
//
//	sigPrivate := append(keySet.PrivateKey, witness.GetRandSecretKey().ToBytesS()...)
//	err = tx.Sign(sigPrivate)
//	if err != nil {
//		return nil, err
//	}
//
//	txBytes, err := json.Marshal(tx)
//	if err != nil {
//		return nil, err
//	}
//
//	result := jsonresult.NewCreateTransactionResult(tx.Hash(), common.EmptyString, txBytes, 0)
//
//	base58Result := result.Base58CheckData
//
//	return []byte(base58Result), nil
//}

//func (tool *DebugTool) CreateTxPrivacyWithOneOfManyProofTampered(privateKey, tokenIDString, paymentString string) ([]byte, error) {
//	keyWallet, senderPaymentAddress, _, coinV1s, _, _, _, err := tool.PrepareTransaction(privateKey, tokenIDString)
//	keySet := &keyWallet.KeySet
//
//	if err != nil {
//		return nil, err
//	}
//
//	amount := uint64(RandIntInterval(0, 1000000000))
//	coinsToSpend, err := ChooseCoinsToSpend(coinV1s, amount, false)
//	if err != nil {
//		return nil, err
//	}
//
//	inputCoins := []coin.PlainCoin{}
//	for i := 0; i < len(coinsToSpend); i++ {
//		tmpCoin := new(coin.PlainCoinV1)
//
//		err = tmpCoin.SetBytes(coinsToSpend[i].Bytes())
//		if err != nil {
//			return nil, err
//		}
//
//		keyImage, err := tmpCoin.ParseKeyImageWithPrivateKey(keySet.PrivateKey)
//		if err != nil {
//			return nil, err
//		}
//		tmpCoin.SetKeyImage(keyImage)
//
//		inputCoins = append(inputCoins, tmpCoin)
//	}
//
//	walletReceiver, err := wallet.Base58CheckDeserialize(paymentString)
//	if err != nil {
//		return nil, err
//	}
//
//	fee := uint64(100)
//
//	tx := new(transaction.TxVersion1)
//
//	paymentInfos, _, err := CreateTxPrivacyInitParams(nil, keySet, walletReceiver.KeySet.PaymentAddress, inputCoins, true, fee, amount, common.PRVCoinID)
//	if err != nil {
//		return nil, err
//	}
//
//	//initializeTxAndParams
//	InitParam(tx, fee, keySet, 1)
//
//	witness, err := tool.InitPaymentWitness(tokenIDString, senderPaymentAddress, inputCoins, paymentInfos, keySet, fee)
//	if err != nil {
//		return nil, err
//	}
//
//	paymentProof, err1 := witness.Prove(true, paymentInfos)
//	if err1 != nil {
//		return nil, err1
//	}
//
//	//Tamper with the one-of-many proof
//	oneOfManyProof := paymentProof.GetOneOfManyProof()
//	oneOfManyProof[0].Statement.Commitments[0] = operation.RandomPoint()
//	paymentProof.SetOneOfManyProof(oneOfManyProof)
//
//	tx.Proof = paymentProof
//
//	sigPrivate := append(keySet.PrivateKey, witness.GetRandSecretKey().ToBytesS()...)
//	err = tx.Sign(sigPrivate)
//	if err != nil {
//		return nil, err
//	}
//
//	txBytes, err := json.Marshal(tx)
//	if err != nil {
//		return nil, err
//	}
//
//	result := jsonresult.NewCreateTransactionResult(tx.Hash(), common.EmptyString, txBytes, 0)
//
//	base58Result := result.Base58CheckData
//
//	return []byte(base58Result), nil
//}

//TxTokenVer1
func (tool *DebugTool) CreateTxTokenPrivacyWithoutSignature(privateKey, tokenIDString, paymentString string) ([]byte, error) {
	b, err := tool.CreateRawTxToken(privateKey, tokenIDString, paymentString, uint64(100), true)

	tx := new(transaction.TxTokenVersion1)
	err = json.Unmarshal(b, &tx)
	if err != nil {
		return nil, err
	}

	r := common.RandInt()
	if r%3 == 0 {
		tx.Tx.SetSig(nil)
	} else if r%3 == 1 {
		tx.TxTokenData.TxNormal.SetSig(nil)
	} else {
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

func (tool *DebugTool) CreateTxTokenPrivacyWithBulletProofCommitmentsTampered(privateKey, tokenIDString, paymentString string)([]byte, error){
	b, err := tool.CreateRawTxToken(privateKey, tokenIDString, paymentString, uint64(100), true)

	tx := new(transaction.TxTokenVersion1)
	err = json.Unmarshal(b, &tx)
	if err != nil {
		return nil, err
	}

	keyWallet, err := wallet.Base58CheckDeserialize(privateKey)
	if err != nil{
		return nil, err
	}

	keySet := &keyWallet.KeySet


	r := common.RandInt()
	if r%2 == 0 {
		txFee, ok  := tx.Tx.(*transaction.TxVersion1)
		if !ok {
			return nil, errors.New("cannot parse txfee")
		}

		paymentProof, ok := txFee.Proof.(*zkp.PaymentProof)
		if !ok {
			return nil, errors.New("cannot parse payment proof of txfee")
		}

		bulletProof := paymentProof.GetAggregatedRangeProof()
		tmpBulletProof, ok := bulletProof.(*aggregatedrange.AggregatedRangeProof)
		if !ok {
			return nil, errors.New("cannot parse bullet proof")
		}
		cmsValue := tmpBulletProof.GetCommitments()
		cmsValue[0] = operation.RandomPoint()
		tmpBulletProof.SetCommitments(cmsValue)
		paymentProof.SetAggregatedRangeProof(tmpBulletProof)

		txFee.Proof = paymentProof

		sigPrivate := append(keySet.PrivateKey, []byte{0,0,0,0}...)

		err = txFee.Sign(sigPrivate)
		if err != nil {
			return nil, err
		}



	} else {
		txNormal, ok  := tx.TxTokenData.TxNormal.(*transaction.TxVersion1)
		if !ok {
			return nil, errors.New("cannot parse txfee")
		}

		paymentProof, ok := txNormal.Proof.(*zkp.PaymentProof)
		if !ok {
			return nil, errors.New("cannot parse payment proof of txfee")
		}

		bulletProof := paymentProof.GetAggregatedRangeProof()
		tmpBulletProof, ok := bulletProof.(*aggregatedrange.AggregatedRangeProof)
		if !ok {
			return nil, errors.New("cannot parse bullet proof")
		}
		cmsValue := tmpBulletProof.GetCommitments()
		cmsValue[0] = operation.RandomPoint()
		tmpBulletProof.SetCommitments(cmsValue)
		paymentProof.SetAggregatedRangeProof(tmpBulletProof)

		txNormal.Proof = paymentProof

		sigPrivate := append(keySet.PrivateKey, []byte{0,0,0,0}...)

		err = txNormal.Sign(sigPrivate)
		if err != nil {
			return nil, err
		}
	}

	txBytes, err := json.Marshal(tx)
	if err != nil {
		return nil, err
	}

	result := jsonresult.NewCreateTransactionResult(tx.Hash(), common.EmptyString, txBytes, 0)

	base58Result := result.Base58CheckData

	return []byte(base58Result), nil
}

func (tool *DebugTool) CreateTxTokenPrivacyWithSNProofCommitmentTampered(privateKey, tokenIDString, paymentString string)([]byte, error){
	b, err := tool.CreateRawTxToken(privateKey, tokenIDString, paymentString, uint64(100), true)

	tx := new(transaction.TxTokenVersion1)
	err = json.Unmarshal(b, &tx)
	if err != nil {
		return nil, err
	}

	keyWallet, err := wallet.Base58CheckDeserialize(privateKey)
	if err != nil{
		return nil, err
	}

	keySet := &keyWallet.KeySet

	r := common.RandInt()
	if r%2 == 0 {
		txFee, ok  := tx.Tx.(*transaction.TxVersion1)
		if !ok {
			return nil, errors.New("cannot parse txfee")
		}

		paymentProof, ok := txFee.Proof.(*zkp.PaymentProof)
		if !ok {
			return nil, errors.New("cannot parse payment proof of txfee")
		}

		//Tamper with the serial number privacy proof
		snPrivacyProof := paymentProof.GetSerialNumberProof()

		snPrivacyProofBytes := snPrivacyProof[0].Bytes()
		for {
			r := RandIntInterval(0, len(snPrivacyProofBytes)-1)
			tmpByte := snPrivacyProofBytes[r]
			tmpRandByte := uint8(common.RandIntInterval(0, 256))
			if tmpByte == tmpRandByte{
				continue
			}

			snPrivacyProofBytes[r] = tmpRandByte
			err = snPrivacyProof[0].SetBytes(snPrivacyProofBytes)
			if err != nil{
				continue
			}else{
				break
			}
		}

		paymentProof.SetSerialNumberProof(snPrivacyProof)

		txFee.Proof = paymentProof

		sigPrivate := append(keySet.PrivateKey, []byte{0,0,0,0}...)

		err = txFee.Sign(sigPrivate)
		if err != nil {
			return nil, err
		}



	} else {
		txNormal, ok  := tx.TxTokenData.TxNormal.(*transaction.TxVersion1)
		if !ok {
			return nil, errors.New("cannot parse txfee")
		}

		paymentProof, ok := txNormal.Proof.(*zkp.PaymentProof)
		if !ok {
			return nil, errors.New("cannot parse payment proof of txfee")
		}

		//Tamper with the serial number privacy proof
		snPrivacyProof := paymentProof.GetSerialNumberProof()

		snPrivacyProofBytes := snPrivacyProof[0].Bytes()
		for {
			r := RandIntInterval(0, len(snPrivacyProofBytes)-1)
			tmpByte := snPrivacyProofBytes[r]
			tmpRandByte := uint8(common.RandIntInterval(0, 256))
			if tmpByte == tmpRandByte{
				continue
			}

			snPrivacyProofBytes[r] = tmpRandByte
			err = snPrivacyProof[0].SetBytes(snPrivacyProofBytes)
			if err != nil{
				continue
			}else{
				break
			}
		}

		txNormal.Proof = paymentProof

		sigPrivate := append(keySet.PrivateKey, []byte{0,0,0,0}...)

		err = txNormal.Sign(sigPrivate)
		if err != nil {
			return nil, err
		}
	}

	txBytes, err := json.Marshal(tx)
	if err != nil {
		return nil, err
	}

	result := jsonresult.NewCreateTransactionResult(tx.Hash(), common.EmptyString, txBytes, 0)

	base58Result := result.Base58CheckData

	return []byte(base58Result), nil
}

func (tool *DebugTool) CreateTxTokenPrivacyWithOneOfManyProofTampered(privateKey, tokenIDString, paymentString string)([]byte, error){
	b, err := tool.CreateRawTxToken(privateKey, tokenIDString, paymentString, uint64(100), true)

	tx := new(transaction.TxTokenVersion1)
	err = json.Unmarshal(b, &tx)
	if err != nil {
		return nil, err
	}

	keyWallet, err := wallet.Base58CheckDeserialize(privateKey)
	if err != nil{
		return nil, err
	}

	keySet := &keyWallet.KeySet

	r := common.RandInt()
	if r%2 == 0 {
		txFee, ok  := tx.Tx.(*transaction.TxVersion1)
		if !ok {
			return nil, errors.New("cannot parse txfee")
		}

		paymentProof, ok := txFee.Proof.(*zkp.PaymentProof)
		if !ok {
			return nil, errors.New("cannot parse payment proof of txfee")
		}

		//Tamper with the one-of-many proof
		oneOfManyProof := paymentProof.GetOneOfManyProof()
		oneOfManyProof[0].Statement.Commitments[0] = operation.RandomPoint()
		paymentProof.SetOneOfManyProof(oneOfManyProof)

		txFee.Proof = paymentProof

		sigPrivate := append(keySet.PrivateKey, []byte{0,0,0,0}...)

		err = txFee.Sign(sigPrivate)
		if err != nil {
			return nil, err
		}



	} else {
		txNormal, ok  := tx.TxTokenData.TxNormal.(*transaction.TxVersion1)
		if !ok {
			return nil, errors.New("cannot parse txfee")
		}

		paymentProof, ok := txNormal.Proof.(*zkp.PaymentProof)
		if !ok {
			return nil, errors.New("cannot parse payment proof of txfee")
		}

		//Tamper with the one-of-many proof
		oneOfManyProof := paymentProof.GetOneOfManyProof()
		oneOfManyProof[0].Statement.Commitments[0] = operation.RandomPoint()
		paymentProof.SetOneOfManyProof(oneOfManyProof)
		txNormal.Proof = paymentProof

		sigPrivate := append(keySet.PrivateKey, []byte{0,0,0,0}...)

		err = txNormal.Sign(sigPrivate)
		if err != nil {
			return nil, err
		}
	}

	txBytes, err := json.Marshal(tx)
	if err != nil {
		return nil, err
	}

	result := jsonresult.NewCreateTransactionResult(tx.Hash(), common.EmptyString, txBytes, 0)

	base58Result := result.Base58CheckData

	return []byte(base58Result), nil
}


//Debug tool functions
func (tool *DebugTool) SendTxNoPrivacyFake(privateKey, tokenIDString, paymentString string, version int8) ([]byte, error) {
	var base58Bytes []byte
	txType := 0

	switch txType {
	case 0: //Create txprivacy without signatures
		return tool.CreateTxNoPrivacyWithoutSignature(privateKey, tokenIDString, paymentString, version)
		//if err != nil {
		//	return nil, err
		//}
		//
		//query := fmt.Sprintf(`{
		//	"jsonrpc": "1.0",
		//	"method": "sendtransaction",
		//	"params": [
		//		"%s"
		//	],
		//	"id": 1
		//}`, string(base58Bytes))

		//return tool.SendPostRequestWithQuery(query)
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

func (tool *DebugTool) SendTxPrivacyFake(privateKey, tokenIDString, paymentString string, txType int64, version int8) ([]byte, error) {
	if tokenIDString != common.PRVIDStr {
		switch txType{
		case 0:
			return tool.CreateTxTokenPrivacyWithoutSignature(privateKey, tokenIDString, paymentString)
			//if err != nil {
			//	return nil, err
			//}
			//query := fmt.Sprintf(`{
			//	"jsonrpc": "1.0",
			//	"method": "sendrawprivacycustomtokentransaction",
			//	"params": [
			//		"%s"
			//	],
			//	"id": 1
			//}`, string(base58Bytes))
			//
			//return tool.SendPostRequestWithQuery(query)
		case 1:
			return tool.CreateTxTokenPrivacyWithBulletProofCommitmentsTampered(privateKey, tokenIDString, paymentString)
			//if err != nil{
			//	return nil, err
			//}
			//
			//query := fmt.Sprintf(`{
			//	"jsonrpc": "1.0",
			//	"method": "sendrawprivacycustomtokentransaction",
			//	"params": [
			//		"%s"
			//	],
			//	"id": 1
			//}`, string(base58Bytes))
			//
			//return tool.SendPostRequestWithQuery(query)
		case 2:
			return tool.CreateTxTokenPrivacyWithSNProofCommitmentTampered(privateKey, tokenIDString, paymentString)
			//if err != nil{
			//	return nil, err
			//}
			//
			//query := fmt.Sprintf(`{
			//	"jsonrpc": "1.0",
			//	"method": "sendrawprivacycustomtokentransaction",
			//	"params": [
			//		"%s"
			//	],
			//	"id": 1
			//}`, string(base58Bytes))
			//
			//return tool.SendPostRequestWithQuery(query)
		case 3:
			return tool.CreateTxTokenPrivacyWithOneOfManyProofTampered(privateKey, tokenIDString, paymentString)
			//if err != nil{
			//	return nil, err
			//}
			//
			//query := fmt.Sprintf(`{
			//	"jsonrpc": "1.0",
			//	"method": "sendrawprivacycustomtokentransaction",
			//	"params": [
			//		"%s"
			//	],
			//	"id": 1
			//}`, string(base58Bytes))
			//
			//return tool.SendPostRequestWithQuery(query)
		default:
			return nil, errors.New("Wrong txType")
		}

	}else{
		switch txType {
		case 0: //Create txprivacy without signatures
			return tool.CreateTxPrivacyWithoutSignature(privateKey, tokenIDString, paymentString, version)
		//	if err != nil {
		//		return nil, err
		//	}
		//
		//	query := fmt.Sprintf(`{
		//	"jsonrpc": "1.0",
		//	"method": "sendtransaction",
		//	"params": [
		//		"%s"
		//	],
		//	"id": 1
		//}`, string(base58Bytes))
		//
		//	return tool.SendPostRequestWithQuery(query)
		case 1: //Create txprivacy with the bulletProof tampered
			return tool.CreateTxPrivacyWithBulletProofCommitmentsTampered(privateKey, tokenIDString, paymentString, version)
		//	if err != nil {
		//		return nil, err
		//	}
		//
		//	query := fmt.Sprintf(`{
		//	"jsonrpc": "1.0",
		//	"method": "sendtransaction",
		//	"params": [
		//		"%s"
		//	],
		//	"id": 1
		//}`, string(base58Bytes))
		//
		//	return tool.SendPostRequestWithQuery(query)
		case 2: //Create txprivacy with snPrivacyProof tampered
			return tool.CreateTxPrivacyWithSNProofCommitmentTampered(privateKey, tokenIDString, paymentString)
		//	if err != nil {
		//		return nil, err
		//	}
		//
		//	query := fmt.Sprintf(`{
		//	"jsonrpc": "1.0",
		//	"method": "sendtransaction",
		//	"params": [
		//		"%s"
		//	],
		//	"id": 1
		//}`, string(base58Bytes))
		//
		//	return tool.SendPostRequestWithQuery(query)
		case 3: //Create txprivacy with one-of-many proofs tampered
			return tool.CreateTxPrivacyWithOneOfManyProofTampered(privateKey, tokenIDString, paymentString)
		//	if err != nil {
		//		return nil, err
		//	}
		//
		//	query := fmt.Sprintf(`{
		//	"jsonrpc": "1.0",
		//	"method": "sendtransaction",
		//	"params": [
		//		"%s"
		//	],
		//	"id": 1
		//}`, string(base58Bytes))
		//
		//	return tool.SendPostRequestWithQuery(query)
		default:
			return nil, errors.New("Wrong txType")
		}
	}
}

func (tool *DebugTool) CreateRawTxToken(privateKey, tokenIDString, paymentString string, amount uint64, isPrivacy bool) ([]byte, error) {
	// fmt.Println("Hi i'm here")
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
				"Privacy": true,
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
	}`, privateKey, tokenIDString, paymentString, amount)
	// fmt.Println("trying to send")
	// fmt.Println(query)

	respondInBytes, err := tool.SendPostRequestWithQuery(query)
	if err != nil {
		return nil, err
	}
	// fmt.Println(string(respondInBytes))


	respond, err := ParseResponse(respondInBytes)
	if err != nil {
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
	if err != nil {
		return nil, err
	}

	return bytearrays, nil
}

func (tool *DebugTool) CreateRawTx(privateKey, paymentString string, amount uint64, isPrivacy bool) ([]byte, error) {
	privIndicator := "-1"
	if isPrivacy{
		privIndicator = "1"
	}
	query := fmt.Sprintf(`{
		"jsonrpc": "1.0",
		"method": "createtransaction",
		"params": [
			"%s",
			{
				"%s":%d
			},
			1,
			%s
		],
		"id": 1
	}`, privateKey, paymentString, amount, privIndicator)

	respondInBytes, err := tool.SendPostRequestWithQuery(query)
	if err != nil {
		return nil, err
	}

	respond, err := ParseResponse(respondInBytes)
	if err != nil {
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
	if err != nil {
		return nil, err
	}

	return bytearrays, nil
}
