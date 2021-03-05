package blockchain

import (
	"errors"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/metadata/rpccaller"
	"github.com/incognitochain/incognito-chain/portal"
	"github.com/incognitochain/incognito-chain/portal/portalv3"
	"github.com/incognitochain/incognito-chain/portal/portalv4"
	btcrelaying "github.com/incognitochain/incognito-chain/relaying/btc"
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


/* ================== Retriever for portal v3 ================== */
// GetPortalParams returns portal params in beaconheight
func (blockchain *BlockChain) GetPortalParams() portal.PortalParams {
	return blockchain.GetConfig().ChainParams.PortalParams
}

func (blockchain *BlockChain) GetPortalParamsV3(beaconHeight uint64) portalv3.PortalParams {
	return blockchain.GetConfig().ChainParams.PortalParams.GetPortalParamsV3(beaconHeight)
}

func (blockchain *BlockChain) GetPortalParamsV4(beaconHeight uint64) portalv4.PortalParams {
	return blockchain.GetConfig().ChainParams.PortalParams.GetPortalParamsV4(beaconHeight)
}

func (blockchain *BlockChain) GetMinAmountPortalToken(tokenIDStr string, beaconHeight uint64) (uint64, error) {
	return blockchain.GetPortalParamsV3(beaconHeight).GetMinAmountPortalToken(tokenIDStr)
}

// IsPortalToken check tokenIDStr is the valid portal token on portal v3 or not
func (blockchain *BlockChain) IsPortalToken(beaconHeight uint64, tokenIDStr string) bool {
	return blockchain.GetPortalParamsV3(beaconHeight).IsPortalToken(tokenIDStr)
}

func (blockchain *BlockChain) IsValidPortalRemoteAddress(tokenIDStr string, remoteAddr string, beaconHeight uint64) (bool, error) {
	portalTokens := blockchain.GetPortalParamsV3(beaconHeight).PortalTokens
	portalToken, ok := portalTokens[tokenIDStr]
	if !ok || portalToken == nil {
		return false, errors.New("Portal token ID is invalid")
	}
	return portalToken.IsValidRemoteAddress(remoteAddr, blockchain)
}

func (blockchain *BlockChain) GetBCHeightBreakPointPortalV3() uint64 {
	return blockchain.config.ChainParams.BCHeightBreakPointPortalV3
}

func (blockchain *BlockChain) GetBNBChainID() string {
	return blockchain.GetConfig().ChainParams.PortalParams.RelayingParam.BNBRelayingHeaderChainID
}

func (blockchain *BlockChain) GetBTCChainID() string {
	return blockchain.GetConfig().ChainParams.PortalParams.RelayingParam.BTCRelayingHeaderChainID
}

func (blockchain *BlockChain) GetBTCHeaderChain() *btcrelaying.BlockChain {
	return blockchain.GetConfig().BTCChain
}

func (blockchain *BlockChain) GetPortalFeederAddress(beaconHeight uint64) string {
	portalParams := blockchain.GetPortalParamsV3(beaconHeight)
	return portalParams.PortalFeederAddress
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

func (blockchain *BlockChain) CheckBlockTimeIsReachedByBeaconHeight(recentBeaconHeight, beaconHeight uint64, duration time.Duration) bool {
	return (recentBeaconHeight+1)-beaconHeight >= blockchain.convertDurationTimeToBeaconBlocks(duration)
}

func (blockchain *BlockChain) IsPortalExchangeRateToken(beaconHeight uint64, tokenIDStr string) bool {
	return blockchain.GetPortalParamsV3(beaconHeight).IsPortalExchangeRateToken(tokenIDStr)
}

func (blockchain *BlockChain) IsSupportedTokenCollateralV3(beaconHeight uint64, externalTokenID string) bool {
	portalParams := blockchain.GetPortalParamsV3(beaconHeight)
	tokenIDs := []string{}
	for _, col := range portalParams.SupportedCollateralTokens {
		tokenIDs = append(tokenIDs, col.ExternalTokenID)
	}

	isSupported, _ := common.SliceExists(tokenIDs, externalTokenID)
	return isSupported
}

func (blockchain *BlockChain) GetPortalETHContractAddrStr(beaconHeight uint64) string {
	portalParams := blockchain.GetPortalParamsV3(beaconHeight)
	return portalParams.PortalETHContractAddressStr
}

// GetBNBHeader calls RPC to fullnode bnb to get latest bnb block height
func (blockchain *BlockChain) GetLatestBNBBlkHeight() (int64, error) {
	portalRelayingParams := blockchain.GetPortalParams().RelayingParam
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

// GetBNBHeader calls RPC to fullnode bnb to get bnb header by block height
func (blockchain *BlockChain) GetBNBHeader(
	blockHeight int64,
) (*types.Header, error) {
	portalRelayingParams := blockchain.GetPortalParams().RelayingParam
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

// Validate portal remote addresses for portal tokens (BTC, BNB)
func (blockchain *BlockChain) ValidatePortalRemoteAddresses(remoteAddresses map[string]string, beaconHeight uint64) (bool, error) {
	if len(remoteAddresses) == 0 {
		return false, errors.New("remote addresses should be at least one address")
	}
	for tokenID, remoteAddr := range remoteAddresses {
		if !blockchain.IsPortalToken(beaconHeight, tokenID) {
			return false, errors.New("TokenID in remote address is invalid")
		}
		if len(remoteAddr) == 0 {
			return false, errors.New("Remote address is invalid")
		}
		isValid, err := blockchain.IsValidPortalRemoteAddress(tokenID, remoteAddr, beaconHeight)
		if !isValid || err != nil {
			return false, fmt.Errorf("Remote address %v is not a valid address of tokenID %v - Error %v", remoteAddr, tokenID, err)
		}
	}

	return true, nil
}

func (blockchain *BlockChain) GetEnableFeatureFlags() map[int]uint64 {
	return blockchain.GetChainParams().EnableFeatureFlags
}

func (blockchain *BlockChain) IsEnableFeature(featureFlag int, epoch uint64) bool {
	enableFeatureFlags := blockchain.config.ChainParams.EnableFeatureFlags
	if enableFeatureFlags[featureFlag] == 0 {
		return false
	}
	return epoch >= enableFeatureFlags[featureFlag]
}

func (blockchain *BlockChain) GetPortalV4MinUnshieldAmount (tokenIDStr string, beaconHeight uint64) uint64 {
	return blockchain.GetPortalParamsV4(beaconHeight).MinUnshieldAmts[tokenIDStr]
}
