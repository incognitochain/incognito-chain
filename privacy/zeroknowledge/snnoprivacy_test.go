package zkp

import (
	"fmt"
	"github.com/big0t/constant-chain/privacy"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
)

func TestPKSNNoPrivacy(t *testing.T) {
	sk := privacy.GenerateSpendingKey([]byte{123})
	skInt := new(big.Int).SetBytes(sk)

	pk := privacy.GeneratePublicKey(sk)
	pkPoint := new(privacy.EllipticPoint)

	err := pkPoint.Decompress(pk)
	if err != nil {
		return
	}

	SND := privacy.RandScalar()

	serialNumber := privacy.PedCom.G[privacy.SK].Derive(skInt, SND)

	witness := new(SNNoPrivacyWitness)
	witness.Set(serialNumber, pkPoint, SND, skInt)

	proof, err := witness.Prove(nil)
	if err != nil {
		return
	}

	proofBytes := proof.Bytes()

	fmt.Printf("Serial number proof size: %v\n", len(proofBytes))

	proof2 := new(SNNoPrivacyProof).Init()
	proof2.SetBytes(proofBytes)

	res := proof2.Verify(nil)

	assert.Equal(t, true, res)
}
