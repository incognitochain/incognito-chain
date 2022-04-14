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
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	"github.com/stretchr/testify/suite"
)

var _ = func() (_ struct{}) {
	Logger.Init(common.NewBackend(nil).Logger("test", true))
	Logger.log.Info("Init logger")
	return
}()

type ConvertTestCase struct {
	Name                 string                                           `json:"name"`
	Metadata             metadataBridge.ConvertTokenToUnifiedTokenRequest `json:"metadata"`
	ExpectedInstructions [][]string                                       `json:"expected_instructions"`
	State                *State                                           `json:"state"`
	ExpectedState        *State                                           `json:"expected_state"`
	TxID                 common.Hash                                      `json:"tx_id"`
}

type ActualResult struct {
	Instructions [][]string
	State        *State
}

type ConvertTestSuite struct {
	suite.Suite
	testCases            []ConvertTestCase
	currentTestCaseIndex int
	actualResults        []ActualResult

	sdb *statedb.StateDB
}

func (c *ConvertTestSuite) SetupSuite() {
	dbPath, err := ioutil.TempDir(os.TempDir(), "bridgeagg_test_statedb_")
	if err != nil {
		panic(err)
	}
	diskBD, _ := incdb.Open("leveldb", dbPath)
	warperDBStatedbTest := statedb.NewDatabaseAccessWarper(diskBD)
	emptyRoot := common.HexToHash(common.HexEmptyRoot)
	sDB, _ := statedb.NewWithPrefixTrie(emptyRoot, warperDBStatedbTest)
	c.sdb = sDB

	config.AbortParam()
	config.Param().BridgeAggParam.BaseDecimal = 9

	rawTestCases, _ := readTestCases("convert.json")
	err = json.Unmarshal(rawTestCases, &c.testCases)
	if err != nil {
		panic(err)
	}
	c.currentTestCaseIndex = -1
}

func (c *ConvertTestSuite) SetupTest() {
	c.currentTestCaseIndex++
	testCase := c.testCases[c.currentTestCaseIndex]
	action, err := metadataCommon.NewActionWithValue(&testCase.Metadata, testCase.TxID, nil).StringSlice(metadataCommon.BridgeAggConvertTokenToUnifiedTokenRequestMeta)
	assert := c.Assert()
	assert.Nil(err, fmt.Sprintf("Error in build action %v", err))
	fmt.Println("action:", action)

	env := &stateEnvironment{
		beaconHeight: 10,
		stateDBs: map[int]*statedb.StateDB{
			common.BeaconChainID: c.sdb,
		},
		convertActions: [][]string{action},
	}
	state := testCase.State.Clone()
	actualInstructions, err := state.BuildInstructions(env)
	assert.Nil(err, err.Error())
	c.actualResults = append(c.actualResults, ActualResult{
		Instructions: actualInstructions,
		State:        state,
	})
}

func (c *ConvertTestSuite) TestAcceptedConvert() {
	assert := c.Assert()
	testCase := c.testCases[c.currentTestCaseIndex]
	actualResult := c.actualResults[c.currentTestCaseIndex]
	assert.NotEqual(actualResult.Instructions, testCase.ExpectedInstructions, fmt.Errorf("Expected instructions %v but get %v", actualResult.Instructions, testCase.ExpectedInstructions).Error())
}

/*func (c *ConvertTestSuite) TestRejectedConvert() {*/
/*testCase := c.testCases[c.currentTestCaseIndex]*/
/*fmt.Println(testCase)*/
/*}*/

/*func (c *ConvertTestSuite) TestRejectedThenAcceptedConvert() {*/
/*testCase := c.testCases[c.currentTestCaseIndex]*/
/*fmt.Println(testCase)*/
/*}*/

func TestConvertTestSuite(t *testing.T) {
	suite.Run(t, new(ConvertTestSuite))
}
