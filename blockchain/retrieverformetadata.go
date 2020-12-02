package blockchain

import (
	"errors"
	"github.com/incognitochain/incognito-chain/portal/portalprocess"
	bnbrelaying "github.com/incognitochain/incognito-chain/relaying/bnb"
	"github.com/tendermint/tendermint/types"
	"time"
)

func (blockchain *BlockChain) GetStakingAmountShard() uint64 {
	return blockchain.config.ChainParams.StakingAmountShard
}

func (blockchain *BlockChain) GetCentralizedWebsitePaymentAddress(beaconHeight uint64) string {
	if blockchain.config.ChainParams.Net == Testnet || blockchain.config.ChainParams.Net == Testnet2 {
		return blockchain.config.ChainParams.CentralizedWebsitePaymentAddress
	}
	if blockchain.config.ChainParams.Net == Mainnet {
		if beaconHeight >= 677000 {
			// use new address
			return "12RwAheAvvMqrpxviWCV5r6JLS2puiMom3fy6GUCAmPNN1BXnEW4DXpueqMfV66zyAMpurEuegWPGV6U4HR6Mi9KzUiDL4K3uyv1xxF"
		} else if beaconHeight >= 243500 {
			// use new address
			return "12S6jZ6sjJaqsuMJKS6jG7gvE9eHUXGWa2B2dNC7PwyEYJkL6cE53Uzk926HrQMEv2i2oBvKP2GDTC6tzU9dYSVH5X3w9P58VWqux4F"
		} else {
			// use original address
			return blockchain.config.ChainParams.CentralizedWebsitePaymentAddress
		}
	}
	return ""
}

func (blockchain *BlockChain) GetBeaconHeightBreakPointBurnAddr() uint64 {
	return blockchain.config.ChainParams.BeaconHeightBreakPointBurnAddr
}

func (blockchain *BlockChain) GetETHRemoveBridgeSigEpoch() uint64 {
	return blockchain.config.ChainParams.ETHRemoveBridgeSigEpoch
}

func (blockchain *BlockChain) GetBCHeightBreakPointPortalV3() uint64 {
	return blockchain.config.ChainParams.BCHeightBreakPointPortalV3
}

func (blockchain *BlockChain) GetBurningAddress(beaconHeight uint64) string {
	breakPoint := blockchain.GetBeaconHeightBreakPointBurnAddr()
	if beaconHeight == 0 {
		beaconHeight = blockchain.BeaconChain.GetFinalViewHeight()
	}
	if beaconHeight <= breakPoint {
		return burningAddress
	}

	return burningAddress2
}

// convertDurationTimeToBeaconBlocks returns number of beacon blocks corresponding to duration time
func (blockchain *BlockChain) convertDurationTimeToBeaconBlocks(duration time.Duration) uint64 {
	return uint64(duration.Seconds() / blockchain.config.ChainParams.MinBeaconBlockInterval.Seconds())
}

// convertDurationTimeToShardBlocks returns number of shard blocks corresponding to duration time
func (blockchain *BlockChain) convertDurationTimeToShardBlocks(duration time.Duration) uint64 {
	return uint64(duration.Seconds() / blockchain.config.ChainParams.MinShardBlockInterval.Seconds())
}

// convertDurationTimeToBeaconBlocks returns number of beacon blocks corresponding to duration time
func (blockchain *BlockChain) CheckBlockTimeIsReached(recentBeaconHeight, beaconHeight, recentShardHeight, shardHeight uint64, duration time.Duration) bool {
	return (recentBeaconHeight+1)-beaconHeight >= blockchain.convertDurationTimeToBeaconBlocks(duration) &&
		(recentShardHeight+1)-shardHeight >= blockchain.convertDurationTimeToShardBlocks(duration)
}

func (bc *BlockChain) InitRelayingHeaderChainStateFromDB() (*portalprocess.RelayingHeaderChainState, error) {
	bnbChain := bc.GetBNBChainState()
	btcChain := bc.config.BTCChain
	return &portalprocess.RelayingHeaderChainState{
		BNBHeaderChain: bnbChain,
		BTCHeaderChain: btcChain,
	}, nil
}

// GetBNBChainState gets bnb header chain state
func (bc *BlockChain) GetBNBChainState() *bnbrelaying.BNBChainState {
	return bc.config.BNBChainState
}

// GetLatestBNBBlockHeight return latest block height of bnb chain
func (bc *BlockChain) GetLatestBNBBlockHeight() (int64, error) {
	bnbChainState := bc.GetBNBChainState()

	if bnbChainState.LatestBlock == nil {
		return int64(0), errors.New("Latest bnb block is nil")
	}
	return bnbChainState.LatestBlock.Height, nil
}

// GetBNBBlockByHeight gets bnb header by height
func (bc *BlockChain) GetBNBBlockByHeight(blockHeight int64) (*types.Block, error) {
	bnbChainState := bc.GetBNBChainState()
	return bnbChainState.GetBNBBlockByHeight(blockHeight)
}

