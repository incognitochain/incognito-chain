package zkp

import (
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
)

func TestPKSNNoPrivacy(t *testing.T) {
	// prepare witness for Serial number no privacy protocol
	sk := privacy.GeneratePrivateKey([]byte{123})
	skInt := new(big.Int).SetBytes(sk)

	pk := privacy.GeneratePublicKey(sk)
	pkPoint := new(privacy.EllipticPoint)
	pkPoint.Decompress(pk)

	SND := privacy.RandScalar()

	serialNumber := privacy.PedCom.G[privacy.SK].Derive(skInt, SND)

	witness := new(SNNoPrivacyWitness)
	witness.Set(serialNumber, pkPoint, SND, skInt)

	// proving
	proof, err := witness.Prove(nil)
	assert.Equal(t, nil, err)

	//validate sanity proof
	isValidSanity := proof.ValidateSanity()
	assert.Equal(t, true, isValidSanity)

	// convert proof to bytes array
	proofBytes := proof.Bytes()
	assert.Equal(t, privacy.SNNoPrivacyProofSize, len(proofBytes))

	// new SNPrivacyProof to set bytes array
	proof2 := new(SNNoPrivacyProof).Init()
	err = proof2.SetBytes(proofBytes)
	assert.Equal(t, nil, err)
	assert.Equal(t, proof, proof2)

	// verify proof
	res := proof2.Verify(nil)
	assert.Equal(t, true, res)
}
