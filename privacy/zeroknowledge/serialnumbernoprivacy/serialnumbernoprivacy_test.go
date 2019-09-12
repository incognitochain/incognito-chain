package serialnumbernoprivacy

import (
	"crypto/rand"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/privacy/zeroknowledge/utils"
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

	var r = rand.Reader
	SND := privacy.RandScalar(r)

	serialNumber := privacy.PedCom.G[privacy.PedersenPrivateKeyIndex].Derive(skInt, SND)

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
	assert.Equal(t, utils.SnNoPrivacyProofSize, len(proofBytes))

	// new SNPrivacyProof to set bytes array
	proof2 := new(SNNoPrivacyProof).Init()
	err = proof2.SetBytes(proofBytes)
	assert.Equal(t, nil, err)
	assert.Equal(t, proof, proof2)

	// verify proof
	res, err := proof2.Verify(nil)
	assert.Equal(t, true, res)
	assert.Equal(t, nil, err)
}
