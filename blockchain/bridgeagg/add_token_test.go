package bridgeagg

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"testing"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/stretchr/testify/suite"
)

type ConfigVault struct {
	ExternalDecimal uint   `json:"external_decimal"`
	ExternalTokenID string `json:"external_token_id"`
	NetworkID       uint   `json:"network_id"`
}

type AddTokenTestData struct {
	TestData
	ConfigedUnifiedTokens map[string]map[string]map[string]ConfigVault `json:"configed_unified_tokens"`
	BeaconHeight          uint64                                       `json:"beacon_height"`
	TriggeredFeature      map[string]uint64                            `json:"triggered_feature"`
	PrivacyTokens         map[common.Hash]struct {
		TokenID common.Hash `json:"token_id"`
	} `json:"privacy_tokens"`
}

type AddTokenTestCase struct {
	Data           AddTokenTestData `json:"data"`
	ExpectedResult ExpectedResult   `json:"expected_result"`
}

type AddTokenTestSuite struct {
	testCases map[string]*AddTokenTestCase
	TestSuite
}

func (a *AddTokenTestSuite) SetupSuite() {
	rawTestCases, _ := readTestCases("add_token.json")
	err := json.Unmarshal(rawTestCases, &a.testCases)
	if err != nil {
		panic(err)
	}
	a.actualResults = make(map[string]ActualResult)
}

func (a *AddTokenTestSuite) SetupTest() {
	dbPath, err := ioutil.TempDir(os.TempDir(), "bridgeagg_test_statedb_")
	if err != nil {
		panic(err)
	}
	diskBD, _ := incdb.Open("leveldb", dbPath)
	warperDBStatedbTest := statedb.NewDatabaseAccessWarper(diskBD)
	emptyRoot := common.HexToHash(common.HexEmptyRoot)
	sDB, _ := statedb.NewWithPrefixTrie(emptyRoot, warperDBStatedbTest)
	a.sDB = sDB
}

func (a *AddTokenTestSuite) BeforeTest(suiteName, testName string) {
	a.currentTestCaseName = testName
	testCase := a.testCases[testName]
	config.AbortUnifiedToken()
	configedUnifiedTokens := make(map[uint64]map[common.Hash]map[common.Hash]config.Vault)
	for checkpointStr, unifiedTokens := range testCase.Data.ConfigedUnifiedTokens {
		checkpoint, err := strconv.ParseUint(checkpointStr, 10, 64)
		if err != nil {
			panic(err)
		}
		configedUnifiedTokens[checkpoint] = make(map[common.Hash]map[common.Hash]config.Vault)
		for unifiedTokenID, vaults := range unifiedTokens {
			unifiedTokenHash, err := common.Hash{}.NewHashFromStr(unifiedTokenID)
			if err != nil {
				panic(err)
			}
			configedUnifiedTokens[checkpoint][*unifiedTokenHash] = make(map[common.Hash]config.Vault)
			for tokenIDStr, vault := range vaults {
				tokenID, err := common.Hash{}.NewHashFromStr(tokenIDStr)
				if err != nil {
					panic(err)
				}
				configedUnifiedTokens[checkpoint][*unifiedTokenHash][*tokenID] = config.Vault{
					ExternalDecimal: vault.ExternalDecimal,
					ExternalTokenID: vault.ExternalTokenID,
					NetworkID:       vault.NetworkID,
				}
			}
		}
	}
	config.SetUnifiedToken(configedUnifiedTokens)
	assert := a.Assert()

	for tokenID, v := range testCase.Data.BridgeTokensInfo {
		err := statedb.UpdateBridgeTokenInfo(a.sDB, tokenID, v.ExternalTokenID(), false, v.Amount(), "+")
		if err != nil {
			panic(err)
		}
	}
	for tokenID := range testCase.Data.PrivacyTokens {
		err := statedb.StorePrivacyToken(a.sDB, tokenID, "", "", 0, false, 0, []byte{}, common.Hash{})
		if err != nil {
			panic(err)
		}
	}
	_, err := a.sDB.Commit(false)
	assert.NoError(err, fmt.Sprintf("Error in commit db %v", err))
}

func (a *AddTokenTestSuite) test() {
	testCase := a.testCases[a.currentTestCaseName]
	assert := a.Assert()
	producerState := testCase.Data.State.Clone()
	producerManager := NewManagerWithValue(producerState)
	processorState := testCase.Data.State.Clone()
	processorManager := NewManagerWithValue(processorState)
	actualInstructions, accumulatedValues, err := producerManager.BuildAddTokenInstruction(
		testCase.Data.BeaconHeight,
		map[int]*statedb.StateDB{
			common.BeaconChainID: a.sDB,
		},
		testCase.Data.AccumulatedValues, testCase.Data.TriggeredFeature,
	)
	assert.NoError(err, fmt.Sprintf("Error in build instructions %v", err))
	err = processorManager.Process(actualInstructions, a.sDB)
	assert.NoError(err, fmt.Sprintf("Error in process instructions %v", err))
	a.actualResults[a.currentTestCaseName] = ActualResult{
		Instructions:      actualInstructions,
		ProducerState:     producerState,
		ProcessorState:    processorState,
		AccumulatedValues: accumulatedValues,
	}
}

func (a *AddTokenTestSuite) AfterTest(suiteName, testName string) {
	assert := a.Assert()
	_, err := a.sDB.Commit(false)
	assert.NoError(err, fmt.Sprintf("Error in commit db %v", err))
	bridgeTokenInfos := make(map[common.Hash]*rawdbv2.BridgeTokenInfo)
	tokens, err := statedb.GetBridgeTokens(a.sDB)
	assert.NoError(err, fmt.Sprintf("Error in get bridge tokens from db %v", err))
	for _, token := range tokens {
		bridgeTokenInfos[*token.TokenID] = token
	}
	expectedBridgeTokensInfo := a.testCases[a.currentTestCaseName].ExpectedResult.BridgeTokensInfo
	assert.Equal(expectedBridgeTokensInfo, bridgeTokenInfos, fmt.Errorf("Expected bridgeTokenInfos %v but get %v", expectedBridgeTokensInfo, bridgeTokenInfos).Error())
}

func (a *AddTokenTestSuite) TestAccepted() {
	a.test()
	assert := a.Assert()
	testCase := a.testCases[a.currentTestCaseName]
	actualResult := a.actualResults[a.currentTestCaseName]
	assert.Equal(testCase.ExpectedResult.Instructions, actualResult.Instructions, fmt.Errorf("Expected instructions %v but get %v", testCase.ExpectedResult.Instructions, actualResult.Instructions).Error())
	assert.Equal(testCase.ExpectedResult.State, actualResult.ProducerState, fmt.Errorf("Expected producer state %v but get %v", testCase.ExpectedResult.State, actualResult.ProducerState).Error())
	assert.Equal(testCase.ExpectedResult.State, actualResult.ProcessorState, fmt.Errorf("Expected processor state %v but get %v", testCase.ExpectedResult.State, actualResult.ProcessorState).Error())
	assert.Equal(testCase.ExpectedResult.AccumulatedValues, actualResult.AccumulatedValues, fmt.Errorf("Expected accumulatedValues %v but get %v", testCase.ExpectedResult.AccumulatedValues, actualResult.AccumulatedValues).Error())
}

/*func (a *AddTokenTestSuite) TestRejectedInvalidBeaconHeight() {*/
/*a.test()*/
/*assert := a.Assert()*/
/*testCase := a.testCases[a.currentTestCaseName]*/
/*actualResult := a.actualResults[a.currentTestCaseName]*/
/*assert.Equal(testCase.ExpectedResult.Instructions, actualResult.Instructions, fmt.Errorf("Expected instructions %v but get %v", testCase.ExpectedResult.Instructions, actualResult.Instructions).Error())*/
/*assert.Equal(testCase.ExpectedResult.State, actualResult.ProducerState, fmt.Errorf("Expected producer state %v but get %v", testCase.ExpectedResult.State, actualResult.ProducerState).Error())*/
/*assert.Equal(testCase.ExpectedResult.State, actualResult.ProcessorState, fmt.Errorf("Expected processor state %v but get %v", testCase.ExpectedResult.State, actualResult.ProcessorState).Error())*/
/*assert.Equal(testCase.ExpectedResult.AccumulatedValues, actualResult.AccumulatedValues, fmt.Errorf("Expected accumulatedValues %v but get %v", testCase.ExpectedResult.AccumulatedValues, actualResult.AccumulatedValues).Error())*/
/*}*/

/*func (a *AddTokenTestSuite) TestRejectedAvailableTokenIDByBridgeTokenIDCheck() {*/
/*a.test()*/
/*assert := a.Assert()*/
/*testCase := a.testCases[a.currentTestCaseName]*/
/*actualResult := a.actualResults[a.currentTestCaseName]*/
/*assert.Equal(testCase.ExpectedResult.Instructions, actualResult.Instructions, fmt.Errorf("Expected instructions %v but get %v", testCase.ExpectedResult.Instructions, actualResult.Instructions).Error())*/
/*assert.Equal(testCase.ExpectedResult.State, actualResult.ProducerState, fmt.Errorf("Expected producer state %v but get %v", testCase.ExpectedResult.State, actualResult.ProducerState).Error())*/
/*assert.Equal(testCase.ExpectedResult.State, actualResult.ProcessorState, fmt.Errorf("Expected processor state %v but get %v", testCase.ExpectedResult.State, actualResult.ProcessorState).Error())*/
/*assert.Equal(testCase.ExpectedResult.AccumulatedValues, actualResult.AccumulatedValues, fmt.Errorf("Expected accumulatedValues %v but get %v", testCase.ExpectedResult.AccumulatedValues, actualResult.AccumulatedValues).Error())*/
/*}*/

/*func (a *AddTokenTestSuite) TestRejectedAvailableTokenIDByPrivacyTokenIDCheck() {*/
/*a.test()*/
/*assert := a.Assert()*/
/*testCase := a.testCases[a.currentTestCaseName]*/
/*actualResult := a.actualResults[a.currentTestCaseName]*/
/*assert.Equal(testCase.ExpectedResult.Instructions, actualResult.Instructions, fmt.Errorf("Expected instructions %v but get %v", testCase.ExpectedResult.Instructions, actualResult.Instructions).Error())*/
/*assert.Equal(testCase.ExpectedResult.State, actualResult.ProducerState, fmt.Errorf("Expected producer state %v but get %v", testCase.ExpectedResult.State, actualResult.ProducerState).Error())*/
/*assert.Equal(testCase.ExpectedResult.State, actualResult.ProcessorState, fmt.Errorf("Expected processor state %v but get %v", testCase.ExpectedResult.State, actualResult.ProcessorState).Error())*/
/*assert.Equal(testCase.ExpectedResult.AccumulatedValues, actualResult.AccumulatedValues, fmt.Errorf("Expected accumulatedValues %v but get %v", testCase.ExpectedResult.AccumulatedValues, actualResult.AccumulatedValues).Error())*/
/*}*/

/*func (a *AddTokenTestSuite) TestRejectedDuplicateUnifiedTokenIDAndTokenID() {*/
/*a.test()*/
/*assert := a.Assert()*/
/*testCase := a.testCases[a.currentTestCaseName]*/
/*actualResult := a.actualResults[a.currentTestCaseName]*/
/*assert.Equal(testCase.ExpectedResult.Instructions, actualResult.Instructions, fmt.Errorf("Expected instructions %v but get %v", testCase.ExpectedResult.Instructions, actualResult.Instructions).Error())*/
/*assert.Equal(testCase.ExpectedResult.State, actualResult.ProducerState, fmt.Errorf("Expected producer state %v but get %v", testCase.ExpectedResult.State, actualResult.ProducerState).Error())*/
/*assert.Equal(testCase.ExpectedResult.State, actualResult.ProcessorState, fmt.Errorf("Expected processor state %v but get %v", testCase.ExpectedResult.State, actualResult.ProcessorState).Error())*/
/*assert.Equal(testCase.ExpectedResult.AccumulatedValues, actualResult.AccumulatedValues, fmt.Errorf("Expected accumulatedValues %v but get %v", testCase.ExpectedResult.AccumulatedValues, actualResult.AccumulatedValues).Error())*/
/*}*/

/*func (a *AddTokenTestSuite) TestRejectedDecimalBy0() {*/
/*a.test()*/
/*assert := a.Assert()*/
/*testCase := a.testCases[a.currentTestCaseName]*/
/*actualResult := a.actualResults[a.currentTestCaseName]*/
/*assert.Equal(testCase.ExpectedResult.Instructions, actualResult.Instructions, fmt.Errorf("Expected instructions %v but get %v", testCase.ExpectedResult.Instructions, actualResult.Instructions).Error())*/
/*assert.Equal(testCase.ExpectedResult.State, actualResult.ProducerState, fmt.Errorf("Expected producer state %v but get %v", testCase.ExpectedResult.State, actualResult.ProducerState).Error())*/
/*assert.Equal(testCase.ExpectedResult.State, actualResult.ProcessorState, fmt.Errorf("Expected processor state %v but get %v", testCase.ExpectedResult.State, actualResult.ProcessorState).Error())*/
/*assert.Equal(testCase.ExpectedResult.AccumulatedValues, actualResult.AccumulatedValues, fmt.Errorf("Expected accumulatedValues %v but get %v", testCase.ExpectedResult.AccumulatedValues, actualResult.AccumulatedValues).Error())*/
/*}*/

/*func (a *AddTokenTestSuite) TestRejectedEmptyTokenIDStr() {*/
/*a.test()*/
/*assert := a.Assert()*/
/*testCase := a.testCases[a.currentTestCaseName]*/
/*actualResult := a.actualResults[a.currentTestCaseName]*/
/*assert.Equal(testCase.ExpectedResult.Instructions, actualResult.Instructions, fmt.Errorf("Expected instructions %v but get %v", testCase.ExpectedResult.Instructions, actualResult.Instructions).Error())*/
/*assert.Equal(testCase.ExpectedResult.State, actualResult.ProducerState, fmt.Errorf("Expected producer state %v but get %v", testCase.ExpectedResult.State, actualResult.ProducerState).Error())*/
/*assert.Equal(testCase.ExpectedResult.State, actualResult.ProcessorState, fmt.Errorf("Expected processor state %v but get %v", testCase.ExpectedResult.State, actualResult.ProcessorState).Error())*/
/*assert.Equal(testCase.ExpectedResult.AccumulatedValues, actualResult.AccumulatedValues, fmt.Errorf("Expected accumulatedValues %v but get %v", testCase.ExpectedResult.AccumulatedValues, actualResult.AccumulatedValues).Error())*/
/*}*/

/*func (a *AddTokenTestSuite) TestRejectedNullUnifiedTokenID() {*/
/*a.test()*/
/*assert := a.Assert()*/
/*testCase := a.testCases[a.currentTestCaseName]*/
/*actualResult := a.actualResults[a.currentTestCaseName]*/
/*assert.Equal(testCase.ExpectedResult.Instructions, actualResult.Instructions, fmt.Errorf("Expected instructions %v but get %v", testCase.ExpectedResult.Instructions, actualResult.Instructions).Error())*/
/*assert.Equal(testCase.ExpectedResult.State, actualResult.ProducerState, fmt.Errorf("Expected producer state %v but get %v", testCase.ExpectedResult.State, actualResult.ProducerState).Error())*/
/*assert.Equal(testCase.ExpectedResult.State, actualResult.ProcessorState, fmt.Errorf("Expected processor state %v but get %v", testCase.ExpectedResult.State, actualResult.ProcessorState).Error())*/
/*assert.Equal(testCase.ExpectedResult.AccumulatedValues, actualResult.AccumulatedValues, fmt.Errorf("Expected accumulatedValues %v but get %v", testCase.ExpectedResult.AccumulatedValues, actualResult.AccumulatedValues).Error())*/
/*}*/

/*func (a *AddTokenTestSuite) TestRejectedEmptyUnifiedTokenID() {*/
/*a.test()*/
/*assert := a.Assert()*/
/*testCase := a.testCases[a.currentTestCaseName]*/
/*actualResult := a.actualResults[a.currentTestCaseName]*/
/*assert.Equal(testCase.ExpectedResult.Instructions, actualResult.Instructions, fmt.Errorf("Expected instructions %v but get %v", testCase.ExpectedResult.Instructions, actualResult.Instructions).Error())*/
/*assert.Equal(testCase.ExpectedResult.State, actualResult.ProducerState, fmt.Errorf("Expected producer state %v but get %v", testCase.ExpectedResult.State, actualResult.ProducerState).Error())*/
/*assert.Equal(testCase.ExpectedResult.State, actualResult.ProcessorState, fmt.Errorf("Expected processor state %v but get %v", testCase.ExpectedResult.State, actualResult.ProcessorState).Error())*/
/*assert.Equal(testCase.ExpectedResult.AccumulatedValues, actualResult.AccumulatedValues, fmt.Errorf("Expected accumulatedValues %v but get %v", testCase.ExpectedResult.AccumulatedValues, actualResult.AccumulatedValues).Error())*/
/*}*/

/*func (a *AddTokenTestSuite) TestRejectedEmptyTokenID() {*/
/*a.test()*/
/*assert := a.Assert()*/
/*testCase := a.testCases[a.currentTestCaseName]*/
/*actualResult := a.actualResults[a.currentTestCaseName]*/
/*assert.Equal(testCase.ExpectedResult.Instructions, actualResult.Instructions, fmt.Errorf("Expected instructions %v but get %v", testCase.ExpectedResult.Instructions, actualResult.Instructions).Error())*/
/*assert.Equal(testCase.ExpectedResult.State, actualResult.ProducerState, fmt.Errorf("Expected producer state %v but get %v", testCase.ExpectedResult.State, actualResult.ProducerState).Error())*/
/*assert.Equal(testCase.ExpectedResult.State, actualResult.ProcessorState, fmt.Errorf("Expected processor state %v but get %v", testCase.ExpectedResult.State, actualResult.ProcessorState).Error())*/
/*assert.Equal(testCase.ExpectedResult.AccumulatedValues, actualResult.AccumulatedValues, fmt.Errorf("Expected accumulatedValues %v but get %v", testCase.ExpectedResult.AccumulatedValues, actualResult.AccumulatedValues).Error())*/
/*}*/

/*func (a *AddTokenTestSuite) TestRejectedNotFoundNetworkID() {*/
/*a.test()*/
/*assert := a.Assert()*/
/*testCase := a.testCases[a.currentTestCaseName]*/
/*actualResult := a.actualResults[a.currentTestCaseName]*/
/*assert.Equal(testCase.ExpectedResult.Instructions, actualResult.Instructions, fmt.Errorf("Expected instructions %v but get %v", testCase.ExpectedResult.Instructions, actualResult.Instructions).Error())*/
/*assert.Equal(testCase.ExpectedResult.State, actualResult.ProducerState, fmt.Errorf("Expected producer state %v but get %v", testCase.ExpectedResult.State, actualResult.ProducerState).Error())*/
/*assert.Equal(testCase.ExpectedResult.State, actualResult.ProcessorState, fmt.Errorf("Expected processor state %v but get %v", testCase.ExpectedResult.State, actualResult.ProcessorState).Error())*/
/*assert.Equal(testCase.ExpectedResult.AccumulatedValues, actualResult.AccumulatedValues, fmt.Errorf("Expected accumulatedValues %v but get %v", testCase.ExpectedResult.AccumulatedValues, actualResult.AccumulatedValues).Error())*/
/*}*/

/*func (a *AddTokenTestSuite) TestRejectedUnifiedTokenIDAvailableInPrivacyTokenList() {*/
/*a.test()*/
/*assert := a.Assert()*/
/*testCase := a.testCases[a.currentTestCaseName]*/
/*actualResult := a.actualResults[a.currentTestCaseName]*/
/*assert.Equal(testCase.ExpectedResult.Instructions, actualResult.Instructions, fmt.Errorf("Expected instructions %v but get %v", testCase.ExpectedResult.Instructions, actualResult.Instructions).Error())*/
/*assert.Equal(testCase.ExpectedResult.State, actualResult.ProducerState, fmt.Errorf("Expected producer state %v but get %v", testCase.ExpectedResult.State, actualResult.ProducerState).Error())*/
/*assert.Equal(testCase.ExpectedResult.State, actualResult.ProcessorState, fmt.Errorf("Expected processor state %v but get %v", testCase.ExpectedResult.State, actualResult.ProcessorState).Error())*/
/*assert.Equal(testCase.ExpectedResult.AccumulatedValues, actualResult.AccumulatedValues, fmt.Errorf("Expected accumulatedValues %v but get %v", testCase.ExpectedResult.AccumulatedValues, actualResult.AccumulatedValues).Error())*/
/*}*/

/*func (a *AddTokenTestSuite) TestRejectedEmptyExternalTokenID() {*/
/*a.test()*/
/*assert := a.Assert()*/
/*testCase := a.testCases[a.currentTestCaseName]*/
/*actualResult := a.actualResults[a.currentTestCaseName]*/
/*assert.Equal(testCase.ExpectedResult.Instructions, actualResult.Instructions, fmt.Errorf("Expected instructions %v but get %v", testCase.ExpectedResult.Instructions, actualResult.Instructions).Error())*/
/*assert.Equal(testCase.ExpectedResult.State, actualResult.ProducerState, fmt.Errorf("Expected producer state %v but get %v", testCase.ExpectedResult.State, actualResult.ProducerState).Error())*/
/*assert.Equal(testCase.ExpectedResult.State, actualResult.ProcessorState, fmt.Errorf("Expected processor state %v but get %v", testCase.ExpectedResult.State, actualResult.ProcessorState).Error())*/
/*assert.Equal(testCase.ExpectedResult.AccumulatedValues, actualResult.AccumulatedValues, fmt.Errorf("Expected accumulatedValues %v but get %v", testCase.ExpectedResult.AccumulatedValues, actualResult.AccumulatedValues).Error())*/
/*}*/

/*func (a *AddTokenTestSuite) TestRejectedDuplicateExternalTokenID() {*/
/*a.test()*/
/*assert := a.Assert()*/
/*testCase := a.testCases[a.currentTestCaseName]*/
/*actualResult := a.actualResults[a.currentTestCaseName]*/
/*assert.Equal(testCase.ExpectedResult.Instructions, actualResult.Instructions, fmt.Errorf("Expected instructions %v but get %v", testCase.ExpectedResult.Instructions, actualResult.Instructions).Error())*/
/*assert.Equal(testCase.ExpectedResult.State, actualResult.ProducerState, fmt.Errorf("Expected producer state %v but get %v", testCase.ExpectedResult.State, actualResult.ProducerState).Error())*/
/*assert.Equal(testCase.ExpectedResult.State, actualResult.ProcessorState, fmt.Errorf("Expected processor state %v but get %v", testCase.ExpectedResult.State, actualResult.ProcessorState).Error())*/
/*assert.Equal(testCase.ExpectedResult.AccumulatedValues, actualResult.AccumulatedValues, fmt.Errorf("Expected accumulatedValues %v but get %v", testCase.ExpectedResult.AccumulatedValues, actualResult.AccumulatedValues).Error())*/
/*}*/

func TestAddTokenTestSuite(t *testing.T) {
	suite.Run(t, new(AddTokenTestSuite))
}
