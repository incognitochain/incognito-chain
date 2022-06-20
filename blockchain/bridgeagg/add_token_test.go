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
	IncTokenID      string `json:"inc_token_id"`
}

type AddTokenTestCase struct {
	ConfigedUnifiedTokens    map[string]map[string]map[string]ConfigVault `json:"configed_unified_tokens"`
	BeaconHeight             uint64                                       `json:"beacon_height"`
	ExpectedBridgeTokensInfo map[common.Hash]*rawdbv2.BridgeTokenInfo     `json:"expected_bridge_tokens_info"`
	PrivacyTokens            map[common.Hash]struct {
		TokenID common.Hash `json:"token_id"`
	} `json:"privacy_tokens"`

	TestCase
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
	configedUnifiedTokens := make(map[uint64]map[common.Hash]map[uint]config.Vault)
	for beaconHeightStr, unifiedTokens := range testCase.ConfigedUnifiedTokens {
		beaconHeight, err := strconv.ParseUint(beaconHeightStr, 10, 64)
		if err != nil {
			panic(err)
		}
		configedUnifiedTokens[beaconHeight] = make(map[common.Hash]map[uint]config.Vault)
		for unifiedTokenID, vaults := range unifiedTokens {
			unifiedTokenHash, err := common.Hash{}.NewHashFromStr(unifiedTokenID)
			if err != nil {
				panic(err)
			}
			configedUnifiedTokens[beaconHeight][*unifiedTokenHash] = make(map[uint]config.Vault)
			for networkIDStr, vault := range vaults {
				networkID, err := strconv.Atoi(networkIDStr)
				if err != nil {
					panic(err)
				}
				configedUnifiedTokens[beaconHeight][*unifiedTokenHash][uint(networkID)] = config.Vault{
					ExternalDecimal: vault.ExternalDecimal,
					ExternalTokenID: vault.ExternalTokenID,
					IncTokenID:      vault.IncTokenID,
				}
			}
		}
	}
	config.SetUnifiedToken(configedUnifiedTokens)
	assert := a.Assert()

	for tokenID, v := range testCase.BridgeTokensInfo {
		err := statedb.UpdateBridgeTokenInfo(a.sDB, tokenID, v.ExternalTokenID(), false, v.Amount(), "+")
		if err != nil {
			panic(err)
		}
	}
	for tokenID := range testCase.PrivacyTokens {
		err := statedb.StorePrivacyToken(a.sDB, tokenID, "", "", 0, false, 0, []byte{}, common.Hash{})
		if err != nil {
			panic(err)
		}
	}
	_, err := a.sDB.Commit(false)
	assert.NoError(err, fmt.Sprintf("Error in commit db %v", err))

	state := NewState()
	producerState := state.Clone()
	processorState := state.Clone()
	actualInstructions, accumulatedValues, err := producerState.BuildAddTokenInstruction(
		testCase.BeaconHeight,
		map[int]*statedb.StateDB{
			common.BeaconChainID: a.sDB,
		},
		testCase.AccumulatedValues,
	)
	assert.NoError(err, fmt.Sprintf("Error in build instructions %v", err))
	err = processorState.Process(actualInstructions, a.sDB)
	assert.NoError(err, fmt.Sprintf("Error in process instructions %v", err))
	a.actualResults[testName] = ActualResult{
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
	expectedBridgeTokensInfo := a.testCases[a.currentTestCaseName].ExpectedBridgeTokensInfo
	assert.Equal(expectedBridgeTokensInfo, bridgeTokenInfos, fmt.Errorf("Expected bridgeTokenInfos %v but get %v", expectedBridgeTokensInfo, bridgeTokenInfos).Error())
}

func (a *AddTokenTestSuite) TestAccepted() {
	assert := a.Assert()
	testCase := a.testCases[a.currentTestCaseName]
	actualResult := a.actualResults[a.currentTestCaseName]
	expectedState := NewState()
	expectedState.unifiedTokenInfos = testCase.ExpectedUnifiedTokens
	assert.Equal(testCase.ExpectedInstructions, actualResult.Instructions, fmt.Errorf("Expected instructions %v but get %v", testCase.ExpectedInstructions, actualResult.Instructions).Error())
	for k, v := range actualResult.ProducerState.unifiedTokenInfos {
		assert.Equal(expectedState.unifiedTokenInfos[k], v, fmt.Errorf("Expected producer state %v but get %v", expectedState, actualResult.ProducerState).Error())
	}
	for k, v := range actualResult.ProcessorState.unifiedTokenInfos {
		assert.Equal(expectedState.unifiedTokenInfos[k], v, fmt.Errorf("Expected processor state %v but get %v", expectedState, actualResult.ProcessorState).Error())
	}
	assert.Equal(testCase.ExpectedAccumulatedValues, actualResult.AccumulatedValues, fmt.Errorf("Expected accumulatedValues %v but get %v", testCase.AccumulatedValues, actualResult.AccumulatedValues).Error())
}

func (a *AddTokenTestSuite) TestRejectedInvalidBeaconHeight() {
	assert := a.Assert()
	testCase := a.testCases[a.currentTestCaseName]
	actualResult := a.actualResults[a.currentTestCaseName]
	expectedState := NewState()
	expectedState.unifiedTokenInfos = testCase.ExpectedUnifiedTokens
	assert.Equal(testCase.ExpectedInstructions, actualResult.Instructions, fmt.Errorf("Expected instructions %v but get %v", testCase.ExpectedInstructions, actualResult.Instructions).Error())
	for k, v := range actualResult.ProducerState.unifiedTokenInfos {
		assert.Equal(expectedState.unifiedTokenInfos[k], v, fmt.Errorf("Expected producer state %v but get %v", expectedState, actualResult.ProducerState).Error())
	}
	for k, v := range actualResult.ProcessorState.unifiedTokenInfos {
		assert.Equal(expectedState.unifiedTokenInfos[k], v, fmt.Errorf("Expected processor state %v but get %v", expectedState, actualResult.ProcessorState).Error())
	}
	assert.Equal(testCase.ExpectedAccumulatedValues, actualResult.AccumulatedValues, fmt.Errorf("Expected accumulatedValues %v but get %v", testCase.AccumulatedValues, actualResult.AccumulatedValues).Error())
}

func (a *AddTokenTestSuite) TestRejectedAvailableTokenIDByBridgeTokenIDCheck() {
	assert := a.Assert()
	testCase := a.testCases[a.currentTestCaseName]
	actualResult := a.actualResults[a.currentTestCaseName]
	expectedState := NewState()
	expectedState.unifiedTokenInfos = testCase.ExpectedUnifiedTokens
	assert.Equal(testCase.ExpectedInstructions, actualResult.Instructions, fmt.Errorf("Expected instructions %v but get %v", testCase.ExpectedInstructions, actualResult.Instructions).Error())
	for k, v := range actualResult.ProducerState.unifiedTokenInfos {
		assert.Equal(expectedState.unifiedTokenInfos[k], v, fmt.Errorf("Expected producer state %v but get %v", expectedState, actualResult.ProducerState).Error())
	}
	for k, v := range actualResult.ProcessorState.unifiedTokenInfos {
		assert.Equal(expectedState.unifiedTokenInfos[k], v, fmt.Errorf("Expected processor state %v but get %v", expectedState, actualResult.ProcessorState).Error())
	}
	assert.Equal(testCase.ExpectedAccumulatedValues, actualResult.AccumulatedValues, fmt.Errorf("Expected accumulatedValues %v but get %v", testCase.AccumulatedValues, actualResult.AccumulatedValues).Error())
}

func (a *AddTokenTestSuite) TestRejectedAvailableTokenIDByPrivacyTokenIDCheck() {
	assert := a.Assert()
	testCase := a.testCases[a.currentTestCaseName]
	actualResult := a.actualResults[a.currentTestCaseName]
	expectedState := NewState()
	expectedState.unifiedTokenInfos = testCase.ExpectedUnifiedTokens
	assert.Equal(testCase.ExpectedInstructions, actualResult.Instructions, fmt.Errorf("Expected instructions %v but get %v", testCase.ExpectedInstructions, actualResult.Instructions).Error())
	for k, v := range actualResult.ProducerState.unifiedTokenInfos {
		assert.Equal(expectedState.unifiedTokenInfos[k], v, fmt.Errorf("Expected producer state %v but get %v", expectedState, actualResult.ProducerState).Error())
	}
	for k, v := range actualResult.ProcessorState.unifiedTokenInfos {
		assert.Equal(expectedState.unifiedTokenInfos[k], v, fmt.Errorf("Expected processor state %v but get %v", expectedState, actualResult.ProcessorState).Error())
	}
	assert.Equal(testCase.ExpectedAccumulatedValues, actualResult.AccumulatedValues, fmt.Errorf("Expected accumulatedValues %v but get %v", testCase.AccumulatedValues, actualResult.AccumulatedValues).Error())
}

func (a *AddTokenTestSuite) TestRejectedDuplicateTokenIDs() {
	assert := a.Assert()
	testCase := a.testCases[a.currentTestCaseName]
	actualResult := a.actualResults[a.currentTestCaseName]
	expectedState := NewState()
	expectedState.unifiedTokenInfos = testCase.ExpectedUnifiedTokens
	assert.Equal(testCase.ExpectedInstructions, actualResult.Instructions, fmt.Errorf("Expected instructions %v but get %v", testCase.ExpectedInstructions, actualResult.Instructions).Error())
	for k, v := range actualResult.ProducerState.unifiedTokenInfos {
		assert.Equal(expectedState.unifiedTokenInfos[k], v, fmt.Errorf("Expected producer state %v but get %v", expectedState, actualResult.ProducerState).Error())
	}
	for k, v := range actualResult.ProcessorState.unifiedTokenInfos {
		assert.Equal(expectedState.unifiedTokenInfos[k], v, fmt.Errorf("Expected processor state %v but get %v", expectedState, actualResult.ProcessorState).Error())
	}
	assert.Equal(testCase.ExpectedAccumulatedValues, actualResult.AccumulatedValues, fmt.Errorf("Expected accumulatedValues %v but get %v", testCase.AccumulatedValues, actualResult.AccumulatedValues).Error())
}

func (a *AddTokenTestSuite) TestRejectedDuplicateUnifiedTokenIDAndTokenID() {
	assert := a.Assert()
	testCase := a.testCases[a.currentTestCaseName]
	actualResult := a.actualResults[a.currentTestCaseName]
	expectedState := NewState()
	expectedState.unifiedTokenInfos = testCase.ExpectedUnifiedTokens
	assert.Equal(testCase.ExpectedInstructions, actualResult.Instructions, fmt.Errorf("Expected instructions %v but get %v", testCase.ExpectedInstructions, actualResult.Instructions).Error())
	for k, v := range actualResult.ProducerState.unifiedTokenInfos {
		assert.Equal(expectedState.unifiedTokenInfos[k], v, fmt.Errorf("Expected producer state %v but get %v", expectedState, actualResult.ProducerState).Error())
	}
	for k, v := range actualResult.ProcessorState.unifiedTokenInfos {
		assert.Equal(expectedState.unifiedTokenInfos[k], v, fmt.Errorf("Expected processor state %v but get %v", expectedState, actualResult.ProcessorState).Error())
	}
	assert.Equal(testCase.ExpectedAccumulatedValues, actualResult.AccumulatedValues, fmt.Errorf("Expected accumulatedValues %v but get %v", testCase.AccumulatedValues, actualResult.AccumulatedValues).Error())
}

func (a *AddTokenTestSuite) TestRejectedDecimalBy0() {
	assert := a.Assert()
	testCase := a.testCases[a.currentTestCaseName]
	actualResult := a.actualResults[a.currentTestCaseName]
	expectedState := NewState()
	expectedState.unifiedTokenInfos = testCase.ExpectedUnifiedTokens
	assert.Equal(testCase.ExpectedInstructions, actualResult.Instructions, fmt.Errorf("Expected instructions %v but get %v", testCase.ExpectedInstructions, actualResult.Instructions).Error())
	for k, v := range actualResult.ProducerState.unifiedTokenInfos {
		assert.Equal(expectedState.unifiedTokenInfos[k], v, fmt.Errorf("Expected producer state %v but get %v", expectedState, actualResult.ProducerState).Error())
	}
	for k, v := range actualResult.ProcessorState.unifiedTokenInfos {
		assert.Equal(expectedState.unifiedTokenInfos[k], v, fmt.Errorf("Expected processor state %v but get %v", expectedState, actualResult.ProcessorState).Error())
	}
	assert.Equal(testCase.ExpectedAccumulatedValues, actualResult.AccumulatedValues, fmt.Errorf("Expected accumulatedValues %v but get %v", testCase.AccumulatedValues, actualResult.AccumulatedValues).Error())
}

func (a *AddTokenTestSuite) TestRejectedEmptyTokenIDStr() {
	assert := a.Assert()
	testCase := a.testCases[a.currentTestCaseName]
	actualResult := a.actualResults[a.currentTestCaseName]
	expectedState := NewState()
	expectedState.unifiedTokenInfos = testCase.ExpectedUnifiedTokens
	assert.Equal(testCase.ExpectedInstructions, actualResult.Instructions, fmt.Errorf("Expected instructions %v but get %v", testCase.ExpectedInstructions, actualResult.Instructions).Error())
	for k, v := range actualResult.ProducerState.unifiedTokenInfos {
		assert.Equal(expectedState.unifiedTokenInfos[k], v, fmt.Errorf("Expected producer state %v but get %v", expectedState, actualResult.ProducerState).Error())
	}
	for k, v := range actualResult.ProcessorState.unifiedTokenInfos {
		assert.Equal(expectedState.unifiedTokenInfos[k], v, fmt.Errorf("Expected processor state %v but get %v", expectedState, actualResult.ProcessorState).Error())
	}
	assert.Equal(testCase.ExpectedAccumulatedValues, actualResult.AccumulatedValues, fmt.Errorf("Expected accumulatedValues %v but get %v", testCase.AccumulatedValues, actualResult.AccumulatedValues).Error())
}

func (a *AddTokenTestSuite) TestRejectedNullUnifiedTokenID() {
	assert := a.Assert()
	testCase := a.testCases[a.currentTestCaseName]
	actualResult := a.actualResults[a.currentTestCaseName]
	expectedState := NewState()
	expectedState.unifiedTokenInfos = testCase.ExpectedUnifiedTokens
	assert.Equal(testCase.ExpectedInstructions, actualResult.Instructions, fmt.Errorf("Expected instructions %v but get %v", testCase.ExpectedInstructions, actualResult.Instructions).Error())
	for k, v := range actualResult.ProducerState.unifiedTokenInfos {
		assert.Equal(expectedState.unifiedTokenInfos[k], v, fmt.Errorf("Expected producer state %v but get %v", expectedState, actualResult.ProducerState).Error())
	}
	for k, v := range actualResult.ProcessorState.unifiedTokenInfos {
		assert.Equal(expectedState.unifiedTokenInfos[k], v, fmt.Errorf("Expected processor state %v but get %v", expectedState, actualResult.ProcessorState).Error())
	}
	assert.Equal(testCase.ExpectedAccumulatedValues, actualResult.AccumulatedValues, fmt.Errorf("Expected accumulatedValues %v but get %v", testCase.AccumulatedValues, actualResult.AccumulatedValues).Error())
}

func (a *AddTokenTestSuite) TestRejectedEmptyUnifiedTokenID() {
	assert := a.Assert()
	testCase := a.testCases[a.currentTestCaseName]
	actualResult := a.actualResults[a.currentTestCaseName]
	expectedState := NewState()
	expectedState.unifiedTokenInfos = testCase.ExpectedUnifiedTokens
	assert.Equal(testCase.ExpectedInstructions, actualResult.Instructions, fmt.Errorf("Expected instructions %v but get %v", testCase.ExpectedInstructions, actualResult.Instructions).Error())
	for k, v := range actualResult.ProducerState.unifiedTokenInfos {
		assert.Equal(expectedState.unifiedTokenInfos[k], v, fmt.Errorf("Expected producer state %v but get %v", expectedState, actualResult.ProducerState).Error())
	}
	for k, v := range actualResult.ProcessorState.unifiedTokenInfos {
		assert.Equal(expectedState.unifiedTokenInfos[k], v, fmt.Errorf("Expected processor state %v but get %v", expectedState, actualResult.ProcessorState).Error())
	}
	assert.Equal(testCase.ExpectedAccumulatedValues, actualResult.AccumulatedValues, fmt.Errorf("Expected accumulatedValues %v but get %v", testCase.AccumulatedValues, actualResult.AccumulatedValues).Error())
}

func (a *AddTokenTestSuite) TestRejectedEmptyTokenID() {
	assert := a.Assert()
	testCase := a.testCases[a.currentTestCaseName]
	actualResult := a.actualResults[a.currentTestCaseName]
	expectedState := NewState()
	expectedState.unifiedTokenInfos = testCase.ExpectedUnifiedTokens
	assert.Equal(testCase.ExpectedInstructions, actualResult.Instructions, fmt.Errorf("Expected instructions %v but get %v", testCase.ExpectedInstructions, actualResult.Instructions).Error())
	for k, v := range actualResult.ProducerState.unifiedTokenInfos {
		assert.Equal(expectedState.unifiedTokenInfos[k], v, fmt.Errorf("Expected producer state %v but get %v", expectedState, actualResult.ProducerState).Error())
	}
	for k, v := range actualResult.ProcessorState.unifiedTokenInfos {
		assert.Equal(expectedState.unifiedTokenInfos[k], v, fmt.Errorf("Expected processor state %v but get %v", expectedState, actualResult.ProcessorState).Error())
	}
	assert.Equal(testCase.ExpectedAccumulatedValues, actualResult.AccumulatedValues, fmt.Errorf("Expected accumulatedValues %v but get %v", testCase.AccumulatedValues, actualResult.AccumulatedValues).Error())
}

func (a *AddTokenTestSuite) TestRejectedNotFoundNetworkID() {
	assert := a.Assert()
	testCase := a.testCases[a.currentTestCaseName]
	actualResult := a.actualResults[a.currentTestCaseName]
	expectedState := NewState()
	expectedState.unifiedTokenInfos = testCase.ExpectedUnifiedTokens
	assert.Equal(testCase.ExpectedInstructions, actualResult.Instructions, fmt.Errorf("Expected instructions %v but get %v", testCase.ExpectedInstructions, actualResult.Instructions).Error())
	for k, v := range actualResult.ProducerState.unifiedTokenInfos {
		assert.Equal(expectedState.unifiedTokenInfos[k], v, fmt.Errorf("Expected producer state %v but get %v", expectedState, actualResult.ProducerState).Error())
	}
	for k, v := range actualResult.ProcessorState.unifiedTokenInfos {
		assert.Equal(expectedState.unifiedTokenInfos[k], v, fmt.Errorf("Expected processor state %v but get %v", expectedState, actualResult.ProcessorState).Error())
	}
	assert.Equal(testCase.ExpectedAccumulatedValues, actualResult.AccumulatedValues, fmt.Errorf("Expected accumulatedValues %v but get %v", testCase.AccumulatedValues, actualResult.AccumulatedValues).Error())
}

func (a *AddTokenTestSuite) TestRejectedUnifiedTokenIDAvailableInPrivacyTokenList() {
	assert := a.Assert()
	testCase := a.testCases[a.currentTestCaseName]
	actualResult := a.actualResults[a.currentTestCaseName]
	expectedState := NewState()
	expectedState.unifiedTokenInfos = testCase.ExpectedUnifiedTokens
	assert.Equal(testCase.ExpectedInstructions, actualResult.Instructions, fmt.Errorf("Expected instructions %v but get %v", testCase.ExpectedInstructions, actualResult.Instructions).Error())
	for k, v := range actualResult.ProducerState.unifiedTokenInfos {
		assert.Equal(expectedState.unifiedTokenInfos[k], v, fmt.Errorf("Expected producer state %v but get %v", expectedState, actualResult.ProducerState).Error())
	}
	for k, v := range actualResult.ProcessorState.unifiedTokenInfos {
		assert.Equal(expectedState.unifiedTokenInfos[k], v, fmt.Errorf("Expected processor state %v but get %v", expectedState, actualResult.ProcessorState).Error())
	}
	assert.Equal(testCase.ExpectedAccumulatedValues, actualResult.AccumulatedValues, fmt.Errorf("Expected accumulatedValues %v but get %v", testCase.AccumulatedValues, actualResult.AccumulatedValues).Error())
}

func (a *AddTokenTestSuite) TestRejectedEmptyExternalTokenID() {
	assert := a.Assert()
	testCase := a.testCases[a.currentTestCaseName]
	actualResult := a.actualResults[a.currentTestCaseName]
	expectedState := NewState()
	expectedState.unifiedTokenInfos = testCase.ExpectedUnifiedTokens
	assert.Equal(testCase.ExpectedInstructions, actualResult.Instructions, fmt.Errorf("Expected instructions %v but get %v", testCase.ExpectedInstructions, actualResult.Instructions).Error())
	for k, v := range actualResult.ProducerState.unifiedTokenInfos {
		assert.Equal(expectedState.unifiedTokenInfos[k], v, fmt.Errorf("Expected producer state %v but get %v", expectedState, actualResult.ProducerState).Error())
	}
	for k, v := range actualResult.ProcessorState.unifiedTokenInfos {
		assert.Equal(expectedState.unifiedTokenInfos[k], v, fmt.Errorf("Expected processor state %v but get %v", expectedState, actualResult.ProcessorState).Error())
	}
	assert.Equal(testCase.ExpectedAccumulatedValues, actualResult.AccumulatedValues, fmt.Errorf("Expected accumulatedValues %v but get %v", testCase.AccumulatedValues, actualResult.AccumulatedValues).Error())
}

func TestAddTokenTestSuite(t *testing.T) {
	suite.Run(t, new(AddTokenTestSuite))
}
