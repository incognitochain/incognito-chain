package zkp

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/ninjadotorg/constant/privacy"
	"github.com/stretchr/testify/assert"
)

func TestPKSNPrivacy(t *testing.T) {
	sk := privacy.GenerateSpendingKey([]byte{123})
	skInt := new(big.Int).SetBytes(sk)
	SND := privacy.RandInt()

	rSK := privacy.RandInt()
	rSND1 := privacy.RandInt()

	serialNumber := privacy.PedCom.G[privacy.SK].Derive(skInt, SND)

	comSK := privacy.PedCom.CommitAtIndex(skInt, rSK, privacy.SK)
	comSND1 := privacy.PedCom.CommitAtIndex(SND, rSND1, privacy.SND)

	witness := new(PKSNPrivacyWitness)
	witness.Set(serialNumber, comSK, comSND1, skInt, rSK, SND, rSND1)

	proof, err := witness.Prove(nil)
	if err != nil {
		return
	}

	proofBytes := proof.Bytes()

	fmt.Printf("Serial number proof size: %v\n", len(proofBytes))

	proof2 := new(PKSNPrivacyProof).Init()
	proof2.SetBytes(proofBytes)

	res := proof2.Verify()

	assert.Equal(t, true, res)
}
