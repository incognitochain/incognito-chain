package privacy

import (
	"crypto/subtle"
	"errors"
	"github.com/incognitochain/incognito-chain/common"
)

// SchnorrPublicKey represents Schnorr Publickey
// PK = G^SK + H^R
type SchnorrPublicKey struct {
	publicKey *Point
	g, h      *Point
}

func (schnorrPubKey SchnorrPublicKey) GetPublicKey() *Point {
	return schnorrPubKey.publicKey
}

// SchnorrPrivateKey represents Schnorr Privatekey
type SchnorrPrivateKey struct {
	privateKey *Scalar
	randomness *Scalar
	publicKey  *SchnorrPublicKey
}

func (schnPrivKey SchnorrPrivateKey) GetPublicKey() *SchnorrPublicKey {
	return schnPrivKey.publicKey
}

// SchnSignature represents Schnorr Signature
type SchnSignature struct {
	e, z1, z2 *Scalar
}

// Set sets Schnorr private key
func (privateKey *SchnorrPrivateKey) Set(sk *Scalar, r *Scalar) {
	privateKey.privateKey = sk
	privateKey.randomness = r
	privateKey.publicKey = new(SchnorrPublicKey)
	privateKey.publicKey.g, _ = new(Point).SetKey(&PedCom.G[PedersenPrivateKeyIndex].key)
	privateKey.publicKey.h, _ = new(Point).SetKey(&PedCom.G[PedersenRandomnessIndex].key)
	privateKey.publicKey.publicKey = new(Point).ScalarMult(PedCom.G[PedersenPrivateKeyIndex], sk)
	privateKey.publicKey.publicKey.Add(privateKey.publicKey.publicKey, new(Point).ScalarMult(PedCom.G[PedersenRandomnessIndex], r))
}

// Set sets Schnorr public key
func (publicKey *SchnorrPublicKey) Set(pk *Point) {
	publicKey.publicKey, _ = new(Point).SetKey(&pk.key)

	publicKey.g, _ = new(Point).SetKey(&PedCom.G[PedersenPrivateKeyIndex].key)
	publicKey.h, _ = new(Point).SetKey(&PedCom.G[PedersenRandomnessIndex].key)
}

//Sign is function which using for signing on hash array by private key
func (privateKey SchnorrPrivateKey) Sign(data []byte) (*SchnSignature, error) {
	if len(data) != common.HashSize {
		return nil, NewPrivacyErr(UnexpectedErr, errors.New("hash length must be 32 bytes"))
	}

	signature := new(SchnSignature)

	// has privacy
	if !privateKey.randomness.IsZero() {
		// generates random numbers s1, s2 in [0, Curve.Params().N - 1]

		s1 := RandomScalar()
		s2 := RandomScalar()

		// t = s1*G + s2*H
		t := new(Point).ScalarMult(privateKey.publicKey.g, s1)
		t.Add(t, new(Point).ScalarMult(privateKey.publicKey.h, s2))

		// E is the hash of elliptic point t and data need to be signed
		msg := append(t.ToBytesS(), data...)

		signature.e = HashToScalar(msg)

		signature.z1 = new(Scalar).Mul(privateKey.privateKey, signature.e)
		signature.z1 = new(Scalar).Sub(s1, signature.z1)

		signature.z2 = new(Scalar).Mul(privateKey.randomness, signature.e)
		signature.z2 = new(Scalar).Sub(s2, signature.z2)

		return signature, nil
	}

	// generates random numbers s, k2 in [0, Curve.Params().N - 1]
	s := RandomScalar()

	// t = s*G
	t := new(Point).ScalarMult(privateKey.publicKey.g, s)

	// E is the hash of elliptic point t and data need to be signed
	msg := append(t.ToBytesS(), data...)
	signature.e = HashToScalar(msg)

	// Z1 = s - e*sk
	signature.z1 = new(Scalar).Mul(privateKey.privateKey, signature.e)
	signature.z1 = new(Scalar).Sub(s, signature.z1)

	signature.z2 = nil

	return signature, nil
}

//Verify is function which using for verify that the given signature was signed by by privatekey of the public key
func (publicKey SchnorrPublicKey) Verify(signature *SchnSignature, data []byte) bool {
	if signature == nil {
		return false
	}
	rv := new(Point).ScalarMult(publicKey.publicKey, signature.e)
	rv.Add(rv, new(Point).ScalarMult(publicKey.g, signature.z1))
	if signature.z2 != nil {
		rv.Add(rv, new(Point).ScalarMult(publicKey.h, signature.z2))
	}
	msg := append(rv.ToBytesS(), data...)

	ev := HashToScalar(msg)
	return subtle.ConstantTimeCompare(ev.ToBytesS(), signature.e.ToBytesS()) == 1
}

func (sig SchnSignature) Bytes() []byte {
	bytes := append(sig.e.ToBytesS(), sig.z1.ToBytesS()...)
	// Z2 is nil when has no privacy
	if sig.z2 != nil {
		bytes = append(bytes, sig.z2.ToBytesS()...)
	}
	return bytes
}

func (sig *SchnSignature) SetBytes(bytes []byte) error {
	if len(bytes) == 0 {
		return NewPrivacyErr(InvalidInputToSetBytesErr, nil)
	}
	sig.e = new(Scalar).FromBytesS(bytes[0:Ed25519KeySize])
	sig.z1 = new(Scalar).FromBytesS(bytes[Ed25519KeySize : 2*Ed25519KeySize])
	if len(bytes) == 3*Ed25519KeySize {
		sig.z2 = new(Scalar).FromBytesS(bytes[2*Ed25519KeySize:])
	} else {
		sig.z2 = nil
	}

	return nil
}
