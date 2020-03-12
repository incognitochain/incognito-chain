package address

import (
	"github.com/incognitochain/incognito-chain/privacy/operation"
)

type ViewAddress struct {
	privateView *operation.Scalar
	publicSpend  *operation.Point
}


func (this *ViewAddress) GetPrivateView() *operation.Scalar {
	return this.privateView
}

func (this *ViewAddress) GetPublicSpend() *operation.Point {
	return this.publicSpend
}

func (this *ViewAddress) GetPublicView() *operation.Point {
	return GetPublic(this.privateView)
}