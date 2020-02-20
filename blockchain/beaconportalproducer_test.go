package blockchain

import (
	"encoding/base64"
	"encoding/json"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/database/lvdb"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/stretchr/testify/suite"
	"strconv"
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
		metadata.PortalCustodianDepositMeta, incogAddressStr, remoteAddresses, depositedAmount)
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
