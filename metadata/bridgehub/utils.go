package bridgehub

import (
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
)

func IsBridgeHubMetaType(metaType int) bool {
	switch metaType {
	case metadataCommon.BridgeHubRegisterBridgeMeta:
		return true
	case metadataCommon.StakePRVRequestMeta:
		return true
	// TODO 0xkraken: add more metadata
	default:
		return false
	}
}
