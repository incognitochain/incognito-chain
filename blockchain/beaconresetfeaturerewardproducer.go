package blockchain

import (
	"encoding/base64"
	"encoding/json"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/metadata"
	"strconv"
)

// beacon build new instruction from instruction received from ShardToBeaconBlock
func buildResetFeatureRewardInst(
	tokenID common.Hash,
	metaType int,
	shardID byte,
	txReqID common.Hash,
) []string {
	resetFeatureRewardContent := metadata.ResetFeatureRewardRequestContent{
		TokenID: tokenID,
		TxReqID: txReqID,
		ShardID: shardID,
	}
	resetFeatureRewardContentBytes, _ := json.Marshal(resetFeatureRewardContent)
	return []string{
		strconv.Itoa(metaType),
		strconv.Itoa(int(shardID)),
		"resetFeatureReward",
		string(resetFeatureRewardContentBytes),
	}
}

func (blockchain *BlockChain) buildInstructionsForResetFeatureReward(
	contentStr string,
	shardID byte,
	metaType int,
) ([][]string, error) {
	// parse instruction
	actionContentBytes, err := base64.StdEncoding.DecodeString(contentStr)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while decoding content string of reset feature reward action: %+v", err)
		return [][]string{}, nil
	}
	var actionData metadata.WithDrawRewardResponseAction
	err = json.Unmarshal(actionContentBytes, &actionData)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while unmarshal reset feature reward action: %+v", err)
		return [][]string{}, nil
	}

	insts := buildResetFeatureRewardInst(actionData.Meta.TokenID, metaType,  actionData.ShardID, actionData.TxReqID)
	return [][]string{insts}, nil
}
