package blockchain

import (
	"encoding/base64"
	"encoding/json"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"os"
	"strconv"
	"fmt"
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
	portalParams PortalParams
}

func (s *PortalTestSuite) SetupTest() {
	dbPath, err := ioutil.TempDir(os.TempDir(), "test_statedb_")
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
		MaxPercentLiquidatedCollateralAmount: 105,
		MaxPercentCustodianRewards:           10,
		MinPercentCustodianRewards:           1,
		MinLockCollateralAmountInEpoch:       5000 * 1e9, // 5000 prv
		MinPercentLockedCollateral:           150,
		TP120:                                120,
		TP130:                                130,
		MinPercentPortingFee:                 0.01,
		MinPercentRedeemFee:                  0.01,
	}
}

/*
 Utility functions
*/

func producerPortalInstructions(
	blockchain *BlockChain,
	beaconHeight uint64,
	insts [][]string,
	portalStateDB *statedb.StateDB,
	currentPortalState *CurrentPortalState,
	portalParams PortalParams,
	shardID byte,
	newMatchedRedeemReqIDs []string,
) ([][]string, error){
	var err error
	var newInst [][]string
	var newInsts [][]string
	for _, inst := range insts {
		switch inst[0] {
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

			/*

			// liquidation custodian run away
		case strconv.Itoa(metadata.PortalLiquidateCustodianMeta):
			newInst, err = blockchain.processPortalLiquidateCustodian(portalStateDB, beaconHeight, inst, currentPortalState, portalParams)
			// portal reward
		case strconv.Itoa(metadata.PortalRewardMeta):
			newInst, err = blockchain.processPortalReward(portalStateDB, beaconHeight, inst, currentPortalState, portalParams)
			// request withdraw reward
		case strconv.Itoa(metadata.PortalRequestWithdrawRewardMeta):
			newInst, err = blockchain.processPortalWithdrawReward(portalStateDB, beaconHeight, inst, currentPortalState, portalParams)
			// expired waiting porting request
		case strconv.Itoa(metadata.PortalExpiredWaitingPortingReqMeta):
			newInst, err = blockchain.processPortalExpiredPortingRequest(portalStateDB, beaconHeight, inst, currentPortalState, portalParams)
			// total custodian reward instruction
		case strconv.Itoa(metadata.PortalTotalRewardCustodianMeta):
			newInst, err = blockchain.processPortalTotalCustodianReward(portalStateDB, beaconHeight, inst, currentPortalState, portalParams)
			//exchange rates
		case strconv.Itoa(metadata.PortalExchangeRatesMeta):
			newInst, err = blockchain.processPortalExchangeRates(portalStateDB, beaconHeight, inst, currentPortalState, portalParams)
			//custodian withdraw
		case strconv.Itoa(metadata.PortalCustodianWithdrawRequestMeta):
			newInst, err = blockchain.processPortalCustodianWithdrawRequest(portalStateDB, beaconHeight, inst, currentPortalState, portalParams)
			//liquidation exchange rates
		case strconv.Itoa(metadata.PortalLiquidateTPExchangeRatesMeta):
			newInst, err = blockchain.processLiquidationTopPercentileExchangeRates(portalStateDB, beaconHeight, inst, currentPortalState, portalParams)
			//liquidation custodian deposit
		case strconv.Itoa(metadata.PortalLiquidationCustodianDepositMetaV2):
			newInst, err = blockchain.processPortalLiquidationCustodianDeposit(portalStateDB, beaconHeight, inst, currentPortalState, portalParams)
			//waiting porting top up
		case strconv.Itoa(metadata.PortalTopUpWaitingPortingRequestMeta):
			newInst, err = blockchain.processPortalTopUpWaitingPorting(portalStateDB, beaconHeight, inst, currentPortalState, portalParams)
			//liquidation user redeem
		case strconv.Itoa(metadata.PortalRedeemLiquidateExchangeRatesMeta):
			newInst, err = blockchain.processPortalRedeemLiquidateExchangeRates(portalStateDB, beaconHeight, inst, currentPortalState, portalParams, updatingInfoByTokenID)

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

func buildPortalCustodianDepositAction(
	incAddressStr string,
	remoteAddress map[string]string,
	depositAmount uint64,
	shardID byte,
) []string {
	data := metadata.PortalCustodianDeposit{
		MetadataBase:    metadata.MetadataBase{
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








/*
	Feature 1: Custodians deposit collateral (PRV)
*/

type TestCaseCustodianDeposit struct {
	custodianIncAddress string
	remoteAddress map[string]string
	depositAmount uint64

	isAccepted bool
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

	// build test cases
	testcases := []TestCaseCustodianDeposit{
		{
			custodianIncAddress: "custodianIncAddress1",
			remoteAddress:       map[string]string{
				"bnb": "bnbAddress1",
				"btc": "btcAddress1",
			},
			depositAmount:       10*1e9,
			isAccepted:          true,
		},
	}

	// build actions from testcases
	insts := buildCustodianDepositActionsFromTcs(testcases, shardID)

	// producer instructions
	newInsts, err := producerPortalInstructions(
		bc, beaconHeight, insts, s.sdb, &s.currentPortalStateForProducer, s.portalParams, shardID, newMatchedRedeemReqIDs)

	// check results
	s.Equal(1, len(newInsts))
	s.Equal(nil, err)

	s.Equal(true, newInsts[0][])


}



/*
	Feature 2: Users create porting request
*/



/*
	Feature 3: Users submit proof to request pTokens after sending public tokens to custodians
*/



/*
	Feature 4:
*/



