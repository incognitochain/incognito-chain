package zkp

import (
	"fmt"
	"github.com/ninjadotorg/constant/privacy-protocol"
	"testing"
)

func TestRand(t* testing.T){
	for i:=0;i<100;i++{
	a:=privacy.RandInt()
	fmt.Println(a)
	}
}
