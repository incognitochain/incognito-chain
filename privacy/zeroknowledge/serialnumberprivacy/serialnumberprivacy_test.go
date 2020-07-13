package serialnumberprivacy

import (
	"fmt"
	"testing"
	"time"

	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/privacy/zeroknowledge/utils"
	"github.com/stretchr/testify/assert"
)

func TestPKSNPrivacy(t *testing.T) {
	for i := 0; i < 1000; i++ {
		sk := privacy.GeneratePrivateKey(privacy.RandBytes(31))
		skScalar := new(privacy.Scalar).FromBytesS(sk)
		if skScalar.ScalarValid() == false {
			fmt.Println("Invalid scala key value")
		}

		SND := privacy.RandomScalar()
		rSK := privacy.RandomScalar()
		rSND := privacy.RandomScalar()

		serialNumber := new(privacy.Point).Derive(privacy.PedCom.G[privacy.PedersenPrivateKeyIndex], skScalar, SND)
		comSK := privacy.PedCom.CommitAtIndex(skScalar, rSK, privacy.PedersenPrivateKeyIndex)
		comSND := privacy.PedCom.CommitAtIndex(SND, rSND, privacy.PedersenSndIndex)

		stmt := new(SerialNumberPrivacyStatement)
		stmt.Set(serialNumber, comSK, comSND)

		witness := new(SNPrivacyWitness)
		witness.Set(stmt, skScalar, rSK, SND, rSND)

		// proving
		start := time.Now()
		proof, err := witness.Prove(nil)
		assert.Equal(t, nil, err)

		end := time.Since(start)
		fmt.Printf("Serial number proving time: %v\n", end)

		//validate sanity proof
		isValidSanity := proof.ValidateSanity()
		assert.Equal(t, true, isValidSanity)

		// convert proof to bytes array
		proofBytes := proof.Bytes()
		assert.Equal(t, utils.SnPrivacyProofSize, len(proofBytes))

		// new SNPrivacyProof to set bytes array
		proof2 := new(SNPrivacyProof).Init()
		err = proof2.SetBytes(proofBytes)
		assert.Equal(t, nil, err)
		assert.Equal(t, proof, proof2)

		start = time.Now()
		res, err := proof2.Verify(nil)
		end = time.Since(start)
		fmt.Printf("Serial number verification time: %v\n", end)
		assert.Equal(t, true, res)
		assert.Equal(t, nil, err)
	}
}
