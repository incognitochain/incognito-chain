package transaction

import "github.com/ninjadotorg/constant/privacy-protocol"

func GetInputCoins(usableTx []*Tx) []*privacy.InputCoin {
	var inputCoins []*privacy.InputCoin
	inCoin := new(privacy.InputCoin)

	for _, tx := range usableTx {
		for _, coin := range tx.Proof.OutputCoins {
			inCoin.CoinDetails = coin.CoinDetails
			inputCoins = append(inputCoins, inCoin)
		}
	}
	return inputCoins
}
