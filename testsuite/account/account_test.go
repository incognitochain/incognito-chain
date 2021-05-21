package account

import (
	"fmt"
	"testing"
)

func TestNewAccountFromPrivatekey(t *testing.T) {
	x, _ := NewAccountFromPrivatekey("112t8sw7CFBzQb33w2uZ9s3aEeKVWYx2LEWQNGDLc6aTNEVoWP2dBihDWYs2gcWfKWUVoeVCKKm6WFz1u5V2VqoXrpSZsqguCN3jSG4nAqTR")
	fmt.Printf("%+v", x)
}
