package blockchain

import (
	"encoding/base64"
	"encoding/json"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/database/lvdb"
	"github.com/incognitochain/incognito-chain/metadata"
	"strconv"
)

// beacon build new instruction from instruction received from ShardToBeaconBlock
func buildBNBHeaderRelayingInst(
	senderAddressStr string,
	header string,
	blockHeight uint64,
	metaType int,
	shardID byte,
	txReqID common.Hash,
	status string,
) []string {
	bnbHeaderRelayingContent := metadata.RelayingBNBHeaderContent{
		IncogAddressStr: senderAddressStr,
		Header:          header,
		TxReqID:         txReqID,
		BlockHeight:     blockHeight,
	}
	bnbHeaderRelayingContentBytes, _ := json.Marshal(bnbHeaderRelayingContent)
	return []string{
		strconv.Itoa(metaType),
		strconv.Itoa(int(shardID)),
		status,
		string(bnbHeaderRelayingContentBytes),
	}
}

// buildInstructionsForBNBHeaderRelaying builds instruction for custodian deposit action
func (blockchain *BlockChain) buildInstructionsForBNBHeaderRelaying(
	contentStr string,
	shardID byte,
	metaType int,
	relayingHeaderChain *RelayingHeaderChainState,
	beaconHeight uint64,
) ([][]string, error) {
	Logger.log.Infof("[RELAYING] beacon relaying producer starting....")
	// parse instruction
	actionContentBytes, err := base64.StdEncoding.DecodeString(contentStr)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while decoding content string of portal custodian deposit action: %+v", err)
		return [][]string{}, nil
	}
	var actionData metadata.RelayingBNBHeaderAction
	err = json.Unmarshal(actionContentBytes, &actionData)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while unmarshal portal custodian deposit action: %+v", err)
		return [][]string{}, nil
	}

	if relayingHeaderChain == nil {
		Logger.log.Warn("WARN - [buildInstructionsForBNBHeaderRelaying]: relayingHeaderChain is null.")
		inst := buildBNBHeaderRelayingInst(
			actionData.Meta.IncogAddressStr,
			actionData.Meta.Header,
			actionData.Meta.BlockHeight,
			actionData.Meta.Type,
			shardID,
			actionData.TxReqID,
			common.RelayingHeaderRejectedChainStatus,
		)
		return [][]string{inst}, nil
	}
	meta := actionData.Meta
	// parse and verify header chain
	headerBytes, err := base64.StdEncoding.DecodeString(meta.Header)
	if err != nil {
		Logger.log.Errorf("Error - [buildInstructionsForBNBHeaderRelaying]: Can not decode header string.%v\n", err)
		inst := buildBNBHeaderRelayingInst(
			actionData.Meta.IncogAddressStr,
			actionData.Meta.Header,
			actionData.Meta.BlockHeight,
			actionData.Meta.Type,
			shardID,
			actionData.TxReqID,
			common.RelayingHeaderRejectedChainStatus,
		)
		return [][]string{inst}, nil
	}

	var newHeader lvdb.BNBHeader
	err = json.Unmarshal(headerBytes, &newHeader)
	if err != nil {
		Logger.log.Errorf("Error - [buildInstructionsForBNBHeaderRelaying]: Can not unmarshal header.%v\n", err)
		inst := buildBNBHeaderRelayingInst(
			actionData.Meta.IncogAddressStr,
			actionData.Meta.Header,
			actionData.Meta.BlockHeight,
			actionData.Meta.Type,
			shardID,
			actionData.TxReqID,
			common.RelayingHeaderRejectedChainStatus,
		)
		return [][]string{inst}, nil
	}

	if newHeader.Header.Height != int64(actionData.Meta.BlockHeight) {
		Logger.log.Errorf("Error - [buildInstructionsForBNBHeaderRelaying]: Block height in metadata is unmatched with block height in new header.")
		inst := buildBNBHeaderRelayingInst(
			actionData.Meta.IncogAddressStr,
			actionData.Meta.Header,
			actionData.Meta.BlockHeight,
			actionData.Meta.Type,
			shardID,
			actionData.TxReqID,
			common.RelayingHeaderRejectedChainStatus,
		)
		return [][]string{inst}, nil
	}

	// if valid, create instruction with status accepted
	// if not, create instruction with status rejected
	latestBNBHeader := relayingHeaderChain.BNBHeaderChain.LatestHeader
	var isValid bool
	var err2 error
	relayingHeaderChain.BNBHeaderChain, isValid, err2 = relayingHeaderChain.BNBHeaderChain.ReceiveNewHeader(
		newHeader.Header, newHeader.LastCommit, blockchain.config.ChainParams.BNBRelayingHeaderChainID)
	if err2 != nil || !isValid {
		Logger.log.Errorf("Error - [buildInstructionsForBNBHeaderRelaying]: Verify new header failed. %v\n", err2)
		inst := buildBNBHeaderRelayingInst(
			actionData.Meta.IncogAddressStr,
			actionData.Meta.Header,
			actionData.Meta.BlockHeight,
			actionData.Meta.Type,
			shardID,
			actionData.TxReqID,
			common.RelayingHeaderRejectedChainStatus,
		)
		return [][]string{inst}, nil
	}

	// check newHeader is a header contain last commit for one of the header in unconfirmed header list or not
	newLatestBNBHeader := relayingHeaderChain.BNBHeaderChain.LatestHeader
	if newLatestBNBHeader != nil && newLatestBNBHeader.Height == 1 && latestBNBHeader == nil {
		inst := buildBNBHeaderRelayingInst(
			actionData.Meta.IncogAddressStr,
			actionData.Meta.Header,
			actionData.Meta.BlockHeight,
			actionData.Meta.Type,
			shardID,
			actionData.TxReqID,
			common.RelayingHeaderConfirmedAcceptedChainStatus,
		)
		return [][]string{inst}, nil
	}

	if newLatestBNBHeader != nil && latestBNBHeader != nil {
		if newLatestBNBHeader.Height == latestBNBHeader.Height + 1 {
			inst := buildBNBHeaderRelayingInst(
				actionData.Meta.IncogAddressStr,
				actionData.Meta.Header,
				actionData.Meta.BlockHeight,
				actionData.Meta.Type,
				shardID,
				actionData.TxReqID,
				common.RelayingHeaderConfirmedAcceptedChainStatus,
			)
			return [][]string{inst}, nil
		}
	}

	inst := buildBNBHeaderRelayingInst(
		actionData.Meta.IncogAddressStr,
		actionData.Meta.Header,
		actionData.Meta.BlockHeight,
		actionData.Meta.Type,
		shardID,
		actionData.TxReqID,
		common.RelayingHeaderUnconfirmedAcceptedChainStatus,
	)
	return [][]string{inst}, nil
}
