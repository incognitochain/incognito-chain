package serialnumbernoprivacy

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/privacy/zeroknowledge/utils"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPKSNNoPrivacy(t *testing.T) {
	for i := 0; i < 1000; i++ {
		// prepare witness for Serial number no privacy protocol
		sk := privacy.GeneratePrivateKey(privacy.RandBytes(10))
		skScalar := new(privacy.Scalar).FromBytesS(sk)
		if skScalar.ScalarValid() == false {
			fmt.Println("Invalid key value")
		}

		pk := privacy.GeneratePublicKey(sk)
		pkPoint, err := new(privacy.Point).FromBytesS(pk)
		if err != nil {
			fmt.Println("Invalid point key valu")
		}
		SND := privacy.RandomScalar()

		serialNumber := new(privacy.Point).Derive(privacy.PedCom.G[privacy.PedersenPrivateKeyIndex], skScalar, SND)

		witness := new(SNNoPrivacyWitness)
		witness.Set(serialNumber, pkPoint, SND, skScalar)

		// proving
		proof, err := witness.Prove(nil)
		assert.Equal(t, nil, err)

		//validate sanity proof
		isValidSanity := proof.ValidateSanity()
		assert.Equal(t, true, isValidSanity)

		// verify proof
		res, err := proof.Verify(nil)
		assert.Equal(t, true, res)
		assert.Equal(t, nil, err)

		// convert proof to bytes array
		proofBytes := proof.Bytes()
		assert.Equal(t, utils.SnNoPrivacyProofSize, len(proofBytes))

		// new SNPrivacyProof to set bytes array
		proof2 := new(SNNoPrivacyProof).Init()
		err = proof2.SetBytes(proofBytes)
		assert.Equal(t, nil, err)
		assert.Equal(t, proof, proof2)

		// verify proof
		res2, err := proof2.Verify(nil)
		assert.Equal(t, true, res2)
		assert.Equal(t, nil, err)
	}

}
