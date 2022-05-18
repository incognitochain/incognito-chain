package common

import (
	"bytes"

	"github.com/incognitochain/incognito-chain/common"
)

type AccumulatedValues struct {
	UniqETHTxsUsed    [][]byte
	UniqBSCTxsUsed    [][]byte
	UniqPRVEVMTxsUsed [][]byte
	UniqPLGTxsUsed    [][]byte
	UniqFTMTxsUsed    [][]byte
	DBridgeTokenPair  map[string][]byte
	CBridgeTokens     []*common.Hash
	InitTokens        []*common.Hash
}

func (ac AccumulatedValues) CanProcessTokenPair(
	externalTokenID []byte,
	incTokenID common.Hash,
) (bool, error) {
	incTokenIDStr := incTokenID.String()
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
	incTokenIDStr := incTokenID.String()
	_, found := ac.DBridgeTokenPair[incTokenIDStr]
	return !found
}

func (ac AccumulatedValues) CanProcessTokenInit(
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
	for _, initializedPTokenID := range ac.InitTokens {
		if initializedPTokenID.String() == pTokenIDStr {
			return false
		}
	}
	return true
}
