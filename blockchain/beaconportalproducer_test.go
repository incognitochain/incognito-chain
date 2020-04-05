package blockchain

import (
	"encoding/base64"
	"encoding/json"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	mocks "github.com/incognitochain/incognito-chain/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"strconv"
	"testing"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type PortalProducerSuite struct {
	suite.Suite
	currentPortalState *CurrentPortalState
}

func (suite *PortalProducerSuite) SetupTest() {
	suite.currentPortalState = &CurrentPortalState{
		CustodianPoolState:     map[string]*statedb.CustodianState{},
		ExchangeRatesRequests:  map[string]*metadata.ExchangeRatesRequestStatus{},
		WaitingPortingRequests: map[string]*statedb.WaitingPortingRequest{},
		WaitingRedeemRequests:  map[string]*statedb.WaitingRedeemRequest{},
	}
}

func buildPortalCustodianDepositAction(
	incogAddressStr string,
	remoteAddresses []statedb.RemoteAddress,
	depositedAmount uint64,
) []string {
	custodianDepositMeta, _ := metadata.NewPortalCustodianDeposit(
		metadata.PortalCustodianDepositMeta,
		incogAddressStr,
		remoteAddresses,
		depositedAmount,
	)

	shardID := byte(0)
	actionContent := metadata.PortalCustodianDepositAction{
		Meta:    *custodianDepositMeta,
		TxReqID: common.Hash{},
		ShardID: shardID,
	}
	actionContentBytes, _ := json.Marshal(actionContent)

	actionContentBase64Str := base64.StdEncoding.EncodeToString(actionContentBytes)
	action := []string{strconv.Itoa(metadata.PortalCustodianDepositMeta), actionContentBase64Str}
	return action
}

//func (suite *PortalProducerSuite) CustodianDepositOnEmptyCustodianPool() {
//	fmt.Println("Testing CustodianDepositOnEmptyCustodianPool")
//
//	// setup suite
//	suite := new(PortalProducerSuite)
//	suite.SetupTest()
//
//	// build new portal custodian deposit action (from shard)
//	action := buildPortalCustodianDepositAction(
//		""
//	)
//}


func TestBuildInstructionsForPortingRequest(t *testing.T)  {
	trieMock := new(mocks.Trie)

	keyPortingRequest := "123456789"
	trieMock.On("GetKey").Return(
		nil,
	).Once()

	trieMock.On("TryGet", []byte(keyPortingRequest)).Return(
		0,
		nil,
	)

	root1 := common.Hash{}
	wrapperDBMock := new(mocks.DatabaseAccessWarper)
	wrapperDBMock.On("OpenPrefixTrie",root1).Return(
		trieMock,
		nil,
	)

	wrapperDBMock.On("CopyTrie",trieMock).Return(
		trieMock,
		nil,
	)


	root := common.Hash{}
	stateDb, err := statedb.NewWithPrefixTrie(root, wrapperDBMock)
	if err != nil || stateDb == nil {
		t.Fatal(err, stateDb)
	}

	beaconBestState := &BeaconBestState{
		featureStateDB: stateDb,
	}

	bestState := &BestState{
		Beacon: beaconBestState,
	}

	blockChain := &BlockChain{
		BestState: bestState,
	}

	//case: wrong input data
	value, err := blockChain.buildInstructionsForPortingRequest(
		"Test",
		1,
		1,
		nil,
		1,
		)

	assert.Equal(t, err, nil)
	assert.Equal(t, value, [][]string{})

	//case: current portal state nil
	meta, _ := metadata.NewPortalUserRegister(
		"123",
		"123",
		"BNB",
		uint64(1000),
		uint64(4),
		metadata.PortalUserRegisterMeta,
	)

	action := metadata.PortalUserRegisterAction{
		Meta : *meta,
		TxReqID: common.HashH([]byte("test")),
		ShardID: 123,
	}

	actionContentBytes, _ := json.Marshal(action)
	actionContentBase64Str := base64.StdEncoding.EncodeToString(actionContentBytes)

	_, err = blockChain.buildInstructionsForPortingRequest(
		actionContentBase64Str,
		1,
		1,
		nil,
		1,
	)

	assert.Equal(t, err, nil)
	assert.Equal(t, value, [][]string{})

	//case: check unique id from record from db
	/*currentPortalState := &CurrentPortalState{
		CustodianPoolState:     map[string]*statedb.CustodianState{},
		ExchangeRatesRequests:  map[string]*metadata.ExchangeRatesRequestStatus{},
		WaitingPortingRequests: map[string]*statedb.WaitingPortingRequest{},
		WaitingRedeemRequests:  map[string]*statedb.WaitingRedeemRequest{},
	}

	instruct, err := blockChain.buildInstructionsForPortingRequest(
		actionContentBase64Str,
		1,
		1,
		currentPortalState,
		1,
	)

	result := instruct[0]
	assert.Equal(t, result[2], "rejected")*/
}