package privacy

import (
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
)

func TestSchnorrMultiSignature(t *testing.T) {
	n := 10
	// generate key sets for n members(s)
	keySets := make([]*MultiSigKeyset, n)
	listPKs := make([]*PublicKey, n)
	for i:=0; i<n; i++{
		keySets[i] = new(MultiSigKeyset)
		privateKey := GeneratePrivateKey(big.NewInt(int64(i)).Bytes())
		publicKey := GeneratePublicKey(privateKey)
		keySets[i].Set(&privateKey, &publicKey)
		listPKs[i] = &publicKey
	}

	// random message to sign
	data := RandScalar().Bytes()

	// each members generates a randomness (public and private) before signing
	secretRandomness := make([]*big.Int, n)
	publicRandomness := make([]*EllipticPoint, n)

	multiSigScheme := new(MultiSigScheme)
	combinedPublicRandomness := new(EllipticPoint).Zero()
	for i :=0 ; i<n ; i++{
		publicRandomness[i], secretRandomness[i] = multiSigScheme.GenerateRandom()
		combinedPublicRandomness = combinedPublicRandomness.Add(publicRandomness[i])
	}

	// each members sign on data
	sigs := make([]*SchnMultiSig, n)
	var err error
	for i, key := range keySets {
		sigs[i], err = key.SignMultiSig(data, listPKs, publicRandomness, secretRandomness[i])

		assert.Equal(t, nil, err)
		assert.Equal(t, SchnMultiSigSize, len(sigs[i].Bytes()))
	}

	// combine all of signatures
	combinedSig := multiSigScheme.CombineMultiSig(sigs)

	// verify combined signature
	isValid := combinedSig.VerifyMultiSig(data, listPKs, listPKs, combinedPublicRandomness)
	assert.Equal(t, true, isValid)

}
