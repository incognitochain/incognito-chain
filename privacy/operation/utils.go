package operation

import (
	"fmt"

	"github.com/incognitochain/incognito-chain/privacy/operation/edwards25519"
)

// MultiScalarMultBuilder is a helper struct to make best use of MultiScalarMult functions. It assumes caller never passes "nil" scalars / points
type MultiScalarMultBuilder struct {
	Scalars []*Scalar
	Points  []*Point
	StaticPointMultBuilder
	useVarTime bool
}

func NewMultBuilder(_useVarTime bool) *MultiScalarMultBuilder {
	return &MultiScalarMultBuilder{
		useVarTime: _useVarTime,
		Scalars:    []*Scalar{},
		Points:     []*Point{},
	}
}

func (b *MultiScalarMultBuilder) Clone() *MultiScalarMultBuilder {
	scLst := make([]*Scalar, len(b.Scalars))
	pLst := make([]*Point, len(b.Scalars))
	for i := range scLst {
		scLst[i] = NewScalar().Set(b.Scalars[i])
		pLst[i] = (&Point{}).Set(b.Points[i])
	}
	return &MultiScalarMultBuilder{
		useVarTime:             b.useVarTime,
		Scalars:                scLst,
		Points:                 pLst,
		StaticPointMultBuilder: *b.StaticPointMultBuilder.Clone(),
	}
}

func (b *MultiScalarMultBuilder) Append(scLst []*Scalar, pLst []*Point) error {
	if len(scLst) != len(pLst) {
		return fmt.Errorf("multiScalarMultBuilder must take same-length slices")
	}
	b.Scalars = append(b.Scalars, scLst...)
	b.Points = append(b.Points, pLst...)
	return nil
}

func (b *MultiScalarMultBuilder) AppendSingle(sc *Scalar, p *Point) {
	b.Append([]*Scalar{sc}, []*Point{p})
}

func (b *MultiScalarMultBuilder) AppendWithMultiplier(b1 *MultiScalarMultBuilder, n *Scalar) error {
	if len(b1.StaticScalars) > 0 {
		if len(b.StaticScalars) == 0 {
			b.WithStaticPoints(b1.StaticPoints)
		}
		if len(b1.StaticScalars) != len(b.StaticScalars) {
			panic(fmt.Errorf("append-with-multiplier: static points length mismatch %d vs %d", len(b1.StaticScalars), len(b.StaticScalars)))
		}
		for i := range b.StaticScalars {
			b.AddStatic(i, NewScalar().Mul(b1.StaticScalars[i], n))
		}
	}
	var scLst []*Scalar
	for _, sc := range b1.Scalars {
		scLst = append(scLst, NewScalar().Mul(sc, n))
	}
	return b.Append(scLst, b1.Points)
}

func (b *MultiScalarMultBuilder) Execute() (result *Point) {
	if b.useVarTime || len(b.StaticScalars) > 0 { // mixed multiscalar-mult currently only supports vartime logic
		result = NewIdentityPoint().MixedVarTimeMultiScalarMult(b.Scalars, b.Points, b.StaticScalars, b.StaticPoints)
	} else {
		result = NewIdentityPoint().MultiScalarMult(b.Scalars, b.Points)
	}
	// reset builder after finalization
	*b = *NewMultBuilder(b.useVarTime)
	return result
}

func (b MultiScalarMultBuilder) Debug() {
	fmt.Printf("multbuilder sizes %d %d %d %d\n", len(b.Scalars), len(b.Points), len(b.StaticScalars), len(b.StaticPoints))
}

type PrecomputedPoint struct {
	p edwards25519.PrecomputedPoint
}

func (pp *PrecomputedPoint) From(p *Point) {
	pp.p.FromP3(&p.p)
}

type StaticPointMultBuilder struct {
	StaticScalars []*Scalar
	StaticPoints  []PrecomputedPoint
}

func (sb *StaticPointMultBuilder) WithStaticPoints(ppLst []PrecomputedPoint) *StaticPointMultBuilder {
	ssLst := make([]*Scalar, len(ppLst))
	for i := 0; i < len(ppLst); i++ {
		ssLst[i] = NewScalar() // initialize the slice with zero-valued scalars
	}

	*sb = StaticPointMultBuilder{ssLst, ppLst}
	return sb
}

func (sb *StaticPointMultBuilder) Clone() *StaticPointMultBuilder {
	ssLst := make([]*Scalar, len(sb.StaticScalars))
	ppLst := make([]PrecomputedPoint, len(sb.StaticScalars))
	for i := range ssLst {
		ssLst[i] = NewScalar().Set(sb.StaticScalars[i])
		ppLst[i] = sb.StaticPoints[i]
	}
	return &StaticPointMultBuilder{
		StaticScalars: ssLst,
		StaticPoints:  ppLst,
	}
}

func (sb *StaticPointMultBuilder) AddStatic(startIndex int, scLst ...*Scalar) error {
	if startIndex < 0 || startIndex+len(scLst) > len(sb.StaticScalars) {
		return fmt.Errorf("staticMultBuilder: append range exceeds static points length")
	}
	for i, sc := range scLst {
		sb.StaticScalars[startIndex+i].Add(sb.StaticScalars[startIndex+i], sc)
	}
	return nil
}

func (sb *StaticPointMultBuilder) SetStatic(startIndex int, scLst ...*Scalar) error {
	if startIndex < 0 || startIndex+len(scLst) > len(sb.StaticScalars) {
		return fmt.Errorf("staticMultBuilder: append range exceeds static points length")
	}
	for i, sc := range scLst {
		sb.StaticScalars[startIndex+i].Set(sc)
	}
	return nil
}

func (sb *StaticPointMultBuilder) MulStatic(startIndex int, scLst ...*Scalar) error {
	if startIndex < 0 || startIndex+len(scLst) > len(sb.StaticScalars) {
		return fmt.Errorf("staticMultBuilder: append range exceeds static points length")
	}
	for i, sc := range scLst {
		sb.StaticScalars[startIndex+i].Mul(sb.StaticScalars[startIndex+i], sc)
	}
	return nil
}
