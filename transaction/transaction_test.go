package transaction

// import(
// 	"math/rand"
// 	"testing"
// 	"time"
// 	"os"
// 	"strconv"

// 	"github.com/incognitochain/incognito-chain/common"
// 	"github.com/incognitochain/incognito-chain/incognitokey"
// 	"github.com/incognitochain/incognito-chain/privacy/key"
// 	"github.com/incognitochain/incognito-chain/privacy/operation"
// )

// func prepareKeySets(numKeySets int) ([]*incognitokey.KeySet, error) {
// 	keySets := make([]*incognitokey.KeySet, numKeySets)
// 	//generate keysets: we want the public key to be in Shard 0
// 	for i := 0; i < numKeySets; i++ {
// 		for {
// 			//generate a private key
// 			privateKey := key.GeneratePrivateKey(common.RandBytes(32))

// 			//make keySets from privateKey
// 			keySet := new(incognitokey.KeySet)
// 			err := keySet.InitFromPrivateKey(&privateKey)

// 			if err != nil {
// 				return nil, err
// 			}

// 			//we want the public key to belong to Shard 0
// 			if keySet.PaymentAddress.Pk[31] == 0 {
// 				keySets[i] = keySet
// 				break
// 			}
// 		}
// 	}
// 	return keySets, nil
// }

// func BenchmarkTxV1BatchVerify(b *testing.B) {
// 	rand.Seed(time.Now().UnixNano())
// 	clargs := os.Args[5:]
// 	// fmt.Println(clargs)

// 	numOfInputs,_ := strconv.Atoi(clargs[0])
// 	numOfOutputs,_ := strconv.Atoi(clargs[1])
// 	numOfPrivateKeys := 50
// 	// fmt.Printf("\n------------------TxVersion2 Verify Benchmark\n")
// 	// fmt.Printf("Number of transactions : %d\n", 1)
// 	// fmt.Printf("Number of inputs       : %d\n", numOfInputs)
// 	// fmt.Printf("Number of outputs      : %d\n", numOfOutputs)
// 	// fmt.Println("Will prepare keys")
// 	keySets, _ := prepareKeySets(numOfPrivateKeys)
// 	numOfTxs := numOfPrivateKeys

// 	var txsForBenchmark []*TxVersion1
// 	for i:=0; i < numOfTxs; i++ {
// 		keySet := keySets[i]
// 		pubKey, _ := new(operation.Point).FromBytesS(keySet.PaymentAddress.Pk)
// 		// fmt.Printf("Will create coins for tx #%d\n",i)
// 		coins, _ := createAndSaveCoinV1s(2*numOfInputs, 0, keySet.PrivateKey, pubKey, dummyDB)

// 		tx := new(TxVersion1)
// 		// r := RandInt() % 90
// 		inputCoins := coins[:numOfInputs]
// 		// fmt.Printf("Will create tx params for tx #%d\n",i)
// 		_, txPrivacyParams, _ := createTxPrivacyInitParams(keySet, inputCoins, true, numOfOutputs)
// 		// assert.Equal(t, nil, err, "createTxPrivacyInitParams returns an error: %v", err)
// 		// fmt.Printf("Finished creating tx params for tx #%d\n",i)
// 		tx.Init(txPrivacyParams)
// 		txsForBenchmark = append(txsForBenchmark, tx)

// 	}

// 	batchLength, _ := strconv.Atoi(clargs[2])
// 	// each loop verifies x transactions as one batch
// 	// so the ops/sec will need to be divided by x afterwards
// 	// for fair comparison
// 	b.ResetTimer()
// 	var pass bool
// 	for loop := 0; loop < b.N; loop++ {
// 		var batchContent []metadata.Transaction
// 		for j:=0;j<batchLength;j++{
// 			chosenIndex := RandInt() % len(txsForBenchmark)
// 			currentTx := txsForBenchmark[chosenIndex]
// 			currentTx.ValidateSanityData(nil,nil,nil,0)
// 			currentTx.ValidateTxWithBlockChain(nil, nil, nil, shardID, dummyDB)

// 			batchContent = append(batchContent, currentTx)
// 		}
// 		batch := NewBatchTransaction(batchContent)
// 		// fmt.Println("About to verify batch")
// 		success, _, _ := batch.Validate(dummyDB, nil)
// 		if !success{
// 			fmt.Println("Something wrong")
// 			panic("Invalid tx batch")
// 		}
// 		pass = true
// 	}
// 	assert.Equal(b,true,pass)
// }

// func BenchmarkTxV1Verify(b *testing.B) {
// 	rand.Seed(time.Now().UnixNano())
// 	clargs := os.Args[5:]
// 	// fmt.Println(clargs)

// 	numOfInputs,_ := strconv.Atoi(clargs[0])
// 	numOfOutputs,_ := strconv.Atoi(clargs[1])
// 	numOfPrivateKeys := 50
// 	numOfTxs := numOfPrivateKeys
// 	// fmt.Printf("\n------------------TxVersion1 Verify Benchmark\n")
// 	// fmt.Printf("Number of transactions : %d\n", numOfTxs)
// 	// fmt.Printf("Number of inputs       : %d\n", numOfInputs)
// 	// fmt.Printf("Number of outputs      : %d\n", numOfOutputs)
// 	keySets, _ := prepareKeySets(numOfPrivateKeys)


// 	var txsForBenchmark []*TxVersion1
// 	for i:=0; i < numOfTxs; i++ {
// 		keySet := keySets[i]
// 		pubKey, _ := new(operation.Point).FromBytesS(keySet.PaymentAddress.Pk)
// 		coins, _ := createAndSaveCoinV1s(2*numOfInputs, 0, keySet.PrivateKey, pubKey, dummyDB)

// 		tx := new(TxVersion1)
// 		// r := RandInt() % 90
// 		inputCoins := coins[:numOfInputs]
// 		_, txPrivacyParams, _ := createTxPrivacyInitParams(keySet, inputCoins, true, numOfOutputs)
// 		// assert.Equal(t, nil, err, "createTxPrivacyInitParams returns an error: %v", err)
// 		tx.Init(txPrivacyParams)
// 		txsForBenchmark = append(txsForBenchmark, tx)

// 	}

// 	// batchLength, _ := strconv.Atoi(clargs[2])
// 	// each loop verifies 20 transactions as one batch
// 	// so the ops/sec will need to be divided by 20 afterwards
// 	// for fair comparison
// 	b.ResetTimer()
// 	for loop := 0; loop < b.N; loop++ {
// 		chosenIndex := RandInt() % len(txsForBenchmark)
// 		currentTx := txsForBenchmark[chosenIndex]

// 		var err error
// 		var isValid bool
// 		isValid, err = currentTx.ValidateSanityData(nil,nil,nil,0)
// 		if !isValid{
// 			panic("Invalid tx sanity")
// 		}
// 		isValid, err = currentTx.ValidateTxByItself(true, dummyDB, nil, nil, byte(0), true, nil, nil)
// 		if !isValid{
// 			panic("Invalid tx")
// 		}
// 		err = currentTx.ValidateTxWithBlockChain(nil, nil, nil, shardID, dummyDB)
// 		if err!=nil{
// 			panic("Invalid tx : double spent")
// 		}
// 	}
// }

// func BenchmarkTxV2Verify(b *testing.B){
// 	rand.Seed(time.Now().UnixNano())
// 	// fmt.Println(os.Args[5:])
// 	clargs := os.Args[5:]
// 	// fmt.Println(clargs)

// 	numOfInputs,_ := strconv.Atoi(clargs[0])
// 	numOfOutputs,_ := strconv.Atoi(clargs[1])
// 	// our setup will cause an extra 'change' output coin to be added so we fix here
// 	numOfOutputs -= 1
// 	numOfPrivateKeys := 50
// 	// fmt.Printf("\n------------------TxVersion2 Verify Benchmark\n")
// 	// fmt.Printf("Number of transactions : %d\n", numOfPrivateKeys)
// 	// fmt.Printf("Number of inputs       : %d\n", numOfInputs)
// 	// fmt.Printf("Number of outputs      : %d\n", numOfOutputs)
// 	preparePaymentKeys(numOfPrivateKeys,nil)
// 	numOfTxs := numOfPrivateKeys
// 	// dummyDB, _ = statedb.NewWithPrefixTrie(emptyRoot, warperDBStatedbTest)

// 	var txsForBenchmark []*TxVersion2
// 	for txInd:=0;txInd<numOfTxs;txInd++{
// 		// pastCoins are coins we forcefully write into the dummyDB to simulate the db having OTAs in the past
// 		// we make sure there are a lot - and a lot - of past coins from all those simulated private keys
// 		pastCoins := make([]coin.Coin, numOfInputs)
// 		for i, _ := range pastCoins {
// 			tempCoin,_ := coin.NewCoinFromPaymentInfo(paymentInfo[txInd])

// 			// to obtain a PlainCoin to feed into input of TX, we need to conceal & decrypt it (it makes sure all fields are right, as opposed to just casting the type to PlainCoin)
// 			tempCoin.ConcealOutputCoin(keySets[txInd].PaymentAddress.GetPublicView())
// 			pastCoins[i] = tempCoin
// 		}
// 		// use the db's interface to write our simulated pastCoins to the database
// 		// we do need to re-format the data into bytes first
// 		forceSaveCoins(dummyDB, pastCoins, 0, common.PRVCoinID, nil)

// 		// in this test, we randomize the length of inputCoins so we feel safe fixing the length of outputCoins to equal len(dummyPrivateKeys)
// 		// since the function `tx.Init` takes output's paymentinfo and creates outputCoins inside of it, we only create the paymentinfo here
// 		paymentInfoOut := make([]*key.PaymentInfo, numOfOutputs)
// 		for i, _ := range paymentInfoOut {
// 			paymentInfoOut[i] = key.InitPaymentInfo(keySets[txInd].PaymentAddress,uint64(3000),[]byte("bench out"))
// 			// fmt.Println(paymentInfo[i])
// 		}
// 		// now we take some of those stored coins to use as TX input
// 		// for the TX to be valid, these inputs must associate to one same private key
// 		// (it's guaranteed by our way of indexing the pastCoins array)
// 		inputCoins := make([]coin.PlainCoin,numOfInputs)
// 		for i,_ := range inputCoins{
// 			inputCoins[i],_ = pastCoins[i].Decrypt(keySets[txInd])
// 		}

// 		// now we calculate the fee = sum(Input) - sum(Output)
// 		// sumIn := uint64(400000*numOfPrivateKeys*numOfInputs)
// 		// sumOut := uint64(3000*numOfOutputs)

// 		initializingParams := tx_generic.NewTxPrivacyInitParams(dummyPrivateKeys[txInd],
// 			paymentInfoOut,inputCoins,
// 			1,true,
// 			dummyDB,
// 			nil,
// 			nil,
// 			[]byte{},
// 		)
// 		// creating the TX object
// 		tx := &TxVersion2{}
// 		// actually making the TX
// 		// `Init` function will also create all necessary proofs and attach them to the TX
// 		tx.Init(initializingParams)

// 		txsForBenchmark = append(txsForBenchmark, tx)
// 	}

// 	b.ResetTimer()
// 	for loop := 0; loop < b.N; loop++ {
// 		chosenIndex := RandInt() % len(txsForBenchmark)
// 		currentTx := txsForBenchmark[chosenIndex]
// 		// verify the TX
// 		// params : hasPrivacy bool, transactionStateDB *statedb.StateDB, bridgeStateDB *statedb.StateDB,
// 		// 			shardID byte (we're testing with only 1 shard),
// 		//			tokenID *common.Hash (set to nil, meaning we use PRV),
// 		//			isBatch bool, isNewTransaction bool
// 		var err error
// 		var isValid bool
// 		isValid, err = currentTx.ValidateSanityData(nil,nil,nil,0)
// 		if !isValid{
// 			panic("Invalid tx sanity")
// 		}
// 		isValid, err = currentTx.ValidateTxByItself(true, dummyDB, nil, nil, byte(0), true, nil, nil)
// 		if !isValid{
// 			panic("Invalid tx")
// 		}
// 		err = currentTx.ValidateTxWithBlockChain(nil, nil, nil, shardID, dummyDB)
// 		if err!=nil{
// 			panic("Invalid tx : double spent")
// 		}
// 	}
// }

// func BenchmarkTxV2BatchVerify(b *testing.B){
// 	rand.Seed(time.Now().UnixNano())
// 	// fmt.Println(os.Args[5:])
// 	clargs := os.Args[5:]
// 	// fmt.Println(clargs)

// 	numOfInputs,_ := strconv.Atoi(clargs[0])
// 	numOfOutputs,_ := strconv.Atoi(clargs[1])
// 	// our setup will cause an extra 'change' output coin to be added so we fix here
// 	numOfOutputs -= 1
// 	numOfPrivateKeys := 50
// 	// fmt.Printf("\n------------------TxVersion2 Verify Benchmark\n")
// 	// fmt.Printf("Number of transactions : %d\n", 1)
// 	// fmt.Printf("Number of inputs       : %d\n", numOfInputs)
// 	// fmt.Printf("Number of outputs      : %d\n", numOfOutputs)
// 	preparePaymentKeys(numOfPrivateKeys,nil)
// 	numOfTxs := numOfPrivateKeys
// 	// dummyDB, _ := statedb.NewWithPrefixTrie(emptyRoot, warperDBStatedbTest)

// 	var txsForBenchmark []*TxVersion2
// 	for txInd:=0;txInd<numOfTxs;txInd++{
// 		// pastCoins are coins we forcefully write into the dummyDB to simulate the db having OTAs in the past
// 		// we make sure there are a lot - and a lot - of past coins from all those simulated private keys
// 		pastCoins := make([]coin.Coin, numOfInputs)
// 		for i, _ := range pastCoins {
// 			tempCoin,_ := coin.NewCoinFromPaymentInfo(paymentInfo[txInd])

// 			// to obtain a PlainCoin to feed into input of TX, we need to conceal & decrypt it (it makes sure all fields are right, as opposed to just casting the type to PlainCoin)
// 			tempCoin.ConcealOutputCoin(keySets[txInd].PaymentAddress.GetPublicView())
// 			pastCoins[i] = tempCoin
// 		}
// 		// use the db's interface to write our simulated pastCoins to the database
// 		// we do need to re-format the data into bytes first
// 		forceSaveCoins(dummyDB, pastCoins, 0, common.PRVCoinID, nil)

// 		// in this test, we randomize the length of inputCoins so we feel safe fixing the length of outputCoins to equal len(dummyPrivateKeys)
// 		// since the function `tx.Init` takes output's paymentinfo and creates outputCoins inside of it, we only create the paymentinfo here
// 		paymentInfoOut := make([]*key.PaymentInfo, numOfOutputs)
// 		for i, _ := range paymentInfoOut {
// 			paymentInfoOut[i] = key.InitPaymentInfo(keySets[txInd].PaymentAddress,uint64(3000),[]byte("bench out"))
// 			// fmt.Println(paymentInfo[i])
// 		}
// 		// now we take some of those stored coins to use as TX input
// 		// for the TX to be valid, these inputs must associate to one same private key
// 		// (it's guaranteed by our way of indexing the pastCoins array)
// 		inputCoins := make([]coin.PlainCoin,numOfInputs)
// 		for i,_ := range inputCoins{
// 			inputCoins[i],_ = pastCoins[i].Decrypt(keySets[txInd])
// 		}

// 		// now we calculate the fee = sum(Input) - sum(Output)
// 		// sumIn := uint64(400000*numOfPrivateKeys*numOfInputs)
// 		// sumOut := uint64(3000*numOfOutputs)

// 		initializingParams := tx_generic.NewTxPrivacyInitParams(dummyPrivateKeys[txInd],
// 			paymentInfoOut,inputCoins,
// 			1,true,
// 			dummyDB,
// 			nil,
// 			nil,
// 			[]byte{},
// 		)
// 		// creating the TX object
// 		tx := &TxVersion2{}
// 		// actually making the TX
// 		// `Init` function will also create all necessary proofs and attach them to the TX
// 		tx.Init(initializingParams)

// 		txsForBenchmark = append(txsForBenchmark, tx)
// 	}

// 	batchLength, _ := strconv.Atoi(clargs[2])
// 	// each loop verifies 20 transactions as one batch
// 	// so the ops/sec will need to be divided by 20 afterwards
// 	// for fair comparison
// 	b.ResetTimer()
// 	var pass bool
// 	for loop := 0; loop < b.N; loop++ {
// 		var batchContent []metadata.Transaction
// 		chosenIndex := RandInt() % len(txsForBenchmark)
// 		for j:=0;j<batchLength;j++{
// 			chosenIndex := (chosenIndex+1)%len(txsForBenchmark)
// 			currentTx := txsForBenchmark[chosenIndex]
// 			currentTx.ValidateSanityData(nil,nil,nil,0)
// 			currentTx.ValidateTxWithBlockChain(nil, nil, nil, shardID, dummyDB)

// 			batchContent = append(batchContent, currentTx)
// 		}
// 		batch := NewBatchTransaction(batchContent)
// 		success, _, _ := batch.Validate(dummyDB, nil)
// 		if !success{
// 			fmt.Println("Something wrong")
// 			panic("Invalid tx batch")
// 		}
// 		pass = true
// 	}
// 	assert.Equal(b,true,pass)
// }