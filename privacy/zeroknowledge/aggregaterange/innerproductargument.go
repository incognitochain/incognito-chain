package aggregaterange

import (
	"errors"
	"fmt"
	"github.com/incognitochain/incognito-chain/privacy"
)

type InnerProductWitness struct {
	a []*privacy.Scalar
	b []*privacy.Scalar

	p *privacy.Point
}

type InnerProductProof struct {
	l []*privacy.Point
	r []*privacy.Point
	a *privacy.Scalar
	b *privacy.Scalar

	p *privacy.Point
}

func (proof InnerProductProof) ValidateSanity() bool {
	if len(proof.l) != len(proof.r) {
		return false
	}

	for i := 0; i < len(proof.l); i++ {
		if !proof.l[i].PointValid() {
			return false
		}

		if !proof.r[i].PointValid() {
			return false
		}
	}

	if !proof.a.ScalarValid() {
		return false
	}
	if !proof.b.ScalarValid() {
		return false
	}

	return proof.p.PointValid()
}

func (proof InnerProductProof) Bytes() []byte {
	var res []byte

	res = append(res, byte(len(proof.l)))
	for _, l := range proof.l {
		res = append(res, l.ToBytesS()...)
	}

	for _, r := range proof.r {
		res = append(res, r.ToBytesS()...)
	}

	res = append(res, proof.a.ToBytesS()...)
	res = append(res, proof.b.ToBytesS()...)
	res = append(res, proof.p.ToBytesS()...)

	return res
}

func (proof *InnerProductProof) SetBytes(bytes []byte) error {
	if len(bytes) == 0 {
		return nil
	}

	lenLArray := int(bytes[0])
	offset := 1
	var err error

	proof.l = make([]*privacy.Point, lenLArray)
	for i := 0; i < lenLArray; i++ {
		proof.l[i], err = new(privacy.Point).FromBytesS(bytes[offset : offset+privacy.Ed25519KeySize])
		if err != nil {
			return err
		}
		offset += privacy.Ed25519KeySize
	}

	proof.r = make([]*privacy.Point, lenLArray)
	for i := 0; i < lenLArray; i++ {
		proof.r[i], err = new(privacy.Point).FromBytesS(bytes[offset : offset+privacy.Ed25519KeySize])
		if err != nil {
			return err
		}
		offset += privacy.Ed25519KeySize
	}

	proof.a = new(privacy.Scalar).FromBytesS(bytes[offset : offset+privacy.Ed25519KeySize])
	offset += privacy.Ed25519KeySize

	proof.b = new(privacy.Scalar).FromBytesS(bytes[offset : offset+privacy.Ed25519KeySize])
	offset += privacy.Ed25519KeySize

	proof.p, err = new(privacy.Point).FromBytesS(bytes[offset : offset+privacy.Ed25519KeySize])
	if err != nil {
		return err
	}

	return nil
}

func (wit InnerProductWitness) Prove(AggParam *bulletproofParams) (*InnerProductProof, error) {
	//var AggParam = newBulletproofParams(1)
	if len(wit.a) != len(wit.b) {
		return nil, errors.New("invalid inputs")
	}

	n := len(wit.a)

	a := make([]*privacy.Scalar, n)
	b := make([]*privacy.Scalar, n)

	for i := range wit.a {
		a[i] = new(privacy.Scalar).Set(wit.a[i])
		b[i] = new(privacy.Scalar).Set(wit.b[i])
	}

	p := new(privacy.Point).Set(wit.p)
	G := make([]*privacy.Point, n)
	H := make([]*privacy.Point, n)
	for i := range G {
		G[i] = new(privacy.Point).Set(AggParam.g[i])
		H[i] = new(privacy.Point).Set(AggParam.h[i])
	}

	proof := new(InnerProductProof)
	proof.l = make([]*privacy.Point, 0)
	proof.r = make([]*privacy.Point, 0)
	proof.p = new(privacy.Point).Set(wit.p)

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

		L, err := encodeVectors(a[:nPrime], b[nPrime:], G[nPrime:], H[:nPrime])
		if err != nil {
			return nil, err
		}
		L.Add(L, new(privacy.Point).ScalarMult(AggParam.u, cL))
		proof.l = append(proof.l, L)

		R, err := encodeVectors(a[nPrime:], b[:nPrime], G[:nPrime], H[nPrime:])
		if err != nil {
			return nil, err
		}
		R.Add(R, new(privacy.Point).ScalarMult(AggParam.u, cR))
		proof.r = append(proof.r, R)

		// calculate challenge x = hash(G || H || u || p ||  l || r)
		x := generateChallengeForAggRange(AggParam, [][]byte{p.ToBytesS(), L.ToBytesS(), R.ToBytesS()})
		xInverse := new(privacy.Scalar).Invert(x)
		xSquare := new(privacy.Scalar).Mul(x, x)
		xSquareInverse := new(privacy.Scalar).Mul(xInverse, xInverse)

		// calculate GPrime, HPrime, PPrime for the next loop
		GPrime := make([]*privacy.Point, nPrime)
		HPrime := make([]*privacy.Point, nPrime)

		for i := range GPrime {
			//GPrime[i] = new(privacy.Point).ScalarMult(G[i], xInverse)
			//GPrime[i].Add(GPrime[i], new(privacy.Point).ScalarMult(G[i+nPrime], x))
			//GPrime[i] = new(privacy.Point).AddPedersen(xInverse, G[i], x, G[i+nPrime])
			GPrime[i] = new(privacy.Point).AddPedersen(xInverse, G[i], x, G[i+nPrime])

			//HPrime[i] = new(privacy.Point).ScalarMult(H[i], x)
			//HPrime[i].Add(HPrime[i], new(privacy.Point).ScalarMult(H[i+nPrime], xInverse))
			HPrime[i] = new(privacy.Point).AddPedersen(x, H[i], xInverse, H[i+nPrime])
		}

		// x^2 * l + P + xInverse^2 * r
		PPrime := new(privacy.Point).AddPedersen(xSquare, L, xSquareInverse, R)
		PPrime.Add(PPrime, p)

		// calculate aPrime, bPrime
		aPrime := make([]*privacy.Scalar, nPrime)
		bPrime := make([]*privacy.Scalar, nPrime)

		for i := range aPrime {
			aPrime[i] = new(privacy.Scalar).Mul(a[i], x)
			aPrime[i] = new(privacy.Scalar).MulAdd(a[i+nPrime], xInverse, aPrime[i])

			bPrime[i] = new(privacy.Scalar).Mul(b[i], xInverse)
			bPrime[i] = new(privacy.Scalar).MulAdd(b[i+nPrime], x, bPrime[i])
		}

		a = aPrime
		b = bPrime
		p.Set(PPrime)
		G = GPrime
		H = HPrime
		n = nPrime
	}

	proof.a = new(privacy.Scalar).Set(a[0])
	proof.b = new(privacy.Scalar).Set(b[0])

	return proof, nil
}

func (proof InnerProductProof) Verify(AggParam *bulletproofParams) bool {
	//var AggParam = newBulletproofParams(1)
	p := new(privacy.Point)
	p.Set(proof.p)

	n := len(AggParam.g)

	G := make([]*privacy.Point, n)
	H := make([]*privacy.Point, n)
	for i := range G {
		G[i] = new(privacy.Point).Set(AggParam.g[i])
		H[i] = new(privacy.Point).Set(AggParam.h[i])
	}

	for i := range proof.l {
		nPrime := n / 2
		// calculate challenge x = hash(G || H || u || p ||  l || r)
		x := generateChallengeForAggRange(AggParam, [][]byte{p.ToBytesS(), proof.l[i].ToBytesS(), proof.r[i].ToBytesS()})
		xInverse := new(privacy.Scalar).Invert(x)
		xSquare := new(privacy.Scalar).Mul(x, x)
		xSquareInverse := new(privacy.Scalar).Mul(xInverse, xInverse)

		// calculate GPrime, HPrime, PPrime for the next loop
		GPrime := make([]*privacy.Point, nPrime)
		HPrime := make([]*privacy.Point, nPrime)

		for j := 0; j < len(GPrime); j++ {
			//GPrime[j] = new(privacy.Point).ScalarMult(G[j], xInverse)
			//GPrime[j].Add(GPrime[j], new(privacy.Point).ScalarMult(G[j+nPrime], x))
			GPrime[j] = new(privacy.Point).AddPedersen(xInverse, G[j], x, G[j+nPrime])

			//HPrime[j] = new(privacy.Point).ScalarMult(H[j], x)
			//HPrime[j].Add(HPrime[j], new(privacy.Point).ScalarMult(H[j+nPrime], xInverse))
			HPrime[j] = new(privacy.Point).AddPedersen(x, H[j], xInverse, H[j+nPrime])
		}

		//PPrime := l.ScalarMul(xSquare).Add(p).Add(r.ScalarMul(xSquareInverse)) // x^2 * l + P + xInverse^2 * r
		PPrime := new(privacy.Point).AddPedersen(xSquare, proof.l[i], xSquareInverse, proof.r[i])
		PPrime.Add(PPrime, p) // x^2 * l + P + xInverse^2 * r

		p = PPrime
		G = GPrime
		H = HPrime
		n = nPrime
	}

	c := new(privacy.Scalar).Mul(proof.a, proof.b)

	rightPoint := new(privacy.Point).AddPedersen(proof.a, G[0], proof.b, H[0])
	rightPoint.Add(rightPoint, new(privacy.Point).ScalarMult(AggParam.u, c))

	res := privacy.IsPointEqual(rightPoint, p)
	if !res {
		privacy.Logger.Log.Error("Inner product argument failed:")
		privacy.Logger.Log.Error("p: %v\n", p)
		privacy.Logger.Log.Error("rightPoint: %v\n", rightPoint)
		fmt.Printf("Inner product argument failed:")
		fmt.Printf("p: %v\n", p)
		fmt.Printf("rightPoint: %v\n", rightPoint)
	}

	return res
}
