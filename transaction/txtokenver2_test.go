package transaction

import (
	"bytes"
	"fmt"
	// "math/big"
	"testing"
	// "io/ioutil"
	// "os"
	"encoding/json"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/privacy/coin"
	"github.com/incognitochain/incognito-chain/privacy/key"
	"github.com/incognitochain/incognito-chain/privacy/operation"
	// "github.com/incognitochain/incognito-chain/trie"
	// "github.com/incognitochain/incognito-chain/incdb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/stretchr/testify/assert"
)

// var _ = func() (_ struct{}) {
// // initialize a `test` db in the OS's tempdir
// // and with it, a db access wrapper that reads/writes our transactions
// 	fmt.Println("This runs before init()!")
// 	Logger.Init(common.NewBackend(nil).Logger("test", true))
// 	dbPath, err := ioutil.TempDir(os.TempDir(), "test_statedb_")
// 	if err != nil {
// 		panic(err)
// 	}
// 	diskBD, _ := incdb.Open("leveldb", dbPath)
// 	warperDBStatedbTest = statedb.NewDatabaseAccessWarper(diskBD)
// 	trie.Logger.Init(common.NewBackend(nil).Logger("test", true))
// 	return
// }()
var (
	hasPrivacyForPRV   bool = true
	hasPrivacyForToken bool = true
	shardID            byte = byte(0)
)

func TestInitAndTransferTxPrivacyToken(t *testing.T) {
	for loop := 0; loop < numOfLoops; loop++ {
		fmt.Printf("\n------------------TxTokenVersion2 Main Test\n")
		var err error
		numOfPrivateKeys := 4
		numOfInputs := 2
		dummyDB, _ := statedb.NewWithPrefixTrie(emptyRoot, warperDBStatedbTest)
		preparePaymentKeys(numOfPrivateKeys, t)

		pastCoins := make([]coin.Coin, (10+numOfInputs)*len(dummyPrivateKeys))
		for i, _ := range pastCoins {
			tempCoin, err := coin.NewCoinFromPaymentInfo(paymentInfo[i%len(dummyPrivateKeys)])
			assert.Equal(t, nil, err)
			assert.Equal(t, false, tempCoin.IsEncrypted())
			tempCoin.ConcealOutputCoin(keySets[i%len(dummyPrivateKeys)].PaymentAddress.GetPublicView())
			assert.Equal(t, true, tempCoin.IsEncrypted())
			assert.Equal(t, true, tempCoin.GetSharedRandom() == nil)
			pastCoins[i] = tempCoin
		}

		// store a bunch of sample OTA coins in PRV
		forceSaveCoins(dummyDB, pastCoins, 0, common.PRVCoinID, t)

		// sample message to receiver
		msgCipherText := []byte("haha dummy ciphertext")
		paramToCreateTx, tokenParam := getParamsForTxTokenInit(pastCoins[0], dummyDB)
		// create tx for token init
		tx := &TxTokenVersion2{}

		fmt.Println("Token Init")
		err = tx.Init(paramToCreateTx)
		assert.Equal(t, nil, err)

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
		assert.Equal(t, true, bytes.Equal([]byte(inf), msgCipherText))
		retrievedFee := tx.GetTxFee()
		assert.Equal(t, uint64(1000), retrievedFee)
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

		testTxTokenV2JsonMarshaler(tx, 25, dummyDB, t)
		// someInvalidTxs := getCorruptedJsonDeserializedTokenTxs(tx, t)
		// for _,theInvalidTx := range someInvalidTxs{
		// 	txSpecific, ok := theInvalidTx.(*TxTokenVersion2)
		// 	assert.Equal(t, true, ok)
		// 	// look for potential panics by calling verify
		// 	isSane, _ := txSpecific.ValidateSanityData(nil, nil, nil, 0)
		// 	// if it doesnt pass sanity then the next validation could panic, it's ok by spec
		// 	if !isSane{
		// 		continue
		// 	}
		// 	txSpecific.ValidateTxByItself(hasPrivacyForPRV, dummyDB, nil, nil, shardID, false, nil, nil)
		// 	txSpecific.ValidateTxWithBlockChain(nil, nil, nil, shardID, dummyDB)
		// }
		// // focused on the above right now
		// continue 

		// testTxTokenV2DeletedProof(tx, dummyDB, t)
		// testTxTokenV2OneFakeOutput(tx, dummyDB, paramToCreateTx, t)

		// save the fee outputs into the db
		// get output coin token from tx
		tokenOutputs := tx.GetTxTokenData().TxNormal.GetProof().GetOutputCoins()
		feeOutputs := tx.GetTxBase().GetProof().GetOutputCoins()
		forceSaveCoins(dummyDB, feeOutputs, 0, common.PRVCoinID, t)

		feeOutputBytesHacked := feeOutputs[0].Bytes()
		tokenOutputBytesHacked := tokenOutputs[0].Bytes()

		// tx token transfer
		paramToCreateTx2, tokenParam2 := getParamForTxTokenTransfer(tx, dummyDB, t)
		_ = tokenParam2
		tx2 := &TxTokenVersion2{}

		fmt.Println("Token Transfer")
		err = tx2.Init(paramToCreateTx2)
		// should fail because db does not have this token yet
		assert.NotEqual(t, nil, err)
		// store the token
		exists := statedb.PrivacyTokenIDExisted(dummyDB, *tx.GetTokenID())
		assert.Equal(t, false, exists)
		statedb.StorePrivacyToken(dummyDB, *tx.GetTokenID(), tokenParam.PropertyName, tokenParam.PropertySymbol, statedb.InitToken, tokenParam.Mintable, tokenParam.Amount, []byte{}, *tx.Hash())

		statedb.StoreCommitments(dummyDB, *tx.GetTokenID(), [][]byte{tokenOutputs[0].GetCommitment().ToBytesS()}, shardID)
		// check it exists
		exists = statedb.PrivacyTokenIDExisted(dummyDB, *tx.GetTokenID())
		assert.Equal(t, true, exists)
		err = tx2.Init(paramToCreateTx2)
		// still fails because the token's `init` coin (10000 T1) is not stored yet
		assert.NotEqual(t, nil, err)
		// add the coin. Tx creation shouldsucceed  now
		forceSaveCoins(dummyDB, tokenOutputs, 0, *tx.GetTokenID(), t)
		err = tx2.Init(paramToCreateTx2)
		assert.Equal(t, nil, err)

		msgCipherText = []byte("doing a transfer")
		assert.Equal(t, true, bytes.Equal(msgCipherText, tx2.GetTxTokenData().TxNormal.GetProof().GetOutputCoins()[0].GetInfo()))

		isValidSanity, err = tx2.ValidateSanityData(nil, nil, nil, 0)
		assert.Equal(t, true, isValidSanity)
		assert.Equal(t, nil, err)

		// before the token init tx is written into db, this should not pass
		isValidTxItself, err = tx2.ValidateTxByItself(hasPrivacyForPRV, dummyDB, nil, nil, shardID, false, nil, nil)
		assert.Equal(t, true, isValidTxItself)
		assert.Equal(t, nil, err)

		err = tx2.ValidateTxWithBlockChain(nil, nil, nil, shardID, dummyDB)
		assert.Equal(t, nil, err)

		//someInvalidTxs := getCorruptedJsonDeserializedTokenTxs(tx2, t)
		//for _,theInvalidTx := range someInvalidTxs{
		//	txSpecific, ok := theInvalidTx.(*TxTokenVersion2)
		//	assert.Equal(t, true, ok)
		//	// look for potential panics by calling verify
		//	isSane, _ := txSpecific.ValidateSanityData(nil, nil, nil, 0)
		//	// if it doesnt pass sanity then the next validation could panic, it's ok by spec
		//	if !isSane{
		//		continue
		//	}
		//	txSpecific.ValidateTxByItself(hasPrivacyForPRV, dummyDB, nil, nil, shardID, false, nil, nil)
		//	txSpecific.ValidateTxWithBlockChain(nil, nil, nil, shardID, dummyDB)
		//}
		testTxTokenV2JsonMarshaler(tx2, 25, dummyDB, t)

		testTxTokenV2DeletedProof(tx2, dummyDB, t)
		testTxTokenV2InvalidFee(tx2, dummyDB, t)
		testTxTokenV2OneFakeOutput(tx2, dummyDB, paramToCreateTx2, t)
		testTxTokenV2OneDoubleSpentInput(tx2, dummyDB, feeOutputBytesHacked, tokenOutputBytesHacked, t)

		testTxTokenV2Salary(tx.GetTokenID(), dummyDB, t)
		testTxTokenV2TransferPRV(dummyDB, t)
	}
}

func testTxTokenV2TransferPRV(db *statedb.StateDB, t *testing.T) {
	pastCoins, err := createAndSaveTokens(100, common.PRVCoinID, keySets, db, 2)
	if err != nil {
		panic(err)
	}

	//Store token onto the database
	err = statedb.StorePrivacyToken(db, common.PRVCoinID, common.PRVCoinName, common.PRVCoinName, statedb.InitToken, true, uint64(10000000000000000), nil, common.Hash{101})
	if err != nil {
		panic(err)
	}
	res := txDatabaseWrapper.privacyTokenIDExisted(db, common.PRVCoinID)
	assert.Equal(t, true, res)


	theInputCoin := pastCoins[:3]

	paramToCreateTx, _, err := createTokenTransferParams(theInputCoin, db, common.PRVCoinID.String(), common.PRVCoinName, common.PRVCoinName, keySets[0])
	if err != nil {
		panic(err)
	}
	paramToCreateTx.tokenParams.TokenTxType = CustomTokenTransfer

	tx := &TxTokenVersion2{}
	err = tx.Init(paramToCreateTx)
	if err != nil {
		panic(err)
	}

	res, err = tx.ValidateSanityData(nil, nil, nil, 0)
	assert.Equal(t, false, res)
}

func testTxTokenV2DeletedProof(txv2 *TxTokenVersion2, db *statedb.StateDB, t *testing.T) {
	// try setting the proof to nil, then verify
	// it should not go through
	savedProof := txv2.GetTxTokenData().TxNormal.GetProof()
	txv2.GetTxTokenData().TxNormal.SetProof(nil)
	isValid, _ := txv2.ValidateSanityData(nil, nil, nil, 0)
	assert.Equal(t, true, isValid)
	isValidTxItself, _ := txv2.ValidateTxByItself(hasPrivacyForPRV, db, nil, nil, shardID, false, nil, nil)
	assert.Equal(t, false, isValidTxItself)
	txv2.GetTxTokenData().TxNormal.SetProof(savedProof)

	savedProof = txv2.GetTxBase().GetProof()
	txv2.GetTxBase().SetProof(nil)
	isValid, _ = txv2.ValidateSanityData(nil, nil, nil, 0)
	assert.Equal(t, true, isValid)
	isValidTxItself, _ = txv2.ValidateTxByItself(hasPrivacyForPRV, db, nil, nil, shardID, false, nil, nil)
	assert.Equal(t, false, isValidTxItself)
	// undo the tampering
	txv2.GetTxBase().SetProof(savedProof)
}

func testTxTokenV2InvalidFee(txv2 *TxTokenVersion2, db *statedb.StateDB, t *testing.T) {
	// a set of init params where fee is changed so mlsag should verify to false
	// let's say someone tried to use this invalid fee for tx
	// we should encounter an error here

	// set fee to increase by 1000PRV
	savedFee := txv2.GetTxBase().GetTxFee()
	txv2.GetTxBase().SetTxFee(savedFee + 1000)

	// sanity should pass
	isValidSanity, err := txv2.ValidateSanityData(nil, nil, nil, 0)
	assert.Equal(t, true, isValidSanity)
	assert.Equal(t, nil, err)

	// should reject at signature since fee & output doesn't sum to input
	isValidTxItself, err := txv2.ValidateTxByItself(hasPrivacyForPRV, db, nil, nil, shardID, false, nil, nil)
	assert.Equal(t, false, isValidTxItself)
	// fmt.Printf("Invalid fee -> %v\n",err)

	// undo the tampering
	txv2.GetTxBase().SetTxFee(savedFee)
	isValidTxItself, _ = txv2.ValidateTxByItself(hasPrivacyForPRV, db, nil, nil, shardID, false, nil, nil)
	assert.Equal(t, true, isValidTxItself)
}

func testTxTokenV2OneFakeOutput(txv2 *TxTokenVersion2, db *statedb.StateDB, params *TxTokenParams, t *testing.T) {
	// tid,err := common.Hash{}.NewHashFromStr(common.PRVCoinID.String())
	// txPrivacyParams := NewTxPrivacyInitParams(
	// 	params.senderKey, params.paymentInfo, params.inputCoin, params.feeNativeCoin,
	// 	true, db, tid, params.metaData, params.info,
	// )

	// similar to the above. All these verifications should fail

	var err error
	outs := txv2.GetTxTokenData().TxNormal.GetProof().GetOutputCoins()
	tokenOutput, ok := outs[0].(*coin.CoinV2)
	savedCoinBytes := tokenOutput.Bytes()
	assert.Equal(t, true, ok)
	tokenOutput.Decrypt(keySets[0])
	// set amount from 69 to 690
	tokenOutput.SetValue(690)
	tokenOutput.SetSharedRandom(operation.RandomScalar())
	tokenOutput.ConcealOutputCoin(keySets[0].PaymentAddress.GetPublicView())
	err = resignUnprovenTxToken([]*incognitokey.KeySet{keySets[0]}, txv2, params, nil)
	assert.Equal(t, nil, err)
	isValid, err := txv2.ValidateTxByItself(true, db, nil, nil, 0, false, nil, nil)
	// verify must fail
	assert.Equal(t, false, isValid)
	// fmt.Printf("Fake output (wrong amount) -> %+v\n",err)
	// undo the tampering
	tokenOutput.SetBytes(savedCoinBytes)
	outs[0] = tokenOutput
	txv2.GetTxTokenData().TxNormal.GetProof().SetOutputCoins(outs)
	err = resignUnprovenTxToken([]*incognitokey.KeySet{keySets[0]}, txv2, params, nil)
	assert.Equal(t, nil, err)
	isValid, err = txv2.ValidateTxByItself(true, db, nil, nil, 0, false, nil, nil)
	assert.Equal(t, true, isValid)

	// now instead of changing amount, we change the OTA public key
	theProof := txv2.GetTxTokenData().TxNormal.GetProof()
	outs = theProof.GetOutputCoins()
	tokenOutput, ok = outs[0].(*coin.CoinV2)
	savedCoinBytes = tokenOutput.Bytes()
	assert.Equal(t, true, ok)
	payInf := &privacy.PaymentInfo{PaymentAddress: keySets[0].PaymentAddress, Amount: uint64(69), Message: []byte("doing a transfer")}
	// totally fresh OTA of the same amount, meant for the same PaymentAddress
	newCoin, err := coin.NewCoinFromPaymentInfo(payInf)
	assert.Equal(t, nil, err)
	newCoin.ConcealOutputCoin(keySets[0].PaymentAddress.GetPublicView())
	theProofSpecific, ok := theProof.(*privacy.ProofV2)
	theBulletProof, ok := theProofSpecific.GetAggregatedRangeProof().(*privacy.AggregatedRangeProofV2)
	cmsv := theBulletProof.GetCommitments()
	cmsv[0] = newCoin.GetCommitment()
	outs[0] = newCoin
	txv2.GetTxTokenData().TxNormal.GetProof().SetOutputCoins(outs)
	err = resignUnprovenTxToken([]*incognitokey.KeySet{keySets[0]}, txv2, params, nil)
	assert.Equal(t, nil, err)
	isValid, err = txv2.ValidateTxByItself(true, db, nil, nil, 0, false, nil, nil)
	// verify must fail
	assert.Equal(t, false, isValid)
	// fmt.Printf("Fake output (wrong receiving OTA) -> %+v\n",err)
	// undo the tampering
	tokenOutput.SetBytes(savedCoinBytes)
	outs[0] = tokenOutput
	cmsv[0] = tokenOutput.GetCommitment()
	txv2.GetTxTokenData().TxNormal.GetProof().SetOutputCoins(outs)
	err = resignUnprovenTxToken([]*incognitokey.KeySet{keySets[0]}, txv2, params, nil)
	assert.Equal(t, nil, err)
	isValid, err = txv2.ValidateTxByItself(true, db, nil, nil, 0, false, nil, nil)
	assert.Equal(t, true, isValid)
}

// happens after txTransfer in test
// we create a second transfer, then try to reuse fee input / token input
func testTxTokenV2OneDoubleSpentInput(tokenTx *TxTokenVersion2, db *statedb.StateDB, feeOutputBytesHacked, tokenOutputBytesHacked []byte, t *testing.T) {
	// save both fee&token outputs from previous tx
	otaBytes := [][]byte{tokenTx.GetTxTokenData().TxNormal.GetProof().GetInputCoins()[0].GetKeyImage().ToBytesS()}
	statedb.StoreSerialNumbers(db, *tokenTx.GetTokenID(), otaBytes, 0)
	otaBytes = [][]byte{tokenTx.GetTxBase().GetProof().GetInputCoins()[0].GetKeyImage().ToBytesS()}
	statedb.StoreSerialNumbers(db, common.PRVCoinID, otaBytes, 0)

	tokenOutputs := tokenTx.GetTxTokenData().TxNormal.GetProof().GetOutputCoins()
	feeOutputs := tokenTx.GetTxBase().GetProof().GetOutputCoins()
	forceSaveCoins(db, feeOutputs, 0, common.PRVCoinID, t)
	forceSaveCoins(db, tokenOutputs, 0, *tokenTx.GetTokenID(), t)

	// firstly, using the output coins to create new tx should be successful
	pr, _ := getParamForTxTokenTransfer(tokenTx, db, t)
	tx := &TxTokenVersion2{}
	err := tx.Init(pr)
	assert.Equal(t, nil, err)
	isValidSanity, err := tx.ValidateSanityData(nil, nil, nil, 0)
	assert.Equal(t, true, isValidSanity)
	assert.Equal(t, nil, err)
	isValidTxItself, err := tx.ValidateTxByItself(hasPrivacyForPRV, db, nil, nil, shardID, false, nil, nil)
	assert.Equal(t, true, isValidTxItself)
	assert.Equal(t, nil, err)
	err = tx.ValidateTxWithBlockChain(nil, nil, nil, 0, db)
	assert.Equal(t, nil, err)

	// now we try to swap in a used input for txfee
	doubleSpendingFeeInput := &coin.CoinV2{}
	doubleSpendingFeeInput.SetBytes(feeOutputBytesHacked)
	pc, _ := doubleSpendingFeeInput.Decrypt(keySets[0])
	pr.inputCoin = []coin.PlainCoin{pc}
	tx = &TxTokenVersion2{}
	err = tx.Init(pr)
	assert.Equal(t, nil, err)
	isValidSanity, err = tx.ValidateSanityData(nil, nil, nil, 0)
	assert.Equal(t, true, isValidSanity)
	assert.Equal(t, nil, err)
	isValidTxItself, err = tx.ValidateTxByItself(hasPrivacyForPRV, db, nil, nil, shardID, false, nil, nil)
	assert.Equal(t, true, isValidTxItself)
	assert.Equal(t, nil, err)
	err = tx.ValidateTxWithBlockChain(nil, nil, nil, 0, db)
	// fmt.Println(err)
	assert.NotEqual(t, nil, err)

	// now we try to swap in a used token input
	doubleSpendingTokenInput := &coin.CoinV2{}
	doubleSpendingTokenInput.SetBytes(tokenOutputBytesHacked)
	pc, _ = doubleSpendingTokenInput.Decrypt(keySets[0])
	pr.tokenParams.TokenInput = []coin.PlainCoin{pc}
	tx = &TxTokenVersion2{}
	err = tx.Init(pr)
	assert.Equal(t, nil, err)
	isValidSanity, err = tx.ValidateSanityData(nil, nil, nil, 0)
	assert.Equal(t, true, isValidSanity)
	assert.Equal(t, nil, err)
	isValidTxItself, err = tx.ValidateTxByItself(hasPrivacyForPRV, db, nil, nil, shardID, false, nil, nil)
	assert.Equal(t, true, isValidTxItself)
	assert.Equal(t, nil, err)
	err = tx.ValidateTxWithBlockChain(nil, nil, nil, 0, db)
	// fmt.Println(err)
	assert.NotEqual(t, nil, err)
	if err == nil {
		fmt.Println(err)
		panic("Test Terminated Early : Double Spent")
	}
}

func getParamForTxTokenTransfer(txTokenInit *TxTokenVersion2, db *statedb.StateDB, t *testing.T) (*TxTokenParams, *TokenParam) {
	transferAmount := uint64(69)
	msgCipherText := []byte("doing a transfer")
	paymentInfo2 := []*privacy.PaymentInfo{{PaymentAddress: keySets[0].PaymentAddress, Amount: transferAmount, Message: msgCipherText}}

	feeOutputs := txTokenInit.GetTxBase().GetProof().GetOutputCoins()
	prvCoinsToPayTransfer := make([]coin.PlainCoin, 0)
	tokenOutputs := txTokenInit.GetTxTokenData().TxNormal.GetProof().GetOutputCoins()
	tokenCoinsToTransfer := make([]coin.PlainCoin, 0)
	for _, c := range feeOutputs {
		pc, _ := c.Decrypt(keySets[0])
		// s,_ := json.Marshal(pc.(*coin.CoinV2))
		fmt.Printf("Tx Fee : %x has received %d in PRV\n", pc.GetPublicKey().ToBytesS(), pc.GetValue())
		prvCoinsToPayTransfer = append(prvCoinsToPayTransfer, pc)
	}
	for _, c := range tokenOutputs {
		pc, err := c.Decrypt(keySets[0])
		// s,_ := json.Marshal(pc.(*coin.CoinV2))
		fmt.Printf("Tx Token : %x has received %d in token T1\n", pc.GetPublicKey().ToBytesS(), pc.GetValue())
		assert.Equal(t, nil, err)
		tokenCoinsToTransfer = append(tokenCoinsToTransfer, pc)
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
		[]*key.PaymentInfo{}, prvCoinsToPayTransfer, 15, tokenParam2, db, nil,
		hasPrivacyForPRV, hasPrivacyForToken, shardID, []byte{}, db)
	return paramToCreateTx2, tokenParam2
}

func getParamsForTxTokenInit(theInputCoin coin.Coin, db *statedb.StateDB) (*TxTokenParams, *TokenParam) {
	msgCipherText := []byte("haha dummy ciphertext")
	initAmount := uint64(10000)
	tokenPayments := []*privacy.PaymentInfo{{PaymentAddress: keySets[0].PaymentAddress, Amount: initAmount, Message: msgCipherText}}

	myOnlyInputCoin, _ := theInputCoin.Decrypt(keySets[0])
	inputCoinsPRV := []coin.PlainCoin{myOnlyInputCoin}
	paymentInfoPRV := []*privacy.PaymentInfo{key.InitPaymentInfo(keySets[0].PaymentAddress, uint64(15000), []byte("test out"))}

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
		paymentInfoPRV, inputCoinsPRV, 1000, tokenParam, db, nil,
		hasPrivacyForPRV, hasPrivacyForToken, shardID, []byte{}, db)
	return paramToCreateTx, tokenParam
}

func testTxTokenV2Salary(tokenID *common.Hash, db *statedb.StateDB, t *testing.T) {
	numOfPrivateKeys := 2
	fmt.Printf("\n------------------TxTokenVersion2 Salary Test\n")
	var err error
	preparePaymentKeys(numOfPrivateKeys, t)

	// create 2 otaCoins, the second one will already be stored in the db
	theCoins := make([]*coin.CoinV2, 2)
	theCoinsGeneric := make([]coin.Coin, 2)
	for i, _ := range theCoins {
		var tempCoin *coin.CoinV2
		var err error
		for {
			tempCoin, err = coin.NewCoinFromPaymentInfo(paymentInfo[i])
			otaPublicKeyBytes := tempCoin.GetPublicKey().ToBytesS()
			// want an OTA in shard 0
			if otaPublicKeyBytes[31] == 0 {
				break
			}
		}
		assert.Equal(t, nil, err)
		assert.Equal(t, false, tempCoin.IsEncrypted())
		// tempCoin.ConcealData(keySets[i].PaymentAddress.GetPublicView())
		// assert.Equal(t, true, tempCoin.IsEncrypted())
		// assert.Equal(t, true, tempCoin.GetSharedRandom() == nil)
		// _, err = tempCoin.Decrypt(keySets[i])
		// assert.Equal(t,nil,err)
		theCoins[i] = tempCoin
		theCoinsGeneric[i] = tempCoin
	}
	forceSaveCoins(db, []coin.Coin{theCoinsGeneric[1]}, 0, *tokenID, t)

	// creating the TX object
	txsal := &TxTokenVersion2{}
	// actually making the salary TX
	err = txsal.InitTxTokenSalary(theCoins[0], dummyPrivateKeys[0], db, nil, tokenID, "Token 1")
	if err != nil {
		fmt.Println(err)
		panic("Test Terminated Early")
	}
	assert.Equal(t, nil, err)

	testTxTokenV2JsonMarshaler(txsal, 25, db, t)
	// someInvalidTxs := getCorruptedJsonDeserializedTokenTxs(txsal, t)
	// for _,theInvalidTx := range someInvalidTxs{
	// 	txSpecific, ok := theInvalidTx.(*TxTokenVersion2)
	// 	// assert.Equal(t, true, ok)
	// 	if !ok{
	// 		// it's a txToken but not ver2 for some reason. We ignore for now
	// 		continue
	// 	}
	// 	// look for potential panics by calling verify
	// 	isSane, _ := txSpecific.ValidateSanityData(nil, nil, nil, 0)
	// 	// if it doesnt pass sanity then the next validation could panic, it's ok by spec
	// 	_ = isSane
	// }
	// isValidSanity, err := txsal.ValidateSanityData(nil, nil, nil, 0)
	// assert.Equal(t, true, isValidSanity)
	// assert.Equal(t, nil, err)

	// verify function for txTokenV2Salary is out of scope, so we exit here

}

func resignUnprovenTxToken(decryptingKeys []*incognitokey.KeySet, txToken *TxTokenVersion2, params *TxTokenParams, nonPrivacyParams *TxPrivacyInitParams) error {
	var err error
	txOuter, ok := txToken.Tx.(*TxVersion2)
	if !ok {
		fmt.Println("Test Error : cast")
		return NewTransactionErr(-1000, nil, "Cast failed")
	}
	txToken.Tx = nil
	txToken.cachedHash = nil

	txn := txToken.TxTokenData.TxNormal.(*TxVersion2)
	txn.cachedHash = nil
	if !ok {
		fmt.Println("Test Error : cast")
		return NewTransactionErr(-1000, nil, "Cast failed")
	}

	if nonPrivacyParams == nil {
		propertyID, _ := common.TokenStringToHash(params.tokenParams.PropertyID)
		paramsInner := NewTxPrivacyInitParams(
			params.senderKey,
			params.tokenParams.Receiver,
			params.tokenParams.TokenInput,
			params.tokenParams.Fee,
			params.hasPrivacyToken,
			params.transactionStateDB,
			propertyID,
			nil,
			nil,
		)
		_ = paramsInner
		paramsOuter := NewTxPrivacyInitParams(
			params.senderKey,
			params.paymentInfo,
			params.inputCoin,
			params.feeNativeCoin,
			params.hasPrivacyCoin,
			params.transactionStateDB,
			&common.PRVCoinID,
			params.metaData,
			params.info,
		)
		txInner, ok := txToken.GetTxTokenData().TxNormal.(*TxVersion2)
		if !ok {
			fmt.Println("Test Error : cast inner")
			return NewTransactionErr(-1000, nil, "Cast failed")
		}

		err = resignUnprovenTx(decryptingKeys, txOuter, paramsOuter, &txToken.TxTokenData)
		err = resignUnprovenTx(decryptingKeys, txInner, paramsInner, nil)

		txToken.Tx = txOuter
		return err
	} else {
		paramsOuter := nonPrivacyParams
		err := resignUnprovenTx(decryptingKeys, txOuter, paramsOuter, &txToken.TxTokenData)
		txToken.Tx = txOuter
		return err
	}

	// txTokenDataHash, err := txToken.TxTokenData.Hash()

}

func resignUnprovenTx(decryptingKeys []*incognitokey.KeySet, tx *TxVersion2, params *TxPrivacyInitParams, tokenData *TxTokenData) error {
	tx.cachedHash = nil
	tx.SetSig(nil)
	tx.SetSigPubKey(nil)
	var err error
	outputCoinsGeneric := tx.GetProof().GetOutputCoins()
	var outputCoins []*coin.CoinV2
	// pre-sign, we need unconcealed outputs
	// so receiver privatekeys here are for simulation
	for ind, c := range outputCoinsGeneric {
		var dk *incognitokey.KeySet
		if len(decryptingKeys) == 1 {
			dk = decryptingKeys[0]
		} else {
			dk = decryptingKeys[ind]
		}
		cv2 := &coin.CoinV2{}
		cv2.SetBytes(c.Bytes())
		cv2.Decrypt(dk)
		// fmt.Printf("Value is %d\n",cv2.GetValue())
		cv2.SetKeyImage(nil)
		outputCoins = append(outputCoins, cv2)
	}
	inputCoins := params.inputCoins

	// tx.Proof, err = privacy_v2.Prove(inputCoins, outputCoins, params.hasPrivacy, params.paymentInfo)
	// if err != nil {
	// 	Logger.Log.Errorf("Error in privacy_v2.Prove, error %v ", err)
	// 	return err
	// }	
	message := tx.Hash()[:]
	if tokenData != nil {
		tdh, err := tokenData.Hash()
		if err != nil {
			panic("Hash failed")
		}
		temp := common.HashH(append(message, tdh[:]...))
		message = temp[:]
	}

	err = tx.signOnMessage(inputCoins, outputCoins, params, message[:])
	jsb, _ := json.MarshalIndent(tx, "", "\t")
	_ = jsb
	// fmt.Printf("Rehashed : %s\n => %v, %v\n", string(jsb), tx.Hash(), txTokenDataHash[:])
	return err
}

func createTokenTransferParams(inputCoins []coin.Coin, db *statedb.StateDB, tokenID, tokenName, symbol string, keySet *incognitokey.KeySet) (*TxTokenParams, *TokenParam, error) {
	var err error

	msgCipherText := []byte("Testing Transfer Token")
	transferAmount := uint64(0)
	plainInputCoins := make([]coin.PlainCoin, len(inputCoins))
	for i, inputCoin := range inputCoins {
		plainInputCoins[i], err = inputCoin.Decrypt(keySet)
		if err != nil {
			return nil, nil, err
		}
		if i != 0 {
			transferAmount += plainInputCoins[i].GetValue()
		}
	}

	tokenPayments := []*privacy.PaymentInfo{{PaymentAddress: keySet.PaymentAddress, Amount: transferAmount, Message: msgCipherText}}


	inputCoinsPRV := []coin.PlainCoin{plainInputCoins[0]}
	paymentInfoPRV := []*privacy.PaymentInfo{key.InitPaymentInfo(keySet.PaymentAddress, uint64(10), []byte("test out"))}

	// token param for init new token
	tokenParam := &TokenParam{
		PropertyID:     tokenID,
		PropertyName:   tokenName,
		PropertySymbol: symbol,
		Amount:         transferAmount,
		TokenTxType:    CustomTokenTransfer,
		Receiver:       tokenPayments,
		TokenInput:     plainInputCoins[1:len(inputCoins)],
		Mintable:       false,
		Fee:            0,
	}

	paramToCreateTx := NewTxTokenParams(&keySet.PrivateKey,
		paymentInfoPRV, inputCoinsPRV, 10, tokenParam, db, nil,
		hasPrivacyForPRV, hasPrivacyForToken, shardID, []byte{}, db)
	return paramToCreateTx, tokenParam, nil
}
