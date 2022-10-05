package common

import (
	"bytes"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/pkg/errors"
)

type AccumulatedValues struct {
	UniqETHTxsUsed    [][]byte
	UniqBSCTxsUsed    [][]byte
	UniqPRVEVMTxsUsed [][]byte
	UniqPLGTxsUsed    [][]byte
	UniqFTMTxsUsed    [][]byte
	UniqAURORATxsUsed [][]byte
	UniqAVAXTxsUsed   [][]byte
	DBridgeTokenPair  map[string][]byte
	CBridgeTokens     []*common.Hash
	InitTokens        []*common.Hash
}

func (ac *AccumulatedValues) UpdateUniqTxsUsed(networkID uint8, uniqTxsUsed [][]byte) (*AccumulatedValues, error) {
	switch networkID {
	case common.ETHNetworkID:
		ac.UniqETHTxsUsed = uniqTxsUsed
	case common.BSCNetworkID:
		ac.UniqBSCTxsUsed = uniqTxsUsed
	case common.PLGNetworkID:
		ac.UniqPLGTxsUsed = uniqTxsUsed
	case common.FTMNetworkID:
		ac.UniqFTMTxsUsed = uniqTxsUsed
	case common.AURORANetworkID:
		ac.UniqAURORATxsUsed = uniqTxsUsed
	case common.AVAXNetworkID:
		ac.UniqAVAXTxsUsed = uniqTxsUsed
	default:
		return nil, errors.New("Invalid networkID")
	}
	return ac, nil
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

func (ac *AccumulatedValues) Clone() *AccumulatedValues {
	res := &AccumulatedValues{}
	res.UniqETHTxsUsed = make([][]byte, len(ac.UniqETHTxsUsed))
	for i, v := range ac.UniqETHTxsUsed {
		res.UniqETHTxsUsed[i] = make([]byte, len(v))
		copy(res.UniqETHTxsUsed[i], v)
	}
	res.UniqBSCTxsUsed = make([][]byte, len(ac.UniqBSCTxsUsed))
	for i, v := range ac.UniqBSCTxsUsed {
		res.UniqBSCTxsUsed[i] = make([]byte, len(v))
		copy(res.UniqBSCTxsUsed[i], v)
	}
	res.UniqPRVEVMTxsUsed = make([][]byte, len(ac.UniqPRVEVMTxsUsed))
	for i, v := range ac.UniqPRVEVMTxsUsed {
		res.UniqPRVEVMTxsUsed[i] = make([]byte, len(v))
		copy(res.UniqPRVEVMTxsUsed[i], v)
	}
	res.UniqPLGTxsUsed = make([][]byte, len(ac.UniqPLGTxsUsed))
	for i, v := range ac.UniqPLGTxsUsed {
		res.UniqPLGTxsUsed[i] = make([]byte, len(v))
		copy(res.UniqPLGTxsUsed[i], v)
	}
	res.UniqFTMTxsUsed = make([][]byte, len(ac.UniqFTMTxsUsed))
	for i, v := range ac.UniqFTMTxsUsed {
		res.UniqFTMTxsUsed[i] = make([]byte, len(v))
		copy(res.UniqFTMTxsUsed[i], v)
	}
	res.CBridgeTokens = make([]*common.Hash, len(ac.CBridgeTokens))
	for _, v := range ac.CBridgeTokens {
		tokenID := &common.Hash{}
		copy(tokenID[:], v[:])
		res.CBridgeTokens = append(res.CBridgeTokens, tokenID)
	}
	res.InitTokens = make([]*common.Hash, len(ac.InitTokens))
	for _, v := range ac.InitTokens {
		tokenID := &common.Hash{}
		copy(tokenID[:], v[:])
		res.InitTokens = append(res.InitTokens, tokenID)
	}
	res.DBridgeTokenPair = make(map[string][]byte)
	for k, v := range ac.DBridgeTokenPair {
		res.DBridgeTokenPair[k] = make([]byte, len(v))
		copy(res.DBridgeTokenPair[k], v)
	}

	return res
}
