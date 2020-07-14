package transaction

import (
	"bytes"
	// "math/big"
	"testing"
	"fmt"
	// "io/ioutil"
	// "os"
	// "encoding/json"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/privacy/key"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/privacy/coin"
	// "github.com/incognitochain/incognito-chain/trie"
	// "github.com/incognitochain/incognito-chain/incdb"
	// "github.com/incognitochain/incognito-chain/incognitokey"
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
	hasPrivacyForPRV bool = true
	hasPrivacyForToken bool = true
	shardID byte = byte(0)
)

func TestInitTxPrivacyToken(t *testing.T) {
	for loop := 0; loop < numOfLoops; loop++ {
		fmt.Printf("\n------------------TxTokenVersion2 Main Test\n")
		var err error
		numOfPrivateKeys := 4
		numOfInputs := 2
		dummyDB, _ := statedb.NewWithPrefixTrie(emptyRoot, warperDBStatedbTest)
		preparePaymentKeys(numOfPrivateKeys,t)
		
		pastCoins := make([]coin.Coin, (10+numOfInputs)*len(dummyPrivateKeys))
		for i, _ := range pastCoins {
			tempCoin,err := coin.NewCoinFromPaymentInfo(paymentInfo[i%len(dummyPrivateKeys)])
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
		paramToCreateTx,tokenParam := getParamsForTxTokenInit(pastCoins[0], dummyDB)
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
		assert.Equal(t,true,bytes.Equal([]byte(inf),msgCipherText))
		retrievedFee := tx.GetTxFee()
		assert.Equal(t, uint64(1000),retrievedFee)
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

		testTxTokenV2DeletedProof(tx, dummyDB, t)

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
		exists := statedb.PrivacyTokenIDExisted(dummyDB,*tx.GetTokenID())
		assert.Equal(t, false, exists)
		statedb.StorePrivacyToken(dummyDB, *tx.GetTokenID(), tokenParam.PropertyName, tokenParam.PropertySymbol, statedb.InitToken, tokenParam.Mintable, tokenParam.Amount, []byte{}, *tx.Hash())

		statedb.StoreCommitments(dummyDB,*tx.GetTokenID(), [][]byte{tokenOutputs[0].GetCommitment().ToBytesS()}, shardID)
		// check it exists
		exists = statedb.PrivacyTokenIDExisted(dummyDB,*tx.GetTokenID())
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
		
		testTxTokenV2DeletedProof(tx2, dummyDB, t)
		testTxTokenV2InvalidFee(dummyDB, paramToCreateTx2, t)
		testTxTokenV2OneFakeOutput(tx2, dummyDB, paramToCreateTx2, t)
		testTxTokenV2OneDoubleSpentInput(tx2, dummyDB, feeOutputBytesHacked, tokenOutputBytesHacked, t)

		testTxTokenV2Salary(tx.GetTokenID(), dummyDB, t)
	}
}

func testTxTokenV2DeletedProof(txv2 *TxTokenVersion2, db *statedb.StateDB, t *testing.T){
	// try setting the proof to nil, then verify
	// it should not go through
	savedProof := txv2.GetTxTokenData().TxNormal.GetProof()
	txv2.GetTxTokenData().TxNormal.SetProof(nil)
	isValid,_ := txv2.ValidateSanityData(nil,nil,nil,0)
	assert.Equal(t,true,isValid)
	isValidTxItself, _ := txv2.ValidateTxByItself(hasPrivacyForPRV, db, nil, nil, shardID, false, nil, nil)
	assert.Equal(t,false,isValidTxItself)
	txv2.GetTxTokenData().TxNormal.SetProof(savedProof)

	savedProof = txv2.GetTxBase().GetProof()
	txv2.GetTxBase().SetProof(nil)
	isValid,_ = txv2.ValidateSanityData(nil,nil,nil,0)
	assert.Equal(t,true,isValid)
	isValidTxItself, _ = txv2.ValidateTxByItself(hasPrivacyForPRV, db, nil, nil, shardID, false, nil, nil)
	assert.Equal(t,false,isValidTxItself)
	txv2.GetTxBase().SetProof(savedProof)
}

func testTxTokenV2InvalidFee(db *statedb.StateDB, params *TxTokenParams, t *testing.T){
	// a set of init params where sum(Input) < fee + sum(Output)
	// let's say someone tried to use this invalid fee for tx
	// we should encounter an error here

	// set fee to 150k
	malInputParams := params
	malInputParams.feeNativeCoin *= 10000
	malTx := &TxTokenVersion2{}
	errMalInit := malTx.Init(malInputParams)
	// shall not pass
	assert.NotEqual(t,nil,errMalInit)
	malInputParams.feeNativeCoin /= 10000
}

func testTxTokenV2OneFakeOutput(txv2 *TxTokenVersion2, db *statedb.StateDB, params *TxTokenParams, t *testing.T){
	// similar to the above. All these verifications should fail
		savedPay := params.tokenParams.Receiver
		params.tokenParams.Receiver = []*key.PaymentInfo{key.InitPaymentInfo(keySets[1].PaymentAddress,uint64(690),[]byte("haha dummy ciphertext"))}
		malInputParams := params
		malTx := &TxTokenVersion2{}
		err := malTx.Init(malInputParams)
		assert.Equal(t,nil,err)
		// fmt.Println(err)
		malTx.GetTxTokenData().TxNormal.SetProof(txv2.GetProof())
		isValid,err := malTx.ValidateTxByItself(true,db,nil,nil,0,false,nil,nil)
		// verify must fail
		assert.Equal(t,false,isValid)
		params.tokenParams.Receiver = savedPay
}

// happens after txTransfer in test
// we create a second transfer, then try to reuse fee input / token input
func testTxTokenV2OneDoubleSpentInput(tokenTx *TxTokenVersion2, db *statedb.StateDB, feeOutputBytesHacked, tokenOutputBytesHacked []byte, t *testing.T){
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
	pr,_ := getParamForTxTokenTransfer(tokenTx, db, t)
	tx := &TxTokenVersion2{}
	err := tx.Init(pr)
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
	doubleSpendingFeeInput := &coin.CoinV2{}
	doubleSpendingFeeInput.SetBytes(feeOutputBytesHacked)
	pc,_ := doubleSpendingFeeInput.Decrypt(keySets[0])
	pr.inputCoin = []coin.PlainCoin{pc}
	tx = &TxTokenVersion2{}
	err = tx.Init(pr)
	assert.Equal(t,nil,err)
	isValidSanity, err = tx.ValidateSanityData(nil, nil, nil, 0)
	assert.Equal(t, true, isValidSanity)
	assert.Equal(t, nil, err)
	isValidTxItself, err = tx.ValidateTxByItself(hasPrivacyForPRV, db, nil, nil, shardID, false, nil, nil)
	assert.Equal(t, true, isValidTxItself)
	assert.Equal(t, nil, err)
	err = tx.ValidateTxWithBlockChain(nil, nil ,nil, 0, db)
	// fmt.Println(err)
	assert.NotEqual(t,nil,err)

	// now we try to swap in a used token input
	doubleSpendingTokenInput := &coin.CoinV2{}
	doubleSpendingTokenInput.SetBytes(tokenOutputBytesHacked)
	pc,_ = doubleSpendingTokenInput.Decrypt(keySets[0])
	pr.tokenParams.TokenInput = []coin.PlainCoin{pc}
	tx = &TxTokenVersion2{}
	err = tx.Init(pr)
	assert.Equal(t,nil,err)
	isValidSanity, err = tx.ValidateSanityData(nil, nil, nil, 0)
	assert.Equal(t, true, isValidSanity)
	assert.Equal(t, nil, err)
	isValidTxItself, err = tx.ValidateTxByItself(hasPrivacyForPRV, db, nil, nil, shardID, false, nil, nil)
	assert.Equal(t, true, isValidTxItself)
	assert.Equal(t, nil, err)
	err = tx.ValidateTxWithBlockChain(nil, nil ,nil, 0, db)
	// fmt.Println(err)
	assert.NotEqual(t,nil,err)
}

func getParamForTxTokenTransfer(txTokenInit *TxTokenVersion2, db *statedb.StateDB, t *testing.T) (*TxTokenParams,*TokenParam){
	transferAmount := uint64(69)
	msgCipherText := []byte("doing a transfer")
	paymentInfo2 := []*privacy.PaymentInfo{{PaymentAddress: keySets[0].PaymentAddress, Amount: transferAmount, Message: msgCipherText}}

	feeOutputs := txTokenInit.GetTxBase().GetProof().GetOutputCoins()
	prvCoinsToPayTransfer := make([]coin.PlainCoin,0)
	tokenOutputs := txTokenInit.GetTxTokenData().TxNormal.GetProof().GetOutputCoins()
	tokenCoinsToTransfer := make([]coin.PlainCoin,0)
	for _,c := range feeOutputs{
		pc,_ := c.Decrypt(keySets[0])
		// s,_ := json.Marshal(pc.(*coin.CoinV2))
		fmt.Printf("Tx Fee : %x has received %d in PRV\n",pc.GetPublicKey().ToBytesS(),pc.GetValue())
		prvCoinsToPayTransfer = append(prvCoinsToPayTransfer,pc)
	}
	for _,c := range tokenOutputs{
		pc,err := c.Decrypt(keySets[0])
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
		[]*key.PaymentInfo{}, prvCoinsToPayTransfer, 15, tokenParam2, db, nil,
		hasPrivacyForPRV, hasPrivacyForToken, shardID, []byte{},db)
	return  paramToCreateTx2, tokenParam2
}

func getParamsForTxTokenInit(theInputCoin coin.Coin, db *statedb.StateDB) (*TxTokenParams,*TokenParam){
	msgCipherText := []byte("haha dummy ciphertext")
	initAmount := uint64(10000)
	tokenPayments := []*privacy.PaymentInfo{{PaymentAddress: keySets[0].PaymentAddress, Amount: initAmount, Message: msgCipherText}}

	myOnlyInputCoin,_ := theInputCoin.Decrypt(keySets[0])
	inputCoinsPRV := []coin.PlainCoin{myOnlyInputCoin}
	paymentInfoPRV := []*privacy.PaymentInfo{key.InitPaymentInfo(keySets[0].PaymentAddress,uint64(15000),[]byte("test out"))}

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
		hasPrivacyForPRV, hasPrivacyForToken, shardID, []byte{},db)
	return paramToCreateTx, tokenParam
}

func testTxTokenV2Salary(tokenID *common.Hash, db *statedb.StateDB, t *testing.T){
	numOfPrivateKeys := 2
	for loop := 0; loop < numOfLoops; loop++ {
		fmt.Printf("\n------------------TxTokenVersion2 Salary Test\n")
		var err error
		preparePaymentKeys(numOfPrivateKeys,t)

		// create 2 otaCoins, the second one will already be stored in the db
		theCoins := make([]*coin.CoinV2, 2)
		theCoinsGeneric := make([]coin.Coin,2)
		for i, _ := range theCoins {
			var tempCoin *coin.CoinV2
			var err error
			for{
				tempCoin,err = coin.NewCoinFromPaymentInfo(paymentInfo[i])
				otaPublicKeyBytes := tempCoin.GetPublicKey().ToBytesS()
				// want an OTA in shard 0
				if otaPublicKeyBytes[31]==0{
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
		txsal := TxTokenVersion2{}
		// actually making the salary TX
		err = txsal.InitTxTokenSalary(theCoins[0], dummyPrivateKeys[0], db, nil, tokenID, "Token 1")
		assert.Equal(t, nil, err)
		// isValidSanity, err := txsal.ValidateSanityData(nil, nil, nil, 0)
		// assert.Equal(t, true, isValidSanity)
		// assert.Equal(t, nil, err)

		// verify function for txTokenV2Salary is out of scope, so we exit here
	}
}