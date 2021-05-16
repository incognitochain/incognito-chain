package blockchain

import (
	"errors"

	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/portal"
	"github.com/incognitochain/incognito-chain/portal/portalrelaying"
	portalprocessv3 "github.com/incognitochain/incognito-chain/portal/portalv3/portalprocess"
	bnbTypes "github.com/tendermint/tendermint/types"
)

// TODO: move to portalrelaying package
func (bc *BlockChain) InitRelayingHeaderChainStateFromDB() (*portalrelaying.RelayingHeaderChainState, error) {
	bnbChain := bc.config.BNBChainState
	btcChain := bc.config.BTCChain
	return &portalrelaying.RelayingHeaderChainState{
		BNBHeaderChain: bnbChain,
		BTCHeaderChain: btcChain,
	}, nil
}

// GetBNBBlockByHeight gets bnb header by height
func (bc *BlockChain) GetBNBBlockByHeight(blockHeight int64) (*bnbTypes.Block, error) {
	bnbChainState := bc.config.BNBChainState
	return bnbChainState.GetBNBBlockByHeight(blockHeight)
}

// GetLatestBNBBlockHeight return latest block height of bnb chain
func (bc *BlockChain) GetLatestBNBBlockHeight() (int64, error) {
	bnbChainState := bc.config.BNBChainState

	if bnbChainState.LatestBlock == nil {
		return int64(0), errors.New("Latest bnb block is nil")
	}
	return bnbChainState.LatestBlock.Height, nil
}

// Beacon producer for portal protocol
func (blockchain *BlockChain) handlePortalInsts(
	stateDB *statedb.StateDB,
	beaconHeight uint64,
	currentPortalState *portalprocessv3.CurrentPortalState,
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
		blockchain, stateDB, beaconHeight, shardHeights, currentPortalState, relayingState,
		rewardForCustodianByEpoch, portalParams, pm, epochBlocks)
}

// Beacon process for portal protocol
func (blockchain *BlockChain) processPortalInstructions(portalStateDB *statedb.StateDB, block *types.BeaconBlock) error {
	// Note: should comment this code if you need to create local chain.
	isSkipPortalV3Ints := false
	if config.Param().Net == config.TestnetNet && block.Header.Height < 1580600 {
		isSkipPortalV3Ints = true
	}
	beaconHeight := block.Header.Height - 1
	relayingState, err := blockchain.InitRelayingHeaderChainStateFromDB()
	if err != nil {
		Logger.log.Error(err)
		return nil
	}
	portalParams := portal.GetPortalParams()
	pm := portal.NewPortalManager()
	epoch := config.Param().EpochParam.NumberOfBlockInEpoch

	err = portal.ProcessPortalInsts(
		blockchain, portalStateDB, relayingState, *portalParams, beaconHeight, block.Body.Instructions, pm, epoch, isSkipPortalV3Ints)
	if err != nil {
		Logger.log.Error(err)
	}

	return nil
}
