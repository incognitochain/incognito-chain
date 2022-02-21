package blockchain

import (
	"reflect"

	"github.com/incognitochain/incognito-chain/metadata"

	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/portal"
	"github.com/incognitochain/incognito-chain/portal/portalrelaying"
	portalprocessv3 "github.com/incognitochain/incognito-chain/portal/portalv3/portalprocess"
	portalprocessv4 "github.com/incognitochain/incognito-chain/portal/portalv4/portalprocess"
)

// Beacon producer for portal protocol
func (blockchain *BlockChain) handlePortalInsts(
	stateDB *statedb.StateDB,
	beaconHeight uint64,
	currentPortalState *portalprocessv3.CurrentPortalState,
	currentPortalStateV4 *portalprocessv4.CurrentPortalStateV4,
	relayingState *portalrelaying.RelayingHeaderChainState,
	rewardForCustodianByEpoch map[common.Hash]uint64,
	portalParams portal.PortalParams,
	pm *portal.PortalManager,
) ([][]string, error) {
	// get shard height of all shards for producer
	shardHeights := map[byte]uint64{}
	for i := 0; i < common.MaxShardNumber; i++ {
		shardHeights[byte(i)] = blockchain.ShardChain[i].multiView.GetBestView().GetHeight()
	}

	epochBlocks := config.Param().EpochParam.NumberOfBlockInEpoch

	return portal.HandlePortalInsts(
		blockchain, stateDB, beaconHeight, shardHeights, currentPortalState, currentPortalStateV4, relayingState,
		rewardForCustodianByEpoch, portalParams, pm, epochBlocks)
}

// Beacon process for portal protocol
func (blockchain *BlockChain) processPortalInstructions(
	portalStateDB *statedb.StateDB, block *types.BeaconBlock,
) (*portalprocessv3.CurrentPortalState, *portalprocessv4.CurrentPortalStateV4, error) {
	// Note: should comment this code if you need to create local chain.
	isSkipPortalV3Ints := false
	if (config.Param().Net == config.LocalNet || config.Param().Net == config.TestnetNet) && block.Header.Height < 1580600 {
		isSkipPortalV3Ints = true
	}
	lastPortalV4State := blockchain.GetBeaconBestState().GetPortalStateV4()
	lastPortalV3State := &portalprocessv3.CurrentPortalState{}
	if !isSkipPortalV3Ints {
		lastPortalV3State = blockchain.GetBeaconBestState().GetPortalStateV3()
	}
	beaconHeight := block.Header.Height - 1
	relayingState, err := portalrelaying.InitRelayingHeaderChainStateFromDB(blockchain.GetBNBHeaderChain(), blockchain.GetBTCHeaderChain())
	if err != nil {
		Logger.log.Error(err)
		return lastPortalV3State, lastPortalV4State, nil
	}
	portalParams := portal.GetPortalParams()
	pm := portal.NewPortalManager()
	epoch := config.Param().EpochParam.NumberOfBlockInEpoch

	newPortalV3State, newPortalV4State, err := portal.ProcessPortalInsts(
		blockchain, lastPortalV3State, lastPortalV4State, portalStateDB, relayingState, *portalParams, beaconHeight,
		block.Body.Instructions, pm, epoch, isSkipPortalV3Ints)
	if err != nil {
		Logger.log.Error(err)
	}

	return newPortalV3State, newPortalV4State, nil
}

func getDiffPortalStateV3(
	previous *portalprocessv3.CurrentPortalState, current *portalprocessv3.CurrentPortalState,
) (diffState *portalprocessv3.CurrentPortalState) {
	if current == nil {
		return nil
	}
	if previous == nil {
		return current
	}

	diffState = &portalprocessv3.CurrentPortalState{
		CustodianPoolState:         map[string]*statedb.CustodianState{},
		WaitingPortingRequests:     map[string]*statedb.WaitingPortingRequest{},
		WaitingRedeemRequests:      map[string]*statedb.RedeemRequest{},
		MatchedRedeemRequests:      map[string]*statedb.RedeemRequest{},
		FinalExchangeRatesState:    nil,
		LiquidationPool:            map[string]*statedb.LiquidationPool{},
		LockedCollateralForRewards: nil,
		ExchangeRatesRequests:      map[string]*metadata.ExchangeRatesRequestStatus{},
	}

	for k, v := range current.CustodianPoolState {
		if _v, ok := previous.CustodianPoolState[k]; !ok || !reflect.DeepEqual(_v, v) {
			diffState.CustodianPoolState[k] = v
		}
	}
	for k, v := range current.WaitingPortingRequests {
		if _v, ok := previous.WaitingPortingRequests[k]; !ok || !reflect.DeepEqual(_v, v) {
			diffState.WaitingPortingRequests[k] = v
		}
	}
	for k, v := range current.WaitingRedeemRequests {
		if _v, ok := previous.WaitingRedeemRequests[k]; !ok || !reflect.DeepEqual(_v, v) {
			diffState.WaitingRedeemRequests[k] = v
		}
	}
	for k, v := range current.MatchedRedeemRequests {
		if _v, ok := previous.MatchedRedeemRequests[k]; !ok || !reflect.DeepEqual(_v, v) {
			diffState.MatchedRedeemRequests[k] = v
		}
	}
	for k, v := range current.LiquidationPool {
		if _v, ok := previous.LiquidationPool[k]; !ok || !reflect.DeepEqual(_v, v) {
			diffState.LiquidationPool[k] = v
		}
	}
	if !reflect.DeepEqual(current.FinalExchangeRatesState, previous.FinalExchangeRatesState) {
		diffState.FinalExchangeRatesState = current.FinalExchangeRatesState
	}
	if !reflect.DeepEqual(current.LockedCollateralForRewards, previous.LockedCollateralForRewards) {
		diffState.LockedCollateralForRewards = current.LockedCollateralForRewards
	}

	return diffState
}

func getDiffPortalStateV4(
	previous *portalprocessv4.CurrentPortalStateV4, current *portalprocessv4.CurrentPortalStateV4,
) (diffState *portalprocessv4.CurrentPortalStateV4) {
	if current == nil {
		return nil
	}
	if previous == nil {
		return current
	}

	diffState = &portalprocessv4.CurrentPortalStateV4{
		UTXOs:                                map[string]map[string]*statedb.UTXO{},
		ShieldingExternalTx:                  map[string]map[string]*statedb.ShieldingRequest{},
		WaitingUnshieldRequests:              map[string]map[string]*statedb.WaitingUnshieldRequest{},
		ProcessedUnshieldRequests:            map[string]map[string]*statedb.ProcessedUnshieldRequestBatch{},
		DeletedUTXOKeyHashes:                 current.DeletedUTXOKeyHashes,
		DeletedWaitingUnshieldReqKeyHashes:   current.DeletedWaitingUnshieldReqKeyHashes,
		DeletedProcessedUnshieldReqKeyHashes: current.DeletedProcessedUnshieldReqKeyHashes,
	}

	for k, v := range current.UTXOs {
		diffState.UTXOs[k] = map[string]*statedb.UTXO{}
		for _k, _v := range v {
			if m, ok := previous.UTXOs[k][_k]; !ok || !reflect.DeepEqual(m, _v) {
				diffState.UTXOs[k][_k] = _v
			}
		}
	}
	for k, v := range current.ShieldingExternalTx {
		diffState.ShieldingExternalTx[k] = map[string]*statedb.ShieldingRequest{}
		for _k, _v := range v {
			if m, ok := previous.ShieldingExternalTx[k][_k]; !ok || !reflect.DeepEqual(m, _v) {
				diffState.ShieldingExternalTx[k][_k] = _v
			}
		}
	}
	for k, v := range current.WaitingUnshieldRequests {
		diffState.WaitingUnshieldRequests[k] = map[string]*statedb.WaitingUnshieldRequest{}
		for _k, _v := range v {
			if m, ok := previous.WaitingUnshieldRequests[k][_k]; !ok || !reflect.DeepEqual(m, _v) {
				diffState.WaitingUnshieldRequests[k][_k] = _v
			}
		}
	}
	for k, v := range current.ProcessedUnshieldRequests {
		diffState.ProcessedUnshieldRequests[k] = map[string]*statedb.ProcessedUnshieldRequestBatch{}
		for _k, _v := range v {
			if m, ok := previous.ProcessedUnshieldRequests[k][_k]; !ok || !reflect.DeepEqual(m, _v) {
				diffState.ProcessedUnshieldRequests[k][_k] = _v
			}
		}
	}

	return diffState
}
