package bridgeagg

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
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
		convertActions: [][]string{actions},
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
	actualInstructions, _, err := producerManager.BuildInstructions(testCase.Data.env)
	assert.Nil(err, fmt.Sprintf("Error in build instructions %v", err))
	err = processorManager.Process(actualInstructions, c.sDB)
	assert.Nil(err, fmt.Sprintf("Error in process instructions %v", err))

	c.actualResults[c.currentTestCaseName] = ConvertActualResult{
		ActualResult: ActualResult{
			Instructions:   actualInstructions,
			ProducerState:  producerState,
			ProcessorState: processorState,
		},
	}

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
		c.testCases[c.currentTestCaseName].ExpectedResult.Statuses = append(c.testCases[c.currentTestCaseName].ExpectedResult.Statuses, status)
	}
}

func (c *ConvertTestSuite) TestAccepted() {
	c.test()
	assert := c.Assert()
	testCase := c.testCases[c.currentTestCaseName]
	actualResult := c.actualResults[c.currentTestCaseName]
	assert.Equal(testCase.ExpectedResult.Instructions, actualResult.Instructions, fmt.Errorf("Expected instructions %v but get %v", actualResult.Instructions, testCase.ExpectedResult.Instructions).Error())
	assert.Equal(testCase.ExpectedResult.State, actualResult.ProducerState, fmt.Errorf("Expected producer state %v but get %v", testCase.ExpectedResult.State, actualResult.ProducerState).Error())
	assert.Equal(testCase.ExpectedResult.State, actualResult.ProcessorState, fmt.Errorf("Expected processor state %v but get %v", testCase.ExpectedResult.State, actualResult.ProcessorState).Error())
	assert.Equal(testCase.ExpectedResult.Statuses, actualResult.Statuses, fmt.Errorf("Expected statuses %v but get %v", testCase.ExpectedResult.Statuses, actualResult.Statuses).Error())
}

func (c *ConvertTestSuite) TestRejectedByInvalidTokenID() {
	c.test()
	assert := c.Assert()
	testCase := c.testCases[c.currentTestCaseName]
	actualResult := c.actualResults[c.currentTestCaseName]
	assert.Equal(testCase.ExpectedResult.Instructions, actualResult.Instructions, fmt.Errorf("Expected instructions %v but get %v", actualResult.Instructions, testCase.ExpectedResult.Instructions).Error())
	assert.Equal(testCase.ExpectedResult.State, actualResult.ProducerState, fmt.Errorf("Expected producer state %v but get %v", testCase.ExpectedResult.State, actualResult.ProducerState).Error())
	assert.Equal(testCase.ExpectedResult.State, actualResult.ProcessorState, fmt.Errorf("Expected processor state %v but get %v", testCase.ExpectedResult.State, actualResult.ProcessorState).Error())
	assert.Equal(testCase.ExpectedResult.Statuses, actualResult.Statuses, fmt.Errorf("Expected statuses %v but get %v", testCase.ExpectedResult.Statuses, actualResult.Statuses).Error())
}

func (c *ConvertTestSuite) TestRejectedByInvalidUnifiedTokenID() {
	c.test()
	assert := c.Assert()
	testCase := c.testCases[c.currentTestCaseName]
	actualResult := c.actualResults[c.currentTestCaseName]
	assert.Equal(testCase.ExpectedResult.Instructions, actualResult.Instructions, fmt.Errorf("Expected instructions %v but get %v", actualResult.Instructions, testCase.ExpectedResult.Instructions).Error())
	assert.Equal(testCase.ExpectedResult.State, actualResult.ProducerState, fmt.Errorf("Expected producer state %v but get %v", testCase.ExpectedResult.State, actualResult.ProducerState).Error())
	assert.Equal(testCase.ExpectedResult.State, actualResult.ProcessorState, fmt.Errorf("Expected processor state %v but get %v", testCase.ExpectedResult.State, actualResult.ProcessorState).Error())
	assert.Equal(testCase.ExpectedResult.Statuses, actualResult.Statuses, fmt.Errorf("Expected statuses %v but get %v", testCase.ExpectedResult.Statuses, actualResult.Statuses).Error())
}

func (c *ConvertTestSuite) TestRejectedThenAccepted() {
	c.test()
	assert := c.Assert()
	testCase := c.testCases[c.currentTestCaseName]
	actualResult := c.actualResults[c.currentTestCaseName]
	assert.Equal(testCase.ExpectedResult.Instructions, actualResult.Instructions, fmt.Errorf("Expected instructions %v but get %v", actualResult.Instructions, testCase.ExpectedResult.Instructions).Error())
	assert.Equal(testCase.ExpectedResult.State, actualResult.ProducerState, fmt.Errorf("Expected producer state %v but get %v", testCase.ExpectedResult.State, actualResult.ProducerState).Error())
	assert.Equal(testCase.ExpectedResult.State, actualResult.ProcessorState, fmt.Errorf("Expected processor state %v but get %v", testCase.ExpectedResult.State, actualResult.ProcessorState).Error())
	assert.Equal(testCase.ExpectedResult.Statuses, actualResult.Statuses, fmt.Errorf("Expected statuses %v but get %v", testCase.ExpectedResult.Statuses, actualResult.Statuses).Error())
}

func TestConvertTestSuite(t *testing.T) {
	suite.Run(t, new(ConvertTestSuite))
}
