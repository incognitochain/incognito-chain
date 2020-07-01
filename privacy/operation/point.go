package operation

import (
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"

	C25519 "github.com/incognitochain/incognito-chain/privacy/operation/curve25519"
)

type Point struct {
	key C25519.Key
}

func RandomPoint() *Point {
	sc := RandomScalar()
	return new(Point).ScalarMultBase(sc)
}

func (p Point) PointValid() bool {
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

	var point C25519.ExtendedGroupElement
	if !point.FromBytes(&p.key) {
		return nil, errors.New("Invalid point value")
	}
	return p, nil
}

func (p *Point) Set(q *Point) *Point {
	if p == nil {
		p = new(Point)
	}
	p.key = q.key
	return p
}

func (p Point) MarshalText() []byte {
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

func (p Point) ToBytes() [Ed25519KeySize]byte {
	return p.key.ToBytes()
}

func (p Point) ToBytesS() []byte {
	slice := p.key.ToBytes()
	return slice[:]
}

func (p *Point) FromBytes(b [Ed25519KeySize]byte) (*Point, error) {
	if p == nil {
		p = new(Point)
	}
	p.key.FromBytes(b)

	var point C25519.ExtendedGroupElement
	if !point.FromBytes(&p.key) {
		return nil, errors.New("Invalid point value")
	}

	return p, nil
}

func (p *Point) FromBytesS(b []byte) (*Point, error) {
	if len(b) != Ed25519KeySize {
		return nil, errors.New("Invalid Ed25519 Key Size")
	}

	if p == nil {
		p = new(Point)
	}
	var array [Ed25519KeySize]byte
	copy(array[:], b)
	p.key.FromBytes(array)

	var point C25519.ExtendedGroupElement
	if !point.FromBytes(&p.key) {
		return nil, errors.New("Invalid point value")
	}

	return p, nil
}

func (p *Point) Identity() *Point {
	if p == nil {
		p = new(Point)
	}
	p.key = C25519.Identity
	return p
}

func (p Point) IsIdentity() bool {
	if p.key == C25519.Identity {
		return true
	}
	return false
}

// does a * G where a is a scalar and G is the curve basepoint
func (p *Point) ScalarMultBase(a *Scalar) *Point {
	if p == nil {
		p = new(Point)
	}
	key := C25519.ScalarmultBase(&a.key)
	p.key = *key
	return p
}

func (p *Point) ScalarMult(pa *Point, a *Scalar) *Point {
	if p == nil {
		p = new(Point)
	}
	key := C25519.ScalarMultKey(&pa.key, &a.key)
	p.key = *key
	return p
}

func (p *Point) MultiScalarMultCached(scalarLs []*Scalar, pointPreComputedLs [][8]C25519.CachedGroupElement) *Point {
	nSc := len(scalarLs)

	if nSc != len(pointPreComputedLs) {
		panic("Cannot MultiscalarMul with different size inputs")
	}

	scalarKeyLs := make([]*C25519.Key, nSc)
	for i := 0; i < nSc; i++ {
		scalarKeyLs[i] = &scalarLs[i].key
	}
	key := C25519.MultiScalarMultKeyCached(pointPreComputedLs, scalarKeyLs)
	res, _ := new(Point).SetKey(key)
	return res
}

func (p *Point) MultiScalarMult(scalarLs []*Scalar, pointLs []*Point) *Point {
	nSc := len(scalarLs)
	nPoint := len(pointLs)

	if nSc != nPoint {
		panic("Cannot MultiscalarMul with different size inputs")
	}

	scalarKeyLs := make([]*C25519.Key, nSc)
	pointKeyLs := make([]*C25519.Key, nSc)
	for i := 0; i < nSc; i++ {
		scalarKeyLs[i] = &scalarLs[i].key
		pointKeyLs[i] = &pointLs[i].key
	}
	key := C25519.MultiScalarMultKey(pointKeyLs, scalarKeyLs)

	res, _ := new(Point).SetKey(key)
	return res
}

func (p *Point) InvertScalarMultBase(a *Scalar) *Point {
	if p == nil {
		p = new(Point)
	}
	inv := new(Scalar).Invert(a)
	p.ScalarMultBase(inv)
	return p
}

func (p *Point) InvertScalarMult(pa *Point, a *Scalar) *Point {
	inv := new(Scalar).Invert(a)
	p.ScalarMult(pa, inv)
	return p
}

func (p *Point) Derive(pa *Point, a *Scalar, b *Scalar) *Point {
	c := new(Scalar).Add(a, b)
	return p.InvertScalarMult(pa, c)
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

// aA + bB
func (p *Point) AddPedersen(a *Scalar, A *Point, b *Scalar, B *Point) *Point {
	if p == nil {
		p = new(Point)
	}

	var A_Precomputed [8]C25519.CachedGroupElement
	Ae := new(C25519.ExtendedGroupElement)
	Ae.FromBytes(&A.key)
	C25519.GePrecompute(&A_Precomputed, Ae)

	var B_Precomputed [8]C25519.CachedGroupElement
	Be := new(C25519.ExtendedGroupElement)
	Be.FromBytes(&B.key)
	C25519.GePrecompute(&B_Precomputed, Be)

	var key C25519.Key
	C25519.AddKeys3_3(&key, &a.key, &A_Precomputed, &b.key, &B_Precomputed)
	p.key = key
	return p
}

func (p *Point) AddPedersenCached(a *Scalar, APreCompute [8]C25519.CachedGroupElement, b *Scalar, BPreCompute [8]C25519.CachedGroupElement) *Point {
	if p == nil {
		p = new(Point)
	}

	var key C25519.Key
	C25519.AddKeys3_3(&key, &a.key, &APreCompute, &b.key, &BPreCompute)
	p.key = key
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

func IsPointEqual(pa *Point, pb *Point) bool {
	tmpa := pa.ToBytesS()
	tmpb := pb.ToBytesS()
	return subtle.ConstantTimeCompare(tmpa, tmpb) == 1
}

func HashToPointFromIndex(index int64, padStr string) *Point {
	array := C25519.GBASE.ToBytes()
	msg := array[:]
	msg = append(msg, []byte(padStr)...)
	msg = append(msg, []byte(string(index))...)

	keyHash := C25519.Key(C25519.Keccak256(msg))
	keyPoint := keyHash.HashToPoint()

	p, _ := new(Point).SetKey(keyPoint)
	return p
}

func HashToPoint(b []byte) *Point {
	keyHash := C25519.Key(C25519.Keccak256(b))
	keyPoint := keyHash.HashToPoint()

	p, _ := new(Point).SetKey(keyPoint)
	return p
}
