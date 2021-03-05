package portalv4

import (
	portaltokensv4 "github.com/incognitochain/incognito-chain/portal/portalv4/portaltokens"
	"time"
)

// todo: add more params for portal v4
type PortalParams struct {
	MultiSigAddresses       map[string]string
	MultiSigScriptHexEncode map[string]string
	PortalTokens            map[string]portaltokensv4.PortalTokenProcessor

	// for unshielding
	DefaultFeeUnshields map[string]uint64 // in nano ptokens
	MinUnshieldAmts     map[string]uint64 // in nano ptokens
	BatchNumBlks        uint

	// for replacement
	PortalReplacementAddress   string
	MaxFeeForEachStep          uint
	TimeSpaceForFeeReplacement time.Duration
}
