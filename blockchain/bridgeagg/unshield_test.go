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

type UnshieldTestCase struct {
	TestCase
	Metadatas        []*metadataBridge.UnshieldRequest `json:"metadatas"`
	ExpectedStatuses []UnshieldStatus                  `json:"expected_statuses"`
	ActualStatues    []UnshieldStatus
}

type UnshieldTestSuite struct {
	testCases map[string]*UnshieldTestCase
	TestSuite
}

func (u *UnshieldTestSuite) SetupSuite() {
	config.AbortParam()
	config.Param().BridgeAggParam.BaseDecimal = 9

	rawTestCases, _ := readTestCases("unshield.json")
	err := json.Unmarshal(rawTestCases, &u.testCases)
	if err != nil {
		panic(err)
	}
	u.actualResults = make(map[string]ActualResult)
}

func (u *UnshieldTestSuite) SetupTest() {
	dbPath, err := ioutil.TempDir(os.TempDir(), "bridgeagg_test_statedb_")
	if err != nil {
		panic(err)
	}
	diskBD, _ := incdb.Open("leveldb", dbPath)
	warperDBStatedbTest := statedb.NewDatabaseAccessWarper(diskBD)
	emptyRoot := common.HexToHash(common.HexEmptyRoot)
	sDB, _ := statedb.NewWithPrefixTrie(emptyRoot, warperDBStatedbTest)
	u.sDB = sDB
}

func (u *UnshieldTestSuite) BeforeTest(suiteName, testName string) {
	u.currentTestCaseName = testName
	testCase := u.testCases[u.currentTestCaseName]
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
		err := statedb.UpdateBridgeTokenInfo(u.sDB, tokenID, v.ExternalTokenID(), false, v.Amount(), "+")
		if err != nil {
			panic(err)
		}
	}

	assert := u.Assert()
	_, err := u.sDB.Commit(false)
	assert.Nil(err, fmt.Sprintf("Error in commit db %v", err))
	env := &stateEnvironment{
		beaconHeight: 10,
		stateDBs: map[int]*statedb.StateDB{
			common.BeaconChainID: u.sDB,
		},
		unshieldActions: [][]string{actions},
	}
	state := NewState()
	state.unifiedTokenInfos = testCase.UnifiedTokens
	producerState := state.Clone()
	processorState := state.Clone()
	processorState.ClearCache()
	actualInstructions, _, err := producerState.BuildInstructions(env)
	assert.Nil(err, fmt.Sprintf("Error in build instructions %v", err))
	err = processorState.Process(actualInstructions, u.sDB)
	assert.Nil(err, fmt.Sprintf("Error in process instructions %v", err))
	u.actualResults[testName] = ActualResult{
		Instructions:   actualInstructions,
		ProducerState:  producerState,
		ProcessorState: processorState,
	}
	for _, txID := range testCase.TxIDs {
		data, err := statedb.GetBridgeAggStatus(
			u.sDB,
			statedb.BridgeAggUnshieldStatusPrefix(),
			txID.Bytes(),
		)
		assert.Nil(err, fmt.Sprintf("get bridge agg status %v", err))
		var status UnshieldStatus
		err = json.Unmarshal(data, &status)
		assert.Nil(err, fmt.Sprintf("parse status error %v", err))
		u.testCases[u.currentTestCaseName].ActualStatues = append(u.testCases[u.currentTestCaseName].ActualStatues, status)
	}
}

func (u *UnshieldTestSuite) TestAcceptedYEqualTo0NativeTokenDepositToSC() {
	assert := u.Assert()
	testCase := u.testCases[u.currentTestCaseName]
	actualResult := u.actualResults[u.currentTestCaseName]
	expectedState := NewState()
	expectedState.unifiedTokenInfos = testCase.ExpectedUnifiedTokens
	expectedStatuses := testCase.ExpectedStatuses
	actualStatuses := testCase.ActualStatues
	assert.Equal(testCase.ExpectedInstructions, actualResult.Instructions, fmt.Errorf("Expected instructions %v but get %v", actualResult.Instructions, testCase.ExpectedInstructions).Error())
	assert.Equal(expectedState, actualResult.ProducerState, fmt.Errorf("Expected producer state %v but get %v", expectedState, actualResult.ProducerState).Error())
	expectedState.ClearCache()
	for i, v := range testCase.TxIDs {
		for _, md := range testCase.Metadatas {
			for index := range md.Data {
				expectedState.processor.UnshieldTxsCache[common.HashH(append(v.Bytes(), common.IntToBytes(index)...))] = testCase.Metadatas[i].TokenID
			}
		}
	}
	assert.Equal(expectedState, actualResult.ProcessorState, fmt.Errorf("Expected processor state %v but get %v", expectedState, actualResult.ProcessorState).Error())
	assert.Equal(actualStatuses, expectedStatuses, fmt.Errorf("Expected statuses %v but get %v", expectedStatuses, actualStatuses).Error())
}

func (u *UnshieldTestSuite) TestAcceptedYEqualTo0NativeTokenWithdrawal() {
	assert := u.Assert()
	testCase := u.testCases[u.currentTestCaseName]
	actualResult := u.actualResults[u.currentTestCaseName]
	expectedState := NewState()
	expectedState.unifiedTokenInfos = testCase.ExpectedUnifiedTokens
	expectedStatuses := testCase.ExpectedStatuses
	actualStatuses := testCase.ActualStatues
	assert.Equal(testCase.ExpectedInstructions, actualResult.Instructions, fmt.Errorf("Expected instructions %v but get %v", actualResult.Instructions, testCase.ExpectedInstructions).Error())
	assert.Equal(expectedState, actualResult.ProducerState, fmt.Errorf("Expected producer state %v but get %v", expectedState, actualResult.ProducerState).Error())
	expectedState.ClearCache()
	for i, v := range testCase.TxIDs {
		for _, md := range testCase.Metadatas {
			for index := range md.Data {
				expectedState.processor.UnshieldTxsCache[common.HashH(append(v.Bytes(), common.IntToBytes(index)...))] = testCase.Metadatas[i].TokenID
			}
		}
	}
	assert.Equal(expectedState, actualResult.ProcessorState, fmt.Errorf("Expected processor state %v but get %v", expectedState, actualResult.ProcessorState).Error())
	assert.Equal(actualStatuses, expectedStatuses, fmt.Errorf("Expected statuses %v but get %v", expectedStatuses, actualStatuses).Error())
}

func (u *UnshieldTestSuite) TestAcceptedYNotEqualTo0NativeTokenDepositToSC() {
	assert := u.Assert()
	testCase := u.testCases[u.currentTestCaseName]
	actualResult := u.actualResults[u.currentTestCaseName]
	expectedState := NewState()
	expectedState.unifiedTokenInfos = testCase.ExpectedUnifiedTokens
	expectedStatuses := testCase.ExpectedStatuses
	actualStatuses := testCase.ActualStatues
	assert.Equal(testCase.ExpectedInstructions, actualResult.Instructions, fmt.Errorf("Expected instructions %v but get %v", actualResult.Instructions, testCase.ExpectedInstructions).Error())
	assert.Equal(expectedState, actualResult.ProducerState, fmt.Errorf("Expected producer state %v but get %v", expectedState, actualResult.ProducerState).Error())
	expectedState.ClearCache()
	for i, v := range testCase.TxIDs {
		for _, md := range testCase.Metadatas {
			for index := range md.Data {
				expectedState.processor.UnshieldTxsCache[common.HashH(append(v.Bytes(), common.IntToBytes(index)...))] = testCase.Metadatas[i].TokenID
			}
		}
	}
	assert.Equal(expectedState, actualResult.ProcessorState, fmt.Errorf("Expected processor state %v but get %v", expectedState, actualResult.ProcessorState).Error())
	assert.Equal(actualStatuses, expectedStatuses, fmt.Errorf("Expected statuses %v but get %v", expectedStatuses, actualStatuses).Error())
}

func (u *UnshieldTestSuite) TestAcceptedYNotEqualTo0NativeTokenWithdrawal() {
	assert := u.Assert()
	testCase := u.testCases[u.currentTestCaseName]
	actualResult := u.actualResults[u.currentTestCaseName]
	expectedState := NewState()
	expectedState.unifiedTokenInfos = testCase.ExpectedUnifiedTokens
	expectedStatuses := testCase.ExpectedStatuses
	actualStatuses := testCase.ActualStatues
	assert.Equal(testCase.ExpectedInstructions, actualResult.Instructions, fmt.Errorf("Expected instructions %v but get %v", actualResult.Instructions, testCase.ExpectedInstructions).Error())
	assert.Equal(expectedState, actualResult.ProducerState, fmt.Errorf("Expected producer state %v but get %v", expectedState, actualResult.ProducerState).Error())
	expectedState.ClearCache()
	for i, v := range testCase.TxIDs {
		for _, md := range testCase.Metadatas {
			for index := range md.Data {
				expectedState.processor.UnshieldTxsCache[common.HashH(append(v.Bytes(), common.IntToBytes(index)...))] = testCase.Metadatas[i].TokenID
			}
		}
	}
	assert.Equal(expectedState, actualResult.ProcessorState, fmt.Errorf("Expected processor state %v but get %v", expectedState, actualResult.ProcessorState).Error())
	assert.Equal(actualStatuses, expectedStatuses, fmt.Errorf("Expected statuses %v but get %v", expectedStatuses, actualStatuses).Error())
}

func (u *UnshieldTestSuite) TestRejectedInvalidNetworkID() {
	assert := u.Assert()
	testCase := u.testCases[u.currentTestCaseName]
	actualResult := u.actualResults[u.currentTestCaseName]
	expectedState := NewState()
	expectedState.unifiedTokenInfos = testCase.ExpectedUnifiedTokens
	expectedStatuses := testCase.ExpectedStatuses
	actualStatuses := testCase.ActualStatues
	assert.Equal(testCase.ExpectedInstructions, actualResult.Instructions, fmt.Errorf("Expected instructions %v but get %v", actualResult.Instructions, testCase.ExpectedInstructions).Error())
	assert.Equal(expectedState, actualResult.ProducerState, fmt.Errorf("Expected producer state %v but get %v", expectedState, actualResult.ProducerState).Error())
	expectedState.ClearCache()
	assert.Equal(expectedState, actualResult.ProcessorState, fmt.Errorf("Expected processor state %v but get %v", expectedState, actualResult.ProcessorState).Error())
	assert.Equal(actualStatuses, expectedStatuses, fmt.Errorf("Expected statuses %v but get %v", expectedStatuses, actualStatuses).Error())
}

func (u *UnshieldTestSuite) TestRejectedInvalidTokenID() {
	assert := u.Assert()
	testCase := u.testCases[u.currentTestCaseName]
	actualResult := u.actualResults[u.currentTestCaseName]
	expectedState := NewState()
	expectedState.unifiedTokenInfos = testCase.ExpectedUnifiedTokens
	expectedStatuses := testCase.ExpectedStatuses
	actualStatuses := testCase.ActualStatues
	assert.Equal(testCase.ExpectedInstructions, actualResult.Instructions, fmt.Errorf("Expected instructions %v but get %v", actualResult.Instructions, testCase.ExpectedInstructions).Error())
	assert.Equal(expectedState, actualResult.ProducerState, fmt.Errorf("Expected producer state %v but get %v", expectedState, actualResult.ProducerState).Error())
	expectedState.ClearCache()
	assert.Equal(expectedState, actualResult.ProcessorState, fmt.Errorf("Expected processor state %v but get %v", expectedState, actualResult.ProcessorState).Error())
	assert.Equal(actualStatuses, expectedStatuses, fmt.Errorf("Expected statuses %v but get %v", expectedStatuses, actualStatuses).Error())
}

func (u *UnshieldTestSuite) TestRejectedNotEnoughExpectedAmount() {
	assert := u.Assert()
	testCase := u.testCases[u.currentTestCaseName]
	actualResult := u.actualResults[u.currentTestCaseName]
	expectedState := NewState()
	expectedState.unifiedTokenInfos = testCase.ExpectedUnifiedTokens
	expectedStatuses := testCase.ExpectedStatuses
	actualStatuses := testCase.ActualStatues
	assert.Equal(testCase.ExpectedInstructions, actualResult.Instructions, fmt.Errorf("Expected instructions %v but get %v", actualResult.Instructions, testCase.ExpectedInstructions).Error())
	assert.Equal(expectedState, actualResult.ProducerState, fmt.Errorf("Expected producer state %v but get %v", expectedState, actualResult.ProducerState).Error())
	expectedState.ClearCache()
	assert.Equal(expectedState, actualResult.ProcessorState, fmt.Errorf("Expected processor state %v but get %v", expectedState, actualResult.ProcessorState).Error())
	assert.Equal(actualStatuses, expectedStatuses, fmt.Errorf("Expected statuses %v but get %v", expectedStatuses, actualStatuses).Error())
}

func (u *UnshieldTestSuite) TestRejectedNotEnoughVaultValue() {
	assert := u.Assert()
	testCase := u.testCases[u.currentTestCaseName]
	actualResult := u.actualResults[u.currentTestCaseName]
	expectedState := NewState()
	expectedState.unifiedTokenInfos = testCase.ExpectedUnifiedTokens
	expectedStatuses := testCase.ExpectedStatuses
	actualStatuses := testCase.ActualStatues
	assert.Equal(testCase.ExpectedInstructions, actualResult.Instructions, fmt.Errorf("Expected instructions %v but get %v", actualResult.Instructions, testCase.ExpectedInstructions).Error())
	assert.Equal(expectedState, actualResult.ProducerState, fmt.Errorf("Expected producer state %v but get %v", expectedState, actualResult.ProducerState).Error())
	expectedState.ClearCache()
	assert.Equal(expectedState, actualResult.ProcessorState, fmt.Errorf("Expected processor state %v but get %v", expectedState, actualResult.ProcessorState).Error())
	assert.Equal(actualStatuses, expectedStatuses, fmt.Errorf("Expected statuses %v but get %v", expectedStatuses, actualStatuses).Error())
}

func (u *UnshieldTestSuite) TestRejectedThenAccepted() {
	assert := u.Assert()
	testCase := u.testCases[u.currentTestCaseName]
	actualResult := u.actualResults[u.currentTestCaseName]
	expectedState := NewState()
	expectedState.unifiedTokenInfos = testCase.ExpectedUnifiedTokens
	expectedStatuses := testCase.ExpectedStatuses
	actualStatuses := testCase.ActualStatues
	assert.Equal(testCase.ExpectedInstructions, actualResult.Instructions, fmt.Errorf("Expected instructions %v but get %v", actualResult.Instructions, testCase.ExpectedInstructions).Error())
	assert.Equal(expectedState, actualResult.ProducerState, fmt.Errorf("Expected producer state %v but get %v", expectedState, actualResult.ProducerState).Error())
	expectedState.ClearCache()
	for i, v := range testCase.TxIDs {
		if i == 0 {
			continue
		}
		for index := range testCase.Metadatas[i].Data {
			expectedState.processor.UnshieldTxsCache[common.HashH(append(v.Bytes(), common.IntToBytes(index)...))] = testCase.Metadatas[i].TokenID
		}
	}
	assert.Equal(expectedState, actualResult.ProcessorState, fmt.Errorf("Expected processor state %v but get %v", expectedState, actualResult.ProcessorState).Error())
	assert.Equal(actualStatuses, expectedStatuses, fmt.Errorf("Expected statuses %v but get %v", expectedStatuses, actualStatuses).Error())
}

func (u *UnshieldTestSuite) TestRejectedByDecimalSmallerThanBaseDecimalYEqualTo0() {
	assert := u.Assert()
	testCase := u.testCases[u.currentTestCaseName]
	actualResult := u.actualResults[u.currentTestCaseName]
	expectedState := NewState()
	expectedState.unifiedTokenInfos = testCase.ExpectedUnifiedTokens
	expectedStatuses := testCase.ExpectedStatuses
	actualStatuses := testCase.ActualStatues
	assert.Equal(testCase.ExpectedInstructions, actualResult.Instructions, fmt.Errorf("Expected instructions %v but get %v", actualResult.Instructions, testCase.ExpectedInstructions).Error())
	assert.Equal(expectedState, actualResult.ProducerState, fmt.Errorf("Expected producer state %v but get %v", expectedState, actualResult.ProducerState).Error())
	expectedState.ClearCache()
	for i, v := range testCase.TxIDs {
		if i == 0 {
			continue
		}
		for index := range testCase.Metadatas[i].Data {
			expectedState.processor.UnshieldTxsCache[common.HashH(append(v.Bytes(), common.IntToBytes(index)...))] = testCase.Metadatas[i].TokenID
		}
	}
	assert.Equal(expectedState, actualResult.ProcessorState, fmt.Errorf("Expected processor state %v but get %v", expectedState, actualResult.ProcessorState).Error())
	assert.Equal(actualStatuses, expectedStatuses, fmt.Errorf("Expected statuses %v but get %v", expectedStatuses, actualStatuses).Error())
}

func (u *UnshieldTestSuite) TestRejected2UnshieldIndexes() {
	assert := u.Assert()
	testCase := u.testCases[u.currentTestCaseName]
	actualResult := u.actualResults[u.currentTestCaseName]
	expectedState := NewState()
	expectedState.unifiedTokenInfos = testCase.ExpectedUnifiedTokens
	expectedStatuses := testCase.ExpectedStatuses
	actualStatuses := testCase.ActualStatues
	assert.Equal(testCase.ExpectedInstructions, actualResult.Instructions, fmt.Errorf("Expected instructions %v but get %v", actualResult.Instructions, testCase.ExpectedInstructions).Error())
	assert.Equal(expectedState, actualResult.ProducerState, fmt.Errorf("Expected producer state %v but get %v", expectedState, actualResult.ProducerState).Error())
	expectedState.ClearCache()
	for i, v := range testCase.TxIDs {
		if i == 0 {
			continue
		}
		for index := range testCase.Metadatas[i].Data {
			expectedState.processor.UnshieldTxsCache[common.HashH(append(v.Bytes(), common.IntToBytes(index)...))] = testCase.Metadatas[i].TokenID
		}
	}
	assert.Equal(expectedState, actualResult.ProcessorState, fmt.Errorf("Expected processor state %v but get %v", expectedState, actualResult.ProcessorState).Error())
	assert.Equal(actualStatuses, expectedStatuses, fmt.Errorf("Expected statuses %v but get %v", expectedStatuses, actualStatuses).Error())
}

/*func (u *UnshieldTestSuite) TestRejectedByDecimalSmallerThanBaseDecimalYNotEqualTo0() {*/
/*assert := u.Assert()*/
/*testCase := u.testCases[u.currentTestCaseName]*/
/*actualResult := u.actualResults[u.currentTestCaseName]*/
/*expectedState := NewState()*/
/*expectedState.unifiedTokenInfos = testCase.ExpectedUnifiedTokens*/
/*expectedStatuses := testCase.ExpectedStatuses*/
/*actualStatuses := testCase.ActualStatues*/
/*assert.Equal(testCase.ExpectedInstructions, actualResult.Instructions, fmt.Errorf("Expected instructions %v but get %v", actualResult.Instructions, testCase.ExpectedInstructions).Error())*/
/*assert.Equal(expectedState, actualResult.ProducerState, fmt.Errorf("Expected producer state %v but get %v", expectedState, actualResult.ProducerState).Error())*/
/*expectedState.ClearCache()*/
/*for i, v := range testCase.TxIDs {*/
/*if i == 0 {*/
/*continue*/
/*}*/
/*for index := range testCase.Metadatas[i].Data {*/
/*expectedState.processor.UnshieldTxsCache[common.HashH(append(v.Bytes(), common.IntToBytes(index)...))] = testCase.Metadatas[i].TokenID*/
/*}*/
/*}*/
/*assert.Equal(expectedState, actualResult.ProcessorState, fmt.Errorf("Expected processor state %v but get %v", expectedState, actualResult.ProcessorState).Error())*/
/*assert.Equal(actualStatuses, expectedStatuses, fmt.Errorf("Expected statuses %v but get %v", expectedStatuses, actualStatuses).Error())*/
/*}*/

func TestUnshieldTestSuite(t *testing.T) {
	suite.Run(t, new(UnshieldTestSuite))
}
