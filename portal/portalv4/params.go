package portalv4

import (
	"errors"
	"time"

	"github.com/incognitochain/incognito-chain/common"
	portalcommonv4 "github.com/incognitochain/incognito-chain/portal/portalv4/common"
	portaltokensv4 "github.com/incognitochain/incognito-chain/portal/portalv4/portaltokens"
)

type PortalParams struct {
	MasterPubKeys                  map[string][][]byte
	NumRequiredSigs                uint
	GeneralMultiSigAddresses       map[string]string // used to received change output coins
	GeneralMultiSigScriptHexEncode map[string]string
	PortalTokens                   map[string]portaltokensv4.PortalTokenProcessor

	// for unshielding
	DefaultFeeUnshields        map[string]uint64 // in nano ptokens
	MinUnshieldAmts            map[string]uint64 // in nano ptokens
	BatchNumBlks               uint
	MinConfirmationIncBlockNum uint
	DustValueThreshold         map[string]uint64 // in nano ptokens

	// for replacement
	PortalReplacementAddress    string
	MaxFeePercentageForEachStep uint
	TimeSpaceForFeeReplacement  time.Duration
}

func (p PortalParams) IsPortalToken(tokenIDStr string) bool {
	isExisted, _ := common.SliceExists(portalcommonv4.PortalV4SupportedIncTokenIDs, tokenIDStr)
	return isExisted
}
func (p PortalParams) GetMinAmountPortalToken(tokenIDStr string) (uint64, error) {
	portalToken, ok := p.PortalTokens[tokenIDStr]
	if !ok {
		return 0, errors.New("TokenID is invalid")
	}
	return portalToken.GetMinTokenAmount(), nil
}
