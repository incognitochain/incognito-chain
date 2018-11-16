package privacy

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"math/big"
)


type SchnPubKey struct {
	PK, G, H EllipticPoint // PK = G^SK + H^R
}

type SchnPrivKey struct {
	SK, R *big.Int
	PK    SchnPubKey
}

type SchnSignature struct {
	E, S1, S2 *big.Int
}

// GenSchnPrivKey generates Schnorr private key
func SchnGenPrivKey() *SchnPrivKey {
	priv := new(SchnPrivKey)
	xBytes := RandBytes(32)
	priv.SK = new(big.Int).SetBytes(xBytes)
	priv.SK.Mod(priv.SK, Curve.Params().N)

	rBytes := RandBytes(32)
	priv.R = new(big.Int).SetBytes(rBytes)
	priv.R.Mod(priv.R, Curve.Params().N)
	priv.PK = *SchnGenPubKey(*priv)

	return priv
}

func SchnGenPubKey(priv SchnPrivKey) *SchnPubKey {
	pub := new(SchnPubKey)

	pub.G = *new(EllipticPoint)
	pub.G.X = Curve.Params().Gx
	pub.G.Y = Curve.Params().Gy

	pub.H = *new(EllipticPoint)
	pub.H.X, pub.H.Y = Curve.ScalarBaseMult(RandBytes(32))
	rH := new(EllipticPoint)
	rH.X, rH.Y = Curve.ScalarMult(pub.H.X, pub.H.Y, priv.R.Bytes())

	pub.PK = *new(EllipticPoint)
	pub.PK.X, pub.PK.Y = Curve.ScalarBaseMult(priv.SK.Bytes())
	pub.PK.X, pub.PK.Y = Curve.Add(pub.PK.X, pub.PK.Y, rH.X, rH.Y)

	return pub
}

func SchnSign(hash []byte, priv SchnPrivKey) (*SchnSignature, error) {
	if len(hash) != 32 {
		return nil, errors.New("Hash length must be 32 bytes")
	}

	signature := new(SchnSignature)

	k1Bytes := RandBytes(32)
	k1 := new(big.Int).SetBytes(k1Bytes)
	k1.Mod(k1, Curve.Params().N)

	k2Bytes := RandBytes(32)
	k2 := new(big.Int).SetBytes(k2Bytes)
	k2.Mod(k2, Curve.Params().N)

	t1 := new(EllipticPoint)
	t1.X, t1.Y = Curve.ScalarMult(priv.PK.G.X, priv.PK.G.Y, k1.Bytes())

	t2 := new(EllipticPoint)
	t2.X, t2.Y = Curve.ScalarMult(priv.PK.H.X, priv.PK.H.Y, k2.Bytes())

	t := new(EllipticPoint)
	t.X, t.Y = Curve.Add(t1.X, t1.Y, t2.X, t2.Y)

	signature.E = Hash(*t, hash)

	xe := new(big.Int)
	xe.Mul(priv.SK, signature.E)
	signature.S1 = new(big.Int)
	signature.S1.Sub(k1, xe)
	signature.S1.Mod(signature.S1, Curve.Params().N)

	re := new(big.Int)
	re.Mul(priv.R, signature.E)
	signature.S2 = new(big.Int)
	signature.S2.Sub(k2, re)
	signature.S2.Mod(signature.S2, Curve.Params().N)

	return signature, nil
}

func SchnVerify(signature *SchnSignature, hash []byte, pub SchnPubKey) bool {
	if len(hash) != 32 {
		return false
	}

	if signature == nil {
		return false
	}

	rv := new(EllipticPoint)
	rv.X, rv.Y = Curve.ScalarMult(pub.G.X, pub.G.Y, signature.S1.Bytes())
	tmp := new(EllipticPoint)
	tmp.X, tmp.Y = Curve.ScalarMult(pub.H.X, pub.H.Y, signature.S2.Bytes())
	rv.X, rv.Y = Curve.Add(rv.X, rv.Y, tmp.X, tmp.Y)
	tmp.X, tmp.Y = Curve.ScalarMult(pub.PK.X, pub.PK.Y, signature.E.Bytes())
	rv.X, rv.Y = Curve.Add(rv.X, rv.Y, tmp.X, tmp.Y)

	ev := Hash(*rv, hash)
	if ev.Cmp(signature.E) == 0 {
		return true
	}

	return false
}

// Hash calculates a hash concatenating a given message bytes with a given EC Point. H(p||m)
func Hash(p EllipticPoint, m []byte) *big.Int {
	var b []byte
	cXBytes := p.X.Bytes()
	cYBytes := p.Y.Bytes()
	b = append(b, cXBytes...)
	b = append(b, cYBytes...)
	b = append(b, m...)
	h := sha256.New()
	h.Write(b)
	hash := h.Sum(nil)
	r := new(big.Int).SetBytes(hash)
	return r
}

func TestSchn() {
	priv := SchnGenPrivKey()

	hash := RandBytes(32)
	fmt.Printf("Hash: %v\n", hash)

	signature, _ := SchnSign(hash, *priv)
	fmt.Printf("Signature: %+v\n", signature)

	res := SchnVerify(signature, hash, priv.PK)
	fmt.Println(res)

}
