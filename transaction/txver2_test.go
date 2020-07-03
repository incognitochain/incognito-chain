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
	"github.com/incognitochain/incognito-chain/privacy/operation"
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
	minPrivateKeys = 8

	maxInputs = 7
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

func TestSigPubKeyCreationAndMarshalling(t *testing.T) {
// here m, n are not very specific so we give them generous range
	m := common.RandInt() % (maxPrivateKeys - minInputs + 1) + minInputs
	n := common.RandInt() % (maxPrivateKeys - minInputs + 1) + minInputs
	var err error
	for i := 0; i < numOfLoops; i += 1 {

		dummyPrivateKeys := []*operation.Scalar{}
		for i := 0; i < m; i += 1 {
			privateKey := operation.RandomScalar()
			dummyPrivateKeys = append(dummyPrivateKeys, privateKey)
		}
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
		txSig.indexes = indexes

		b, err := txSig.Bytes()
		assert.Equal(t, nil, err, "Should not have any bug when txSig.ToBytes")

		txSig2 := new(TxSigPubKeyVer2)
		err = txSig2.SetBytes(b)
		assert.Equal(t, nil, err, "Should not have any bug when txSig.FromBytes")

		b2, err := txSig2.Bytes()
		assert.Equal(t, nil, err, "Should not have any bug when txSig2.ToBytes")
		assert.Equal(t, true, bytes.Equal(b, b2))

		n1 := len(txSig.indexes)
		m1 := len(txSig.indexes[0])
		n2 := len(txSig2.indexes)
		m2 := len(txSig2.indexes[0])

		assert.Equal(t, n1, n2, "Two indexes length should be equal")
		assert.Equal(t, m1, m2, "Two indexes length should be equal")
		for i := 0; i < n; i += 1 {
			for j := 0; j < m; j += 1 {
				b1 := txSig.indexes[i][j].Bytes()
				b2 := txSig2.indexes[i][j].Bytes()
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
// first, we randomize some dummy private keys in the form of bytes32
		dummyPrivateKeys := []*key.PrivateKey{}
		for i := 0; i < numOfPrivateKeys; i += 1 {
			privateKey := key.GeneratePrivateKey(common.RandBytes(32))
			dummyPrivateKeys = append(dummyPrivateKeys, &privateKey)
		}

// then use each privatekey to derive Incognito keyset (various keys for everything inside the protocol)
		keySets := make([]*incognitokey.KeySet,len(dummyPrivateKeys))
		paymentInfo := make([]*key.PaymentInfo, len(dummyPrivateKeys))
		for i,_ := range keySets{
			keySets[i] = new(incognitokey.KeySet)
			err := keySets[i].InitFromPrivateKey(dummyPrivateKeys[i])
			assert.Equal(t, nil, err)

			paymentInfo[i] = key.InitPaymentInfo(keySets[i].PaymentAddress, uint64(12345679), []byte("test salary"))
		}
// create 2 otaCoins, the second one will already be stored in the db
		theCoins := make([]*coin.CoinV2, 2)
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
			tempCoin.ConcealData(keySets[i].PaymentAddress.GetPublicView())
			assert.Equal(t, true, tempCoin.IsEncrypted())
			assert.Equal(t, true, tempCoin.GetSharedRandom() == nil)
			_, err = tempCoin.Decrypt(keySets[i])
			assert.Equal(t,nil,err)
			theCoins[i] = tempCoin
		}
		coinsToBeSaved := make([][]byte, 0)
		otas := make([][]byte, 0)
		assert.Equal(t,2,int(theCoins[1].GetVersion()))
		coinsToBeSaved = append(coinsToBeSaved, theCoins[1].Bytes())
		otas = append(otas, theCoins[1].GetPublicKey().ToBytesS())
		err := statedb.StoreOTACoinsAndOnetimeAddresses(dummyDB, common.PRVCoinID, 0, coinsToBeSaved, otas, 0)

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
// first, we randomize some dummy private keys in the form of bytes32
		dummyPrivateKeys := []*key.PrivateKey{}
		for i := 0; i < numOfPrivateKeys; i += 1 {
			privateKey := key.GeneratePrivateKey(common.RandBytes(32))
			dummyPrivateKeys = append(dummyPrivateKeys, &privateKey)
		}

// then use each privatekey to derive Incognito keyset (various keys for everything inside the protocol)
		keySets := make([]*incognitokey.KeySet,len(dummyPrivateKeys))
// PaymentInfo is like `intent` for making Coin.
// the paymentInfo slice here will be used to create pastCoins & inputCoins 
// we populate `value` fields with some arbitrary, big-enough constant (here, 4000*len)
// `message` field can be anything
		paymentInfo := make([]*key.PaymentInfo, len(dummyPrivateKeys))
		for i,_ := range keySets{
			keySets[i] = new(incognitokey.KeySet)
			err := keySets[i].InitFromPrivateKey(dummyPrivateKeys[i])
			assert.Equal(t, nil, err)

			paymentInfo[i] = key.InitPaymentInfo(keySets[i].PaymentAddress, uint64(4000*len(dummyPrivateKeys)), []byte("test in"))
		}		

		// for i, _ := range dummyPrivateKeys {
		// 	senderPrivateKeyBytes := []byte(*dummyPrivateKeys[0])
		// 	paymentInfo[i] = key.InitPaymentInfo(key.GeneratePaymentAddress(senderPrivateKeyBytes[:]),uint64(3000+1000*i),[]byte("test in"))
		// 	// fmt.Println(paymentInfo[i])
		// }
// pastCoins are coins we forcefully write into the dummyDB to simulate the db having OTAs in the past
// we make sure there are a lot - and a lot - of past coins from all those simulated private keys
		pastCoins := make([]coin.PlainCoin, (10+numOfInputs)*len(dummyPrivateKeys))
		for i, _ := range pastCoins {
			tempCoin,err := coin.NewCoinFromPaymentInfo(paymentInfo[i%len(dummyPrivateKeys)])
			assert.Equal(t, nil, err)
			assert.Equal(t, false, tempCoin.IsEncrypted())

// to obtain a PlainCoin to feed into input of TX, we need to conceal & decrypt it (it makes sure all fields are right, as opposed to just casting the type to PlainCoin)
			tempCoin.ConcealData(keySets[i%len(dummyPrivateKeys)].PaymentAddress.GetPublicView())
			assert.Equal(t, true, tempCoin.IsEncrypted())
			assert.Equal(t, true, tempCoin.GetSharedRandom() == nil)
			pastCoins[i], err = tempCoin.Decrypt(keySets[i%len(dummyPrivateKeys)])
			assert.Equal(t, nil, err)
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
		coinsToBeSaved := make([][]byte, 0)
		otas := make([][]byte, 0)
		for _, inputCoin := range pastCoins {
			if inputCoin.GetVersion() != 2 {
				continue
			}
			coinsToBeSaved = append(coinsToBeSaved, inputCoin.Bytes())
			otas = append(otas, inputCoin.GetPublicKey().ToBytesS())
		}
		err := statedb.StoreOTACoinsAndOnetimeAddresses(dummyDB, common.PRVCoinID, 0, coinsToBeSaved, otas, 0)

// now we take some of those stored coins to use as TX input
// for the TX to be valid, these inputs must associate to one same private key
// (it's guaranteed by our way of indexing the pastCoins array)
		inputCoins := make([]coin.PlainCoin,numOfInputs)
		for i,_ := range inputCoins{
			inputCoins[i] = pastCoins[i*len(dummyPrivateKeys)]
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
		// jsonBytes,err := json.Marshal(tx)

		// thatTx := &TxVersion2{}
		// json.Unmarshal(jsonBytes,thatTx)
		// thatJsonBytes, err := json.Marshal(thatTx)

		// assert.Equal(t,true,bytes.Equal(jsonBytes,thatJsonBytes))
		// x := tx.GetProof().GetOutputCoins()

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

// first, try setting the proof to nil, then verify
// it should not go through
		tx_specific := tx
		savedProof := tx_specific.Proof
		tx_specific.Proof = nil
		isValid,err = tx.ValidateSanityData(nil,nil,nil,0)
		// fmt.Printf("Nullified the proof, see what happens\n%v\n%v",isValid,err)
		assert.NotEqual(t,nil,err)
		assert.Equal(t,false,isValid)
		tx_specific.Proof = savedProof

		// tx_specific,ok := tx.(*TxVersion2)
		// assert.Equal(t,true,ok)
		// v,err := json.Marshal(tx_specific)
		// fmt.Printf("TX is %s\n",string(v))
		
		// err = json.Unmarshal(v,tx_specific)
		// v,err = json.Marshal(tx_specific)
		// fmt.Printf("Now TX is %s\n",string(v))

		// a set of init params where sum(Input) < fee + sum(Output)
		// let's say someone tried to use this invalid fee for tx
		// we should encounter an error here
		malFeeParams := NewTxPrivacyInitParams(dummyPrivateKeys[0],
			paymentInfoOut,inputCoins,
			sumIn-sumOut+1111,true,
			dummyDB,
			nil,
			nil,
			[]byte{},
		)
		malTx := &TxVersion2{}
		errMalInit := malTx.Init(malFeeParams)
		assert.NotEqual(t,nil,errMalInit)
		isValid,errMalVerify := malTx.ValidateTransaction(true,dummyDB,nil,0,nil,false,true)
		assert.NotEqual(t,nil,errMalVerify)

// likewise, if someone took an already proven tx and swaps one input coin 
// for another random coin from outside, the tx cannot go through
// (here we only meddle with coin-changing - not adding/removing - since length checks are included within mlsag)
		changed := common.RandInt() % numOfInputs
		saved := inputCoins[changed]
		inputCoins[changed] = pastCoins[len(dummyPrivateKeys)*(numOfInputs+1)]
		malInputParams := NewTxPrivacyInitParams(dummyPrivateKeys[0],
			paymentInfoOut,inputCoins,
			1,true,
			dummyDB,
			nil,
			nil,
			[]byte{},
		)
		malTx = &TxVersion2{}
		err = malTx.Init(malInputParams)
		assert.Equal(t,nil,err)
		malTx.SetProof(tx.GetProof())
		isValid,err = malTx.ValidateTransaction(true,dummyDB,nil,0,nil,false,true)
		// verify must fail
		assert.NotEqual(t,nil,err)
		inputCoins[changed] = saved

// we try the same for output. All these verifications should fail
		changed = common.RandInt() % len(dummyPrivateKeys)
		savedPay := paymentInfoOut[changed]
		paymentInfoOut[changed] = key.InitPaymentInfo(keySets[len(dummyPrivateKeys)-1].PaymentAddress,uint64(1400),[]byte("test out mal"))
		malInputParams = NewTxPrivacyInitParams(dummyPrivateKeys[0],
			paymentInfoOut,inputCoins,
			1,true,
			dummyDB,
			nil,
			nil,
			[]byte{},
		)
		malTx = &TxVersion2{}
		err = malTx.Init(malInputParams)
		assert.Equal(t,nil,err)
		malTx.SetProof(tx.GetProof())
		isValid,err = malTx.ValidateTransaction(true,dummyDB,nil,0,nil,false,true)
		// verify must fail
		assert.NotEqual(t,nil,err)
		paymentInfoOut[changed] = savedPay

	}
}