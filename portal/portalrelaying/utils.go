package portalrelaying

import (
	"errors"

	bnbrelaying "github.com/incognitochain/incognito-chain/relaying/bnb"
	btcrelaying "github.com/incognitochain/incognito-chain/relaying/btc"
	bnbTypes "github.com/tendermint/tendermint/types"
)

func InitRelayingHeaderChainStateFromDB(bnbChain *bnbrelaying.BNBChainState, btcChain *btcrelaying.BlockChain) (*RelayingHeaderChainState, error) {
	return &RelayingHeaderChainState{
		BNBHeaderChain: bnbChain,
		BTCHeaderChain: btcChain,
	}, nil
}

// GetBNBBlockByHeight gets bnb header by height
func GetBNBBlockByHeight(bnbChainState *bnbrelaying.BNBChainState, blockHeight int64) (*bnbTypes.Block, error) {
	return bnbChainState.GetBNBBlockByHeight(blockHeight)
}

// GetLatestBNBBlockHeight return latest block height of bnb chain
func GetLatestBNBBlockHeight(bnbChainState *bnbrelaying.BNBChainState) (int64, error) {
	if bnbChainState.LatestBlock == nil {
		return int64(0), errors.New("Latest bnb block is nil")
	}
	return bnbChainState.LatestBlock.Height, nil
}
