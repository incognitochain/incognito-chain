package pdex

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"testing"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incdb"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	metadataPdexv3 "github.com/incognitochain/incognito-chain/metadata/pdexv3"
	"github.com/incognitochain/incognito-chain/privacy"
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
	privacy.LoggerV2.Init(logger)
	transaction.Logger.Init(logger)

	dbPath, _ := ioutil.TempDir(os.TempDir(), "test_statedb_")
	d, _ := incdb.Open("leveldb", dbPath)
	testWarper = statedb.NewDatabaseAccessWarper(d)
	testDB, _ = statedb.NewWithPrefixTrie(emptyRoot, testWarper)
}

func setTestTradeConfig() {
	config.AbortParam()
	config.Param().PDexParams.Pdexv3BreakPointHeight = 1
	config.Param().PDexParams.ProtocolFundAddress = "12svfkP6w5UDJDSCwqH978PvqiqBxKmUnA9em9yAYWYJVRv7wuXY1qhhYpPAm4BDz2mLbFrRmdK3yRhnTqJCZXKHUmoi7NV83HCH2YFpctHNaDdkSiQshsjw2UFUuwdEvcidgaKmF3VJpY5f8RdN"
	config.Param().EpochParam.NumberOfBlockInEpoch = 50
}

func TestProduceTrade(t *testing.T) {
	setTestTradeConfig()
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

			env := skipToProduce([]metadataCommon.Metadata{&testdata.Metadata}, 0)
			testState := mustReadState("test_state.json")
			testState.params = &Params{
				DefaultFeeRateBPS: 30,
			}
			temp := &StateFormatter{}
			temp.FromState(testState)

			instructions, err := testState.BuildInstructions(env)
			NoError(t, err)

			encodedResult, _ := json.Marshal(TestResult{instructions})
			Equal(t, testcase.Expected, string(encodedResult))
		})
	}
}

func TestProduceSameBlockTrades(t *testing.T) {
	setTestTradeConfig()
	type TestData struct {
		Metadata []metadataPdexv3.TradeRequest `json:"metadata"`
	}

	type TestResult struct {
		Instructions [][]string `json:"instructions"`
	}

	var testcases []Testcase
	testcases = append(testcases, produceSameBlockTradesTestcases...)
	for _, testcase := range testcases {
		t.Run(testcase.Name, func(t *testing.T) {
			var testdata TestData
			err := json.Unmarshal([]byte(testcase.Data), &testdata)
			NoError(t, err)

			var mds []metadataCommon.Metadata
			for _, md := range testdata.Metadata {
				var temp metadataPdexv3.TradeRequest = md
				mds = append(mds, &temp)
			}

			env := skipToProduce(mds, 0)
			testState := mustReadState("test_state.json")
			testState.params = &Params{
				DefaultFeeRateBPS: 30,
			}
			temp := &StateFormatter{}
			temp.FromState(testState)

			instructions, err := testState.BuildInstructions(env)
			NoError(t, err)

			encodedResult, _ := json.Marshal(TestResult{instructions})
			Equal(t, testcase.Expected, string(encodedResult))
		})
	}
}

func TestProcessTrade(t *testing.T) {
	setTestTradeConfig()
	type TestData struct {
		Instructions [][]string `json:"instructions"`
	}

	type TestResult StateFormatter

	var testcases []Testcase
	testcases = append(testcases, processTradeTestcases...)
	for _, testcase := range testcases {
		t.Run(testcase.Name, func(t *testing.T) {
			var testdata TestData
			err := json.Unmarshal([]byte(testcase.Data), &testdata)
			NoError(t, err)

			env := skipToProcess(testdata.Instructions)
			testState := mustReadState("test_state.json")
			err = testState.Process(env)
			NoError(t, err)

			temp := (&StateFormatter{}).FromState(testState)
			encodedResult, _ := json.Marshal(TestResult(*temp))
			Equal(t, testcase.Expected, string(encodedResult))
		})
	}
}

func TestBuildResponseTrade(t *testing.T) {
	setTestTradeConfig()
	type TestData struct {
		Instructions [][]string `json:"instructions"`
	}

	type TestResult struct {
		Tx metadataCommon.Transaction `json:"tx"`
	}

	var testcases []Testcase
	testcases = append(testcases, buildResponseTradeTestcases...)
	var blankPrivateKey privacy.PrivateKey = make([]byte, 32)
	// use a fixed, non-zero private key for testing
	blankPrivateKey[3] = 10

	var blankShardID byte = 0
	for _, testcase := range testcases {
		t.Run(testcase.Name, func(t *testing.T) {
			var testdata TestData
			err := json.Unmarshal([]byte(testcase.Data), &testdata)
			NoError(t, err)

			myInstruction := testdata.Instructions[0]
			metaType, err := strconv.Atoi(myInstruction[0])
			NoError(t, err)

			tx, err := (&TxBuilderV2{}).Build(
				metaType,
				myInstruction,
				&blankPrivateKey,
				blankShardID,
				testDB,
			)
			NoError(t, err)
			txv2, ok := tx.(*transaction.TxTokenVersion2)
			True(t, ok)
			mintedCoin, ok := txv2.TokenData.
				Proof.GetOutputCoins()[0].(*privacy.CoinV2)
			True(t, ok)

			var expectedTx transaction.TxTokenVersion2
			err = json.Unmarshal([]byte(testcase.Expected), &expectedTx)
			NoError(t, err)
			expectedMintedCoin, ok := expectedTx.TokenData.Proof.GetOutputCoins()[0].(*privacy.CoinV2)
			True(t, ok)
			// check token id, receiver & value
			Equal(t, expectedTx.TokenData.PropertyID, txv2.TokenData.PropertyID)
			True(t, bytes.Equal(expectedMintedCoin.GetPublicKey().ToBytesS(),
				mintedCoin.GetPublicKey().ToBytesS()))
			Equal(t, expectedMintedCoin.GetValue(), mintedCoin.GetValue())
		})
	}
}

func TestGetPRVRate(t *testing.T) {
	setTestTradeConfig()
	type TestData map[string]*PoolPairState

	var testcases []Testcase
	testcases = append(testcases, prvRateTestcases...)
	for _, testcase := range testcases {
		t.Run(testcase.Name, func(t *testing.T) {
			var testdata TestData
			err := json.Unmarshal([]byte(testcase.Data), &testdata)
			NoError(t, err)

			chosenPoolMap := getTokenPricesAgainstPRV(testdata, 0)
			Equal(t, len(chosenPoolMap), 1) // testcases must be 1-pair only
			for _, result := range chosenPoolMap {
				encodedResult, _ := json.Marshal(result)
				Equal(t, testcase.Expected, string(encodedResult))
			}
		})
	}
}

func TestIgnoreSmallPRVPool(t *testing.T) {
	setTestTradeConfig()
	type TestData struct {
		MinPRVReserve uint64
		Pools         map[string]*PoolPairState
	}

	var testcases []Testcase
	testcases = append(testcases, minPRVReserveTestcases...)
	for _, testcase := range testcases {
		t.Run(testcase.Name, func(t *testing.T) {
			var testdata TestData
			err := json.Unmarshal([]byte(testcase.Data), &testdata)
			NoError(t, err)

			chosenPoolMap := getTokenPricesAgainstPRV(testdata.Pools, testdata.MinPRVReserve)
			if len(chosenPoolMap) == 0 {
				Equal(t, testcase.Expected, "")
			} else {
				Equal(t, len(chosenPoolMap), 1) // testcases must be 1-pair only
				for _, result := range chosenPoolMap {
					encodedResult, _ := json.Marshal(result)
					Equal(t, testcase.Expected, string(encodedResult))
				}
			}
		})
	}
}

func skipToProduce(mds []metadataCommon.Metadata, shardID byte) StateEnvironment {
	var txLst []metadataCommon.Transaction
	for _, md := range mds {
		mytx := &transaction.TxVersion2{}
		valEnv := tx_generic.DefaultValEnv()
		valEnv = tx_generic.WithShardID(valEnv, int(shardID))
		mytx.SetMetadata(md)
		mytx.SetValidationEnv(valEnv)
		txLst = append(txLst, mytx)
	}

	return NewStateEnvBuilder().
		BuildPrevBeaconHeight(10).
		BuildListTxs(map[byte][]metadataCommon.Transaction{shardID: txLst}).
		BuildBCHeightBreakPointPrivacyV2(0).
		BuildStateDB(testDB).
		Build()
}

func skipToProcess(instructions [][]string) StateEnvironment {
	return NewStateEnvBuilder().
		BuildBeaconInstructions(instructions).
		BuildStateDB(testDB).
		Build()
}

type Testcase struct {
	Name          string `json:"name"`
	Data          string `json:"data"`
	Expected      string `json:"expected"`
	ExpectSuccess bool   `json:"expectSuccess"`
}

// format a pool, discarding data irrelevant to this test
type PoolFormatter struct {
	State     *rawdbv2.Pdexv3PoolPair `json:"state"`
	Orderbook Orderbook               `json:"orderbook"`
}

type StateFormatter struct {
	PoolPairs map[string]PoolFormatter `json:"poolPairs"`
}

func (sf *StateFormatter) State() *stateV2 {
	s := newStateV2WithValue(nil, nil, make(map[string]*PoolPairState),
		&Params{}, nil, make(map[string]uint64))
	for k, v := range sf.PoolPairs {
		s.poolPairs[k] = &PoolPairState{state: *v.State, orderbook: v.Orderbook}
	}
	return s
}

func (sf *StateFormatter) FromState(s *stateV2) *StateFormatter {
	sf.PoolPairs = make(map[string]PoolFormatter)
	for k, v := range s.poolPairs {
		sf.PoolPairs[k] = PoolFormatter{State: &v.state, Orderbook: v.orderbook}
	}
	return sf
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

	var temp StateFormatter
	err = json.Unmarshal(raw, &temp)
	if err != nil {
		panic(err)
	}
	return temp.State()
}

var sortOrderTestcases = mustReadTestcases("sort_orders.json")
var produceTradeTestcases = mustReadTestcases("produce_trade.json")
var produceSameBlockTradesTestcases = mustReadTestcases("produce_same_block_trades.json")
var processTradeTestcases = mustReadTestcases("process_trade.json")
var buildResponseTradeTestcases = mustReadTestcases("response_trade.json")
var prvRateTestcases = mustReadTestcases("prv_rate.json")
var minPRVReserveTestcases = mustReadTestcases("min_prv_reserve.json")
