// HOLD DEBUG FUNCTIONS

package mlsag

import (
	"fmt"
)

func (this *Mlsag) debugRing() {
	fmt.Println("================")
	fmt.Println("Here comes Private Keys")
	fmt.Println(this.privateKeys)

	fmt.Println("================")
	fmt.Println("Here comes the ring")
	fmt.Println(this.K)

	fmt.Println("================")
	fmt.Println("Here comes Pi")
	fmt.Println(this.pi)

	fmt.Println("================")
	for i := 0; i < len(this.privateKeys); i += 1 {
		fmt.Printf("Checking Ring[%d]\n", i)
		fmt.Println(this.K.keys[this.pi][i])
		fmt.Println(*parsePublicKey(this.privateKeys[i]) == this.K.keys[this.pi][i])
		// fmt.Println(parsePublicKey(this.privateKeys[i]) == this.K.keys[this.pi][i].ToBytes())
	}
}
