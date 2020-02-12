// HOLD DEBUG FUNCTIONS

package mlsag

import (
	"fmt"

	C25519 "github.com/incognitochain/incognito-chain/privacy/curve25519"
)

func debugRing(privateKeys []C25519.Key, K [][]C25519.Key, pi int) {
	fmt.Println("================")
	fmt.Println("Here comes Private Keys")
	fmt.Println(privateKeys)

	fmt.Println("================")
	fmt.Println("Here comes the ring")
	fmt.Println(K)

	fmt.Println("================")
	fmt.Println("Here comes Pi")
	fmt.Println(pi)

	fmt.Println("================")
	for i := 0; i < len(privateKeys); i += 1 {
		fmt.Printf("Checking Ring[%d]\n", i)
		fmt.Println(K[pi][i].ToBytes())
		fmt.Println(parsePublicKey(privateKeys[i]) == K[pi][i].ToBytes())
	}
}
