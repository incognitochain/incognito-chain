package portal

import (
	"github.com/incognitochain/incognito-chain/portal/portalrelaying"
	"github.com/incognitochain/incognito-chain/portal/portalv3"
	"github.com/incognitochain/incognito-chain/portal/portalv4"
	"sort"
)

type PortalParams struct {
	RelayingParam  portalrelaying.RelayingParams
	PortalParamsV3 map[uint64]portalv3.PortalParams
	PortalParamsV4 map[uint64]portalv4.PortalParams
}

func (p PortalParams) GetPortalParamsV3(beaconHeight uint64) portalv3.PortalParams {
	portalParamMap := p.PortalParamsV3
	// only has one value - default value
	if len(portalParamMap) == 1 {
		return portalParamMap[0]
	}

	bchs := []uint64{}
	for bch := range portalParamMap {
		bchs = append(bchs, bch)
	}
	sort.Slice(bchs, func(i, j int) bool {
		return bchs[i] < bchs[j]
	})

	bchKey := bchs[len(bchs)-1]
	for i := len(bchs) - 1; i >= 0; i-- {
		if beaconHeight < bchs[i] {
			continue
		}
		bchKey = bchs[i]
		break
	}

	return portalParamMap[bchKey]
}

func (p PortalParams) GetPortalParamsV4(beaconHeight uint64) portalv4.PortalParams {
	portalParamMap := p.PortalParamsV4
	// only has one value - default value
	if len(portalParamMap) == 1 {
		return portalParamMap[0]
	}

	bchs := []uint64{}
	for bch := range portalParamMap {
		bchs = append(bchs, bch)
	}
	sort.Slice(bchs, func(i, j int) bool {
		return bchs[i] < bchs[j]
	})

	bchKey := bchs[len(bchs)-1]
	for i := len(bchs) - 1; i >= 0; i-- {
		if beaconHeight < bchs[i] {
			continue
		}
		bchKey = bchs[i]
		break
	}

	return portalParamMap[bchKey]
}
