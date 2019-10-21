package bls

import (
	"crypto/sha256"
	"crypto/subtle"
	"github.com/go-ethereum-master/crypto/bn256/cloudflare"
	"io"
	"math/big"
)

type PrivateKey struct {
	x *big.Int
}

type PublicKey struct {
	gx *bn256.G2
}

type Signature struct {
	s *bn256.G1
}

// elliptic curve y²=x³+3
var g2gen = new(bn256.G2).ScalarBaseMult(big.NewInt(1))
var g1gen = new(bn256.G1).ScalarBaseMult(big.NewInt(1))
var curveB = new(big.Int).SetInt64(3)

func GenerateKey(rand io.Reader) (*PublicKey, *PrivateKey, error) {
	x, gx, err := bn256.RandomG2(rand)
	if err != nil {
		return nil, nil, err
	}

	return &PublicKey{gx}, &PrivateKey{x}, nil
}

// DefendRogueAttack returns a vector t that is used as random numbers a_i
func DefendRogueAttack(pubKeys []*PublicKey) []*big.Int {
	var bGroupPubKey []byte
	for _, pub := range pubKeys {
		bGroupPubKey = append(bGroupPubKey, pub.gx.Marshal()...)
	}

	var t []*big.Int
	for i := 0; i < len(pubKeys); i++ {
		t = append(t, HashToScalar(append(pubKeys[i].gx.Marshal(), bGroupPubKey...)))
	}
	return t
}

// Sign returns a signature h^x \in G1
func Sign(privateKey *PrivateKey, message []byte) *Signature {
	h := HashToG1(message)
	hx := new(bn256.G1).ScalarMult(h, privateKey.x)
	return &Signature{hx}
}

// Verify verifies a signature.  Returns false if sig is not a valid signature.
func Verify(key *PublicKey, message []byte, sig *Signature) bool {
	h := HashToG1(message)

	u := bn256.Pair(h, key.gx)
	v := bn256.Pair(sig.s, g2gen)
	return subtle.ConstantTimeCompare(u.Marshal(), v.Marshal()) == 1
}

func VerifyCompressSig(key *PublicKey, message []byte, compressedSig []byte) bool {
	s, err := DecompressG1(compressedSig)
	if err != nil {
		return false
	}

	return Verify(key, message, &Signature{s})
}

// Aggregate combines signatures on distinct messages.  The messages must
// be distinct, otherwise the scheme is vulnerable to chosen-key attack.
func Aggregate(sigs []*Signature, pubKeys []*PublicKey) *Signature {
	pointLs := make([]*bn256.G1, len(sigs))

	for i := 0; i < len(sigs); i++ {
		pointLs[i] = new(bn256.G1).Set(sigs[i].s)
	}
	t := DefendRogueAttack(pubKeys)

	sigma := MultiScalarMultG1(pointLs, t)

	return &Signature{sigma}
}

// Verify verifies an aggregate signature.  Returns false if sig is not a valid signature.
func VerifyAgg(pubKeys []*PublicKey, message []byte, sig *Signature) bool {
	pointLs := make([]*bn256.G2, len(pubKeys))

	for i := 0; i < len(pubKeys); i++ {
		pointLs[i] = new(bn256.G2).Set(pubKeys[i].gx)
	}
	t := DefendRogueAttack(pubKeys)

	apk := MultiScalarMultG2(pointLs, t)

	h := HashToG1(message)
	u := bn256.Pair(h, apk)
	v := bn256.Pair(sig.s, g2gen)

	return subtle.ConstantTimeCompare(u.Marshal(), v.Marshal()) == 1
}

func VerifyAggCompressSig(pubKeys []*PublicKey, message []byte, compressedSig []byte) bool {
	s, err := DecompressG1(compressedSig)
	if err != nil {
		return false
	}
	return VerifyAgg(pubKeys, message, &Signature{s})
}

func BatchVerifyDistinct(pubKeys []*PublicKey, messages [][]byte, sigs []*Signature) bool {

	if len(pubKeys) != len(messages) || len(messages) != len(sigs) {
		return false
	}

	if !distinct(messages) {
		return false
	}
	/*
		aggSig := new(bn256.G1).Set(sigs[0].s)
		aggPub := new(bn256.G2).Set(pubKeys[0].gx)
		aggMsg := new(bn256.G1).Set(HashToG1(messages[0]))

		for i:= 1; i< len(messages); i++ {
			aggSig.Add(aggSig, sigs[i].s)
			aggPub.Add(aggPub, pubKeys[i].gx)
			aggMsg.Add(aggMsg, HashToG1(messages[i]))
		}
		u := bn256.Pair(aggMsg, aggPub)
		v := bn256.Pair(aggSig, g2gen)
	*/

	aggSig := new(bn256.G1).Set(sigs[0].s)
	u := bn256.Pair(HashToG1(messages[0]), pubKeys[0].gx)
	for i := 1; i < len(messages); i++ {
		aggSig.Add(aggSig, sigs[i].s)
		u.Add(u, bn256.Pair(HashToG1(messages[i]), pubKeys[i].gx))
	}
	v := bn256.Pair(aggSig, g2gen)

	return subtle.ConstantTimeCompare(u.Marshal(), v.Marshal()) == 1
}

func distinct(msgs [][]byte) bool {
	m := make(map[[32]byte]bool)
	for _, msg := range msgs {
		h := sha256.Sum256(msg)
		if m[h] {
			return false
		}
		m[h] = true
	}
	return true
}
