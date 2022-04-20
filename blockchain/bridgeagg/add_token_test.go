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
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/stretchr/testify/suite"
)

type AddTokenTestCase struct {
	UnifiedTokens map[uint64]map[common.Hash]map[uint]config.Vault `json:"unified_tokens"`
	BeaconHeight  uint64                                           `json:"beacon_height"`
	TestCase
}

func (a *AddTokenTestSuite) BeforeTest(suiteName, testName string) {
	testCase := a.testCases[testName]
	config.AbortUnifiedToken()
	config.SetUnifiedToken(testCase.UnifiedTokens)

	assert := a.Assert()

	accumulatedValues := &metadata.AccumulatedValues{
		UniqETHTxsUsed:    [][]byte{},
		UniqBSCTxsUsed:    [][]byte{},
		UniqPLGTxsUsed:    [][]byte{},
		UniqPRVEVMTxsUsed: [][]byte{},
		UniqFTMTxsUsed:    [][]byte{},
		DBridgeTokenPair:  map[string][]byte{},
		CBridgeTokens:     []*common.Hash{},
		InitTokens:        []*common.Hash{},
	}

	state := NewState()
	producerState := state.Clone()
	processorState := state.Clone()
	actualInstruction, accumulatedValues, err := producerState.BuildAddTokenInstruction(
		testCase.BeaconHeight,
		map[int]*statedb.StateDB{
			common.BeaconChainID: a.sDB,
		},
		accumulatedValues,
	)
	assert.NoError(err, fmt.Sprintf("Error in build instructions %v", err))
	err = processorState.Process([][]string{actualInstruction}, a.sDB)
	assert.NoError(err, fmt.Sprintf("Error in process instructions %v", err))
	a.actualResults[testName] = ActualResult{
		Instructions:      [][]string{actualInstruction},
		ProducerState:     producerState,
		ProcessorState:    processorState,
		AccumulatedValues: accumulatedValues,
	}
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

func (a *AddTokenTestSuite) TestAccepted() {
	assert := a.Assert()
	testCase := a.testCases[a.currentTestCaseName]
	actualResult := a.actualResults[a.currentTestCaseName]
	expectedState := NewState()
	expectedState.unifiedTokenInfos = testCase.ExpectedUnifiedTokens
	assert.Equal(testCase.ExpectedInstructions, actualResult.Instructions, fmt.Errorf("Expected instructions %v but get %v", actualResult.Instructions, testCase.ExpectedInstructions).Error())
	assert.Equal(expectedState, actualResult.ProducerState, fmt.Errorf("Expected producer state %v but get %v", expectedState, actualResult.ProducerState).Error())
	assert.Equal(expectedState, actualResult.ProcessorState, fmt.Errorf("Expected processor state %v but get %v", expectedState, actualResult.ProcessorState).Error())
}

func TestAddTokenTestSuite(t *testing.T) {
	suite.Run(t, new(AddTokenTestSuite))
}
