package privacy

import (
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	C25519 "github.com/deroproject/derosuite/crypto"
)

type Point struct {
	key C25519.Key
}

func (p *Point) PointValid() bool {
	var point C25519.ExtendedGroupElement
	return point.FromBytes(&p.key)
}

func (p Point) GetKey() C25519.Key {
	return p.key
}

func (p *Point) SetKey(a *C25519.Key) (*Point, error) {
	if p == nil {
		p = new(Point)
	}
	p.key = *a
	if p.PointValid() == false {
		return nil, errors.New("Invalid key value")
	}
	return p, nil
}

func (p Point) MarshalText() ([]byte) {
	return []byte(fmt.Sprintf("%x", p.key[:]))
}

func (p *Point) UnmarshalText(data []byte) (*Point, error) {
	if p == nil {
		p = new(Point)
	}

	byteSlice, _ := hex.DecodeString(string(data))
	if len(byteSlice) != Ed25519KeySize {
		return nil, errors.New("Incorrect key size")
	}
	copy(p.key[:], byteSlice)
	return p, nil
}

func (p *Point) Zero() *Point {
	if p == nil {
		p = new(Point)
	}
	copy(p.key[:], C25519.Zero.String())
	return p
}

// does a * G where a is a scalar and G is the curve basepoint
func (p *Point) ScalarMulBase(a *Scalar) *Point {
	if p == nil {
		p = new(Point)
	}

	key := C25519.ScalarmultBase(a.key)
	p.key = key
	return p
}

func (p *Point) ScalarMul(pa *Point, a *Scalar) *Point {
	if p == nil {
		p = new(Point)
	}
	key := C25519.ScalarMultKey(&pa.key, &a.key)
	p.SetKey(key)
	return p
}

func (p *Point) InvertScalarMulBase(a *Scalar) *Point {
	if p == nil {
		p = new(Point)
	}
	inv := new(Scalar).Invert(a)
	p.ScalarMulBase(inv)
	return p
}

func (p *Point) InvertScalarMul(pa *Point, a *Scalar) *Point {
	inv := new(Scalar).Invert(a)
	p.ScalarMul(pa,inv)
	return p
}

func (p *Point) Add(pa, pb *Point) *Point {
	if p == nil {
		p = new(Point)
	}
	res := p.key
	C25519.AddKeys(&res, &pa.key, &pb.key)
	p.key = res
	return p
}

func (p *Point) Sub(pa, pb *Point) *Point {
	if p == nil {
		p = new(Point)
	}
	res := p.key
	C25519.SubKeys(&res, &pa.key, &pb.key)
	p.key = res
	return p
}

func IsEqual(pa *Point, pb *Point) bool {
	tmpa, errora := pa.key.MarshalText()
	tmpb, errorb := pb.key.MarshalText()
	if errora != nil || errorb != nil {
		return false
	}
	return subtle.ConstantTimeCompare(tmpa, tmpb) == 1
}

func RandomPoint() *Point {
	p := new(Point)
	sc := RandomScalar()
	p = new(Point).ScalarMulBase(sc)
	return p
}

func HashToPoint(index int64) *Point {
	msg, _ := C25519.GBASE.MarshalText()
	msg = append(msg,[]byte(CStringBulletProof)...)
	msg = append(msg,[]byte(string(index))...)

	keyHash := C25519.Key(C25519.Keccak256(msg))
	keyPoint := keyHash.HashToPoint()

	p, error := new(Point).SetKey(&keyPoint)
	if error != nil {
		return nil
	}
	return p
}







