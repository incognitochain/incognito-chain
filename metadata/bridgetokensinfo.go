package metadata

import "github.com/incognitochain/incognito-chain/common"

type UpdatingInfo struct {
	CountUpAmt      uint64
	DeductAmt       uint64
	TokenID         common.Hash
	ExternalTokenID []byte
	IsCentralized   bool
}

func UpdatePortalBridgeTokenInfo(tokenInfos map[common.Hash]UpdatingInfo, tokenID string, amount uint64, isDeduct bool) error {
	incTokenID, err := common.Hash{}.NewHashFromStr(tokenID)
	if err != nil {
		Logger.log.Errorf("[UpdatePortalBridgeTokenInfo]: Can not new hash from incTokenID: %+v", err)
		return nil
	}
	updatingInfo, found := tokenInfos[*incTokenID]
	if !found {
		updatingInfo = UpdatingInfo{
			CountUpAmt:      0,
			DeductAmt:       0,
			TokenID:         *incTokenID,
			ExternalTokenID: nil,
			IsCentralized:   true,
		}
	}

	if isDeduct {
		updatingInfo.DeductAmt += amount
	} else {
		updatingInfo.CountUpAmt += amount
	}

	tokenInfos[*incTokenID] = updatingInfo
	return nil
}
