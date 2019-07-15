package privacy

import (
	"errors"
	"math/big"

	"github.com/incognitochain/incognito-chain/common"
)

//SchnPubKey represents Schnorr Publickey
// PK = G^SK + H^R
type SchnPubKey struct {
	PK, G, H *EllipticPoint
}

//SchnPrivKey represents Schnorr Privatekey
type SchnPrivKey struct {
	SK, R  *big.Int
	PubKey *SchnPubKey
}

//SchnSignature represents Schnorr Signature
type SchnSignature struct {
	E, Z1, Z2 *big.Int
}

// Set sets Schnorr private key
func (priKey *SchnPrivKey) Set(sk *big.Int, r *big.Int) {
	priKey.SK = sk
	priKey.R = r
	priKey.PubKey = new(SchnPubKey)
	priKey.PubKey.G = new(EllipticPoint)
	priKey.PubKey.G.Set(PedCom.G[SK].X, PedCom.G[SK].Y)

	priKey.PubKey.H = new(EllipticPoint)
	priKey.PubKey.H.Set(PedCom.G[RAND].X, PedCom.G[RAND].Y)
	priKey.PubKey.PK = PedCom.G[SK].ScalarMult(sk).Add(PedCom.G[RAND].ScalarMult(r))
}

// Set sets Schnorr public key
func (pubKey *SchnPubKey) Set(pk *EllipticPoint) {
	pubKey.PK = new(EllipticPoint)
	pubKey.PK.Set(pk.X, pk.Y)

	pubKey.G = new(EllipticPoint)
	pubKey.G.Set(PedCom.G[SK].X, PedCom.G[SK].Y)

	pubKey.H = new(EllipticPoint)
	pubKey.H.Set(PedCom.G[RAND].X, PedCom.G[RAND].Y)
}

//Sign is function which using for signing on hash array by private key
func (priKey SchnPrivKey) Sign(data []byte) (*SchnSignature, error) {
	if len(data) != common.HashSize {
		return nil, NewPrivacyErr(UnexpectedErr, errors.New("hash length must be 32 bytes"))
	}

	signature := new(SchnSignature)

	// has privacy
	if priKey.R.Cmp(big.NewInt(0)) != 0 {
		// generates random numbers s1, s2 in [0, Curve.Params().N - 1]
		s1 := RandScalar()
		s2 := RandScalar()

		// t = s1*G + s2*H
		t := priKey.PubKey.G.ScalarMult(s1).Add(priKey.PubKey.H.ScalarMult(s2))

		// E is the hash of elliptic point t and data need to be signed
		signature.E = Hash(*t, data)

		signature.Z1 = new(big.Int).Sub(s1, new(big.Int).Mul(priKey.SK, signature.E))
		signature.Z1.Mod(signature.Z1, Curve.Params().N)

		signature.Z2 = new(big.Int).Sub(s2, new(big.Int).Mul(priKey.R, signature.E))
		signature.Z2.Mod(signature.Z2, Curve.Params().N)

		return signature, nil
	}

	// generates random numbers s, k2 in [0, Curve.Params().N - 1]
	s := RandScalar()

	// t = s*G
	t := priKey.PubKey.G.ScalarMult(s)

	// E is the hash of elliptic point t and data need to be signed
	signature.E = Hash(*t, data)

	// Z1 = s - e*sk
	signature.Z1 = new(big.Int).Sub(s, new(big.Int).Mul(priKey.SK, signature.E))
	signature.Z1.Mod(signature.Z1, Curve.Params().N)

	return signature, nil
}

//Verify is function which using for verify that the given signature was signed by by privatekey of the public key
func (pubKey SchnPubKey) Verify(signature *SchnSignature, data []byte) bool {
	if signature == nil {
		return false
	}

	rv := pubKey.G.ScalarMult(signature.Z1).Add(pubKey.H.ScalarMult(signature.Z2))
	rv = rv.Add(pubKey.PK.ScalarMult(signature.E))

	ev := Hash(*rv, data)
	return ev.Cmp(signature.E) == 0
}

func (sig *SchnSignature) Bytes() []byte {
	bytes := append(AddPaddingBigInt(sig.E, BigIntSize), AddPaddingBigInt(sig.Z1, BigIntSize)...)
	// Z2 is nil when has no privacy
	if sig.Z2 != nil {
		bytes = append(bytes, AddPaddingBigInt(sig.Z2, BigIntSize)...)
	}
	return bytes
}

func (sig *SchnSignature) SetBytes(bytes []byte) {
	sig.E = new(big.Int).SetBytes(bytes[0:BigIntSize])
	sig.Z1 = new(big.Int).SetBytes(bytes[BigIntSize : 2*BigIntSize])
	sig.Z2 = new(big.Int).SetBytes(bytes[2*BigIntSize:])
}

// Hash calculates a hash concatenating a given message bytes with a given EC Point. H(p||m)
func Hash(p EllipticPoint, m []byte) *big.Int {
	b := append(AddPaddingBigInt(p.X, BigIntSize), AddPaddingBigInt(p.Y, BigIntSize)...)
	b = append(b, m...)

	return new(big.Int).SetBytes(common.HashB(b))
}
