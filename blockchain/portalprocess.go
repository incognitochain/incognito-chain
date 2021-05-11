package blockchain

import (
	"reflect"

	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
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

	epochBlocks := blockchain.config.ChainParams.Epoch

	return portal.HandlePortalInsts(
		blockchain, stateDB, beaconHeight, shardHeights, currentPortalState, currentPortalStateV4, relayingState,
		rewardForCustodianByEpoch, portalParams, pm, epochBlocks)
}

// Beacon process for portal protocol
func (blockchain *BlockChain) processPortalInstructions(
	portalStateDB *statedb.StateDB, block *types.BeaconBlock,
) (*portalprocessv4.CurrentPortalStateV4, error) {
	// Note: should comment this code if you need to create local chain.
	isSkipPortalV3Ints := false
	if (blockchain.config.ChainParams.Net == Testnet || blockchain.config.ChainParams.Net == Testnet2) && block.Header.Height < 1580600 {
		isSkipPortalV3Ints = true
	}
	// get the last portalv4 state
	clonedBeaconBestState, err := blockchain.GetClonedBeaconBestState()
	if err != nil {
		Logger.log.Error(err)
		return nil, nil
	}
	lastPortalV4State := clonedBeaconBestState.portalStateV4
	beaconHeight := block.Header.Height - 1
	relayingState, err := portalrelaying.InitRelayingHeaderChainStateFromDB(blockchain.GetBNBHeaderChain(), blockchain.GetBTCHeaderChain())
	if err != nil {
		Logger.log.Error(err)
		return lastPortalV4State, nil
	}
	portalParams := blockchain.GetPortalParams()
	pm := portal.NewPortalManager()
	epoch := blockchain.config.ChainParams.Epoch

	newPortalV4State, err := portal.ProcessPortalInsts(
		blockchain, lastPortalV4State, portalStateDB, relayingState, portalParams, beaconHeight,
		block.Body.Instructions, pm, epoch, isSkipPortalV3Ints,
	)
	if err != nil {
		Logger.log.Error(err)
	}

	return newPortalV4State, nil
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
		UTXOs:                     map[string]map[string]*statedb.UTXO{},
		ShieldingExternalTx:       map[string]map[string]*statedb.ShieldingRequest{},
		WaitingUnshieldRequests:   map[string]map[string]*statedb.WaitingUnshieldRequest{},
		ProcessedUnshieldRequests: map[string]map[string]*statedb.ProcessedUnshieldRequestBatch{},
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
