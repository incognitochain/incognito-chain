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
// Ciphertext contains 2 EllipticPoint R, C
// R = G*k
// C = pk*k + M
// k <-- Zn is a secret random number
type ElGamalPublicKeyEncryption interface {
	PubKeyGen() *ElGamalPubKey
	ElGamalEnc(plainPoint *EllipticPoint) *ElGamalCipherText
	ElGamalDec(cipher *ElGamalCipherText) *EllipticPoint
}

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

type ElGamalCipherText struct {
	R, C *EllipticPoint
}

func (elgamalCipher *ElGamalCipherText) Set(R, C *EllipticPoint) {
	elgamalCipher.C = C
	elgamalCipher.R = R
}

func (elgamalCipher *ElGamalCipherText) SetBytes(bytearr []byte) {
	if len(bytearr) == 0{
		return
	}
	if elgamalCipher.C == nil {
		elgamalCipher.C = new(EllipticPoint)
	}
	if elgamalCipher.R == nil {
		elgamalCipher.R = new(EllipticPoint)
	}
	elgamalCipher.C.Decompress(bytearr[:33])
	elgamalCipher.R.Decompress(bytearr[33:])
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

func (cipherText *ElGamalCipherText) Bytes() []byte {
	if cipherText.R.IsEqual(new(EllipticPoint).Zero()){
		return []byte{}
	}
	res := append(cipherText.C.Compress(), cipherText.R.Compress()...)
	return res
}

//func (cipherText *ElGamalCipherText) SetBytes(bytes []byte)  (err error) {
//	cipherText.C, err = DecompressKey(bytes[0:33])
//	res := append(cipherText.C.Compress(), cipherText.R.Compress()...)
//	return res
//}

func (priv *ElGamalPrivKey) PubKeyGen() *ElGamalPubKey {
	elGamalPubKey := new(ElGamalPubKey)
	publicKeyValue := new(EllipticPoint)
	publicKeyValue.X, publicKeyValue.Y = (*priv.Curve).ScalarMult((*priv.Curve).Params().Gx, (*priv.Curve).Params().Gy, priv.X.Bytes())
	elGamalPubKey.Set(priv.Curve, publicKeyValue)
	return elGamalPubKey
}

func (pub *ElGamalPubKey) ElGamalEnc(plainPoint *EllipticPoint) *ElGamalCipherText {
	rRnd, _ := rand.Int(rand.Reader, (*pub.Curve).Params().N)
	RRnd := new(EllipticPoint)
	RRnd.X, RRnd.Y = (*pub.Curve).ScalarMult((*pub.Curve).Params().Gx, (*pub.Curve).Params().Gy, rRnd.Bytes())
	Cipher := new(EllipticPoint)
	Cipher.X, Cipher.Y = (*pub.Curve).ScalarMult(pub.H.X, pub.H.Y, rRnd.Bytes())
	Cipher.X, Cipher.Y = (*pub.Curve).Add(Cipher.X, Cipher.Y, plainPoint.X, plainPoint.Y)
	elgamalCipher := new(ElGamalCipherText)
	elgamalCipher.Set(RRnd, Cipher)
	return elgamalCipher
}

func (priv *ElGamalPrivKey) ElGamalDec(cipher *ElGamalCipherText) *EllipticPoint {

	plainPoint := new(EllipticPoint)
	inversePrivKey := new(big.Int)
	inversePrivKey.Set((*priv.Curve).Params().N)
	inversePrivKey.Sub(inversePrivKey, priv.X)

	// fmt.Println(big.NewInt(0).Mod(big.NewInt(0).Add(inversePrivKey, priv.H), Curve.Params().N))
	plainPoint.X, plainPoint.Y = (*priv.Curve).ScalarMult(cipher.R.X, cipher.R.Y, inversePrivKey.Bytes())
	plainPoint.X, plainPoint.Y = (*priv.Curve).Add(plainPoint.X, plainPoint.Y, cipher.C.X, cipher.C.Y)
	return plainPoint
}

func TestElGamalPubKeyEncryption() bool {
	privKey := new(ElGamalPrivKey)
	privKey.Curve = new(elliptic.Curve)
	*privKey.Curve = elliptic.P256()
	privKey.X, _ = rand.Int(rand.Reader, (*privKey.Curve).Params().N)
	pubKey := privKey.PubKeyGen()

	mess := new(EllipticPoint)
	mess.Randomize()

	cipher := pubKey.ElGamalEnc(mess)
	ciphernew := new(ElGamalCipherText)
	ciphernew.SetBytes(cipher.Bytes())
	decPoint := privKey.ElGamalDec(ciphernew)
	return mess.IsEqual(decPoint)
}
