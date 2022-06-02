package bridgeagg

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/stretchr/testify/suite"
)

type HandleWaitingUnshieldTestData struct {
	TestData
}

type HandleWaitingUnshieldExpectedResult struct {
	ExpectedResult
}

type HandleWaitingUnshieldActualResult struct {
	ActualResult
}

type HandleWaitingUnshieldTestCase struct {
	Data           HandleWaitingUnshieldTestData       `json:"data"`
	ExpectedResult HandleWaitingUnshieldExpectedResult `json:"expected_result"`
}

type HandleWaitingUnshieldTestSuite struct {
	testCases map[string]*HandleWaitingUnshieldTestCase
	TestSuite
	actualResults map[string]HandleWaitingUnshieldActualResult
}

func (h *HandleWaitingUnshieldTestSuite) SetupSuite() {
	config.AbortParam()
	config.Param().BridgeAggParam.BaseDecimal = 9

	rawTestCases, _ := readTestCases("handle_waiting_unshield_request.json")
	err := json.Unmarshal(rawTestCases, &h.testCases)
	if err != nil {
		panic(err)
	}
	h.actualResults = make(map[string]HandleWaitingUnshieldActualResult)
}

func (h *HandleWaitingUnshieldTestSuite) SetupTest() {
	dbPath, err := ioutil.TempDir(os.TempDir(), "bridgeagg_test_statedb_")
	if err != nil {
		panic(err)
	}
	diskBD, _ := incdb.Open("leveldb", dbPath)
	warperDBStatedbTest := statedb.NewDatabaseAccessWarper(diskBD)
	emptyRoot := common.HexToHash(common.HexEmptyRoot)
	sDB, _ := statedb.NewWithPrefixTrie(emptyRoot, warperDBStatedbTest)
	h.sDB = sDB
}

func (h *HandleWaitingUnshieldTestSuite) BeforeTest(suiteName, testName string) {
	h.currentTestCaseName = testName
	testCase := h.testCases[h.currentTestCaseName]

	assert := h.Assert()
	_, err := h.sDB.Commit(false)
	assert.Nil(err, fmt.Sprintf("Error in commit db %v", err))
	env := &stateEnvironment{
		beaconHeight: 10,
		stateDBs: map[int]*statedb.StateDB{
			common.BeaconChainID: h.sDB,
		},
		accumulatedValues: testCase.Data.AccumulatedValues,
	}
	h.testCases[h.currentTestCaseName].Data.env = env
}

func (h *HandleWaitingUnshieldTestSuite) test() {
	testCase := h.testCases[h.currentTestCaseName]
	assert := h.Assert()
	producerState := testCase.Data.State.Clone()
	producerManager := NewManagerWithValue(producerState)
	processorState := testCase.Data.State.Clone()
	processorManager := NewManagerWithValue(processorState)
	actualInstructions, accumulatedValues, err := producerManager.BuildInstructions(testCase.Data.env)
	assert.Nil(err, fmt.Sprintf("Error in build instructions %v", err))
	err = processorManager.Process(actualInstructions, h.sDB)
	assert.Nil(err, fmt.Sprintf("Error in process instructions %v", err))

	h.actualResults[h.currentTestCaseName] = HandleWaitingUnshieldActualResult{
		ActualResult: ActualResult{
			Instructions:      actualInstructions,
			ProducerState:     producerManager.state,
			ProcessorState:    processorManager.state,
			AccumulatedValues: accumulatedValues,
		},
	}
}

func (h *HandleWaitingUnshieldTestSuite) AfterTest(suiteName, testName string) {
	assert := h.Assert()
	_, err := h.sDB.Commit(false)
	assert.NoError(err, fmt.Sprintf("Error in commit db %v", err))
	bridgeTokenInfos := make(map[common.Hash]*rawdbv2.BridgeTokenInfo)
	tokens, err := statedb.GetBridgeTokens(h.sDB)
	assert.NoError(err, fmt.Sprintf("Error in get bridge tokens from db %v", err))
	for _, token := range tokens {
		bridgeTokenInfos[*token.TokenID] = token
	}
	expectedBridgeTokensInfo := h.testCases[h.currentTestCaseName].ExpectedResult.BridgeTokensInfo
	assert.Equal(expectedBridgeTokensInfo, bridgeTokenInfos, fmt.Errorf("Expected bridgeTokenInfos %v but get %v", expectedBridgeTokensInfo, bridgeTokenInfos).Error())
}

func (h *HandleWaitingUnshieldTestSuite) TestNotEnoughVault() {
	h.test()
	assert := h.Assert()
	testCase := h.testCases[h.currentTestCaseName]
	actualResult := h.actualResults[h.currentTestCaseName]
	assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.Instructions, actualResult.Instructions))
	assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.State, actualResult.ProducerState))
	assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.State, actualResult.ProcessorState))
	assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.AccumulatedValues, actualResult.AccumulatedValues))
}

/*func (h *HandleWaitingUnshieldTestSuite) TestWaiting() {*/
/*h.test()*/
/*assert := h.Assert()*/
/*testCase := h.testCases[h.currentTestCaseName]*/
/*actualResult := h.actualResults[h.currentTestCaseName]*/
/*assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.Instructions, actualResult.Instructions))*/
/*assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.State, actualResult.ProducerState))*/
/*assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.State, actualResult.ProcessorState))*/
/*assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.AccumulatedValues, actualResult.AccumulatedValues))*/
/*}*/

/*func (h *HandleWaitingUnshieldTestSuite) TestFilled() {*/
/*h.test()*/
/*assert := h.Assert()*/
/*testCase := h.testCases[h.currentTestCaseName]*/
/*actualResult := h.actualResults[h.currentTestCaseName]*/
/*assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.Instructions, actualResult.Instructions))*/
/*assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.State, actualResult.ProducerState))*/
/*assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.State, actualResult.ProcessorState))*/
/*assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.AccumulatedValues, actualResult.AccumulatedValues))*/
/*}*/

/*func (h *HandleWaitingUnshieldTestSuite) TestAccepted() {*/
/*h.test()*/
/*assert := h.Assert()*/
/*testCase := h.testCases[h.currentTestCaseName]*/
/*actualResult := h.actualResults[h.currentTestCaseName]*/
/*assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.Instructions, actualResult.Instructions))*/
/*assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.State, actualResult.ProducerState))*/
/*assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.State, actualResult.ProcessorState))*/
/*assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.AccumulatedValues, actualResult.AccumulatedValues))*/
/*}*/

func TestHandleWaitingUnshieldTestSuite(t *testing.T) {
	suite.Run(t, new(HandleWaitingUnshieldTestSuite))
}
