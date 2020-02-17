package mlsag

import (
	"encoding/hex"
	"errors"

	"github.com/incognitochain/incognito-chain/privacy"
)

type Signature struct {
	c         privacy.Scalar     // 32 bytes
	keyImages []privacy.Point    // 32 * size bytes
	r         [][]privacy.Scalar // 32 * size_1 * size_2 bytes
}

func (this *Signature) ToHex() (string, error) {
	b, err := this.ToBytes()
	if err != nil {
		return "", errors.New("Error in MLSAG Signature ToHex: the signature is broken (size of keyImages and r differ")
	}
	return hex.EncodeToString(b), nil
}

func (this *Signature) ToBytes() ([]byte, error) {
	var b []byte

	// Number of private keys should be up to 2^8 only (1 byte)
	var leng byte = byte(len(this.keyImages))

	b = append(b, leng)
	b = append(b, this.c.ToBytesS()...)
	for i := 0; i < int(leng); i += 1 {
		b = append(b, this.keyImages[i].ToBytesS()...)
	}
	for i := 0; i < len(this.r); i += 1 {
		if int(leng) != len(this.r[i]) {
			return []byte{}, errors.New("Error in MLSAG Signature ToBytes: the signature is broken (size of keyImages and r differ)")
		}
		for j := 0; j < int(leng); j += 1 {
			b = append(b, this.r[i][j].ToBytesS()...)
		}
	}
	return b, nil
}

func (this *Signature) FromHex(s string) (*Signature, error) {
	b, err := hex.DecodeString(s)
	if err != nil {
		return nil, errors.New("Error in MLSAG Signature FromHex: the signature hex is broken")
	}
	return this.FromBytes(b)
}

// Get from byte and store to signature
func (this *Signature) FromBytes(b []byte) (*Signature, error) {
	if len(b)%32 != 1 {
		return nil, errors.New("Error in MLSAG Signature FromBytes: the signature byte is broken (missing byte)")
	}
	var m int = int(b[0])
	lenArr := len(b) - 33 - m*32
	n := lenArr / 32 / m

	if len(b) != 1+(1+m+m*n)*32 {
		return nil, errors.New("Error in MLSAG Signature FromBytes: the signature byte is broken (some scalar is missing)")
	}

	c := new(privacy.Scalar).FromBytesS(b[1:33])

	index := 33
	keyImages := make([]privacy.Point, m)
	for i := 0; i < m; i += 1 {
		val, err := new(privacy.Point).FromBytesS(b[index : index+32])
		if err != nil {
			return nil, errors.New("Error in MLSAG Signature FromBytes: the signature byte is broken (keyImages is broken)")
		}
		keyImages[i] = *val
		index += 32
	}

	r := make([][]privacy.Scalar, n)
	for i := 0; i < n; i += 1 {
		row := make([]privacy.Scalar, m)
		for j := 0; j < m; j += 1 {
			row[j] = *new(privacy.Scalar).FromBytesS(b[index : index+32])
			index += 32
		}
		r[i] = row
	}

	if this == nil {
		this = new(Signature)
	}
	this.c = *c
	this.keyImages = keyImages
	this.r = r

	return this, nil
}
