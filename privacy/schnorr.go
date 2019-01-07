package privacy

import (
	"errors"
	"github.com/ninjadotorg/constant/common"
	"math/big"
)

//SchnPubKey denoted Schnorr Publickey
type SchnPubKey struct {
	PK, G, H *EllipticPoint // PK = G^SK + H^Randomness
}

//SchnPrivKey denoted Schnorr Privatekey
type SchnPrivKey struct {
	SK, R  *big.Int
	PubKey *SchnPubKey
}

//SchnSignature denoted Schnorr Signature
type SchnSignature struct {
	E, S1, S2 *big.Int
}

//GenKey generates PriKey and PubKey
func (priKey *SchnPrivKey) GenKey() {
	if priKey == nil {
		priKey = new(SchnPrivKey)
	}

	xBytes := RandBytes(SpendingKeySize)
	priKey.SK = new(big.Int).SetBytes(xBytes)
	priKey.SK.Mod(priKey.SK, Curve.Params().N)

	rBytes := RandBytes(SpendingKeySize)
	priKey.R = new(big.Int).SetBytes(rBytes)
	priKey.R.Mod(priKey.R, Curve.Params().N)

	priKey.PubKey = new(SchnPubKey)

	priKey.PubKey.G = new(EllipticPoint)
	priKey.PubKey.G.Set(Curve.Params().Gx, Curve.Params().Gy)

	priKey.PubKey.H = priKey.PubKey.G.ScalarMult(RandInt())
	rH := priKey.PubKey.H.ScalarMult(priKey.R)

	priKey.PubKey.PK = priKey.PubKey.G.ScalarMult(priKey.SK).Add(rH)
}

func (priKey *SchnPrivKey) Set(sk *big.Int, r *big.Int) {
	priKey.SK = sk
	priKey.R = r
	priKey.PubKey = new(SchnPubKey)
	priKey.PubKey.G = new(EllipticPoint)
	priKey.PubKey.G.Set(PedCom.G[SK].X, PedCom.G[SK].Y)

	if r.Cmp(big.NewInt(0)) == 0 {
		priKey.PubKey.PK = PedCom.G[SK].ScalarMult(sk)
	} else {
		priKey.PubKey.H = new(EllipticPoint)
		priKey.PubKey.H.Set(PedCom.G[RAND].X, PedCom.G[RAND].Y)
		priKey.PubKey.PK = PedCom.G[SK].ScalarMult(sk).Add(PedCom.G[RAND].ScalarMult(r))
	}
}

func (pubKey *SchnPubKey) Set(pk *EllipticPoint) {
	pubKey.PK = new(EllipticPoint)
	pubKey.PK.Set(pk.X, pk.Y)

	pubKey.G = new(EllipticPoint)
	pubKey.G.Set(PedCom.G[SK].X, PedCom.G[SK].Y)

	pubKey.H = new(EllipticPoint)
	pubKey.H.Set(PedCom.G[RAND].X, PedCom.G[RAND].Y)
}

//Sign is function which using for sign on hash array by private key
func (priKey SchnPrivKey) Sign(hash []byte) (*SchnSignature, error) {
	if len(hash) != common.HashSize {
		return nil, NewPrivacyErr(UnexpectedErr, errors.New("Hash length must be 32 bytes"))
	}

	genPoint := new(EllipticPoint)
	genPoint.Set(Curve.Params().Gx, Curve.Params().Gy)

	signature := new(SchnSignature)

	// has privacy
	if priKey.R.Cmp(big.NewInt(0)) != 0 {
		// generates random numbers k1, k2 in [0, Curve.Params().N - 1]
		k1 := RandInt()
		k2 := RandInt()

		// t1 = G^k1
		t1 := priKey.PubKey.G.ScalarMult(k1)

		// t2 = H^k2
		t2 := priKey.PubKey.H.ScalarMult(k2)

		// t = t1 + t2
		t := t1.Add(t2)

		// E is the hash of elliptic point t and data need to be signed
		signature.E = Hash(*t, hash)

		// xe = Sk * e
		xe := new(big.Int)
		xe.Mul(priKey.SK, signature.E)

		signature.S1 = new(big.Int).Sub(k1, xe)
		signature.S1.Mod(signature.S1, Curve.Params().N)

		// re = Randomness * e
		re := new(big.Int).Mul(priKey.R, signature.E)

		signature.S2 = new(big.Int).Sub(k2, re)
		signature.S2.Mod(signature.S2, Curve.Params().N)

		return signature, nil
	}

	// generates random numbers k1, k2 in [0, Curve.Params().N - 1]
	k1 := RandInt()

	// t1 = G^k1
	t1 := priKey.PubKey.G.ScalarMult(k1)

	// E is the hash of elliptic point t and data need to be signed
	signature.E = Hash(*t1, hash)

	// xe = Sk * e
	xe := new(big.Int)
	xe.Mul(priKey.SK, signature.E)

	signature.S1 = new(big.Int).Sub(k1, xe)
	signature.S1.Mod(signature.S1, Curve.Params().N)

	return signature, nil
}

//Verify is function which using for verify that the given signature was signed by by privatekey of the public key
func (pub SchnPubKey) Verify(signature *SchnSignature, hash []byte) bool {
	if len(hash) != common.HashSize {
		return false
	}

	if signature == nil {
		return false
	}

	rv := pub.G.ScalarMult(signature.S1).Add(pub.H.ScalarMult(signature.S2))
	rv = rv.Add(pub.PK.ScalarMult(signature.E))

	ev := Hash(*rv, hash)
	if ev.Cmp(signature.E) == 0 {
		return true
	}

	return false
}

func (sig *SchnSignature) Bytes() []byte {
	var bytes []byte
	bytes = append(AddPaddingBigInt(sig.E, BigIntSize), AddPaddingBigInt(sig.S1, BigIntSize)...)
	// S2 is nil when has no privacy
	if sig.S2 != nil {
		bytes = append(bytes, AddPaddingBigInt(sig.S2, BigIntSize)...)
	}
	return bytes
}

func (sig *SchnSignature) SetBytes(bytes []byte) {
	sig.E = new(big.Int).SetBytes(bytes[0:BigIntSize])
	sig.S1 = new(big.Int).SetBytes(bytes[BigIntSize : 2*BigIntSize])
	sig.S2 = new(big.Int).SetBytes(bytes[2*BigIntSize:])
}

// Hash calculates a hash concatenating a given message bytes with a given EC Point. H(p||m)
func Hash(p EllipticPoint, m []byte) *big.Int {
	var b []byte

	b = append(AddPaddingBigInt(p.X, BigIntSize), AddPaddingBigInt(p.Y, BigIntSize)...)
	b = append(b, m...)

	return new(big.Int).SetBytes(common.HashB(b))
}
