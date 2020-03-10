package address

import (
	"github.com/incognitochain/incognito-chain/privacy/operation"
)

type PrivateAddress struct {
	privateSpend *operation.Scalar
	privateView  *operation.Scalar
}

// Get Public Key from Private Key
func GetPublicKey(privateKey *operation.Scalar) *operation.Point {
	return new(operation.Point).ScalarMultBase(privateKey)
}

func (this *PrivateAddress) GetPrivateSpend() *operation.Scalar {
	return this.privateSpend
}

func (this *PrivateAddress) GetPrivateView() *operation.Scalar {
	return this.privateView
}

func (this *PrivateAddress) GetPublicSpend() *operation.Point {
	return GetPublicKey(this.privateSpend)
}

func (this *PrivateAddress) GetPublicView() *operation.Point {
	return GetPublicKey(this.privateView)
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
