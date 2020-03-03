// HOLD CONSTANTS

package mlsag

import (
	"github.com/incognitochain/incognito-chain/privacy/operation"
	C25519 "github.com/incognitochain/incognito-chain/privacy/operation/curve25519"
)

var CurveOrder *operation.Scalar = new(operation.Scalar).SetKeyUnsafe(&C25519.L)
