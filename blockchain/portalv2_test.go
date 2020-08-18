package blockchain

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"math"
	"math/big"
	"os"
	"strconv"
	"testing"
	"time"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type PortalTestSuite struct {
	suite.Suite
	currentPortalStateForProducer CurrentPortalState
	currentPortalStateForProcess  CurrentPortalState
	sdb                           *statedb.StateDB
	portalParams                  PortalParams
}

func (s *PortalTestSuite) SetupTest() {
	dbPath, err := ioutil.TempDir(os.TempDir(), "portal_test_statedb_")
	if err != nil {
		panic(err)
	}
	diskBD, _ := incdb.Open("leveldb", dbPath)
	warperDBStatedbTest := statedb.NewDatabaseAccessWarper(diskBD)
	emptyRoot := common.HexToHash(common.HexEmptyRoot)
	stateDB, _ := statedb.NewWithPrefixTrie(emptyRoot, warperDBStatedbTest)

	s.sdb = stateDB
	s.currentPortalStateForProducer = CurrentPortalState{
		CustodianPoolState:         map[string]*statedb.CustodianState{},
		WaitingPortingRequests:     map[string]*statedb.WaitingPortingRequest{},
		WaitingRedeemRequests:      map[string]*statedb.RedeemRequest{},
		MatchedRedeemRequests:      map[string]*statedb.RedeemRequest{},
		FinalExchangeRatesState:    new(statedb.FinalExchangeRatesState),
		LiquidationPool:            map[string]*statedb.LiquidationPool{},
		LockedCollateralForRewards: new(statedb.LockedCollateralState),
		ExchangeRatesRequests:      map[string]*metadata.ExchangeRatesRequestStatus{},
	}
	s.currentPortalStateForProcess = CurrentPortalState{
		CustodianPoolState:         map[string]*statedb.CustodianState{},
		WaitingPortingRequests:     map[string]*statedb.WaitingPortingRequest{},
		WaitingRedeemRequests:      map[string]*statedb.RedeemRequest{},
		MatchedRedeemRequests:      map[string]*statedb.RedeemRequest{},
		FinalExchangeRatesState:    new(statedb.FinalExchangeRatesState),
		LiquidationPool:            map[string]*statedb.LiquidationPool{},
		LockedCollateralForRewards: new(statedb.LockedCollateralState),
		ExchangeRatesRequests:      map[string]*metadata.ExchangeRatesRequestStatus{},
	}
	s.portalParams = PortalParams{
		TimeOutCustodianReturnPubToken:       1 * time.Hour,
		TimeOutWaitingPortingRequest:         1 * time.Hour,
		TimeOutWaitingRedeemRequest:          10 * time.Minute,
		MaxPercentLiquidatedCollateralAmount: 120,
		MaxPercentCustodianRewards:           10,
		MinPercentCustodianRewards:           1,
		MinLockCollateralAmountInEpoch:       5000 * 1e9, // 5000 prv
		MinPercentLockedCollateral:           200,
		TP120:                                120,
		TP130:                                130,
		MinPercentPortingFee:                 0.01,
		MinPercentRedeemFee:                  0.01,
	}
}

/*
 Utility functions
*/

func exchangeRates(amount uint64, tokenIDFrom string, tokenIDTo string, finalExchangeRate *statedb.FinalExchangeRatesState) uint64 {
	rateFrom := finalExchangeRate.Rates()[tokenIDFrom].Amount
	rateTo := finalExchangeRate.Rates()[tokenIDTo].Amount
	tmp := new(big.Int).Mul(new(big.Int).SetUint64(amount), new(big.Int).SetUint64(rateFrom))
	return new(big.Int).Div(tmp, new(big.Int).SetUint64(rateTo)).Uint64()
}

func getLockedCollateralAmount(
	portingAmount uint64, tokenID string, finalExchangeRate *statedb.FinalExchangeRatesState, percent uint64) uint64 {

	amount := new(big.Int).Mul(new(big.Int).SetUint64(portingAmount), new(big.Int).SetUint64(percent))
	amount = amount.Div(amount, new(big.Int).SetUint64(100))
	return exchangeRates(amount.Uint64(), tokenID, common.PRVIDStr, finalExchangeRate)
}

func getMinFee(amount uint64, tokenID string, finalExchangeRate *statedb.FinalExchangeRatesState, percent float64) uint64{
	amountInPRV := exchangeRates(amount, tokenID, common.PRVIDStr, finalExchangeRate)
	fee := float64(amountInPRV) * percent / float64(100)
	return uint64(math.Round(fee))
}

func producerPortalInstructions(
	blockchain *BlockChain,
	beaconHeight uint64,
	insts [][]string,
	portalStateDB *statedb.StateDB,
	currentPortalState *CurrentPortalState,
	portalParams PortalParams,
	shardID byte,
	newMatchedRedeemReqIDs []string,
) ([][]string, error) {
	var err error
	var newInst [][]string
	var newInsts [][]string
	for _, inst := range insts {
		switch inst[0] {
		//exchange rates
		case strconv.Itoa(metadata.PortalExchangeRatesMeta):
			newInst, err = blockchain.buildInstructionsForExchangeRates(
				inst[1], shardID, metadata.PortalExchangeRatesMeta, currentPortalState, beaconHeight, portalParams)

		// custodians deposit collateral
		case strconv.Itoa(metadata.PortalCustodianDepositMeta):
			newInst, err = blockchain.buildInstructionsForCustodianDeposit(
				inst[1], shardID, metadata.PortalCustodianDepositMeta, currentPortalState, beaconHeight, portalParams)
		// porting request
		case strconv.Itoa(metadata.PortalUserRegisterMeta):
			newInst, err = blockchain.buildInstructionsForPortingRequest(
				portalStateDB, inst[1], shardID, metadata.PortalUserRegisterMeta, currentPortalState, beaconHeight, portalParams)
		// submit proof to request ptokens
		case strconv.Itoa(metadata.PortalUserRequestPTokenMeta):
			newInst, err = blockchain.buildInstructionsForReqPTokens(
				portalStateDB, inst[1], shardID, metadata.PortalUserRequestPTokenMeta, currentPortalState, beaconHeight, portalParams)

		// redeem request
		case strconv.Itoa(metadata.PortalRedeemRequestMeta):
			newInst, err = blockchain.buildInstructionsForRedeemRequest(
				portalStateDB, inst[1], shardID, metadata.PortalRedeemRequestMeta, currentPortalState, beaconHeight, portalParams)
		// custodian request matching waiting redeem requests
		case strconv.Itoa(metadata.PortalReqMatchingRedeemMeta):
			newInst, newMatchedRedeemReqIDs, err = blockchain.buildInstructionsForReqMatchingRedeem(
				portalStateDB, inst[1], shardID, metadata.PortalReqMatchingRedeemMeta, currentPortalState, beaconHeight, portalParams, newMatchedRedeemReqIDs)
		// submit proof to request unlock collateral
		case strconv.Itoa(metadata.PortalRequestUnlockCollateralMeta):
			newInst, err = blockchain.buildInstructionsForReqUnlockCollateral(
				portalStateDB, inst[1], shardID, metadata.PortalRequestUnlockCollateralMeta, currentPortalState, beaconHeight, portalParams)

		// custodian request withdraw reward
		case strconv.Itoa(metadata.PortalRequestWithdrawRewardMeta):
			newInst, err = blockchain.buildInstructionsForReqWithdrawPortalReward(
				inst[1], shardID, metadata.PortalRequestWithdrawRewardMeta, currentPortalState, beaconHeight)
		// custodian request withdraw collaterals
		case strconv.Itoa(metadata.PortalCustodianWithdrawRequestMeta):
			newInst, err = blockchain.buildInstructionsForCustodianWithdraw(
				inst[1], shardID, metadata.PortalCustodianWithdrawRequestMeta, currentPortalState, beaconHeight, portalParams)

		// custodian top-up collaterals for holding public tokens
		case strconv.Itoa(metadata.PortalLiquidationCustodianDepositMetaV2):
			newInst, err = blockchain.buildInstructionsForLiquidationCustodianDeposit(
				inst[1], shardID, metadata.PortalCustodianWithdrawRequestMeta, currentPortalState, beaconHeight, portalParams)
		// custodian top-up collaterals for waiting porting requests
		case strconv.Itoa(metadata.PortalTopUpWaitingPortingRequestMeta):
			newInst, err = blockchain.buildInstsForTopUpWaitingPorting(
				inst[1], shardID, metadata.PortalTopUpWaitingPortingRequestMeta, currentPortalState, beaconHeight, portalParams)
		// user redeem ptokens from liquidation pool to get PRV
		case strconv.Itoa(metadata.PortalRedeemLiquidateExchangeRatesMeta):
			newInst, err = blockchain.buildInstructionsForLiquidationRedeemPTokenExchangeRates(
				inst[1], shardID, metadata.PortalRedeemLiquidateExchangeRatesMeta, currentPortalState, beaconHeight, portalParams)

			/*
					//// portal reward
				case strconv.Itoa(metadata.PortalRewardMeta):
					newInst, err = blockchain.buildPortalRewardsInsts(portalStateDB, beaconHeight, inst, currentPortalState, portalParams)

					// liquidation custodian run away
				case strconv.Itoa(metadata.PortalLiquidateCustodianMeta):
					newInst, err = blockchain.processPortalLiquidateCustodian(portalStateDB, beaconHeight, inst, currentPortalState, portalParams)

					// expired waiting porting request
				case strconv.Itoa(metadata.PortalExpiredWaitingPortingReqMeta):
					newInst, err = blockchain.processPortalExpiredPortingRequest(portalStateDB, beaconHeight, inst, currentPortalState, portalParams)
					// total custodian reward instruction
				case strconv.Itoa(metadata.PortalTotalRewardCustodianMeta):
					newInst, err = blockchain.processPortalTotalCustodianReward(portalStateDB, beaconHeight, inst, currentPortalState, portalParams)

					//liquidation exchange rates
				case strconv.Itoa(metadata.PortalLiquidateTPExchangeRatesMeta):
					newInst, err = blockchain.processLiquidationTopPercentileExchangeRates(portalStateDB, beaconHeight, inst, currentPortalState, portalParams)

			*/
		}

		if err != nil {
			Logger.log.Error(err)
			return newInsts, err
		}

		newInsts = append(newInsts, newInst...)
	}

	return newInsts, nil
}

func processPortalInstructions(
	blockchain *BlockChain,
	beaconHeight uint64,
	insts [][]string,
	portalStateDB *statedb.StateDB,
	currentPortalState *CurrentPortalState,
	portalParams PortalParams,
	updatingInfoByTokenID map[common.Hash]UpdatingInfo,
) error {
	var err error
	for _, inst := range insts {
		if len(inst) < 4 {
			continue // Not error, just not Portal instruction
		}

		var err error
		switch inst[0] {
		//porting request
		case strconv.Itoa(metadata.PortalUserRegisterMeta):
			err = blockchain.processPortalUserRegister(portalStateDB, beaconHeight, inst, currentPortalState, portalParams)
		//exchange rates
		case strconv.Itoa(metadata.PortalExchangeRatesMeta):
			err = blockchain.processPortalExchangeRates(portalStateDB, beaconHeight, inst, currentPortalState, portalParams)
		//custodian withdraw
		case strconv.Itoa(metadata.PortalCustodianWithdrawRequestMeta):
			err = blockchain.processPortalCustodianWithdrawRequest(portalStateDB, beaconHeight, inst, currentPortalState, portalParams)
		//liquidation exchange rates
		case strconv.Itoa(metadata.PortalLiquidateTPExchangeRatesMeta):
			err = blockchain.processLiquidationTopPercentileExchangeRates(portalStateDB, beaconHeight, inst, currentPortalState, portalParams)
		//liquidation custodian deposit
		case strconv.Itoa(metadata.PortalLiquidationCustodianDepositMetaV2):
			err = blockchain.processPortalLiquidationCustodianDeposit(portalStateDB, beaconHeight, inst, currentPortalState, portalParams)
		//waiting porting top up
		case strconv.Itoa(metadata.PortalTopUpWaitingPortingRequestMeta):
			err = blockchain.processPortalTopUpWaitingPorting(portalStateDB, beaconHeight, inst, currentPortalState, portalParams)
		//liquidation user redeem
		case strconv.Itoa(metadata.PortalRedeemLiquidateExchangeRatesMeta):
			err = blockchain.processPortalRedeemLiquidateExchangeRates(portalStateDB, beaconHeight, inst, currentPortalState, portalParams, updatingInfoByTokenID)
		//custodian deposit
		case strconv.Itoa(metadata.PortalCustodianDepositMeta):
			err = blockchain.processPortalCustodianDeposit(portalStateDB, beaconHeight, inst, currentPortalState, portalParams)
		// request ptoken
		case strconv.Itoa(metadata.PortalUserRequestPTokenMeta):
			err = blockchain.processPortalUserReqPToken(portalStateDB, beaconHeight, inst, currentPortalState, portalParams, updatingInfoByTokenID)
		// redeem request
		case strconv.Itoa(metadata.PortalRedeemRequestMeta):
			err = blockchain.processPortalRedeemRequest(portalStateDB, beaconHeight, inst, currentPortalState, portalParams, updatingInfoByTokenID)
		// request unlock collateral
		case strconv.Itoa(metadata.PortalRequestUnlockCollateralMeta):
			err = blockchain.processPortalUnlockCollateral(portalStateDB, beaconHeight, inst, currentPortalState, portalParams)
		// liquidation custodian run away
		case strconv.Itoa(metadata.PortalLiquidateCustodianMeta):
			err = blockchain.processPortalLiquidateCustodian(portalStateDB, beaconHeight, inst, currentPortalState, portalParams)
		// portal reward
		case strconv.Itoa(metadata.PortalRewardMeta):
			err = blockchain.processPortalReward(portalStateDB, beaconHeight, inst, currentPortalState, portalParams)
		// request withdraw reward
		case strconv.Itoa(metadata.PortalRequestWithdrawRewardMeta):
			err = blockchain.processPortalWithdrawReward(portalStateDB, beaconHeight, inst, currentPortalState, portalParams)
		// expired waiting porting request
		case strconv.Itoa(metadata.PortalExpiredWaitingPortingReqMeta):
			err = blockchain.processPortalExpiredPortingRequest(portalStateDB, beaconHeight, inst, currentPortalState, portalParams)
		// total custodian reward instruction
		case strconv.Itoa(metadata.PortalTotalRewardCustodianMeta):
			err = blockchain.processPortalTotalCustodianReward(portalStateDB, beaconHeight, inst, currentPortalState, portalParams)
		// custodian request matching waiting redeem requests
		case strconv.Itoa(metadata.PortalReqMatchingRedeemMeta):
			err = blockchain.processPortalReqMatchingRedeem(portalStateDB, beaconHeight, inst, currentPortalState, portalParams)
		case strconv.Itoa(metadata.PortalPickMoreCustodianForRedeemMeta):
			err = blockchain.processPortalPickMoreCustodiansForTimeOutWaitingRedeemReq(portalStateDB, beaconHeight, inst, currentPortalState, portalParams)
		}

		if err != nil {
			Logger.log.Error(err)
			return nil
		}
	}

	//save final exchangeRates
	blockchain.pickExchangesRatesFinal(currentPortalState)

	// update info of bridge portal token
	for _, updatingInfo := range updatingInfoByTokenID {
		var updatingAmt uint64
		var updatingType string
		if updatingInfo.countUpAmt > updatingInfo.deductAmt {
			updatingAmt = updatingInfo.countUpAmt - updatingInfo.deductAmt
			updatingType = "+"
		}
		if updatingInfo.countUpAmt < updatingInfo.deductAmt {
			updatingAmt = updatingInfo.deductAmt - updatingInfo.countUpAmt
			updatingType = "-"
		}
		err := statedb.UpdateBridgeTokenInfo(
			portalStateDB,
			updatingInfo.tokenID,
			updatingInfo.externalTokenID,
			updatingInfo.isCentralized,
			updatingAmt,
			updatingType,
		)
		if err != nil {
			return err
		}
	}

	// store updated currentPortalState to leveldb with new beacon height
	err = storePortalStateToDB(portalStateDB, currentPortalState)
	if err != nil {
		Logger.log.Error(err)
	}

	return nil
}

func buildPortalRelayExchangeRateAction(
	incAddressStr string,
	rates []*metadata.ExchangeRateInfo,
	shardID byte,
) []string {
	data := metadata.PortalExchangeRates{
		MetadataBase: metadata.MetadataBase{
			Type: metadata.PortalExchangeRatesMeta,
		},
		SenderAddress: incAddressStr,
		Rates:         rates,
	}

	actionContent := metadata.PortalExchangeRatesAction{
		Meta:    data,
		TxReqID: common.Hash{},
		ShardID: shardID,
	}
	actionContentBytes, _ := json.Marshal(actionContent)
	actionContentBase64Str := base64.StdEncoding.EncodeToString(actionContentBytes)
	return []string{strconv.Itoa(metadata.PortalExchangeRatesMeta), actionContentBase64Str}
}

func buildPortalCustodianDepositAction(
	incAddressStr string,
	remoteAddress map[string]string,
	depositAmount uint64,
	shardID byte,
) []string {
	data := metadata.PortalCustodianDeposit{
		MetadataBase: metadata.MetadataBase{
			Type: metadata.PortalCustodianDepositMeta,
		},
		IncogAddressStr: incAddressStr,
		RemoteAddresses: remoteAddress,
		DepositedAmount: depositAmount,
	}

	actionContent := metadata.PortalCustodianDepositAction{
		Meta:    data,
		TxReqID: common.Hash{},
		ShardID: shardID,
	}
	actionContentBytes, _ := json.Marshal(actionContent)
	actionContentBase64Str := base64.StdEncoding.EncodeToString(actionContentBytes)
	return []string{strconv.Itoa(metadata.PortalCustodianDepositMeta), actionContentBase64Str}
}

func buildPortalUserRegisterAction(
	portingID string,
	incAddressStr string,
	pTokenID string,
	portingAmount uint64,
	portingFee uint64,
	shardID byte,
) []string {
	data := metadata.PortalUserRegister{
		MetadataBase: metadata.MetadataBase{
			Type: metadata.PortalUserRegisterMeta,
		},
		UniqueRegisterId: portingID,
		IncogAddressStr:  incAddressStr,
		PTokenId:         pTokenID,
		RegisterAmount:   portingAmount,
		PortingFee:       portingFee,
	}

	actionContent := metadata.PortalUserRegisterAction{
		Meta:    data,
		TxReqID: common.Hash{},
		ShardID: shardID,
	}
	actionContentBytes, _ := json.Marshal(actionContent)
	actionContentBase64Str := base64.StdEncoding.EncodeToString(actionContentBytes)
	return []string{strconv.Itoa(metadata.PortalUserRegisterMeta), actionContentBase64Str}
}

/*
	Feature 0: Relay exchange rate
*/
type TestCaseRelayExchangeRate struct {
	senderAddressStr string
	rates            []*metadata.ExchangeRateInfo
}

func buildPortalExchangeRateActionsFromTcs(tcs []TestCaseRelayExchangeRate, shardID byte) [][]string {
	insts := [][]string{}

	for _, tc := range tcs {
		inst := buildPortalRelayExchangeRateAction(tc.senderAddressStr, tc.rates, shardID)
		insts = append(insts, inst)
	}

	return insts
}

func (s *PortalTestSuite) TestRelayExchangeRate() {
	fmt.Println("Running TestRelayExchangeRate - beacon height 999 ...")
	bc := new(BlockChain)
	beaconHeight := uint64(999)
	shardID := byte(0)
	newMatchedRedeemReqIDs := []string{}
	updatingInfoByTokenID := map[common.Hash]UpdatingInfo{}

	// build test cases
	testcases := []TestCaseRelayExchangeRate{
		// valid
		{
			senderAddressStr: "feeder1",
			rates: []*metadata.ExchangeRateInfo{
				{
					PTokenID: common.PRVIDStr,
					Rate:     1000000,
				},
				{
					PTokenID: common.PortalBNBIDStr,
					Rate:     20000000,
				},
				{
					PTokenID: common.PortalBTCIDStr,
					Rate:     10000000000,
				},
			},
		},
	}

	// build actions from testcases
	insts := buildPortalExchangeRateActionsFromTcs(testcases, shardID)

	// producer instructions
	newInsts, err := producerPortalInstructions(
		bc, beaconHeight, insts, s.sdb, &s.currentPortalStateForProducer, s.portalParams, shardID, newMatchedRedeemReqIDs)

	// process new instructions
	err = processPortalInstructions(
		bc, beaconHeight, newInsts, s.sdb, &s.currentPortalStateForProcess, s.portalParams, updatingInfoByTokenID)

	// check results
	s.Equal(1, len(newInsts))
	s.Equal(nil, err)

	//exchangeRateKey := statedb.GeneratePortalFinalExchangeRatesStateObjectKey().String()

	finalExchangeRate := statedb.NewFinalExchangeRatesStateWithValue(
		map[string]statedb.FinalExchangeRatesDetail{
			common.PRVIDStr:       {Amount: 1000000},
			common.PortalBNBIDStr: {Amount: 20000000},
			common.PortalBTCIDStr: {Amount: 10000000000},
		})

	s.Equal(finalExchangeRate, s.currentPortalStateForProcess.FinalExchangeRatesState)
}

/*
	Feature 1: Custodians deposit collateral (PRV)
*/

type TestCaseCustodianDeposit struct {
	custodianIncAddress string
	remoteAddress       map[string]string
	depositAmount       uint64
}

func buildCustodianDepositActionsFromTcs(tcs []TestCaseCustodianDeposit, shardID byte) [][]string {
	insts := [][]string{}

	for _, tc := range tcs {
		inst := buildPortalCustodianDepositAction(tc.custodianIncAddress, tc.remoteAddress, tc.depositAmount, shardID)
		insts = append(insts, inst)
	}

	return insts
}

func (s *PortalTestSuite) TestCustodianDepositCollateral() {
	fmt.Println("Running TestCustodianDepositCollateral - beacon height 1000 ...")
	bc := new(BlockChain)
	beaconHeight := uint64(1000)
	shardID := byte(0)
	newMatchedRedeemReqIDs := []string{}
	updatingInfoByTokenID := map[common.Hash]UpdatingInfo{}

	// build test cases
	testcases := []TestCaseCustodianDeposit{
		// valid
		{
			custodianIncAddress: "custodianIncAddress1",
			remoteAddress: map[string]string{
				common.PortalBNBIDStr: "bnbAddress1",
				common.PortalBTCIDStr: "btcAddress1",
			},
			depositAmount: 5000 * 1e9,
		},
		// custodian deposit more with new remote addresses
		// expect don't change to new remote addresses,
		// custodian is able to update new remote addresses when total collaterals is empty
		{
			custodianIncAddress: "custodianIncAddress1",
			remoteAddress: map[string]string{
				common.PortalBNBIDStr: "bnbAddress2",
				common.PortalBTCIDStr: "btcAddress2",
			},
			depositAmount: 2000 * 1e9,
		},
		// new custodian supply only bnb address
		{
			custodianIncAddress: "custodianIncAddress2",
			remoteAddress: map[string]string{
				common.PortalBNBIDStr: "bnbAddress2",
			},
			depositAmount: 1000 * 1e9,
		},
		// new custodian supply only btc address
		{
			custodianIncAddress: "custodianIncAddress3",
			remoteAddress: map[string]string{
				common.PortalBTCIDStr: "btcAddress3",
			},
			depositAmount: 10000 * 1e9,
		},
	}

	// build actions from testcases
	insts := buildCustodianDepositActionsFromTcs(testcases, shardID)

	// producer instructions
	newInsts, err := producerPortalInstructions(
		bc, beaconHeight, insts, s.sdb, &s.currentPortalStateForProducer, s.portalParams, shardID, newMatchedRedeemReqIDs)

	// process new instructions
	err = processPortalInstructions(
		bc, beaconHeight, newInsts, s.sdb, &s.currentPortalStateForProcess, s.portalParams, updatingInfoByTokenID)

	// check results
	s.Equal(4, len(newInsts))
	s.Equal(nil, err)

	custodianKey1 := statedb.GenerateCustodianStateObjectKey("custodianIncAddress1").String()
	custodianKey2 := statedb.GenerateCustodianStateObjectKey("custodianIncAddress2").String()
	custodianKey3 := statedb.GenerateCustodianStateObjectKey("custodianIncAddress3").String()

	custodian1 := statedb.NewCustodianStateWithValue(
		"custodianIncAddress1", 7000*1e9, 7000*1e9,
		nil, nil,
		map[string]string{
			common.PortalBNBIDStr: "bnbAddress1",
			common.PortalBTCIDStr: "btcAddress1",
		}, nil)

	custodian2 := statedb.NewCustodianStateWithValue(
		"custodianIncAddress2", 1000*1e9, 1000*1e9,
		nil, nil,
		map[string]string{
			common.PortalBNBIDStr: "bnbAddress2",
		}, nil)

	custodian3 := statedb.NewCustodianStateWithValue(
		"custodianIncAddress3", 10000*1e9, 10000*1e9,
		nil, nil,
		map[string]string{
			common.PortalBTCIDStr: "btcAddress3",
		}, nil)

	s.Equal(3, len(s.currentPortalStateForProducer.CustodianPoolState))
	s.Equal(custodian1, s.currentPortalStateForProducer.CustodianPoolState[custodianKey1])
	s.Equal(custodian2, s.currentPortalStateForProducer.CustodianPoolState[custodianKey2])
	s.Equal(custodian3, s.currentPortalStateForProducer.CustodianPoolState[custodianKey3])

	s.Equal(s.currentPortalStateForProcess, s.currentPortalStateForProducer)
}

/*
	Feature 2: Users create porting request
*/

type TestCaseRequestPorting struct {
	portingID     string
	incAddressStr string
	pTokenID      string
	portingAmount uint64
	portingFee    uint64
}

func buildRequestPortingActionsFromTcs(tcs []TestCaseRequestPorting, shardID byte) [][]string {
	insts := [][]string{}

	for i, tc := range tcs {
		inst := buildPortalUserRegisterAction(
			tc.portingID, tc.incAddressStr, tc.pTokenID, tc.portingAmount, tc.portingFee, shardID)
		insts = append(insts, inst)
		fmt.Printf("***** Test case %v *****\n", i)
		fmt.Printf("Porting ID: %v\n", tc.portingID)
		fmt.Printf("tc.incAddressStr: %v\n", tc.incAddressStr)
		fmt.Printf("tc.pTokenID: %v\n", tc.pTokenID)
		fmt.Printf("tc.portingAmount: %v\n", tc.portingAmount)
		fmt.Printf("tc.portingFee: %v\n", tc.portingFee)
	}

	return insts
}

func cloneCustodians(custodians map[string]*statedb.CustodianState) map[string]*statedb.CustodianState{
	newCustodians := make(map[string]*statedb.CustodianState, len(custodians))
	for key, cus := range custodians {
		newCustodians[key] = statedb.NewCustodianStateWithValue(
			cus.GetIncognitoAddress(),
			cus.GetTotalCollateral(),
			cus.GetFreeCollateral(),
			cus.GetHoldingPublicTokens(),
			cus.GetLockedAmountCollateral(),
			cus.GetRemoteAddresses(),
			cus.GetRewardAmount(),
		)
	}
	return newCustodians
}

func (s *PortalTestSuite) SetupTestRequestPorting() {
	custodianKey1 := statedb.GenerateCustodianStateObjectKey("custodianIncAddress1").String()
	custodianKey2 := statedb.GenerateCustodianStateObjectKey("custodianIncAddress2").String()
	custodianKey3 := statedb.GenerateCustodianStateObjectKey("custodianIncAddress3").String()

	custodian1 := statedb.NewCustodianStateWithValue(
		"custodianIncAddress1", 7000*1e9, 7000*1e9,
		nil, nil,
		map[string]string{
			common.PortalBNBIDStr: "bnbAddress1",
			common.PortalBTCIDStr: "btcAddress1",
		}, nil)

	custodian2 := statedb.NewCustodianStateWithValue(
		"custodianIncAddress2", 1000*1e9, 1000*1e9,
		nil, nil,
		map[string]string{
			common.PortalBNBIDStr: "bnbAddress2",
		}, nil)

	custodian3 := statedb.NewCustodianStateWithValue(
		"custodianIncAddress3", 10000*1e9, 10000*1e9,
		nil, nil,
		map[string]string{
			common.PortalBTCIDStr: "btcAddress3",
		}, nil)

	custodians := map[string]*statedb.CustodianState{
		custodianKey1: custodian1,
		custodianKey2: custodian2,
		custodianKey3: custodian3,
	}

	finalExchangeRate := statedb.NewFinalExchangeRatesStateWithValue(
		map[string]statedb.FinalExchangeRatesDetail{
			common.PRVIDStr:       {Amount: 1000000},
			common.PortalBNBIDStr: {Amount: 20000000},
			common.PortalBTCIDStr: {Amount: 10000000000},
		})
	s.currentPortalStateForProducer = CurrentPortalState{
		CustodianPoolState:      custodians,
		FinalExchangeRatesState: finalExchangeRate,
		WaitingPortingRequests:     map[string]*statedb.WaitingPortingRequest{},
		WaitingRedeemRequests:      map[string]*statedb.RedeemRequest{},
		MatchedRedeemRequests:      map[string]*statedb.RedeemRequest{},
		LiquidationPool:            map[string]*statedb.LiquidationPool{},
		LockedCollateralForRewards: new(statedb.LockedCollateralState),
		ExchangeRatesRequests:      map[string]*metadata.ExchangeRatesRequestStatus{},
	}
	s.currentPortalStateForProcess = CurrentPortalState{
		CustodianPoolState:      cloneCustodians(custodians),
		FinalExchangeRatesState: finalExchangeRate,
		WaitingPortingRequests:     map[string]*statedb.WaitingPortingRequest{},
		WaitingRedeemRequests:      map[string]*statedb.RedeemRequest{},
		MatchedRedeemRequests:      map[string]*statedb.RedeemRequest{},
		LiquidationPool:            map[string]*statedb.LiquidationPool{},
		LockedCollateralForRewards: new(statedb.LockedCollateralState),
		ExchangeRatesRequests:      map[string]*metadata.ExchangeRatesRequestStatus{},
	}
}

func (s *PortalTestSuite) TestRequestPorting() {
	fmt.Println("Running TestRequestPorting - beacon height 1001 ...")
	bc := new(BlockChain)
	beaconHeight := uint64(1001)
	shardID := byte(0)
	newMatchedRedeemReqIDs := []string{}
	updatingInfoByTokenID := map[common.Hash]UpdatingInfo{}

	s.SetupTestRequestPorting()

	// build test cases
	testcases := []TestCaseRequestPorting{
		// valid porting request with 0.01% porting fee
		{
			portingID:     "porting-bnb-1",
			incAddressStr: "userIncAddress1",
			pTokenID:      common.PortalBNBIDStr,
			portingAmount: 1 * 1e9,
			portingFee:    getMinFee(
				1 * 1e9, common.PortalBNBIDStr,
				s.currentPortalStateForProducer.FinalExchangeRatesState, s.portalParams.MinPercentPortingFee),
		},
		// invalid porting request with duplicate porting ID
		{
			portingID:     "porting-bnb-1",
			incAddressStr: "userIncAddress2",
			pTokenID:      common.PortalBNBIDStr,
			portingAmount: 1 * 1e9,
			portingFee:    getMinFee(
				1 * 1e9, common.PortalBNBIDStr,
				s.currentPortalStateForProducer.FinalExchangeRatesState, s.portalParams.MinPercentPortingFee),
		},
		// invalid porting request with invalid porting fee
		{
			portingID:     "porting-bnb-2",
			incAddressStr: "userIncAddress2",
			pTokenID:      common.PortalBNBIDStr,
			portingAmount: 1 * 1e9,
			portingFee:    getMinFee(
				1 * 1e9, common.PortalBNBIDStr,
				s.currentPortalStateForProducer.FinalExchangeRatesState, s.portalParams.MinPercentPortingFee) - 1,
		},
		// valid porting request with >0.01% porting fee
		{
			portingID:     "porting-btc-2",
			incAddressStr: "userIncAddress2",
			pTokenID:      common.PortalBTCIDStr,
			portingAmount: 0.1*1e9,
			portingFee:    getMinFee(
				0.1*1e9, common.PortalBTCIDStr,
				s.currentPortalStateForProducer.FinalExchangeRatesState, s.portalParams.MinPercentPortingFee) + 1,
		},
		// invalid porting request: custodians' collateral not enough for the porting amount
		{
			portingID:     "porting-btc-3",
			incAddressStr: "userIncAddress3",
			pTokenID:      common.PortalBTCIDStr,
			portingAmount: 1*1e9,
			portingFee:    getMinFee(
				1*1e9, common.PortalBTCIDStr,
				s.currentPortalStateForProducer.FinalExchangeRatesState, s.portalParams.MinPercentPortingFee),
		},
	}

	// build actions from testcases
	insts := buildRequestPortingActionsFromTcs(testcases, shardID)

	// producer instructions
	newInsts, err := producerPortalInstructions(
		bc, beaconHeight - 1, insts, s.sdb, &s.currentPortalStateForProducer, s.portalParams, shardID, newMatchedRedeemReqIDs)
	s.Equal(nil, err)

	// process new instructions
	err = processPortalInstructions(
		bc, beaconHeight - 1, newInsts, s.sdb, &s.currentPortalStateForProcess, s.portalParams, updatingInfoByTokenID)

	// check results
	s.Equal(5, len(newInsts))
	s.Equal(nil, err)

	wPortingReqKey1 := statedb.GeneratePortalWaitingPortingRequestObjectKey("porting-bnb-1").String()
	lockedCollateralAmount1 := getLockedCollateralAmount(
		1*1e9, common.PortalBNBIDStr,
		s.currentPortalStateForProducer.FinalExchangeRatesState, s.portalParams.MinPercentLockedCollateral)
	portingFee1 := getMinFee(
		1 * 1e9, common.PortalBNBIDStr,
		s.currentPortalStateForProducer.FinalExchangeRatesState, s.portalParams.MinPercentPortingFee)
	wPortingRequest1 := statedb.NewWaitingPortingRequestWithValue(
		"porting-bnb-1", common.Hash{}, common.PortalBNBIDStr,
		"userIncAddress1", 1*1e9,
		[]*statedb.MatchingPortingCustodianDetail{
			{
				IncAddress:             "custodianIncAddress1",
				RemoteAddress:          "bnbAddress1",
				Amount:                 1 * 1e9,
				LockedAmountCollateral: lockedCollateralAmount1,
			},
		}, portingFee1, beaconHeight)

	wPortingReqKey2 := statedb.GeneratePortalWaitingPortingRequestObjectKey("porting-btc-2").String()
	lockedCollateralAmount2 := getLockedCollateralAmount(
		0.1*1e9, common.PortalBTCIDStr,
		s.currentPortalStateForProducer.FinalExchangeRatesState, s.portalParams.MinPercentLockedCollateral)
	portingFee2 := getMinFee(
		0.1 * 1e9, common.PortalBTCIDStr,
		s.currentPortalStateForProducer.FinalExchangeRatesState, s.portalParams.MinPercentPortingFee) + 1
	wPortingRequest2 := statedb.NewWaitingPortingRequestWithValue(
		"porting-btc-2", common.Hash{}, common.PortalBTCIDStr,
		"userIncAddress2", 0.1*1e9,
		[]*statedb.MatchingPortingCustodianDetail{
			{
				IncAddress:             "custodianIncAddress3",
				RemoteAddress:          "btcAddress3",
				Amount:                 0.1 * 1e9,
				LockedAmountCollateral: lockedCollateralAmount2,
			},
		}, portingFee2, beaconHeight)

	fmt.Printf("lockedCollateralAmount1: %v\n", lockedCollateralAmount1)
	fmt.Printf("portingFee1: %v\n", portingFee1)
	fmt.Printf("lockedCollateralAmount2: %v\n", lockedCollateralAmount2)
	fmt.Printf("portingFee2: %v\n", portingFee2)

	s.Equal(2, len(s.currentPortalStateForProducer.WaitingPortingRequests))
	s.Equal(wPortingRequest1, s.currentPortalStateForProducer.WaitingPortingRequests[wPortingReqKey1])
	s.Equal(wPortingRequest2, s.currentPortalStateForProducer.WaitingPortingRequests[wPortingReqKey2])

	fmt.Printf("wPortingReqKey2: %v\n", s.currentPortalStateForProducer.WaitingPortingRequests[wPortingReqKey2].Custodians()[0].IncAddress)

	// custodian state after matching porting requests
	custodianKey1 := statedb.GenerateCustodianStateObjectKey("custodianIncAddress1").String()
	custodianKey2 := statedb.GenerateCustodianStateObjectKey("custodianIncAddress2").String()
	custodianKey3 := statedb.GenerateCustodianStateObjectKey("custodianIncAddress3").String()

	custodian1 := statedb.NewCustodianStateWithValue(
		"custodianIncAddress1", 7000*1e9, 7000*1e9-lockedCollateralAmount1,
		nil,
		map[string]uint64{
			common.PortalBNBIDStr: lockedCollateralAmount1,
		},
		map[string]string{
			common.PortalBNBIDStr: "bnbAddress1",
			common.PortalBTCIDStr: "btcAddress1",
		},nil)

	custodian2 := statedb.NewCustodianStateWithValue(
		"custodianIncAddress2", 1000*1e9, 1000*1e9,
		nil, nil,
		map[string]string{
			common.PortalBNBIDStr: "bnbAddress2",
		}, nil)

	custodian3 := statedb.NewCustodianStateWithValue(
		"custodianIncAddress3", 10000*1e9, 10000*1e9 - lockedCollateralAmount2,
		nil,
		map[string]uint64{
			common.PortalBTCIDStr: lockedCollateralAmount2,
		},
		map[string]string{
			common.PortalBTCIDStr: "btcAddress3",
		}, nil)

	s.Equal(custodian1, s.currentPortalStateForProducer.CustodianPoolState[custodianKey1])
	s.Equal(custodian2, s.currentPortalStateForProducer.CustodianPoolState[custodianKey2])
	s.Equal(custodian3, s.currentPortalStateForProducer.CustodianPoolState[custodianKey3])

	s.Equal(s.currentPortalStateForProcess, s.currentPortalStateForProducer)
}

/*
	Feature 3: Users submit proof to request pTokens after sending public tokens to custodians
*/

/*
	Feature 4:
*/

func TestPortalSuite(t *testing.T) {
	suite.Run(t, new(PortalTestSuite))
}
