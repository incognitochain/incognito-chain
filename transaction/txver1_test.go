package transaction

import (
	"bytes"
	"fmt"
	// "encoding/json"
	"testing"
	"time"
	"os"
	"math/rand"
	// "math/big"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/privacy/coin"
	"github.com/incognitochain/incognito-chain/privacy/key"
	"github.com/incognitochain/incognito-chain/privacy/operation"
	"github.com/incognitochain/incognito-chain/metadata"
	zkp "github.com/incognitochain/incognito-chain/privacy/privacy_v1/zeroknowledge"
	"github.com/incognitochain/incognito-chain/privacy/privacy_v1/zeroknowledge/aggregatedrange"
	"github.com/incognitochain/incognito-chain/privacy/privacy_v1/zeroknowledge/serialnumberprivacy"
	"github.com/stretchr/testify/assert"
)

func createAndSaveCoinV1s(numCoins, numEquals int, privKey privacy.PrivateKey, pubKey *operation.Point, dummyDB *statedb.StateDB) ([]coin.PlainCoin, error) {
	//amount := uint64(common.RandIntInterval(0, 1000000))
	amount := uint64(numCoins * 1000)
	outCoins := []coin.PlainCoin{}
	for i := 0; i < numEquals; i++ {
		coin, err := createSamplePlainCoinV1(privKey, pubKey, 1000, nil)
		if err != nil {
			return nil, err
		}
		outCoins = append(outCoins, coin)
	}
	tmpOutCoins, err := createSamplePlainCoinsFromTotalAmount(privKey, pubKey, amount, numCoins-numEquals, 1)
	for _, coin := range tmpOutCoins {
		outCoins = append(outCoins, coin)
	}
	if err != nil {
		return nil, err
	}

	//save coins and commitment indices onto the database
	commitmentsToBeSaved := [][]byte{}
	coinsToBeSaved := [][]byte{}
	for _, outCoin := range outCoins {
		coinsToBeSaved = append(coinsToBeSaved, outCoin.Bytes())
		commitmentsToBeSaved = append(commitmentsToBeSaved, outCoin.GetCommitment().ToBytesS())
	}
	err = statedb.StoreOutputCoins(dummyDB, common.PRVCoinID, pubKey.ToBytesS(), coinsToBeSaved, shardID)
	if err != nil {
		return nil, err
	}
	err = statedb.StoreCommitments(dummyDB, common.PRVCoinID, commitmentsToBeSaved, shardID)
	if err != nil {
		return nil, err
	}

	return outCoins, nil
}

func createAndSaveTokenCoinV1s(numCoins, numEquals int, privKey privacy.PrivateKey, pubKey *operation.Point, tokenID common.Hash, dummyDB *statedb.StateDB) ([]coin.PlainCoin, error) {
	//amount := uint64(common.RandIntInterval(0, 1000000))
	amount := uint64(numCoins * 1000)
	outCoins := []coin.PlainCoin{}
	for i := 0; i < numEquals; i++ {
		coin, err := createSamplePlainCoinV1(privKey, pubKey, 1000, nil)
		if err != nil {
			return nil, err
		}
		outCoins = append(outCoins, coin)
	}
	tmpOutCoins, err := createSamplePlainCoinsFromTotalAmount(privKey, pubKey, amount, numCoins-numEquals, 1)
	for _, coin := range tmpOutCoins {
		outCoins = append(outCoins, coin)
	}
	if err != nil {
		return nil, err
	}

	//save coins and commitment indices onto the database
	commitmentsToBeSaved := [][]byte{}
	coinsToBeSaved := [][]byte{}
	for _, outCoin := range outCoins {
		coinsToBeSaved = append(coinsToBeSaved, outCoin.Bytes())
		commitmentsToBeSaved = append(commitmentsToBeSaved, outCoin.GetCommitment().ToBytesS())
	}
	err = statedb.StoreOutputCoins(dummyDB, tokenID, pubKey.ToBytesS(), coinsToBeSaved, shardID)
	if err != nil {
		return nil, err
	}
	err = statedb.StoreCommitments(dummyDB, tokenID, commitmentsToBeSaved, shardID)
	if err != nil {
		return nil, err
	}

	return outCoins, nil
}

func createTxPrivacyInitParams(keySet *incognitokey.KeySet, inputCoins []coin.PlainCoin, hasPrivacy bool, numOutputs int) (*incognitokey.KeySet, *TxPrivacyInitParams, error) {
	//initialize payment info of input coins
	paymentInfos := make([]*key.PaymentInfo, numOutputs)
	sumAmount := uint64(0)
	for _, inputCoin := range inputCoins {
		sumAmount += inputCoin.GetValue()
	}
	amount := sumAmount/uint64(numOutputs)
	for i:= 0; i< numOutputs - 1; i++ {
		paymentInfos[i] = key.InitPaymentInfo(keySet.PaymentAddress, amount, []byte("blahblah"))
		sumAmount -= amount
	}
	paymentInfos[numOutputs-1] = key.InitPaymentInfo(keySet.PaymentAddress, sumAmount, []byte("blahblah"))

	//create privacyinitparam
	txPrivacyInitParam := NewTxPrivacyInitParams(&keySet.PrivateKey,
		paymentInfos,
		inputCoins,
		0,
		hasPrivacy,
		dummyDB,
		&common.PRVCoinID,
		nil,
		nil)

	return keySet, txPrivacyInitParam, nil
}

func TestTxVersion1_ValidateTransaction(t *testing.T) {
	keySets, err := prepareKeySets(1)
	assert.Equal(t, nil, err, "prepareKeySets returns an error: %v", err)
	keySet := keySets[0]

	pubKey, err := new(operation.Point).FromBytesS(keySet.PaymentAddress.Pk)
	assert.Equal(t, nil, err, "Cannot parse public key")

	for i:=0; i < numTests; i++ {
		numOfInputs := RandInt() % (maxInputs - minInputs + 1) + minInputs
		numOfOutputs := RandInt() % (maxInputs - minInputs + 1) + minInputs
		coins, err := createAndSaveCoinV1s(100, 0, keySet.PrivateKey, pubKey, dummyDB)
		assert.Equal(t, nil, err, "createAndSaevCoinV1s returns an error: %v", err)

		tx := new(TxVersion1)

		r := common.RandInt() % (100 - numOfInputs)

		inputCoins := coins[r:r+numOfInputs]

		_, txPrivacyParams, err := createTxPrivacyInitParams(keySet, inputCoins, true, numOfOutputs)
		assert.Equal(t, nil, err, "createTxPrivacyInitParams returns an error: %v", err)

		err = tx.Init(txPrivacyParams)
		assert.Equal(t, nil, err, "validateTxParams returns an error: %v", err)

		res, err := tx.ValidateSanityData(nil, nil, nil, 0)
		assert.Equal(t, nil, err, "ValidateSanityData returns an error: %v", err)
		assert.Equal(t, true, res)

		res, err = tx.ValidateTransaction(true, dummyDB, dummyDB, 0, &common.PRVCoinID, false, false)
		assert.Equal(t, nil, err, "ValidateTransaction returns an error: %v", err)
		assert.Equal(t, true, res)

		testTxV1JsonMarshaler(tx, 25, dummyDB, t)
	}
}

func TestTxVersion1_InputCoinReplication(t *testing.T) {
	keySets, err := prepareKeySets(1)
	assert.Equal(t, nil, err, "prepareKeySets returns an error: %v", err)
	keySet := keySets[0]

	pubKey, err := new(operation.Point).FromBytesS(keySet.PaymentAddress.Pk)
	assert.Equal(t, nil, err, "Cannot parse public key")

	for i:=0; i< numTests; i++ {
		coins, err := createAndSaveCoinV1s(100, 2, keySet.PrivateKey, pubKey, dummyDB)
		assert.Equal(t, nil, err, "createAndSaevCoinV1s returns an error: %v", err)

		// fmt.Println(coins[0].GetValue())

		tx := new(TxVersion1)
		//choose some input coins to spend => make sure that inputCoins[0] and inputCoins[1] have the same amount
		inputCoins := coins[:10]

		_, txPrivacyParams, err := createTxPrivacyInitParams(keySet, inputCoins, true, 1)
		assert.Equal(t, nil, err, "createTxPrivacyInitParams returns an error: %v", err)

		err = validateTxParams(txPrivacyParams)
		assert.Equal(t, nil, err, "validateTxParams returns an error: %v", err)

		err = tx.initializeTxAndParams(txPrivacyParams)
		assert.Equal(t, nil, err, "Init returns an error: %v", err)

		paymentWitnessParamPtr, err := tx.initializePaymentWitnessParam(txPrivacyParams)
		assert.Equal(t, nil, err, "initializePaymentWitnessParam returns an error: %v", err)

		//copy the first coin to the second
		sz := privacy.CommitmentRingSize
		paymentWitnessParamPtr.InputCoins[1] = paymentWitnessParamPtr.InputCoins[0]
		paymentWitnessParamPtr.MyCommitmentIndices[1] = paymentWitnessParamPtr.MyCommitmentIndices[0]
		for i := 0; i < sz; i++ {
			paymentWitnessParamPtr.CommitmentIndices[i+sz] = paymentWitnessParamPtr.CommitmentIndices[i]
			paymentWitnessParamPtr.Commitments[i+sz] = paymentWitnessParamPtr.Commitments[i]
		}

		err = tx.proveAndSignCore(txPrivacyParams, paymentWitnessParamPtr)
		assert.Equal(t, nil, err, "proveAndSignCore returns an error: %v")

		res, err := tx.ValidateSanityData(nil,nil,nil,0)
		assert.Equal(t, false, res)
		if res {
			assert.Equal(t, true, bytes.Equal(paymentWitnessParamPtr.InputCoins[1].Bytes(), paymentWitnessParamPtr.InputCoins[0].Bytes()))
		}
	}
}

func TestTxVersion1_BulletProofCommitmentConsistency(t *testing.T) {
	keySets, err := prepareKeySets(1)
	assert.Equal(t, nil, err, "prepareKeySets returns an error: %v", err)
	keySet := keySets[0]

	pubKey, err := new(operation.Point).FromBytesS(keySet.PaymentAddress.Pk)
	assert.Equal(t, nil, err, "Cannot parse public key")

	coins, err := createAndSaveCoinV1s(numTests*50, 0, keySet.PrivateKey, pubKey, dummyDB)
	assert.Equal(t, nil, err, "createAndSaevCoinV1s returns an error: %v", err)

	for i:=0;i<numTests;i++ {
		//create 2 transactions
		tx1 := new(TxVersion1)
		tx2 := new(TxVersion1)
		inputCoins := coins[i*20 : i*20+10]

		_, txPrivacyParams, err := createTxPrivacyInitParams(keySet, inputCoins, true, 1)
		assert.Equal(t, nil, err, "createTxPrivacyInitParams returns an error: %v", err)

		err = tx1.Init(txPrivacyParams)
		assert.Equal(t, nil, err, "Init returns an error: %v", err)

		res, err := tx1.ValidateSanityData(nil, nil, nil, 0)
		assert.Equal(t, nil, err, "ValidateSanityData returns an error: %v", err)
		assert.Equal(t, true, res)

		res, err = tx1.ValidateTransaction(true, dummyDB, dummyDB, 0, &common.PRVCoinID, false, false)
		assert.Equal(t, nil, err, "ValidateTransaction returns an error: %v", err)
		assert.Equal(t, true, res)

		proof, ok := tx1.Proof.(*zkp.PaymentProof)
		assert.Equal(t, true, ok, "cannot parse proof")

		//retrieve the bulletproof
		oldAggregatedProof, ok := proof.GetAggregatedRangeProof().(*aggregatedrange.AggregatedRangeProof)
		assert.Equal(t, true, ok, "cannot parse bulletproof")

		//create a new transaction and reuse the bulletproof
		inputCoins2 := coins[i*20+10 : i*20+20]

		_, newTxPrivacyParams, err := createTxPrivacyInitParams(keySet, inputCoins2, true, 1)
		assert.Equal(t, nil, err, "createTxPrivacyInitParams returns an error: %v", err)

		err = validateTxParams(newTxPrivacyParams)
		assert.Equal(t, nil, err, "validateTxParams returns an error: %v", err)

		err = tx2.initializeTxAndParams(newTxPrivacyParams)
		assert.Equal(t, nil, err, "Init returns an error: %v", err)

		newPaymentWitnessParamPtr, err := tx2.initializePaymentWitnessParam(newTxPrivacyParams)
		assert.Equal(t, nil, err, "initializePaymentWitnessParam returns an error: %v", err)

		err = tx2.proveAndSignCore(newTxPrivacyParams, newPaymentWitnessParamPtr)
		assert.Equal(t, nil, err, "proveAndSignCore returns an error: %v", err)

		proof, ok = tx2.Proof.(*zkp.PaymentProof)
		assert.Equal(t, true, ok, "cannot parse proof")

		//assert that tx2 and tx1 have different bulletproofs
		assert.Equal(t, false, bytes.Equal(oldAggregatedProof.Bytes(), proof.Bytes()), "tx2 and tx1 have the same bulletproof")

		proof.SetAggregatedRangeProof(oldAggregatedProof)

		tx2.Sig = nil
		err = tx2.sign()
		assert.Equal(t, nil, err, "sign returns an error: %v", err)

		res, err = tx2.ValidateSanityData(nil, nil, nil, 0)
		//assert.Equal(t, nil, err, "ValidateTransaction returns an error: %v", err)
		assert.Equal(t, false, res)

	}
}

func TestTxVersion1_SerialNumberProofConsistency(t *testing.T) {
	keySets, err := prepareKeySets(1)
	assert.Equal(t, nil, err, "prepareKeySets returns an error: %v", err)
	keySet := keySets[0]

	pubKey, err := new(operation.Point).FromBytesS(keySet.PaymentAddress.Pk)
	assert.Equal(t, nil, err, "Cannot parse public key")

	coins, err := createAndSaveCoinV1s(numTests*50, 0, keySet.PrivateKey, pubKey, dummyDB)
	assert.Equal(t, nil, err, "createAndSaevCoinV1s returns an error: %v", err)

	for i := 0; i < numTests; i++ {
		//create 2 transactions
		tx1 := new(TxVersion1)
		tx2 := new(TxVersion1)
		inputCoins := coins[i*20 : i*20+10]

		_, txPrivacyParams, err := createTxPrivacyInitParams(keySet, inputCoins, true, 1)
		assert.Equal(t, nil, err, "createTxPrivacyInitParams returns an error: %v", err)

		err = tx1.Init(txPrivacyParams)
		assert.Equal(t, nil, err, "Init returns an error at test number %v: %v", i, err)

		res, err := tx1.ValidateSanityData(nil, nil, nil, 0)
		assert.Equal(t, nil, err, "ValidateSanityData returns an error: %v", err)
		assert.Equal(t, true, res)

		res, err = tx1.ValidateTransaction(true, dummyDB, dummyDB, 0, &common.PRVCoinID, false, false)
		assert.Equal(t, nil, err, "ValidateTransaction returns an error: %v", err)
		assert.Equal(t, true, res)

		proof, ok := tx1.Proof.(*zkp.PaymentProof)
		assert.Equal(t, true, ok, "cannot parse proof")

		//retrieve the serial number proof
		oldSNProof := proof.GetSerialNumberProof()

		//create a new transaction and reuse the serial number proof
		inputCoins2 := coins[i*20+10 : i*20+20]

		_, newTxPrivacyParams, err := createTxPrivacyInitParams(keySet, inputCoins2, true, 1)
		assert.Equal(t, nil, err, "createTxPrivacyInitParams returns an error: %v", err)

		err = validateTxParams(newTxPrivacyParams)
		assert.Equal(t, nil, err, "validateTxParams returns an error: %v", err)

		err = tx2.initializeTxAndParams(newTxPrivacyParams)
		assert.Equal(t, nil, err, "Init returns an error: %v", err)

		newPaymentWitnessParamPtr, err := tx2.initializePaymentWitnessParam(newTxPrivacyParams)
		assert.Equal(t, nil, err, "initializePaymentWitnessParam returns an error: %v", err)

		err = tx2.proveAndSignCore(newTxPrivacyParams, newPaymentWitnessParamPtr)
		assert.Equal(t, nil, err, "proveAndSignCore returns an error: %v", err)

		proof, ok = tx2.Proof.(*zkp.PaymentProof)
		assert.Equal(t, true, ok, "cannot parse proof")

		//check that the SNProof has changed
		assert.Equal(t, false, bytes.Equal(oldSNProof[0].Bytes(), proof.GetSerialNumberProof()[0].Bytes()))
		assert.Equal(t, false, bytes.Equal(oldSNProof[1].Bytes(), proof.GetSerialNumberProof()[1].Bytes()))
		proof.SetSerialNumberProof(oldSNProof)

		//re-sign the transaction
		tx2.Sig = nil
		err = tx2.sign()
		assert.Equal(t, nil, err, "sign returns an error: %v", err)

		res, err = tx2.ValidateSanityData(nil, nil, nil, 0)
		assert.Equal(t, false, res)
	}
}

func TestTxVersion1_OneOutOfManyProofConsistency(t *testing.T) {
	keySets, err := prepareKeySets(1)
	assert.Equal(t, nil, err, "prepareKeySets returns an error: %v", err)
	keySet := keySets[0]

	pubKey, err := new(operation.Point).FromBytesS(keySet.PaymentAddress.Pk)
	assert.Equal(t, nil, err, "Cannot parse public key")

	coins, err := createAndSaveCoinV1s(numTests*50, 0, keySet.PrivateKey, pubKey, dummyDB)
	assert.Equal(t, nil, err, "createAndSaevCoinV1s returns an error: %v", err)

	for i := 0; i < numTests; i++ {
		//create 2 transactions
		tx1 := new(TxVersion1)
		tx2 := new(TxVersion1)
		inputCoins := coins[i*20 : i*20+10]

		_, txPrivacyParams, err := createTxPrivacyInitParams(keySet, inputCoins, true, 1)
		assert.Equal(t, nil, err, "createTxPrivacyInitParams returns an error: %v", err)

		err = tx1.Init(txPrivacyParams)
		assert.Equal(t, nil, err, "Init returns an error at test number %v: %v", i, err)

		res, err := tx1.ValidateSanityData(nil, nil, nil, 0)
		assert.Equal(t, nil, err, "ValidateSanityData returns an error: %v", err)
		assert.Equal(t, true, res)

		res, err = tx1.ValidateTransaction(true, dummyDB, dummyDB, 0, &common.PRVCoinID, false, false)
		assert.Equal(t, nil, err, "ValidateTransaction returns an error: %v", err)
		assert.Equal(t, true, res)

		proof, ok := tx1.Proof.(*zkp.PaymentProof)
		assert.Equal(t, true, ok, "cannot parse proof")

		//retrieve the serial number proof
		oldOneOfMany := proof.GetOneOfManyProof()

		//create a new transaction and reuse the serial number proof
		inputCoins2 := coins[i*20+10 : i*20+20]

		_, newTxPrivacyParams, err := createTxPrivacyInitParams(keySet, inputCoins2, true, 1)
		assert.Equal(t, nil, err, "createTxPrivacyInitParams returns an error: %v", err)

		err = validateTxParams(newTxPrivacyParams)
		assert.Equal(t, nil, err, "validateTxParams returns an error: %v", err)

		err = tx2.initializeTxAndParams(newTxPrivacyParams)
		assert.Equal(t, nil, err, "Init returns an error: %v", err)

		newPaymentWitnessParamPtr, err := tx2.initializePaymentWitnessParam(newTxPrivacyParams)
		assert.Equal(t, nil, err, "initializePaymentWitnessParam returns an error: %v", err)

		err = tx2.proveAndSignCore(newTxPrivacyParams, newPaymentWitnessParamPtr)
		assert.Equal(t, nil, err, "proveAndSignCore returns an error: %v", err)

		proof, ok = tx2.Proof.(*zkp.PaymentProof)
		assert.Equal(t, true, ok, "cannot parse proof")

		//check that the SNProof has changed
		assert.Equal(t, false, bytes.Equal(oldOneOfMany[0].Bytes(), proof.GetSerialNumberProof()[0].Bytes()))
		assert.Equal(t, false, bytes.Equal(oldOneOfMany[1].Bytes(), proof.GetSerialNumberProof()[1].Bytes()))
		proof.SetOneOfManyProof(oldOneOfMany)

		//re-sign the transaction
		tx2.Sig = nil
		err = tx2.sign()
		assert.Equal(t, nil, err, "sign returns an error: %v", err)

		res, err = tx2.ValidateSanityData(nil, nil, nil, 0)
		assert.Equal(t, nil, err, "ValidateSanityData returns an error: %v", err)
		assert.Equal(t, true, res)

		res, err = tx2.ValidateTransaction(true, dummyDB, dummyDB, 0, &common.PRVCoinID, false, false)
		//assert.Equal(t, nil, err, "ValidateTransaction returns an error: %v", err)
		assert.Equal(t, false, res)
	}
}

func TestTxVersion1_SerialNumberNoPrivacyProofConsistency(t *testing.T) {
	keySets, err := prepareKeySets(1)
	assert.Equal(t, nil, err, "prepareKeySets returns an error: %v", err)
	keySet := keySets[0]

	pubKey, err := new(operation.Point).FromBytesS(keySet.PaymentAddress.Pk)
	assert.Equal(t, nil, err, "Cannot parse public key")

	coins, err := createAndSaveCoinV1s(numTests*50, 0, keySet.PrivateKey, pubKey, dummyDB)
	assert.Equal(t, nil, err, "createAndSaevCoinV1s returns an error: %v", err)

	for i := 0; i < numTests; i++ {
		//create 2 transactions
		tx1 := new(TxVersion1)
		tx2 := new(TxVersion1)
		inputCoins := coins[i*20 : i*20+10]

		_, txPrivacyParams, err := createTxPrivacyInitParams(keySet, inputCoins, false, 1)
		assert.Equal(t, nil, err, "createTxPrivacyInitParams returns an error: %v", err)
		txPrivacyParams.hasPrivacy = false // set hasPrivacy to false to obtain serialNumberNoPrivacyProof

		err = tx1.Init(txPrivacyParams)
		assert.Equal(t, nil, err, "Init returns an error at test number %v: %v", i, err)

		res, err := tx1.ValidateSanityData(nil, nil, nil, 0)
		assert.Equal(t, nil, err, "ValidateSanityData returns an error: %v", err)
		assert.Equal(t, true, res)

		res, err = tx1.ValidateTransaction(false, dummyDB, dummyDB, 0, &common.PRVCoinID, false, false)
		assert.Equal(t, nil, err, "ValidateTransaction returns an error: %v", err)
		assert.Equal(t, true, res)

		proof, ok := tx1.Proof.(*zkp.PaymentProof)
		assert.Equal(t, true, ok, "cannot parse proof")

		//retrieve the serial number proof
		oldSNNoPrivacyProof := proof.GetSerialNumberNoPrivacyProof()

		//create a new transaction and reuse the serial number proof
		inputCoins2 := coins[i*20+10 : i*20+20]

		_, newTxPrivacyParams, err := createTxPrivacyInitParams(keySet, inputCoins2, false, 1)
		assert.Equal(t, nil, err, "createTxPrivacyInitParams returns an error: %v", err)
		newTxPrivacyParams.hasPrivacy = false // set hasPrivacy to false to obtain serialNumberNoPrivacyProof

		err = validateTxParams(newTxPrivacyParams)
		assert.Equal(t, nil, err, "validateTxParams returns an error: %v", err)

		err = tx2.initializeTxAndParams(newTxPrivacyParams)
		assert.Equal(t, nil, err, "Init returns an error: %v", err)

		newPaymentWitnessParamPtr, err := tx2.initializePaymentWitnessParam(newTxPrivacyParams)
		assert.Equal(t, nil, err, "initializePaymentWitnessParam returns an error: %v", err)

		err = tx2.proveAndSignCore(newTxPrivacyParams, newPaymentWitnessParamPtr)
		assert.Equal(t, nil, err, "proveAndSignCore returns an error: %v", err)

		proof, ok = tx2.Proof.(*zkp.PaymentProof)
		assert.Equal(t, true, ok, "cannot parse proof")

		//check that the SNProof has changed
		if len(oldSNNoPrivacyProof) > 0{
			assert.Equal(t, false, bytes.Equal(oldSNNoPrivacyProof[0].Bytes(), proof.GetSerialNumberNoPrivacyProof()[0].Bytes()))
			proof.SetSerialNumberNoPrivacyProof(oldSNNoPrivacyProof)
		}

		//re-sign the transaction
		tx2.Sig = nil
		err = tx2.sign()
		assert.Equal(t, nil, err, "sign returns an error: %v", err)

		res, err = tx2.ValidateSanityData(nil, nil, nil, 0)
		assert.Equal(t, nil, err, "ValidateSanityData returns an error: %v", err)
		assert.Equal(t, true, res) //This should fail

		res, err = tx2.ValidateTransaction(false, dummyDB, dummyDB, 0, &common.PRVCoinID, false, false)
		//assert.Equal(t, nil, err, "ValidateTransaction returns an error: %v", err)
		assert.Equal(t, true, res)
	}
}

func TestTxVersion1_OutputTampered(t *testing.T) {
	//This test will attempt to create a transaction ver1 which has output value larger than sum of input values
	keySets, err := prepareKeySets(1)
	assert.Equal(t, nil, err, "prepareKeySets returns an error: %v", err)
	keySet := keySets[0]

	pubKey, err := new(operation.Point).FromBytesS(keySet.PaymentAddress.Pk)
	assert.Equal(t, nil, err, "Cannot parse public key")

	coins, err := createAndSaveCoinV1s(numTests*50, 0, keySet.PrivateKey, pubKey, dummyDB)
	assert.Equal(t, nil, err, "createAndSaevCoinV1s returns an error: %v", err)

	for i := 0; i < numTests; i++ {
		//create 2 transactions
		tx := new(TxVersion1)
		inputCoins := coins[i*10 : i*10 + 10]

		_, txPrivacyParams, err := createTxPrivacyInitParams(keySet, inputCoins, true, 1)
		assert.Equal(t, nil, err, "createTxPrivacyInitParams returns an error: %v", err)

		//create an output whose amount is larger than sum of all input amounts
		txPrivacyParams.paymentInfo[0].Amount += 1000

		//initialize transaction and param (initializeTxAndParams without updateParamsWhenOverBalance)
		tx.sigPrivKey = keySet.PrivateKey
		if tx.LockTime == 0 {
			tx.LockTime = time.Now().Unix()
		}
		tx.Fee = txPrivacyParams.fee
		tx.Type = common.TxNormalType
		tx.Metadata = txPrivacyParams.metaData
		tx.PubKeyLastByteSender = keySet.PaymentAddress.Pk[len(keySet.PaymentAddress.Pk)-1]
		tx.Version = 1
		tx.Info, err = getTxInfo(txPrivacyParams.info)
		assert.Equal(t, nil, err, "getTxInfo returns an error: %v", err)

		err = tx.prove(txPrivacyParams)
		assert.Equal(t, nil, err, "tx.Prove returns an error: %v", err)

		proof, ok := tx.Proof.(*zkp.PaymentProof)
		assert.Equal(t, true, ok, "cannot parse payment proof")


		cmOutputValueSum := proof.GetCommitmentOutputValue()
		cmInputValues := proof.GetCommitmentInputValue()
		cmInputSND := proof.GetCommitmentInputSND()
		serialNumberProof := proof.GetSerialNumberProof()

		tmpSNProof := serialnumberprivacy.Copy(*serialNumberProof[0])

		cmInputValueSum := new(operation.Point).Identity()
		for _, cmInputValue := range cmInputValues {
			cmInputValueSum.Add(cmInputValueSum, cmInputValue)
		}

		tmpDiff := new(operation.Point).Sub(cmOutputValueSum[0], cmInputValueSum)
		cmInputValues[0].Add(cmInputValues[0], tmpDiff)
		cmInputSND[0].Sub(cmInputSND[0], tmpDiff)

		proof.SetCommitmentInputValue(cmInputValues)
		proof.SetCommitmentInputSND(cmInputSND)
		serialNumberProof[0] = tmpSNProof
		proof.SetSerialNumberProof(serialNumberProof)


		tx.SetProof(proof)

		tx.Sig = nil
		err = tx.sign()

		assert.Equal(t, nil, err, "tx.sign returns an error: %v", err)

		res, err := tx.ValidateSanityData(nil, nil, nil, 0)
		assert.Equal(t, nil, err, "ValidateSanityData returns an error: %v", err)
		assert.Equal(t, true, res)

		res, err = tx.ValidateTransaction(true, dummyDB, dummyDB, 0, &common.PRVCoinID, false, false)
		assert.Equal(t, true, res, err)

	}
}

func BenchmarkTxV1BatchVerify(b *testing.B) {
	rand.Seed(time.Now().UnixNano())
	clargs := os.Args[5:]
	// fmt.Println(clargs)

	numOfInputs,_ := strconv.Atoi(clargs[0])
	numOfOutputs,_ := strconv.Atoi(clargs[1])
	numOfPrivateKeys := 50
	// fmt.Printf("\n------------------TxVersion2 Verify Benchmark\n")
	// fmt.Printf("Number of transactions : %d\n", 1)
	// fmt.Printf("Number of inputs       : %d\n", numOfInputs)
	// fmt.Printf("Number of outputs      : %d\n", numOfOutputs)
	// fmt.Println("Will prepare keys")
	keySets, _ := prepareKeySets(numOfPrivateKeys)
	numOfTxs := numOfPrivateKeys

	var txsForBenchmark []*TxVersion1
	for i:=0; i < numOfTxs; i++ {
		keySet := keySets[i]
		pubKey, _ := new(operation.Point).FromBytesS(keySet.PaymentAddress.Pk)
		// fmt.Printf("Will create coins for tx #%d\n",i)
		coins, _ := createAndSaveCoinV1s(2*numOfInputs, 0, keySet.PrivateKey, pubKey, dummyDB)

		tx := new(TxVersion1)
		// r := RandInt() % 90
		inputCoins := coins[:numOfInputs]
		// fmt.Printf("Will create tx params for tx #%d\n",i)
		_, txPrivacyParams, _ := createTxPrivacyInitParams(keySet, inputCoins, true, numOfOutputs)
		// assert.Equal(t, nil, err, "createTxPrivacyInitParams returns an error: %v", err)
		// fmt.Printf("Finished creating tx params for tx #%d\n",i)
		tx.Init(txPrivacyParams)
		txsForBenchmark = append(txsForBenchmark, tx)

	}

	batchLength, _ := strconv.Atoi(clargs[2])
	// each loop verifies x transactions as one batch
	// so the ops/sec will need to be divided by x afterwards
	// for fair comparison
	b.ResetTimer()
	var pass bool
	for loop := 0; loop < b.N; loop++ {
		var batchContent []metadata.Transaction
		for j:=0;j<batchLength;j++{
			chosenIndex := RandInt() % len(txsForBenchmark)
			currentTx := txsForBenchmark[chosenIndex]
			currentTx.ValidateSanityData(nil,nil,nil,0)
			currentTx.ValidateTxWithBlockChain(nil, nil, nil, shardID, dummyDB)

			batchContent = append(batchContent, currentTx)
		}
		batch := NewBatchTransaction(batchContent)
		// fmt.Println("About to verify batch")
		success, _, _ := batch.Validate(dummyDB, nil)
		if !success{
			fmt.Println("Something wrong")
			panic("Invalid tx batch")
		}
		pass = true
	}
	assert.Equal(b,true,pass)
}

func BenchmarkTxV1Verify(b *testing.B) {
	rand.Seed(time.Now().UnixNano())
	clargs := os.Args[5:]
	// fmt.Println(clargs)

	numOfInputs,_ := strconv.Atoi(clargs[0])
	numOfOutputs,_ := strconv.Atoi(clargs[1])
	numOfPrivateKeys := 50
	numOfTxs := numOfPrivateKeys
	// fmt.Printf("\n------------------TxVersion1 Verify Benchmark\n")
	// fmt.Printf("Number of transactions : %d\n", numOfTxs)
	// fmt.Printf("Number of inputs       : %d\n", numOfInputs)
	// fmt.Printf("Number of outputs      : %d\n", numOfOutputs)
	keySets, _ := prepareKeySets(numOfPrivateKeys)


	var txsForBenchmark []*TxVersion1
	for i:=0; i < numOfTxs; i++ {
		keySet := keySets[i]
		pubKey, _ := new(operation.Point).FromBytesS(keySet.PaymentAddress.Pk)
		coins, _ := createAndSaveCoinV1s(2*numOfInputs, 0, keySet.PrivateKey, pubKey, dummyDB)

		tx := new(TxVersion1)
		// r := RandInt() % 90
		inputCoins := coins[:numOfInputs]
		_, txPrivacyParams, _ := createTxPrivacyInitParams(keySet, inputCoins, true, numOfOutputs)
		// assert.Equal(t, nil, err, "createTxPrivacyInitParams returns an error: %v", err)
		tx.Init(txPrivacyParams)
		txsForBenchmark = append(txsForBenchmark, tx)

	}

	// batchLength, _ := strconv.Atoi(clargs[2])
	// each loop verifies 20 transactions as one batch
	// so the ops/sec will need to be divided by 20 afterwards
	// for fair comparison
	b.ResetTimer()
	for loop := 0; loop < b.N; loop++ {
		chosenIndex := RandInt() % len(txsForBenchmark)
		currentTx := txsForBenchmark[chosenIndex]

		var err error
		var isValid bool
		isValid, err = currentTx.ValidateSanityData(nil,nil,nil,0)
		if !isValid{
			panic("Invalid tx sanity")
		}
		isValid, err = currentTx.ValidateTxByItself(true, dummyDB, nil, nil, byte(0), true, nil, nil)
		if !isValid{
			panic("Invalid tx")
		}
		err = currentTx.ValidateTxWithBlockChain(nil, nil, nil, shardID, dummyDB)
		if err!=nil{
			panic("Invalid tx : double spent")
		}
	}
}

func testTxV1JsonMarshaler(tx *TxVersion1, count int, db *statedb.StateDB, t *testing.T){
	someInvalidTxs := getCorruptedJsonDeserializedTxs(tx, count, t)
	for _,theInvalidTx := range someInvalidTxs{
		txSpecific, ok := theInvalidTx.(*TxVersion1)
		if !ok{
			fmt.Println("Skipping a transaction from wrong version")
			continue
		}
		// look for potential panics by calling verify
		isSane, _ := txSpecific.ValidateSanityData(nil, nil, nil, 0)
		// if it doesnt pass sanity then the next validation could panic, it's ok by spec
		if !isSane{
			continue
		}
		isValid, _ := txSpecific.ValidateTxByItself(hasPrivacyForPRV, db, nil, nil, shardID, false, nil, nil)
		if !isValid{
			continue
		}
		txSpecific.ValidateTxWithBlockChain(nil, nil, nil, shardID, db)
	}
}

func testTxTokenV1JsonMarshaler(tx *TxTokenVersion1, count int, db *statedb.StateDB, t *testing.T){
	someInvalidTxs := getCorruptedJsonDeserializedTokenTxs(tx, count, t)
	for _,theInvalidTx := range someInvalidTxs{
		txSpecific, ok := theInvalidTx.(*TxTokenVersion1)
		if !ok{
			fmt.Println("Skipping a transaction from wrong version")
			continue
		}
		// look for potential panics by calling verify
		isSane, _ := txSpecific.ValidateSanityData(nil, nil, nil, 0)
		// if it doesnt pass sanity then the next validation could panic, it's ok by spec
		if !isSane{
			continue
		}
		isValid, _ := txSpecific.ValidateTxByItself(hasPrivacyForPRV, db, nil, nil, shardID, false, nil, nil)
		if !isValid{
			continue
		}
		txSpecific.ValidateTxWithBlockChain(nil, nil, nil, shardID, db)
	}
}