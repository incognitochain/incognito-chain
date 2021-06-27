package blockchain

import (
	"sort"
	"strconv"

	"github.com/incognitochain/incognito-chain/blockchain/pdex"
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/config"
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
			metadata.IssuingBSCRequestMeta,
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
	shardStates map[byte][]types.ShardState,
) ([][]string, error) {
	// transfrom beacon height for pdex process
	beaconBestState.pdeState.TransformKeyWithNewBeaconHeight(beaconHeight - 1)

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
				var uniqTx []byte
				newInst, uniqTx, err = blockchain.buildInstructionsForIssuingBridgeReq(
					beaconBestState,
					featureStateDB,
					contentStr,
					shardID,
					metaType,
					accumulatedValues,
					accumulatedValues.UniqETHTxsUsed,
					config.Param().EthContractAddressStr,
					"",
					statedb.IsETHTxHashIssued,
				)
				if uniqTx != nil {
					accumulatedValues.UniqETHTxsUsed = append(accumulatedValues.UniqETHTxsUsed, uniqTx)
				}
			case metadata.IssuingBSCRequestMeta:
				var uniqTx []byte
				newInst, uniqTx, err = blockchain.buildInstructionsForIssuingBridgeReq(
					beaconBestState,
					featureStateDB,
					contentStr,
					shardID,
					metaType,
					accumulatedValues,
					accumulatedValues.UniqBSCTxsUsed,
					config.Param().BscContractAddressStr,
					common.BSCPrefix,
					statedb.IsBSCTxHashIssued,
				)
				if uniqTx != nil {
					accumulatedValues.UniqBSCTxsUsed = append(accumulatedValues.UniqBSCTxsUsed, uniqTx)
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

	txHashes := []common.Hash{}

	for _, key := range keys {
		for _, shardState := range shardStates[byte(key)] {
			txHashes = append(txHashes, shardState.PDETxHashes()...)
		}
	}

	pdeStateEnv := pdex.
		NewStateEnvBuilder().
		BuildBeaconHeight(beaconHeight - 1).
		BuildContributionActions(pdeContributionActions).
		BuildPRVRequiredContributionActions(pdePRVRequiredContributionActions).
		BuildTradeActions(pdeTradeActions).
		BuildCrossPoolTradeActions(pdeCrossPoolTradeActions).
		BuildWithdrawalActions(pdeWithdrawalActions).
		BuildFeeWithdrawalActions(pdeFeeWithdrawalActions).
		BuildTxHashes(txHashes).
		Build()

	pdeInstructions, err := beaconBestState.pdeState.BuildInstructions(pdeStateEnv)
	if err != nil {
		Logger.log.Error(err)
		return utils.EmptyStringMatrix, err
	}
	instructions = append(instructions, pdeInstructions...)

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
