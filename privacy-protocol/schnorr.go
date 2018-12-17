package privacy

import (
	"crypto/sha256"
	"errors"
	"fmt"
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
	xBytes := RandBytes(32)
	priKey.SK = new(big.Int).SetBytes(xBytes)
	priKey.SK.Mod(priKey.SK, Curve.Params().N)

	rBytes := RandBytes(32)
	priKey.R = new(big.Int).SetBytes(rBytes)
	priKey.R.Mod(priKey.R, Curve.Params().N)

	priKey.PubKey = new(SchnPubKey)

	priKey.PubKey.G = new(EllipticPoint)
	priKey.PubKey.G.X, priKey.PubKey.G.Y = Curve.Params().Gx, Curve.Params().Gy

	priKey.PubKey.H = new(EllipticPoint)
	priKey.PubKey.H.X, priKey.PubKey.H.Y = Curve.ScalarBaseMult(RandBytes(32))
	rH := new(EllipticPoint)
	rH.X, rH.Y = Curve.ScalarMult(priKey.PubKey.H.X, priKey.PubKey.H.Y, priKey.R.Bytes())

	priKey.PubKey.PK = new(EllipticPoint)
	priKey.PubKey.PK.X, priKey.PubKey.PK.Y = Curve.ScalarBaseMult(priKey.SK.Bytes())
	priKey.PubKey.PK.X, priKey.PubKey.PK.Y = Curve.Add(priKey.PubKey.PK.X, priKey.PubKey.PK.Y, rH.X, rH.Y)
}

//Sign is function which using for sign on hash array by private key
func (priKey SchnPrivKey) Sign(hash []byte) (*SchnSignature, error) {
	if len(hash) != 32 {
		return nil, errors.New("Hash length must be 32 bytes")
	}

	genPoint := *new(EllipticPoint)
	genPoint.X = Curve.Params().Gx
	genPoint.Y = Curve.Params().Gy

	signature := new(SchnSignature)

	// generates random numbers k1, k2 in [0, Curve.Params().N - 1]
	k1Bytes := RandBytes(32)
	k1 := new(big.Int).SetBytes(k1Bytes)
	k1.Mod(k1, Curve.Params().N)

	k2Bytes := RandBytes(32)
	k2 := new(big.Int).SetBytes(k2Bytes)
	k2.Mod(k2, Curve.Params().N)

	// t1 = G^k1
	t1 := new(EllipticPoint)
	t1.X, t1.Y = Curve.ScalarMult(priKey.PubKey.G.X, priKey.PubKey.G.Y, k1.Bytes())

	// t2 = H^k2
	t2 := new(EllipticPoint)
	t2.X, t2.Y = Curve.ScalarMult(priKey.PubKey.H.X, priKey.PubKey.H.Y, k2.Bytes())

	// t = t1 + t2
	t := new(EllipticPoint)
	t.X, t.Y = Curve.Add(t1.X, t1.Y, t2.X, t2.Y)

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
	if len(hash) != 32 {
		return false
	}

	if signature == nil {
		return false
	}

	fmt.Printf("VERIFY 2 ------ PUBLICKEY: %+v\n", pub.PK)

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

func (sig *SchnSignature) ToBytes() []byte {
	temp := sig.E.Bytes()
	for i:=0; i<BigIntSize-len(temp); i++{
		temp = append([]byte{0}, temp...)
	}
	res:=temp
	temp = sig.S1.Bytes()
	for i:=0; i<BigIntSize-len(temp); i++{
		temp = append([]byte{0}, temp...)
	}
	res = append(res, temp...)
	temp = sig.S2.Bytes()
	for i:=0; i<BigIntSize-len(temp); i++{
		temp = append([]byte{0}, temp...)
	}
	res = append(res, temp...)
	//res := append(sig.E.Bytes(), sig.S1.Bytes()...)
	//res = append(res, sig.S2.Bytes()...)
	return res
}

func (sig *SchnSignature) FromBytes(bytes []byte) {
	sig.E = new(big.Int).SetBytes(bytes[0:32])
	sig.S1 = new(big.Int).SetBytes(bytes[32:64])
	sig.S2 = new(big.Int).SetBytes(bytes[64:96])
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
	schnPrivKey := new(SchnPrivKey)
	schnPrivKey.KeyGen()

	hash := RandBytes(32)
	fmt.Printf("Hash: %v\n", hash)

	signature, _ := schnPrivKey.Sign(hash)
	fmt.Printf("Signature: %+v\n", signature)

	res := schnPrivKey.PubKey.Verify(signature, hash)
	fmt.Println(res)

}
