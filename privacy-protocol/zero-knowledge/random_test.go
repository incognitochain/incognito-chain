package zkp

import (
	"fmt"
	"github.com/ninjadotorg/constant/privacy-protocol"
	"testing"
)

func TestRand(t* testing.T){
	a:=privacy.RandInt()
	fmt.Println(a.Sub(a,privacy.Curve.Params().N))
}
