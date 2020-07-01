package transaction

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/privacy/coin"
	"github.com/incognitochain/incognito-chain/privacy/key"
	"github.com/incognitochain/incognito-chain/privacy/operation"
	"github.com/incognitochain/incognito-chain/trie"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
)

//var (
//	// num of private keys
//	maxPrivateKeys = 15
//	minPrivateKeys = 8
//
//	maxInputs = 7
//	minInputs = 1
//	numOfLoops = 10
//)
//var (
//	warperDBStatedbTest statedb.DatabaseAccessWarper
//	emptyRoot           = common.HexToHash(common.HexEmptyRoot)
//	prefixA             = "serialnumber"
//	prefixB             = "serialnumberderivator"
//	prefixC             = "serial"
//	prefixD             = "commitment"
//	prefixE             = "outputcoin"
//	keysA               = []common.Hash{}
//	keysB               = []common.Hash{}
//	keysC               = []common.Hash{}
//	keysD               = []common.Hash{}
//	keysE               = []common.Hash{}
//	valuesA             = [][]byte{}
//	valuesB             = [][]byte{}
//	valuesC             = [][]byte{}
//	valuesD             = [][]byte{}
//	valuesE             = [][]byte{}
//
//	limit100000 = 100000
//	limit10000  = 10000
//	limit1000   = 1000
//	limit100    = 100
//	limit1      = 1
//)

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

func createSampleInputCoin(pubKey *operation.Point, amount uint64, msg []byte) (*coin.PlainCoinV1, error) {
	c := new(coin.PlainCoinV1).Init()

	c.SetValue(amount)
	c.SetInfo(msg)
	c.SetPublicKey(pubKey)
	c.SetSNDerivator(operation.RandomScalar())
	c.SetRandomness(operation.RandomScalar())
	err := c.CommitAll()

	if err != nil {
		return nil, err
	}

	return c, nil
}

func createConversionParams(numInputs, numOutputs int)(*statedb.StateDB, *incognitokey.KeySet,
	[]*privacy.PaymentInfo, *TxConvertVer1ToVer2InitParams, error){
	var senderSK privacy.PrivateKey
	var keySet *incognitokey.KeySet

	//create a sample test DB
	testDB, _ := statedb.NewWithPrefixTrie(emptyRoot, warperDBStatedbTest)

	//generate keyset: we want the public key to be in Shard 0
	for {
		//generate a private key
		senderSK = key.GeneratePrivateKey(common.RandBytes(32))

		//make keySets from privateKey
		keySet = new(incognitokey.KeySet)
		err := keySet.InitFromPrivateKey(&senderSK)

		if err != nil {
			return nil, nil, nil, nil, err
		}

		//we want the public key to belong to Shard 0
		if keySet.PaymentAddress.Pk[31] == 0 {
			break
		}
	}

	//create input coins
	inputCoins := make([]coin.PlainCoin, numInputs)
	sumInput := uint64(0)
	var err error

	pubKey, err := new(operation.Point).FromBytesS(keySet.PaymentAddress.Pk)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	for i := 0; i< numInputs; i++ {
		amount := uint64(common.RandIntInterval(0, 1000))
		inputCoins[i], err = createSampleInputCoin(pubKey, amount, nil)
		if err != nil {
			return nil, nil, nil, nil, err
		}
		sumInput += amount
	}

	sumOutput := uint64(0)

	paymentInfo := make([]*key.PaymentInfo, numOutputs)

	for i :=0;i< numOutputs;i++{
		amount := uint64(common.RandIntInterval(0, int(sumInput-sumOutput)))
		paymentInfo[i] = key.InitPaymentInfo(keySet.PaymentAddress, amount, nil)
		sumOutput += amount
	}

	//calculate sample fee
	fee := sumInput - sumOutput

	//create conversion params
	txConvertParams := NewTxConvertVer1ToVer2InitParams(
		&keySet.PrivateKey,
		paymentInfo,
		inputCoins,
		fee,
		testDB,
		nil, // use for prv coin -> nil is valid
		nil,
		nil,
	)

	return testDB, keySet, paymentInfo, txConvertParams, nil
}

func TestInitializeTxConversion(t *testing.T) {

	_, _, _, txConvertParams, err := createConversionParams(10, 5)

	assert.Equal(t, nil, err, "createConversionParams returns an error: %v", err)

	err = validateTxConvertVer1ToVer2Params(txConvertParams)

	txConversion := new(TxVersion2)

	assert.Equal(t, nil, err, "validateTxConvertVer1ToVer2Params returns an error: %v", err)

	err = initializeTxConversion(txConversion, txConvertParams)

	assert.Equal(t, nil, err, "initializeTxConversion returns an error: %v", err)

}

func TestProveConversion(t *testing.T){
}


