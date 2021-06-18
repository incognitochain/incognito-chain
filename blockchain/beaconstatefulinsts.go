package blockchain

import (
	"math/big"
	"sort"
	"strconv"

	"github.com/incognitochain/incognito-chain/blockchain/pdex"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/instruction"
	"github.com/incognitochain/incognito-chain/portal"
	portalprocessv3 "github.com/incognitochain/incognito-chain/portal/portalv3/portalprocess"
	"github.com/incognitochain/incognito-chain/utils"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
)

// build instructions at beacon chain before syncing to shards
func collectStatefulActions(
	shardBlockInstructions [][]string,
) [][]string {
	// stateful instructions are dependently processed with results of instructioins before them in shards2beacon blocks
	statefulInsts := [][]string{}
	for _, inst := range shardBlockInstructions {
		if len(inst) < 2 {
			continue
		}
		if instruction.IsConsensusInstruction(inst[0]) {
			continue
		}

		metaType, err := strconv.Atoi(inst[0])
		if err != nil {
			Logger.log.Error(err)
			continue
		}
		switch metaType {
		case metadata.InitTokenRequestMeta,
			metadata.IssuingRequestMeta,
			metadata.IssuingETHRequestMeta,
			metadata.PDEContributionMeta,
			metadata.PDETradeRequestMeta,
			metadata.PDEWithdrawalRequestMeta,
			metadata.PDEFeeWithdrawalRequestMeta,
			metadata.PDEPRVRequiredContributionRequestMeta,
			metadata.PDECrossPoolTradeRequestMeta,
			metadata.PortalCustodianDepositMeta,
			metadata.PortalRequestPortingMeta,
			metadata.PortalUserRequestPTokenMeta,
			metadata.PortalExchangeRatesMeta,
			metadata.PortalUnlockOverRateCollateralsMeta,
			metadata.RelayingBNBHeaderMeta,
			metadata.RelayingBTCHeaderMeta,
			metadata.PortalCustodianWithdrawRequestMeta,
			metadata.PortalRedeemRequestMeta,
			metadata.PortalRequestUnlockCollateralMeta,
			metadata.PortalRequestUnlockCollateralMetaV3,
			metadata.PortalLiquidateCustodianMeta,
			metadata.PortalLiquidateCustodianMetaV3,
			metadata.PortalRequestWithdrawRewardMeta,
			metadata.PortalRedeemFromLiquidationPoolMeta,
			metadata.PortalCustodianTopupMetaV2,
			metadata.PortalCustodianTopupResponseMeta,
			metadata.PortalReqMatchingRedeemMeta,
			metadata.PortalTopUpWaitingPortingRequestMeta,
			metadata.PortalCustodianDepositMetaV3,
			metadata.PortalCustodianWithdrawRequestMetaV3,
			metadata.PortalRedeemFromLiquidationPoolMetaV3,
			metadata.PortalCustodianTopupMetaV3,
			metadata.PortalTopUpWaitingPortingRequestMetaV3,
			metadata.PortalRequestPortingMetaV3,
			metadata.PortalRedeemRequestMetaV3:
			statefulInsts = append(statefulInsts, inst)

		default:
			continue
		}
	}
	return statefulInsts
}

func (blockchain *BlockChain) buildStatefulInstructions(
	beaconBestState *BeaconBestState,
	featureStateDB *statedb.StateDB,
	statefulActionsByShardID map[byte][][]string,
	beaconHeight uint64,
	rewardForCustodianByEpoch map[common.Hash]uint64,
	portalParams portal.PortalParams,
) ([][]string, error) {

	/*currentPDEState, err := InitCurrentPDEStateFromDB(featureStateDB, beaconBestState.pdeState, beaconHeight-1)*/
	//if err != nil {
	//Logger.log.Error(err)
	//return utils.EmptyStringMatrix, err
	/*}*/

	pm := portal.NewPortalManager()
	currentPortalStateV3, err := portalprocessv3.InitCurrentPortalStateFromDB(featureStateDB)
	if err != nil {
		Logger.log.Error(err)
		return utils.EmptyStringMatrix, err
	}
	relayingHeaderState, err := blockchain.InitRelayingHeaderChainStateFromDB()
	if err != nil {
		Logger.log.Error(err)
		return utils.EmptyStringMatrix, err
	}

	accumulatedValues := &metadata.AccumulatedValues{
		UniqETHTxsUsed:   [][]byte{},
		DBridgeTokenPair: map[string][]byte{},
		CBridgeTokens:    []*common.Hash{},
	}
	instructions := [][]string{}

	// Start pde instructions handler
	pdeContributionActions := [][]string{}
	pdePRVRequiredContributionActions := [][]string{}
	pdeTradeActions := [][]string{}
	pdeCrossPoolTradeActions := [][]string{}
	pdeWithdrawalActions := [][]string{}
	pdeFeeWithdrawalActions := [][]string{}

	var keys []int
	for k := range statefulActionsByShardID {
		keys = append(keys, int(k))
	}

	sort.Ints(keys)
	for _, value := range keys {
		shardID := byte(value)
		actions := statefulActionsByShardID[shardID]
		for _, action := range actions {
			metaType, err := strconv.Atoi(action[0])
			if err != nil {
				continue
			}
			contentStr := action[1]
			newInst := [][]string{}

			// group portal instructions
			isCollected := portal.CollectPortalInstructions(pm, metaType, action, shardID)
			if isCollected {
				continue
			}

			switch metaType {
			case metadata.InitTokenRequestMeta:
				newInst, err = blockchain.buildInstructionsForTokenInitReq(beaconBestState, featureStateDB, contentStr, shardID, metaType, accumulatedValues)
				if err != nil {
					Logger.log.Error(err)
					continue
				}
			case metadata.IssuingRequestMeta:
				newInst, err = blockchain.buildInstructionsForIssuingReq(beaconBestState, featureStateDB, contentStr, shardID, metaType, accumulatedValues)
				if err != nil {
					Logger.log.Error(err)
					continue
				}
			case metadata.IssuingETHRequestMeta:
				newInst, err = blockchain.buildInstructionsForIssuingETHReq(beaconBestState, featureStateDB, contentStr, shardID, metaType, accumulatedValues)
				if err != nil {
					Logger.log.Error(err)
					continue
				}

			case metadata.PDEContributionMeta:
				pdeContributionActions = append(pdeContributionActions, action)
			case metadata.PDEPRVRequiredContributionRequestMeta:
				pdePRVRequiredContributionActions = append(pdePRVRequiredContributionActions, action)
			case metadata.PDETradeRequestMeta:
				pdeTradeActions = append(pdeTradeActions, action)
			case metadata.PDECrossPoolTradeRequestMeta:
				pdeCrossPoolTradeActions = append(pdeCrossPoolTradeActions, action)
			case metadata.PDEWithdrawalRequestMeta:
				pdeWithdrawalActions = append(pdeWithdrawalActions, action)
			case metadata.PDEFeeWithdrawalRequestMeta:
				pdeFeeWithdrawalActions = append(pdeFeeWithdrawalActions, action)
			default:
				continue
			}
			if len(newInst) > 0 {
				instructions = append(instructions, newInst...)
			}
		}
	}

	if beaconBestState.pDEXState.Version() <= pdex.ForceWithPrvVersion {
		pdexStateEnv := pdex.
			NewStateEnvBuilder().
			BuildContributionActions(pdeContributionActions).
			BuildPRVRequiredContributionActions(pdePRVRequiredContributionActions).
			BuildTradeActions(pdeTradeActions).
			BuildCrossPoolTradeActions(pdeCrossPoolTradeActions).
			BuildWithdrawalActions(pdeWithdrawalActions).
			BuildFeeWithdrawalActions(pdeFeeWithdrawalActions).
			Build()

		pdeInstructions, err := beaconBestState.pDEXState.Update(pdexStateEnv)
		if err != nil {
			Logger.log.Error(err)
			return utils.EmptyStringMatrix, err
		}
		instructions = append(instructions, pdeInstructions...)
	}

	// TODO: @tin check here again
	/*pdeInsts, err := blockchain.handlePDEInsts(*/
	//beaconHeight-1, currentPDEState,
	//pdeContributionActionsByShardID,
	//pdePRVRequiredContributionActionsByShardID,
	//pdeTradeActionsByShardID,
	//pdeCrossPoolTradeActionsByShardID,
	//pdeWithdrawalActionsByShardID,
	//pdeFeeWithdrawalActionsByShardID,
	//)

	//if err != nil {
	//Logger.log.Error(err)
	//return instructions, err
	/*}*/

	// handle portal instructions
	// include portal v3, portal relaying header chain
	portalInsts, err := blockchain.handlePortalInsts(
		featureStateDB,
		beaconHeight-1,
		currentPortalStateV3,
		relayingHeaderState,
		rewardForCustodianByEpoch,
		portalParams,
		pm,
	)

	if err != nil {
		Logger.log.Error(err)
		return instructions, err
	}
	if len(portalInsts) > 0 {
		instructions = append(instructions, portalInsts...)
	}

	return instructions, nil
}

func isTradingFairContainsPRV(
	tokenIDToSellStr string,
	tokenIDToBuyStr string,
) bool {
	return tokenIDToSellStr == common.PRVCoinID.String() || tokenIDToBuyStr == common.PRVCoinID.String()
}

func isPoolPairExisting(
	beaconHeight uint64,
	currentPDEState *CurrentPDEState,
	token1IDStr string,
	token2IDStr string,
) bool {
	poolPairKey := string(rawdbv2.BuildPDEPoolForPairKey(beaconHeight, token1IDStr, token2IDStr))
	poolPair, found := currentPDEState.PDEPoolPairs[poolPairKey]
	if !found || poolPair == nil || poolPair.Token1PoolValue == 0 || poolPair.Token2PoolValue == 0 {
		return false
	}
	return true
}

func calcTradeValue(
	pdePoolPair *rawdbv2.PDEPoolForPair,
	tokenIDStrToSell string,
	sellAmount uint64,
) (uint64, uint64, uint64) {
	tokenPoolValueToBuy := pdePoolPair.Token1PoolValue
	tokenPoolValueToSell := pdePoolPair.Token2PoolValue
	if pdePoolPair.Token1IDStr == tokenIDStrToSell {
		tokenPoolValueToSell = pdePoolPair.Token1PoolValue
		tokenPoolValueToBuy = pdePoolPair.Token2PoolValue
	}
	invariant := big.NewInt(0)
	invariant.Mul(new(big.Int).SetUint64(tokenPoolValueToSell), new(big.Int).SetUint64(tokenPoolValueToBuy))
	newTokenPoolValueToSell := big.NewInt(0)
	newTokenPoolValueToSell.Add(new(big.Int).SetUint64(tokenPoolValueToSell), new(big.Int).SetUint64(sellAmount))

	newTokenPoolValueToBuy := big.NewInt(0).Div(invariant, newTokenPoolValueToSell).Uint64()
	modValue := big.NewInt(0).Mod(invariant, newTokenPoolValueToSell)
	if modValue.Cmp(big.NewInt(0)) != 0 {
		newTokenPoolValueToBuy++
	}
	if tokenPoolValueToBuy <= newTokenPoolValueToBuy {
		return uint64(0), uint64(0), uint64(0)
	}
	return tokenPoolValueToBuy - newTokenPoolValueToBuy, newTokenPoolValueToBuy, newTokenPoolValueToSell.Uint64()
}

func prepareInfoForSorting(
	currentPDEState *CurrentPDEState,
	beaconHeight uint64,
	tradeAction metadata.PDECrossPoolTradeRequestAction,
) (uint64, uint64) {
	prvIDStr := common.PRVCoinID.String()
	tradeMeta := tradeAction.Meta
	sellAmount := tradeMeta.SellAmount
	tradingFee := tradeMeta.TradingFee
	if tradeMeta.TokenIDToSellStr == prvIDStr {
		return tradingFee, sellAmount
	}
	poolPairKey := string(rawdbv2.BuildPDEPoolForPairKey(beaconHeight, prvIDStr, tradeMeta.TokenIDToSellStr))
	poolPair, _ := currentPDEState.PDEPoolPairs[poolPairKey]
	sellAmount, _, _ = calcTradeValue(poolPair, tradeMeta.TokenIDToSellStr, sellAmount)
	return tradingFee, sellAmount
}

func (blockchain *BlockChain) handlePDEInsts(
	beaconHeight uint64,
	currentPDEState *CurrentPDEState,
	pdeContributionActionsByShardID map[byte][][]string,
	pdePRVRequiredContributionActionsByShardID map[byte][][]string,
	pdeTradeActionsByShardID map[byte][][]string,
	pdeCrossPoolTradeActionsByShardID map[byte][][]string,
	pdeWithdrawalActionsByShardID map[byte][][]string,
	pdeFeeWithdrawalActionsByShardID map[byte][][]string,
) ([][]string, error) {
	instructions := [][]string{}

	// handle withdrawal
	var wrKeys []int
	for k := range pdeWithdrawalActionsByShardID {
		wrKeys = append(wrKeys, int(k))
	}
	sort.Ints(wrKeys)
	for _, value := range wrKeys {
		shardID := byte(value)
		actions := pdeWithdrawalActionsByShardID[shardID]
		for _, action := range actions {
			contentStr := action[1]
			newInst, err := blockchain.buildInstructionsForPDEWithdrawal(contentStr, shardID, metadata.PDEWithdrawalRequestMeta, currentPDEState, beaconHeight)
			if err != nil {
				Logger.log.Error(err)
				continue
			}
			if len(newInst) > 0 {
				instructions = append(instructions, newInst...)
			}
		}
	}

	// handle contribution
	var ctKeys []int
	for k := range pdeContributionActionsByShardID {
		ctKeys = append(ctKeys, int(k))
	}
	sort.Ints(ctKeys)
	for _, value := range ctKeys {
		shardID := byte(value)
		actions := pdeContributionActionsByShardID[shardID]
		for _, action := range actions {
			contentStr := action[1]
			newInst, err := blockchain.buildInstructionsForPDEContribution(contentStr, shardID, metadata.PDEContributionMeta, currentPDEState, beaconHeight, false)
			if err != nil {
				Logger.log.Error(err)
				continue
			}
			if len(newInst) > 0 {
				instructions = append(instructions, newInst...)
			}
		}
	}

	// handle prv required contribution
	var prvRequiredContribKeys []int
	for k := range pdePRVRequiredContributionActionsByShardID {
		prvRequiredContribKeys = append(prvRequiredContribKeys, int(k))
	}
	sort.Ints(prvRequiredContribKeys)
	for _, value := range prvRequiredContribKeys {
		shardID := byte(value)
		actions := pdePRVRequiredContributionActionsByShardID[shardID]
		for _, action := range actions {
			contentStr := action[1]
			newInst, err := blockchain.buildInstructionsForPDEContribution(contentStr, shardID, metadata.PDEPRVRequiredContributionRequestMeta, currentPDEState, beaconHeight, true)
			if err != nil {
				Logger.log.Error(err)
				continue
			}
			if len(newInst) > 0 {
				instructions = append(instructions, newInst...)
			}
		}
	}
	return instructions, nil
}
