package transaction

import(
	"math/rand"
	"testing"
	"time"
	"os"
	"strconv"
	"fmt"
	"io/ioutil"
	// "encoding/json"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/trie"
	"github.com/incognitochain/incognito-chain/incdb"
	_ "github.com/incognitochain/incognito-chain/incdb/lvdb"
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
	diskDB incdb.Database
	bridgeDB *statedb.StateDB
	activeLogger common.Logger
	inactiveLogger common.Logger
)

var _ = func() (_ struct{}) {
	// initialize a `test` db in the OS's tempdir
	// and with it, a db access wrapper that reads/writes our transactions
	fmt.Println("This runs before init()!")
	testLogFile, _ := os.OpenFile("test-log.txt",os.O_RDWR|os.O_TRUNC|os.O_CREATE, 0755)

	inactiveLogger = common.NewBackend(nil).Logger("test", true)
	activeLogger = common.NewBackend(testLogFile).Logger("test", false)
	// activeLogger = inactiveLogger
	activeLogger.SetLevel(common.LevelDebug)
	privacy.LoggerV1.Init(inactiveLogger)
	privacy.LoggerV2.Init(activeLogger)
	// can switch between the 2 loggers to mute logs as one wishes
	utils.Logger.Init(activeLogger)
	bulletproofs.Logger.Init(common.NewBackend(nil).Logger("test", true))
	
	// bridgeDB  = dummyDB.Copy()
	trie.Logger.Init(common.NewBackend(nil).Logger("test", true))
	return
}()

func init() {
	dbPath, err := ioutil.TempDir(os.TempDir(), "test_statedb_")
	if err != nil {
		panic(err)
	}
	diskDB, err = incdb.Open("leveldb", dbPath)
	if err != nil {
		panic(err)
	}

	warperDBStatedbTest = statedb.NewDatabaseAccessWarper(diskDB)
	dummyDB, _ = statedb.NewWithPrefixTrie(emptyRoot, warperDBStatedbTest)
}

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
		// jsb, _ := json.Marshal(c)
		// fmt.Printf("Coin is %s\n", string(jsb))
		coinsInBytes = append(coinsInBytes, c.Bytes())
		otas = append(otas, c.GetPublicKey().ToBytesS())
	}
	// fmt.Printf("Db is %v\n", db)
	err := statedb.StoreOTACoinsAndOnetimeAddresses(db, tokenID, height, coinsInBytes, otas, shardID)
	if err!=nil{
		panic(err)
	}
}

func BenchmarkQueryCoinV1(b *testing.B) {
	rand.Seed(time.Now().UnixNano())
	clargs := os.Args[len(os.Args)-2:]
	// fmt.Println(clargs)

	numOfCoinsTotal, _ := strconv.Atoi(clargs[0])
	numOfCoinsPerKey, _ := strconv.Atoi(clargs[1])
	numOfPrivateKeys := numOfCoinsTotal / numOfCoinsPerKey
	// numOfReads,_ := strconv.Atoi(clargs[2])
	
	fmt.Printf("\n------------------CoinV1 Query Benchmark\n")
	fmt.Printf("Number of coins in db   : %d\n", numOfCoinsTotal)
	fmt.Printf("Number of coins per key : %d\n", numOfCoinsPerKey)
	keySets, _ := prepareKeySets(numOfPrivateKeys)


	// var coins []*privacy.CoinV1 = make([]*privacy.CoinV1, numOfCoinsPerKey)
	for _, keySet := range keySets {
		pubKey, _ := new(operation.Point).FromBytesS(keySet.PaymentAddress.Pk)
		pcs, _ := createSamplePlainCoinsFromTotalAmount(keySet.PrivateKey, pubKey, 1000, numOfCoinsPerKey, 1)
		coinsToBeSaved := [][]byte{}
		for _, plainCoin := range pcs {
			encryptedCoin := &privacy.CoinV1{}
			encryptedCoin.Init()
			var ok bool
			encryptedCoin.CoinDetails, ok = plainCoin.(*privacy.PlainCoinV1)
			assert.Equal(b, true, ok)
			err := encryptedCoin.Encrypt(keySet.PaymentAddress.Tk)
			if err != nil {
				panic(err)
			}
			// fmt.Printf("Store coin %x by key %x\n", encryptedCoin.Bytes(), keySet.PaymentAddress.Pk)
			coinsToBeSaved = append(coinsToBeSaved, encryptedCoin.Bytes())
		}
		
		err := statedb.StoreOutputCoins(dummyDB, common.PRVCoinID, keySet.PaymentAddress.Pk, coinsToBeSaved, shardID)
		assert.Equal(b, nil, err)
	}
	newRoot, err := dummyDB.Commit(true)
	assert.Equal(b, nil, err)
	err = warperDBStatedbTest.TrieDB().Commit(newRoot, false)
	assert.Equal(b, nil, err)
	coinDB, err := statedb.NewWithPrefixTrie(newRoot, warperDBStatedbTest)
	if err != nil {
		panic(err)
	}
	// fmt.Println(coinDB)

	// each loop reads all output coins for a public key
	b.ResetTimer()
	for loop := 0; loop < b.N; loop++ {
		chosenIndex := rand.Int() % len(keySets)
		ks := keySets[chosenIndex]
		// fmt.Printf("Get coin by key %x\n", ks.PaymentAddress.Pk)
		// pubKey, _ := new(operation.Point).FromBytesS(ks.PaymentAddress.Pk)
		results, err := utils.QueryDbCoinVer1(ks.PaymentAddress.Pk, shardID, &common.PRVCoinID, coinDB)
		if err != nil {
			panic(err)
		}
		assert.Equal(b, numOfCoinsPerKey, len(results))
		assert.Equal(b, nil, err)
	}
}


func prepareKeysAndPaymentsV2(count int) ([]*incognitokey.KeySet, []*key.PaymentInfo) {
	// create many random private keys
	// then use each privatekey to derive Incognito keyset (various keys for everything inside the protocol)
	// we ensure they all belong in shard 0 for this test

	// PaymentInfo is like `intent` for making Coin.
	// the paymentInfo slice here will be used to create pastCoins & inputCoins
	// we populate `value` fields with some arbitrary, big-enough constant (here, 4000*len)
	// `message` field can be anything
	dummyPrivateKeys := make([]*key.PrivateKey,count)
	keySets := make([]*incognitokey.KeySet,len(dummyPrivateKeys))
	paymentInfos := make([]*key.PaymentInfo, len(dummyPrivateKeys))
	for i := 0; i < count; i += 1 {
		for{
			privateKey := key.GeneratePrivateKey(common.RandBytes(32))
			dummyPrivateKeys[i] = &privateKey
			keySets[i] = new(incognitokey.KeySet)
			err := keySets[i].InitFromPrivateKey(dummyPrivateKeys[i])
			if err != nil {
				panic(err)
			}

			paymentInfos[i] = key.InitPaymentInfo(keySets[i].PaymentAddress, uint64(4000*len(dummyPrivateKeys)), []byte("test in"))
			pkb := []byte(paymentInfos[i].PaymentAddress.Pk)
			if common.GetShardIDFromLastByte(pkb[len(pkb)-1])==shardID{
				break
			}
		}
	}
	fmt.Println("Key & PaymentInfo generation finished")
	return keySets, paymentInfos
}

func getCoinFilterByOTAKey() utils.CoinMatcher{
    return func(c *privacy.CoinV2, kvargs map[string]interface{}) bool{
        entry, exists := kvargs["otaKey"]
        if !exists{
            return false
        }
        vk, ok := entry.(privacy.OTAKey)
        if !ok{
            return false
        }
        ks := &incognitokey.KeySet{}
        ks.OTAKey = vk

        pass, _ := c.DoesCoinBelongToKeySet(ks)
        return pass
    }
}

func BenchmarkQueryCoinV2(b *testing.B) {
	rand.Seed(time.Now().UnixNano())
	clargs := os.Args[len(os.Args)-3:]
	// fmt.Println(clargs)

	numOfCoinsTotal, _ := strconv.Atoi(clargs[0])
	numOfCoinsPerKey, _ := strconv.Atoi(clargs[1])
	num := 1
	if len(clargs) >= 3 {
		num, _ = strconv.Atoi(clargs[2])
	}
	maxHeight := uint64(num)
	numOfPrivateKeys := numOfCoinsTotal / numOfCoinsPerKey
	// numOfReads,_ := strconv.Atoi(clargs[2])
	
	fmt.Printf("\n------------------CoinV2 Query Benchmark\n")
	fmt.Printf("Number of coins in db   : %d\n", numOfCoinsTotal)
	fmt.Printf("Number of coins per key : %d\n", numOfCoinsPerKey)
	fmt.Printf("Spread between height 0-%d\n", maxHeight - 1)

	keySets, paymentInfos := prepareKeysAndPaymentsV2(numOfPrivateKeys)

	pastCoins := make([]coin.Coin, numOfCoinsPerKey * numOfPrivateKeys)
	for i, _ := range pastCoins {
		tempCoin, err := coin.NewCoinFromPaymentInfo(paymentInfos[i % numOfPrivateKeys])
		assert.Equal(b, nil, err)
		assert.Equal(b, false, tempCoin.IsEncrypted())
		tempCoin.ConcealOutputCoin(keySets[i % numOfPrivateKeys].PaymentAddress.GetPublicView())
		assert.Equal(b, true, tempCoin.IsEncrypted())
		assert.Equal(b, true, tempCoin.GetSharedRandom() == nil)
		// fmt.Printf("Add coin by key %x\n", keySets[i % numOfPrivateKeys].OTAKey)
		pastCoins[i] = tempCoin
		storeCoinV2(dummyDB, []coin.Coin{tempCoin}, 0, common.PRVCoinID, uint64(i) % maxHeight)
	}	

	newRoot, err := dummyDB.Commit(true)
	assert.Equal(b, nil, err)
	err = warperDBStatedbTest.TrieDB().Commit(newRoot, false)
	assert.Equal(b, nil, err)
	coinDB, err := statedb.NewWithPrefixTrie(newRoot, warperDBStatedbTest)
	if err != nil {
		panic(err)
	}

	// each loop reads all output coins for a public key
	b.ResetTimer()
	for loop := 0; loop < b.N; loop++ {
		chosenIndex := rand.Int() % len(keySets)
		ks := keySets[chosenIndex]
		// fmt.Printf("Get coin by key %x\n", ks.OTAKey)
		otaKey := ks.OTAKey
		results, err := utils.QueryDbCoinVer2(otaKey, shardID, &common.PRVCoinID, 0, maxHeight, coinDB, getCoinFilterByOTAKey())
		assert.Equal(b, len(results), numOfCoinsPerKey)
		if err != nil {
			panic(err)
		}
		assert.Equal(b, nil, err)
	}
}