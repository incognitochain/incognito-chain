package zkp

import (
	"github.com/ninjadotorg/constant/privacy"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
)

func TestPKComProduct(t *testing.T) {
	index := privacy.VALUE
	G := privacy.PedCom.G[index]

	witnessA := privacy.RandInt()

	x := new(big.Int).ModInverse(witnessA, privacy.Curve.Params().N)

	r1Int := privacy.RandInt()

	ipCm := new(PKComProductWitness)
	invAmulG := new(privacy.EllipticPoint)
	*invAmulG = *G.ScalarMult(x)

	ipCm.Set(witnessA, r1Int, invAmulG, &index)
	proof, _ := ipCm.Prove()
	proofBytes := proof.Bytes()

	proof2 := new(PKComProductProof)
	proof2.SetBytes(proofBytes)

	res := proof2.Verify()
	assert.Equal(t, true, res)
}
