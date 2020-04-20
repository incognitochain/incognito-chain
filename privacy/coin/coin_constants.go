package coin

const (
	MaxSizeInfoCoin = 255
	CoinVersion1    = 1
	CoinVersion2    = 2
)

func getMin(a, b int) int {
	if a < b {
		return a
	}
	return b
}
