package zkp

import (
	"github.com/ninjadotorg/constant/privacy"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
)

func TestPKSNPrivacy(t *testing.T) {
	sk := privacy.GenerateSpendingKey([]byte{123})
	skInt := new(big.Int).SetBytes(sk)
	SND := privacy.RandInt()

	rSK := privacy.RandInt()
	rSND1 := privacy.RandInt()
	rSND2 := privacy.RandInt()

	serialNumber := privacy.PedCom.G[privacy.SK].Derive(skInt, SND)

	comSK := privacy.PedCom.CommitAtIndex(skInt, rSK, privacy.SK)
	comSND1 := privacy.PedCom.CommitAtIndex(SND, rSND1, privacy.SND)
	comSND2 := privacy.PedCom.CommitAtIndex(SND, rSND2, privacy.SK)

	witness := new(PKSNPrivacyWitness)
	witness.Set(serialNumber, comSK, comSND1, comSND2, skInt, rSK, SND, rSND1, rSND2)

	proof, err := witness.Prove()
	if err != nil{
		return
	}

	proofBytes := proof.Bytes()

	proof2 := new(PKSNPrivacyProof).Init()
	proof2.SetBytes(proofBytes)

	res := proof2.Verify()

	assert.Equal(t, true, res)
}
