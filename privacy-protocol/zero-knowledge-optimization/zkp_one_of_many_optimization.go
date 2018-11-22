package zkpoptimization

import (
	"fmt"
	"math/big"
	"math/bits"
	"sync"
	"time"

	"github.com/ninjadotorg/constant/privacy-protocol"
)

// PKOneOfManyProtocol is a protocol for Zero-knowledge Proof of Knowledge of one out of many commitments containing 0
// include Witness: CommitedValue, r []byte
type PKOneOfManyProtocol struct {
	witnesses [][]byte
}

// PKOneOfManyProof contains Proof's value
type PKOneOfManyProof struct {
	cl, ca, cb, cd [][]byte
	f, za, zb      [][]byte
	zd             []byte
	//val1 EllipticPoint
}

// SetWitness sets Witness
func (pro *PKOneOfManyProtocol) SetWitness(witnesses [][]byte) {
	pro.witnesses = make([][]byte, len(witnesses))
	var wg sync.WaitGroup
	for i := 0; i < len(witnesses); i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			pro.witnesses[i] = make([]byte, len(witnesses[i]))
			copy(pro.witnesses[i], witnesses[i])
		}(i)
	}
	wg.Wait()
}

//res is variable which contains result's value in each function
var res privacy.EllipticPoint

//wg is WaitGroup using in each function
var wg sync.WaitGroup

//addRes: the function, which is protected by mutex, inc res Point by 1 point
func addRes(wgaddRes *sync.WaitGroup, m *sync.Mutex, tmp privacy.EllipticPoint, i int, k int) {
	defer wgaddRes.Done()
	m.Lock()
	// fmt.Println(res.X, " ", i, " ", k)
	res.X, res.Y = privacy.Curve.Add(res.X, res.Y, tmp.X, tmp.Y)
	m.Unlock()
}

// Prove creates proof for one out of many commitments containing 0
func (pro *PKOneOfManyProtocol) Prove(commitments [][]byte, indexIsZero int, commitmentValue []byte, rand []byte, index byte) (*PKOneOfManyProof, error) {

	N := len(commitments)
	proof := new(PKOneOfManyProof)
	// Check the number of PedersenCommitment list's elements
	// st := time.Now()
	// temp := 1
	// n := 0
	// for (temp < N) && (n <= 16) {
	// 	temp = temp << 1
	// 	n++
	// }
	// if temp != N {
	// 	return nil, fmt.Errorf("the number of PedersenCommitment list's elements must be power of two and less than 2^16")
	// }
	// end := time.Now()
	// fmt.Println(end.Sub(st))

	// st = time.Now()
	n := bits.TrailingZeros(uint(N))
	if (n > 16) || (n+bits.LeadingZeros(uint(N)) != 63) {
		return nil, fmt.Errorf("the number of PedersenCommitment list's elements must be power of two and less than 2^16")
	}
	// end = time.Now()
	// fmt.Println(end.Sub(st))

	// fmt.Println(n1 - n)

	// Check indexIsZero
	if indexIsZero > N || indexIsZero < 0 {
		return nil, fmt.Errorf("Index is zero must be Index in list of commitments")
	}

	// Check Index
	if index < 1 || index > 4 {
		return nil, fmt.Errorf("Index must be between 1 and 4")
	}

	// represent indexIsZero in binary
	//indexIsZeroBinary := privacy.ConvertIntToBinary(indexIsZero, n)

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

	temp := indexIsZero
	for j := n; j >= 1; j-- {
		indexInt := big.NewInt(int64(temp & 1))
		temp = temp >> 1
		wg.Add(1)
		go func(j int, indexInt *big.Int) {
			// Generate random numbers
			defer wg.Done()
			r[j] = make([]byte, 32)
			r[j] = privacy.RandBytes(32)
			a[j] = make([]byte, 32)
			a[j] = privacy.RandBytes(32)
			s[j] = make([]byte, 32)
			s[j] = privacy.RandBytes(32)
			t[j] = make([]byte, 32)
			t[j] = privacy.RandBytes(32)
			u[j-1] = make([]byte, 32)
			u[j-1] = privacy.RandBytes(32)

			// convert indexIsZeroBinary[j] to big.Int
			//indexInt := big.NewInt(int64(indexIsZeroBinary[j-1]))

			// Calculate cl, ca, cb, cd
			// cl = Com(l, r)
			proof.cl[j] = make([]byte, 34)
			proof.cl[j] = privacy.Pcm.CommitAtIndex(indexInt.Bytes(), r[j], index)

			// ca = Com(a, s)
			proof.ca[j] = make([]byte, 34)
			proof.ca[j] = privacy.Pcm.CommitAtIndex(a[j], s[j], index)

			// cb = Com(la, t)
			la := new(big.Int)
			la.Mul(indexInt, new(big.Int).SetBytes(a[j]))
			la.Mod(la, privacy.Curve.Params().N)
			proof.cb[j] = make([]byte, 34)
			proof.cb[j] = privacy.Pcm.CommitAtIndex(la.Bytes(), t[j], index)
		}(j, indexInt)
	}

	wg.Wait()

	//

	//wgchild1 is WaitGroup using in child loop
	var wgchild1 sync.WaitGroup
	var wgchild2 sync.WaitGroup
	var mutex sync.Mutex

	commitPoints := make([]*privacy.EllipticPoint, N)
	// Calculate: cd_k = ci^pi,k
	for k := 0; k < n; k++ {
		// Calculate pi,k which is coefficient of x^k in polynomial pi(x)
		res = privacy.EllipticPoint{X: big.NewInt(0), Y: big.NewInt(0)}
		tmp := privacy.EllipticPoint{X: big.NewInt(0), Y: big.NewInt(0)}
		// var err error

		for i := 0; i < N; i++ {
			wgchild1.Add(1)
			go func(i, k int, mutex *sync.Mutex) {
				defer wgchild1.Done()
				commitPoints[i] = new(privacy.EllipticPoint)
				commitPoints[i], _ = privacy.DecompressCommitment(commitments[i])
				// commitPoints[i], err = privacy.DecompressCommitment(commitments[i])
				// if err != nil {
				// 	return nil, err
				// }

				//iBinary := privacy.ConvertIntToBinary(i, n)
				pik := GetCoefficient(i, k, n, a, indexIsZero)
				//mutex.

				tmp.X, tmp.Y = privacy.Curve.ScalarMult(commitPoints[i].X, commitPoints[i].Y, pik.Bytes())
				wgchild2.Add(1)
				go func(tmp privacy.EllipticPoint) { addRes(&wgchild2, mutex, tmp, i, k) }(tmp)

				// mutex.Lock()
				// fmt.Println(res.X, " ", i, " ", k)
				// res.X, res.Y = privacy.Curve.Add(res.X, res.Y, tmp.X, tmp.Y)
				// mutex.Unlock()
				// wg1.Done()
			}(i, k, &mutex)
		}
		wgchild1.Wait()
		wgchild2.Wait()
		comZero := privacy.Pcm.CommitAtIndex(big.NewInt(0).Bytes(), u[k], index)
		comZeroPoint, _ := privacy.DecompressCommitment(comZero)
		// comZeroPoint, err := privacy.DecompressCommitment(comZero)
		// if err != nil {
		// 	return nil, err
		// }
		res.X, res.Y = privacy.Curve.Add(res.X, res.Y, comZeroPoint.X, comZeroPoint.Y)
		cd := res.CompressPoint()
		proof.cd[k] = make([]byte, 33)
		copy(proof.cd[k], cd)
	}
	wg.Wait()
	// Calculate x
	x := big.NewInt(0)

	for j := 1; j <= n; j++ {
		x.SetBytes(privacy.Pcm.GetHashOfValues([][]byte{x.Bytes(), proof.cl[j], proof.ca[j], proof.cb[j], proof.cd[j-1]}))
	}
	x.Mod(x, privacy.Curve.Params().N)

	// Calculate za, zb zd
	proof.f = make([][]byte, n+1)
	proof.za = make([][]byte, n+1)
	proof.zb = make([][]byte, n+1)
	proof.zd = make([]byte, 32)
	temp = indexIsZero
	for j := n; j >= 1; j-- {
		tempid := byte(temp & 1)
		temp = temp >> 1
		wg.Add(1)
		go func(j int) {
			defer wg.Done()
			// f = lx + a
			fInt := big.NewInt(0)
			fInt.Mul(big.NewInt(int64(tempid)), x)
			fInt.Add(fInt, new(big.Int).SetBytes(a[j]))
			fInt.Mod(fInt, privacy.Curve.Params().N)
			proof.f[j] = make([]byte, 32)
			proof.f[j] = fInt.Bytes()

			// za = s + rx
			zaInt := big.NewInt(0)
			zaInt.Mul(new(big.Int).SetBytes(r[j]), x)
			zaInt.Add(zaInt, new(big.Int).SetBytes(s[j]))
			proof.za[j] = make([]byte, 32)
			proof.za[j] = zaInt.Bytes()

			// zb = r(x - f) + t
			zbInt := big.NewInt(0)
			zbInt.Sub(privacy.Curve.Params().N, fInt)
			zbInt.Add(zbInt, x)
			zbInt.Mul(zbInt, new(big.Int).SetBytes(r[j]))
			zbInt.Add(zbInt, new(big.Int).SetBytes(t[j]))
			zbInt.Mod(zbInt, privacy.Curve.Params().N)
			proof.zb[j] = make([]byte, 32)
			proof.zb[j] = zbInt.Bytes()
		}(j)
	}
	wg.Wait()
	zdInt := big.NewInt(0)
	zdInt.Exp(x, big.NewInt(int64(n)), privacy.Curve.Params().N)
	zdInt.Mul(zdInt, new(big.Int).SetBytes(rand))

	uxInt := big.NewInt(0)
	sumInt := big.NewInt(0)
	for k := 0; k < n; k++ {
		uxInt.Exp(x, big.NewInt(int64(k)), privacy.Curve.Params().N)
		uxInt.Mul(uxInt, new(big.Int).SetBytes(u[k]))
		sumInt.Add(sumInt, uxInt)
		sumInt.Mod(sumInt, privacy.Curve.Params().N)
	}

	sumInt.Sub(privacy.Curve.Params().N, sumInt)

	zdInt.Add(zdInt, sumInt)
	zdInt.Mod(zdInt, privacy.Curve.Params().N)
	proof.zd = zdInt.Bytes()

	return proof, nil
}

func (pro *PKOneOfManyProtocol) Verify(commitments [][]byte, proof *PKOneOfManyProof, index byte, rand []byte) bool {
	// var wg sync.WaitGroup
	N := len(commitments)
	n := bits.TrailingZeros(uint(N))
	if (n > 16) || (n+bits.LeadingZeros(uint(N)) != 63) {
		return false
	}
	clPoint := make([]*privacy.EllipticPoint, n+1)
	caPoint := make([]*privacy.EllipticPoint, n+1)
	cbPoint := make([]*privacy.EllipticPoint, n+1)
	cdPoint := make([]*privacy.EllipticPoint, n)
	var err error

	// Calculate x
	x := big.NewInt(0)
	for j := 1; j <= n; j++ {
		x.SetBytes(privacy.Pcm.GetHashOfValues([][]byte{x.Bytes(), proof.cl[j], proof.ca[j], proof.cb[j], proof.cd[j-1]}))
	}
	x.Mod(x, privacy.Curve.Params().N)
	//fmt.Printf("x Verify: %v\n", x)

	for i := 1; i <= n; i++ {
		// Decompress cl from bytes array to Elliptic
		clPoint[i] = new(privacy.EllipticPoint)
		clPoint[i], err = privacy.DecompressCommitment(proof.cl[i])
		if err != nil {
			return false
		}
		// Decompress ca from bytes array to Elliptic
		caPoint[i] = new(privacy.EllipticPoint)
		caPoint[i], err = privacy.DecompressCommitment(proof.ca[i])
		if err != nil {
			return false
		}
		// Decompress cb from bytes array to Elliptic
		cbPoint[i] = new(privacy.EllipticPoint)
		cbPoint[i], err = privacy.DecompressCommitment(proof.cb[i])
		if err != nil {
			return false
		}

		// Decompress cd from bytes array to Elliptic
		cdPoint[i-1] = new(privacy.EllipticPoint)
		cdPoint[i-1], err = privacy.DecompressKey(proof.cd[i-1])
		if err != nil {
			return false
		}

		// Check cl^x * ca = Com(f, za)
		leftPoint1 := new(privacy.EllipticPoint)
		leftPoint1.X, leftPoint1.Y = privacy.Curve.ScalarMult(clPoint[i].X, clPoint[i].Y, x.Bytes())
		leftPoint1.X, leftPoint1.Y = privacy.Curve.Add(leftPoint1.X, leftPoint1.Y, caPoint[i].X, caPoint[i].Y)

		rightPoint1 := new(privacy.EllipticPoint)
		right1 := privacy.Pcm.CommitAtIndex(proof.f[i], proof.za[i], index)
		rightPoint1, err = privacy.DecompressCommitment(right1)
		if err != nil {
			return false
		}

		//fmt.Printf("Left point 1 X: %v\n", leftPoint1.X)
		//fmt.Printf("Right point 1 X: %v\n", rightPoint1.X)
		//fmt.Printf("Left point 1 Y: %v\n", leftPoint1.Y)
		//fmt.Printf("Right point 1 Y: %v\n", rightPoint1.Y)

		if leftPoint1.X.Cmp(rightPoint1.X) != 0 || leftPoint1.Y.Cmp(rightPoint1.Y) != 0 {
			return false
		}

		// Check cl^(x-f) * cb = Com(0, zb)
		leftPoint2 := new(privacy.EllipticPoint)
		xSubF := new(big.Int)
		tmp := new(big.Int).SetBytes(proof.f[i])
		//fmt.Printf("tmp: %v\n", tmp)
		xSubF.Sub(x, tmp)
		xSubF.Mod(xSubF, privacy.Curve.Params().N)
		leftPoint2.X, leftPoint2.Y = privacy.Curve.ScalarMult(clPoint[i].X, clPoint[i].Y, xSubF.Bytes())
		leftPoint2.X, leftPoint2.Y = privacy.Curve.Add(leftPoint2.X, leftPoint2.Y, cbPoint[i].X, cbPoint[i].Y)

		rightPoint2 := new(privacy.EllipticPoint)
		right2 := privacy.Pcm.CommitAtIndex(big.NewInt(0).Bytes(), proof.zb[i], index)
		rightPoint2, err = privacy.DecompressCommitment(right2)
		if err != nil {
			return false
		}

		//fmt.Printf("Left point 2 X: %v\n", leftPoint2.X)
		//fmt.Printf("Right point 2 X: %v\n", rightPoint2.X)
		//fmt.Printf("Left point 2 Y: %v\n", leftPoint2.Y)
		//fmt.Printf("Right point 2 Y: %v\n", rightPoint2.Y)

		if leftPoint2.X.Cmp(rightPoint2.X) != 0 || leftPoint2.Y.Cmp(rightPoint2.Y) != 0 {
			return false
		}
	}

	commitPoints := make([]*privacy.EllipticPoint, N)
	leftPoint3 := privacy.EllipticPoint{X: big.NewInt(0), Y: big.NewInt(0)}
	//res = leftPoint3
	leftPoint32 := privacy.EllipticPoint{X: big.NewInt(0), Y: big.NewInt(0)}
	rightPoint3 := new(privacy.EllipticPoint)

	// var mutex sync.Mutex
	// var wgchild1 sync.WaitGroup
	for i := 0; i < N; i++ {
		tmpPoint := privacy.EllipticPoint{X: big.NewInt(0), Y: big.NewInt(0)}
		//wg.Add(1)
		//go func(i int, mutex *sync.Mutex) {
		// defer wg.Done()
		//iBinary := privacy.ConvertIntToBinary(i, n)
		temp := i
		commitPoints[i] = new(privacy.EllipticPoint)
		commitPoints[i], _ = privacy.DecompressCommitment(commitments[i])
		// if err != nil {
		// 	return false
		// }

		exp := big.NewInt(1)
		fji := big.NewInt(1)
		for j := n; j >= 1; j-- {
			if temp&1 == 1 {
				fji.SetBytes(proof.f[j])
			} else {
				fji.Sub(x, new(big.Int).SetBytes(proof.f[j]))
				fji.Mod(fji, privacy.Curve.Params().N)
			}

			exp.Mul(exp, fji)
			exp.Mod(exp, privacy.Curve.Params().N)
			temp = temp >> 1
		}

		tmpPoint.X, tmpPoint.Y = privacy.Curve.ScalarMult(commitPoints[i].X, commitPoints[i].Y, exp.Bytes())
		leftPoint3.X, leftPoint3.Y = privacy.Curve.Add(leftPoint3.X, leftPoint3.Y, tmpPoint.X, tmpPoint.Y)
		// wgchild1.Add(1)
		// go func(tmpPoint privacy.EllipticPoint) { addRes(&wgchild1, mutex, tmpPoint, 1, 1) }(tmpPoint)

		// }(i, &mutex)

	}

	// wg.Wait()
	// wgchild1.Wait()

	// leftPoint3 = res
	for k := 0; k < n; k++ {
		tmpPoint := privacy.EllipticPoint{X: big.NewInt(0), Y: big.NewInt(0)}
		xk := big.NewInt(0)
		xk.Exp(x, big.NewInt(int64(k)), privacy.Curve.Params().N)

		xk.Sub(privacy.Curve.Params().N, xk)

		tmpPoint.X, tmpPoint.Y = privacy.Curve.ScalarMult(cdPoint[k].X, cdPoint[k].Y, xk.Bytes())
		leftPoint32.X, leftPoint32.Y = privacy.Curve.Add(leftPoint32.X, leftPoint32.Y, tmpPoint.X, tmpPoint.Y)
	}

	leftPoint3.X, leftPoint3.Y = privacy.Curve.Add(leftPoint3.X, leftPoint3.Y, leftPoint32.X, leftPoint32.Y)

	rightValue3 := privacy.Pcm.CommitAtIndex(big.NewInt(0).Bytes(), proof.zd, index)
	rightPoint3, _ = privacy.DecompressCommitment(rightValue3)

	fmt.Printf("Left point 3 X: %v\n", leftPoint3.X)
	fmt.Printf("Right point 3 X: %v\n", rightPoint3.X)
	fmt.Printf("Left point 3 Y: %v\n", leftPoint3.Y)
	fmt.Printf("Right point 3 Y: %v\n", rightPoint3.Y)
	if leftPoint3.X.Cmp(rightPoint3.X) != 0 || leftPoint3.Y.Cmp(rightPoint3.Y) != 0 {
		return false
	}

	return true
}

//TestPKOneOfMany test protocol for one of many PedersenCommitment is PedersenCommitment to zero
func TestPKOneOfMany() {
	privacy.Pcm.InitCommitment()
	pk := new(PKOneOfManyProtocol)

	indexIsZero := 23

	// list of commitments
	commitments := make([][]byte, 32)
	serialNumbers := make([][]byte, 32)
	randoms := make([][]byte, 32)

	for i := 0; i < 32; i++ {
		serialNumbers[i] = privacy.RandBytes(32)
		randoms[i] = privacy.RandBytes(32)
		commitments[i] = make([]byte, 34)
		commitments[i] = privacy.Pcm.CommitAtIndex(serialNumbers[i], randoms[i], privacy.SND)
	}

	// create PedersenCommitment to zero at indexIsZero
	serialNumbers[indexIsZero] = big.NewInt(0).Bytes()
	commitments[indexIsZero] = privacy.Pcm.CommitAtIndex(serialNumbers[indexIsZero], randoms[indexIsZero], privacy.SND)

	start := time.Now()
	proof, err := pk.Prove(commitments, indexIsZero, commitments[indexIsZero], randoms[indexIsZero], privacy.SND)
	if err != nil {
		fmt.Println(err)
	}

	resbool := pk.Verify(commitments, proof, privacy.SND, randoms[indexIsZero])
	end := time.Now()
	fmt.Printf("%v_+_\n", end.Sub(start))
	fmt.Println(resbool)
}

// Get coefficient of x^k in polynomial pi(x)
// func GetCoefficient(iBinary []byte, k int, n int, a [][]byte, l []byte) *big.Int {
// 	res := privacy.Poly{big.NewInt(1)}
// 	var fji privacy.Poly

// 	for j := 1; j <= n; j++ {
// 		fj := privacy.Poly{new(big.Int).SetBytes(a[j]), big.NewInt(int64(l[j-1]))}

// 		if iBinary[j-1] == 0 {
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

// GetCoefficient of x^k in polynomial pi(x)
func GetCoefficient(index, k, n int, a [][]byte, indexIsZero int) *big.Int {
	resPoly := privacy.Poly{big.NewInt(1)}
	var fji privacy.Poly
	tempIndex := index
	tempindexIsZero := indexIsZero
	for j := n; j >= 1; j-- {
		fj := privacy.Poly{new(big.Int).SetBytes(a[j]), big.NewInt(int64(tempindexIsZero & 1))}
		if tempIndex&1 == 0 {
			fji = privacy.Poly{big.NewInt(0), big.NewInt(1)}.Sub(fj, privacy.Curve.Params().N)
		} else {
			fji = fj
		}
		resPoly = resPoly.Mul(fji, privacy.Curve.Params().N)
		tempIndex = tempIndex >> 1
		tempindexIsZero = tempindexIsZero >> 1
	}

	if resPoly.GetDegree() < k {
		return big.NewInt(0)
	}
	return resPoly[k]
}

//func TestGetCoefficient() {
//	i := 3
//	n := 4
//	l := 2
//	k := 0
//	iBinary := privacy.ConvertIntToBinary(i, n)
//
//	fmt.Printf("iBinary: %v\n", iBinary)
//	fmt.Printf("Binary 3: %v\n", iBinary[3])
//
//	lBinary := privacy.ConvertIntToBinary(l, n)
//
//	// a[i] = 2
//	a := make([][]byte, n+1)
//	for i:= 1; i<=n ;i++{
//		a[i] = big.NewInt(2).Bytes()
//		fmt.Printf("a[%v]: %v\n", i, a[i])
//	}
//
//	//
//	co := GetCoefficient(iBinary, k, n, a, lBinary)
//	fmt.Printf("Co: %v\n", co)
//}
