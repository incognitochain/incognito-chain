// HOLD CONSTANTS

package mlsag

import (
	"github.com/incognitochain/incognito-chain/privacy"
	C25519 "github.com/incognitochain/incognito-chain/privacy/curve25519"
)

var CurveOrder *privacy.Scalar = new(privacy.Scalar).SetKeyUnsafe(&C25519.L)
