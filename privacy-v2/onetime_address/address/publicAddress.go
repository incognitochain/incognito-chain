package address

import (
	"github.com/incognitochain/incognito-chain/privacy"
)

type PublicAddress struct {
	publicSpend *privacy.Point
	publicView  *privacy.Point
}

// Get Public Key from Private Key
func GetPublicSpend(privateKey *privacy.Scalar) *privacy.Point {
	return new(privacy.Point).ScalarMultBase(privateKey)
}

func (this *PublicAddress) GetPublicSpend() *privacy.Point {
	return this.publicSpend
}

func (this *PublicAddress) GetPublicView() *privacy.Point {
	return this.publicView
}
