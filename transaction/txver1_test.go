package transaction

import (
	"bytes"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/privacy/coin"
	"github.com/incognitochain/incognito-chain/privacy/operation"
	zkp "github.com/incognitochain/incognito-chain/privacy/privacy_v1/zeroknowledge"
	"github.com/incognitochain/incognito-chain/privacy/privacy_v1/zeroknowledge/aggregatedrange"
	"github.com/stretchr/testify/assert"
	"testing"
)

func createAndSaveCoinV1s(numCoins, numEquals int, privKey privacy.PrivateKey, pubKey *operation.Point, testDB *statedb.StateDB) ([]coin.PlainCoin, error){
	//amount := uint64(common.RandIntInterval(0, 1000000))
	amount := uint64(1000000)
	outCoins := []coin.PlainCoin{}
	for i:=0;i<numEquals;i++{
		coin, err := createSamplePlainCoinV1(privKey, pubKey, 1000, nil)
		if err != nil {
			return nil, err
		}
		outCoins = append(outCoins, coin)
	}
	tmpOutCoins, err := createSamplePlainCoinsFromTotalAmount(privKey, pubKey, amount, numCoins - numEquals, 1)
	for _, coin :=range tmpOutCoins{
		outCoins = append(outCoins, coin)
	}
	if err !=nil {
		return nil, err
	}

	//save coins and commitment indices onto the database
	commitmentsToBeSaved := [][]byte{}
	coinsToBeSaved := [][]byte{}
	for _, outCoin := range outCoins{
		coinsToBeSaved = append(coinsToBeSaved, outCoin.Bytes())
		commitmentsToBeSaved = append(commitmentsToBeSaved, outCoin.GetCommitment().ToBytesS())
	}
	err = statedb.StoreOutputCoins(testDB, common.PRVCoinID, pubKey.ToBytesS(), coinsToBeSaved, shardID)
	if err != nil {
		return nil, err
	}
	err = statedb.StoreCommitments(testDB, common.PRVCoinID, commitmentsToBeSaved, shardID)
	if err != nil {
		return nil, err
	}

	return outCoins, nil
}

func createTxPrivacyInitParams(keySet *incognitokey.KeySet, inputCoins []coin.PlainCoin) (*incognitokey.KeySet, *TxPrivacyInitParams, error) {
	//initialize payment info of input coins
	paymentInfos := coin.CreatePaymentInfosFromPlainCoinsAndAddress(inputCoins, keySet.PaymentAddress, nil)


	//create privacyinitparam
	txPrivacyInitParam := NewTxPrivacyInitParams(&keySet.PrivateKey,
		paymentInfos,
		inputCoins,
		0,
		true,
		testDB,
		&common.PRVCoinID,
		nil,
		nil)

	return keySet, txPrivacyInitParam, nil
}

func TestTxVersion1_ValidateTransaction(t *testing.T){
	keySets, err := prepareKeySets(1)
	assert.Equal(t, nil, err, "prepareKeySets returns an error: %v", err)
	keySet := keySets[0]

	pubKey, err := new(operation.Point).FromBytesS(keySet.PaymentAddress.Pk)
	assert.Equal(t, nil, err, "Cannot parse public key")

	coins, err := createAndSaveCoinV1s(100, 0, keySet.PrivateKey, pubKey, testDB)

	tx := new(TxVersion1)

	//choose some input coins to spend
	inputCoins := coins[:10]
	_, txPrivacyParams, err := createTxPrivacyInitParams(keySet, inputCoins)
	assert.Equal(t, nil, err, "createTxPrivacyInitParams returns an error: %v", err)

	err = validateTxParams(txPrivacyParams)
	assert.Equal(t, nil, err, "validateTxParams returns an error: %v", err)

	err = tx.initializeTxAndParams(txPrivacyParams)
	assert.Equal(t, nil, err, "Init returns an error: %v", err)

	paymentWitnessParamPtr, err := tx.initializePaymentWitnessParam(txPrivacyParams)
	assert.Equal(t, nil, err, "initializePaymentWitnessParam returns an error: %v", err)

	err = tx.proveAndSignCore(txPrivacyParams, paymentWitnessParamPtr)
	assert.Equal(t, nil, err, "proveAndSignCore returns an error: %v")

	res, err := tx.ValidateTransaction(true, testDB, testDB, 0, &common.PRVCoinID, false, false)
	assert.Equal(t, nil, err, "ValidateTransaction returns an error: %v", err)
	assert.Equal(t, true, res)
}

func TestTxVersion1_InputCoinReplication(t *testing.T) {
	keySets, err := prepareKeySets(1)
	assert.Equal(t, nil, err, "prepareKeySets returns an error: %v", err)
	keySet := keySets[0]

	pubKey, err := new(operation.Point).FromBytesS(keySet.PaymentAddress.Pk)
	assert.Equal(t, nil, err, "Cannot parse public key")

	coins, err := createAndSaveCoinV1s(100, 2, keySet.PrivateKey, pubKey, testDB)

	fmt.Println(coins[0].GetValue())

	tx := new(TxVersion1)
	//choose some input coins to spend => make sure that inputCoins[0] and inputCoins[1] have the same amount
	inputCoins := coins[:10]

	_, txPrivacyParams, err := createTxPrivacyInitParams(keySet, inputCoins)
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
	for i :=0;i< sz ; i++ {
		paymentWitnessParamPtr.CommitmentIndices[i+sz] = paymentWitnessParamPtr.CommitmentIndices[i]
		paymentWitnessParamPtr.Commitments[i+sz] = paymentWitnessParamPtr.Commitments[i]
	}

	err = tx.proveAndSignCore(txPrivacyParams, paymentWitnessParamPtr)
	assert.Equal(t, nil, err, "proveAndSignCore returns an error: %v")


	res, err := tx.ValidateTransaction(true, testDB, testDB, 0, &common.PRVCoinID, false, false)

	//assert.Equal(t, nil, err, "ValidateTransaction returns an error: %v", err)
	//if res is true => successfully double-spent!!!
	assert.Equal(t, true, res) // <= this test should failed
	if res {
		assert.Equal(t, true, bytes.Equal(paymentWitnessParamPtr.InputCoins[1].Bytes(), paymentWitnessParamPtr.InputCoins[0].Bytes()))
	}
}

func TestTxVersion1_BulletProofCommitmentConsistency(t *testing.T) {
	keySets, err := prepareKeySets(1)
	assert.Equal(t, nil, err, "prepareKeySets returns an error: %v", err)
	keySet := keySets[0]

	pubKey, err := new(operation.Point).FromBytesS(keySet.PaymentAddress.Pk)
	assert.Equal(t, nil, err, "Cannot parse public key")

	coins, err := createAndSaveCoinV1s(100, 0, keySet.PrivateKey, pubKey, testDB)

	//create 2 transactions
	tx1 := new(TxVersion1)
	tx2 := new(TxVersion1)
	inputCoins := coins[:10]
	_, txPrivacyParams, err := createTxPrivacyInitParams(keySet, inputCoins)
	assert.Equal(t, nil, err, "createTxPrivacyInitParams returns an error: %v", err)

	err = validateTxParams(txPrivacyParams)
	assert.Equal(t, nil, err, "validateTxParams returns an error: %v", err)

	err = tx1.initializeTxAndParams(txPrivacyParams)
	assert.Equal(t, nil, err, "Init returns an error: %v", err)

	paymentWitnessParamPtr, err := tx1.initializePaymentWitnessParam(txPrivacyParams)
	assert.Equal(t, nil, err, "initializePaymentWitnessParam returns an error: %v", err)

	err = tx1.proveAndSignCore(txPrivacyParams, paymentWitnessParamPtr)
	assert.Equal(t, nil, err, "proveAndSignCore returns an error: %v", err)

	res, err := tx1.ValidateTransaction(true, testDB, testDB, 0, &common.PRVCoinID, false, false)
	assert.Equal(t, nil, err, "ValidateTransaction returns an error: %v", err)
	assert.Equal(t, true, res)

	proof, ok := tx1.Proof.(*zkp.PaymentProof)
	assert.Equal(t, true, ok, "cannot parse proof")

	//retrieve the bulletproof
	oldAggregatedProof, ok := proof.GetAggregatedRangeProof().(*aggregatedrange.AggregatedRangeProof)
	assert.Equal(t, true, ok, "cannot parse bulletproof")

	//create a new transaction and reuse the bulletproof
	//choose some input coins to spend
	inputCoins2 := coins[10:20]
	_, newTxPrivacyParams, err := createTxPrivacyInitParams(keySet, inputCoins2)
	fmt.Println("BUGLOG2 output", newTxPrivacyParams.paymentInfo[0].Amount)
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
	proof.SetAggregatedRangeProof(oldAggregatedProof)

	tx2.Sig = nil
	err = tx2.sign()
	assert.Equal(t, nil, err, "sign returns an error: %v", err)

	res, err = tx2.ValidateTransaction(true, testDB, testDB, 0, &common.PRVCoinID, false, false)
	//assert.Equal(t, nil, err, "ValidateTransaction returns an error: %v", err)
	assert.Equal(t, true, res) // <= this test should failed
}

func TestTxVersion1_SerialNumberCommitmentConsistency(t *testing.T){
	keySets, err := prepareKeySets(1)
	assert.Equal(t, nil, err, "prepareKeySets returns an error: %v", err)
	keySet := keySets[0]

	pubKey, err := new(operation.Point).FromBytesS(keySet.PaymentAddress.Pk)
	assert.Equal(t, nil, err, "Cannot parse public key")

	coins, err := createAndSaveCoinV1s(100, 0, keySet.PrivateKey, pubKey, testDB)

	//create 2 transactions
	tx1 := new(TxVersion1)
	tx2 := new(TxVersion1)
	inputCoins := coins[:10]
	_, txPrivacyParams, err := createTxPrivacyInitParams(keySet, inputCoins)
	assert.Equal(t, nil, err, "createTxPrivacyInitParams returns an error: %v", err)

	err = validateTxParams(txPrivacyParams)
	assert.Equal(t, nil, err, "validateTxParams returns an error: %v", err)

	err = tx1.initializeTxAndParams(txPrivacyParams)
	assert.Equal(t, nil, err, "Init returns an error: %v", err)

	paymentWitnessParamPtr, err := tx1.initializePaymentWitnessParam(txPrivacyParams)
	assert.Equal(t, nil, err, "initializePaymentWitnessParam returns an error: %v", err)

	err = tx1.proveAndSignCore(txPrivacyParams, paymentWitnessParamPtr)
	assert.Equal(t, nil, err, "proveAndSignCore returns an error: %v", err)

	res, err := tx1.ValidateTransaction(true, testDB, testDB, 0, &common.PRVCoinID, false, false)
	assert.Equal(t, nil, err, "ValidateTransaction returns an error: %v", err)
	assert.Equal(t, true, res)

	proof, ok := tx1.Proof.(*zkp.PaymentProof)
	assert.Equal(t, true, ok, "cannot parse proof")

	//retrieve the bulletproof
	oldSNProof := proof.GetSerialNumberProof()

	//create a new transaction and reuse the serial number proof
	//choose some input coins to spend
	inputCoins2 := coins[10:20]
	_, newTxPrivacyParams, err := createTxPrivacyInitParams(keySet, inputCoins2)
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

	res, err = tx2.ValidateTransaction(true, testDB, testDB, 0, &common.PRVCoinID, false, false)
	//assert.Equal(t, nil, err, "ValidateTransaction returns an error: %v", err)
	assert.Equal(t, false, res) // <= this test should failed
}
