package transaction

import (
	"bytes"
	"math/big"
	"testing"
	"fmt"
	"io/ioutil"
	"os"
	"encoding/json"
	"unicode"
	"math/rand"
	"time"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/privacy/operation"
	"github.com/incognitochain/incognito-chain/privacy/key"
	"github.com/incognitochain/incognito-chain/privacy/coin"
	"github.com/incognitochain/incognito-chain/privacy"
	// "github.com/incognitochain/incognito-chain/privacy/privacy_v2"
	"github.com/incognitochain/incognito-chain/privacy/privacy_v2/bulletproofs"
	"github.com/incognitochain/incognito-chain/trie"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/stretchr/testify/assert"
)

var (
	// num of private keys
	maxPrivateKeys = 20
	minPrivateKeys = 1

	maxInputs = 20
	minInputs = 1

	maxTries = 100
	numOfLoops = 5000
)
var (
	warperDBStatedbTest statedb.DatabaseAccessWarper
	emptyRoot           = common.HexToHash(common.HexEmptyRoot)
	prefixA             = "serialnumber"
	prefixB             = "serialnumberderivator"
	prefixC             = "serial"
	prefixD             = "commitment"
	prefixE             = "outputcoin"
	keysA               = []common.Hash{}
	keysB               = []common.Hash{}
	keysC               = []common.Hash{}
	keysD               = []common.Hash{}
	keysE               = []common.Hash{}
	valuesA             = [][]byte{}
	valuesB             = [][]byte{}
	valuesC             = [][]byte{}
	valuesD             = [][]byte{}
	valuesE             = [][]byte{}

	limit100000 = 100000
	limit10000  = 10000
	limit1000   = 1000
	limit100    = 100
	limit1      = 1

	dummyDB *statedb.StateDB
	bridgeDB *statedb.StateDB
	dummyPrivateKeys []*key.PrivateKey
	keySets []*incognitokey.KeySet
	paymentInfo []*key.PaymentInfo
)

var _ = func() (_ struct{}) {
// initialize a `test` db in the OS's tempdir
// and with it, a db access wrapper that reads/writes our transactions
	fmt.Println("This runs before init()!")
	// numOfInputs = RandInt() % (maxInputs - minInputs + 1) + minInputs
	// numOfOutputs = RandInt() % (maxInputs - minInputs + 1) + minInputs
	// fmt.Printf("\n------------------TxVersion2 Verify Benchmark\n")
	// fmt.Printf("Number of transactions : %d\n", 1)
	// fmt.Printf("Number of inputs       : %d\n", numOfInputs)
	// fmt.Printf("Number of outputs      : %d\n", numOfOutputs)
	privacy.LoggerV1.Init(common.NewBackend(nil).Logger("test", true))
	privacy.LoggerV2.Init(common.NewBackend(nil).Logger("test", true))
	// fmt.Println("And then")
	// fmt.Println(privacy.LoggerV2)
	// fmt.Println(privacy_v2.Logger)
	Logger.Init(common.NewBackend(nil).Logger("test", true))
	bulletproofs.Logger.Init(common.NewBackend(nil).Logger("test", true))
	dbPath, err := ioutil.TempDir(os.TempDir(), "test_statedb_")
	if err != nil {
		panic(err)
	}
	diskBD, _ := incdb.Open("leveldb", dbPath)
	warperDBStatedbTest = statedb.NewDatabaseAccessWarper(diskBD)
	dummyDB, _ = statedb.NewWithPrefixTrie(emptyRoot, warperDBStatedbTest)
	bridgeDB  = dummyDB.Copy()
	trie.Logger.Init(common.NewBackend(nil).Logger("test", true))
	return
}()

func forceSaveCoins(db *statedb.StateDB, coinsToBeSaved []coin.Coin, shardID byte, tokenID common.Hash, t *testing.T){
	coinsInBytes := make([][]byte, 0)
	otas := make([][]byte, 0)
	for _,c := range coinsToBeSaved{
		if t!=nil{
			assert.Equal(t,2,int(c.GetVersion()))
		}
		coinsInBytes = append(coinsInBytes, c.Bytes())
		otas = append(otas, c.GetPublicKey().ToBytesS())
	}
	err := statedb.StoreOTACoinsAndOnetimeAddresses(db, tokenID, 0, coinsInBytes, otas, shardID)
	if t!=nil{
		assert.Equal(t,nil,err)
	}
}

func preparePaymentKeys(count int, t *testing.T){
	// create many random private keys
	// then use each privatekey to derive Incognito keyset (various keys for everything inside the protocol)
	// we ensure they all belong in shard 0 for this test
		
	// PaymentInfo is like `intent` for making Coin.
	// the paymentInfo slice here will be used to create pastCoins & inputCoins 
	// we populate `value` fields with some arbitrary, big-enough constant (here, 4000*len)
	// `message` field can be anything
	dummyPrivateKeys = make([]*key.PrivateKey,count)
	keySets = make([]*incognitokey.KeySet,len(dummyPrivateKeys))
	paymentInfo = make([]*key.PaymentInfo, len(dummyPrivateKeys))
	for i := 0; i < count; i += 1 {
		for{
			privateKey := key.GeneratePrivateKey(common.RandBytes(32))
			dummyPrivateKeys[i] = &privateKey
			keySets[i] = new(incognitokey.KeySet)
			err := keySets[i].InitFromPrivateKey(dummyPrivateKeys[i])
			if t!=nil{
				assert.Equal(t, nil, err)
			}
			paymentInfo[i] = key.InitPaymentInfo(keySets[i].PaymentAddress, uint64(400000*len(dummyPrivateKeys)), []byte("test in"))
			pkb := []byte(paymentInfo[i].PaymentAddress.Pk)
			if common.GetShardIDFromLastByte(pkb[len(pkb)-1])==shardID{
				break
			}
		}
	}
	// fmt.Println("Key & PaymentInfo generation finished")
}

func TestSigPubKeyCreationAndMarshalling(t *testing.T) {
// here m, n are not very specific so we give them generous range
	m := RandInt() % (maxPrivateKeys - minInputs + 1) + minInputs
	n := RandInt() % (maxPrivateKeys - minInputs + 1) + minInputs
	var err error
	for i := 0; i < numOfLoops; i += 1 {
		fmt.Printf("\n------------------TxTokenVersion2 SigPubKey Test\n")
		maxLen := new(big.Int)
		maxLen.SetString("1000000000000000000", 10)
		indexes := make([][]*big.Int, n)
		for i := 0; i < n; i += 1 {
			row := make([]*big.Int, m)
			for j := 0; j < m; j += 1 {
				row[j], err = common.RandBigIntMaxRange(maxLen)
				assert.Equal(t, nil, err, "Should not have any bug when Randomizing Int Max Range")
			}
			indexes[i] = row
		}

		txSig := new(TxSigPubKeyVer2)
		txSig.Indexes = indexes

		b, err := txSig.Bytes()
		assert.Equal(t, nil, err, "Should not have any bug when txSig.ToBytes")

		txSig2 := new(TxSigPubKeyVer2)
		err = txSig2.SetBytes(b)
		assert.Equal(t, nil, err, "Should not have any bug when txSig.FromBytes")

		b2, err := txSig2.Bytes()
		assert.Equal(t, nil, err, "Should not have any bug when txSig2.ToBytes")
		assert.Equal(t, true, bytes.Equal(b, b2))

		n1 := len(txSig.Indexes)
		m1 := len(txSig.Indexes[0])
		n2 := len(txSig2.Indexes)
		m2 := len(txSig2.Indexes[0])

		assert.Equal(t, n1, n2, "Two Indexes length should be equal")
		assert.Equal(t, m1, m2, "Two Indexes length should be equal")
		for i := 0; i < n; i += 1 {
			for j := 0; j < m; j += 1 {
				b1 := txSig.Indexes[i][j].Bytes()
				b2 := txSig2.Indexes[i][j].Bytes()
				assert.Equal(t, true, bytes.Equal(b1, b2), "Indexes[i][j] should be equal for every i j")
			}
		}
	}
	fmt.Println("SigPubKey Marshalling Test successful")
}

// tx salary is just a validator printing block rewards in PRV, without privacy
// no need for dummy input
func TestTxV2Salary(t *testing.T){
	numOfPrivateKeys := 2
	
	for loop := 0; loop < numOfLoops; loop++ {
		fmt.Printf("\n------------------TxVersion2 Salary Test\n")
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
			tempCoin.ConcealOutputCoin(keySets[i].PaymentAddress.GetPublicView())
			assert.Equal(t, true, tempCoin.IsEncrypted())
			assert.Equal(t, true, tempCoin.GetSharedRandom() == nil)
			_, err = tempCoin.Decrypt(keySets[i])
			assert.Equal(t,nil,err)
			theCoins[i] = tempCoin
			theCoinsGeneric[i] = tempCoin
		}
		forceSaveCoins(dummyDB, []coin.Coin{theCoinsGeneric[1]}, 0, common.PRVCoinID, t)

		// creating the TX object
		tx := &TxVersion2{}
		// actually making the salary TX
		err = tx.InitTxSalary(theCoins[0], dummyPrivateKeys[0], dummyDB, nil)

		isValid,err := tx.ValidateTxSalary(dummyDB)
		assert.Equal(t,nil,err)
		assert.Equal(t,true,isValid)

		testTxV2JsonMarshaler(tx, 50, dummyDB, t)
		// someInvalidTxs := getCorruptedJsonDeserializedTxs(tx, t)
		// for _,theInvalidTx := range someInvalidTxs{
		// 	txSpecific, ok := theInvalidTx.(*TxVersion2)
		// 	assert.Equal(t, true, ok)
		// 	// look for potential panics by calling verify
		// 	isSane, _ := txSpecific.ValidateSanityData(nil,nil,nil,0)
		// 	// if it doesnt pass sanity then the next validation could panic, it's ok by spec
		// 	if !isSane{
		// 		continue
		// 	}
		// 	txSpecific.ValidateTxByItself(true, dummyDB, nil, nil, byte(0), true, nil, nil)
		// }

		malTx := &TxVersion2{}
		// this other coin is already in db so it must be rejected
		err = malTx.InitTxSalary(theCoins[1], dummyPrivateKeys[0], dummyDB, nil)
		assert.NotEqual(t,nil,err)
	}
}

func TestTxV2ProveWithPrivacy(t *testing.T){
	numOfPrivateKeys := RandInt() % (maxPrivateKeys - minPrivateKeys + 1) + minPrivateKeys
	numOfInputs := RandInt() % (maxInputs - minInputs + 1) + minInputs
	// dummyDB, _ = statedb.NewWithPrefixTrie(emptyRoot, warperDBStatedbTest)
	for loop := 0; loop < numOfLoops; loop++ {
		fmt.Printf("\n------------------TxVersion2 Main Test\n")
		var err error
		preparePaymentKeys(numOfPrivateKeys,t)

		// pastCoins are coins we forcefully write into the dummyDB to simulate the db having OTAs in the past
		// we make sure there are a lot - and a lot - of past coins from all those simulated private keys
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

		// in this test, we randomize the length of inputCoins so we feel safe fixing the length of outputCoins to equal len(dummyPrivateKeys)
		// since the function `tx.Init` takes output's paymentinfo and creates outputCoins inside of it, we only create the paymentinfo here
		paymentInfoOut := make([]*key.PaymentInfo, len(dummyPrivateKeys))
		for i, _ := range dummyPrivateKeys {
			paymentInfoOut[i] = key.InitPaymentInfo(keySets[i].PaymentAddress,uint64(3000),[]byte("test out"))
			// fmt.Println(paymentInfo[i])
		}

		// use the db's interface to write our simulated pastCoins to the database
		// we do need to re-format the data into bytes first
		forceSaveCoins(dummyDB, pastCoins, 0, common.PRVCoinID, t)


		// now we take some of those stored coins to use as TX input
		// for the TX to be valid, these inputs must associate to one same private key
		// (it's guaranteed by our way of indexing the pastCoins array)
		inputCoins := make([]coin.PlainCoin,numOfInputs)
		for i,_ := range inputCoins{
			var err error
			inputCoins[i],err = pastCoins[i*len(dummyPrivateKeys)].Decrypt(keySets[0])
			assert.Equal(t,nil,err)
		}

		// now we calculate the fee = sum(Input) - sum(Output)
		sumIn := uint64(4000*len(dummyPrivateKeys)*numOfInputs)
		sumOut := uint64(3000*len(paymentInfoOut))
		assert.Equal(t,true,sumIn > sumOut)

		initializingParams := NewTxPrivacyInitParams(dummyPrivateKeys[0],
			paymentInfoOut,inputCoins,
			sumIn-sumOut,true,
			dummyDB,
			nil,
			nil,
			[]byte{},
		)
		// creating the TX object
		tx := &TxVersion2{}
		// actually making the TX
		// `Init` function will also create all necessary proofs and attach them to the TX
		err = tx.Init(initializingParams)
		assert.Equal(t,nil,err)

		// verify the TX
		// params : hasPrivacy bool, transactionStateDB *statedb.StateDB, bridgeStateDB *statedb.StateDB, 
		// 			shardID byte (we're testing with only 1 shard), 
		//			tokenID *common.Hash (set to nil, meaning we use PRV),
		//			isBatch bool, isNewTransaction bool
		// isValid,err := tx.ValidateSanityData(nil,nil,nil,0)
		// assert.Equal(t,nil,err)
		// assert.Equal(t,true,isValid)
		// isValid,err = tx.ValidateTransaction(true,dummyDB,nil,0,nil,false,true)
		// isValid,err = tx.ValidateTxByItself(true, dummyDB, nil, nil, byte(0), true, nil, nil)
		// assert.Equal(t,nil,err)
		// assert.Equal(t,true,isValid)

		// first, test the json marshaller
		testTxV2JsonMarshaler(tx, 25, dummyDB, t)

		// testTxV2DeletedProof(tx, t)
		// testTxV2DuplicateInput(dummyDB, inputCoins, paymentInfoOut, t)
		// testTxV2InvalidFee(dummyDB, inputCoins, paymentInfoOut, t)
		// testTxV2OneFakeInput(tx, dummyDB, initializingParams, pastCoins, t)
		// testTxV2OneFakeOutput(tx, dummyDB, initializingParams, paymentInfoOut, t)
		// testTxV2OneDoubleSpentInput(dummyDB, inputCoins, paymentInfoOut, pastCoins, t)
	}
}

func testTxV2DeletedProof(txv2 *TxVersion2, t *testing.T){
	// try setting the proof to nil, then verify
	// it should not go through
	savedProof := txv2.Proof
	txv2.Proof = nil
	isValid,err := txv2.ValidateSanityData(nil,nil,nil,0)
	assert.NotEqual(t,nil,err)
	assert.Equal(t,false,isValid)
	txv2.Proof = savedProof
}

func testTxV2DuplicateInput(db *statedb.StateDB, inputCoins []coin.PlainCoin, paymentInfoOut []*key.PaymentInfo, t *testing.T){
	dup := &coin.CoinV2{}
	dup.SetBytes(inputCoins[0].Bytes())
	// used the same coin twice in inputs
	malInputCoins := append(inputCoins,dup)
	malFeeParams := NewTxPrivacyInitParams(dummyPrivateKeys[0],
		paymentInfoOut,malInputCoins,
		10,true,
		db,
		nil,
		nil,
		[]byte{},
	)
	malTx := &TxVersion2{}
	errMalInit := malTx.Init(malFeeParams)
	assert.Equal(t,nil,errMalInit)
	// sanity should be fine
	isValid,err := malTx.ValidateSanityData(nil,nil,nil,0)
	assert.Equal(t,nil,err)
	assert.Equal(t,true,isValid)
	// validate should reject due to Verify() in PaymentProofV2
	isValid,_ = malTx.ValidateTxByItself(true, db, nil, nil, byte(0), true, nil, nil)
	assert.Equal(t,false,isValid)
}

func testTxV2InvalidFee(db *statedb.StateDB, inputCoins []coin.PlainCoin, paymentInfoOut []*key.PaymentInfo, t *testing.T){
	// a set of init params where sum(Input) < fee + sum(Output)
	// let's say someone tried to use this invalid fee for tx
	// we should encounter an error here
	sumIn := uint64(4000*len(dummyPrivateKeys)*len(inputCoins))
	sumOut := uint64(3000*len(paymentInfoOut))
	assert.Equal(t,true,sumIn > sumOut)
	malFeeParams := NewTxPrivacyInitParams(dummyPrivateKeys[0],
		paymentInfoOut,inputCoins,
		sumIn-sumOut+1111,true,
		db,
		nil,
		nil,
		[]byte{},
	)
	malTx := &TxVersion2{}
	errMalInit := malTx.Init(malFeeParams)
	assert.NotEqual(t,nil,errMalInit)
	isValid,errMalVerify := malTx.ValidateTxByItself(true, db, nil, nil, byte(0), true, nil, nil)
	assert.NotEqual(t,nil,errMalVerify)
	assert.Equal(t,false,isValid)
}

func testTxV2OneFakeInput(txv2 *TxVersion2, db *statedb.StateDB, params *TxPrivacyInitParams, pastCoins []coin.Coin, t *testing.T){
	// likewise, if someone took an already proven tx and swaps one input coin 
	// for another random coin from outside, the tx cannot go through
	// (here we only meddle with coin-changing - not adding/removing - since length checks are included within mlsag)
	var err error
	theProof := txv2.GetProof()
	inputCoins := theProof.GetInputCoins()
	numOfInputs := len(inputCoins)
	changed := RandInt() % numOfInputs
	saved := inputCoins[changed]
	inputCoins[changed],_ = pastCoins[len(dummyPrivateKeys)*(numOfInputs+1)].Decrypt(keySets[0])
	theProof.SetInputCoins(inputCoins)
	// malInputParams := NewTxPrivacyInitParams(dummyPrivateKeys[0],
	// 	paymentInfoOut,inputCoins,
	// 	1,true,
	// 	db,
	// 	nil,
	// 	nil,
	// 	[]byte{},
	// )
	err = resignUnprovenTx(keySets, txv2, params, nil)
	assert.Equal(t,nil,err)
	isValid,err := txv2.ValidateSanityData(nil,nil,nil,0)
	assert.Equal(t,nil,err)
	assert.Equal(t,true,isValid)
	isValid,err = txv2.ValidateTxByItself(true, db, nil, nil, byte(0), true, nil, nil)
	// should fail at signature since mlsag needs commitments from inputs
	// fmt.Printf("One fake valid input -> %v\n",err)
	assert.Equal(t,false,isValid)
	inputCoins[changed] = saved
	theProof.SetInputCoins(inputCoins)
	err = resignUnprovenTx(keySets, txv2, params, nil)
	isValid,err = txv2.ValidateSanityData(nil,nil,nil,0)
	assert.Equal(t,nil,err)
	assert.Equal(t,true,isValid)
	isValid,err = txv2.ValidateTxByItself(true, db, nil, nil, byte(0), true, nil, nil)
	assert.Equal(t,nil,err)
	assert.Equal(t,true,isValid)


	// saved = inputCoins[changed]
	// inputCoins[changed] = nil
	// malTx.GetProof().SetInputCoins(inputCoins)
	// isValid,err = malTx.ValidateTxByItself(true, db, nil, nil, byte(0), true, nil, nil)
	// // verify must fail
	// assert.NotEqual(t,nil,err)
	// assert.Equal(t,false,isValid)
	// inputCoins[changed] = saved
}

func testTxV2OneFakeOutput(txv2 *TxVersion2, db *statedb.StateDB, params *TxPrivacyInitParams, paymentInfoOut []*key.PaymentInfo, t *testing.T){
	// similar to the above. All these verifications should fail
	var err error
	outs := txv2.GetProof().GetOutputCoins()
	prvOutput,ok := outs[0].(*coin.CoinV2)
	savedCoinBytes := prvOutput.Bytes()
	assert.Equal(t,true,ok)
	prvOutput.Decrypt(keySets[0])
	// set amount to something wrong
	prvOutput.SetValue(6996)
	prvOutput.SetSharedRandom(operation.RandomScalar())
	prvOutput.ConcealOutputCoin(keySets[0].PaymentAddress.GetPublicView())
	err = resignUnprovenTx(keySets, txv2, params, nil)
	assert.Equal(t,nil,err)
	isValid,err := txv2.ValidateSanityData(nil,nil,nil,0)
	assert.Equal(t,nil,err)
	assert.Equal(t,true,isValid)
	isValid,err = txv2.ValidateTxByItself(true, db, nil, nil, byte(0), true, nil, nil)
	// verify must fail
	assert.Equal(t,false,isValid)
	// fmt.Printf("Fake output (wrong amount) -> %v\n",err)
	// undo the tampering
	prvOutput.SetBytes(savedCoinBytes)
	outs[0] = prvOutput
	txv2.GetProof().SetOutputCoins(outs)
	err = resignUnprovenTx(keySets, txv2, params, nil)
	assert.Equal(t,nil,err)
	isValid,_ = txv2.ValidateTxByItself(true, db, nil, nil, byte(0), true, nil, nil)
	assert.Equal(t,true,isValid)

	// now instead of changing amount, we change the OTA public key
	theProof := txv2.GetProof()
	outs = theProof.GetOutputCoins()
	prvOutput,ok = outs[0].(*coin.CoinV2)
	savedCoinBytes = prvOutput.Bytes()
	assert.Equal(t,true,ok)
	payInf := paymentInfoOut[0]
	// totally fresh OTA of the same amount, meant for the same PaymentAddress
	newCoin,err  := coin.NewCoinFromPaymentInfo(payInf)
	assert.Equal(t,nil,err)
	newCoin.ConcealOutputCoin(keySets[0].PaymentAddress.GetPublicView())
	theProofSpecific, ok := theProof.(*privacy.ProofV2)
	theBulletProof, ok := theProofSpecific.GetAggregatedRangeProof().(*privacy.AggregatedRangeProofV2)
	cmsv := theBulletProof.GetCommitments()
	cmsv[0] = newCoin.GetCommitment()
	outs[0] = newCoin
	txv2.GetProof().SetOutputCoins(outs)
	err = resignUnprovenTx(keySets, txv2, params, nil)
	assert.Equal(t,nil,err)
	isValid,err = txv2.ValidateSanityData(nil,nil,nil,0)
	assert.Equal(t,nil,err)
	assert.Equal(t,true,isValid)
	isValid,err = txv2.ValidateTxByItself(true, db, nil, nil, byte(0), true, nil, nil)
	// verify must fail
	assert.Equal(t,false,isValid)
	// fmt.Printf("Fake output (wrong receiving OTA) -> %v\n",err)
	// undo the tampering
	prvOutput.SetBytes(savedCoinBytes)
	outs[0] = prvOutput
	cmsv[0] = prvOutput.GetCommitment()
	txv2.GetProof().SetOutputCoins(outs)
	err = resignUnprovenTx(keySets, txv2, params, nil)
	assert.Equal(t,nil,err)
	isValid,_ = txv2.ValidateTxByItself(true, db, nil, nil, byte(0), true, nil, nil)
	assert.Equal(t,true,isValid)

}

func testTxV2OneDoubleSpentInput(db *statedb.StateDB, inputCoins []coin.PlainCoin, paymentInfoOut []*key.PaymentInfo, pastCoins []coin.Coin, t *testing.T){
	// similar to the above. All these verifications should fail
		changed := RandInt() % len(inputCoins)
		malInputParams := NewTxPrivacyInitParams(dummyPrivateKeys[0],
			paymentInfoOut,inputCoins,
			1,true,
			db,
			nil,
			nil,
			[]byte{},
		)
		malTx := &TxVersion2{}
		err := malTx.Init(malInputParams)
		assert.Equal(t,nil,err)
		otaBytes := malTx.GetProof().GetInputCoins()[changed].GetKeyImage().ToBytesS()
		statedb.StoreSerialNumbers(db, common.PRVCoinID, [][]byte{otaBytes}, 0)
		isValid,err := malTx.ValidateTxByItself(true, db, nil, nil, byte(0), true, nil, nil)
		// verify by itself passes
		assert.Equal(t,nil,err)
		assert.Equal(t,true,isValid)

		// verify with blockchain fails
		err = malTx.ValidateTxWithBlockChain(nil, nil ,nil, 0, db)
		assert.NotEqual(t,nil,err)
		
}

func testTxV2JsonMarshaler(tx *TxVersion2, count int, db *statedb.StateDB, t *testing.T){
	someInvalidTxs := getCorruptedJsonDeserializedTxs(tx, count, t)
	for _,theInvalidTx := range someInvalidTxs{
		txSpecific, ok := theInvalidTx.(*TxVersion2)
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
		isSane, _ = txSpecific.ValidateTxByItself(hasPrivacyForPRV, db, nil, nil, shardID, false, nil, nil)
		if !isSane{
			continue
		}
		txSpecific.ValidateTxWithBlockChain(nil, nil, nil, shardID, db)
	}
}

func testTxTokenV2JsonMarshaler(tx *TxTokenVersion2, count int, db *statedb.StateDB, t *testing.T){
	someInvalidTxs := getCorruptedJsonDeserializedTokenTxs(tx, count, t)
	for _,theInvalidTx := range someInvalidTxs{
		txSpecific, ok := theInvalidTx.(*TxTokenVersion2)
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
		isSane, _ = txSpecific.ValidateTxByItself(hasPrivacyForPRV, db, nil, nil, shardID, false, nil, nil)
		if !isSane{
			continue
		}
		txSpecific.ValidateTxWithBlockChain(nil, nil, nil, shardID, db)
	}
}


func getRandomDigit() rune{

	ind := RandInt() % 10
	return rune(int(rune('0'))+ind)
}

func getRandomLetter() rune{
	ind := RandInt() % 52
	if ind < 26{
		return rune(int(rune('A'))+ind)
	}else{
		return rune(int(rune('a'))+ind-26)
	}
}

func getCorruptedJsonDeserializedTxs(tx metadata.Transaction, maxJsonChanges int, t *testing.T) []metadata.Transaction{
	jsonBytes, err := json.Marshal(tx)
	assert.Equal(t, nil, err)

	reconstructedTx, err := NewTransactionFromJsonBytes(jsonBytes)
	assert.Equal(t, nil, err)
	jsonBytesAgain, err := json.Marshal(reconstructedTx)
	assert.Equal(t, true, bytes.Equal(jsonBytes, jsonBytesAgain))
	var result []metadata.Transaction
	// json bytes are readable strings
	// we try to malleify a letter / digit
	for i:=0; i<maxJsonChanges; i++{
		// let the changes stack up many times to exhaust more cases
		s := string(jsonBytesAgain)
		theRunes := []rune(s)
		corruptedIndex := RandInt() % len(theRunes)
		for j:=maxTries;j>0;j--{
			if j==0{
				fmt.Printf("Strange letterless TX with json form : %s\n",s)
				panic("End")
			}
			if unicode.IsLetter(theRunes[corruptedIndex]) || unicode.IsDigit(theRunes[corruptedIndex]){
				break
			}
			// not letter -> retry
			corruptedIndex = RandInt() % len(theRunes)
		}
		// replace this letter with a random one
		if unicode.IsLetter(theRunes[corruptedIndex]){
			theRunes[corruptedIndex] = getRandomLetter()
		}else{
			theRunes[corruptedIndex] = getRandomDigit()
		}


		reconstructedTx, err = NewTransactionFromJsonBytes([]byte(string(theRunes)))
		if err != nil{
			// fmt.Printf("A byte array failed to deserialize\n")
			continue
		}
		result = append(result,reconstructedTx)
	}
	// fmt.Printf("Made %d dummy faulty txs\n",len(result))
	return result
}

func getCorruptedJsonDeserializedTokenTxs(tx TransactionToken, maxJsonChanges int,t *testing.T) []TransactionToken{
	jsonBytes, err := json.Marshal(tx)
	assert.Equal(t, nil, err)

	reconstructedTx, err := NewTransactionTokenFromJsonBytes(jsonBytes)
	assert.Equal(t, nil, err)
	jsonBytesAgain, err := json.Marshal(reconstructedTx)
	assert.Equal(t, true, bytes.Equal(jsonBytes, jsonBytesAgain))
	var result []TransactionToken
	// json bytes are readable strings
	// we try to malleify a letter / digit
	for i:=0; i<maxJsonChanges; i++{
		// let the changes stack up many times to exhaust more cases
		s := string(jsonBytesAgain)
		theRunes := []rune(s)
		corruptedIndex := RandInt() % len(theRunes)
		for j:=maxTries;j>0;j--{
			if j==0{
				fmt.Printf("Strange letterless TX with json form : %s\n",s)
				panic("End")
			}
			if unicode.IsLetter(theRunes[corruptedIndex]) || unicode.IsDigit(theRunes[corruptedIndex]){
				break
			}
			// not letter -> retry
			corruptedIndex = RandInt() % len(theRunes)
		}
		// replace this letter with a random one
		if unicode.IsLetter(theRunes[corruptedIndex]){
			theRunes[corruptedIndex] = getRandomLetter()
		}else{
			theRunes[corruptedIndex] = getRandomDigit()
		}


		reconstructedTx, err = NewTransactionTokenFromJsonBytes([]byte(string(theRunes)))
		if err != nil{
			// fmt.Printf("A byte array failed to deserialize\n")
			continue
		}
		result = append(result,reconstructedTx)
	}
	return result
}

func RandInt() int {
	return rand.Int()
}

func BenchmarkTxV2Verify(b *testing.B){
	rand.Seed(time.Now().UnixNano())
	// fmt.Println(os.Args[5:])
	clargs := os.Args[5:]
	// fmt.Println(clargs)

	numOfInputs,_ := strconv.Atoi(clargs[0])
	numOfOutputs,_ := strconv.Atoi(clargs[1])
	// our setup will cause an extra 'change' output coin to be added so we fix here
	numOfOutputs -= 1
	numOfPrivateKeys := 50
	// fmt.Printf("\n------------------TxVersion2 Verify Benchmark\n")
	// fmt.Printf("Number of transactions : %d\n", numOfPrivateKeys)
	// fmt.Printf("Number of inputs       : %d\n", numOfInputs)
	// fmt.Printf("Number of outputs      : %d\n", numOfOutputs)
	preparePaymentKeys(numOfPrivateKeys,nil)
	numOfTxs := numOfPrivateKeys
	// dummyDB, _ = statedb.NewWithPrefixTrie(emptyRoot, warperDBStatedbTest)
	
	var txsForBenchmark []*TxVersion2
	for txInd:=0;txInd<numOfTxs;txInd++{ 
		// pastCoins are coins we forcefully write into the dummyDB to simulate the db having OTAs in the past
		// we make sure there are a lot - and a lot - of past coins from all those simulated private keys
		pastCoins := make([]coin.Coin, numOfInputs)
		for i, _ := range pastCoins {
			tempCoin,_ := coin.NewCoinFromPaymentInfo(paymentInfo[txInd])

			// to obtain a PlainCoin to feed into input of TX, we need to conceal & decrypt it (it makes sure all fields are right, as opposed to just casting the type to PlainCoin)
			tempCoin.ConcealOutputCoin(keySets[txInd].PaymentAddress.GetPublicView())
			pastCoins[i] = tempCoin
		}
		// use the db's interface to write our simulated pastCoins to the database
		// we do need to re-format the data into bytes first
		forceSaveCoins(dummyDB, pastCoins, 0, common.PRVCoinID, nil)

		// in this test, we randomize the length of inputCoins so we feel safe fixing the length of outputCoins to equal len(dummyPrivateKeys)
		// since the function `tx.Init` takes output's paymentinfo and creates outputCoins inside of it, we only create the paymentinfo here
		paymentInfoOut := make([]*key.PaymentInfo, numOfOutputs)
		for i, _ := range paymentInfoOut {
			paymentInfoOut[i] = key.InitPaymentInfo(keySets[txInd].PaymentAddress,uint64(3000),[]byte("bench out"))
			// fmt.Println(paymentInfo[i])
		}
		// now we take some of those stored coins to use as TX input
		// for the TX to be valid, these inputs must associate to one same private key
		// (it's guaranteed by our way of indexing the pastCoins array)
		inputCoins := make([]coin.PlainCoin,numOfInputs)
		for i,_ := range inputCoins{
			inputCoins[i],_ = pastCoins[i].Decrypt(keySets[txInd])
		}

		// now we calculate the fee = sum(Input) - sum(Output)
		// sumIn := uint64(400000*numOfPrivateKeys*numOfInputs)
		// sumOut := uint64(3000*numOfOutputs)

		initializingParams := NewTxPrivacyInitParams(dummyPrivateKeys[txInd],
			paymentInfoOut,inputCoins,
			1,true,
			dummyDB,
			nil,
			nil,
			[]byte{},
		)
		// creating the TX object
		tx := &TxVersion2{}
		// actually making the TX
		// `Init` function will also create all necessary proofs and attach them to the TX
		tx.Init(initializingParams)

		txsForBenchmark = append(txsForBenchmark, tx)
	}

	b.ResetTimer()
	for loop := 0; loop < b.N; loop++ {
		chosenIndex := RandInt() % len(txsForBenchmark)
		currentTx := txsForBenchmark[chosenIndex]
		// verify the TX
		// params : hasPrivacy bool, transactionStateDB *statedb.StateDB, bridgeStateDB *statedb.StateDB, 
		// 			shardID byte (we're testing with only 1 shard), 
		//			tokenID *common.Hash (set to nil, meaning we use PRV),
		//			isBatch bool, isNewTransaction bool
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

func BenchmarkTxV2BatchVerify(b *testing.B){
	rand.Seed(time.Now().UnixNano())
	// fmt.Println(os.Args[5:])
	clargs := os.Args[5:]
	// fmt.Println(clargs)

	numOfInputs,_ := strconv.Atoi(clargs[0])
	numOfOutputs,_ := strconv.Atoi(clargs[1])
	// our setup will cause an extra 'change' output coin to be added so we fix here
	numOfOutputs -= 1
	numOfPrivateKeys := 50
	// fmt.Printf("\n------------------TxVersion2 Verify Benchmark\n")
	// fmt.Printf("Number of transactions : %d\n", 1)
	// fmt.Printf("Number of inputs       : %d\n", numOfInputs)
	// fmt.Printf("Number of outputs      : %d\n", numOfOutputs)
	preparePaymentKeys(numOfPrivateKeys,nil)
	numOfTxs := numOfPrivateKeys
	// dummyDB, _ := statedb.NewWithPrefixTrie(emptyRoot, warperDBStatedbTest)
	
	var txsForBenchmark []*TxVersion2
	for txInd:=0;txInd<numOfTxs;txInd++{ 
		// pastCoins are coins we forcefully write into the dummyDB to simulate the db having OTAs in the past
		// we make sure there are a lot - and a lot - of past coins from all those simulated private keys
		pastCoins := make([]coin.Coin, numOfInputs)
		for i, _ := range pastCoins {
			tempCoin,_ := coin.NewCoinFromPaymentInfo(paymentInfo[txInd])

			// to obtain a PlainCoin to feed into input of TX, we need to conceal & decrypt it (it makes sure all fields are right, as opposed to just casting the type to PlainCoin)
			tempCoin.ConcealOutputCoin(keySets[txInd].PaymentAddress.GetPublicView())
			pastCoins[i] = tempCoin
		}
		// use the db's interface to write our simulated pastCoins to the database
		// we do need to re-format the data into bytes first
		forceSaveCoins(dummyDB, pastCoins, 0, common.PRVCoinID, nil)

		// in this test, we randomize the length of inputCoins so we feel safe fixing the length of outputCoins to equal len(dummyPrivateKeys)
		// since the function `tx.Init` takes output's paymentinfo and creates outputCoins inside of it, we only create the paymentinfo here
		paymentInfoOut := make([]*key.PaymentInfo, numOfOutputs)
		for i, _ := range paymentInfoOut {
			paymentInfoOut[i] = key.InitPaymentInfo(keySets[txInd].PaymentAddress,uint64(3000),[]byte("bench out"))
			// fmt.Println(paymentInfo[i])
		}
		// now we take some of those stored coins to use as TX input
		// for the TX to be valid, these inputs must associate to one same private key
		// (it's guaranteed by our way of indexing the pastCoins array)
		inputCoins := make([]coin.PlainCoin,numOfInputs)
		for i,_ := range inputCoins{
			inputCoins[i],_ = pastCoins[i].Decrypt(keySets[txInd])
		}

		// now we calculate the fee = sum(Input) - sum(Output)
		// sumIn := uint64(400000*numOfPrivateKeys*numOfInputs)
		// sumOut := uint64(3000*numOfOutputs)

		initializingParams := NewTxPrivacyInitParams(dummyPrivateKeys[txInd],
			paymentInfoOut,inputCoins,
			1,true,
			dummyDB,
			nil,
			nil,
			[]byte{},
		)
		// creating the TX object
		tx := &TxVersion2{}
		// actually making the TX
		// `Init` function will also create all necessary proofs and attach them to the TX
		tx.Init(initializingParams)

		txsForBenchmark = append(txsForBenchmark, tx)
	}

	batchLength, _ := strconv.Atoi(clargs[2])
	// each loop verifies 20 transactions as one batch
	// so the ops/sec will need to be divided by 20 afterwards
	// for fair comparison
	b.ResetTimer()
	var pass bool
	for loop := 0; loop < b.N; loop++ {
		var batchContent []metadata.Transaction
		chosenIndex := RandInt() % len(txsForBenchmark)
		for j:=0;j<batchLength;j++{
			chosenIndex := (chosenIndex+1)%len(txsForBenchmark)
			currentTx := txsForBenchmark[chosenIndex]
			currentTx.ValidateSanityData(nil,nil,nil,0)
			currentTx.ValidateTxWithBlockChain(nil, nil, nil, shardID, dummyDB)
			
			batchContent = append(batchContent, currentTx)
		}
		batch := NewBatchTransaction(batchContent)
		success, _, _ := batch.Validate(dummyDB, nil)
		if !success{
			fmt.Println("Something wrong")
			panic("Invalid tx batch")
		}
		pass = true
	}
	assert.Equal(b,true,pass)
}