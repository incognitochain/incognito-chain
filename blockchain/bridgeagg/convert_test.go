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
	metadataBridge "github.com/incognitochain/incognito-chain/metadata/bridge"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	"github.com/stretchr/testify/suite"
)

type ConvertTestData struct {
	TestData
	Metadatas []*metadataBridge.ConvertTokenToUnifiedTokenRequest `json:"metadatas"`
}

type ConvertExpectedResult struct {
	ExpectedResult
	Statuses []ConvertStatus `json:"statuses"`
}

type ConvertActualResult struct {
	ActualResult
	Statuses []ConvertStatus
}

type ConvertTestCase struct {
	Data           ConvertTestData       `json:"data"`
	ExpectedResult ConvertExpectedResult `json:"expected_result"`
}

type ConvertTestSuite struct {
	testCases map[string]*ConvertTestCase
	TestSuite
	actualResults map[string]ConvertActualResult
}

func (c *ConvertTestSuite) SetupSuite() {
	config.AbortParam()
	config.Param().BridgeAggParam.BaseDecimal = 9

	rawTestCases, _ := readTestCases("convert.json")
	err := json.Unmarshal(rawTestCases, &c.testCases)
	if err != nil {
		panic(err)
	}
	c.actualResults = make(map[string]ConvertActualResult)
}

func (c *ConvertTestSuite) SetupTest() {
	dbPath, err := ioutil.TempDir(os.TempDir(), "bridgeagg_test_statedb_")
	if err != nil {
		panic(err)
	}
	diskBD, _ := incdb.Open("leveldb", dbPath)
	warperDBStatedbTest := statedb.NewDatabaseAccessWarper(diskBD)
	emptyRoot := common.HexToHash(common.HexEmptyRoot)
	sDB, _ := statedb.NewWithPrefixTrie(emptyRoot, warperDBStatedbTest)
	c.sDB = sDB
}

func (c *ConvertTestSuite) BeforeTest(suiteName, testName string) {
	c.currentTestCaseName = testName
	testCase := c.testCases[c.currentTestCaseName]
	actions := []string{}
	for i, v := range testCase.Data.Metadatas {
		content, err := metadataCommon.NewActionWithValue(v, testCase.Data.TxIDs[i], nil).StringSlice(metadataCommon.BridgeAggConvertTokenToUnifiedTokenRequestMeta)
		if err != nil {
			panic(err)
		}
		actions = append(actions, content[1])
	}
	for tokenID, v := range testCase.Data.BridgeTokensInfo {
		err := statedb.UpdateBridgeTokenInfo(c.sDB, tokenID, v.ExternalTokenID(), false, v.Amount(), "+")
		if err != nil {
			panic(err)
		}
	}
	env := &stateEnvironment{
		beaconHeight: 10,
		stateDBs: map[int]*statedb.StateDB{
			common.BeaconChainID: c.sDB,
		},
		convertActions:    [][]string{actions},
		accumulatedValues: testCase.Data.AccumulatedValues,
	}
	c.testCases[c.currentTestCaseName].Data.env = env
}

func (c *ConvertTestSuite) test() {
	testCase := c.testCases[c.currentTestCaseName]
	assert := c.Assert()
	producerState := testCase.Data.State.Clone()
	producerManager := NewManagerWithValue(producerState)
	processorState := testCase.Data.State.Clone()
	processorManager := NewManagerWithValue(processorState)
	actualInstructions, accumulatedValues, err := producerManager.BuildInstructions(testCase.Data.env)
	assert.Nil(err, fmt.Sprintf("Error in build instructions %v", err))
	err = processorManager.Process(actualInstructions, c.sDB)
	assert.Nil(err, fmt.Sprintf("Error in process instructions %v", err))

	c.actualResults[c.currentTestCaseName] = ConvertActualResult{
		ActualResult: ActualResult{
			Instructions:      actualInstructions,
			ProducerState:     producerState,
			ProcessorState:    processorState,
			AccumulatedValues: accumulatedValues,
		},
	}

	actualResult := c.actualResults[c.currentTestCaseName]
	for _, txID := range testCase.Data.TxIDs {
		data, err := statedb.GetBridgeAggStatus(
			c.sDB,
			statedb.BridgeAggConvertStatusPrefix(),
			txID.Bytes(),
		)
		assert.Nil(err, fmt.Sprintf("get bridge agg status %v", err))
		var status ConvertStatus
		err = json.Unmarshal(data, &status)
		assert.Nil(err, fmt.Sprintf("parse status error %v", err))
		actualResult.Statuses = append(actualResult.Statuses, status)
	}
	c.actualResults[c.currentTestCaseName] = actualResult
}

func (c *ConvertTestSuite) AfterTest(suiteName, testName string) {
	assert := c.Assert()
	_, err := c.sDB.Commit(false)
	assert.NoError(err, fmt.Sprintf("Error in commit db %v", err))
	bridgeTokenInfos := make(map[common.Hash]*rawdbv2.BridgeTokenInfo)
	tokens, err := statedb.GetBridgeTokens(c.sDB)
	assert.NoError(err, fmt.Sprintf("Error in get bridge tokens from db %v", err))
	for _, token := range tokens {
		bridgeTokenInfos[*token.TokenID] = token
	}
	expectedBridgeTokensInfo := c.testCases[c.currentTestCaseName].ExpectedResult.BridgeTokensInfo
	assert.Equal(expectedBridgeTokensInfo, bridgeTokenInfos, fmt.Errorf("Expected bridgeTokenInfos %v but get %v", expectedBridgeTokensInfo, bridgeTokenInfos).Error())
}

func (c *ConvertTestSuite) TestAcceptedGreaterBaseDecimal() {
	c.test()
	assert := c.Assert()
	testCase := c.testCases[c.currentTestCaseName]
	actualResult := c.actualResults[c.currentTestCaseName]
	assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.Instructions, actualResult.Instructions))
	assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.State, actualResult.ProducerState))
	assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.State, actualResult.ProcessorState))
	assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.Statuses, actualResult.Statuses))
	assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.AccumulatedValues, actualResult.AccumulatedValues))
}

func (c *ConvertTestSuite) TestAcceptedSmallerBaseDecimal() {
	c.test()
	assert := c.Assert()
	testCase := c.testCases[c.currentTestCaseName]
	actualResult := c.actualResults[c.currentTestCaseName]
	assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.Instructions, actualResult.Instructions))
	assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.State, actualResult.ProducerState))
	assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.State, actualResult.ProcessorState))
	assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.Statuses, actualResult.Statuses))
	assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.AccumulatedValues, actualResult.AccumulatedValues))
}

func (c *ConvertTestSuite) TestRejectedByInvalidTokenID() {
	c.test()
	assert := c.Assert()
	testCase := c.testCases[c.currentTestCaseName]
	actualResult := c.actualResults[c.currentTestCaseName]
	assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.Instructions, actualResult.Instructions))
	assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.State, actualResult.ProducerState))
	assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.State, actualResult.ProcessorState))
	assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.Statuses, actualResult.Statuses))
	assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.AccumulatedValues, actualResult.AccumulatedValues))
}

func (c *ConvertTestSuite) TestRejectedByInvalidUnifiedTokenID() {
	c.test()
	assert := c.Assert()
	testCase := c.testCases[c.currentTestCaseName]
	actualResult := c.actualResults[c.currentTestCaseName]
	assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.Instructions, actualResult.Instructions))
	assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.State, actualResult.ProducerState))
	assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.State, actualResult.ProcessorState))
	assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.Statuses, actualResult.Statuses))
	assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.AccumulatedValues, actualResult.AccumulatedValues))
}

func (c *ConvertTestSuite) TestRejectedThenAccepted() {
	c.test()
	assert := c.Assert()
	testCase := c.testCases[c.currentTestCaseName]
	actualResult := c.actualResults[c.currentTestCaseName]
	assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.Instructions, actualResult.Instructions))
	assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.State, actualResult.ProducerState))
	assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.State, actualResult.ProcessorState))
	assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.Statuses, actualResult.Statuses))
	assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.AccumulatedValues, actualResult.AccumulatedValues))
}

func TestConvertTestSuite(t *testing.T) {
	suite.Run(t, new(ConvertTestSuite))
}
