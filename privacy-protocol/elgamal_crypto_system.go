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
	res := append(cipherText.C.Compress(), cipherText.R.Compress()...)
	return res
}

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
	return mess.IsEqual(*decPoint)
}

// // Copyright 2011 The Go Authors. All rights reserved.
// // Use of this source code is governed by a BSD-style
// // license that can be found in the LICENSE file.

// // Package elgamal implements ElGamal encryption, suitable for OpenPGP,
// // as specified in "A Public-Key Cryptosystem and a Signature Scheme Based on
// // Discrete Logarithms," IEEE Transactions on Information Theory, v. IT-31,
// // n. 4, 1985, pp. 469-472.
// //
// // This form of ElGamal embeds PKCS#1 v1.5 padding, which may make it
// // unsuitable for other protocols. RSA should be used in preference in any
// // case.

// import (
// "crypto/rand"
// "crypto/subtle"
// "errors"
// "io"
// "math/big"
// )

// // PublicKey represents an ElGamal public key.
// // Em nho cmt tung thanh phan trong struct PublicKey la gi nhe
// //
// type PublicKey struct {
// 	G, P, Y *big.Int
// }

// // PrivvateKey represents an ElGamal Privvate key.
// type PrivvateKey struct {
// 	PublicKey
// 	X *big.Int
// }

// // Encrypt encrypts the given message to the given public key. The result is a
// // pair of integers. Errors can result from reading random, or because msg is
// // too large to be encrypted to the public key.
// func Encrypt(random io.Reader, pub *PublicKey, msg []byte) (c1, c2 *big.Int, err error) {
// 	pLen := (pub.P.BitLen() + 7) / 8
// 	if len(msg) > pLen-11 {
// 		err = errors.New("elgamal: message too long")
// 		return
// 	}

// 	// EM = 0x02 || PS || 0x00 || M
// 	em := make([]byte, pLen-1)
// 	em[0] = 2
// 	ps, mm := em[1:len(em)-len(msg)-1], em[len(em)-len(msg):]
// 	err = nonZeroRandomBytes(ps, random)
// 	if err != nil {
// 		return
// 	}
// 	em[len(em)-len(msg)-1] = 0
// 	copy(mm, msg)

// 	m := new(big.Int).SetBytes(em)

// 	k, err := rand.Int(random, pub.P)
// 	if err != nil {
// 		return
// 	}

// 	c1 = new(big.Int).Exp(pub.G, k, pub.P)
// 	s := new(big.Int).Exp(pub.Y, k, pub.P)
// 	c2 = s.Mul(s, m)
// 	c2.Mod(c2, pub.P)

// 	return
// }

// // Decrypt takes two integers, resulting from an ElGamal encryption, and
// // returns the plaintext of the message. An error can result only if the
// // ciphertext is invalid. Users should keep in mind that this is a padding
// // oracle and thus, if exposed to an adaptive chosen ciphertext attack, can
// // be used to break the cryptosystem.  See ``Chosen Ciphertext Attacks
// // Against Protocols Based on the RSA Encryption Standard PKCS #1'', Daniel
// // Bleichenbacher, Advances in Cryptology (Crypto '98),
// func Decrypt(Privv *PrivvateKey, c1, c2 *big.Int) (msg []byte, err error) {
// 	s := new(big.Int).Exp(c1, Privv.X, Privv.P)
// 	s.ModInverse(s, Privv.P)
// 	s.Mul(s, c2)
// 	s.Mod(s, Privv.P)
// 	em := s.Bytes()

// 	// return nil, errors.New("fuck")

// 	firstByteIsTwo := subtle.ConstantTimeByteEq(em[0], 2)

// 	// The remainder of the plaintext must be a string of non-zero random
// 	// octets, followed by a 0, followed by the message.
// 	//   lookingForIndex: 1 iff we are still looking for the zero.
// 	//   index: the offset of the first zero byte.
// 	var lookingForIndex, index int
// 	lookingForIndex = 1

// 	for i := 1; i < len(em); i++ {
// 		equals0 := subtle.ConstantTimeByteEq(em[i], 0)
// 		index = subtle.ConstantTimeSelect(lookingForIndex&equals0, i, index)
// 		lookingForIndex = subtle.ConstantTimeSelect(equals0, 0, lookingForIndex)
// 	}

// 	if firstByteIsTwo != 1 || lookingForIndex != 0 || index < 9 {
// 		return nil, errors.New("elgamal: decryption error")
// 	}
// 	return em[index+1:], nil
// }

// // nonZeroRandomBytes fills the given slice with non-zero random octets.
// func nonZeroRandomBytes(s []byte, rand io.Reader) (err error) {
// 	_, err = io.ReadFull(rand, s)
// 	if err != nil {
// 		return
// 	}

// 	for i := 0; i < len(s); i++ {
// 		for s[i] == 0 {
// 			_, err = io.ReadFull(rand, s[i:i+1])
// 			if err != nil {
// 				return
// 			}
// 		}
// 	}

// 	return
// }
