package address

import (
	"github.com/incognitochain/incognito-chain/privacy"
)

type PrivateAddress struct {
	privateSpend *privacy.Scalar
	privateView  *privacy.Scalar
}

// Get Public Key from Private Key
func GetPublicKey(privateKey *privacy.Scalar) *privacy.Point {
	return new(privacy.Point).ScalarMultBase(privateKey)
}

func (this *PrivateAddress) GetPrivateSpend() *privacy.Scalar {
	return this.privateSpend
}

func (this *PrivateAddress) GetPrivateView() *privacy.Scalar {
	return this.privateView
}

func (this *PrivateAddress) GetPublicSpend() *privacy.Point {
	return GetPublicKey(this.privateSpend)
}

func (this *PrivateAddress) GetPublicView() *privacy.Point {
	return GetPublicKey(this.privateView)
}

// Get public address coresponding to this private address
func (this *PrivateAddress) GetPublicAddress() *PublicAddress {
	result := new(PublicAddress)
	result.publicSpend = this.GetPublicSpend()
	result.publicView = this.GetPublicView()
	return result
}

func GenerateRandomAddress() *PrivateAddress {
	result := new(PrivateAddress)
	result.privateSpend = privacy.RandomScalar()
	result.privateView = privacy.RandomScalar()
	return result
}
