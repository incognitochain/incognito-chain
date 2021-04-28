package blockchain

import (
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
func (blockchain *BlockChain) processPortalInstructions(portalStateDB *statedb.StateDB, block *types.BeaconBlock) error {
	// Note: should comment this code if you need to create local chain.
	isSkipPortalV3Ints := false
	if (blockchain.config.ChainParams.Net == Testnet || blockchain.config.ChainParams.Net == Testnet2) && block.Header.Height < 1580600 {
		isSkipPortalV3Ints = true
	}
	beaconHeight := block.Header.Height - 1
	relayingState, err := portalrelaying.InitRelayingHeaderChainStateFromDB(blockchain.GetBNBHeaderChain(), blockchain.GetBTCHeaderChain())
	if err != nil {
		Logger.log.Error(err)
		return nil
	}
	portalParams := blockchain.GetPortalParams()
	pm := portal.NewPortalManager()
	epoch := blockchain.config.ChainParams.Epoch

	err = portal.ProcessPortalInsts(
		blockchain, portalStateDB, relayingState, portalParams, beaconHeight, block.Body.Instructions, pm, epoch, isSkipPortalV3Ints)
	if err != nil {
		Logger.log.Error(err)
	}

	return nil
}
