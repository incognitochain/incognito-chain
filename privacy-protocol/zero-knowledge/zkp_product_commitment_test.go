package zkp

import (
	"fmt"
	"github.com/ninjadotorg/constant/privacy-protocol"

	//	"github.com/ninjadotorg/constant/privacy-protocol"
	"math/big"
	"testing"
)

func TestPKComProduct(t *testing.T) {
	res := true
	for i:=0;i<100;i++ {
		index := privacy.VALUE
		G:=privacy.PedCom.G[index]
		witnessA := new(big.Int).SetBytes(privacy.RandBytes(32))
		x:=new(big.Int)
		x.ModInverse(witnessA,privacy.Curve.Params().N)
		rA := privacy.RandBytes(32)
		r1Int := new(big.Int).SetBytes(rA)
		r1Int.Mod(r1Int, privacy.Curve.Params().N)
		ipCm:= new(PKComProductWitness)
		invAmulG:=new(privacy.EllipticPoint)
		*invAmulG = *G.ScalarMul(x)
		ipCm.Set(witnessA,r1Int,invAmulG,&index)
		proof,_:= ipCm.Prove()
		//proof.Print()
		//fmt.Println(proof.index)
		//fmt.Println()
		proof_bytes:= proof.Bytes()
		fmt.Println(len(proof_bytes))
		Vproof:=new(PKComProductProof)
		Vproof.SetBytes(proof_bytes)
		//fmt.Println("------------------------------------")
		//Vproof.Print()
		//fmt.Println(Vproof.index)


		res = Vproof.Verify();
		fmt.Printf("Test %d is %t\n",i,res)
	}
}
