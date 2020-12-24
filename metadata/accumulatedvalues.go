package metadata

import (
	"bytes"

	"github.com/incognitochain/incognito-chain/common"
)

type AccumulatedValues struct {
	UniqETHTxsUsed       [][]byte
	DBridgeTokenPair     map[string][]byte
	CBridgeTokens        []*common.Hash
	InitializedPTokenIDs []string
}

func (ac AccumulatedValues) CanProcessTokenPair(
	externalTokenID []byte,
	incTokenID common.Hash,
) (bool, error) {
	incTokenIDStr := incTokenID.String()
	for _, initializedPTokenID := range ac.InitializedPTokenIDs {
		if initializedPTokenID == incTokenIDStr {
			return false, nil
		}
	}

	for _, tokenID := range ac.CBridgeTokens {
		if bytes.Equal(tokenID[:], incTokenID[:]) {
			return false, nil
		}
	}
	bridgeTokenPair := ac.DBridgeTokenPair
	if existedExtTokenID, found := bridgeTokenPair[incTokenIDStr]; found {
		if bytes.Equal(existedExtTokenID, externalTokenID) {
			return true, nil
		}
		return false, nil
	}
	for _, existedExtTokenID := range bridgeTokenPair {
		if !bytes.Equal(existedExtTokenID, externalTokenID) {
			continue
		}
		return false, nil
	}
	return true, nil
}

func (ac AccumulatedValues) CanProcessCIncToken(
	incTokenID common.Hash,
) bool {
	for _, initializedPTokenID := range ac.InitializedPTokenIDs {
		if initializedPTokenID == incTokenID.String() {
			return false
		}
	}

	incTokenIDStr := incTokenID.String()
	_, found := ac.DBridgeTokenPair[incTokenIDStr]
	return !found
}

func (ac AccumulatedValues) CanProcessPTokenInitialization(
	pTokenID common.Hash,
) bool {
	pTokenIDStr := pTokenID.String()
	_, found := ac.DBridgeTokenPair[pTokenIDStr]
	if found {
		return false
	}
	for _, cTokenID := range ac.CBridgeTokens {
		if bytes.Equal(cTokenID[:], pTokenID[:]) {
			return false
		}
	}
	for _, initializedPTokenID := range ac.InitializedPTokenIDs {
		if initializedPTokenID == pTokenIDStr {
			return false
		}
	}
	return true
}
