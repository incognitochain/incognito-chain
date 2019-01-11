package privacy

import (
	"fmt"
	"testing"
)

func TestKnapSack(t *testing.T){
	values := []uint64{10, 23, 3, 1, 56, 9}

	moneyNeedToSpent := uint64(34)
	target := uint64(0)
	for i :=0; i<len(values); i++{
		target += values[i]
	}

	target -= moneyNeedToSpent

	choose := Knapsack(values, target)

	for i, choose := range choose{
		if !choose{
			fmt.Printf("%v ", values[i])
		}
	}
}
