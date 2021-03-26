package portalrelaying

import (
	"encoding/base64"
	"encoding/json"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/tendermint/tendermint/types"
	"strconv"
)

type RelayingProcessor interface {
	GetActions() [][]string
	PutAction(action []string)
	BuildRelayingInst(
		bc metadata.ChainRetriever,
		relayingHeaderAction metadata.RelayingHeaderAction,
		relayingState *RelayingHeaderChainState,
	) [][]string
	BuildHeaderRelayingInst(
		senderAddressStr string,
		header string,
		blockHeight uint64,
		metaType int,
		shardID byte,
		txReqID common.Hash,
		status string,
	) []string
}

type RelayingChain struct {
	Actions [][]string
}
type RelayingBNBChain struct {
	*RelayingChain
}
type RelayingBTCChain struct {
	*RelayingChain
}

func (rChain *RelayingChain) GetActions() [][]string {
	return rChain.Actions
}
func (rChain *RelayingChain) PutAction(action []string) {
	rChain.Actions = append(rChain.Actions, action)
}

// buildHeaderRelayingInst builds a new instruction from action received from ShardToBeaconBlock
func (rChain *RelayingChain) BuildHeaderRelayingInst(
	senderAddressStr string,
	header string,
	blockHeight uint64,
	metaType int,
	shardID byte,
	txReqID common.Hash,
	status string,
) []string {
	headerRelayingContent := metadata.RelayingHeaderContent{
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

func (rbnbChain *RelayingBNBChain) BuildRelayingInst(
	bc metadata.ChainRetriever,
	relayingHeaderAction metadata.RelayingHeaderAction,
	relayingHeaderChain *RelayingHeaderChainState,
) [][]string {
	meta := relayingHeaderAction.Meta
	// parse bnb block header
	headerBytes, err := base64.StdEncoding.DecodeString(meta.Header)
	if err != nil {
		Logger.log.Errorf("Error - [buildInstructionsForBNBHeaderRelaying]: Cannot decode header string.%v\n", err)
		inst := rbnbChain.BuildHeaderRelayingInst(
			relayingHeaderAction.Meta.IncogAddressStr,
			relayingHeaderAction.Meta.Header,
			relayingHeaderAction.Meta.BlockHeight,
			relayingHeaderAction.Meta.Type,
			relayingHeaderAction.ShardID,
			relayingHeaderAction.TxReqID,
			RelayingHeaderRejectedChainStatus,
		)
		return [][]string{inst}
	}

	var newBlock types.Block
	err = json.Unmarshal(headerBytes, &newBlock)
	if err != nil {
		Logger.log.Errorf("Error - [buildInstructionsForBNBHeaderRelaying]: Cannot unmarshal header.%v\n", err)
		inst := rbnbChain.BuildHeaderRelayingInst(
			relayingHeaderAction.Meta.IncogAddressStr,
			relayingHeaderAction.Meta.Header,
			relayingHeaderAction.Meta.BlockHeight,
			relayingHeaderAction.Meta.Type,
			relayingHeaderAction.ShardID,
			relayingHeaderAction.TxReqID,
			RelayingHeaderRejectedChainStatus,
		)
		return [][]string{inst}
	}

	if newBlock.Header.Height != int64(relayingHeaderAction.Meta.BlockHeight) {
		Logger.log.Errorf("Error - [buildInstructionsForBNBHeaderRelaying]: Block height in metadata is unmatched with block height in new header.")
		inst := rbnbChain.BuildHeaderRelayingInst(
			relayingHeaderAction.Meta.IncogAddressStr,
			relayingHeaderAction.Meta.Header,
			relayingHeaderAction.Meta.BlockHeight,
			relayingHeaderAction.Meta.Type,
			relayingHeaderAction.ShardID,
			relayingHeaderAction.TxReqID,
			RelayingHeaderRejectedChainStatus,
		)
		return [][]string{inst}
	}

	inst := rbnbChain.BuildHeaderRelayingInst(
		relayingHeaderAction.Meta.IncogAddressStr,
		relayingHeaderAction.Meta.Header,
		relayingHeaderAction.Meta.BlockHeight,
		relayingHeaderAction.Meta.Type,
		relayingHeaderAction.ShardID,
		relayingHeaderAction.TxReqID,
		RelayingHeaderConsideringChainStatus,
	)
	return [][]string{inst}
}

func (rbtcChain *RelayingBTCChain) BuildRelayingInst(
	bc metadata.ChainRetriever,
	relayingHeaderAction metadata.RelayingHeaderAction,
	relayingState *RelayingHeaderChainState,
) [][]string {
	Logger.log.Info("[BTC Relaying] - Processing buildRelayingInst...")
	inst := rbtcChain.BuildHeaderRelayingInst(
		relayingHeaderAction.Meta.IncogAddressStr,
		relayingHeaderAction.Meta.Header,
		relayingHeaderAction.Meta.BlockHeight,
		relayingHeaderAction.Meta.Type,
		relayingHeaderAction.ShardID,
		relayingHeaderAction.TxReqID,
		RelayingHeaderConsideringChainStatus,
	)
	return [][]string{inst}
}