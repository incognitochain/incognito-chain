package client

import (
	"fmt"
	"testing"
)

func TestHSigCRH(t *testing.T) {
	seed := [32]byte{1, 2, 3}
	nf1 := [32]byte{4, 5, 6}
	nf2 := [32]byte{7, 8, 9}
	pubKey := [32]byte{0, 1, 2}

	result := HSigCRH(seed[:], nf1[:], nf2[:], pubKey[:])
	fmt.Printf("%x\n", result)
}
