package privacy

import (
	"fmt"
	"math/big"
)

// PKComZeroOneProtocol is a protocol for Zero-knowledge Proof of Knowledge of committed zero or one
// include witnesses: commitedValue, r []byte
type PKComZeroOneProtocol struct {
	witnesses [][]byte
}

// PKComZeroOneProof contains proof's value
type PKComZeroOneProof struct {
	ca, cb    []byte // 34 bytes
	f, za, zb []byte //32 bytes
}

// SetWitness sets witnesses
func (pro *PKComZeroOneProtocol) SetWitness(witnesses [][]byte) {
	pro.witnesses = make([][]byte, len(witnesses))
	for i := 0; i < len(witnesses); i++ {
		pro.witnesses[i] = make([]byte, len(witnesses[i]))
		copy(pro.witnesses[i], witnesses[i])
	}
}

// Prove creates a proof for commitment to zero or one
func (pro *PKComZeroOneProtocol) Prove(commitmentValue []byte, index byte) (*PKComZeroOneProof, error) {
	// Check index
	// indexInt := int(index)
	// fmt.Printf("index int: %v\n", indexInt)
	if index < 0 || index > 2 {
		return nil, fmt.Errorf("index must be between 0 and 2")
	}

	// Check value's witnessth is zero or one
	witness := big.NewInt(0)
	witness.SetBytes(pro.witnesses[0])
	fmt.Printf("witness: %v\n", witness)

	if witness.Cmp(big.NewInt(0)) != 0 && witness.Cmp(big.NewInt(0)) != 1 {
		return nil, fmt.Errorf("witness must be zero or one")
	}

	proof := new(PKComZeroOneProof)

	// Generate random numbers
	a := RandBytes(32)
	s := RandBytes(32)
	t := RandBytes(32)

	// Calculate ca, cb
	proof.ca = make([]byte, 34)
	proof.ca = Pcm.CommitSpecValue(a, s, index)
	am := big.NewInt(0)
	am.Mul(new(big.Int).SetBytes(a), witness)

	proof.cb = make([]byte, 34)
	proof.cb = Pcm.CommitSpecValue(am.Bytes(), t, index)

	// Calculate x = hash (G0||G1||G2||G3||ca||cb||cm)
	x := big.NewInt(0)
	x.SetBytes(Pcm.getHashOfValues([][]byte{proof.ca, proof.cb, commitmentValue}))
	x.Mod(x, Curve.Params().N)

	// Calculate f = mx + a
	f := big.NewInt(0)
	f.Mul(witness, x)
	f.Mod(f, Curve.Params().N)
	f.Add(f, new(big.Int).SetBytes(a))
	f.Mod(f, Curve.Params().N)
	proof.f = make([]byte, 32)
	copy(proof.f, f.Bytes())

	// Calculate za = rx + s
	za := big.NewInt(0)
	za.Mul(new(big.Int).SetBytes(pro.witnesses[1]), x)
	za.Mod(za, Curve.Params().N)
	za.Add(za, new(big.Int).SetBytes(s))
	za.Mod(za, Curve.Params().N)
	proof.za = make([]byte, 32)
	copy(proof.za, za.Bytes())

	// Calculate zb = r(x-f) + t
	zb := big.NewInt(0)
	xSubF := new(big.Int).Sub(x, f)
	xSubF.Mod(xSubF, Curve.Params().N)
	zb.Mul(new(big.Int).SetBytes(pro.witnesses[1]), xSubF)
	zb.Mod(zb, Curve.Params().N)
	zb.Add(zb, new(big.Int).SetBytes(t))
	zb.Mod(zb, Curve.Params().N)
	proof.zb = make([]byte, 32)
	copy(proof.zb, zb.Bytes())

	return proof, nil
}

// Verify verifies the proof for commitment to zero or one
func (pro *PKComZeroOneProtocol) Verify(proof *PKComZeroOneProof, commitmentValue []byte, index byte) bool {
	//Decompress commitment  value
	comPoint, err := DecompressCommitment(commitmentValue)
	if err != nil {
		fmt.Printf("Can not decompress commitment value to ECC point")
		return false
	}

	// Calculate x = hash (G0||G1||G2||G3||ca||cb||cm)
	x := big.NewInt(0)
	x.SetBytes(Pcm.getHashOfValues([][]byte{proof.ca, proof.cb, commitmentValue}))
	x.Mod(x, Curve.Params().N)

	// Decompress ca, cb of proof
	caPoint, err := DecompressCommitment(proof.ca)
	if err != nil {
		fmt.Printf("Can not decompress proof ca to ECC point")
		return false
	}
	cbPoint, err := DecompressCommitment(proof.cb)
	if err != nil {
		fmt.Printf("Can not decompress proof cb to ECC point")
		return false
	}

	// Calculate leftPoint1 = c^x * ca
	leftPoint1 := EllipticPoint{big.NewInt(0), big.NewInt(0)}
	leftPoint1.X, leftPoint1.Y = Curve.ScalarMult(comPoint.X, comPoint.Y, x.Bytes())
	leftPoint1.X, leftPoint1.Y = Curve.Add(leftPoint1.X, leftPoint1.Y, caPoint.X, caPoint.Y)

	// Calculate rightPoint1 = Com(f, za)
	rightValue1 := Pcm.CommitSpecValue(proof.f, proof.za, index)
	rightPoint1, err := DecompressCommitment(rightValue1)
	if err != nil {
		fmt.Printf("Can not decompress comitment for f")
		return false
	}

	// Calculate leftPoint2 = c^(x-f) * cb
	leftPoint2 := EllipticPoint{big.NewInt(0), big.NewInt(0)}
	xSubF := new(big.Int)
	xSubF.Sub(x, new(big.Int).SetBytes(proof.f))
	xSubF.Mod(xSubF, Curve.Params().N)
	leftPoint2.X, leftPoint2.Y = Curve.ScalarMult(comPoint.X, comPoint.Y, xSubF.Bytes())
	leftPoint2.X, leftPoint2.Y = Curve.Add(leftPoint2.X, leftPoint2.Y, cbPoint.X, cbPoint.Y)

	// Calculate rightPoint1 = Com(0, zb)
	rightValue2 := Pcm.CommitSpecValue(big.NewInt(0).Bytes(), proof.zb, index)
	rightPoint2, err := DecompressCommitment(rightValue2)
	if err != nil {
		fmt.Printf("Can not decompress comitment for zero")
		return false
	}

	if leftPoint1.X.Cmp(rightPoint1.X) == 0 && leftPoint1.Y.Cmp(rightPoint1.Y) == 0 && leftPoint2.X.Cmp(rightPoint2.X) == 0 && leftPoint2.Y.Cmp(rightPoint2.Y) == 0 {
		return true
	}

	return false
}

// TestPKComZeroOne tests prove and verify function for PK for commitment to zero or one
func TestPKComZeroOne() {

	Pcm.InitCommitment()
	spendingKey := GenerateSpendingKey(new(big.Int).SetInt64(123).Bytes())
	fmt.Printf("\nSpending key: %v\n", spendingKey)

	pubKey := GeneratePublicKey(spendingKey)
	serialNumber := RandBytes(32)

	value := make([]byte, 32, 32)
	valueInt := big.NewInt(1)
	value = valueInt.Bytes()

	r := RandBytes(32)
	coin := Coin{
		PublicKey:      pubKey,
		SerialNumber:   serialNumber,
		CoinCommitment: nil,
		R:              r,
		Value:          value,
	}
	coin.CommitAll()
	fmt.Println(coin.CoinCommitment)

	partialCommitment := Pcm.CommitSpecValue(coin.Value, coin.R, 1)

	witness := [][]byte{
		coin.Value,
		coin.R,
	}

	var zk PKComZeroOneProtocol

	zk.SetWitness(witness)
	proof, _ := zk.Prove(partialCommitment, 1)

	fmt.Printf("Proof: %+v\n", proof)

	res := zk.Verify(proof, partialCommitment, 1)
	fmt.Println(res)
}
