package privacy

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestIsPowerOfTwo(t *testing.T) {

	testcase1 := IsPowerOfTwo(64)
	assert.Equal(t, true, testcase1)

	testcase2 := IsPowerOfTwo(124)
	assert.Equal(t, false, testcase2)

	testcase3 := IsPowerOfTwo(0)
	assert.Equal(t, false, testcase3)
}
