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
	metadataMocks "github.com/incognitochain/incognito-chain/metadata/common/mocks"
	"github.com/stretchr/testify/suite"
)

type ModifyRewardReserveTestCase struct {
	Metadatas        []*metadataBridge.ModifyRewardReserve `json:"metadatas"`
	ExpectedStatuses []ModifyRewardReserveStatus           `json:"expected_statuses"`
	ActualStatues    []ModifyRewardReserveStatus
	TestCase
}

type ModifyRewardReserveTestSuite struct {
	testCases map[string]*ModifyRewardReserveTestCase
	TestSuite
}

func (m *ModifyRewardReserveTestSuite) SetupSuite() {
	config.AbortParam()
	config.Param().BridgeAggParam.BaseDecimal = 9

	rawTestCases, _ := readTestCases("modify_reward_reserve.json")
	err := json.Unmarshal(rawTestCases, &m.testCases)
	if err != nil {
		panic(err)
	}
	m.actualResults = make(map[string]ActualResult)
}

func (m *ModifyRewardReserveTestSuite) SetupTest() {
	dbPath, err := ioutil.TempDir(os.TempDir(), "bridgeagg_test_statedb_")
	if err != nil {
		panic(err)
	}
	diskBD, _ := incdb.Open("leveldb", dbPath)
	warperDBStatedbTest := statedb.NewDatabaseAccessWarper(diskBD)
	emptyRoot := common.HexToHash(common.HexEmptyRoot)
	sDB, _ := statedb.NewWithPrefixTrie(emptyRoot, warperDBStatedbTest)
	m.sDB = sDB
}

func (m *ModifyRewardReserveTestSuite) BeforeTest(suiteName, testName string) {
	m.currentTestCaseName = testName
	testCase := m.testCases[m.currentTestCaseName]
	actions := []string{}
	for i, v := range testCase.Metadatas {
		tx := &metadataMocks.Transaction{}
		tx.On("Hash").Return(&testCase.TxIDs[i])
		tmpActions, err := v.BuildReqActions(tx, nil, nil, nil, 0, 100)
		if err != nil {
			panic(err)
		}
		reqActions := []string{}
		for _, v := range tmpActions {
			reqActions = append(reqActions, v[1])
		}
		actions = append(actions, reqActions...)
	}
	for tokenID, v := range testCase.BridgeTokensInfo {
		err := statedb.UpdateBridgeTokenInfo(m.sDB, tokenID, v.ExternalTokenID(), false, v.Amount(), "+")
		if err != nil {
			panic(err)
		}
	}

	assert := m.Assert()
	env := &stateEnvironment{
		beaconHeight: 10,
		stateDBs: map[int]*statedb.StateDB{
			common.BeaconChainID: m.sDB,
		},
		modifyRewardReserveActions: [][]string{actions},
	}
	state := NewState()
	state.unifiedTokenInfos = testCase.UnifiedTokens
	producerState := state.Clone()
	processorState := state.Clone()
	actualInstructions, _, err := producerState.BuildInstructions(env)
	assert.Nil(err, fmt.Sprintf("Error in build instructions %v", err))
	err = processorState.Process(actualInstructions, m.sDB)
	assert.Nil(err, fmt.Sprintf("Error in process instructions %v", err))
	m.actualResults[testName] = ActualResult{
		Instructions:   actualInstructions,
		ProducerState:  producerState,
		ProcessorState: processorState,
	}
	for _, txID := range testCase.TxIDs {
		data, err := statedb.GetBridgeAggStatus(
			m.sDB,
			statedb.BridgeAggRewardReserveModifyingStatusPrefix(),
			txID.Bytes(),
		)
		assert.Nil(err, fmt.Sprintf("get bridge agg status %v", err))
		var status ModifyRewardReserveStatus
		err = json.Unmarshal(data, &status)
		assert.Nil(err, fmt.Sprintf("parse status error %v", err))
		m.testCases[m.currentTestCaseName].ActualStatues = append(m.testCases[m.currentTestCaseName].ActualStatues, status)
	}
}

func (m *ModifyRewardReserveTestSuite) TestAccepted() {
	assert := m.Assert()
	testCase := m.testCases[m.currentTestCaseName]
	actualResult := m.actualResults[m.currentTestCaseName]
	expectedState := NewState()
	expectedState.unifiedTokenInfos = testCase.ExpectedUnifiedTokens
	expectedStatuses := testCase.ExpectedStatuses
	actualStatuses := testCase.ActualStatues
	assert.Equal(testCase.ExpectedInstructions, actualResult.Instructions, fmt.Errorf("Expected instructions %v but get %v", actualResult.Instructions, testCase.ExpectedInstructions).Error())
	assert.Equal(expectedState, actualResult.ProducerState, fmt.Errorf("Expected producer state %v but get %v", expectedState, actualResult.ProducerState).Error())
	assert.Equal(expectedState, actualResult.ProcessorState, fmt.Errorf("Expected processor state %v but get %v", expectedState, actualResult.ProcessorState).Error())
	assert.Equal(actualStatuses, expectedStatuses, fmt.Errorf("Expected statuses %v but get %v", expectedStatuses, actualStatuses).Error())
}

func (m *ModifyRewardReserveTestSuite) TestRejectedInvalidTokenID() {
	assert := m.Assert()
	testCase := m.testCases[m.currentTestCaseName]
	actualResult := m.actualResults[m.currentTestCaseName]
	expectedState := NewState()
	expectedState.unifiedTokenInfos = testCase.ExpectedUnifiedTokens
	expectedStatuses := testCase.ExpectedStatuses
	actualStatuses := testCase.ActualStatues
	assert.Equal(testCase.ExpectedInstructions, actualResult.Instructions, fmt.Errorf("Expected instructions %v but get %v", actualResult.Instructions, testCase.ExpectedInstructions).Error())
	for k, v := range actualResult.ProducerState.unifiedTokenInfos {
		assert.Equal(expectedState.unifiedTokenInfos[k], v, fmt.Errorf("Expected producer state %v but get %v", expectedState, actualResult.ProducerState).Error())
	}
	assert.Equal(expectedState, actualResult.ProcessorState, fmt.Errorf("Expected processor state %v but get %v", expectedState, actualResult.ProcessorState).Error())
	assert.Equal(actualStatuses, expectedStatuses, fmt.Errorf("Expected statuses %v but get %v", expectedStatuses, actualStatuses).Error())
}

func TestModifyRewardReserveTestSuite(t *testing.T) {
	suite.Run(t, new(ModifyRewardReserveTestSuite))
}
