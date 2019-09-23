package aggregaterange

import (
	"errors"
	"sync"

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
		res = append(res, l.ToBytes()[:]...)
	}

	for _, r := range proof.r {
		res = append(res, r.ToBytes()[:]...)
	}

	res = append(res, proof.a.ToBytes()[:]...)
	res = append(res, proof.b.ToBytes()[:]...)
	res = append(res, proof.p.ToBytes()[:]...)

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
		proof.l[i], err = new(privacy.Point).FromBytes(privacy.SliceToArray(bytes[offset : offset+privacy.Ed25519KeySize]))
		if err != nil{
			return err
		}
		offset += privacy.Ed25519KeySize
	}

	proof.r = make([]*privacy.Point, lenLArray)
	for i := 0; i < lenLArray; i++ {
		proof.r[i], err = new(privacy.Point).FromBytes(privacy.SliceToArray(bytes[offset : offset+privacy.Ed25519KeySize]))
		if err != nil{
			return err
		}
		offset += privacy.Ed25519KeySize
	}

	proof.a = new(privacy.Scalar).FromBytes(privacy.SliceToArray(bytes[offset : offset+privacy.Ed25519KeySize]))
	offset += privacy.Ed25519KeySize

	proof.b = new(privacy.Scalar).FromBytes(privacy.SliceToArray(bytes[offset : offset+privacy.Ed25519KeySize]))
	offset += privacy.Ed25519KeySize

	proof.p, err = new(privacy.Point).FromBytes(privacy.SliceToArray(bytes[offset : offset+privacy.Ed25519KeySize]))
	if err != nil{
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

		L, err := encodeVectors(a[:nPrime], b[nPrime:], G[nPrime:], H[:nPrime])
		if err != nil {
			return nil, err
		}
		L = L.Add(L, new(privacy.Point).ScalarMult(AggParam.u, cL))
		proof.l = append(proof.l, L)

		R, err := encodeVectors(a[nPrime:], b[:nPrime], G[:nPrime], H[nPrime:])
		if err != nil {
			return nil, err
		}
		R = R.Add(R, new(privacy.Point).ScalarMult(AggParam.u, cR))
		proof.r = append(proof.r, R)

		// calculate challenge x = hash(G || H || u || p ||  l || r)
		x := generateChallengeForAggRange(AggParam, [][]byte{p.ToBytes()[:], L.ToBytes()[:], R.ToBytes()[:]})
		xInverse := new(privacy.Scalar).Invert(x)

		// calculate GPrime, HPrime, PPrime for the next loop
		GPrime := make([]*privacy.Point, nPrime)
		HPrime := make([]*privacy.Point, nPrime)

		for i := range GPrime {
			GPrime[i] = new(privacy.Point).ScalarMult(G[i], xInverse)
			GPrime[i].Add(GPrime[i], new(privacy.Point).ScalarMult(G[i+nPrime], x))

			HPrime[i] = new(privacy.Point).ScalarMult(H[i], x)
			HPrime[i].Add(HPrime[i], new(privacy.Point).ScalarMult(H[i+nPrime], xInverse))
		}

		xSquare := new(privacy.Scalar).Mul(x, x)
		xSquareInverse := new(privacy.Scalar).Invert(xSquare)

		// x^2 * l + P + xInverse^2 * r
		PPrime := new(privacy.Point).ScalarMult(L, xSquare)
		PPrime.Add(PPrime, p)
		PPrime.Add(PPrime, new(privacy.Point).ScalarMult(R, xSquareInverse))

		// calculate aPrime, bPrime
		aPrime := make([]*privacy.Scalar, nPrime)
		bPrime := make([]*privacy.Scalar, nPrime)

		for i := range aPrime {
			aPrime[i] = new(privacy.Scalar).Mul(a[i], x)
			aPrime[i].Add(aPrime[i], new(privacy.Scalar).Mul(a[i+nPrime], xInverse))

			bPrime[i] = new(privacy.Scalar).Mul(b[i], xInverse)
			bPrime[i].Add(bPrime[i], new(privacy.Scalar).Mul(b[i+nPrime], x))
		}

		a = aPrime
		b = bPrime
		p.Set(PPrime)
		G = GPrime
		H = HPrime
		n = nPrime
	}

	proof.a = a[0]
	proof.b = b[0]

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
		x := generateChallengeForAggRange(AggParam, [][]byte{p.ToBytes()[:], proof.l[i].ToBytes()[:], proof.r[i].ToBytes()[:]})
		xInverse := new(privacy.Scalar).Invert(x)

		// calculate GPrime, HPrime, PPrime for the next loop
		GPrime := make([]*privacy.Point, nPrime)
		HPrime := make([]*privacy.Point, nPrime)

		var wg sync.WaitGroup
		wg.Add(len(GPrime) * 2)
		for i := 0; i < len(GPrime); i++ {
			go func(i int, wg *sync.WaitGroup) {
				defer wg.Done()
				GPrime[i] = new(privacy.Point).ScalarMult(G[i], xInverse)
				GPrime[i].Add(GPrime[i], new(privacy.Point).ScalarMult(G[i+nPrime], x))
			}(i, &wg)
			go func(i int, wg *sync.WaitGroup) {
				defer wg.Done()
				HPrime[i] = new(privacy.Point).ScalarMult(H[i], x)
				HPrime[i].Add(HPrime[i], new(privacy.Point).ScalarMult(H[i+nPrime], xInverse))
			}(i, &wg)
		}
		wg.Wait()

		xSquare := new(privacy.Scalar).Mul(x, x)
		xSquareInverse := new(privacy.Scalar).Invert(xSquare)

		//PPrime := l.ScalarMul(xSquare).Add(p).Add(r.ScalarMul(xSquareInverse)) // x^2 * l + P + xInverse^2 * r
		var temp1, temp2 *privacy.Point
		wg.Add(2)
		go func(wg *sync.WaitGroup) {
			defer wg.Done()
			temp1 = new(privacy.Point).ScalarMult(proof.l[i], xSquare)
		}(&wg)
		go func(wg *sync.WaitGroup) {
			defer wg.Done()
			temp2 = new(privacy.Point).ScalarMult(proof.r[i], xSquareInverse)
		}(&wg)
		wg.Wait()
		PPrime := new(privacy.Point).Add(temp1, p)
		PPrime.Add(PPrime, temp2) // x^2 * l + P + xInverse^2 * r

		p = PPrime
		G = GPrime
		H = HPrime
		n = nPrime
	}

	c := new(privacy.Scalar).Mul(proof.a, proof.b)

	rightPoint := new(privacy.Point).ScalarMult(G[0], proof.a)
	rightPoint.Add(rightPoint, new(privacy.Point).ScalarMult(H[0], proof.b))
	rightPoint.Add(rightPoint, new(privacy.Point).ScalarMult(AggParam.u, c))

	res := privacy.IsEqual(rightPoint, p)
	if !res {
		privacy.Logger.Log.Error("Inner product argument failed:")
		privacy.Logger.Log.Error("p: %v\n", p)
		privacy.Logger.Log.Error("rightPoint: %v\n", rightPoint)
	}

	return res
}
