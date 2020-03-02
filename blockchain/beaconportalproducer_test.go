package blockchain

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/database"
	"github.com/incognitochain/incognito-chain/database/lvdb"
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
		CustodianPoolState:     map[string]*lvdb.CustodianState{},
		ExchangeRatesRequests:  map[string]*lvdb.ExchangeRatesRequest{},
		WaitingPortingRequests: map[string]*lvdb.PortingRequest{},
		WaitingRedeemRequests:  map[string]*lvdb.RedeemRequest{},
	}
}

func buildPortalCustodianDepositAction(
	incogAddressStr string,
	remoteAddresses map[string]string,
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
	databaseInterface := new(mocks.DatabaseInterface)

	keyPortingRequest := lvdb.NewPortingRequestKeyForValidation("123")
	databaseInterface.On("GetItemPortalByPrefix", []byte(keyPortingRequest)).Return(
		nil,
		database.NewDatabaseError(database.GetItemPortalByPrefixError, errors.New("data not found")),
	).Once()

	config := &Config{
		DataBase: databaseInterface,
	}

	blockChain := &BlockChain{
		config: *config,
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
	currentPortalState := &CurrentPortalState{
		CustodianPoolState:     map[string]*lvdb.CustodianState{},
		ExchangeRatesRequests:  map[string]*lvdb.ExchangeRatesRequest{},
		WaitingPortingRequests: map[string]*lvdb.PortingRequest{},
		WaitingRedeemRequests:  map[string]*lvdb.RedeemRequest{},
	}

	instruct, err := blockChain.buildInstructionsForPortingRequest(
		actionContentBase64Str,
		1,
		1,
		currentPortalState,
		1,
	)

	result := instruct[0]
	assert.Equal(t, result[2], "LoadDataFailed")
}