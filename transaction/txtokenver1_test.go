package transaction

import (
	"testing"
	"fmt"
	"bytes"


	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/privacy/key"
	"github.com/incognitochain/incognito-chain/privacy/coin"
	"github.com/incognitochain/incognito-chain/privacy/operation"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	// "github.com/incognitochain/incognito-chain/privacy/privacy_v1/hybridencryption"
	// "github.com/incognitochain/incognito-chain/wallet"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/stretchr/testify/assert"
)

func TestInitAndTransferTxV1PrivacyToken(t *testing.T) {
	fmt.Printf("\n------------------TxTokenVersion1 Main Test\n")
	for loop := 0; loop < numOfLoops; loop++ {
		var err error
		numOfPrivateKeys := 50
		// numOfInputs := 2
		dummyDB, _ := statedb.NewWithPrefixTrie(emptyRoot, warperDBStatedbTest)
		preparePaymentKeys(numOfPrivateKeys,t)
		
		pastCoins := make([]coin.PlainCoin, 50)
		for i, _ := range pastCoins {
			pubKey, err := new(operation.Point).FromBytesS(keySets[i%numOfPrivateKeys].PaymentAddress.Pk)
			assert.Equal(t,nil,err)
			// tempCoin,err := coin.NewCoinFromPaymentInfo(paymentInfo[i%len(dummyPrivateKeys)])
			c, err := createAndSaveCoinV1s(1, 0, keySets[i%numOfPrivateKeys].PrivateKey, pubKey, dummyDB)
			assert.Equal(t,nil,err)
			pastCoins[i] = c[0]
			// tempCoin.ConcealOutputCoin(keySets[i%len(dummyPrivateKeys)].PaymentAddress.GetPublicView())
		}

		// store a bunch of sample OTA coins in PRV
		// forceSaveCoins(dummyDB, pastCoins, 0, common.PRVCoinID, t)

		// sample message to receiver
		msgCipherText := []byte("haha dummy ciphertext")
		paramToCreateTx,tokenParam := getParamsForTxV1TokenInit(pastCoins[0], dummyDB)
		// create tx for token init
		tx := &TxTokenVersion1{}
		
		err = tx.Init(paramToCreateTx)
		assert.Equal(t, nil, err)
		if err!=nil{
			fmt.Printf("Fatal Error : %v\n",err)
			panic("Test Terminated Early")
		}

		// convert to JSON string and revert
		txJsonString := tx.JSONString()
		txHash := tx.Hash()
		tx1 := new(TxTokenBase)
		tx1.UnmarshalJSON([]byte(txJsonString))
		txHash1 := tx1.Hash()
		assert.Equal(t, txHash, txHash1)

		// size checks
		txActualSize := tx.GetTxActualSize()
		assert.Greater(t, txActualSize, uint64(0))
		sigPubKey := tx.GetSigPubKey()
		assert.Equal(t, common.SigPubKeySize, len(sigPubKey))
		// param checks
		inf := tx.GetTxTokenData().TxNormal.GetProof().GetOutputCoins()[0].GetInfo()
		assert.Equal(t,true,bytes.Equal([]byte(inf),msgCipherText))
		retrievedFee := tx.GetTxFee()
		assert.Equal(t, uint64(10),retrievedFee)
		theAmount := tx.GetTxTokenData().GetAmount()
		assert.Equal(t, tokenParam.Amount, theAmount)
		isUniqueReceiver, _, uniqueAmount, tokenID := tx.GetTransferData()
		assert.Equal(t, true, isUniqueReceiver)
		assert.Equal(t, tokenParam.Amount, uniqueAmount)
		assert.Equal(t, tx.GetTokenID(), tokenID)

		// sanity check
		isValidSanity, err := tx.ValidateSanityData(nil, nil, nil, 0)
		assert.Equal(t, true, isValidSanity)
		assert.Equal(t, nil, err)
		// validate signatures, proofs, etc. Only do after sanity checks are passed
		isValidTxItself, err := tx.ValidateTxByItself(hasPrivacyForPRV, dummyDB, nil, nil, shardID, false, nil, nil)
		assert.Equal(t, true, isValidTxItself)
		assert.Equal(t, nil, err)
		// check double spend using `blockchain data` in this db
		err = tx.ValidateTxWithBlockChain(nil, nil, nil, shardID, dummyDB)
		assert.Equal(t, nil, err)

		// testTxTokenV1JsonMarshaler(tx, 25, dummyDB, t)

		testTxTokenV1InitFakeOutput(tx, dummyDB, paramToCreateTx, t)

		// save the fee outputs into the db
		// get output coin token from tx
		tokenOutputs := tx.GetTxTokenData().TxNormal.GetProof().GetOutputCoins()
		feeOutputs := tx.GetTxBase().GetProof().GetOutputCoins()
		forceSaveCoinsV1(dummyDB, feeOutputs, 0, common.PRVCoinID, t)

		feeOutputBytesHacked := feeOutputs[0].Bytes()
		tokenOutputBytesHacked := tokenOutputs[0].Bytes()
		
		// tx token transfer
		paramToCreateTx2, tokenParam2 := getParamForTxV1TokenTransfer(tx, dummyDB, t)
		_ = tokenParam2
		tx2 := &TxTokenVersion1{}
		fmt.Println("Token Transfer")
		err = tx2.Init(paramToCreateTx2)
		// should fail because db does not have this token yet
		assert.NotEqual(t, nil, err)
		// store the token
		exists := statedb.PrivacyTokenIDExisted(dummyDB,*tx.GetTokenID())
		assert.Equal(t, false, exists)
		statedb.StorePrivacyToken(dummyDB, *tx.GetTokenID(), tokenParam.PropertyName, tokenParam.PropertySymbol, statedb.InitToken, tokenParam.Mintable, tokenParam.Amount, []byte{}, *tx.Hash())

		// statedb.StoreCommitments(dummyDB,*tx.GetTokenID(), [][]byte{tokenOutputs[0].GetCommitment().ToBytesS()}, shardID)
		// check it exists
		exists = statedb.PrivacyTokenIDExisted(dummyDB,*tx.GetTokenID())
		assert.Equal(t, true, exists)
		tx2 = &TxTokenVersion1{}
		paramToCreateTx2, tokenParam2 = getParamForTxV1TokenTransfer(tx, dummyDB, t)
		err = tx2.Init(paramToCreateTx2)
		// still fails because the token's `init` coin (10000 T1) is not stored yet
		isValidSanity, err = tx2.ValidateSanityData(nil, nil, nil, 0)
		assert.Equal(t, true, isValidSanity)
		assert.Equal(t, nil, err)

		isValidTxItself, err = tx2.ValidateTxByItself(hasPrivacyForPRV, dummyDB, nil, nil, shardID, false, nil, nil)
		assert.Equal(t, false, isValidTxItself)
		
		// add the coin. Tx creation should succeed now
		forceSaveCoinsV1(dummyDB, tokenOutputs, 0, *tx.GetTokenID(), t)
		tx2 = &TxTokenVersion1{}
		paramToCreateTx2, tokenParam2 = getParamForTxV1TokenTransfer(tx, dummyDB, t)
		err = tx2.Init(paramToCreateTx2)
		if err != nil{
			fmt.Println(err)
			panic("END")
		}
		assert.Equal(t, nil, err)

		msgCipherText = []byte("doing a transfer")
		assert.Equal(t, true, bytes.Equal(msgCipherText, tx2.GetTxTokenData().TxNormal.GetProof().GetOutputCoins()[0].GetInfo()))

		isValidSanity, err = tx2.ValidateSanityData(nil, nil, nil, 0)
		assert.Equal(t, true, isValidSanity)
		assert.Equal(t, nil, err)

		isValidTxItself, err = tx2.ValidateTxByItself(hasPrivacyForPRV, dummyDB, nil, nil, shardID, false, nil, nil)
		assert.Equal(t, true, isValidTxItself)
		assert.Equal(t, nil, err)
		
		err = tx2.ValidateTxWithBlockChain(nil, nil, nil, shardID, dummyDB)
		assert.Equal(t, nil, err)

		_ = feeOutputBytesHacked
		_ = tokenOutputBytesHacked
	
		testTxTokenV1JsonMarshaler(tx2, 25, dummyDB, t)


		// testTxTokenV2InvalidFee(tx2, dummyDB, t)
		// testTxTokenV2OneFakeOutput(tx2, dummyDB, paramToCreateTx2, t)
		// testTxTokenV2OneDoubleSpentInput(tx2, dummyDB, feeOutputBytesHacked, tokenOutputBytesHacked, t)

		// testTxTokenV2Salary(tx.GetTokenID(), dummyDB, t)
	}
}

func resignTxV1(txv1_generic metadata.Transaction){
	txv1, ok := txv1_generic.(*TxVersion1)
	if !ok{
		panic("Error when casting")
	}
	txv1.cachedHash = nil
	txv1.SetSig(nil)
	txv1.SetSigPubKey(nil)
	err := txv1.sign()
	if err!=nil{
		// if it fails, something's wrong
		panic("Error when resigning")
	}
}

// not used
func testTxTokenV1DeletedProof(txv1 *TxTokenVersion1, db *statedb.StateDB, t *testing.T){
	// try setting the proof to nil, then verify
	// it should not go through
	inner := txv1.GetTxTokenData().TxNormal
	savedProof := inner.GetProof()
	inner.SetProof(nil)
	resignTxV1(inner)
	isValid,_ := txv1.ValidateSanityData(nil,nil,nil,0)
	assert.Equal(t,true,isValid)
	isValidTxItself, _ := txv1.ValidateTxByItself(hasPrivacyForPRV, db, nil, nil, shardID, false, nil, nil)
	assert.Equal(t,false,isValidTxItself)
	// undo the tampering
	inner.SetProof(savedProof)
	resignTxV1(inner)
	isValidTxItself, err := txv1.ValidateTxByItself(hasPrivacyForPRV, db, nil, nil, shardID, false, nil, nil)
	assert.Equal(t, true, isValidTxItself)
	assert.Equal(t, nil, err)

	outer := txv1.GetTxBase()
	savedProof = outer.GetProof()
	outer.SetProof(nil)
	resignTxV1(outer)
	isValid,_ = txv1.ValidateSanityData(nil,nil,nil,0)
	assert.Equal(t,true,isValid)
	isValidTxItself, _ = txv1.ValidateTxByItself(hasPrivacyForPRV, db, nil, nil, shardID, false, nil, nil)
	assert.Equal(t,false,isValidTxItself)
	// undo the tampering
	outer.SetProof(savedProof)
	resignTxV1(outer)
	isValidTxItself, err = txv1.ValidateTxByItself(hasPrivacyForPRV, db, nil, nil, shardID, false, nil, nil)
	assert.Equal(t, true, isValidTxItself)
	assert.Equal(t, nil, err)
}

func testTxTokenV1InitFakeOutput(txv1 *TxTokenVersion1, db *statedb.StateDB, params *TxTokenParams, t *testing.T){
	var err error
	outs := txv1.GetTxTokenData().TxNormal.GetProof().GetOutputCoins()
	tokenOutput,ok := outs[0].(*coin.CoinV1)
	savedCoinBytes := tokenOutput.Bytes()
	assert.Equal(t,true,ok)
	// fmt.Printf("Encrypted? %v\n",tokenOutput.IsEncrypted())
	pc := tokenOutput.CoinDetails
	// set amount from x to 690
	// fmt.Printf("Value : %d\n",txv1.GetTxTokenData().TxNormal.GetProof().(*privacy.ProofV1).GetOutputCoins()[0].GetValue())
	pc.SetValue(690)
	pc.CommitAll()
	// fmt.Printf("Value : %d\n",txv1.GetTxTokenData().TxNormal.GetProof().(*privacy.ProofV1).GetOutputCoins()[0].GetValue())
	
	inner, ok := txv1.GetTxTokenData().TxNormal.(*TxVersion1)
	assert.Equal(t,true,ok)
	inner.Proof.SetOutputCoins(coin.ArrayCoinV1ToCoin([]*coin.CoinV1{tokenOutput}))
	inner.cachedHash = nil
	// isSane,_ := txv1.ValidateSanityData(nil,nil,nil,0)
	// assert.Equal(t,false,isSane)
	isValid,err := inner.ValidateTxByItself(true,db,nil,nil,0,false,nil,nil)
	// verify must fail
	assert.Equal(t,false,isValid)

	pc.SetValue(10000)
	pc.CommitAll()
	inner.Proof.SetOutputCoins(coin.ArrayCoinV1ToCoin([]*coin.CoinV1{tokenOutput}))
	inner.cachedHash = nil
	_ = savedCoinBytes
	_ = err
	// fmt.Printf("Fake output (wrong amount) -> %+v\n",err)
	// // undo the tampering
	// tokenOutput.SetBytes(savedCoinBytes)
	// outs[0] = tokenOutput
	// txv1.GetTxTokenData().TxNormal.GetProof().SetOutputCoins(outs)
	// // resignTxV1(txv1.GetTxTokenData().TxNormal)
	// assert.Equal(t,nil,err)
	// isValid,err = txv1.ValidateTxByItself(true,db,nil,nil,0,false,nil,nil)
	// assert.Equal(t,true,isValid)
}

func testTxTokenV1OneFakeOutput(txv1 *TxTokenVersion1, db *statedb.StateDB, params *TxTokenParams, t *testing.T){
	var err error
	outs := txv1.GetTxTokenData().TxNormal.GetProof().GetOutputCoins()
	tokenOutput,ok := outs[0].(*coin.CoinV1)
	savedCoinBytes := tokenOutput.Bytes()
	assert.Equal(t,true,ok)
	pc := tokenOutput.CoinDetails
	// set amount from x to 690
	pc.SetValue(690)
	resignTxV1(txv1.GetTxTokenData().TxNormal)
	assert.Equal(t,nil,err)
	isValid,err := txv1.ValidateTxByItself(true,db,nil,nil,0,false,nil,nil)
	// verify must fail
	assert.Equal(t,false,isValid)
	fmt.Printf("Fake output (wrong amount) -> %+v\n",err)
	// undo the tampering
	tokenOutput.SetBytes(savedCoinBytes)
	outs[0] = tokenOutput
	txv1.GetTxTokenData().TxNormal.GetProof().SetOutputCoins(outs)
	resignTxV1(txv1.GetTxTokenData().TxNormal)
	assert.Equal(t,nil,err)
	isValid,err = txv1.ValidateTxByItself(true,db,nil,nil,0,false,nil,nil)
	assert.Equal(t,true,isValid)

	// now instead of changing amount, we change the OTA public key
	// theProof := txv1.GetTxTokenData().TxNormal.GetProof()
	// outs = theProof.GetOutputCoins()
	// tokenOutput,ok = outs[0].(*coin.CoinV2)
	// savedCoinBytes = tokenOutput.Bytes()
	// assert.Equal(t,true,ok)
	// payInf := &privacy.PaymentInfo{PaymentAddress: keySets[0].PaymentAddress, Amount: uint64(69), Message: []byte("doing a transfer")}
	// // totally fresh OTA of the same amount, meant for the same PaymentAddress
	// newCoin,err  := coin.NewCoinFromPaymentInfo(payInf)
	// assert.Equal(t,nil,err)
	// newCoin.ConcealOutputCoin(keySets[0].PaymentAddress.GetPublicView())
	// theProofSpecific, ok := theProof.(*privacy.ProofV2)
	// theBulletProof, ok := theProofSpecific.GetAggregatedRangeProof().(*privacy.AggregatedRangeProofV2)
	// cmsv := theBulletProof.GetCommitments()
	// cmsv[0] = newCoin.GetCommitment()
	// outs[0] = newCoin
	// txv1.GetTxTokenData().TxNormal.GetProof().SetOutputCoins(outs)
	// err = resignUnprovenTxToken([]*incognitokey.KeySet{keySets[0]}, txv1, params, nil)
	// assert.Equal(t,nil,err)
	// isValid,err = txv1.ValidateTxByItself(true,db,nil,nil,0,false,nil,nil)
	// // verify must fail
	// assert.Equal(t,false,isValid)
	// // fmt.Printf("Fake output (wrong receiving OTA) -> %+v\n",err)
	// // undo the tampering
	// tokenOutput.SetBytes(savedCoinBytes)
	// outs[0] = tokenOutput
	// cmsv[0] = tokenOutput.GetCommitment()
	// txv1.GetTxTokenData().TxNormal.GetProof().SetOutputCoins(outs)
	// err = resignUnprovenTxToken([]*incognitokey.KeySet{keySets[0]}, txv1, params, nil)
	// assert.Equal(t,nil,err)
	// isValid,err = txv1.ValidateTxByItself(true,db,nil,nil,0,false,nil,nil)
	// assert.Equal(t,true,isValid)
}

func getParamsForTxV1TokenInit(theInputCoin coin.PlainCoin, db *statedb.StateDB) (*TxTokenParams,*TokenParam){
	msgCipherText := []byte("haha dummy ciphertext")
	initAmount := uint64(10000)
	tokenPayments := []*privacy.PaymentInfo{{PaymentAddress: keySets[0].PaymentAddress, Amount: initAmount, Message: msgCipherText}}

	myOnlyInputCoin := theInputCoin
	inputCoinsPRV := []coin.PlainCoin{myOnlyInputCoin}
	paymentInfoPRV := []*privacy.PaymentInfo{key.InitPaymentInfo(keySets[0].PaymentAddress,uint64(990),[]byte("test out"))}

	// token param for init new token
	tokenParam := &TokenParam{
		PropertyID:     "",
		PropertyName:   "Token 1",
		PropertySymbol: "T1",
		Amount:         initAmount,
		TokenTxType:    CustomTokenInit,
		Receiver:       tokenPayments,
		TokenInput:     []coin.PlainCoin{},
		Mintable:       false,
		Fee:            0,
	}

	paramToCreateTx := NewTxTokenParams(&keySets[0].PrivateKey,
		paymentInfoPRV, inputCoinsPRV, 10, tokenParam, db, nil,
		hasPrivacyForPRV, hasPrivacyForToken, shardID, []byte{},db)
	return paramToCreateTx, tokenParam
}

func getParamForTxV1TokenTransfer(txTokenInit *TxTokenVersion1, db *statedb.StateDB, t *testing.T) (*TxTokenParams,*TokenParam){
	transferAmount := uint64(69)
	msgCipherText := []byte("doing a transfer")
	paymentInfo2 := []*privacy.PaymentInfo{{PaymentAddress: keySets[0].PaymentAddress, Amount: transferAmount, Message: msgCipherText}}
	paymentInfoFee := []*privacy.PaymentInfo{{PaymentAddress: keySets[0].PaymentAddress, Amount: 850, Message: msgCipherText}}

	feeOutputs := txTokenInit.GetTxBase().GetProof().GetOutputCoins()
	prvCoinsToPayTransfer := make([]coin.PlainCoin,0)
	tokenOutputs := txTokenInit.GetTxTokenData().TxNormal.GetProof().GetOutputCoins()
	tokenCoinsToTransfer := make([]coin.PlainCoin,0)
	for _,c := range feeOutputs{
		cloneCoin := coin.CoinV1{}
		cloneCoin.SetBytes(c.Bytes())
		pc,_ := cloneCoin.Decrypt(keySets[0])
		// s,_ := json.Marshal(pc.(*coin.CoinV2))
		fmt.Printf("Tx Fee : %x has received %d in PRV\n",pc.GetPublicKey().ToBytesS(),pc.GetValue())
		prvCoinsToPayTransfer = append(prvCoinsToPayTransfer,pc)
	}
	for _,c := range tokenOutputs{
		cloneCoin := coin.CoinV1{}
		cloneCoin.SetBytes(c.Bytes())
		pc,err := cloneCoin.Decrypt(keySets[0])
		// s,_ := json.Marshal(pc.(*coin.CoinV2))
		fmt.Printf("Tx Token : %x has received %d in token T1\n",pc.GetPublicKey().ToBytesS(),pc.GetValue())
		assert.Equal(t,nil,err)
		tokenCoinsToTransfer = append(tokenCoinsToTransfer,pc)
	}
	// // token param for transfer token
	tokenParam2 := &TokenParam{
		PropertyID:     txTokenInit.GetTokenID().String(),
		PropertyName:   "Token 1",
		PropertySymbol: "T1",
		Amount:         transferAmount,
		TokenTxType:    CustomTokenTransfer,
		Receiver:       paymentInfo2,
		TokenInput:     tokenCoinsToTransfer,
		Mintable:       false,
		Fee:            0,
	}

	paramToCreateTx2 := NewTxTokenParams(&keySets[0].PrivateKey,
		paymentInfoFee, prvCoinsToPayTransfer, 140, tokenParam2, db, nil,
		hasPrivacyForPRV, hasPrivacyForToken, shardID, []byte{},db)
	return  paramToCreateTx2, tokenParam2
}

func forceSaveCoinsV1(db *statedb.StateDB, coinsToBeSaved []coin.Coin, shardID byte, tokenID common.Hash, t *testing.T){
	// coinsInBytes := make([][]byte, 0)
	// publicKeys := make([][]byte,0)
	commitmentsInBytes := make([][]byte, 0)
	for _,c := range coinsToBeSaved{
		if t!=nil{
			assert.Equal(t,1,int(c.GetVersion()))
		}
		err := statedb.StoreOutputCoins(db, tokenID, c.GetPublicKey().ToBytesS(), [][]byte{c.Bytes()}, shardID)
		if t!=nil{
			assert.Equal(t,nil,err)
		}
		// coinsInBytes = append(coinsInBytes, c.Bytes())
		// publicKeys = append(publicKeys, c.GetPublicKey().ToBytesS())
		commitmentsInBytes = append(commitmentsInBytes, c.GetCommitment().ToBytesS())
	}

	// err = statedb.StoreOutputCoins(dummyDB, common.PRVCoinID, publicKeys, coinsToBeSaved, shardID)
	// if err != nil {
	// 	return nil, err
	// }
	err := statedb.StoreCommitments(db, tokenID, commitmentsInBytes, shardID)
	if t != nil {
		assert.Equal(t,nil,err)
	}
}

// func TestInitTxV1PrivacyToken(t *testing.T) {
// 	for i := 0; i < 1; i++ {
// 		//Generate sender private key & receiver payment address
// 		seed := privacy.RandomScalar().ToBytesS()
// 		masterKey, _ := wallet.NewMasterKey(seed)
// 		childSender, _ := masterKey.NewChildKey(uint32(1))
// 		privKeyB58 := childSender.Base58CheckSerialize(wallet.PriKeyType)
// 		childReceiver, _ := masterKey.NewChildKey(uint32(2))
// 		paymentAddressB58 := childReceiver.Base58CheckSerialize(wallet.PaymentAddressType)

// 		// sender key
// 		senderKey, err := wallet.Base58CheckDeserialize(privKeyB58)
// 		assert.Equal(t, nil, err)

// 		err = senderKey.KeySet.InitFromPrivateKey(&senderKey.KeySet.PrivateKey)
// 		assert.Equal(t, nil, err)

// 		//receiver key
// 		receiverKey, _ := wallet.Base58CheckDeserialize(paymentAddressB58)
// 		receiverPaymentAddress := receiverKey.KeySet.PaymentAddress

// 		shardID := common.GetShardIDFromLastByte(senderKey.KeySet.PaymentAddress.Pk[len(senderKey.KeySet.PaymentAddress.Pk)-1])

// 		// message to receiver
// 		msg := "Incognito-chain"
// 		receiverTK, _ := new(privacy.Point).FromBytesS(senderKey.KeySet.PaymentAddress.Tk)
// 		msgCipherText, _ := hybridencryption.HybridEncrypt([]byte(msg), receiverTK)

// 		initAmount := uint64(10000)
// 		tokenPayments := []*privacy.PaymentInfo{{PaymentAddress: senderKey.KeySet.PaymentAddress, Amount: initAmount, Message: msgCipherText.Bytes()}}

// 		inputCoinsPRV := []coin.PlainCoin{}
// 		paymentInfoPRV := []*privacy.PaymentInfo{}

// 		// token param for init new token
// 		tokenParam := &TokenParam{
// 			PropertyID:     "",
// 			PropertyName:   "Token 1",
// 			PropertySymbol: "Token 1",
// 			Amount:         initAmount,
// 			TokenTxType:    CustomTokenInit,
// 			Receiver:       tokenPayments,
// 			TokenInput:     []*coin.PlainCoinV1{},
// 			Mintable:       false,
// 			Fee:            0,
// 		}

// 		hasPrivacyForPRV := false
// 		hasPrivacyForToken := false

// 		paramToCreateTx := NewTxTokenParams(&senderKey.KeySet.PrivateKey,
// 			paymentInfoPRV, inputCoinsPRV, 0, tokenParam, db, nil,
// 			hasPrivacyForPRV, hasPrivacyForToken, shardID, []byte{})

// 		// init tx
// 		tx := new(TxTokenBase)
// 		err = tx.Init(paramToCreateTx)
// 		assert.Equal(t, nil, err)

// 		assert.Equal(t, len(msgCipherText.Bytes()), len(tx.TxPrivacyTokenDataVersion1.TxNormal.Proof.GetOutputCoins()[0].CoinDetails.GetInfo()))

// 		//fmt.Printf("Tx: %v\n", tx.TxPrivacyTokenDataVersion1.TxNormal.Proof.GetOutputCoins()[0].CoinDetails.GetInfo())

// 		// convert to JSON string and revert
// 		txJsonString := tx.JSONString()
// 		txHash := tx.Hash()

// 		tx1 := new(TxTokenBase)
// 		tx1.UnmarshalJSON([]byte(txJsonString))
// 		txHash1 := tx1.Hash()
// 		assert.Equal(t, txHash, txHash1)

// 		// get actual tx size
// 		txActualSize := tx.GetTxActualSize()
// 		assert.Greater(t, txActualSize, uint64(0))

// 		txPrivacyTokenActualSize := tx.GetTxPrivacyTokenActualSize()
// 		assert.Greater(t, txPrivacyTokenActualSize, uint64(0))

// 		//isValidFee := tx.CheckTransactionFee(uint64(0))
// 		//assert.Equal(t, true, isValidFee)

// 		//isValidFeeToken := tx.CheckTransactionFeeByFeeToken(uint64(0))
// 		//assert.Equal(t, true, isValidFeeToken)
// 		//
// 		//isValidFeeTokenForTokenData := tx.CheckTransactionFeeByFeeTokenForTokenData(uint64(0))
// 		//assert.Equal(t, true, isValidFeeTokenForTokenData)

// 		isValidType := tx.ValidateType()
// 		assert.Equal(t, true, isValidType)

// 		//err = tx.ValidateTxWithCurrentMempool(nil)
// 		//assert.Equal(t, nil, err)

// 		err = tx.ValidateTxWithBlockChain(nil, shardID, nil, nil, db)
// 		assert.Equal(t, nil, err)

// 		isValidSanity, err := tx.ValidateSanityData(nil, nil, nil)
// 		assert.Equal(t, true, isValidSanity)
// 		assert.Equal(t, nil, err)

// 		isValidTxItself, err := tx.ValidateTxByItself(hasPrivacyForPRV, db, nil, shardID, nil, nil)
// 		assert.Equal(t, true, isValidTxItself)
// 		assert.Equal(t, nil, err)

// 		//isValidTx, err := tx.ValidateTransaction(hasPrivacyForPRV, db, shardID, tx.GetTokenID())
// 		//fmt.Printf("Err: %v\n", err)
// 		//assert.Equal(t, true, isValidTx)
// 		//assert.Equal(t, nil, err)

// 		_ = tx.GetProof()
// 		//assert.Equal(t, nil, proof)

// 		pubKeyReceivers, amounts := tx.GetTokenReceivers()
// 		assert.Equal(t, 1, len(pubKeyReceivers))
// 		assert.Equal(t, 1, len(amounts))
// 		assert.Equal(t, initAmount, amounts[0])

// 		isUniqueReceiver, uniquePubKey, uniqueAmount, tokenID := tx.GetTransferData()
// 		assert.Equal(t, true, isUniqueReceiver)
// 		assert.Equal(t, initAmount, uniqueAmount)
// 		assert.Equal(t, tx.GetTokenID(), tokenID)
// 		receiverPubKeyBytes := make([]byte, common.PublicKeySize)
// 		copy(receiverPubKeyBytes, senderKey.KeySet.PaymentAddress.Pk)
// 		assert.Equal(t, uniquePubKey, receiverPubKeyBytes)

// 		//TODO: Fix IsCoinBurining
// 		//isCoinBurningTx := tx.IsCoinsBurning()
// 		//assert.Equal(t, false, isCoinBurningTx)

// 		txValue := tx.CalculateTxValue()
// 		assert.Equal(t, initAmount, txValue)

// 		listSerialNumber := tx.ListSerialNumbersHashH()
// 		assert.Equal(t, 0, len(listSerialNumber))

// 		sigPubKey := tx.GetSigPubKey()
// 		assert.Equal(t, common.SigPubKeySize, len(sigPubKey))

// 		// store init tx

// 		// get output coin token from tx
// 		//outputCoins := ConvertOutputCoinToInputCoin(tx.TxPrivacyTokenDataVersion1.TxNormal.Proof.GetOutputCoins())

// 		// calculate serial number for input coins
// 		serialNumber := new(privacy.Point).Derive(privacy.PedCom.G[privacy.PedersenPrivateKeyIndex],
// 			new(privacy.Scalar).FromBytesS(senderKey.KeySet.PrivateKey),
// 			outputCoins[0].GetSNDerivator())
// 		outputCoins[0].SetKeyImage(serialNumber)

// 		db.StorePrivacyToken(*tx.GetTokenID(), tx.Hash()[:])
// 		db.StoreCommitments(*tx.GetTokenID(), senderKey.KeySet.PaymentAddress.Pk[:], [][]byte{outputCoins[0].CoinDetails.GetCoinCommitment().ToBytesS()}, shardID)

// 		//listTokens, err := db.ListPrivacyToken()
// 		//assert.Equal(t, nil, err)
// 		//assert.Equal(t, 1, len(listTokens))

// 		transferAmount := uint64(10)

// 		paymentInfo2 := []*privacy.PaymentInfo{{PaymentAddress: receiverPaymentAddress, Amount: transferAmount, Message: msgCipherText.Bytes()}}

// 		// token param for transfer token
// 		tokenParam2 := &TokenParam{
// 			PropertyID:     tx.GetTokenID().String(),
// 			PropertyName:   "Token 1",
// 			PropertySymbol: "Token 1",
// 			Amount:         transferAmount,
// 			TokenTxType:    CustomTokenTransfer,
// 			Receiver:       paymentInfo2,
// 			TokenInput:     outputCoins,
// 			Mintable:       false,
// 			Fee:            0,
// 		}

// 		paramToCreateTx2 := NewTxTokenParams(&senderKey.KeySet.PrivateKey,
// 			paymentInfoPRV, inputCoinsPRV, 0, tokenParam2, db, nil,
// 			hasPrivacyForPRV, true, shardID, []byte{})

// 		// init tx
// 		tx2 := new(TxTokenBase)
// 		err = tx2.Init(paramToCreateTx2)
// 		assert.Equal(t, nil, err)

// 		assert.Equal(t, len(msgCipherText.Bytes()), len(tx2.TxPrivacyTokenDataVersion1.TxNormal.Proof.GetOutputCoins()[0].CoinDetails.GetInfo()))

// 		err = tx2.ValidateTxWithBlockChain(nil, shardID, nil, nil, db)
// 		assert.Equal(t, nil, err)

// 		isValidSanity, err = tx2.ValidateSanityData(nil, nil, nil)
// 		assert.Equal(t, true, isValidSanity)
// 		assert.Equal(t, nil, err)

// 		isValidTxItself, err = tx2.ValidateTxByItself(hasPrivacyForPRV, db, nil, shardID, nil, nil)
// 		assert.Equal(t, true, isValidTxItself)
// 		assert.Equal(t, nil, err)

// 		txValue2 := tx2.CalculateTxValue()
// 		assert.Equal(t, uint64(0), txValue2)
// 	}
// }