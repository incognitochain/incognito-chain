// HOLD CONSTANTS

package mlsag

import (
	"math/big"

	"github.com/incognitochain/incognito-chain/privacy"
)

var lEdwardCached *privacy.Scalar

// Get order of Edwards curve: 2^252 + 27742317777372353535851937790883648493
// And store it into global variable lEdwardCached
func getLEdward() *privacy.Scalar {
	if lEdwardCached != nil {
		return lEdwardCached
	}
	orderEd, _ := new(big.Int).SetString("7237005577332262213973186563042994240857116359379907606001950938285454250989", 10)

	orderScalarEd := new(privacy.Scalar).FromBytesS(orderEd.Bytes())
	keyEd := privacy.Reverse(orderScalarEd.GetKey())
	lEd := new(privacy.Scalar).SetKeyUnsafe(&keyEd)

	lEdwardCached = lEd
	return lEd
}
