package tx_ver2

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
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
	minPrivateKeys = 2
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
			for i := range theCoins {
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
		Convey("prepare keys", func() {
			dummyPrivateKeys, keySets, paymentInfo = preparePaymentKeys(numOfPrivateKeys)
		})

		Convey("create & store UTXOs", func() {
			// pastCoins are coins we manually store in the dummyDB to simulate the db having OTAs from chaindata
			pastCoins = make([]coin.Coin, (10+numOfInputs)*len(dummyPrivateKeys))
			for i := range pastCoins {
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
			for i := range dummyPrivateKeys {
				paymentInfoOut[i] = key.InitPaymentInfo(keySets[i].PaymentAddress, uint64(3000), []byte("test out"))
			}
		})

		Convey("decrypt inputs", func() {
			// now we take some of those stored coins to use as TX input
			// for the TX to be valid, these inputs must associate to one same private key
			// (it's guaranteed by our way of indexing the pastCoins array)
			inputCoins = make([]coin.PlainCoin, numOfInputs)
			for i := range inputCoins {
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
			var err error
			tx, err = tx.startVerifyTx(dummyDB)
			So(err, ShouldBeNil)
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

func TestTx_FromCompactBytes(t *testing.T) {
	encodedTxStr := "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"
	encodedTx, _, err := base58.Base58Check{}.Decode(encodedTxStr)
	if err != nil {
		panic(err)
	}

	tx := new(Tx)
	err = json.Unmarshal(encodedTx, &tx)
	if err != nil {
		panic(err)
	}

	compactBytes, err := tx.ToCompactBytes()
	if err != nil {
		panic(err)
	}

	reductionRate := 1 - float64(len(compactBytes))/float64(len(encodedTx))
	fmt.Printf("jsonSize: %v, compactSize: %v, reductionRate: %v\n", len(encodedTx), len(compactBytes), reductionRate)

	newTx := new(Tx)
	err = newTx.FromCompactBytes(compactBytes)
	if err != nil {
		panic(err)
	}
	assert.Equal(t, tx.Hash().String(), newTx.Hash().String(), "tx hashes mismatch")
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
	err := malTx.Init(malFeeParams)
	So(err, ShouldBeNil)
	malTx, err = malTx.startVerifyTx(db)
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
		sumIn-sumOut, true,
		db,
		nil,
		nil,
		[]byte{},
	)
	malTx := &Tx{}
	err := malTx.Init(malFeeParams)
	So(err, ShouldBeNil)
	malTx.Fee = sumIn-sumOut+1111
	malTx, err = malTx.startVerifyTx(db)
	So(err, ShouldBeNil)
	boolParams := make(map[string]bool)
	boolParams["hasPrivacy"] = hasPrivacyForPRV
	boolParams["isNewTransaction"] = true
	isValid, errMalVerify := malTx.ValidateTxByItself(boolParams, db, nil, nil, byte(0), nil, nil)
	So(errMalVerify, ShouldNotBeNil)
	So(isValid, ShouldBeFalse)
}

func testTxV2OneFakeInput(txv2 *Tx, privateKeys []*privacy.PrivateKey, keySets []*incognitokey.KeySet, db *statedb.StateDB, params *tx_generic.TxPrivacyInitParams, pastCoins []coin.Coin) {
	jsb, _ := json.MarshalIndent(txv2, "", "\t")
	logger.Debugf("debug original tx %s %s", txv2.Hash().String(), string(jsb))
	// likewise, if someone took an already proven tx and swaps one input coin
	// for another random coin from outside, the tx cannot go through
	// (here we only meddle with coin-changing - not adding/removing - since length checks are included within mlsag)
	var err error
	inputCoins := txv2.GetProof().GetInputCoins()
	numOfInputs := len(inputCoins)
	changed := RandInt() % numOfInputs
	saved := inputCoins[changed]
	inputCoins[changed], _ = pastCoins[len(privateKeys)*(numOfInputs+1)].Decrypt(keySets[0])
	txv2.GetProof().SetInputCoins(inputCoins)

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
	txv2.GetProof().SetInputCoins(inputCoins)
	err = resignUnprovenTx(keySets, txv2, params, nil, false)
	isValid, err = txv2.ValidateSanityData(nil, nil, nil, 0)
	So(err, ShouldBeNil)
	So(isValid, ShouldBeTrue)
	jsb, _ = json.MarshalIndent(txv2, "", "\t")
	logger.Debugf("debug tx after recover %s %s", txv2.Hash().String(), string(jsb))
	isValid, err = txv2.ValidateTxByItself(boolParams, db, nil, nil, byte(0), nil, nil)
	So(err, ShouldBeNil)
	So(isValid, ShouldBeTrue)
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
	outs = txv2.GetProof().GetOutputCoins()
	prvOutput, ok = outs[0].(*coin.CoinV2)
	savedCoinBytes = prvOutput.Bytes()
	So(ok, ShouldBeTrue)
	payInf := paymentInfoOut[0]
	// totally fresh OTA of the same amount, meant for the same PaymentAddress
	newCoin, err := coin.NewCoinFromPaymentInfo(payInf)
	So(err, ShouldBeNil)
	newCoin.ConcealOutputCoin(keySets[0].PaymentAddress.GetPublicView())
	txv2.GetProof().(*privacy.ProofV2).GetAggregatedRangeProof().(*privacy.AggregatedRangeProofV2).GetCommitments()[0] = newCoin.GetCommitment()
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
	txv2.GetProof().(*privacy.ProofV2).GetAggregatedRangeProof().(*privacy.AggregatedRangeProofV2).GetCommitments()[0] = prvOutput.GetCommitment()
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
	malTx, err = malTx.startVerifyTx(db)
	So(err, ShouldBeNil)
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
	if err != nil {
		return err
	}

	jsb, _ := json.MarshalIndent(tx, "", "\t")
	logger.Debugf("Resigning TX for testing : Rehash message %s\n => %v", string(jsb), tx.Hash())

	temp, err := tx.startVerifyTx(params.StateDB)
	if err != nil {
		return err
	}
	*tx = *temp
	return nil
}

func (tx *Tx) startVerifyTx(db *statedb.StateDB) (*Tx, error) {
	marshaledTx, _ := json.Marshal(tx)
	result := &Tx{}
	err := json.Unmarshal(marshaledTx, result)
	if err != nil {
		return nil, err
	}
	marshaledTx2, _ := json.Marshal(result)
	if !bytes.Equal(marshaledTx, marshaledTx2) {
		return nil, fmt.Errorf("marshal output inconsistent %s", marshaledTx)
	}
	err = result.LoadData(db)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (tx *TxToken) startVerifyTx(db *statedb.StateDB) (*TxToken, error) {
	marshaledTx, _ := json.Marshal(tx)
	result := &TxToken{}
	err := json.Unmarshal(marshaledTx, result)
	if err != nil {
		return nil, err
	}
	marshaledTx2, _ := json.Marshal(result)
	if !bytes.Equal(marshaledTx, marshaledTx2) {
		return nil, fmt.Errorf("marshal output inconsistent %s", marshaledTx)
	}
	err = result.LoadData(db)
	if err != nil {
		return nil, err
	}
	return result, nil
}
