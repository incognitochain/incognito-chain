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

func TestInitTxPrivacyToken(t *testing.T) {
	for loop := 0; loop < numOfLoops; loop++ {
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

// to obtain a PlainCoin to feed into input of TX, we need to conceal & decrypt it (it makes sure all fields are right, as opposed to just casting the type to PlainCoin)
			tempCoin.ConcealOutputCoin(keySets[i%len(dummyPrivateKeys)].PaymentAddress.GetPublicView())
			assert.Equal(t, true, tempCoin.IsEncrypted())
			assert.Equal(t, true, tempCoin.GetSharedRandom() == nil)
			pastCoins[i] = tempCoin
		}

// use the db's interface to write our simulated pastCoins to the database
// we do need to re-format the data into bytes first
		forceSaveCoins(dummyDB, pastCoins, 0, common.PRVCoinID,t)

		// message to receiver
		msgCipherText := []byte("haha dummy ciphertext")
		// receiverTK, _ := new(privacy.Point).FromBytesS(keySets[0].PaymentAddress.Tk)

		initAmount := uint64(10000)
		tokenPayments := []*privacy.PaymentInfo{{PaymentAddress: keySets[0].PaymentAddress, Amount: initAmount, Message: msgCipherText}}

		myOnlyInputCoin,err := pastCoins[0].Decrypt(keySets[0])
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

		hasPrivacyForPRV := true
		hasPrivacyForToken := true
		shardID := byte(0)
		paramToCreateTx := NewTxTokenParams(&keySets[0].PrivateKey,
			paymentInfoPRV, inputCoinsPRV, 1000, tokenParam, dummyDB, nil,
			hasPrivacyForPRV, hasPrivacyForToken, shardID, []byte{},nil)

		// init tx
		tx := &TxTokenVersion2{}
		err = tx.Init(paramToCreateTx)
		// fmt.Println(err)
		assert.Equal(t, nil, err)

		// assert.Equal(t, len(msgCipherText), 
		inf := tx.TxTokenData.TxNormal.GetProof().GetOutputCoins()[0].GetInfo()

		assert.Equal(t,true,bytes.Equal([]byte(inf),msgCipherText))
		// convert to JSON string and revert
		txJsonString := tx.JSONString()
		txHash := tx.Hash()
		tx1 := new(TxTokenBase)
		tx1.UnmarshalJSON([]byte(txJsonString))
		txHash1 := tx1.Hash()
		assert.Equal(t, txHash, txHash1)

		// // get actual tx size
		txActualSize := tx.GetTxActualSize()
		assert.Greater(t, txActualSize, uint64(0))
		// txPrivacyTokenActualSize := tx.GetTxPrivacyTokenActualSize()
		// assert.Greater(t, txPrivacyTokenActualSize, uint64(0))
		retrievedFee := tx.GetTxFee()
		assert.Equal(t, uint64(1000),retrievedFee)
		err = tx.ValidateTxWithBlockChain(nil, nil, nil, shardID, dummyDB)
		assert.Equal(t, nil, err)

		isValidSanity, err := tx.ValidateSanityData(nil, nil, nil, 0)
		assert.Equal(t, true, isValidSanity)
		assert.Equal(t, nil, err)
		isValidTxItself, err := tx.ValidateTxByItself(hasPrivacyForPRV, dummyDB, nil, nil, shardID, false, nil, nil)
		assert.Equal(t, true, isValidTxItself)
		assert.Equal(t, nil, err)

		// testTxTokenV2DeletedProof(tx,t)

		theAmount := tx.GetTxTokenData().GetAmount()

		assert.Equal(t, initAmount, theAmount)

		isUniqueReceiver, _, uniqueAmount, tokenID := tx.GetTransferData()
		assert.Equal(t, true, isUniqueReceiver)
		assert.Equal(t, initAmount, uniqueAmount)
		assert.Equal(t, tx.GetTokenID(), tokenID)

		sigPubKey := tx.GetSigPubKey()
		assert.Equal(t, common.SigPubKeySize, len(sigPubKey))

		// save the token-init tx's effects into the db
		// get output coin token from tx
		tokenOutputs := tx.GetTxTokenData().TxNormal.GetProof().GetOutputCoins()
		feeOutputs := tx.GetTxBase().GetProof().GetOutputCoins()
		forceSaveCoins(dummyDB, tokenOutputs, 0, *tx.GetTokenID(), t)
		forceSaveCoins(dummyDB, feeOutputs, 0, common.PRVCoinID, t)
		// err = statedb.StorePrivacyTokenTx(dummyDB,*tx.GetTokenID(), *tx.Hash())
		statedb.StorePrivacyToken(dummyDB, *tx.GetTokenID(), tokenParam.PropertyName, tokenParam.PropertySymbol, statedb.InitToken, tokenParam.Mintable, tokenParam.Amount, []byte{}, *tx.Hash())
		assert.Equal(t,nil,err)
		statedb.StoreCommitments(dummyDB,*tx.GetTokenID(), [][]byte{tokenOutputs[0].GetCommitment().ToBytesS()}, shardID)

		// token check
		exists := statedb.PrivacyTokenIDExisted(dummyDB,*tx.GetTokenID())
		// val,exists := tokenMap[*tx.GetTokenID()]
		fmt.Printf("Token created : %v\n",exists)
		assert.Equal(t, true, exists)

		transferAmount := uint64(69)
		paymentInfo2 := []*privacy.PaymentInfo{{PaymentAddress: keySets[1].PaymentAddress, Amount: transferAmount, Message: msgCipherText}}

		prvCoinsToPayTransfer := make([]coin.PlainCoin,0)
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
			PropertyID:     tx.GetTokenID().String(),
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
			[]*key.PaymentInfo{}, prvCoinsToPayTransfer, 15, tokenParam2, dummyDB, nil,
			hasPrivacyForPRV, hasPrivacyForToken, shardID, []byte{},nil)

		// // init tx
		tx2 := &TxTokenVersion2{}
		err = tx2.Init(paramToCreateTx2)
		// fmt.Println(err)
		assert.Equal(t, nil, err)

		assert.Equal(t, true, bytes.Equal(msgCipherText, tx2.GetTxTokenData().TxNormal.GetProof().GetOutputCoins()[0].GetInfo()))


		err = tx2.ValidateTxWithBlockChain(nil, nil, nil, shardID, dummyDB)
		assert.Equal(t, nil, err)

		isValidSanity, err = tx2.ValidateSanityData(nil, nil, nil, 0)
		assert.Equal(t, true, isValidSanity)
		assert.Equal(t, nil, err)

		isValidTxItself, err = tx2.ValidateTxByItself(hasPrivacyForPRV, dummyDB, nil, nil, shardID, false, nil, nil)
		assert.Equal(t, true, isValidTxItself)
		assert.Equal(t, nil, err)
		
		testTxTokenV2InvalidFee(dummyDB, paramToCreateTx2, t)
		testTxTokenV2OneFakeOutput(tx2, dummyDB, paramToCreateTx2, t)
		testTxTokenV2OneDoubleSpentInput(tx2, dummyDB, t)
	}
}

// TODO: explore the case where proofs in txnormal or txbase are nulled out, where the tx then still passes sanity check
// and then panic inside Verify function
func testTxTokenV2DeletedProof(txv2 *TxTokenVersion2, t *testing.T){
	// try setting the proof to nil, then verify
	// it should not go through
	savedProof := txv2.GetTxTokenData().TxNormal.GetProof()
	txv2.GetTxTokenData().TxNormal.SetProof(nil)
	isValid,_ := txv2.ValidateSanityData(nil,nil,nil,0)
	assert.Equal(t,false,isValid)
	txv2.GetTxTokenData().TxNormal.SetProof(savedProof)

	savedProof = txv2.GetTxBase().GetProof()
	txv2.GetTxBase().SetProof(nil)
	isValid,_ = txv2.ValidateSanityData(nil,nil,nil,0)
	assert.Equal(t,false,isValid)
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

func testTxTokenV2OneDoubleSpentInput(tokenTx *TxTokenVersion2, db *statedb.StateDB, t *testing.T){
		otaBytes := tokenTx.GetTxTokenData().TxNormal.GetProof().GetInputCoins()[0].GetKeyImage().ToBytesS()
		statedb.StoreSerialNumbers(db, *tokenTx.GetTokenID(), [][]byte{otaBytes}, 0)
		isValid,err := tokenTx.ValidateTransaction(true,db,nil,0,nil,false,true)
		// verify by itself passes
		assert.Equal(t,nil,err)
		assert.Equal(t,true,isValid)

		// verify with blockchain fails
		err = tokenTx.ValidateTxWithBlockChain(nil, nil ,nil, 0, db)
		assert.NotEqual(t,nil,err)
		
}