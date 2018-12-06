package blockchain

const Decimals = uint64(10000) // Each float number is multiplied by this value to store as uint64

func GetInterestAmount(principle, interestRate uint64) uint64 {
	return principle * interestRate / Decimals
}
