package privacy

import (
	"fmt"
	"testing"
)

func TestKnapSack(t *testing.T){
	//values := []uint64{1000000000, 10}
	//
	//moneyNeedToSpent := uint64(10)
	//target := uint64(0)
	//for i :=0; i<len(values); i++{
	//	target += values[i]
	//}
	//
	//target -= moneyNeedToSpent
	//
	//choose := Knapsack(values, target)
	//
	//for i, choose := range choose{
	//	if !choose{
	//		fmt.Printf("%v ", values[i])
	//	}
	//}



	//n := 1000
	//
	//values := make([]uint64, n)
	//sum := uint64(0)
	//for i:=0; i<n ; i++{
	//	values[i] = new(big.Int).SetBytes(RandBytes(2)).Uint64()
	//	sum += values[i]
	//}
	//
	//fmt.Printf("Sum values: %v\n", sum)
	//
	//target := uint64(100000)
	//fmt.Printf("Target: %v\n", target)
	//_ = Knapsack(values, target)
	n := 10000

	outCoins := make([]*OutputCoin, n)
	for i:=0; i<n ; i++{
		outCoins[i] = new(OutputCoin).Init()
		//outCoins[i].CoinDetails.Value = new(big.Int).SetBytes(RandBytes(1)).Uint64()
		outCoins[i].CoinDetails.Value = 10
	}

	amount := uint64(20)

	resultOutputCoins := make([]*OutputCoin, 0)
	remainOutputCoins := make([]*OutputCoin, 0)
	totalResultOutputCoinAmount := uint64(0)

	// Calculate sum of all output coins' value
	sumValue := uint64(0)
	valuesKnapsack := make([]uint64, 0)
	outCoinKnapsack := make([]*OutputCoin, 0)
	valuesUnknapsack := make([]uint64, 0)
	outCoinUnknapsack := make([]*OutputCoin, 0)

	for _, outCoin := range outCoins {
		if outCoin.CoinDetails.Value > amount{
			valuesUnknapsack = append(valuesUnknapsack, outCoin.CoinDetails.Value)
			outCoinUnknapsack = append(outCoinUnknapsack, outCoin)
		} else {
			sumValue += outCoin.CoinDetails.Value
			valuesKnapsack = append(valuesKnapsack, outCoin.CoinDetails.Value)
			outCoinKnapsack = append(outCoinKnapsack, outCoin)
		}
	}

	// target
	target := int64(sumValue - amount)
	if target > 0 {
		choices := Knapsack(valuesKnapsack, uint64(target))
		for i, choice := range choices {
			if !choice {
				totalResultOutputCoinAmount += outCoinKnapsack[i].CoinDetails.Value
				resultOutputCoins = append(resultOutputCoins, outCoinKnapsack[i])
			} else {
				remainOutputCoins = append(remainOutputCoins, outCoinKnapsack[i])
			}
		}
		for _, outCoin := range outCoinUnknapsack{
			remainOutputCoins = append(remainOutputCoins, outCoin)
		}
	} else if target == 0{
		totalResultOutputCoinAmount = sumValue
		resultOutputCoins = outCoinKnapsack
		remainOutputCoins = outCoinUnknapsack
	} else{
		if len(outCoinUnknapsack) == 0{
			fmt.Printf("Not enough coin")
		} else{
			indexMin := min(valuesUnknapsack)
			resultOutputCoins = append(resultOutputCoins, outCoinUnknapsack[indexMin])
			totalResultOutputCoinAmount = valuesUnknapsack[indexMin]
			for i, outCoin := range outCoinUnknapsack{
				if i != indexMin{
					remainOutputCoins = append(remainOutputCoins, outCoin)
				}
			}
			for _, outCoin := range outCoinKnapsack{
				remainOutputCoins = append(remainOutputCoins, outCoin)
			}
		}
	}

	fmt.Printf("output all : \n")
	for _, coin := range outCoins{
		fmt.Printf("%v, ", coin.CoinDetails.Value)
	}
	fmt.Printf("\n res: \n")
	for _, coin := range resultOutputCoins{
		fmt.Printf("%v, ", coin.CoinDetails.Value)
	}
	fmt.Printf("\n remain output coin: \n")
	for _, coin := range remainOutputCoins{
		fmt.Printf("%v, ", coin.CoinDetails.Value)
	}
	fmt.Printf("\n \n")


}
