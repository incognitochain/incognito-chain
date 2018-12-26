package privacy

import (
	"errors"
	"github.com/ninjadotorg/constant/common"
	"math/big"
)

//1

//SignScheme contains some algorithms for digital signature scheme
type SignScheme interface {
	KeyGen()                //Generate PriKey and PubKey
	GetPubkey() *SchnPubKey //return Publickey belong to the PrivateKey
	Sign(hash []byte) (*SchnSignature, error)
	Verify(signature *SchnSignature, hash []byte) bool
}

/*---------------------------------------------------------------------------------------------------------*/

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

//KeyGen generates PriKey and PubKey
func (priKey *SchnPrivKey) KeyGen() {
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

	priKey.PubKey.H = new(EllipticPoint)
	priKey.PubKey.H.X, priKey.PubKey.H.Y = Curve.ScalarBaseMult(RandBytes(SpendingKeySize))
	rH := priKey.PubKey.H.ScalarMult(priKey.R)

	priKey.PubKey.PK = priKey.PubKey.G.ScalarMult(priKey.SK).Add(rH)
}

//Sign is function which using for sign on hash array by private key
func (priKey SchnPrivKey) Sign(hash []byte) (*SchnSignature, error) {
	if len(hash) != common.HashSize {
		return nil, NewPrivacyErr(UnexpectedErr, errors.New("Hash length must be 32 bytes"))
	}

	genPoint := new(EllipticPoint)
	genPoint.Set(Curve.Params().Gx, Curve.Params().Gy)

	signature := new(SchnSignature)

	// generates random numbers k1, k2 in [0, Curve.Params().N - 1]
	k1 := new(big.Int).SetBytes(RandBytes(SpendingKeySize))
	k1.Mod(k1, Curve.Params().N)

	k2 := new(big.Int).SetBytes(RandBytes(SpendingKeySize))
	k2.Mod(k2, Curve.Params().N)

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

	signature.S1 = new(big.Int)
	signature.S1.Sub(k1, xe)
	signature.S1.Mod(signature.S1, Curve.Params().N)

	// re = Randomness * e
	re := new(big.Int)
	re.Mul(priKey.R, signature.E)

	signature.S2 = new(big.Int)
	signature.S2.Sub(k2, re)
	signature.S2.Mod(signature.S2, Curve.Params().N)

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

	rv := pub.G.ScalarMult(signature.S1)
	rv = rv.Add(pub.H.ScalarMult(signature.S2))
	rv = rv.Add(pub.PK.ScalarMult(signature.E))

	ev := Hash(*rv, hash)
	if ev.Cmp(signature.E) == 0 {
		return true
	}

	return false
}

func (sig *SchnSignature) Bytes() []byte {
	temp := sig.E.Bytes()
	for i := 0; i < BigIntSize-len(temp); i++ {
		temp = append([]byte{0}, temp...)
	}
	res := temp
	temp = sig.S1.Bytes()
	for i := 0; i < BigIntSize-len(temp); i++ {
		temp = append([]byte{0}, temp...)
	}
	res = append(res, temp...)
	temp = sig.S2.Bytes()
	for i := 0; i < BigIntSize-len(temp); i++ {
		temp = append([]byte{0}, temp...)
	}
	res = append(res, temp...)
	return res
}

func (sig *SchnSignature) SetBytes(bytes []byte) {
	sig.E = new(big.Int).SetBytes(bytes[0:32])
	sig.S1 = new(big.Int).SetBytes(bytes[32:64])
	sig.S2 = new(big.Int).SetBytes(bytes[64:96])
}

// Hash calculates a hash concatenating a given message bytes with a given EC Point. H(p||m)
func Hash(p EllipticPoint, m []byte) *big.Int {
	var b []byte

	b = append(p.X.Bytes(), p.Y.Bytes()...)
	b = append(b, m...)

	return new(big.Int).SetBytes(common.HashB(b))
}


