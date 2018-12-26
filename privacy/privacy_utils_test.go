package privacy

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"math"
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

func TestRandInt(t *testing.T) {
	for {
		a := RandInt()
		threshold_test := 0.01
		length := a.BitLen()
		zero_count := 0
		one_count := 0
		for i := 0; i < length; i++ {
			if (a.Bit(i) == 1) {
				one_count++
			}
			if (a.Bit(i) == 0) {
				zero_count++
			}
		}
		fmt.Println(zero_count, one_count)
		if math.Abs(1-float64(zero_count)/float64(one_count)) <= threshold_test {
			fmt.Println("true")
			break
		}
		fmt.Println("false")
	}
}
