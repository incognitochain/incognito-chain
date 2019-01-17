package privacy

import (
	"fmt"
	"math/big"
	"sort"
	"testing"
	"time"
)

func TestKnapSack(t *testing.T) {
	n := 100

	outCoins := make([]*OutputCoin, n)
	for i := 0; i < n-3; i++ {
		outCoins[i] = new(OutputCoin).Init()
		//outCoins[i].CoinDetails.Value = new(big.Int).SetBytes(RandBytes(1)).Uint64()
		outCoins[i].CoinDetails.Value = 1
	}

	outCoins[n-3] = new(OutputCoin).Init()
	outCoins[n-3].CoinDetails.Value = 270
	outCoins[n-2] = new(OutputCoin).Init()
	outCoins[n-2].CoinDetails.Value = 230
	outCoins[n-1] = new(OutputCoin).Init()
	outCoins[n-1].CoinDetails.Value = 300

	amount := uint64(200)

	resultOutputCoins := make([]*OutputCoin, 0)
	remainOutputCoins := make([]*OutputCoin, 0)
	totalResultOutputCoinAmount := uint64(0)

	// Calculate sum of all output coins' value
	sumValueKnapsack := uint64(0)
	valuesKnapsack := make([]uint64, 0)
	outCoinKnapsack := make([]*OutputCoin, 0)
	outCoinUnknapsack := make([]*OutputCoin, 0)

	for _, outCoin := range outCoins {
		if outCoin.CoinDetails.Value > amount {
			outCoinUnknapsack = append(outCoinUnknapsack, outCoin)
		} else {
			sumValueKnapsack += outCoin.CoinDetails.Value
			valuesKnapsack = append(valuesKnapsack, outCoin.CoinDetails.Value)
			outCoinKnapsack = append(outCoinKnapsack, outCoin)
		}
	}

	// target
	target := int64(sumValueKnapsack - amount)

	fmt.Printf("Target: %v\n", target)

	if target > 1000 {
		choices := Greedy(outCoins, amount)
		for i, choice := range choices {
			if choice {
				totalResultOutputCoinAmount += outCoins[i].CoinDetails.Value
				resultOutputCoins = append(resultOutputCoins, outCoins[i])
			} else {
				remainOutputCoins = append(remainOutputCoins, outCoins[i])
			}
		}
	} else if target > 0 {
		choices := Knapsack(valuesKnapsack, uint64(target))
		for i, choice := range choices {
			if !choice {
				totalResultOutputCoinAmount += outCoinKnapsack[i].CoinDetails.Value
				resultOutputCoins = append(resultOutputCoins, outCoinKnapsack[i])
			} else {
				remainOutputCoins = append(remainOutputCoins, outCoinKnapsack[i])
			}
		}
		remainOutputCoins = append(remainOutputCoins, outCoinUnknapsack...)
	} else if target == 0 {
		totalResultOutputCoinAmount = sumValueKnapsack
		resultOutputCoins = outCoinKnapsack
		remainOutputCoins = outCoinUnknapsack
	} else {
		if len(outCoinUnknapsack) == 0 {
			fmt.Printf("Not enough coin")
		} else {
			sort.Slice(outCoinUnknapsack, func(i, j int) bool {
				return outCoinUnknapsack[i].CoinDetails.Value < outCoinUnknapsack[j].CoinDetails.Value
			})
			resultOutputCoins = append(resultOutputCoins, outCoinUnknapsack[0])
			totalResultOutputCoinAmount = outCoinUnknapsack[0].CoinDetails.Value
			for i := 1; i < len(outCoinUnknapsack); i++ {
				remainOutputCoins = append(remainOutputCoins, outCoinUnknapsack[i])
			}
			remainOutputCoins = append(remainOutputCoins, outCoinKnapsack...)
		}
	}

	fmt.Printf("output all : \n")
	for _, coin := range outCoins {
		fmt.Printf("%v, ", coin.CoinDetails.Value)
	}
	fmt.Printf("\n res: \n")
	for _, coin := range resultOutputCoins {
		fmt.Printf("%v, ", coin.CoinDetails.Value)
	}
	fmt.Printf("\n remain output coin: \n")
	for _, coin := range remainOutputCoins {
		fmt.Printf("%v, ", coin.CoinDetails.Value)
	}
	fmt.Printf("\n \n")
}

func TestGreedy(t *testing.T) {
	n := 1000000

	outCoins := make([]*OutputCoin, n)
	values := make([]uint64, 0)
	for i := 0; i < n; i++ {
		outCoins[i] = new(OutputCoin).Init()
		//outCoins[i].CoinDetails.Value = new(big.Int).SetBytes(RandBytes(1)).Uint64()
		outCoins[i].CoinDetails.Value = new(big.Int).Add(new(big.Int).SetBytes(RandBytes(1)), big.NewInt(1)).Uint64()
		values = append(values, outCoins[i].CoinDetails.Value)
	}

	amount := uint64(20)

	start := time.Now()

	choices := Greedy(outCoins, amount)

	end := time.Since(start)
	fmt.Printf("Greedy time: %v\n", end)

	for i, choice := range choices {
		if choice {
			fmt.Printf("%v ", outCoins[i].CoinDetails.Value)
		} else {
			break
		}
	}
}
