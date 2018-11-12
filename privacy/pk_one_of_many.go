package privacy

import (
	"fmt"
	"math/big"
)

// PKOneOfManyProtocol is a protocol for Zero-knowledge Proof of Knowledge of one out of many commitments containing 0
// include witnesses: commitedValue, r []byte
type PKOneOfManyProtocol struct {
	witnesses [][]byte
}

// PKOneOfManyProof contains proof's value
type PKOneOfManyProof struct {
	cl, ca, cb, cd [][]byte
	f, za, zb [][]byte
	zd []byte

}

// SetWitness sets witnesses
func (pro *PKOneOfManyProtocol) SetWitness(witnesses [][]byte) {
	pro.witnesses = make([][]byte, len(witnesses))
	for i := 0; i < len(witnesses); i++ {
		pro.witnesses[i] = make([]byte, len(witnesses[i]))
		copy(pro.witnesses[i], witnesses[i])
	}
}

// Prove creates proof for one out of many commitments containing 0
func (pro *PKOneOfManyProtocol) Prove(commitments [][]byte, indexIsZero int, rand []byte, commitmentValue []byte, index byte) (*PKOneOfManyProof, error) {
	n := len(commitments)
	proof := new(PKOneOfManyProof)
	// Check the number of commitment list's elements
	if !IsPowerOfTwo(n) {
		return nil, fmt.Errorf("the number of commitment list's elements must be power of two")
	}

	// Check indexIsZero
	if indexIsZero > n || index < 0 {
		return nil, fmt.Errorf("index is zero must be index in list of commitments")
	}

	// Check index
	if index < 0 || index > 2 {
		return nil, fmt.Errorf("index must be between 0 and 2")
	}

	// represent indexIsZero in binary
	indexIsZeroBinary := ConvertIntToBinany(indexIsZero)

	//
	r := make([][]byte, n+1)
	a := make([][]byte, n+1)
	s := make([][]byte, n+1)
	t := make([][]byte, n+1)
	u := make([][]byte, n)

	proof.cl = make([][]byte, n+1)
	proof.ca = make([][]byte, n+1)
	proof.cb = make([][]byte, n+1)
	proof.cd = make([][]byte, n)

	for j := 1; j <= n; j++ {
		// Generate random numbers
		r[j] = make([]byte, 32)
		r[j] = RandBytes(32)
		a[j] = make([]byte, 32)
		a[j] = RandBytes(32)
		s[j] = make([]byte, 32)
		s[j] = RandBytes(32)
		t[j] = make([]byte, 32)
		t[j] = RandBytes(32)
		u[j-1] = make([]byte, 32)
		u[j-1] = RandBytes(32)

		// convert indexIsZeroBinary[j] to big.Int
		indexInt := big.NewInt(int64(indexIsZeroBinary[j-1]))

		// Calculate cl, ca, cb, cd
		// cl = Com(l, r)
		proof.cl[j] = make([]byte, 34)
		proof.cl[j] = Pcm.CommitSpecValue(indexInt.Bytes(), r[j], index)

		// ca = Com(a, s)
		proof.ca[j] = make([]byte, 34)
		proof.ca[j] = Pcm.CommitSpecValue(a[j], s[j], index)

		// cb = Com(la, t)
		la := new(big.Int)
		la.Mul(indexInt, new(big.Int).SetBytes(a[j]))
		proof.cb[j] = make([]byte, 34)
		proof.cb[j] = Pcm.CommitSpecValue(la.Bytes(), t[j], index)

	}

	//nInt := big.NewInt(int64(n))

	// cd_k =
	// Calculate: ci^pi,k
	commitPoints := make([]*EllipticPoint, n)
	for k := 0; k < n; k++{
		// Calculate pi,k which is coefficient of x^k in polynomial pi(x)
		res := EllipticPoint{X: big.NewInt(0), Y: big.NewInt(0)}
		tmp := EllipticPoint{X: big.NewInt(0), Y: big.NewInt(0)}
		//tmp2 := big.NewInt(0)
		for i:=0; i<n; i++{
			commitPoints[i] = new(EllipticPoint)
			commitPoints[i], _ = DecompressCommitment(commitments[i])
			//if err != nil{
			//	return nil, err
			//}

			// represent i in binary
			iBinary := ConvertIntToBinany(i)
			pik := GetCoefficient(iBinary, k, n, a, indexIsZeroBinary)
			//pik := make([]byte, 32)
			tmp.X, tmp.Y = Curve.ScalarMult(commitPoints[i].X, commitPoints[i].Y, pik.Bytes())
			res.X, res.Y = Curve.Add(res.X, res.Y, tmp.X, tmp.Y)
		}
		comZero := Pcm.CommitSpecValue(big.NewInt(0).Bytes(), u[k], index)
		comZeroPoint, err := DecompressCommitment(comZero)
		if err != nil{
			return nil, err
		}

		res.X, res.Y = Curve.Add(res.X, res.Y, comZeroPoint.X, comZeroPoint.Y)
		cd := CompressKey(res)
		proof.cd[k] = make([]byte, 33)
		copy(proof.cd[k], cd)
	}


	// Calculate x
	x := big.NewInt(0)
	for j:=1; j<=n; j++{
		x.SetBytes(Pcm.getHashOfValues([][]byte{x.Bytes(), proof.cl[j], proof.ca[j], proof.cb[j], proof.cd[j-1]}))
	}

	fmt.Printf("x Prove: %v\n", x)
	//x.Mod(x, Curve.Params().N)

	// Calculate za, zb zd
	//res := Poly{big.NewInt(1)}
	//var fji Poly
	proof.f = make([][]byte, n+1)
	proof.za = make([][]byte, n+1)
	proof.zb = make([][]byte, n+1)
	proof.zd = make([]byte, 32)

	for j:=1; j<=n; j++{
		// f = lx + a
		fInt:=  big.NewInt(0)
		fInt.Mul(big.NewInt(int64(indexIsZeroBinary[j-1])), x)
		fInt.Add(fInt, new(big.Int).SetBytes(a[j]))
		proof.f[j] = make([]byte, 32)
		proof.f[j] = fInt.Bytes()
		//copy(proof.f[j], fInt.Bytes())

		// za = s + rx
		zaInt := big.NewInt(0)
		zaInt.Mul(new(big.Int).SetBytes(r[j]), x)
		zaInt.Add(zaInt, new(big.Int).SetBytes(s[j]))
		proof.za[j] = make([]byte, 32)
		proof.za[j] = zaInt.Bytes()
		//copy(proof.za[j], zaInt.Bytes())

		// zb = r(x - f) + t
		zbInt := big.NewInt(0)
		zbInt.Sub(x, fInt)
		zbInt.Mul(zbInt, new(big.Int).SetBytes(r[j]))
		zbInt.Add(zbInt, new(big.Int).SetBytes(t[j]))
		proof.zb[j] = make([]byte, 32)
		proof.zb[j] = zbInt.Bytes()
		//copy(proof.zb[j], zbInt.Bytes())

	}

	zdInt := big.NewInt(0)
	zdInt.Exp(x, big.NewInt(int64(n)), nil )
	zdInt.Mul(zdInt, new(big.Int).SetBytes(rand))

	uxInt := big.NewInt(0)
	sumInt := big.NewInt(0)
	for k:=0; k<n; k++{
		uxInt.Exp(x, big.NewInt(int64(k)), nil )
		uxInt.Mul(uxInt, new(big.Int).SetBytes(u[k]))
		sumInt.Add(sumInt, uxInt)
	}

	zdInt.Sub(zdInt, sumInt)
	proof.zd = zdInt.Bytes()
	//copy(proof.zd, zdInt.Bytes())

	return proof, nil
}


func (pro *PKOneOfManyProtocol) Verify(commitments [][]byte, proof *PKOneOfManyProof, index byte) bool{
	n := len(commitments)
	clPoint := make([]*EllipticPoint, n+1)
	caPoint := make([]*EllipticPoint, n+1)
	cbPoint := make([]*EllipticPoint, n+1)
	cdPoint := make([]*EllipticPoint, n)
	var err error

	// Calculate x
	x := big.NewInt(0)
	for j:=1; j<=n; j++{
		x.SetBytes(Pcm.getHashOfValues([][]byte{x.Bytes(), proof.cl[j], proof.ca[j], proof.cb[j], proof.cd[j-1]}))
	}
	fmt.Printf("x Verify: %v\n", x)

	for i:=1; i<=n; i++{
		// Decompress cl from bytes array to Elliptic
		clPoint[i] = new(EllipticPoint)
		clPoint[i], err = DecompressCommitment(proof.cl[i])
		if err != nil {
			return false
		}
		// Decompress ca from bytes array to Elliptic
		caPoint[i] = new(EllipticPoint)
		caPoint[i], err = DecompressCommitment(proof.ca[i])
		if err != nil {
			return false
		}
		// Decompress cb from bytes array to Elliptic
		cbPoint[i] = new(EllipticPoint)
		cbPoint[i], err = DecompressCommitment(proof.cb[i])
		if err != nil {
			return false
		}

		// Decompress cd from bytes array to Elliptic
		cdPoint[i-1] = new(EllipticPoint)
		cdPoint[i-1], err = DecompressKey(proof.cd[i-1])
		if err != nil {
			return false
		}


		// Check cl^x * ca = Com(f, za)
		leftPoint1 := new(EllipticPoint)
		leftPoint1.X, leftPoint1.Y = Curve.ScalarMult(clPoint[i].X, clPoint[i].Y, x.Bytes())
		leftPoint1.X, leftPoint1.Y = Curve.Add(leftPoint1.X, leftPoint1.Y, caPoint[i].X, caPoint[i].Y)

		rightPoint1 := new(EllipticPoint)
		right1 := Pcm.CommitSpecValue(proof.f[i], proof.za[i], index)
		rightPoint1, err = DecompressCommitment(right1)
		if err != nil {
			return false
		}

		fmt.Printf("Left point 1 X: %v\n", leftPoint1.X)
		fmt.Printf("Right point 1 X: %v\n", rightPoint1.X)
		fmt.Printf("Left point 1 Y: %v\n", leftPoint1.Y)
		fmt.Printf("Right point 1 Y: %v\n", rightPoint1.Y)

		if leftPoint1.X.Cmp(rightPoint1.X) != 0 || leftPoint1.Y.Cmp(rightPoint1.Y) != 0 {
			return false
		}

		// Check cl^(x-f) * cb = Com(0, zb)
		leftPoint2 := new(EllipticPoint)
		xSubF := new(big.Int)
		tmp := new(big.Int).SetBytes(proof.f[i])
		fmt.Printf("tmp: %v\n", tmp)
		xSubF.Sub(x, tmp)
		xSubF.Mod(xSubF, Curve.Params().N)
		leftPoint2.X, leftPoint2.Y = Curve.ScalarMult(clPoint[i].X, clPoint[i].Y, xSubF.Bytes())
		leftPoint2.X, leftPoint2.Y = Curve.Add(leftPoint2.X, leftPoint2.Y, cbPoint[i].X, cbPoint[i].Y)

		rightPoint2 := new(EllipticPoint)
		right2 := Pcm.CommitSpecValue(big.NewInt(0).Bytes(), proof.zb[i], index)
		rightPoint2, err = DecompressCommitment(right2)
		if err != nil {
			return false
		}

		fmt.Printf("Left point 2 X: %v\n", leftPoint2.X)
		fmt.Printf("Right point 2 X: %v\n", rightPoint2.X)
		fmt.Printf("Left point 2 Y: %v\n", leftPoint2.Y)
		fmt.Printf("Right point 2 Y: %v\n", rightPoint2.Y)

		//if leftPoint1.X.Cmp(Curve.Params().P) == 1 {
		//	fmt.Printf("Wrong!")
		//}

		if leftPoint2.X.Cmp(rightPoint2.X) != 0 || leftPoint2.Y.Cmp(rightPoint2.Y) != 0 {
			return false
		}

	}

	commitPoints := make([]*EllipticPoint, n)
	leftPoint3 := EllipticPoint{X: big.NewInt(0), Y: big.NewInt(0)}
	leftPoint32 := EllipticPoint{X: big.NewInt(0), Y: big.NewInt(0)}
	rightPoint3 := new(EllipticPoint)
	tmpPoint := new(EllipticPoint)

	for i :=0; i< n; i++{
		iBinary := ConvertIntToBinany(i)
		commitPoints[i] = new(EllipticPoint)
		commitPoints[i], err = DecompressCommitment(commitments[i])
		if err != nil {
			return false
		}

		exp := big.NewInt(1)
		fji := big.NewInt(1)
		for j := 1; j<=n; j++{
			if iBinary[j-1] == 0{
				fji.SetBytes(proof.f[j])
			} else{
				fji.Sub(x, new(big.Int).SetBytes(proof.f[j]))
			}
			exp.Mul(exp, fji)
		}
		tmpPoint.X, tmpPoint.Y = Curve.ScalarMult(commitPoints[i].X, commitPoints[i].Y, exp.Bytes())
		leftPoint3.X, leftPoint3.Y = Curve.Add(leftPoint3.X, leftPoint3.Y, tmpPoint.X, tmpPoint.Y)
	}

	for k:=0; k < n; k++{
		xk := big.NewInt(0)
		xk.Exp(x, big.NewInt(int64(k)), nil)
		xk.Sub(big.NewInt(0), xk)

		tmpPoint.X, tmpPoint.Y = Curve.ScalarMult(cdPoint[k].X, cdPoint[k].Y, xk.Bytes())
		leftPoint32.X, leftPoint32.Y = Curve.Add(leftPoint32.X, leftPoint32.Y, tmpPoint.X, tmpPoint.Y)
	}

	leftPoint3.X, leftPoint3.Y = Curve.Add(leftPoint3.X, leftPoint3.Y, leftPoint32.X, leftPoint32.Y )

	right3 := Pcm.CommitSpecValue(big.NewInt(0).Bytes(), proof.zd, index)
	rightPoint3, err = DecompressCommitment(right3)


	fmt.Printf("Left point 3 X: %v\n", leftPoint3.X)
	fmt.Printf("Right point 3 X: %v\n", rightPoint3.X)
	fmt.Printf("Left point 3 Y: %v\n", leftPoint3.Y)
	fmt.Printf("Right point 3 Y: %v\n", rightPoint3.Y)
	if leftPoint3.X.Cmp(rightPoint3.X) != 0 || leftPoint3.Y.Cmp(rightPoint3.Y) != 0 {
		return false
	}
	return true
}


//TestPKOneOfMany test protocol for one of many commitment is commitment to zero
func TestPKOneOfMany() {
	Pcm.InitCommitment()
	pk := new(PKOneOfManyProtocol)

	indexIsZero := 23

	// list of commitments
	commitments := make([][]byte, 32)
	serialNumbers := make([][]byte, 32)
	randoms := make([][]byte, 32)

	for i := 0; i < 32; i++ {
		serialNumbers[i] = RandBytes(32)
		randoms[i] = RandBytes(32)
		commitments[i] = make([]byte, 34)
		commitments[i] = Pcm.CommitSpecValue(serialNumbers[i], randoms[i], SN_CM)
	}

	// create commitment to zero at indexIsZero

	serialNumbers[indexIsZero] = big.NewInt(0).Bytes()
	commitments[indexIsZero] = Pcm.CommitSpecValue(serialNumbers[indexIsZero], randoms[indexIsZero], SN_CM)
	proof, err := pk.Prove(commitments, indexIsZero, commitments[indexIsZero], randoms[indexIsZero], SN_CM)

	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("proof.cl: %+v\n", proof.cl)
	fmt.Printf("proof.ca: %+v\n", proof.ca)
	fmt.Printf("proof.cb: %+v\n", proof.cb)
	fmt.Printf("proof.cd: %+v\n", proof.cd)

	res := pk.Verify(commitments, proof, SN_CM)
	fmt.Println(res)
}

// Get coefficient of x^k in polynomial pi(x)
func GetCoefficient(iBinary []byte , k int, n int, a [][]byte, l []byte ) *big.Int{
	fmt.Printf("i binary: %v\n", iBinary)
	res := Poly{big.NewInt(1)}
	var fji Poly
	for j:=1; j<=n; j++{
		fj := Poly{new(big.Int).SetBytes(a[j]), big.NewInt(int64(l[j-1])) }
		fmt.Printf("Poly fj : %v\n", fj.String())
		if iBinary[j-1] == 0 {
			fji = Poly{big.NewInt(0), big.NewInt(1)}.Sub(fj, nil)
			fmt.Printf("Poly fji : %v\n", fji.String())
		} else{
			fji = fj
		}
		res = res.Mul(fji, Curve.Params().N)
	}
	if res.GetDegree() < k+1 {
		return big.NewInt(0)
	}
	fmt.Printf("res: %v", res[k+1])
	return res[k+1]
}
