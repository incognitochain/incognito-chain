package zkp

import (
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/stretchr/testify/assert"
)

func TestPKSNPrivacy(t *testing.T) {
	// prepare witness for Serial number privacy protocol
	sk := privacy.GeneratePrivateKey([]byte{123})
	skInt := new(big.Int).SetBytes(sk)
	SND := privacy.RandScalar()

	rSK := privacy.RandScalar()
	rSND := privacy.RandScalar()

	serialNumber := privacy.PedCom.G[privacy.SK].Derive(skInt, SND)

	comSK := privacy.PedCom.CommitAtIndex(skInt, rSK, privacy.SK)
	comSND := privacy.PedCom.CommitAtIndex(SND, rSND, privacy.SND)

	stmt := new(SNPrivacyStatement)
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
	assert.Equal(t, privacy.SNPrivacyProofSize, len(proofBytes))

	// new SNPrivacyProof to set bytes array
	proof2 := new(SNPrivacyProof).Init()
	err = proof2.SetBytes(proofBytes)
	assert.Equal(t, nil, err)
	assert.Equal(t, proof, proof2)

	start = time.Now()
	res := proof2.Verify(nil)
	end = time.Since(start)
	fmt.Printf("Serial number verification time: %v\n", end)
	assert.Equal(t, true, res)
}
