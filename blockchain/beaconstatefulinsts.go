package blockchain

import (
	"sort"
	"strconv"

	"github.com/incognitochain/incognito-chain/blockchain/bridgeagg"
	"github.com/incognitochain/incognito-chain/blockchain/pdex"
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/utils"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/instruction"
	"github.com/incognitochain/incognito-chain/metadata"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	"github.com/incognitochain/incognito-chain/portal"
	"github.com/incognitochain/incognito-chain/portal/portalrelaying"
	portalprocessv3 "github.com/incognitochain/incognito-chain/portal/portalv3/portalprocess"
	portalprocessv4 "github.com/incognitochain/incognito-chain/portal/portalv4/portalprocess"
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
			metadata.IssuingBSCRequestMeta,
			metadata.IssuingPRVERC20RequestMeta,
			metadata.IssuingPRVBEP20RequestMeta,
			metadata.IssuingPLGRequestMeta,
			metadata.IssuingFantomRequestMeta,
			metadata.IssuingAuroraRequestMeta,
			metadata.IssuingAvaxRequestMeta,
			metadata.IssuingNearRequestMeta,
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
			metadata.PortalRedeemRequestMetaV3,
			metadataCommon.PortalV4ShieldingRequestMeta,
			metadataCommon.PortalV4UnshieldingRequestMeta,
			metadataCommon.PortalV4FeeReplacementRequestMeta,
			metadataCommon.PortalV4SubmitConfirmedTxMeta,
			metadataCommon.PortalV4ConvertVaultRequestMeta,
			metadataCommon.BridgeAggModifyParamMeta,
			metadataCommon.BridgeAggConvertTokenToUnifiedTokenRequestMeta,
			metadataCommon.IssuingUnifiedTokenRequestMeta,
			metadataCommon.BurningUnifiedTokenRequestMeta,
			metadataCommon.BriHubRegisterBridgeMeta:
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
	shardStates map[byte][]types.ShardState,
	allPdexTxs map[uint]map[byte][]metadata.Transaction,
	pdexReward uint64,
) ([][]string, error) {
	// transfrom beacon height for pdex process
	pdeVersions := []int{}
	for version := range beaconBestState.pdeStates {
		pdeVersions = append(pdeVersions, int(version))
	}
	sort.Ints(pdeVersions)

	for _, version := range pdeVersions {
		if version == pdex.BasicVersion {
			hasTx := false
			for _, v := range allPdexTxs[uint(version)] {
				if len(v) != 0 {
					hasTx = true
					break
				}
			}
			if !hasTx {
				continue
			}
			beaconBestState.pdeStates[uint(version)].TransformKeyWithNewBeaconHeight(beaconHeight - 1)
		}
	}

	pm := portal.NewPortalManager()
	currentPortalStateV3, err := portalprocessv3.InitCurrentPortalStateFromDB(
		featureStateDB, beaconBestState.portalStateV3)
	if err != nil {
		Logger.log.Error(err)
		return utils.EmptyStringMatrix, err
	}
	relayingHeaderState, err := portalrelaying.InitRelayingHeaderChainStateFromDB(blockchain.GetBNBHeaderChain(), blockchain.GetBTCHeaderChain())
	if err != nil {
		Logger.log.Error(err)
	}
	currentPortalStateV4, err := portalprocessv4.InitCurrentPortalStateV4FromDB(
		featureStateDB,
		beaconBestState.portalStateV4,
		portalParams.GetPortalParamsV4(beaconHeight))
	if err != nil {
		Logger.log.Error(err)
		return utils.EmptyStringMatrix, err
	}

	accumulatedValues := &metadata.AccumulatedValues{
		UniqETHTxsUsed:   [][]byte{},
		UniqBSCTxsUsed:   [][]byte{},
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

	// bridge agg actions collector
	// unshieldActions := make([][]string, beaconBestState.ActiveShards)
	shieldActions := make([][]string, beaconBestState.ActiveShards)
	convertActions := make([][]string, beaconBestState.ActiveShards)
	modifyParamActions := make([][]string, beaconBestState.ActiveShards)
	sDBs, err := blockchain.getStateDBsForVerifyTokenID(beaconBestState)
	if err != nil {
		Logger.log.Error(err)
		return utils.EmptyStringMatrix, err
	}

	newInsts, newAccumulatedValues, err := beaconBestState.bridgeAggManager.BuildAddTokenInstruction(beaconHeight, sDBs, accumulatedValues, beaconBestState.TriggeredFeature)
	if err != nil {
		return [][]string{}, err
	}
	if len(newInsts) > 0 {
		instructions = append(instructions, newInsts...)
		accumulatedValues = newAccumulatedValues
	}

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

			// group portal instructions (both portal relaying, portal v3, portal v4)
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
				var uniqTxs [][]byte
				newInst, uniqTxs, err = blockchain.buildInstructionsForIssuingBridgeReq(
					sDBs,
					contentStr,
					shardID,
					metaType,
					accumulatedValues,
					accumulatedValues.UniqETHTxsUsed,
					config.Param().EthContractAddressStr,
					"",
					statedb.IsETHTxHashIssued,
					false,
				)
				if uniqTxs != nil {
					accumulatedValues.UniqETHTxsUsed = append(accumulatedValues.UniqETHTxsUsed, uniqTxs...)
				}
			case metadata.IssuingBSCRequestMeta:
				var uniqTxs [][]byte
				newInst, uniqTxs, err = blockchain.buildInstructionsForIssuingBridgeReq(
					sDBs,
					contentStr,
					shardID,
					metaType,
					accumulatedValues,
					accumulatedValues.UniqBSCTxsUsed,
					config.Param().BscContractAddressStr,
					common.BSCPrefix,
					statedb.IsBSCTxHashIssued,
					false,
				)
				if uniqTxs != nil {
					accumulatedValues.UniqBSCTxsUsed = append(accumulatedValues.UniqBSCTxsUsed, uniqTxs...)
				}
			case metadata.IssuingPRVERC20RequestMeta:
				var uniqTxs [][]byte
				newInst, uniqTxs, err = blockchain.buildInstructionsForIssuingBridgeReq(
					sDBs,
					contentStr,
					shardID,
					metaType,
					accumulatedValues,
					accumulatedValues.UniqPRVEVMTxsUsed,
					config.Param().PRVERC20ContractAddressStr,
					"",
					statedb.IsPRVEVMTxHashIssued,
					true,
				)
				if uniqTxs != nil {
					accumulatedValues.UniqPRVEVMTxsUsed = append(accumulatedValues.UniqPRVEVMTxsUsed, uniqTxs...)
				}
			case metadata.IssuingPRVBEP20RequestMeta:
				var uniqTxs [][]byte
				newInst, uniqTxs, err = blockchain.buildInstructionsForIssuingBridgeReq(
					sDBs,
					contentStr,
					shardID,
					metaType,
					accumulatedValues,
					accumulatedValues.UniqPRVEVMTxsUsed,
					config.Param().PRVBEP20ContractAddressStr,
					"",
					statedb.IsPRVEVMTxHashIssued,
					true,
				)
				if uniqTxs != nil {
					accumulatedValues.UniqPRVEVMTxsUsed = append(accumulatedValues.UniqPRVEVMTxsUsed, uniqTxs...)
				}
			case metadata.IssuingPLGRequestMeta:
				var uniqTxs [][]byte
				newInst, uniqTxs, err = blockchain.buildInstructionsForIssuingBridgeReq(
					sDBs,
					contentStr,
					shardID,
					metaType,
					accumulatedValues,
					accumulatedValues.UniqPLGTxsUsed,
					config.Param().PlgContractAddressStr,
					common.PLGPrefix,
					statedb.IsPLGTxHashIssued,
					false,
				)
				if uniqTxs != nil {
					accumulatedValues.UniqPLGTxsUsed = append(accumulatedValues.UniqPLGTxsUsed, uniqTxs...)
				}
			case metadata.IssuingFantomRequestMeta:
				var uniqTxs [][]byte
				newInst, uniqTxs, err = blockchain.buildInstructionsForIssuingBridgeReq(
					sDBs,
					contentStr,
					shardID,
					metaType,
					accumulatedValues,
					accumulatedValues.UniqFTMTxsUsed,
					config.Param().FtmContractAddressStr,
					common.FTMPrefix,
					statedb.IsFTMTxHashIssued,
					false,
				)
				if uniqTxs != nil {
					accumulatedValues.UniqFTMTxsUsed = append(accumulatedValues.UniqFTMTxsUsed, uniqTxs...)
				}

			case metadata.IssuingAuroraRequestMeta:
				var uniqTxs [][]byte
				newInst, uniqTxs, err = blockchain.buildInstructionsForIssuingBridgeReq(
					sDBs,
					contentStr,
					shardID,
					metaType,
					accumulatedValues,
					accumulatedValues.UniqAURORATxsUsed,
					config.Param().AuroraContractAddressStr,
					common.AURORAPrefix,
					statedb.IsAURORATxHashIssued,
					false,
				)
				if uniqTxs != nil {
					accumulatedValues.UniqAURORATxsUsed = append(accumulatedValues.UniqAURORATxsUsed, uniqTxs...)
				}

			case metadata.IssuingAvaxRequestMeta:
				var uniqTxs [][]byte
				newInst, uniqTxs, err = blockchain.buildInstructionsForIssuingBridgeReq(
					sDBs,
					contentStr,
					shardID,
					metaType,
					accumulatedValues,
					accumulatedValues.UniqAVAXTxsUsed,
					config.Param().AvaxContractAddressStr,
					common.AVAXPrefix,
					statedb.IsAVAXTxHashIssued,
					false,
				)
				if uniqTxs != nil {
					accumulatedValues.UniqAVAXTxsUsed = append(accumulatedValues.UniqAVAXTxsUsed, uniqTxs...)
				}

			case metadata.IssuingNearRequestMeta:
				var uniqTx []byte
				newInst, uniqTx, err = blockchain.buildInstructionsForIssuingWasmBridgeReq(
					sDBs,
					contentStr,
					shardID,
					metaType,
					accumulatedValues,
					accumulatedValues.UniqNEARTxsUsed,
					config.Param().NearContractAddressStr,
					common.NEARPrefix,
					statedb.IsNEARTxHashIssued,
				)
				if uniqTx != nil {
					accumulatedValues.UniqNEARTxsUsed = append(accumulatedValues.UniqNEARTxsUsed, uniqTx)
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
			case metadataCommon.BridgeAggModifyParamMeta:
				modifyParamActions[shardID] = append(modifyParamActions[shardID], contentStr)
			case metadataCommon.BridgeAggConvertTokenToUnifiedTokenRequestMeta:
				convertActions[shardID] = append(convertActions[shardID], contentStr)
			case metadataCommon.IssuingUnifiedTokenRequestMeta:
				shieldActions[shardID] = append(shieldActions[shardID], contentStr)
			default:
				continue
			}
			if len(newInst) > 0 {
				instructions = append(instructions, newInst...)
			}
		}
	}

	pdeStateEnv := pdex.
		NewStateEnvBuilder().
		BuildPrevBeaconHeight(beaconHeight - 1).
		BuildContributionActions(pdeContributionActions).
		BuildPRVRequiredContributionActions(pdePRVRequiredContributionActions).
		BuildTradeActions(pdeTradeActions).
		BuildCrossPoolTradeActions(pdeCrossPoolTradeActions).
		BuildWithdrawalActions(pdeWithdrawalActions).
		BuildFeeWithdrawalActions(pdeFeeWithdrawalActions).
		BuildListTxs(allPdexTxs[pdex.AmplifierVersion]).
		BuildBCHeightBreakPointPrivacyV2(config.Param().BCHeightBreakPointPrivacyV2).
		BuildPdexv3BreakPoint(config.Param().PDexParams.Pdexv3BreakPointHeight).
		BuildReward(pdexReward).
		Build()

	for _, version := range pdeVersions {
		pdeInstructions, err := beaconBestState.pdeStates[uint(version)].BuildInstructions(pdeStateEnv)
		if err != nil {
			return utils.EmptyStringMatrix, err
		}
		instructions = append(instructions, pdeInstructions...)
	}

	// handle portal instructions
	// include portal v3, portal v4, portal relaying header chain
	portalInsts, err := blockchain.handlePortalInsts(
		featureStateDB,
		beaconHeight-1,
		currentPortalStateV3,
		currentPortalStateV4,
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

	// Bridge aggregator instructions (don't build unshield instructions here)
	bridgeAggEnv := bridgeagg.
		NewStateEnvBuilder().
		BuildConvertActions(convertActions).
		BuildModifyParamActions(modifyParamActions).
		BuildShieldActions(shieldActions).
		BuildAccumulatedValues(accumulatedValues).
		BuildBeaconHeight(beaconHeight).
		BuildStateDBs(sDBs).
		Build()
	bridgeAggInsts, newAccumulatedValues, err := beaconBestState.bridgeAggManager.BuildInstructions(bridgeAggEnv)
	if err != nil {
		return instructions, err
	}
	if len(bridgeAggInsts) > 0 {
		instructions = append(instructions, bridgeAggInsts...)
	}
	accumulatedValues = newAccumulatedValues

	return instructions, nil
}
