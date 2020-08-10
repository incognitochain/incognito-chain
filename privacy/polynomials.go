package privacy

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/privacy/curve25519"
	"math/big"
	"math/rand"
	"time"
)

// Data structure for a polynomial
// Just an array in Reverse
// f(x) = 3x^3 + 2x + 1 => [1 2 0 3]
type Poly []*big.Int

// Helper function for generating a polynomial with given integers
func newPoly(coeffs ...int) (p Poly) {
	p = make([]*big.Int, len(coeffs))
	for i := 0; i < len(coeffs); i++ {
		p[i] = big.NewInt(int64(coeffs[i]))
	}
	p.trim()
	return
}

func ScalarToBigInt(sc *Scalar) *big.Int {
	keyR := Reverse(sc.key)
	keyRByte := keyR.ToBytes()
	bi := new(big.Int).SetBytes(keyRByte[:])
	return bi
}

func BigIntToScalar(bi *big.Int) *Scalar {
	biByte := common.AddPaddingBigInt(bi, Ed25519KeySize)
	var key curve25519.Key
	key.FromBytes(SliceToArray(biByte))
	keyR := Reverse(key)
	sc, err := new(Scalar).SetKey(&keyR)
	if err != nil {
		return nil
	}
	return sc
}

// Returns a polynomial with random coefficients
// You can give the degree of the polynomial
// A random coefficients have a [0, 2^bits) integer
func randomPoly(degree, bits int64) (p Poly) {
	p = make(Poly, degree+1)
	rr := rand.New(rand.NewSource(time.Now().UnixNano()))
	exp := big.NewInt(2)
	exp.Exp(exp, big.NewInt(bits), nil)
	for i := 0; i <= p.GetDegree(); i++ {
		p[i] = new(big.Int)
		p[i].Rand(rr, exp)
	}
	p.trim()
	return
}

// trim() makes sure that the highest coefficient never has zero value
// when you add or subtract two polynomials, sometimes the highest coefficient goes zero
// if you don't remove the highest and zero coefficient, GetDegree() returns the wrong result
func (p *Poly) trim() {
	var last int = 0
	for i := p.GetDegree(); i > 0; i-- { // why i > 0, not i >=0? do not remove the constant
		if (*p)[i].Sign() != 0 {
			last = i
			break
		}
	}
	*p = (*p)[:(last + 1)]
}

// isZero() checks if P = 0
func (p *Poly) isZero() bool {
	if p.GetDegree() == 0 && (*p)[0].Cmp(big.NewInt(0)) == 0 {
		return true
	}
	return false
}

// GetDegree returns the degree
// if p = x^3 + 2x^2 + 5, GetDegree() returns 3
func (p Poly) GetDegree() int {
	return len(p) - 1
}

// pretty print
func (p Poly) String() (s string) {
	s = "["
	for i := len(p) - 1; i >= 0; i-- {
		switch p[i].Sign() {
		case -1:
			if i == len(p)-1 {
				s += "-"
			} else {
				s += " - "
			}
			if i == 0 || p[i].Int64() != -1 {
				s += p[i].String()[1:]
			}
		case 0:
			continue
		case 1:
			if i < len(p)-1 {
				s += " + "
			}
			if i == 0 || p[i].Int64() != 1 {
				s += p[i].String()
			}
		}
		if i > 0 {
			s += "x"
			if i > 1 {
				s += "^" + fmt.Sprintf("%d", i)
			}
		}
	}
	if s == "[" {
		s += "0"
	}
	s += "]"
	return
}

// Compare() compares two polynomials and returns -1, 0, or 1
// if P == Q, returns 0
// if P > Q, returns 1
// if P < Q, returns -1
func (p *Poly) compare(q *Poly) int {
	switch {
	case p.GetDegree() > q.GetDegree():
		return 1
	case p.GetDegree() < q.GetDegree():
		return -1
	}
	for i := 0; i <= p.GetDegree(); i++ {
		switch (*p)[i].Cmp((*q)[i]) {
		case 1:
			return 1
		case -1:
			return -1
		}
	}
	return 0
}

// Add() adds two polynomials
// modulo m can be nil
func (p Poly) add(q Poly, m *big.Int) Poly {
	if p.compare(&q) < 0 {
		return q.add(p, m)
	}
	var r Poly = make([]*big.Int, len(p))
	for i := 0; i < len(q); i++ {
		a := new(big.Int)
		a.Add(p[i], q[i])
		r[i] = a
	}
	for i := len(q); i < len(p); i++ {
		a := new(big.Int)
		a.Set(p[i])
		r[i] = a
	}
	if m != nil {
		for i := 0; i < len(p); i++ {
			r[i].Mod(r[i], m)
		}
	}
	r.trim()
	return r
}

// Neg() returns a polynomial Q = -P
func (p *Poly) neg() Poly {
	var q Poly = make([]*big.Int, len(*p))
	for i := 0; i < len(*p); i++ {
		b := new(big.Int)
		b.Neg((*p)[i])
		q[i] = b
	}
	return q
}

// Clone() does deep-copy
// adjust increases the degree of copied polynomial
// adjust cannot have a negative integer
// for example, P = x + 1 and adjust = 2, Clone() returns x^3 + x^2
func (p Poly) clone(adjust int) Poly {
	var q Poly = make([]*big.Int, len(p)+adjust)
	if adjust < 0 {
		return newPoly(0)
	}
	for i := 0; i < adjust; i++ {
		q[i] = big.NewInt(0)
	}
	for i := adjust; i < len(p)+adjust; i++ {
		b := new(big.Int)
		b.Set(p[i-adjust])
		q[i] = b
	}
	return q
}

// sanitize() does modular arithmetic with m
func (p *Poly) sanitize(m *big.Int) {
	if m == nil {
		return
	}
	for i := 0; i <= (*p).GetDegree(); i++ {
		(*p)[i].Mod((*p)[i], m)
	}
	p.trim()
}

// Sub() subtracts P from Q
// Since we already have Add(), Sub() does Add(P, -Q)
func (p Poly) Sub(q Poly, m *big.Int) Poly {
	r := q.neg()
	return p.add(r, m)
}

// P * Q
func (p Poly) Mul(q Poly, m *big.Int) Poly {
	if m != nil {
		p.sanitize(m)
		q.sanitize(m)
	}
	var r Poly = make([]*big.Int, p.GetDegree()+q.GetDegree()+1)
	for i := 0; i < len(r); i++ {
		r[i] = big.NewInt(0)
	}
	for i := 0; i < len(p); i++ {
		for j := 0; j < len(q); j++ {
			a := new(big.Int)
			a.Mul(p[i], q[j])
			a.Add(a, r[i+j])
			if m != nil {
				a = new(big.Int).Mod(a, m)
			}
			r[i+j] = a
		}
	}
	r.trim()
	return r
}

// returns (P / Q, P % Q)
func (p Poly) div(q Poly, m *big.Int) (quo, rem Poly) {
	if m != nil {
		p.sanitize(m)
		q.sanitize(m)
	}
	if p.GetDegree() < q.GetDegree() || q.isZero() {
		quo = newPoly(0)
		rem = p.clone(0)
		return
	}
	quo = make([]*big.Int, p.GetDegree()-q.GetDegree()+1)
	rem = p.clone(0)
	for i := 0; i < len(quo); i++ {
		quo[i] = big.NewInt(0)
	}
	t := p.clone(0)
	qd := q.GetDegree()
	for {
		td := t.GetDegree()
		rd := td - qd
		if rd < 0 || t.isZero() {
			rem = t
			break
		}
		r := new(big.Int)
		if m != nil {
			r.ModInverse(q[qd], m)
			r.Mul(r, t[td])
			r.Mod(r, m)
		} else {
			r.Div(t[td], q[qd])
		}
		// if r == 0, it means that the highest coefficient of the result is not an integer
		// this polynomial library handles integer coefficients
		if r.Cmp(big.NewInt(0)) == 0 {
			quo = newPoly(0)
			rem = p.clone(0)
			return
		}
		u := q.clone(rd)
		for i := rd; i < len(u); i++ {
			u[i].Mul(u[i], r)
			if m != nil {
				u[i].Mod(u[i], m)
			}
		}
		t = t.Sub(u, m)
		t.trim()
		quo[rd] = r
	}
	quo.trim()
	rem.trim()
	return
}

// returns the greatest common divisor(GCD) of P and Q (Euclidean algorithm)
func (p Poly) gcd(q Poly, m *big.Int) Poly {
	if p.compare(&q) < 0 {
		return q.gcd(p, m)
	}
	if q.isZero() {
		return p
	} else {
		_, rem := p.div(q, m)
		return q.gcd(rem, m)
	}
}

// Eval() returns p(x) where x is the given big integer
func (p Poly) eval(x *big.Int, m *big.Int) (y *big.Int) {
	y = big.NewInt(0)
	accx := big.NewInt(1)
	xd := new(big.Int)
	for i := 0; i <= p.GetDegree(); i++ {
		xd.Mul(accx, p[i])
		y.Add(y, xd)
		accx.Mul(accx, x)
		if m != nil {
			y.Mod(y, m)
			accx.Mod(accx, m)
		}
	}
	return y
}
