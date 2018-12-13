package transaction

import (
	"github.com/ninjadotorg/constant/privacy-protocol"
)

// ConvertOutputCoinToInputCoin - convert output coin from old tx to input coin for new tx
func ConvertOutputCoinToInputCoin(usableOutputsOfOld []*privacy.OutputCoin) []*privacy.InputCoin {
	var inputCoins []*privacy.InputCoin
	inCoin := new(privacy.InputCoin)

	for _, coin := range usableOutputsOfOld {
		inCoin.CoinDetails = coin.CoinDetails
		inputCoins = append(inputCoins, inCoin)
	}
	return inputCoins
}
