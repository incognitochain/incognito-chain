package operation

import (
	"encoding/hex"
	"fmt"

	"github.com/incognitochain/incognito-chain/privacy/operation/edwards25519"
	v1 "github.com/incognitochain/incognito-chain/privacy/operation/v1"
)

type Point struct {
	p edwards25519.Point
}

func NewGeneratorPoint() *Point {
	return &Point{*edwards25519.NewGeneratorPoint()}
}

func NewIdentityPoint() *Point {
	return &Point{*edwards25519.NewIdentityPoint()}
}

// PointValid checks that p belongs to the group. It does need to be a valid Point object first, or this will panic
func (p Point) PointValid() bool {
	id := edwards25519.NewIdentityPoint()
	if p.p.Equal(id) == 1 {
		return true
	}
	return p.p.MultByCofactor(&p.p).Equal(id) != 1
}

func (p *Point) Set(q *Point) *Point {
	p.p.Set(&q.p)
	return p
}

func (p Point) String() string {
	return fmt.Sprintf("%x", p.ToBytesS())
}

func (p Point) MarshalText() []byte {
	return []byte(hex.EncodeToString(p.ToBytesS()))
}

func (p *Point) UnmarshalText(data []byte) (*Point, error) {
	byteSlice, _ := hex.DecodeString(string(data))
	if len(byteSlice) != Ed25519KeySize {
		return nil, fmt.Errorf("invalid point byte size")
	}
	_, err := p.p.SetBytes(byteSlice)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (p Point) ToBytesS() []byte {
	return p.p.Bytes()
}

func (p *Point) FromBytesS(b []byte) (*Point, error) {
	if len(b) != Ed25519KeySize {
		return nil, fmt.Errorf("invalid point byte Size")
	}
	_, err := p.p.SetBytes(b)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (p Point) ToBytes() (result [32]byte) {
	copy(result[:], p.p.Bytes())
	return result
}

func (p *Point) FromBytes(bArr [32]byte) (*Point, error) {
	_, err := p.p.SetBytes(bArr[:])
	if err != nil {
		return nil, err
	}
	return p, nil
}

func RandomPoint() *Point {
	sc := RandomScalar()
	return (&Point{}).ScalarMultBase(sc)
}

func (p *Point) Identity() *Point {
	p.p = *edwards25519.NewIdentityPoint()
	return p
}

func (p Point) IsIdentity() bool {
	return p.p.Equal(edwards25519.NewIdentityPoint()) == 1
}

// does a * G where a is a scalar and G is the curve basepoint
func (p *Point) ScalarMultBase(a *Scalar) *Point {
	p.p.ScalarBaseMult(&a.sc)
	return p
}

func (p *Point) ScalarMult(pa *Point, a *Scalar) *Point {
	p.p.ScalarMult(&a.sc, &pa.p)
	return p
}

func (p *Point) MultiScalarMult(scalarLs []*Scalar, pointLs []*Point) *Point {
	l := len(scalarLs)
	// must take inputs of the same length
	if l != len(pointLs) {
		panic("Cannot MultiscalarMul with different size inputs")
	}

	scalarKeyLs := make([]*edwards25519.Scalar, l)
	pointKeyLs := make([]*edwards25519.Point, l)
	for i := 0; i < l; i++ {
		scalarKeyLs[i] = &scalarLs[i].sc
		pointKeyLs[i] = &pointLs[i].p
	}
	// need to be valid point to call MultiScalarMult
	p.p = *edwards25519.NewIdentityPoint()
	p.p.MultiScalarMult(scalarKeyLs, pointKeyLs)
	return p
}

func (p *Point) VarTimeMultiScalarMult(scalarLs []*Scalar, pointLs []*Point) *Point {
	l := len(scalarLs)
	// must take inputs of the same length
	if l != len(pointLs) {
		panic("Cannot MultiscalarMul with different size inputs")
	}

	scalarKeyLs := make([]*edwards25519.Scalar, l)
	pointKeyLs := make([]*edwards25519.Point, l)
	for i := 0; i < l; i++ {
		scalarKeyLs[i] = &scalarLs[i].sc
		pointKeyLs[i] = &pointLs[i].p
	}
	p.p = *edwards25519.NewIdentityPoint()
	p.p.VarTimeMultiScalarMult(scalarKeyLs, pointKeyLs)
	return p
}

func (p *Point) MixedVarTimeMultiScalarMult(scalarLs []*Scalar, pointLs []*Point, staticScalarLs []*Scalar, staticPointLs []PrecomputedPoint) *Point {
	l := len(scalarLs)
	l1 := len(staticScalarLs)
	// must take inputs of the same length
	if l != len(pointLs) || l1 != len(staticPointLs) {
		panic("Cannot MultiscalarMul with different size inputs")
	}

	scalarKeyLs := make([]*edwards25519.Scalar, l)
	pointKeyLs := make([]*edwards25519.Point, l)
	for i := 0; i < l; i++ {
		scalarKeyLs[i] = &scalarLs[i].sc
		pointKeyLs[i] = &pointLs[i].p
	}

	ssLst := make([]*edwards25519.Scalar, l1)
	ppLst := make([]edwards25519.PrecomputedPoint, l1)
	for i := 0; i < len(staticScalarLs); i++ {
		ssLst[i] = &staticScalarLs[i].sc
		ppLst[i] = staticPointLs[i].p
	}
	p.p = *edwards25519.NewIdentityPoint()
	p.p.MixedVarTimeMultiScalarMult(scalarKeyLs, pointKeyLs, ssLst, ppLst)
	return p
}

func (p *Point) Derive(pa *Point, a *Scalar, b *Scalar) *Point {
	temp := NewScalar().Add(a, b)
	return p.ScalarMult(pa, temp.Invert(temp))
}

func (p Point) GetKey() [32]byte {
	return p.ToBytes()
}

func (p *Point) SetKey(bArr *[32]byte) (*Point, error) {
	return p.FromBytes(*bArr)
}

func (p *Point) Add(pa, pb *Point) *Point {
	p.p.Add(&pa.p, &pb.p)
	return p
}

func (p *Point) Sub(pa, pb *Point) *Point {
	p.p.Subtract(&pa.p, &pb.p)
	return p
}

// aA + bB
func (p *Point) AddPedersen(a *Scalar, A *Point, b *Scalar, B *Point) *Point {
	return p.MultiScalarMult([]*Scalar{a, b}, []*Point{A, B})
}

func IsPointEqual(pa *Point, pb *Point) bool {
	return pa.p.Equal(&pb.p) == 1
}

func HashToPointFromIndex(index int32, padStr string) *Point {
	msg := edwards25519.NewGeneratorPoint().Bytes()
	msg = append(msg, []byte(padStr)...)
	msg = append(msg, []byte(fmt.Sprintf("%c", index))...)

	return HashToPoint(msg)
}

// legacy map-to-point
func HashToPoint(b []byte) *Point {
	temp := v1.HashToPoint(b)
	result := &Point{}
	result.FromBytesS(temp.ToBytesS())
	return result
}
