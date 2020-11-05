package blockchain

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	typesBNB "github.com/binance-chain/go-sdk/common/types"
	"github.com/binance-chain/go-sdk/types/msg"
	txBNB "github.com/binance-chain/go-sdk/types/tx"
	eCommon "github.com/ethereum/go-ethereum/common"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/relaying/bnb"
	"github.com/stretchr/testify/suite"
	"github.com/tendermint/tendermint/types"
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
type PortalTestSuiteV3 struct {
	suite.Suite
	currentPortalStateForProducer CurrentPortalState
	currentPortalStateForProcess  CurrentPortalState
	sdb                           *statedb.StateDB
	blockChain                    *BlockChain
}

const USER_INC_ADDRESS_1 = "12S5pBBRDf1GqfRHouvCV86sWaHzNfvakAWpVMvNnWu2k299xWCgQzLLc9wqPYUHfMYGDprPvQ794dbi6UU1hfRN4tPiU61txWWenhC"
const USER_INC_ADDRESS_2 = "12S1a8VnkwhDTQWZ5PhdpySwiFZj7p8sKdG7oAQFZ3dLsWaV6fhDWk5aSFHpt1jcPBjY4sYgwqAqRzx3oTYDZCvCei1LSCdJARXWiyK"
const USER_INC_ADDRESS_3 = "12Rs38qgRQZXdeKgy9QS2ebg6pHA875eJcxQRof3QAkhBFaBDKovGbxvPqV2wa2k128vQY7ahuXLSNcUvs2g52QsFFbdiaqQUWKLdc4"
const USER_BNB_ADDRESS_1 = "tbnb1d90lad6rg5ldh8vxgtuwzxd8n6rhhx7mfqek38"
const USER_BNB_ADDRESS_2 = "tbnb172pnrmd0409237jwlq5qjhw2s2r7lq6ukmaeke"

const CUS_INC_ADDRESS_1 = "12RuEdPjq4yxivzm8xPxRVHmkL74t4eAdUKPdKKhMEnpxPH3k8GEyULbwq4hjwHWmHQr7MmGBJsMpdCHsYAqNE18jipWQwciBf9yqvQ"
const CUS_INC_ADDRESS_2 = "12Rwz4HXkVABgRnSb5Gfu1FaJ7auo3fLNXVGFhxx1dSytxHpWhbkimT1Mv5Z2oCMsssSXTVsapY8QGBZd2J4mPiCTzJAtMyCzb4dDcy"
const CUS_INC_ADDRESS_3 = "12S4NL3DZ1KoprFRy1k5DdYSXUq81NtxFKdvUTP3PLqQypWzceL5fBBwXooAsX5s23j7cpb1Za37ddmfSaMpEJDPsnJGZuyWTXJSZZ5"
const CUS_BNB_ADDRESS_1 = "tbnb19cmxazhx5ujlhhlvj9qz0wv8a4vvsx8vuy9cyc"
const CUS_BNB_ADDRESS_2 = "tbnb1zyqrky9zcumc2e4smh3xwh2u8kudpdc56gafuk"
const CUS_BNB_ADDRESS_3 = "tbnb1n5lrzass9l28djvv7drst53dcw7y9yj4pyvksf"
const CUS_BTC_ADDRESS_1 = "btc-address-1"
const CUS_BTC_ADDRESS_2 = "btc-address-2"
const CUS_BTC_ADDRESS_3 = "btc-address-3"
const CUS_ETH_ADDRESS_1 = "eth-address-1"
const CUS_ETH_ADDRESS_2 = "eth-address-2"
const CUS_ETH_ADDRESS_3 = "eth-address-3"

const USDT_ID = "64fbdbc6bf5b228814b58706d91ed03777f0edf6"
const ETH_ID = "0000000000000000000000000000000000000000"

const BNB_NODE_URL = "https://data-seed-pre-0-s3.binance.org:443"

func (s *PortalTestSuiteV3) SetupTest() {
	dbPath, err := ioutil.TempDir(os.TempDir(), "portal_test_statedb_")
	if err != nil {
		panic(err)
	}
	diskBD, _ := incdb.Open("leveldb", dbPath)
	warperDBStatedbTest := statedb.NewDatabaseAccessWarper(diskBD)
	emptyRoot := common.HexToHash(common.HexEmptyRoot)
	stateDB, _ := statedb.NewWithPrefixTrie(emptyRoot, warperDBStatedbTest)

	s.sdb = stateDB

	finalExchangeRate := statedb.NewFinalExchangeRatesStateWithValue(
		map[string]statedb.FinalExchangeRatesDetail{
			common.PRVIDStr:       {Amount: 1000000},
			common.PortalBNBIDStr: {Amount: 20000000},
			common.PortalBTCIDStr: {Amount: 10000000000},
			ETH_ID:                {Amount: 400000000},
			USDT_ID:               {Amount: 1000000},
		})
	s.currentPortalStateForProducer = CurrentPortalState{
		CustodianPoolState:         map[string]*statedb.CustodianState{},
		WaitingPortingRequests:     map[string]*statedb.WaitingPortingRequest{},
		WaitingRedeemRequests:      map[string]*statedb.RedeemRequest{},
		MatchedRedeemRequests:      map[string]*statedb.RedeemRequest{},
		FinalExchangeRatesState:    finalExchangeRate,
		LiquidationPool:            map[string]*statedb.LiquidationPool{},
		LockedCollateralForRewards: new(statedb.LockedCollateralState),
		ExchangeRatesRequests:      map[string]*metadata.ExchangeRatesRequestStatus{},
	}
	s.currentPortalStateForProcess = CurrentPortalState{
		CustodianPoolState:         map[string]*statedb.CustodianState{},
		WaitingPortingRequests:     map[string]*statedb.WaitingPortingRequest{},
		WaitingRedeemRequests:      map[string]*statedb.RedeemRequest{},
		MatchedRedeemRequests:      map[string]*statedb.RedeemRequest{},
		FinalExchangeRatesState:    finalExchangeRate,
		LiquidationPool:            map[string]*statedb.LiquidationPool{},
		LockedCollateralForRewards: new(statedb.LockedCollateralState),
		ExchangeRatesRequests:      map[string]*metadata.ExchangeRatesRequestStatus{},
	}
	s.blockChain = &BlockChain{
		config: Config{
			ChainParams: &Params{
				MinBeaconBlockInterval: 40 * time.Second,
				MinShardBlockInterval:  40 * time.Second,
				Epoch:                  100,
				PortalTokens: map[string]PortalTokenProcessor{
					common.PortalBTCIDStr: &PortalBTCTokenProcessor{
						&PortalToken{
							ChainID: "Bitcoin-Testnet",
						},
					},
					common.PortalBNBIDStr: &PortalBNBTokenProcessor{
						&PortalToken{
							ChainID: "Binance-Chain-Ganges",
						},
					},
				},
				PortalParams: map[uint64]PortalParams{
					0: {
						TimeOutCustodianReturnPubToken:       24 * time.Hour,
						TimeOutWaitingPortingRequest:         24 * time.Hour,
						TimeOutWaitingRedeemRequest:          15 * time.Minute,
						MaxPercentLiquidatedCollateralAmount: 120,
						MaxPercentCustodianRewards:           10,
						MinPercentCustodianRewards:           1,
						MinLockCollateralAmountInEpoch:       10000 * 1e6, // 10000 usd
						MinPercentLockedCollateral:           200,
						TP120:                                120,
						TP130:                                130,
						MinPercentPortingFee:                 0.01,
						MinPercentRedeemFee:                  0.01,
						SupportedCollateralTokens:            getSupportedPortalCollateralsTestnet(),
					},
				},
				BNBFullNodeProtocol:      TestnetBNBFullNodeProtocol,
				BNBFullNodeHost:          TestnetBNBFullNodeHost,
				BNBFullNodePort:          TestnetBNBFullNodePort,
				BNBRelayingHeaderChainID: TestnetBNBChainID,
			},
		},
	}
}

/*
 Utility functions
*/

func exchangeRates(amount uint64, tokenIDFrom string, tokenIDTo string, finalExchangeRate *statedb.FinalExchangeRatesState) uint64 {
	convertTool := NewPortalExchangeRateTool(finalExchangeRate, getSupportedPortalCollateralsTestnet())
	res, _ := convertTool.Convert(tokenIDFrom, tokenIDTo, amount)
	return res
}

func getLockedCollateralAmount(
	portingAmount uint64, tokenID string, collateralTokenID string, finalExchangeRate *statedb.FinalExchangeRatesState, percent uint64) uint64 {
	amount := upPercent(portingAmount, percent)
	return exchangeRates(amount, tokenID, collateralTokenID, finalExchangeRate)
}

func getMinFee(amount uint64, tokenID string, finalExchangeRate *statedb.FinalExchangeRatesState, percent float64) uint64 {
	amountInPRV := exchangeRates(amount, tokenID, common.PRVIDStr, finalExchangeRate)
	fee := float64(amountInPRV) * percent / float64(100)
	return uint64(math.Round(fee))
}

func getUnlockAmount(totalLockedAmount uint64, totalPTokenAmount uint64, pTokenAmount uint64) uint64 {
	amount := new(big.Int).Mul(new(big.Int).SetUint64(pTokenAmount), new(big.Int).SetUint64(totalLockedAmount))
	amount = amount.Div(amount, new(big.Int).SetUint64(totalPTokenAmount))
	return amount.Uint64()
}

func (s *PortalTestSuiteV3) TestGetLockedCollateralAmount() {
	portingAmount := uint64(100 * 1e9)
	tokenID := common.PortalBNBIDStr
	collateralTokenID := USDT_ID

	percent := s.blockChain.GetPortalParams(0).MinPercentLockedCollateral
	amount := getLockedCollateralAmount(portingAmount, tokenID, collateralTokenID, s.currentPortalStateForProducer.FinalExchangeRatesState, percent)
	fmt.Println("Result from TestGetLockedCollateralAmount: ", amount)
}

func (s *PortalTestSuiteV3) TestGetMinFee() {
	amount := uint64(150 * 1e9)
	tokenID := common.PortalBNBIDStr
	percent := s.blockChain.GetPortalParams(0).MinPercentPortingFee

	fee := getMinFee(amount, tokenID, s.currentPortalStateForProducer.FinalExchangeRatesState, percent)
	fmt.Println("Result from TestGetMinFee: ", fee)
}

func (s *PortalTestSuiteV3) TestGetUnlockAmount() {
	totalLockedAmount := uint64(40000000000)
	totalPTokenAmount := uint64(1 * 1e9)
	pTokenAmount := uint64(0.3 * 1e9)

	unlockAmount := getUnlockAmount(totalLockedAmount, totalPTokenAmount, pTokenAmount)
	fmt.Println("Result from TestGetUnlockAmount: ", unlockAmount)
}

func (s *PortalTestSuiteV3) TestExchangeRate() {
	//s.currentPortalStateForProducer.FinalExchangeRatesState = statedb.NewFinalExchangeRatesStateWithValue(
	//	map[string]statedb.FinalExchangeRatesDetail{
	//		common.PRVIDStr:       {Amount: 1000000},
	//		common.PortalBNBIDStr: {Amount: 40000000},
	//		common.PortalBTCIDStr: {Amount: 10000000000},
	//	})
	//
	//s.currentPortalStateForProducer.FinalExchangeRatesState = s.
	amount := uint64(12500000000 * 2)
	tokenIDFrom := common.PortalBNBIDStr
	tokenIDTo := USDT_ID
	convertAmount := exchangeRates(amount, tokenIDFrom, tokenIDTo, s.currentPortalStateForProducer.FinalExchangeRatesState)
	//convertAmount = convertAmount * 120 / 100
	fmt.Println("Result from TestExchangeRate: ", convertAmount)
}

func cloneMap(m map[string]uint64) map[string]uint64 {
	if m == nil {
		return nil
	}
	newMap := make(map[string]uint64, len(m))
	for k, v := range m {
		newMap[k] = v
	}
	return newMap
}

func cloneMapOfMap(m map[string]map[string]uint64) map[string]map[string]uint64 {
	if m == nil {
		return nil
	}
	newMap := make(map[string]map[string]uint64, len(m))
	for k, v := range m {
		newMap[k] = cloneMap(v)
	}
	return newMap
}

func cloneCustodians(custodians map[string]*statedb.CustodianState) map[string]*statedb.CustodianState {
	newCustodians := make(map[string]*statedb.CustodianState, len(custodians))
	for key, cus := range custodians {
		newCustodians[key] = statedb.NewCustodianStateWithValue(
			cus.GetIncognitoAddress(),
			cus.GetTotalCollateral(),
			cus.GetFreeCollateral(),
			cloneMap(cus.GetHoldingPublicTokens()),
			cloneMap(cus.GetLockedAmountCollateral()),
			cus.GetRemoteAddresses(),
			cloneMap(cus.GetRewardAmount()),
			cloneMap(cus.GetTotalTokenCollaterals()),
			cloneMap(cus.GetFreeTokenCollaterals()),
			cloneMapOfMap(cus.GetLockedTokenCollaterals()),
		)
	}
	return newCustodians
}

func cloneMatchingPortingCustodians(custodians []*statedb.MatchingPortingCustodianDetail) []*statedb.MatchingPortingCustodianDetail {
	newMatchingCustodians := make([]*statedb.MatchingPortingCustodianDetail, len(custodians))
	for i, cus := range custodians {
		newMatchingCustodians[i] = &statedb.MatchingPortingCustodianDetail{
			IncAddress:             cus.IncAddress,
			RemoteAddress:          cus.RemoteAddress,
			Amount:                 cus.Amount,
			LockedAmountCollateral: cus.LockedAmountCollateral,
			LockedTokenCollaterals: cus.LockedTokenCollaterals,
		}
	}
	return newMatchingCustodians
}

func cloneWPortingRequests(wPortingReqs map[string]*statedb.WaitingPortingRequest) map[string]*statedb.WaitingPortingRequest {
	newReqs := make(map[string]*statedb.WaitingPortingRequest, len(wPortingReqs))
	for key, req := range wPortingReqs {
		newReqs[key] = statedb.NewWaitingPortingRequestWithValue(
			req.UniquePortingID(),
			req.TxReqID(),
			req.TokenID(),
			req.PorterAddress(),
			req.Amount(),
			cloneMatchingPortingCustodians(req.Custodians()),
			req.PortingFee(),
			req.BeaconHeight(),
			req.ShardHeight(),
			req.ShardID(),
		)
	}
	return newReqs
}

//func cloneRedeemRequests(redeemReqs map[string]*statedb.RedeemRequest) map[string]*statedb.RedeemRequest {
//	newReqs := make(map[string]*statedb.RedeemRequest, len(redeemReqs))
//	for key, req := range redeemReqs {
//		newReqs[key] = statedb.NewRedeemRequestWithValue(
//			req.GetUniqueRedeemID(),
//			req.GetTokenID(),
//			req.GetRedeemerAddress(),
//			req.GetRedeemerRemoteAddress(),
//			req.GetRedeemAmount(),
//			req.GetCustodians(),
//			req.GetRedeemFee(),
//			req.GetBeaconHeight(),
//			req.GetTxReqID(),
//		)
//	}
//	return newReqs
//}

// buildBNBProofFromTxs build a bnb proof for unit tests
func buildBNBProofFromTxs(blockHeight int64, txs *types.Txs, indexTx int) *bnb.BNBProof {
	proof := txs.Proof(indexTx)

	return &bnb.BNBProof{
		Proof:       &proof,
		BlockHeight: blockHeight,
	}
}

func createSendMsg(fromAddr string, transferInfo map[string]int64) msg.SendMsg {
	fromAccAddr, _ := typesBNB.AccAddressFromHex(fromAddr)

	transfer := make([]msg.Transfer, 0)
	totalAmountTransfer := int64(0)
	for toAddrStr, amount := range transferInfo {
		toAddr, _ := typesBNB.AccAddressFromHex(toAddrStr)
		transfer = append(transfer, msg.Transfer{
			ToAddr: toAddr,
			Coins: typesBNB.Coins{
				typesBNB.Coin{
					Denom:  bnb.DenomBNB,
					Amount: amount,
				},
			},
		})
		totalAmountTransfer += amount
	}

	fromCoins := typesBNB.Coins{
		typesBNB.Coin{
			Denom:  bnb.DenomBNB,
			Amount: totalAmountTransfer,
		},
	}

	sendMsg := msg.CreateSendMsg(fromAccAddr, fromCoins, transfer)
	return sendMsg
}

func createTxs(fromAddr string, tranferInfo map[string]int64, memo string) *types.Txs {
	// create SendMsg
	sendMsg := createSendMsg(fromAddr, tranferInfo)

	// create StdTx
	stdTx := txBNB.NewStdTx([]msg.Msg{sendMsg}, []txBNB.StdSignature{}, memo, int64(0), []byte{})

	txBytes, _ := types.GetCodec().MarshalBinaryLengthPrefixed(stdTx)
	txs := &types.Txs{txBytes}
	return txs
}

// createMemo create memo for porting tx or redeem tx
// if custodianIncAddr is empty, create memo for porting
func createMemo(id string, custodianIncAddr string) string {
	type PortingMemoBNB struct {
		PortingID string `json:"PortingID"`
	}

	type RedeemMemoBNB struct {
		RedeemID                  string `json:"RedeemID"`
		CustodianIncognitoAddress string `json:"CustodianIncognitoAddress"`
	}

	memo := ""

	if custodianIncAddr == "" {
		memoPorting := PortingMemoBNB{PortingID: id}
		memoPortingBytes, _ := json.Marshal(memoPorting)
		memo = base64.StdEncoding.EncodeToString(memoPortingBytes)
	} else {
		memoRedeem := RedeemMemoBNB{RedeemID: id, CustodianIncognitoAddress: custodianIncAddr}
		memoRedeemBytes, _ := json.Marshal(memoRedeem)
		memoRedeemHash := common.HashB(memoRedeemBytes)
		memo = base64.StdEncoding.EncodeToString(memoRedeemHash)
	}

	return memo
}

func buildBNBProof(blockHeight int64, fromAddr string, transferInfo map[string]int64, id string, msg string) (string, []byte) {
	indexTx := 0

	// build memo attach to tx
	memo := createMemo(id, msg)

	txs := createTxs(fromAddr, transferInfo, memo)

	bnbProof := buildBNBProofFromTxs(blockHeight, txs, indexTx)
	bnbProofBytes, _ := json.Marshal(bnbProof)
	bnbProofStr := base64.StdEncoding.EncodeToString(bnbProofBytes)
	rootHash := txs.Hash()

	fmt.Println("Result from TestBuildBNBProof bnbProofStr: ", bnbProofStr)
	fmt.Printf("Result from TestBuildBNBProof rootHash: %#v\n", txs.Hash())

	return bnbProofStr, rootHash
}

//todo:
func (s *PortalTestSuiteV3) TestBuildBNBProof() {
	//buildBNBProof()
}

type instructionForProducer struct {
	inst         []string
	optionalData map[string]interface{}
}

func producerPortalInstructions(
	blockchain *BlockChain,
	beaconHeight uint64,
	insts []instructionForProducer,
	currentPortalState *CurrentPortalState,
	portalParams PortalParams,
	shardID byte,
	pm *portalManager,
) ([][]string, error) {
	var newInsts [][]string

	for _, item := range insts {
		inst := item.inst
		optionalData := item.optionalData

		metaType, _ := strconv.Atoi(inst[0])
		contentStr := inst[1]
		portalProcessor := pm.portalInstructions[metaType]
		newInst, err := portalProcessor.buildNewInsts(
			blockchain,
			contentStr,
			shardID,
			currentPortalState,
			beaconHeight,
			portalParams,
			optionalData,
		)
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
		// ============ Exchange rate ============
		case strconv.Itoa(metadata.PortalExchangeRatesMeta):
			err = blockchain.processPortalExchangeRates(portalStateDB, beaconHeight, inst, currentPortalState, portalParams)

		// ============ Custodian ============
		// custodian deposit collateral
		case strconv.Itoa(metadata.PortalCustodianDepositMeta):
			err = blockchain.processPortalCustodianDeposit(portalStateDB, beaconHeight, inst, currentPortalState, portalParams)
		// custodian withdraw collateral
		case strconv.Itoa(metadata.PortalCustodianWithdrawRequestMeta):
			err = blockchain.processPortalCustodianWithdrawRequest(portalStateDB, beaconHeight, inst, currentPortalState, portalParams)
		// custodian deposit collateral v3
		case strconv.Itoa(metadata.PortalCustodianDepositMetaV3):
			err = blockchain.processPortalCustodianDepositV3(portalStateDB, beaconHeight, inst, currentPortalState, portalParams)
		// custodian request withdraw collateral v3
		case strconv.Itoa(metadata.PortalCustodianWithdrawRequestMetaV3):
			err = blockchain.processPortalCustodianWithdrawV3(portalStateDB, beaconHeight, inst, currentPortalState, portalParams)

		// ============ Porting flow ============
		// porting request
		case strconv.Itoa(metadata.PortalUserRegisterMeta):
			err = blockchain.processPortalUserRegister(portalStateDB, beaconHeight, inst, currentPortalState, portalParams)
		// request ptoken
		case strconv.Itoa(metadata.PortalUserRequestPTokenMeta):
			err = blockchain.processPortalUserReqPToken(portalStateDB, beaconHeight, inst, currentPortalState, portalParams, updatingInfoByTokenID)

		// ============ Redeem flow ============
		// redeem request
		case strconv.Itoa(metadata.PortalRedeemRequestMeta):
			err = blockchain.processPortalRedeemRequest(portalStateDB, beaconHeight, inst, currentPortalState, portalParams, updatingInfoByTokenID)
		// custodian request matching waiting redeem requests
		case strconv.Itoa(metadata.PortalReqMatchingRedeemMeta):
			err = blockchain.processPortalReqMatchingRedeem(portalStateDB, beaconHeight, inst, currentPortalState, portalParams)
		case strconv.Itoa(metadata.PortalPickMoreCustodianForRedeemMeta):
			err = blockchain.processPortalPickMoreCustodiansForTimeOutWaitingRedeemReq(portalStateDB, beaconHeight, inst, currentPortalState, portalParams)
		// request unlock collateral
		case strconv.Itoa(metadata.PortalRequestUnlockCollateralMeta), strconv.Itoa(metadata.PortalRequestUnlockCollateralMetaV3):
			err = blockchain.processPortalUnlockCollateral(portalStateDB, beaconHeight, inst, currentPortalState, portalParams)

		// ============ Liquidation ============
		// liquidation custodian run away
		case strconv.Itoa(metadata.PortalLiquidateCustodianMeta), strconv.Itoa(metadata.PortalLiquidateCustodianMetaV3):
			err = blockchain.processPortalLiquidateCustodian(portalStateDB, beaconHeight, inst, currentPortalState, portalParams)
		//liquidation exchange rates
		case strconv.Itoa(metadata.PortalLiquidateTPExchangeRatesMeta):
			err = blockchain.processLiquidationTopPercentileExchangeRates(portalStateDB, beaconHeight, inst, currentPortalState, portalParams)
		// custodian topup
		case strconv.Itoa(metadata.PortalCustodianTopupMetaV2):
			err = blockchain.processPortalCustodianTopup(portalStateDB, beaconHeight, inst, currentPortalState, portalParams)
		// top up for waiting porting
		case strconv.Itoa(metadata.PortalTopUpWaitingPortingRequestMeta):
			err = blockchain.processPortalTopUpWaitingPorting(portalStateDB, beaconHeight, inst, currentPortalState, portalParams)
		// redeem from liquidation pool
		case strconv.Itoa(metadata.PortalRedeemFromLiquidationPoolMeta):
			err = blockchain.processPortalRedeemLiquidateExchangeRates(portalStateDB, beaconHeight, inst, currentPortalState, portalParams, updatingInfoByTokenID)
		// expired waiting porting request
		case strconv.Itoa(metadata.PortalExpiredWaitingPortingReqMeta):
			err = blockchain.processPortalExpiredPortingRequest(portalStateDB, beaconHeight, inst, currentPortalState, portalParams)

		// liquidation by exchange rate v3
		case strconv.Itoa(metadata.PortalLiquidateByRatesMetaV3):
			err = blockchain.processLiquidationByExchangeRatesV3(portalStateDB, beaconHeight, inst, currentPortalState, portalParams)
		// redeem from liquidation pool v3
		case strconv.Itoa(metadata.PortalRedeemFromLiquidationPoolMetaV3):
			err = blockchain.processPortalRedeemFromLiquidationPoolV3(portalStateDB, beaconHeight, inst, currentPortalState, portalParams, updatingInfoByTokenID)
		// custodian topup v3
		case strconv.Itoa(metadata.PortalCustodianTopupMetaV3):
			err = blockchain.processPortalCustodianTopupV3(portalStateDB, beaconHeight, inst, currentPortalState, portalParams)
		// top up for waiting porting v3
		case strconv.Itoa(metadata.PortalTopUpWaitingPortingRequestMetaV3):
			err = blockchain.processPortalTopUpWaitingPortingV3(portalStateDB, beaconHeight, inst, currentPortalState, portalParams)

		// ============ Reward ============
		// portal reward
		case strconv.Itoa(metadata.PortalRewardMeta), strconv.Itoa(metadata.PortalRewardMetaV3):
			err = blockchain.processPortalReward(portalStateDB, beaconHeight, inst, currentPortalState, portalParams)
		// request withdraw reward
		case strconv.Itoa(metadata.PortalRequestWithdrawRewardMeta):
			err = blockchain.processPortalWithdrawReward(portalStateDB, beaconHeight, inst, currentPortalState, portalParams)
		// total custodian reward instruction
		case strconv.Itoa(metadata.PortalTotalRewardCustodianMeta):
			err = blockchain.processPortalTotalCustodianReward(portalStateDB, beaconHeight, inst, currentPortalState, portalParams)

		// ============ Portal smart contract ============
		// todo: add more metadata need to unlock token from sc
		case strconv.Itoa(metadata.PortalCustodianWithdrawConfirmMetaV3),
			strconv.Itoa(metadata.PortalRedeemFromLiquidationPoolConfirmMetaV3),
			strconv.Itoa(metadata.PortalLiquidateRunAwayCustodianConfirmMetaV3):
			err = blockchain.processPortalConfirmWithdrawInstV3(portalStateDB, beaconHeight, inst, currentPortalState, portalParams)
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

func buildPortalCustodianDepositActionV3(
	remoteAddress map[string]string,
	blockHash eCommon.Hash,
	txIndex uint,
	proofStrs []string,
	shardID byte,
) []string {
	data := metadata.PortalCustodianDepositV3{
		MetadataBase: metadata.MetadataBase{
			Type: metadata.PortalCustodianDepositMetaV3,
		},
		RemoteAddresses: remoteAddress,
		BlockHash:       blockHash,
		TxIndex:         txIndex,
		ProofStrs:       proofStrs,
	}

	actionContent := metadata.PortalCustodianDepositActionV3{
		Meta:    data,
		TxReqID: common.Hash{},
		ShardID: shardID,
	}
	actionContentBytes, _ := json.Marshal(actionContent)
	actionContentBase64Str := base64.StdEncoding.EncodeToString(actionContentBytes)
	return []string{strconv.Itoa(metadata.PortalCustodianDepositMetaV3), actionContentBase64Str}
}

func buildPortalCustodianWithdrawActionV3(
	custodianIncAddress string,
	custodianExternalAddress string,
	externalTokenID string,
	amount uint64,
	shardID byte,
) []string {
	data := metadata.PortalCustodianWithdrawRequestV3{
		MetadataBase: metadata.MetadataBase{
			Type: metadata.PortalCustodianWithdrawRequestMetaV3,
		},
		CustodianIncAddress:      custodianIncAddress,
		CustodianExternalAddress: custodianExternalAddress,
		ExternalTokenID:          externalTokenID,
		Amount:                   amount,
	}

	actionContent := metadata.PortalCustodianWithdrawRequestActionV3{
		Meta:    data,
		TxReqID: common.Hash{},
		ShardID: shardID,
	}
	actionContentBytes, _ := json.Marshal(actionContent)
	actionContentBase64Str := base64.StdEncoding.EncodeToString(actionContentBytes)
	return []string{strconv.Itoa(metadata.PortalCustodianWithdrawRequestMetaV3), actionContentBase64Str}
}

func buildPortalUserRegisterAction(
	portingID string,
	incAddressStr string,
	pTokenID string,
	portingAmount uint64,
	portingFee uint64,
	shardID byte,
	shardHeight uint64,
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
		Meta:        data,
		TxReqID:     common.Hash{},
		ShardID:     shardID,
		ShardHeight: shardHeight,
	}
	actionContentBytes, _ := json.Marshal(actionContent)
	actionContentBase64Str := base64.StdEncoding.EncodeToString(actionContentBytes)
	return []string{strconv.Itoa(metadata.PortalUserRegisterMeta), actionContentBase64Str}
}

func buildPortalUserReqPTokenAction(
	portingID string,
	incAddressStr string,
	pTokenID string,
	portingAmount uint64,
	portingProof string,
	shardID byte,
) []string {
	data := metadata.PortalRequestPTokens{
		MetadataBase: metadata.MetadataBase{
			Type: metadata.PortalUserRequestPTokenMeta,
		},
		UniquePortingID: portingID,
		TokenID:         pTokenID,
		IncogAddressStr: incAddressStr,
		PortingAmount:   portingAmount,
		PortingProof:    portingProof,
	}

	actionContent := metadata.PortalRequestPTokensAction{
		Meta:    data,
		TxReqID: common.Hash{},
		ShardID: shardID,
	}
	actionContentBytes, _ := json.Marshal(actionContent)
	actionContentBase64Str := base64.StdEncoding.EncodeToString(actionContentBytes)
	return []string{strconv.Itoa(metadata.PortalUserRequestPTokenMeta), actionContentBase64Str}
}

func buildPortalTopupCustodianAction(
	incAddressStr string,
	ptokenID string,
	depositAmount uint64,
	shardID byte,
	freeCollateralAmount uint64,
) []string {
	data := metadata.PortalLiquidationCustodianDepositV2{
		MetadataBase: metadata.MetadataBase{
			Type: metadata.PortalCustodianTopupMetaV2,
		},
		IncogAddressStr:      incAddressStr,
		PTokenId:             ptokenID,
		DepositedAmount:      depositAmount,
		FreeCollateralAmount: freeCollateralAmount,
	}

	actionContent := metadata.PortalLiquidationCustodianDepositActionV2{
		Meta:    data,
		TxReqID: common.Hash{},
		ShardID: shardID,
	}
	actionContentBytes, _ := json.Marshal(actionContent)
	actionContentBase64Str := base64.StdEncoding.EncodeToString(actionContentBytes)
	return []string{strconv.Itoa(metadata.PortalCustodianTopupMetaV2), actionContentBase64Str}
}

func buildTopupWaitingPortingAction(
	incAddressStr string,
	portingID string,
	ptokenID string,
	depositAmount uint64,
	shardID byte,
	freeCollateralAmount uint64,
) []string {
	data := metadata.PortalTopUpWaitingPortingRequest{
		MetadataBase: metadata.MetadataBase{
			Type: metadata.PortalTopUpWaitingPortingRequestMeta,
		},
		IncogAddressStr:      incAddressStr,
		PortingID:            portingID,
		PTokenID:             ptokenID,
		DepositedAmount:      depositAmount,
		FreeCollateralAmount: freeCollateralAmount,
	}

	actionContent := metadata.PortalTopUpWaitingPortingRequestAction{
		Meta:    data,
		TxReqID: common.Hash{},
		ShardID: shardID,
	}
	actionContentBytes, _ := json.Marshal(actionContent)
	actionContentBase64Str := base64.StdEncoding.EncodeToString(actionContentBytes)
	return []string{strconv.Itoa(metadata.PortalTopUpWaitingPortingRequestMeta), actionContentBase64Str}
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

//func (s *PortalTestSuiteV3) TestRelayExchangeRate() {
//	fmt.Println("Running TestRelayExchangeRate - beacon height 999 ...")
//	bc := s.blockChain
//	pm := NewPortalManager()
//	beaconHeight := uint64(999)
//	shardID := byte(0)
//	updatingInfoByTokenID := map[common.Hash]UpdatingInfo{}
//
//	// build test cases
//	testcases := []TestCaseRelayExchangeRate{
//		// valid
//		{
//			senderAddressStr: "feeder1",
//			rates: []*metadata.ExchangeRateInfo{
//				{
//					PTokenID: common.PRVIDStr,
//					Rate:     1000000,
//				},
//				{
//					PTokenID: common.PortalBNBIDStr,
//					Rate:     20000000,
//				},
//				{
//					PTokenID: common.PortalBTCIDStr,
//					Rate:     10000000000,
//				},
//			},
//		},
//	}
//
//	// build actions from testcases
//	insts := buildPortalExchangeRateActionsFromTcs(testcases, shardID)
//
//	// producer instructions
//	newInsts, err := producerPortalInstructions(
//		bc, beaconHeight, insts, s.sdb, &s.currentPortalStateForProducer, s.blockChain.GetPortalParams(0), shardID, pm)
//
//	// process new instructions
//	err = processPortalInstructions(
//		bc, beaconHeight, newInsts, s.sdb, &s.currentPortalStateForProcess, s.blockChain.GetPortalParams(0), updatingInfoByTokenID)
//
//	// check results
//	s.Equal(1, len(newInsts))
//	s.Equal(nil, err)
//
//	//exchangeRateKey := statedb.GeneratePortalFinalExchangeRatesStateObjectKey().String()
//
//	finalExchangeRate := statedb.NewFinalExchangeRatesStateWithValue(
//		map[string]statedb.FinalExchangeRatesDetail{
//			common.PRVIDStr:       {Amount: 1000000},
//			common.PortalBNBIDStr: {Amount: 20000000},
//			common.PortalBTCIDStr: {Amount: 10000000000},
//		})
//
//	s.Equal(finalExchangeRate, s.currentPortalStateForProcess.FinalExchangeRatesState)
//}

/*
	Feature 1: Custodians deposit collateral (PRV)
*/

type TestCaseCustodianDeposit struct {
	custodianIncAddress string
	remoteAddress       map[string]string
	depositAmount       uint64
}

type ExpectedResultCustodianDeposit struct {
	custodianPool  map[string]*statedb.CustodianState
	numBeaconInsts uint
}

func buildTestCaseAndExpectedResultCustodianDeposit() ([]TestCaseCustodianDeposit, *ExpectedResultCustodianDeposit) {
	testcases := []TestCaseCustodianDeposit{
		// valid
		{
			custodianIncAddress: CUS_INC_ADDRESS_1,
			remoteAddress: map[string]string{
				common.PortalBNBIDStr: CUS_BNB_ADDRESS_1,
				common.PortalBTCIDStr: CUS_BTC_ADDRESS_1,
			},
			depositAmount: 5000 * 1e9,
		},
		// custodian deposit more with new remote addresses
		// expect don't change to new remote addresses,
		// custodian is able to update new remote addresses when total collaterals is empty
		{
			custodianIncAddress: CUS_INC_ADDRESS_1,
			remoteAddress: map[string]string{
				common.PortalBNBIDStr: CUS_BNB_ADDRESS_2,
				common.PortalBTCIDStr: CUS_BTC_ADDRESS_2,
			},
			depositAmount: 2000 * 1e9,
		},
		// new custodian supply only bnb address
		{
			custodianIncAddress: CUS_INC_ADDRESS_2,
			remoteAddress: map[string]string{
				common.PortalBNBIDStr: CUS_BNB_ADDRESS_2,
			},
			depositAmount: 1000 * 1e9,
		},
		// new custodian supply only btc address
		{
			custodianIncAddress: CUS_INC_ADDRESS_3,
			remoteAddress: map[string]string{
				common.PortalBTCIDStr: CUS_BTC_ADDRESS_3,
			},
			depositAmount: 10000 * 1e9,
		},
	}

	// build expected results
	custodianKey1 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_1).String()
	custodianKey2 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_2).String()
	custodianKey3 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_3).String()

	custodian1 := statedb.NewCustodianState()
	custodian1.SetIncognitoAddress(CUS_INC_ADDRESS_1)
	custodian1.SetRemoteAddresses(
		map[string]string{
			common.PortalBNBIDStr: CUS_BNB_ADDRESS_1,
			common.PortalBTCIDStr: CUS_BTC_ADDRESS_1,
		})
	custodian1.SetTotalCollateral(7000 * 1e9)
	custodian1.SetFreeCollateral(7000 * 1e9)

	custodian2 := statedb.NewCustodianState()
	custodian2.SetIncognitoAddress(CUS_INC_ADDRESS_2)
	custodian2.SetRemoteAddresses(
		map[string]string{
			common.PortalBNBIDStr: CUS_BNB_ADDRESS_2,
		})
	custodian2.SetTotalCollateral(1000 * 1e9)
	custodian2.SetFreeCollateral(1000 * 1e9)

	custodian3 := statedb.NewCustodianState()
	custodian3.SetIncognitoAddress(CUS_INC_ADDRESS_3)
	custodian3.SetRemoteAddresses(
		map[string]string{
			common.PortalBTCIDStr: CUS_BTC_ADDRESS_3,
		})
	custodian3.SetTotalCollateral(10000 * 1e9)
	custodian3.SetFreeCollateral(10000 * 1e9)

	expectedRes := &ExpectedResultCustodianDeposit{
		custodianPool: map[string]*statedb.CustodianState{
			custodianKey1: custodian1,
			custodianKey2: custodian2,
			custodianKey3: custodian3,
		},
		numBeaconInsts: 4,
	}

	return testcases, expectedRes
}

func buildCustodianDepositActionsFromTcs(tcs []TestCaseCustodianDeposit, shardID byte) []instructionForProducer {
	insts := []instructionForProducer{}

	for _, tc := range tcs {
		inst := buildPortalCustodianDepositAction(tc.custodianIncAddress, tc.remoteAddress, tc.depositAmount, shardID)
		insts = append(insts, instructionForProducer{
			inst:         inst,
			optionalData: nil,
		})
	}

	return insts
}

func (s *PortalTestSuiteV3) TestCustodianDepositCollateral() {
	fmt.Println("Running TestCustodianDepositCollateral - beacon height 1000 ...")
	bc := s.blockChain
	pm := NewPortalManager()
	beaconHeight := uint64(1000)
	shardID := byte(0)
	updatingInfoByTokenID := map[common.Hash]UpdatingInfo{}

	// build test cases
	testcases, expectedRes := buildTestCaseAndExpectedResultCustodianDeposit()

	// build actions from testcases
	instsForProducer := buildCustodianDepositActionsFromTcs(testcases, shardID)

	// producer instructions
	newInsts, err := producerPortalInstructions(
		bc, beaconHeight, instsForProducer, &s.currentPortalStateForProducer, s.blockChain.GetPortalParams(0), shardID, pm)

	// process new instructions
	err = processPortalInstructions(
		bc, beaconHeight, newInsts, s.sdb, &s.currentPortalStateForProcess, s.blockChain.GetPortalParams(0), updatingInfoByTokenID)

	// check results
	s.Equal(expectedRes.numBeaconInsts, uint(len(newInsts)))
	s.Equal(nil, err)
	s.Equal(expectedRes.custodianPool, s.currentPortalStateForProducer.CustodianPoolState)

	s.Equal(s.currentPortalStateForProcess, s.currentPortalStateForProducer)
}

/*
	Feature 2: Custodians deposit collateral (ETH/ERC20)
*/

type TestCaseCustodianDepositV3 struct {
	custodianIncAddress string
	remoteAddress       map[string]string
	depositAmount       uint64
	collateralTokenID   string
	blockHash           eCommon.Hash
	txIndex             uint
	proofStrs           []string
}

type ExpectedResultCustodianDepositV3 struct {
	custodianPool  map[string]*statedb.CustodianState
	numBeaconInsts uint
}

func buildTestCaseAndExpectedResultCustodianDepositV3() ([]TestCaseCustodianDepositV3, *ExpectedResultCustodianDepositV3) {
	// build test cases
	//todo: build the eth proof
	testcases := []TestCaseCustodianDepositV3{
		// valid
		{
			custodianIncAddress: CUS_INC_ADDRESS_1,
			remoteAddress: map[string]string{
				common.PortalBNBIDStr: CUS_BNB_ADDRESS_1,
				common.PortalBTCIDStr: CUS_BTC_ADDRESS_1,
			},
			depositAmount:     10 * 1e9,
			collateralTokenID: ETH_ID,
			blockHash:         eCommon.Hash{},
			txIndex:           0,
			proofStrs:         nil,
		},
		// the existed custodian deposits other token more with new remote addresses
		// expect don't change to new remote addresses,
		// custodian is able to update new remote addresses when total collaterals is empty
		{
			custodianIncAddress: CUS_INC_ADDRESS_1,
			remoteAddress: map[string]string{
				common.PortalBNBIDStr: CUS_BNB_ADDRESS_2,
				common.PortalBTCIDStr: CUS_BTC_ADDRESS_2,
			},
			depositAmount:     500 * 1e6,
			collateralTokenID: USDT_ID,
			blockHash:         eCommon.Hash{},
			txIndex:           0,
			proofStrs:         nil,
		},
		// invalid: submit the used proof
		{
			custodianIncAddress: CUS_INC_ADDRESS_1,
			remoteAddress: map[string]string{
				common.PortalBNBIDStr: CUS_BNB_ADDRESS_1,
				common.PortalBTCIDStr: CUS_BTC_ADDRESS_1,
			},
			depositAmount:     10 * 1e9,
			collateralTokenID: ETH_ID,
			blockHash:         eCommon.Hash{},
			txIndex:           0,
			proofStrs:         nil,
		},
		// new custodian deposit ERC20 (USDT)
		{
			custodianIncAddress: CUS_INC_ADDRESS_2,
			remoteAddress: map[string]string{
				common.PortalBNBIDStr: CUS_BNB_ADDRESS_2,
			},
			depositAmount:     2000 * 1e6,
			collateralTokenID: USDT_ID,
			blockHash:         eCommon.Hash{},
			txIndex:           0,
			proofStrs:         nil,
		},
		// invalid: collateral tokenID is not supported
		{
			custodianIncAddress: CUS_INC_ADDRESS_3,
			remoteAddress: map[string]string{
				common.PortalBTCIDStr: CUS_BTC_ADDRESS_3,
			},
			depositAmount:     10000 * 1e9,
			collateralTokenID: "0x0050f638abfb0e5dfd794933ffff3b3350ebb6f4",
			blockHash:         eCommon.Hash{},
			txIndex:           0,
			proofStrs:         nil,
		},
	}

	// build expected results
	custodianKey1 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_1).String()
	custodianKey2 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_2).String()

	custodian1 := statedb.NewCustodianState()
	custodian1.SetIncognitoAddress(CUS_INC_ADDRESS_1)
	custodian1.SetRemoteAddresses(
		map[string]string{
			common.PortalBNBIDStr: CUS_BNB_ADDRESS_1,
			common.PortalBTCIDStr: CUS_BTC_ADDRESS_1,
		})
	custodian1.SetTotalTokenCollaterals(
		map[string]uint64{
			ETH_ID:  10 * 1e9,
			USDT_ID: 500 * 1e6,
		})
	custodian1.SetFreeTokenCollaterals(
		map[string]uint64{
			ETH_ID:  10 * 1e9,
			USDT_ID: 500 * 1e6,
		})

	custodian2 := statedb.NewCustodianState()
	custodian2.SetIncognitoAddress(CUS_INC_ADDRESS_2)
	custodian2.SetRemoteAddresses(
		map[string]string{
			common.PortalBNBIDStr: CUS_BNB_ADDRESS_2,
		})
	custodian2.SetTotalTokenCollaterals(
		map[string]uint64{
			USDT_ID: 2000 * 1e6,
		})
	custodian2.SetFreeTokenCollaterals(
		map[string]uint64{
			USDT_ID: 2000 * 1e6,
		})

	expectedRes := &ExpectedResultCustodianDepositV3{
		custodianPool: map[string]*statedb.CustodianState{
			custodianKey1: custodian1,
			custodianKey2: custodian2,
		},
		numBeaconInsts: 5,
	}

	return testcases, expectedRes
}

func buildCustodianDepositActionsV3FromTcs(tcs []TestCaseCustodianDepositV3, shardID byte) []instructionForProducer {
	insts := []instructionForProducer{}

	for _, tc := range tcs {
		inst := buildPortalCustodianDepositActionV3(tc.remoteAddress, tc.blockHash, tc.txIndex, tc.proofStrs, shardID)
		insts = append(insts, instructionForProducer{
			inst:         inst,
			optionalData: nil,
		})
	}

	return insts
}

//todo:
func (s *PortalTestSuiteV3) TestCustodianDepositCollateralV3() {
	return
	fmt.Println("Running TestCustodianDepositCollateral - beacon height 1000 ...")
	bc := s.blockChain
	pm := NewPortalManager()
	beaconHeight := uint64(1000)
	shardID := byte(0)
	updatingInfoByTokenID := map[common.Hash]UpdatingInfo{}

	// build test cases and expected results
	testcases, expectedRes := buildTestCaseAndExpectedResultCustodianDepositV3()

	// build actions from testcases
	instsForProducer := buildCustodianDepositActionsV3FromTcs(testcases, shardID)

	// producer instructions
	newInsts, err := producerPortalInstructions(
		bc, beaconHeight, instsForProducer, &s.currentPortalStateForProducer, s.blockChain.GetPortalParams(0), shardID, pm)

	// process new instructions
	err = processPortalInstructions(
		bc, beaconHeight, newInsts, s.sdb, &s.currentPortalStateForProcess, s.blockChain.GetPortalParams(0), updatingInfoByTokenID)

	// check results
	s.Equal(expectedRes.numBeaconInsts, uint(len(newInsts)))
	s.Equal(nil, err)
	s.Equal(expectedRes.custodianPool, s.currentPortalStateForProducer.CustodianPoolState)

	s.Equal(s.currentPortalStateForProcess, s.currentPortalStateForProducer)
}

/*
	Feature 2: Custodians withdraw free collaterals (ETH/ERC20)
*/

type TestCaseCustodianWithdrawV3 struct {
	custodianIncAddress      string
	custodianExternalAddress string
	externalTokenID          string
	amount                   uint64
}

type ExpectedResultCustodianWithdrawV3 struct {
	custodianPool  map[string]*statedb.CustodianState
	numBeaconInsts uint
}

func (s *PortalTestSuiteV3) SetupTestCustodianWithdrawV3() {
	custodianKey1 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_1).String()
	custodianKey2 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_2).String()
	custodianKey3 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_3).String()

	custodian1 := statedb.NewCustodianState()
	custodian1.SetIncognitoAddress(CUS_INC_ADDRESS_1)
	custodian1.SetRemoteAddresses(
		map[string]string{
			common.PortalBNBIDStr: CUS_BNB_ADDRESS_1,
			common.PortalBTCIDStr: CUS_BTC_ADDRESS_1,
		})
	custodian1.SetTotalCollateral(1000 * 1e9)
	custodian1.SetFreeCollateral(1000 * 1e9)
	custodian1.SetTotalTokenCollaterals(
		map[string]uint64{
			ETH_ID:  10 * 1e9,
			USDT_ID: 500 * 1e6,
		})
	custodian1.SetFreeTokenCollaterals(
		map[string]uint64{
			ETH_ID:  10 * 1e9,
			USDT_ID: 500 * 1e6,
		})

	custodian2 := statedb.NewCustodianState()
	custodian2.SetIncognitoAddress(CUS_INC_ADDRESS_2)
	custodian2.SetRemoteAddresses(
		map[string]string{
			common.PortalBNBIDStr: CUS_BNB_ADDRESS_2,
		})
	custodian2.SetTotalTokenCollaterals(
		map[string]uint64{
			USDT_ID: 2000 * 1e6,
		})
	custodian2.SetFreeTokenCollaterals(
		map[string]uint64{
			USDT_ID: 2000 * 1e6,
		})

	custodian3 := statedb.NewCustodianState()
	custodian3.SetIncognitoAddress(CUS_INC_ADDRESS_3)
	custodian3.SetRemoteAddresses(
		map[string]string{
			common.PortalBTCIDStr: CUS_BTC_ADDRESS_2,
		})
	custodian3.SetTotalCollateral(1000 * 1e9)
	custodian3.SetFreeCollateral(1000 * 1e9)

	custodianPool := map[string]*statedb.CustodianState{
		custodianKey1: custodian1,
		custodianKey2: custodian2,
		custodianKey3: custodian3,
	}

	s.currentPortalStateForProducer.CustodianPoolState = custodianPool
	s.currentPortalStateForProcess.CustodianPoolState = cloneCustodians(custodianPool)
}

func buildTestCaseAndExpectedResultCustodianWithdrawV3() ([]TestCaseCustodianWithdrawV3, *ExpectedResultCustodianWithdrawV3) {
	// build test cases
	//todo: build the eth proof
	testcases := []TestCaseCustodianWithdrawV3{
		// valid: withdrawal amount less than free collateral amount (ETH)
		{
			custodianIncAddress:      CUS_INC_ADDRESS_1,
			custodianExternalAddress: CUS_ETH_ADDRESS_1,
			externalTokenID:          ETH_ID,
			amount:                   1 * 1e9,
		},
		// valid: withdrawal amount equal to free collateral amount (ERC20)
		{
			custodianIncAddress:      CUS_INC_ADDRESS_1,
			custodianExternalAddress: CUS_ETH_ADDRESS_1,
			externalTokenID:          USDT_ID,
			amount:                   500 * 1e6,
		},
		// invalid: withdrawal amount greater than free collateral amount
		{
			custodianIncAddress:      CUS_INC_ADDRESS_2,
			custodianExternalAddress: CUS_ETH_ADDRESS_2,
			externalTokenID:          USDT_ID,
			amount:                   3000 * 1e6,
		},
		// invalid: invalid collateral tokenID
		{
			custodianIncAddress:      CUS_INC_ADDRESS_2,
			custodianExternalAddress: CUS_ETH_ADDRESS_2,
			externalTokenID:          ETH_ID,
			amount:                   1 * 1e9,
		},
	}

	// build expected results
	custodianKey1 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_1).String()
	custodianKey2 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_2).String()
	custodianKey3 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_3).String()

	custodian1 := statedb.NewCustodianState()
	custodian1.SetIncognitoAddress(CUS_INC_ADDRESS_1)
	custodian1.SetRemoteAddresses(
		map[string]string{
			common.PortalBNBIDStr: CUS_BNB_ADDRESS_1,
			common.PortalBTCIDStr: CUS_BTC_ADDRESS_1,
		})
	custodian1.SetTotalCollateral(1000 * 1e9)
	custodian1.SetFreeCollateral(1000 * 1e9)
	custodian1.SetTotalTokenCollaterals(
		map[string]uint64{
			ETH_ID:  9 * 1e9,
			USDT_ID: 0 * 1e6,
		})
	custodian1.SetFreeTokenCollaterals(
		map[string]uint64{
			ETH_ID:  9 * 1e9,
			USDT_ID: 0 * 1e6,
		})

	custodian2 := statedb.NewCustodianState()
	custodian2.SetIncognitoAddress(CUS_INC_ADDRESS_2)
	custodian2.SetRemoteAddresses(
		map[string]string{
			common.PortalBNBIDStr: CUS_BNB_ADDRESS_2,
		})
	custodian2.SetTotalTokenCollaterals(
		map[string]uint64{
			USDT_ID: 2000 * 1e6,
		})
	custodian2.SetFreeTokenCollaterals(
		map[string]uint64{
			USDT_ID: 2000 * 1e6,
		})

	custodian3 := statedb.NewCustodianState()
	custodian3.SetIncognitoAddress(CUS_INC_ADDRESS_3)
	custodian3.SetRemoteAddresses(
		map[string]string{
			common.PortalBTCIDStr: CUS_BTC_ADDRESS_2,
		})
	custodian3.SetTotalCollateral(1000 * 1e9)
	custodian3.SetFreeCollateral(1000 * 1e9)

	expectedRes := &ExpectedResultCustodianWithdrawV3{
		custodianPool: map[string]*statedb.CustodianState{
			custodianKey1: custodian1,
			custodianKey2: custodian2,
			custodianKey3: custodian3,
		},
		numBeaconInsts: 6,
	}
	// todo: check confirm withdraw instructions

	return testcases, expectedRes
}

func buildCustodianWithdrawActionsV3FromTcs(tcs []TestCaseCustodianWithdrawV3, shardID byte) []instructionForProducer {
	insts := []instructionForProducer{}

	for _, tc := range tcs {
		inst := buildPortalCustodianWithdrawActionV3(tc.custodianIncAddress, tc.custodianExternalAddress, tc.externalTokenID, tc.amount, shardID)
		insts = append(insts, instructionForProducer{
			inst:         inst,
			optionalData: nil,
		})
	}

	return insts
}

func (s *PortalTestSuiteV3) TestCustodianWithdrawCollateralV3() {
	fmt.Println("Running TestCustodianWithdrawCollateralV3 - beacon height 1000 ...")
	bc := s.blockChain
	pm := NewPortalManager()
	beaconHeight := uint64(1000)
	shardID := byte(0)
	updatingInfoByTokenID := map[common.Hash]UpdatingInfo{}

	s.SetupTestCustodianWithdrawV3()

	// build test cases and expected results
	testcases, expectedRes := buildTestCaseAndExpectedResultCustodianWithdrawV3()

	// build actions from testcases
	instsForProducer := buildCustodianWithdrawActionsV3FromTcs(testcases, shardID)

	// producer instructions
	newInsts, err := producerPortalInstructions(
		bc, beaconHeight, instsForProducer, &s.currentPortalStateForProducer, s.blockChain.GetPortalParams(0), shardID, pm)

	// process new instructions
	err = processPortalInstructions(
		bc, beaconHeight, newInsts, s.sdb, &s.currentPortalStateForProcess, s.blockChain.GetPortalParams(0), updatingInfoByTokenID)

	// check results
	s.Equal(expectedRes.numBeaconInsts, uint(len(newInsts)))
	s.Equal(nil, err)
	s.Equal(expectedRes.custodianPool, s.currentPortalStateForProducer.CustodianPoolState)

	s.Equal(s.currentPortalStateForProcess, s.currentPortalStateForProducer)
}

/*
	Feature 3: Users create porting request
*/
type TestCaseRequestPorting struct {
	portingID     string
	incAddressStr string
	pTokenID      string
	portingAmount uint64
	portingFee    uint64
	isExisted     bool
}

type ExpectedResultPortingRequest struct {
	waitingPortingRes map[string]*statedb.WaitingPortingRequest
	custodianPool     map[string]*statedb.CustodianState
	numBeaconInsts    uint
}

func (s *PortalTestSuiteV3) SetupTestPortingRequest() {
	custodianKey1 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_1).String()
	custodianKey2 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_2).String()
	custodianKey3 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_3).String()

	custodian1 := statedb.NewCustodianState()
	custodian1.SetIncognitoAddress(CUS_INC_ADDRESS_1)
	custodian1.SetRemoteAddresses(
		map[string]string{
			common.PortalBNBIDStr: CUS_BNB_ADDRESS_1,
			common.PortalBTCIDStr: CUS_BTC_ADDRESS_1,
		})
	custodian1.SetTotalCollateral(1000 * 1e9)
	custodian1.SetFreeCollateral(1000 * 1e9)
	custodian1.SetTotalTokenCollaterals(
		map[string]uint64{
			ETH_ID:  10 * 1e9,
			USDT_ID: 500 * 1e6,
		})
	custodian1.SetFreeTokenCollaterals(
		map[string]uint64{
			ETH_ID:  10 * 1e9,
			USDT_ID: 500 * 1e6,
		})

	custodian2 := statedb.NewCustodianState()
	custodian2.SetIncognitoAddress(CUS_INC_ADDRESS_2)
	custodian2.SetRemoteAddresses(
		map[string]string{
			common.PortalBNBIDStr: CUS_BNB_ADDRESS_2,
		})
	custodian2.SetTotalTokenCollaterals(
		map[string]uint64{
			USDT_ID: 2000 * 1e6,
		})
	custodian2.SetFreeTokenCollaterals(
		map[string]uint64{
			USDT_ID: 2000 * 1e6,
		})

	custodian3 := statedb.NewCustodianState()
	custodian3.SetIncognitoAddress(CUS_INC_ADDRESS_3)
	custodian3.SetRemoteAddresses(
		map[string]string{
			common.PortalBTCIDStr: CUS_BTC_ADDRESS_2,
		})
	custodian3.SetTotalCollateral(1000 * 1e9)
	custodian3.SetFreeCollateral(1000 * 1e9)

	custodianPool := map[string]*statedb.CustodianState{
		custodianKey1: custodian1,
		custodianKey2: custodian2,
		custodianKey3: custodian3,
	}

	s.currentPortalStateForProducer.CustodianPoolState = custodianPool
	s.currentPortalStateForProcess.CustodianPoolState = cloneCustodians(custodianPool)
}

func buildTestCaseAndExpectedResultPortingRequest() ([]TestCaseRequestPorting, *ExpectedResultPortingRequest) {
	beaconHeight := uint64(1001)
	shardHeight := uint64(1001)
	shardID := byte(0)
	// build test cases
	testcases := []TestCaseRequestPorting{
		// valid porting request: match to one custodian, 0.01% porting fee
		{
			portingID:     "porting-bnb-1",
			incAddressStr: USER_INC_ADDRESS_1,
			pTokenID:      common.PortalBNBIDStr,
			portingAmount: 1 * 1e9,
			portingFee:    2000000,
			isExisted:     false,
		},
		// valid porting request: match to many custodians
		{
			portingID:     "porting-bnb-2",
			incAddressStr: USER_INC_ADDRESS_2,
			pTokenID:      common.PortalBNBIDStr,
			portingAmount: 150 * 1e9,
			portingFee:    300000000,
			isExisted:     false,
		},
		// invalid porting request with duplicate porting ID
		{
			portingID:     "porting-bnb-1",
			incAddressStr: USER_INC_ADDRESS_1,
			pTokenID:      common.PortalBNBIDStr,
			portingAmount: 1 * 1e9,
			portingFee:    2000000,
			isExisted:     true,
		},
		// invalid porting request with invalid porting fee
		{
			portingID:     "porting-bnb-3",
			incAddressStr: USER_INC_ADDRESS_3,
			pTokenID:      common.PortalBNBIDStr,
			portingAmount: 1 * 1e9,
			portingFee:    99900,
			isExisted:     false,
		},
		// invalid porting request: total collaterals of the custodians are not enough for the porting amount
		{
			portingID:     "porting-btc-4",
			incAddressStr: USER_INC_ADDRESS_3,
			pTokenID:      common.PortalBTCIDStr,
			portingAmount: 10 * 1e9,
			portingFee:    1000000000,
			isExisted:     false,
		},
	}

	// build expected results
	// custodian state after matching porting requests
	custodianKey1 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_1).String()
	custodianKey2 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_2).String()
	custodianKey3 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_3).String()

	custodian1 := statedb.NewCustodianState()
	custodian1.SetIncognitoAddress(CUS_INC_ADDRESS_1)
	custodian1.SetRemoteAddresses(
		map[string]string{
			common.PortalBNBIDStr: CUS_BNB_ADDRESS_1,
			common.PortalBTCIDStr: CUS_BTC_ADDRESS_1,
		})
	custodian1.SetTotalCollateral(1000 * 1e9)
	custodian1.SetFreeCollateral(0)
	custodian1.SetTotalTokenCollaterals(
		map[string]uint64{
			ETH_ID:  10 * 1e9,
			USDT_ID: 500 * 1e6,
		})
	custodian1.SetFreeTokenCollaterals(
		map[string]uint64{
			ETH_ID:  0,
			USDT_ID: 0,
		})
	custodian1.SetLockedAmountCollateral(
		map[string]uint64{
			common.PortalBNBIDStr: 1000 * 1e9,
		})
	custodian1.SetLockedTokenCollaterals(
		map[string]map[string]uint64{
			common.PortalBNBIDStr: {
				ETH_ID:  10 * 1e9,
				USDT_ID: 500 * 1e6,
			},
		})

	custodian2 := statedb.NewCustodianState()
	custodian2.SetIncognitoAddress(CUS_INC_ADDRESS_2)
	custodian2.SetRemoteAddresses(
		map[string]string{
			common.PortalBNBIDStr: CUS_BNB_ADDRESS_2,
		})
	custodian2.SetTotalTokenCollaterals(
		map[string]uint64{
			USDT_ID: 2000 * 1e6,
		})
	custodian2.SetFreeTokenCollaterals(
		map[string]uint64{
			USDT_ID: 1460 * 1e6,
		})
	custodian2.SetLockedTokenCollaterals(map[string]map[string]uint64{
		common.PortalBNBIDStr: {
			USDT_ID: 540 * 1e6,
		},
	})

	custodian3 := statedb.NewCustodianState()
	custodian3.SetIncognitoAddress(CUS_INC_ADDRESS_3)
	custodian3.SetRemoteAddresses(
		map[string]string{
			common.PortalBTCIDStr: CUS_BTC_ADDRESS_2,
		})
	custodian3.SetTotalCollateral(1000 * 1e9)
	custodian3.SetFreeCollateral(1000 * 1e9)

	// waiting porting requests
	wPortingReqKey1 := statedb.GeneratePortalWaitingPortingRequestObjectKey("porting-bnb-1").String()
	wPortingRequest1 := statedb.NewWaitingPortingRequestWithValue(
		"porting-bnb-1", common.Hash{}, common.PortalBNBIDStr,
		USER_INC_ADDRESS_1, 1*1e9,
		[]*statedb.MatchingPortingCustodianDetail{
			{
				IncAddress:             custodian2.GetIncognitoAddress(),
				RemoteAddress:          custodian2.GetRemoteAddresses()[common.PortalBNBIDStr],
				Amount:                 1 * 1e9,
				LockedAmountCollateral: 0,
				LockedTokenCollaterals: map[string]uint64{
					USDT_ID: 40000000,
				},
			},
		}, 2000000, beaconHeight, shardHeight, shardID)

	wPortingReqKey2 := statedb.GeneratePortalWaitingPortingRequestObjectKey("porting-bnb-2").String()
	wPortingRequest2 := statedb.NewWaitingPortingRequestWithValue(
		"porting-bnb-2", common.Hash{}, common.PortalBNBIDStr,
		USER_INC_ADDRESS_2, 150*1e9,
		[]*statedb.MatchingPortingCustodianDetail{
			{
				IncAddress:             custodian1.GetIncognitoAddress(),
				RemoteAddress:          custodian1.GetRemoteAddresses()[common.PortalBNBIDStr],
				Amount:                 137500000000,
				LockedAmountCollateral: 1000000000000,
				LockedTokenCollaterals: map[string]uint64{
					ETH_ID:  10000000000,
					USDT_ID: 500000000,
				},
			},
			{
				IncAddress:             custodian2.GetIncognitoAddress(),
				RemoteAddress:          custodian2.GetRemoteAddresses()[common.PortalBNBIDStr],
				Amount:                 12500000000,
				LockedAmountCollateral: 0,
				LockedTokenCollaterals: map[string]uint64{
					USDT_ID: 500000000,
				},
			},
		}, 300000000, beaconHeight, shardHeight, shardID)

	expectedRes := &ExpectedResultPortingRequest{
		custodianPool: map[string]*statedb.CustodianState{
			custodianKey1: custodian1,
			custodianKey2: custodian2,
			custodianKey3: custodian3,
		},
		waitingPortingRes: map[string]*statedb.WaitingPortingRequest{
			wPortingReqKey1: wPortingRequest1,
			wPortingReqKey2: wPortingRequest2,
		},
		numBeaconInsts: 5,
	}

	return testcases, expectedRes
}

func buildRequestPortingActionsFromTcs(tcs []TestCaseRequestPorting, shardID byte) []instructionForProducer {
	insts := []instructionForProducer{}

	//todo: shardHeight
	for _, tc := range tcs {
		inst := buildPortalUserRegisterAction(
			tc.portingID, tc.incAddressStr, tc.pTokenID, tc.portingAmount, tc.portingFee, shardID, 1001)
		insts = append(insts, instructionForProducer{
			inst:         inst,
			optionalData: map[string]interface{}{"isExistPortingID": tc.isExisted},
		})
	}

	return insts
}

func (s *PortalTestSuiteV3) TestPortingRequest() {
	fmt.Println("Running TestPortingRequest - beacon height 1001 ...")
	bc := s.blockChain
	pm := NewPortalManager()
	beaconHeight := uint64(1001)
	shardID := byte(0)
	updatingInfoByTokenID := map[common.Hash]UpdatingInfo{}

	s.SetupTestPortingRequest()

	// build test cases
	testcases, expectedRes := buildTestCaseAndExpectedResultPortingRequest()

	// build actions from testcases
	instsForProducer := buildRequestPortingActionsFromTcs(testcases, shardID)

	// producer instructions
	newInsts, err := producerPortalInstructions(
		bc, beaconHeight-1, instsForProducer, &s.currentPortalStateForProducer, s.blockChain.GetPortalParams(0), shardID, pm)
	s.Equal(nil, err)

	// process new instructions
	err = processPortalInstructions(
		bc, beaconHeight-1, newInsts, s.sdb, &s.currentPortalStateForProcess, s.blockChain.GetPortalParams(0), updatingInfoByTokenID)

	// check results
	s.Equal(expectedRes.numBeaconInsts, uint(len(newInsts)))
	s.Equal(nil, err)
	s.Equal(expectedRes.custodianPool, s.currentPortalStateForProducer.CustodianPoolState)
	s.Equal(expectedRes.waitingPortingRes, s.currentPortalStateForProducer.WaitingPortingRequests)

	s.Equal(s.currentPortalStateForProcess, s.currentPortalStateForProducer)
}

/*
	Feature 3: Users submit proof to request pTokens after sending public tokens to custodians
*/
type TestCaseRequestPtokens struct {
	portingID     string
	incAddressStr string
	pTokenID      string
	portingAmount uint64

	txID         string
	portingProof string
}

type ExpectedResultRequestPTokens struct {
	waitingPortingRes map[string]*statedb.WaitingPortingRequest
	custodianPool     map[string]*statedb.CustodianState
	numBeaconInsts    uint
	statusInsts       []string
}

func (s *PortalTestSuiteV3) SetupTestRequestPtokens() {
	beaconHeight := uint64(1002)
	shardHeight := uint64(1002)
	shardID := byte(0)

	custodianKey1 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_1).String()
	custodianKey2 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_2).String()
	custodianKey3 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_3).String()

	custodian1 := statedb.NewCustodianState()
	custodian1.SetIncognitoAddress(CUS_INC_ADDRESS_1)
	custodian1.SetRemoteAddresses(
		map[string]string{
			common.PortalBNBIDStr: CUS_BNB_ADDRESS_1,
			common.PortalBTCIDStr: CUS_BTC_ADDRESS_1,
		})
	custodian1.SetTotalCollateral(1000 * 1e9)
	custodian1.SetFreeCollateral(0)
	custodian1.SetTotalTokenCollaterals(
		map[string]uint64{
			ETH_ID:  10 * 1e9,
			USDT_ID: 500 * 1e6,
		})
	custodian1.SetFreeTokenCollaterals(
		map[string]uint64{
			ETH_ID:  0,
			USDT_ID: 0,
		})
	custodian1.SetLockedAmountCollateral(
		map[string]uint64{
			common.PortalBNBIDStr: 1000 * 1e9,
		})
	custodian1.SetLockedTokenCollaterals(
		map[string]map[string]uint64{
			common.PortalBNBIDStr: {
				ETH_ID:  10 * 1e9,
				USDT_ID: 500 * 1e6,
			},
		})

	custodian2 := statedb.NewCustodianState()
	custodian2.SetIncognitoAddress(CUS_INC_ADDRESS_2)
	custodian2.SetRemoteAddresses(
		map[string]string{
			common.PortalBNBIDStr: CUS_BNB_ADDRESS_2,
		})
	custodian2.SetTotalTokenCollaterals(
		map[string]uint64{
			USDT_ID: 2000 * 1e6,
		})
	custodian2.SetFreeTokenCollaterals(
		map[string]uint64{
			USDT_ID: 1460 * 1e6,
		})
	custodian2.SetLockedTokenCollaterals(map[string]map[string]uint64{
		common.PortalBNBIDStr: {
			USDT_ID: 540 * 1e6,
		},
	})

	custodian3 := statedb.NewCustodianState()
	custodian3.SetIncognitoAddress(CUS_INC_ADDRESS_3)
	custodian3.SetRemoteAddresses(
		map[string]string{
			common.PortalBTCIDStr: CUS_BTC_ADDRESS_2,
		})
	custodian3.SetTotalCollateral(1000 * 1e9)
	custodian3.SetFreeCollateral(1000 * 1e9)

	// waiting porting requests
	wPortingReqKey1 := statedb.GeneratePortalWaitingPortingRequestObjectKey("porting-bnb-1").String()
	wPortingRequest1 := statedb.NewWaitingPortingRequestWithValue(
		"porting-bnb-1", common.Hash{}, common.PortalBNBIDStr,
		USER_INC_ADDRESS_1, 1*1e9,
		[]*statedb.MatchingPortingCustodianDetail{
			{
				IncAddress:             custodian2.GetIncognitoAddress(),
				RemoteAddress:          custodian2.GetRemoteAddresses()[common.PortalBNBIDStr],
				Amount:                 1 * 1e9,
				LockedAmountCollateral: 0,
				LockedTokenCollaterals: map[string]uint64{
					USDT_ID: 40000000,
				},
			},
		}, 2000000, beaconHeight, shardHeight, shardID)

	wPortingReqKey2 := statedb.GeneratePortalWaitingPortingRequestObjectKey("porting-bnb-2").String()
	wPortingRequest2 := statedb.NewWaitingPortingRequestWithValue(
		"porting-bnb-2", common.Hash{}, common.PortalBNBIDStr,
		USER_INC_ADDRESS_2, 150*1e9,
		[]*statedb.MatchingPortingCustodianDetail{
			{
				IncAddress:             custodian1.GetIncognitoAddress(),
				RemoteAddress:          custodian1.GetRemoteAddresses()[common.PortalBNBIDStr],
				Amount:                 137500000000,
				LockedAmountCollateral: 1000000000000,
				LockedTokenCollaterals: map[string]uint64{
					ETH_ID:  10000000000,
					USDT_ID: 500000000,
				},
			},
			{
				IncAddress:             custodian2.GetIncognitoAddress(),
				RemoteAddress:          custodian2.GetRemoteAddresses()[common.PortalBNBIDStr],
				Amount:                 12500000000,
				LockedAmountCollateral: 0,
				LockedTokenCollaterals: map[string]uint64{
					USDT_ID: 500000000,
				},
			},
		}, 300000000, beaconHeight, shardHeight, shardID)

	custodians := map[string]*statedb.CustodianState{
		custodianKey1: custodian1,
		custodianKey2: custodian2,
		custodianKey3: custodian3,
	}

	wPortingRequests := map[string]*statedb.WaitingPortingRequest{
		wPortingReqKey1: wPortingRequest1,
		wPortingReqKey2: wPortingRequest2,
	}

	s.currentPortalStateForProducer.CustodianPoolState = custodians
	s.currentPortalStateForProducer.WaitingPortingRequests = wPortingRequests

	s.currentPortalStateForProcess.CustodianPoolState = cloneCustodians(custodians)
	s.currentPortalStateForProcess.WaitingPortingRequests = cloneWPortingRequests(wPortingRequests)
}

func buildTestCaseAndExpectedResultRequestPTokens() ([]TestCaseRequestPtokens, *ExpectedResultRequestPTokens) {
	// build test cases
	testcases := []TestCaseRequestPtokens{
		// valid request ptokens: porting request matching one custodian
		{
			portingID:     "porting-bnb-1",
			incAddressStr: USER_INC_ADDRESS_1,
			pTokenID:      common.PortalBNBIDStr,
			portingAmount: 1 * 1e9,
			txID:          "A71E509490912AA82E138361D63253601C4CD20F9508F6880738FEF495CE42C6",
			portingProof:  "",
		},
		// valid request ptokens: porting request matching many custodians
		{
			portingID:     "porting-bnb-2",
			incAddressStr: USER_INC_ADDRESS_2,
			pTokenID:      common.PortalBNBIDStr,
			portingAmount: 150 * 1e9,
			txID:          "37B26ADEFB57534C9F38C07D948B8241099E02A0D3CAFA60B4400BD200197B52",
			portingProof:  "",
		},
		// invalid request ptokens: resubmit porting proof
		{
			portingID:     "porting-bnb-1",
			incAddressStr: USER_INC_ADDRESS_1,
			pTokenID:      common.PortalBNBIDStr,
			portingAmount: 1 * 1e9,
			txID:          "A71E509490912AA82E138361D63253601C4CD20F9508F6880738FEF495CE42C6",
			portingProof:  "",
		},
		// invalid request ptokens: invalid porting proof
		{
			portingID:     "porting-bnb-1",
			incAddressStr: USER_INC_ADDRESS_1,
			pTokenID:      common.PortalBNBIDStr,
			portingAmount: 1 * 1e9,
			txID:          "B186FDFBB781D118E25FB37E85BEB83865C3C575485DC43C28220DD7F91BD70A",
			portingProof:  "",
		},
	}

	// build porting proof for testcases
	for i, tc := range testcases {
		proof, err := bnb.BuildProofFromTxID(tc.txID, BNB_NODE_URL)
		if err != nil {
			fmt.Errorf("err build proof: %v", err)
		}
		fmt.Printf("Proof: %v\n", proof)

		testcases[i].portingProof = proof
	}

	// build expected results
	// custodian state after matching porting requests
	custodianKey1 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_1).String()
	custodianKey2 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_2).String()
	custodianKey3 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_3).String()

	custodian1 := statedb.NewCustodianState()
	custodian1.SetIncognitoAddress(CUS_INC_ADDRESS_1)
	custodian1.SetRemoteAddresses(
		map[string]string{
			common.PortalBNBIDStr: CUS_BNB_ADDRESS_1,
			common.PortalBTCIDStr: CUS_BTC_ADDRESS_1,
		})
	custodian1.SetTotalCollateral(1000 * 1e9)
	custodian1.SetFreeCollateral(0)
	custodian1.SetTotalTokenCollaterals(
		map[string]uint64{
			ETH_ID:  10 * 1e9,
			USDT_ID: 500 * 1e6,
		})
	custodian1.SetFreeTokenCollaterals(
		map[string]uint64{
			ETH_ID:  0,
			USDT_ID: 0,
		})
	custodian1.SetLockedAmountCollateral(
		map[string]uint64{
			common.PortalBNBIDStr: 1000 * 1e9,
		})
	custodian1.SetLockedTokenCollaterals(
		map[string]map[string]uint64{
			common.PortalBNBIDStr: {
				ETH_ID:  10 * 1e9,
				USDT_ID: 500 * 1e6,
			},
		})
	custodian1.SetHoldingPublicTokens(
		map[string]uint64{
			common.PortalBNBIDStr: 137500000000,
		})

	custodian2 := statedb.NewCustodianState()
	custodian2.SetIncognitoAddress(CUS_INC_ADDRESS_2)
	custodian2.SetRemoteAddresses(
		map[string]string{
			common.PortalBNBIDStr: CUS_BNB_ADDRESS_2,
		})
	custodian2.SetTotalTokenCollaterals(
		map[string]uint64{
			USDT_ID: 2000 * 1e6,
		})
	custodian2.SetFreeTokenCollaterals(
		map[string]uint64{
			USDT_ID: 1460 * 1e6,
		})
	custodian2.SetLockedTokenCollaterals(
		map[string]map[string]uint64{
			common.PortalBNBIDStr: {
				USDT_ID: 540 * 1e6,
			}})
	custodian2.SetHoldingPublicTokens(
		map[string]uint64{
			common.PortalBNBIDStr: 13500000000,
		})

	custodian3 := statedb.NewCustodianState()
	custodian3.SetIncognitoAddress(CUS_INC_ADDRESS_3)
	custodian3.SetRemoteAddresses(
		map[string]string{
			common.PortalBTCIDStr: CUS_BTC_ADDRESS_2,
		})
	custodian3.SetTotalCollateral(1000 * 1e9)
	custodian3.SetFreeCollateral(1000 * 1e9)

	expectedRes := &ExpectedResultRequestPTokens{
		custodianPool: map[string]*statedb.CustodianState{
			custodianKey1: custodian1,
			custodianKey2: custodian2,
			custodianKey3: custodian3,
		},
		waitingPortingRes: map[string]*statedb.WaitingPortingRequest{},
		numBeaconInsts:    4,
		statusInsts: []string{
			common.PortalReqPTokensAcceptedChainStatus,
			common.PortalReqPTokensAcceptedChainStatus,
			common.PortalReqPTokensRejectedChainStatus,
			common.PortalReqPTokensRejectedChainStatus,
		},
	}

	return testcases, expectedRes
}

func buildRequestPtokensActionsFromTcs(tcs []TestCaseRequestPtokens, shardID byte) []instructionForProducer {
	insts := []instructionForProducer{}

	for _, tc := range tcs {
		inst := buildPortalUserReqPTokenAction(
			tc.portingID, tc.incAddressStr, tc.pTokenID, tc.portingAmount, tc.portingProof, shardID)
		insts = append(insts, instructionForProducer{
			inst:         inst,
			optionalData: nil,
		})
	}

	return insts
}

func (s *PortalTestSuiteV3) TestRequestPtokens() {
	fmt.Println("Running TestRequestPtokens - beacon height 1002 ...")
	bc := s.blockChain
	pm := NewPortalManager()
	beaconHeight := uint64(1002)
	shardID := byte(0)
	updatingInfoByTokenID := map[common.Hash]UpdatingInfo{}

	s.SetupTestRequestPtokens()

	// build test cases
	testcases, expectedResult := buildTestCaseAndExpectedResultRequestPTokens()

	// build actions from testcases
	instsForProducer := buildRequestPtokensActionsFromTcs(testcases, shardID)

	// producer instructions
	newInsts, err := producerPortalInstructions(
		bc, beaconHeight-1, instsForProducer, &s.currentPortalStateForProducer, s.blockChain.GetPortalParams(0), shardID, pm)
	s.Equal(nil, err)

	// process new instructions
	err = processPortalInstructions(
		bc, beaconHeight-1, newInsts, s.sdb, &s.currentPortalStateForProcess, s.blockChain.GetPortalParams(0), updatingInfoByTokenID)

	// check results
	s.Equal(expectedResult.numBeaconInsts, uint(len(newInsts)))
	s.Equal(nil, err)

	for i, inst := range newInsts {
		s.Equal(expectedResult.statusInsts[i], inst[2])
	}

	s.Equal(expectedResult.custodianPool, s.currentPortalStateForProducer.CustodianPoolState)
	s.Equal(expectedResult.waitingPortingRes, s.currentPortalStateForProducer.WaitingPortingRequests)

	s.Equal(s.currentPortalStateForProcess, s.currentPortalStateForProducer)

}

//
///*
//	Feature 4: auto-liquidation: the custodians don't send back public token to the users
//*/
//func (s *PortalTestSuiteV3) SetupTestAutoLiquidation() {
//	custodianKey1 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_1).String()
//	custodianKey2 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_2).String()
//	custodianKey3 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_3).String()
//	custodianKey4 := statedb.GenerateCustodianStateObjectKey("custodianIncAddress4").String()
//
//	custodian1 := statedb.NewCustodianStateWithValue(
//		CUS_INC_ADDRESS_1, 7000*1e9, 6920000000000,
//		map[string]uint64{
//			common.PortalBNBIDStr: 0,
//		},
//		map[string]uint64{
//			common.PortalBNBIDStr: 80000000000,
//		},
//		map[string]string{
//			common.PortalBNBIDStr: CUS_BNB_ADDRESS_1,
//			common.PortalBTCIDStr: CUS_BTC_ADDRESS_1,
//		},
//		map[string]uint64{}, map[string]uint64{}, map[string]uint64{}, map[string]map[string]uint64{})
//
//	custodian2 := statedb.NewCustodianStateWithValue(
//		CUS_INC_ADDRESS_2, 1000*1e9, 960000000000,
//		map[string]uint64{
//			common.PortalBNBIDStr: 0.6 * 1e9,
//		},
//		map[string]uint64{
//			common.PortalBNBIDStr: 40000000000,
//		},
//		map[string]string{
//			common.PortalBNBIDStr: CUS_BNB_ADDRESS_2,
//		},
//		map[string]uint64{}, map[string]uint64{}, map[string]uint64{}, map[string]map[string]uint64{})
//
//	custodian3 := statedb.NewCustodianStateWithValue(
//		CUS_INC_ADDRESS_3, 10000*1e9, 8000000000000,
//		map[string]uint64{
//			common.PortalBTCIDStr: 0.1 * 1e9,
//		},
//		map[string]uint64{
//			common.PortalBTCIDStr: 2000000000000,
//		},
//		map[string]string{
//			common.PortalBTCIDStr: CUS_BTC_ADDRESS_3,
//		},
//		map[string]uint64{}, map[string]uint64{}, map[string]uint64{}, map[string]map[string]uint64{})
//
//	custodian4 := statedb.NewCustodianStateWithValue(
//		"custodianIncAddress4", 5000*1e9, 4960000000000,
//		map[string]uint64{
//		},
//		map[string]uint64{
//			common.PortalBNBIDStr: 40000000000,
//		},
//		map[string]string{
//			common.PortalBNBIDStr: "bnbAddress4",
//		},
//		map[string]uint64{}, map[string]uint64{}, map[string]uint64{}, map[string]map[string]uint64{})
//
//	custodians := map[string]*statedb.CustodianState{
//		custodianKey1: custodian1,
//		custodianKey2: custodian2,
//		custodianKey3: custodian3,
//		custodianKey4: custodian4,
//	}
//
//	redeemReqKey1 := statedb.GenerateMatchedRedeemRequestObjectKey("redeem-bnb-1").String()
//	redeemRequest1 := statedb.NewRedeemRequestWithValue(
//		"redeem-bnb-1", common.PortalBNBIDStr,
//		USER_INC_ADDRESS_1, "userCUS_BNB_ADDRESS_1", 2.3*1e9,
//		[]*statedb.MatchingRedeemCustodianDetail{
//			statedb.NewMatchingRedeemCustodianDetailWithValue(CUS_INC_ADDRESS_1, CUS_BNB_ADDRESS_1, 2*1e9),
//			statedb.NewMatchingRedeemCustodianDetailWithValue(CUS_INC_ADDRESS_2, CUS_BNB_ADDRESS_2, 0.3*1e9),
//		}, 4600000, 1000, common.Hash{}, 0, 1000, "")
//
//	redeemReqKey2 := statedb.GenerateMatchedRedeemRequestObjectKey("redeem-btc-2").String()
//	redeemRequest2 := statedb.NewRedeemRequestWithValue(
//		"redeem-btc-2", common.PortalBTCIDStr,
//		USER_INC_ADDRESS_2, "userCUS_BTC_ADDRESS_2", 0.03*1e9,
//		[]*statedb.MatchingRedeemCustodianDetail{
//			statedb.NewMatchingRedeemCustodianDetailWithValue(CUS_INC_ADDRESS_3, CUS_BTC_ADDRESS_3, 0.03*1e9),
//		}, 30000000, 1500, common.Hash{}, 0, 1000, "")
//
//	matchedRedeemRequest := map[string]*statedb.RedeemRequest{
//		redeemReqKey1: redeemRequest1,
//		redeemReqKey2: redeemRequest2,
//	}
//
//	wRedeemReqKey3 := statedb.GenerateWaitingRedeemRequestObjectKey("redeem-bnb-3").String()
//	wRedeemRequest3 := statedb.NewRedeemRequestWithValue(
//		"redeem-bnb-3", common.PortalBNBIDStr,
//		USER_INC_ADDRESS_1, "userCUS_BNB_ADDRESS_1", 0.1*1e9,
//		[]*statedb.MatchingRedeemCustodianDetail{
//			statedb.NewMatchingRedeemCustodianDetailWithValue(CUS_INC_ADDRESS_2, CUS_BNB_ADDRESS_2, 0.1*1e9),
//		}, 4600000, 1500, common.Hash{}, 0, 1000,  "")
//
//	wRedeemRequests := map[string]*statedb.RedeemRequest{
//		wRedeemReqKey3: wRedeemRequest3,
//	}
//
//	wPortingReqKey1 := statedb.GeneratePortalWaitingPortingRequestObjectKey("porting-bnb-1").String()
//	wPortingRequest1 := statedb.NewWaitingPortingRequestWithValue(
//		"porting-bnb-1", common.Hash{}, common.PortalBNBIDStr,
//		USER_INC_ADDRESS_1, 1*1e9,
//		[]*statedb.MatchingPortingCustodianDetail{
//			{
//				IncAddress:             "custodianIncAddress4",
//				RemoteAddress:          "bnbAddress4",
//				Amount:                 1 * 1e9,
//				LockedAmountCollateral: 40000000000,
//			},
//		}, 2000000, 1500, 1500, 0)
//
//	wPortingRequests := map[string]*statedb.WaitingPortingRequest{
//		wPortingReqKey1: wPortingRequest1,
//	}
//
//	s.currentPortalStateForProducer.CustodianPoolState = custodians
//	s.currentPortalStateForProducer.MatchedRedeemRequests = matchedRedeemRequest
//	s.currentPortalStateForProducer.WaitingRedeemRequests = wRedeemRequests
//	s.currentPortalStateForProducer.WaitingPortingRequests = wPortingRequests
//
//	s.currentPortalStateForProcess.CustodianPoolState = cloneCustodians(custodians)
//	s.currentPortalStateForProcess.MatchedRedeemRequests = cloneRedeemRequests(matchedRedeemRequest)
//	s.currentPortalStateForProcess.WaitingRedeemRequests = cloneRedeemRequests(wRedeemRequests)
//	s.currentPortalStateForProcess.WaitingPortingRequests = cloneWPortingRequests(wPortingRequests)
//}
//
//func (s *PortalTestSuiteV3) TestAutoLiquidationCustodian() {
//	fmt.Println("Running TestAutoLiquidationCustodian - beacon height 3161 ...")
//	bc := s.blockChain
//	beaconHeight := uint64(3161) // ~ after 24 hours from redeem request
//	//shardID := byte(0)
//	//newMatchedRedeemReqIDs := []string{}
//	updatingInfoByTokenID := map[common.Hash]UpdatingInfo{}
//
//	s.SetupTestAutoLiquidation()
//
//	// producer instructions
//	newInsts, err := bc.checkAndBuildInstForCustodianLiquidation(
//		beaconHeight-1, &s.currentPortalStateForProducer, s.blockChain.GetPortalParams(0))
//	s.Equal(nil, err)
//
//	// process new instructions
//	err = processPortalInstructions(
//		bc, beaconHeight-1, newInsts, s.sdb, &s.currentPortalStateForProcess, s.blockChain.GetPortalParams(0), updatingInfoByTokenID)
//
//	// check results
//	s.Equal(2, len(newInsts))
//	s.Equal(nil, err)
//
//	//// remain waiting porting request
//	//redeemReqKey2 := statedb.GenerateMatchedRedeemRequestObjectKey("redeem-btc-2").String()
//	s.Equal(1, len(s.currentPortalStateForProducer.MatchedRedeemRequests))
//	s.Equal(1, len(s.currentPortalStateForProducer.WaitingRedeemRequests))
//	//s.Equal(1, len(s.currentPortalStateForProducer.WaitingPortingRequests))
//
//	//custodian state after auto liquidation
//	custodianKey1 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_1).String()
//	custodianKey2 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_2).String()
//	custodianKey3 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_3).String()
//
//	custodian1 := statedb.NewCustodianStateWithValue(
//		CUS_INC_ADDRESS_1, 6952000000000, 6952000000000,
//		map[string]uint64{
//			common.PortalBNBIDStr: 0,
//		},
//		map[string]uint64{
//			common.PortalBNBIDStr: 0,
//		},
//		map[string]string{
//			common.PortalBNBIDStr: CUS_BNB_ADDRESS_1,
//			common.PortalBTCIDStr: CUS_BTC_ADDRESS_1,
//		},
//		map[string]uint64{}, map[string]uint64{}, map[string]uint64{}, map[string]map[string]uint64{})
//
//	custodian2 := statedb.NewCustodianStateWithValue(
//		CUS_INC_ADDRESS_2, 992800000000, 964800000000,
//		map[string]uint64{
//			common.PortalBNBIDStr: 0.6 * 1e9,
//		},
//		map[string]uint64{
//			common.PortalBNBIDStr: 28000000000,
//		},
//		map[string]string{
//			common.PortalBNBIDStr: CUS_BNB_ADDRESS_2,
//		},
//		map[string]uint64{}, map[string]uint64{}, map[string]uint64{}, map[string]map[string]uint64{})
//
//	custodian3 := statedb.NewCustodianStateWithValue(
//		CUS_INC_ADDRESS_3, 10000*1e9, 8000000000000,
//		map[string]uint64{
//			common.PortalBTCIDStr: 0.1 * 1e9,
//		},
//		map[string]uint64{
//			common.PortalBTCIDStr: 2000000000000,
//		},
//		map[string]string{
//			common.PortalBTCIDStr: CUS_BTC_ADDRESS_3,
//		},
//		map[string]uint64{}, map[string]uint64{}, map[string]uint64{}, map[string]map[string]uint64{})
//
//	s.Equal(custodian1, s.currentPortalStateForProducer.CustodianPoolState[custodianKey1])
//	s.Equal(custodian2, s.currentPortalStateForProducer.CustodianPoolState[custodianKey2])
//	s.Equal(custodian3, s.currentPortalStateForProducer.CustodianPoolState[custodianKey3])
//
//	s.Equal(s.currentPortalStateForProcess, s.currentPortalStateForProducer)
//}
//
///*
//	Feature 5: auto-liquidation: the proportion between the collateral and public token is drop down below 120%
//*/
//func (s *PortalTestSuiteV3) SetupTestAutoLiquidationByExchangeRate() {
//	s.SetupTestAutoLiquidation()
//	finalExchangeRate := statedb.NewFinalExchangeRatesStateWithValue(
//		map[string]statedb.FinalExchangeRatesDetail{
//			common.PRVIDStr:       {Amount: 1000000},
//			common.PortalBNBIDStr: {Amount: 40000000},
//			common.PortalBTCIDStr: {Amount: 10000000000},
//		})
//
//	s.currentPortalStateForProducer.FinalExchangeRatesState = finalExchangeRate
//	s.currentPortalStateForProcess.FinalExchangeRatesState = finalExchangeRate
//}
//
//func (s *PortalTestSuiteV3) TestAutoLiquidationByExchangeRate() {
//	fmt.Println("Running TestAutoLiquidationCustodian - beacon height 1501 ...")
//	bc := s.blockChain
//	beaconHeight := uint64(1501)
//	//shardID := byte(0)
//	//newMatchedRedeemReqIDs := []string{}
//	updatingInfoByTokenID := map[common.Hash]UpdatingInfo{}
//
//	s.SetupTestAutoLiquidationByExchangeRate()
//
//	// producer instructions
//	newInsts, err := buildInstForLiquidationTopPercentileExchangeRates(
//		beaconHeight-1, &s.currentPortalStateForProducer, s.blockChain.GetPortalParams(0))
//	s.Equal(nil, err)
//
//	// process new instructions
//	err = processPortalInstructions(
//		bc, beaconHeight-1, newInsts, s.sdb, &s.currentPortalStateForProcess, s.blockChain.GetPortalParams(0), updatingInfoByTokenID)
//
//	// check results
//	s.Equal(2, len(newInsts))
//	s.Equal(nil, err)
//
//	// remain waiting redeem requests and matched redeem requests
//	s.Equal(2, len(s.currentPortalStateForProducer.MatchedRedeemRequests))
//	s.Equal(0, len(s.currentPortalStateForProducer.WaitingRedeemRequests))
//	s.Equal(1, len(s.currentPortalStateForProducer.WaitingPortingRequests))
//
//	//custodian state after auto liquidation
//	custodianKey1 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_1).String()
//	custodianKey2 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_2).String()
//	custodianKey3 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_3).String()
//	custodianKey4 := statedb.GenerateCustodianStateObjectKey("custodianIncAddress4").String()
//
//	custodian1 := statedb.NewCustodianStateWithValue(
//		CUS_INC_ADDRESS_1, 7000*1e9, 6920000000000,
//		map[string]uint64{
//			common.PortalBNBIDStr: 0,
//		},
//		map[string]uint64{
//			common.PortalBNBIDStr: 80000000000,
//		},
//		map[string]string{
//			common.PortalBNBIDStr: CUS_BNB_ADDRESS_1,
//			common.PortalBTCIDStr: CUS_BTC_ADDRESS_1,
//		}, map[string]uint64{}, map[string]uint64{}, map[string]uint64{}, map[string]map[string]uint64{})
//
//	custodian2 := statedb.NewCustodianStateWithValue(
//		CUS_INC_ADDRESS_2, 972000000000, 960000000000,
//		map[string]uint64{
//			common.PortalBNBIDStr: 0,
//		},
//		map[string]uint64{
//			common.PortalBNBIDStr: 12000000000,
//		},
//		map[string]string{
//			common.PortalBNBIDStr: CUS_BNB_ADDRESS_2,
//		},
//		map[string]uint64{}, map[string]uint64{}, map[string]uint64{}, map[string]map[string]uint64{})
//
//	custodian3 := statedb.NewCustodianStateWithValue(
//		CUS_INC_ADDRESS_3, 10000*1e9, 8000000000000,
//		map[string]uint64{
//			common.PortalBTCIDStr: 0.1 * 1e9,
//		},
//		map[string]uint64{
//			common.PortalBTCIDStr: 2000000000000,
//		},
//		map[string]string{
//			common.PortalBTCIDStr: CUS_BTC_ADDRESS_3,
//		},
//		map[string]uint64{}, map[string]uint64{}, map[string]uint64{}, map[string]map[string]uint64{})
//
//	custodian4 := statedb.NewCustodianStateWithValue(
//		"custodianIncAddress4", 5000*1e9, 4960000000000,
//		map[string]uint64{
//		},
//		map[string]uint64{
//			common.PortalBNBIDStr: 40000000000,
//		},
//		map[string]string{
//			common.PortalBNBIDStr: "bnbAddress4",
//		},
//		map[string]uint64{}, map[string]uint64{}, map[string]uint64{}, map[string]map[string]uint64{})
//
//	rates := map[string]statedb.LiquidationPoolDetail{
//		common.PortalBNBIDStr: {
//			CollateralAmount: 28000000000,
//			PubTokenAmount:   0.7 * 1e9,
//		},
//	}
//	liquidationPool := statedb.NewLiquidationPoolWithValue(rates)
//	liquidationPoolKey := statedb.GeneratePortalLiquidationPoolObjectKey().String()
//
//	s.Equal(custodian1, s.currentPortalStateForProducer.CustodianPoolState[custodianKey1])
//	s.Equal(custodian2, s.currentPortalStateForProducer.CustodianPoolState[custodianKey2])
//	s.Equal(custodian3, s.currentPortalStateForProducer.CustodianPoolState[custodianKey3])
//	s.Equal(custodian4, s.currentPortalStateForProducer.CustodianPoolState[custodianKey4])
//	s.Equal(liquidationPool, s.currentPortalStateForProducer.LiquidationPool[liquidationPoolKey])
//
//	s.Equal(s.currentPortalStateForProcess, s.currentPortalStateForProducer)
//}
//
///*
//	Feature 6: auto-liquidation: the custodian top up the collaterals
//*/
//func (s *PortalTestSuiteV3) SetupTestTopupCustodian() {
//	s.SetupTestAutoLiquidationByExchangeRate()
//}
//
//type TestCaseTopupCustodian struct {
//	incAddressStr        string
//	ptokenID             string
//	depositAmount        uint64
//	freeCollateralAmount uint64
//}
//
//func buildTopupCustodianActionsFromTcs(tcs []TestCaseTopupCustodian, shardID byte) [][]string {
//	insts := [][]string{}
//
//	for _, tc := range tcs {
//		inst := buildPortalTopupCustodianAction(tc.incAddressStr, tc.ptokenID, tc.depositAmount, shardID, tc.freeCollateralAmount)
//		insts = append(insts, inst)
//	}
//
//	return insts
//}
//
//type TestCaseTopupWaitingPorting struct {
//	incAddressStr        string
//	portingID            string
//	ptokenID             string
//	depositAmount        uint64
//	freeCollateralAmount uint64
//}
//
//func buildTopupWaitingPortingActionsFromTcs(tcs []TestCaseTopupWaitingPorting, shardID byte) [][]string {
//	insts := [][]string{}
//
//	for _, tc := range tcs {
//		inst := buildTopupWaitingPortingAction(tc.incAddressStr, tc.portingID, tc.ptokenID, tc.depositAmount, shardID, tc.freeCollateralAmount)
//		insts = append(insts, inst)
//	}
//
//	return insts
//}
//func (s *PortalTestSuiteV3) TestTopupCustodian() {
//	fmt.Println("Running TestAutoLiquidationCustodian - beacon height 1501 ...")
//	bc := s.blockChain
//	beaconHeight := uint64(1501)
//	pm := NewPortalManager()
//	shardID := byte(0)
//	updatingInfoByTokenID := map[common.Hash]UpdatingInfo{}
//
//	s.SetupTestAutoLiquidationByExchangeRate()
//
//	// build test cases for topup custodian
//	testcases := []TestCaseTopupCustodian{
//		// topup by burning more collaterals
//		{
//			incAddressStr:        CUS_INC_ADDRESS_2,
//			ptokenID:             common.PortalBNBIDStr,
//			depositAmount:        500 * 1e9,
//			freeCollateralAmount: 0,
//		},
//		// topup by using free collaterals
//		{
//			incAddressStr:        CUS_INC_ADDRESS_2,
//			ptokenID:             common.PortalBNBIDStr,
//			depositAmount:        0,
//			freeCollateralAmount: 500 * 1e9,
//		},
//	}
//
//	// build actions from testcases
//	insts := buildTopupCustodianActionsFromTcs(testcases, shardID)
//
//	// build test cases for topup waiting porting
//	testcases2 := []TestCaseTopupWaitingPorting{
//		// topup by burning more collaterals
//		{
//			incAddressStr:        "custodianIncAddress4",
//			portingID:            "porting-bnb-1",
//			ptokenID:             common.PortalBNBIDStr,
//			depositAmount:        20 * 1e9,
//			freeCollateralAmount: 0,
//		},
//		// topup by using free collaterals
//		{
//			incAddressStr:        "custodianIncAddress4",
//			portingID:            "porting-bnb-1",
//			ptokenID:             common.PortalBNBIDStr,
//			depositAmount:        0,
//			freeCollateralAmount: 50 * 1e9,
//		},
//	}
//
//	// build actions from testcases2
//	insts2 := buildTopupWaitingPortingActionsFromTcs(testcases2, shardID)
//
//	insts = append(insts, insts2...)
//
//	// producer instructions
//	newInsts, err := producerPortalInstructions(
//		bc, beaconHeight, insts, s.sdb, &s.currentPortalStateForProducer, s.blockChain.GetPortalParams(0), shardID, pm)
//
//	// check liquidation by exchange rates
//	newInstsForLiquidationByExchangeRate, err := buildInstForLiquidationTopPercentileExchangeRates(
//		beaconHeight-1, &s.currentPortalStateForProducer, s.blockChain.GetPortalParams(0))
//
//	s.Equal(0, len(newInstsForLiquidationByExchangeRate))
//
//	// process new instructions
//	err = processPortalInstructions(
//		bc, beaconHeight-1, newInsts, s.sdb, &s.currentPortalStateForProcess, s.blockChain.GetPortalParams(0), updatingInfoByTokenID)
//
//	// check results
//	s.Equal(4, len(newInsts))
//	s.Equal(nil, err)
//
//	// remain waiting redeem requests and matched redeem requests
//	s.Equal(2, len(s.currentPortalStateForProducer.MatchedRedeemRequests))
//	s.Equal(1, len(s.currentPortalStateForProducer.WaitingRedeemRequests))
//	s.Equal(1, len(s.currentPortalStateForProducer.WaitingPortingRequests))
//
//	//custodian state after auto liquidation
//	custodianKey1 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_1).String()
//	custodianKey2 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_2).String()
//	custodianKey3 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_3).String()
//	custodianKey4 := statedb.GenerateCustodianStateObjectKey("custodianIncAddress4").String()
//
//	custodian1 := statedb.NewCustodianStateWithValue(
//		CUS_INC_ADDRESS_1, 7000*1e9, 6920000000000,
//		map[string]uint64{
//			common.PortalBNBIDStr: 0,
//		},
//		map[string]uint64{
//			common.PortalBNBIDStr: 80000000000,
//		},
//		map[string]string{
//			common.PortalBNBIDStr: CUS_BNB_ADDRESS_1,
//			common.PortalBTCIDStr: CUS_BTC_ADDRESS_1,
//		},
//		map[string]uint64{}, map[string]uint64{}, map[string]uint64{}, map[string]map[string]uint64{})
//
//	custodian2 := statedb.NewCustodianStateWithValue(
//		CUS_INC_ADDRESS_2, 1500*1e9, 460000000000,
//		map[string]uint64{
//			common.PortalBNBIDStr: 0.6 * 1e9,
//		},
//		map[string]uint64{
//			common.PortalBNBIDStr: 1040000000000,
//		},
//		map[string]string{
//			common.PortalBNBIDStr: CUS_BNB_ADDRESS_2,
//		},
//		map[string]uint64{}, map[string]uint64{}, map[string]uint64{}, map[string]map[string]uint64{})
//
//	custodian3 := statedb.NewCustodianStateWithValue(
//		CUS_INC_ADDRESS_3, 10000*1e9, 8000000000000,
//		map[string]uint64{
//			common.PortalBTCIDStr: 0.1 * 1e9,
//		},
//		map[string]uint64{
//			common.PortalBTCIDStr: 2000000000000,
//		},
//		map[string]string{
//			common.PortalBTCIDStr: CUS_BTC_ADDRESS_3,
//		},
//		map[string]uint64{}, map[string]uint64{}, map[string]uint64{}, map[string]map[string]uint64{})
//
//	custodian4 := statedb.NewCustodianStateWithValue(
//		"custodianIncAddress4", 5020*1e9, 4910000000000,
//		map[string]uint64{
//		},
//		map[string]uint64{
//			common.PortalBNBIDStr: 110000000000,
//		},
//		map[string]string{
//			common.PortalBNBIDStr: "bnbAddress4",
//		},
//		map[string]uint64{}, map[string]uint64{}, map[string]uint64{}, map[string]map[string]uint64{})
//
//	wPortingReqKey1 := statedb.GeneratePortalWaitingPortingRequestObjectKey("porting-bnb-1").String()
//	wPortingRequest1 := statedb.NewWaitingPortingRequestWithValue(
//		"porting-bnb-1", common.Hash{}, common.PortalBNBIDStr,
//		USER_INC_ADDRESS_1, 1*1e9,
//		[]*statedb.MatchingPortingCustodianDetail{
//			{
//				IncAddress:             "custodianIncAddress4",
//				RemoteAddress:          "bnbAddress4",
//				Amount:                 1 * 1e9,
//				LockedAmountCollateral: 110000000000,
//			},
//		}, 2000000, 1500, 1500, 0)
//
//	wPortingRequests := map[string]*statedb.WaitingPortingRequest{
//		wPortingReqKey1: wPortingRequest1,
//	}
//
//	s.Equal(custodian1, s.currentPortalStateForProducer.CustodianPoolState[custodianKey1])
//	s.Equal(custodian2, s.currentPortalStateForProducer.CustodianPoolState[custodianKey2])
//	s.Equal(custodian3, s.currentPortalStateForProducer.CustodianPoolState[custodianKey3])
//	s.Equal(custodian4, s.currentPortalStateForProducer.CustodianPoolState[custodianKey4])
//	s.Equal(0, len(s.currentPortalStateForProducer.LiquidationPool))
//	s.Equal(wPortingRequests, s.currentPortalStateForProducer.WaitingPortingRequests)
//
//	s.Equal(s.currentPortalStateForProcess, s.currentPortalStateForProducer)
//}
//
///**
//	Feature 7: Porting request timeout - after 21 hours
//**/
//
//func (s *PortalTestSuiteV3) SetupTestPortingRequestExpired() {
//	s.SetupTestRequestPtokens()
//}
//
//func (s *PortalTestSuiteV3) TestPortingRequestExpired() {
//	fmt.Println("Running TestPortingRequestExpired - beacon height 3161 ...")
//	bc := s.blockChain
//	beaconHeight := uint64(3161) // after 24 hours from requesting porting (bch = 100)
//	//shardID := byte(0)
//	//newMatchedRedeemReqIDs := []string{}
//	updatingInfoByTokenID := map[common.Hash]UpdatingInfo{}
//
//	s.SetupTestPortingRequestExpired()
//
//	// producer instructions
//	newInsts, err := bc.checkAndBuildInstForExpiredWaitingPortingRequest(
//		beaconHeight-1, &s.currentPortalStateForProducer,s.blockChain.GetPortalParams(0))
//	s.Equal(nil, err)
//
//	// process new instructions
//	err = processPortalInstructions(
//		bc, beaconHeight-1, newInsts, s.sdb, &s.currentPortalStateForProcess, s.blockChain.GetPortalParams(0), updatingInfoByTokenID)
//
//	// check results
//	s.Equal(3, len(newInsts))
//	s.Equal(nil, err)
//
//	// remain waiting redeem requests and matched redeem requests
//	s.Equal(1, len(s.currentPortalStateForProducer.WaitingPortingRequests))
//
//	//custodian state after auto liquidation
//	custodianKey1 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_1).String()
//	custodianKey2 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_2).String()
//	custodianKey3 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_3).String()
//
//	custodian1 := statedb.NewCustodianStateWithValue(
//		CUS_INC_ADDRESS_1, 7000*1e9, 7000*1e9,
//		map[string]uint64{},
//		map[string]uint64{
//			common.PortalBNBIDStr: 0,
//		},
//		map[string]string{
//			common.PortalBNBIDStr: CUS_BNB_ADDRESS_1,
//			common.PortalBTCIDStr: CUS_BTC_ADDRESS_1,
//		},  map[string]uint64{}, map[string]uint64{}, map[string]uint64{}, map[string]map[string]uint64{})
//
//	custodian2 := statedb.NewCustodianStateWithValue(
//		CUS_INC_ADDRESS_2, 1000*1e9, 1000*1e9,
//		map[string]uint64{}, map[string]uint64{},
//		map[string]string{
//			common.PortalBNBIDStr: CUS_BNB_ADDRESS_2,
//		},  map[string]uint64{}, map[string]uint64{}, map[string]uint64{}, map[string]map[string]uint64{})
//
//	custodian3 := statedb.NewCustodianStateWithValue(
//		CUS_INC_ADDRESS_3, 10000*1e9, 5000000000000,
//		map[string]uint64{},
//		map[string]uint64{
//			common.PortalBTCIDStr: 5000000000000,
//		},
//		map[string]string{
//			common.PortalBTCIDStr: CUS_BTC_ADDRESS_3,
//		}, map[string]uint64{}, map[string]uint64{}, map[string]uint64{}, map[string]map[string]uint64{})
//
//	s.Equal(custodian1, s.currentPortalStateForProducer.CustodianPoolState[custodianKey1])
//	s.Equal(custodian2, s.currentPortalStateForProducer.CustodianPoolState[custodianKey2])
//	s.Equal(custodian3, s.currentPortalStateForProducer.CustodianPoolState[custodianKey3])
//
//	s.Equal(s.currentPortalStateForProcess, s.currentPortalStateForProducer)
//}
//========================

/**
	Feature 8: Custodian rewards from DAO funds and porting/redeem fee
**/

//func (s *PortalTestSuiteV3) SetupTestCustodianRewards() {
//	custodianKey1 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_1).String()
//	custodianKey2 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_2).String()
//	custodianKey3 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_3).String()
//
//	custodian1 := statedb.NewCustodianStateWithValue(
//		CUS_INC_ADDRESS_1, 7000000000000, 6708000000000,
//		map[string]uint64{
//			common.PortalBNBIDStr: 0,
//		},
//		map[string]uint64{
//			common.PortalBNBIDStr: 292000000000,
//		},
//		map[string]string{
//			common.PortalBNBIDStr: CUS_BNB_ADDRESS_1,
//			common.PortalBTCIDStr: CUS_BTC_ADDRESS_1,
//		},  map[string]uint64{}, map[string]uint64{}, map[string]uint64{}, map[string]map[string]uint64{})
//
//	custodian2 := statedb.NewCustodianStateWithValue(
//		CUS_INC_ADDRESS_2, 1000000000000, 972000000000,
//		map[string]uint64{
//			common.PortalBNBIDStr: 0,
//		},
//		map[string]uint64{
//			common.PortalBNBIDStr: 28000000000,
//		},
//		map[string]string{
//			common.PortalBNBIDStr: CUS_BNB_ADDRESS_2,
//		},  map[string]uint64{}, map[string]uint64{}, map[string]uint64{}, map[string]map[string]uint64{})
//
//	custodian3 := statedb.NewCustodianStateWithValue(
//		CUS_INC_ADDRESS_3, 10000000000000, 7988000000000,
//		nil,
//		map[string]uint64{
//			common.PortalBTCIDStr: 2012000000000,
//		},
//		map[string]string{
//			common.PortalBTCIDStr: CUS_BTC_ADDRESS_3,
//		},  map[string]uint64{}, map[string]uint64{}, map[string]uint64{}, map[string]map[string]uint64{})
//
//	custodians := map[string]*statedb.CustodianState{
//		custodianKey1: custodian1,
//		custodianKey2: custodian2,
//		custodianKey3: custodian3,
//	}
//
//	// waiting porting requests
//	wPortingReqKey1 := statedb.GeneratePortalWaitingPortingRequestObjectKey("porting-bnb-1").String()
//	wPortingRequest1 := statedb.NewWaitingPortingRequestWithValue(
//		"porting-bnb-1", common.Hash{}, common.PortalBNBIDStr,
//		USER_INC_ADDRESS_1, 1*1e9,
//		[]*statedb.MatchingPortingCustodianDetail{
//			{
//				IncAddress:             CUS_INC_ADDRESS_1,
//				RemoteAddress:          CUS_BNB_ADDRESS_1,
//				Amount:                 0.3 * 1e9,
//				LockedAmountCollateral: 12000000000,
//			},
//			{
//				IncAddress:             CUS_INC_ADDRESS_2,
//				RemoteAddress:          CUS_BNB_ADDRESS_1,
//				Amount:                 0.7 * 1e9,
//				LockedAmountCollateral: 28000000000,
//			},
//		}, 2000000, 1000, 1000, 0)
//
//	wPortingReqKey2 := statedb.GeneratePortalWaitingPortingRequestObjectKey("porting-btc-2").String()
//	wPortingRequest2 := statedb.NewWaitingPortingRequestWithValue(
//		"porting-btc-2", common.Hash{}, common.PortalBTCIDStr,
//		USER_INC_ADDRESS_2, 0.1*1e9,
//		[]*statedb.MatchingPortingCustodianDetail{
//			{
//				IncAddress:             CUS_INC_ADDRESS_3,
//				RemoteAddress:          CUS_BTC_ADDRESS_3,
//				Amount:                 0.1 * 1e9,
//				LockedAmountCollateral: 2000000000000,
//			},
//		}, 100000001, 1000, 1000, 0)
//
//	wPortingReqKey3 := statedb.GeneratePortalWaitingPortingRequestObjectKey("porting-bnb-3").String()
//	wPortingRequest3 := statedb.NewWaitingPortingRequestWithValue(
//		"porting-bnb-3", common.Hash{}, common.PortalBNBIDStr,
//		USER_INC_ADDRESS_2, 5*1e9,
//		[]*statedb.MatchingPortingCustodianDetail{
//			{
//				IncAddress:             CUS_INC_ADDRESS_1,
//				RemoteAddress:          CUS_BNB_ADDRESS_1,
//				Amount:                 5 * 1e9,
//				LockedAmountCollateral: 200000000000,
//			},
//		}, 2000000, 900, 900, 0)
//
//	wPortingRequests := map[string]*statedb.WaitingPortingRequest{
//		wPortingReqKey1: wPortingRequest1,
//		wPortingReqKey2: wPortingRequest2,
//		wPortingReqKey3: wPortingRequest3,
//	}
//
//	// matched redeem requests
//	redeemReqKey1 := statedb.GenerateMatchedRedeemRequestObjectKey("redeem-bnb-1").String()
//	redeemRequest1 := statedb.NewRedeemRequestWithValue(
//		"redeem-bnb-1", common.PortalBNBIDStr,
//		USER_INC_ADDRESS_1, "userCUS_BNB_ADDRESS_1", 2.3*1e9,
//		[]*statedb.MatchingRedeemCustodianDetail{
//			statedb.NewMatchingRedeemCustodianDetailWithValue(CUS_INC_ADDRESS_1, CUS_BNB_ADDRESS_1, 2*1e9),
//			statedb.NewMatchingRedeemCustodianDetailWithValue(CUS_INC_ADDRESS_2, CUS_BNB_ADDRESS_2, 0.3*1e9),
//		}, 4600000, 990, common.Hash{}, 0, 990, "")
//
//	matchedRedeemRequest := map[string]*statedb.RedeemRequest{
//		redeemReqKey1: redeemRequest1,
//	}
//
//	// locked collaterals
//	lockedCollateralDetail := map[string]uint64{
//		CUS_INC_ADDRESS_1: 292000000000,
//		CUS_INC_ADDRESS_2: 28000000000,
//		CUS_INC_ADDRESS_3: 2012000000000,
//	}
//	totalLockedCollateralInEpoch := uint64(2332000000000)
//	s.currentPortalStateForProducer.LockedCollateralForRewards = statedb.NewLockedCollateralStateWithValue(
//		totalLockedCollateralInEpoch, lockedCollateralDetail)
//	s.currentPortalStateForProducer.CustodianPoolState = custodians
//	s.currentPortalStateForProducer.WaitingPortingRequests = wPortingRequests
//	s.currentPortalStateForProducer.MatchedRedeemRequests = matchedRedeemRequest
//
//	s.currentPortalStateForProcess.CustodianPoolState = cloneCustodians(custodians)
//	s.currentPortalStateForProcess.WaitingPortingRequests = cloneWPortingRequests(wPortingRequests)
//	s.currentPortalStateForProcess.MatchedRedeemRequests = cloneRedeemRequests(matchedRedeemRequest)
//	s.currentPortalStateForProcess.LockedCollateralForRewards = statedb.NewLockedCollateralStateWithValue(
//		totalLockedCollateralInEpoch, cloneMap(lockedCollateralDetail))
//}
//
//func (s *PortalTestSuiteV3) TestCustodianRewards() {
//	fmt.Println("Running TestCustodianRewards - beacon height 1000 ...")
//	bc := s.blockChain
//	beaconHeight := uint64(1000) // after 24 hours from requesting porting (bch = 100)
//	//shardID := byte(0)
//	newMatchedRedeemReqIDs := []string{"redeem-bnb-1"}
//	rewardForCustodianByEpoch := map[common.Hash]uint64{
//		common.PRVCoinID: 100000000000, // 100 prv
//		common.Hash{1}:   200000,
//	}
//	updatingInfoByTokenID := map[common.Hash]UpdatingInfo{}
//
//	s.SetupTestCustodianRewards()
//
//	// producer instructions
//	newInsts, err := bc.buildPortalRewardsInsts(
//		beaconHeight-1, &s.currentPortalStateForProducer, rewardForCustodianByEpoch, newMatchedRedeemReqIDs)
//	s.Equal(nil, err)
//
//	// process new instructions
//	err = processPortalInstructions(
//		bc, beaconHeight-1, newInsts, s.sdb, &s.currentPortalStateForProcess, s.blockChain.GetPortalParams(0), updatingInfoByTokenID)
//
//	// check results
//	s.Equal(2, len(newInsts))
//	s.Equal(nil, err)
//
//	//custodian state after auto liquidation
//	custodianKey1 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_1).String()
//	custodianKey2 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_2).String()
//	custodianKey3 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_3).String()
//
//	reward1 := map[string]uint64{
//		common.PRVIDStr:         12526040824,
//		common.Hash{1}.String(): 25044,
//	}
//
//	reward2 := map[string]uint64{
//		common.PRVIDStr:         1202686106,
//		common.Hash{1}.String(): 2401,
//	}
//	reward3 := map[string]uint64{
//		common.PRVIDStr:         86377873071,
//		common.Hash{1}.String(): 172555,
//	}
//
//	s.Equal(reward1, s.currentPortalStateForProducer.CustodianPoolState[custodianKey1].GetRewardAmount())
//	s.Equal(reward2, s.currentPortalStateForProducer.CustodianPoolState[custodianKey2].GetRewardAmount())
//	s.Equal(reward3, s.currentPortalStateForProducer.CustodianPoolState[custodianKey3].GetRewardAmount())
//
//	s.Equal(reward1, s.currentPortalStateForProcess.CustodianPoolState[custodianKey1].GetRewardAmount())
//	s.Equal(reward2, s.currentPortalStateForProcess.CustodianPoolState[custodianKey2].GetRewardAmount())
//	s.Equal(reward3, s.currentPortalStateForProcess.CustodianPoolState[custodianKey3].GetRewardAmount())
//}

func TestPortalSuiteV3(t *testing.T) {
	suite.Run(t, new(PortalTestSuiteV3))
}
