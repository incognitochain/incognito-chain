package gomobile

import (
	"encoding/base64"
	"fmt"
	"github.com/incognitochain/incognito-chain/consensus/blsmultisig"
)


// GenerateBLSKeyPairFromSeed generates BLS key pair from seed
func GenerateBLSKeyPairFromSeed(args string) string {
	// convert seed from string to bytes array
	fmt.Printf("args: %v\n", args)
	seed := []byte(args)
	fmt.Printf("Bytes: %v\n", seed)

	// generate  bls key
	privateKey, publicKey := blsmultisig.KeyGen(seed)

	// append key pair to one bytes array
	keyPairBytes := []byte{}
	keyPairBytes = append(keyPairBytes, privateKey.Bytes()...)
	keyPairBytes = append(keyPairBytes, blsmultisig.CmprG2(publicKey)...)

	//  base64.StdEncoding.EncodeToString()
	keyPairEncode := base64.StdEncoding.EncodeToString(keyPairBytes)

	return keyPairEncode
}