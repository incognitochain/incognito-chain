package blockchain

import (
	"encoding/base64"
	"encoding/json"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"math/big"
	"sort"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
)

// build instructions at beacon chain before syncing to shards
func (blockchain *BlockChain) collectStatefulActions(
	shardBlockInstructions [][]string,
) [][]string {
	// stateful instructions are dependently processed with results of instructioins before them in shards2beacon blocks
	statefulInsts := [][]string{}
	for _, inst := range shardBlockInstructions {
		if len(inst) < 2 {
			continue
		}
		if inst[0] == SetAction || inst[0] == StakeAction || inst[0] == SwapAction || inst[0] == RandomAction || inst[0] == AssignAction {
			continue
		}

		metaType, err := strconv.Atoi(inst[0])
		if err != nil {
			Logger.log.Error(err)
			continue
		}
		switch metaType {
		case metadata.IssuingRequestMeta, metadata.IssuingETHRequestMeta,
			metadata.PDEContributionMeta, metadata.PDETradeRequestMeta,
			metadata.PDEWithdrawalRequestMeta:
			statefulInsts = append(statefulInsts, inst)

		default:
			continue
		}
	}
	return statefulInsts
}

func groupPDEActionsByShardID(
	pdeActionsByShardID map[byte][][]string,
	action []string,
	shardID byte,
) map[byte][][]string {
	_, found := pdeActionsByShardID[shardID]
	if !found {
		pdeActionsByShardID[shardID] = [][]string{action}
	} else {
		pdeActionsByShardID[shardID] = append(pdeActionsByShardID[shardID], action)
	}
	return pdeActionsByShardID
}

func (blockchain *BlockChain) buildStatefulInstructions(stateDB *statedb.StateDB, statefulActionsByShardID map[byte][][]string, beaconHeight uint64) [][]string {
	currentPDEState, err := InitCurrentPDEStateFromDB(stateDB, beaconHeight-1)
	if err != nil {
		Logger.log.Error(err)
	}
	accumulatedValues := &metadata.AccumulatedValues{
		UniqETHTxsUsed:   [][]byte{},
		DBridgeTokenPair: map[string][]byte{},
		CBridgeTokens:    []*common.Hash{},
	}
	instructions := [][]string{}
	pdeContributionActionsByShardID := map[byte][][]string{}
	pdeTradeActionsByShardID := map[byte][][]string{}
	pdeWithdrawalActionsByShardID := map[byte][][]string{}

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
			switch metaType {
			case metadata.IssuingRequestMeta:
				newInst, err = blockchain.buildInstructionsForIssuingReq(stateDB, contentStr, shardID, metaType, accumulatedValues)

			case metadata.IssuingETHRequestMeta:
				newInst, err = blockchain.buildInstructionsForIssuingETHReq(stateDB, contentStr, shardID, metaType, accumulatedValues)

			case metadata.PDEContributionMeta:
				pdeContributionActionsByShardID = groupPDEActionsByShardID(
					pdeContributionActionsByShardID,
					action,
					shardID,
				)
			case metadata.PDETradeRequestMeta:
				pdeTradeActionsByShardID = groupPDEActionsByShardID(
					pdeTradeActionsByShardID,
					action,
					shardID,
				)
			case metadata.PDEWithdrawalRequestMeta:
				pdeWithdrawalActionsByShardID = groupPDEActionsByShardID(
					pdeWithdrawalActionsByShardID,
					action,
					shardID,
				)
			default:
				continue
			}
			if err != nil {
				Logger.log.Error(err)
				continue
			}
			if len(newInst) > 0 {
				instructions = append(instructions, newInst...)
			}
		}
	}
	pdeInsts, err := blockchain.handlePDEInsts(
		beaconHeight-1, currentPDEState,
		pdeContributionActionsByShardID,
		pdeTradeActionsByShardID,
		pdeWithdrawalActionsByShardID,
	)
	if err != nil {
		Logger.log.Error(err)
		return instructions
	}
	if len(pdeInsts) > 0 {
		instructions = append(instructions, pdeInsts...)
	}
	return instructions
}

func sortPDETradeInstsByFee(
	beaconHeight uint64,
	currentPDEState *CurrentPDEState,
	pdeTradeActionsByShardID map[byte][][]string,
) []metadata.PDETradeRequestAction {
	tradesByPairs := make(map[string][]metadata.PDETradeRequestAction)

	var keys []int
	for k := range pdeTradeActionsByShardID {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)
	for _, value := range keys {
		shardID := byte(value)
		actions := pdeTradeActionsByShardID[shardID]
		for _, action := range actions {
			contentStr := action[1]
			contentBytes, err := base64.StdEncoding.DecodeString(contentStr)
			if err != nil {
				Logger.log.Errorf("ERROR: an error occured while decoding content string of pde trade action: %+v", err)
				continue
			}
			var pdeTradeReqAction metadata.PDETradeRequestAction
			err = json.Unmarshal(contentBytes, &pdeTradeReqAction)
			if err != nil {
				Logger.log.Errorf("ERROR: an error occured while unmarshaling pde trade action: %+v", err)
				continue
			}
			tradeMeta := pdeTradeReqAction.Meta
			poolPairKey := string(rawdbv2.BuildPDEPoolForPairKey(beaconHeight, tradeMeta.TokenIDToBuyStr, tradeMeta.TokenIDToSellStr))
			tradesByPair, found := tradesByPairs[poolPairKey]
			if !found {
				tradesByPairs[poolPairKey] = []metadata.PDETradeRequestAction{pdeTradeReqAction}
			} else {
				tradesByPairs[poolPairKey] = append(tradesByPair, pdeTradeReqAction)
			}
		}
	}

	notExistingPairTradeActions := []metadata.PDETradeRequestAction{}
	sortedExistingPairTradeActions := []metadata.PDETradeRequestAction{}

	var ppKeys []string
	for k := range tradesByPairs {
		ppKeys = append(ppKeys, k)
	}
	sort.Strings(ppKeys)
	for _, poolPairKey := range ppKeys {
		tradeActions := tradesByPairs[poolPairKey]
		poolPair, found := currentPDEState.PDEPoolPairs[poolPairKey]
		if !found || poolPair == nil {
			notExistingPairTradeActions = append(notExistingPairTradeActions, tradeActions...)
			continue
		}
		if poolPair.Token1PoolValue == 0 || poolPair.Token2PoolValue == 0 {
			notExistingPairTradeActions = append(notExistingPairTradeActions, tradeActions...)
			continue
		}

		// sort trade actions by trading fee
		sort.Slice(tradeActions, func(i, j int) bool {
			// comparing a/b to c/d is equivalent with comparing a*d to c*b
			firstItemProportion := big.NewInt(0)
			firstItemProportion.Mul(
				big.NewInt(int64(tradeActions[i].Meta.TradingFee)),
				big.NewInt(int64(tradeActions[j].Meta.SellAmount)),
			)
			secondItemProportion := big.NewInt(0)
			secondItemProportion.Mul(
				big.NewInt(int64(tradeActions[j].Meta.TradingFee)),
				big.NewInt(int64(tradeActions[i].Meta.SellAmount)),
			)
			return firstItemProportion.Cmp(secondItemProportion) == 1
		})
		sortedExistingPairTradeActions = append(sortedExistingPairTradeActions, tradeActions...)
	}
	return append(sortedExistingPairTradeActions, notExistingPairTradeActions...)
}

func (blockchain *BlockChain) handlePDEInsts(
	beaconHeight uint64,
	currentPDEState *CurrentPDEState,
	pdeContributionActionsByShardID map[byte][][]string,
	pdeTradeActionsByShardID map[byte][][]string,
	pdeWithdrawalActionsByShardID map[byte][][]string,
) ([][]string, error) {
	instructions := [][]string{}
	sortedTradesActions := sortPDETradeInstsByFee(
		beaconHeight,
		currentPDEState,
		pdeTradeActionsByShardID,
	)
	for _, tradeAction := range sortedTradesActions {
		actionContentBytes, _ := json.Marshal(tradeAction)
		actionContentBase64Str := base64.StdEncoding.EncodeToString(actionContentBytes)
		newInst, err := blockchain.buildInstructionsForPDETrade(actionContentBase64Str, tradeAction.ShardID, metadata.PDETradeRequestMeta, currentPDEState, beaconHeight)
		if err != nil {
			Logger.log.Error(err)
			continue
		}
		if len(newInst) > 0 {
			instructions = append(instructions, newInst...)
		}
	}

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
			newInst, err := blockchain.buildInstructionsForPDEContribution(contentStr, shardID, metadata.PDEContributionMeta, currentPDEState, beaconHeight)
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
