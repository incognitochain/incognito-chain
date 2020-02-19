package aggregaterange

import (
	"errors"
	"github.com/incognitochain/incognito-chain/privacy"
	"math"
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
		if !proof.l[i].PointValid() || !proof.r[i].PointValid() {
			return false
		}
	}

	if !proof.a.ScalarValid() || !proof.b.ScalarValid() {
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

func (wit InnerProductWitness) Prove(aggParam *bulletproofParams) (*InnerProductProof, error) {
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
		G[i] = new(privacy.Point).Set(aggParam.g[i])
		H[i] = new(privacy.Point).Set(aggParam.h[i])
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
		L.Add(L, new(privacy.Point).ScalarMult(aggParam.u, cL))
		proof.l = append(proof.l, L)

		R, err := encodeVectors(a[nPrime:], b[:nPrime], G[:nPrime], H[nPrime:])
		if err != nil {
			return nil, err
		}
		R.Add(R, new(privacy.Point).ScalarMult(aggParam.u, cR))
		proof.r = append(proof.r, R)

		// calculate challenge x = hash(G || H || u || x || l || r)
		x := generateChallenge([][]byte{aggParam.cs, p.ToBytesS(), L.ToBytesS(), R.ToBytesS()})
		//x := generateChallengeOld(aggParam, [][]byte{p.ToBytesS(), L.ToBytesS(), R.ToBytesS()})
		xInverse := new(privacy.Scalar).Invert(x)
		xSquare := new(privacy.Scalar).Mul(x, x)
		xSquareInverse := new(privacy.Scalar).Mul(xInverse, xInverse)

		// calculate GPrime, HPrime, PPrime for the next loop
		GPrime := make([]*privacy.Point, nPrime)
		HPrime := make([]*privacy.Point, nPrime)

		for i := range GPrime {
			GPrime[i] = new(privacy.Point).AddPedersen(xInverse, G[i], x, G[i+nPrime])
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

func (proof InnerProductProof) Verify(aggParam *bulletproofParams) bool {
	//var aggParam = newBulletproofParams(1)
	p := new(privacy.Point)
	p.Set(proof.p)
	n := len(aggParam.g)
	G := make([]*privacy.Point, n)
	H := make([]*privacy.Point, n)
	s := make([]*privacy.Scalar, n)
	sInverse := make([]*privacy.Scalar, n)

	for i := range G {
		G[i] = new(privacy.Point).Set(aggParam.g[i])
		H[i] = new(privacy.Point).Set(aggParam.h[i])
		s[i] = new(privacy.Scalar).FromUint64(1)
		sInverse[i] = new(privacy.Scalar).FromUint64(1)
	}
	logN := int(math.Log2(float64(n)))
	xList := make([]*privacy.Scalar, logN)
	xInverseList := make([]*privacy.Scalar, logN)
	xSquareList := make([]*privacy.Scalar, logN)
	xInverseSquare_List := make([]*privacy.Scalar, logN)

	//a*s ; b*s^-1
	for i := range proof.l {
		// calculate challenge x = hash(hash(G || H || u || p) || x || l || r)
		xList[i] = generateChallenge([][]byte{aggParam.cs, p.ToBytesS(), proof.l[i].ToBytesS(), proof.r[i].ToBytesS()})
		xInverseList[i] = new(privacy.Scalar).Invert(xList[i])
		xSquareList[i] = new(privacy.Scalar).Mul(xList[i], xList[i])
		xInverseSquare_List[i] = new(privacy.Scalar).Mul(xInverseList[i], xInverseList[i])

		//Update s, s^-1
		for j := 0; j < n; j++ {
			if j&int(math.Pow(2, float64(logN-i-1))) != 0 {
				s[j] = new(privacy.Scalar).Mul(s[j], xList[i])
				sInverse[j] = new(privacy.Scalar).Mul(sInverse[j], xInverseList[i])
			} else {
				s[j] = new(privacy.Scalar).Mul(s[j], xInverseList[i])
				sInverse[j] = new(privacy.Scalar).Mul(sInverse[j], xList[i])
			}
		}
		PPrime := new(privacy.Point).AddPedersen(xSquareList[i], proof.l[i], xInverseSquare_List[i], proof.r[i])
		PPrime.Add(PPrime, p)
		p = PPrime
	}

	// Compute (g^s)^a (h^-s)^b u^(ab) = p l^(x^2) r^(-x^2)
	c := new(privacy.Scalar).Mul(proof.a, proof.b)
	rightHSPart1 := new(privacy.Point).MultiScalarMult(s, G)
	rightHSPart1.ScalarMult(rightHSPart1, proof.a)
	rightHSPart2 := new(privacy.Point).MultiScalarMult(sInverse, H)
	rightHSPart2.ScalarMult(rightHSPart2, proof.b)
	rightHS := new(privacy.Point).Add(rightHSPart1, rightHSPart2)
	rightHS.Add(rightHS, new(privacy.Point).ScalarMult(aggParam.u, c))

	leftHSPart1 := new(privacy.Point).MultiScalarMult(xSquareList, proof.l)
	leftHSPart2 := new(privacy.Point).MultiScalarMult(xInverseSquare_List, proof.r)
	leftHS := new(privacy.Point).Add(leftHSPart1, leftHSPart2)
	leftHS.Add(leftHS, proof.p)

	res := privacy.IsPointEqual(rightHS, leftHS)
	if !res {
		privacy.Logger.Log.Error("Inner product argument failed:")
		privacy.Logger.Log.Error("LHS: %v\n", leftHS)
		privacy.Logger.Log.Error("RHS: %v\n", rightHS)
	}

	return res
}

func VerifyBatchingInnerProductProofs(proofs []*InnerProductProof, csList [][]byte) bool {
	batchSize := len(proofs)
	// Generate list of random value
	sum_abAlpha := new(privacy.Scalar).FromUint64(0)
	pList := make([]*privacy.Point, 0)
	alphaList := make([]*privacy.Scalar, 0)
	LList := make([]*privacy.Point, 0)
	nXSquareList := make([]*privacy.Scalar, 0)
	RList := make([]*privacy.Point, 0)
	nXInverseSquareList := make([]*privacy.Scalar, 0)

	maxN := 0
	asAlphaList := make([]*privacy.Scalar, len(AggParam.g))
	bsInverseAlphaList := make([]*privacy.Scalar, len(AggParam.g))
	for k := 0; k < len(AggParam.g); k++ {
		asAlphaList[k] = new(privacy.Scalar).FromUint64(0)
		bsInverseAlphaList[k] = new(privacy.Scalar).FromUint64(0)
	}
	for i := 0; i < batchSize; i++ {
		alpha := privacy.RandomScalar()
		abAlpha := new(privacy.Scalar).Mul(proofs[i].a, proofs[i].b)
		abAlpha.Mul(abAlpha, alpha)
		sum_abAlpha.Add(sum_abAlpha, abAlpha)

		//prod_PAlpha.Add(prod_PAlpha, new(privacy.Point).ScalarMult(proofs[i].p,alpha))
		pList = append(pList, proofs[i].p)
		alphaList = append(alphaList, alpha)

		n := int(math.Pow(2, float64(len(proofs[i].l))))
		if maxN < n {
			maxN = n
		}
		logN := int(math.Log2(float64(n)))
		s := make([]*privacy.Scalar, n)
		sInverse := make([]*privacy.Scalar, n)
		xList := make([]*privacy.Scalar, logN)
		xInverseList := make([]*privacy.Scalar, logN)
		xSquareList := make([]*privacy.Scalar, logN)
		xSquareAlphaList := make([]*privacy.Scalar, logN)
		xInverseSquareList := make([]*privacy.Scalar, logN)
		xInverseSquareAlphaList := make([]*privacy.Scalar, logN)

		for k := 0; k < n; k++ {
			s[k] = new(privacy.Scalar).Mul(alpha, proofs[i].a)
			sInverse[k] = new(privacy.Scalar).Mul(alpha, proofs[i].b)
		}

		p := new(privacy.Point).Set(proofs[i].p)
		for j := 0; j < len(proofs[i].l); j++ {
			// calculate challenge x = hash(hash(G || H || u || p) || x || l || r)
			xList[j] = generateChallenge([][]byte{csList[i], p.ToBytesS(), proofs[i].l[j].ToBytesS(), proofs[i].r[j].ToBytesS()})
			xInverseList[j] = new(privacy.Scalar).Invert(xList[j])
			xSquareList[j] = new(privacy.Scalar).Mul(xList[j], xList[j])
			xSquareAlphaList[j] = new(privacy.Scalar).Mul(xSquareList[j], alpha)
			xInverseSquareList[j] = new(privacy.Scalar).Mul(xInverseList[j], xInverseList[j])
			xInverseSquareAlphaList[j] = new(privacy.Scalar).Mul(xInverseSquareList[j], alpha)

			pPrime := new(privacy.Point).AddPedersen(xSquareList[j], proofs[i].l[j], xInverseSquareList[j], proofs[i].r[j])
			pPrime.Add(pPrime, p)
			p = pPrime

			//Update s, s^-1
			for k := 0; k < n; k++ {
				if k&int(math.Pow(2, float64(logN-j-1))) != 0 {
					s[k] = new(privacy.Scalar).Mul(s[k], xList[j])
					sInverse[k] = new(privacy.Scalar).Mul(sInverse[k], xInverseList[j])
				} else {
					s[k] = new(privacy.Scalar).Mul(s[k], xInverseList[j])
					sInverse[k] = new(privacy.Scalar).Mul(sInverse[k], xList[j])
				}
			}
		}
		for k := 0; k < n; k++ {
			asAlphaList[k].Add(asAlphaList[k], s[k])
			bsInverseAlphaList[k].Add(bsInverseAlphaList[k], sInverse[k])
		}

		LList = append(LList, proofs[i].l...)
		nXSquareList = append(nXSquareList, xSquareAlphaList...)
		RList = append(RList, proofs[i].r...)
		nXInverseSquareList = append(nXInverseSquareList, xInverseSquareAlphaList...)
	}

	gAlphaAS := new(privacy.Point).MultiScalarMult(asAlphaList[0:maxN], AggParam.g[0:maxN])
	hAlphaBSInverse := new(privacy.Point).MultiScalarMult(bsInverseAlphaList[0:maxN], AggParam.h[0:maxN])
	LHS := new(privacy.Point).Add(gAlphaAS, hAlphaBSInverse)
	LHS.Add(LHS, new(privacy.Point).ScalarMult(AggParam.u, sum_abAlpha))
	//fmt.Println("LHS:", LHS )

	prod_PAlpha := new(privacy.Point).MultiScalarMult(alphaList, pList)
	prod_LX := new(privacy.Point).MultiScalarMult(nXSquareList, LList)
	prod_RX := new(privacy.Point).MultiScalarMult(nXInverseSquareList, RList)

	RHS := new(privacy.Point).Add(prod_LX, prod_RX)
	RHS.Add(RHS, prod_PAlpha)
	//fmt.Println("RHS:", RHS)

	res := privacy.IsPointEqual(RHS, LHS)
	if !res {
		privacy.Logger.Log.Error("Inner product argument failed:")
		privacy.Logger.Log.Error("LHS: %v\n", LHS)
		privacy.Logger.Log.Error("RHS: %v\n", RHS)
	}

	return res
}
