// MAIN IMPLEMENTATION OF MLSAG

package mlsag

import (
	"crypto/sha256"
	"errors"

	"github.com/incognitochain/incognito-chain/privacy"
)

type Mlsag struct {
	K           *Ring
	pi          int
	keyImages   []privacy.Point
	privateKeys []privacy.Scalar
}

func (this *Mlsag) IsEmpty() bool {
	return this.K.IsEmpty() || len(this.privateKeys) == 0
}

func (this *Mlsag) createRandomChallenges() (alpha []privacy.Scalar, r [][]privacy.Scalar) {
	m := len(this.privateKeys)
	n := len(this.K.keys)

	alpha = make([]privacy.Scalar, m)
	for i := 0; i < m; i += 1 {
		alpha[i] = *privacy.RandomScalar()
	}
	r = make([][]privacy.Scalar, n)
	for i := 0; i < n; i += 1 {
		r[i] = make([]privacy.Scalar, m)
		if i == this.pi {
			continue
		}
		for j := 0; j < m; j += 1 {
			r[i][j] = *privacy.RandomScalar()
		}
	}
	return
}

func NewMlsagWithDefinedRing(privateKeys []privacy.Scalar, K *Ring, pi, numFake int) (mlsag *Mlsag) {
	if mlsag == nil {
		mlsag = new(Mlsag)
	}

	mlsag.K = K
	mlsag.pi = pi
	mlsag.privateKeys = privateKeys
	mlsag.keyImages = parseKeyImages(privateKeys) // 1st step in monero paper

	return
}

// func NewMlsagWithRandomRing(privateKeys []privacy.Scalar, numFake int) (mlsag *Mlsag) {
// 	if mlsag == nil {
// 		mlsag = new(Mlsag)
// 	}

// 	mlsag.privateKeys = privateKeys
// 	mlsag.pi = common.RandInt() % numFake
// 	mlsag.keyImages = parseKeyImages(privateKeys) // 1st step in monero paper
// 	mlsag.K = NewRandomRing(privateKeys, numFake, mlsag.pi)
// 	return
// }

func SignWithMlsag(mlsag Mlsag, message string) (*Signature, error) {
	return mlsag.Sign(message)
}

// func SignWithOneKey(privateKey privacy.Scalar, message string, numFake int) (*Signature, error) {
// 	keys := []privacy.Scalar{privateKey}
// 	mlsag := NewMlsagWithRandomRing(keys, numFake)
// 	return mlsag.Sign(message)
// }

// // SignWithRandomRing will generate random ring and use MLSAG to sign the message.
// func SignWithRandomRing(privateKeys []privacy.Scalar, message string, numFake int) *Signature {
// 	mlsag := NewMlsagWithRandomRing(privateKeys, numFake)
// 	return mlsag.Sign(message)
// }

func calculateFirstC(digest [sha256.Size]byte, alpha []privacy.Scalar, K []privacy.Point) (*privacy.Scalar, error) {
	if len(alpha) != len(K) {
		return nil, errors.New("Error in MLSAG: Calculating first C must have length of alpha be the same with length of ring K")
	}
	var b []byte
	b = append(b, digest[:]...)
	for i := 0; i < len(K); i += 1 {
		alphaG := new(privacy.Point).ScalarMultBase(&alpha[i])

		H := privacy.HashToPoint(K[i].ToBytesS())
		alphaH := new(privacy.Point).ScalarMult(H, &alpha[i])

		b = append(b, alphaG.ToBytesS()...)
		b = append(b, alphaH.ToBytesS()...)
	}

	if len(b) != 32*(1+len(alpha)*2) {
		return nil, errors.New("Error in MLSAG: Something is wrong while calculating first C")
	}

	return privacy.HashToScalar(b), nil
}

func calculateNextC(digest [sha256.Size]byte, r []privacy.Scalar, c *privacy.Scalar, K []privacy.Point, keyImages []privacy.Point) (*privacy.Scalar, error) {
	if len(r) != len(K) || len(r) != len(keyImages) {
		return nil, errors.New("Error in MLSAG: Calculating next C must have length of r be the same with length of ring K and same with length of keyImages")
	}
	var b []byte
	b = append(b, digest[:]...)

	// Below is the mathematics within the Monero paper:
	// If you are reviewing my code, please refer to paper
	// rG: r*G
	// cK: c*K
	// rG_cK: rG + cK
	//
	// HK: H_p(K_i)
	// rHK: r_i*H_p(K_i)
	// cKI: c*K~ (KI as keyImage)
	// rHK_cKI: rHK + cKI
	for i := 0; i < len(K); i += 1 {
		rG := new(privacy.Point).ScalarMultBase(&r[i])
		cK := new(privacy.Point).ScalarMult(&K[i], c)
		rG_cK := new(privacy.Point).Add(rG, cK)

		HK := privacy.HashToPoint(K[i].ToBytesS())
		rHK := new(privacy.Point).ScalarMult(HK, &r[i])
		cKI := new(privacy.Point).ScalarMult(&keyImages[i], c)
		rHK_cKI := new(privacy.Point).Add(rHK, cKI)

		b = append(b, rG_cK.ToBytesS()...)
		b = append(b, rHK_cKI.ToBytesS()...)
	}

	if len(b) != 32*(1+len(K)*2) {
		return nil, errors.New("Error in MLSAG: Something is wrong while calculating next C")
	}

	return privacy.HashToScalar(b), nil
}

func (this *Mlsag) calculateC(digest [sha256.Size]byte, alpha []privacy.Scalar, r *[][]privacy.Scalar) ([]*privacy.Scalar, error) {
	m := len(this.privateKeys)
	n := len(this.K.keys)

	c := make([]*privacy.Scalar, n)
	firstC, err := calculateFirstC(
		digest,
		alpha,
		this.K.keys[this.pi],
	)
	if err != nil {
		return nil, err
	}

	var i int = (this.pi + 1) % n
	c[i] = firstC
	for next := (i + 1) % n; i != this.pi; {
		nextC, err := calculateNextC(
			digest,
			(*r)[i], c[i],
			(*this.K).keys[i],
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
		ck := new(privacy.Scalar).Mul(c[this.pi], &this.privateKeys[i])
		(*r)[this.pi][i] = *new(privacy.Scalar).Sub(&alpha[i], ck)
	}

	return c, nil
}

func (this *Mlsag) Sign(message string) (*Signature, error) {
	digest := hashToNum([]byte(message))
	alpha, r := this.createRandomChallenges()    // step 2 in paper
	c, err := this.calculateC(digest, alpha, &r) // step 3 and 4 in paper

	if err != nil {
		return nil, err
	}
	return &Signature{
		c[0], r, this.keyImages,
	}, nil
}

// check l*KI = 0 by checking KI is a valid point
func verifyKeyImages(keyImages []privacy.Point) bool {
	var check bool = true
	for i := 0; i < len(keyImages); i += 1 {
		lKI := new(privacy.Point).ScalarMult(&keyImages[i], getLEdward())
		check = check && lKI.IsIdentity()
	}
	return check
}

func verifyRing(sig *Signature, K *Ring, message string) (bool, error) {
	digest := hashToNum([]byte(message))

	c := sig.c
	cBefore := sig.c
	for i := 0; i < len(sig.r); i += 1 {
		nextC, err := calculateNextC(
			digest,
			sig.r[i], c,
			K.keys[i],
			sig.keyImages,
		)
		if err != nil {
			return false, err
		}
		c = nextC
	}
	return privacy.Compare(c, cBefore) == 0, nil
}

func Verify(sig *Signature, K *Ring, message string) (bool, error) {
	b1 := verifyKeyImages(sig.keyImages)
	b2, err := verifyRing(sig, K, message)
	return (b1 && b2), err
}
