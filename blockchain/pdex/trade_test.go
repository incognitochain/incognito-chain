package pdex

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"strconv"
	"testing"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incdb"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	metadataMocks "github.com/incognitochain/incognito-chain/metadata/common/mocks"
	metadataPdexv3 "github.com/incognitochain/incognito-chain/metadata/pdexv3"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/transaction"
	"github.com/incognitochain/incognito-chain/transaction/tx_generic"
	. "github.com/stretchr/testify/assert"
)

var (
	_                          = fmt.Print
	testWarper                 statedb.DatabaseAccessWarper
	emptyRoot                  = common.HexToHash(common.HexEmptyRoot)
	testDB                     *statedb.StateDB
	logger                     common.Logger
	DefaultTestMaxOrdersPerNft uint = 20
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
			err := json.Unmarshal(testcase.Data, &testdata)
			NoError(t, err)
			var expected TestResult
			err = json.Unmarshal(testcase.Expected, &expected)
			NoError(t, err)
			testState := mustReadState("test_state.json")

			env := skipToProduce([]metadataCommon.Metadata{&testdata.Metadata}, 0)
			instructions, err := testState.BuildInstructions(env)
			NoError(t, err)
			Equal(t, expected, TestResult{instructions})
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
			err := json.Unmarshal(testcase.Data, &testdata)
			NoError(t, err)
			var expected TestResult
			err = json.Unmarshal(testcase.Expected, &expected)
			NoError(t, err)
			testState := mustReadState("test_state.json")

			var mds []metadataCommon.Metadata
			for _, md := range testdata.Metadata {
				var temp metadataPdexv3.TradeRequest = md
				mds = append(mds, &temp)
			}

			env := skipToProduce(mds, 0)
			instructions, err := testState.BuildInstructions(env)
			NoError(t, err)
			Equal(t, expected, TestResult{instructions})
		})
	}
}

func TestProduceTradeWithFee(t *testing.T) {
	setTestTradeConfig()
	type TestData struct {
		Metadata metadataPdexv3.TradeRequest `json:"metadata"`
	}

	type TestResult struct {
		Instructions [][]string `json:"instructions"`
	}

	var testcases []Testcase
	testcases = append(testcases, produceTradeWithFeeTestCases...)
	for _, testcase := range testcases {
		t.Run(testcase.Name, func(t *testing.T) {
			var testdata TestData
			err := json.Unmarshal(testcase.Data, &testdata)
			NoError(t, err)
			var expected TestResult
			err = json.Unmarshal(testcase.Expected, &expected)
			NoError(t, err)
			testState := mustReadState("test_multiple_pools_state.json")

			env := skipToProduce([]metadataCommon.Metadata{&testdata.Metadata}, 0)
			instructions, err := testState.BuildInstructions(env)
			NoError(t, err)
			Equal(t, expected, TestResult{instructions})
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
			err := json.Unmarshal(testcase.Data, &testdata)
			NoError(t, err)
			var expected TestResult
			err = json.Unmarshal(testcase.Expected, &expected)
			NoError(t, err)
			testState := mustReadState("test_state.json")

			env := skipToProcess(testdata.Instructions)
			err = testState.Process(env)
			NoError(t, err)
			result := (&StateFormatter{}).FromState(testState)
			EqualValues(t, expected, *result)
		})
	}
}

func TestProcessOrderReward(t *testing.T) {
	setTestTradeConfig()
	type TestData struct {
		Instructions [][]string `json:"instructions"`
	}

	type TestResult StateFormatter

	var testcases []Testcase
	testcases = append(testcases, processTradeOrderRewardTestcases...)
	for _, testcase := range testcases {
		t.Run(testcase.Name, func(t *testing.T) {
			var testdata TestData
			err := json.Unmarshal(testcase.Data, &testdata)
			NoError(t, err)
			var expected TestResult
			err = json.Unmarshal(testcase.Expected, &expected)
			NoError(t, err)
			testState := mustReadState("test_state_order_reward.json", "params.json")

			env := skipToProcess(testdata.Instructions)
			err = testState.Process(env)
			NoError(t, err)
			result := (&StateFormatter{}).FromState(testState)

			EqualValues(t, expected, *result)
		})
	}
}

func TestBuildResponseTrade(t *testing.T) {
	setTestTradeConfig()
	type TestData struct {
		Instructions [][]string `json:"instructions"`
	}

	type TestResult = transaction.TxTokenVersion2

	var testcases []Testcase
	testcases = append(testcases, buildResponseTradeTestcases...)
	var blankPrivateKey privacy.PrivateKey = make([]byte, 32)
	// use a fixed, non-zero private key for testing
	blankPrivateKey[3] = 10

	var blankShardID byte = 0
	for _, testcase := range testcases {
		t.Run(testcase.Name, func(t *testing.T) {
			var testdata TestData
			err := json.Unmarshal(testcase.Data, &testdata)
			NoError(t, err)
			var expected TestResult
			err = json.Unmarshal(testcase.Expected, &expected)
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
				10,
			)
			NoError(t, err)
			txv2, ok := tx.(*transaction.TxTokenVersion2)
			True(t, ok)
			mintedCoin, ok := txv2.TokenData.Proof.GetOutputCoins()[0].(*privacy.CoinV2)
			True(t, ok)

			expectedMintedCoin, ok := expected.TokenData.Proof.GetOutputCoins()[0].(*privacy.CoinV2)
			True(t, ok)
			// check token id, receiver & value
			Equal(t, expected.TokenData.PropertyID, txv2.TokenData.PropertyID)
			True(t, bytes.Equal(expectedMintedCoin.GetPublicKey().ToBytesS(),
				mintedCoin.GetPublicKey().ToBytesS()))
			Equal(t, expectedMintedCoin.GetValue(), mintedCoin.GetValue())
		})
	}
}

func TestGetPRVRate(t *testing.T) {
	setTestTradeConfig()
	type TestData map[string]*PoolPairState
	type TestResult = [3]*big.Int

	var testcases []Testcase
	testcases = append(testcases, prvRateTestcases...)
	for _, testcase := range testcases {
		t.Run(testcase.Name, func(t *testing.T) {
			var testdata TestData
			err := json.Unmarshal(testcase.Data, &testdata)
			NoError(t, err)
			var expected TestResult
			err = json.Unmarshal(testcase.Expected, &expected)
			NoError(t, err)

			chosenPoolMap := getTokenPricesAgainstPRV(testdata, 0)
			Equal(t, len(chosenPoolMap), 1) // testcases must be 1-pair only
			for _, result := range chosenPoolMap {
				Equal(t, expected, result)
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
	type TestResult = [3]*big.Int

	var testcases []Testcase
	testcases = append(testcases, minPRVReserveTestcases...)
	for _, testcase := range testcases {
		t.Run(testcase.Name, func(t *testing.T) {
			var testdata TestData
			err := json.Unmarshal(testcase.Data, &testdata)
			NoError(t, err)
			var expected TestResult
			errParseExpectedResult := json.Unmarshal(testcase.Expected, &expected)

			chosenPoolMap := getTokenPricesAgainstPRV(testdata.Pools, testdata.MinPRVReserve)
			if testcase.ExpectFailure {
				Equal(t, 0, len(chosenPoolMap))
			} else {
				NoError(t, errParseExpectedResult)
				Equal(t, 1, len(chosenPoolMap)) // testcases must be 1-pair only
				for _, result := range chosenPoolMap {
					Equal(t, expected, result)
				}
			}
		})
	}
}

func TestProduceFeeInPRVTrade(t *testing.T) {
	setTestTradeConfig()
	type TestData struct {
		Metadata metadataPdexv3.TradeRequest `json:"metadata"`
	}

	type TestResult struct {
		Instructions [][]string `json:"instructions"`
	}

	var testcases []Testcase = mustReadTestcases("produce_trade_fee_prv.json")
	for _, testcase := range testcases {
		t.Run(testcase.Name, func(t *testing.T) {
			var testdata TestData
			err := json.Unmarshal(testcase.Data, &testdata)
			NoError(t, err)
			var expected TestResult
			err = json.Unmarshal(testcase.Expected, &expected)
			NoError(t, err)
			testState := mustReadState("test_state.json")

			env := mockTxsForProducer([]metadataCommon.Metadata{&testdata.Metadata}, 0, true)
			instructions, err := testState.BuildInstructions(env)
			NoError(t, err)
			Equal(t, expected, TestResult{instructions})
		})
	}
}

func mockTxsForProducer(mds []metadataCommon.Metadata, shardID byte, burningPRV bool) StateEnvironment {
	var txLst []metadataCommon.Transaction
	for _, md := range mds {
		// for compatibility within tests, use the actual Hash() function; mock others when necessary
		mytx := &transaction.TxVersion2{}
		valEnv := tx_generic.DefaultValEnv()
		valEnv = tx_generic.WithShardID(valEnv, int(shardID))
		mytx.SetMetadata(md)

		mocktx := &metadataMocks.Transaction{}
		mocktx.On("GetMetadata").Return(md)
		mocktx.On("GetMetadataType").Return(md.GetType())
		mocktx.On("GetValidationEnv").Return(valEnv)
		mocktx.On("Hash").Return(mytx.Hash())
		// default for trade: set isBurn to true, fee is in sellToken
		var burnedPRVCoin privacy.Coin = &privacy.CoinV2{}
		if !burningPRV {
			burnedPRVCoin = nil
		}
		mocktx.On("GetTxFullBurnData").Return(true, burnedPRVCoin, nil, nil, nil)
		txLst = append(txLst, mocktx)
	}

	return NewStateEnvBuilder().
		BuildPrevBeaconHeight(10).
		BuildListTxs(map[byte][]metadataCommon.Transaction{shardID: txLst}).
		BuildBCHeightBreakPointPrivacyV2(0).
		BuildStateDB(testDB).
		Build()
}

func skipToProduce(mds []metadataCommon.Metadata, shardID byte) StateEnvironment {
	return mockTxsForProducer(mds, shardID, false)
}

func skipToProcess(instructions [][]string) StateEnvironment {
	return NewStateEnvBuilder().
		BuildBeaconInstructions(instructions).
		BuildStateDB(testDB).
		Build()
}

type Testcase struct {
	Name          string          `json:"name"`
	Data          json.RawMessage `json:"data"`
	Expected      json.RawMessage `json:"expected"`
	ExpectFailure bool            `json:"fail"`
}

// format a pool, discarding data irrelevant to this test
type PoolFormatter struct {
	State        *rawdbv2.Pdexv3PoolPair       `json:"state"`
	Orderbook    Orderbook                     `json:"orderbook"`
	OrderRewards map[string]*OrderReward       `json:"orderrewards"`
	MakingVolume map[common.Hash]*MakingVolume `json:"makingvolume"`
}

type StateFormatter struct {
	PoolPairs map[string]PoolFormatter `json:"poolPairs"`
}

func (sf *StateFormatter) State(params *Params) *stateV2 {
	s := newStateV2WithValue(
		nil, nil, make(map[string]*PoolPairState),
		params,
		nil, make(map[string]uint64),
	)
	for k, v := range sf.PoolPairs {
		s.poolPairs[k] = &PoolPairState{
			state:          *v.State,
			orderbook:      v.Orderbook,
			orderRewards:   v.OrderRewards,
			makingVolume:   v.MakingVolume,
			lpFeesPerShare: map[common.Hash]*big.Int{},
		}
	}
	return s
}

func (sf *StateFormatter) FromState(s *stateV2) *StateFormatter {
	sf.PoolPairs = make(map[string]PoolFormatter)
	for k, v := range s.poolPairs {
		sf.PoolPairs[k] = PoolFormatter{
			State:        &v.state,
			Orderbook:    v.orderbook,
			OrderRewards: v.orderRewards,
			MakingVolume: v.makingVolume,
		}
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

func mustReadState(filename string, paramsFiles ...string) *stateV2 {
	raw, err := ioutil.ReadFile("testdata/" + filename)
	if err != nil {
		panic(err)
	}

	var temp StateFormatter
	err = json.Unmarshal(raw, &temp)
	if err != nil {
		panic(err)
	}
	params := &Params{
		MaxOrdersPerNft:   DefaultTestMaxOrdersPerNft,
		DefaultFeeRateBPS: 30,
	}
	if len(paramsFiles) > 0 {
		rawParams, err := ioutil.ReadFile("testdata/" + paramsFiles[0])
		if err != nil {
			panic(err)
		}
		err = json.Unmarshal(rawParams, params)
		if err != nil {
			panic(err)
		}
	}
	return temp.State(params)
}

var sortOrderTestcases = mustReadTestcases("sort_orders.json")
var produceTradeTestcases = mustReadTestcases("produce_trade.json")
var produceSameBlockTradesTestcases = mustReadTestcases("produce_same_block_trades.json")
var produceTradeWithFeeTestCases = mustReadTestcases("produce_trade_with_fee.json")
var processTradeTestcases = mustReadTestcases("process_trade.json")
var processTradeOrderRewardTestcases = mustReadTestcases("process_trade_order_reward.json")
var buildResponseTradeTestcases = mustReadTestcases("response_trade.json")
var prvRateTestcases = mustReadTestcases("prv_rate.json")
var minPRVReserveTestcases = mustReadTestcases("min_prv_reserve.json")
