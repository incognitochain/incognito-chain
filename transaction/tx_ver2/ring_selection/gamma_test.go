package ring_selection

import (
	"fmt"
	"testing"
)

func TestNewGammaPicker(t *testing.T) {
	shape := 0.6168096414830899
	scale := 34565.54623186987
	gp := NewGammaPicker(shape, scale)

	mean := shape * scale
	variant := shape * scale * scale
	fmt.Println(mean, variant)

	for i := 0; i < 1000; i++ {
		s := gp.Rand()
		fmt.Println(i, s)
	}
}
