package mlsag

import "github.com/incognitochain/incognito-chain/privacy"

type Signature struct {
	c         *privacy.Scalar
	r         [][]privacy.Scalar
	keyImages []privacy.Point
}
