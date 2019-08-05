package privacy

import (
	"errors"
	"math/big"

	"github.com/incognitochain/incognito-chain/common"
)

//SchnorrPubKey represents Schnorr Publickey
// PK = G^SK + H^R
type SchnorrPubKey struct {
	PK, G, H *EllipticPoint
}

//SchnorrPrivateKey represents Schnorr Privatekey
type SchnorrPrivateKey struct {
	secretKey, r *big.Int
	publicKey    *SchnorrPubKey
}

func (schnPrivKey SchnorrPrivateKey) GetPublicKey() *SchnorrPubKey {
	return schnPrivKey.publicKey
}

//SchnSignature represents Schnorr Signature
type SchnSignature struct {
	e, z1, z2 *big.Int
}

// Set sets Schnorr private key
func (priKey *SchnorrPrivateKey) Set(sk *big.Int, r *big.Int) {
	priKey.secretKey = sk
	priKey.r = r
	priKey.publicKey = new(SchnorrPubKey)
	priKey.publicKey.G = new(EllipticPoint)
	priKey.publicKey.G.Set(PedCom.G[SK].X, PedCom.G[SK].Y)

	priKey.publicKey.H = new(EllipticPoint)
	priKey.publicKey.H.Set(PedCom.G[RAND].X, PedCom.G[RAND].Y)
	priKey.publicKey.PK = PedCom.G[SK].ScalarMult(sk).Add(PedCom.G[RAND].ScalarMult(r))
}

// Set sets Schnorr public key
func (pubKey *SchnorrPubKey) Set(pk *EllipticPoint) {
	pubKey.PK = new(EllipticPoint)
	pubKey.PK.Set(pk.X, pk.Y)

	pubKey.G = new(EllipticPoint)
	pubKey.G.Set(PedCom.G[SK].X, PedCom.G[SK].Y)

	pubKey.H = new(EllipticPoint)
	pubKey.H.Set(PedCom.G[RAND].X, PedCom.G[RAND].Y)
}

//Sign is function which using for signing on hash array by private key
func (priKey SchnorrPrivateKey) Sign(data []byte) (*SchnSignature, error) {
	if len(data) != common.HashSize {
		return nil, NewPrivacyErr(UnexpectedErr, errors.New("hash length must be 32 bytes"))
	}

	signature := new(SchnSignature)

	// has privacy
	if priKey.r.Cmp(big.NewInt(0)) != 0 {
		// generates random numbers s1, s2 in [0, Curve.Params().N - 1]
		s1 := RandScalar()
		s2 := RandScalar()

		// t = s1*G + s2*H
		t := priKey.publicKey.G.ScalarMult(s1).Add(priKey.publicKey.H.ScalarMult(s2))

		// E is the hash of elliptic point t and data need to be signed
		signature.e = Hash(*t, data)

		signature.z1 = new(big.Int).Sub(s1, new(big.Int).Mul(priKey.secretKey, signature.e))
		signature.z1.Mod(signature.z1, Curve.Params().N)

		signature.z2 = new(big.Int).Sub(s2, new(big.Int).Mul(priKey.r, signature.e))
		signature.z2.Mod(signature.z2, Curve.Params().N)

		return signature, nil
	}

	// generates random numbers s, k2 in [0, Curve.Params().N - 1]
	s := RandScalar()

	// t = s*G
	t := priKey.publicKey.G.ScalarMult(s)

	// E is the hash of elliptic point t and data need to be signed
	signature.e = Hash(*t, data)

	// Z1 = s - e*sk
	signature.z1 = new(big.Int).Sub(s, new(big.Int).Mul(priKey.secretKey, signature.e))
	signature.z1.Mod(signature.z1, Curve.Params().N)

	return signature, nil
}

//Verify is function which using for verify that the given signature was signed by by privatekey of the public key
func (pubKey SchnorrPubKey) Verify(signature *SchnSignature, data []byte) bool {
	if signature == nil {
		return false
	}

	rv := pubKey.G.ScalarMult(signature.z1).Add(pubKey.H.ScalarMult(signature.z2))
	rv = rv.Add(pubKey.PK.ScalarMult(signature.e))

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
}

// Hash calculates a hash concatenating a given message bytes with a given EC Point. H(p||m)
func Hash(p EllipticPoint, m []byte) *big.Int {
	b := append(common.AddPaddingBigInt(p.X, common.BigIntSize), common.AddPaddingBigInt(p.Y, common.BigIntSize)...)
	b = append(b, m...)

	return new(big.Int).SetBytes(common.HashB(b))
}
