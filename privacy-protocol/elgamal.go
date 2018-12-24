package privacy

import (
	"crypto/elliptic"
	"crypto/rand"
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

func (ciphertext *ElGamalCiphertext) SetBytes(bytearr []byte) {
	if len(bytearr) == 0 {
		return
	}
	if ciphertext.C2 == nil {
		ciphertext.C2 = new(EllipticPoint)
	}
	if ciphertext.C1 == nil {
		ciphertext.C1 = new(EllipticPoint)
	}
	ciphertext.C1.Decompress(bytearr[:33])
	ciphertext.C2.Decompress(bytearr[33:])
}

func (priv *ElGamalPrivKey) PubKeyGen() *ElGamalPubKey {
	elGamalPubKey := new(ElGamalPubKey)
	publicKeyValue := new(EllipticPoint)
	publicKeyValue.X, publicKeyValue.Y = (*priv.Curve).ScalarMult((*priv.Curve).Params().Gx, (*priv.Curve).Params().Gy, priv.X.Bytes())
	elGamalPubKey.Set(priv.Curve, publicKeyValue)
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


func TestElGamalPubKeyEncryption() bool {
	privKey := new(ElGamalPrivKey)
	privKey.Curve = new(elliptic.Curve)
	*privKey.Curve = elliptic.P256()
	privKey.X, _ = rand.Int(rand.Reader, (*privKey.Curve).Params().N)
	pubKey := privKey.PubKeyGen()

	mess := new(EllipticPoint)
	mess.Randomize()

	cipher := pubKey.Encrypt(mess)
	ciphernew := new(ElGamalCiphertext)
	ciphernew.SetBytes(cipher.Bytes())
	decPoint := privKey.Decrypt(ciphernew)
	return mess.IsEqual(decPoint)
}
