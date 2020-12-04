package blockchain

import (
	"errors"
	"fmt"
	"github.com/incognitochain/incognito-chain/metadata/rpccaller"
	"github.com/incognitochain/incognito-chain/portal/portalprocess"
	bnbrelaying "github.com/incognitochain/incognito-chain/relaying/bnb"
	"github.com/tendermint/tendermint/rpc/client"
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


// IsSupportedTokenCollateralV3 check externalTokenID is the supported token collateral on portal v3 or not
func (blockchain *BlockChain) IsSupportedTokenCollateralV3(beaconHeight uint64, externalTokenID string) bool {
	return blockchain.GetPortalParams(beaconHeight).IsSupportedTokenCollateralV3(externalTokenID)
}

// IsPortalExchangeRateToken check tokenID is the valid token can be relayed the rate on portal v3 or not
func (blockchain *BlockChain) IsPortalExchangeRateToken(beaconHeight uint64, tokenID string) bool {
	return blockchain.GetPortalParams(beaconHeight).IsPortalExchangeRateToken(tokenID)
}

// IsPortalExchangeRateToken check tokenIDStr is the valid portal token on portal v3 or not
func (blockchain *BlockChain) IsPortalToken(beaconHeight uint64, tokenIDStr string) bool {
	return blockchain.GetPortalParams(beaconHeight).IsPortalExchangeRateToken(tokenIDStr)
}

// GetBNBHeader calls RPC to fullnode bnb to get bnb header by block height
func (blockchain *BlockChain) GetBNBHeader(
	blockHeight int64,
) (*types.Header, error) {
	portalRelayingParams := blockchain.GetPortalParams(0).RelayingParams
	bnbFullNodeAddress := rpccaller.BuildRPCServerAddress(
		portalRelayingParams.BNBFullNodeProtocol,
		portalRelayingParams.BNBFullNodeHost,
		portalRelayingParams.BNBFullNodePort,
	)
	bnbClient := client.NewHTTP(bnbFullNodeAddress, "/websocket")
	result, err := bnbClient.Block(&blockHeight)
	if err != nil {
		Logger.log.Errorf("An error occured during calling status method: %s", err)
		return nil, fmt.Errorf("error occured during calling status method: %s", err)
	}
	return &result.Block.Header, nil
}

// GetBNBDataHash calls RPC to fullnode bnb to get bnb data hash in header
func (blockchain *BlockChain) GetBNBDataHash(
	blockHeight int64,
) ([]byte, error) {
	header, err := blockchain.GetBNBHeader(blockHeight)
	if err != nil {
		return nil, err
	}
	if header.DataHash == nil {
		return nil, errors.New("Data hash is nil")
	}
	return header.DataHash, nil
}

// GetBNBHeader calls RPC to fullnode bnb to get latest bnb block height
func (blockchain *BlockChain) GetLatestBNBBlkHeight() (int64, error) {
	portalRelayingParams := blockchain.GetPortalParams(0).RelayingParams
	bnbFullNodeAddress := rpccaller.BuildRPCServerAddress(
		portalRelayingParams.BNBFullNodeProtocol,
		portalRelayingParams.BNBFullNodeHost,
		portalRelayingParams.BNBFullNodePort,
	)
	bnbClient := client.NewHTTP(bnbFullNodeAddress, "/websocket")
	result, err := bnbClient.Status()
	if err != nil {
		Logger.log.Errorf("An error occured during calling status method: %s", err)
		return 0, fmt.Errorf("error occured during calling status method: %s", err)
	}
	return result.SyncInfo.LatestBlockHeight, nil
}