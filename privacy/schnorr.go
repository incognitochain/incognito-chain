package privacy

import (
	"crypto/rand"
	"errors"
	"math/big"

	"github.com/incognitochain/incognito-chain/common"
)

// SchnorrPublicKey represents Schnorr Publickey
// PK = G^SK + H^R
type SchnorrPublicKey struct {
	publicKey *EllipticPoint
	g, h      *EllipticPoint
}

func (schnorrPubKey SchnorrPublicKey) GetPublicKey() *EllipticPoint {
	return schnorrPubKey.publicKey
}

// SchnorrPrivateKey represents Schnorr Privatekey
type SchnorrPrivateKey struct {
	privateKey *big.Int
	randomness *big.Int
	publicKey  *SchnorrPublicKey
}

func (schnPrivKey SchnorrPrivateKey) GetPublicKey() *SchnorrPublicKey {
	return schnPrivKey.publicKey
}

// SchnSignature represents Schnorr Signature
type SchnSignature struct {
	e, z1, z2 *big.Int
}

// Set sets Schnorr private key
func (privateKey *SchnorrPrivateKey) Set(sk *big.Int, r *big.Int) {
	privateKey.privateKey = sk
	privateKey.randomness = r
	privateKey.publicKey = new(SchnorrPublicKey)
	privateKey.publicKey.g = new(EllipticPoint)
	privateKey.publicKey.g.Set(PedCom.G[PedersenPrivateKeyIndex].x, PedCom.G[PedersenPrivateKeyIndex].y)

	privateKey.publicKey.h = new(EllipticPoint)
	privateKey.publicKey.h.Set(PedCom.G[PedersenRandomnessIndex].x, PedCom.G[PedersenRandomnessIndex].y)
	privateKey.publicKey.publicKey = PedCom.G[PedersenPrivateKeyIndex].ScalarMult(sk).Add(PedCom.G[PedersenRandomnessIndex].ScalarMult(r))
}

// Set sets Schnorr public key
func (publicKey *SchnorrPublicKey) Set(pk *EllipticPoint) {
	publicKey.publicKey = new(EllipticPoint)
	publicKey.publicKey.Set(pk.x, pk.y)

	publicKey.g = new(EllipticPoint)
	publicKey.g.Set(PedCom.G[PedersenPrivateKeyIndex].x, PedCom.G[PedersenPrivateKeyIndex].y)

	publicKey.h = new(EllipticPoint)
	publicKey.h.Set(PedCom.G[PedersenRandomnessIndex].x, PedCom.G[PedersenRandomnessIndex].y)
}

//Sign is function which using for signing on hash array by private key
func (privateKey SchnorrPrivateKey) Sign(data []byte) (*SchnSignature, error) {
	if len(data) != common.HashSize {
		return nil, NewPrivacyErr(UnexpectedErr, errors.New("hash length must be 32 bytes"))
	}

	signature := new(SchnSignature)

	var r = rand.Reader
	// has privacy
	if privateKey.randomness.Cmp(big.NewInt(0)) != 0 {
		// generates random numbers s1, s2 in [0, Curve.Params().N - 1]

		s1 := RandScalar(r)
		s2 := RandScalar(r)

		// t = s1*G + s2*H
		t := privateKey.publicKey.g.ScalarMult(s1).Add(privateKey.publicKey.h.ScalarMult(s2))

		// E is the hash of elliptic point t and data need to be signed
		signature.e = Hash(*t, data)

		signature.z1 = new(big.Int).Sub(s1, new(big.Int).Mul(privateKey.privateKey, signature.e))
		signature.z1.Mod(signature.z1, Curve.Params().N)

		signature.z2 = new(big.Int).Sub(s2, new(big.Int).Mul(privateKey.randomness, signature.e))
		signature.z2.Mod(signature.z2, Curve.Params().N)

		return signature, nil
	}

	// generates random numbers s, k2 in [0, Curve.Params().N - 1]
	s := RandScalar(r)

	// t = s*G
	t := privateKey.publicKey.g.ScalarMult(s)

	// E is the hash of elliptic point t and data need to be signed
	signature.e = Hash(*t, data)

	// Z1 = s - e*sk
	signature.z1 = new(big.Int).Sub(s, new(big.Int).Mul(privateKey.privateKey, signature.e))
	signature.z1.Mod(signature.z1, Curve.Params().N)

	return signature, nil
}

//Verify is function which using for verify that the given signature was signed by by privatekey of the public key
func (publicKey SchnorrPublicKey) Verify(signature *SchnSignature, data []byte) bool {
	if signature == nil {
		return false
	}

	rv := publicKey.g.ScalarMult(signature.z1).Add(publicKey.h.ScalarMult(signature.z2))
	rv = rv.Add(publicKey.publicKey.ScalarMult(signature.e))

	ev := Hash(*rv, data)
	return ev.Cmp(signature.e) == 0
}

func (sig SchnSignature) Bytes() []byte {
	bytes := append(common.AddPaddingBigInt(sig.e, common.BigIntSize), common.AddPaddingBigInt(sig.z1, common.BigIntSize)...)
	// Z2 is nil when has no privacy
	if sig.z2 != nil {
		bytes = append(bytes, common.AddPaddingBigInt(sig.z2, common.BigIntSize)...)
	}
	return bytes
}

func (sig *SchnSignature) SetBytes(bytes []byte) error {
	if len(bytes) == 0 {
		return NewPrivacyErr(InvalidInputToSetBytesErr, nil)
	}

	sig.e = new(big.Int).SetBytes(bytes[0:common.BigIntSize])
	sig.z1 = new(big.Int).SetBytes(bytes[common.BigIntSize : 2*common.BigIntSize])
	sig.z2 = new(big.Int).SetBytes(bytes[2*common.BigIntSize:])
	return nil
}

// Hash calculates a hash concatenating a given message bytes with a given EC Point. H(p||m)
func Hash(p EllipticPoint, m []byte) *big.Int {
	b := append(common.AddPaddingBigInt(p.x, common.BigIntSize), common.AddPaddingBigInt(p.y, common.BigIntSize)...)
	b = append(b, m...)

	return new(big.Int).SetBytes(common.HashB(b))
}
