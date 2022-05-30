package operation

import (
	"fmt"
	"math/rand"
	"testing"

	. "github.com/stretchr/testify/assert"
)

var staticPointsPrecompute, staticPoints = randStaticPoints()

func randStaticPoints() ([]PrecomputedPoint, []*Point) {
	sLen := rand.Int()%100 + 1
	ppLst := make([]PrecomputedPoint, sLen)
	spLst := make([]*Point, sLen)
	for i := range spLst {
		spLst[i] = RandomPoint()
		ppLst[i].From(spLst[i])
	}
	return ppLst, spLst
}

func randMult(useVarTime bool) (*MultiScalarMultBuilder, []*Point) {
	dLen := rand.Int()%100 + 1

	scLst := make([]*Scalar, dLen)
	pLst := make([]*Point, dLen)
	for i := range scLst {
		scLst[i] = RandomScalar()
		pLst[i] = RandomPoint()
	}

	ssMap := make(map[int]*Scalar)
	for i := range staticPointsPrecompute {
		ssMap[i] = RandomScalar()
	}

	b := MultiScalarMultBuilder{
		useVarTime: useVarTime,
		scalars:    scLst,
		points:     pLst,
	}
	b.StaticScalars = ssMap
	b.StaticPoints = staticPointsPrecompute
	return &b, staticPoints
}

func TestMultBuilder(t *testing.T) {
	b, spLst := randMult(true)
	allScalars := make([]*Scalar, len(b.StaticPoints))
	for i := range allScalars {
		sc, exists := b.StaticScalars[i]
		True(t, exists)
		allScalars[i] = sc
	}
	allScalars = append(append([]*Scalar{}, b.scalars...), allScalars...)
	allPoints := append(append([]*Point{}, b.points...), spLst...)
	expected1 := NewIdentityPoint().MultiScalarMult(allScalars, allPoints)
	expected2 := NewIdentityPoint().VarTimeMultiScalarMult(allScalars, allPoints)

	fmt.Println(expected1)
	fmt.Println(expected2)
	b.Debug()
	actualResult := b.Eval()

	True(t, IsPointEqual(actualResult, expected1))
	True(t, IsPointEqual(actualResult, expected2))
}

func TestMultBuilderConcat(t *testing.T) {
	mbLen := rand.Int()%10 + 1
	multBuilderLst := make([]*MultiScalarMultBuilder, mbLen)
	scaleFactorLst := make([]*Scalar, mbLen)
	for i := range multBuilderLst {
		multBuilderLst[i], _ = randMult(true)
		scaleFactorLst[i] = RandomScalar()
	}

	expectedSum := multBuilderLst[0].Clone().Eval()
	actualSum := multBuilderLst[0].Clone()
	for i := range multBuilderLst[1:] {
		temp := multBuilderLst[i].Clone().Eval()
		expectedSum.Add(expectedSum, temp.ScalarMult(temp, scaleFactorLst[i]))

		actualSum.ConcatScaled(multBuilderLst[i].Clone(), scaleFactorLst[i])
		fmt.Println(expectedSum)
		actualSum.Debug()
		True(t, IsPointEqual(actualSum.Clone().Eval(), expectedSum))
	}
}

func TestInvalidMultBuilder(t *testing.T) {
	b, _ := randMult(true)
	b1 := b.Clone()
	b1.scalars = append(b1.scalars, RandomScalar())
	Panics(t, func() {
		b1.Eval()
	})
}
