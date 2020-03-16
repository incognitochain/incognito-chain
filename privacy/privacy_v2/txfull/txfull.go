package txfull

import (
	"errors"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/privacy/coin"
	"github.com/incognitochain/incognito-chain/privacy/address"
	"github.com/incognitochain/incognito-chain/privacy/operation"
	"github.com/incognitochain/incognito-chain/privacy/privacy_util"
	"github.com/incognitochain/incognito-chain/privacy/privacy_v2/mlsag"

)

type RingCTFull struct {
	inputs         []*coin.Coin_v2
	privateAddress *address.PrivateAddress
	sumBlindOutput *operation.Scalar
	outputs        []*coin.Coin_v2
	toAddress      []*address.PublicAddress
}

type Member struct {
	publicKey *operation.Point
	commitment *operation.Point
}

func NewMember(publicKey *operation.Point, commitment *operation.Point) *Member {
	return &Member{
		publicKey,
		commitment,
	}
}

func NewRingCTFull(inputs []*coin.Coin_v2, privateAddress *address.PrivateAddress, sumBlindOutput *operation.Scalar,
	outputs []*coin.Coin_v2, toAddress []*address.PublicAddress) *RingCTFull {
	return &RingCTFull{
		inputs,
		privateAddress,
		sumBlindOutput,
		outputs,
		toAddress,
	}
}

func (this *RingCTFull) CreateRingCTFull(members [][]*Member) (*mlsag.Ring, []*operation.Scalar, int, error) {
	m := len(this.inputs)
	if m != privacy_util.RingSize {
		return nil, nil, 0, errors.New("Invalid input members for MLSAG")
	}
	pi := common.RandInt() % privacy_util.RingSize

	decoyPubKeys := make([][]*operation.Point, privacy_util.RingSize)
	txPrivateKeys := make([]*operation.Scalar, m + 1)
	sumBlindInput := new(operation.Scalar).FromUint64(0)

	for i := 0; i < privacy_util.RingSize; i ++ {
		decoyPubKeys[i] = make([]*operation.Point, m + 1)
		sumInCommitment := new(operation.Point).Identity()
		for j := 0; j < m; j ++ {
			if i != pi {
				if blind, err := getBlindInput(this.privateAddress, this.inputs[j]); err != nil {
					return nil, nil, 0, errors.New("Cannot get blind from Input")
				} else {
					sumBlindInput.Add(sumBlindInput, blind)
				}
				sumInCommitment.Add(sumInCommitment, members[i][j].commitment)
				decoyPubKeys[i][j] = members[i][j].publicKey
			} else {
				txPrivateKeys[j] = getTxPrivateKey(this.privateAddress, this.inputs[j])
				sumInCommitment.Add(sumInCommitment, this.inputs[j].GetCommitment())
				decoyPubKeys[i][j] = parsePublicKey(txPrivateKeys[j])
			}

		}
		sumOutCommitment := getSumCommitment(this.outputs)
		decoyPubKeys[i][m+1] = new(operation.Point).Sub(sumInCommitment, sumOutCommitment)
	}
	lastPrivateKey := new(operation.Scalar).Sub(sumBlindInput, this.sumBlindOutput)
	ringCTFull := mlsag.InitRing(decoyPubKeys)
	return ringCTFull, append(txPrivateKeys, lastPrivateKey), pi, nil
}
