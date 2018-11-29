package zkp

import (
	"fmt"
	"github.com/ninjadotorg/constant/privacy-protocol"
	"math/big"
	"testing"
)

func TestFunc(t* testing.T){
	for i:=0;i<1000;i++ {
		r := new(big.Int).SetBytes(privacy.RandBytes(32))
		v := new(big.Int).SetBytes(privacy.RandBytes(32))
		x := privacy.PedCom.CommitAtIndex(v, r, 0)
		fmt.Println(x)
	}
}

