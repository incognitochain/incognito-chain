package address

import (
	"github.com/incognitochain/incognito-chain/privacy"
)

type PublicAddress struct {
	publicSpend *privacy.Point
	publicView  *privacy.Point
}

func (this *PublicAddress) GetPublicSpend() *privacy.Point {
	return this.publicSpend
}

func (this *PublicAddress) GetPublicView() *privacy.Point {
	return this.publicView
}
