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
	Name         string                                           `json:"name"`
	Metadata     metadataBridge.ConvertTokenToUnifiedTokenRequest `json:"metadata"`
	Instructions [][]string                                       `json:"instructions"`
	State        *State                                           `json:"state"`
}

type ConvertTestSuite struct {
	suite.Suite
	producerState  *State
	processorState *State
	testCases      []ConvertTestCase

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
	c.processorState = NewState()
	c.producerState = NewState()

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
	fmt.Println(0)
	Logger.log.Info(c.testCases)
}

func (c *ConvertTestSuite) TestAcceptedConvert() {
}

func (c *ConvertTestSuite) TestRejectedConvert() {

}

func (c *ConvertTestSuite) TestRejectedThenAcceptedConvert() {

}

func TestConvertTestSuite(t *testing.T) {
	suite.Run(t, new(ConvertTestSuite))
}
