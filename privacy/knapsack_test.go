package privacy

import (
	"fmt"
	"math/big"
	"testing"
	"time"
)

func TestKnapSack(t *testing.T){
	start := time.Now()

	values := make([]uint64, 1000)
	for i:=0; i<len(values); i++{
		values[i] = new(big.Int).SetBytes(RandBytes(1)).Uint64()
	}

	moneyNeedToSpent := uint64(350)
	target := uint64(0)
	for i :=0; i<len(values); i++{
		target += values[i]
	}

	target -= moneyNeedToSpent


	choose := Knapsack(values, target)


	fmt.Printf("choose: %v\n", choose)


	for i, choose := range choose{
		if !choose{
			fmt.Printf("%v ", values[i])
		}
	}
	end := time.Since(start)
	fmt.Printf("Knapsack time: %v\n", end)
}
