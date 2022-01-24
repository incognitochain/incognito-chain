package privacy_v2

import (
	"fmt"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/privacy/coin"
	"github.com/incognitochain/incognito-chain/privacy/operation"
	"github.com/incognitochain/incognito-chain/privacy/proof"
)

var (
	tokenPrivAttributeMap = initTokenPrivacyAttributes()
)

type TokenAttributes struct {
	Private  bool
	BurnOnly bool
}

// ValidateNonPrivateTransfer verifies TXs with identifiable inputs, based on that input's token attributes
func ValidateNonPrivateTransfer(tokenID common.Hash, proof proof.Proof) (bool, error) {
	att, found := tokenPrivAttributeMap[tokenID]
	if found && att.BurnOnly {
		// a burn-only token can only be burned
		for _, outcoin := range proof.GetOutputCoins() {
			receiverPk := outcoin.GetPublicKey().ToBytesS()
			if !common.IsPublicKeyBurningAddress(receiverPk) {
				return false, fmt.Errorf("cannot transfer burn-only token %v", tokenID)
			}
		}
	}
	return true, nil
}

// ContainsNonPrivacyToken detects a non-private coin in a ring signature
func ContainsNonPrivateToken(ringCoins [][]*coin.CoinV2) (bool, bool, *common.Hash, *coin.CoinV2) {
	assetTagMap := MapPlainAssetTags(tokenPrivAttributeMap)
	for _, row := range ringCoins {
		for _, c := range row {
			tokenID, _ := c.GetTokenId(nil, assetTagMap) // error ignored; only relevant when keySet is present
			if tokenID != nil {
				att, found := tokenPrivAttributeMap[*tokenID]
				if found && !att.Private {
					return true, att.BurnOnly, tokenID, c
				}
			}
		}
	}
	return false, false, nil, nil
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

func initTokenPrivacyAttributes() map[common.Hash]TokenAttributes {
	result := make(map[common.Hash]TokenAttributes)
	result[common.PdexAccessCoinID] = TokenAttributes{Private: false, BurnOnly: true}
	return result
}
