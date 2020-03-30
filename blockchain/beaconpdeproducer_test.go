package blockchain

// Basic imports
import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"strconv"
	"testing"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type PDEProducerSuite struct {
	suite.Suite
	currentPDEState *CurrentPDEState
}

func (suite *PDEProducerSuite) SetupTest() {
	suite.currentPDEState = &CurrentPDEState{
		WaitingPDEContributions: make(map[string]*rawdbv2.PDEContribution),
		PDEPoolPairs:            make(map[string]*rawdbv2.PDEPoolForPair),
		PDEShares:               make(map[string]uint64),
	}
}

func buildPDETradeReqAction(
	tokenIDToBuyStr string,
	tokenIDToSellStr string,
	sellAmount uint64,
	traderAddressStr string,
) []string {
	metadataBase := metadata.MetadataBase{
		Type: metadata.PDETradeRequestMeta,
	}
	pdeTradeRequest := metadata.PDETradeRequest{
		TokenIDToBuyStr:  tokenIDToBuyStr,
		TokenIDToSellStr: tokenIDToSellStr,
		SellAmount:       sellAmount,
		TraderAddressStr: traderAddressStr,
	}
	pdeTradeRequest.MetadataBase = metadataBase
	actionContent := metadata.PDETradeRequestAction{
		Meta:    pdeTradeRequest,
		TxReqID: common.Hash{},
		ShardID: 1,
	}
	actionContentBytes, _ := json.Marshal(actionContent)
	actionContentBase64Str := base64.StdEncoding.EncodeToString(actionContentBytes)
	return []string{strconv.Itoa(metadata.PDETradeRequestMeta), actionContentBase64Str}
}

// All methods that begin with "Test" are run as tests within a
// suite.
func (suite *PDEProducerSuite) TestTradeOnNoAnyExistedPair() {
	fmt.Println("Running testcase: TestTradeOnNoAnyExistedPair")
	reqAction := buildPDETradeReqAction(
		"0000000000000000000000000000000000000000000000000000000000000005",
		"0000000000000000000000000000000000000000000000000000000000000006",
		10000000000,
		"12S2jM1TBbX2V5TBTvpJkJmsdaYxbCspGNedQkvJpYcbnV4gad7FDEbzY9P3zbpZRJTsGD5vxJRia3UiiUwMUbXbjfgezewq6rtPNtj",
	)
	// insts := [][]string{reqAction}
	metaType, _ := strconv.Atoi(reqAction[0])
	contentStr := reqAction[1]
	shardID := byte(1)
	beaconHeight := uint64(1001)
	bc := &BlockChain{}
	newInsts, err := bc.buildInstructionsForPDETrade(contentStr, shardID, metaType, suite.currentPDEState, beaconHeight-1)
	suite.Equal(err, nil)
	suite.Equal(len(newInsts), 1)
	suite.Equal(len(newInsts[0]), 4)
	suite.Equal(newInsts[0][0], strconv.Itoa(metaType))
	suite.Equal(newInsts[0][1], strconv.Itoa(int(shardID)))
	suite.Equal(newInsts[0][2], "refund")
	suite.Equal(newInsts[0][3], contentStr)
	suite.Equal(len(suite.currentPDEState.PDEPoolPairs), 0)
}

func (suite *PDEProducerSuite) TestTradeOnUnexistedPair() {
	fmt.Println("Running testcase: TestTradeOnUnexistedPair")
	beaconHeight := uint64(1001)
	pair := rawdbv2.PDEPoolForPair{
		Token1IDStr:     "0000000000000000000000000000000000000000000000000000000000000005",
		Token1PoolValue: 500000000000,
		Token2IDStr:     "0000000000000000000000000000000000000000000000000000000000000007",
		Token2PoolValue: 60000000000000,
	}
	pairKey := string(rawdbv2.BuildPDEPoolForPairKey(beaconHeight-1, pair.Token1IDStr, pair.Token2IDStr))
	suite.currentPDEState.PDEPoolPairs = map[string]*rawdbv2.PDEPoolForPair{
		pairKey: &pair,
	}
	reqAction := buildPDETradeReqAction(
		"0000000000000000000000000000000000000000000000000000000000000005",
		"0000000000000000000000000000000000000000000000000000000000000006",
		10000000000,
		"12S2jM1TBbX2V5TBTvpJkJmsdaYxbCspGNedQkvJpYcbnV4gad7FDEbzY9P3zbpZRJTsGD5vxJRia3UiiUwMUbXbjfgezewq6rtPNtj",
	)
	metaType, _ := strconv.Atoi(reqAction[0])
	contentStr := reqAction[1]
	shardID := byte(1)
	bc := &BlockChain{}
	newInsts, err := bc.buildInstructionsForPDETrade(contentStr, shardID, metaType, suite.currentPDEState, beaconHeight-1)
	suite.Equal(err, nil)
	suite.Equal(len(newInsts), 1)
	suite.Equal(len(newInsts[0]), 4)
	suite.Equal(newInsts[0][0], strconv.Itoa(metaType))
	suite.Equal(newInsts[0][1], strconv.Itoa(int(shardID)))
	suite.Equal(newInsts[0][2], "refund")
	suite.Equal(newInsts[0][3], contentStr)

	remainingTk1PoolVal := suite.currentPDEState.PDEPoolPairs[pairKey].Token1PoolValue
	remainingTk2PoolVal := suite.currentPDEState.PDEPoolPairs[pairKey].Token2PoolValue
	suite.Equal(remainingTk1PoolVal, pair.Token1PoolValue)
	suite.Equal(remainingTk2PoolVal, pair.Token2PoolValue)
}

func (suite *PDEProducerSuite) TestBuyToken1OnExistedPair() {
	fmt.Println("Running testcase: TestBuyToken1OnExistedPair")
	beaconHeight := uint64(1001)
	pair := rawdbv2.PDEPoolForPair{
		Token1IDStr:     "0000000000000000000000000000000000000000000000000000000000000005",
		Token1PoolValue: 500000000000,
		Token2IDStr:     "0000000000000000000000000000000000000000000000000000000000000007",
		Token2PoolValue: 60000000000000,
	}
	pairKey := string(rawdbv2.BuildPDEPoolForPairKey(beaconHeight-1, pair.Token1IDStr, pair.Token2IDStr))
	suite.currentPDEState.PDEPoolPairs = map[string]*rawdbv2.PDEPoolForPair{
		pairKey: &pair,
	}

	reqAction := buildPDETradeReqAction(
		"0000000000000000000000000000000000000000000000000000000000000005",
		"0000000000000000000000000000000000000000000000000000000000000007",
		10000000000,
		"12S2jM1TBbX2V5TBTvpJkJmsdaYxbCspGNedQkvJpYcbnV4gad7FDEbzY9P3zbpZRJTsGD5vxJRia3UiiUwMUbXbjfgezewq6rtPNtj",
	)
	metaType, _ := strconv.Atoi(reqAction[0])
	contentStr := reqAction[1]
	shardID := byte(1)
	bc := &BlockChain{}

	pdeTradeAcceptedContent := metadata.PDETradeAcceptedContent{
		TraderAddressStr: "12S2jM1TBbX2V5TBTvpJkJmsdaYxbCspGNedQkvJpYcbnV4gad7FDEbzY9P3zbpZRJTsGD5vxJRia3UiiUwMUbXbjfgezewq6rtPNtj",
		TokenIDToBuyStr:  "0000000000000000000000000000000000000000000000000000000000000005",
		ReceiveAmount:    83111182,
		Token1IDStr:      "0000000000000000000000000000000000000000000000000000000000000005",
		Token2IDStr:      "0000000000000000000000000000000000000000000000000000000000000007",
		ShardID:          shardID,
		RequestedTxID:    common.Hash{},
	}
	pdeTradeAcceptedContent.Token1PoolValueOperation = metadata.TokenPoolValueOperation{
		Operator: "-",
		Value:    83111182,
	}
	pdeTradeAcceptedContent.Token2PoolValueOperation = metadata.TokenPoolValueOperation{
		Operator: "+",
		Value:    10000000000,
	}
	pdeTradeAcceptedContentBytes, _ := json.Marshal(pdeTradeAcceptedContent)
	newInsts, err := bc.buildInstructionsForPDETrade(contentStr, shardID, metaType, suite.currentPDEState, beaconHeight-1)

	suite.Equal(err, nil)
	suite.Equal(len(newInsts), 1)
	suite.Equal(len(newInsts[0]), 4)
	suite.Equal(newInsts[0][0], strconv.Itoa(metaType))
	suite.Equal(newInsts[0][1], strconv.Itoa(int(shardID)))
	suite.Equal(newInsts[0][2], "accepted")
	suite.Equal(newInsts[0][3], string(pdeTradeAcceptedContentBytes))

	remainingTk1PoolVal := suite.currentPDEState.PDEPoolPairs[pairKey].Token1PoolValue
	remainingTk2PoolVal := suite.currentPDEState.PDEPoolPairs[pairKey].Token2PoolValue
	suite.Equal(remainingTk1PoolVal, uint64(500000000000-83111182))
	suite.Equal(remainingTk2PoolVal, uint64(60000000000000+10000000000))
}

func (suite *PDEProducerSuite) TestBuyToken2OnExistedPair() {
	fmt.Println("Running testcase: TestBuyToken2OnExistedPair")
	beaconHeight := uint64(1001)
	pair := rawdbv2.PDEPoolForPair{
		Token1IDStr:     "0000000000000000000000000000000000000000000000000000000000000005",
		Token1PoolValue: 500000000000,
		Token2IDStr:     "0000000000000000000000000000000000000000000000000000000000000007",
		Token2PoolValue: 60000000000000,
	}
	pairKey := string(rawdbv2.BuildPDEPoolForPairKey(beaconHeight-1, pair.Token1IDStr, pair.Token2IDStr))
	suite.currentPDEState.PDEPoolPairs = map[string]*rawdbv2.PDEPoolForPair{
		pairKey: &pair,
	}

	reqAction := buildPDETradeReqAction(
		"0000000000000000000000000000000000000000000000000000000000000007",
		"0000000000000000000000000000000000000000000000000000000000000005",
		10000000000,
		"12S2jM1TBbX2V5TBTvpJkJmsdaYxbCspGNedQkvJpYcbnV4gad7FDEbzY9P3zbpZRJTsGD5vxJRia3UiiUwMUbXbjfgezewq6rtPNtj",
	)
	metaType, _ := strconv.Atoi(reqAction[0])
	contentStr := reqAction[1]
	shardID := byte(1)
	bc := &BlockChain{}

	pdeTradeAcceptedContent := metadata.PDETradeAcceptedContent{
		TraderAddressStr: "12S2jM1TBbX2V5TBTvpJkJmsdaYxbCspGNedQkvJpYcbnV4gad7FDEbzY9P3zbpZRJTsGD5vxJRia3UiiUwMUbXbjfgezewq6rtPNtj",
		TokenIDToBuyStr:  "0000000000000000000000000000000000000000000000000000000000000007",
		ReceiveAmount:    1173586940536,
		Token1IDStr:      "0000000000000000000000000000000000000000000000000000000000000005",
		Token2IDStr:      "0000000000000000000000000000000000000000000000000000000000000007",
		ShardID:          shardID,
		RequestedTxID:    common.Hash{},
	}
	pdeTradeAcceptedContent.Token2PoolValueOperation = metadata.TokenPoolValueOperation{
		Operator: "-",
		Value:    1173586940536,
	}
	pdeTradeAcceptedContent.Token1PoolValueOperation = metadata.TokenPoolValueOperation{
		Operator: "+",
		Value:    10000000000,
	}
	pdeTradeAcceptedContentBytes, _ := json.Marshal(pdeTradeAcceptedContent)
	newInsts, err := bc.buildInstructionsForPDETrade(contentStr, shardID, metaType, suite.currentPDEState, beaconHeight-1)

	suite.Equal(err, nil)
	suite.Equal(len(newInsts), 1)
	suite.Equal(len(newInsts[0]), 4)
	suite.Equal(newInsts[0][0], strconv.Itoa(metaType))
	suite.Equal(newInsts[0][1], strconv.Itoa(int(shardID)))
	suite.Equal(newInsts[0][2], "accepted")
	suite.Equal(newInsts[0][3], string(pdeTradeAcceptedContentBytes))

	remainingTk1PoolVal := suite.currentPDEState.PDEPoolPairs[pairKey].Token1PoolValue
	remainingTk2PoolVal := suite.currentPDEState.PDEPoolPairs[pairKey].Token2PoolValue
	suite.Equal(remainingTk1PoolVal, uint64(500000000000+10000000000))
	suite.Equal(remainingTk2PoolVal, uint64(60000000000000-1173586940536))
}

func (suite *PDEProducerSuite) TestSellVerySmallAmtOnExistedPair() {
	fmt.Println("Running testcase: TestSellVerySmallAmtOnExistedPair")
	beaconHeight := uint64(1001)
	pair := rawdbv2.PDEPoolForPair{
		Token1IDStr:     "0000000000000000000000000000000000000000000000000000000000000005",
		Token1PoolValue: 500000000000,
		Token2IDStr:     "0000000000000000000000000000000000000000000000000000000000000007",
		Token2PoolValue: 60000000000000,
	}
	pairKey := string(rawdbv2.BuildPDEPoolForPairKey(beaconHeight-1, pair.Token1IDStr, pair.Token2IDStr))
	suite.currentPDEState.PDEPoolPairs = map[string]*rawdbv2.PDEPoolForPair{
		pairKey: &pair,
	}

	reqAction := buildPDETradeReqAction(
		"0000000000000000000000000000000000000000000000000000000000000005",
		"0000000000000000000000000000000000000000000000000000000000000007",
		1,
		"12S2jM1TBbX2V5TBTvpJkJmsdaYxbCspGNedQkvJpYcbnV4gad7FDEbzY9P3zbpZRJTsGD5vxJRia3UiiUwMUbXbjfgezewq6rtPNtj",
	)
	metaType, _ := strconv.Atoi(reqAction[0])
	contentStr := reqAction[1]
	shardID := byte(1)
	bc := &BlockChain{}

	newInsts, err := bc.buildInstructionsForPDETrade(contentStr, shardID, metaType, suite.currentPDEState, beaconHeight-1)

	suite.Equal(err, nil)
	suite.Equal(len(newInsts), 1)
	suite.Equal(len(newInsts[0]), 4)
	suite.Equal(newInsts[0][0], strconv.Itoa(metaType))
	suite.Equal(newInsts[0][1], strconv.Itoa(int(shardID)))
	suite.Equal(newInsts[0][2], "refund")
	suite.Equal(newInsts[0][3], contentStr)

	remainingTk1PoolVal := suite.currentPDEState.PDEPoolPairs[pairKey].Token1PoolValue
	remainingTk2PoolVal := suite.currentPDEState.PDEPoolPairs[pairKey].Token2PoolValue
	suite.Equal(remainingTk1PoolVal, pair.Token1PoolValue)
	suite.Equal(remainingTk2PoolVal, pair.Token2PoolValue)
}

func buildPDEWithdrawReqAction(
	withdrawerAddressStr string,
	withdrawalToken1IDStr string,
	withdrawalShare1Amt uint64,
	withdrawalToken2IDStr string,
	withdrawalShare2Amt uint64,
) []string {
	metadataBase := metadata.MetadataBase{
		Type: metadata.PDEWithdrawalRequestMeta,
	}
	// PLEASE UPDATE THIS TEST
	pdeWithdrawalRequest := metadata.PDEWithdrawalRequest{
		WithdrawerAddressStr:  withdrawerAddressStr,
		WithdrawalToken1IDStr: withdrawalToken1IDStr,
		WithdrawalShareAmt:    withdrawalShare1Amt,
		WithdrawalToken2IDStr: withdrawalToken2IDStr,
		// TMP FIX FOR TEST
		// WithdrawalShare2Amt:   withdrawalShare2Amt,
	}
	pdeWithdrawalRequest.MetadataBase = metadataBase
	actionContent := metadata.PDEWithdrawalRequestAction{
		Meta:    pdeWithdrawalRequest,
		TxReqID: common.Hash{},
		ShardID: 1,
	}
	actionContentBytes, _ := json.Marshal(actionContent)
	actionContentBase64Str := base64.StdEncoding.EncodeToString(actionContentBytes)
	return []string{strconv.Itoa(metadata.PDEWithdrawalRequestMeta), actionContentBase64Str}
}

func (suite *PDEProducerSuite) TestWithdrawOnExistedPair() {
	fmt.Println("Running testcase: TestWithdrawOnExistedPair")
	beaconHeight := uint64(1001)
	pair := rawdbv2.PDEPoolForPair{
		Token1IDStr:     "0000000000000000000000000000000000000000000000000000000000000005",
		Token1PoolValue: 500000000000,
		Token2IDStr:     "0000000000000000000000000000000000000000000000000000000000000007",
		Token2PoolValue: 60000000000000,
	}
	pairKey := string(rawdbv2.BuildPDEPoolForPairKey(beaconHeight-1, pair.Token1IDStr, pair.Token2IDStr))
	suite.currentPDEState.PDEPoolPairs = map[string]*rawdbv2.PDEPoolForPair{
		pairKey: &pair,
	}

	shareKey1 := string(rawdbv2.BuildPDESharesKey(
		beaconHeight-1,
		"0000000000000000000000000000000000000000000000000000000000000005",
		"0000000000000000000000000000000000000000000000000000000000000007",
		"0000000000000000000000000000000000000000000000000000000000000005",
		"12S2jM1TBbX2V5TBTvpJkJmsdaYxbCspGNedQkvJpYcbnV4gad7FDEbzY9P3zbpZRJTsGD5vxJRia3UiiUwMUbXbjfgezewq6rtPNtj",
	))
	shareKey2 := string(rawdbv2.BuildPDESharesKey(
		beaconHeight-1,
		"0000000000000000000000000000000000000000000000000000000000000005",
		"0000000000000000000000000000000000000000000000000000000000000007",
		"0000000000000000000000000000000000000000000000000000000000000007",
		"12S2jM1TBbX2V5TBTvpJkJmsdaYxbCspGNedQkvJpYcbnV4gad7FDEbzY9P3zbpZRJTsGD5vxJRia3UiiUwMUbXbjfgezewq6rtPNtj",
	))
	suite.currentPDEState.PDEShares = map[string]uint64{
		shareKey1: 500000000000,
		shareKey2: 60000000000000,
	}

	reqAction := buildPDEWithdrawReqAction(
		"12S2jM1TBbX2V5TBTvpJkJmsdaYxbCspGNedQkvJpYcbnV4gad7FDEbzY9P3zbpZRJTsGD5vxJRia3UiiUwMUbXbjfgezewq6rtPNtj",
		"0000000000000000000000000000000000000000000000000000000000000005",
		250000000000,
		"0000000000000000000000000000000000000000000000000000000000000007",
		30000000000000,
	)
	metaType, _ := strconv.Atoi(reqAction[0])
	contentStr := reqAction[1]
	shardID := byte(1)
	bc := &BlockChain{}
	wdAcceptedContent1 := metadata.PDEWithdrawalAcceptedContent{
		WithdrawalTokenIDStr: "0000000000000000000000000000000000000000000000000000000000000005",
		WithdrawerAddressStr: "12S2jM1TBbX2V5TBTvpJkJmsdaYxbCspGNedQkvJpYcbnV4gad7FDEbzY9P3zbpZRJTsGD5vxJRia3UiiUwMUbXbjfgezewq6rtPNtj",
		DeductingPoolValue:   250000000000,
		DeductingShares:      250000000000,
		PairToken1IDStr:      "0000000000000000000000000000000000000000000000000000000000000005",
		PairToken2IDStr:      "0000000000000000000000000000000000000000000000000000000000000007",
		TxReqID:              common.Hash{},
		ShardID:              shardID,
	}
	wdAcceptedContent2 := metadata.PDEWithdrawalAcceptedContent{
		WithdrawalTokenIDStr: "0000000000000000000000000000000000000000000000000000000000000007",
		WithdrawerAddressStr: "12S2jM1TBbX2V5TBTvpJkJmsdaYxbCspGNedQkvJpYcbnV4gad7FDEbzY9P3zbpZRJTsGD5vxJRia3UiiUwMUbXbjfgezewq6rtPNtj",
		DeductingPoolValue:   30000000000000,
		DeductingShares:      30000000000000,
		PairToken1IDStr:      "0000000000000000000000000000000000000000000000000000000000000005",
		PairToken2IDStr:      "0000000000000000000000000000000000000000000000000000000000000007",
		TxReqID:              common.Hash{},
		ShardID:              shardID,
	}
	wdAcceptedContent1Bytes, err := json.Marshal(wdAcceptedContent1)
	wdAcceptedContent2Bytes, err := json.Marshal(wdAcceptedContent2)
	newInsts, err := bc.buildInstructionsForPDEWithdrawal(
		contentStr,
		shardID,
		metaType,
		suite.currentPDEState,
		beaconHeight-1,
	)
	suite.Equal(err, nil)
	suite.Equal(len(newInsts), 2)
	suite.Equal(len(newInsts[0]), 4)
	suite.Equal(newInsts[0][0], strconv.Itoa(metaType))
	suite.Equal(newInsts[0][1], strconv.Itoa(int(shardID)))
	suite.Equal(newInsts[0][2], "accepted")
	suite.Equal(newInsts[0][3], string(wdAcceptedContent1Bytes))
	suite.Equal(newInsts[1][3], string(wdAcceptedContent2Bytes))

	remainingTk1PoolVal := suite.currentPDEState.PDEPoolPairs[pairKey].Token1PoolValue
	remainingTk2PoolVal := suite.currentPDEState.PDEPoolPairs[pairKey].Token2PoolValue
	remainingTk1Shares := suite.currentPDEState.PDEShares[shareKey1]
	remainingTk2Shares := suite.currentPDEState.PDEShares[shareKey2]
	suite.Equal(remainingTk1PoolVal, uint64(250000000000))
	suite.Equal(remainingTk2PoolVal, uint64(30000000000000))
	suite.Equal(remainingTk1Shares, uint64(250000000000))
	suite.Equal(remainingTk2Shares, uint64(30000000000000))
}

func (suite *PDEProducerSuite) TestWithdrawOnToken1OfExistedPair() {
	fmt.Println("Running testcase: TestWithdrawOnToken1OfExistedPair")
	beaconHeight := uint64(1001)
	pair := rawdbv2.PDEPoolForPair{
		Token1IDStr:     "0000000000000000000000000000000000000000000000000000000000000005",
		Token1PoolValue: 500000000000,
		Token2IDStr:     "0000000000000000000000000000000000000000000000000000000000000007",
		Token2PoolValue: 60000000000000,
	}
	pairKey := string(rawdbv2.BuildPDEPoolForPairKey(beaconHeight-1, pair.Token1IDStr, pair.Token2IDStr))
	suite.currentPDEState.PDEPoolPairs = map[string]*rawdbv2.PDEPoolForPair{
		pairKey: &pair,
	}

	shareKey1 := string(rawdbv2.BuildPDESharesKey(
		beaconHeight-1,
		"0000000000000000000000000000000000000000000000000000000000000005",
		"0000000000000000000000000000000000000000000000000000000000000007",
		"0000000000000000000000000000000000000000000000000000000000000005",
		"12S2jM1TBbX2V5TBTvpJkJmsdaYxbCspGNedQkvJpYcbnV4gad7FDEbzY9P3zbpZRJTsGD5vxJRia3UiiUwMUbXbjfgezewq6rtPNtj",
	))
	shareKey2 := string(rawdbv2.BuildPDESharesKey(
		beaconHeight-1,
		"0000000000000000000000000000000000000000000000000000000000000005",
		"0000000000000000000000000000000000000000000000000000000000000007",
		"0000000000000000000000000000000000000000000000000000000000000007",
		"12S2jM1TBbX2V5TBTvpJkJmsdaYxbCspGNedQkvJpYcbnV4gad7FDEbzY9P3zbpZRJTsGD5vxJRia3UiiUwMUbXbjfgezewq6rtPNtj",
	))
	suite.currentPDEState.PDEShares = map[string]uint64{
		shareKey1: 500000000000,
		shareKey2: 60000000000000,
	}

	reqAction := buildPDEWithdrawReqAction(
		"12S2jM1TBbX2V5TBTvpJkJmsdaYxbCspGNedQkvJpYcbnV4gad7FDEbzY9P3zbpZRJTsGD5vxJRia3UiiUwMUbXbjfgezewq6rtPNtj",
		"0000000000000000000000000000000000000000000000000000000000000005",
		250000000000,
		"0000000000000000000000000000000000000000000000000000000000000007",
		0,
	)
	metaType, _ := strconv.Atoi(reqAction[0])
	contentStr := reqAction[1]
	shardID := byte(1)
	bc := &BlockChain{}
	wdAcceptedContent1 := metadata.PDEWithdrawalAcceptedContent{
		WithdrawalTokenIDStr: "0000000000000000000000000000000000000000000000000000000000000005",
		WithdrawerAddressStr: "12S2jM1TBbX2V5TBTvpJkJmsdaYxbCspGNedQkvJpYcbnV4gad7FDEbzY9P3zbpZRJTsGD5vxJRia3UiiUwMUbXbjfgezewq6rtPNtj",
		DeductingPoolValue:   250000000000,
		DeductingShares:      250000000000,
		PairToken1IDStr:      "0000000000000000000000000000000000000000000000000000000000000005",
		PairToken2IDStr:      "0000000000000000000000000000000000000000000000000000000000000007",
		TxReqID:              common.Hash{},
		ShardID:              shardID,
	}
	wdAcceptedContent1Bytes, err := json.Marshal(wdAcceptedContent1)
	newInsts, err := bc.buildInstructionsForPDEWithdrawal(
		contentStr,
		shardID,
		metaType,
		suite.currentPDEState,
		beaconHeight-1,
	)
	suite.Equal(err, nil)
	suite.Equal(len(newInsts), 1)
	suite.Equal(len(newInsts[0]), 4)
	suite.Equal(newInsts[0][0], strconv.Itoa(metaType))
	suite.Equal(newInsts[0][1], strconv.Itoa(int(shardID)))
	suite.Equal(newInsts[0][2], "accepted")
	suite.Equal(newInsts[0][3], string(wdAcceptedContent1Bytes))

	remainingTk1PoolVal := suite.currentPDEState.PDEPoolPairs[pairKey].Token1PoolValue
	remainingTk2PoolVal := suite.currentPDEState.PDEPoolPairs[pairKey].Token2PoolValue
	remainingTk1Shares := suite.currentPDEState.PDEShares[shareKey1]
	remainingTk2Shares := suite.currentPDEState.PDEShares[shareKey2]
	suite.Equal(remainingTk1PoolVal, uint64(250000000000))
	suite.Equal(remainingTk2PoolVal, uint64(60000000000000))
	suite.Equal(remainingTk1Shares, uint64(250000000000))
	suite.Equal(remainingTk2Shares, uint64(60000000000000))
}

func (suite *PDEProducerSuite) TestWithdrawOnUnexistedPair() {
	fmt.Println("Running testcase: TestWithdrawOnUnexistedPair")
	beaconHeight := uint64(1001)
	pair := rawdbv2.PDEPoolForPair{
		Token1IDStr:     "0000000000000000000000000000000000000000000000000000000000000005",
		Token1PoolValue: 500000000000,
		Token2IDStr:     "0000000000000000000000000000000000000000000000000000000000000007",
		Token2PoolValue: 60000000000000,
	}
	pairKey := string(rawdbv2.BuildPDEPoolForPairKey(beaconHeight-1, pair.Token1IDStr, pair.Token2IDStr))
	suite.currentPDEState.PDEPoolPairs = map[string]*rawdbv2.PDEPoolForPair{
		pairKey: &pair,
	}

	shareKey1 := string(rawdbv2.BuildPDESharesKey(
		beaconHeight-1,
		"0000000000000000000000000000000000000000000000000000000000000005",
		"0000000000000000000000000000000000000000000000000000000000000007",
		"0000000000000000000000000000000000000000000000000000000000000005",
		"12S2jM1TBbX2V5TBTvpJkJmsdaYxbCspGNedQkvJpYcbnV4gad7FDEbzY9P3zbpZRJTsGD5vxJRia3UiiUwMUbXbjfgezewq6rtPNtj",
	))
	shareKey2 := string(rawdbv2.BuildPDESharesKey(
		beaconHeight-1,
		"0000000000000000000000000000000000000000000000000000000000000005",
		"0000000000000000000000000000000000000000000000000000000000000007",
		"0000000000000000000000000000000000000000000000000000000000000007",
		"12S2jM1TBbX2V5TBTvpJkJmsdaYxbCspGNedQkvJpYcbnV4gad7FDEbzY9P3zbpZRJTsGD5vxJRia3UiiUwMUbXbjfgezewq6rtPNtj",
	))
	suite.currentPDEState.PDEShares = map[string]uint64{
		shareKey1: 500000000000,
		shareKey2: 60000000000000,
	}

	reqAction := buildPDEWithdrawReqAction(
		"12S2jM1TBbX2V5TBTvpJkJmsdaYxbCspGNedQkvJpYcbnV4gad7FDEbzY9P3zbpZRJTsGD5vxJRia3UiiUwMUbXbjfgezewq6rtPNtj",
		"0000000000000000000000000000000000000000000000000000000000000005",
		250000000000,
		"0000000000000000000000000000000000000000000000000000000000000008",
		0,
	)
	metaType, _ := strconv.Atoi(reqAction[0])
	contentStr := reqAction[1]
	shardID := byte(1)
	bc := &BlockChain{}
	newInsts, err := bc.buildInstructionsForPDEWithdrawal(
		contentStr,
		shardID,
		metaType,
		suite.currentPDEState,
		beaconHeight-1,
	)
	suite.Equal(err, nil)
	suite.Equal(len(newInsts), 0)

	remainingTk1PoolVal := suite.currentPDEState.PDEPoolPairs[pairKey].Token1PoolValue
	remainingTk2PoolVal := suite.currentPDEState.PDEPoolPairs[pairKey].Token2PoolValue
	remainingTk1Shares := suite.currentPDEState.PDEShares[shareKey1]
	remainingTk2Shares := suite.currentPDEState.PDEShares[shareKey2]
	suite.Equal(remainingTk1PoolVal, uint64(500000000000))
	suite.Equal(remainingTk2PoolVal, uint64(60000000000000))
	suite.Equal(remainingTk1Shares, uint64(500000000000))
	suite.Equal(remainingTk2Shares, uint64(60000000000000))
}

func (suite *PDEProducerSuite) TestWithdrawExceededSharesOnToken2OfExistedPair() {
	fmt.Println("Running testcase: TestWithdrawExceededSharesOnToken2OfExistedPair")
	beaconHeight := uint64(1001)
	pair := rawdbv2.PDEPoolForPair{
		Token1IDStr:     "0000000000000000000000000000000000000000000000000000000000000005",
		Token1PoolValue: 500000000000,
		Token2IDStr:     "0000000000000000000000000000000000000000000000000000000000000007",
		Token2PoolValue: 60000000000000,
	}
	pairKey := string(rawdbv2.BuildPDEPoolForPairKey(beaconHeight-1, pair.Token1IDStr, pair.Token2IDStr))
	suite.currentPDEState.PDEPoolPairs = map[string]*rawdbv2.PDEPoolForPair{
		pairKey: &pair,
	}

	shareKey1 := string(rawdbv2.BuildPDESharesKey(
		beaconHeight-1,
		"0000000000000000000000000000000000000000000000000000000000000005",
		"0000000000000000000000000000000000000000000000000000000000000007",
		"0000000000000000000000000000000000000000000000000000000000000005",
		"12S2jM1TBbX2V5TBTvpJkJmsdaYxbCspGNedQkvJpYcbnV4gad7FDEbzY9P3zbpZRJTsGD5vxJRia3UiiUwMUbXbjfgezewq6rtPNtj",
	))
	shareKey2 := string(rawdbv2.BuildPDESharesKey(
		beaconHeight-1,
		"0000000000000000000000000000000000000000000000000000000000000005",
		"0000000000000000000000000000000000000000000000000000000000000007",
		"0000000000000000000000000000000000000000000000000000000000000007",
		"12S2jM1TBbX2V5TBTvpJkJmsdaYxbCspGNedQkvJpYcbnV4gad7FDEbzY9P3zbpZRJTsGD5vxJRia3UiiUwMUbXbjfgezewq6rtPNtj",
	))
	shareKey3 := string(rawdbv2.BuildPDESharesKey(
		beaconHeight-1,
		"0000000000000000000000000000000000000000000000000000000000000005",
		"0000000000000000000000000000000000000000000000000000000000000007",
		"0000000000000000000000000000000000000000000000000000000000000007",
		"12S2jM1TBbX2V5TBTvpJkJmsdaYxbCspGNedQkvJpYcbnV4gad7FDEbzY9P3zbpZRJTsGD5vxJRia3UiiUwMUbXbjfgezewq6rtPNtj-new",
	))
	suite.currentPDEState.PDEShares = map[string]uint64{
		shareKey1: 500000000000,
		shareKey2: 60000000000000,
		shareKey3: 20000000000000,
	}

	reqAction := buildPDEWithdrawReqAction(
		"12S2jM1TBbX2V5TBTvpJkJmsdaYxbCspGNedQkvJpYcbnV4gad7FDEbzY9P3zbpZRJTsGD5vxJRia3UiiUwMUbXbjfgezewq6rtPNtj",
		"0000000000000000000000000000000000000000000000000000000000000005",
		0,
		"0000000000000000000000000000000000000000000000000000000000000007",
		60000000000000+1000,
	)
	metaType, _ := strconv.Atoi(reqAction[0])
	contentStr := reqAction[1]
	shardID := byte(1)
	bc := &BlockChain{}
	wdAcceptedContent2 := metadata.PDEWithdrawalAcceptedContent{
		WithdrawalTokenIDStr: "0000000000000000000000000000000000000000000000000000000000000007",
		WithdrawerAddressStr: "12S2jM1TBbX2V5TBTvpJkJmsdaYxbCspGNedQkvJpYcbnV4gad7FDEbzY9P3zbpZRJTsGD5vxJRia3UiiUwMUbXbjfgezewq6rtPNtj",
		DeductingPoolValue:   45000000000000,
		DeductingShares:      60000000000000,
		PairToken1IDStr:      "0000000000000000000000000000000000000000000000000000000000000005",
		PairToken2IDStr:      "0000000000000000000000000000000000000000000000000000000000000007",
		TxReqID:              common.Hash{},
		ShardID:              shardID,
	}
	wdAcceptedContent2Bytes, err := json.Marshal(wdAcceptedContent2)
	newInsts, err := bc.buildInstructionsForPDEWithdrawal(
		contentStr,
		shardID,
		metaType,
		suite.currentPDEState,
		beaconHeight-1,
	)
	suite.Equal(err, nil)
	suite.Equal(len(newInsts), 1)
	suite.Equal(len(newInsts[0]), 4)
	suite.Equal(newInsts[0][0], strconv.Itoa(metaType))
	suite.Equal(newInsts[0][1], strconv.Itoa(int(shardID)))
	suite.Equal(newInsts[0][2], "accepted")
	suite.Equal(newInsts[0][3], string(wdAcceptedContent2Bytes))

	remainingTk1PoolVal := suite.currentPDEState.PDEPoolPairs[pairKey].Token1PoolValue
	remainingTk2PoolVal := suite.currentPDEState.PDEPoolPairs[pairKey].Token2PoolValue
	remainingTk1Shares := suite.currentPDEState.PDEShares[shareKey1]
	remainingTk2Shares := suite.currentPDEState.PDEShares[shareKey2]
	remainingTk3Shares := suite.currentPDEState.PDEShares[shareKey3]
	suite.Equal(remainingTk1PoolVal, uint64(500000000000))
	suite.Equal(remainingTk2PoolVal, uint64(15000000000000))
	suite.Equal(remainingTk1Shares, uint64(500000000000))
	suite.Equal(remainingTk2Shares, uint64(0))
	suite.Equal(remainingTk3Shares, uint64(20000000000000))
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestPDEProducerSuite(t *testing.T) {
	fmt.Println("Initialized...")
	suite.Run(t, new(PDEProducerSuite))
}
