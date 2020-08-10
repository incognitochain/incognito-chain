package oneoutofmany

import (
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/privacy/zeroknowledge/utils"
	"github.com/pkg/errors"
	"math/big"
)

// This protocol proves in zero-knowledge that one-out-of-N commitments contains 0

// Statement to be proved
type OneOutOfManyStatement struct {
	Commitments []*privacy.Point
}

// Statement's witness
type OneOutOfManyWitness struct {
	stmt        *OneOutOfManyStatement
	rand        *privacy.Scalar
	indexIsZero uint64
}

// Statement's proof
type OneOutOfManyProof struct {
	Statement      *OneOutOfManyStatement
	cl, ca, cb, cd []*privacy.Point
	f, za, zb      []*privacy.Scalar
	zd             *privacy.Scalar
}

func (proof OneOutOfManyProof) ValidateSanity() bool {
	if len(proof.cl) != privacy.CommitmentRingSizeExp || len(proof.ca) != privacy.CommitmentRingSizeExp ||
		len(proof.cb) != privacy.CommitmentRingSizeExp || len(proof.cd) != privacy.CommitmentRingSizeExp ||
		len(proof.f) != privacy.CommitmentRingSizeExp || len(proof.za) != privacy.CommitmentRingSizeExp ||
		len(proof.zb) != privacy.CommitmentRingSizeExp {
		return false
	}

	for i := 0; i < len(proof.cl); i++ {
		if !proof.cl[i].PointValid() {
			return false
		}
		if !proof.ca[i].PointValid() {
			return false
		}
		if !proof.cb[i].PointValid() {
			return false
		}
		if !proof.cd[i].PointValid() {
			return false
		}

		if !proof.f[i].ScalarValid() {
			return false
		}
		if !proof.za[i].ScalarValid() {
			return false
		}
		if !proof.zb[i].ScalarValid() {
			return false
		}
	}

	return proof.zd.ScalarValid()
}

func (proof OneOutOfManyProof) isNil() bool {
	if proof.cl == nil {
		return true
	}
	if proof.ca == nil {
		return true
	}
	if proof.cb == nil {
		return true
	}
	if proof.cd == nil {
		return true
	}
	if proof.f == nil {
		return true
	}
	if proof.za == nil {
		return true
	}
	if proof.zb == nil {
		return true
	}
	return proof.zd == nil
}

func (proof *OneOutOfManyProof) Init() *OneOutOfManyProof {
	proof.zd = new(privacy.Scalar)
	proof.Statement = new(OneOutOfManyStatement)

	return proof
}

// Set sets Statement
func (stmt *OneOutOfManyStatement) Set(commitments []*privacy.Point) {
	stmt.Commitments = commitments
}

// Set sets Witness
func (wit *OneOutOfManyWitness) Set(commitments []*privacy.Point, rand *privacy.Scalar, indexIsZero uint64) {
	wit.stmt = new(OneOutOfManyStatement)
	wit.stmt.Set(commitments)

	wit.indexIsZero = indexIsZero
	wit.rand = rand
}

// Set sets Proof
func (proof *OneOutOfManyProof) Set(
	commitments []*privacy.Point,
	cl, ca, cb, cd []*privacy.Point,
	f, za, zb []*privacy.Scalar,
	zd *privacy.Scalar) {

	proof.Statement = new(OneOutOfManyStatement)
	proof.Statement.Set(commitments)

	proof.cl, proof.ca, proof.cb, proof.cd = cl, ca, cb, cd
	proof.f, proof.za, proof.zb = f, za, zb
	proof.zd = zd
}

// Bytes converts one of many proof to bytes array
func (proof OneOutOfManyProof) Bytes() []byte {
	// if proof is nil, return an empty array
	if proof.isNil() {
		return []byte{}
	}

	// N = 2^n
	n := privacy.CommitmentRingSizeExp

	var bytes []byte

	// convert array cl to bytes array
	for i := 0; i < n; i++ {
		bytes = append(bytes, proof.cl[i].ToBytesS()...)
	}
	// convert array ca to bytes array
	for i := 0; i < n; i++ {
		//fmt.Printf("proof.ca[i]: %v\n", proof.ca[i])
		//fmt.Printf("proof.ca[i]: %v\n", proof.ca[i].Compress())
		bytes = append(bytes, proof.ca[i].ToBytesS()...)
	}

	// convert array cb to bytes array
	for i := 0; i < n; i++ {
		bytes = append(bytes, proof.cb[i].ToBytesS()...)
	}

	// convert array cd to bytes array
	for i := 0; i < n; i++ {
		bytes = append(bytes, proof.cd[i].ToBytesS()...)
	}

	// convert array f to bytes array
	for i := 0; i < n; i++ {
		bytes = append(bytes, proof.f[i].ToBytesS()...)
	}

	// convert array za to bytes array
	for i := 0; i < n; i++ {
		bytes = append(bytes, proof.za[i].ToBytesS()...)
	}

	// convert array zb to bytes array
	for i := 0; i < n; i++ {
		bytes = append(bytes, proof.zb[i].ToBytesS()...)
	}

	// convert array zd to bytes array
	bytes = append(bytes, proof.zd.ToBytesS()...)

	return bytes
}

// SetBytes converts an array of bytes to an object of OneOutOfManyProof
func (proof *OneOutOfManyProof) SetBytes(bytes []byte) error {
	if len(bytes) == 0 {
		return nil
	}

	n := privacy.CommitmentRingSizeExp

	offset := 0
	var err error

	// get cl array
	proof.cl = make([]*privacy.Point, n)
	for i := 0; i < n; i++ {
		proof.cl[i], err = new(privacy.Point).FromBytesS(bytes[offset : offset+privacy.Ed25519KeySize])
		if err != nil {
			return err
		}
		offset = offset + privacy.Ed25519KeySize
	}

	// get ca array
	proof.ca = make([]*privacy.Point, n)
	for i := 0; i < n; i++ {
		proof.ca[i], err = new(privacy.Point).FromBytesS(bytes[offset : offset+privacy.Ed25519KeySize])
		if err != nil {
			return err
		}
		offset = offset + privacy.Ed25519KeySize
	}

	// get cb array
	proof.cb = make([]*privacy.Point, n)
	for i := 0; i < n; i++ {
		proof.cb[i], err = new(privacy.Point).FromBytesS(bytes[offset : offset+privacy.Ed25519KeySize])
		if err != nil {
			return err
		}
		offset = offset + privacy.Ed25519KeySize
	}

	// get cd array
	proof.cd = make([]*privacy.Point, n)
	for i := 0; i < n; i++ {
		proof.cd[i], err = new(privacy.Point).FromBytesS(bytes[offset : offset+privacy.Ed25519KeySize])
		if err != nil {
			return err
		}
		offset = offset + privacy.Ed25519KeySize
	}

	// get f array
	proof.f = make([]*privacy.Scalar, n)
	for i := 0; i < n; i++ {
		proof.f[i] = new(privacy.Scalar).FromBytesS(bytes[offset : offset+privacy.Ed25519KeySize])
		offset = offset + privacy.Ed25519KeySize
	}

	// get za array
	proof.za = make([]*privacy.Scalar, n)
	for i := 0; i < n; i++ {
		proof.za[i] = new(privacy.Scalar).FromBytesS(bytes[offset : offset+privacy.Ed25519KeySize])
		offset = offset + privacy.Ed25519KeySize
	}

	// get zb array
	proof.zb = make([]*privacy.Scalar, n)
	for i := 0; i < n; i++ {
		proof.zb[i] = new(privacy.Scalar).FromBytesS(bytes[offset : offset+privacy.Ed25519KeySize])
		offset = offset + privacy.Ed25519KeySize
	}

	// get zd
	proof.zd = new(privacy.Scalar).FromBytesS(bytes[offset : offset+privacy.Ed25519KeySize])

	return nil
}

// Prove produces a proof for the statement
func (wit OneOutOfManyWitness) Prove() (*OneOutOfManyProof, error) {
	// Check the number of Commitment list's elements
	N := len(wit.stmt.Commitments)
	if N != privacy.CommitmentRingSize {
		return nil, errors.New("the number of Commitment list's elements must be equal to CMRingSize")
	}

	n := privacy.CommitmentRingSizeExp

	// Check indexIsZero
	if wit.indexIsZero > uint64(N) {
		return nil, errors.New("Index is zero must be Index in list of commitments")
	}

	// represent indexIsZero in binary
	indexIsZeroBinary := privacy.ConvertIntToBinary(int(wit.indexIsZero), n)

	//
	r := make([]*privacy.Scalar, n)
	a := make([]*privacy.Scalar, n)
	s := make([]*privacy.Scalar, n)
	t := make([]*privacy.Scalar, n)
	u := make([]*privacy.Scalar, n)

	cl := make([]*privacy.Point, n)
	ca := make([]*privacy.Point, n)
	cb := make([]*privacy.Point, n)
	cd := make([]*privacy.Point, n)

	for j := 0; j < n; j++ {
		// Generate random numbers
		r[j] = privacy.RandomScalar()
		a[j] = privacy.RandomScalar()
		s[j] = privacy.RandomScalar()
		t[j] = privacy.RandomScalar()
		u[j] = privacy.RandomScalar()

		// convert indexIsZeroBinary[j] to privacy.Scalar
		indexInt := new(privacy.Scalar).FromUint64(uint64(indexIsZeroBinary[j]))

		// Calculate cl, ca, cb, cd
		// cl = Com(l, r)
		cl[j] = privacy.PedCom.CommitAtIndex(indexInt, r[j], privacy.PedersenPrivateKeyIndex)

		// ca = Com(a, s)
		ca[j] = privacy.PedCom.CommitAtIndex(a[j], s[j], privacy.PedersenPrivateKeyIndex)

		// cb = Com(la, t)
		la := new(privacy.Scalar).Mul(indexInt, a[j])
		//la.Mod(la, privacy.Curve.Params().N)
		cb[j] = privacy.PedCom.CommitAtIndex(la, t[j], privacy.PedersenPrivateKeyIndex)
	}

	// Calculate: cd_k = ci^pi,k
	for k := 0; k < n; k++ {
		// Calculate pi,k which is coefficient of x^k in polynomial pi(x)
		cd[k] = new(privacy.Point).Identity()

		for i := 0; i < N; i++ {
			iBinary := privacy.ConvertIntToBinary(i, n)
			pik := getCoefficient(iBinary, k, n, a, indexIsZeroBinary)
			cd[k].Add(cd[k], new(privacy.Point).ScalarMult(wit.stmt.Commitments[i], pik))
		}

		cd[k].Add(cd[k], privacy.PedCom.CommitAtIndex(new(privacy.Scalar).FromUint64(0), u[k], privacy.PedersenPrivateKeyIndex))
	}

	// Calculate x
	x := new(privacy.Scalar).FromUint64(0)
	for j := 0; j < n; j++ {
		x = utils.GenerateChallenge([][]byte{
			x.ToBytesS(),
			cl[j].ToBytesS(),
			ca[j].ToBytesS(),
			cb[j].ToBytesS(),
			cd[j].ToBytesS(),
		})
	}

	// Calculate za, zb zd
	za := make([]*privacy.Scalar, n)
	zb := make([]*privacy.Scalar, n)

	f := make([]*privacy.Scalar, n)

	for j := 0; j < n; j++ {
		// f = lx + a
		f[j] = new(privacy.Scalar).Mul(new(privacy.Scalar).FromUint64(uint64(indexIsZeroBinary[j])), x)
		f[j].Add(f[j], a[j])

		// za = s + rx
		za[j] = new(privacy.Scalar).Mul(r[j], x)
		za[j].Add(za[j], s[j])

		// zb = r(x - f) + t
		zb[j] = new(privacy.Scalar).Sub(x, f[j])
		zb[j].Mul(zb[j], r[j])
		zb[j].Add(zb[j], t[j])
	}

	// zd = rand * x^n - sum_{k=0}^{n-1} u[k] * x^k
	xi := new(privacy.Scalar).FromUint64(1)
	sum := new(privacy.Scalar).FromUint64(0)
	for k := 0; k < n; k++ {
		tmp := new(privacy.Scalar).Mul(xi, u[k])
		sum.Add(sum, tmp)
		xi.Mul(xi, x)
	}
	zd := new(privacy.Scalar).Mul(xi, wit.rand)
	zd.Sub(zd, sum)

	proof := new(OneOutOfManyProof).Init()
	proof.Set(wit.stmt.Commitments, cl, ca, cb, cd, f, za, zb, zd)

	return proof, nil
}

// Verify verifies a proof output by Prove
func (proof OneOutOfManyProof) Verify() (bool, error) {
	N := len(proof.Statement.Commitments)

	// the number of Commitment list's elements must be equal to CMRingSize
	if N != privacy.CommitmentRingSize {
		return false, errors.New("Invalid length of commitments list in one out of many proof")
	}
	n := privacy.CommitmentRingSizeExp

	//Calculate x
	x := new(privacy.Scalar).FromUint64(0)

	for j := 0; j < n; j++ {
		x = utils.GenerateChallenge([][]byte{x.ToBytesS(), proof.cl[j].ToBytesS(), proof.ca[j].ToBytesS(), proof.cb[j].ToBytesS(), proof.cd[j].ToBytesS()})
	}

	for i := 0; i < n; i++ {
		//Check cl^x * ca = Com(f, za)
		leftPoint1 := new(privacy.Point).ScalarMult(proof.cl[i], x)
		leftPoint1.Add(leftPoint1, proof.ca[i])

		rightPoint1 := privacy.PedCom.CommitAtIndex(proof.f[i], proof.za[i], privacy.PedersenPrivateKeyIndex)

		if !privacy.IsPointEqual(leftPoint1, rightPoint1) {
			privacy.Logger.Log.Errorf("verify one out of many proof statement 1 failed")
			return false, errors.New("verify one out of many proof statement 1 failed")
		}

		//Check cl^(x-f) * cb = Com(0, zb)
		xSubF := new(privacy.Scalar).Sub(x, proof.f[i])

		leftPoint2 := new(privacy.Point).ScalarMult(proof.cl[i], xSubF)
		leftPoint2.Add(leftPoint2, proof.cb[i])
		rightPoint2 := privacy.PedCom.CommitAtIndex(new(privacy.Scalar).FromUint64(0), proof.zb[i], privacy.PedersenPrivateKeyIndex)

		if !privacy.IsPointEqual(leftPoint2, rightPoint2) {
			privacy.Logger.Log.Errorf("verify one out of many proof statement 2 failed")
			return false, errors.New("verify one out of many proof statement 2 failed")
		}
	}

	leftPoint3 := new(privacy.Point).Identity()
	leftPoint32 := new(privacy.Point).Identity()

	for i := 0; i < N; i++ {
		iBinary := privacy.ConvertIntToBinary(i, n)

		exp := new(privacy.Scalar).FromUint64(1)
		fji := new(privacy.Scalar).FromUint64(1)
		for j := 0; j < n; j++ {
			if iBinary[j] == 1 {
				fji.Set(proof.f[j])
			} else {
				fji.Sub(x, proof.f[j])
			}

			exp.Mul(exp, fji)
		}

		leftPoint3.Add(leftPoint3, new(privacy.Point).ScalarMult(proof.Statement.Commitments[i], exp))
	}

	tmp2 := new(privacy.Scalar).FromUint64(1)
	for k := 0; k < n; k++ {
		xk := new(privacy.Scalar).Sub(new(privacy.Scalar).FromUint64(0), tmp2)
		leftPoint32.Add(leftPoint32, new(privacy.Point).ScalarMult(proof.cd[k], xk))
		tmp2.Mul(tmp2, x)
	}

	leftPoint3.Add(leftPoint3, leftPoint32)

	rightPoint3 := privacy.PedCom.CommitAtIndex(new(privacy.Scalar).FromUint64(0), proof.zd, privacy.PedersenPrivateKeyIndex)

	if !privacy.IsPointEqual(leftPoint3, rightPoint3) {
		privacy.Logger.Log.Errorf("verify one out of many proof statement 3 failed")
		return false, errors.New("verify one out of many proof statement 3 failed")
	}

	return true, nil
}

// Get coefficient of x^k in the polynomial p_i(x)
func getCoefficient(iBinary []byte, k int, n int, scLs []*privacy.Scalar, l []byte) *privacy.Scalar {

	a := make([]*big.Int, len(scLs))
	for i := 0; i < len(scLs); i++ {
		a[i] = privacy.ScalarToBigInt(scLs[i])
	}

	//AP2
	curveOrder := privacy.LInt
	res := privacy.Poly{big.NewInt(1)}
	var fji privacy.Poly
	for j := n - 1; j >= 0; j-- {
		fj := privacy.Poly{a[j], big.NewInt(int64(l[j]))}
		if iBinary[j] == 0 {
			fji = privacy.Poly{big.NewInt(0), big.NewInt(1)}.Sub(fj, curveOrder)
		} else {
			fji = fj
		}
		res = res.Mul(fji, curveOrder)
	}

	var sc2 *privacy.Scalar
	if res.GetDegree() < k {
		sc2 = new(privacy.Scalar).FromUint64(0)
	} else {
		sc2 = privacy.BigIntToScalar(res[k])
	}
	return sc2
}

func getCoefficientInt(iBinary []byte, k int, n int, a []*big.Int, l []byte) *big.Int {
	res := privacy.Poly{big.NewInt(1)}
	var fji privacy.Poly

	for j := n - 1; j >= 0; j-- {
		fj := privacy.Poly{a[j], big.NewInt(int64(l[j]))}
		if iBinary[j] == 0 {
			fji = privacy.Poly{big.NewInt(0), big.NewInt(1)}.Sub(fj, privacy.LInt)
		} else {
			fji = fj
		}

		res = res.Mul(fji, privacy.LInt)
	}

	if res.GetDegree() < k {
		return big.NewInt(0)
	}
	return res[k]
}
