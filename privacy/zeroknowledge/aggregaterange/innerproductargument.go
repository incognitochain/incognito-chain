package aggregaterange

import (
	"errors"
	"math/big"
	"sync"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/privacy"
)

type InnerProductWitness struct {
	a []*big.Int
	b []*big.Int

	p *privacy.EllipticPoint
}

type InnerProductProof struct {
	l []*privacy.EllipticPoint
	r []*privacy.EllipticPoint
	a *big.Int
	b *big.Int

	p *privacy.EllipticPoint
}

func (proof InnerProductProof) ValidateSanity() bool {
	if len(proof.l) != len(proof.r) {
		return false
	}

	for i := 0; i < len(proof.l); i++ {
		if !proof.l[i].IsSafe() {
			return false
		}

		if !proof.r[i].IsSafe() {
			return false
		}
	}

	if proof.a.BitLen() > 256 {
		return false
	}
	if proof.b.BitLen() > 256 {
		return false
	}

	return proof.p.IsSafe()
}

func (proof InnerProductProof) Bytes() []byte {
	var res []byte

	res = append(res, byte(len(proof.l)))
	for _, l := range proof.l {
		res = append(res, l.Compress()...)
	}

	for _, r := range proof.r {
		res = append(res, r.Compress()...)
	}

	res = append(res, common.AddPaddingBigInt(proof.a, common.BigIntSize)...)
	res = append(res, common.AddPaddingBigInt(proof.b, common.BigIntSize)...)
	res = append(res, proof.p.Compress()...)

	return res
}

func (proof *InnerProductProof) SetBytes(bytes []byte) error {
	if len(bytes) == 0 {
		return nil
	}

	lenLArray := int(bytes[0])
	offset := 1

	proof.l = make([]*privacy.EllipticPoint, lenLArray)
	for i := 0; i < lenLArray; i++ {
		proof.l[i] = new(privacy.EllipticPoint)
		proof.l[i].Decompress(bytes[offset : offset+privacy.CompressedEllipticPointSize])
		offset += privacy.CompressedEllipticPointSize
	}

	proof.r = make([]*privacy.EllipticPoint, lenLArray)
	for i := 0; i < lenLArray; i++ {
		proof.r[i] = new(privacy.EllipticPoint)
		proof.r[i].Decompress(bytes[offset : offset+privacy.CompressedEllipticPointSize])
		offset += privacy.CompressedEllipticPointSize
	}

	proof.a = new(big.Int).SetBytes(bytes[offset : offset+common.BigIntSize])
	offset += common.BigIntSize

	proof.b = new(big.Int).SetBytes(bytes[offset : offset+common.BigIntSize])
	offset += common.BigIntSize

	proof.p = new(privacy.EllipticPoint)
	proof.p.Decompress(bytes[offset : offset+privacy.CompressedEllipticPointSize])

	return nil
}

func (wit InnerProductWitness) Prove(AggParam *bulletproofParams) (*InnerProductProof, error) {
	//var AggParam = newBulletproofParams(1)
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
	p.Set(wit.p.GetX(), wit.p.GetY())

	G := make([]*privacy.EllipticPoint, n)
	H := make([]*privacy.EllipticPoint, n)
	for i := range G {
		G[i] = new(privacy.EllipticPoint)
		G[i].Set(AggParam.g[i].GetX(), AggParam.g[i].GetY())

		H[i] = new(privacy.EllipticPoint)
		H[i].Set(AggParam.h[i].GetX(), AggParam.h[i].GetY())
	}

	proof := new(InnerProductProof)
	proof.l = make([]*privacy.EllipticPoint, 0)
	proof.r = make([]*privacy.EllipticPoint, 0)
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
		L = L.Add(AggParam.u.ScalarMult(cL))
		proof.l = append(proof.l, L)

		R, err := encodeVectors(a[nPrime:], b[:nPrime], G[:nPrime], H[nPrime:])
		if err != nil {
			return nil, err
		}
		R = R.Add(AggParam.u.ScalarMult(cR))
		proof.r = append(proof.r, R)

		// calculate challenge x = hash(G || H || u || p ||  l || r)
		x := generateChallengeForAggRange(AggParam, [][]byte{p.Compress(), L.Compress(), R.Compress()})
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

		// x^2 * l + P + xInverse^2 * r
		PPrime := L.ScalarMult(xSquare).Add(p).Add(R.ScalarMult(xSquareInverse))

		// calculate aPrime, bPrime
		aPrime := make([]*big.Int, nPrime)
		bPrime := make([]*big.Int, nPrime)

		for i := range aPrime {
			aPrime[i] = new(big.Int).Mul(a[i], x)
			aPrime[i].Add(aPrime[i], new(big.Int).Mul(a[i+nPrime], xInverse))
			aPrime[i].Mod(aPrime[i], privacy.Curve.Params().N)

			bPrime[i] = new(big.Int).Mul(b[i], xInverse)
			bPrime[i].Add(bPrime[i], new(big.Int).Mul(b[i+nPrime], x))
			bPrime[i].Mod(bPrime[i], privacy.Curve.Params().N)
		}

		a = aPrime
		b = bPrime
		p.Set(PPrime.GetX(), PPrime.GetY())
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
	p := new(privacy.EllipticPoint)
	p.Set(proof.p.GetX(), proof.p.GetY())

	n := len(AggParam.g)

	G := make([]*privacy.EllipticPoint, n)
	H := make([]*privacy.EllipticPoint, n)
	for i := range G {
		G[i] = new(privacy.EllipticPoint)
		G[i].Set(AggParam.g[i].GetX(), AggParam.g[i].GetY())

		H[i] = new(privacy.EllipticPoint)
		H[i].Set(AggParam.h[i].GetX(), AggParam.h[i].GetY())
	}

	for i := range proof.l {
		nPrime := n / 2
		// calculate challenge x = hash(G || H || u || p ||  l || r)
		x := generateChallengeForAggRange(AggParam, [][]byte{p.Compress(), proof.l[i].Compress(), proof.r[i].Compress()})
		xInverse := new(big.Int).ModInverse(x, privacy.Curve.Params().N)

		// calculate GPrime, HPrime, PPrime for the next loop
		GPrime := make([]*privacy.EllipticPoint, nPrime)
		HPrime := make([]*privacy.EllipticPoint, nPrime)

		var wg sync.WaitGroup
		wg.Add(len(GPrime) * 2)
		for i := 0; i < len(GPrime); i++ {
			go func(i int, wg *sync.WaitGroup) {
				defer wg.Done()
				GPrime[i] = G[i].ScalarMult(xInverse).Add(G[i+nPrime].ScalarMult(x))
			}(i, &wg)
			go func(i int, wg *sync.WaitGroup) {
				defer wg.Done()
				HPrime[i] = H[i].ScalarMult(x).Add(H[i+nPrime].ScalarMult(xInverse))
			}(i, &wg)
		}
		wg.Wait()

		xSquare := new(big.Int).Mul(x, x)
		xSquareInverse := new(big.Int).ModInverse(xSquare, privacy.Curve.Params().N)

		//PPrime := l.ScalarMul(xSquare).Add(p).Add(r.ScalarMul(xSquareInverse)) // x^2 * l + P + xInverse^2 * r
		var temp1, temp2 *privacy.EllipticPoint
		wg.Add(2)
		go func(wg *sync.WaitGroup) {
			defer wg.Done()
			temp1 = proof.l[i].ScalarMult(xSquare)
		}(&wg)
		go func(wg *sync.WaitGroup) {
			defer wg.Done()
			temp2 = proof.r[i].ScalarMult(xSquareInverse)
		}(&wg)
		wg.Wait()
		PPrime := temp1.Add(p).Add(temp2) // x^2 * l + P + xInverse^2 * r

		p = PPrime
		G = GPrime
		H = HPrime
		n = nPrime
	}

	c := new(big.Int).Mul(proof.a, proof.b)

	rightPoint := G[0].ScalarMult(proof.a)
	rightPoint = rightPoint.Add(H[0].ScalarMult(proof.b))
	rightPoint = rightPoint.Add(AggParam.u.ScalarMult(c))

	res := rightPoint.IsEqual(p)
	if !res {
		privacy.Logger.Log.Error("Inner product argument failed:")
		privacy.Logger.Log.Error("p: %v\n", p)
		privacy.Logger.Log.Error("rightPoint: %v\n", rightPoint)
	}

	return res
}
