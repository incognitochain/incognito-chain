package blockchain

import (
	"github.com/magiconair/properties/assert"
	"testing"
)

func TestCalculatePortingFees(t *testing.T)  {
	result := calculatePortingFees(3106511852580)
	assert.Equal(t, result, uint64(310651185))
}
