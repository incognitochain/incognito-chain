package blsmultisig

import (
	"fmt"

	"golang.org/x/crypto/bn256"
)

// Compress is the ...
func Compress() {
	fmt.Println(bn256.Order.BitLen())
}

// Decompress is
func Decompress() {
	fmt.Println(bn256.Order.Bits())
}
