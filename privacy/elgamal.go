package privacy

import (
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
	H *EllipticPoint
}

type ElGamalPrivKey struct {
	X *big.Int
}

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

// Bytes -  returned value always is 66 bytes
func (ciphertext *ElGamalCiphertext) Bytes() []byte {
	if ciphertext.C1.IsEqual(new(EllipticPoint).Zero()) {
		return []byte{}
	}
	res := append(ciphertext.C1.Compress(), ciphertext.C2.Compress()...)
	return res
}

func (ciphertext *ElGamalCiphertext) SetBytes(bytes []byte) error {
	if len(bytes) == 0 {
		return nil
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

func (priv *ElGamalPrivKey) GenPubKey() *ElGamalPubKey {
	elGamalPubKey := new(ElGamalPubKey)

	pubKey := new(EllipticPoint)
	pubKey.Set(Curve.Params().Gx, Curve.Params().Gy)
	pubKey = pubKey.ScalarMult(priv.X)

	elGamalPubKey.Set(pubKey)
	return elGamalPubKey
}

func (pub *ElGamalPubKey) Encrypt(plaintext *EllipticPoint) *ElGamalCiphertext {
	randomness := RandScalar()

	ciphertext := new(ElGamalCiphertext)

	ciphertext.C1 = new(EllipticPoint).Zero()
	ciphertext.C1.Set(Curve.Params().Gx, Curve.Params().Gy)
	ciphertext.C1 = ciphertext.C1.ScalarMult(randomness)

	ciphertext.C2 = plaintext.Add(pub.H.ScalarMult(randomness))

	return ciphertext
}

func (priv *ElGamalPrivKey) Decrypt(ciphertext *ElGamalCiphertext) (*EllipticPoint, error) {
	return ciphertext.C2.Sub(ciphertext.C1.ScalarMult(priv.X))
}

//func ElGamalEncrypt(pubKey []byte, data *EllipticPoint) ([]byte, error) {
//	elgamalPub := ElGamalPubKey{
//		H: new(EllipticPoint),
//	}
//
//	err := elgamalPub.H.Decompress(pubKey)
//	if err != nil {
//		return nil, err
//	}
//
//	cipher := elgamalPub.Encrypt(data)
//	return cipher.Bytes(), nil
//}
//
//func ElGamalDecrypt(privKey []byte, cipher []byte) (*EllipticPoint, error) {
//	elgamalPri := ElGamalPrivKey{
//		X: new(big.Int),
//	}
//	elgamalPri.X.SetBytes(privKey)
//
//	ciphertext := &ElGamalCiphertext{}
//	ciphertext.SetBytes(cipher)
//
//	result, err := elgamalPri.Decrypt(ciphertext)
//	return result, err
//}
