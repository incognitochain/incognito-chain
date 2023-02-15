package bridgehub

import (
	"encoding/base64"
	"encoding/json"
	"errors"

	"github.com/incognitochain/incognito-chain/common"
	metadataBriHub "github.com/incognitochain/incognito-chain/metadata/bridgehub"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
)

func buildBriHubRegisterBridgeInst(
	meta metadataBriHub.RegisterBridgeRequest,
	shardID byte,
	txReqID common.Hash,
	bridgeID string,
	status string,
	errorType int,
) ([]string, error) {
	content := metadataBriHub.RegisterBridgeContentInst{
		ExtChainID:       meta.ExtChainID,
		BridgePoolPubKey: meta.BridgePoolPubKey,
		ValidatorPubKeys: meta.ValidatorPubKeys,
		VaultAddress:     meta.VaultAddress,
		Signature:        meta.Signature,
		BridgeID:         bridgeID,
	}
	contentBytes, _ := json.Marshal(content)

	contentStr := ""
	if status == common.AcceptedStatusStr {
		contentStr = base64.StdEncoding.EncodeToString(contentBytes)
	} else if status == common.RejectedStatusStr {
		contentStr, _ = metadataCommon.NewRejectContentWithValue(txReqID, ErrCodeMessage[errorType].Code, contentBytes).String()
	} else {
		return nil, errors.New("Invalid instructtion status")
	}

	inst := metadataCommon.NewInstructionWithValue(
		metadataCommon.BriHubRegisterBridgeMeta,
		status,
		shardID,
		contentStr,
	)
	return inst.StringSlice(), nil
}
