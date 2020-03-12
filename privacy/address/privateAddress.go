package address

import (
	"github.com/incognitochain/incognito-chain/privacy/operation"
)

type PrivateAddress struct {
	privateSpend *operation.Scalar
	privateView  *operation.Scalar
}

// Get Public Key from Private Key
func GetPublic(private *operation.Scalar) *operation.Point {
	return new(operation.Point).ScalarMultBase(private)
}

func (this *PrivateAddress) GetPrivateSpend() *operation.Scalar {
	return this.privateSpend
}

func (this *PrivateAddress) GetPrivateView() *operation.Scalar {
	return this.privateView
}

func (this *PrivateAddress) GetPublicSpend() *operation.Point {
	return GetPublic(this.privateSpend)
}

func (this *PrivateAddress) GetPublicView() *operation.Point {
	return GetPublic(this.privateView)
}

// Get public address coresponding to this private address
func (this *PrivateAddress) GetPublicAddress() *PublicAddress {
	return &PublicAddress{
		this.GetPublicSpend(),
		this.GetPublicView(),
	}
}

func GenerateRandomAddress() *PrivateAddress {
	return &PrivateAddress{
		operation.RandomScalar(),
		operation.RandomScalar(),
	}
}
