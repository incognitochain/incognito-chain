package serialnumberprivacy

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/privacy/zeroknowledge/utils"
	"github.com/stretchr/testify/assert"
)

func TestPKSNPrivacy(t *testing.T) {
	// prepare witness for Serial number privacy protocol
	sk := privacy.GeneratePrivateKey([]byte{123})
	skInt := new(big.Int).SetBytes(sk)
	var r = rand.Reader
	SND := privacy.RandScalar(r)

	rSK := privacy.RandScalar(r)
	rSND := privacy.RandScalar(r)

	serialNumber := privacy.PedCom.G[privacy.PedersenPrivateKeyIndex].Derive(skInt, SND)

	comSK := privacy.PedCom.CommitAtIndex(skInt, rSK, privacy.PedersenPrivateKeyIndex)
	comSND := privacy.PedCom.CommitAtIndex(SND, rSND, privacy.PedersenSndIndex)

	stmt := new(SerialNumberPrivacyStatement)
	stmt.Set(serialNumber, comSK, comSND)

	witness := new(SNPrivacyWitness)
	witness.Set(stmt, skInt, rSK, SND, rSND)

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
