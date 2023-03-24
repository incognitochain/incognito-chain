package bridgehub

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/incognitochain/incognito-chain/common"
	metadataBridgeHub "github.com/incognitochain/incognito-chain/metadata/bridgehub"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	"github.com/incognitochain/incognito-chain/privacy"
)

type RegisterBridgeStatus struct {
	Status           byte           `json:"Status"`
	BridgeID         string         `json:"BridgeID"`
	BridgePoolPubKey string         `json:"BridgePoolPubKey"` // TSS pubkey
	ValidatorPubKeys []string       `json:"ValidatorPubKeys"` // pubkey to build TSS key
	VaultAddress     map[int]string `json:"VaultAddress"`     // vault to receive external assets
	Signature        string         `json:"Signature"`        // TSS sig
	ErrorCode        int            `json:"ErrorCode,omitempty"`
}

func buildBridgeHubRegisterBridgeInst(
	meta metadataBridgeHub.RegisterBridgeRequest,
	shardID byte,
	txReqID common.Hash,
	status string,
	errorType int,
) ([]string, error) {
	content := metadataBridgeHub.RegisterBridgeContentInst{
		ValidatorPubKeys: meta.ValidatorPubKeys,
		VaultAddress:     meta.VaultAddress,
		TxReqID:          txReqID.String(),
		BridgePoolPubKey: meta.BridgePoolPubKey,
		Signature:        meta.Signature,
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

type StakeBridgeStatus struct {
	Status           byte                   `json:"Status"`
	StakeAmount      uint64                 `json:"StakeAmount"` // must be equal to vout value
	TokenID          common.Hash            `json:"TokenID"`
	StakerAddress    privacy.PaymentAddress `json:"StakerAddress"`
	BridgePubKey     string                 `json:"BridgePubKey"` // staker's key
	BridgePoolPubKey string                 `json:"BridgePoolPubKey"`
	BridgePubKeys    []string               `json:"BridgePubKeys"`
	ErrorCode        int                    `json:"ErrorCode,omitempty"`
}

func buildBridgeHubStakeInst(
	meta metadataBridgeHub.StakePRVRequest,
	shardID byte,
	txReqID common.Hash,
	status string,
	errorType int,
) ([]string, error) {
	content := metadataBridgeHub.StakePRVRequestContentInst{
		StakeAmount:      meta.StakeAmount,
		TokenID:          meta.TokenID,
		BridgePubKey:     meta.BridgePubKey,
		BridgePoolPubKey: meta.BridgePoolPubKey,
		TxReqID:          txReqID.String(),
		Staker:           meta.Staker,
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
		metadataCommon.StakePRVRequestMeta,
		status,
		shardID,
		contentStr,
	)
	return inst.StringSlice(), nil
}

func newBridgeHubNetworkInfo(vaultAddresses map[int]string) map[int]*BridgeNetwork {
	temp := make(map[int]*BridgeNetwork)
	for v, k := range vaultAddresses {
		temp[v] = &BridgeNetwork{
			VaultAddress: k,
			NetworkId:    v,
		}
	}

	return temp
}
