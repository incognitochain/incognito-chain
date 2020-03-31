// MAIN IMPLEMENTATION OF MLSAG

package mlsag

import (
	"crypto/sha256"
	"errors"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/privacy/operation"
	C25519 "github.com/incognitochain/incognito-chain/privacy/operation/curve25519"
)

var CurveOrder *operation.Scalar = new(operation.Scalar).SetKeyUnsafe(&C25519.L)

type Ring struct {
	keys [][]*operation.Point
}

func NewRing(keys [][]*operation.Point) *Ring {
	return &Ring{keys}
}

func (ring Ring) ToBytes() ([]byte, error) {
	k := ring.keys
	if len(k) == 0 {
		return nil, errors.New("RingToBytes: Ring is empty")
	}
	// Make sure that the ring size is a rectangle row*column
	for i := 1; i < len(k); i += 1 {
		if len(k[i]) != len(k[0]) {
			return nil, errors.New("RingToBytes: Ring is not a proper rectangle row*column")
		}
	}
	n := len(k)
	m := len(k[0])
	b := make([]byte, 2)
	if n > 255 || m > 255 {
		return nil, errors.New("RingToBytes: Ring size is too large")
	}
	b[0] = byte(n)
	b[1] = byte(m)
	for i := 0; i < n; i += 1 {
		for j := 0; j < m; j += 1 {
			b = append(b, k[i][j].ToBytesS()...)
		}
	}
	return b, nil
}

func (ring *Ring) FromBytes(b []byte) (*Ring, error) {
	if len(b) < 2 {
		return nil, errors.New("RingFromBytes: byte length is too short")
	}
	n := int(b[0])
	m := int(b[1])
	if len(b) != operation.Ed25519KeySize*n*m+2 {
		return nil, errors.New("RingFromBytes: byte length is not correct")
	}
	offset := 2
	key := make([][]*operation.Point, 0)
	for i := 0; i < n; i += 1 {
		curRow := make([]*operation.Point, m)
		for j := 0; j < m; j += 1 {
			currentByte := b[offset : offset+operation.Ed25519KeySize]
			currentPoint, err := new(operation.Point).FromBytesS(currentByte)
			if err != nil {
				return nil, errors.New("RingFromBytes: byte contains incorrect point")
			}
			curRow = append(curRow, currentPoint)
		}
		key = append(key, curRow)
	}
	ring = NewRing(key)
	return ring, nil
}

type Mlsag struct {
	R           *Ring
	pi          int
	keyImages   []*operation.Point
	privateKeys []*operation.Scalar
}

func NewMlsag(privateKeys []*operation.Scalar, R *Ring, pi int) *Mlsag {
	return &Mlsag{
		R,
		pi,
		parseKeyImages(privateKeys),
		privateKeys,
	}
}

// Parse public key from private key
func parsePublicKey(privateKey *operation.Scalar) *operation.Point {
	return new(operation.Point).ScalarMultBase(privateKey)
}

func parseKeyImages(privateKeys []*operation.Scalar) []*operation.Point {
	m := len(privateKeys)

	result := make([]*operation.Point, m)
	for i := 0; i < m; i += 1 {
		publicKey := parsePublicKey(privateKeys[i])
		hashPoint := operation.HashToPoint(publicKey.ToBytesS())
		result[i] = new(operation.Point).ScalarMult(hashPoint, privateKeys[i])
	}
	return result
}

func (this *Mlsag) createRandomChallenges() (alpha []*operation.Scalar, r [][]*operation.Scalar) {
	m := len(this.privateKeys)
	n := len(this.R.keys)

	alpha = make([]*operation.Scalar, m)
	for i := 0; i < m; i += 1 {
		alpha[i] = operation.RandomScalar()
	}
	r = make([][]*operation.Scalar, n)
	for i := 0; i < n; i += 1 {
		r[i] = make([]*operation.Scalar, m)
		if i == this.pi {
			continue
		}
		for j := 0; j < m; j += 1 {
			r[i][j] = operation.RandomScalar()
		}
	}
	return
}

func calculateFirstC(digest [sha256.Size]byte, alpha []*operation.Scalar, K []*operation.Point) (*operation.Scalar, error) {
	if len(alpha) != len(K) {
		Logger.log.Error("Calculating first C must have length of alpha be the same with length of ring R")
		return nil, errors.New("Error in MLSAG: Calculating first C must have length of alpha be the same with length of ring R")
	}
	var b []byte
	b = append(b, digest[:]...)
	for i := 0; i < len(K); i += 1 {
		alphaG := new(operation.Point).ScalarMultBase(alpha[i])

		H := operation.HashToPoint(K[i].ToBytesS())
		alphaH := new(operation.Point).ScalarMult(H, alpha[i])

		b = append(b, alphaG.ToBytesS()...)
		b = append(b, alphaH.ToBytesS()...)
	}
	return operation.HashToScalar(b), nil
}

func calculateNextC(digest [sha256.Size]byte, r []*operation.Scalar, c *operation.Scalar, K []*operation.Point, keyImages []*operation.Point) (*operation.Scalar, error) {
	if len(r) != len(K) || len(r) != len(keyImages) {
		Logger.log.Error("Calculating next C must have length of r be the same with length of ring R and same with length of keyImages")
		return nil, errors.New("Error in MLSAG: Calculating next C must have length of r be the same with length of ring R and same with length of keyImages")
	}
	var b []byte
	b = append(b, digest[:]...)

	// Below is the mathematics within the Monero paper:
	// If you are reviewing my code, please refer to paper
	// rG: r*G
	// cK: c*R
	// rG_cK: rG + cK
	//
	// HK: H_p(K_i)
	// rHK: r_i*H_p(K_i)
	// cKI: c*R~ (KI as keyImage)
	// rHK_cKI: rHK + cKI
	for i := 0; i < len(K); i += 1 {
		rG := new(operation.Point).ScalarMultBase(r[i])
		cK := new(operation.Point).ScalarMult(K[i], c)
		rG_cK := new(operation.Point).Add(rG, cK)

		HK := operation.HashToPoint(K[i].ToBytesS())
		rHK := new(operation.Point).ScalarMult(HK, r[i])
		cKI := new(operation.Point).ScalarMult(keyImages[i], c)
		rHK_cKI := new(operation.Point).Add(rHK, cKI)

		b = append(b, rG_cK.ToBytesS()...)
		b = append(b, rHK_cKI.ToBytesS()...)
	}
	return operation.HashToScalar(b), nil
}

func (this *Mlsag) calculateC(digest [HashSize]byte, alpha []*operation.Scalar, r [][]*operation.Scalar) ([]*operation.Scalar, error) {
	m := len(this.privateKeys)
	n := len(this.R.keys)

	c := make([]*operation.Scalar, n)
	firstC, err := calculateFirstC(
		digest,
		alpha,
		this.R.keys[this.pi],
	)
	if err != nil {
		return nil, err
	}

	var i int = (this.pi + 1) % n
	c[i] = firstC
	for next := (i + 1) % n; i != this.pi; {
		nextC, err := calculateNextC(
			digest,
			r[i], c[i],
			(*this.R).keys[i],
			this.keyImages,
		)
		if err != nil {
			return nil, err
		}
		c[next] = nextC
		i = next
		next = (next + 1) % n
	}

	for i := 0; i < m; i += 1 {
		ck := new(operation.Scalar).Mul(c[this.pi], this.privateKeys[i])
		r[this.pi][i] = new(operation.Scalar).Sub(alpha[i], ck)
	}

	return c, nil
}

// check l*KI = 0 by checking KI is a valid point
func verifyKeyImages(keyImages []*operation.Point) bool {
	var check bool = true
	for i := 0; i < len(keyImages); i += 1 {
		lKI := new(operation.Point).ScalarMult(keyImages[i], CurveOrder)
		check = check && lKI.IsIdentity()
	}
	return check
}

func verifyRing(sig *MlsagSig, R *Ring, message []byte) (bool, error) {
	digest := common.Keccak256(message)
	c := sig.c
	cBefore := sig.c
	for i := 0; i < len(sig.r); i += 1 {
		nextC, err := calculateNextC(
			digest,
			sig.r[i], &c,
			R.keys[i],
			sig.keyImages,
		)
		if err != nil {
			return false, err
		}
		c = *nextC
	}
	return c == cBefore, nil
}

func Verify(sig *MlsagSig, K *Ring, message []byte) (bool, error) {
	b1 := verifyKeyImages(sig.keyImages)
	b2, err := verifyRing(sig, K, message)
	return (b1 && b2), err
}

func (this *Mlsag) Sign(message []byte) (*MlsagSig, error) {
	digest := common.Keccak256(message)
	alpha, r := this.createRandomChallenges()   // step 2 in paper
	c, err := this.calculateC(digest, alpha, r) // step 3 and 4 in paper

	if err != nil {
		return nil, err
	}
	return &MlsagSig{
		*c[0], this.keyImages, r,
	}, nil
}
