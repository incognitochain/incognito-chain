package bridgeagg

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/incognitochain/incognito-chain/metadata"
	metadataBridge "github.com/incognitochain/incognito-chain/metadata/bridge"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	metadataMocks "github.com/incognitochain/incognito-chain/metadata/common/mocks"
	"github.com/incognitochain/incognito-chain/metadata/evmcaller"
	"github.com/stretchr/testify/suite"
)

type TestCase struct {
	Name         string
	BeaconHeight uint64                                              `json:"BeaconHeight"`
	ConvertReqs  []*metadataBridge.ConvertTokenToUnifiedTokenRequest `json:"ConvertReqs"`
	ShieldReqs   []*metadataBridge.ShieldRequest                     `json:"ShieldReqs"`
	UnshieldReqs []*metadataBridge.UnshieldRequest                   `json:"UnshieldReqs"`
}

type StateInfo struct {
	BridgeAggState    *State                                        `json:"BridgeAggState"`
	BridgeTokensInfo  map[common.Hash]*statedb.BridgeTokenInfoState `json:"BridgeTokensInfo"`
	AccumulatedValues *metadata.AccumulatedValues                   `json:"AccumulatedValues"`
}

type SequenceTestSuite struct {
	suite.Suite
	sDB                  *statedb.StateDB
	InitializedStateInfo StateInfo `json:"InitializedStateInfo"`
	TestCases            []TestCase
	ExpectedResult       []StateInfo
}

type SequenceData struct {
	InitializedStateInfo StateInfo   `json:"InitializedStateInfo"`
	TestCases            []TestCase  `json:"TestCases"`
	ExpectedStateInfos   []StateInfo `json:"ExpectedStateInfos"`
}

func (s *SequenceTestSuite) SetupSuite() {

}

func (s *SequenceTestSuite) SetupTest() {
	// init evm caller
	evmcaller.InitCacher()

	// init config
	config.AbortParam()
	config.Param().BridgeAggParam.BaseDecimal = 9
	config.Param().BridgeAggParam.MaxLenOfPath = 3
	config.Param().BridgeAggParam.PercentFeeDecimal = 1e6
	config.Param().GethParam.Host = []string{"https://kovan.infura.io/v3/1138a1e99b154b10bae5c382ad894361"}
	config.Param().BSCParam.Host = []string{"https://data-seed-prebsc-1-s1.binance.org:8545"}
	config.Param().PLGParam.Host = []string{"https://polygon-mumbai.g.alchemy.com/v2/V8SP0S8Q-sT35ca4VKH3Iwyvh8K8wTRn"}
	config.Param().FTMParam.Host = []string{"https://rpc.testnet.fantom.network"}
	config.Param().EthContractAddressStr = "0xf90860014c7e13dE4A86B81c54dDE797820c72bE"
	config.Param().BscContractAddressStr = "0x9ee4E8FE2D977c8869F2e29736b21e8fc6FF830E"
	config.Param().PlgContractAddressStr = "0x141aa0C4c7d27f7B526a254202F979B60F056333"
	config.Param().FtmContractAddressStr = "0x526768A37feD86Fd8D5D72ca78913DFF64AC5E15"
	common.MaxShardNumber = 1

	sequenceData := SequenceData{}

	rawSequenceData, _ := readTestCases("sequence.json")
	err := json.Unmarshal(rawSequenceData, &sequenceData)
	if err != nil {
		panic(err)
	}

	s.InitializedStateInfo = sequenceData.InitializedStateInfo
	s.TestCases = sequenceData.TestCases
	s.ExpectedResult = sequenceData.ExpectedStateInfos

	dbPath, err := ioutil.TempDir(os.TempDir(), "bridgeagg_test_statedb_")
	if err != nil {
		panic(err)
	}
	diskBD, _ := incdb.Open("leveldb", dbPath)
	warperDBStatedbTest := statedb.NewDatabaseAccessWarper(diskBD)

	emptyRoot := common.HexToHash(common.HexEmptyRoot)
	sDB, _ := statedb.NewWithPrefixTrie(emptyRoot, warperDBStatedbTest)
	s.sDB = sDB

	fmt.Printf("s: %+v\n", s)
}

func (s *SequenceTestSuite) TestSequence() {
	convertActions := make([][]string, 1)
	shieldActions := make([][]string, 1)
	unshieldActionForProducers := []UnshieldActionForProducer{}
	accumulatedValues := &metadataCommon.AccumulatedValues{
		UniqETHTxsUsed:    [][]byte{},
		UniqBSCTxsUsed:    [][]byte{},
		UniqPRVEVMTxsUsed: [][]byte{},
		UniqPLGTxsUsed:    [][]byte{},
		UniqFTMTxsUsed:    [][]byte{},
		DBridgeTokenPair:  map[string][]byte{},
		CBridgeTokens:     []*common.Hash{},
		InitTokens:        []*common.Hash{},
	}

	managerProducer := NewManagerWithValue(s.InitializedStateInfo.BridgeAggState)
	managerProcess := NewManagerWithValue(s.InitializedStateInfo.BridgeAggState.Clone())
	for i, tc := range s.TestCases {
		beaconHeight := tc.BeaconHeight
		// build action
		for i, convertReq := range tc.ConvertReqs {
			txId := common.HashH([]byte("Convert" + fmt.Sprint(i)))
			content, _ := metadataCommon.NewActionWithValue(convertReq, txId, nil).StringSlice(metadataCommon.BridgeAggConvertTokenToUnifiedTokenRequestMeta)
			convertActions[0] = append(convertActions[0], content[1])
		}

		for i, shieldReq := range tc.ShieldReqs {
			tx := &metadataMocks.Transaction{}
			txId := common.HashH([]byte("Shield" + fmt.Sprint(i)))
			tx.On("Hash").Return(&txId)

			content, _ := shieldReq.BuildReqActions(tx, nil, nil, nil, 0, 0)
			shieldActions[0] = append(shieldActions[0], content[0][1])
		}

		for i, unshieldReq := range tc.UnshieldReqs {
			txId := common.HashH([]byte("Unshield" + fmt.Sprint(i)))
			action := metadataCommon.NewActionWithValue(unshieldReq, txId, nil)
			unshieldActionForProducers = append(unshieldActionForProducers, UnshieldActionForProducer{
				Action:       *action,
				ShardID:      0,
				BeaconHeight: beaconHeight,
			})
		}

		// build bridge agg enviroment
		bridgeAggEnv :=
			NewStateEnvBuilder().
				BuildConvertActions(convertActions).
				BuildShieldActions(shieldActions).
				BuildAccumulatedValues(accumulatedValues).
				BuildBeaconHeight(beaconHeight).
				BuildStateDBs(map[int]*statedb.StateDB{-1: s.sDB}).
				Build()

		// beacon producer
		inst, accumulatedValues, err := managerProducer.BuildInstructions(bridgeAggEnv)
		s.Assert().Equal(nil, err, "Beacon producer instructions error")

		unshieldInsts, err := managerProducer.BuildNewUnshieldInstructions(s.sDB, beaconHeight, unshieldActionForProducers)
		s.Assert().Equal(nil, err, "Beacon producer unshield instructions error")

		newInsts := append(inst, unshieldInsts...)

		err = managerProcess.Process(newInsts, s.sDB)
		s.Assert().Equal(nil, err, "Beacon process instructions error")

		s.Assert().Equal(true, reflect.DeepEqual(managerProducer.state, s.ExpectedResult[i].BridgeAggState), "Different state producer")
		s.Assert().Equal(true, reflect.DeepEqual(managerProducer.state, managerProcess.state), "Different state producer and state process")
		s.Assert().Equal(true, reflect.DeepEqual(accumulatedValues, s.ExpectedResult[i].AccumulatedValues), "Different accumulated value")
		s.Assert().Equal(true, reflect.DeepEqual(len(newInsts), 6), "Different instructions length")
	}
}

func TestSequenceBridgeAggSuite(t *testing.T) {
	suite.Run(t, new(SequenceTestSuite))
}
