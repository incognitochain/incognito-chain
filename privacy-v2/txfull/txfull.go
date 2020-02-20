package txfull

import (
	"errors"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/privacy-v2/mlsag"
	ota "github.com/incognitochain/incognito-chain/privacy-v2/onetime_address"

	"github.com/incognitochain/incognito-chain/privacy-v2/onetime_address/address"
)

type RingCTFull struct {
	inputs         []ota.UTXO
	fromAddress    []address.PrivateAddress
	sumBlindOutput *privacy.Scalar
	outputs        []ota.UTXO
	toAddress      []address.PublicAddress
}

func NewRingCTFull(inputs *[]ota.UTXO, fromAddress *[]address.PrivateAddress, sumBlindOutput *privacy.Scalar, outputs *[]ota.UTXO, toAddress *[]address.PublicAddress) *RingCTFull {
	return &RingCTFull{
		*inputs,
		*fromAddress,
		sumBlindOutput,
		*outputs,
		*toAddress,
	}
}

// Create and return: ring, privatekey (of transaction), pi number of the ring, error
func (this *RingCTFull) CreateRandomRing(message string) (*mlsag.Ring, *[]privacy.Scalar, int, error) {
	// TODO
	// In real system should change numFake
	numFake := 10
	pi := common.RandInt() % numFake

	// Generating Ring without commitment then add later
	priv := *this.getPrivateKeyOfInputs()
	ring := mlsag.NewRandomRing(&priv, numFake, pi)

	// Generate privateKey with commitment
	sumBlindInput, err := getSumBlindInput(this)
	if err != nil {
		return nil, nil, 0, errors.New("Error in RingCTFull CreateRandomRing: the RingCTFull is broken (private key not associate with transaction")
	}
	privCommitment := new(privacy.Scalar).Sub(sumBlindInput, this.sumBlindOutput)
	priv = append(priv, *privCommitment)

	// Add commitment to ring
	sumInputsCom := getSumCommitment(this.inputs)
	sumOutputCom := getSumCommitment(this.outputs)
	for i := 0; i < numFake; i += 1 {
		val := new(privacy.Point)
		if i == pi {
			val = val.Sub(sumInputsCom, sumOutputCom)
		} else {
			// TODO
			// Random scalar should be changed when use in real product
			// Should get randomly in database the commitments
			val = val.Sub(privacy.RandomPoint(), sumOutputCom)
		}
		ring.AppendToRow(i, val)
	}

	return ring, &priv, pi, nil
}
