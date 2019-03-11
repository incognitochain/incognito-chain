package zkp

import (
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/constant-money/constant-chain/privacy"
	"github.com/stretchr/testify/assert"
)

func TestPKSNPrivacy(t *testing.T) {
	sk := privacy.GenerateSpendingKey([]byte{123})
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

	start := time.Now()
	proof, err := witness.Prove(nil)
	if err != nil {
		return
	}
	end := time.Since(start)
	fmt.Printf("Serial number proving time: %v\n", end)

	proofBytes := proof.Bytes()

	fmt.Printf("Serial number proof size: %v\n", len(proofBytes))

	proof2 := new(SNPrivacyProof).Init()
	proof2.SetBytes(proofBytes)

	start = time.Now()
	res := proof2.Verify(nil)
	end = time.Since(start)
	fmt.Printf("Serial number verification time: %v\n", end)

	assert.Equal(t, true, res)
}
