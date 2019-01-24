package zkp

import (
	"encoding/binary"
	"fmt"
	"github.com/pkg/errors"
	"math/big"
	"time"

	"github.com/ninjadotorg/constant/privacy"
)

// OneOutOfManyWitness is a protocol for Zero-knowledge Proof of Knowledge of one out of many commitments containing 0
// include Witness: CommitedValue, r []byte
type OneOutOfManyWitness struct {
	rand        *big.Int
	indexIsZero uint64
	// general info
	commitments       []*privacy.EllipticPoint
	commitmentIndices []uint64
	index             byte
}

// OneOutOfManyProof contains Proof's value
type OneOutOfManyProof struct {
	cl, ca, cb, cd []*privacy.EllipticPoint
	f, za, zb      []*big.Int
	zd             *big.Int
	// general info
	Commitments       []*privacy.EllipticPoint
	CommitmentIndices []uint64
	index             byte
}

func (pro *OneOutOfManyProof) isNil() bool {
	if pro.cl == nil {
		return true
	}
	if pro.ca == nil {
		return true
	}
	if pro.cb == nil {
		return true
	}
	if pro.cd == nil {
		return true
	}
	if pro.f == nil {
		return true
	}
	if pro.za == nil {
		return true
	}
	if pro.zb == nil {
		return true
	}
	if pro.zd == nil {
		return true
	}
	if pro.CommitmentIndices == nil {
		return true
	}
	return false
}

func (pro *OneOutOfManyProof) Init() *OneOutOfManyProof {
	pro.zd = new(big.Int)
	return pro
}

// Set sets Witness
func (wit *OneOutOfManyWitness) Set(
	commitments []*privacy.EllipticPoint,
	commitmentIndices []uint64,
	rand *big.Int,
	indexIsZero uint64,
	index byte) {

	wit.commitmentIndices = commitmentIndices
	wit.commitments = commitments
	wit.indexIsZero = indexIsZero
	wit.rand = rand
	wit.index = index
}

// Set sets Proof
func (pro *OneOutOfManyProof) Set(
	commitmentIndexs []uint64,
	commitments []*privacy.EllipticPoint,
	cl, ca, cb, cd []*privacy.EllipticPoint,
	f, za, zb []*big.Int,
	zd *big.Int,
	index byte) {

	pro.CommitmentIndices = commitmentIndexs
	pro.Commitments = commitments
	pro.cl, pro.ca, pro.cb, pro.cd = cl, ca, cb, cd
	pro.f, pro.za, pro.zb = f, za, zb
	pro.zd = zd
	pro.index = index
}

// Bytes converts one of many proof to bytes array
func (pro *OneOutOfManyProof) Bytes() []byte {
	// if proof is nil, return an empty array
	if pro.isNil() {
		return []byte{}
	}

	// N = 2^n
	N := privacy.CMRingSize
	n := privacy.CMRingSizeExp

	var bytes []byte

	// convert array cl to bytes array
	for i := 0; i < n; i++ {
		bytes = append(bytes, pro.cl[i].Compress()...)
	}
	// convert array ca to bytes array
	for i := 0; i < n; i++ {
		bytes = append(bytes, pro.ca[i].Compress()...)
	}

	// convert array cb to bytes array
	for i := 0; i < n; i++ {
		bytes = append(bytes, pro.cb[i].Compress()...)
	}

	// convert array cd to bytes array
	for i := 0; i < n; i++ {
		bytes = append(bytes, pro.cd[i].Compress()...)
	}

	// convert array f to bytes array
	for i := 0; i < n; i++ {
		bytes = append(bytes, privacy.AddPaddingBigInt(pro.f[i], privacy.BigIntSize)...)
	}

	// convert array za to bytes array
	for i := 0; i < n; i++ {
		bytes = append(bytes, privacy.AddPaddingBigInt(pro.za[i], privacy.BigIntSize)...)
	}

	// convert array zb to bytes array
	for i := 0; i < n; i++ {
		bytes = append(bytes, privacy.AddPaddingBigInt(pro.zb[i], privacy.BigIntSize)...)
	}

	// convert array zd to bytes array
	bytes = append(bytes, privacy.AddPaddingBigInt(pro.zd, privacy.BigIntSize)...)

	// convert commitment index to bytes array
	for i := 0; i < N; i++ {
		commitmentIndexBytes := make([]byte, privacy.Uint64Size)
		binary.LittleEndian.PutUint64(commitmentIndexBytes, pro.CommitmentIndices[i])
		bytes = append(bytes, commitmentIndexBytes...)
	}

	// append index
	bytes = append(bytes, pro.index)

	return bytes
}

// SetBytes convert from bytes array to OneOutOfManyProof
func (pro *OneOutOfManyProof) SetBytes(bytes []byte) error {
	if pro == nil {
		pro = pro.Init()
	}

	if len(bytes) == 0 {
		return nil
	}

	N := privacy.CMRingSize
	n := privacy.CMRingSizeExp

	offset := 0
	var err error

	// get cl array
	pro.cl = make([]*privacy.EllipticPoint, n)
	for i := 0; i < n; i++ {
		pro.cl[i], err = privacy.DecompressKey(bytes[offset : offset+privacy.CompressedPointSize])
		if err != nil {
			return err
		}
		offset = offset + privacy.CompressedPointSize
	}
	// get ca array
	pro.ca = make([]*privacy.EllipticPoint, n)
	for i := 0; i < n; i++ {
		pro.ca[i], err = privacy.DecompressKey(bytes[offset : offset+privacy.CompressedPointSize])
		if err != nil {
			return err
		}
		offset = offset + privacy.CompressedPointSize
	}
	// get cb array
	pro.cb = make([]*privacy.EllipticPoint, n)
	for i := 0; i < n; i++ {
		pro.cb[i], err = privacy.DecompressKey(bytes[offset : offset+privacy.CompressedPointSize])
		if err != nil {
			return err
		}
		offset = offset + privacy.CompressedPointSize
	}

	// get cd array
	pro.cd = make([]*privacy.EllipticPoint, n)
	for i := 0; i < n; i++ {
		pro.cd[i], err = privacy.DecompressKey(bytes[offset : offset+privacy.CompressedPointSize])
		if err != nil {
			return err
		}
		offset = offset + privacy.CompressedPointSize
	}

	// get f array
	pro.f = make([]*big.Int, n)
	for i := 0; i < n; i++ {
		pro.f[i] = new(big.Int).SetBytes(bytes[offset : offset+privacy.BigIntSize])
		offset = offset + privacy.BigIntSize
	}

	// get za array
	pro.za = make([]*big.Int, n)
	for i := 0; i < n; i++ {
		pro.za[i] = new(big.Int).SetBytes(bytes[offset : offset+privacy.BigIntSize])
		offset = offset + privacy.BigIntSize
	}

	// get zb array
	pro.zb = make([]*big.Int, n)
	for i := 0; i < n; i++ {
		pro.zb[i] = new(big.Int).SetBytes(bytes[offset : offset+privacy.BigIntSize])
		offset = offset + privacy.BigIntSize
	}

	// get zd
	pro.zd = new(big.Int).SetBytes(bytes[offset : offset+privacy.BigIntSize])
	offset = offset + privacy.BigIntSize

	// get commitments list
	pro.CommitmentIndices = make([]uint64, N)
	for i := 0; i < N; i++ {
		pro.CommitmentIndices[i] = binary.LittleEndian.Uint64(bytes[offset : offset+privacy.Uint64Size])
		offset = offset + privacy.Uint64Size
	}

	//get index
	pro.index = bytes[len(bytes)-1]
	return nil
}

// Prove creates proof for one out of many commitments containing 0
func (wit *OneOutOfManyWitness) Prove() (*OneOutOfManyProof, error) {
	start :=time.Now()
	// Check the number of Commitment list's elements
	N := len(wit.commitments)
	if N != privacy.CMRingSize {
		return nil, errors.New("the number of Commitment list's elements must be equal to CMRingSize")
	}

	n := privacy.CMRingSizeExp

	// Check indexIsZero
	if wit.indexIsZero > uint64(N) {
		return nil, errors.New("Index is zero must be Index in list of commitments")
	}

	// Check Index
	if wit.index < privacy.SK || wit.index > privacy.RAND {
		return nil, errors.New("Index must be between index SK and index RAND")
	}

	// represent indexIsZero in binary
	indexIsZeroBinary := privacy.ConvertIntToBinary(int(wit.indexIsZero), n)

	//
	r := make([]*big.Int, n)
	a := make([]*big.Int, n)
	s := make([]*big.Int, n)
	t := make([]*big.Int, n)
	u := make([]*big.Int, n)

	cl := make([]*privacy.EllipticPoint, n)
	ca := make([]*privacy.EllipticPoint, n)
	cb := make([]*privacy.EllipticPoint, n)
	cd := make([]*privacy.EllipticPoint, n)

	for j := n - 1; j >= 0; j-- {
		// Generate random numbers
		r[j] = privacy.RandInt()
		a[j] = privacy.RandInt()
		s[j] = privacy.RandInt()
		t[j] = privacy.RandInt()
		u[j] = privacy.RandInt()

		// convert indexIsZeroBinary[j] to big.Int
		indexInt := big.NewInt(int64(indexIsZeroBinary[j]))

		// Calculate cl, ca, cb, cd
		// cl = Com(l, r)
		cl[j] = privacy.PedCom.CommitAtIndex(indexInt, r[j], wit.index)

		// ca = Com(a, s)
		ca[j] = privacy.PedCom.CommitAtIndex(a[j], s[j], wit.index)

		// cb = Com(la, t)
		la := new(big.Int).Mul(indexInt, a[j])
		la.Mod(la, privacy.Curve.Params().N)
		cb[j] = privacy.PedCom.CommitAtIndex(la, t[j], wit.index)
	}

	// Calculate: cd_k = ci^pi,k
	for k := 0; k < n; k++ {
		// Calculate pi,k which is coefficient of x^k in polynomial pi(x)
		cd[k] = new(privacy.EllipticPoint).Zero()

		for i := 0; i < N; i++ {
			iBinary := privacy.ConvertIntToBinary(i, n)
			pik := GetCoefficient(iBinary, k, n, a, indexIsZeroBinary)
			cd[k] = cd[k].Add(wit.commitments[i].ScalarMult(pik))
		}

		cd[k] = cd[k].Add(privacy.PedCom.CommitAtIndex(big.NewInt(0), u[k], wit.index))
	}

	// Calculate x
	x := big.NewInt(0)
	for j := 0; j <= n-1; j++ {
		x = generateChallengeFromByte([][]byte{x.Bytes(), cl[j].Compress(), ca[j].Compress(), cb[j].Compress(), cd[j].Compress()})
	}

	// Calculate za, zb zd
	za := make([]*big.Int, n)
	zb := make([]*big.Int, n)
	zd := new(big.Int)
	f := make([]*big.Int, n)

	for j := n - 1; j >= 0; j-- {
		// f = lx + a
		f[j] = new(big.Int).Mul(big.NewInt(int64(indexIsZeroBinary[j])), x)
		f[j].Add(f[j], a[j])
		f[j].Mod(f[j], privacy.Curve.Params().N)

		// za = s + rx
		za[j] = new(big.Int).Mul(r[j], x)
		za[j].Add(za[j], s[j])
		za[j].Mod(za[j], privacy.Curve.Params().N)

		// zb = r(x - f) + t
		zb[j] = new(big.Int).Sub(x, f[j])
		zb[j].Mul(zb[j], r[j])
		zb[j].Add(zb[j], t[j])
		zb[j].Mod(zb[j], privacy.Curve.Params().N)
	}

	// zd = rand * x^n - sum_{k=0}^{n-1} u[k] * x^k
	zd.Exp(x, big.NewInt(int64(n)), privacy.Curve.Params().N)
	zd.Mul(zd, wit.rand)

	uxInt := big.NewInt(0)
	sumInt := big.NewInt(0)
	for k := 0; k < n; k++ {
		uxInt.Exp(x, big.NewInt(int64(k)), privacy.Curve.Params().N)
		uxInt.Mul(uxInt, u[k])
		sumInt.Add(sumInt, uxInt)
		sumInt.Mod(sumInt, privacy.Curve.Params().N)
	}

	zd.Sub(zd, sumInt)
	zd.Mod(zd, privacy.Curve.Params().N)

	proof := new(OneOutOfManyProof).Init()
	proof.Set(wit.commitmentIndices, wit.commitments, cl, ca, cb, cd, f, za, zb, zd, wit.index)

	end := time.Since(start)
	fmt.Printf("One out of many proving time: %v\n", end)
	return proof, nil
}

func (pro *OneOutOfManyProof) Verify() bool {
	N := len(pro.Commitments)
	//N := 8

	// the number of Commitment list's elements must be equal to CMRingSize
	if N != privacy.CMRingSize {
		return false
	}
	n := privacy.CMRingSizeExp

	//Calculate x
	x := big.NewInt(0)

	for j := 0; j <= n-1; j++ {
		x = generateChallengeFromByte([][]byte{x.Bytes(), pro.cl[j].Compress(), pro.ca[j].Compress(), pro.cb[j].Compress(), pro.cd[j].Compress()})
	}

	for i := 0; i < n; i++ {
		// Check cl^x * ca = Com(f, za)
		leftPoint1 := pro.cl[i].ScalarMult(x).Add(pro.ca[i])
		rightPoint1 := privacy.PedCom.CommitAtIndex(pro.f[i], pro.za[i], pro.index)

		if !leftPoint1.IsEqual(rightPoint1) {
			return false
		}

		// Check cl^(x-f) * cb = Com(0, zb)
		xSubF := new(big.Int).Sub(x, pro.f[i])
		xSubF.Mod(xSubF, privacy.Curve.Params().N)

		leftPoint2 := pro.cl[i].ScalarMult(xSubF).Add(pro.cb[i])
		rightPoint2 := privacy.PedCom.CommitAtIndex(big.NewInt(0), pro.zb[i], pro.index)

		if !leftPoint2.IsEqual(rightPoint2) {
			return false
		}
	}

	leftPoint3 := new(privacy.EllipticPoint).Zero()
	leftPoint32 := new(privacy.EllipticPoint).Zero()

	for i := 0; i < N; i++ {
		iBinary := privacy.ConvertIntToBinary(i, n)

		exp := big.NewInt(1)
		fji := big.NewInt(1)
		for j := n - 1; j >= 0; j-- {
			if iBinary[j] == 1 {
				fji.Set(pro.f[j])
			} else {
				fji.Sub(x, pro.f[j])
				fji.Mod(fji, privacy.Curve.Params().N)
			}

			exp.Mul(exp, fji)
			exp.Mod(exp, privacy.Curve.Params().N)
		}

		leftPoint3 = leftPoint3.Add(pro.Commitments[i].ScalarMult(exp))
	}

	for k := 0; k < n; k++ {
		xk := big.NewInt(0).Exp(x, big.NewInt(int64(k)), privacy.Curve.Params().N)
		xk.Sub(privacy.Curve.Params().N, xk)

		leftPoint32 = leftPoint32.Add(pro.cd[k].ScalarMult(xk))
	}

	leftPoint3 = leftPoint3.Add(leftPoint32)

	rightPoint3 := privacy.PedCom.CommitAtIndex(big.NewInt(0), pro.zd, pro.index)

	return leftPoint3.IsEqual(rightPoint3)
}

// Get coefficient of x^k in polynomial pi(x)
func GetCoefficient(iBinary []byte, k int, n int, a []*big.Int, l []byte) *big.Int {
	res := privacy.Poly{big.NewInt(1)}
	var fji privacy.Poly

	for j := n - 1; j >= 0; j-- {
		fj := privacy.Poly{a[j], big.NewInt(int64(l[j]))}
		if iBinary[j] == 0 {
			fji = privacy.Poly{big.NewInt(0), big.NewInt(1)}.Sub(fj, privacy.Curve.Params().N)
		} else {
			fji = fj
		}
		res = res.Mul(fji, privacy.Curve.Params().N)
	}

	if res.GetDegree() < k {
		return big.NewInt(0)
	}
	return res[k]
}
