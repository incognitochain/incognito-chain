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
	rSND := privacy.RandInt()

	serialNumber := privacy.PedCom.G[privacy.SK].Derive(skInt, SND)

	comSK := privacy.PedCom.CommitAtIndex(skInt, rSK, privacy.SK)
	comSND := privacy.PedCom.CommitAtIndex(SND, rSND, privacy.SND)

	stmt := new(SNPrivacyStatement)
	stmt.Set(serialNumber, comSK, comSND)
	witness := new(SNPrivacyWitness)
	witness.Set(stmt, skInt, rSK, SND, rSND)

	proof, err := witness.Prove(nil)
	if err != nil {
		return
	}

	proofBytes := proof.Bytes()

	fmt.Printf("Serial number proof size: %v\n", len(proofBytes))

	proof2 := new(SNPrivacyProof).Init()
	proof2.SetBytes(proofBytes)

	res := proof2.Verify(nil)

	assert.Equal(t, true, res)
}
