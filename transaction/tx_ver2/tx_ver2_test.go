package tx_ver2

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"math/rand"
	"os"
	"testing"
	"unicode"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/privacy/coin"
	"github.com/incognitochain/incognito-chain/privacy/key"
	"github.com/incognitochain/incognito-chain/privacy/operation"
	"github.com/incognitochain/incognito-chain/privacy/privacy_v2/bulletproofs"
	"github.com/incognitochain/incognito-chain/transaction/tx_generic"
	"github.com/incognitochain/incognito-chain/transaction/utils"
	"github.com/incognitochain/incognito-chain/trie"
	. "github.com/smartystreets/goconvey/convey"
	// "github.com/stretchr/testify/assert"
)

var (
	// num of private keys
	maxPrivateKeys = 10
	minPrivateKeys = 1
	maxInputs      = 10
	minInputs      = 1
	maxTries       = 100

	allowModifiedTXsToPass      = false
	hasPrivacyForPRV       bool = true
	hasPrivacyForToken     bool = true
	shardID                byte = byte(0)

	positiveTestsFileName = "./testdata/accepted.txt"
	negativeTestsFileName = "./testdata/rejected.txt"
	b58                   = base58.Base58Check{}
)

// variables for initializing stateDB for test
var (
	warperDBStatedbTest statedb.DatabaseAccessWarper
	emptyRoot           = common.HexToHash(common.HexEmptyRoot)
	dummyDB             *statedb.StateDB
	bridgeDB            *statedb.StateDB
	logger              common.Logger
)

func init() {
	fmt.Println("Initializing")
	// initialize a `test` db in the OS's tempdir
	// and with it, a db access wrapper that reads/writes our transactions
	common.MaxShardNumber = 1
	testLogFile, _ := os.OpenFile("test-log.txt", os.O_RDWR|os.O_TRUNC|os.O_CREATE, 0755)
	logger = common.NewBackend(testLogFile).Logger("test", false)
	logger.SetLevel(common.LevelDebug)
	privacy.LoggerV1.Init(logger)
	privacy.LoggerV2.Init(logger)
	// can switch between the 2 loggers to mute logs as one wishes
	utils.Logger.Init(logger)
	bulletproofs.Logger.Init(common.NewBackend(nil).Logger("test", true))
	dbPath, _ := ioutil.TempDir(os.TempDir(), "test_statedb_")
	diskBD, _ := incdb.Open("leveldb", dbPath)
	warperDBStatedbTest = statedb.NewDatabaseAccessWarper(diskBD)
	dummyDB, _ = statedb.NewWithPrefixTrie(emptyRoot, warperDBStatedbTest)
	bridgeDB = dummyDB.Copy()
	trie.Logger.Init(common.NewBackend(nil).Logger("test", true))
}

func storeCoins(db *statedb.StateDB, coinsToBeSaved []coin.Coin, shardID byte, tokenID common.Hash) error {
	coinsInBytes := make([][]byte, 0)
	otas := make([][]byte, 0)
	for _, c := range coinsToBeSaved {
		So(int(c.GetVersion()), ShouldEqual, 2)
		coinsInBytes = append(coinsInBytes, c.Bytes())
		otas = append(otas, c.GetPublicKey().ToBytesS())
	}
	return statedb.StoreOTACoinsAndOnetimeAddresses(db, tokenID, 0, coinsInBytes, otas, shardID)
}

func preparePaymentKeys(count int) ([]*privacy.PrivateKey, []*incognitokey.KeySet, []*key.PaymentInfo) {
	// create many random private keys
	// then use each privatekey to derive Incognito keyset (various keys for everything inside the protocol)
	// we ensure they all belong in shard 0 for this test

	// PaymentInfo is like `intent` for making Coin.
	// the paymentInfo slice here will be used to create pastCoins & inputCoins
	// we populate `value` fields with some arbitrary, big-enough constant (here, 4000*len)
	// `message` field can be anything
	dummyPrivateKeys := make([]*privacy.PrivateKey, count)
	keySets := make([]*incognitokey.KeySet, len(dummyPrivateKeys))
	paymentInfo := make([]*key.PaymentInfo, len(dummyPrivateKeys))
	for i := 0; i < count; i += 1 {
		for {
			privateKey := key.GeneratePrivateKey(common.RandBytes(32))
			dummyPrivateKeys[i] = &privateKey
			keySets[i] = new(incognitokey.KeySet)
			err := keySets[i].InitFromPrivateKey(dummyPrivateKeys[i])
			paymentInfo[i] = key.InitPaymentInfo(keySets[i].PaymentAddress, uint64(4000*len(dummyPrivateKeys)), []byte("test in"))

			pkb := []byte(paymentInfo[i].PaymentAddress.Pk)
			if common.GetShardIDFromLastByte(pkb[len(pkb)-1]) == shardID {
				So(err, ShouldBeNil)
				break
			}
		}
	}
	return dummyPrivateKeys, keySets, paymentInfo
}

func TestSigPubKeyCreationAndMarshalling(t *testing.T) {
	Convey("Tx - Public Key Marshalling Test", t, func() {
		// here m, n are not very specific so we give them generous range
		m := RandInt()%(maxPrivateKeys-minInputs+1) + minInputs
		n := RandInt()%(maxPrivateKeys-minInputs+1) + minInputs
		var err error
		maxLen := new(big.Int)
		maxLen.SetString("1000000000000000000", 10)
		indexes := make([][]*big.Int, n)

		for i := 0; i < n; i += 1 {
			row := make([]*big.Int, m)
			for j := 0; j < m; j += 1 {
				row[j], err = common.RandBigIntMaxRange(maxLen)
				So(err, ShouldBeNil)
			}
			indexes[i] = row
		}
		txSig := new(SigPubKey)
		txSig.Indexes = indexes
		b, err := txSig.Bytes()
		Convey("txSig.ToBytes", func() {
			So(err, ShouldBeNil)
		})

		txSig2 := new(SigPubKey)
		err = txSig2.SetBytes(b)
		Convey("txSig.FromBytes", func() {
			So(err, ShouldBeNil)
		})

		b2, err := txSig2.Bytes()
		Convey("txSig.ToBytes again", func() {
			So(err, ShouldBeNil)
			So(bytes.Equal(b, b2), ShouldBeTrue)
		})

		n1 := len(txSig.Indexes)
		m1 := len(txSig.Indexes[0])
		n2 := len(txSig2.Indexes)
		m2 := len(txSig2.Indexes[0])
		Convey("dimensions should match", func() {
			So(n1, ShouldEqual, n2)
			So(m1, ShouldEqual, m2)

		})
		Convey("elements should match", func() {
			for i := 0; i < n; i += 1 {
				for j := 0; j < m; j += 1 {
					b1 := txSig.Indexes[i][j].Bytes()
					b2 := txSig2.Indexes[i][j].Bytes()
					So(bytes.Equal(b1, b2), ShouldBeTrue)
				}
			}
		})
	})
}

func TestTxV2Salary(t *testing.T) {
	var numOfPrivateKeys int
	theCoins := make([]*coin.CoinV2, 2)
	theCoinsGeneric := make([]coin.Coin, 2)
	var dummyPrivateKeys []*privacy.PrivateKey
	var keySets []*incognitokey.KeySet
	var paymentInfo []*privacy.PaymentInfo
	tx := &Tx{}

	Convey("Tx Salary Test", t, func() {
		numOfPrivateKeys = RandInt()%(maxPrivateKeys-minPrivateKeys+1) + minPrivateKeys
		Convey("prepare keys", func() {
			dummyPrivateKeys, keySets, paymentInfo = preparePaymentKeys(numOfPrivateKeys)
		})
		Convey("create salary coins", func() {
			// create 2 otaCoins, the second one will already be stored in the db
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
				So(err, ShouldBeNil)
				So(tempCoin.IsEncrypted(), ShouldBeFalse)
				tempCoin.ConcealOutputCoin(keySets[i].PaymentAddress.GetPublicView())
				So(tempCoin.IsEncrypted(), ShouldBeTrue)
				So(tempCoin.GetSharedRandom() == nil, ShouldBeTrue)
				_, err = tempCoin.Decrypt(keySets[i])
				So(err, ShouldBeNil)
				theCoins[i] = tempCoin
				theCoinsGeneric[i] = tempCoin
			}
			So(storeCoins(dummyDB, []coin.Coin{theCoinsGeneric[1]}, 0, common.PRVCoinID), ShouldBeNil)
		})

		Convey("create salary TX", func() {
			// actually making the salary TX
			err := tx.InitTxSalary(theCoins[0], dummyPrivateKeys[0], dummyDB, nil)
			So(err, ShouldBeNil)
		})
		Convey("verify salary TX", func() {
			isValid, err := tx.ValidateTxSalary(dummyDB)
			So(err, ShouldBeNil)
			So(isValid, ShouldBeTrue)
			testTxV2JsonMarshaler(tx, 10, dummyDB)
			malTx := &Tx{}
			// this other coin is already in db so it must be rejected
			err = malTx.InitTxSalary(theCoins[1], dummyPrivateKeys[0], dummyDB, nil)
			So(err, ShouldNotBeNil)
		})
	})
}

func TestPrivacyV2TxPRV(t *testing.T) {
	var numOfPrivateKeys int
	var numOfInputs int
	tx := &Tx{}
	var dummyPrivateKeys []*privacy.PrivateKey
	var keySets []*incognitokey.KeySet
	var paymentInfo []*privacy.PaymentInfo
	var pastCoins []coin.Coin
	var paymentInfoOut []*privacy.PaymentInfo
	var inputCoins []coin.PlainCoin
	var initializingParams *tx_generic.TxPrivacyInitParams

	Convey("Tx PRV Main Test", t, func() {
		numOfPrivateKeys = RandInt()%(maxPrivateKeys-minPrivateKeys+1) + minPrivateKeys
		numOfInputs = RandInt()%(maxInputs-minInputs+1) + minInputs
		Convey("prepare keys", func(){
			dummyPrivateKeys, keySets, paymentInfo = preparePaymentKeys(numOfPrivateKeys)
		})

		Convey("create & store UTXOs", func() {
			// pastCoins are coins we manually store in the dummyDB to simulate the db having OTAs from chaindata
			pastCoins = make([]coin.Coin, (10+numOfInputs)*len(dummyPrivateKeys))
			for i, _ := range pastCoins {
				tempCoin, err := coin.NewCoinFromPaymentInfo(paymentInfo[i%len(dummyPrivateKeys)])
				So(err, ShouldBeNil)
				So(tempCoin.IsEncrypted(), ShouldBeFalse)

				// to obtain a PlainCoin to feed into input of TX, we need to conceal & decrypt it (it makes sure all fields are right, as opposed to just casting the type to PlainCoin)
				tempCoin.ConcealOutputCoin(keySets[i%len(dummyPrivateKeys)].PaymentAddress.GetPublicView())
				So(tempCoin.IsEncrypted(), ShouldBeTrue)
				So(tempCoin.GetSharedRandom() == nil, ShouldBeTrue)
				pastCoins[i] = tempCoin
			}
			// use the db's interface to write our simulated pastCoins to the database
			So(storeCoins(dummyDB, pastCoins, 0, common.PRVCoinID), ShouldBeNil)
		})

		Convey("prepare payment info", func() {
			// in this test, we randomize the length of inputCoins & fix the length of outputCoins to len(dummyPrivateKeys)
			paymentInfoOut = make([]*privacy.PaymentInfo, len(dummyPrivateKeys))
			for i, _ := range dummyPrivateKeys {
				paymentInfoOut[i] = key.InitPaymentInfo(keySets[i].PaymentAddress, uint64(3000), []byte("test out"))
			}
		})

		Convey("decrypt inputs", func() {
			// now we take some of those stored coins to use as TX input
			// for the TX to be valid, these inputs must associate to one same private key
			// (it's guaranteed by our way of indexing the pastCoins array)
			inputCoins = make([]coin.PlainCoin, numOfInputs)
			for i, _ := range inputCoins {
				var err error
				inputCoins[i], err = pastCoins[i*len(dummyPrivateKeys)].Decrypt(keySets[0])
				So(err, ShouldBeNil)
			}
		})

		Convey("create TX params", func() {
			// now we calculate the fee = sum(Input) - sum(Output)
			sumIn := uint64(4000 * len(dummyPrivateKeys) * numOfInputs)
			sumOut := uint64(3000 * len(dummyPrivateKeys))
			var fee uint64 = 100
			initializingParams = tx_generic.NewTxPrivacyInitParams(dummyPrivateKeys[0],
				paymentInfoOut, inputCoins,
				fee, hasPrivacyForPRV,
				dummyDB,
				&common.PRVCoinID,
				nil,
				[]byte{},
			)
			So(sumIn >= sumOut, ShouldBeTrue)
		})

		Convey("create transaction", func() {
			// actually making the TX
			// `Init` function will also create all necessary proofs and attach them to the TX
			err := tx.Init(initializingParams)
			if err != nil {
				panic(err)
			}
			So(err, ShouldBeNil)
		})

		Convey("should verify & accept transaction", func() {
			// verify the TX
			isValid, err := tx.ValidateSanityData(nil, nil, nil, 0)
			So(err, ShouldBeNil)
			So(isValid, ShouldBeTrue)

			boolParams := make(map[string]bool)
			boolParams["hasPrivacy"] = hasPrivacyForPRV
			boolParams["isNewTransaction"] = true
			// isValid,err = tx.ValidateTransaction(true,dummyDB,nil,0,nil,false,true)
			isValid, err = tx.ValidateTxByItself(boolParams, dummyDB, nil, nil, shardID, nil, nil)
			if err != nil {
				panic(err)
			}
			So(err, ShouldBeNil)
			So(isValid, ShouldBeTrue)
			err = tx.ValidateTxWithBlockChain(nil, nil, nil, shardID, dummyDB)
			if err != nil {
				panic(err)
			}
		})

		Convey("should reject tampered TXs", func() {
			// first, test the json marshaller
			testTxV2JsonMarshaler(tx, 10, dummyDB)
			// then apply some TX tampering templates
			// testTxV2DeletedProof(tx)
			testTxV2DuplicateInput(dummyDB, dummyPrivateKeys, inputCoins, paymentInfoOut)
			testTxV2InvalidFee(dummyDB, dummyPrivateKeys, inputCoins, paymentInfoOut)
			testTxV2OneFakeInput(tx, dummyPrivateKeys, keySets, dummyDB, initializingParams, pastCoins)
			testTxV2OneFakeOutput(tx, keySets, dummyDB, initializingParams, paymentInfoOut)
			testTxV2OneDoubleSpentInput(dummyPrivateKeys, dummyDB, inputCoins, paymentInfoOut, pastCoins)
		})
	})
}

func testTxV2DeletedProof(txv2 *Tx) {
	// try setting the proof to nil, then verify
	// it should not go through
	savedProof := txv2.Proof
	txv2.Proof = nil
	isValid, err := txv2.ValidateSanityData(nil, nil, nil, 0)
	So(err, ShouldNotBeNil)
	So(isValid, ShouldBeFalse)
	txv2.Proof = savedProof
}

func testTxV2DuplicateInput(db *statedb.StateDB, privateKeys []*privacy.PrivateKey, inputCoins []coin.PlainCoin, paymentInfoOut []*key.PaymentInfo) {
	dup := &coin.CoinV2{}
	dup.SetBytes(inputCoins[0].Bytes())
	// used the same coin twice in inputs
	malInputCoins := append(inputCoins, dup)
	malFeeParams := tx_generic.NewTxPrivacyInitParams(privateKeys[0],
		paymentInfoOut, malInputCoins,
		10, true,
		db,
		nil,
		nil,
		[]byte{},
	)
	malTx := &Tx{}
	errMalInit := malTx.Init(malFeeParams)
	So(errMalInit, ShouldBeNil)
	// sanity should be fine
	isValid, err := malTx.ValidateSanityData(nil, nil, nil, 0)
	So(err, ShouldBeNil)
	So(isValid, ShouldBeTrue)

	boolParams := make(map[string]bool)
	boolParams["hasPrivacy"] = hasPrivacyForPRV
	boolParams["isNewTransaction"] = true
	// validate should reject due to Verify() in PaymentProofV2
	isValid, _ = malTx.ValidateTxByItself(boolParams, db, nil, nil, byte(0), nil, nil)
	So(isValid, ShouldBeFalse)
}

func testTxV2InvalidFee(db *statedb.StateDB, privateKeys []*privacy.PrivateKey, inputCoins []coin.PlainCoin, paymentInfoOut []*key.PaymentInfo) {
	// a set of init params where sum(Input) < fee + sum(Output)
	// let's say someone tried to use this invalid fee for tx
	// we should encounter an error here
	sumIn := uint64(4000 * len(privateKeys) * len(inputCoins))
	sumOut := uint64(3000 * len(paymentInfoOut))
	So(sumIn, ShouldBeGreaterThan, sumOut)
	malFeeParams := tx_generic.NewTxPrivacyInitParams(privateKeys[0],
		paymentInfoOut, inputCoins,
		sumIn-sumOut+1111, true,
		db,
		nil,
		nil,
		[]byte{},
	)
	malTx := &Tx{}
	_ = malTx.Init(malFeeParams)

	boolParams := make(map[string]bool)
	boolParams["hasPrivacy"] = hasPrivacyForPRV
	boolParams["isNewTransaction"] = true
	isValid, errMalVerify := malTx.ValidateTxByItself(boolParams, db, nil, nil, byte(0), nil, nil)
	So(errMalVerify, ShouldNotBeNil)
	So(isValid, ShouldBeFalse)
}

func testTxV2OneFakeInput(txv2 *Tx, privateKeys []*privacy.PrivateKey, keySets []*incognitokey.KeySet, db *statedb.StateDB, params *tx_generic.TxPrivacyInitParams, pastCoins []coin.Coin) {
	// likewise, if someone took an already proven tx and swaps one input coin
	// for another random coin from outside, the tx cannot go through
	// (here we only meddle with coin-changing - not adding/removing - since length checks are included within mlsag)
	var err error
	theProof := txv2.GetProof()
	inputCoins := theProof.GetInputCoins()
	numOfInputs := len(inputCoins)
	changed := RandInt() % numOfInputs
	saved := inputCoins[changed]
	inputCoins[changed], _ = pastCoins[len(privateKeys)*(numOfInputs+1)].Decrypt(keySets[0])
	theProof.SetInputCoins(inputCoins)

	err = resignUnprovenTx(keySets, txv2, params, nil, false)
	So(err, ShouldBeNil)
	isValid, err := txv2.ValidateSanityData(nil, nil, nil, 0)
	So(err, ShouldBeNil)
	So(isValid, ShouldBeTrue)

	boolParams := make(map[string]bool)
	boolParams["hasPrivacy"] = hasPrivacyForPRV
	boolParams["isNewTransaction"] = true
	isValid, err = txv2.ValidateTxByItself(boolParams, db, nil, nil, byte(0), nil, nil)
	// should fail at signature since mlsag needs commitments from inputs
	logger.Debugf("TEST RESULT : One faked valid input -> %v", err)
	So(isValid, ShouldBeFalse)
	inputCoins[changed] = saved
	theProof.SetInputCoins(inputCoins)
	err = resignUnprovenTx(keySets, txv2, params, nil, false)
	isValid, err = txv2.ValidateSanityData(nil, nil, nil, 0)
	So(err, ShouldBeNil)
	So(isValid, ShouldBeTrue)
	isValid, err = txv2.ValidateTxByItself(boolParams, db, nil, nil, byte(0), nil, nil)
	So(err, ShouldBeNil)
	So(isValid, ShouldBeTrue)

	// saved = inputCoins[changed]
	// inputCoins[changed] = nil
	// malTx.GetProof().SetInputCoins(inputCoins)
	// isValid,err = malTx.ValidateTxByItself(hasPrivacyForPRV, db, nil, nil, byte(0), true, nil, nil)
	// // verify must fail
	// assert.NotEqual(t,nil,err)
	// assert.Equal(t,false,isValid)
	// inputCoins[changed] = saved
}

func testTxV2OneFakeOutput(txv2 *Tx, keySets []*incognitokey.KeySet, db *statedb.StateDB, params *tx_generic.TxPrivacyInitParams, paymentInfoOut []*key.PaymentInfo) {
	// similar to the above. All these verifications should fail
	var err error
	outs := txv2.GetProof().GetOutputCoins()
	prvOutput, ok := outs[0].(*coin.CoinV2)
	savedCoinBytes := prvOutput.Bytes()
	So(ok, ShouldBeTrue)
	prvOutput.Decrypt(keySets[0])
	// set amount to something wrong
	prvOutput.SetValue(6996)
	prvOutput.SetSharedRandom(operation.RandomScalar())
	prvOutput.ConcealOutputCoin(keySets[0].PaymentAddress.GetPublicView())
	err = resignUnprovenTx(keySets, txv2, params, nil, false)
	isValid := err == nil
	boolParams := make(map[string]bool)
	boolParams["hasPrivacy"] = hasPrivacyForPRV
	boolParams["isNewTransaction"] = true
	if isValid {
		isValid, _ = txv2.ValidateTxByItself(boolParams, db, nil, nil, byte(0), nil, nil)
		// verify must fail
	}
	So(isValid, ShouldBeFalse)
	// undo the tampering
	prvOutput.SetBytes(savedCoinBytes)
	outs[0] = prvOutput
	txv2.GetProof().SetOutputCoins(outs)
	err = resignUnprovenTx(keySets, txv2, params, nil, false)
	So(err, ShouldBeNil)

	isValid, _ = txv2.ValidateTxByItself(boolParams, db, nil, nil, byte(0), nil, nil)
	So(isValid, ShouldBeTrue)
	// now instead of changing amount, we change the OTA public key
	theProof := txv2.GetProof()
	outs = theProof.GetOutputCoins()
	prvOutput, ok = outs[0].(*coin.CoinV2)
	savedCoinBytes = prvOutput.Bytes()
	So(ok, ShouldBeTrue)
	payInf := paymentInfoOut[0]
	// totally fresh OTA of the same amount, meant for the same PaymentAddress
	newCoin, err := coin.NewCoinFromPaymentInfo(payInf)
	So(err, ShouldBeNil)
	newCoin.ConcealOutputCoin(keySets[0].PaymentAddress.GetPublicView())
	theProofSpecific, ok := theProof.(*privacy.ProofV2)
	theBulletProof, ok := theProofSpecific.GetAggregatedRangeProof().(*privacy.AggregatedRangeProofV2)
	cmsv := theBulletProof.GetCommitments()
	cmsv[0] = newCoin.GetCommitment()
	outs[0] = newCoin
	txv2.GetProof().SetOutputCoins(outs)
	err = resignUnprovenTx(keySets, txv2, params, nil, false)
	So(err, ShouldBeNil)
	isValid, err = txv2.ValidateSanityData(nil, nil, nil, 0)
	So(err, ShouldBeNil)
	So(isValid, ShouldBeTrue)

	isValid, err = txv2.ValidateTxByItself(boolParams, db, nil, nil, byte(0), nil, nil)
	// verify must fail
	So(isValid, ShouldBeFalse)
	// undo the tampering
	prvOutput.SetBytes(savedCoinBytes)
	outs[0] = prvOutput
	cmsv[0] = prvOutput.GetCommitment()
	txv2.GetProof().SetOutputCoins(outs)
	err = resignUnprovenTx(keySets, txv2, params, nil, false)
	So(err, ShouldBeNil)
	isValid, _ = txv2.ValidateTxByItself(boolParams, db, nil, nil, byte(0), nil, nil)
	So(isValid, ShouldBeTrue)
}

func testTxV2OneDoubleSpentInput(privateKeys []*privacy.PrivateKey, db *statedb.StateDB, inputCoins []coin.PlainCoin, paymentInfoOut []*key.PaymentInfo, pastCoins []coin.Coin) {
	// similar to the above. All these verifications should fail
	changed := RandInt() % len(inputCoins)
	malInputParams := tx_generic.NewTxPrivacyInitParams(privateKeys[0],
		paymentInfoOut, inputCoins,
		1, true,
		db,
		nil,
		nil,
		[]byte{},
	)
	malTx := &Tx{}
	err := malTx.Init(malInputParams)
	So(err, ShouldBeNil)
	otaBytes := malTx.GetProof().GetInputCoins()[changed].GetKeyImage().ToBytesS()
	statedb.StoreSerialNumbers(db, common.ConfidentialAssetID, [][]byte{otaBytes}, 0)

	boolParams := make(map[string]bool)
	boolParams["hasPrivacy"] = hasPrivacyForPRV
	boolParams["isBatch"] = true
	boolParams["isNewTransaction"] = true
	isValid, err := malTx.ValidateTxByItself(boolParams, db, nil, nil, byte(0), nil, nil)
	// verify by itself passes
	if err != nil {
		panic(err)
	}
	So(err, ShouldBeNil)
	So(isValid, ShouldBeTrue)

	// verify with blockchain fails
	err = malTx.ValidateTxWithBlockChain(nil, nil, nil, 0, db)
	So(err, ShouldNotBeNil)

}

func testTxV2JsonMarshaler(tx *Tx, count int, db *statedb.StateDB) {
	boolParams := make(map[string]bool)
	boolParams["hasPrivacy"] = hasPrivacyForPRV
	var payloadTx *Tx
	defer func() {
		if r := recover(); r != nil {
			jsb, _ := json.Marshal(payloadTx)
			fmt.Printf("Payload: %s\n", string(jsb))
			panic("Bad Raw TX caught")
		}
	}()
	for i := 0; i < count; i++ {
		someInvalidTxs := getCorruptedJsonDeserializedTxs(tx, count)
		for _, theInvalidTx := range someInvalidTxs {
			txSpecific, ok := theInvalidTx.(*Tx)
			if !ok {
				continue
			}
			payloadTx = txSpecific
			// look for potential panics by calling verify
			isValid, _ := txSpecific.ValidateSanityData(nil, nil, nil, 0)
			// if it doesnt pass sanity then the next validation could panic, it's ok by spec
			if !isValid {
				continue
			}
			isValid, _ = txSpecific.ValidateTxByItself(boolParams, db, nil, nil, shardID, nil, nil)
			if !isValid {
				continue
			}
			errAlreadyInChain := txSpecific.ValidateTxWithBlockChain(nil, nil, nil, shardID, db)
			if !allowModifiedTXsToPass && errAlreadyInChain == nil {
				// make sure it's different
				s1 := formatTx(tx)
				s2 := formatTx(txSpecific)
				if bytes.Equal([]byte(s1), []byte(s2)) {
					continue
				}
				// the forged TX somehow is valid after all 3 checks, we caught a bug
				Printf("Original TX : %s\nChanged TX (still valid) : %s\n", s1, s2)
				panic("END TEST : a mal-TX was accepted")
			}

			// look for potential panics by calling verify
			isSane, _ := txSpecific.ValidateSanityData(nil, nil, nil, 0)
			// if it doesnt pass sanity then the next validation could panic, it's ok by spec
			if !isSane {
				continue
			}

			isSane, _ = txSpecific.ValidateTxByItself(boolParams, db, nil, nil, shardID, nil, nil)
			if !isSane {
				continue
			}
			txSpecific.ValidateTxWithBlockChain(nil, nil, nil, shardID, db)
		}
	}
}

func testTxTokenV2JsonMarshaler(tx *TxToken, count int, db *statedb.StateDB) {
	boolParams := make(map[string]bool)
	boolParams["hasPrivacy"] = hasPrivacyForPRV
	for i := 0; i < count; i++ {
		someInvalidTxs := getCorruptedJsonDeserializedTokenTxs(tx, count)
		for _, theInvalidTx := range someInvalidTxs {
			txSpecific, ok := theInvalidTx.(*TxToken)
			if !ok {
				continue
			}
			// look for potential panics by calling verify
			isValid, _ := txSpecific.ValidateSanityData(nil, nil, nil, 0)
			// if it doesnt pass sanity then the next validation could panic, it's ok by spec
			if !isValid {
				continue
			}
			isValid, _ = txSpecific.ValidateTxByItself(boolParams, db, nil, nil, shardID, nil, nil)
			if !isValid {
				continue
			}
			errAlreadyInChain := txSpecific.ValidateTxWithBlockChain(nil, nil, nil, shardID, db)
			if !allowModifiedTXsToPass && errAlreadyInChain == nil {
				// make sure it's different
				s1 := formatTx(tx)
				s2 := formatTx(txSpecific)
				if bytes.Equal([]byte(s1), []byte(s2)) {
					continue
				}
				// the forged TX somehow is valid after all 3 checks, we caught a bug
				Printf("Original TX : %s\nChanged TX (still valid) : %s\n", s1, s2)
				panic("END TEST : a mal-TXTOKEN was accepted")
			}

			// look for potential panics by calling verify
			isSane, _ := txSpecific.ValidateSanityData(nil, nil, nil, 0)
			// if it doesnt pass sanity then the next validation could panic, it's ok by spec
			if !isSane {
				continue
			}

			isSane, _ = txSpecific.ValidateTxByItself(boolParams, db, nil, nil, shardID, nil, nil)
			if !isSane {
				continue
			}
			txSpecific.ValidateTxWithBlockChain(nil, nil, nil, shardID, db)
		}
	}
}

func getRandomDigit() rune {

	ind := RandInt() % 10
	return rune(int(rune('0')) + ind)
}

func getRandomLetter() rune {
	ind := RandInt() % 52
	if ind < 26 {
		return rune(int(rune('A')) + ind)
	} else {
		return rune(int(rune('a')) + ind - 26)
	}
}

func getCorruptedJsonDeserializedTxs(tx *Tx, maxJsonChanges int) []metadata.Transaction {
	jsonBytes, err := json.Marshal(tx)
	So(err, ShouldBeNil)
	reconstructedTx := &Tx{}
	err = json.Unmarshal(jsonBytes, reconstructedTx)
	So(err, ShouldBeNil)
	jsonBytesAgain, err := json.Marshal(reconstructedTx)
	So(bytes.Equal(jsonBytes, jsonBytesAgain), ShouldBeTrue)
	var result []metadata.Transaction
	// json bytes are readable strings
	// we try to malleify a letter / digit
	// if that char is part of a key then it's equivalent to deleting that attribute
	s := string(jsonBytesAgain)
	theRunes := []rune(s)
	var payloadTx []byte
	defer func() {
		if r := recover(); r != nil {
			s := base58.Base58Check{}.Encode(payloadTx, 0)
			fmt.Printf("Payload: %s\n", s)
			panic("Bad Raw TX caught")
		}
	}()
	for i := 0; i < maxJsonChanges; i++ {
		// let the changes stack up many times to exhaust more cases
		corruptedIndex := RandInt() % len(theRunes)
		for j := maxTries; j > 0; j-- {
			if j == 0 {
				logger.Warnf("Max changes exceeded with : %s\n", s)
				return result
			}
			if unicode.IsLetter(theRunes[corruptedIndex]) || unicode.IsDigit(theRunes[corruptedIndex]) {
				break
			}
			// not letter -> retry
			corruptedIndex = RandInt() % len(theRunes)
		}
		// replace this letter with a random one
		var newRune rune
		if unicode.IsLetter(theRunes[corruptedIndex]) {
			newRune = getRandomLetter()
		} else {
			newRune = getRandomDigit()
		}
		if theRunes[corruptedIndex] == newRune {
			// remove that char
			theRunes = append(theRunes[:corruptedIndex], theRunes[corruptedIndex+1:]...)
		} else {
			theRunes[corruptedIndex] = newRune
		}
		temp := &Tx{}
		payloadTx = []byte(string(theRunes))
		err := json.Unmarshal([]byte(string(theRunes)), temp)
		if err != nil {
			continue
		}
		result = append(result, temp)
	}
	return result
}

func getCorruptedJsonDeserializedTokenTxs(tx *TxToken, maxJsonChanges int) []tx_generic.TransactionToken {
	jsonBytes, err := json.Marshal(tx)
	So(err, ShouldBeNil)
	reconstructedTx := &TxToken{}
	err = json.Unmarshal(jsonBytes, reconstructedTx)
	So(err, ShouldBeNil)
	jsonBytesAgain, err := json.Marshal(reconstructedTx)
	So(bytes.Equal(jsonBytes, jsonBytesAgain), ShouldBeTrue)
	var result []tx_generic.TransactionToken

	s := string(jsonBytesAgain)
	theRunes := []rune(s)
	for i := 0; i < maxJsonChanges; i++ {
		corruptedIndex := RandInt() % len(theRunes)
		for j := maxTries; j > 0; j-- {
			if j == 0 {
				logger.Warnf("Max changes exceeded with : %s\n", s)
				return result
			}
			if unicode.IsLetter(theRunes[corruptedIndex]) || unicode.IsDigit(theRunes[corruptedIndex]) {
				break
			}
			// not letter -> retry
			corruptedIndex = RandInt() % len(theRunes)
		}
		// replace this letter with a random one
		var newRune rune
		if unicode.IsLetter(theRunes[corruptedIndex]) {
			newRune = getRandomLetter()
		} else {
			newRune = getRandomDigit()
		}
		if theRunes[corruptedIndex] == newRune {
			// remove that char
			theRunes = append(theRunes[:corruptedIndex], theRunes[corruptedIndex+1:]...)
		} else {
			theRunes[corruptedIndex] = newRune
		}
		temp := &TxToken{}
		err := json.Unmarshal([]byte(string(theRunes)), temp)
		if err != nil {
			continue
		}
		result = append(result, temp)
	}
	return result
}

func RandInt() int {
	return rand.Int()
}

func formatTx(tx metadata.Transaction) string {
	jsb, _ := json.Marshal(tx)
	return string(jsb)
}

func resignUnprovenTx(decryptingKeys []*incognitokey.KeySet, tx *Tx, params *tx_generic.TxPrivacyInitParams, tokenData *TxTokenDataVersion2, isCA bool) error {
	tx.SetCachedHash(nil)
	tx.SetSig(nil)
	tx.SetSigPubKey(nil)
	var err error
	outputCoinsGeneric := tx.GetProof().GetOutputCoins()
	var outputCoins []*coin.CoinV2
	// pre-sign, we need unconcealed outputs
	// so receiver privatekeys here are for simulation
	var sharedSecrets []*operation.Point
	for ind, c := range outputCoinsGeneric {
		var dk *incognitokey.KeySet = decryptingKeys[ind%len(decryptingKeys)]
		mySkBytes := dk.PrivateKey[:]
		cv2 := &coin.CoinV2{}
		cv2.SetBytes(c.Bytes())
		cv2.Decrypt(dk)
		sharedSecret, err := cv2.RecomputeSharedSecret(mySkBytes)
		if err != nil {
			logger.Errorf("TEST : Cannot compute shared secret for coin %v", cv2.Bytes())
			return err
		}
		sharedSecrets = append(sharedSecrets, sharedSecret)
		outputCoins = append(outputCoins, cv2)
	}
	inputCoins := params.InputCoins

	message := tx.Hash()[:]
	if tokenData != nil {
		tdh, err := tokenData.Hash()
		if err != nil {
			panic("Hash failed")
		}
		temp := common.HashH(append(message, tdh[:]...))
		message = temp[:]
	}

	if isCA {
		utils.Logger.Log.Warnf("Re-sign a CA transaction")
		err = tx.signCA(inputCoins, outputCoins, sharedSecrets, params, message[:])
	} else {
		utils.Logger.Log.Warnf("Re-sign a non-CA transaction")
		err = tx.signOnMessage(inputCoins, outputCoins, params, message[:])
	}

	jsb, _ := json.MarshalIndent(tx, "", "\t")
	logger.Debugf("Resigning TX for testing : Rehash message %s\n => %v", string(jsb), tx.Hash())
	return err
}
