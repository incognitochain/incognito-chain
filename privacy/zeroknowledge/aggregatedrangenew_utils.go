package zkp

import (
	"errors"
	"github.com/ninjadotorg/constant/privacy"
	"math"
	"math/big"
)


type InnerProductWitness struct {
	a []*big.Int
	b []*big.Int
	u *privacy.EllipticPoint
	p *privacy.EllipticPoint
}


type InnerProductProof struct {

	u *privacy.EllipticPoint
	p *privacy.EllipticPoint
}


func (wit * InnerProductWitness) Prove() (*InnerProductProof, error){
	if len(wit.a) != len(wit.b) {
		return nil, errors.New("invalid inputs")
	}

	//n := len(wit.a)
	//
	//for n > 1{
	//	n2 := n/2
	//
	//	cL, err := innerProduct(wit.a[:n2], wit.b[:n2])
	//	if err != nil{
	//		return  nil, err
	//	}
	//
	//	cR, err := innerProduct(wit.a[n2:], wit.b[n2:])
	//	if err != nil{
	//		return  nil, err
	//	}
	//
	//}



	return nil, nil
}

func pad(l int) int {
	deg := 0
	for l > 0 {
		if l%2 == 0 {
			deg++
			l = l / 2
		} else {
			break
		}
	}
	i := 0
	for {
		if math.Pow(2, float64(i)) < float64(l) {
			i++
		} else {
			l = int(math.Pow(2, float64(i+deg)))
			break
		}
	}
	return l
}


/*-----------------------------Vector Functions-----------------------------*/
// The length here always has to be a power of two
// innerProduct calculates inner product between two vectors a and b
func innerProduct(a []*big.Int, b []*big.Int) (*big.Int, error) {
	if len(a) != len(b) {
		return nil, errors.New("InnerProduct: Arrays not of the same length")
	}

	c := big.NewInt(0)
	tmp := new(big.Int)

	for i := range a {
		c.Add(c, tmp.Mul(a[i], b[i]))
	}
	c.Mod(c, privacy.Curve.Params().N)

	return c, nil
}
