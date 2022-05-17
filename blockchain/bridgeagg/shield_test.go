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
	metadataBridge "github.com/incognitochain/incognito-chain/metadata/bridge"
	metadataMocks "github.com/incognitochain/incognito-chain/metadata/common/mocks"
	"github.com/incognitochain/incognito-chain/metadata/evmcaller"
	"github.com/stretchr/testify/suite"
)

type ShieldTestCase struct {
	TestCase
	Metadatas                 []*metadataBridge.ShieldRequest `json:"metadatas"`
	ExpectedStatuses          []ShieldStatus                  `json:"expected_statuses"`
	ActualStatues             []ShieldStatus
	ExpectedAccumulatedValues *metadata.AccumulatedValues `json:"expected_accumulated_values"`
	ActualAccumulatedValues   *metadata.AccumulatedValues
}

type ShieldTestSuite struct {
	testCases map[string]*ShieldTestCase
	TestSuite
}

func (s *ShieldTestSuite) SetupSuite() {
	config.AbortParam()
	config.Param().BridgeAggParam.BaseDecimal = 9
	evmcaller.InitCacher()
	config.Param().BSCParam.Host = []string{"https://data-seed-prebsc-1-s1.binance.org:8545"}
	config.Param().PLGParam.Host = []string{"https://polygon-mumbai.g.alchemy.com/v2/V8SP0S8Q-sT35ca4VKH3Iwyvh8K8wTRn"}
	config.Param().GethParam.Host = []string{"https://kovan.infura.io/v3/1138a1e99b154b10bae5c382ad894361"}
	config.Param().FTMParam.Host = []string{"https://rpc.testnet.fantom.network"}
	config.Param().EthContractAddressStr = "0x7bebc8445c6223b41b7bb4b0ae9742e2fd2f47f3"
	config.Param().BscContractAddressStr = "0xb51B25e6a0AEEC950379795bD80E2d42Bd7726Fb"
	config.Param().PlgContractAddressStr = "0x76eEE3fF9C8E651c669d7cfb69D10A67856325De"
	config.Param().FtmContractAddressStr = "0x526768A37feD86Fd8D5D72ca78913DFF64AC5E15"
	common.MaxShardNumber = 8

	rawTestCases, _ := readTestCases("shield.json")
	err := json.Unmarshal(rawTestCases, &s.testCases)
	if err != nil {
		panic(err)
	}
	s.actualResults = make(map[string]ActualResult)
}

func (s *ShieldTestSuite) SetupTest() {
	dbPath, err := ioutil.TempDir(os.TempDir(), "bridgeagg_test_statedb_")
	if err != nil {
		panic(err)
	}
	diskBD, _ := incdb.Open("leveldb", dbPath)
	warperDBStatedbTest := statedb.NewDatabaseAccessWarper(diskBD)
	emptyRoot := common.HexToHash(common.HexEmptyRoot)
	sDB, _ := statedb.NewWithPrefixTrie(emptyRoot, warperDBStatedbTest)
	s.sDB = sDB
}

func (s *ShieldTestSuite) BeforeTest(suiteName, testName string) {
	s.currentTestCaseName = testName
	testCase := s.testCases[s.currentTestCaseName]
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
		err := statedb.UpdateBridgeTokenInfo(s.sDB, tokenID, v.ExternalTokenID(), false, v.Amount(), "+")
		if err != nil {
			panic(err)
		}
	}

	assert := s.Assert()
	env := &stateEnvironment{
		beaconHeight: 10,
		stateDBs: map[int]*statedb.StateDB{
			common.BeaconChainID: s.sDB,
		},
		shieldActions:     [][]string{actions},
		accumulatedValues: testCase.AccumulatedValues,
	}
	state := NewState()
	state.unifiedTokenVaults = testCase.UnifiedTokens
	producerState := state.Clone()
	processorState := state.Clone()
	actualInstructions, ac, err := producerState.BuildInstructions(env)
	s.testCases[s.currentTestCaseName].ActualAccumulatedValues = ac
	assert.Nil(err, fmt.Sprintf("Error in build instructions %v", err))
	err = processorState.Process(actualInstructions, s.sDB)
	assert.Nil(err, fmt.Sprintf("Error in process instructions %v", err))
	s.actualResults[s.currentTestCaseName] = ActualResult{
		Instructions:   actualInstructions,
		ProducerState:  producerState,
		ProcessorState: processorState,
	}
	for _, txID := range testCase.TxIDs {
		shieldStatus := ShieldStatus{}
		prefixValues := [][]byte{
			{},
			{common.BoolToByte(false)},
			{common.BoolToByte(true)},
		}
		for _, prefixValue := range prefixValues {
			suffix := append(txID.Bytes(), prefixValue...)
			data, err := statedb.GetBridgeAggStatus(
				s.sDB,
				statedb.BridgeAggShieldStatusPrefix(),
				suffix,
			)
			if err != nil {
				continue
			}
			status := ShieldStatus{}
			err = json.Unmarshal(data, &status)
			assert.Nil(err, fmt.Sprintf("parse status error %v", err))
			shieldStatus.Status = status.Status
			if status.Status == common.RejectedStatusByte {
				shieldStatus.Data = nil
				shieldStatus.ErrorCode = status.ErrorCode
			} else {
				if len(shieldStatus.Data) == 0 {
					shieldStatus.Data = make([]ShieldStatusData, len(status.Data))
				}
				for i, v := range status.Data {
					shieldStatus.Data[i].Amount += v.Amount
					shieldStatus.Data[i].Reward += v.Reward
				}
			}
		}
		s.testCases[s.currentTestCaseName].ActualStatues = append(s.testCases[s.currentTestCaseName].ActualStatues, shieldStatus)
	}
}

func (s *ShieldTestSuite) TestAcceptedNotEqualTo0NativeToken() {
	assert := s.Assert()
	testCase := s.testCases[s.currentTestCaseName]
	actualResult := s.actualResults[s.currentTestCaseName]
	expectedState := NewState()
	expectedState.unifiedTokenVaults = testCase.ExpectedUnifiedTokens
	expectedStatuses := testCase.ExpectedStatuses
	actualStatuses := testCase.ActualStatues
	expectedAccumulatedValues := testCase.ExpectedAccumulatedValues
	actualAccumulatedValues := testCase.ActualAccumulatedValues
	assert.Equal(testCase.ExpectedInstructions, actualResult.Instructions, fmt.Errorf("Expected instructions %v but get %v", actualResult.Instructions, testCase.ExpectedInstructions).Error())
	assert.Equal(expectedState, actualResult.ProducerState, fmt.Errorf("Expected producer state %v but get %v", expectedState, actualResult.ProducerState).Error())
	assert.Equal(expectedState, actualResult.ProcessorState, fmt.Errorf("Expected processor state %v but get %v", expectedState, actualResult.ProcessorState).Error())
	assert.Equal(actualStatuses, expectedStatuses, fmt.Errorf("Expected statuses %v but get %v", expectedStatuses, actualStatuses).Error())
	assert.Equal(actualAccumulatedValues, expectedAccumulatedValues, fmt.Errorf("Expected accumulatedValues %v but get %v", expectedAccumulatedValues, actualAccumulatedValues).Error())
}

func (s *ShieldTestSuite) TestAcceptedNotEqualTo0NotNativeToken() {
	assert := s.Assert()
	testCase := s.testCases[s.currentTestCaseName]
	actualResult := s.actualResults[s.currentTestCaseName]
	expectedState := NewState()
	expectedState.unifiedTokenVaults = testCase.ExpectedUnifiedTokens
	expectedStatuses := testCase.ExpectedStatuses
	actualStatuses := testCase.ActualStatues
	expectedAccumulatedValues := testCase.ExpectedAccumulatedValues
	actualAccumulatedValues := testCase.ActualAccumulatedValues
	assert.Equal(testCase.ExpectedInstructions, actualResult.Instructions, fmt.Errorf("Expected instructions %v but get %v", actualResult.Instructions, testCase.ExpectedInstructions).Error())
	assert.Equal(expectedState, actualResult.ProducerState, fmt.Errorf("Expected producer state %v but get %v", expectedState, actualResult.ProducerState).Error())
	assert.Equal(expectedState, actualResult.ProcessorState, fmt.Errorf("Expected processor state %v but get %v", expectedState, actualResult.ProcessorState).Error())
	assert.Equal(actualStatuses, expectedStatuses, fmt.Errorf("Expected statuses %v but get %v", expectedStatuses, actualStatuses).Error())
	assert.Equal(actualAccumulatedValues, expectedAccumulatedValues, fmt.Errorf("Expected accumulatedValues %v but get %v", expectedAccumulatedValues, actualAccumulatedValues).Error())
}

func (s *ShieldTestSuite) TestAcceptedYEqualTo0NativeToken() {
	assert := s.Assert()
	testCase := s.testCases[s.currentTestCaseName]
	actualResult := s.actualResults[s.currentTestCaseName]
	expectedState := NewState()
	expectedState.unifiedTokenVaults = testCase.ExpectedUnifiedTokens
	expectedStatuses := testCase.ExpectedStatuses
	actualStatuses := testCase.ActualStatues
	expectedAccumulatedValues := testCase.ExpectedAccumulatedValues
	actualAccumulatedValues := testCase.ActualAccumulatedValues
	assert.Equal(testCase.ExpectedInstructions, actualResult.Instructions, fmt.Errorf("Expected instructions %v but get %v", actualResult.Instructions, testCase.ExpectedInstructions).Error())
	assert.Equal(expectedState, actualResult.ProducerState, fmt.Errorf("Expected producer state %v but get %v", expectedState, actualResult.ProducerState).Error())
	assert.Equal(expectedState, actualResult.ProcessorState, fmt.Errorf("Expected processor state %v but get %v", expectedState, actualResult.ProcessorState).Error())
	assert.Equal(expectedStatuses, actualStatuses, fmt.Errorf("Expected statuses %v but get %v", expectedStatuses, actualStatuses).Error())
	assert.Equal(expectedAccumulatedValues, actualAccumulatedValues, fmt.Errorf("Expected accumulatedValues %v but get %v", expectedAccumulatedValues, actualAccumulatedValues).Error())
}

func (s *ShieldTestSuite) TestAcceptedYEqualTo0NotNativeTokenDecimalGreaterBaseExternalDecimal() {
	assert := s.Assert()
	testCase := s.testCases[s.currentTestCaseName]
	actualResult := s.actualResults[s.currentTestCaseName]
	expectedState := NewState()
	expectedState.unifiedTokenVaults = testCase.ExpectedUnifiedTokens
	expectedStatuses := testCase.ExpectedStatuses
	actualStatuses := testCase.ActualStatues
	expectedAccumulatedValues := testCase.ExpectedAccumulatedValues
	actualAccumulatedValues := testCase.ActualAccumulatedValues
	assert.Equal(testCase.ExpectedInstructions, actualResult.Instructions, fmt.Errorf("Expected instructions %v but get %v", actualResult.Instructions, testCase.ExpectedInstructions).Error())
	assert.Equal(expectedState, actualResult.ProducerState, fmt.Errorf("Expected producer state %v but get %v", expectedState, actualResult.ProducerState).Error())
	assert.Equal(expectedState, actualResult.ProcessorState, fmt.Errorf("Expected processor state %v but get %v", expectedState, actualResult.ProcessorState).Error())
	assert.Equal(expectedStatuses, actualStatuses, fmt.Errorf("Expected statuses %v but get %v", expectedStatuses, actualStatuses).Error())
	assert.Equal(expectedAccumulatedValues, actualAccumulatedValues, fmt.Errorf("Expected accumulatedValues %v but get %v", expectedAccumulatedValues, actualAccumulatedValues).Error())
}

func (s *ShieldTestSuite) TestAcceptedYEqualTo0NotNativeTokenDecimalSmallerBaseExternalDecimal() {
	assert := s.Assert()
	testCase := s.testCases[s.currentTestCaseName]
	actualResult := s.actualResults[s.currentTestCaseName]
	expectedState := NewState()
	expectedState.unifiedTokenVaults = testCase.ExpectedUnifiedTokens
	expectedStatuses := testCase.ExpectedStatuses
	actualStatuses := testCase.ActualStatues
	expectedAccumulatedValues := testCase.ExpectedAccumulatedValues
	actualAccumulatedValues := testCase.ActualAccumulatedValues
	assert.Equal(testCase.ExpectedInstructions, actualResult.Instructions, fmt.Errorf("Expected instructions %v but get %v", actualResult.Instructions, testCase.ExpectedInstructions).Error())
	assert.Equal(expectedState, actualResult.ProducerState, fmt.Errorf("Expected producer state %v but get %v", expectedState, actualResult.ProducerState).Error())
	assert.Equal(expectedState, actualResult.ProcessorState, fmt.Errorf("Expected processor state %v but get %v", expectedState, actualResult.ProcessorState).Error())
	assert.Equal(expectedStatuses, actualStatuses, fmt.Errorf("Expected statuses %v but get %v", expectedStatuses, actualStatuses).Error())
	assert.Equal(expectedAccumulatedValues, actualAccumulatedValues, fmt.Errorf("Expected accumulatedValues %v but get %v", expectedAccumulatedValues, actualAccumulatedValues).Error())
}

func (s *ShieldTestSuite) TestRejectedInvalidExternalTokenID() {
	assert := s.Assert()
	testCase := s.testCases[s.currentTestCaseName]
	actualResult := s.actualResults[s.currentTestCaseName]
	expectedState := NewState()
	expectedState.unifiedTokenVaults = testCase.ExpectedUnifiedTokens
	expectedStatuses := testCase.ExpectedStatuses
	actualStatuses := testCase.ActualStatues
	expectedAccumulatedValues := testCase.ExpectedAccumulatedValues
	actualAccumulatedValues := testCase.ActualAccumulatedValues
	assert.Equal(testCase.ExpectedInstructions, actualResult.Instructions, fmt.Errorf("Expected instructions %v but get %v", actualResult.Instructions, testCase.ExpectedInstructions).Error())
	assert.Equal(expectedState, actualResult.ProducerState, fmt.Errorf("Expected producer state %v but get %v", expectedState, actualResult.ProducerState).Error())
	assert.Equal(expectedState, actualResult.ProcessorState, fmt.Errorf("Expected processor state %v but get %v", expectedState, actualResult.ProcessorState).Error())
	assert.Equal(expectedStatuses, actualStatuses, fmt.Errorf("Expected statuses %v but get %v", expectedStatuses, actualStatuses).Error())
	assert.Equal(expectedAccumulatedValues, actualAccumulatedValues, fmt.Errorf("Expected accumulatedValues %v but get %v", expectedAccumulatedValues, actualAccumulatedValues).Error())
}

func (s *ShieldTestSuite) TestRejectedInvalidTokenID() {
	assert := s.Assert()
	testCase := s.testCases[s.currentTestCaseName]
	actualResult := s.actualResults[s.currentTestCaseName]
	expectedState := NewState()
	expectedState.unifiedTokenVaults = testCase.ExpectedUnifiedTokens
	expectedStatuses := testCase.ExpectedStatuses
	actualStatuses := testCase.ActualStatues
	expectedAccumulatedValues := testCase.ExpectedAccumulatedValues
	actualAccumulatedValues := testCase.ActualAccumulatedValues
	assert.Equal(testCase.ExpectedInstructions, actualResult.Instructions, fmt.Errorf("Expected instructions %v but get %v", actualResult.Instructions, testCase.ExpectedInstructions).Error())
	assert.Equal(expectedState, actualResult.ProducerState, fmt.Errorf("Expected producer state %v but get %v", expectedState, actualResult.ProducerState).Error())
	assert.Equal(expectedState, actualResult.ProcessorState, fmt.Errorf("Expected processor state %v but get %v", expectedState, actualResult.ProcessorState).Error())
	assert.Equal(expectedStatuses, actualStatuses, fmt.Errorf("Expected statuses %v but get %v", expectedStatuses, actualStatuses).Error())
	assert.Equal(expectedAccumulatedValues, actualAccumulatedValues, fmt.Errorf("Expected accumulatedValues %v but get %v", expectedAccumulatedValues, actualAccumulatedValues).Error())
}

func (s *ShieldTestSuite) TestRejectedTwoProofs() {
	assert := s.Assert()
	testCase := s.testCases[s.currentTestCaseName]
	actualResult := s.actualResults[s.currentTestCaseName]
	expectedState := NewState()
	expectedState.unifiedTokenVaults = testCase.ExpectedUnifiedTokens
	expectedStatuses := testCase.ExpectedStatuses
	actualStatuses := testCase.ActualStatues
	expectedAccumulatedValues := testCase.ExpectedAccumulatedValues
	actualAccumulatedValues := testCase.ActualAccumulatedValues
	assert.Equal(testCase.ExpectedInstructions, actualResult.Instructions, fmt.Errorf("Expected instructions %v but get %v", actualResult.Instructions, testCase.ExpectedInstructions).Error())
	assert.Equal(expectedState, actualResult.ProducerState, fmt.Errorf("Expected producer state %v but get %v", expectedState, actualResult.ProducerState).Error())
	assert.Equal(expectedState, actualResult.ProcessorState, fmt.Errorf("Expected processor state %v but get %v", expectedState, actualResult.ProcessorState).Error())
	assert.Equal(expectedStatuses, actualStatuses, fmt.Errorf("Expected statuses %v but get %v", expectedStatuses, actualStatuses).Error())
	assert.Equal(expectedAccumulatedValues, actualAccumulatedValues, fmt.Errorf("Expected accumulatedValues %v but get %v", expectedAccumulatedValues, actualAccumulatedValues).Error())
}

func (s *ShieldTestSuite) TestRejectedTwoProofsInOneRequest() {
	assert := s.Assert()
	testCase := s.testCases[s.currentTestCaseName]
	actualResult := s.actualResults[s.currentTestCaseName]
	expectedState := NewState()
	expectedState.unifiedTokenVaults = testCase.ExpectedUnifiedTokens
	expectedStatuses := testCase.ExpectedStatuses
	actualStatuses := testCase.ActualStatues
	expectedAccumulatedValues := testCase.ExpectedAccumulatedValues
	actualAccumulatedValues := testCase.ActualAccumulatedValues
	assert.Equal(testCase.ExpectedInstructions, actualResult.Instructions, fmt.Errorf("Expected instructions %v but get %v", actualResult.Instructions, testCase.ExpectedInstructions).Error())
	assert.Equal(expectedState, actualResult.ProducerState, fmt.Errorf("Expected producer state %v but get %v", expectedState, actualResult.ProducerState).Error())
	assert.Equal(expectedState, actualResult.ProcessorState, fmt.Errorf("Expected processor state %v but get %v", expectedState, actualResult.ProcessorState).Error())
	assert.Equal(expectedStatuses, actualStatuses, fmt.Errorf("Expected statuses %v but get %v", expectedStatuses, actualStatuses).Error())
	assert.Equal(expectedAccumulatedValues, actualAccumulatedValues, fmt.Errorf("Expected accumulatedValues %v but get %v", expectedAccumulatedValues, actualAccumulatedValues).Error())
}

func (s *ShieldTestSuite) TestRejectedInvalidIncTokenID() {
	assert := s.Assert()
	testCase := s.testCases[s.currentTestCaseName]
	actualResult := s.actualResults[s.currentTestCaseName]
	expectedState := NewState()
	expectedState.unifiedTokenVaults = testCase.ExpectedUnifiedTokens
	expectedStatuses := testCase.ExpectedStatuses
	actualStatuses := testCase.ActualStatues
	expectedAccumulatedValues := testCase.ExpectedAccumulatedValues
	actualAccumulatedValues := testCase.ActualAccumulatedValues
	assert.Equal(testCase.ExpectedInstructions, actualResult.Instructions, fmt.Errorf("Expected instructions %v but get %v", actualResult.Instructions, testCase.ExpectedInstructions).Error())
	assert.Equal(expectedState, actualResult.ProducerState, fmt.Errorf("Expected producer state %v but get %v", expectedState, actualResult.ProducerState).Error())
	assert.Equal(expectedState, actualResult.ProcessorState, fmt.Errorf("Expected processor state %v but get %v", expectedState, actualResult.ProcessorState).Error())
	assert.Equal(expectedStatuses, actualStatuses, fmt.Errorf("Expected statuses %v but get %v", expectedStatuses, actualStatuses).Error())
	assert.Equal(expectedAccumulatedValues, actualAccumulatedValues, fmt.Errorf("Expected accumulatedValues %v but get %v", expectedAccumulatedValues, actualAccumulatedValues).Error())
}

func TestShieldTestSuite(t *testing.T) {
	suite.Run(t, new(ShieldTestSuite))
}
