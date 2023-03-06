// Package operation allows for basic manipulation of scalars & group elements
package operation

import (
	"encoding/hex"
	"fmt"

	"github.com/incognitochain/incognito-chain/privacy/operation/edwards25519"
	v1 "github.com/incognitochain/incognito-chain/privacy/operation/v1"
)

// Point is a wrapper for `edwards25519.Point`, representing a point on the curve.
// It needs to be instantiated via constructor; while `new(Point)` can only be used as receiver.
type Point struct {
	p edwards25519.Point
}

// NewGeneratorPoint returns a new instance of the curve generator point
func NewGeneratorPoint() *Point {
	return &Point{*edwards25519.NewGeneratorPoint()}
}

// NewIdentityPoint returns a new instance of the curve identity point
func NewIdentityPoint() *Point {
	return &Point{*edwards25519.NewIdentityPoint()}
}

// PointValid checks that `p` belongs to the group (p first needs to be a valid Point object)
func (p *Point) PointValid() bool {
	if p == nil {
		return false
	}
	id := edwards25519.NewIdentityPoint()
	if p.p.Equal(id) == 1 {
		return true
	}
	return edwards25519.NewIdentityPoint().MultByCofactor(&p.p).Equal(id) != 1
}

// Set sets the value of `p` to that of `q`, then returns `p`
func (p *Point) Set(q *Point) *Point {
	p.p.Set(&q.p)
	return p
}

// String returns the hex string representation of `p`
func (p Point) String() string {
	return hex.EncodeToString(p.ToBytesS())
}

// MarshalText returns the hex string representation of `p` but as bytes
func (p Point) MarshalText() []byte {
	return []byte(p.String())
}

// UnmarshalText decodes a Point from its hex string form and sets `p` to it, then returns p
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

// ToBytesS marshals `p` into a byte slice
func (p *Point) ToBytesS() (result []byte) {
	defer func() {
		if r := recover(); r != nil {
			var b [32]byte
			result = b[:]
		}
	}()
	result = p.p.Bytes()
	return result
}

// FromBytesS unmarshals `p` from a byte slice, then returns `p`
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

// ToBytes marshals `p` into a byte array
func (p *Point) ToBytes() (result [32]byte) {
	copy(result[:], p.ToBytesS())
	return result
}

// FromBytes unmarshals `p` from a byte array, then returns `p`
func (p *Point) FromBytes(bArr [32]byte) (*Point, error) {
	return p.FromBytesS(bArr[:])
}

// RandomPoint returns a random point in the group, using `crypto/`'s randomness
func RandomPoint() *Point {
	sc := RandomScalar()
	return (&Point{}).ScalarMultBase(sc)
}

// Identity sets `p`'s value to identity
func (p *Point) Identity() *Point {
	p.p = *edwards25519.NewIdentityPoint()
	return p
}

// IsIdentity checks if Point `p` is equal to identity
func (p Point) IsIdentity() bool {
	return p.p.Equal(edwards25519.NewIdentityPoint()) == 1
}

// ScalarMultBase sets `p = a * G` where a is a scalar and G is the curve basepoint, then returns `p`
func (p *Point) ScalarMultBase(a *Scalar) *Point {
	p.p.ScalarBaseMult(&a.sc)
	return p
}

// ScalarMultBase sets `p = a * p_a`, then returns `p`
func (p *Point) ScalarMult(pa *Point, a *Scalar) *Point {
	p.p.ScalarMult(&a.sc, &pa.p)
	return p
}

// MultiScalarMult performs a multi-scalar multiplication on the group; sets and returns `p = sum(scalarLs[i] * pointLs[i])`.
// The caller must pass inputs of matching length to it.
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

// VarTimeMultiScalarMult is a multiscalar-mult variant that uses variable-time logic instead of constant-time.
// The caller must pass inputs of matching length to it.
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

// MixedVarTimeMultiScalarMult is a multiscalar-mult variant that uses variable-time logic and takes static (precomputed) points in combination with dynamic points.
// The caller must pass scalar lists of matching length for dynamic & static points respectively.
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
	ppLst := make([]*edwards25519.PrecomputedPoint, l1)
	for i := 0; i < len(staticScalarLs); i++ {
		ssLst[i] = &staticScalarLs[i].sc
		ppLst[i] = staticPointLs[i].p
	}
	p.p = *edwards25519.NewIdentityPoint()
	p.p.MixedVarTimeMultiScalarMult(scalarKeyLs, pointKeyLs, ssLst, ppLst)
	return p
}

// Derive is a privacy-v1 legacy function; it performs the SN-derivation algorithm, then sets and returns `p`
func (p *Point) Derive(pa *Point, a *Scalar, b *Scalar) *Point {
	temp := NewScalar().Add(a, b)
	return p.ScalarMult(pa, temp.Invert(temp))
}

// GetKey is a legacy function & alias of `ToBytes`
func (p Point) GetKey() [32]byte {
	return p.ToBytes()
}

// SetKey is a legacy function & alias of `FromBytes`
func (p *Point) SetKey(bArr *[32]byte) (*Point, error) {
	return p.FromBytes(*bArr)
}

// Add sets `p = p_a + p_b`, then returns `p`
func (p *Point) Add(pa, pb *Point) *Point {
	p.p.Add(&pa.p, &pb.p)
	return p
}

// Sub sets `p = p_a - p_b`, then returns `p`
func (p *Point) Sub(pa, pb *Point) *Point {
	p.p.Subtract(&pa.p, &pb.p)
	return p
}

// AddPedersen computes a Pedersen commitment; it sets `p = aA + bB`, then returns `p`
//
//nolint:gocritic // using capitalized variable name
func (p *Point) AddPedersen(a *Scalar, A *Point, b *Scalar, B *Point) *Point {
	return p.MultiScalarMult([]*Scalar{a, b}, []*Point{A, B})
}

// IsPointEqual checks the equality of 2 `Point`s
func IsPointEqual(pa *Point, pb *Point) bool {
	return pa.p.Equal(&pb.p) == 1
}

// HashToPointFromIndex is a legacy function; it maps an index to a Point
func HashToPointFromIndex(index int32, padStr string) *Point {
	msg := edwards25519.NewGeneratorPoint().Bytes()
	msg = append(msg, []byte(padStr)...)
	msg = append(msg, []byte(fmt.Sprintf("%c", index))...)
	return HashToPoint(msg)
}

// HashToPoint is a legacy map-to-point implementation
func HashToPoint(b []byte) *Point {
	temp := v1.HashToPoint(b)
	result := &Point{}
	result.FromBytesS(temp.ToBytesS()) //nolint // legacy Point marshals to valid bytes
	return result
}
