package transaction

import "github.com/ninjadotorg/constant/privacy-protocol"

// ConvertOutputCoinToInputCoin - convert output coin from old tx to input coin for new tx
func ConvertOutputCoinToInputCoin(usableTxOfOld []*Tx) []*privacy.InputCoin {
	var inputCoins []*privacy.InputCoin
	inCoin := new(privacy.InputCoin)

	for _, tx := range usableTxOfOld {
		for _, coin := range tx.Proof.OutputCoins {
			inCoin.CoinDetails = coin.CoinDetails
			inputCoins = append(inputCoins, inCoin)
		}
	}
	return inputCoins
}
