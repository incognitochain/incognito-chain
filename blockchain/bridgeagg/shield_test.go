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
	"github.com/incognitochain/incognito-chain/metadata/evmcaller"
	"github.com/stretchr/testify/suite"
)

type ShieldTestData struct {
	TestData
	Metadatas []*metadataBridge.ShieldRequest `json:"metadatas"`
}

type ShieldExpectedResult struct {
	ExpectedResult
	Statuses []ShieldStatus `json:"statuses"`
}

type ShieldActualResult struct {
	ActualResult
	Statuses []ShieldStatus
}

type ShieldTestCase struct {
	Data           ShieldTestData       `json:"data"`
	ExpectedResult ShieldExpectedResult `json:"expected_result"`
}

type ShieldTestSuite struct {
	testCases map[string]*ShieldTestCase
	TestSuite
	actualResults map[string]ShieldActualResult
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
	s.actualResults = make(map[string]ShieldActualResult)
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
		err := statedb.UpdateBridgeTokenInfo(s.sDB, tokenID, v.ExternalTokenID(), false, v.Amount(), "+")
		if err != nil {
			panic(err)
		}
	}

	env := &stateEnvironment{
		beaconHeight: 10,
		stateDBs: map[int]*statedb.StateDB{
			common.BeaconChainID: s.sDB,
		},
		shieldActions:     [][]string{actions},
		accumulatedValues: testCase.Data.AccumulatedValues,
	}
	s.testCases[s.currentTestCaseName].Data.env = env
}

func (s *ShieldTestSuite) test() {
	testCase := s.testCases[s.currentTestCaseName]
	assert := s.Assert()
	producerState := testCase.Data.State.Clone()
	producerManager := NewManagerWithValue(producerState)
	processorState := testCase.Data.State.Clone()
	processorManager := NewManagerWithValue(processorState)
	actualInstructions, accumulatedValues, err := producerManager.BuildInstructions(testCase.Data.env)
	assert.Nil(err, fmt.Sprintf("Error in build instructions %v", err))
	err = processorManager.Process(actualInstructions, s.sDB)
	assert.Nil(err, fmt.Sprintf("Error in process instructions %v", err))

	s.actualResults[s.currentTestCaseName] = ShieldActualResult{
		ActualResult: ActualResult{
			Instructions:      actualInstructions,
			ProducerState:     producerState,
			ProcessorState:    processorState,
			AccumulatedValues: accumulatedValues,
		},
	}

	for _, txID := range testCase.Data.TxIDs {
		data, err := statedb.GetBridgeAggStatus(
			s.sDB,
			statedb.BridgeAggShieldStatusPrefix(),
			txID.Bytes(),
		)
		assert.Nil(err, fmt.Sprintf("get bridge agg status %v", err))
		var status ShieldStatus
		err = json.Unmarshal(data, &status)
		assert.Nil(err, fmt.Sprintf("parse status error %v", err))
		s.testCases[s.currentTestCaseName].ExpectedResult.Statuses = append(s.testCases[s.currentTestCaseName].ExpectedResult.Statuses, status)
	}
}

func (s *ShieldTestSuite) TestAcceptedNativeToken() {
	s.test()
	assert := s.Assert()
	testCase := s.testCases[s.currentTestCaseName]
	actualResult := s.actualResults[s.currentTestCaseName]
	assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.Instructions, actualResult.Instructions))
	assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.State, actualResult.ProducerState))
	assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.State, actualResult.ProcessorState))
	assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.Statuses, actualResult.Statuses))
	assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.AccumulatedValues, actualResult.AccumulatedValues))
}

/*func (s *ShieldTestSuite) TestAcceptedNotNativeToken() {*/
/*s.test()*/
/*assert := s.Assert()*/
/*testCase := s.testCases[s.currentTestCaseName]*/
/*actualResult := s.actualResults[s.currentTestCaseName]*/
/*assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.Instructions, actualResult.Instructions))*/
/*assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.State, actualResult.ProducerState))*/
/*assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.State, actualResult.ProcessorState))*/
/*assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.Statuses, actualResult.Statuses))*/
/*assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.AccumulatedValues, actualResult.AccumulatedValues))*/
/*}*/

/*func (s *ShieldTestSuite) TestRejectedInvalidExternalTokenID() {*/
/*s.test()*/
/*assert := s.Assert()*/
/*testCase := s.testCases[s.currentTestCaseName]*/
/*actualResult := s.actualResults[s.currentTestCaseName]*/
/*assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.Instructions, actualResult.Instructions))*/
/*assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.State, actualResult.ProducerState))*/
/*assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.State, actualResult.ProcessorState))*/
/*assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.Statuses, actualResult.Statuses))*/
/*assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.AccumulatedValues, actualResult.AccumulatedValues))*/
/*}*/

/*func (s *ShieldTestSuite) TestRejectedInvalidTokenID() {*/
/*s.test()*/
/*assert := s.Assert()*/
/*testCase := s.testCases[s.currentTestCaseName]*/
/*actualResult := s.actualResults[s.currentTestCaseName]*/
/*assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.Instructions, actualResult.Instructions))*/
/*assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.State, actualResult.ProducerState))*/
/*assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.State, actualResult.ProcessorState))*/
/*assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.Statuses, actualResult.Statuses))*/
/*assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.AccumulatedValues, actualResult.AccumulatedValues))*/
/*}*/

/*func (s *ShieldTestSuite) TestRejectedTwoProofs() {*/
/*s.test()*/
/*assert := s.Assert()*/
/*testCase := s.testCases[s.currentTestCaseName]*/
/*actualResult := s.actualResults[s.currentTestCaseName]*/
/*assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.Instructions, actualResult.Instructions))*/
/*assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.State, actualResult.ProducerState))*/
/*assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.State, actualResult.ProcessorState))*/
/*assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.Statuses, actualResult.Statuses))*/
/*assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.AccumulatedValues, actualResult.AccumulatedValues))*/
/*}*/

/*func (s *ShieldTestSuite) TestRejectedTwoProofsInOneRequest() {*/
/*s.test()*/
/*assert := s.Assert()*/
/*testCase := s.testCases[s.currentTestCaseName]*/
/*actualResult := s.actualResults[s.currentTestCaseName]*/
/*assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.Instructions, actualResult.Instructions))*/
/*assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.State, actualResult.ProducerState))*/
/*assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.State, actualResult.ProcessorState))*/
/*assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.Statuses, actualResult.Statuses))*/
/*assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.AccumulatedValues, actualResult.AccumulatedValues))*/
/*}*/

/*func (s *ShieldTestSuite) TestRejectedInvalidIncTokenID() {*/
/*s.test()*/
/*assert := s.Assert()*/
/*testCase := s.testCases[s.currentTestCaseName]*/
/*actualResult := s.actualResults[s.currentTestCaseName]*/
/*assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.Instructions, actualResult.Instructions))*/
/*assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.State, actualResult.ProducerState))*/
/*assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.State, actualResult.ProcessorState))*/
/*assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.Statuses, actualResult.Statuses))*/
/*assert.Nil(CheckInterfacesIsEqual(testCase.ExpectedResult.AccumulatedValues, actualResult.AccumulatedValues))*/
/*}*/

func TestShieldTestSuite(t *testing.T) {
	suite.Run(t, new(ShieldTestSuite))
}
