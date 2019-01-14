package zkp

import (
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
)

//TestInnerProduct test inner product calculation
func TestInnerProduct(t *testing.T) {
	n := 2
	a := make([]*big.Int, n)
	b := make([]*big.Int, n)

	for i:=0; i<n; i++{
		a[i]= big.NewInt(10)
		b[i]= big.NewInt(20)
	}

	c, _ := innerProduct(a, b)
	assert.Equal(t, big.NewInt(400), c)
}
