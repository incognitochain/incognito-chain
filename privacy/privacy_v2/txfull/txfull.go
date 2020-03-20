package txfull

import (
	"errors"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/privacy/address"
	"github.com/incognitochain/incognito-chain/privacy/coin"
	"github.com/incognitochain/incognito-chain/privacy/operation"
	"github.com/incognitochain/incognito-chain/privacy/privacy_util"
	"github.com/incognitochain/incognito-chain/privacy/privacy_v2/mlsag"
)

type RingCTFull struct {
	inputs         []*coin.CoinV2
	privateAddress *address.PrivateAddress
	sumBlindOutput *operation.Scalar
	outputs        []*coin.CoinV2
	toAddress      []*address.PublicAddress
}

type Member struct {
	publicKey  *operation.Point
	commitment *operation.Point
}

func NewMember(publicKey *operation.Point, commitment *operation.Point) *Member {
	return &Member{
		publicKey,
		commitment,
	}
}

func NewRingCTFull(inputs []*coin.CoinV2, privateAddress *address.PrivateAddress, sumBlindOutput *operation.Scalar,
	outputs []*coin.CoinV2, toAddress []*address.PublicAddress) *RingCTFull {
	return &RingCTFull{
		inputs,
		privateAddress,
		sumBlindOutput,
		outputs,
		toAddress,
	}
}

func (this *RingCTFull) CreateRingCTFull(decoyMembers [][]*Member) (*mlsag.Mlsag, error) {
	m := len(this.inputs)
	if m != privacy_util.RingSize {
		return nil, errors.New("Invalid input decoyMembers for MLSAG")
	}
	pi := common.RandInt() % privacy_util.RingSize

	ringInput := make([][]*operation.Point, privacy_util.RingSize)
	txPrivateKeys := make([]*operation.Scalar, m+1)
	sumBlindInput := new(operation.Scalar).FromUint64(0)

	for i := 0; i < privacy_util.RingSize; i++ {
		ringInput[i] = make([]*operation.Point, m+1)
		sumInCommitment := new(operation.Point).Identity()
		for j := 0; j < m; j++ {
			if i != pi {
				if blind, err := getBlindInput(this.privateAddress, this.inputs[j]); err != nil {
					return nil, errors.New("Cannot get blind from Input")
				} else {
					sumBlindInput.Add(sumBlindInput, blind)
				}
				sumInCommitment.Add(sumInCommitment, decoyMembers[i][j].commitment)
				ringInput[i][j] = decoyMembers[i][j].publicKey
			} else {
				txPrivateKeys[j] = getTxPrivateKey(this.privateAddress, this.inputs[j])
				sumInCommitment.Add(sumInCommitment, this.inputs[j].GetCommitment())
				ringInput[i][j] = parsePublicKey(txPrivateKeys[j])
			}

		}
		sumOutCommitment := getSumCommitment(this.outputs)
		ringInput[i][m+1] = new(operation.Point).Sub(sumInCommitment, sumOutCommitment)
	}
	lastPrivateKey := new(operation.Scalar).Sub(sumBlindInput, this.sumBlindOutput)
	mlsagObj := mlsag.NewMlsag(append(txPrivateKeys, lastPrivateKey), mlsag.NewRing(ringInput), pi)
	return mlsagObj, nil
}

func RingCTFullProve(mlsagObj mlsag.Mlsag,  message []byte) (*mlsag.MlsagSig, error) {
	mlsagSig, err := mlsagObj.Sign(message)
	if err != nil {
		return nil, errors.New("Cannot create Mlsag Signature")
	}
	return mlsagSig, nil
}

func RingCTFullVerify(mlsgaSig *mlsag.MlsagSig, R *mlsag.Ring, message []byte) bool {
	res, err := mlsag.Verify(mlsgaSig, R, message)
	if err != nil {
		return false
	} else {
		return res
	}
}