package zkp

// import (
// 	"crypto/rand"
// 	"fmt"
// 	"math/big"
// 	"time"

// 	"github.com/ninjadotorg/constant/privacy-protocol"
// )

// //// PKOneOfManyWitness is a protocol for Zero-knowledge Proof of Knowledge of one out of many commitments containing 0
// //// include Witness: CommitedValue, r []byte
// type PKOneOfManyWitness struct {
// 	commitments []*privacy.EllipticPoint
// 	rand        *big.Int
// 	indexIsZero *int
// 	index       *int
// }

// //// PKOneOfManyProof contains Proof's value
// type PKOneOfManyProof struct {
// 	commitments    []*privacy.EllipticPoint
// 	cl, ca, cb, cd []*privacy.EllipticPoint
// 	f, za, zb      []*big.Int
// 	zd             *big.Int
// }

// // Set sets Witness
// func (wit *PKOneOfManyWitness) Set(
// 	commitments []*privacy.EllipticPoint,
// 	rand *big.Int,
// 	indexIsZero *int) {

// 	wit.commitments = commitments
// 	wit.indexIsZero = indexIsZero
// 	wit.rand = rand

// 	// if wit.commitments == nil {
// 	// 	make([]*privacy.EllipticPoint, len(commitments))
// 	// }

// 	// for i := 0; i < len(commitments); i++ {
// 	// 	if wit.commitments[i].X == nil {
// 	// 		wit.commitments[i].X = big.NewInt(0)
// 	// 	}
// 	// 	*(wit.commitments[i].X) = *(commitments[i].X)

// 	// 	if wit.commitments[i].Y == nil {
// 	// 		wit.commitments[i].Y = big.NewInt(0)
// 	// 	}
// 	// 	*(wit.commitments[i].Y) = *(commitments[i].Y)
// 	// }

// 	// copy(wit.rand, rand)
// 	// wit.indexIsZero = indexIsZero

// }

// // Set sets Proof
// func (pro *PKOneOfManyProof) Set(
// 	commitments []*privacy.EllipticPoint,
// 	cl, ca, cb, cd []*privacy.EllipticPoint,
// 	f, za, zb []*big.Int,
// 	zd *big.Int) {

// 	pro.commitments = commitments
// 	pro.cl, pro.ca, pro.cb, pro.cd = cl, ca, cb, cd
// 	pro.f, pro.za, pro.zb = f, za, zb
// 	pro.zd = zd
// }

// // Prove creates proof for one out of many commitments containing 0
// func (wit *PKOneOfManyWitness) Prove( /*commitments [][]byte, indexIsZero int, commitmentValue []byte, index byte*/ ) (*PKOneOfManyProof, error) {
// 	// proof := new(PKOneOfManyProof)

// 	// Check the number of Commitment list's elements
// 	// N = 2^n
// 	N := len(wit.commitments)
// 	temp := 1
// 	n := 0
// 	for temp < N {
// 		temp = temp << 1
// 		n++
// 	}

// 	if temp != N {
// 		return nil, fmt.Errorf("the number of Commitment list's elements must be power of two")
// 	}

// 	// Check indexIsZero
// 	if *wit.indexIsZero > N || *wit.indexIsZero < 0 {
// 		return nil, fmt.Errorf("Index is zero must be Index in list of commitments")
// 	}

// 	// Check Index
// 	if *wit.index < 1 || *wit.index > 4 {
// 		return nil, fmt.Errorf("Index must be between 1 and 4")
// 	}

// 	// represent indexIsZero in binary
// 	indexIsZeroBinary := privacy.ConvertIntToBinary(*wit.indexIsZero, n)

// 	//
// 	r := make([]*big.Int, n)
// 	a := make([]*big.Int, n)
// 	s := make([]*big.Int, n)
// 	t := make([]*big.Int, n)
// 	u := make([]*big.Int, n)

// 	cl := make([]privacy.EllipticPoint, n)
// 	ca := make([]privacy.EllipticPoint, n)
// 	cb := make([]privacy.EllipticPoint, n)
// 	cd := make([]privacy.EllipticPoint, n)

// 	for j := n - 1; j >= 0; j-- {
// 		// Generate random numbers
// 		r[j], _ = rand.Int(rand.Reader, privacy.Curve.Params().N)
// 		a[j], _ = rand.Int(rand.Reader, privacy.Curve.Params().N)
// 		s[j], _ = rand.Int(rand.Reader, privacy.Curve.Params().N)
// 		t[j], _ = rand.Int(rand.Reader, privacy.Curve.Params().N)
// 		u[j], _ = rand.Int(rand.Reader, privacy.Curve.Params().N)
// 		// r[j] = make([]byte, 32)
// 		// r[j] = privacy.RandBytes(32)
// 		// a[j] = make([]byte, 32)
// 		// a[j] = privacy.RandBytes(32)
// 		// s[j] = make([]byte, 32)
// 		// s[j] = privacy.RandBytes(32)
// 		// t[j] = make([]byte, 32)
// 		// t[j] = privacy.RandBytes(32)
// 		// u[j] = make([]byte, 32)
// 		// u[j] = privacy.RandBytes(32)

// 		// convert indexIsZeroBinary[j] to big.Int
// 		indexInt := big.NewInt(int64(indexIsZeroBinary[j]))

// 		// Calculate cl, ca, cb, cd
// 		// cl = Com(l, r)
// 		cl[j] = privacy.PedCom.CommitAtIndex(indexInt, r[j], index)

// 		// ca = Com(a, s)
// 		ca[j] = privacy.PedCom.CommitAtIndex(a[j], s[j], index)

// 		// cb = Com(la, t)
// 		la := new(big.Int)
// 		la.Mul(indexInt, a[j])
// 		la.Mod(la, privacy.Curve.Params().N)
// 		cb[j] = privacy.PedCom.CommitAtIndex(la, t[j], index)
// 	}

// 	//
// 	// Calculate: cd_k = ci^pi,k
// 	commitPoints := make([]privacy.EllipticPoint, N)
// 	for k := 0; k < n; k++ {
// 		// Calculate pi,k which is coefficient of x^k in polynomial pi(x)
// 		res := privacy.EllipticPoint{X: big.NewInt(0), Y: big.NewInt(0)}
// 		tmp := privacy.EllipticPoint{X: big.NewInt(0), Y: big.NewInt(0)}

// 		var err error

// 		for i := 0; i < N; i++ {
// 			// commitPoints[i] = new(privacy.EllipticPoint)
// 			// commitPoints[i], err = privacy.DecompressCommitment(commitments[i])
// 			//fmt.Printf("i %v k %v %v\n", i, k, *commitPoints[i])
// 			// if err != nil {
// 			// 	return nil, err
// 			// }

// 			iBinary := privacy.ConvertIntToBinary(i, n)
// 			pik := GetCoefficient(iBinary, k, n, a, indexIsZeroBinary)
// 			//fmt.Printf("i %v k %v n %v %v\n", i, k, n, pik)
// 			tmp.X, tmp.Y = privacy.Curve.ScalarMult(wit.commitments[i].X, wit.commitments[i].Y, pik.Bytes())
// 			res.X, res.Y = privacy.Curve.Add(res.X, res.Y, tmp.X, tmp.Y)
// 			// fmt.Printf("i %v k %v %v\n", i, k, tmp)
// 		}

// 		comZero := privacy.PedCom.CommitAtIndex(big.NewInt(0), u[k], index)
// 		// comZeroPoint, err := privacy.DecompressCommitment(comZero)
// 		// if err != nil {
// 		// 	return nil, err
// 		// }
// 		res.X, res.Y = privacy.Curve.Add(res.X, res.Y, comZero.X, comZero.Y)
// 		cd[k] = res
// 	}

// 	// Calculate x
// 	x := big.NewInt(0)

// 	for j := 0; j <= n-1; j++ {
// 		*x = *GenerateChallenge([][]byte{x.Bytes(), cl[j].Compress(), ca[j].Compress(), cb[j].Compress(), cd[j].Compress()})
// 		// x.SetBytes(privacy.Elcm.GetHashOfValues([][]byte{x.Bytes(), cl[j], ca[j], cb[j], cd[j]}))
// 		x.Mod(x, privacy.Curve.Params().N)
// 	}

// 	// Calculate za, zb zd

// 	za := make([]big.Int, n)
// 	zb := make([]big.Int, n)
// 	zd := make(big.Int)
// 	f := make([]big.Int, n)

// 	for j := n - 1; j >= 0; j-- {
// 		// f = lx + a
// 		// fInt := big.NewInt(0)
// 		f[j].Mul(big.NewInt(int64(indexIsZeroBinary[j])), x)
// 		// fInt.Mul(big.NewInt(int64(indexIsZeroBinary[j])), x)
// 		f[j].Add(f[j], a[j])
// 		// fInt.Add(fInt, new(big.Int).SetBytes(a[j]))
// 		f[j].Mod(f[j], privacy.Curve.Params().N)
// 		// fInt.Mod(fInt, privacy.Curve.Params().N)
// 		// f[j] = make([]byte, 32)
// 		// f[j] = fInt.Bytes()

// 		// za = s + rx
// 		za[j].Mul(r[j], x)
// 		// zaInt := big.NewInt(0)
// 		// zaInt.Mul(new(big.Int).SetBytes(r[j]), x)
// 		za[j].Add(za[j], s[j])
// 		// zaInt.Add(zaInt, new(big.Int).SetBytes(s[j]))
// 		// za[j] = make([]byte, 32)
// 		// za[j] = zaInt.Bytes()

// 		// zb = r(x - f) + t
// 		// zbInt := big.NewInt(0)
// 		zb[j].Sub(privacy.Curve.Params().N, f[j])
// 		// zbInt.Sub(privacy.Curve.Params().N, fInt)
// 		zb[j].Add(zb[j], x)
// 		// zbInt.Add(zbInt, x)
// 		zb[j].Mul(zb[j], r[j])
// 		// zbInt.Mul(zbInt, new(big.Int).SetBytes(r[j]))
// 		zb[j].Add(zb[j], t[j])
// 		// zbInt.Add(zbInt, new(big.Int).SetBytes(t[j]))
// 		zb[j].Mod(zb[j], privacy.Curve.Params().N)
// 		// zbInt.Mod(zbInt, privacy.Curve.Params().N)
// 		// zb[j] = make([]byte, 32)
// 		// zb[j] = zbInt.Bytes()

// 	}

// 	// zdInt := big.NewInt(0)
// 	zd.Exp(x, big.NewInt(int64(n)), privacy.Curve.Params().N)
// 	zd.Mul(zd, wit.rand)
// 	// zdInt.Mul(zdInt, new(big.Int).SetBytes(rand))

// 	uxInt := big.NewInt(0)
// 	sumInt := big.NewInt(0)
// 	for k := 0; k < n; k++ {
// 		uxInt.Exp(x, big.NewInt(int64(k)), privacy.Curve.Params().N)
// 		uxInt.Mul(uxInt, u[k])
// 		sumInt.Add(sumInt, uxInt)
// 		sumInt.Mod(sumInt, privacy.Curve.Params().N)
// 	}

// 	sumInt.Sub(privacy.Curve.Params().N, sumInt)

// 	zd.Add(zd, sumInt)
// 	zd.Mod(zd, privacy.Curve.Params().N)
// 	// zd = zdInt.Bytes()
// 	var proof PKOneOfManyProof
// 	proof.Set(&wit.commitments, &cl, &ca, &cb, &cd, &f, &za, &zb, &zd)
// 	return &proof, nil
// }

// func (pro *PKOneOfManyProof) Verify( /*commitments [][]byte, proof *PKOneOfManyProof, index byte, rand []byte*/ ) bool {
// 	N := len(pro.commitments)

// 	temp := 1
// 	n := 0
// 	for temp < N {
// 		temp = temp << 1
// 		n++
// 	}
// 	// clPoint := make([]*privacy.EllipticPoint, n)
// 	// caPoint := make([]*privacy.EllipticPoint, n)
// 	// cbPoint := make([]*privacy.EllipticPoint, n)
// 	// cdPoint := make([]*privacy.EllipticPoint, n)
// 	var err error

// 	// Calculate x
// 	// x := big.NewInt(0)
// 	// for j := 0; j <= n-1; j++ {
// 	// 	x.SetBytes(privacy.Elcm.GetHashOfValues([][]byte{x.Bytes(), proof.cl[j], proof.ca[j], proof.cb[j], proof.cd[j]}))
// 	// 	x.Mod(x, privacy.Curve.Params().N)
// 	// }

// 	for j := 0; j <= n-1; j++ {
// 		*x = *GenerateChallenge([][]byte{x.Bytes(), cl[j].Compress(), ca[j].Compress(), cb[j].Compress(), cd[j].Compress()})
// 		// x.SetBytes(privacy.Elcm.GetHashOfValues([][]byte{x.Bytes(), cl[j], ca[j], cb[j], cd[j]}))
// 		x.Mod(x, privacy.Curve.Params().N)
// 	}

// 	//fmt.Printf("x Verify: %v\n", x)

// 	for i := 0; i < n; i++ {
// 		// Decompress cl from bytes array to Elliptic
// 		// clPoint[i] = new(privacy.EllipticPoint)
// 		// clPoint[i], err = privacy.DecompressCommitment(proof.cl[i])
// 		// if err != nil {
// 		// 	return false
// 		// }
// 		// Decompress ca from bytes array to Elliptic
// 		// caPoint[i] = new(privacy.EllipticPoint)
// 		// caPoint[i], err = privacy.DecompressCommitment(proof.ca[i])
// 		// if err != nil {
// 		// 	return false
// 		// }
// 		// Decompress cb from bytes array to Elliptic
// 		// cbPoint[i] = new(privacy.EllipticPoint)
// 		// cbPoint[i], err = privacy.DecompressCommitment(proof.cb[i])
// 		// if err != nil {
// 		// 	return false
// 		// }

// 		// // Decompress cd from bytes array to Elliptic
// 		// cdPoint[i] = new(privacy.EllipticPoint)
// 		// cdPoint[i], err = privacy.DecompressKey(proof.cd[i])
// 		// if err != nil {
// 		// 	return false
// 		// }

// 		// Check cl^x * ca = Com(f, za)
// 		leftPoint1 := new(privacy.EllipticPoint)
// 		leftPoint1.X, leftPoint1.Y = privacy.Curve.ScalarMult(pro.cl[i].X, pro.cl[i].Y, x.Bytes())
// 		leftPoint1.X, leftPoint1.Y = privacy.Curve.Add(leftPoint1.X, leftPoint1.Y, pro.ca[i].X, pro.ca[i].Y)

// 		rightPoint1 := privacy.PedCom.CommitAtIndex(pro.f[i], pro.za[i], index)
// 		//fmt.Printf("Left point 1 X: %v\n", leftPoint1.X)
// 		//fmt.Printf("Right point 1 X: %v\n", rightPoint1.X)
// 		//fmt.Printf("Left point 1 Y: %v\n", leftPoint1.Y)
// 		//fmt.Printf("Right point 1 Y: %v\n", rightPoint1.Y)

// 		if leftPoint1.X.Cmp(rightPoint1.X) != 0 || leftPoint1.Y.Cmp(rightPoint1.Y) != 0 {
// 			return false
// 		}

// 		// Check cl^(x-f) * cb = Com(0, zb)
// 		leftPoint2 := new(privacy.EllipticPoint)
// 		xSubF := new(big.Int)
// 		// tmp := new(big.Int).SetBytes(proof.f[i])
// 		//fmt.Printf("tmp: %v\n", tmp)
// 		xSubF.Sub(x, pro.f[i])
// 		xSubF.Mod(xSubF, privacy.Curve.Params().N)
// 		leftPoint2.X, leftPoint2.Y = privacy.Curve.ScalarMult(pro.cl[i].X, pro.cl[i].Y, xSubF.Bytes())
// 		leftPoint2.X, leftPoint2.Y = privacy.Curve.Add(leftPoint2.X, leftPoint2.Y, pro.cb[i].X, pro.cb[i].Y)

// 		// rightPoint2 := new(privacy.EllipticPoint)
// 		rightPoint2 := privacy.Elcm.CommitSpecValue(big.NewInt(0), pro.zb[i], index)
// 		// rightPoint2, err = privacy.DecompressCommitment(right2)
// 		// if err != nil {
// 		// 	return false
// 		// }

// 		//fmt.Printf("Left point 2 X: %v\n", leftPoint2.X)
// 		//fmt.Printf("Right point 2 X: %v\n", rightPoint2.X)
// 		//fmt.Printf("Left point 2 Y: %v\n", leftPoint2.Y)
// 		//fmt.Printf("Right point 2 Y: %v\n", rightPoint2.Y)

// 		if leftPoint2.X.Cmp(rightPoint2.X) != 0 || leftPoint2.Y.Cmp(rightPoint2.Y) != 0 {
// 			return false
// 		}
// 	}

// 	// commitPoints := make([]*privacy.EllipticPoint, N)
// 	leftPoint3 := privacy.EllipticPoint{X: big.NewInt(0), Y: big.NewInt(0)}
// 	leftPoint32 := privacy.EllipticPoint{X: big.NewInt(0), Y: big.NewInt(0)}
// 	// rightPoint3 := new(privacy.EllipticPoint)
// 	tmpPoint := new(privacy.EllipticPoint)

// 	for i := 0; i < N; i++ {
// 		iBinary := privacy.ConvertIntToBinary(i, n)
// 		// commitPoints[i] = new(privacy.EllipticPoint)
// 		// commitPoints[i], err = privacy.DecompressCommitment(pro.commitments[i])
// 		// if err != nil {
// 		// 	return false
// 		// }

// 		exp := big.NewInt(1)
// 		fji := big.NewInt(1)
// 		for j := n - 1; j >= 0; j-- {
// 			if iBinary[j] == 1 {
// 				*fji = *pro.f[j]
// 			} else {
// 				fji.Sub(x, pro.f[j])
// 				fji.Mod(fji, privacy.Curve.Params().N)
// 			}

// 			exp.Mul(exp, fji)
// 			exp.Mod(exp, privacy.Curve.Params().N)
// 		}

// 		tmpPoint.X, tmpPoint.Y = privacy.Curve.ScalarMult(pro.commitments[i].X, pro.commitments[i].Y, exp.Bytes())
// 		leftPoint3.X, leftPoint3.Y = privacy.Curve.Add(leftPoint3.X, leftPoint3.Y, tmpPoint.X, tmpPoint.Y)
// 	}

// 	for k := 0; k < n; k++ {
// 		xk := big.NewInt(0)
// 		xk.Exp(x, big.NewInt(int64(k)), privacy.Curve.Params().N)

// 		xk.Sub(privacy.Curve.Params().N, xk)

// 		tmpPoint.X, tmpPoint.Y = privacy.Curve.ScalarMult(pro.cd[k].X, pro.cd[k].Y, xk.Bytes())
// 		leftPoint32.X, leftPoint32.Y = privacy.Curve.Add(leftPoint32.X, leftPoint32.Y, tmpPoint.X, tmpPoint.Y)
// 	}

// 	leftPoint3.X, leftPoint3.Y = privacy.Curve.Add(leftPoint3.X, leftPoint3.Y, leftPoint32.X, leftPoint32.Y)

// 	rightPoint3 := privacy.PedCom.CommitAtIndex(big.NewInt(0), pro.zd, index)
// 	// rightPoint3, _ = privacy.DecompressCommitment(rightValue3)

// 	// fmt.Printf("Left point 3 X: %v\n", leftPoint3.X)
// 	// fmt.Printf("Right point 3 X: %v\n", rightPoint3.X)
// 	// fmt.Printf("Left point 3 Y: %v\n", leftPoint3.Y)
// 	// fmt.Printf("Right point 3 Y: %v\n", rightPoint3.Y)
// 	if leftPoint3.X.Cmp(rightPoint3.X) != 0 || leftPoint3.Y.Cmp(rightPoint3.Y) != 0 {
// 		return false
// 	}

// 	return true
// }

// //TestPKOneOfMany test protocol for one of many Commitment is Commitment to zero
// func TestPKOneOfMany() bool {
// 	// privacy.Elcm.InitCommitment()
// 	pk := new(PKOneOfManyProtocol)

// 	indexIsZero := 23

// 	// list of commitments
// 	commitments := make([][]byte, 32)
// 	serialNumbers := make([][]byte, 32)
// 	randoms := make([][]byte, 32)

// 	for i := 0; i < 32; i++ {
// 		serialNumbers[i] = privacy.RandBytes(32)
// 		randoms[i] = privacy.RandBytes(32)
// 		commitments[i] = make([]byte, 34)
// 		commitments[i] = privacy.Elcm.CommitSpecValue(serialNumbers[i], randoms[i], privacy.SN_CM)
// 	}
// 	// fmt.Printf("%v\n", commitments[indexIsZero])
// 	// fmt.Printf("%v\n", randoms[indexIsZero])
// 	// create Commitment to zero at indexIsZero
// 	serialNumbers[indexIsZero] = big.NewInt(0).Bytes()
// 	commitments[indexIsZero] = privacy.Elcm.CommitSpecValue(serialNumbers[indexIsZero], randoms[indexIsZero], privacy.SN_CM)
// 	// fmt.Printf("%v\n", commitments[indexIsZero])
// 	// fmt.Printf("%v\n", randoms[indexIsZero])
// 	start := time.Now()
// 	proof, err := pk.Prove(commitments, indexIsZero, commitments[indexIsZero], randoms[indexIsZero], privacy.SN_CM)
// 	if err != nil {
// 		fmt.Println(err)
// 	}

// 	res := pk.Verify(commitments, proof, privacy.SN_CM, randoms[indexIsZero])
// 	end := time.Now()
// 	fmt.Printf("%v_+_\n", end.Sub(start))
// 	//fmt.Println(res)
// 	return res
// }

// // Get coefficient of x^k in polynomial pi(x)
// func GetCoefficient(iBinary []byte, k int, n int, a []big.Int, l []byte) *big.Int {
// 	res := privacy.Poly{big.NewInt(1)}
// 	var fji privacy.Poly

// 	for j := n - 1; j >= 0; j-- {
// 		// fj := privacy.Poly{new(big.Int).SetBytes(a[j]), big.NewInt(int64(l[j]))}
// 		fj := privacy.Poly{a[j], big.NewInt(int64(l[j]))}
// 		if iBinary[j] == 0 {
// 			fji = privacy.Poly{big.NewInt(0), big.NewInt(1)}.Sub(fj, privacy.Curve.Params().N)
// 		} else {
// 			fji = fj
// 		}
// 		res = res.Mul(fji, privacy.Curve.Params().N)
// 	}

// 	if res.GetDegree() < k {
// 		return big.NewInt(0)
// 	}
// 	return res[k]
// }
