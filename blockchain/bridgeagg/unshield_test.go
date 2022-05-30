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

type UnshieldTestData struct {
	TestData
	Metadatas []*metadataBridge.UnshieldRequest `json:"metadatas"`
}

type UnshieldExpectedResult struct {
	ExpectedResult
	Statuses []UnshieldStatus `json:"statuses"`
}

type UnshieldActualResult struct {
	ActualResult
	Statuses []UnshieldStatus
}

type UnshieldTestCase struct {
	Data           UnshieldTestData       `json:"data"`
	ExpectedResult UnshieldExpectedResult `json:"expected_result"`
}

type UnshieldTestSuite struct {
	testCases map[string]*UnshieldTestCase
	TestSuite
	actualResults map[string]UnshieldActualResult
}

func (u *UnshieldTestSuite) SetupSuite() {
	config.AbortParam()
	config.Param().BridgeAggParam.BaseDecimal = 9

	rawTestCases, _ := readTestCases("unshield.json")
	err := json.Unmarshal(rawTestCases, &u.testCases)
	if err != nil {
		panic(err)
	}
	u.actualResults = make(map[string]UnshieldActualResult)
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
	for i, v := range testCase.Data.Metadatas {
		tx := &metadataMocks.Transaction{}
		tx.On("Hash").Return(&testCase.Data.TxIDs[i])
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
	for tokenID, v := range testCase.Data.BridgeTokensInfo {
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
	u.testCases[u.currentTestCaseName].Data.env = env
}

func (u *UnshieldTestSuite) test() {
	testCase := u.testCases[u.currentTestCaseName]
	assert := u.Assert()
	producerState := testCase.Data.State.Clone()
	producerManager := NewManagerWithValue(producerState)
	processorState := testCase.Data.State.Clone()
	processorManager := NewManagerWithValue(processorState)
	actualInstructions, accumulatedValues, err := producerManager.BuildInstructions(testCase.Data.env)
	assert.Nil(err, fmt.Sprintf("Error in build instructions %v", err))
	err = processorManager.Process(actualInstructions, u.sDB)
	assert.Nil(err, fmt.Sprintf("Error in process instructions %v", err))

	u.actualResults[u.currentTestCaseName] = UnshieldActualResult{
		ActualResult: ActualResult{
			Instructions:      actualInstructions,
			ProducerState:     producerState,
			ProcessorState:    processorState,
			AccumulatedValues: accumulatedValues,
		},
	}

	actualResult := u.actualResults[u.currentTestCaseName]
	for _, txID := range testCase.Data.TxIDs {
		data, err := statedb.GetBridgeAggStatus(
			u.sDB,
			statedb.BridgeAggUnshieldStatusPrefix(),
			txID.Bytes(),
		)
		assert.Nil(err, fmt.Sprintf("get bridge agg status %v", err))
		var status UnshieldStatus
		err = json.Unmarshal(data, &status)
		assert.Nil(err, fmt.Sprintf("parse status error %v", err))
		actualResult.Statuses = append(actualResult.Statuses, status)
	}
	u.actualResults[u.currentTestCaseName] = actualResult
}

func (u *UnshieldTestSuite) TestAcceptedEnoughVault() {
	u.test()
	assert := u.Assert()
	testCase := u.testCases[u.currentTestCaseName]
	actualResult := u.actualResults[u.currentTestCaseName]
	assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.Instructions, actualResult.Instructions))
	assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.State, actualResult.ProducerState))
	assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.State, actualResult.ProcessorState))
	assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.Statuses, actualResult.Statuses))
	assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.AccumulatedValues, actualResult.AccumulatedValues))
}

/*func (u *UnshieldTestSuite) TestAcceptedYEqualTo0NativeTokenWithdrawal() {*/
/*u.test()*/
/*assert := u.Assert()*/
/*testCase := u.testCases[u.currentTestCaseName]*/
/*actualResult := u.actualResults[u.currentTestCaseName]*/
/*assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.Instructions, actualResult.Instructions))*/
/*assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.State, actualResult.ProducerState))*/
/*assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.State, actualResult.ProcessorState))*/
/*assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.Statuses, actualResult.Statuses))*/
/*assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.AccumulatedValues, actualResult.AccumulatedValues))*/
/*}*/

/*func (u *UnshieldTestSuite) TestAcceptedYNotEqualTo0NativeTokenDepositToSC() {*/
/*u.test()*/
/*assert := u.Assert()*/
/*testCase := u.testCases[u.currentTestCaseName]*/
/*actualResult := u.actualResults[u.currentTestCaseName]*/
/*assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.Instructions, actualResult.Instructions))*/
/*assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.State, actualResult.ProducerState))*/
/*assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.State, actualResult.ProcessorState))*/
/*assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.Statuses, actualResult.Statuses))*/
/*assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.AccumulatedValues, actualResult.AccumulatedValues))*/
/*}*/

/*func (u *UnshieldTestSuite) TestAcceptedYNotEqualTo0NativeTokenWithdrawal() {*/
/*u.test()*/
/*assert := u.Assert()*/
/*testCase := u.testCases[u.currentTestCaseName]*/
/*actualResult := u.actualResults[u.currentTestCaseName]*/
/*assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.Instructions, actualResult.Instructions))*/
/*assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.State, actualResult.ProducerState))*/
/*assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.State, actualResult.ProcessorState))*/
/*assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.Statuses, actualResult.Statuses))*/
/*assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.AccumulatedValues, actualResult.AccumulatedValues))*/
/*}*/

/*func (u *UnshieldTestSuite) TestRejectedInvalidIncTokenID() {*/
/*u.test()*/
/*assert := u.Assert()*/
/*testCase := u.testCases[u.currentTestCaseName]*/
/*actualResult := u.actualResults[u.currentTestCaseName]*/
/*assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.Instructions, actualResult.Instructions))*/
/*assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.State, actualResult.ProducerState))*/
/*assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.State, actualResult.ProcessorState))*/
/*assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.Statuses, actualResult.Statuses))*/
/*assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.AccumulatedValues, actualResult.AccumulatedValues))*/
/*}*/

/*func (u *UnshieldTestSuite) TestRejectedInvalidTokenID() {*/
/*u.test()*/
/*assert := u.Assert()*/
/*testCase := u.testCases[u.currentTestCaseName]*/
/*actualResult := u.actualResults[u.currentTestCaseName]*/
/*assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.Instructions, actualResult.Instructions))*/
/*assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.State, actualResult.ProducerState))*/
/*assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.State, actualResult.ProcessorState))*/
/*assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.Statuses, actualResult.Statuses))*/
/*assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.AccumulatedValues, actualResult.AccumulatedValues))*/
/*}*/

/*func (u *UnshieldTestSuite) TestRejectedNotEnoughExpectedAmount() {*/
/*u.test()*/
/*assert := u.Assert()*/
/*testCase := u.testCases[u.currentTestCaseName]*/
/*actualResult := u.actualResults[u.currentTestCaseName]*/
/*assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.Instructions, actualResult.Instructions))*/
/*assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.State, actualResult.ProducerState))*/
/*assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.State, actualResult.ProcessorState))*/
/*assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.Statuses, actualResult.Statuses))*/
/*assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.AccumulatedValues, actualResult.AccumulatedValues))*/
/*}*/

/*func (u *UnshieldTestSuite) TestRejectedNotEnoughVaultValue() {*/
/*u.test()*/
/*assert := u.Assert()*/
/*testCase := u.testCases[u.currentTestCaseName]*/
/*actualResult := u.actualResults[u.currentTestCaseName]*/
/*assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.Instructions, actualResult.Instructions))*/
/*assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.State, actualResult.ProducerState))*/
/*assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.State, actualResult.ProcessorState))*/
/*assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.Statuses, actualResult.Statuses))*/
/*assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.AccumulatedValues, actualResult.AccumulatedValues))*/
/*}*/

/*func (u *UnshieldTestSuite) TestRejectedThenAccepted() {*/
/*u.test()*/
/*assert := u.Assert()*/
/*testCase := u.testCases[u.currentTestCaseName]*/
/*actualResult := u.actualResults[u.currentTestCaseName]*/
/*assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.Instructions, actualResult.Instructions))*/
/*assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.State, actualResult.ProducerState))*/
/*assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.State, actualResult.ProcessorState))*/
/*assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.Statuses, actualResult.Statuses))*/
/*assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.AccumulatedValues, actualResult.AccumulatedValues))*/
/*}*/

/*func (u *UnshieldTestSuite) TestRejectedByDecimalSmallerThanBaseDecimalYEqualTo0() {*/
/*u.test()*/
/*assert := u.Assert()*/
/*testCase := u.testCases[u.currentTestCaseName]*/
/*actualResult := u.actualResults[u.currentTestCaseName]*/
/*assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.Instructions, actualResult.Instructions))*/
/*assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.State, actualResult.ProducerState))*/
/*assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.State, actualResult.ProcessorState))*/
/*assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.Statuses, actualResult.Statuses))*/
/*assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.AccumulatedValues, actualResult.AccumulatedValues))*/
/*}*/

/*func (u *UnshieldTestSuite) TestRejected2UnshieldIndexes() {*/
/*u.test()*/
/*assert := u.Assert()*/
/*testCase := u.testCases[u.currentTestCaseName]*/
/*actualResult := u.actualResults[u.currentTestCaseName]*/
/*assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.Instructions, actualResult.Instructions))*/
/*assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.State, actualResult.ProducerState))*/
/*assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.State, actualResult.ProcessorState))*/
/*assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.Statuses, actualResult.Statuses))*/
/*assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.AccumulatedValues, actualResult.AccumulatedValues))*/
/*}*/

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
