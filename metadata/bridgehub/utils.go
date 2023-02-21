package bridgehub

import (
	"encoding/base64"
	"encoding/json"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
)

func IsBridgeHubMetaType(metaType int) bool {
	switch metaType {
	case metadataCommon.BridgeHubRegisterBridgeMeta:
		return true
	// TODO 0xkraken: add more metadata
	default:
		return false
	}
}

func DecodeContent(content string, action interface{}) error {
	contentBytes, err := base64.StdEncoding.DecodeString(content)
	if err != nil {
		return err
	}
	return json.Unmarshal(contentBytes, &action)
}
