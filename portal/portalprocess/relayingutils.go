package portalprocess

import (
	"encoding/base64"
	"encoding/json"
	"github.com/incognitochain/incognito-chain/basemeta"
	"github.com/incognitochain/incognito-chain/common"
	metadata2 "github.com/incognitochain/incognito-chain/portal/metadata"
	bnbrelaying "github.com/incognitochain/incognito-chain/relaying/bnb"
	btcrelaying "github.com/incognitochain/incognito-chain/relaying/btc"
	"github.com/tendermint/tendermint/types"
	"strconv"
)

type relayingChain struct {
	actions [][]string
}
type relayingBNBChain struct {
	*relayingChain
}
type relayingBTCChain struct {
	*relayingChain
}

func (rChain *relayingChain) getActions() [][]string {
	return rChain.actions
}
func (rChain *relayingChain) putAction(action []string) {
	rChain.actions = append(rChain.actions, action)
}

// buildHeaderRelayingInst builds a new instruction from action received from ShardToBeaconBlock
func (rChain *relayingChain) buildHeaderRelayingInst(
	senderAddressStr string,
	header string,
	blockHeight uint64,
	metaType int,
	shardID byte,
	txReqID common.Hash,
	status string,
) []string {
	headerRelayingContent := metadata2.RelayingHeaderContent{
		IncogAddressStr: senderAddressStr,
		Header:          header,
		TxReqID:         txReqID,
		BlockHeight:     blockHeight,
	}
	headerRelayingContentBytes, _ := json.Marshal(headerRelayingContent)
	return []string{
		strconv.Itoa(metaType),
		strconv.Itoa(int(shardID)),
		status,
		string(headerRelayingContentBytes),
	}
}

func (rbnbChain *relayingBNBChain) buildRelayingInst(
	bc basemeta.ChainRetriever,
	relayingHeaderAction metadata2.RelayingHeaderAction,
	relayingHeaderChain *RelayingHeaderChainState,
) [][]string {
	meta := relayingHeaderAction.Meta
	// parse bnb block header
	headerBytes, err := base64.StdEncoding.DecodeString(meta.Header)
	if err != nil {
		Logger.log.Errorf("Error - [buildInstructionsForBNBHeaderRelaying]: Cannot decode header string.%v\n", err)
		inst := rbnbChain.buildHeaderRelayingInst(
			relayingHeaderAction.Meta.IncogAddressStr,
			relayingHeaderAction.Meta.Header,
			relayingHeaderAction.Meta.BlockHeight,
			relayingHeaderAction.Meta.Type,
			relayingHeaderAction.ShardID,
			relayingHeaderAction.TxReqID,
			common.RelayingHeaderRejectedChainStatus,
		)
		return [][]string{inst}
	}

	var newBlock types.Block
	err = json.Unmarshal(headerBytes, &newBlock)
	if err != nil {
		Logger.log.Errorf("Error - [buildInstructionsForBNBHeaderRelaying]: Cannot unmarshal header.%v\n", err)
		inst := rbnbChain.buildHeaderRelayingInst(
			relayingHeaderAction.Meta.IncogAddressStr,
			relayingHeaderAction.Meta.Header,
			relayingHeaderAction.Meta.BlockHeight,
			relayingHeaderAction.Meta.Type,
			relayingHeaderAction.ShardID,
			relayingHeaderAction.TxReqID,
			common.RelayingHeaderRejectedChainStatus,
		)
		return [][]string{inst}
	}

	if newBlock.Header.Height != int64(relayingHeaderAction.Meta.BlockHeight) {
		Logger.log.Errorf("Error - [buildInstructionsForBNBHeaderRelaying]: Block height in metadata is unmatched with block height in new header.")
		inst := rbnbChain.buildHeaderRelayingInst(
			relayingHeaderAction.Meta.IncogAddressStr,
			relayingHeaderAction.Meta.Header,
			relayingHeaderAction.Meta.BlockHeight,
			relayingHeaderAction.Meta.Type,
			relayingHeaderAction.ShardID,
			relayingHeaderAction.TxReqID,
			common.RelayingHeaderRejectedChainStatus,
		)
		return [][]string{inst}
	}

	inst := rbnbChain.buildHeaderRelayingInst(
		relayingHeaderAction.Meta.IncogAddressStr,
		relayingHeaderAction.Meta.Header,
		relayingHeaderAction.Meta.BlockHeight,
		relayingHeaderAction.Meta.Type,
		relayingHeaderAction.ShardID,
		relayingHeaderAction.TxReqID,
		common.RelayingHeaderConsideringChainStatus,
	)
	return [][]string{inst}
}

func (rbtcChain *relayingBTCChain) buildRelayingInst(
	bc basemeta.ChainRetriever,
	relayingHeaderAction metadata2.RelayingHeaderAction,
	relayingState *RelayingHeaderChainState,
) [][]string {
	Logger.log.Info("[BTC Relaying] - Processing buildRelayingInst...")
	inst := rbtcChain.buildHeaderRelayingInst(
		relayingHeaderAction.Meta.IncogAddressStr,
		relayingHeaderAction.Meta.Header,
		relayingHeaderAction.Meta.BlockHeight,
		relayingHeaderAction.Meta.Type,
		relayingHeaderAction.ShardID,
		relayingHeaderAction.TxReqID,
		common.RelayingHeaderConsideringChainStatus,
	)
	return [][]string{inst}
}



type RelayingHeaderChainState struct {
	BNBHeaderChain *bnbrelaying.BNBChainState
	BTCHeaderChain *btcrelaying.BlockChain
}

