package pdex

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"testing"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incdb"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	metadataPdexv3 "github.com/incognitochain/incognito-chain/metadata/pdexv3"
	"github.com/incognitochain/incognito-chain/transaction"
	"github.com/incognitochain/incognito-chain/transaction/tx_generic"
	. "github.com/stretchr/testify/assert"
)

var (
	_          = fmt.Print
	testWarper statedb.DatabaseAccessWarper
	emptyRoot  = common.HexToHash(common.HexEmptyRoot)
	testDB     *statedb.StateDB
	logger     common.Logger
)

func init() {
	// initialize a `test` db in the OS's tempdir
	// and with it, a db access wrapper that reads/writes our transactions
	common.MaxShardNumber = 1
	testLogFile, _ := os.OpenFile("test.log", os.O_RDWR|os.O_TRUNC|os.O_CREATE, 0755)
	logger = common.NewBackend(testLogFile).Logger("test", false)
	logger.SetLevel(common.LevelDebug)
	Logger.Init(logger)

	dbPath, _ := ioutil.TempDir(os.TempDir(), "test_statedb_")
	d, _ := incdb.Open("leveldb", dbPath)
	testWarper = statedb.NewDatabaseAccessWarper(d)
	testDB, _ = statedb.NewWithPrefixTrie(emptyRoot, testWarper)
}

func TestSortOrder(t *testing.T) {
	type TestData struct {
		Orders []*Order `json:"orders"`
	}

	type TestResult struct {
		Orders []*Order `json:"orders"`
	}

	var testcases []Testcase
	testcases = append(testcases, sortOrderTestcases...)

	testState := newStateV2WithValue(nil, nil, make(map[string]*PoolPairState),
		Params{}, nil, map[string]bool{})
	blankPairID := "pair0"
	testState.poolPairs[blankPairID] = &PoolPairState{orderbook: Orderbook{[]*Order{}}}
	for _, testcase := range testcases {
		t.Run(testcase.Name, func(t *testing.T) {
			var testdata TestData
			err := json.Unmarshal([]byte(testcase.Data), &testdata)
			NoError(t, err)
			// get a random permutation of orders in test data for inserting
			// since this test inserts items at random order, it is not compatible for testing equality-breaking of orders
			perm := rand.Perm(len(testdata.Orders))
			var orderbookPerm []*Order
			for _, newInd := range perm {
				orderbookPerm = append(orderbookPerm, testdata.Orders[newInd])
			}
			testdata.Orders = orderbookPerm

			for _, item := range testdata.Orders {
				pair := testState.poolPairs[blankPairID]
				pair.orderbook.InsertOrder(item)
				testState.poolPairs[blankPairID] = pair
			}
			encodedResult, _ := json.Marshal(TestResult{testState.poolPairs[blankPairID].orderbook.orders})
			Equal(t, testcase.Expected, string(encodedResult))
		})
	}
}

func TestProduceTrade(t *testing.T) {
	type TestData struct {
		Metadata metadataPdexv3.TradeRequest `json:"metadata"`
	}

	type TestResult struct {
		Instructions [][]string `json:"instructions"`
	}

	var testcases []Testcase
	testcases = append(testcases, produceTradeTestcases...)
	for _, testcase := range testcases {
		t.Run(testcase.Name, func(t *testing.T) {
			var testdata TestData
			err := json.Unmarshal([]byte(testcase.Data), &testdata)
			NoError(t, err)

			env := metadataToBeacon(&testdata.Metadata, 0)
			testState := mustReadState("test_state.json")
			instructions, err := testState.BuildInstructions(env)

			encodedResult, _ := json.Marshal(TestResult{instructions})
			Equal(t, testcase.Expected, string(encodedResult))
		})
	}
}

func metadataToBeacon(md metadataCommon.Metadata, shardID byte) StateEnvironment {
	mytx := &transaction.TxVersion2{}
	valEnv := tx_generic.DefaultValEnv()
	valEnv = tx_generic.WithShardID(valEnv, int(shardID))
	mytx.SetMetadata(md)
	mytx.SetValidationEnv(valEnv)

	return NewStateEnvBuilder().
		BuildBeaconHeight(10).
		BuildListTxs(map[byte][]metadataCommon.Transaction{shardID: []metadataCommon.Transaction{mytx}}).
		BuildBCHeightBreakPointPrivacyV2(0).
		Build()
}

type Testcase struct {
	Name          string `json:"name"`
	Data          string `json:"data"`
	Expected      string `json:"expected"`
	ExpectSuccess bool   `json:"expectSuccess"`
}

func mustReadTestcases(filename string) []Testcase {
	raw, err := ioutil.ReadFile("testdata/" + filename)
	if err != nil {
		panic(err)
	}
	var results []Testcase = make([]Testcase, 30)
	err = json.Unmarshal(raw, &results)
	if err != nil {
		panic(err)
	}
	return results
}

func mustReadState(filename string) *stateV2 {
	raw, err := ioutil.ReadFile("testdata/" + filename)
	if err != nil {
		panic(err)
	}

	var temp struct {
		PoolPairs  map[string]rawdbv2.Pdexv3PoolPair `json:"poolPairs"`
		Orderbooks map[string]Orderbook              `json:"orderbooks"`
		Params     Params                            `json:"params"`
	}

	err = json.Unmarshal(raw, &temp)
	if err != nil {
		panic(err)
	}

	s := newStateV2WithValue(nil, nil, make(map[string]*PoolPairState),
		Params{}, nil, map[string]bool{})
	for k, v := range temp.PoolPairs {
		s.poolPairs[k] = &PoolPairState{state: v, orderbook: temp.Orderbooks[k]}
	}
	return s
}

var sortOrderTestcases = mustReadTestcases("sort_orders.json")
var produceTradeTestcases = mustReadTestcases("produce_trade.json")
