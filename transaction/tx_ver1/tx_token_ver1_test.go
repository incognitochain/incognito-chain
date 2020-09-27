package tx_ver1

import (
	"testing"
	"fmt"
	"bytes"

	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/transaction/tx_generic"
	"github.com/incognitochain/incognito-chain/transaction/utils"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/stretchr/testify/assert"
)

func preparePaymentKeys(count int, t *testing.T){
	// create many random private keys
	// then use each privatekey to derive Incognito keyset (various keys for everything inside the protocol)
	// we ensure they all belong in shard 0 for this test

	// PaymentInfo is like `intent` for making Coin.
	// the paymentInfo slice here will be used to create pastCoins & inputCoins
	// we populate `value` fields with some arbitrary, big-enough constant (here, 4000*len)
	// `message` field can be anything
	dummyPrivateKeys = make([]*privacy.PrivateKey,count)
	keySets = make([]*incognitokey.KeySet,len(dummyPrivateKeys))
	paymentInfo = make([]*privacy.PaymentInfo, len(dummyPrivateKeys))
	for i := 0; i < count; i += 1 {
		for{
			privateKey := privacy.GeneratePrivateKey(common.RandBytes(32))
			dummyPrivateKeys[i] = &privateKey
			keySets[i] = new(incognitokey.KeySet)
			err := keySets[i].InitFromPrivateKey(dummyPrivateKeys[i])
			if t!=nil{
				assert.Equal(t, nil, err)
			}
			paymentInfo[i] = &privacy.PaymentInfo{PaymentAddress: keySets[i].PaymentAddress, Amount: uint64(4000*len(dummyPrivateKeys)), Message: []byte("test in")}
			pkb := []byte(paymentInfo[i].PaymentAddress.Pk)
			if common.GetShardIDFromLastByte(pkb[len(pkb)-1])==shardID{
				break
			}
		}
	}
	// fmt.Println("Key & PaymentInfo generation finished")
}

func TestInitAndTransferTxV1PrivacyToken(t *testing.T) {
	fmt.Printf("\n------------------TxTokenVersion1 Main Test\n")
	for loop := 0; loop < numOfLoops; loop++ {
		var err error
		numOfPrivateKeys := 10
		// numOfInputs := 2
		preparePaymentKeys(numOfPrivateKeys,t)

		pastCoins := make([]privacy.PlainCoin, 50)
		for i, _ := range pastCoins {
			pubKey, err := new(privacy.Point).FromBytesS(keySets[i%numOfPrivateKeys].PaymentAddress.Pk)
			assert.Equal(t,nil,err)
			// tempCoin,err := privacy.NewCoinFromPaymentInfo(paymentInfo[i%len(dummyPrivateKeys)])
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
		tx := &TxToken{}
		fmt.Println("Token Init")
		err = tx.Init(paramToCreateTx)
		assert.Equal(t, nil, err)
		if err!=nil{
			fmt.Printf("Fatal Error : %v\n",err)
			panic("Test Terminated Early")
		}

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

		testTxTokenV1JsonMarshaler(tx, 25, dummyDB, t)
		// TODO : resolve the issue where signature in txNormal is not verified
		// testTxTokenV1InitFakeOutput(tx, dummyDB, paramToCreateTx, t)

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
		tx2 := &TxToken{}
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
		tx2 = &TxToken{}
		paramToCreateTx2, tokenParam2 = getParamForTxV1TokenTransfer(tx, dummyDB, t)
		err = tx2.Init(paramToCreateTx2)
		// still fails because the token's `init` coin (10000 T1) is not stored yet
		isValidSanity, err = tx2.ValidateSanityData(nil, nil, nil, 0)
		assert.Equal(t, true, isValidSanity)
		assert.Equal(t, nil, err)

		isValidTxItself, err = tx2.ValidateTxByItself(hasPrivacyForPRV, dummyDB, nil, nil, shardID, false, nil, nil)
		assert.Equal(t, false, isValidTxItself)

		// add the coin . Tx creation should succeed now
		forceSaveCoinsV1(dummyDB, tokenOutputs, 0, *tx.GetTokenID(), t)
		tx2 = &TxToken{}
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

		testTxTokenV1DuplicateInput(tx, dummyDB, t)
		testTxTokenV1InvalidFee(tx2, dummyDB, t)
		testTxTokenV1TransferFakeOutput(tx2, dummyDB, paramToCreateTx2, t)
		testTxTokenV1OneDoubleSpentInput(tx2, dummyDB, feeOutputBytesHacked, tokenOutputBytesHacked, t)

	}
}

func resignTxV1(txv1_generic metadata.Transaction){
	txv1, ok := txv1_generic.(*Tx)
	if !ok{
		panic("Error when casting")
	}
	txv1.SetCachedHash(nil)
	txv1.SetSig(nil)
	txv1.SetSigPubKey(nil)
	err := txv1.sign()
	if err!=nil{
		// if it fails, something's wrong
		panic("Error when resigning")
	}
}

// not used
func testTxTokenV1DeletedProof(txv1 *TxToken, db *statedb.StateDB, t *testing.T){
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

func testTxTokenV1InitFakeOutput(txv1 *TxToken, db *statedb.StateDB, params *tx_generic.TxTokenParams, t *testing.T){
	outs := txv1.GetTxTokenData().TxNormal.GetProof().GetOutputCoins()
	tokenOutput,ok := outs[0].(*privacy.CoinV1)
	assert.Equal(t,true,ok)
	// fmt.Printf("Encrypted? %v\n",tokenOutput.IsEncrypted())
	pc := tokenOutput.CoinDetails
	// set amount from x to 690
	pc.SetValue(690)
	pc.CommitAll()

	inner, ok := txv1.GetTxTokenData().TxNormal.(*Tx)
	assert.Equal(t,true,ok)
	inner.Proof.SetOutputCoins(outs)
	inner.SetCachedHash(nil)
	// isSane,_ := txv1.ValidateSanityData(nil,nil,nil,0)
	// assert.Equal(t,false,isSane)
	isValid,err := txv1.ValidateTxByItself(true,db,nil,nil,0,false,nil,nil)
	// verify must fail
	assert.Equal(t,false,isValid)
	fmt.Printf("Fake token init (wrong amount) -> %v\n",err)

	pc.SetValue(10000)
	pc.CommitAll()
	inner.Proof.SetOutputCoins(outs)
	inner.SetCachedHash(nil)
	isValid,_ = txv1.ValidateTxByItself(true,db,nil,nil,0,false,nil,nil)
	assert.Equal(t,true,isValid)
}

func testTxTokenV1InvalidFee(txv1 *TxToken, db *statedb.StateDB, t *testing.T){
	// a set of init params where fee is changed so mlsag should verify to false
	// let's say someone tried to use this invalid fee for tx
	// we should encounter an error here

	// set fee to increase by 1000PRV
	savedFee := txv1.GetTxBase().GetTxFee()
	txv1.GetTxBase().SetTxFee(savedFee + 1000)

	// sanity should pass
	isValidSanity, err := txv1.ValidateSanityData(nil, nil, nil, 0)
	assert.Equal(t, true, isValidSanity)
	assert.Equal(t, nil, err)

	// should reject at signature since fee & output doesn't sum to input
	isValidTxItself, err := txv1.ValidateTxByItself(hasPrivacyForPRV, db, nil, nil, shardID, false, nil, nil)
	assert.Equal(t, false, isValidTxItself)
	fmt.Printf("Invalid fee -> %v\n",err)

	// undo the tampering
	txv1.GetTxBase().SetTxFee(savedFee)
	isValidTxItself, _ = txv1.ValidateTxByItself(hasPrivacyForPRV, db, nil, nil, shardID, false, nil, nil)
	assert.Equal(t, true, isValidTxItself)
}

func testTxTokenV1TransferFakeOutput(txv1 *TxToken, db *statedb.StateDB, params *tx_generic.TxTokenParams, t *testing.T){
	outs := txv1.GetTxTokenData().TxNormal.GetProof().GetOutputCoins()
	tokenOutput,ok := outs[0].(*privacy.CoinV1)
	assert.Equal(t,true,ok)
	clonedCoin := &privacy.CoinV1{}
	err := clonedCoin.SetBytes(tokenOutput.Bytes())
	assert.Equal(t,nil,err)
	pcGeneric, err := clonedCoin.Decrypt(keySets[0])
	assert.Equal(t,nil,err)
	pc, ok := pcGeneric.(*privacy.PlainCoinV1)
	assert.Equal(t,true,ok)
	// set amount from x to 690
	pc.SetValue(690)
	pc.CommitAll()
	forgedCoin := &privacy.CoinV1{}
	forgedCoin.CoinDetails = pc
	err = forgedCoin.Encrypt(keySets[0].PaymentAddress.Tk)

	inner, ok := txv1.GetTxTokenData().TxNormal.(*Tx)
	assert.Equal(t,true,ok)
	outs[0] = forgedCoin
	inner.Proof.SetOutputCoins(outs)
	inner.SetCachedHash(nil)
	resignTxV1(inner)
	isValid,err := txv1.ValidateTxByItself(hasPrivacyForPRV,db,nil,nil,0,false,nil,nil)
	// verify must fail
	assert.Equal(t,false,isValid)
	fmt.Printf("Fake output (wrong amount) -> %v\n",err)
	outs[0] = tokenOutput
	inner.Proof.SetOutputCoins(outs)
	inner.SetCachedHash(nil)
	resignTxV1(inner)
	isValid,_ = txv1.ValidateTxByItself(hasPrivacyForPRV,db,nil,nil,0,false,nil,nil)
	assert.Equal(t,true,isValid)

	// now instead of changing amount, we change the receiving public key
	// this time we use the old commitment
	aDifferentPk, err := (&privacy.Point{}).FromBytesS(keySets[2].PaymentAddress.Pk)
	assert.Equal(t,nil,err)
	clonedCoin = &privacy.CoinV1{}
	err = clonedCoin.SetBytes(tokenOutput.Bytes())
	// assert.Equal(t,nil,err)
	// pcGeneric, err = clonedCoin.Decrypt(keySets[0])
	// assert.Equal(t,nil,err)
	pc = clonedCoin.CoinDetails
	assert.Equal(t,true,ok)
	pc.SetPublicKey(aDifferentPk)
	// pc.SetCommitment(tokenOutput.GetCommitment())
	forgedCoin = clonedCoin
	// forgedCoin.CoinDetails = pc
	// err = forgedCoin.Encrypt(keySets[2].PaymentAddress.Tk)

	outs[0] = forgedCoin
	inner.Proof.SetOutputCoins(outs)
	inner.SetCachedHash(nil)
	resignTxV1(inner)
	isValid,err = txv1.ValidateTxByItself(hasPrivacyForPRV,db,nil,nil,0,false,nil,nil)
	// verify must fail
	assert.Equal(t,false,isValid)
	fmt.Printf("Fake output (wrong receiver) -> %v\n",err)
	outs[0] = tokenOutput
	inner.Proof.SetOutputCoins(outs)
	inner.SetCachedHash(nil)
	resignTxV1(inner)
	isValid,_ = txv1.ValidateTxByItself(hasPrivacyForPRV,db,nil,nil,0,false,nil,nil)
	assert.Equal(t,true,isValid)
}

func testTxTokenV1OneDoubleSpentInput(tokenTx *TxToken, db *statedb.StateDB, feeOutputBytesHacked, tokenOutputBytesHacked []byte, t *testing.T){
	// save both fee&token outputs from previous tx
	otaBytes := [][]byte{tokenTx.GetTxTokenData().TxNormal.GetProof().GetInputCoins()[0].GetKeyImage().ToBytesS()}
	statedb.StoreSerialNumbers(db, *tokenTx.GetTokenID(), otaBytes, 0)
	otaBytes = [][]byte{tokenTx.GetTxBase().GetProof().GetInputCoins()[0].GetKeyImage().ToBytesS()}
	statedb.StoreSerialNumbers(db, common.PRVCoinID, otaBytes, 0)

	tokenOutputs := tokenTx.GetTxTokenData().TxNormal.GetProof().GetOutputCoins()
	feeOutputs := tokenTx.GetTxBase().GetProof().GetOutputCoins()
	forceSaveCoinsV1(db, feeOutputs, 0, common.PRVCoinID, t)
	forceSaveCoinsV1(db, tokenOutputs, 0, *tokenTx.GetTokenID(), t)

	// firstly, using the output coins to create new tx should be successful
	pr,_ := getParamForTxV1TokenTransfer(tokenTx, db, t)
	tx := &TxToken{}
	err := tx.Init(pr)
	if err !=  nil{
		fmt.Println(err)
		panic("END")
	}
	assert.Equal(t,nil,err)
	isValidSanity, err := tx.ValidateSanityData(nil, nil, nil, 0)
	assert.Equal(t, true, isValidSanity)
	assert.Equal(t, nil, err)
	isValidTxItself, err := tx.ValidateTxByItself(hasPrivacyForPRV, db, nil, nil, shardID, false, nil, nil)
	assert.Equal(t, true, isValidTxItself)
	assert.Equal(t, nil, err)
	err = tx.ValidateTxWithBlockChain(nil, nil ,nil, 0, db)
	assert.Equal(t,nil,err)

	// now we try to swap in a used input for txfee
	doubleSpendingFeeInput := &privacy.CoinV1{}
	doubleSpendingFeeInput.SetBytes(feeOutputBytesHacked)
	pc,_ := doubleSpendingFeeInput.Decrypt(keySets[0])
	pr,_ = getParamForTxV1TokenTransfer(tokenTx, db, t)
	pr.InputCoin = []privacy.PlainCoin{pc}
	tx = &TxToken{}
	err = tx.Init(pr)
	assert.Equal(t,nil,err)
	isValidSanity, err = tx.ValidateSanityData(nil, nil, nil, 0)
	assert.Equal(t, true, isValidSanity)
	assert.Equal(t, nil, err)
	isValidTxItself, err = tx.ValidateTxByItself(hasPrivacyForPRV, db, nil, nil, shardID, false, nil, nil)
	assert.Equal(t, true, isValidTxItself)
	assert.Equal(t, nil, err)
	err = tx.ValidateTxWithBlockChain(nil, nil ,nil, 0, db)
	fmt.Printf("Double spent input (fee) -> %v\n",err)
	assert.NotEqual(t,nil,err)

	// now we try to swap in a used token input
	doubleSpendingTokenInput := &privacy.CoinV1{}
	doubleSpendingTokenInput.SetBytes(tokenOutputBytesHacked)
	pc,_ = doubleSpendingTokenInput.Decrypt(keySets[0])
	pr,_ = getParamForTxV1TokenTransfer(tokenTx, db, t)
	pr.TokenParams.TokenInput = []privacy.PlainCoin{pc}

	tx = &TxToken{}
	err = tx.Init(pr)
	assert.Equal(t,nil,err)
	isValidSanity, err = tx.ValidateSanityData(nil, nil, nil, 0)
	assert.Equal(t, true, isValidSanity)
	assert.Equal(t, nil, err)
	isValidTxItself, err = tx.ValidateTxByItself(hasPrivacyForPRV, db, nil, nil, shardID, false, nil, nil)
	assert.Equal(t, true, isValidTxItself)
	assert.Equal(t, nil, err)
	err = tx.ValidateTxWithBlockChain(nil, nil ,nil, 0, db)
	fmt.Printf("Double spent input (token) -> %v\n",err)
	assert.NotEqual(t,nil,err)
	if err==nil{
		fmt.Println(err)
		panic("Test Terminated Early : Double Spent")
	}
}

func testTxTokenV1DuplicateInput(txTokenInit *TxToken, db *statedb.StateDB, t *testing.T){
	transferAmount := uint64(696)
	msgCipherText := []byte("doing a double-spend")

	feeOutputs := txTokenInit.GetTxBase().GetProof().GetOutputCoins()
	dup := &privacy.CoinV1{}
	dup.SetBytes(feeOutputs[0].Bytes())
	tokenOutputs := []privacy.Coin{dup}
	prvCoinsToPayTransfer := make([]privacy.PlainCoin,0)
	tokenCoinsToTransfer := make([]privacy.PlainCoin,0)
	var inputAmountFee uint64
	for _,c := range feeOutputs{
		cloneCoin := privacy.CoinV1{}
		cloneCoin.SetBytes(c.Bytes())
		pc,_ := cloneCoin.Decrypt(keySets[0])
		if inputAmountFee==0{
			inputAmountFee = pc.GetValue()
		}
		// s,_ := json.Marshal(pc.(*privacy.CoinV1))
		// fmt.Printf("Tx Fee : %x has received %d in PRV\n",pc.GetPublicKey().ToBytesS(),pc.GetValue())
		prvCoinsToPayTransfer = append(prvCoinsToPayTransfer,pc)
	}
	for _,c := range tokenOutputs{
		cloneCoin := privacy.CoinV1{}
		cloneCoin.SetBytes(c.Bytes())
		pc,err := cloneCoin.Decrypt(keySets[0])
		// s,_ := json.Marshal(pc.(*privacy.CoinV1))
		// fmt.Printf("Tx Token : %x has received %d in token T1\n",pc.GetPublicKey().ToBytesS(),pc.GetValue())
		assert.Equal(t,nil,err)
		tokenCoinsToTransfer = append(tokenCoinsToTransfer,pc)
	}

	paymentInfo2 := []*privacy.PaymentInfo{{PaymentAddress: keySets[0].PaymentAddress, Amount: transferAmount, Message: msgCipherText}}
	paymentInfoFee := []*privacy.PaymentInfo{{PaymentAddress: keySets[0].PaymentAddress, Amount: inputAmountFee-140, Message: msgCipherText}}
	// // token param for transfer token
	tokenParam2 := &tx_generic.TokenParam{
		PropertyID:     "0000000000000000000000000000000000000000000000000000000000000004",
		PropertyName:   "Token 1",
		PropertySymbol: "PRV",
		Amount:         transferAmount,
		TokenTxType:    utils.CustomTokenTransfer,
		Receiver:       paymentInfo2,
		TokenInput:     tokenCoinsToTransfer,
		Mintable:       false,
		Fee:            0,
	}

	existed := statedb.PrivacyTokenIDExisted(db, common.PRVCoinID)
	if !existed{
		errStore := statedb.StorePrivacyToken(db, common.PRVCoinID, tokenParam2.PropertyName, tokenParam2.PropertySymbol, statedb.InitToken, tokenParam2.Mintable, tokenParam2.Amount, []byte{}, *txTokenInit.Hash())
		assert.Equal(t,nil,errStore)
	}
	malParams := tx_generic.NewTxTokenParams(&keySets[0].PrivateKey,
		paymentInfoFee, prvCoinsToPayTransfer, 140, tokenParam2, db, nil,
		hasPrivacyForPRV, hasPrivacyForToken, shardID, []byte{},db)

	malTx := &TxToken{}
	errMalInit := malTx.Init(malParams)
	assert.Equal(t,nil,errMalInit)
	// sanity should be fine
	isSane,err := malTx.ValidateSanityData(nil,nil,nil,0)
	if isSane{
		fmt.Println("Passed Sanity Test")
		panic("Test Terminated Early")
	}
	_ = err
	fmt.Printf("Token-fee double spend -> %v\n",err)
	assert.Equal(t,false,isSane)
	// validate should reject due to Verify() in PaymentProofV1
	// isValid,_ = malTx.ValidateTxByItself(true, db, nil, nil, byte(0), true, nil, nil)
	// assert.Equal(t,false,isValid)
}

func getParamsForTxV1TokenInit(theInputCoin privacy.PlainCoin, db *statedb.StateDB) (*tx_generic.TxTokenParams,*tx_generic.TokenParam){
	msgCipherText := []byte("haha dummy ciphertext")
	initAmount := uint64(10000)
	tokenPayments := []*privacy.PaymentInfo{{PaymentAddress: keySets[0].PaymentAddress, Amount: initAmount, Message: msgCipherText}}

	myOnlyInputCoin := theInputCoin
	inputCoinsPRV := []privacy.PlainCoin{myOnlyInputCoin}
	paymentInfoPRV := []*privacy.PaymentInfo{ &privacy.PaymentInfo{PaymentAddress: keySets[0].PaymentAddress, Amount: uint64(990), Message: []byte("test out")}}

	// token param for init new token
	tokenParam := &tx_generic.TokenParam{
		PropertyID:     "",
		PropertyName:   "Token 1",
		PropertySymbol: "T1",
		Amount:         initAmount,
		TokenTxType:    utils.CustomTokenInit,
		Receiver:       tokenPayments,
		TokenInput:     []privacy.PlainCoin{},
		Mintable:       false,
		Fee:            0,
	}

	paramToCreateTx := tx_generic.NewTxTokenParams(&keySets[0].PrivateKey,
		paymentInfoPRV, inputCoinsPRV, 10, tokenParam, db, nil,
		hasPrivacyForPRV, hasPrivacyForToken, shardID, []byte{},db)
	return paramToCreateTx, tokenParam
}

func getParamForTxV1TokenTransfer(txTokenInit *TxToken, db *statedb.StateDB, t *testing.T) (*tx_generic.TxTokenParams,*tx_generic.TokenParam){
	transferAmount := uint64(69)
	msgCipherText := []byte("doing a transfer")

	feeOutputs := txTokenInit.GetTxBase().GetProof().GetOutputCoins()
	prvCoinsToPayTransfer := make([]privacy.PlainCoin,0)
	tokenOutputs := txTokenInit.GetTxTokenData().TxNormal.GetProof().GetOutputCoins()
	tokenCoinsToTransfer := make([]privacy.PlainCoin,0)
	var inputAmountFee uint64
	for _,c := range feeOutputs{
		cloneCoin := privacy.CoinV1{}
		cloneCoin.SetBytes(c.Bytes())
		pc,_ := cloneCoin.Decrypt(keySets[0])
		if inputAmountFee==0{
			inputAmountFee = pc.GetValue()
		}
		// s,_ := json.Marshal(pc.(*privacy.CoinV1))
		// fmt.Printf("Tx Fee : %x has received %d in PRV\n",pc.GetPublicKey().ToBytesS(),pc.GetValue())
		prvCoinsToPayTransfer = append(prvCoinsToPayTransfer,pc)
	}
	for _,c := range tokenOutputs{
		cloneCoin := privacy.CoinV1{}
		cloneCoin.SetBytes(c.Bytes())
		pc,err := cloneCoin.Decrypt(keySets[0])
		// s,_ := json.Marshal(pc.(*privacy.CoinV1))
		// fmt.Printf("Tx Token : %x has received %d in token T1\n",pc.GetPublicKey().ToBytesS(),pc.GetValue())
		assert.Equal(t,nil,err)
		tokenCoinsToTransfer = append(tokenCoinsToTransfer,pc)
	}

	paymentInfo2 := []*privacy.PaymentInfo{{PaymentAddress: keySets[0].PaymentAddress, Amount: transferAmount, Message: msgCipherText}}
	paymentInfoFee := []*privacy.PaymentInfo{{PaymentAddress: keySets[0].PaymentAddress, Amount: inputAmountFee-140, Message: msgCipherText}}
	// // token param for transfer token
	tokenParam2 := &tx_generic.TokenParam{
		PropertyID:     txTokenInit.GetTokenID().String(),
		PropertyName:   "Token 1",
		PropertySymbol: "T1",
		Amount:         transferAmount,
		TokenTxType:    utils.CustomTokenTransfer,
		Receiver:       paymentInfo2,
		TokenInput:     tokenCoinsToTransfer,
		Mintable:       false,
		Fee:            0,
	}

	paramToCreateTx2 := tx_generic.NewTxTokenParams(&keySets[0].PrivateKey,
		paymentInfoFee, prvCoinsToPayTransfer, 140, tokenParam2, db, nil,
		hasPrivacyForPRV, hasPrivacyForToken, shardID, []byte{},db)
	return  paramToCreateTx2, tokenParam2
}

func forceSaveCoinsV1(db *statedb.StateDB, coinsToBeSaved []privacy.Coin, shardID byte, tokenID common.Hash, t *testing.T){
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