package privacy

import (
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
	"time"
	"fmt"
)

func TestSchnorrMultiSignature(t *testing.T) {
	n := 100
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
	sigs := make([]*SchnMultiSig, n )
	var err error
	start1 := time.Now()
	for i := 0; i<n; i++{
		sigs[i], err = keySets[i].SignMultiSig(data, listPKs, publicRandomness, secretRandomness[i])

		assert.Equal(t, nil, err)
		assert.Equal(t, SchnMultiSigSize, len(sigs[i].Bytes()))
	}

	end1 := time.Since(start1)
	fmt.Printf("Time1: %v\n", end1)


	// combine all of signatures
	start2 := time.Now()
	combinedSig := multiSigScheme.CombineMultiSig(sigs)
	end2 := time.Since(start2)
	fmt.Printf("Time2: %v\n", end2)

	// verify combined signature
	start3 := time.Now()
	listCombinedPKs := listPKs[:n]
	isValid := combinedSig.VerifyMultiSig(data, listPKs, listCombinedPKs, combinedPublicRandomness)
	end3 := time.Since(start3)
	fmt.Printf("Time3: %v\n", end3)
	assert.Equal(t, true, isValid)

}
