package privacy_v2

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/privacy/operation"
)

type TokenAttributes struct {
	Private  bool
	BurnOnly bool
}

func MapPlainAssetTags(m map[common.Hash]TokenAttributes) map[string]*common.Hash {
	result := make(map[string]*common.Hash)
	for id, _ := range m {
		assetTag := operation.HashToPoint(id[:])
		var tokenID common.Hash = id
		result[assetTag.String()] = &tokenID
	}
	return result
}
