package serialnumbernoprivacy

import (
	"fmt"
	"testing"

	"github.com/incognitochain/incognito-chain/privacy/key"
	"github.com/incognitochain/incognito-chain/privacy/operation"
	"github.com/incognitochain/incognito-chain/privacy/privacy_util"
	"github.com/incognitochain/incognito-chain/privacy/privacy_v1/zeroknowledge/utils"
	"github.com/stretchr/testify/assert"
)

func TestPKSNNoPrivacy(t *testing.T) {
	for i := 0; i < 1000; i++ {
		// prepare witness for Serial number no privacy protocol
		sk := key.GeneratePrivateKey(privacy_util.RandBytes(10))
		skScalar := new(operation.Scalar).FromBytesS(sk)
		if skScalar.ScalarValid() == false {
			fmt.Println("Invalid key value")
		}

		pk := key.GeneratePublicKey(sk)
		pkPoint, err := new(operation.Point).FromBytesS(pk)
		if err != nil {
			fmt.Println("Invalid point key valu")
		}
		SND := operation.RandomScalar()

		serialNumber := new(operation.Point).Derive(operation.PedCom.G[operation.PedersenPrivateKeyIndex], skScalar, SND)

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
