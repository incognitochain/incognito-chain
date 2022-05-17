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

type ConvertTestCase struct {
	Metadatas        []*metadataBridge.ConvertTokenToUnifiedTokenRequest `json:"metadatas"`
	ExpectedStatuses []ConvertStatus                                     `json:"expected_statuses"`
	ActualStatues    []ConvertStatus
	TestCase
}

type ConvertTestSuite struct {
	testCases map[string]*ConvertTestCase
	TestSuite
}

func (c *ConvertTestSuite) SetupSuite() {
	config.AbortParam()
	config.Param().BridgeAggParam.BaseDecimal = 9

	rawTestCases, _ := readTestCases("convert.json")
	err := json.Unmarshal(rawTestCases, &c.testCases)
	if err != nil {
		panic(err)
	}
	c.actualResults = make(map[string]ActualResult)
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
	for i, v := range testCase.Metadatas {
		content, err := metadataCommon.NewActionWithValue(v, testCase.TxIDs[i], nil).StringSlice(metadataCommon.BridgeAggConvertTokenToUnifiedTokenRequestMeta)
		if err != nil {
			panic(err)
		}
		actions = append(actions, content[1])
	}
	for tokenID, v := range testCase.BridgeTokensInfo {
		err := statedb.UpdateBridgeTokenInfo(c.sDB, tokenID, v.ExternalTokenID(), false, v.Amount(), "+")
		if err != nil {
			panic(err)
		}
	}

	assert := c.Assert()
	env := &stateEnvironment{
		beaconHeight: 10,
		stateDBs: map[int]*statedb.StateDB{
			common.BeaconChainID: c.sDB,
		},
		convertActions: [][]string{actions},
	}
	state := NewState()
	state.unifiedTokenVaults = testCase.UnifiedTokens
	producerState := state.Clone()
	processorState := state.Clone()
	actualInstructions, _, err := producerState.BuildInstructions(env)
	assert.Nil(err, fmt.Sprintf("Error in build instructions %v", err))
	err = processorState.Process(actualInstructions, c.sDB)
	assert.Nil(err, fmt.Sprintf("Error in process instructions %v", err))

	c.actualResults[testName] = ActualResult{
		Instructions:   actualInstructions,
		ProducerState:  producerState,
		ProcessorState: processorState,
	}
	for _, txID := range testCase.TxIDs {
		data, err := statedb.GetBridgeAggStatus(
			c.sDB,
			statedb.BridgeAggConvertStatusPrefix(),
			txID.Bytes(),
		)
		assert.Nil(err, fmt.Sprintf("get bridge agg status %v", err))
		var status ConvertStatus
		err = json.Unmarshal(data, &status)
		assert.Nil(err, fmt.Sprintf("parse status error %v", err))
		c.testCases[c.currentTestCaseName].ActualStatues = append(c.testCases[c.currentTestCaseName].ActualStatues, status)
	}
}

func (c *ConvertTestSuite) TestAccepted() {
	assert := c.Assert()
	testCase := c.testCases[c.currentTestCaseName]
	actualResult := c.actualResults[c.currentTestCaseName]
	expectedState := NewState()
	expectedState.unifiedTokenVaults = testCase.ExpectedUnifiedTokens
	expectedStatuses := testCase.ExpectedStatuses
	actualStatuses := testCase.ActualStatues
	assert.Equal(testCase.ExpectedInstructions, actualResult.Instructions, fmt.Errorf("Expected instructions %v but get %v", actualResult.Instructions, testCase.ExpectedInstructions).Error())
	for k, v := range actualResult.ProducerState.unifiedTokenVaults {
		assert.Equal(expectedState.unifiedTokenVaults[k], v, fmt.Errorf("Expected producer state %v but get %v", expectedState, actualResult.ProducerState).Error())
	}
	assert.Equal(expectedState, actualResult.ProcessorState, fmt.Errorf("Expected processor state %v but get %v", expectedState, actualResult.ProcessorState).Error())
	assert.Equal(actualStatuses, expectedStatuses, fmt.Errorf("Expected statuses %v but get %v", expectedStatuses, actualStatuses).Error())
}

/*func (c *ConvertTestSuite) TestRejectedByInvalidTokenID() {*/
/*assert := c.Assert()*/
/*testCase := c.testCases[c.currentTestCaseName]*/
/*actualResult := c.actualResults[c.currentTestCaseName]*/
/*expectedState := NewState()*/
/*expectedState.unifiedTokenInfos = testCase.ExpectedUnifiedTokens*/
/*expectedStatuses := testCase.ExpectedStatuses*/
/*actualStatuses := testCase.ActualStatues*/
/*assert.Equal(testCase.ExpectedInstructions, actualResult.Instructions, fmt.Errorf("Expected instructions %v but get %v", actualResult.Instructions, testCase.ExpectedInstructions).Error())*/
/*for k, v := range actualResult.ProducerState.unifiedTokenInfos {*/
/*assert.Equal(expectedState.unifiedTokenInfos[k], v, fmt.Errorf("Expected producer state %v but get %v", expectedState, actualResult.ProducerState).Error())*/
/*}*/
/*assert.Equal(expectedState, actualResult.ProcessorState, fmt.Errorf("Expected processor state %v but get %v", expectedState, actualResult.ProcessorState).Error())*/
/*assert.Equal(actualStatuses, expectedStatuses, fmt.Errorf("Expected statuses %v but get %v", expectedStatuses, actualStatuses).Error())*/
/*}*/

/*func (c *ConvertTestSuite) TestRejectedByInvalidUnifiedTokenID() {*/
/*assert := c.Assert()*/
/*testCase := c.testCases[c.currentTestCaseName]*/
/*actualResult := c.actualResults[c.currentTestCaseName]*/
/*expectedState := NewState()*/
/*expectedState.unifiedTokenInfos = testCase.ExpectedUnifiedTokens*/
/*expectedStatuses := testCase.ExpectedStatuses*/
/*actualStatuses := testCase.ActualStatues*/
/*assert.Equal(testCase.ExpectedInstructions, actualResult.Instructions, fmt.Errorf("Expected instructions %v but get %v", actualResult.Instructions, testCase.ExpectedInstructions).Error())*/
/*for k, v := range actualResult.ProducerState.unifiedTokenInfos {*/
/*assert.Equal(expectedState.unifiedTokenInfos[k], v, fmt.Errorf("Expected producer state %v but get %v", expectedState, actualResult.ProducerState).Error())*/
/*}*/
/*assert.Equal(expectedState, actualResult.ProcessorState, fmt.Errorf("Expected processor state %v but get %v", expectedState, actualResult.ProcessorState).Error())*/
/*assert.Equal(actualStatuses, expectedStatuses, fmt.Errorf("Expected statuses %v but get %v", expectedStatuses, actualStatuses).Error())*/
/*}*/

/*func (c *ConvertTestSuite) TestRejectedThenAccepted() {*/
/*assert := c.Assert()*/
/*testCase := c.testCases[c.currentTestCaseName]*/
/*actualResult := c.actualResults[c.currentTestCaseName]*/
/*expectedState := NewState()*/
/*expectedState.unifiedTokenInfos = testCase.ExpectedUnifiedTokens*/
/*expectedStatuses := testCase.ExpectedStatuses*/
/*actualStatuses := testCase.ActualStatues*/
/*assert.Equal(testCase.ExpectedInstructions, actualResult.Instructions, fmt.Errorf("Expected instructions %v but get %v", actualResult.Instructions, testCase.ExpectedInstructions).Error())*/
/*for k, v := range actualResult.ProducerState.unifiedTokenInfos {*/
/*assert.Equal(expectedState.unifiedTokenInfos[k], v, fmt.Errorf("Expected producer state %v but get %v", expectedState, actualResult.ProducerState).Error())*/
/*}*/
/*assert.Equal(expectedState, actualResult.ProcessorState, fmt.Errorf("Expected processor state %v but get %v", expectedState, actualResult.ProcessorState).Error())*/
/*assert.Equal(actualStatuses, expectedStatuses, fmt.Errorf("Expected statuses %v but get %v", expectedStatuses, actualStatuses).Error())*/
/*}*/

func TestConvertTestSuite(t *testing.T) {
	suite.Run(t, new(ConvertTestSuite))
}
