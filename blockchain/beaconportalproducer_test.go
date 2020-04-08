package blockchain

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
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
		CustodianPoolState:      map[string]*statedb.CustodianState{},
		ExchangeRatesRequests:   map[string]*metadata.ExchangeRatesRequestStatus{},
		FinalExchangeRatesState: map[string]*statedb.FinalExchangeRatesState{},
		WaitingPortingRequests:  map[string]*statedb.WaitingPortingRequest{},
		WaitingRedeemRequests:   map[string]*statedb.WaitingRedeemRequest{},
	}
}

/************************ Porting request test ************************/
type PortingRequestExcepted struct {
	Metadata    string
	ChainStatus string
	Custodian1  []string
	Custodian2  []string
}

type PortingRequestTestCase struct {
	TestCaseName string
	Input        func() metadata.PortalUserRegisterAction
	Output       func() PortingRequestExcepted
}

func (suite *PortalProducerSuite) SetupPortingRequest(beaconHeight uint64) {
	rates := make(map[string]statedb.FinalExchangeRatesDetail)
	rates["b832e5d3b1f01a4f0623f7fe91d6673461e1f5d37d91fe78c5c2e6183ff39696"] = statedb.FinalExchangeRatesDetail{
		Amount: 8000000000,
	}
	rates["b2655152784e8639fa19521a7035f331eea1f1e911b2f3200a507ebb4554387b"] = statedb.FinalExchangeRatesDetail{
		Amount: 20000000,
	}
	rates["0000000000000000000000000000000000000000000000000000000000000004"] = statedb.FinalExchangeRatesDetail{
		Amount: 500000,
	}

	exchangeRates := make(map[string]*statedb.FinalExchangeRatesState)
	exchangeRatesKey := statedb.GeneratePortalFinalExchangeRatesStateObjectKey(beaconHeight)
	exchangeRates[exchangeRatesKey.String()] = statedb.NewFinalExchangeRatesStateWithValue(rates)

	suite.currentPortalState.FinalExchangeRatesState = exchangeRates
}

func (suite *PortalProducerSuite) SetupOneCustodian(beaconHeight uint64) {
	remoteAddresses := make([]statedb.RemoteAddress, 0)
	remoteAddresses = append(
		remoteAddresses,
		*statedb.NewRemoteAddressWithValue("b2655152784e8639fa19521a7035f331eea1f1e911b2f3200a507ebb4554387b", "bnb136ns6lfw4zs5hg4n85vdthaad7hq5m4gtkgf234"),
	)

	custodianKey := statedb.GenerateCustodianStateObjectKey(beaconHeight, "12RuEdPjq4yxivzm8xPxRVHmkL74t4eAdUKPdKKhMEnpxPH3k8GEyULbwq4hjwHWmHQr7MmGBJsMpdCHsYAqNE18jipWQwciBf9yqvQ")
	newCustodian := statedb.NewCustodianStateWithValue(
		"12RuEdPjq4yxivzm8xPxRVHmkL74t4eAdUKPdKKhMEnpxPH3k8GEyULbwq4hjwHWmHQr7MmGBJsMpdCHsYAqNE18jipWQwciBf9yqvQ",
		100000,
		100000,
		nil,
		nil,
		remoteAddresses,
		0,
	)

	custodian := make(map[string]*statedb.CustodianState)
	custodian[custodianKey.String()] = newCustodian
	suite.currentPortalState.CustodianPoolState = custodian
}

func (suite *PortalProducerSuite) SetupMultipleCustodian(beaconHeight uint64) {
	remoteAddresses := make([]statedb.RemoteAddress, 0)
	remoteAddresses = append(
		remoteAddresses,
		*statedb.NewRemoteAddressWithValue("b2655152784e8639fa19521a7035f331eea1f1e911b2f3200a507ebb4554387b", "bnb136ns6lfw4zs5hg4n85vdthaad7hq5m4gtkgf234"),
	)

	custodianKey := statedb.GenerateCustodianStateObjectKey(beaconHeight, "12RuEdPjq4yxivzm8xPxRVHmkL74t4eAdUKPdKKhMEnpxPH3k8GEyULbwq4hjwHWmHQr7MmGBJsMpdCHsYAqNE18jipWQwciBf9yqvQ")
	newCustodian := statedb.NewCustodianStateWithValue(
		"12RuEdPjq4yxivzm8xPxRVHmkL74t4eAdUKPdKKhMEnpxPH3k8GEyULbwq4hjwHWmHQr7MmGBJsMpdCHsYAqNE18jipWQwciBf9yqvQ",
		100000,
		100000,
		nil,
		nil,
		remoteAddresses,
		0,
	)

	custodianKey2 := statedb.GenerateCustodianStateObjectKey(beaconHeight, "12Rwz4HXkVABgRnSb5Gfu1FaJ7auo3fLNXVGFhxx1dSytxHpWhbkimT1Mv5Z2oCMsssSXTVsapY8QGBZd2J4mPiCTzJAtMyCzb4dDcy")
	newCustodian2 := statedb.NewCustodianStateWithValue(
		"12Rwz4HXkVABgRnSb5Gfu1FaJ7auo3fLNXVGFhxx1dSytxHpWhbkimT1Mv5Z2oCMsssSXTVsapY8QGBZd2J4mPiCTzJAtMyCzb4dDcy",
		90000,
		90000,
		nil,
		nil,
		remoteAddresses,
		0,
	)

	custodian := make(map[string]*statedb.CustodianState)
	custodian[custodianKey.String()] = newCustodian
	custodian[custodianKey2.String()] = newCustodian2
	suite.currentPortalState.CustodianPoolState = custodian
}

func (suite *PortalProducerSuite) SetupMockBlockChain(trieMock *mocks.Trie) *BlockChain {
	root := common.Hash{}
	wrapperDBMock := new(mocks.DatabaseAccessWarper)
	wrapperDBMock.On("OpenPrefixTrie", root).Return(
		trieMock,
		nil,
	)

	wrapperDBMock.On("CopyTrie", trieMock).Return(
		trieMock,
		nil,
	)

	root1 := common.Hash{}
	stateDb, _ := statedb.NewWithPrefixTrie(root1, wrapperDBMock)

	beaconBestState := &BeaconBestState{
		featureStateDB: stateDb,
	}

	bestState := &BestState{
		Beacon: beaconBestState,
	}

	blockChain := &BlockChain{
		BestState: bestState,
	}

	return blockChain
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

func (suite *PortalProducerSuite) TestBuildInstructionsForPortingRequest() {
	/*
	happyCases := []PortingRequestTestCase{
		{
			"happy_case_1",
			func() metadata.PortalUserRegisterAction {
				meta, _ := metadata.NewPortalUserRegister(
					"1",
					"12S5pBBRDf1GqfRHouvCV86sWaHzNfvakAWpVMvNnWu2k299xWCgQzLLc9wqPYUHfMYGDprPvQ794dbi6UU1hfRN4tPiU61txWWenhC", //100.000 prv
					"b2655152784e8639fa19521a7035f331eea1f1e911b2f3200a507ebb4554387b",
					1000,
					4,
					metadata.PortalUserRegisterMeta,
				)

				actionContent := metadata.PortalUserRegisterAction{
					Meta:    *meta,
					TxReqID: *meta.Hash(),
					ShardID: 1,
				}
				return actionContent
			},
			func() PortingRequestExcepted {
				return PortingRequestExcepted{
					Metadata:    strconv.Itoa(metadata.PortalUserRegisterMeta),
					ChainStatus: common.PortalPortingRequestAcceptedChainStatus,
					Custodian1: []string{
						"12RuEdPjq4yxivzm8xPxRVHmkL74t4eAdUKPdKKhMEnpxPH3k8GEyULbwq4hjwHWmHQr7MmGBJsMpdCHsYAqNE18jipWQwciBf9yqvQ", //address
						"40000", //free collateral
						"1000",  //hold pToken
						"60000", //lock prv amount
					},
				}
			},
		},
		{
			"happy_case_2",
			func() metadata.PortalUserRegisterAction {
				meta, _ := metadata.NewPortalUserRegister(
					"2",
					"12S5pBBRDf1GqfRHouvCV86sWaHzNfvakAWpVMvNnWu2k299xWCgQzLLc9wqPYUHfMYGDprPvQ794dbi6UU1hfRN4tPiU61txWWenhC", //100.000 prv
					"b2655152784e8639fa19521a7035f331eea1f1e911b2f3200a507ebb4554387b",
					100,
					4,
					metadata.PortalUserRegisterMeta,
				)

				actionContent := metadata.PortalUserRegisterAction{
					Meta:    *meta,
					TxReqID: *meta.Hash(),
					ShardID: 1,
				}
				return actionContent
			},
			func() PortingRequestExcepted {
				return PortingRequestExcepted{
					Metadata:    strconv.Itoa(metadata.PortalUserRegisterMeta),
					ChainStatus: common.PortalPortingRequestAcceptedChainStatus,
					Custodian1: []string{
						"12RuEdPjq4yxivzm8xPxRVHmkL74t4eAdUKPdKKhMEnpxPH3k8GEyULbwq4hjwHWmHQr7MmGBJsMpdCHsYAqNE18jipWQwciBf9yqvQ", //address
						"34000", //free collateral
						"1100",  //hold pToken
						"66000", //lock prv amount
					},
				}
			},
		},
	}

	//reset
	suite.SetupTest()
	suite.SetupPortingRequest(1)
	suite.SetupOneCustodian(1)
	suite.verifyPortingRequest(happyCases)
*/
	pickMultipleCustodianCases := []PortingRequestTestCase{
		{
			"pick_multiple_custodian_case_1",
			func() metadata.PortalUserRegisterAction {
				meta, _ := metadata.NewPortalUserRegister(
					"1",
					"12S5pBBRDf1GqfRHouvCV86sWaHzNfvakAWpVMvNnWu2k299xWCgQzLLc9wqPYUHfMYGDprPvQ794dbi6UU1hfRN4tPiU61txWWenhC", //100.000 prv
					"b2655152784e8639fa19521a7035f331eea1f1e911b2f3200a507ebb4554387b",
					2000,
					8,
					metadata.PortalUserRegisterMeta,
				)

				actionContent := metadata.PortalUserRegisterAction{
					Meta:    *meta,
					TxReqID: *meta.Hash(),
					ShardID: 1,
				}
				return actionContent
			},
			func() PortingRequestExcepted {
				return PortingRequestExcepted{
					Metadata:    strconv.Itoa(metadata.PortalUserRegisterMeta),
					ChainStatus: common.PortalPortingRequestAcceptedChainStatus,
					Custodian1: []string{
						"12RuEdPjq4yxivzm8xPxRVHmkL74t4eAdUKPdKKhMEnpxPH3k8GEyULbwq4hjwHWmHQr7MmGBJsMpdCHsYAqNE18jipWQwciBf9yqvQ", //address
						"0", //free collateral
						"1667",  //hold pToken
						"100000", //lock prv amount
					},
					Custodian2: []string{
						"12Rwz4HXkVABgRnSb5Gfu1FaJ7auo3fLNXVGFhxx1dSytxHpWhbkimT1Mv5Z2oCMsssSXTVsapY8QGBZd2J4mPiCTzJAtMyCzb4dDcy", //address
						"70000", //free collateral
						"333",  //hold pToken
						"20000", //lock prv amount
					},
				}
			},
		},
	    {
			"pick_multiple_custodian_case_2",
			func() metadata.PortalUserRegisterAction {
				meta, _ := metadata.NewPortalUserRegister(
					"2",
					"12S5pBBRDf1GqfRHouvCV86sWaHzNfvakAWpVMvNnWu2k299xWCgQzLLc9wqPYUHfMYGDprPvQ794dbi6UU1hfRN4tPiU61txWWenhC", //100.000 prv
					"b2655152784e8639fa19521a7035f331eea1f1e911b2f3200a507ebb4554387b",
					1000,
					4,
					metadata.PortalUserRegisterMeta,
				)

				actionContent := metadata.PortalUserRegisterAction{
					Meta:    *meta,
					TxReqID: *meta.Hash(),
					ShardID: 1,
				}
				return actionContent
			},
			func() PortingRequestExcepted {
				return PortingRequestExcepted{
					Metadata:    strconv.Itoa(metadata.PortalUserRegisterMeta),
					ChainStatus: common.PortalPortingRequestAcceptedChainStatus,
					Custodian1: []string{
						"12RuEdPjq4yxivzm8xPxRVHmkL74t4eAdUKPdKKhMEnpxPH3k8GEyULbwq4hjwHWmHQr7MmGBJsMpdCHsYAqNE18jipWQwciBf9yqvQ", //address
						"0", //free collateral
						"1667",  //hold pToken
						"100000", //lock prv amount
					},
					Custodian2: []string{
						"12Rwz4HXkVABgRnSb5Gfu1FaJ7auo3fLNXVGFhxx1dSytxHpWhbkimT1Mv5Z2oCMsssSXTVsapY8QGBZd2J4mPiCTzJAtMyCzb4dDcy", //address
						"10000", //free collateral
						"1333",  //hold pToken
						"80000", //lock prv amount
					},
				}
			},
		},
	}

	//reset
	suite.SetupTest()
	suite.SetupPortingRequest(1)
	suite.SetupMultipleCustodian(1)
	suite.verifyPortingRequest(pickMultipleCustodianCases)
}

func (suite *PortalProducerSuite) verifyPortingRequest(testCases []PortingRequestTestCase) {
	trieMock := new(mocks.Trie)
	beaconHeight := uint64(1)

	for _, testCase := range testCases {
		actionContentBytes, _ := json.Marshal(testCase.Input())
		actionContentBase64Str := base64.StdEncoding.EncodeToString(actionContentBytes)

		key := statedb.GeneratePortalStatusObjectKey(statedb.PortalPortingRequestStatusPrefix(), []byte(testCase.Input().Meta.UniqueRegisterId))
		trieMock.On("TryGet", key[:]).Return(nil, nil)

		blockChain := suite.SetupMockBlockChain(trieMock)

		value, err := blockChain.buildInstructionsForPortingRequest(
			actionContentBase64Str,
			testCase.Input().ShardID,
			testCase.Input().Meta.Type,
			suite.currentPortalState,
			beaconHeight,
		)

		fmt.Printf("Testcase %v: instruction %#v", testCase.TestCaseName, value)
		fmt.Println()

		assert.Equal(suite.T(), err, nil)

		if len(testCase.Output().Metadata) > 0 {
			assert.Equal(suite.T(), testCase.Output().Metadata, value[0][0])
		}

		assert.Equal(suite.T(), strconv.Itoa(1), value[0][1])

		if len(testCase.Output().ChainStatus) > 0 {
			assert.Equal(suite.T(), testCase.Output().ChainStatus, value[0][2])
		}

		assert.NotNil(suite.T(), value[0][3])

		//test current portal state
		var portingRequestContent metadata.PortalPortingRequestContent
		json.Unmarshal([]byte(value[0][3]), &portingRequestContent)

		prettyJSON, _ := json.MarshalIndent(portingRequestContent, "", "  ")
		fmt.Printf("Porting request result: %s\n", string(prettyJSON))

		for _, itemCustodian := range portingRequestContent.Custodian {
			//update custodian state
			custodianKey := statedb.GenerateCustodianStateObjectKey(beaconHeight, itemCustodian.IncAddress)
			custodian := suite.currentPortalState.CustodianPoolState[custodianKey.String()]

			if testCase.Output().Custodian1 != nil && itemCustodian.IncAddress == testCase.Output().Custodian1[0] {
				holdPublicToken := custodian.GetHoldingPublicTokens()
				lockedAmountCollateral := custodian.GetLockedAmountCollateral()
				freeCollateral := custodian.GetFreeCollateral()

				fmt.Println("custodian 1")
				fmt.Println(testCase.Output().Custodian1)
				i1, _ := strconv.ParseUint(testCase.Output().Custodian1[1], 10, 64)
				i2, _ := strconv.ParseUint(testCase.Output().Custodian1[2], 10, 64)
				i3, _ := strconv.ParseUint(testCase.Output().Custodian1[3], 10, 64)

				if i1 != freeCollateral {    //free collateral
					suite.T().Errorf("free collateral is not equal, %v != %v", i1, freeCollateral)
				}

				if i2 != holdPublicToken["b2655152784e8639fa19521a7035f331eea1f1e911b2f3200a507ebb4554387b"] {
					suite.T().Errorf("hold public token is not equal, %v != %v", i2, holdPublicToken["b2655152784e8639fa19521a7035f331eea1f1e911b2f3200a507ebb4554387b"])
				}

				if i3 != lockedAmountCollateral["b2655152784e8639fa19521a7035f331eea1f1e911b2f3200a507ebb4554387b"] {
					suite.T().Errorf("lock amount collateral is not equal, %v != %v", i3, lockedAmountCollateral["b2655152784e8639fa19521a7035f331eea1f1e911b2f3200a507ebb4554387b"])
				}
			}

			if testCase.Output().Custodian2 != nil && itemCustodian.IncAddress == testCase.Output().Custodian2[0] {
				holdPublicToken := custodian.GetHoldingPublicTokens()
				lockedAmountCollateral := custodian.GetLockedAmountCollateral()
				freeCollateral := custodian.GetFreeCollateral()

				fmt.Println("custodian 2")
				fmt.Println(testCase.Output().Custodian2)
				i1, _ := strconv.ParseUint(testCase.Output().Custodian2[1], 10, 64)
				i2, _ := strconv.ParseUint(testCase.Output().Custodian2[2], 10, 64)
				i3, _ := strconv.ParseUint(testCase.Output().Custodian2[3], 10, 64)

				if i1 != freeCollateral {    //free collateral
					suite.T().Errorf("free collateral is not equal, %v != %v", i1, freeCollateral)
				}

				if i2 != holdPublicToken["b2655152784e8639fa19521a7035f331eea1f1e911b2f3200a507ebb4554387b"] {
					suite.T().Errorf("hold public token is not equal, %v != %v", i2, holdPublicToken["b2655152784e8639fa19521a7035f331eea1f1e911b2f3200a507ebb4554387b"])
				}

				if i3 != lockedAmountCollateral["b2655152784e8639fa19521a7035f331eea1f1e911b2f3200a507ebb4554387b"] {
					suite.T().Errorf("lock amount collateral is not equal, %v != %v", i3, lockedAmountCollateral["b2655152784e8639fa19521a7035f331eea1f1e911b2f3200a507ebb4554387b"])
				}
			}
		}
	}
}

/************************ Custodian deposit test ************************/

/************************ Run suite test ************************/
// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestPortalProducerSuite(t *testing.T) {
	suite.Run(t, new(PortalProducerSuite))
}
