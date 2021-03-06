package transaction

import(
	"math/rand"
	"testing"
	"time"
	"os"
	"strconv"
	"fmt"
	"io/ioutil"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/trie"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/incognitochain/incognito-chain/incognitokey"

	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/privacy/coin"
	"github.com/incognitochain/incognito-chain/privacy/key"
	"github.com/incognitochain/incognito-chain/privacy/operation"
	"github.com/incognitochain/incognito-chain/privacy/privacy_v2/bulletproofs"
	"github.com/incognitochain/incognito-chain/transaction/utils"

	"github.com/stretchr/testify/assert"
)

var (
	// num of private keys
	maxPrivateKeys = 10
	minPrivateKeys = 1
	maxInputs = 10
	minInputs = 1
	maxTries = 100
	numOfLoops = 10

	allowModifiedTXsToPass = false
	hasPrivacyForPRV   bool = true
	hasPrivacyForToken bool = true
	shardID            byte = byte(0)

	positiveTestsFileName = "./testdata/accepted.txt"
	negativeTestsFileName = "./testdata/rejected.txt"
	b58 = base58.Base58Check{}
)
// variables for initializing stateDB for test
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
	activeLogger common.Logger
	inactiveLogger common.Logger
)

var _ = func() (_ struct{}) {
	// initialize a `test` db in the OS's tempdir
	// and with it, a db access wrapper that reads/writes our transactions
	fmt.Println("This runs before init()!")
	testLogFile, err := os.OpenFile("test-log.txt",os.O_RDWR|os.O_TRUNC|os.O_CREATE, 0755)

	inactiveLogger = common.NewBackend(nil).Logger("test", true)
	activeLogger = common.NewBackend(testLogFile).Logger("test", false)
	// activeLogger = inactiveLogger
	activeLogger.SetLevel(common.LevelDebug)
	privacy.LoggerV1.Init(inactiveLogger)
	privacy.LoggerV2.Init(activeLogger)
	// can switch between the 2 loggers to mute logs as one wishes
	utils.Logger.Init(activeLogger)
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

func prepareKeySets(numKeySets int) ([]*incognitokey.KeySet, error) {
	keySets := make([]*incognitokey.KeySet, numKeySets)
	//generate keysets: we want the public key to be in Shard 0
	for i := 0; i < numKeySets; i++ {
		for {
			//generate a private key
			privateKey := key.GeneratePrivateKey(common.RandBytes(32))

			//make keySets from privateKey
			keySet := new(incognitokey.KeySet)
			err := keySet.InitFromPrivateKey(&privateKey)

			if err != nil {
				return nil, err
			}

			//we want the public key to belong to Shard 0
			if keySet.PaymentAddress.Pk[31] == 0 {
				keySets[i] = keySet
				break
			}
		}
	}
	return keySets, nil
}

func createSamplePlainCoinsFromTotalAmount(senderSK privacy.PrivateKey, pubkey *operation.Point, amount uint64, count, version int) ([]privacy.PlainCoin, error) {
	coinList := []coin.PlainCoin{}
	if version == coin.CoinVersion1 {
		for i := 0; i < count; i++ {
			theCoin, err := createSamplePlainCoinV1(senderSK, pubkey, amount, nil)
			if err != nil {
				return nil, err
			}
			coinList = append(coinList, theCoin)
		}
	}
	return coinList, nil
}

func createSamplePlainCoinV1(privKey privacy.PrivateKey, pubKey *operation.Point, amount uint64, msg []byte) (*coin.PlainCoinV1, error) {
	c := new(coin.PlainCoinV1).Init()

	c.SetValue(amount)
	c.SetInfo(msg)
	c.SetPublicKey(pubKey)
	c.SetSNDerivator(operation.RandomScalar())
	c.SetRandomness(operation.RandomScalar())

	//Derive serial number from snDerivator
	c.SetKeyImage(new(operation.Point).Derive(privacy.PedCom.G[0], new(operation.Scalar).FromBytesS(privKey), c.GetSNDerivator()))

	//Create commitment
	err := c.CommitAll()

	if err != nil {
		return nil, err
	}

	return c, nil
}

func storeCoinV2(db *statedb.StateDB, coinsToBeSaved []coin.Coin, shardID byte, tokenID common.Hash, height uint64){
	coinsInBytes := make([][]byte, 0)
	otas := make([][]byte, 0)
	for _,c := range coinsToBeSaved{
		coinsInBytes = append(coinsInBytes, c.Bytes())
		otas = append(otas, c.GetPublicKey().ToBytesS())
	}
	err := statedb.StoreOTACoinsAndOnetimeAddresses(db, tokenID, height, coinsInBytes, otas, shardID)
	if err!=nil{
		panic(err)
	}
}

func BenchmarkQueryCoinV1(b *testing.B) {
	rand.Seed(time.Now().UnixNano())
	clargs := os.Args[len(os.Args)-2:]
	// fmt.Println(clargs)

	numOfPrivateKeys, _ := strconv.Atoi(clargs[0])
	numOfCoinsPerKey, _ := strconv.Atoi(clargs[1])
	// numOfReads,_ := strconv.Atoi(clargs[2])
	
	fmt.Printf("\n------------------CoinV1 Query Benchmark\n")
	fmt.Printf("Number of keys : %d\n", numOfPrivateKeys)
	fmt.Printf("Number of coins       : %d\n", numOfPrivateKeys * numOfCoinsPerKey)
	// fmt.Printf("Number of outputs      : %d\n", numOfOutputs)
	keySets, _ := prepareKeySets(numOfPrivateKeys)


	// var coins []*privacy.CoinV1 = make([]*privacy.CoinV1, numOfCoinsPerKey)
	for _, keySet := range keySets {
		pubKey, _ := new(operation.Point).FromBytesS(keySet.PaymentAddress.Pk)
		coins, _ := createSamplePlainCoinsFromTotalAmount(keySet.PrivateKey, pubKey, 1000, numOfCoinsPerKey, 1)
		coinsToBeSaved := [][]byte{}
		for _, outCoin := range coins {
			coinsToBeSaved = append(coinsToBeSaved, outCoin.Bytes())
		}
		err := statedb.StoreOutputCoins(dummyDB, common.PRVCoinID, pubKey.ToBytesS(), coinsToBeSaved, shardID)
		assert.Equal(b, nil, err)
	}

	// each loop reads all output coins for a public key
	b.ResetTimer()
	for loop := 0; loop < b.N; loop++ {
		chosenIndex := rand.Int() % len(keySets)
		ks := keySets[chosenIndex]
		pubKey, _ := new(operation.Point).FromBytesS(ks.PaymentAddress.Pk)
		_, err := utils.QueryDbCoinVer1(pubKey.ToBytesS(), shardID, &common.PRVCoinID, dummyDB)
		assert.Equal(b, nil, err)
	}
}
