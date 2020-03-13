package txfull

import (
	"errors"
	"github.com/incognitochain/incognito-chain/privacy/privacy_util"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/privacy/operation"
	"github.com/incognitochain/incognito-chain/privacy/privacy_v2/mlsag"

	"github.com/incognitochain/incognito-chain/privacy/address"
	"github.com/incognitochain/incognito-chain/privacy/privacy_v2/onetime_address/utxo"
)

type RingCTFull struct {
	inputs         []*utxo.Utxo
	privateAddress *address.PrivateAddress
	sumBlindOutput *operation.Scalar
	outputs        []*utxo.Utxo
	toAddress      []*address.PublicAddress
}

func NewRingCTFull(inputs []*utxo.Utxo, privateAddress *address.PrivateAddress, sumBlindOutput *operation.Scalar, outputs *[]utxo.Utxo, toAddress *[]address.PublicAddress) *RingCTFull {
	return &RingCTFull{
		inputs,
		privateAddress,
		sumBlindOutput,
		outputs,
		toAddress,
	}
}

// Create and return: ring, privatekey (of transaction), pi of the ring, error
func (this *RingCTFull) CreateRandomRing() (*mlsag.Ring, *[]operation.Scalar, int, error) {
	// TODO
	// In real system should change numFake
	numFake :=  privacy_util.RingSize
	pi := common.RandInt() % numFake

	// Generating Ring without commitment then add later
	privKeys := *getTxPrivateKeys(this.privateAddress, &this.inputs)
	ring := mlsag.NewRandomRing(&privKeys, numFake, pi)

	// Generate privateKey with commitment
	sumBlindInput, err := getSumBlindInput(this)
	if err != nil {
		return nil, nil, 0, errors.New("Error in RingCTFull CreateRandomRing: the RingCTFull is broken (private key not associate with transaction")
	}
	privCommitment := new(operation.Scalar).Sub(sumBlindInput, this.sumBlindOutput)
	priv = append(priv, *privCommitment)

	// Add commitment to ring
	sumInputsCom := getSumCommitment(this.inputs)
	sumOutputCom := getSumCommitment(this.outputs)
	for i := 0; i < numFake; i += 1 {
		val := new(operation.Point)
		if i == pi {
			val = val.Sub(sumInputsCom, sumOutputCom)
		} else {
			// TODO
			// Random scalar should be changed when use in real product
			// Should get randomly in database the commitments
			val = val.Sub(operation.RandomPoint(), sumOutputCom)
		}
		ring.AppendToRow(i, val)
	}

	return ring, &priv, pi, nil
}
