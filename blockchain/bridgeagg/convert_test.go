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
	"github.com/stretchr/testify/suite"
)

var _ = func() (_ struct{}) {
	Logger.Init(common.NewBackend(nil).Logger("test", true))
	Logger.log.Info("Init logger")
	return
}()

type ConvertTestCase struct {
	Name           string                                           `json:"name"`
	Metadata       metadataBridge.ConvertTokenToUnifiedTokenRequest `json:"metadata"`
	Instructions   [][]string                                       `json:"instructions"`
	ProducerState  *State                                           `json:"producer_state"`
	ProcessorState *State                                           `json:"processor_state"`
}

type ConvertTestSuite struct {
	suite.Suite
	producerState        []*State
	processorState       []*State
	testCases            []ConvertTestCase
	currentTestCaseIndex int

	sdb *statedb.StateDB
	env *stateEnvironment
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

	env := &stateEnvironment{
		beaconHeight: 10,
		stateDBs: map[int]*statedb.StateDB{
			common.BeaconChainID: sDB,
		},
	}
	c.env = env
	rawTestCases, _ := readTestCases("convert.json")
	err = json.Unmarshal(rawTestCases, &c.testCases)
	if err != nil {
		panic(err)
	}
	c.currentTestCaseIndex = -1
}

func (c *ConvertTestSuite) SetupTest() {
	c.currentTestCaseIndex++
}

func (c *ConvertTestSuite) TestAcceptedConvert() {
	testCase := c.testCases[c.currentTestCaseIndex]
	fmt.Println(testCase)
}

func (c *ConvertTestSuite) TestRejectedConvert() {
	testCase := c.testCases[c.currentTestCaseIndex]
	fmt.Println(testCase)
}

func (c *ConvertTestSuite) TestRejectedThenAcceptedConvert() {
	testCase := c.testCases[c.currentTestCaseIndex]
	fmt.Println(testCase)
}

func TestConvertTestSuite(t *testing.T) {
	suite.Run(t, new(ConvertTestSuite))
}
