package transaction

import (
	"github.com/ninjadotorg/constant/privacy-protocol"
	"math/big"
)

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

// ECDSASigToByteArray converts signature to byte array
func ECDSASigToByteArray(r, s *big.Int) (sig []byte) {
	sig = append(sig, r.Bytes()...)
	sig = append(sig, s.Bytes()...)
	return
}

// FromByteArrayToECDSASig converts a byte array to signature
func FromByteArrayToECDSASig(sig []byte) (r, s *big.Int) {
	r = new(big.Int).SetBytes(sig[0:32])
	s = new(big.Int).SetBytes(sig[32:64])
	return
}
