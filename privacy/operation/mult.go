package operation

import (
	"fmt"

	"github.com/incognitochain/incognito-chain/privacy/operation/edwards25519"
)

// MultiScalarMultBuilder is a helper struct to make best use of MultiScalarMult functions. The idea is to delay the invocation of Point multiplications
// so we can combine them into 1 large `MultiScalarMult` operation, which is faster.
//
// MultiScalarMultBuilder keeps track of lists of scalars & points of guaranteed matching length; it evaluates to
// `sum(scalars[i] * points[i]) + sum(static_scalars[i] * static_points[i])`.
// Static members are precomputed.
//
// This struct assumes caller never passes "nil" scalars / points
type MultiScalarMultBuilder struct {
	scalars []*Scalar
	points  []*Point
	staticPointMultBuilder
	useVarTime bool
}

// NewMultBuilder returns a new instance of MultiScalarMultBuilder, with fields initialized to zero except static ones (which need a call to WithStaticPoints)
func NewMultBuilder(_useVarTime bool) *MultiScalarMultBuilder {
	return &MultiScalarMultBuilder{
		useVarTime: _useVarTime,
		scalars:    []*Scalar{},
		points:     []*Point{},
	}
}

// Clone returns an exact clone of b, with newly allocated members
func (b *MultiScalarMultBuilder) Clone() *MultiScalarMultBuilder {
	scLst := make([]*Scalar, len(b.scalars))
	pLst := make([]*Point, len(b.scalars))
	for i := range scLst {
		scLst[i] = NewScalar().Set(b.scalars[i])
		pLst[i] = (&Point{}).Set(b.points[i])
	}
	return &MultiScalarMultBuilder{
		useVarTime:             b.useVarTime,
		scalars:                scLst,
		points:                 pLst,
		staticPointMultBuilder: *b.staticPointMultBuilder.Clone(),
	}
}

// Append appends more scalars and points to b, making sure their lengths match.
// It does not clone but consumes its inputs.
//
// This adds `sum(scLst[i] * pLst[i])` to evaluation of b.
func (b *MultiScalarMultBuilder) Append(scLst []*Scalar, pLst []*Point) error {
	if len(scLst) != len(pLst) {
		return fmt.Errorf("multiScalarMultBuilder must take same-length slices")
	}
	b.scalars = append(b.scalars, scLst...)
	b.points = append(b.points, pLst...)
	return nil
}

// MustAppend appends more scalars and points to b. Caller must pass slices of matching lengths or it panics.
// It does not clone but consumes its inputs.
//
// This adds `sum(scLst[i] * pLst[i])` to evaluation of b.
func (b *MultiScalarMultBuilder) MustAppend(scLst []*Scalar, pLst []*Point) {
	if len(scLst) != len(pLst) {
		panic(fmt.Errorf("multiScalarMultBuilder must take same-length slices"))
	}
	b.scalars = append(b.scalars, scLst...)
	b.points = append(b.points, pLst...)
}

// AppendSingle appends 1 scalar & 1 point to b.
//
// This adds `sc * p` to evaluation of b.
func (b *MultiScalarMultBuilder) AppendSingle(sc *Scalar, p *Point) {
	b.MustAppend([]*Scalar{sc}, []*Point{p})
}

// ConcatScaled first scales the evaluation of `b1` by `n`, then concatenates it to the fields of b.
//
// Dynamic members are appended, while static members are summed by-column.
//
// This adds `sum(scalars_1[i] * points_1[i]) + sum(static_scalars_1[i] * static_points_1[i])` to the evaluation of b.
func (b *MultiScalarMultBuilder) ConcatScaled(b1 *MultiScalarMultBuilder, n *Scalar) error {
	if len(b1.StaticScalars) > 0 {
		if len(b.StaticPoints) == 0 {
			b.WithStaticPoints(b1.StaticPoints)
		}
		if len(b1.StaticPoints) != len(b.StaticPoints) {
			return fmt.Errorf("append-with-multiplier: static points length mismatch %d vs %d", len(b1.StaticScalars), len(b.StaticScalars))
		}
		for k, v := range b1.StaticScalars {
			vn := NewScalar().Mul(v, n)
			sc, exists := b.StaticScalars[k]
			if !exists {
				sc = NewScalar()
				b.StaticScalars[k] = sc
			}
			sc.Add(sc, vn)
		}
	}
	var scLst []*Scalar
	for _, sc := range b1.scalars {
		scLst = append(scLst, NewScalar().Mul(sc, n))
	}
	return b.Append(scLst, b1.points)
}

// Eval invokes a multiscalar-mult function to finalize the builder `b`, returning the resulted Point.
//
// It chooses the constant-time or variable-time variant of multiscalar-mult based on the contents of `b`; it clears `b` after execution.
func (b *MultiScalarMultBuilder) Eval() (result *Point) {
	if b.useVarTime || len(b.StaticScalars) > 0 { // mixed multiscalar-mult currently only supports vartime logic
		var simplifiedStaticScalars []*Scalar
		var simplifiedStaticPoints []PrecomputedPoint
		for k, v := range b.StaticScalars {
			simplifiedStaticScalars = append(simplifiedStaticScalars, v)
			simplifiedStaticPoints = append(simplifiedStaticPoints, b.StaticPoints[k])
		}
		result = NewIdentityPoint().MixedVarTimeMultiScalarMult(b.scalars, b.points, simplifiedStaticScalars, simplifiedStaticPoints)
	} else {
		result = NewIdentityPoint().MultiScalarMult(b.scalars, b.points)
	}
	// reset builder after finalization
	*b = *NewMultBuilder(b.useVarTime)
	return result
}

func (b MultiScalarMultBuilder) Debug() {
	fmt.Printf("multbuilder of value %v, sizes %d %d %d %d\n", b.Clone().Eval(), len(b.scalars), len(b.points), len(b.StaticScalars), len(b.StaticPoints))
}

// PrecomputedPoint wraps the struct edwards25519.PrecomputedPoint which stores precomputed field elements to speed up computation
type PrecomputedPoint struct {
	p *edwards25519.PrecomputedPoint
}

// From populates this PrecomputedPoint to the value of Point p
func (pp *PrecomputedPoint) From(q *Point) *PrecomputedPoint {
	pp.p = &edwards25519.PrecomputedPoint{}
	pp.p.FromP3(&q.p)
	return pp
}

type staticPointMultBuilder struct {
	StaticScalars map[int]*Scalar
	StaticPoints  []PrecomputedPoint
}

// WithStaticPoints sets the list of static Points for the builder.
//
// The slice is not cloned since PrecomputedPoint is immutable.
func (sb *staticPointMultBuilder) WithStaticPoints(ppLst []PrecomputedPoint) *staticPointMultBuilder {
	ssLst := make(map[int]*Scalar)
	*sb = staticPointMultBuilder{ssLst, ppLst}
	return sb
}

func (sb *staticPointMultBuilder) Clone() *staticPointMultBuilder {
	ssLst := make(map[int]*Scalar)
	for k, v := range sb.StaticScalars {
		ssLst[k] = NewScalar().Set(v)
	}
	return &staticPointMultBuilder{
		StaticScalars: ssLst,
		StaticPoints:  sb.StaticPoints,
	}
}

// SetStatic is the equivalent of Append for static members. Static Points are identified using index.
func (sb *staticPointMultBuilder) SetStatic(startIndex int, scLst ...*Scalar) error {
	if startIndex < 0 || startIndex+len(scLst) > len(sb.StaticPoints) {
		return fmt.Errorf("staticMultBuilder: append range exceeds static points length")
	}
	for i, sc := range scLst {
		sb.StaticScalars[startIndex+i] = NewScalar().Set(sc)
	}
	return nil
}

// MustSetStatic is the equivalent of MustAppend for static members. Static Points are identified using index.
func (sb *staticPointMultBuilder) MustSetStatic(startIndex int, scLst ...*Scalar) {
	if startIndex < 0 || startIndex+len(scLst) > len(sb.StaticPoints) {
		panic(fmt.Errorf("staticMultBuilder: append range exceeds static points length"))
	}
	for i, sc := range scLst {
		sb.StaticScalars[startIndex+i] = NewScalar().Set(sc)
	}
}
