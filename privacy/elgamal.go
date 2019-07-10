package privacy

import (
	"github.com/pkg/errors"
	"math/big"
)

// ElGamalPubKey represents to public key in ElGamal encryption
// H = G^X, X is private key
type ElGamalPubKey struct {
	H *EllipticPoint
}

// ElGamalPrivKey represents to private key in ElGamal encryption
type ElGamalPrivKey struct {
	X *big.Int
}

// ElGamalCiphertext represents to ciphertext in ElGamal encryption
// in which C1 = G^k and C2 = H^k * message
// k is a random number (32 bytes), message is an elliptic point
type ElGamalCiphertext struct {
	C1, C2 *EllipticPoint
}

func (ciphertext *ElGamalCiphertext) Set(C1, C2 *EllipticPoint) {
	ciphertext.C1 = C1
	ciphertext.C2 = C2
}

func (pub *ElGamalPubKey) Set(H *EllipticPoint) {
	pub.H = H
}

func (priv *ElGamalPrivKey) Set(x *big.Int) {
	priv.X = x
}

// Bytes converts ciphertext to 66-byte array
func (ciphertext *ElGamalCiphertext) Bytes() []byte {
	if ciphertext.C1.IsEqual(new(EllipticPoint).Zero()) {
		return []byte{}
	}
	res := append(ciphertext.C1.Compress(), ciphertext.C2.Compress()...)
	return res
}

// SetBytes reverts 66-byte array to ciphertext
func (ciphertext *ElGamalCiphertext) SetBytes(bytes []byte) error {
	if len(bytes) == 0 {
		return errors.New("Length bytes array of Elgamal ciphertext is empty")
	}

	ciphertext.C1 = new(EllipticPoint)
	ciphertext.C2 = new(EllipticPoint)

	err := ciphertext.C1.Decompress(bytes[:CompressedPointSize])
	if err != nil {
		return err
	}
	err = ciphertext.C2.Decompress(bytes[CompressedPointSize:])
	if err != nil {
		return err
	}
	return nil
}

// Encrypt encrypts plaintext (is an elliptic point) using public key ElGamal
// returns ElGamal ciphertext
func (pub *ElGamalPubKey) Encrypt(plaintext *EllipticPoint) *ElGamalCiphertext {
	randomness := RandScalar()

	ciphertext := new(ElGamalCiphertext)

	ciphertext.C1 = new(EllipticPoint).Zero()
	ciphertext.C1.Set(Curve.Params().Gx, Curve.Params().Gy)
	ciphertext.C1 = ciphertext.C1.ScalarMult(randomness)

	ciphertext.C2 = plaintext.Add(pub.H.ScalarMult(randomness))

	return ciphertext
}

// Decrypt receives a ciphertext and
// decrypts it using private key ElGamal
// and returns plain text in elliptic point
func (priv *ElGamalPrivKey) Decrypt(ciphertext *ElGamalCiphertext) (*EllipticPoint, error) {
	return ciphertext.C2.Sub(ciphertext.C1.ScalarMult(priv.X))
}
