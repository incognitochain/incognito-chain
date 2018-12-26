package zkp

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/ninjadotorg/constant/privacy"
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

	serialNumber := privacy.Eval(skInt, SND, privacy.PedCom.G[privacy.SK])

	comSK := privacy.PedCom.CommitAtIndex(skInt, rSK, privacy.SK)
	comSND1 := privacy.PedCom.CommitAtIndex(SND, rSND1, privacy.SND)
	comSND2 := privacy.PedCom.CommitAtIndex(SND, rSND2, privacy.SK)

	witness := new(PKSNPrivacyWitness)
	witness.Set(serialNumber, comSK, comSND1, comSND2, skInt, rSK, SND, rSND1, rSND2)

	proof, err := witness.Prove()
	if err != nil{
		fmt.Println(err)
	}

	proofBytes := proof.Bytes()

	proof2 := new(PKSNPrivacyProof).Init()
	proof2.SetBytes(proofBytes)

	res := proof2.Verify()

	assert.Equal(t, true, res)
}
