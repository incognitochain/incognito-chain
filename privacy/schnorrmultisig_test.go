package privacy

import (
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/incognitochain/incognito-chain/common"

	"github.com/stretchr/testify/assert"
)

func printHex(x []byte) {
	temp := x
	tempLen := len(temp)
	tmp := [32]byte{0}
	for j := 31; j >= 0; j-- {
		tempLen--
		tmp[j] = temp[tempLen]
	}
	fmt.Printf("\"0x")
	for j := 0; j < 32; j++ {
		fmt.Printf("%x%x", tmp[j]>>4, tmp[j]%16)
	}
	fmt.Printf("\", ")
}

func TestSchnorrMultiSignature(t *testing.T) {
	n := 8
	// generate key sets for n members(s)
	keySets := make([]*MultiSigKeyset, n)
	listPKs := make([]*PublicKey, n)
	listPkPoints := make([]EllipticPoint, n)
	for i := 0; i < n; i++ {
		keySets[i] = new(MultiSigKeyset)
		privateKey := GeneratePrivateKey(big.NewInt(int64(i)).Bytes())
		publicKey := GeneratePublicKey(privateKey)
		keySets[i].Set(&privateKey, &publicKey)
		listPKs[i] = &publicKey
		listPkPoints[i].Decompress(*listPKs[i])
	}

	fmt.Println("X Coor point")
	for i := 0; i < n; i++ {
		printHex(listPkPoints[i].X.Bytes())
	}
	fmt.Printf("\n")
	fmt.Println("Y Coor point")
	for i := 0; i < n; i++ {
		printHex(listPkPoints[i].Y.Bytes())
	}
	fmt.Printf("\n")

	// random message to sign
	data := []byte{50, 30, 179, 190, 122, 161, 29, 184, 20, 123, 94, 62, 60, 134, 200, 20, 250, 211, 152, 16, 131, 222, 168, 160, 188, 237, 76, 113, 44, 220, 78, 42}

	// each members generates a randomness (public and private) before signing
	secretRandomness := make([]*big.Int, n)
	publicRandomness := make([]*EllipticPoint, n)

	multiSigScheme := new(MultiSigScheme)
	combinedPublicRandomness := new(EllipticPoint).Zero()
	for i := 0; i < n; i++ {
		seed := big.NewInt(int64(i + 10))
		publicRandomness[i], secretRandomness[i] = multiSigScheme.GenerateRandomFromSeed(seed)
		combinedPublicRandomness = combinedPublicRandomness.Add(publicRandomness[i])
	}

	fmt.Printf("XR: ")
	printHex(combinedPublicRandomness.X.Bytes())
	fmt.Println()
	fmt.Printf("YR: ")
	printHex(combinedPublicRandomness.Y.Bytes())
	fmt.Println()

	fmt.Printf("R: %+v\n", combinedPublicRandomness.Compress())

	// each members sign on data
	sigs := make([]*SchnMultiSig, n)
	var err error
	start1 := time.Now()
	for i := 0; i < n; i++ {
		sigs[i] = keySets[i].SignMultiSig(data, listPKs, publicRandomness, secretRandomness[i])

		assert.Equal(t, nil, err)
		assert.Equal(t, SchnMultiSigSize, len(sigs[i].Bytes()))
	}

	end1 := time.Since(start1)
	fmt.Printf("Time1: %v\n", end1)

	// combine all of signatures
	start2 := time.Now()
	combinedSig := multiSigScheme.CombineMultiSig(sigs)
	fmt.Println("Sig:")
	printHex(combinedSig.S.Bytes())
	fmt.Println("---")

	end2 := time.Since(start2)
	fmt.Printf("Time2: %v\n", end2)

	// verify combined signature
	start3 := time.Now()
	listCombinedPKs := listPKs[:n]
	isValid := combinedSig.VerifyMultiSig(data, listPKs, listCombinedPKs, combinedPublicRandomness)
	end3 := time.Since(start3)
	fmt.Printf("Time3: %v\n", end3)
	assert.Equal(t, true, isValid)

	printHex(common.HashB(data))
}
