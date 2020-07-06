package transaction

import (
	"bytes"
	"math/big"
	"testing"
	"fmt"
	"io/ioutil"
	"os"
	// "encoding/json"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	// "github.com/incognitochain/incognito-chain/privacy/operation"
	"github.com/incognitochain/incognito-chain/privacy/key"
	"github.com/incognitochain/incognito-chain/privacy/coin"
	"github.com/incognitochain/incognito-chain/trie"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/stretchr/testify/assert"
)

var (
	// num of private keys
	maxPrivateKeys = 15
	minPrivateKeys = 6

	maxInputs = 10
	minInputs = 1
	numOfLoops = 1
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

	dummyPrivateKeys []*key.PrivateKey
	keySets []*incognitokey.KeySet
	paymentInfo []*key.PaymentInfo
)

var _ = func() (_ struct{}) {
// initialize a `test` db in the OS's tempdir
// and with it, a db access wrapper that reads/writes our transactions
	fmt.Println("This runs before init()!")
	Logger.Init(common.NewBackend(nil).Logger("test", true))
	dbPath, err := ioutil.TempDir(os.TempDir(), "test_statedb_")
	if err != nil {
		panic(err)
	}
	diskBD, _ := incdb.Open("leveldb", dbPath)
	warperDBStatedbTest = statedb.NewDatabaseAccessWarper(diskBD)
	trie.Logger.Init(common.NewBackend(nil).Logger("test", true))
	return
}()

func forceSaveCoins(db *statedb.StateDB, coinsToBeSaved []coin.Coin, shardID byte, tokenID common.Hash, t *testing.T){
	coinsInBytes := make([][]byte, 0)
	otas := make([][]byte, 0)
	for _,c := range coinsToBeSaved{
		assert.Equal(t,2,int(c.GetVersion()))
		coinsInBytes = append(coinsInBytes, c.Bytes())
		otas = append(otas, c.GetPublicKey().ToBytesS())
	}
	err := statedb.StoreOTACoinsAndOnetimeAddresses(db, tokenID, 0, coinsInBytes, otas, shardID)
	assert.Equal(t,nil,err)
}

func preparePaymentKeys(count int, t *testing.T){
	dummyPrivateKeys = make([]*key.PrivateKey,count)
	for i := 0; i < count; i += 1 {
		privateKey := key.GeneratePrivateKey(common.RandBytes(32))
		dummyPrivateKeys[i] = &privateKey
	}

// then use each privatekey to derive Incognito keyset (various keys for everything inside the protocol)
	keySets = make([]*incognitokey.KeySet,len(dummyPrivateKeys))
// PaymentInfo is like `intent` for making Coin.
// the paymentInfo slice here will be used to create pastCoins & inputCoins 
// we populate `value` fields with some arbitrary, big-enough constant (here, 4000*len)
// `message` field can be anything
	paymentInfo = make([]*key.PaymentInfo, len(dummyPrivateKeys))
	for i,_ := range keySets{
		keySets[i] = new(incognitokey.KeySet)
		err := keySets[i].InitFromPrivateKey(dummyPrivateKeys[i])
		assert.Equal(t, nil, err)

		paymentInfo[i] = key.InitPaymentInfo(keySets[i].PaymentAddress, uint64(4000*len(dummyPrivateKeys)), []byte("test in"))
	}
}

func TestSigPubKeyCreationAndMarshalling(t *testing.T) {
// here m, n are not very specific so we give them generous range
	m := common.RandInt() % (maxPrivateKeys - minInputs + 1) + minInputs
	n := common.RandInt() % (maxPrivateKeys - minInputs + 1) + minInputs
	var err error
	for i := 0; i < numOfLoops; i += 1 {
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
	dummyDB, _ := statedb.NewWithPrefixTrie(emptyRoot, warperDBStatedbTest)
	for loop := 0; loop < numOfLoops; loop++ {
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
		tx := TxVersion2{}
// actually making the salary TX
		err = tx.InitTxSalary(theCoins[0], dummyPrivateKeys[0], dummyDB, nil)

		isValid,err := tx.ValidateTxSalary(dummyDB)
		assert.Equal(t,nil,err)
		assert.Equal(t,true,isValid)

// this other coin is already in db so it must be rejected
		err = tx.InitTxSalary(theCoins[1], dummyPrivateKeys[0], dummyDB, nil)
		assert.NotEqual(t,nil,err)
	}
}

func TestTxV2ProveWithPrivacy(t *testing.T){
	numOfPrivateKeys := common.RandInt() % (maxPrivateKeys - minPrivateKeys + 1) + minPrivateKeys
	numOfInputs := common.RandInt() % (maxInputs - minInputs + 1) + minInputs
	dummyDB, _ := statedb.NewWithPrefixTrie(emptyRoot, warperDBStatedbTest)
	for loop := 0; loop < numOfLoops; loop++ {
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

// verify the TX
// params : hasPrivacy bool, transactionStateDB *statedb.StateDB, bridgeStateDB *statedb.StateDB, 
// 			shardID byte (we're testing with only 1 shard), 
//			tokenID *common.Hash (set to nil, meaning we use PRV),
//			isBatch bool, isNewTransaction bool
		isValid,err := tx.ValidateSanityData(nil,nil,nil,0)
		assert.Equal(t,nil,err)
		assert.Equal(t,true,isValid)
		isValid,err = tx.ValidateTransaction(true,dummyDB,nil,0,nil,false,true)
		assert.Equal(t,nil,err)
		assert.Equal(t,true,isValid)

		testTxV2DeletedProof(tx, t)
		testTxV2InvalidFee(dummyDB, inputCoins, paymentInfoOut, t)
		testTxV2OneFakeInput(tx, dummyDB, inputCoins, paymentInfoOut, pastCoins, t)
		testTxV2OneFakeOutput(tx, dummyDB, inputCoins, paymentInfoOut, pastCoins, t)
		testTxV2OneDoubleSpentInput(dummyDB, inputCoins, paymentInfoOut, pastCoins, t)
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
	isValid,errMalVerify := malTx.ValidateTransaction(true,db,nil,0,nil,false,true)
	assert.NotEqual(t,nil,errMalVerify)
	assert.Equal(t,false,isValid)
}

func testTxV2OneFakeInput(txv2 *TxVersion2, db *statedb.StateDB, inputCoins []coin.PlainCoin, paymentInfoOut []*key.PaymentInfo, pastCoins []coin.Coin, t *testing.T){
	// likewise, if someone took an already proven tx and swaps one input coin 
	// for another random coin from outside, the tx cannot go through
	// (here we only meddle with coin-changing - not adding/removing - since length checks are included within mlsag)
	numOfInputs := len(inputCoins)
	changed := common.RandInt() % numOfInputs
	saved := inputCoins[changed]
	inputCoins[changed],_ = pastCoins[len(dummyPrivateKeys)*(numOfInputs+1)].Decrypt(keySets[0])
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
	malTx.SetProof(txv2.GetProof())
	isValid,err := malTx.ValidateTransaction(true,db,nil,0,nil,false,true)
	// verify must fail
	assert.NotEqual(t,nil,err)
	assert.Equal(t,false,isValid)
	inputCoins[changed] = saved
}

func testTxV2OneFakeOutput(txv2 *TxVersion2, db *statedb.StateDB, inputCoins []coin.PlainCoin, paymentInfoOut []*key.PaymentInfo, pastCoins []coin.Coin, t *testing.T){
	// similar to the above. All these verifications should fail
		changed := common.RandInt() % len(dummyPrivateKeys)
		savedPay := paymentInfoOut[changed]
		paymentInfoOut[changed] = key.InitPaymentInfo(keySets[len(dummyPrivateKeys)-1].PaymentAddress,uint64(1400),[]byte("test out mal"))
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
		malTx.SetProof(txv2.GetProof())
		isValid,err := malTx.ValidateTransaction(true,db,nil,0,nil,false,true)
		// verify must fail
		assert.NotEqual(t,nil,err)
		assert.Equal(t,false,isValid)
		paymentInfoOut[changed] = savedPay
}

func testTxV2OneDoubleSpentInput(db *statedb.StateDB, inputCoins []coin.PlainCoin, paymentInfoOut []*key.PaymentInfo, pastCoins []coin.Coin, t *testing.T){
	// similar to the above. All these verifications should fail
		changed := common.RandInt() % len(inputCoins)
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
		isValid,err := malTx.ValidateTransaction(true,db,nil,0,nil,false,true)
		// verify by itself passes
		assert.Equal(t,nil,err)
		assert.Equal(t,true,isValid)

		// verify with blockchain fails
		err = malTx.ValidateTxWithBlockChain(nil, nil ,nil, 0, db)
		assert.NotEqual(t,nil,err)
		
}