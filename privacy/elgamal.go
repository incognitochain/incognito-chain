package privacy

import (
	"crypto/elliptic"
	"math/big"
)

// ElGamalPublicKeyEncryption
// PrivateKey sk <-- Zn
// PublicKey  pk <-- G*sk
// Plaintext M is a EllipticPoint
// Ciphertext contains 2 EllipticPoint C1, C2
// C1 = G*k
// C2 = pk*k + M
// k <-- Zn is a secret random number

// ElGamalPubKey ...
// H = G^X
type ElGamalPubKey struct {
	Curve *elliptic.Curve
	H     *EllipticPoint
}

type ElGamalPrivKey struct {
	Curve *elliptic.Curve
	X     *big.Int
}

type ElGamalCiphertext struct {
	C1, C2 *EllipticPoint
}

func (ciphertext *ElGamalCiphertext) Set(C1, C2 *EllipticPoint) {
	ciphertext.C1 = C1
	ciphertext.C2 = C2
}

func (pub *ElGamalPubKey) Set(
	Curve *elliptic.Curve,
	H *EllipticPoint) {
	pub.Curve = Curve
	pub.H = H
}

func (priv *ElGamalPrivKey) Set(
	Curve *elliptic.Curve,
	Value *big.Int) {
	priv.Curve = Curve
	priv.X = Value
}

func (ciphertext *ElGamalCiphertext) Bytes() []byte {
	if ciphertext.C1.IsEqual(new(EllipticPoint).Zero()){
		return []byte{}
	}
	res := append(ciphertext.C1.Compress(), ciphertext.C2.Compress()...)
	return res
}

func (ciphertext *ElGamalCiphertext) SetBytes(bytearr []byte) error {
	if len(bytearr) == 0 {
		return nil
	}

	if ciphertext.C2 == nil {
		ciphertext.C2 = new(EllipticPoint)
	}
	if ciphertext.C1 == nil {
		ciphertext.C1 = new(EllipticPoint)
	}

	err := ciphertext.C1.Decompress(bytearr[:33])
	if err != nil{
		return err
	}
	err = ciphertext.C2.Decompress(bytearr[33:])
	if err != nil{
		return err
	}
	return nil
}

func (priv *ElGamalPrivKey) GenPubKey() *ElGamalPubKey {
	elGamalPubKey := new(ElGamalPubKey)

	pubKey := new(EllipticPoint)
	pubKey.Set((*priv.Curve).Params().Gx, (*priv.Curve).Params().Gy)
	pubKey = pubKey.ScalarMult(priv.X)

	elGamalPubKey.Set(priv.Curve, pubKey)
	return elGamalPubKey
}

func (pub *ElGamalPubKey) Encrypt(plaintext *EllipticPoint) *ElGamalCiphertext {
	randomness := RandInt()

	ciphertext := new(ElGamalCiphertext)

	ciphertext.C1 = new(EllipticPoint).Zero()
	ciphertext.C1.Set((*pub.Curve).Params().Gx, (*pub.Curve).Params().Gy)
	ciphertext.C1 = ciphertext.C1.ScalarMult(randomness)

	ciphertext.C2 = plaintext.Add((*pub.H).ScalarMult(randomness))

	return ciphertext
}

func (priv *ElGamalPrivKey) Decrypt(ciphertext *ElGamalCiphertext) *EllipticPoint {
	return ciphertext.C2.Sub(ciphertext.C1.ScalarMult(priv.X))
}
