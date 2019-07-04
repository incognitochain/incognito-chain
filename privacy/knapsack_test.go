package privacy

import (
	"errors"
	"fmt"
	"math/big"
	"math/rand"
	"sort"
	"testing"
	"time"
)

func createRandomTx(seed int) (outCoins []*OutputCoin, amount uint64) {
	n := 100000
	amount = uint64(rand.Uint32())
	outCoins = make([]*OutputCoin, 0)
	for i := 0; i < n; i++ {
		newCoin := new(OutputCoin).Init()
		newCoin.CoinDetails.Value = uint64(rand.Uint32())
		outCoins = append(outCoins, newCoin)
	}
	return outCoins, amount
}

func newBestCoinAlgorithm(outCoins []*OutputCoin, amount uint64) (resultOutputCoins []*OutputCoin, remainOutputCoins []*OutputCoin, totalResultOutputCoinAmount uint64, err error) {
	resultOutputCoins = make([]*OutputCoin, 0)
	remainOutputCoins = make([]*OutputCoin, 0)
	totalResultOutputCoinAmount = uint64(0)

	// either take the smallest coins, or a single largest one
	var outCoinOverLimit *OutputCoin
	outCoinsUnderLimit := make([]*OutputCoin, 0)

	for _, outCoin := range outCoins {
		if outCoin.CoinDetails.Value < amount {
			outCoinsUnderLimit = append(outCoinsUnderLimit, outCoin)
		} else if outCoinOverLimit == nil {
			outCoinOverLimit = outCoin
		} else if outCoinOverLimit.CoinDetails.Value > outCoin.CoinDetails.Value {
			remainOutputCoins = append(remainOutputCoins, outCoin)
		} else {
			remainOutputCoins = append(remainOutputCoins, outCoinOverLimit)
			outCoinOverLimit = outCoin
		}
	}

	sort.Slice(outCoinsUnderLimit, func(i, j int) bool {
		return outCoinsUnderLimit[i].CoinDetails.Value < outCoinsUnderLimit[j].CoinDetails.Value
	})

	for _, outCoin := range outCoinsUnderLimit {
		if totalResultOutputCoinAmount < amount {
			totalResultOutputCoinAmount += outCoin.CoinDetails.Value
			resultOutputCoins = append(resultOutputCoins, outCoin)
		} else {
			remainOutputCoins = append(remainOutputCoins, outCoin)
		}
	}

	if outCoinOverLimit != nil && (outCoinOverLimit.CoinDetails.Value > 2*amount || totalResultOutputCoinAmount < amount) {
		remainOutputCoins = append(remainOutputCoins, resultOutputCoins...)
		resultOutputCoins = []*OutputCoin{outCoinOverLimit}
		totalResultOutputCoinAmount = outCoinOverLimit.CoinDetails.Value
	} else if outCoinOverLimit != nil {
		remainOutputCoins = append(remainOutputCoins, outCoinOverLimit)
	}

	if totalResultOutputCoinAmount < amount {
		return resultOutputCoins, remainOutputCoins, totalResultOutputCoinAmount, errors.New("Not enough coin")
	} else {
		return resultOutputCoins, remainOutputCoins, totalResultOutputCoinAmount, nil
	}
}

func oldBestCoinAlgorithm(outCoins []*OutputCoin, amount uint64) (resultOutputCoins []*OutputCoin, remainOutputCoins []*OutputCoin, totalResultOutputCoinAmount uint64, err error) {
	resultOutputCoins = make([]*OutputCoin, 0)
	remainOutputCoins = make([]*OutputCoin, 0)
	totalResultOutputCoinAmount = uint64(0)

	// just choose output coins have value less than amount for Knapsack algorithm
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

	// if target > 1000, using Greedy algorithm
	// if target > 0, using Knapsack algorithm to choose coins
	// if target == 0, coins need to be spent is coins for Knapsack, we don't need to run Knapsack to find solution
	// if target < 0, instead of using Knapsack, we get the coin that has value is minimum in list unKnapsack coins
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
			return resultOutputCoins, remainOutputCoins, totalResultOutputCoinAmount, errors.New("Not enough coin")
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

	return resultOutputCoins, remainOutputCoins, totalResultOutputCoinAmount, nil
}

func TestAlgorithms(t *testing.T) {
	tot1, tot2 := 0, 0
	outCoins, _ := createRandomTx(1000000)
	for i := 1; i <= 100; i++ {
		_, amount := createRandomTx(i)
		res1, rem1, _, err1 := oldBestCoinAlgorithm(outCoins, amount)
		res2, rem2, _, err2 := newBestCoinAlgorithm(outCoins, amount)

		if err1 == nil && err2 == nil && len(res1)+len(rem1) != len(res2)+len(rem2) {
			t.Errorf("Incorrect length: %v %v", len(res1)+len(rem1), len(res2)+len(rem2))
		}

		if err1 != nil && err2 == nil || err1 == nil && err2 != nil {
			t.Errorf("Incorrect feedback")
		}

		tot1 += len(res1)
		tot2 += len(res2)
	}
	Logger.Log.Infof("%v %v\n", tot1, tot2)
}

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

	Logger.Log.Infof("Target: %v\n", target)

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
			Logger.Log.Infof("Not enough coin")
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
	_ = totalResultOutputCoinAmount
	Logger.Log.Infof("output all : \n")
	for _, coin := range outCoins {
		Logger.Log.Infof("%v, ", coin.CoinDetails.Value)
	}
	Logger.Log.Infof("\n res: \n")
	for _, coin := range resultOutputCoins {
		Logger.Log.Infof("%v, ", coin.CoinDetails.Value)
	}
	Logger.Log.Infof("\n remain output coin: \n")
	for _, coin := range remainOutputCoins {
		Logger.Log.Infof("%v, ", coin.CoinDetails.Value)
	}
	Logger.Log.Infof("\n \n")
}

func TestGreedy(t *testing.T) {
	n := 1000

	outCoins := make([]*OutputCoin, n)
	values := make([]uint64, 0)

	for i := 0; i < n; i++ {
		outCoins[i] = new(OutputCoin).Init()
		//outCoins[i].CoinDetails.Value = new(big.Int).SetBytes(RandBytes(1)).Uint64()
		outCoins[i].CoinDetails.Value = new(big.Int).Add(new(big.Int).SetBytes(RandBytes(2)), big.NewInt(1)).Uint64()
		values = append(values, outCoins[i].CoinDetails.Value)
		Logger.Log.Infof("%v ", outCoins[i].CoinDetails.Value)
	}
	fmt.Println()

	amount := uint64(2000)

	start := time.Now()
	choices := Greedy(outCoins, amount)
	end := time.Since(start)
	Logger.Log.Infof("Greedy time: %v\n", end)

	for i, choice := range choices {
		if choice {
			Logger.Log.Infof("%v ", outCoins[i].CoinDetails.Value)
		} else {
			break
		}
	}
}
