package bridgehub

import (
	"encoding/base64"
	"encoding/json"
	"errors"

	"github.com/incognitochain/incognito-chain/common"
	metadataBridgeHub "github.com/incognitochain/incognito-chain/metadata/bridgehub"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
)

type RegisterBridgeStatus struct {
	Status           byte     `json:"Status"`
	BridgeID         string   `json:"BridgeID"`
	ExtChainID       string   `json:"ExtChainID"`
	BridgePoolPubKey string   `json:"BridgePoolPubKey"` // TSS pubkey
	ValidatorPubKeys []string `json:"ValidatorPubKeys"` // pubkey to build TSS key
	VaultAddress     string   `json:"VaultAddress"`     // vault to receive external assets
	Signature        string   `json:"Signature"`        // TSS sig
	ErrorCode        int      `json:"ErrorCode,omitempty"`
}

func buildBridgeHubRegisterBridgeInst(
	meta metadataBridgeHub.RegisterBridgeRequest,
	shardID byte,
	txReqID common.Hash,
	bridgeID string,
	status string,
	errorType int,
) ([]string, error) {
	content := metadataBridgeHub.RegisterBridgeContentInst{
		ExtChainID:       meta.ExtChainID,
		BridgePoolPubKey: meta.BridgePoolPubKey,
		ValidatorPubKeys: meta.ValidatorPubKeys,
		VaultAddress:     meta.VaultAddress,
		Signature:        meta.Signature,
		BridgeID:         bridgeID,
		TxReqID:          txReqID.String(),
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
		metadataCommon.BridgeHubRegisterBridgeMeta,
		status,
		shardID,
		contentStr,
	)
	return inst.StringSlice(), nil
}
