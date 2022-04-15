package bridgeagg

import (
	"encoding/json"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataBridge "github.com/incognitochain/incognito-chain/metadata/bridge"
	"github.com/stretchr/testify/suite"
)

type ShieldTestCase struct {
	Metadatas             []*metadataBridge.ShieldRequest `json:"metadatas"`
	ExpectedInstructions  [][]string                      `json:"expected_instructions"`
	UnifiedTokens         map[common.Hash]map[uint]*Vault `json:"unified_tokens"`
	ExpectedUnifiedTokens map[common.Hash]map[uint]*Vault `json:"expected_unified_tokens"`
	TxID                  common.Hash                     `json:"tx_id"`
}

type ShieldTestSuite struct {
	suite.Suite
	testCases            []ShieldTestCase
	currentTestCaseIndex int
	actualResults        []ActualResult

	sdb *statedb.StateDB
}

func (s *ShieldTestSuite) SetupSuite() {
	config.AbortParam()
	config.Param().BridgeAggParam.BaseDecimal = 9

	rawTestCases, _ := readTestCases("convert.json")
	err := json.Unmarshal(rawTestCases, &s.testCases)
	if err != nil {
		panic(err)
	}
	s.currentTestCaseIndex = -1
}

/*func (s *ShieldTestSuite) SetupTest() {*/
/*dbPath, err := ioutil.TempDir(os.TempDir(), "bridgeagg_test_statedb_")*/
/*if err != nil {*/
/*panic(err)*/
/*}*/
/*diskBD, _ := incdb.Open("leveldb", dbPath)*/
/*warperDBStatedbTest := statedb.NewDatabaseAccessWarper(diskBD)*/
/*emptyRoot := common.HexToHash(common.HexEmptyRoot)*/
/*sDB, _ := statedb.NewWithPrefixTrie(emptyRoot, warperDBStatedbTest)*/
/*s.sdb = sDB*/

/*s.currentTestCaseIndex++*/
/*testCase := s.testCases[s.currentTestCaseIndex]*/
/*actions := []string{}*/
/*for _, v := range testCase.Metadatas {*/
/*content, err := metadataCommon.NewActionWithValue(v, testCase.TxID, nil).StringSlice(metadataCommon.BridgeAggConvertTokenToUnifiedTokenRequestMeta)*/
/*if err != nil {*/
/*panic(err)*/
/*}*/
/*err = statedb.UpdateBridgeTokenInfo(sDB, v.TokenID, []byte("123"), false, v.Amount+100, "+")*/
/*if err != nil {*/
/*panic(err)*/
/*}*/
/*actions = append(actions, content[1])*/
/*}*/

/*assert := c.Assert()*/
/*env := &stateEnvironment{*/
/*beaconHeight: 10,*/
/*stateDBs: map[int]*statedb.StateDB{*/
/*common.BeaconChainID: c.sdb,*/
/*},*/
/*convertActions: [][]string{actions},*/
/*}*/
/*state := NewState()*/
/*state.unifiedTokenInfos = testCase.UnifiedTokens*/
/*producerState := state.Clone()*/
/*processorState := state.Clone()*/
/*actualInstructions, err := producerState.BuildInstructions(env)*/
/*assert.Nil(err, fmt.Sprintf("Error in build instructions %v", err))*/
/*err = processorState.Process(actualInstructions, c.sdb)*/
/*assert.Nil(err, fmt.Sprintf("Error in process instructions %v", err))*/
/*c.actualResults = append(c.actualResults, ActualResult{*/
/*Instructions:   actualInstructions,*/
/*ProducerState:  producerState,*/
/*ProcessorState: processorState,*/
/*})*/
/*}*/

/*func (c *ConvertTestSuite) TestAcceptedConvert() {*/
/*assert := c.Assert()*/
/*testCase := c.testCases[c.currentTestCaseIndex]*/
/*actualResult := c.actualResults[c.currentTestCaseIndex]*/
/*expectedState := NewState()*/
/*expectedState.unifiedTokenInfos = testCase.ExpectedUnifiedTokens*/
/*assert.Equal(testCase.ExpectedInstructions, actualResult.Instructions, fmt.Errorf("Expected instructions %v but get %v", actualResult.Instructions, testCase.ExpectedInstructions).Error())*/
/*assert.Equal(expectedState, actualResult.ProducerState, fmt.Errorf("Expected producer state %v but get %v", actualResult.ProducerState, expectedState).Error())*/
/*assert.Equal(expectedState, actualResult.ProducerState, fmt.Errorf("Expected processor state %v but get %v", actualResult.ProcessorState, expectedState).Error())*/
/*}*/

/*func (c *ConvertTestSuite) TestRejectedConvert() {*/
/*assert := c.Assert()*/
/*testCase := c.testCases[c.currentTestCaseIndex]*/
/*actualResult := c.actualResults[c.currentTestCaseIndex]*/
/*expectedState := NewState()*/
/*expectedState.unifiedTokenInfos = testCase.ExpectedUnifiedTokens*/
/*assert.Equal(testCase.ExpectedInstructions, actualResult.Instructions, fmt.Errorf("Expected instructions %v but get %v", actualResult.Instructions, testCase.ExpectedInstructions).Error())*/
/*assert.Equal(expectedState, actualResult.ProducerState, fmt.Errorf("Expected producer state %v but get %v", actualResult.ProducerState, expectedState).Error())*/
/*assert.Equal(expectedState, actualResult.ProducerState, fmt.Errorf("Expected processor state %v but get %v", actualResult.ProcessorState, expectedState).Error())*/
/*}*/

/*func (c *ConvertTestSuite) TestRejectedThenAcceptedConvert() {*/
/*assert := c.Assert()*/
/*testCase := c.testCases[c.currentTestCaseIndex]*/
/*actualResult := c.actualResults[c.currentTestCaseIndex]*/
/*expectedState := NewState()*/
/*expectedState.unifiedTokenInfos = testCase.ExpectedUnifiedTokens*/
/*assert.Equal(testCase.ExpectedInstructions, actualResult.Instructions, fmt.Errorf("Expected instructions %v but get %v", actualResult.Instructions, testCase.ExpectedInstructions).Error())*/
/*assert.Equal(expectedState, actualResult.ProducerState, fmt.Errorf("Expected producer state %v but get %v", actualResult.ProducerState, expectedState).Error())*/
/*assert.Equal(expectedState, actualResult.ProducerState, fmt.Errorf("Expected processor state %v but get %v", actualResult.ProcessorState, expectedState).Error())*/
/*}*/

/*func TestConvertTestSuite(t *testing.T) {*/
/*suite.Run(t, new(ConvertTestSuite))*/
/*}*/
