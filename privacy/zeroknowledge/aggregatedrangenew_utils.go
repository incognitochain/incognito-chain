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
	L []*privacy.EllipticPoint
	R []*privacy.EllipticPoint
	a *big.Int
	b *big.Int

	u *privacy.EllipticPoint
	p *privacy.EllipticPoint
}

func (wit *InnerProductWitness) Prove() (*InnerProductProof, error) {
	if len(wit.a) != len(wit.b) {
		return nil, errors.New("invalid inputs")
	}

	n := len(wit.a)

	a := make([]*big.Int, n)
	b := make([]*big.Int, n)

	for i := range wit.a {
		a[i] = new(big.Int)
		a[i].Set(wit.a[i])

		b[i] = new(big.Int)
		b[i].Set(wit.b[i])
	}

	p := new(privacy.EllipticPoint)
	p.Set(wit.p.X, wit.p.Y)

	G := make([]*privacy.EllipticPoint, n)
	H := make([]*privacy.EllipticPoint, n)
	for i := range G {
		G[i] = new(privacy.EllipticPoint)
		G[i].Set(AggParam.G[i].X, AggParam.G[i].Y)

		H[i] = new(privacy.EllipticPoint)
		H[i].Set(AggParam.H[i].X, AggParam.H[i].Y)
	}

	proof := new(InnerProductProof)
	proof.L = make([]*privacy.EllipticPoint, 0)
	proof.R = make([]*privacy.EllipticPoint, 0)
	proof.u = wit.u
	proof.p = wit.p

	for n > 1 {
		nPrime := n / 2

		cL, err := innerProduct(a[:nPrime], b[nPrime:])
		if err != nil {
			return nil, err
		}

		cR, err := innerProduct(a[nPrime:], b[:nPrime])
		if err != nil {
			return nil, err
		}

		L, err := EncodeVectors(a[:nPrime], b[nPrime:], G[nPrime:], H[:nPrime])
		if err != nil {
			return nil, err
		}
		L = L.Add(wit.u.ScalarMult(cL))
		proof.L = append(proof.L, L)

		R, err := EncodeVectors(a[nPrime:], b[:nPrime], G[:nPrime], H[nPrime:])
		if err != nil {
			return nil, err
		}
		R = R.Add(wit.u.ScalarMult(cR))
		proof.R = append(proof.R, R)

		// calculate challenge x = hash(G || H || u || p ||  L || R)
		x := generateChallengeForAggRange([]*privacy.EllipticPoint{wit.u, p, L, R})
		xInverse := new(big.Int).ModInverse(x, privacy.Curve.Params().N)

		// calculate GPrime, HPrime, PPrime for the next loop
		GPrime := make([]*privacy.EllipticPoint, nPrime)
		HPrime := make([]*privacy.EllipticPoint, nPrime)

		for i := range GPrime {
			GPrime[i] = G[i].ScalarMult(xInverse).Add(G[i+nPrime].ScalarMult(x))
			HPrime[i] = H[i].ScalarMult(x).Add(H[i+nPrime].ScalarMult(xInverse))
		}

		xSquare := new(big.Int).Mul(x, x)
		xSquareInverse := new(big.Int).ModInverse(xSquare, privacy.Curve.Params().N)

		//PPrime := L.ScalarMult(xSquare).Add(p).Add(R.ScalarMult(xSquareInverse)) // x^2 * L + P + xInverse^2 * R
		PPrime := L.ScalarMult(xSquare).Add(p).Add(R.ScalarMult(xSquareInverse)) // x^2 * L + P + xInverse^2 * R

		// calculate aPrime, bPrime
		aPrime := make([]*big.Int, nPrime)
		bPrime := make([]*big.Int, nPrime)
		tmp := new(big.Int)

		for i := range aPrime {
			aPrime[i] = new(big.Int).Mul(a[i+nPrime], x)
			aPrime[i].Add(aPrime[i], tmp.Mul(a[i], xInverse))
			aPrime[i].Mod(aPrime[i], privacy.Curve.Params().N)

			bPrime[i] = new(big.Int).Mul(b[i+nPrime], xInverse)
			bPrime[i].Add(bPrime[i], tmp.Mul(b[i], x))
			bPrime[i].Mod(bPrime[i], privacy.Curve.Params().N)
		}

		a = aPrime
		b = bPrime
		p.Set(PPrime.X, PPrime.Y)
		G = GPrime
		H = HPrime
		n = nPrime
	}

	proof.a = a[0]
	proof.b = b[0]

	return proof, nil
}

func (proof *InnerProductProof) Verify() bool {
	p := new(privacy.EllipticPoint)
	p.Set(proof.p.X, proof.p.Y)

	n := privacy.MaxExp

	G := make([]*privacy.EllipticPoint, n)
	H := make([]*privacy.EllipticPoint, n)
	for i := range G {
		G[i] = new(privacy.EllipticPoint)
		G[i].Set(AggParam.G[i].X, AggParam.G[i].Y)

		H[i] = new(privacy.EllipticPoint)
		H[i].Set(AggParam.H[i].X, AggParam.H[i].Y)
	}

	for i := range proof.L {
		nPrime := n / 2
		// calculate challenge x = hash(G || H || u || p ||  L || R)
		x := generateChallengeForAggRange([]*privacy.EllipticPoint{proof.u, p, proof.L[i], proof.R[i]})
		xInverse := new(big.Int).ModInverse(x, privacy.Curve.Params().N)

		// calculate GPrime, HPrime, PPrime for the next loop
		GPrime := make([]*privacy.EllipticPoint, nPrime)
		HPrime := make([]*privacy.EllipticPoint, nPrime)

		for i := range GPrime {
			GPrime[i] = G[i].ScalarMult(xInverse).Add(G[i+nPrime].ScalarMult(x))
			HPrime[i] = H[i].ScalarMult(x).Add(H[i+nPrime].ScalarMult(xInverse))
		}

		xSquare := new(big.Int).Mul(x, x)
		xSquareInverse := new(big.Int).ModInverse(xSquare, privacy.Curve.Params().N)

		//PPrime := L.ScalarMult(xSquare).Add(p).Add(R.ScalarMult(xSquareInverse)) // x^2 * L + P + xInverse^2 * R
		PPrime := proof.L[i].ScalarMult(xSquare).Add(p).Add(proof.R[i].ScalarMult(xSquareInverse)) // x^2 * L + P + xInverse^2 * R

		p = PPrime
		G = GPrime
		H = HPrime
		n = nPrime
	}

	c := new(big.Int).Mul(proof.a, proof.b)

	rightPoint := G[0].ScalarMult(proof.a)
	rightPoint = rightPoint.Add(H[0].ScalarMult(proof.b))
	rightPoint = rightPoint.Add(proof.u.ScalarMult(c))

	if rightPoint.IsEqual(p) {
		return true
	}
	return false
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

//func hadamardProduct(a []*big.Int, b []*big.Int) ([]*big.Int, error) {
//	if len(a) != len(b) {
//		return nil, errors.New("InnerProduct: Arrays not of the same length")
//	}
//
//	c := make([]*big.Int, len(a))
//	for i := 0; i < len(c); i++ {
//		c[i] = new(big.Int).Mul(a[i], b[i])
//		c[i].Mod(c[i], privacy.Curve.Params().N)
//	}
//
//	return c, nil
//}
