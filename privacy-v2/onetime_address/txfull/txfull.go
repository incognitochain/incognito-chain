package txfull

import (
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

func NewRingCTFull(inputs []ota.UTXO, fromAddress []address.PrivateAddress, sumBlindOutput *privacy.Scalar, outputs []ota.UTXO, toAddress []address.PublicAddress) *RingCTFull {
	return &RingCTFull{
		inputs,
		fromAddress,
		sumBlindOutput,
		outputs,
		toAddress,
	}
}

func (this *RingCTFull) Sign(message string) (*mlsag.Signature, error) {
	// TODO
	// In real system should change numFake
	numFake := 10

	// Generating Ring without commitment then add later
	pi := common.RandInt() % numFake
	privateKeys := getPrivateKeyOfInputs(this)
	ring := mlsag.NewRandomRing(privateKeys, numFake, pi)

	// Generate privateKey with commitment
	sumBlindInput, err := getSumBlindInput(this)
	if err != nil {
		return nil, err
	}
	privCommitment := new(privacy.Scalar).Sub(sumBlindInput, this.sumBlindOutput)
	privateKeys = append(privateKeys, *privCommitment)

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
		ring.AppendToRow(i, *val)
	}

	signer := mlsag.NewMlsagWithDefinedRing(privateKeys, ring, pi)
	return signer.Sign(message)
}
