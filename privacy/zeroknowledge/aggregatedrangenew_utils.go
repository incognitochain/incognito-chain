package zkp

import (
	"github.com/ninjadotorg/constant/privacy"
	"math/big"
)


type InnerProductWitness struct {
	a []*big.Int
	b []*big.Int
	u *privacy.EllipticPoint
	p *privacy.EllipticPoint
	g []*privacy.EllipticPoint
	h []*privacy.EllipticPoint
}


type InnerProductProof struct {


	u *privacy.EllipticPoint
	p *privacy.EllipticPoint
	g []*privacy.EllipticPoint
	h []*privacy.EllipticPoint
}


//func (wit * InnerProductWitness) Prove(){
//	if len(wit.a) != len(wit.b) || len(wit.g) != len(wit.h) || len(wit.a) != len(wit.b)
//}
