package zkp

import (
	"github.com/ninjadotorg/constant/privacy"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
)

func TestPKSNNoPrivacy(t *testing.T) {
	sk := privacy.GenerateSpendingKey([]byte{123})
	skInt := new(big.Int).SetBytes(sk)

	pk := privacy.GeneratePublicKey(sk)
	pkPoint, err := privacy.DecompressKey(pk)
	if err != nil{
		return
	}

	SND := privacy.RandInt()

	serialNumber := privacy.Eval(skInt, SND, privacy.PedCom.G[privacy.SK])

	witness := new(SNNoPrivacyWitness)
	witness.Set(serialNumber, pkPoint, SND, skInt)

	proof, err := witness.Prove()
	if err != nil{
		return
	}

	proofBytes := proof.Bytes()

	proof2 := new(SNNoPrivacyProof).Init()
	proof2.SetBytes(proofBytes)

	res := proof2.Verify()

	assert.Equal(t, true, res)
}
