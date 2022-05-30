//nolint:gocritic,revive // skip some linters since this file has some capitalized, underscored
// variable names to match names in the crypto protocol
package mlsag

import (
	"errors"
	"github.com/incognitochain/incognito-chain/privacy/operation"
)

// Sig is the ring signature that appears on transactions.
type Sig struct {
	c         *operation.Scalar     // 32 bytes
	keyImages []*operation.Point    // 32 * size bytes
	r         [][]*operation.Scalar // 32 * size_1 * size_2 bytes
}

func NewMlsagSig(c *operation.Scalar, keyImages []*operation.Point, r [][]*operation.Scalar) (*Sig, error) {
	if len(r) == 0 {
		return nil, errors.New("Cannot create new mlsag signature, length of r is not correct")
	}
	if len(keyImages) != len(r[0]) {
		return nil, errors.New("Cannot create new mlsag signature, length of keyImages is not correct")
	}
	res := new(Sig)
	res.SetC(c)
	res.SetR(r)
	res.SetKeyImages(keyImages)
	return res, nil
}

func (sig Sig) GetC() *operation.Scalar          { return sig.c }
func (sig Sig) GetKeyImages() []*operation.Point { return sig.keyImages }
func (sig Sig) GetR() [][]*operation.Scalar      { return sig.r }

func (sig *Sig) SetC(c *operation.Scalar)                  { sig.c = c }
func (sig *Sig) SetKeyImages(keyImages []*operation.Point) { sig.keyImages = keyImages }
func (sig *Sig) SetR(r [][]*operation.Scalar)              { sig.r = r }

func (sig *Sig) ToBytes() ([]byte, error) {
	b := []byte{MlsagPrefix}

	if sig.c != nil {
		b = append(b, operation.Ed25519KeySize)
		b = append(b, sig.c.ToBytesS()...)
	} else {
		b = append(b, 0)
	}

	if sig.keyImages != nil {
		if len(sig.keyImages) > MaxSizeByte {
			return nil, errors.New("Length of key image is too large > 255")
		}
		lenKeyImage := byte(len(sig.keyImages) & 0xFF)
		b = append(b, lenKeyImage)
		for i := 0; i < int(lenKeyImage); i++ {
			b = append(b, sig.keyImages[i].ToBytesS()...)
		}
	} else {
		b = append(b, 0)
	}

	if sig.r != nil {
		n := len(sig.r)
		if n == 0 {
			b = append(b, 0)
			b = append(b, 0)
			return b, nil
		}
		m := len(sig.r[0])
		if n > MaxSizeByte || m > MaxSizeByte {
			return nil, errors.New("Length of R of mlsagSig is too large > 255")
		}
		b = append(b, byte(n&0xFF))
		b = append(b, byte(m&0xFF))
		for i := 0; i < n; i++ {
			if m != len(sig.r[i]) {
				return []byte{}, errors.New("Error in MLSAG MlsagSig ToBytes: the signature is broken (size of keyImages and r differ)")
			}
			for j := 0; j < m; j++ {
				b = append(b, sig.r[i][j].ToBytesS()...)
			}
		}
	} else {
		b = append(b, 0)
		b = append(b, 0)
	}

	return b, nil
}

// Get from byte and store to signature
func (sig *Sig) FromBytes(b []byte) (*Sig, error) {
	if len(b) == 0 {
		return nil, errors.New("Length of byte is empty, cannot setbyte mlsagSig")
	}
	if b[0] != MlsagPrefix {
		return nil, errors.New("The signature byte is broken (first byte is not mlsag)")
	}

	offset := 1
	if b[offset] != operation.Ed25519KeySize {
		return nil, errors.New("Cannot parse value C, byte length of C is wrong")
	}
	offset++
	if offset+operation.Ed25519KeySize > len(b) {
		return nil, errors.New("Cannot parse value C, byte is too small")
	}
	C := new(operation.Scalar).FromBytesS(b[offset : offset+operation.Ed25519KeySize])
	if !C.ScalarValid() {
		return nil, errors.New("Cannot parse value C, invalid scalar")
	}
	offset += operation.Ed25519KeySize

	if offset >= len(b) {
		return nil, errors.New("Cannot parse length of keyimage, byte is too small")
	}
	lenKeyImages := int(b[offset])
	offset++
	keyImages := make([]*operation.Point, lenKeyImages)
	for i := 0; i < lenKeyImages; i++ {
		if offset+operation.Ed25519KeySize > len(b) {
			return nil, errors.New("Cannot parse keyimage of mlsagSig, byte is too small")
		}
		var err error
		keyImages[i], err = new(operation.Point).FromBytesS(b[offset : offset+operation.Ed25519KeySize])
		if err != nil {
			return nil, errors.New("Cannot convert byte to operation point keyimage")
		}
		offset += operation.Ed25519KeySize
	}

	if offset+2 > len(b) {
		return nil, errors.New("Cannot parse length of R, byte is too small")
	}
	n := int(b[offset])
	m := int(b[offset+1])
	offset += 2

	R := make([][]*operation.Scalar, n)
	for i := 0; i < n; i++ {
		R[i] = make([]*operation.Scalar, m)
		for j := 0; j < m; j++ {
			if offset+operation.Ed25519KeySize > len(b) {
				return nil, errors.New("Cannot parse R of mlsagSig, byte is too small")
			}
			sc := new(operation.Scalar).FromBytesS(b[offset : offset+operation.Ed25519KeySize])
			if !sc.ScalarValid() {
				return nil, errors.New("Cannot parse R of mlsagSig, invalid scalar")
			}
			R[i][j] = sc
			offset += operation.Ed25519KeySize
		}
	}

	if sig == nil {
		sig = new(Sig)
	}
	sig.SetC(C)
	sig.SetKeyImages(keyImages)
	sig.SetR(R)
	return sig, nil
}
