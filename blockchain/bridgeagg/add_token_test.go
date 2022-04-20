package bridgeagg

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/stretchr/testify/suite"
)

type AddTokenTestCase struct {
	UnifiedTokens map[uint64]map[common.Hash]map[uint]config.Vault `json:"unified_tokens"`
	TestCase
	ExpectedAccumulatedValues *metadata.AccumulatedValues `json:"expected_accumulated_values"`
	ActualAccumulatedValues   *metadata.AccumulatedValues
}

func (a *AddTokenTestSuite) BeforeTest(suiteName, testName string) {
	fmt.Println("suiteName:", suiteName)
	fmt.Println("testName:", testName)
}

type AddTokenTestSuite struct {
	testCases map[string]AddTokenTestCase
	TestSuite
}

func (a *AddTokenTestSuite) SetupSuite() {
	rawTestCases, _ := readTestCases("add_token.json")
	err := json.Unmarshal(rawTestCases, &a.testCases)
	if err != nil {
		panic(err)
	}
}

/*func (a *AddTokenTestSuite) SetupTest() {*/
/*dbPath, err := ioutil.TempDir(os.TempDir(), "bridgeagg_test_statedb_")*/
/*if err != nil {*/
/*panic(err)*/
/*}*/
/*diskBD, _ := incdb.Open("leveldb", dbPath)*/
/*warperDBStatedbTest := statedb.NewDatabaseAccessWarper(diskBD)*/
/*emptyRoot := common.HexToHash(common.HexEmptyRoot)*/
/*sDB, _ := statedb.NewWithPrefixTrie(emptyRoot, warperDBStatedbTest)*/

/*m.currentTestCaseIndex++*/
/*testCase := m.testCases[m.currentTestCaseIndex]*/
/*actions := []string{}*/
/*for i, v := range testCase.Metadatas {*/
/*tx := &metadataMocks.Transaction{}*/
/*tx.On("Hash").Return(&testCase.TxIDs[i])*/
/*tmpActions, err := v.BuildReqActions(tx, nil, nil, nil, 0, 100)*/
/*if err != nil {*/
/*panic(err)*/
/*}*/
/*reqActions := []string{}*/
/*for _, v := range tmpActions {*/
/*reqActions = append(reqActions, v[1])*/
/*}*/
/*actions = append(actions, reqActions...)*/
/*}*/
/*for tokenID, v := range testCase.BridgeTokensInfo {*/
/*err := statedb.UpdateBridgeTokenInfo(sDB, tokenID, v.ExternalTokenID(), false, v.Amount(), "+")*/
/*if err != nil {*/
/*panic(err)*/
/*}*/
/*}*/

/*assert := m.Assert()*/
/*env := &stateEnvironment{*/
/*beaconHeight: 10,*/
/*stateDBs: map[int]*statedb.StateDB{*/
/*common.BeaconChainID: sDB,*/
/*},*/
/*modifyRewardReserveActions: [][]string{actions},*/
/*}*/
/*state := NewState()*/
/*state.unifiedTokenInfos = testCase.UnifiedTokens*/
/*producerState := state.Clone()*/
/*processorState := state.Clone()*/
/*actualInstructions, _, err := producerState.BuildInstructions(env)*/
/*assert.Nil(err, fmt.Sprintf("Error in build instructions %v", err))*/
/*err = processorState.Process(actualInstructions, sDB)*/
/*assert.Nil(err, fmt.Sprintf("Error in process instructions %v", err))*/

/*m.actualResults = append(m.actualResults, ActualResult{*/
/*Instructions:   actualInstructions,*/
/*ProducerState:  producerState,*/
/*ProcessorState: processorState,*/
/*})*/
/*for _, txID := range testCase.TxIDs {*/
/*data, err := statedb.GetBridgeAggStatus(*/
/*sDB,*/
/*statedb.BridgeAggRewardReserveModifyingStatusPrefix(),*/
/*txID.Bytes(),*/
/*)*/
/*assert.Nil(err, fmt.Sprintf("get bridge agg status %v", err))*/
/*var status ModifyRewardReserveStatus*/
/*err = json.Unmarshal(data, &status)*/
/*assert.Nil(err, fmt.Sprintf("parse status error %v", err))*/
/*m.testCases[m.currentTestCaseIndex].ActualStatues = append(m.testCases[m.currentTestCaseIndex].ActualStatues, status)*/
/*}*/
/*}*/

/*func (m *ModifyRewardReserveTestSuite) TestAccepted() {*/
/*assert := m.Assert()*/
/*testCase := m.testCases[m.currentTestCaseIndex]*/
/*actualResult := m.actualResults[m.currentTestCaseIndex]*/
/*expectedState := NewState()*/
/*expectedState.unifiedTokenInfos = testCase.ExpectedUnifiedTokens*/
/*expectedStatuses := testCase.ExpectedStatuses*/
/*actualStatuses := testCase.ActualStatues*/
/*assert.Equal(testCase.ExpectedInstructions, actualResult.Instructions, fmt.Errorf("Expected instructions %v but get %v", actualResult.Instructions, testCase.ExpectedInstructions).Error())*/
/*assert.Equal(expectedState, actualResult.ProducerState, fmt.Errorf("Expected producer state %v but get %v", expectedState, actualResult.ProducerState).Error())*/
/*assert.Equal(expectedState, actualResult.ProcessorState, fmt.Errorf("Expected processor state %v but get %v", expectedState, actualResult.ProcessorState).Error())*/
/*assert.Equal(actualStatuses, expectedStatuses, fmt.Errorf("Expected statuses %v but get %v", expectedStatuses, actualStatuses).Error())*/
/*}*/

func TestAddTokenTestSuite(t *testing.T) {
	suite.Run(t, new(AddTokenTestSuite))
}
