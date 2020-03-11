package address

import (
	"errors"
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

func GenerateOTAs(noOTAs int, pubViewKey, pubSpendKey *operation.Point) (*operation.Point, []*operation.Point, error) {
	if noOTAs <= 0 {
		return nil, nil, errors.New("invalid number of generating OTAs")
	}
	r := operation.RandomScalar()
	R := new(operation.Point).ScalarMultBase(r)
	OTAs := make([]*operation.Point, noOTAs)
	for i := 0; i < noOTAs; i++ {
		ss := new(operation.Point).ScalarMult(pubViewKey, r)
		randSS := operation.HashToScalar(append(ss.ToBytesS(), byte(i)))
		OTAs[i] = new(operation.Point).Add(pubSpendKey, new(operation.Point).ScalarMultBase(randSS))
	}
	return R, OTAs, nil
}

func CompareOTA(privViewKey *operation.Scalar, R, pubSpendKey, OTA *operation.Point) (bool, int) {
	maxOutputCoin := 16
	ss := new(operation.Point).ScalarMult(R, privViewKey)
	for i := 0; i < maxOutputCoin; i++ {
		randSS := operation.HashToScalar(append(ss.ToBytesS(), byte(i)))
		tmpOTA := new(operation.Point).Add(pubSpendKey, new(operation.Point).ScalarMultBase(randSS))
		if operation.IsPointEqual(tmpOTA, OTA) {
			return true, i
		}
	}
	return false, 0
}

func GetPrivateTxKey(privSpendKey, privViewKey *operation.Scalar, R *operation.Point, index int) (*operation.Scalar, error) {
	if privViewKey.ScalarValid() != true || privSpendKey.ScalarValid() != true {
		return nil, errors.New("invalid Inputs")
	}
	ss := new(operation.Point).ScalarMult(R, privViewKey)
	privTxKey := new(operation.Scalar).Add(operation.HashToScalar(append(ss.ToBytesS(), byte(index))), privSpendKey)
	return privTxKey, nil
}