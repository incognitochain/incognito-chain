package privacy

//Todo: 0xKraken

// max returns maximum of two unsigned integer 64 bits
func max(a uint64, b uint64) uint64 {
	if a > b {
		return a
	} else {
		return b
	}
}

// knapsackDPAlg solves Knapsack problem using dynamic programing to collect unspent output coins for spending protocol
func Knapsack(values []uint64, target uint64) []bool {
	n := len(values)
	choices := make([]bool, n)

	K := make([][]uint64, n+1)

	for i := 0; i <= n; i++ {
		K[i] = make([]uint64, target+1)
		for w := uint64(0); w <= target; w++ {
			if i == 0 || w == 0 {
				K[i][w] = 0
			} else if values[i-1] <= w {
				K[i][w] = max(1 + K[i-1][w-values[i-1]], K[i-1][w])
			} else {
				K[i][w] = K[i-1][w]
			}
		}
	}

	targetResult := K[n][target]

	w := target
	for i := n; i > 0 && targetResult > 0; i-- {
		if targetResult == K[i-1][w] {
			choices[i-1] = false
		} else {
			choices[i-1]= true
			targetResult -= 1
			w = w - values[i-1]
		}
	}

	return choices
}
