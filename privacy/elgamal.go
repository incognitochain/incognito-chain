package privacy

import (
	"crypto/rand"
	"math/big"
)

// elGamalPublicKey represents to public key in ElGamal encryption
// H = G^X, X is private key
type elGamalPublicKey struct {
	h *EllipticPoint
}

// elGamalPrivateKey represents to private key in ElGamal encryption
type elGamalPrivateKey struct {
	x *big.Int
}

// elGamalCipherText represents to ciphertext in ElGamal encryption
// in which C1 = G^k and C2 = H^k * message
// k is a random number (32 bytes), message is an elliptic point
type elGamalCipherText struct {
	c1, c2 *EllipticPoint
}

func (ciphertext *elGamalCipherText) set(c1, c2 *EllipticPoint) {
	ciphertext.c1 = c1
	ciphertext.c2 = c2
}

func (pub *elGamalPublicKey) set(H *EllipticPoint) {
	pub.h = H
}

func (pub elGamalPublicKey) GetH() *EllipticPoint {
	return pub.h
}

func (priv *elGamalPrivateKey) set(x *big.Int) {
	priv.x = x
}

func (priv elGamalPrivateKey) GetX() *big.Int {
	return priv.x
}

// Bytes converts ciphertext to 66-byte array
func (ciphertext elGamalCipherText) Bytes() []byte {
	zero := new(EllipticPoint)
	zero.Zero()
	if ciphertext.c1.IsEqual(zero) {
		return []byte{}
	}
	res := append(ciphertext.c1.Compress(), ciphertext.c2.Compress()...)
	return res
}

// SetBytes reverts 66-byte array to ciphertext
func (ciphertext *elGamalCipherText) SetBytes(bytes []byte) error {
	if len(bytes) == 0 {
		return NewPrivacyErr(InvalidInputToSetBytesErr, nil)
	}

	ciphertext.c1 = new(EllipticPoint)
	ciphertext.c2 = new(EllipticPoint)

	err := ciphertext.c1.Decompress(bytes[:CompressedEllipticPointSize])
	if err != nil {
		return err
	}
	err = ciphertext.c2.Decompress(bytes[CompressedEllipticPointSize:])
	if err != nil {
		return err
	}
	return nil
}

// encrypt encrypts plaintext (is an elliptic point) using public key ElGamal
// returns ElGamal ciphertext
func (pub elGamalPublicKey) encrypt(plaintext *EllipticPoint) *elGamalCipherText {
	var r = rand.Reader
	randomness := RandScalar(r)

	ciphertext := new(elGamalCipherText)

	ciphertext.c1 = new(EllipticPoint)
	ciphertext.c1.Zero()
	ciphertext.c1.Set(Curve.Params().Gx, Curve.Params().Gy)
	ciphertext.c1 = ciphertext.c1.ScalarMult(randomness)

	ciphertext.c2 = plaintext.Add(pub.h.ScalarMult(randomness))

	return ciphertext
}

// decrypt receives a ciphertext and
// decrypts it using private key ElGamal
// and returns plain text in elliptic point
func (priv elGamalPrivateKey) decrypt(ciphertext *elGamalCipherText) (*EllipticPoint, error) {
	return ciphertext.c2.Sub(ciphertext.c1.ScalarMult(priv.x))
}
